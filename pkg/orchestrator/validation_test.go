package orchestrator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationOrchestrator(t *testing.T) {
	t.Run("should validate existing CSV files successfully", func(t *testing.T) {
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

		// Create valid CSV file
		csvContent := `id,name
user-1,John Doe
user-2,Jane Smith`
		err = os.WriteFile(filepath.Join(tempDir, "User.csv"), []byte(csvContent), 0644)
		require.NoError(t, err)

		options := ValidationOptions{
			GenerateDiagram: true,
		}

		result, err := RunValidation(def, tempDir, options)

		// Should succeed with no issues
		assert.NoError(t, err, "Validation should succeed")
		assert.NotNil(t, result, "Should return validation result")
		assert.Equal(t, 1, result.FilesValidated, "Should validate 1 CSV file")
		assert.Equal(t, 2, result.RecordsValidated, "Should validate 2 records")
		assert.Empty(t, result.ValidationErrors, "Should have no validation errors")

		// Check diagram generation
		if options.GenerateDiagram {
			assert.True(t, result.DiagramGenerated, "Should generate diagram when requested")
		}
	})

	t.Run("should detect and report validation issues", func(t *testing.T) {
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
					FromAttribute: "User.roleId",
					ToAttribute:   "Role.id",
				},
			},
		}

		tempDir, err := os.MkdirTemp("", "validation_test_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Create CSV files with validation issues
		userCSV := `id,roleId
user-1,invalid-role
user-1,another-invalid-role` // Duplicate ID + invalid FK
		err = os.WriteFile(filepath.Join(tempDir, "User.csv"), []byte(userCSV), 0644)
		require.NoError(t, err)

		roleCSV := `id
role-1
role-2` // Valid roles that don't match FK references
		err = os.WriteFile(filepath.Join(tempDir, "Role.csv"), []byte(roleCSV), 0644)
		require.NoError(t, err)

		options := ValidationOptions{}

		result, err := RunValidation(def, tempDir, options)

		// Should collect all validation issues, not fail
		assert.NoError(t, err, "Should not fail fatally - should collect errors")
		assert.NotNil(t, result, "Should return validation result")
		assert.NotEmpty(t, result.ValidationErrors, "Should detect validation issues")

		// Should detect both duplicate unique values and invalid FK references
		assert.GreaterOrEqual(t, len(result.ValidationErrors), 2, "Should detect multiple validation issues")

		// Check error types - debug what errors we actually get
		t.Logf("Validation errors found: %d", len(result.ValidationErrors))
		for i, errMsg := range result.ValidationErrors {
			t.Logf("Error %d: %s", i, errMsg)
		}

		hasUniqueError := false
		hasFKError := false
		for _, errMsg := range result.ValidationErrors {
			if contains(errMsg, "duplicate") {
				hasUniqueError = true
			}
			if contains(errMsg, "does not exist") {
				hasFKError = true
			}
		}
		assert.True(t, hasUniqueError, "Should detect duplicate unique value")
		assert.True(t, hasFKError, "Should detect invalid FK reference")
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

		// Empty directory
		tempDir, err := os.MkdirTemp("", "validation_test_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		options := ValidationOptions{}

		result, err := RunValidation(def, tempDir, options)

		// Should handle missing files gracefully
		assert.NoError(t, err, "Should not fail fatally for missing files")
		assert.NotNil(t, result, "Should return validation result")
		assert.Equal(t, 0, result.FilesValidated, "Should validate 0 files")
		assert.NotEmpty(t, result.ValidationErrors, "Should report missing file as validation error")
		assert.Contains(t, result.ValidationErrors[0], "not found", "Should mention missing CSV file")
	})

	t.Run("should support validation options", func(t *testing.T) {
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

		tempDir, err := os.MkdirTemp("", "validation_test_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Create valid CSV
		csvContent := `id
user-1`
		err = os.WriteFile(filepath.Join(tempDir, "User.csv"), []byte(csvContent), 0644)
		require.NoError(t, err)

		// Test with diagram generation disabled
		options := ValidationOptions{
			GenerateDiagram: false,
		}

		result, err := RunValidation(def, tempDir, options)

		assert.NoError(t, err, "Validation should succeed")
		assert.False(t, result.DiagramGenerated, "Should not generate diagram when disabled")
		assert.Empty(t, result.DiagramPath, "Should not provide diagram path when disabled")

		// Test with diagram generation enabled
		options.GenerateDiagram = true
		result, err = RunValidation(def, tempDir, options)

		assert.NoError(t, err, "Validation should succeed")
		assert.True(t, result.DiagramGenerated, "Should generate diagram when enabled")
		assert.NotEmpty(t, result.DiagramPath, "Should provide diagram path when enabled")
	})
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
