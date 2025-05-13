package generators

import (
	"strings"
	"testing"
)

func TestGenerateUniqueValue(t *testing.T) {
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
			result1 := generator.generateUniqueValue(entityID, field, baseValue)

			// Verify it's a UUID format string (36 chars, contains hyphens)
			if len(result1) != 36 || !strings.Contains(result1, "-") {
				t.Errorf("Expected UUID format for %s, got: %s", field, result1)
			}

			// Check it was marked as used
			attrKey := entityID + ":" + field
			if !generator.usedUniqueValues[attrKey][result1] {
				t.Errorf("Value %s was not marked as used in the map for %s", result1, field)
			}

			// Second call should generate a different UUID
			result2 := generator.generateUniqueValue(entityID, field, baseValue)
			if result1 == result2 {
				t.Errorf("Expected different UUIDs on successive calls for %s", field)
			}
		}
	})

	// Test case 2: Non-UUID fields that don't have duplicates
	t.Run("NonDuplicateFields", func(t *testing.T) {
		entityID := "test-entity"
		attrName := "unique_name"
		baseValue := "test-value"

		// First call should return the base value since it's not used yet
		result := generator.generateUniqueValue(entityID, attrName, baseValue)

		if result != baseValue {
			t.Errorf("Expected first call to return the base value, got: %s", result)
		}

		// Verify the value was marked as used
		attrKey := entityID + ":" + attrName
		if !generator.usedUniqueValues[attrKey][baseValue] {
			t.Errorf("Value was not marked as used in the map")
		}
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
		result := generator.generateUniqueValue(entityID, attrName, baseValue)

		if result == baseValue {
			t.Errorf("Expected a modified value to avoid duplicates, got the same value")
		}

		if !strings.HasPrefix(result, baseValue+"_") {
			t.Errorf("Expected value with suffix, got: %s", result)
		}

		// Check that the new value is marked as used
		if !generator.usedUniqueValues[attrKey][result] {
			t.Errorf("New value %s was not marked as used", result)
		}
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
		result := generator.generateUniqueValue(entityID, attrName, baseValue)

		if result == baseValue {
			t.Errorf("Expected a modified value to avoid duplicates, got the same value")
		}

		if !strings.HasPrefix(result, "value_") {
			t.Errorf("Expected value with replaced suffix, got: %s", result)
		}

		// Check that the new value is marked as used
		if !generator.usedUniqueValues[attrKey][result] {
			t.Errorf("New value %s was not marked as used", result)
		}
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
		result := generator.generateUniqueValue(entityID, attrName, baseValue)

		if result == baseValue || result == baseValue+"_0" {
			t.Errorf("Expected a different suffix to avoid duplicates, got: %s", result)
		}

		// Check that the new value is marked as used
		if !generator.usedUniqueValues[attrKey][result] {
			t.Errorf("New value %s was not marked as used", result)
		}
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
		result := freshGen.generateUniqueValue(entityID, attrName, baseValue)

		// Check that the result is as expected
		if result != baseValue {
			t.Errorf("Expected base value to be returned, got: %s", result)
		}

		// Check that the maps were initialized
		attrKey := entityID + ":" + attrName
		if freshGen.usedUniqueValues[attrKey] == nil {
			t.Errorf("Used values map was not initialized for new entity/attribute")
		}

		// Check that the value was marked as used
		if !freshGen.usedUniqueValues[attrKey][baseValue] {
			t.Errorf("Value was not marked as used in the newly initialized map")
		}
	})
}
