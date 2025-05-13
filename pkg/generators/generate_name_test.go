package generators

import (
	"strings"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
)

func TestGenerateName(t *testing.T) {
	// Create a test generator with predefined entity data
	generator := NewCSVGenerator("test-output", 10, false)

	// Setup different entity types to test different name generation branches
	entityTypes := []struct {
		entityID   string
		entityName string
		expected   string // partial string that should be contained in the result, empty means any non-empty result is valid
	}{
		{
			entityID:   "user",
			entityName: "User",
			expected:   "", // Person name, can't predict exact output
		},
		{
			entityID:   "role",
			entityName: "Role",
			expected:   "", // Job title, can't predict exact output
		},
		{
			entityID:   "team",
			entityName: "Team",
			expected:   "", // Department name, can't predict exact output but should be non-empty
		},
		{
			entityID:   "application",
			entityName: "Application",
			expected:   "", // App name, can't predict exact output
		},
		{
			entityID:   "product",
			entityName: "Product",
			expected:   "", // Product name, can't predict exact output
		},
		{
			entityID:   "location",
			entityName: "Location",
			expected:   "", // Location name, can't predict exact output
		},
		{
			entityID:   "company",
			entityName: "Company",
			expected:   "", // Company name, can't predict exact output
		},
		{
			entityID:   "project",
			entityName: "Project",
			expected:   "", // Project name, can't predict exact output
		},
		{
			entityID:   "category",
			entityName: "Category",
			expected:   "", // Category name, can't predict exact output
		},
		{
			entityID:   "unknown",
			entityName: "Unknown",
			expected:   "", // Default case, should use company name
		},
	}

	// Add all entity types to the generator
	for _, et := range entityTypes {
		generator.EntityData[et.entityID] = &models.CSVData{
			ExternalId: et.entityID,
			EntityName: et.entityName,
			Headers:    []string{"id", "name"},
			Rows:       [][]string{{"1", "Test"}},
		}
	}

	// Test name generation for each entity type
	for i, et := range entityTypes {
		t.Run(et.entityName, func(t *testing.T) {
			result := generator.generateName(i)

			// Verify the result is not empty
			if result == "" {
				t.Errorf("generateName returned empty string for entity type %s", et.entityName)
			}

			// If an expected substring is specified, check for it
			if et.expected != "" && !strings.Contains(result, et.expected) {
				t.Errorf("generateName for %s should contain '%s', got '%s'", et.entityName, et.expected, result)
			}

			// Check that the name is sanitized
			if strings.Contains(result, "\"") || strings.Contains(result, ",") {
				t.Errorf("Name should be sanitized, but found quotes or commas: %s", result)
			}
		})
	}

	// Test branch of function where entityID is empty/not found
	t.Run("EmptyEntityID", func(t *testing.T) {
		// Create a generator with empty entity data
		emptyGenerator := NewCSVGenerator("test-output", 10, false)

		// This should hit the fallback path
		result := emptyGenerator.generateName(0)

		if result == "" {
			t.Errorf("generateName should return a default name when entity not found")
		}
	})

	// Test the sanitizeName function directly
	t.Run("SanitizeName", func(t *testing.T) {
		cases := []struct {
			input    string
			expected string
		}{
			{"Normal Name", "Normal Name"},
			{"Name, with comma", "Name- with comma"},
			{"Name \"with\" quotes", "Name 'with' quotes"},
			{"Name, \"with\" both", "Name- 'with' both"},
		}

		for _, c := range cases {
			result := sanitizeName(c.input)
			if result != c.expected {
				t.Errorf("sanitizeName(%s) = %s; want %s", c.input, result, c.expected)
			}
		}
	})

	// Test corner cases for the findEntityByIndex method
	t.Run("FindEntityByIndex", func(t *testing.T) {
		// Clear the generator and set specific entities
		clearGenerator := NewCSVGenerator("test-output", 10, false)
		clearGenerator.EntityData["entity1"] = &models.CSVData{
			ExternalId: "entity1",
			EntityName: "Entity One",
		}

		// The function should return the first entity it finds
		result := clearGenerator.findEntityByIndex(0)
		if result != "entity1" {
			t.Errorf("findEntityByIndex should return 'entity1', got '%s'", result)
		}

		// Empty generator should return empty string
		emptyGenerator := NewCSVGenerator("test-output", 10, false)
		emptyResult := emptyGenerator.findEntityByIndex(0)
		if emptyResult != "" {
			t.Errorf("findEntityByIndex should return empty string when no entities, got '%s'", emptyResult)
		}
	})
}
