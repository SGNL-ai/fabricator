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

	t.Run("N:1 relationship should provide clustered target values", func(t *testing.T) {
		// Set up Users → Department (N:1) relationship
		def := &parser.SORDefinition{
			DisplayName: "N:1 Selection Test",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
						{Name: "dept_id", ExternalId: "dept_id", Type: "String", UniqueId: false}, // FK non-unique
					},
				},
				"department": {
					DisplayName: "Department",
					ExternalId:  "Department",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true}, // PK unique
					},
				},
			},
			Relationships: map[string]parser.Relationship{
				"user_dept": {
					Name:          "user_department",
					FromAttribute: "user.dept_id",  // Non-unique FK
					ToAttribute:   "department.id", // Unique PK
				},
			},
		}

		graph, err := NewGraph(def, 10)
		require.NoError(t, err)

		// Set up test data - 3 departments and 10 users
		entities := graph.GetAllEntities()
		userEntity := entities["user"]
		deptEntity := entities["department"]

		// Add 10 users (source entity needs rows)
		for i := 0; i < 10; i++ {
			err := userEntity.AddRow(NewRow(map[string]string{
				"id": "user-" + string(rune('0'+i)),
			}))
			require.NoError(t, err)
		}

		// Add 3 departments (target entity)
		for i := 0; i < 3; i++ {
			err := deptEntity.AddRow(NewRow(map[string]string{
				"id": "dept-" + string(rune('X'+i)),
			}))
			require.NoError(t, err)
		}

		// Get the relationship
		relationships := graph.GetAllRelationships()
		require.Len(t, relationships, 1)
		relationship := relationships[0]

		// Test N:1 target selection with auto-cardinality enabled
		selectedValues := make([]string, 0)
		for sourceRowIndex := 0; sourceRowIndex < 10; sourceRowIndex++ {
			targetValue, err := relationship.GetTargetValueForSourceRow(sourceRowIndex, true) // autoCardinality=true
			require.NoError(t, err, "Should be able to get target value")
			selectedValues = append(selectedValues, targetValue)
		}

		// Count department usage
		deptUsage := make(map[string]int)
		for _, deptId := range selectedValues {
			deptUsage[deptId]++
		}

		// CRITICAL: N:1 clustering tests
		// 1. All values should be valid department IDs
		validDeptIds := []string{"dept-X", "dept-Y", "dept-Z"}
		for _, value := range selectedValues {
			assert.Contains(t, validDeptIds, value, "Should return valid department ID")
		}

		// 2. With auto-cardinality, should create clustering (some departments more popular)
		// NOT uniform distribution where each department gets exactly 3-4 users
		t.Logf("Department usage with auto-cardinality: %v", deptUsage)

		// Should have variation in department usage (realistic clustering)
		usageCounts := make([]int, 0)
		for _, count := range deptUsage {
			usageCounts = append(usageCounts, count)
		}

		// With power law clustering, should see variation in department sizes
		// (This will fail with current round-robin but should pass with power law)
		hasVariation := false
		if len(usageCounts) > 1 {
			firstCount := usageCounts[0]
			for _, count := range usageCounts {
				if count != firstCount {
					hasVariation = true
					break
				}
			}
		}

		assert.True(t, hasVariation, "Auto-cardinality should create varied clustering, not uniform distribution")

		// At least one department should have multiple users (clustering)
		hasMultipleUsers := false
		for _, count := range deptUsage {
			if count > 1 {
				hasMultipleUsers = true
				break
			}
		}
		assert.True(t, hasMultipleUsers, "N:1 should create clustering with some departments having multiple users")
	})

	t.Run("round-robin should provide predictable distribution", func(t *testing.T) {
		// Test that non-auto-cardinality provides predictable round-robin
		def := &parser.SORDefinition{
			DisplayName: "Round-robin Test",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
						{Name: "role_id", ExternalId: "role_id", Type: "String", UniqueId: false},
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
					FromAttribute: "user.role_id",
					ToAttribute:   "role.id",
				},
			},
		}

		graph, err := NewGraph(def, 6)
		require.NoError(t, err)

		// Set up 6 users and 3 roles
		entities := graph.GetAllEntities()
		userEntity := entities["user"]
		roleEntity := entities["role"]

		// Add 6 users (source entity needs rows)
		for i := 0; i < 6; i++ {
			err := userEntity.AddRow(NewRow(map[string]string{
				"id": "user-" + string(rune('0'+i)),
			}))
			require.NoError(t, err)
		}

		// Add 3 roles (target entity)
		for i := 0; i < 3; i++ {
			err := roleEntity.AddRow(NewRow(map[string]string{
				"id": "role-" + string(rune('1'+i)),
			}))
			require.NoError(t, err)
		}

		// Get relationship
		relationships := graph.GetAllRelationships()
		require.Len(t, relationships, 1)
		relationship := relationships[0]

		// Test round-robin selection with auto-cardinality disabled
		selectedValues := make([]string, 0)
		for sourceRowIndex := 0; sourceRowIndex < 6; sourceRowIndex++ {
			targetValue, err := relationship.GetTargetValueForSourceRow(sourceRowIndex, false) // autoCardinality=false
			require.NoError(t, err, "Should be able to get target value")
			selectedValues = append(selectedValues, targetValue)
		}

		// CRITICAL: Round-robin should create predictable pattern
		expected := []string{"role-1", "role-2", "role-3", "role-1", "role-2", "role-3"}
		assert.Equal(t, expected, selectedValues, "Round-robin should create predictable cycling pattern")

		// Count role usage - should be perfectly uniform
		roleUsage := make(map[string]int)
		for _, roleId := range selectedValues {
			roleUsage[roleId]++
		}

		// Each role should be used exactly twice (6 users / 3 roles = 2 each)
		for roleId, count := range roleUsage {
			assert.Equal(t, 2, count, "Round-robin should distribute evenly - role '%s' should be used exactly 2 times", roleId)
		}
	})

	t.Run("selectTargetIndex should choose correct algorithm", func(t *testing.T) {
		// Create a test relationship with mocked entities for direct testing
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock source entity with 100 rows
		mockSource := NewMockEntityInterface(ctrl)
		mockSource.EXPECT().GetRowCount().Return(100).AnyTimes()

		// Create mock target entity with 5 rows
		mockTarget := NewMockEntityInterface(ctrl)
		mockTarget.EXPECT().GetRowCount().Return(5).AnyTimes()

		// Create relationship for direct testing
		relationship := &Relationship{
			id:           "test_rel",
			cardinality:  ManyToOne, // Non-unique FK → unique PK
			sourceEntity: mockSource,
			targetEntity: mockTarget,
		}

		// Test round-robin selection (autoCardinality=false)
		roundRobinResults := make([]int, 6)
		for i := 0; i < 6; i++ {
			roundRobinResults[i] = relationship.selectTargetIndex(i, 3, false)
		}
		expected := []int{0, 1, 2, 0, 1, 2} // Perfect cycling
		assert.Equal(t, expected, roundRobinResults, "Round-robin should cycle predictably")

		// Test power law selection (autoCardinality=true) with larger sample
		powerLawResults := make([]int, 100)
		for i := 0; i < 100; i++ {
			powerLawResults[i] = relationship.selectTargetIndex(i, 5, true)
		}

		// Power law should create clustering (not uniform distribution)
		indexCounts := make(map[int]int)
		for _, index := range powerLawResults {
			indexCounts[index]++
		}

		t.Logf("Power law index distribution: %v", indexCounts)

		// Should have variation in index usage (some indices more popular)
		assert.Greater(t, len(indexCounts), 1, "Power law should use multiple indices")

		// Should show clustering effect (some indices used more than others)
		hasVariation := false
		if len(indexCounts) > 1 {
			counts := make([]int, 0)
			for _, count := range indexCounts {
				counts = append(counts, count)
			}
			firstCount := counts[0]
			for _, count := range counts {
				if count != firstCount {
					hasVariation = true
					break
				}
			}
		}
		assert.True(t, hasVariation, "Power law should create varied index usage, not uniform")

		// All indices should be valid
		for _, index := range powerLawResults {
			assert.GreaterOrEqual(t, index, 0, "Index should be >= 0")
			assert.Less(t, index, 5, "Index should be < targetCount")
		}
	})

	t.Run("powerLawIndex should create clustering distribution", func(t *testing.T) {
		// Create a test relationship with mocked entities for direct powerLawIndex testing
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock source entity with 20 rows
		mockSource := NewMockEntityInterface(ctrl)
		mockSource.EXPECT().GetRowCount().Return(20).AnyTimes()

		// Create relationship for direct testing
		relationship := &Relationship{
			id:           "power_law_test",
			sourceEntity: mockSource,
		}

		// Test powerLawIndex directly with different source indices
		results := make([]int, 20)
		for i := 0; i < 20; i++ {
			results[i] = relationship.powerLawIndex(i, 5) // 20 calls, 5 targets
		}

		t.Logf("Power law results for indices 0-19: %v", results)

		// Count index usage
		indexCounts := make(map[int]int)
		for _, index := range results {
			indexCounts[index]++
		}

		t.Logf("Power law index counts: %v", indexCounts)

		// Should create clustering (not uniform 4,4,4,4,4 distribution)
		assert.Greater(t, len(indexCounts), 1, "Power law should use multiple different indices")

		// Should show power law effect (some indices much more popular)
		hasVariation := false
		if len(indexCounts) > 1 {
			counts := make([]int, 0)
			for _, count := range indexCounts {
				counts = append(counts, count)
			}
			firstCount := counts[0]
			for _, count := range counts {
				if count != firstCount {
					hasVariation = true
					break
				}
			}
		}
		assert.True(t, hasVariation, "Power law should create varied usage (some indices more popular)")

		// All indices should be valid
		for _, index := range results {
			assert.GreaterOrEqual(t, index, 0, "Index should be >= 0")
			assert.Less(t, index, 5, "Index should be < targetCount")
		}
	})
}
