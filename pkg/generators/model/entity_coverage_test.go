package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestEntityCoverageMethods tests methods that need coverage to reach 85% threshold
func TestEntityCoverageMethods(t *testing.T) {
	t.Run("preAllocateRows should resize entity capacity", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGraph := NewMockGraphInterface(ctrl)
		mockGraph.EXPECT().GetExpectedDataVolume().Return(10).AnyTimes()

		idAttr := &Attribute{
			name:       "id",
			externalID: "id",
			isUnique:   true,
		}

		entity, err := newEntity("test", "Test", "Test Entity", "Description", []AttributeInterface{idAttr}, mockGraph)
		require.NoError(t, err)

		// Initially should have capacity for 10 rows (from graph data volume)
		initialCap := cap(entity.(*Entity).rows)

		// preAllocateRows with larger capacity
		entity.(*Entity).preAllocateRows(1000)
		newCap := cap(entity.(*Entity).rows)

		// Should have increased capacity
		assert.Greater(t, newCap, initialCap, "Should increase capacity when requested size is larger")

		// preAllocateRows with smaller capacity should not decrease
		entity.(*Entity).preAllocateRows(5)
		finalCap := cap(entity.(*Entity).rows)
		assert.Equal(t, newCap, finalCap, "Should not decrease capacity when requested size is smaller")
	})

	t.Run("RemoveRow should remove row and update hash maps", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGraph := NewMockGraphInterface(ctrl)
		mockGraph.EXPECT().GetExpectedDataVolume().Return(100).AnyTimes()

		// Create entity with PK and FK attributes
		idAttr := &Attribute{
			name:           "id",
			externalID:     "id",
			isUnique:       true,
			isRelationship: false,
		}
		fkAttr := &Attribute{
			name:           "user_id",
			externalID:     "user_id",
			isUnique:       false,
			isRelationship: true,
			relatedEntity:  "user",
			relatedAttr:    "id",
		}

		entity, err := newEntity("test", "Test", "Test Entity", "Description", []AttributeInterface{idAttr, fkAttr}, mockGraph)
		require.NoError(t, err)

		// Add test rows
		err = entity.AddRow(NewRow(map[string]string{"id": "1", "user_id": "user-1"}))
		require.NoError(t, err)
		err = entity.AddRow(NewRow(map[string]string{"id": "2", "user_id": "user-2"}))
		require.NoError(t, err)

		// Verify initial state
		assert.Equal(t, 2, entity.GetRowCount(), "Should have 2 rows initially")
		assert.True(t, entity.CheckKeyExists("1"), "PK '1' should be in hash map")
		assert.True(t, entity.CheckKeyExists("2"), "PK '2' should be in hash map")

		// Remove first row
		err = entity.RemoveRow(0)
		require.NoError(t, err)

		// Verify row was removed
		assert.Equal(t, 1, entity.GetRowCount(), "Should have 1 row after removal")

		// Verify PK hash map was updated
		assert.False(t, entity.CheckKeyExists("1"), "PK '1' should be removed from hash map")
		assert.True(t, entity.CheckKeyExists("2"), "PK '2' should still be in hash map")

		// Verify remaining row has correct data
		remainingRow := entity.GetRowByIndex(0)
		require.NotNil(t, remainingRow)
		assert.Equal(t, "2", remainingRow.GetValue("id"), "Remaining row should have id '2'")

		// Test error cases
		err = entity.RemoveRow(-1)
		assert.Error(t, err, "Should error on negative index")

		err = entity.RemoveRow(1)
		assert.Error(t, err, "Should error on index out of range")
	})

	t.Run("IsForeignKeyUnique should check composite key uniqueness", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGraph := NewMockGraphInterface(ctrl)
		mockGraph.EXPECT().GetExpectedDataVolume().Return(100).AnyTimes()

		// Create junction table entity with 2 FK attributes
		idAttr := &Attribute{
			name:           "id",
			externalID:     "id",
			isUnique:       true,
			isRelationship: false,
		}
		userFKAttr := &Attribute{
			name:           "user_id",
			externalID:     "user_id",
			isUnique:       false,
			isRelationship: true,
			relatedEntity:  "user",
			relatedAttr:    "id",
		}
		groupFKAttr := &Attribute{
			name:           "group_id",
			externalID:     "group_id",
			isUnique:       false,
			isRelationship: true,
			relatedEntity:  "group",
			relatedAttr:    "id",
		}

		entity, err := newEntity("membership", "Membership", "User Group Membership", "Junction table",
			[]AttributeInterface{idAttr, userFKAttr, groupFKAttr}, mockGraph)
		require.NoError(t, err)

		// Test unique composite key
		uniqueRow := NewRow(map[string]string{"id": "1", "user_id": "user-1", "group_id": "group-A"})
		assert.True(t, entity.IsForeignKeyUnique(uniqueRow), "New composite key should be unique")

		// Add the row to index it
		err = entity.AddRow(uniqueRow)
		require.NoError(t, err)

		// Test duplicate composite key
		duplicateRow := NewRow(map[string]string{"id": "2", "user_id": "user-1", "group_id": "group-A"})
		assert.False(t, entity.IsForeignKeyUnique(duplicateRow), "Duplicate composite key should not be unique")

		// Test different composite key
		differentRow := NewRow(map[string]string{"id": "3", "user_id": "user-1", "group_id": "group-B"})
		assert.True(t, entity.IsForeignKeyUnique(differentRow), "Different composite key should be unique")

		// Test entity with no FK attributes
		nonJunctionEntity, err := newEntity("simple", "Simple", "Simple Entity", "Non-junction",
			[]AttributeInterface{idAttr}, mockGraph)
		require.NoError(t, err)

		simpleRow := NewRow(map[string]string{"id": "1"})
		assert.True(t, nonJunctionEntity.IsForeignKeyUnique(simpleRow), "Non-junction entity should always be unique")
	})

	t.Run("getRows should return internal rows slice", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGraph := NewMockGraphInterface(ctrl)
		mockGraph.EXPECT().GetExpectedDataVolume().Return(100).AnyTimes()

		idAttr := &Attribute{
			name:       "id",
			externalID: "id",
			isUnique:   true,
		}

		entity, err := newEntity("test", "Test", "Test Entity", "Description", []AttributeInterface{idAttr}, mockGraph)
		require.NoError(t, err)

		// Test empty entity
		rows := entity.(*Entity).getRows()
		assert.Len(t, rows, 0, "Empty entity should return empty rows slice")

		// Add some rows
		err = entity.AddRow(NewRow(map[string]string{"id": "1"}))
		require.NoError(t, err)
		err = entity.AddRow(NewRow(map[string]string{"id": "2"}))
		require.NoError(t, err)

		// Test populated entity
		rows = entity.(*Entity).getRows()
		assert.Len(t, rows, 2, "Should return all rows")
		assert.Equal(t, "1", rows[0].GetValue("id"), "First row should have id '1'")
		assert.Equal(t, "2", rows[1].GetValue("id"), "Second row should have id '2'")
	})

	t.Run("buildCompositeKey edge cases", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGraph := NewMockGraphInterface(ctrl)
		mockGraph.EXPECT().GetExpectedDataVolume().Return(100).AnyTimes()

		// Create attributes for testing composite key building
		fkAttr1 := &Attribute{
			name:           "user_id",
			externalID:     "user_id",
			isRelationship: true,
		}
		fkAttr2 := &Attribute{
			name:           "group_id",
			externalID:     "group_id",
			isRelationship: true,
		}
		fkAttr3 := &Attribute{
			name:           "role_id",
			externalID:     "role_id",
			isRelationship: true,
		}

		idAttr := &Attribute{
			name:       "id",
			externalID: "id",
			isUnique:   true,
		}

		entity, err := newEntity("test", "Test", "Test Entity", "Description",
			[]AttributeInterface{idAttr, fkAttr1, fkAttr2, fkAttr3}, mockGraph)
		require.NoError(t, err)

		// Test composite key with multiple FK attributes (tests the "if i > 0" branch)
		row := NewRow(map[string]string{
			"id":       "1",
			"user_id":  "user-1",
			"group_id": "group-A",
			"role_id":  "role-X",
		})

		compositeKey := entity.(*Entity).buildCompositeKey(row, []AttributeInterface{fkAttr1, fkAttr2, fkAttr3})
		expected := "user-1|group-A|role-X" // Should join with "|" separator
		assert.Equal(t, expected, compositeKey, "Should build composite key with pipe separators")

		// Test single FK attribute (no separator needed)
		singleKey := entity.(*Entity).buildCompositeKey(row, []AttributeInterface{fkAttr1})
		assert.Equal(t, "user-1", singleKey, "Single FK should not have separators")
	})
}
