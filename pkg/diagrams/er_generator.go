package diagrams

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"
)

// Entity represents a node in the ER diagram
type Entity struct {
	ID         string
	Name       string
	ExternalID string
}

// Relationship represents an edge in the ER diagram
type Relationship struct {
	ID            string
	DisplayName   string
	FromEntity    string
	ToEntity      string
	FromAttribute string
	ToAttribute   string
	FromIsUnique  bool
	ToIsUnique    bool
	PathBased     bool
}

// ERDiagramGenerator handles the generation of ER diagrams
type ERDiagramGenerator struct {
	Definition    *models.SORDefinition
	Entities      map[string]Entity
	Relationships []Relationship
}

// NewERDiagramGenerator creates a new ERDiagramGenerator instance
func NewERDiagramGenerator(definition *models.SORDefinition) *ERDiagramGenerator {
	return &ERDiagramGenerator{
		Definition:    definition,
		Entities:      make(map[string]Entity),
		Relationships: []Relationship{},
	}
}

// IsGraphvizAvailable checks if the user has graphviz (dot) installed
// Using a variable to enable mocking in tests
var IsGraphvizAvailable = func() bool {
	cmd := execCommand("which", "dot")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// execCommand holds exec.Command function to allow mocking in tests
var execCommand = exec.Command

// createTemp holds os.CreateTemp function to allow mocking in tests
var createTemp = os.CreateTemp

// GenerateERDiagram creates an ER diagram from the SOR definition
// If Graphviz is available, it generates an SVG file directly
// Otherwise, it generates just a DOT file
func GenerateERDiagram(def *models.SORDefinition, outputPath string) error {
	generator := NewERDiagramGenerator(def)
	return generator.Generate(outputPath)
}

// Generate creates the ER diagram as a DOT file
func (g *ERDiagramGenerator) Generate(outputPath string) error {
	// Create the output directory if needed
	err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Extract entity and relationship data
	g.extractEntities()
	g.extractRelationships()

	// Create a graph using dominikbraun/graph
	entityGraph := graph.New(graph.StringHash, graph.Directed())

	// Add entities as vertices with attributes for styling
	for id, entity := range g.Entities {
		// Create vertex attribute map for styling
		attributes := map[string]string{
			"label":     entity.Name,
			"shape":     "ellipse",
			"style":     "filled",
			"fillcolor": "#AED6F1",
			"color":     "#2c3e50",
			// "fontcolor": "white",
			// "fontname":  "Arial",
			// "fontsize":  "14",
			// "penwidth":  "1.5",
			// "width":     "2.0",
			// "height":    "1.0",
		}

		// Add vertex with attributes
		err := entityGraph.AddVertex(id, graph.VertexAttributes(attributes))
		if err != nil {
			return fmt.Errorf("failed to add vertex for entity %s: %w", id, err)
		}
	}

	// Add relationships as edges
	// Use a map to prevent duplicate edges between the same entities
	edgeMap := make(map[string]bool)

	for _, rel := range g.Relationships {
		// Create a unique key for this edge
		edgeKey := fmt.Sprintf("%s->%s", rel.FromEntity, rel.ToEntity)

		// Skip if we've already added an edge between these entities
		if edgeMap[edgeKey] {
			continue
		}

		// Create edge attribute map
		attributes := map[string]string{
			"label":     rel.DisplayName,
			"fontname":  "Arial",
			"fontsize":  "12",
			"fontcolor": "#333333",
			"color":     "#6c7a89",
			"penwidth":  "1.2",
			"dir":       "forward",
			"arrowhead": "normal",
		}

		// Add style attribute for path-based relationships
		if rel.PathBased {
			attributes["style"] = "dashed"
		}

		// Add the edge with attributes
		err := entityGraph.AddEdge(rel.FromEntity, rel.ToEntity, graph.EdgeAttributes(attributes))
		if err != nil {
			return fmt.Errorf("failed to add edge for relationship %s -> %s: %w",
				rel.FromEntity, rel.ToEntity, err)
		}

		// Mark this edge as added
		edgeMap[edgeKey] = true
	}

	// Generate DOT representation with styling for the overall graph
	var dotBuf bytes.Buffer
	err = draw.DOT(entityGraph, &dotBuf,
		// draw.GraphAttribute("rankdir", "LR"),
		draw.GraphAttribute("concentrate", "true"),
		draw.GraphAttribute("splines", "curved"),
		draw.GraphAttribute("overlap", "scalexy"),
		draw.GraphAttribute("nodesep", "0.8"),
		draw.GraphAttribute("ranksep", "1.5"),
		draw.GraphAttribute("label", g.Definition.DisplayName),
		draw.GraphAttribute("fontname", "Arial"),
		draw.GraphAttribute("fontsize", "14"),
		draw.GraphAttribute("pad", "0.5"),
		draw.GraphAttribute("dpi", "72"),
	)
	if err != nil {
		return fmt.Errorf("failed to generate DOT file: %w", err)
	}

	// Ensure the output path has the correct extension based on whether we'll generate SVG or DOT
	isSvgOutput := IsGraphvizAvailable() && filepath.Ext(outputPath) == ".svg"

	// Create a temporary DOT file
	tmpDotFile, err := createTemp("", "er-diagram-*.dot")
	if err != nil {
		return fmt.Errorf("failed to create temporary DOT file: %w", err)
	}
	defer func() { _ = os.Remove(tmpDotFile.Name()) }() // Clean up the temporary file when done

	// Write DOT content to temporary file
	if _, err := tmpDotFile.Write(dotBuf.Bytes()); err != nil {
		return fmt.Errorf("failed to write to temporary DOT file: %w", err)
	}
	if err := tmpDotFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary DOT file: %w", err)
	}

	// If Graphviz is available and we want SVG output, use it to generate the SVG
	if isSvgOutput {
		// Use the dot command to convert DOT to SVG
		cmd := execCommand("dot", "-Tsvg", tmpDotFile.Name(), "-o", outputPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run Graphviz dot command: %w", err)
		}
		return nil
	}

	// If Graphviz isn't available or we explicitly want DOT output, just copy the DOT file
	dotOutputPath := outputPath
	if filepath.Ext(outputPath) != ".dot" {
		// If output path doesn't have .dot extension, change it
		dotOutputPath = strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".dot"
	}

	// Copy the DOT content to the final location
	dotContent, err := os.ReadFile(tmpDotFile.Name())
	if err != nil {
		return fmt.Errorf("failed to read temporary DOT file: %w", err)
	}

	err = os.WriteFile(dotOutputPath, dotContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to write DOT file: %w", err)
	}

	return nil
}

// extractEntities extracts entity information from the definition
func (g *ERDiagramGenerator) extractEntities() {
	for id, entity := range g.Definition.Entities {
		// Use DisplayName if available, otherwise ExternalId
		displayName := entity.DisplayName
		if displayName == "" {
			// Extract just the entity name from ExternalId if it contains a namespace
			parts := strings.Split(entity.ExternalId, "/")
			displayName = parts[len(parts)-1]
		}

		g.Entities[id] = Entity{
			ID:         id,
			Name:       displayName,
			ExternalID: entity.ExternalId,
		}
	}
}

// extractRelationships extracts relationship information
func (g *ERDiagramGenerator) extractRelationships() {
	// Create a map to easily find entities by attribute alias
	aliasToEntity := make(map[string]struct {
		EntityID      string
		AttributeName string
		IsUnique      bool
	})

	// Build the attribute alias map
	for entityID, entity := range g.Definition.Entities {
		for _, attr := range entity.Attributes {
			aliasToEntity[attr.AttributeAlias] = struct {
				EntityID      string
				AttributeName string
				IsUnique      bool
			}{
				EntityID:      entityID,
				AttributeName: attr.ExternalId,
				IsUnique:      attr.UniqueId,
			}
		}
	}

	// Process each relationship
	for relID, relationship := range g.Definition.Relationships {
		// Skip invalid relationships where entities wouldn't be found
		if relationship.FromAttribute == "" || relationship.ToAttribute == "" {
			if len(relationship.Path) == 0 {
				// Skip this relationship as it's missing required attributes
				continue
			}
		}

		// Handle direct relationships
		if len(relationship.Path) == 0 {
			fromAttr, fromOK := aliasToEntity[relationship.FromAttribute]
			toAttr, toOK := aliasToEntity[relationship.ToAttribute]

			if fromOK && toOK {
				// Skip self-referential relationships to avoid loops
				if fromAttr.EntityID == toAttr.EntityID {
					continue
				}

				displayName := relationship.DisplayName
				if displayName == "" {
					displayName = relationship.Name
				}

				g.Relationships = append(g.Relationships, Relationship{
					ID:            relID,
					DisplayName:   displayName,
					FromEntity:    fromAttr.EntityID,
					ToEntity:      toAttr.EntityID,
					FromAttribute: fromAttr.AttributeName,
					ToAttribute:   toAttr.AttributeName,
					FromIsUnique:  fromAttr.IsUnique,
					ToIsUnique:    toAttr.IsUnique,
					PathBased:     false,
				})
			}
		} else if len(relationship.Path) > 0 {
			// For path-based relationships, simplify to avoid potential recursion issues
			// Only process if we have at least one valid path element
			if len(relationship.Path) > 0 {
				// Get the first relationship in the path
				firstRelName := relationship.Path[0].Relationship
				firstRel, exists := g.Definition.Relationships[firstRelName]
				if !exists {
					continue
				}

				// Get source and target entities
				fromInfo, fromOK := aliasToEntity[firstRel.FromAttribute]

				// Get the last relationship in the path for target info
				lastRelName := relationship.Path[len(relationship.Path)-1].Relationship
				lastRel, exists := g.Definition.Relationships[lastRelName]
				if !exists {
					continue
				}

				toInfo, toOK := aliasToEntity[lastRel.ToAttribute]

				if fromOK && toOK {
					// Skip self-referential relationships to avoid loops
					if fromInfo.EntityID == toInfo.EntityID {
						continue
					}

					displayName := relationship.DisplayName
					if displayName == "" {
						displayName = relationship.Name
					}

					g.Relationships = append(g.Relationships, Relationship{
						ID:            relID,
						DisplayName:   displayName,
						FromEntity:    fromInfo.EntityID,
						ToEntity:      toInfo.EntityID,
						FromAttribute: fromInfo.AttributeName,
						ToAttribute:   toInfo.AttributeName,
						FromIsUnique:  fromInfo.IsUnique,
						ToIsUnique:    toInfo.IsUnique,
						PathBased:     true,
					})
				}
			}
		}
	}
}
