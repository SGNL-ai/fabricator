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
						{Name: "name", ExternalId: "name", Type: "String"},
						{Name: "profile_id", ExternalId: "profile_id", Type: "String"}, // FK field
					},
				},
				"profile": {
					DisplayName: "Profile",
					ExternalId:  "Profile",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
						{Name: "bio", ExternalId: "bio", Type: "String"},
					},
				},
			},
			Relationships: map[string]parser.Relationship{
				"user_profile": {
					DisplayName:   "User Profile",
					Name:          "user_profile",
					FromAttribute: "user.profile_id",
					ToAttribute:   "profile.id",
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

		// CRITICAL TEST: Foreign key relationships should work in orchestrator too
		if result.ValidationSummary != nil {
			assert.Empty(t, result.ValidationSummary.Errors,
				"Should have no FK validation errors - found: %v", result.ValidationSummary.Errors)
		}
	})

	t.Run("should work with attributeAlias relationships", func(t *testing.T) {
		def := &parser.SORDefinition{
			DisplayName: "AttributeAlias Test SOR",
			Description: "Test attributeAlias relationship format",
			Entities: map[string]parser.Entity{
				"user-entity": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "userId", ExternalId: "uuid", Type: "String", UniqueId: true, AttributeAlias: "user-primary-key-alias"},
						{Name: "name", ExternalId: "name", Type: "String", AttributeAlias: "user-name-alias"},
					},
				},
				"profile-entity": {
					DisplayName: "Profile",
					ExternalId:  "Profile",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true, AttributeAlias: "profile-primary-key-alias"},
						{Name: "userId", ExternalId: "uuid", Type: "String", AttributeAlias: "profile-user-id-alias"},
					},
				},
			},
			Relationships: map[string]parser.Relationship{
				"user-to-profile": {
					DisplayName:   "USER_TO_PROFILE",
					Name:          "user_to_profile",
					FromAttribute: "user-primary-key-alias",
					ToAttribute:   "profile-user-id-alias",
				},
			},
		}

		tempDir, err := os.MkdirTemp("", "attr_alias_test_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		options := GenerationOptions{
			DataVolume:      2,
			ValidateResults: true,
		}

		result, err := RunGeneration(def, tempDir, options)

		assert.NoError(t, err, "Generation should succeed with attributeAlias")
		assert.NotNil(t, result.ValidationSummary, "Should include validation summary")

		// CRITICAL TEST: AttributeAlias relationships should work
		if result.ValidationSummary != nil {
			assert.Empty(t, result.ValidationSummary.Errors,
				"AttributeAlias FK relationships should be valid - found: %v", result.ValidationSummary.Errors)
		}
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
