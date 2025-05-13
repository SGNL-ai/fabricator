package generators

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
)

func TestValidateUniqueValuesExpanded(t *testing.T) {
	// Test case 1: Entity with no unique attributes
	t.Run("NoUniqueAttributes", func(t *testing.T) {
		// Setup test data
		generator := NewCSVGenerator("test-output", 10, false)

		// Create entity with no unique attributes
		entity := models.CSVData{
			ExternalId: "entity",
			EntityName: "Entity",
			Headers:    []string{"id", "name"},
			Rows: [][]string{
				{"entity-1", "Entity 1"},
				{"entity-2", "Entity 2"},
			},
		}

		generator.EntityData["entity"] = &entity

		// No unique attributes set in the map
		generator.uniqueIdAttributes["entity"] = []string{}

		// Call function under test
		results := generator.ValidateUniqueValues()

		// Should not report any errors
		if len(results) > 0 {
			t.Errorf("Expected no errors for entity with no unique attributes, but found %d errors", len(results))
		}
	})

	// Test case 2: Missing unique attribute in headers
	t.Run("MissingUniqueAttributeInHeaders", func(t *testing.T) {
		// Setup test data
		generator := NewCSVGenerator("test-output", 10, false)

		// Create entity
		entity := models.CSVData{
			ExternalId: "entity",
			EntityName: "Entity",
			Headers:    []string{"id", "name"}, // Missing "email" from headers
			Rows: [][]string{
				{"entity-1", "Entity 1"},
				{"entity-2", "Entity 2"},
			},
		}

		generator.EntityData["entity"] = &entity

		// Set up a unique attribute that doesn't exist in headers
		generator.uniqueIdAttributes["entity"] = []string{"email"}

		// Call function under test
		results := generator.ValidateUniqueValues()

		// Should report errors about missing attribute
		if len(results) == 0 {
			t.Errorf("Expected errors for missing unique attribute in headers, but found none")
		} else if len(results[0].Messages) == 0 {
			t.Errorf("Expected error messages for missing unique attribute, but found none")
		}
	})

	// Test case 3: Empty values in unique attribute
	t.Run("EmptyValuesInUniqueAttribute", func(t *testing.T) {
		// Setup test data
		generator := NewCSVGenerator("test-output", 10, false)

		// Create entity with empty values in a unique attribute
		entity := models.CSVData{
			ExternalId: "entity",
			EntityName: "Entity",
			Headers:    []string{"id", "email"},
			Rows: [][]string{
				{"entity-1", "user1@example.com"},
				{"entity-2", ""}, // Empty email
			},
		}

		generator.EntityData["entity"] = &entity

		// Set up email as a unique attribute
		generator.uniqueIdAttributes["entity"] = []string{"email"}

		// Call function under test
		results := generator.ValidateUniqueValues()

		// Should report errors about empty value
		if len(results) == 0 {
			t.Errorf("Expected errors for empty values in unique attribute, but found none")
		} else if len(results[0].Messages) == 0 {
			t.Errorf("Expected error messages for empty values, but found none")
		}
	})

	// Test case 4: Duplicate values in unique attribute
	t.Run("DuplicateValuesInUniqueAttribute", func(t *testing.T) {
		// Setup test data
		generator := NewCSVGenerator("test-output", 10, false)

		// Create entity with duplicate values in a unique attribute
		entity := models.CSVData{
			ExternalId: "entity",
			EntityName: "Entity",
			Headers:    []string{"id", "email"},
			Rows: [][]string{
				{"entity-1", "same@example.com"},
				{"entity-2", "same@example.com"}, // Duplicate email
				{"entity-3", "different@example.com"},
			},
		}

		generator.EntityData["entity"] = &entity

		// Set up email as a unique attribute
		generator.uniqueIdAttributes["entity"] = []string{"email"}

		// Call function under test
		results := generator.ValidateUniqueValues()

		// Should report errors about duplicate values
		if len(results) == 0 {
			t.Errorf("Expected errors for duplicate values in unique attribute, but found none")
		} else if len(results[0].Messages) == 0 {
			t.Errorf("Expected error messages for duplicate values, but found none")
		}
	})

	// Test case 5: Multiple unique attributes with issues
	t.Run("MultipleUniqueAttributesWithIssues", func(t *testing.T) {
		// Setup test data
		generator := NewCSVGenerator("test-output", 10, false)

		// Create entity with multiple unique attributes and various issues
		entity := models.CSVData{
			ExternalId: "entity",
			EntityName: "Entity",
			Headers:    []string{"id", "email", "username"},
			Rows: [][]string{
				{"entity-1", "user1@example.com", "user1"},
				{"entity-2", "user2@example.com", "user2"},
				{"entity-3", "user1@example.com", ""},      // Duplicate email, empty username
				{"entity-4", "user4@example.com", "user2"}, // Duplicate username
			},
		}

		generator.EntityData["entity"] = &entity

		// Set up multiple unique attributes
		generator.uniqueIdAttributes["entity"] = []string{"email", "username"}

		// Call function under test
		results := generator.ValidateUniqueValues()

		// Should report multiple issues
		if len(results) == 0 {
			t.Errorf("Expected errors for multiple issues, but found none")
		} else if len(results[0].Messages) < 2 {
			t.Errorf("Expected multiple error messages, but found %d", len(results[0].Messages))
		}
	})

	// Test case 6: Test getEntityFileName function with nil input
	t.Run("GetEntityFileNameWithNilInput", func(t *testing.T) {
		// Call the function with nil input
		result := getEntityFileName(nil)

		// Should return "unknown"
		if result != "unknown" {
			t.Errorf("Expected 'unknown' for nil input to getEntityFileName, got '%s'", result)
		}
	})

	// Test case 7: Test getEntityFileName with and without namespace prefix
	t.Run("GetEntityFileNameWithAndWithoutNamespace", func(t *testing.T) {
		// Test with namespace prefix
		withPrefix := &models.CSVData{
			ExternalId: "Namespace/Entity",
		}
		resultWithPrefix := getEntityFileName(withPrefix)
		if resultWithPrefix != "Entity.csv" {
			t.Errorf("Expected 'Entity.csv' for 'Namespace/Entity', got '%s'", resultWithPrefix)
		}

		// Test without namespace prefix
		withoutPrefix := &models.CSVData{
			ExternalId: "Entity",
		}
		resultWithoutPrefix := getEntityFileName(withoutPrefix)
		if resultWithoutPrefix != "Entity.csv" {
			t.Errorf("Expected 'Entity.csv' for 'Entity', got '%s'", resultWithoutPrefix)
		}
	})
}
