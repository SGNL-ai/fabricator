package generators

import (
	"os"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
)

func TestValidationFunctions(t *testing.T) {
	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "validation-functions-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Basic setup for validation testing
	t.Run("Basic validation functions", func(t *testing.T) {
		// Create a test generator
		generator := NewCSVGenerator(tempDir, 5, true)

		// Set up some test entities and relationships
		generator.EntityData = map[string]*models.CSVData{
			"user": {
				EntityName: "User",
				Headers:    []string{"id", "name"},
				Rows: [][]string{
					{"user1", "User 1"},
					{"user2", "User 2"},
				},
			},
			"order": {
				EntityName: "Order",
				Headers:    []string{"id", "userId", "amount"},
				Rows: [][]string{
					{"order1", "user1", "100"},
					{"order2", "user2", "200"},
					{"order3", "user1", "300"},
					{"order4", "invalid", "400"}, // Invalid reference
				},
			},
		}

		// Set up a relationship
		relationshipLink := models.RelationshipLink{
			FromEntityID:      "order",
			ToEntityID:        "user",
			FromAttribute:     "userId",
			ToAttribute:       "id",
			IsFromAttributeID: false,
			IsToAttributeID:   true,
		}

		generator.relationshipMap = map[string][]models.RelationshipLink{
			"order": {relationshipLink},
		}

		// Set up tracking for unique attributes
		generator.uniqueIdAttributes = map[string][]string{
			"user":  {"id"},
			"order": {"id"},
		}

		// Test ValidateRelationships
		results := generator.ValidateRelationships()
		if len(results) == 0 {
			t.Errorf("Expected validation issues but got none")
		} else {
			// Should find 1 issue with 1 invalid row
			found := false
			for _, result := range results {
				if result.FromEntity == "order" && result.ToEntity == "user" && result.InvalidRows == 1 {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected to find 1 invalid row in order->user relationship")
			}
		}

		// Test ValidateUniqueValues with no duplicates
		errors := generator.ValidateUniqueValues()
		if len(errors) > 0 {
			t.Errorf("Expected no uniqueness errors, but got %d", len(errors))
		}

		// Add a duplicate ID to test uniqueness validation
		generator.EntityData["user"].Rows = append(
			generator.EntityData["user"].Rows,
			[]string{"user1", "Duplicate User"}, // Duplicate ID
		)

		// Now validate uniqueness again
		errors = generator.ValidateUniqueValues()
		if len(errors) == 0 {
			t.Errorf("Expected uniqueness errors after adding duplicate, but got none")
		}
	})
}
