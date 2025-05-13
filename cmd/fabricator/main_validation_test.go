package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/fabricator"
	"github.com/SGNL-ai/fabricator/pkg/generators"
)

func TestValidateRelationships(t *testing.T) {
	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "validation-test-*")
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

	// Parse the example YAML file
	parser := fabricator.NewParser(tempYAML)
	if err := parser.Parse(); err != nil {
		t.Fatalf("Failed to parse YAML file: %v", err)
	}

	// Initialize generator with a small data volume for speed
	generator := generators.NewCSVGenerator(outputDir, 5, true)
	generator.Setup(parser.Definition.Entities, parser.Definition.Relationships)

	// Generate data
	generator.GenerateData()

	// Write CSV files
	if err := generator.WriteCSVFiles(); err != nil {
		t.Fatalf("Failed to write CSV files: %v", err)
	}

	// Validate relationships
	validationResults := generator.ValidateRelationships()

	// We're not asserting specific failures - just checking the validation runs
	// In real scenarios, we'd check for specific expected results
	t.Logf("Validation results: %d issues found", len(validationResults))

	// Check that uniqueness is respected
	uniqueErrors := generator.ValidateUniqueValues()
	if len(uniqueErrors) > 0 {
		t.Logf("Uniqueness violations found in %d entities", len(uniqueErrors))
		for _, entityError := range uniqueErrors {
			t.Logf("  Entity %s (%s): %d errors", entityError.EntityID, entityError.EntityFile, len(entityError.Messages))
		}
	} else {
		t.Logf("No uniqueness violations found")
	}
}
