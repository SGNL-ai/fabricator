package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParserValidation(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "parser-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	t.Run("Parse valid YAML", func(t *testing.T) {
		// Create a valid YAML file
		yamlContent := `displayName: Test SOR
description: Test System of Record
entities:
  user:
    displayName: User
    externalId: User
    attributes:
      - name: id
        externalId: id
        type: String
        uniqueId: true`

		yamlPath := filepath.Join(tempDir, "valid.yaml")
		if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write YAML file: %v", err)
		}

		parser := NewParser(yamlPath)
		err := parser.Parse()
		if err != nil {
			t.Errorf("Parse() failed on valid YAML: %v", err)
		}

		// Verify parsed data
		if parser.Definition.DisplayName != "Test SOR" {
			t.Errorf("Expected DisplayName 'Test SOR', got '%s'", parser.Definition.DisplayName)
		}
	})

	t.Run("Parse invalid YAML", func(t *testing.T) {
		// Create an invalid YAML file
		invalidYAML := `displayName: Test SOR
description: Test System of Record
entities:
  user:
    displayName: User
    # Missing required externalId field`

		yamlPath := filepath.Join(tempDir, "invalid.yaml")
		if err := os.WriteFile(yamlPath, []byte(invalidYAML), 0644); err != nil {
			t.Fatalf("Failed to write YAML file: %v", err)
		}

		parser := NewParser(yamlPath)
		err := parser.Parse()
		if err == nil {
			t.Error("Parse() should fail on invalid YAML")
		}
	})

	t.Run("Parse non-existent file", func(t *testing.T) {
		parser := NewParser("/nonexistent/file.yaml")
		err := parser.Parse()
		if err == nil {
			t.Error("Parse() should fail for non-existent file")
		}
	})

	t.Run("Parse YAML with invalid schema", func(t *testing.T) {
		// Create YAML that fails schema validation
		invalidSchema := `displayName: Test SOR
# Missing required description field
entities:
  user:
    displayName: User
    externalId: User
    attributes:
      - name: id
        externalId: id
        type: String
        uniqueId: true`

		yamlPath := filepath.Join(tempDir, "invalid_schema.yaml")
		if err := os.WriteFile(yamlPath, []byte(invalidSchema), 0644); err != nil {
			t.Fatalf("Failed to write YAML file: %v", err)
		}

		parser := NewParser(yamlPath)
		err := parser.Parse()
		if err == nil {
			t.Error("Parse() should fail on schema validation")
		}
		if !strings.Contains(err.Error(), "schema validation") {
			t.Errorf("Expected schema validation error, got: %v", err)
		}
	})

	t.Run("Parse YAML missing display name", func(t *testing.T) {
		// Missing DisplayName - should fail schema validation
		invalidYAML := `# Missing displayName
description: Test description
entities:
  user:
    displayName: User
    externalId: User
    attributes:
      - name: id
        externalId: id
        type: String
        uniqueId: true`

		yamlPath := filepath.Join(tempDir, "missing_display_name.yaml")
		if err := os.WriteFile(yamlPath, []byte(invalidYAML), 0644); err != nil {
			t.Fatalf("Failed to write YAML file: %v", err)
		}

		parser := NewParser(yamlPath)
		err := parser.Parse()
		if err == nil {
			t.Error("Parse() should fail when DisplayName is missing")
		}
	})

	t.Run("Parse YAML missing description", func(t *testing.T) {
		// Missing Description - should fail schema validation
		invalidYAML := `displayName: Test SOR
# Missing description
entities:
  user:
    displayName: User
    externalId: User
    attributes:
      - name: id
        externalId: id
        type: String
        uniqueId: true`

		yamlPath := filepath.Join(tempDir, "missing_description.yaml")
		if err := os.WriteFile(yamlPath, []byte(invalidYAML), 0644); err != nil {
			t.Fatalf("Failed to write YAML file: %v", err)
		}

		parser := NewParser(yamlPath)
		err := parser.Parse()
		if err == nil {
			t.Error("Parse() should fail when Description is missing")
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("Valid definition", func(t *testing.T) {
		parser := &Parser{
			Definition: &SORDefinition{
				DisplayName: "Test SOR",
				Description: "Test description",
				Entities: map[string]Entity{
					"user": {
						DisplayName: "User",
						ExternalId:  "User",
						Attributes: []Attribute{
							{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
						},
					},
				},
			},
		}
		err := parser.validate()
		if err != nil {
			t.Errorf("validate() failed on valid definition: %v", err)
		}
	})

	t.Run("Empty entities", func(t *testing.T) {
		parser := &Parser{
			Definition: &SORDefinition{
				DisplayName: "Test SOR",
				Description: "Test description",
				Entities:    map[string]Entity{}, // Empty entities
			},
		}
		err := parser.validate()
		if err == nil {
			t.Error("validate() should fail when Entities is empty")
		}
	})
}
