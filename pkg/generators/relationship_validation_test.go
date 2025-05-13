package generators

import (
	"os"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
)

func TestRelationshipValidation(t *testing.T) {
	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "relationship-validation-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Test 1: Validate a valid relationship
	t.Run("Valid relationship", func(t *testing.T) {
		// Create a test generator with auto-cardinality enabled
		generator := NewCSVGenerator(tempDir, 10, true)

		// Create test entities with a valid relationship
		generator.EntityData = map[string]*models.CSVData{
			"user": {
				EntityName: "User",
				Headers:    []string{"id", "name"},
				Rows: [][]string{
					{"user1", "User One"},
					{"user2", "User Two"},
					{"user3", "User Three"},
				},
			},
			"order": {
				EntityName: "Order",
				Headers:    []string{"id", "userId", "amount"},
				Rows: [][]string{
					{"order1", "user1", "100"},
					{"order2", "user2", "200"},
					{"order3", "user1", "300"},
					{"order4", "user3", "400"},
				},
			},
		}

		// Set up a relationship from Order.userId to User.id (N:1)
		link := models.RelationshipLink{
			FromEntityID:      "order",
			ToEntityID:        "user",
			FromAttribute:     "userId",
			ToAttribute:       "id",
			IsFromAttributeID: false,
			IsToAttributeID:   true,
		}

		// Set up the relationship map
		generator.relationshipMap = map[string][]models.RelationshipLink{
			"order": {link},
		}

		// Initialize unique attributes tracking
		generator.uniqueIdAttributes = map[string][]string{
			"user":  {"id"},
			"order": {"id"},
		}

		// Validate the relationship
		results := generator.ValidateRelationships()

		// Should have no errors
		if len(results) > 0 {
			t.Errorf("Expected no validation errors, but got %d issues", len(results))
			for _, result := range results {
				for _, err := range result.Errors {
					t.Errorf("Validation error: %s", err)
				}
			}
		}
	})

	// Test 2: Invalid relationship with broken references
	t.Run("Invalid relationship", func(t *testing.T) {
		// Create a test generator with auto-cardinality enabled
		generator := NewCSVGenerator(tempDir, 10, true)

		// Create test entities with an invalid relationship
		generator.EntityData = map[string]*models.CSVData{
			"user": {
				EntityName: "User",
				Headers:    []string{"id", "name"},
				Rows: [][]string{
					{"user1", "User One"},
					{"user2", "User Two"},
				},
			},
			"order": {
				EntityName: "Order",
				Headers:    []string{"id", "userId", "amount"},
				Rows: [][]string{
					{"order1", "user1", "100"},
					{"order2", "user2", "200"},
					{"order3", "user3", "300"}, // Invalid reference - user3 doesn't exist
					{"order4", "user4", "400"}, // Invalid reference - user4 doesn't exist
				},
			},
		}

		// Set up a relationship from Order.userId to User.id (N:1)
		link := models.RelationshipLink{
			FromEntityID:      "order",
			ToEntityID:        "user",
			FromAttribute:     "userId",
			ToAttribute:       "id",
			IsFromAttributeID: false,
			IsToAttributeID:   true,
		}

		// Set up the relationship map
		generator.relationshipMap = map[string][]models.RelationshipLink{
			"order": {link},
		}

		// Initialize unique attributes tracking
		generator.uniqueIdAttributes = map[string][]string{
			"user":  {"id"},
			"order": {"id"},
		}

		// Validate the relationship
		results := generator.ValidateRelationships()

		// Should have errors
		if len(results) == 0 {
			t.Errorf("Expected validation errors, but got none")
		} else {
			result := results[0]
			if result.InvalidRows != 2 {
				t.Errorf("Expected 2 invalid rows, got %d", result.InvalidRows)
			}
			if result.TotalRows != 4 {
				t.Errorf("Expected 4 total rows, got %d", result.TotalRows)
			}
		}
	})

	// Test 3: Unique attribute validation
	t.Run("Unique attribute validation", func(t *testing.T) {
		// Create a test generator
		generator := NewCSVGenerator(tempDir, 10, true)

		// Create test entities with duplicate values in unique fields
		generator.EntityData = map[string]*models.CSVData{
			"user": {
				EntityName: "User",
				Headers:    []string{"id", "email", "name"},
				Rows: [][]string{
					{"user1", "user1@example.com", "User One"},
					{"user2", "user2@example.com", "User Two"},
					{"user3", "user1@example.com", "User Three"}, // Duplicate email
					{"user4", "user4@example.com", "User Four"},
				},
			},
		}

		// Initialize unique attributes tracking
		generator.uniqueIdAttributes = map[string][]string{
			"user": {"id", "email"}, // Both id and email should be unique
		}

		// Initialize tracking maps
		generator.usedUniqueValues = map[string]map[string]bool{
			"user:id":    {"user1": true, "user2": true, "user3": true, "user4": true},
			"user:email": {"user1@example.com": true, "user2@example.com": true, "user4@example.com": true},
		}

		// Validate unique values
		errors := generator.ValidateUniqueValues()

		// Should have errors for the duplicate email
		if len(errors) == 0 {
			t.Errorf("Expected validation errors for duplicate email, but got none")
		} else {
			foundUserEntity := false
			foundDuplicateError := false

			for _, entityError := range errors {
				if entityError.EntityID == "user" {
					foundUserEntity = true

					for _, msg := range entityError.Messages {
						if msg == "Attribute email has 1 duplicate values" {
							foundDuplicateError = true
							break
						}
					}
				}
			}

			if !foundUserEntity {
				t.Errorf("Expected errors for user entity, but found none")
			}

			if !foundDuplicateError {
				t.Errorf("Expected error about duplicate email, but didn't find it")
			}
		}
	})
}
