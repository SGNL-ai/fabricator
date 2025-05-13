package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestValidationOnlyMode(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "fabricator-validation-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a minimal YAML file
	yamlContent := `
displayName: Test SOR
description: Test system of record
entities:
  entity1:
    displayName: TestEntity
    externalId: Test/Entity
    description: A test entity
    attributes:
      - name: id
        externalId: id
        uniqueId: true
        attributeAlias: test-id
      - name: name
        externalId: name
        attributeAlias: test-name
`
	yamlPath := filepath.Join(tempDir, "test.yaml")
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test YAML file: %v", err)
	}

	// Create output directory
	outputPath := filepath.Join(tempDir, "output")
	err = os.MkdirAll(outputPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Create a sample CSV file
	csvContent := "id,name\n1,Entity One\n2,Entity Two\n"
	csvPath := filepath.Join(outputPath, "Entity.csv")
	err = os.WriteFile(csvPath, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write sample CSV file: %v", err)
	}

	// Create a test diagram file
	diagramPath := filepath.Join(outputPath, "Test_SOR.dot")
	err = os.WriteFile(diagramPath, []byte("digraph { Test_SOR }"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test diagram file: %v", err)
	}

	// Disable color for testing
	oldNoColor := color.NoColor
	defer func() { color.NoColor = oldNoColor }()
	color.NoColor = true

	// Save and restore validateOnly flag
	oldValidateOnly := validateOnly
	defer func() { validateOnly = oldValidateOnly }()
	validateOnly = true

	// Run the application in validation-only mode
	err = run(yamlPath, outputPath, 10, false)

	// If there was an error, fail the test
	if err != nil {
		t.Fatalf("run() in validation-only mode failed: %v", err)
	}

	// CSV file should remain unchanged
	modifiedCSVContent, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("Failed to read CSV file after validation: %v", err)
	}

	// Check that the CSV file was not modified
	if string(modifiedCSVContent) != csvContent {
		t.Error("CSV file was modified in validation-only mode")
	}
}

func TestValidateOnlyWithMissingCSV(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "fabricator-missing-csv-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a minimal YAML file
	yamlContent := `
displayName: Test SOR
description: Test system of record
entities:
  entity1:
    displayName: TestEntity
    externalId: Test/Entity
    description: A test entity
    attributes:
      - name: id
        externalId: id
        uniqueId: true
        attributeAlias: test-id
`
	yamlPath := filepath.Join(tempDir, "test.yaml")
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test YAML file: %v", err)
	}

	// Create empty output directory (no CSV files)
	outputPath := filepath.Join(tempDir, "empty_output")
	err = os.MkdirAll(outputPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Disable color for testing
	oldNoColor := color.NoColor
	defer func() { color.NoColor = oldNoColor }()
	color.NoColor = true

	// Save and restore validateOnly flag
	oldValidateOnly := validateOnly
	defer func() { validateOnly = oldValidateOnly }()
	validateOnly = true

	// Run the application in validation-only mode
	err = run(yamlPath, outputPath, 10, false)

	// Should get an error (no CSV files to validate)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "no matching csv files") {
		t.Errorf("Expected 'no matching CSV files' error, got: %v", err)
	}
}

func TestValidateOnlyWithInvalidCSV(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "fabricator-invalid-csv-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a minimal YAML file
	yamlContent := `
displayName: Test SOR
description: Test system of record
entities:
  entity1:
    displayName: TestEntity
    externalId: Test/Entity
    description: A test entity
    attributes:
      - name: id
        externalId: id
        uniqueId: true
        attributeAlias: test-id
      - name: name
        externalId: name
        attributeAlias: test-name
  entity2:
    displayName: RelatedEntity
    externalId: Test/Related
    description: A related entity
    attributes:
      - name: id
        externalId: id
        uniqueId: true
        attributeAlias: related-id
      - name: entity1Id
        externalId: entity1Id
        attributeAlias: entity1-id
relationships:
  rel1:
    name: test_relationship
    fromAttribute: test-id
    toAttribute: entity1-id
`
	yamlPath := filepath.Join(tempDir, "test.yaml")
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test YAML file: %v", err)
	}

	// Create output directory
	outputPath := filepath.Join(tempDir, "output")
	err = os.MkdirAll(outputPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Create a sample CSV file with valid data
	entity1CSV := "id,name\n1,Entity One\n2,Entity Two\n"
	entity1Path := filepath.Join(outputPath, "Entity.csv")
	err = os.WriteFile(entity1Path, []byte(entity1CSV), 0644)
	if err != nil {
		t.Fatalf("Failed to write Entity.csv file: %v", err)
	}

	// Create a related CSV file with an invalid reference
	relatedCSV := "id,entity1Id\na,999\nb,888\n" // References to entities that don't exist
	relatedPath := filepath.Join(outputPath, "Related.csv")
	err = os.WriteFile(relatedPath, []byte(relatedCSV), 0644)
	if err != nil {
		t.Fatalf("Failed to write Related.csv file: %v", err)
	}

	// Disable color for testing
	oldNoColor := color.NoColor
	defer func() { color.NoColor = oldNoColor }()
	color.NoColor = true

	// Save and restore validateOnly flag
	oldValidateOnly := validateOnly
	defer func() { validateOnly = oldValidateOnly }()
	validateOnly = true

	// Make sure validation is enabled
	oldValidateRelationships := validateRelationships
	defer func() { validateRelationships = oldValidateRelationships }()
	validateRelationships = true

	// Run the application in validation-only mode
	err = run(yamlPath, outputPath, 10, false)

	// Should not return an error as validation issues are reported but don't fail the program
	if err != nil {
		t.Errorf("run() with invalid CSV relationships shouldn't have returned an error: %v", err)
	}

	// CSV files should remain unchanged
	modifiedEntityCSV, err := os.ReadFile(entity1Path)
	if err != nil {
		t.Fatalf("Failed to read Entity.csv after validation: %v", err)
	}
	if string(modifiedEntityCSV) != entity1CSV {
		t.Error("Entity.csv was modified in validation-only mode")
	}

	modifiedRelatedCSV, err := os.ReadFile(relatedPath)
	if err != nil {
		t.Fatalf("Failed to read Related.csv after validation: %v", err)
	}
	if string(modifiedRelatedCSV) != relatedCSV {
		t.Error("Related.csv was modified in validation-only mode")
	}
}

func TestCommandLineFlagsWithValidateOnly(t *testing.T) {
	// Test the validateOnly flag
	withFlagValues(t, map[string]string{
		"-f":              "test.yaml",
		"--validate-only": "",
	}, func() {
		if inputFile != "test.yaml" {
			t.Errorf("Expected inputFile to be 'test.yaml', got '%s'", inputFile)
		}

		if !validateOnly {
			t.Error("Expected validateOnly to be true, got false")
		}
	})

	// Test combination with other flags
	withFlagValues(t, map[string]string{
		"-f":              "test.yaml",
		"--validate-only": "",
		"-d=false":        "", // Disable diagram - boolean flags need the '=false' syntax
	}, func() {
		if inputFile != "test.yaml" {
			t.Errorf("Expected inputFile to be 'test.yaml', got '%s'", inputFile)
		}

		if !validateOnly {
			t.Error("Expected validateOnly to be true, got false")
		}

		if generateDiagram {
			t.Error("Expected generateDiagram to be false, got true")
		}
	})
}
