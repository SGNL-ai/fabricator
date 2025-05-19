package generators

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnsureUniqueValue(t *testing.T) {
	// Create test generator
	generator := NewCSVGenerator("test-output", 10, false)

	// Test case 1: UUID-based fields
	t.Run("UUIDFields", func(t *testing.T) {
		// Test for both uuid and id fields
		uuidFields := []string{"uuid", "guid", "user_uuid", "entity_id"}
		for _, field := range uuidFields {
			entityID := "test-entity"
			baseValue := "test-value"

			// First call to generate a value
			result1 := generator.ensureUniqueValue(entityID, field, baseValue)

			// Verify it's a UUID format string (36 chars, contains hyphens)
			assert.Len(t, result1, 36, "Expected UUID format for %s with length 36", field)
			assert.Contains(t, result1, "-", "Expected UUID format for %s to contain hyphens", field)

			// Check it was marked as used
			attrKey := entityID + ":" + field
			assert.True(t, generator.usedUniqueValues[attrKey][result1],
				"Value %s should be marked as used in the map for %s", result1, field)

			// Second call should generate a different UUID
			result2 := generator.ensureUniqueValue(entityID, field, baseValue)
			assert.NotEqual(t, result1, result2, "Expected different UUIDs on successive calls for %s", field)
		}
	})

	// Test case 2: Non-UUID fields that don't have duplicates
	t.Run("NonDuplicateFields", func(t *testing.T) {
		entityID := "test-entity"
		attrName := "unique_name"
		baseValue := "test-value"

		// First call should return the base value since it's not used yet
		result := generator.ensureUniqueValue(entityID, attrName, baseValue)

		assert.Equal(t, baseValue, result, "Expected first call to return the base value")

		// Verify the value was marked as used
		attrKey := entityID + ":" + attrName
		assert.True(t, generator.usedUniqueValues[attrKey][baseValue],
			"Value should be marked as used in the map")
	})

	// Test case 3: Handling duplicates by adding suffixes
	t.Run("HandleDuplicates", func(t *testing.T) {
		entityID := "test-entity"
		attrName := "unique_field"
		baseValue := "duplicate-value"

		// Pre-populate the used values map
		attrKey := entityID + ":" + attrName
		if generator.usedUniqueValues[attrKey] == nil {
			generator.usedUniqueValues[attrKey] = make(map[string]bool)
		}
		generator.usedUniqueValues[attrKey][baseValue] = true

		// Generate a value - should get a suffix
		result := generator.ensureUniqueValue(entityID, attrName, baseValue)

		assert.NotEqual(t, baseValue, result, "Expected a modified value to avoid duplicates")
		assert.True(t, strings.HasPrefix(result, baseValue+"_"),
			"Expected value with suffix, got: %s", result)

		// Check that the new value is marked as used
		assert.True(t, generator.usedUniqueValues[attrKey][result],
			"New value %s should be marked as used", result)
	})

	// Test case 4: Handle values with existing numeric suffixes
	t.Run("HandleExistingNumericSuffixes", func(t *testing.T) {
		entityID := "test-entity"
		attrName := "unique_field"
		baseValue := "value_123"

		// Pre-populate the used values map
		attrKey := entityID + ":" + attrName
		if generator.usedUniqueValues[attrKey] == nil {
			generator.usedUniqueValues[attrKey] = make(map[string]bool)
		}
		generator.usedUniqueValues[attrKey][baseValue] = true

		// Generate a value - should replace the numeric suffix
		result := generator.ensureUniqueValue(entityID, attrName, baseValue)

		assert.NotEqual(t, baseValue, result, "Expected a modified value to avoid duplicates")
		assert.True(t, strings.HasPrefix(result, "value_"),
			"Expected value with replaced suffix, got: %s", result)

		// Check that the new value is marked as used
		assert.True(t, generator.usedUniqueValues[attrKey][result],
			"New value %s should be marked as used", result)
	})

	// Test case 5: Handle multiple levels of duplicates
	t.Run("HandleMultipleDuplicateLevels", func(t *testing.T) {
		entityID := "test-entity"
		attrName := "multi_dupe_field"
		baseValue := "base-value"

		// Pre-populate the used values map with the base value and first attempt
		attrKey := entityID + ":" + attrName
		if generator.usedUniqueValues[attrKey] == nil {
			generator.usedUniqueValues[attrKey] = make(map[string]bool)
		}
		generator.usedUniqueValues[attrKey][baseValue] = true
		generator.usedUniqueValues[attrKey][baseValue+"_0"] = true

		// Generate a value - should try the next suffix
		result := generator.ensureUniqueValue(entityID, attrName, baseValue)

		assert.NotEqual(t, baseValue, result, "Expected a different value to avoid duplicates")
		assert.NotEqual(t, baseValue+"_0", result, "Expected a different suffix to avoid duplicates")

		// Check that the new value is marked as used
		assert.True(t, generator.usedUniqueValues[attrKey][result],
			"New value %s should be marked as used", result)
	})

	// Test case 6: Test initializing the used values map
	t.Run("InitializeUsedValuesMap", func(t *testing.T) {
		// Create a fresh generator
		freshGen := NewCSVGenerator("test-output", 10, false)

		// Use an entity and attribute that doesn't have a map yet
		entityID := "new-entity"
		attrName := "new-field"
		baseValue := "test-value"

		// This should initialize the map
		result := freshGen.ensureUniqueValue(entityID, attrName, baseValue)

		// Check that the result is as expected
		assert.Equal(t, baseValue, result, "Expected base value to be returned")

		// Check that the maps were initialized
		attrKey := entityID + ":" + attrName
		assert.NotNil(t, freshGen.usedUniqueValues[attrKey],
			"Used values map should be initialized for new entity/attribute")

		// Check that the value was marked as used
		assert.True(t, freshGen.usedUniqueValues[attrKey][baseValue],
			"Value should be marked as used in the newly initialized map")
	})
}
