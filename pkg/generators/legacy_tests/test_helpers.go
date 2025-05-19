package generators

import (
	"os"
	"strings"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHelper provides common functions and fixtures for tests
type TestHelper struct {
	T *testing.T
}

// NewTestHelper creates a new test helper with the given testing.T instance
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{T: t}
}

// CreateTempDir creates a temporary directory and returns its path along with a cleanup function
func (h *TestHelper) CreateTempDir(prefix string) (string, func()) {
	tempDir, err := os.MkdirTemp("", prefix+"-*")
	require.NoError(h.T, err, "Failed to create temporary directory")
	return tempDir, func() { _ = os.RemoveAll(tempDir) }
}

// CreateTestGenerator creates a new CSVGenerator with the specified parameters
func (h *TestHelper) CreateTestGenerator(outputDir string, dataVolume int, autoCardinality bool) *CSVGenerator {
	return NewCSVGenerator(outputDir, dataVolume, autoCardinality)
}

// CreateBasicEntities creates a map of basic test entities commonly used in tests
func (h *TestHelper) CreateBasicEntities() map[string]*models.CSVData {
	return map[string]*models.CSVData{
		"user": {
			ExternalId: "User",
			EntityName: "User",
			Headers:    []string{"id", "name", "email"},
			Rows: [][]string{
				{"user1", "User One", "user1@example.com"},
				{"user2", "User Two", "user2@example.com"},
			},
		},
		"order": {
			ExternalId: "Order",
			EntityName: "Order",
			Headers:    []string{"id", "userId", "amount"},
			Rows: [][]string{
				{"order1", "", "100"},
				{"order2", "", "200"},
			},
		},
	}
}

// CreateModelEntities creates a map of model entities for YAML-based tests
func (h *TestHelper) CreateModelEntities() map[string]models.Entity {
	return map[string]models.Entity{
		"user": {
			DisplayName: "User",
			ExternalId:  "User",
			Description: "User entity",
			Attributes: []models.Attribute{
				{
					Name:       "id",
					ExternalId: "id",
					Type:       "String",
					UniqueId:   true,
				},
				{
					Name:       "name",
					ExternalId: "name",
					Type:       "String",
				},
				{
					Name:       "email",
					ExternalId: "email",
					Type:       "String",
				},
			},
		},
		"order": {
			DisplayName: "Order",
			ExternalId:  "Order",
			Description: "Order entity",
			Attributes: []models.Attribute{
				{
					Name:       "id",
					ExternalId: "id",
					Type:       "String",
					UniqueId:   true,
				},
				{
					Name:       "userId",
					ExternalId: "userId",
					Type:       "String",
				},
				{
					Name:       "amount",
					ExternalId: "amount",
					Type:       "String",
				},
			},
		},
	}
}

// CreateTestRelationships creates a map of common test relationships
func (h *TestHelper) CreateTestRelationships() map[string]models.Relationship {
	return map[string]models.Relationship{
		"order_user": {
			DisplayName:   "order_user",
			Name:          "order_user",
			FromAttribute: "Order.userId",
			ToAttribute:   "User.id",
		},
	}
}

// CreateRelationshipLink creates a relationship link between entities
func (h *TestHelper) CreateRelationshipLink(fromEntity, toEntity, fromAttr, toAttr string) models.RelationshipLink {
	return models.RelationshipLink{
		FromEntityID:  fromEntity,
		ToEntityID:    toEntity,
		FromAttribute: fromAttr,
		ToAttribute:   toAttr,
	}
}

// SetupBasicGenerator creates a generator with common test entities already loaded
func (h *TestHelper) SetupBasicGenerator(tempDir string, dataVolume int, autoCardinality bool) *CSVGenerator {
	generator := h.CreateTestGenerator(tempDir, dataVolume, autoCardinality)
	
	// Add entities to the generator
	entities := h.CreateBasicEntities()
	for key, entity := range entities {
		generator.EntityData[key] = entity
	}
	
	return generator
}

// SetupModelBasedGenerator creates a generator with model entities and relationships
func (h *TestHelper) SetupModelBasedGenerator(tempDir string, dataVolume int, autoCardinality bool) *CSVGenerator {
	generator := h.CreateTestGenerator(tempDir, dataVolume, autoCardinality)
	
	// Setup the generator with model entities and relationships
	entities := h.CreateModelEntities()
	relationships := h.CreateTestRelationships()
	
	err := generator.Setup(entities, relationships)
	require.NoError(h.T, err, "Failed to setup generator with model entities")
	
	return generator
}

// VerifyCSVFile checks that a CSV file exists with the expected headers
func (h *TestHelper) VerifyCSVFile(filePath, expectedHeaders string) {
	_, err := os.Stat(filePath)
	assert.False(h.T, os.IsNotExist(err), "Expected file %s was not created", filePath)
	
	content, err := os.ReadFile(filePath)
	require.NoError(h.T, err, "Failed to read CSV file: %s", filePath)
	
	lines := strings.Split(string(content), "\n")
	assert.True(h.T, strings.HasPrefix(lines[0], expectedHeaders),
		"Header line should begin with expected headers")
}

// VerifyEntityHasRows checks that an entity has at least the expected number of rows
func (h *TestHelper) VerifyEntityHasRows(entity *models.CSVData, expectedRowCount int) {
	assert.GreaterOrEqual(h.T, len(entity.Rows), expectedRowCount, 
		"Entity should have at least %d rows", expectedRowCount)
}

// ExtractValues extracts values from a specific column in a set of rows
func ExtractValues(rows [][]string, colIndex int) map[string]bool {
	values := make(map[string]bool)
	for _, row := range rows {
		values[row[colIndex]] = true
	}
	return values
}

// FindColumnIndex finds the index of a column by name in a header slice
func FindColumnIndex(headers []string, name string) int {
	for i, header := range headers {
		if header == name {
			return i
		}
	}
	return -1
}

// VerifyRelationshipConsistency checks that references in one entity point to valid values in another
func (h *TestHelper) VerifyRelationshipConsistency(sourceEntity, targetEntity *models.CSVData, 
	sourceFieldIdx, targetFieldIdx int) {
	
	// Build a set of valid target values
	validTargetValues := ExtractValues(targetEntity.Rows, targetFieldIdx)
	
	// Check that all source rows reference valid target values
	for i, row := range sourceEntity.Rows {
		sourceValue := row[sourceFieldIdx]
		
		// Check that the value is not empty
		assert.NotEmpty(h.T, sourceValue, 
			"Row %d has empty value for relationship field", i)
		
		// Check that the value is valid
		assert.True(h.T, validTargetValues[sourceValue], 
			"Row %d has invalid reference value: %s", i, sourceValue)
	}
}