package pipeline

import (
	"fmt"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
)

// IDGeneratorInterface defines the interface for ID generation
type IDGeneratorInterface interface {
	GenerateIDs(graph *model.Graph, dataVolume int) error
}

// RelationshipLinkerInterface defines the interface for relationship linking
type RelationshipLinkerInterface interface {
	LinkRelationships(graph *model.Graph, autoCardinality bool) error
}

// FieldGeneratorInterface defines the interface for field generation
type FieldGeneratorInterface interface {
	GenerateFields(graph *model.Graph) error
}

// ValidatorInterface defines the interface for graph-level validation
type ValidatorInterface interface {
	ValidateRelationships(graph *model.Graph) []string
}

// CSVWriterInterface defines the interface for writing CSV files
type CSVWriterInterface interface {
	WriteFiles(graph *model.Graph) error
}

// DataGenerator coordinates the entire CSV generation process using the pipeline components
type DataGenerator struct {
	// Pipeline components
	idGenerator        IDGeneratorInterface
	relationshipLinker RelationshipLinkerInterface
	fieldGenerator     FieldGeneratorInterface
	validator          ValidatorInterface
	csvWriter          CSVWriterInterface

	// Configuration
	dataVolume      int
	outputDir       string
	autoCardinality bool
}

// NewDataGenerator creates a new DataGenerator with all pipeline components
func NewDataGenerator(outputDir string, dataVolume int, autoCardinality bool) *DataGenerator {
	return &DataGenerator{
		idGenerator:        NewIDGenerator(),
		relationshipLinker: NewRelationshipLinker(),
		fieldGenerator:     NewFieldGenerator(),
		validator:          NewValidation(),
		csvWriter:          NewCSVWriter(outputDir),

		dataVolume:      dataVolume,
		outputDir:       outputDir,
		autoCardinality: autoCardinality,
	}
}

// Generate executes the full pipeline to generate data for all entities
func (g *DataGenerator) Generate(graph *model.Graph) error {
	// Step 1: Generate all identifier fields in topological order
	if err := g.idGenerator.GenerateIDs(graph, g.dataVolume); err != nil {
		return fmt.Errorf("ID generation failed: %w", err)
	}

	// Step 2: Establish relationship structure between entities
	if err := g.relationshipLinker.LinkRelationships(graph, g.autoCardinality); err != nil {
		return fmt.Errorf("relationship linking failed: %w", err)
	}

	// Step 3: Fill in remaining non-relationship fields
	if err := g.fieldGenerator.GenerateFields(graph); err != nil {
		return fmt.Errorf("field generation failed: %w", err)
	}

	// Optional validation step
	relationshipErrors := g.validator.ValidateRelationships(graph)
	if len(relationshipErrors) > 0 {
		// Just log the validation errors rather than failing
		fmt.Printf("WARNING: Found %d relationship validation errors\n", len(relationshipErrors))
		for _, err := range relationshipErrors {
			fmt.Printf("- %s\n", err)
		}
	}

	// Note: Unique value validation is handled by AddRow during data generation
	// No need for separate unique value validation since AddRow rejects duplicates

	// Write CSV files
	if err := g.csvWriter.WriteFiles(graph); err != nil {
		return fmt.Errorf("CSV file writing failed: %w", err)
	}

	return nil
}
