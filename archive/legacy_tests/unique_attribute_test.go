package generators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUniqueAttribute(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name            string
		entityID        string
		attrName        string
		uniqueIdMap     map[string][]string
		expectedOutcome bool
	}{
		{
			name:            "attribute_is_unique",
			entityID:        "entity1",
			attrName:        "id",
			uniqueIdMap:     map[string][]string{"entity1": {"id", "email"}},
			expectedOutcome: true,
		},
		{
			name:            "attribute_is_not_unique",
			entityID:        "entity1",
			attrName:        "name",
			uniqueIdMap:     map[string][]string{"entity1": {"id", "email"}},
			expectedOutcome: false,
		},
		{
			name:            "entity_not_in_map",
			entityID:        "entity2",
			attrName:        "id",
			uniqueIdMap:     map[string][]string{"entity1": {"id", "email"}},
			expectedOutcome: false,
		},
		{
			name:            "empty_unique_map_for_entity",
			entityID:        "entity3",
			attrName:        "id",
			uniqueIdMap:     map[string][]string{"entity3": {}},
			expectedOutcome: false,
		},
		{
			name:            "empty_unique_map",
			entityID:        "entity1",
			attrName:        "id",
			uniqueIdMap:     map[string][]string{},
			expectedOutcome: false,
		},
		{
			name:            "nil_unique_map",
			entityID:        "entity1",
			attrName:        "id",
			uniqueIdMap:     nil,
			expectedOutcome: false,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsUniqueAttribute(tc.entityID, tc.attrName, tc.uniqueIdMap)
			assert.Equal(t, tc.expectedOutcome, result,
				"IsUniqueAttribute(%s, %s, map) = %v; want %v",
				tc.entityID, tc.attrName, result, tc.expectedOutcome)
		})
	}
}
