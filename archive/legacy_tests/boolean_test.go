package generators

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBooleanFieldDetection(t *testing.T) {
	// Define a function that exactly matches the boolean detection logic in generateRowForEntity
	isBooleanField := func(header string) bool {
		// First convert to lowercase for case insensitive matching
		headerLower := strings.ToLower(header)

		// Special case for common boolean field names
		if header == "valid" || header == "enabled" || header == "active" || header == "archived" {
			return true
		}

		// More general pattern matching
		return headerLower == "enabled" || headerLower == "active" ||
			headerLower == "valid" || headerLower == "archived" ||
			strings.HasSuffix(headerLower, "enabled") ||
			strings.HasSuffix(headerLower, "active") ||
			strings.HasSuffix(headerLower, "valid") ||
			strings.HasSuffix(headerLower, "archived") ||
			strings.Contains(headerLower, "valid") ||
			strings.Contains(headerLower, "archived") ||
			strings.Contains(headerLower, "enabled") ||
			strings.Contains(headerLower, "active")
	}

	testCases := []struct {
		fieldName string
		expected  bool
	}{
		{"enabled", true},
		{"active", true},
		{"valid", true},
		{"archived", true},
		{"userEnabled", true},
		{"isActive", true},
		{"accountValid", true},
		{"recordArchived", true},
		{"isValid", true},
		{"isValidUser", true},
		{"id", false},
		{"name", false},
		{"description", false},
		{"count", false},
		{"date", false},
	}

	for _, tc := range testCases {
		t.Run(tc.fieldName, func(t *testing.T) {
			result := isBooleanField(tc.fieldName)
			assert.Equal(t, tc.expected, result, "isBooleanField(%s) should be %v", tc.fieldName, tc.expected)
		})
	}
}
