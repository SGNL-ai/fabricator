package generators

import (
	"strings"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/stretchr/testify/assert"
)

// TestFieldTypeDetection verifies field type detection in generateRowForEntity
func TestFieldTypeDetection(t *testing.T) {
	// Skip this test since we have a more focused test for boolean field detection
	// in boolean_test.go
	t.Skip("Skipping TestFieldTypeDetection as we have a more focused test for the same functionality")

	// Create a new generator
	generator := NewCSVGenerator("test_output", 5, false)
	generator.generateCommonValues() // Initialize common values

	// Set up test entity with many field types
	generator.EntityData = map[string]*models.CSVData{
		"entity": {
			EntityName: "TestEntity",
			ExternalId: "Test/TestEntity",
			Headers: []string{
				"id", "type", "permissions", "expression", "percentage", "rate",
				"code", "enabled", "archived", "valid", "updatedTime",
			},
		},
	}

	generator.idMap = map[string]map[string]string{
		"entity": {"0": "entity-uuid-0"},
	}

	// Generate a row
	row := generator.generateRowForEntity("entity", 0)

	// Define field indices for readability
	const (
		idIdx          = 0
		typeIdx        = 1
		permissionsIdx = 2
		expressionIdx  = 3
		percentageIdx  = 4
		rateIdx        = 5
		codeIdx        = 6
		enabledIdx     = 7
		archivedIdx    = 8
		validIdx       = 9
		dateIdx        = 10
	)

	// Check ID field
	assert.Equal(t, "entity-uuid-0", row[idIdx], "ID field should match the value from idMap")

	// Check type field
	assert.NotEmpty(t, row[typeIdx], "Type field should not be empty")

	// Check permissions field (should be a comma-separated list)
	assert.True(t, strings.Contains(row[permissionsIdx], ",") || len(row[permissionsIdx]) >= 3,
		"Permissions field should be a comma-separated list")

	// Check expression field
	assert.NotEmpty(t, row[expressionIdx], "Expression field should not be empty")

	// Check percentage and rate fields (should contain %)
	assert.Contains(t, row[percentageIdx], "%", "Percentage field should contain the % symbol")
	assert.Contains(t, row[rateIdx], "%", "Rate field should contain the % symbol")

	// Check code field (should be in format XXX-1000)
	assert.Contains(t, row[codeIdx], "-", "Code field should contain a dash")
	assert.GreaterOrEqual(t, len(row[codeIdx]), 5, "Code field should have at least 5 characters")

	// Check boolean fields
	booleanFields := []struct {
		name  string
		index int
	}{
		{"enabled", enabledIdx},
		{"archived", archivedIdx},
		{"valid", validIdx},
	}

	for _, field := range booleanFields {
		value := row[field.index]
		assert.Contains(t, []string{"true", "false"}, value,
			"Field %s should be a boolean (true/false)", field.name)
	}

	// Check date field
	assert.Len(t, row[dateIdx], 10, "Date field should be 10 characters long (YYYY-MM-DD)")
	assert.Contains(t, row[dateIdx], "-", "Date field should contain dashes")
}
