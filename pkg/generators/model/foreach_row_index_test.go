package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestForEachRowWithIndex tests that ForEachRow provides row index to the callback
func TestForEachRowWithIndex(t *testing.T) {
	t.Run("should provide correct row indices during iteration", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock graph
		mockGraph := NewMockGraphInterface(ctrl)
		mockGraph.EXPECT().GetExpectedDataVolume().Return(100).AnyTimes()

		// Create test attribute
		idAttr := &Attribute{
			name:       "id",
			externalID: "id",
			isUnique:   true,
		}

		// Create entity
		entity, err := newEntity("test", "Test", "Test Entity", "Description", []AttributeInterface{idAttr}, mockGraph)
		require.NoError(t, err)

		// Add 3 test rows
		for i := 0; i < 3; i++ {
			err := entity.AddRow(NewRow(map[string]string{
				"id": "row-" + string(rune('0'+i)),
			}))
			require.NoError(t, err)
		}

		// Track indices provided to callback
		providedIndices := make([]int, 0)
		providedRowIds := make([]string, 0)

		// Test ForEachRow with index
		err = entity.ForEachRow(func(row *Row, index int) error {
			providedIndices = append(providedIndices, index)
			providedRowIds = append(providedRowIds, row.GetValue("id"))
			return nil
		})
		require.NoError(t, err)

		// Verify correct indices were provided
		expectedIndices := []int{0, 1, 2}
		assert.Equal(t, expectedIndices, providedIndices, "Should provide correct row indices")

		// Verify rows were processed in order
		expectedRowIds := []string{"row-0", "row-1", "row-2"}
		assert.Equal(t, expectedRowIds, providedRowIds, "Should process rows in order")

		// Verify all rows are still present after iteration
		assert.Equal(t, 3, entity.GetRowCount(), "Should maintain all rows after ForEachRow")
	})

	t.Run("should handle empty entity gracefully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGraph := NewMockGraphInterface(ctrl)
		mockGraph.EXPECT().GetExpectedDataVolume().Return(100).AnyTimes()

		idAttr := &Attribute{
			name:       "id",
			externalID: "id",
			isUnique:   true,
		}

		entity, err := newEntity("empty", "Empty", "Empty Entity", "Description", []AttributeInterface{idAttr}, mockGraph)
		require.NoError(t, err)

		// ForEachRow on empty entity should not call the function
		callCount := 0
		err = entity.ForEachRow(func(row *Row, index int) error {
			callCount++
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 0, callCount, "Should not call function for empty entity")
	})
}
