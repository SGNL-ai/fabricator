package generators

import (
	"os"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
)

func TestAutoCardinalityRelationshipConsistency(t *testing.T) {
	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "relationship-consistency-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Test case: Demonstrate a foreign key reference issue with auto-cardinality
	t.Run("Auto-cardinality foreign key consistency", func(t *testing.T) {
		// Create a test generator with auto-cardinality enabled
		generator := NewCSVGenerator(tempDir, 50, true)

		// Define a simple user-order relationship (1:N)
		entities := map[string]models.Entity{
			"user": {
				DisplayName: "User",
				ExternalId:  "Users/User",
				Attributes: []models.Attribute{
					{
						Name:           "id",
						ExternalId:     "id",
						Description:    "User ID",
						Type:           "String",
						UniqueId:       true,
						AttributeAlias: "user_id",
					},
					{
						Name:        "name",
						ExternalId:  "name",
						Description: "User name",
						Type:        "String",
					},
				},
			},
			"order": {
				DisplayName: "Order",
				ExternalId:  "Orders/Order",
				Attributes: []models.Attribute{
					{
						Name:           "id",
						ExternalId:     "id",
						Description:    "Order ID",
						Type:           "String",
						UniqueId:       true,
						AttributeAlias: "order_id",
					},
					{
						Name:           "userId",
						ExternalId:     "userId",
						Description:    "Reference to user",
						Type:           "String",
						UniqueId:       false,
						AttributeAlias: "order_userId",
					},
					{
						Name:        "product",
						ExternalId:  "product",
						Description: "Ordered product",
						Type:        "String",
					},
				},
			},
		}

		// Set up a one-to-many relationship from User to Order (i.e., each order references a user)
		relationships := map[string]models.Relationship{
			"user_orders": {
				DisplayName:   "User Orders",
				Name:          "user_orders",
				FromAttribute: "user_id",      // User's ID field (PK)
				ToAttribute:   "order_userId", // Order's userId field (FK)
			},
		}

		// Set up the generator
		generator.Setup(entities, relationships)

		// Debug: print the relationship map to understand how it's set up
		t.Logf("Relationship map: %+v", generator.relationshipMap)

		// Generate the data
		generator.GenerateData()

		// Debug: print a few rows from each entity to inspect the data
		t.Logf("User entity:")
		for i, row := range generator.EntityData["user"].Rows[:5] {
			t.Logf("  Row %d: %v", i, row)
		}

		t.Logf("Order entity:")
		for i, row := range generator.EntityData["order"].Rows[:5] {
			t.Logf("  Row %d: %v", i, row)
		}

		// Verify relationship consistency in generated data
		validationResults := generator.ValidateRelationships()

		// Log all the valid user IDs
		validUserIds := make(map[string]bool)
		for _, row := range generator.EntityData["user"].Rows {
			// User.id is in the first column (index 0)
			validUserIds[row[0]] = true
		}
		t.Logf("Valid user IDs count: %d", len(validUserIds))

		// The test should fail if there are relationship consistency issues
		if len(validationResults) > 0 {
			for _, result := range validationResults {
				t.Errorf("Relationship consistency issue: %s (%s) → %s (%s): %d invalid references out of %d rows",
					result.FromEntity, result.FromEntityFile,
					result.ToEntity, result.ToEntityFile,
					result.InvalidRows, result.TotalRows)

				// Show a few detailed errors
				maxErrorsToShow := 5
				for i, err := range result.Errors {
					if i < maxErrorsToShow {
						t.Errorf("  Error: %s", err)
					} else {
						t.Errorf("  ... and %d more errors", len(result.Errors)-maxErrorsToShow)
						break
					}
				}
			}
		}

		// Check uniqueness constraints
		uniqueErrors := generator.ValidateUniqueValues()
		if len(uniqueErrors) > 0 {
			for _, entityError := range uniqueErrors {
				t.Errorf("Uniqueness errors in entity %s (%s)", entityError.EntityID, entityError.EntityFile)
				for _, msg := range entityError.Messages {
					t.Errorf("  Error: %s", msg)
				}
			}
		}
	})
}
