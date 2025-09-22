package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanNameForFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "should clean spaces",
			input:    "Test SOR Name",
			expected: "Test_SOR_Name",
		},
		{
			name:     "should clean slashes",
			input:    "App/SOR/Name",
			expected: "App_SOR_Name",
		},
		{
			name:     "should clean special characters",
			input:    "SOR: Name*With?Special<Chars>",
			expected: "SOR__Name_With_Special_Chars_",
		},
		{
			name:     "should handle backslashes",
			input:    "Windows\\Path\\Name",
			expected: "Windows_Path_Name",
		},
		{
			name:     "should handle pipe characters",
			input:    "Name|With|Pipes",
			expected: "Name_With_Pipes",
		},
		{
			name:     "should handle quotes",
			input:    "Name\"With\"Quotes",
			expected: "Name_With_Quotes",
		},
		{
			name:     "should handle empty string",
			input:    "",
			expected: "entity_relationship_diagram",
		},
		{
			name:     "should handle only special characters",
			input:    ":/\\?*|<>\"",
			expected: "entity_relationship_diagram",
		},
		{
			name:     "should preserve valid characters",
			input:    "Valid-Name_123",
			expected: "Valid-Name_123",
		},
		{
			name:     "should handle mixed valid and invalid",
			input:    "Valid/Name: With-Spaces_123",
			expected: "Valid_Name__With-Spaces_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanNameForFilename(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
