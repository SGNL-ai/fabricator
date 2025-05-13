package generators

import (
	"testing"
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
			name:            "attribute is unique",
			entityID:        "entity1",
			attrName:        "id",
			uniqueIdMap:     map[string][]string{"entity1": {"id", "email"}},
			expectedOutcome: true,
		},
		{
			name:            "attribute is not unique",
			entityID:        "entity1",
			attrName:        "name",
			uniqueIdMap:     map[string][]string{"entity1": {"id", "email"}},
			expectedOutcome: false,
		},
		{
			name:            "entity not in map",
			entityID:        "entity2",
			attrName:        "id",
			uniqueIdMap:     map[string][]string{"entity1": {"id", "email"}},
			expectedOutcome: false,
		},
		{
			name:            "empty unique map for entity",
			entityID:        "entity3",
			attrName:        "id",
			uniqueIdMap:     map[string][]string{"entity3": {}},
			expectedOutcome: false,
		},
		{
			name:            "empty unique map",
			entityID:        "entity1",
			attrName:        "id",
			uniqueIdMap:     map[string][]string{},
			expectedOutcome: false,
		},
		{
			name:            "nil unique map",
			entityID:        "entity1",
			attrName:        "id",
			uniqueIdMap:     nil,
			expectedOutcome: false,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isUniqueAttribute(tc.entityID, tc.attrName, tc.uniqueIdMap)
			if result != tc.expectedOutcome {
				t.Errorf("isUniqueAttribute(%s, %s, map) = %v; want %v",
					tc.entityID, tc.attrName, result, tc.expectedOutcome)
			}
		})
	}
}
