package generators

import (
	"os"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAutoCardinalityRelationshipConsistency verifies that relationships
// maintain consistency when auto-cardinality is enabled.
func TestAutoCardinalityRelationshipConsistency(t *testing.T) {
	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "relationship-consistency-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	defer func() { _ = os.RemoveAll(tempDir) }()

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

	// Set up a one-to-many relationship from User to Order
	relationships := map[string]models.Relationship{
		"user_orders": {
			DisplayName:   "User Orders",
			Name:          "user_orders",
			FromAttribute: "user_id",      // User's ID field (PK)
			ToAttribute:   "order_userId", // Order's userId field (FK)
		},
	}

	// Set up the generator
	err = generator.Setup(entities, relationships)
	require.NoError(t, err, "Failed to set up entities and relationships")

	// Generate the data
	err = generator.GenerateData()
	require.NoError(t, err, "Failed to generate data")

	// Verify relationship consistency in generated data
	validationResults := generator.ValidateRelationships()

	// There should be no relationship consistency issues
	assert.Empty(t, validationResults, "There should be no relationship consistency issues")

	// Check uniqueness constraints for unique IDs
	uniqueErrors := generator.ValidateUniqueValues()
	assert.Empty(t, uniqueErrors, "There should be no uniqueness errors")

	// Verify that orders reference valid user IDs
	validUserIds := make(map[string]bool)
	for _, row := range generator.EntityData["user"].Rows {
		// User.id is in the first column (index 0)
		validUserIds[row[0]] = true
	}

	userIdIndex := 1 // Order.userId is the second column (index 1)
	invalidReferences := 0
	
	for _, row := range generator.EntityData["order"].Rows {
		if userIdIndex < len(row) {
			userId := row[userIdIndex]
			if userId != "" && !validUserIds[userId] {
				invalidReferences++
			}
		}
	}

	assert.Zero(t, invalidReferences, "There should be no orders with invalid user references")
}
