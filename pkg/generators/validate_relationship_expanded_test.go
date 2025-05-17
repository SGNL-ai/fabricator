package generators

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestValidateRelationshipExpanded(t *testing.T) {
	// Test case 1: Missing entity data
	t.Run("MissingEntityData", func(t *testing.T) {
		// Setup test data
		generator := NewCSVGenerator("test-output", 10, false)

		// Only add data for the from entity
		fromEntity := models.CSVData{
			ExternalId: "from",
			EntityName: "From Entity",
			Headers:    []string{"id", "name"},
			Rows: [][]string{
				{"from-1", "From 1"},
			},
		}
		generator.EntityData["from"] = &fromEntity

		// Create relationship link with a to-entity that doesn't exist
		link := models.RelationshipLink{
			FromEntityID:  "from",
			ToEntityID:    "missing", // This entity doesn't exist
			FromAttribute: "id",
			ToAttribute:   "fromId",
		}

		// Call function under test
		result := generator.validateRelationship("from", link)

		// Check that it identified the missing entity
		assert.NotEmpty(t, result.Errors, "Expected errors for missing entity data")
	})

	// Test case 2: Missing attribute columns
	t.Run("MissingAttributeColumns", func(t *testing.T) {
		// Setup test data
		generator := NewCSVGenerator("test-output", 10, false)

		// Create both entities
		fromEntity := models.CSVData{
			ExternalId: "from",
			EntityName: "From Entity",
			Headers:    []string{"id", "name"}, // Missing the attribute we'll reference
			Rows: [][]string{
				{"from-1", "From 1"},
			},
		}

		toEntity := models.CSVData{
			ExternalId: "to",
			EntityName: "To Entity",
			Headers:    []string{"id", "name"}, // Missing the attribute we'll reference
			Rows: [][]string{
				{"to-1", "To 1"},
			},
		}

		generator.EntityData["from"] = &fromEntity
		generator.EntityData["to"] = &toEntity

		// Create relationship link with attributes that don't exist
		link := models.RelationshipLink{
			FromEntityID:  "from",
			ToEntityID:    "to",
			FromAttribute: "missingAttr",        // Doesn't exist in headers
			ToAttribute:   "anotherMissingAttr", // Doesn't exist in headers
		}

		// Call function under test
		result := generator.validateRelationship("from", link)

		// Check that it identified the missing attributes
		assert.NotEmpty(t, result.Errors, "Expected errors for missing attribute columns")
	})

	// Test case 3: Empty foreign key
	t.Run("EmptyForeignKey", func(t *testing.T) {
		// Setup test data
		generator := NewCSVGenerator("test-output", 10, false)

		// Create primary key entity
		primaryEntity := models.CSVData{
			ExternalId: "primary",
			EntityName: "Primary Entity",
			Headers:    []string{"id", "name"},
			Rows: [][]string{
				{"primary-1", "Primary 1"},
			},
		}

		// Create foreign key entity with empty reference
		foreignEntity := models.CSVData{
			ExternalId: "foreign",
			EntityName: "Foreign Entity",
			Headers:    []string{"id", "primaryId", "name"},
			Rows: [][]string{
				{"foreign-1", "", "Foreign 1"}, // Empty foreign key
			},
		}

		generator.EntityData["primary"] = &primaryEntity
		generator.EntityData["foreign"] = &foreignEntity

		// Create relationship link from primary to foreign
		link := models.RelationshipLink{
			FromEntityID:  "primary",
			ToEntityID:    "foreign",
			FromAttribute: "id",        // Primary key
			ToAttribute:   "primaryId", // Foreign key
		}

		// Call function under test - primary key to foreign key
		result := generator.validateRelationship("primary", link)

		// Should flag the empty foreign key
		assert.NotEmpty(t, result.Errors, "Expected errors for empty foreign key")
		assert.Greater(t, result.InvalidRows, 0, "Expected invalid rows for empty foreign key")
	})

	// Test case 4: Invalid primary key reference
	t.Run("InvalidPrimaryKeyReference", func(t *testing.T) {
		// Setup test data
		generator := NewCSVGenerator("test-output", 10, false)

		// Create primary key entity
		primaryEntity := models.CSVData{
			ExternalId: "primary",
			EntityName: "Primary Entity",
			Headers:    []string{"id", "name"},
			Rows: [][]string{
				{"primary-1", "Primary 1"},
			},
		}

		// Create foreign key entity with invalid reference
		foreignEntity := models.CSVData{
			ExternalId: "foreign",
			EntityName: "Foreign Entity",
			Headers:    []string{"id", "primaryId", "name"},
			Rows: [][]string{
				{"foreign-1", "invalid-ref", "Foreign 1"}, // Invalid reference
			},
		}

		generator.EntityData["primary"] = &primaryEntity
		generator.EntityData["foreign"] = &foreignEntity

		// Create relationship link from primary to foreign
		link := models.RelationshipLink{
			FromEntityID:  "primary",
			ToEntityID:    "foreign",
			FromAttribute: "id",        // Primary key
			ToAttribute:   "primaryId", // Foreign key
		}

		// Call function under test - primary key to foreign key
		result := generator.validateRelationship("primary", link)

		// Should flag the invalid reference
		assert.NotEmpty(t, result.Errors, "Expected errors for invalid primary key reference")
		assert.Greater(t, result.InvalidRows, 0, "Expected invalid rows for invalid primary key reference")
	})

	// Test case 5: Empty source value
	t.Run("EmptySourceValue", func(t *testing.T) {
		// Setup test data
		generator := NewCSVGenerator("test-output", 10, false)

		// Create from entity with empty value
		fromEntity := models.CSVData{
			ExternalId: "from",
			EntityName: "From Entity",
			Headers:    []string{"id", "toRef", "name"},
			Rows: [][]string{
				{"from-1", "", "From 1"}, // Empty reference
			},
		}

		// Create to entity
		toEntity := models.CSVData{
			ExternalId: "to",
			EntityName: "To Entity",
			Headers:    []string{"ref", "name"},
			Rows: [][]string{
				{"to-ref", "To 1"},
			},
		}

		generator.EntityData["from"] = &fromEntity
		generator.EntityData["to"] = &toEntity

		// Create relationship link - regular case, not primary/foreign keys
		link := models.RelationshipLink{
			FromEntityID:  "from",
			ToEntityID:    "to",
			FromAttribute: "toRef", // Regular attribute
			ToAttribute:   "ref",   // Regular attribute
		}

		// Call function under test
		result := generator.validateRelationship("from", link)

		// Should flag the empty source value
		assert.NotEmpty(t, result.Errors, "Expected errors for empty source value")
		assert.Greater(t, result.InvalidRows, 0, "Expected invalid rows for empty source value")
	})

	// Test case 6: Invalid target reference
	t.Run("InvalidTargetReference", func(t *testing.T) {
		// Setup test data
		generator := NewCSVGenerator("test-output", 10, false)

		// Create from entity with invalid reference
		fromEntity := models.CSVData{
			ExternalId: "from",
			EntityName: "From Entity",
			Headers:    []string{"id", "toRef", "name"},
			Rows: [][]string{
				{"from-1", "invalid-ref", "From 1"}, // Invalid reference
			},
		}

		// Create to entity
		toEntity := models.CSVData{
			ExternalId: "to",
			EntityName: "To Entity",
			Headers:    []string{"ref", "name"},
			Rows: [][]string{
				{"to-ref", "To 1"}, // Doesn't match the "invalid-ref" value
			},
		}

		generator.EntityData["from"] = &fromEntity
		generator.EntityData["to"] = &toEntity

		// Create relationship link - regular case, not primary/foreign keys
		link := models.RelationshipLink{
			FromEntityID:  "from",
			ToEntityID:    "to",
			FromAttribute: "toRef", // Regular attribute
			ToAttribute:   "ref",   // Regular attribute
		}

		// Call function under test
		result := generator.validateRelationship("from", link)

		// Should flag the invalid reference
		assert.NotEmpty(t, result.Errors, "Expected errors for invalid target reference")
		assert.Greater(t, result.InvalidRows, 0, "Expected invalid rows for invalid target reference")
	})
}