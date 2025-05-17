package generators

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/dominikbraun/graph"
	"github.com/fatih/color"
	"github.com/google/uuid"
)

// CSVGenerator handles the generation of CSV files
type CSVGenerator struct {
	OutputDir          string
	DataVolume         int
	EntityData         map[string]*models.CSVData
	idMap              map[string]map[string]string // Maps between entities for consistent relationships
	relationshipMap    map[string][]models.RelationshipLink
	generatedValues    map[string][]string         // Store generated values by type
	namespacePrefix    string                      // Store the common namespace prefix
	AutoCardinality    bool                        // Whether to enable automatic cardinality detection
	usedUniqueValues   map[string]map[string]bool  // Track used values for fields with uniqueId=true by entity:attribute
	uniqueIdAttributes map[string][]string         // Track attributes with uniqueId=true by entity
	existingFiles      map[string]bool             // Track existing CSV files in validation-only mode
	dependencyGraph    graph.Graph[string, string] // Entity dependency graph for topological generation
}

// NewCSVGenerator creates a new CSVGenerator instance
func NewCSVGenerator(outputDir string, dataVolume int, autoCardinality bool) *CSVGenerator {
	// Set a seed for consistent generation
	seed := time.Now().UnixNano()
	// In Go 1.20+ rand.Seed is deprecated but we'll use it for compatibility
	// with older Go versions
	gofakeit.Seed(seed)

	return &CSVGenerator{
		OutputDir:          outputDir,
		DataVolume:         dataVolume,
		EntityData:         make(map[string]*models.CSVData),
		idMap:              make(map[string]map[string]string),
		relationshipMap:    make(map[string][]models.RelationshipLink),
		generatedValues:    make(map[string][]string),
		AutoCardinality:    autoCardinality,
		usedUniqueValues:   make(map[string]map[string]bool),
		uniqueIdAttributes: make(map[string][]string),
		existingFiles:      make(map[string]bool),
		dependencyGraph:    nil, // Will be initialized in Setup
	}
}

// Setup prepares the generator with the necessary data
func (g *CSVGenerator) Setup(entities map[string]models.Entity, relationships map[string]models.Relationship) error {
	// Extract the common namespace prefix
	g.extractNamespacePrefix(entities)

	// Initialize entity data and track unique attributes
	for id, entity := range entities {
		csvData := &models.CSVData{
			ExternalId:  entity.ExternalId,
			EntityName:  entity.DisplayName,
			Description: entity.Description,
		}

		// Extract headers from attributes and track unique attributes
		headers := []string{}
		g.uniqueIdAttributes[id] = []string{} // Initialize the slice for this entity
		g.usedUniqueValues[id] = make(map[string]bool)

		for _, attr := range entity.Attributes {
			headers = append(headers, attr.ExternalId)

			// Track attributes with uniqueId=true
			if attr.UniqueId {
				g.uniqueIdAttributes[id] = append(g.uniqueIdAttributes[id], attr.ExternalId)
				// Initialize the map to track used values for this attribute
				attrKey := id + ":" + attr.ExternalId
				if g.usedUniqueValues[attrKey] == nil {
					g.usedUniqueValues[attrKey] = make(map[string]bool)
				}
			}
		}
		csvData.Headers = headers

		g.EntityData[id] = csvData
		g.idMap[id] = make(map[string]string)
	}

	// Process relationships
	g.processRelationships(entities, relationships)

	// Build entity dependency graph for generation order
	dependencyGraph, err := g.buildEntityDependencyGraph(entities, relationships)
	if err != nil {
		return fmt.Errorf("failed to build entity dependency graph: %w", err)
	}

	// Store the dependency graph for use during generation
	g.dependencyGraph = dependencyGraph

	// Log the generation order (needed for later, but don't display)
	_, err = g.getTopologicalOrder(dependencyGraph)
	if err != nil {
		return fmt.Errorf("failed to determine entity generation order: %w", err)
	}

	// Pre-generate some common values for generic data types
	g.generateCommonValues()

	// Identify existing CSV files in the output directory
	files, err := os.ReadDir(g.OutputDir)
	if err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".csv") {
				g.existingFiles[file.Name()] = true
			}
		}
	}

	return nil
}

// extractNamespacePrefix finds the common namespace prefix from entity external IDs
func (g *CSVGenerator) extractNamespacePrefix(entities map[string]models.Entity) {
	// Find the first entity with a prefix in its externalId
	for _, entity := range entities {
		if strings.Contains(entity.ExternalId, "/") {
			parts := strings.Split(entity.ExternalId, "/")
			if len(parts) > 0 {
				g.namespacePrefix = parts[0]
			}
			break
		}
	}
}

// processRelationships analyzes the relationships between entities
func (g *CSVGenerator) processRelationships(entities map[string]models.Entity, relationships map[string]models.Relationship) {
	// Create maps to find entities and attributes by different identifiers
	// Map of attribute alias to (entity ID, attribute name, uniqueId)
	attributeAliasMap := make(map[string]struct {
		EntityID      string
		AttributeName string
		UniqueID      bool
	})

	// Map to lookup entities/attributes by "Entity.Attribute" pattern (for YAML without attributeAlias)
	entityAttributeMap := make(map[string]struct {
		EntityID      string
		AttributeName string
		UniqueID      bool
	})

	// Build the attribute maps
	for entityID, entity := range entities {
		for _, attr := range entity.Attributes {
			// Handle attributeAlias case (when it exists)
			if attr.AttributeAlias != "" {
				attributeAliasMap[attr.AttributeAlias] = struct {
					EntityID      string
					AttributeName string
					UniqueID      bool
				}{
					EntityID:      entityID,
					AttributeName: attr.Name,
					UniqueID:      attr.UniqueId,
				}
			}

			// Also build Entity.Attribute map for YAMLs without attributeAlias
			entityKey := entity.ExternalId + "." + attr.ExternalId
			entityAttributeMap[entityKey] = struct {
				EntityID      string
				AttributeName string
				UniqueID      bool
			}{
				EntityID:      entityID,
				AttributeName: attr.Name,
				UniqueID:      attr.UniqueId,
			}
		}
	}

	// Process each relationship
	for _, relationship := range relationships {
		// Skip path-based relationships for now
		if len(relationship.Path) > 0 {
			continue
		}

		var fromAttr, toAttr struct {
			EntityID      string
			AttributeName string
			UniqueID      bool
		}
		var fromOk, toOk bool

		// Try to match using attribute aliases first
		if relationship.FromAttribute != "" && relationship.ToAttribute != "" {
			fromAttr, fromOk = attributeAliasMap[relationship.FromAttribute]
			toAttr, toOk = attributeAliasMap[relationship.ToAttribute]
		}

		// If attribute alias lookup fails, try Entity.Attribute pattern
		if !fromOk || !toOk {
			// Handle the format used in SW-Assertions-Only-0.1.0.yaml
			// Relationships defined like "Entity.attribute"
			if strings.Contains(relationship.FromAttribute, ".") && strings.Contains(relationship.ToAttribute, ".") {
				fromAttr, fromOk = entityAttributeMap[relationship.FromAttribute]
				toAttr, toOk = entityAttributeMap[relationship.ToAttribute]
			}
		}

		// If we found both the from and to attributes, create the relationship
		if fromOk && toOk {
			// Add the relationship link with uniqueId information
			link := models.RelationshipLink{
				FromEntityID:      fromAttr.EntityID,
				ToEntityID:        toAttr.EntityID,
				FromAttribute:     fromAttr.AttributeName,
				ToAttribute:       toAttr.AttributeName,
				IsFromAttributeID: fromAttr.UniqueID,
				IsToAttributeID:   toAttr.UniqueID,
			}

			g.relationshipMap[fromAttr.EntityID] = append(g.relationshipMap[fromAttr.EntityID], link)
		}
	}
}

// LoadExistingCSVFiles loads existing CSV files from the output directory for validation
func (g *CSVGenerator) LoadExistingCSVFiles() error {
	// Check if output directory exists
	if _, err := os.Stat(g.OutputDir); os.IsNotExist(err) {
		return fmt.Errorf("output directory does not exist: %s", g.OutputDir)
	}

	// Load CSV files in the output directory
	files, err := os.ReadDir(g.OutputDir)
	if err != nil {
		return fmt.Errorf("failed to read output directory: %w", err)
	}

	// Map to track which entities have been loaded
	foundEntities := make(map[string]bool)

	// Counter for successfully loaded files
	loadedFiles := 0

	// Find and load CSV files for each entity
	for entityID, entityData := range g.EntityData {
		// Determine the expected filename for this entity
		entityFilename := GetEntityFileName(entityData)
		entityLoaded := false

		// Look for an existing CSV file matching this entity
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".csv") {
				continue
			}

			// Check if this is the file for the current entity
			if file.Name() == entityFilename {
				// Read and parse the CSV file
				filePath := filepath.Join(g.OutputDir, file.Name())
				fileData, err := os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("failed to read CSV file %s: %w", filePath, err)
				}

				// Parse the CSV content
				reader := csv.NewReader(strings.NewReader(string(fileData)))
				records, err := reader.ReadAll()
				if err != nil {
					return fmt.Errorf("failed to parse CSV file %s: %w", filePath, err)
				}

				// Verify we have at least a header row
				if len(records) == 0 {
					color.Yellow("Warning: Empty CSV file for entity %s (%s)", entityID, filePath)
					continue
				}

				// Update entity data with parsed CSV
				entityData.Headers = records[0]
				entityData.Rows = records[1:]

				foundEntities[entityID] = true
				entityLoaded = true
				loadedFiles++

				color.Green("✓ Loaded %s: %d rows", entityFilename, len(entityData.Rows))
				break
			}
		}

		if !entityLoaded {
			color.Yellow("Warning: No CSV file found for entity %s (expected %s)", entityID, entityFilename)
		}
	}

	// Check if any files were loaded
	if loadedFiles == 0 {
		return fmt.Errorf("no matching CSV files found in directory: %s", g.OutputDir)
	}

	// Report loading summary
	color.Green("Successfully loaded %d CSV files for validation", loadedFiles)
	if len(foundEntities) < len(g.EntityData) {
		color.Yellow("Warning: %d entities do not have corresponding CSV files",
			len(g.EntityData)-len(foundEntities))
	}

	return nil
}

// generateCommonValues pre-generates common test data values

// GenerateData creates random data for all entities
// following a topological ordering of dependencies
func (g *CSVGenerator) GenerateData() error {
	// First, generate IDs that will be consistent across relationships
	g.generateConsistentIds()

	// If the dependency graph wasn't built during setup, this is an error
	if g.dependencyGraph == nil {
		return fmt.Errorf("dependency graph not initialized; Setup must be called before GenerateData")
	}

	// Get a topological ordering of entities for generation
	// This ensures that entities are generated in dependency order
	entityOrder, err := g.getTopologicalOrder(g.dependencyGraph)
	if err != nil {
		return fmt.Errorf("failed to determine entity generation order: %w", err)
	}

	// Generate data for each entity in topological order
	for _, entityID := range entityOrder {
		csvData := g.EntityData[entityID]
		rows := [][]string{}

		// Generating data (removed from output)
		for i := 0; i < g.DataVolume; i++ {
			row := g.generateRowForEntity(entityID, i)
			rows = append(rows, row)
		}

		csvData.Rows = rows

		// Make relationships consistent after each entity is generated
		// This ensures that dependent entities have correct references
		g.makeRelationshipsConsistentForEntity(entityID)
	}

	return nil
}

// generateConsistentIds ensures that IDs used in relationships are consistent
func (g *CSVGenerator) generateConsistentIds() {
	// Initialize the random number generator with local source
	// No need to call Seed as of Go 1.20+

	// Generate consistent IDs for each entity
	for entityID := range g.EntityData {
		for i := 0; i < g.DataVolume; i++ {
			g.idMap[entityID][strconv.Itoa(i)] = uuid.New().String()
		}
	}
}

// generateRowForEntity creates a row of data for a specific entity
func (g *CSVGenerator) generateRowForEntity(entityID string, index int) []string {
	// Get the entity info
	csvData := g.EntityData[entityID]
	headers := csvData.Headers

	row := make([]string, len(headers))

	// Use pre-generated IDs to ensure consistency across relationships
	idKey := strconv.Itoa(index)
	preGenID := ""
	if id, exists := g.idMap[entityID][idKey]; exists {
		preGenID = id
	} else {
		preGenID = uuid.New().String()
		// Store it for future reference
		g.idMap[entityID][idKey] = preGenID
	}

	for i, header := range headers {
		// Check if this attribute is marked as uniqueId=true
		isUnique := false
		for _, uniqueAttr := range g.uniqueIdAttributes[entityID] {
			if header == uniqueAttr {
				isUnique = true
				break
			}
		}

		// For primary key field (determined by uniqueId flag)
		// Use the pre-generated ID for this entity/row
		if isUnique {
			row[i] = preGenID
			continue
		}

		// For all other fields, use our field generator
		fieldType := DetectFieldType(header)

		req := FieldRequest{
			EntityID:    entityID,
			Header:      header,
			HeaderIndex: i,
			RowIndex:    index,
			IsUnique:    isUnique,
		}

		row[i] = g.GenerateFieldValue(req, fieldType)
	}

	return row
}

// ensureRelationshipConsistency enforces consistency across related entities
func (g *CSVGenerator) ensureRelationshipConsistency() {
	// Iterate through relationships and ensure consistency
	for fromEntityID, links := range g.relationshipMap {
		for _, link := range links {
			g.makeRelationshipsConsistent(fromEntityID, link)
		}
	}
}

// makeRelationshipsConsistent ensures that related entities have consistent values
func (g *CSVGenerator) makeRelationshipsConsistent(fromEntityID string, link models.RelationshipLink) {
	// Process relationship via the relationship handler
	ctx := NewRelationshipContext(g, fromEntityID, link)
	
	// If we couldn't create a valid context, stop processing
	if ctx == nil {
		return
	}
	
	// Handle relationship based on its type (one-to-many, many-to-one, etc.)
	// The handler will update the entity data directly
	g.handleRelationship(ctx)
}



// WriteCSVFiles writes all generated data to CSV files
func (g *CSVGenerator) WriteCSVFiles() error {
	// Create the output directory if it doesn't exist
	err := os.MkdirAll(g.OutputDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write each entity's data to a CSV file
	for _, csvData := range g.EntityData {
		// Get the filename based on the entity's external ID
		filename := GetEntityFileName(csvData)

		filePath := filepath.Join(g.OutputDir, filename)

		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", filePath, err)
		}
		defer func() { _ = file.Close() }()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		// Write headers
		err = writer.Write(csvData.Headers)
		if err != nil {
			return fmt.Errorf("failed to write headers to %s: %w", filePath, err)
		}

		// Write data rows
		for _, row := range csvData.Rows {
			err = writer.Write(row)
			if err != nil {
				return fmt.Errorf("failed to write row to %s: %w", filePath, err)
			}
		}

		color.Green("✓ Generated %s with %d rows", filename, len(csvData.Rows))
	}

	return nil
}
