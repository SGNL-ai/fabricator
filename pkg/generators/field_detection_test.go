package generators

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/google/uuid"
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

	// Set up idMap
	expectedID := "test-id-1234"
	generator.idMap = map[string]map[string]string{
		"entity": {"0": expectedID},
	}

	// Generate row
	row := generator.generateRowForEntity("entity", 0)

	// Check ID field
	if row[0] != expectedID {
		t.Errorf("Expected ID to be '%s', got '%s'", expectedID, row[0])
	}

	// Test case where idMap is missing for this index
	generator.idMap["entity"] = map[string]string{} // empty map

	// Create a new UUID to compare
	newID := uuid.New().String()
	t.Logf("Using new UUID %s for comparison", newID)

	// Note: The test is now simply verifying that an ID is generated when one doesn't exist
	// instead of trying to predict exactly what it will be
	row = generator.generateRowForEntity("entity", 0)

	// Check that *some* ID was generated
	if row[0] == "" {
		t.Errorf("Expected non-empty ID when missing from idMap")
	}

	if len(row[0]) < 10 {
		t.Errorf("Expected ID to be a UUID-like value, got '%s'", row[0])
	}
}
