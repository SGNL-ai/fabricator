package util

import "strings"

// CleanNameForFilename creates a filesystem-safe name from a display name
func CleanNameForFilename(name string) string {
	// Replace spaces and slashes with underscores
	cleaned := strings.ReplaceAll(name, " ", "_")
	cleaned = strings.ReplaceAll(cleaned, "/", "_")

	// Remove other potentially problematic characters
	cleaned = strings.ReplaceAll(cleaned, ":", "_")
	cleaned = strings.ReplaceAll(cleaned, "\\", "_")
	cleaned = strings.ReplaceAll(cleaned, "?", "_")
	cleaned = strings.ReplaceAll(cleaned, "*", "_")
	cleaned = strings.ReplaceAll(cleaned, "|", "_")
	cleaned = strings.ReplaceAll(cleaned, "<", "_")
	cleaned = strings.ReplaceAll(cleaned, ">", "_")
	cleaned = strings.ReplaceAll(cleaned, "\"", "_")

	// If the name is empty or only underscores, use a default
	if cleaned == "" || strings.Trim(cleaned, "_") == "" {
		cleaned = "entity_relationship_diagram"
	}
	return cleaned
}
