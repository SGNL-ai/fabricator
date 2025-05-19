package generators

import (
	"path/filepath"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUsingHelpers demonstrates how to use the test helper functions
func TestUsingHelpers(t *testing.T) {
	// Create a test helper
	helper := NewTestHelper(t)

	// Create a temporary directory for test output
	tempDir, cleanup := helper.CreateTempDir("helper-example")
	defer cleanup()

	// Use the helper to create a pre-configured generator
	generator := helper.SetupBasicGenerator(tempDir, 5, false)

	// Set up a relationship from Order.userId to User.id
	link := helper.CreateRelationshipLink("order", "user", "userId", "id")

	// Set up the relationship map
	generator.relationshipMap = map[string][]models.RelationshipLink{
		"order": {link},
	}

	// Make relationships consistent (this is the core functionality we're testing)
	generator.makeRelationshipsConsistent("order", link)

	// Get the relevant data
	orderData := generator.EntityData["order"]
	userData := generator.EntityData["user"]

	// Find column indices
	userIdIdx := FindColumnIndex(orderData.Headers, "userId")
	idIdx := FindColumnIndex(userData.Headers, "id")

	// Verify the relationship consistency
	helper.VerifyRelationshipConsistency(orderData, userData, userIdIdx, idIdx)

	// Write CSV files
	err := generator.WriteCSVFiles()
	require.NoError(t, err, "WriteCSVFiles should not fail")

	// Verify the CSV files were created with expected content
	helper.VerifyCSVFile(filepath.Join(tempDir, "User.csv"), "id,name,email")
	helper.VerifyCSVFile(filepath.Join(tempDir, "Order.csv"), "id,userId,amount")

	// Additional custom assertions can be added as needed
	assert.Len(t, orderData.Rows, 2, "Should have 2 orders")
}

// TestSetupModelBasedGenerator demonstrates the model-based generator setup
func TestSetupModelBasedGenerator(t *testing.T) {
	// Create a test helper
	helper := NewTestHelper(t)

	// Create a temporary directory
	tempDir, cleanup := helper.CreateTempDir("model-generator")
	defer cleanup()

	// Use the helper to create a model-based generator
	generator := helper.SetupModelBasedGenerator(tempDir, 5, false)

	// Generate data
	err := generator.GenerateData()
	require.NoError(t, err, "Data generation should not fail")

	// Write CSV files
	err = generator.WriteCSVFiles()
	require.NoError(t, err, "Writing CSV files should not fail")

	// Verify files exist with correct headers
	helper.VerifyCSVFile(filepath.Join(tempDir, "User.csv"), "id,name,email")
	helper.VerifyCSVFile(filepath.Join(tempDir, "Order.csv"), "id,userId,amount")
}
