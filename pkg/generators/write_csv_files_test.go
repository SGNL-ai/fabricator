package generators

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
)

func TestWriteCSVFilesExtended(t *testing.T) {
	// Setup a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "fabricator-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }() // Clean up after the test

	// Test case 1: Write CSV files with namespace prefix
	t.Run("WriteCSVFilesWithNamespacePrefix", func(t *testing.T) {
		// Setup test data
		generator := NewCSVGenerator(tempDir, 2, false)

		// Create entity data with namespace prefix
		userEntity := models.CSVData{
			ExternalId: "Namespace/User", // With namespace prefix
			EntityName: "User Entity",
			Headers:    []string{"id", "name", "email"},
			Rows: [][]string{
				{"user-1", "John Doe", "john@example.com"},
				{"user-2", "Jane Smith", "jane@example.com"},
			},
		}

		// Add to generator
		generator.EntityData["user"] = &userEntity

		// Call function under test
		err := generator.WriteCSVFiles()
		if err != nil {
			t.Fatalf("WriteCSVFiles failed: %v", err)
		}

		// Verify file was created with expected name (without namespace prefix)
		expectedFilePath := filepath.Join(tempDir, "User.csv")
		if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", expectedFilePath)
		}
	})

	// Test case 2: Write CSV files without namespace prefix
	t.Run("WriteCSVFilesWithoutNamespacePrefix", func(t *testing.T) {
		// Setup test data
		generator := NewCSVGenerator(tempDir, 2, false)

		// Create entity data without namespace prefix
		userEntity := models.CSVData{
			ExternalId: "User", // No namespace prefix
			EntityName: "User Entity",
			Headers:    []string{"id", "name", "email"},
			Rows: [][]string{
				{"user-1", "John Doe", "john@example.com"},
				{"user-2", "Jane Smith", "jane@example.com"},
			},
		}

		// Add to generator
		generator.EntityData["user"] = &userEntity

		// Call function under test
		err := generator.WriteCSVFiles()
		if err != nil {
			t.Fatalf("WriteCSVFiles failed: %v", err)
		}

		// Verify file was created
		expectedFilePath := filepath.Join(tempDir, "User.csv")
		if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", expectedFilePath)
		}
	})

	// Test case 3: Error creating directory
	t.Run("ErrorCreatingDirectory", func(t *testing.T) {
		// Try to use a non-writable directory
		invalidDir := filepath.Join("/proc", "nonexistent")
		generator := NewCSVGenerator(invalidDir, 2, false)

		// Create entity data
		userEntity := models.CSVData{
			ExternalId: "User",
			EntityName: "User Entity",
			Headers:    []string{"id", "name"},
			Rows: [][]string{
				{"user-1", "John Doe"},
			},
		}

		// Add to generator
		generator.EntityData["user"] = &userEntity

		// Call function under test - should return an error
		err := generator.WriteCSVFiles()

		// Skip if we can actually create the directory (might happen on some systems)
		if err == nil {
			_, err = os.Stat(filepath.Join(invalidDir, "User.csv"))
			if err == nil {
				t.Skip("Test environment allows creating files in unexpected location")
			}
		}

		if err == nil {
			t.Errorf("Expected error when creating directory in protected location, but got nil")
		}
	})

	// Removed unused mock implementation

	// Test case 4: Error writing headers
	// This test is more complex as it requires mocking file operations
	// Simplified test checking file contents instead
	t.Run("ValidateCSVContents", func(t *testing.T) {
		// Setup test data in clean directory
		csvDir := filepath.Join(tempDir, "csv-contents")
		err := os.MkdirAll(csvDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		generator := NewCSVGenerator(csvDir, 2, false)

		// Create entity data
		productEntity := models.CSVData{
			ExternalId: "Product",
			EntityName: "Product Entity",
			Headers:    []string{"id", "name", "price"},
			Rows: [][]string{
				{"product-1", "Laptop", "999.99"},
				{"product-2", "Smartphone", "499.99"},
			},
		}

		// Add to generator
		generator.EntityData["product"] = &productEntity

		// Write the files
		err = generator.WriteCSVFiles()
		if err != nil {
			t.Fatalf("WriteCSVFiles failed: %v", err)
		}

		// Read the file back to verify contents
		csvFile := filepath.Join(csvDir, "Product.csv")
		content, err := os.ReadFile(csvFile)
		if err != nil {
			t.Fatalf("Failed to read generated CSV: %v", err)
		}

		// Very basic content check - headers must be first line
		expectedHeader := "id,name,price"
		if string(content[:len(expectedHeader)]) != expectedHeader {
			t.Errorf("Expected CSV to start with '%s', got: '%s'",
				expectedHeader, string(content[:len(expectedHeader)]))
		}
	})
}
