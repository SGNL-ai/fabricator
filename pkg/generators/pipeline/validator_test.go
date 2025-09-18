package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test validation-only mode behavior - loading existing CSV files and validating them
func TestValidationProcessor(t *testing.T) {
	t.Run("should load valid CSV files and report no errors", func(t *testing.T) {
		// Create test definition
		def := &parser.SORDefinition{
			DisplayName: "Test SOR",
			Description: "Test Description",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
						{Name: "name", ExternalId: "name", Type: "String"},
						{Name: "email", ExternalId: "email", Type: "String"},
					},
				},
				"role": {
					DisplayName: "Role",
					ExternalId:  "Role",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
						{Name: "name", ExternalId: "name", Type: "String"},
					},
				},
			},
		}

		// Create test directory with valid CSV files
		tempDir, err := os.MkdirTemp("", "validation_test_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Create valid User.csv
		userCSV := `id,name,email
user-1,John Doe,john@example.com
user-2,Jane Smith,jane@example.com`
		err = os.WriteFile(filepath.Join(tempDir, "User.csv"), []byte(userCSV), 0644)
		require.NoError(t, err)

		// Create valid Role.csv
		roleCSV := `id,name
role-1,Admin
role-2,User`
		err = os.WriteFile(filepath.Join(tempDir, "Role.csv"), []byte(roleCSV), 0644)
		require.NoError(t, err)

		// Test validation-only mode
		processor := NewValidationProcessor()
		errors, err := processor.ValidateExistingCSVFiles(def, tempDir)

		// Should succeed with no validation errors
		assert.NoError(t, err, "Should successfully load and validate CSV files")
		assert.Empty(t, errors, "Should have no validation errors for valid data")
	})

	t.Run("should detect duplicate unique values in CSV files", func(t *testing.T) {
		def := &parser.SORDefinition{
			DisplayName: "Test SOR",
			Description: "Test Description",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
						{Name: "name", ExternalId: "name", Type: "String"},
					},
				},
			},
		}

		tempDir, err := os.MkdirTemp("", "validation_test_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Create CSV with duplicate unique IDs
		userCSV := `id,name
user-1,John Doe
user-1,Jane Smith` // Duplicate ID
		err = os.WriteFile(filepath.Join(tempDir, "User.csv"), []byte(userCSV), 0644)
		require.NoError(t, err)

		processor := NewValidationProcessor()
		errors, err := processor.ValidateExistingCSVFiles(def, tempDir)

		// Should collect the duplicate unique value error, not fail fatally
		assert.NoError(t, err, "Should not fail fatally - should collect errors")
		assert.NotEmpty(t, errors, "Should collect duplicate unique value error")
		assert.Contains(t, errors[0], "duplicate", "Error should mention duplicate value")
	})

	t.Run("should detect invalid foreign key references", func(t *testing.T) {
		def := &parser.SORDefinition{
			DisplayName: "Test SOR",
			Description: "Test Description",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
						{Name: "roleId", ExternalId: "roleId", Type: "String"},
					},
				},
				"role": {
					DisplayName: "Role",
					ExternalId:  "Role",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
					},
				},
			},
			Relationships: map[string]parser.Relationship{
				"user_role": {
					DisplayName:   "User Role",
					Name:          "user_role",
					FromAttribute: "user.roleId",
					ToAttribute:   "role.id",
				},
			},
		}

		tempDir, err := os.MkdirTemp("", "validation_test_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Create CSV with invalid FK references
		userCSV := `id,roleId
user-1,role-999
user-2,role-888` // Invalid role IDs
		err = os.WriteFile(filepath.Join(tempDir, "User.csv"), []byte(userCSV), 0644)
		require.NoError(t, err)

		roleCSV := `id
role-1
role-2` // Valid roles, but don't match the FK values
		err = os.WriteFile(filepath.Join(tempDir, "Role.csv"), []byte(roleCSV), 0644)
		require.NoError(t, err)

		processor := NewValidationProcessor()
		errors, err := processor.ValidateExistingCSVFiles(def, tempDir)

		// Should load successfully but report FK validation errors
		assert.NoError(t, err, "Should load CSV files successfully")
		assert.NotEmpty(t, errors, "Should detect invalid FK references")
		assert.GreaterOrEqual(t, len(errors), 2, "Should detect at least 2 invalid FK references")

		// Check error content
		for _, errMsg := range errors {
			assert.Contains(t, errMsg, "does not exist", "Should mention FK doesn't exist")
		}
	})

	t.Run("should handle missing CSV files gracefully", func(t *testing.T) {
		def := &parser.SORDefinition{
			DisplayName: "Test SOR",
			Description: "Test Description",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
					},
				},
			},
		}

		// Empty directory - no CSV files
		tempDir, err := os.MkdirTemp("", "validation_test_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		processor := NewValidationProcessor()
		errors, err := processor.ValidateExistingCSVFiles(def, tempDir)

		// Should collect missing file errors, not fail fatally
		assert.NoError(t, err, "Should not fail fatally - should collect errors")
		assert.NotEmpty(t, errors, "Should collect missing file errors")
		assert.Contains(t, errors[0], "not found", "Should mention missing CSV file")
	})

	t.Run("should handle malformed CSV files", func(t *testing.T) {
		def := &parser.SORDefinition{
			DisplayName: "Test SOR",
			Description: "Test Description",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
						{Name: "name", ExternalId: "name", Type: "String"},
					},
				},
			},
		}

		tempDir, err := os.MkdirTemp("", "validation_test_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Create malformed CSV (wrong number of columns)
		malformedCSV := `id,name
user-1,John Doe
user-2,Jane,Smith,Extra` // Too many columns in row 2
		err = os.WriteFile(filepath.Join(tempDir, "User.csv"), []byte(malformedCSV), 0644)
		require.NoError(t, err)

		processor := NewValidationProcessor()
		errors, err := processor.ValidateExistingCSVFiles(def, tempDir)

		// Should collect malformed CSV errors, not fail fatally
		assert.NoError(t, err, "Should not fail fatally - should collect errors")
		assert.NotEmpty(t, errors, "Should collect malformed CSV errors")
		assert.Contains(t, errors[0], "wrong number of fields", "Error should mention field count mismatch")
	})
}
