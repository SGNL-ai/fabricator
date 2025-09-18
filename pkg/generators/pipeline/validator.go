package pipeline

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/SGNL-ai/fabricator/pkg/models"
)

// CSVLoaderInterface defines the interface for loading CSV files
type CSVLoaderInterface interface {
	LoadCSVFiles(graph *model.Graph, directory string) []string
}

// ValidationProcessorInterface defines the interface for validation-only mode
type ValidationProcessorInterface interface {
	ValidateExistingCSVFiles(def *models.SORDefinition, directory string) ([]string, error)
}

// CSVLoader handles loading existing CSV files into the model
type CSVLoader struct {
	// Configuration options can be added here
}

// ValidationProcessor handles validation-only mode workflows
type ValidationProcessor struct {
	csvLoader CSVLoaderInterface
	validator ValidatorInterface
}

// NewCSVLoader creates a new CSV loader
func NewCSVLoader() CSVLoaderInterface {
	return &CSVLoader{}
}

// NewValidationProcessor creates a new validation processor
func NewValidationProcessor() ValidationProcessorInterface {
	return &ValidationProcessor{
		csvLoader: NewCSVLoader(),
		validator: NewValidation(),
	}
}

// ValidateExistingCSVFiles validates existing CSV files without generating new data
// Returns all validation issues found - does not stop on first error
func (p *ValidationProcessor) ValidateExistingCSVFiles(def *models.SORDefinition, directory string) ([]string, error) {
	var allErrors []string

	// Create graph from definition
	graphInterface, err := model.NewGraph(def)
	if err != nil {
		return []string{fmt.Sprintf("failed to create graph: %v", err)}, nil
	}

	graph, ok := graphInterface.(*model.Graph)
	if !ok {
		return []string{"failed to convert graph to concrete type"}, nil
	}

	// Load existing CSV files into the graph (collect all loading errors)
	loadErrors := p.csvLoader.LoadCSVFiles(graph, directory)
	for _, errMsg := range loadErrors {
		allErrors = append(allErrors, errMsg)
	}

	// Continue validation even if some files failed to load
	// Validate FK relationships using post-generation validation
	for _, entity := range graph.GetAllEntities() {
		fkErrors := entity.ValidateAllForeignKeys()
		for _, errMsg := range fkErrors {
			allErrors = append(allErrors, fmt.Sprintf("entity %s: %s", entity.GetExternalID(), errMsg))
		}
	}

	// Validate graph-level relationships
	relationshipErrors := p.validator.ValidateRelationships(graph)
	allErrors = append(allErrors, relationshipErrors...)

	return allErrors, nil
}

// LoadCSVFiles loads existing CSV files into the graph entities
// Returns all errors found - does not stop on first error
func (l *CSVLoader) LoadCSVFiles(graph *model.Graph, directory string) []string {
	var errors []string

	// Check if directory exists
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("directory %s does not exist", directory))
		return errors
	}

	// Load CSV file for each entity
	for entityID, entity := range graph.GetAllEntities() {
		// Determine CSV filename from entity external ID
		filename := l.getCSVFilename(entity.GetExternalID())
		csvPath := filepath.Join(directory, filename)

		// Check if CSV file exists
		if _, err := os.Stat(csvPath); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("CSV file not found for entity %s: %s", entityID, csvPath))
			continue
		}

		// Load CSV data into entity
		if err := l.loadEntityCSV(entity, csvPath); err != nil {
			errors = append(errors, fmt.Sprintf("failed to load CSV for entity %s: %v", entityID, err))
			continue
		}
	}

	return errors
}

// loadEntityCSV loads a single CSV file into an entity
func (l *CSVLoader) loadEntityCSV(entity model.EntityInterface, csvPath string) error {
	file, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file %s: %w", csvPath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV file %s: %w", csvPath, err)
	}

	if len(records) == 0 {
		return fmt.Errorf("CSV file %s is empty", csvPath)
	}

	// First row is headers
	headers := records[0]
	dataRows := records[1:]

	// Load each data row into the entity
	for i, row := range dataRows {
		if len(row) != len(headers) {
			return fmt.Errorf("CSV file %s row %d has %d columns, expected %d", csvPath, i+1, len(row), len(headers))
		}

		// Create row data map
		rowData := make(map[string]string)
		for j, value := range row {
			rowData[headers[j]] = value
		}

		// Add row to entity (AddRow validation will catch duplicates, etc.)
		if err := entity.AddRow(model.NewRow(rowData)); err != nil {
			return fmt.Errorf("validation failed for CSV file %s row %d: %w", csvPath, i+1, err)
		}
	}

	return nil
}

// getCSVFilename determines the CSV filename from entity external ID
func (l *CSVLoader) getCSVFilename(externalID string) string {
	// Handle namespace format (e.g., "Sample/EntityName" -> "EntityName.csv")
	if strings.Contains(externalID, "/") {
		parts := strings.Split(externalID, "/")
		return parts[len(parts)-1] + ".csv"
	}
	// Simple format (e.g., "EntityName" -> "EntityName.csv")
	return externalID + ".csv"
}