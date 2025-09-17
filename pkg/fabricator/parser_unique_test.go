package fabricator

import (
	"os"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"gopkg.in/yaml.v3"
)

func TestUniqueAttributeValidation(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name()) // Ignore error on cleanup
	}()

	// Test 1: Entity with a uniqueId attribute should pass validation
	t.Run("Entity with uniqueId attribute", func(t *testing.T) {
		definition := models.SORDefinition{
			DisplayName:               "Test SOR",
			Description:               "Test SOR for validation",
			DefaultSyncFrequency:      "DAILY",
			DefaultSyncMinInterval:    1,
			DefaultApiCallFrequency:   "HOURLY",
			DefaultApiCallMinInterval: 1,
			AdapterConfig:             "test",
			Entities: map[string]models.Entity{
				"entity1": {
					DisplayName: "Entity1",
					ExternalId:  "Test/Entity1",
					Description: "Test entity",
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
							UniqueId:   false,
						},
					},
				},
			},
		}

		// Write the YAML to the temp file
		yamlData, err := yaml.Marshal(definition)
		if err != nil {
			t.Fatalf("Failed to marshal YAML: %v", err)
		}

		if err := os.WriteFile(tmpFile.Name(), yamlData, 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		// Create a parser and parse the file
		parser := NewParser(tmpFile.Name())
		err = parser.Parse()
		if err != nil {
			t.Errorf("Validation failed unexpectedly: %v", err)
		}
	})

	// Test 2: Entity without a uniqueId attribute should fail validation
	t.Run("Entity without uniqueId attribute", func(t *testing.T) {
		definition := models.SORDefinition{
			DisplayName:               "Test SOR",
			Description:               "Test SOR for validation",
			DefaultSyncFrequency:      "DAILY",
			DefaultSyncMinInterval:    1,
			DefaultApiCallFrequency:   "HOURLY",
			DefaultApiCallMinInterval: 1,
			AdapterConfig:             "test",
			Entities: map[string]models.Entity{
				"entity1": {
					DisplayName: "Entity1",
					ExternalId:  "Test/Entity1",
					Description: "Test entity",
					Attributes: []models.Attribute{
						{
							Name:       "id",
							ExternalId: "id",
							Type:       "String",
							UniqueId:   false,
						},
						{
							Name:       "name",
							ExternalId: "name",
							Type:       "String",
							UniqueId:   false,
						},
					},
				},
			},
		}

		// Write the YAML to the temp file
		yamlData, err := yaml.Marshal(definition)
		if err != nil {
			t.Fatalf("Failed to marshal YAML: %v", err)
		}

		if err := os.WriteFile(tmpFile.Name(), yamlData, 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		// Create a parser and parse the file
		parser := NewParser(tmpFile.Name())
		err = parser.Parse()
		if err == nil {
			t.Error("Validation should have failed but didn't")
		}
	})

	// Test 3: Verify uniqueId default value is false
	t.Run("Default uniqueId value should be false", func(t *testing.T) {
		yamlContent := `
displayName: Test SOR
description: Test SOR for validation
entities:
  entity1:
    displayName: Entity1
    externalId: Test/Entity1
    description: Test entity
    attributes:
      - name: id
        externalId: id
        type: String
        uniqueId: true
      - name: name
        externalId: name
        type: String
        # uniqueId is intentionally omitted and should default to false
`

		if err := os.WriteFile(tmpFile.Name(), []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		// Create a parser and parse the file
		parser := NewParser(tmpFile.Name())
		err = parser.Parse()
		if err != nil {
			t.Errorf("Validation failed unexpectedly: %v", err)
		}

		// Check that the omitted uniqueId defaulted to false
		nameAttr := parser.Definition.Entities["entity1"].Attributes[1]
		if nameAttr.UniqueId {
			t.Error("UniqueId field should default to false when omitted")
		}
	})
}
