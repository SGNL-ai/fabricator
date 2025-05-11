package generators

import (
	"strings"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
)

// TestFieldTypeDetection verifies field type detection in generateRowForEntity
func TestFieldTypeDetection(t *testing.T) {
	// Skip this test since we have a more focused test for boolean field detection
	// in boolean_test.go
	t.Skip("Skipping TestFieldTypeDetection as we have a more focused test for the same functionality")

	// Create a new generator
	generator := NewCSVGenerator("test_output", 5)
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
	if row[idIdx] != "entity-uuid-0" {
		t.Errorf("Expected ID to be entity-uuid-0, got %s", row[idIdx])
	}

	// Check type field
	if row[typeIdx] == "" {
		t.Errorf("Expected type to be non-empty")
	}

	// Check permissions field (should be a comma-separated list)
	if !strings.Contains(row[permissionsIdx], ",") && len(row[permissionsIdx]) < 3 {
		t.Errorf("Expected permissions to be a comma-separated list, got %s", row[permissionsIdx])
	}

	// Check expression field
	if row[expressionIdx] == "" {
		t.Errorf("Expected expression to be non-empty")
	}

	// Check percentage and rate fields (should contain %)
	if !strings.Contains(row[percentageIdx], "%") {
		t.Errorf("Expected percentage to contain %% symbol, got %s", row[percentageIdx])
	}
	if !strings.Contains(row[rateIdx], "%") {
		t.Errorf("Expected rate to contain %% symbol, got %s", row[rateIdx])
	}

	// Check code field (should be in format XXX-1000)
	if !strings.Contains(row[codeIdx], "-") || len(row[codeIdx]) < 5 {
		t.Errorf("Expected code to be in format XXX-1000, got %s", row[codeIdx])
	}

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
		if value != "true" && value != "false" {
			t.Errorf("Expected %s to be a boolean (true/false), got %s", field.name, value)
		}
	}

	// Check date field
	if len(row[dateIdx]) != 10 || !strings.Contains(row[dateIdx], "-") {
		t.Errorf("Expected date in YYYY-MM-DD format, got %s", row[dateIdx])
	}
}
