package orchestrator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SGNL-ai/fabricator/pkg/diagrams"
	"github.com/SGNL-ai/fabricator/pkg/fabricator"
	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/SGNL-ai/fabricator/pkg/generators/pipeline"
	"github.com/SGNL-ai/fabricator/pkg/parser"
	"github.com/SGNL-ai/fabricator/pkg/util"
)

// GenerationOptions configures the data generation process
type GenerationOptions struct {
	DataVolume      int
	AutoCardinality bool
	GenerateDiagram bool
	ValidateResults bool
}

// GenerationResult contains the results of data generation
type GenerationResult struct {
	EntitiesProcessed int
	RecordsPerEntity  int
	TotalRecords      int
	CSVFilesGenerated int
	DiagramGenerated  bool
	DiagramPath       string
	ValidationSummary *ValidationSummary
}

// ValidationSummary contains validation results
type ValidationSummary struct {
	Errors                 []string
	RelationshipIssues     int
	UniqueConstraintIssues int
}

// RunGeneration orchestrates the complete data generation workflow
func RunGeneration(def *parser.SORDefinition, outputDir string, options GenerationOptions) (*GenerationResult, error) {
	result := &GenerationResult{
		RecordsPerEntity: options.DataVolume,
	}

	// Create graph from definition with data volume for memory optimization
	graphInterface, err := model.NewGraph(def, options.DataVolume)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity graph: %w", err)
	}

	graph, ok := graphInterface.(*model.Graph)
	if !ok {
		return nil, fmt.Errorf("failed to convert graph to concrete type")
	}

	// Display parsing statistics from the constructed graph
	statistics := graph.GetStatistics()
	fabricator.PrintGraphStatistics(statistics)

	// Initialize and run the data generation pipeline
	generator := pipeline.NewDataGenerator(outputDir, options.DataVolume, options.AutoCardinality)
	if err := generator.Generate(graph); err != nil {
		return nil, fmt.Errorf("data generation failed: %w", err)
	}

	// Count generated files
	files, err := os.ReadDir(outputDir)
	if err == nil {
		for _, file := range files {
			if filepath.Ext(file.Name()) == ".csv" {
				result.CSVFilesGenerated++
			}
		}
	}

	// Calculate results
	result.EntitiesProcessed = len(def.Entities)
	result.TotalRecords = result.EntitiesProcessed * options.DataVolume

	// Generate ER diagram if requested
	if options.GenerateDiagram {
		diagramPath, err := generateERDiagram(def, outputDir)
		if err == nil {
			result.DiagramGenerated = true
			result.DiagramPath = diagramPath
		}
	}

	// Run validation if requested
	if options.ValidateResults {
		validator := pipeline.NewValidation()
		relationshipErrors := validator.ValidateRelationships(graph)

		result.ValidationSummary = &ValidationSummary{
			Errors:             relationshipErrors,
			RelationshipIssues: len(relationshipErrors),
		}
	}

	return result, nil
}

// generateERDiagram creates an ER diagram for the SOR
func generateERDiagram(def *parser.SORDefinition, outputDir string) (string, error) {
	// Create diagram filename based on SOR name
	diagramName := util.CleanNameForFilename(def.DisplayName)

	// Determine extension based on Graphviz availability
	extension := ".dot"
	if diagrams.IsGraphvizAvailable() {
		extension = ".svg"
	}

	diagramPath := filepath.Join(outputDir, diagramName+extension)

	// Generate the diagram
	err := diagrams.GenerateERDiagram(def, diagramPath)
	if err != nil {
		return "", err
	}

	return diagramPath, nil
}
