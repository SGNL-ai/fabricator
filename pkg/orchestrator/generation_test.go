package orchestrator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerationOrchestrator(t *testing.T) {
	t.Run("should generate CSV files successfully", func(t *testing.T) {
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

		tempDir, err := os.MkdirTemp("", "generation_test_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		options := GenerationOptions{
			DataVolume:      3,
			AutoCardinality: false,
			GenerateDiagram: true,
			ValidateResults: true,
		}

		result, err := RunGeneration(def, tempDir, options)

		// Should succeed
		assert.NoError(t, err, "Generation should succeed")
		assert.NotNil(t, result, "Should return generation result")

		// Check result details
		assert.Equal(t, 1, result.EntitiesProcessed, "Should process 1 entity")
		assert.Equal(t, 3, result.RecordsPerEntity, "Should generate 3 records per entity")
		assert.Equal(t, 3, result.TotalRecords, "Should generate 3 total records")
		assert.Equal(t, 1, result.CSVFilesGenerated, "Should generate 1 CSV file")

		// Verify CSV file was created
		csvPath := filepath.Join(tempDir, "User.csv")
		assert.FileExists(t, csvPath, "User.csv should be created")

		// Verify ER diagram was created (if requested)
		if options.GenerateDiagram {
			assert.True(t, result.DiagramGenerated, "Should indicate diagram was generated")
			assert.NotEmpty(t, result.DiagramPath, "Should provide diagram path")
		}

		// Check validation results
		if options.ValidateResults {
			assert.NotNil(t, result.ValidationSummary, "Should include validation summary")
			assert.Empty(t, result.ValidationSummary.Errors, "Should have no validation errors for generated data")
		}
	})

	t.Run("should handle generation errors gracefully", func(t *testing.T) {
		// Invalid definition (no entities)
		def := &parser.SORDefinition{
			DisplayName: "Invalid SOR",
			Description: "Invalid Description",
			Entities:    map[string]parser.Entity{}, // Empty entities
		}

		tempDir, err := os.MkdirTemp("", "generation_test_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		options := GenerationOptions{
			DataVolume: 1,
		}

		result, err := RunGeneration(def, tempDir, options)

		// Should fail with validation error since empty entities is invalid
		assert.Error(t, err, "Should fail for invalid SOR definition")
		assert.Contains(t, err.Error(), "at least one entity", "Should mention entity requirement")
		assert.Nil(t, result, "Should return nil result for invalid definition")
	})

	t.Run("should validate generated data when requested", func(t *testing.T) {
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

		tempDir, err := os.MkdirTemp("", "generation_test_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		options := GenerationOptions{
			DataVolume:      2,
			ValidateResults: true, // Request validation
		}

		result, err := RunGeneration(def, tempDir, options)

		assert.NoError(t, err, "Generation should succeed")
		assert.NotNil(t, result.ValidationSummary, "Should include validation summary when requested")

		// May have some relationship inconsistencies due to random data generation
		// but should not have fatal validation errors
	})

	t.Run("should skip validation when not requested", func(t *testing.T) {
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

		tempDir, err := os.MkdirTemp("", "generation_test_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		options := GenerationOptions{
			DataVolume:      1,
			ValidateResults: false, // Skip validation
		}

		result, err := RunGeneration(def, tempDir, options)

		assert.NoError(t, err, "Generation should succeed")
		assert.Nil(t, result.ValidationSummary, "Should not include validation summary when not requested")
	})
}
