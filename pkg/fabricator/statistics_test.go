package fabricator

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/stretchr/testify/assert"
)

func TestPrintGraphStatistics(t *testing.T) {
	t.Run("should display statistics without panicking", func(t *testing.T) {
		stats := &model.GraphStatistics{
			SORName:                "Test SOR",
			Description:            "Test Description",
			EntityCount:            2,
			TotalAttributes:        5,
			UniqueAttributes:       2,
			IndexedAttributes:      3,
			ListAttributes:         1,
			RelationshipCount:      2,
			DirectRelationships:    1,
			PathBasedRelationships: 1,
			NamespaceFormats: map[string]int{
				"TestApp": 2,
			},
		}

		// This test mainly ensures the function doesn't panic
		// Since it outputs to console, we can't easily test the output content
		assert.NotPanics(t, func() {
			PrintGraphStatistics(stats)
		}, "PrintGraphStatistics should not panic")
	})

	t.Run("should handle empty namespace formats", func(t *testing.T) {
		stats := &model.GraphStatistics{
			SORName:          "Empty SOR",
			Description:      "Empty Description",
			NamespaceFormats: map[string]int{},
		}

		assert.NotPanics(t, func() {
			PrintGraphStatistics(stats)
		}, "Should handle empty namespace formats without panicking")
	})

	t.Run("should handle multiple namespace formats", func(t *testing.T) {
		stats := &model.GraphStatistics{
			SORName:     "Multi SOR",
			Description: "Multi Description",
			NamespaceFormats: map[string]int{
				"App1":           2,
				"App2":           3,
				"(no namespace)": 1,
			},
		}

		assert.NotPanics(t, func() {
			PrintGraphStatistics(stats)
		}, "Should handle multiple namespace formats without panicking")
	})
}
