package model

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestRelationshipTargetSelection tests the new methods for relationship-based FK selection
func TestRelationshipTargetSelection(t *testing.T) {
	testCases := []struct {
		name            string
		cardinality     string
		autoCardinality bool
		sourceRows      int
		targetRows      int
		expectedPattern string
		testFunc        func(t *testing.T, selectedValues []string, targetRows int)
	}{
		// Test edge cases with different row counts
		{
			name:            "N:1 round-robin with 1 source, 1 target",
			cardinality:     ManyToOne,
			autoCardinality: false,
			sourceRows:      1,
			targetRows:      1,
			expectedPattern: "single assignment",
			testFunc: func(t *testing.T, selectedValues []string, targetRows int) {
				assert.Equal(t, []string{"target-0"}, selectedValues, "Single row should get target-0")
			},
		},
		{
			name:            "N:1 round-robin with 1 source, 3 targets",
			cardinality:     ManyToOne,
			autoCardinality: false,
			sourceRows:      1,
			targetRows:      3,
			expectedPattern: "first target selection",
			testFunc: func(t *testing.T, selectedValues []string, targetRows int) {
				assert.Equal(t, []string{"target-0"}, selectedValues, "First source row should get target-0")
			},
		},
		{
			name:            "N:1 power-law with 1 source, 3 targets",
			cardinality:     ManyToOne,
			autoCardinality: true,
			sourceRows:      1,
			targetRows:      3,
			expectedPattern: "power law single selection",
			testFunc: func(t *testing.T, selectedValues []string, targetRows int) {
				assert.Len(t, selectedValues, 1, "Should have 1 selection")
				assert.NotEmpty(t, selectedValues[0], "Should select some target")
				// Don't assert specific target since power law can vary
			},
		},
		{
			name:            "1:1 with autoCardinality=true",
			cardinality:     OneToOne,
			autoCardinality: true,
			sourceRows:      5,
			targetRows:      5,
			expectedPattern: "unique assignment",
			testFunc: func(t *testing.T, selectedValues []string, targetRows int) {
				// Should have unique values (no duplicates)
				uniqueValues := make(map[string]bool)
				for _, value := range selectedValues {
					assert.False(t, uniqueValues[value], "1:1 should not reuse target values")
					uniqueValues[value] = true
				}
				assert.Len(t, uniqueValues, len(selectedValues), "All values should be unique in 1:1")
			},
		},
		{
			name:            "1:1 with autoCardinality=false",
			cardinality:     OneToOne,
			autoCardinality: false,
			sourceRows:      4,
			targetRows:      4,
			expectedPattern: "round-robin unique",
			testFunc: func(t *testing.T, selectedValues []string, targetRows int) {
				// Should be round-robin: target-0, target-1, target-2, target-3
				expected := []string{"target-0", "target-1", "target-2", "target-3"}
				assert.Equal(t, expected, selectedValues, "1:1 round-robin should cycle uniquely")
			},
		},
		{
			name:            "N:1 with autoCardinality=true",
			cardinality:     ManyToOne,
			autoCardinality: true,
			sourceRows:      10,
			targetRows:      3,
			expectedPattern: "power law clustering",
			testFunc: func(t *testing.T, selectedValues []string, targetRows int) {
				// Should show clustering (some targets more popular)
				targetCounts := make(map[string]int)
				for _, value := range selectedValues {
					targetCounts[value]++
				}
				// Should have variation in target usage
				counts := make([]int, 0)
				for _, count := range targetCounts {
					counts = append(counts, count)
				}
				hasVariation := false
				if len(counts) > 1 {
					firstCount := counts[0]
					for _, count := range counts {
						if count != firstCount {
							hasVariation = true
							break
						}
					}
				}
				assert.True(t, hasVariation, "N:1 auto-cardinality should create clustering variation")
			},
		},
		{
			name:            "N:1 with autoCardinality=false",
			cardinality:     ManyToOne,
			autoCardinality: false,
			sourceRows:      6,
			targetRows:      3,
			expectedPattern: "round-robin cycling",
			testFunc: func(t *testing.T, selectedValues []string, targetRows int) {
				// Should be predictable round-robin: target-0, target-1, target-2, target-0, target-1, target-2
				expected := []string{"target-0", "target-1", "target-2", "target-0", "target-1", "target-2"}
				assert.Equal(t, expected, selectedValues, "N:1 round-robin should cycle predictably")
			},
		},
		{
			name:            "1:N with autoCardinality=true",
			cardinality:     OneToMany,
			autoCardinality: true,
			sourceRows:      8,
			targetRows:      5,
			expectedPattern: "power law clustering",
			testFunc: func(t *testing.T, selectedValues []string, targetRows int) {
				// Should show clustering behavior
				targetCounts := make(map[string]int)
				for _, value := range selectedValues {
					targetCounts[value]++
				}
				assert.Greater(t, len(targetCounts), 0, "Should select some targets")
				t.Logf("1:N auto-cardinality distribution: %v", targetCounts)
			},
		},
		{
			name:            "1:N with autoCardinality=false",
			cardinality:     OneToMany,
			autoCardinality: false,
			sourceRows:      6,
			targetRows:      4,
			expectedPattern: "round-robin cycling",
			testFunc: func(t *testing.T, selectedValues []string, targetRows int) {
				// Should be predictable round-robin
				expected := []string{"target-0", "target-1", "target-2", "target-3", "target-0", "target-1"}
				assert.Equal(t, expected, selectedValues, "1:N round-robin should cycle predictably")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mock source entity
			mockSource := NewMockEntityInterface(ctrl)
			mockSource.EXPECT().GetRowCount().Return(tc.sourceRows).AnyTimes()

			// Create mock target entity with controlled responses
			mockTarget := NewMockEntityInterface(ctrl)
			mockTarget.EXPECT().GetRowCount().Return(tc.targetRows).AnyTimes()
			mockTarget.EXPECT().GetRowByIndex(gomock.Any()).DoAndReturn(func(index int) *Row {
				if index >= 0 && index < tc.targetRows {
					return NewRow(map[string]string{"id": "target-" + string(rune('0'+index))})
				}
				return nil
			}).AnyTimes()

			// Create mock target attribute
			mockTargetAttr := NewMockAttributeInterface(ctrl)
			mockTargetAttr.EXPECT().GetName().Return("id").AnyTimes()

			// Create relationship with mocks
			relationship := &Relationship{
				id:           tc.name + "_rel",
				cardinality:  tc.cardinality,
				sourceEntity: mockSource,
				targetEntity: mockTarget,
				targetAttr:   mockTargetAttr,
			}

			// Test target selection
			selectedValues := make([]string, tc.sourceRows)
			for i := 0; i < tc.sourceRows; i++ {
				value, err := relationship.GetTargetValueForSourceRow(i, tc.autoCardinality)
				require.NoError(t, err, "Should successfully get target value for row %d", i)
				selectedValues[i] = value
			}

			t.Logf("%s results: %v", tc.expectedPattern, selectedValues)

			// Run cardinality-specific validation
			tc.testFunc(t, selectedValues, tc.targetRows)
		})
	}

	t.Run("should work with real entities not just mocks", func(t *testing.T) {
		// Test using real entity setup to expose key mismatch issues
		def := &parser.SORDefinition{
			DisplayName: "Real Entity Test",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
						{Name: "roleId", ExternalId: "roleId", Type: "String", UniqueId: false},
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
					Name:          "user_role",
					FromAttribute: "User.roleId",
					ToAttribute:   "Role.id",
				},
			},
		}

		graph, err := NewGraph(def, 5)
		require.NoError(t, err)

		// Set up real entities with real rows
		entities := graph.GetAllEntities()
		userEntity := entities["User"]
		roleEntity := entities["Role"]

		// Add 1 user (source)
		err = userEntity.AddRow(NewRow(map[string]string{
			"id":     "user-1",
			"roleId": "", // Empty FK
		}))
		require.NoError(t, err)

		// Add 1 role (target)
		err = roleEntity.AddRow(NewRow(map[string]string{
			"id": "role-1",
		}))
		require.NoError(t, err)

		// Get the relationship
		relationships := graph.GetAllRelationships()
		require.Len(t, relationships, 1)
		relationship := relationships[0]

		// Test with real entity setup - this should expose the key mismatch bug
		targetValue, err := relationship.GetTargetValueForSourceRow(0, false)
		require.NoError(t, err, "Should successfully get target value with real entities")
		assert.Equal(t, "role-1", targetValue, "Should return role-1 from real entity row")
	})

	t.Run("should validate mock parameter usage", func(t *testing.T) {
		// Test that verifies what parameters the relationship is actually using
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock source entity that validates GetRowCount is called
		mockSource := NewMockEntityInterface(ctrl)
		mockSource.EXPECT().GetRowCount().Return(5).Times(1) // Should be called exactly once

		// Create mock target entity that validates specific GetRowByIndex calls
		mockTarget := NewMockEntityInterface(ctrl)
		mockTarget.EXPECT().GetRowCount().Return(3).Times(1)
		// Expect specific index call with round-robin: sourceIndex 0 → targetIndex 0
		mockTarget.EXPECT().GetRowByIndex(0).Return(NewRow(map[string]string{"id": "target-0"})).Times(1)

		// Create mock target attribute that validates GetName is called
		mockTargetAttr := NewMockAttributeInterface(ctrl)
		mockTargetAttr.EXPECT().GetName().Return("id").Times(1) // Should be called to get attribute name

		// Create relationship
		relationship := &Relationship{
			id:           "validation_test",
			cardinality:  ManyToOne,
			sourceEntity: mockSource,
			targetEntity: mockTarget,
			targetAttr:   mockTargetAttr,
		}

		// Call with round-robin (autoCardinality=false)
		// This should call: GetRowCount() on both entities, GetRowByIndex(0) on target, GetName() on attribute
		targetValue, err := relationship.GetTargetValueForSourceRow(0, false)
		require.NoError(t, err)
		assert.Equal(t, "target-0", targetValue)

		// All mock expectations should be satisfied automatically by gomock
	})

}
