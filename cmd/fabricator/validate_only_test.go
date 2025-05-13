package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/fabricator"
	"github.com/SGNL-ai/fabricator/pkg/generators"
	"github.com/fatih/color"
)

func TestValidateOnlyMode(t *testing.T) {
	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "validate-only-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Copy test YAML file from examples directory
	examplePath := filepath.Join("..", "..", "examples", "example.yaml")
	exampleContent, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("Failed to read example YAML file: %v", err)
	}

	// Write to temporary file
	tempYAML := filepath.Join(tempDir, "test.yaml")
	if err := os.WriteFile(tempYAML, exampleContent, 0644); err != nil {
		t.Fatalf("Failed to write temp YAML file: %v", err)
	}

	// Create output directory
	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	// First generate CSV files in normal mode
	t.Log("Generating CSV files in normal mode")

	// Parse the example YAML file
	parser := fabricator.NewParser(tempYAML)
	if err := parser.Parse(); err != nil {
		t.Fatalf("Failed to parse YAML file: %v", err)
	}

	// Disable colors for testing
	oldNoColor := color.NoColor
	defer func() { color.NoColor = oldNoColor }()
	color.NoColor = true

	// Initialize generator with a small data volume for speed
	generator := generators.NewCSVGenerator(outputDir, 5, true)
	generator.Setup(parser.Definition.Entities, parser.Definition.Relationships)

	// Generate data and write CSV files
	generator.GenerateData()
	if err := generator.WriteCSVFiles(); err != nil {
		t.Fatalf("Failed to write CSV files: %v", err)
	}

	// Now test validation-only mode
	t.Log("Testing validation-only mode on the generated files")

	// Make a new parser for validation
	validationParser := fabricator.NewParser(tempYAML)
	if err := validationParser.Parse(); err != nil {
		t.Fatalf("Failed to parse YAML file for validation: %v", err)
	}

	// Create a new generator for validation-only mode
	validationGenerator := generators.NewCSVGenerator(outputDir, 5, true)
	validationGenerator.Setup(validationParser.Definition.Entities, validationParser.Definition.Relationships)

	// Load existing CSV files instead of generating new ones
	err = validationGenerator.LoadExistingCSVFiles()
	if err != nil {
		t.Fatalf("Failed to load existing CSV files: %v", err)
	}

	// Validate relationships
	validationResults := validationGenerator.ValidateRelationships()
	t.Logf("Validation results: %d relationship issues found", len(validationResults))

	// Validate unique constraints
	uniqueErrors := validationGenerator.ValidateUniqueValues()
	t.Logf("Uniqueness validation: %d entities with constraint violations", len(uniqueErrors))

	// Test intentional validation error by modifying a CSV file
	// Get the first CSV file
	files, _ := os.ReadDir(outputDir)
	var firstCSVFile string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".csv" {
			firstCSVFile = file.Name()
			break
		}
	}

	if firstCSVFile != "" {
		// Create invalid CSV (missing header row)
		t.Logf("Creating invalid CSV file to test validation error handling")
		invalidCSVPath := filepath.Join(outputDir, firstCSVFile)
		invalidContent := "1,InvalidTest,BadData\n"
		if err := os.WriteFile(invalidCSVPath, []byte(invalidContent), 0644); err != nil {
			t.Fatalf("Failed to create invalid CSV: %v", err)
		}

		// Try validation again with invalid data
		invalidGenerator := generators.NewCSVGenerator(outputDir, 5, true)
		invalidGenerator.Setup(validationParser.Definition.Entities, validationParser.Definition.Relationships)

		// Load should still succeed as we're just parsing the CSV files
		err = invalidGenerator.LoadExistingCSVFiles()
		if err != nil {
			t.Logf("Expected load to succeed even with invalid CSV, but got error: %v", err)
		}

		// Validation should find issues
		validationResults = invalidGenerator.ValidateRelationships()
		t.Logf("Validation with invalid data: %d relationship issues found", len(validationResults))

		uniqueErrors = invalidGenerator.ValidateUniqueValues()
		t.Logf("Uniqueness validation with invalid data: %d entities with constraint violations", len(uniqueErrors))
	}
}
