package generators

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/stretchr/testify/assert"
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
		assert.Empty(t, results, "Expected no errors for entity with no unique attributes")
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
		assert.NotEmpty(t, results, "Expected errors for missing unique attribute in headers")
		assert.NotEmpty(t, results[0].Messages, "Expected error messages for missing unique attribute")
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
		assert.NotEmpty(t, results, "Expected errors for empty values in unique attribute")
		assert.NotEmpty(t, results[0].Messages, "Expected error messages for empty values")
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
		assert.NotEmpty(t, results, "Expected errors for duplicate values in unique attribute")
		assert.NotEmpty(t, results[0].Messages, "Expected error messages for duplicate values")
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
		assert.NotEmpty(t, results, "Expected errors for multiple issues")
		assert.GreaterOrEqual(t, len(results[0].Messages), 2, "Expected multiple error messages")
	})

	// Test case 6: Test GetEntityFileName function with nil input
	t.Run("GetEntityFileNameWithNilInput", func(t *testing.T) {
		// Call the function with nil input
		result := GetEntityFileName(nil)

		// Should return "unknown"
		assert.Equal(t, "unknown", result, "Expected 'unknown' for nil input to GetEntityFileName")
	})

	// Test case 7: Test GetEntityFileName with and without namespace prefix
	t.Run("GetEntityFileNameWithAndWithoutNamespace", func(t *testing.T) {
		// Test with namespace prefix
		withPrefix := &models.CSVData{
			ExternalId: "Namespace/Entity",
		}
		resultWithPrefix := GetEntityFileName(withPrefix)
		assert.Equal(t, "Entity.csv", resultWithPrefix, "Expected 'Entity.csv' for 'Namespace/Entity'")

		// Test without namespace prefix
		withoutPrefix := &models.CSVData{
			ExternalId: "Entity",
		}
		resultWithoutPrefix := GetEntityFileName(withoutPrefix)
		assert.Equal(t, "Entity.csv", resultWithoutPrefix, "Expected 'Entity.csv' for 'Entity'")
	})
}
