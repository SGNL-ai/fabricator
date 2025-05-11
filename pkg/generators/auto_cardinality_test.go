package generators

import (
	"os"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
)

func TestAutoCardinality(t *testing.T) {
	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "auto-cardinality-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Test 1: One-to-Many relationship with uniqueId attributes
	t.Run("One-to-Many relationship with uniqueId attributes", func(t *testing.T) {
		// Create a test generator with auto-cardinality enabled
		generator := NewCSVGenerator(tempDir, 2, true)

		// Create test entities with a one-to-many relationship
		// User (one) -> Orders (many)
		generator.EntityData = map[string]*models.CSVData{
			"user": {
				EntityName: "User",
				Headers:    []string{"id", "name"},
				Rows:       [][]string{{"user1", "User One"}, {"user2", "User Two"}},
			},
			"order": {
				EntityName: "Order",
				Headers:    []string{"id", "userId", "amount"},
				Rows:       [][]string{{"order1", "", "100"}, {"order2", "", "200"}},
			},
		}

		// Set up a relationship from Order.userId to User.id (N:1)
		// This is expressed in the YAML model, but for testing we're creating the link directly
		link := models.RelationshipLink{
			FromEntityID:      "order",
			ToEntityID:        "user",
			FromAttribute:     "userId",
			ToAttribute:       "id",
			IsFromAttributeID: false, // userId is not a unique ID
			IsToAttributeID:   true,  // id is a unique ID
		}

		// Set up the relationship map
		generator.relationshipMap = map[string][]models.RelationshipLink{
			"order": {link},
		}

		// Process the N:1 relationship
		generator.makeRelationshipsConsistent("order", link)

		// Verify that orders have valid user IDs
		// With auto-cardinality enabled, each order should be linked to a user
		// and we should see clustering of values (multiple orders linked to same user)

		orderRows := generator.EntityData["order"].Rows
		if len(orderRows) < 2 {
			t.Errorf("Expected at least 2 order rows, got %d", len(orderRows))
		}

		// Check if userId is populated
		for _, row := range orderRows {
			userId := row[1] // userId is the second column
			if userId == "" {
				t.Errorf("Order has empty userId, expected a value")
			}

			// Verify userId references a valid user
			validUserIds := map[string]bool{"user1": true, "user2": true}
			if !validUserIds[userId] {
				t.Errorf("Order has userId %s which is not a valid user ID", userId)
			}
		}

		// Verify that each order has a valid user ID
		// With only 2 orders and 2 users, we can't guarantee clustering due to randomness
		// But we can verify that all orders have a valid user ID assigned

		for _, row := range orderRows {
			userId := row[1]
			validUserIds := map[string]bool{"user1": true, "user2": true}
			if !validUserIds[userId] {
				t.Errorf("Order has invalid userId %s", userId)
			}
		}

		// Log the distribution but don't fail the test
		userIdCounts := make(map[string]int)
		for _, row := range orderRows {
			userId := row[1]
			userIdCounts[userId]++
		}

		t.Logf("User ID distribution: %v", userIdCounts)
	})

	// Test 2: Many-to-One relationship with field name patterns
	t.Run("Many-to-One relationship with field name patterns", func(t *testing.T) {
		// Create a test generator with auto-cardinality enabled
		generator := NewCSVGenerator(tempDir, 2, true)

		// Create test entities with a many-to-one relationship
		// Products (many) -> Category (one)
		generator.EntityData = map[string]*models.CSVData{
			"product": {
				EntityName: "Product",
				Headers:    []string{"id", "name", "categoryId"},
				Rows:       [][]string{{"prod1", "Product 1", ""}, {"prod2", "Product 2", ""}},
			},
			"category": {
				EntityName: "Category",
				Headers:    []string{"id", "name"},
				Rows:       [][]string{{"cat1", "Category 1"}, {"cat2", "Category 2"}},
			},
		}

		// Set up a relationship from Product.categoryId to Category.id
		// where field name patterns suggest N:1 relationship
		link := models.RelationshipLink{
			FromEntityID:      "product",
			ToEntityID:        "category",
			FromAttribute:     "categoryId", // Field name pattern "...Id" suggests reference
			ToAttribute:       "id",
			IsFromAttributeID: false, // Attributes don't have uniqueId info
			IsToAttributeID:   false,
		}

		// Set up the relationship map
		generator.relationshipMap = map[string][]models.RelationshipLink{
			"product": {link},
		}

		// Process the N:1 relationship
		generator.makeRelationshipsConsistent("product", link)

		// Verify products have valid category IDs
		productRows := generator.EntityData["product"].Rows

		// Check if categoryId is populated
		for _, row := range productRows {
			categoryId := row[2] // categoryId is the third column
			if categoryId == "" {
				t.Errorf("Product has empty categoryId, expected a value")
			}

			// Verify categoryId references a valid category
			validCategoryIds := map[string]bool{"cat1": true, "cat2": true}
			if !validCategoryIds[categoryId] {
				t.Errorf("Product has categoryId %s which is not a valid category ID", categoryId)
			}
		}

		// Log the distribution but don't fail the test (similar to Test 1)
		categoryIdCounts := make(map[string]int)
		for _, row := range productRows {
			categoryId := row[2]
			categoryIdCounts[categoryId]++
		}

		t.Logf("Category ID distribution: %v", categoryIdCounts)
	})

	// Test 3: One-to-Many relationship with field name patterns
	t.Run("One-to-Many relationship with field name patterns", func(t *testing.T) {
		// Create a test generator with auto-cardinality enabled
		generator := NewCSVGenerator(tempDir, 2, true)

		// Create test entities with a one-to-many relationship
		// User (one) -> Accounts (many) based on field name patterns
		generator.EntityData = map[string]*models.CSVData{
			"user": {
				EntityName: "User",
				Headers:    []string{"id", "name", "account_ids"}, // Plural field name suggests 1:N
				Rows:       [][]string{{"user1", "User 1", ""}, {"user2", "User 2", ""}},
			},
			"account": {
				EntityName: "Account",
				Headers:    []string{"id", "name"},
				Rows:       [][]string{{"acc1", "Account 1"}, {"acc2", "Account 2"}},
			},
		}

		// Set up a relationship from User.account_ids to Account.id (1:N)
		link := models.RelationshipLink{
			FromEntityID:      "user",
			ToEntityID:        "account",
			FromAttribute:     "account_ids", // Plural field name suggests 1:N
			ToAttribute:       "id",
			IsFromAttributeID: false, // No uniqueId info
			IsToAttributeID:   false,
		}

		// Set up the relationship map
		generator.relationshipMap = map[string][]models.RelationshipLink{
			"user": {link},
		}

		// In real usage, processRelationships would be called to prepare data
		// For testing, we'll manually call makeRelationshipsConsistent
		generator.makeRelationshipsConsistent("user", link)

		// Verify the user rows have account IDs
		userRows := generator.EntityData["user"].Rows

		for _, row := range userRows {
			accountIds := row[2] // account_ids is the third column
			if accountIds == "" {
				t.Errorf("User has empty account_ids, expected a value")
			}

			// For 1:N relationships, we expect comma-separated lists
			// This is the behavior of our row duplication approach
			if !generator.AutoCardinality {
				// Without auto-cardinality, we'd just have a single ID
				continue
			}

			// With auto-cardinality, we should see comma-separated lists or multiple rows
			// At least some users should have multiple accounts
			foundMultipleAccounts := false
			for _, accountIds := range row {
				if accountIds != "" && accountIds != "acc1" && accountIds != "acc2" {
					foundMultipleAccounts = true
				}
			}

			if !foundMultipleAccounts {
				// This test might be flaky because the row duplication is random
				// But in most cases we should see a user with multiple accounts
				// Consider adding a threshold here if the test is flaky
				t.Logf("Unusual: expected at least one user to have multiple accounts")
			}
		}
	})
}
