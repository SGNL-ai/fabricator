package generators

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestIDMapConsistency ensures that ID fields are properly handled
func TestIDMapConsistency(t *testing.T) {
	// Create a generator
	generator := NewCSVGenerator("test_output", 3, false)
	generator.generateCommonValues() // Initialize common values

	// Set up entity data
	generator.EntityData = map[string]*models.CSVData{
		"entity": {
			EntityName: "TestEntity",
			ExternalId: "Test/TestEntity",
			Headers:    []string{"id", "name"},
		},
	}
	
	// Mark the id field as a unique attribute
	generator.uniqueIdAttributes = map[string][]string{
		"entity": {"id"},
	}

	// Set up idMap
	expectedID := "test-id-1234"
	generator.idMap = map[string]map[string]string{
		"entity": {"0": expectedID},
	}

	// Generate row
	row := generator.generateRowForEntity("entity", 0)

	// Check ID field
	assert.Equal(t, expectedID, row[0], "Generated row should have expected ID value")

	// Test case where idMap is missing for this index
	generator.idMap["entity"] = map[string]string{} // empty map

	// Create a new UUID to compare
	newID := uuid.New().String()
	t.Logf("Using new UUID %s for comparison", newID)

	// Note: The test is now simply verifying that an ID is generated when one doesn't exist
	// instead of trying to predict exactly what it will be
	row = generator.generateRowForEntity("entity", 0)

	// Check that *some* ID was generated
	assert.NotEmpty(t, row[0], "Expected non-empty ID when missing from idMap")

	// Verify we have a UUID-like value (at least 36 characters long)
	assert.GreaterOrEqual(t, len(row[0]), 36, 
		"Expected ID to be a UUID-like value with at least 36 characters, got '%s'", row[0])
}