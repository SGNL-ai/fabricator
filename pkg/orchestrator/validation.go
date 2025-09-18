package orchestrator

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SGNL-ai/fabricator/pkg/fabricator"
	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/SGNL-ai/fabricator/pkg/generators/pipeline"
	"github.com/SGNL-ai/fabricator/pkg/parser"
)

// ValidationOptions configures the validation process
type ValidationOptions struct {
	GenerateDiagram bool
}

// ValidationResult contains the results of validation-only mode
type ValidationResult struct {
	FilesValidated   int
	RecordsValidated int
	ValidationErrors []string
	DiagramGenerated bool
	DiagramPath      string
}

// RunValidation orchestrates the validation-only workflow
func RunValidation(def *parser.SORDefinition, outputDir string, options ValidationOptions) (*ValidationResult, error) {
	result := &ValidationResult{}

	// Create graph from definition to get statistics
	graphInterface, err := model.NewGraph(def)
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

	// Use ValidationProcessor to load and validate CSV files
	processor := pipeline.NewValidationProcessor()
	validationErrors, err := processor.ValidateExistingCSVFiles(def, outputDir)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Count files and records validated
	result.ValidationErrors = validationErrors
	result.FilesValidated, result.RecordsValidated = countValidatedData(outputDir)

	// Generate ER diagram if requested
	if options.GenerateDiagram {
		diagramPath, err := generateERDiagram(def, outputDir)
		if err == nil {
			result.DiagramGenerated = true
			result.DiagramPath = diagramPath
		}
	}

	return result, nil
}

// countValidatedData counts CSV files and records in the directory
func countValidatedData(directory string) (int, int) {
	files, err := os.ReadDir(directory)
	if err != nil {
		return 0, 0
	}

	filesCount := 0
	recordsCount := 0

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".csv" {
			filesCount++

			// Count records in this CSV file
			csvPath := filepath.Join(directory, file.Name())
			if csvFile, err := os.Open(csvPath); err == nil {
				if reader := csv.NewReader(csvFile); reader != nil {
					if records, err := reader.ReadAll(); err == nil && len(records) > 1 {
						recordsCount += len(records) - 1 // Exclude header row
					}
				}
				_ = csvFile.Close()
			}
		}
	}

	return filesCount, recordsCount
}
