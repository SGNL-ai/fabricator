package generators

import (
	"os"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
)

func TestUniqueAttributeValues(t *testing.T) {
	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "unique-attributes-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Test case: Entity with multiple unique attributes
	t.Run("Entity with multiple unique attributes", func(t *testing.T) {
		// Create a test generator
		generator := NewCSVGenerator(tempDir, 50, false)

		// Define an entity with multiple attributes marked as uniqueId=true
		entities := map[string]models.Entity{
			"entity1": {
				DisplayName: "TestEntity",
				ExternalId:  "Test/TestEntity",
				Attributes: []models.Attribute{
					{
						Name:           "id",
						ExternalId:     "id",
						Description:    "Unique identifier",
						Type:           "String",
						UniqueId:       true,
						AttributeAlias: "attr1",
					},
					{
						Name:           "code",
						ExternalId:     "code",
						Description:    "Unique code",
						Type:           "String",
						UniqueId:       true,
						AttributeAlias: "attr2",
					},
					{
						Name:           "name",
						ExternalId:     "name",
						Description:    "Display name",
						Type:           "String",
						UniqueId:       false,
						AttributeAlias: "attr3",
					},
				},
			},
		}

		// Set up the generator (including tracking uniqueId attributes)
		generator.Setup(entities, map[string]models.Relationship{})

		// Generate data
		for i := 0; i < generator.DataVolume; i++ {
			row := generator.generateRowForEntity("entity1", i)
			generator.EntityData["entity1"].Rows = append(generator.EntityData["entity1"].Rows, row)
		}

		// Verify that each unique attribute has unique values across all rows
		uniqueValues := make(map[string]map[string]bool)
		for _, attr := range entities["entity1"].Attributes {
			if attr.UniqueId {
				uniqueValues[attr.ExternalId] = make(map[string]bool)
			}
		}

		rows := generator.EntityData["entity1"].Rows
		headers := generator.EntityData["entity1"].Headers

		for _, row := range rows {
			for i, header := range headers {
				if _, isUnique := uniqueValues[header]; isUnique {
					value := row[i]
					// Check if we've seen this value before
					if uniqueValues[header][value] {
						t.Errorf("Duplicate value %s found for unique attribute %s", value, header)
					}
					uniqueValues[header][value] = true
				}
			}
		}

		// Verify that the tracking maps were properly populated
		t.Run("Verify uniqueIdAttributes map", func(t *testing.T) {
			uniqueAttrs, exists := generator.uniqueIdAttributes["entity1"]
			if !exists {
				t.Error("uniqueIdAttributes map does not contain entity1")
			}

			expectedUniqueAttrs := []string{"id", "code"}
			for _, attr := range expectedUniqueAttrs {
				found := false
				for _, uniqueAttr := range uniqueAttrs {
					if uniqueAttr == attr {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("uniqueIdAttributes map does not contain attribute %s", attr)
				}
			}
		})
	})

	// Test case: Auto-cardinality with unique attributes
	t.Run("Auto-cardinality with unique attributes", func(t *testing.T) {
		// Create a test generator with auto-cardinality enabled
		generator := NewCSVGenerator(tempDir, 20, true)

		// Define entities for a one-to-many relationship
		entities := map[string]models.Entity{
			"user": {
				DisplayName: "User",
				ExternalId:  "Test/User",
				Attributes: []models.Attribute{
					{
						Name:           "id",
						ExternalId:     "id",
						Description:    "Unique user ID",
						Type:           "String",
						UniqueId:       true,
						AttributeAlias: "user_id",
					},
					{
						Name:           "email",
						ExternalId:     "email",
						Description:    "User email address",
						Type:           "String",
						UniqueId:       true,
						AttributeAlias: "user_email",
					},
					{
						Name:           "name",
						ExternalId:     "name",
						Description:    "User name",
						Type:           "String",
						UniqueId:       false,
						AttributeAlias: "user_name",
					},
				},
			},
			"order": {
				DisplayName: "Order",
				ExternalId:  "Test/Order",
				Attributes: []models.Attribute{
					{
						Name:           "id",
						ExternalId:     "id",
						Description:    "Unique order ID",
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
						Name:           "orderNumber",
						ExternalId:     "orderNumber",
						Description:    "Unique order number",
						Type:           "String",
						UniqueId:       true,
						AttributeAlias: "order_number",
					},
				},
			},
		}

		// Set up a one-to-many relationship from User to Order
		relationships := map[string]models.Relationship{
			"user_orders": {
				DisplayName:   "User Orders",
				Name:          "user_orders",
				FromAttribute: "user_id",
				ToAttribute:   "order_userId",
			},
		}

		// Set up the generator
		generator.Setup(entities, relationships)

		// Generate the data
		generator.GenerateData()

		// Verify that all unique attributes have unique values even after the relationship processing
		userUniqueValues := make(map[string]map[string]bool)
		for _, attr := range entities["user"].Attributes {
			if attr.UniqueId {
				userUniqueValues[attr.ExternalId] = make(map[string]bool)
			}
		}

		orderUniqueValues := make(map[string]map[string]bool)
		for _, attr := range entities["order"].Attributes {
			if attr.UniqueId {
				orderUniqueValues[attr.ExternalId] = make(map[string]bool)
			}
		}

		// Check user unique values
		userRows := generator.EntityData["user"].Rows
		userHeaders := generator.EntityData["user"].Headers

		for _, row := range userRows {
			for i, header := range userHeaders {
				if _, isUnique := userUniqueValues[header]; isUnique {
					value := row[i]
					if userUniqueValues[header][value] {
						t.Errorf("Duplicate value %s found for unique user attribute %s", value, header)
					}
					userUniqueValues[header][value] = true
				}
			}
		}

		// Check order unique values
		orderRows := generator.EntityData["order"].Rows
		orderHeaders := generator.EntityData["order"].Headers

		for _, row := range orderRows {
			for i, header := range orderHeaders {
				if _, isUnique := orderUniqueValues[header]; isUnique {
					value := row[i]
					if orderUniqueValues[header][value] {
						t.Errorf("Duplicate value %s found for unique order attribute %s", value, header)
					}
					orderUniqueValues[header][value] = true
				}
			}
		}

		// Check that relationships are correct
		userIdCol := -1
		for i, header := range userHeaders {
			if header == "id" {
				userIdCol = i
				break
			}
		}

		orderUserIdCol := -1
		for i, header := range orderHeaders {
			if header == "userId" {
				orderUserIdCol = i
				break
			}
		}

		if userIdCol >= 0 && orderUserIdCol >= 0 {
			// Print debug information
			t.Logf("User data - %d rows", len(userRows))
			for i, row := range userRows {
				if i < 3 { // Print just a few rows to avoid overwhelming logs
					t.Logf("User %d: id = %s", i, row[userIdCol])
				}
			}
			t.Logf("Order data - %d rows", len(orderRows))
			for i, row := range orderRows {
				if i < 3 { // Print just a few rows
					t.Logf("Order %d: userId = %s", i, row[orderUserIdCol])
				}
			}

			// Build a map of valid user IDs
			validUserIds := make(map[string]bool)
			for _, row := range userRows {
				validUserIds[row[userIdCol]] = true
			}
			t.Logf("Valid user IDs count: %d", len(validUserIds))

			// Count valid and invalid references
			validRefs := 0
			invalidRefs := 0

			// Check that all orders reference valid users
			for _, row := range orderRows {
				userId := row[orderUserIdCol]
				if validUserIds[userId] {
					validRefs++
				} else {
					invalidRefs++
					if invalidRefs <= 3 { // Limit error messages
						t.Errorf("Order references invalid userId: %s", userId)
					}
				}
			}

			// Print summary
			if invalidRefs > 0 {
				t.Errorf("Summary: %d valid references, %d invalid references", validRefs, invalidRefs)
			} else {
				t.Logf("All references are valid: %d total", validRefs)
			}
		}
	})
}
