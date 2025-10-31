package config

import "fmt"

// ValidationError represents a configuration validation failure.
// It provides structured error information with actionable suggestions
// for users to fix configuration issues.
type ValidationError struct {
	// EntityID is the problematic entity external_id (if applicable)
	EntityID string

	// Field is the problematic field name (if applicable)
	Field string

	// Value is the invalid value
	Value interface{}

	// Message is a human-readable description
	Message string

	// Suggestion is an actionable fix suggestion
	Suggestion string
}

// Error implements the error interface.
// Returns a formatted error message with a suggestion for fixing the issue.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s\nSuggestion: %s", e.Message, e.Suggestion)
}
