package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name           string
		err            *ValidationError
		expectedMsg    string
		expectedSugg   string
		expectsMessage bool
		expectsSugg    bool
	}{
		{
			name: "Complete validation error with all fields",
			err: &ValidationError{
				EntityID:   "users",
				Field:      "count",
				Value:      -5,
				Message:    "Invalid count for entity 'users': -5 (expected positive integer)",
				Suggestion: "Use a number like 100, 1000, etc.",
			},
			expectedMsg:    "Invalid count for entity 'users': -5 (expected positive integer)",
			expectedSugg:   "Use a number like 100, 1000, etc.",
			expectsMessage: true,
			expectsSugg:    true,
		},
		{
			name: "Validation error with minimal fields",
			err: &ValidationError{
				Message:    "Configuration file not found",
				Suggestion: "Check the file path and ensure the file exists",
			},
			expectedMsg:    "Configuration file not found",
			expectedSugg:   "Check the file path and ensure the file exists",
			expectsMessage: true,
			expectsSugg:    true,
		},
		{
			name: "Validation error with empty suggestion",
			err: &ValidationError{
				EntityID:   "groups",
				Message:    "Invalid entity reference",
				Suggestion: "",
			},
			expectedMsg:    "Invalid entity reference",
			expectsMessage: true,
			expectsSugg:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()

			if tt.expectsMessage {
				assert.Contains(t, errMsg, tt.expectedMsg, "Error message should contain the message")
			}

			if tt.expectsSugg && tt.expectedSugg != "" {
				assert.Contains(t, errMsg, tt.expectedSugg, "Error message should contain the suggestion")
				assert.Contains(t, errMsg, "Suggestion:", "Error should include 'Suggestion:' label")
			}

			// Verify format: "Message\nSuggestion: suggestion"
			if tt.expectsMessage && tt.expectsSugg && tt.expectedSugg != "" {
				parts := strings.Split(errMsg, "\n")
				assert.GreaterOrEqual(t, len(parts), 2, "Error should have message and suggestion on separate lines")
			}
		})
	}
}

func TestValidationError_Fields(t *testing.T) {
	err := &ValidationError{
		EntityID:   "permissions",
		Field:      "count",
		Value:      0,
		Message:    "Count cannot be zero",
		Suggestion: "Use a positive integer",
	}

	assert.Equal(t, "permissions", err.EntityID, "EntityID should match")
	assert.Equal(t, "count", err.Field, "Field should match")
	assert.Equal(t, 0, err.Value, "Value should match")
	assert.Equal(t, "Count cannot be zero", err.Message, "Message should match")
	assert.Equal(t, "Use a positive integer", err.Suggestion, "Suggestion should match")
}
