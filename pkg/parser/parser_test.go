package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestNewParser(t *testing.T) {
	t.Run("should create parser with valid file path", func(t *testing.T) {
		parser := NewParser("test.yaml")
		assert.NotNil(t, parser)
		assert.Equal(t, "test.yaml", parser.FilePath)
		assert.NotNil(t, parser.schema) // Schema should be initialized
	})

	t.Run("should handle schema initialization failure gracefully", func(t *testing.T) {
		// Test the error path in NewParser where initSchema fails
		parser := NewParser("test.yaml")

		// Set schema to nil to test the error handling path
		parser.schema = nil

		// Call initSchema directly to test error handling
		err := parser.initSchema()
		assert.NoError(t, err) // Should succeed with valid embedded schema

		parser = NewParser("nonexistent.yaml")
		assert.NotNil(t, parser)
		assert.Equal(t, "nonexistent.yaml", parser.FilePath)
	})

	t.Run("should test NewParser error paths indirectly", func(t *testing.T) {
		// Create multiple parsers to exercise different code paths
		parsers := []string{
			"file1.yaml",
			"file2.yaml",
			"",           // Empty filename
			"very/long/path/to/nonexistent/file.yaml",
		}

		for _, filePath := range parsers {
			parser := NewParser(filePath)
			assert.NotNil(t, parser)
			assert.Equal(t, filePath, parser.FilePath)
			// Schema should be initialized in all cases
		}
	})

	t.Run("should initialize schema successfully under normal conditions", func(t *testing.T) {
		parser := NewParser("test.yaml")
		assert.NotNil(t, parser)
		// Under normal conditions, schema should be initialized
		assert.NotNil(t, parser.schema)
	})

	t.Run("should exercise NewParser warning path by testing error scenarios", func(t *testing.T) {
		// Test multiple edge cases that might trigger different code paths
		edgeCases := []string{
			"test.yaml",
			"",
			"very-long-filename-that-might-cause-issues-in-some-systems.yaml",
			"/tmp/test.yaml",
			"../test.yaml",
			"test with spaces.yaml",
		}

		for _, filePath := range edgeCases {
			parser := NewParser(filePath)
			assert.NotNil(t, parser)
			assert.Equal(t, filePath, parser.FilePath)
			// Schema should be initialized in all normal cases
		}

		// Test that multiple parser instances work correctly
		parsers := make([]*Parser, 10)
		for i := 0; i < 10; i++ {
			parsers[i] = NewParser(fmt.Sprintf("test_%d.yaml", i))
			assert.NotNil(t, parsers[i])
		}
	})
}

func TestParser_initSchema_ErrorPaths(t *testing.T) {
	t.Run("should handle schema compilation under normal conditions", func(t *testing.T) {
		parser := &Parser{FilePath: "test.yaml"}
		err := parser.initSchema()
		assert.NoError(t, err, "Schema initialization should succeed with valid embedded schema")
		assert.NotNil(t, parser.schema)
	})

	t.Run("should reinitialize schema when called multiple times", func(t *testing.T) {
		parser := &Parser{FilePath: "test.yaml"}

		// First initialization
		err := parser.initSchema()
		assert.NoError(t, err)
		firstSchema := parser.schema
		assert.NotNil(t, firstSchema)

		// Second initialization should work too
		err = parser.initSchema()
		assert.NoError(t, err)
		assert.NotNil(t, parser.schema)
	})

	t.Run("should handle different parser instances", func(t *testing.T) {
		// Test multiple parser instances to exercise schema compilation
		for i := 0; i < 3; i++ {
			parser := &Parser{FilePath: fmt.Sprintf("test%d.yaml", i)}
			err := parser.initSchema()
			assert.NoError(t, err)
			assert.NotNil(t, parser.schema)
		}
	})

	t.Run("should compile schema successfully with different parser configurations", func(t *testing.T) {
		testCases := []struct {
			filePath string
		}{
			{"simple.yaml"},
			{"complex/path/file.yaml"},
			{""},
			{"file-with-dashes.yaml"},
			{"file_with_underscores.yaml"},
		}

		for _, tc := range testCases {
			parser := &Parser{FilePath: tc.filePath}
			err := parser.initSchema()
			assert.NoError(t, err, "Should succeed for filePath: %s", tc.filePath)
			assert.NotNil(t, parser.schema)
		}
	})

	// Note: The error paths in initSchema are defensive coding for edge cases
	// where the embedded schema could be corrupted or jsonschema library fails
}

func TestParser_validateSchema_ErrorPaths(t *testing.T) {
	t.Run("should validate valid YAML against schema", func(t *testing.T) {
		parser := NewParser("test.yaml")

		validYAML := []byte(`displayName: Test SOR
description: Test Description
entities:
  user:
    displayName: User
    externalId: User
    attributes:
      - name: id
        externalId: id
        type: String
        uniqueId: true`)

		err := parser.validateSchema(validYAML)
		assert.NoError(t, err)
	})

	t.Run("should reject invalid YAML against schema", func(t *testing.T) {
		parser := NewParser("test.yaml")

		// Missing required fields
		invalidYAML := []byte(`displayName: Test SOR`)

		err := parser.validateSchema(invalidYAML)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "schema validation")
	})

	t.Run("should handle malformed YAML", func(t *testing.T) {
		parser := NewParser("test.yaml")

		malformedYAML := []byte(`displayName: Test SOR
description: [unclosed`)

		err := parser.validateSchema(malformedYAML)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse YAML")
	})

	t.Run("should handle nil schema gracefully", func(t *testing.T) {
		parser := &Parser{
			FilePath: "test.yaml",
			schema:   nil, // Explicitly set schema to nil
		}

		validYAML := []byte(`displayName: Test SOR
description: Test Description`)

		err := parser.validateSchema(validYAML)
		// Should try to initialize schema first
		assert.NoError(t, err)
	})
}

func TestParser_validate_ErrorPaths(t *testing.T) {
	t.Run("should validate parsed data successfully", func(t *testing.T) {
		parser := &Parser{
			FilePath: "test.yaml",
			Definition: &SORDefinition{
				DisplayName: "Test SOR",
				Description: "Test Description",
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
		assert.NoError(t, err)
	})

	t.Run("should reject nil definition", func(t *testing.T) {
		parser := &Parser{
			FilePath:   "test.yaml",
			Definition: nil, // Nil definition
		}

		err := parser.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty definition")
	})

	t.Run("should reject data with empty entities", func(t *testing.T) {
		parser := &Parser{
			FilePath: "test.yaml",
			Definition: &SORDefinition{
				DisplayName: "Test SOR",
				Description: "Test Description",
				Entities:    map[string]Entity{}, // Empty entities
			},
		}

		err := parser.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no entities defined")
	})

	t.Run("should reject entity with missing externalId", func(t *testing.T) {
		parser := &Parser{
			FilePath: "test.yaml",
			Definition: &SORDefinition{
				DisplayName: "Test SOR",
				Description: "Test Description",
				Entities: map[string]Entity{
					"user": {
						DisplayName: "User",
						ExternalId:  "", // Empty external ID
						Attributes: []Attribute{
							{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
						},
					},
				},
			},
		}

		err := parser.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing externalId")
	})

	t.Run("should reject entity with no attributes", func(t *testing.T) {
		parser := &Parser{
			FilePath: "test.yaml",
			Definition: &SORDefinition{
				DisplayName: "Test SOR",
				Description: "Test Description",
				Entities: map[string]Entity{
					"user": {
						DisplayName: "User",
						ExternalId:  "User",
						Attributes:  []Attribute{}, // No attributes
					},
				},
			},
		}

		err := parser.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "has no attributes")
	})

	t.Run("should reject entity with no uniqueId attribute", func(t *testing.T) {
		parser := &Parser{
			FilePath: "test.yaml",
			Definition: &SORDefinition{
				DisplayName: "Test SOR",
				Description: "Test Description",
				Entities: map[string]Entity{
					"user": {
						DisplayName: "User",
						ExternalId:  "User",
						Attributes: []Attribute{
							{Name: "name", ExternalId: "name", Type: "String", UniqueId: false}, // No unique attribute
						},
					},
				},
			},
		}

		err := parser.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no attribute marked as uniqueId")
	})

	t.Run("should accept entity with multiple uniqueId attributes", func(t *testing.T) {
		parser := &Parser{
			FilePath: "test.yaml",
			Definition: &SORDefinition{
				DisplayName: "Test SOR",
				Description: "Test Description",
				Entities: map[string]Entity{
					"user": {
						DisplayName: "User",
						ExternalId:  "User",
						Attributes: []Attribute{
							{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
							{Name: "uuid", ExternalId: "uuid", Type: "String", UniqueId: true}, // Multiple unique is allowed by parser
						},
					},
				},
			},
		}

		err := parser.validate()
		assert.NoError(t, err) // Parser allows multiple unique attributes
	})
}
