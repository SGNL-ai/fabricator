package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T009: Test LoadConfiguration with valid YAML
func TestLoadConfiguration_ValidYAML(t *testing.T) {
	// Create a temporary valid config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "valid_config.yaml")

	validYAML := `users: 1000
groups: 50
permissions: 200
`
	err := os.WriteFile(configPath, []byte(validYAML), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Load the configuration
	config, err := LoadConfiguration(configPath)

	// Assertions
	require.NoError(t, err, "LoadConfiguration should not return an error for valid YAML")
	assert.NotNil(t, config, "Configuration should not be nil")
	assert.Equal(t, 1000, config.EntityCounts["users"], "Users count should be 1000")
	assert.Equal(t, 50, config.EntityCounts["groups"], "Groups count should be 50")
	assert.Equal(t, 200, config.EntityCounts["permissions"], "Permissions count should be 200")
	assert.Equal(t, configPath, config.SourceFile, "SourceFile should match the config path")
	assert.False(t, config.LoadedAt.IsZero(), "LoadedAt timestamp should be set")
}

// T010: Test LoadConfiguration with invalid YAML
func TestLoadConfiguration_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid_config.yaml")

	invalidYAML := `users: 1000
groups: [this is not valid
permissions: 200
`
	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Load the configuration
	config, err := LoadConfiguration(configPath)

	// Assertions
	assert.Error(t, err, "LoadConfiguration should return an error for invalid YAML")
	assert.Nil(t, config, "Configuration should be nil on error")

	// Check that it's a ValidationError
	var valErr *ValidationError
	assert.ErrorAs(t, err, &valErr, "Error should be a ValidationError")
	if valErr != nil {
		assert.Contains(t, valErr.Message, "Invalid YAML syntax", "Error message should mention invalid YAML")
		assert.Contains(t, valErr.Suggestion, "YAML", "Suggestion should mention YAML validation")
	}
}

// T011: Test LoadConfiguration with non-existent file
func TestLoadConfiguration_NonExistentFile(t *testing.T) {
	nonExistentPath := "/tmp/does_not_exist_12345.yaml"

	// Load the configuration
	config, err := LoadConfiguration(nonExistentPath)

	// Assertions
	assert.Error(t, err, "LoadConfiguration should return an error for non-existent file")
	assert.Nil(t, config, "Configuration should be nil on error")

	// Check that it's a ValidationError
	var valErr *ValidationError
	assert.ErrorAs(t, err, &valErr, "Error should be a ValidationError")
	if valErr != nil {
		assert.Contains(t, valErr.Message, "not found", "Error message should mention file not found")
		assert.Contains(t, valErr.Suggestion, "init-count-config", "Suggestion should mention template generation")
	}
}

// T012: Test GetCount with existing entity
func TestGetCount_ExistingEntity(t *testing.T) {
	config := &CountConfiguration{
		EntityCounts: map[string]int{
			"users":       1000,
			"groups":      50,
			"permissions": 200,
		},
	}

	count := config.GetCount("users", 100)
	assert.Equal(t, 1000, count, "GetCount should return the configured count for existing entity")

	count = config.GetCount("groups", 100)
	assert.Equal(t, 50, count, "GetCount should return the configured count for groups")
}

// T013: Test GetCount with missing entity (should return default)
func TestGetCount_MissingEntity(t *testing.T) {
	config := &CountConfiguration{
		EntityCounts: map[string]int{
			"users": 1000,
		},
	}

	count := config.GetCount("nonexistent", 100)
	assert.Equal(t, 100, count, "GetCount should return default count for missing entity")

	count = config.GetCount("another_missing", 500)
	assert.Equal(t, 500, count, "GetCount should return the provided default count")
}

// T014: Test GetCount with zero value in map (should return default)
func TestGetCount_ZeroValue(t *testing.T) {
	config := &CountConfiguration{
		EntityCounts: map[string]int{
			"users":  1000,
			"groups": 0, // Explicitly set to zero
		},
	}

	count := config.GetCount("groups", 100)
	assert.Equal(t, 100, count, "GetCount should return default count when map has zero value")
}

// T015: Test Validate against matching entities (success)
func TestValidate_MatchingEntities(t *testing.T) {
	config := &CountConfiguration{
		EntityCounts: map[string]int{
			"users":       1000,
			"groups":      50,
			"permissions": 200,
		},
	}

	sorEntities := []string{"users", "groups", "permissions", "roles"}

	err := config.Validate(sorEntities)
	assert.NoError(t, err, "Validate should not return error when all entities exist in SOR")
}

// T016: Test Validate against mismatched entities (error)
func TestValidate_MismatchedEntities(t *testing.T) {
	config := &CountConfiguration{
		EntityCounts: map[string]int{
			"users":       1000,
			"nonexistent": 50, // This entity doesn't exist in SOR
		},
	}

	sorEntities := []string{"users", "groups", "permissions"}

	err := config.Validate(sorEntities)
	assert.Error(t, err, "Validate should return error for non-existent entity")

	var valErr *ValidationError
	assert.ErrorAs(t, err, &valErr, "Error should be a ValidationError")
	if valErr != nil {
		assert.Equal(t, "nonexistent", valErr.EntityID, "EntityID should be the non-existent entity")
		assert.Contains(t, valErr.Message, "not found in SOR", "Error should mention entity not found in SOR")
		assert.Contains(t, valErr.Suggestion, "Remove 'nonexistent'", "Suggestion should mention removing the entity")
	}
}

// T017: Test Validate with negative count values
func TestValidate_NegativeCount(t *testing.T) {
	config := &CountConfiguration{
		EntityCounts: map[string]int{
			"users":  -5,
			"groups": 50,
		},
	}

	sorEntities := []string{"users", "groups"}

	err := config.Validate(sorEntities)
	assert.Error(t, err, "Validate should return error for negative count")

	var valErr *ValidationError
	assert.ErrorAs(t, err, &valErr, "Error should be a ValidationError")
	if valErr != nil {
		assert.Equal(t, "users", valErr.EntityID, "EntityID should be the entity with negative count")
		assert.Equal(t, "count", valErr.Field, "Field should be 'count'")
		assert.Contains(t, valErr.Message, "expected positive integer", "Error should mention positive integer requirement")
	}
}

// T018: Test Validate with zero count values
func TestValidate_ZeroCount(t *testing.T) {
	config := &CountConfiguration{
		EntityCounts: map[string]int{
			"users":  0,
			"groups": 50,
		},
	}

	sorEntities := []string{"users", "groups"}

	err := config.Validate(sorEntities)
	assert.Error(t, err, "Validate should return error for zero count")

	var valErr *ValidationError
	assert.ErrorAs(t, err, &valErr, "Error should be a ValidationError")
	if valErr != nil {
		assert.Equal(t, "users", valErr.EntityID, "EntityID should be the entity with zero count")
		assert.Contains(t, valErr.Message, "expected positive integer", "Error should mention positive integer requirement")
	}
}

// T019: Test Validate with non-integer values would be caught by YAML parser
// This test verifies that LoadConfiguration handles non-integer gracefully
func TestLoadConfiguration_NonIntegerValues(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "non_integer_config.yaml")

	// YAML with string value where integer expected
	nonIntegerYAML := `users: "not a number"
groups: 50
`
	err := os.WriteFile(configPath, []byte(nonIntegerYAML), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Load the configuration
	config, err := LoadConfiguration(configPath)

	// YAML parser should catch this
	assert.Error(t, err, "LoadConfiguration should return error for non-integer values")
	assert.Nil(t, config, "Configuration should be nil on error")
}

// Additional test: HasEntity method
func TestHasEntity(t *testing.T) {
	config := &CountConfiguration{
		EntityCounts: map[string]int{
			"users":       1000,
			"permissions": 200,
		},
	}

	assert.True(t, config.HasEntity("users"), "HasEntity should return true for existing entity")
	assert.True(t, config.HasEntity("permissions"), "HasEntity should return true for permissions")
	assert.False(t, config.HasEntity("groups"), "HasEntity should return false for missing entity")
	assert.False(t, config.HasEntity("nonexistent"), "HasEntity should return false for nonexistent entity")
}
