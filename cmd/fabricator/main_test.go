package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
)

// Helper function to save and restore command-line flags for testing
func withFlagValues(t *testing.T, flags map[string]string, fn func()) {
	// Save current flag values
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Create new args
	newArgs := []string{"fabricator"}
	for name, value := range flags {
		if value == "" {
			// For boolean flags (which don't need values)
			// Handle boolean flags with format "-flag=false"
			if strings.Contains(name, "=") {
				newArgs = append(newArgs, name)
			} else {
				newArgs = append(newArgs, name)
			}
		} else {
			// For flags with values
			newArgs = append(newArgs, name, value)
		}
	}
	os.Args = newArgs

	// Reset flags for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Re-initialize global variables
	showVersion = false
	inputFile = ""
	outputDir = "output"
	dataVolume = 100
	autoCardinality = false
	validateOnly = false
	generateDiagram = true

	// Re-register flags
	flag.BoolVar(&showVersion, "v", false, "Display version information")
	flag.BoolVar(&showVersion, "version", false, "Display version information")

	flag.StringVar(&inputFile, "f", "", "Path to the YAML definition file (required)")
	flag.StringVar(&inputFile, "file", "", "Path to the YAML definition file (required)")

	flag.StringVar(&outputDir, "o", "output", "Directory to store generated CSV files")
	flag.StringVar(&outputDir, "output", "output", "Directory to store generated CSV files")

	flag.IntVar(&dataVolume, "n", 100, "Number of rows to generate for each entity")
	flag.IntVar(&dataVolume, "num-rows", 100, "Number of rows to generate for each entity")

	flag.BoolVar(&autoCardinality, "a", false, "Enable automatic cardinality detection for relationships")
	flag.BoolVar(&autoCardinality, "auto-cardinality", false, "Enable automatic cardinality detection for relationships")

	// Add validateOnly flag
	flag.BoolVar(&validateOnly, "validate-only", false, "Validate existing CSV files without generating new data")

	// Add diagram generation flag
	flag.BoolVar(&generateDiagram, "d", true, "Generate Entity-Relationship diagram")
	flag.BoolVar(&generateDiagram, "diagram", true, "Generate Entity-Relationship diagram")

	// Parse flags
	flag.Parse()

	// Run the test function
	fn()
}

func TestCommandLineFlagParsing(t *testing.T) {
	// Test short form flags
	withFlagValues(t, map[string]string{
		"-f": "test.yaml",
		"-o": "testdir",
		"-n": "42",
	}, func() {
		if inputFile != "test.yaml" {
			t.Errorf("Expected inputFile to be 'test.yaml', got '%s'", inputFile)
		}

		if outputDir != "testdir" {
			t.Errorf("Expected outputDir to be 'testdir', got '%s'", outputDir)
		}

		if dataVolume != 42 {
			t.Errorf("Expected dataVolume to be 42, got %d", dataVolume)
		}
	})

	// Test long form flags
	withFlagValues(t, map[string]string{
		"--file":     "other.yaml",
		"--output":   "otherdir",
		"--num-rows": "123",
	}, func() {
		if inputFile != "other.yaml" {
			t.Errorf("Expected inputFile to be 'other.yaml', got '%s'", inputFile)
		}

		if outputDir != "otherdir" {
			t.Errorf("Expected outputDir to be 'otherdir', got '%s'", outputDir)
		}

		if dataVolume != 123 {
			t.Errorf("Expected dataVolume to be 123, got %d", dataVolume)
		}
	})

	// Test version flag (short form)
	withFlagValues(t, map[string]string{
		"-v": "",
	}, func() {
		if !showVersion {
			t.Error("Expected showVersion to be true, got false")
		}
	})

	// Test version flag (long form)
	withFlagValues(t, map[string]string{
		"--version": "",
	}, func() {
		if !showVersion {
			t.Error("Expected showVersion to be true, got false")
		}
	})
}

// Test validation-only mode CLI behavior
func TestValidationOnlyMode(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "fabricator-validation-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a minimal YAML file
	yamlContent := `displayName: Test SOR
description: Test system of record
entities:
  entity1:
    displayName: TestEntity
    externalId: Test/Entity
    description: A test entity
    attributes:
      - name: id
        externalId: id
        type: String
        uniqueId: true
      - name: name
        externalId: name
        type: String`

	yamlPath := filepath.Join(tempDir, "test.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML file: %v", err)
	}

	// Create output directory and CSV file
	outputPath := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	csvContent := `id,name
entity-1,Test Name
entity-2,Another Name`
	csvPath := filepath.Join(outputPath, "Entity.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to write CSV file: %v", err)
	}

	// Save current flag values
	oldValidateOnly := validateOnly
	defer func() { validateOnly = oldValidateOnly }()
	validateOnly = true

	// Run the application in validation-only mode
	err = run(yamlPath, outputPath, 10, false)

	// Should succeed - new validation-only mode doesn't fail fatally
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
	tempDir, err := os.MkdirTemp("", "fabricator-missing-csv-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	yamlContent := `displayName: Test SOR
description: Test system of record
entities:
  entity1:
    displayName: TestEntity
    externalId: Test/Entity
    description: A test entity
    attributes:
      - name: id
        externalId: id
        type: String
        uniqueId: true`

	yamlPath := filepath.Join(tempDir, "test.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML file: %v", err)
	}

	// Create empty output directory (no CSV files)
	outputPath := filepath.Join(tempDir, "empty_output")
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	oldValidateOnly := validateOnly
	defer func() { validateOnly = oldValidateOnly }()
	validateOnly = true

	// Run the application in validation-only mode
	err = run(yamlPath, outputPath, 10, false)

	// New behavior: should succeed but report validation issues
	// (collect errors instead of failing fatally)
	if err != nil {
		t.Errorf("run() should not fail fatally for missing CSV files, got: %v", err)
	}
	// The missing files should be reported as validation issues in the output
}

// TestPrintHeaderFunctionExists just checks that the function exists and runs
func TestPrintHeaderFunctionExists(t *testing.T) {
	// Just call the function to make sure it doesn't panic
	printHeader()
}

// TestPrintHeader tests that the function exists and can be called without errors
func TestPrintHeader(t *testing.T) {
	// Since the actual function uses the color package to print directly to terminal,
	// it's difficult to capture and verify that output in a test.
	// Instead, we just verify that the function exists and runs without panicking,
	// which is enough for a unit test. The visual output can be manually verified.

	// Temporarily disable color for testing
	oldNoColor := color.NoColor
	defer func() { color.NoColor = oldNoColor }()
	color.NoColor = true

	// Run the function to make sure it doesn't panic
	printHeader()

	// Test passes if we reach this point without panicking
	// This is a trivial test but ensures the function exists and is callable
}

// TestPrintUsage tests the printUsage function
func TestPrintUsage(t *testing.T) {
	// Use a more robust approach that will work with direct console output
	// Rather than trying to capture output, we just check that the function
	// doesn't panic, which is sufficient for test coverage

	// Run the function - if it panics, the test will fail
	printUsage()

	// Test passes if we reach this point (function ran without panic)
	// For code coverage purposes, this is considered a successful test
}

func TestRunWithInvalidYAML(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "fabricator-test-invalid-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create an invalid YAML file
	yamlContent := `
displayName: Invalid YAML
description: This YAML is invalid
entities:
  - This is not valid YAML for our parser
`
	yamlPath := filepath.Join(tempDir, "invalid.yaml")
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test YAML file: %v", err)
	}

	// Create output directory
	outputPath := filepath.Join(tempDir, "output")

	// Run the application with the invalid YAML
	err = run(yamlPath, outputPath, 2, false)

	// This should return an error
	if err == nil {
		t.Error("run() with invalid YAML should have returned an error")
	}
}

func TestRunWithNonexistentYAML(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "fabricator-test-nonexistent-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Use a non-existent YAML file
	yamlPath := filepath.Join(tempDir, "nonexistent.yaml")

	// Create output directory
	outputPath := filepath.Join(tempDir, "output")

	// Run the application with the non-existent YAML
	err = run(yamlPath, outputPath, 2, false)

	// This should return an error
	if err == nil {
		t.Error("run() with non-existent YAML should have returned an error")
	}
}

func TestBasicFunctionalityWithMinimalYAML(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "fabricator-test-*")
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
        type: String
        uniqueId: true
        attributeAlias: test-id
      - name: name
        externalId: name
        type: String
        attributeAlias: test-name
`
	yamlPath := filepath.Join(tempDir, "test.yaml")
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test YAML file: %v", err)
	}

	// Create an output directory
	outputPath := filepath.Join(tempDir, "output")

	// Run the application with the test YAML
	// We can't actually run main() because it may call os.Exit(),
	// but we can call run() directly
	err = run(yamlPath, outputPath, 2, false)

	// If there was an error, fail the test
	if err != nil {
		t.Fatalf("run() failed: %v", err)
	}

	// Check that the output directory was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output directory was not created")
	}

	// Check that the CSV file was created
	csvPath := filepath.Join(outputPath, "Entity.csv")
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Error("CSV file was not created")
	}

	// Read the CSV file
	csvContent, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	// Check that the CSV file has the expected format
	lines := strings.Split(string(csvContent), "\n")

	// Header line
	if !strings.HasPrefix(lines[0], "id,name") {
		t.Errorf("Expected header line to begin with 'id,name', got %s", lines[0])
	}

	// Should have 3 lines (header + 2 data rows + empty line at end)
	if len(lines) != 4 {
		t.Errorf("Expected 4 lines in CSV (header + 2 data rows + empty), got %d", len(lines))
	}

	// Each data row should have 2 values (id and name)
	for i := 1; i <= 2; i++ {
		if i >= len(lines) {
			t.Fatalf("Not enough lines in CSV")
		}

		values := strings.Split(lines[i], ",")
		if len(values) != 2 {
			t.Errorf("Expected 2 values in data row %d, got %d", i, len(values))
		}

		// ID should not be empty
		if values[0] == "" {
			t.Errorf("Expected non-empty ID in row %d", i)
		}

		// Name should not be empty
		if values[1] == "" {
			t.Errorf("Expected non-empty name in row %d", i)
		}
	}
}
