package config

import (
	"fmt"
	"io"
	"time"
)

// ConfigurationTemplate holds data for generating a count config template.
// It represents the structure needed to generate a user-friendly YAML
// configuration file with comments and default values.
type ConfigurationTemplate struct {
	// Entities to include in template
	Entities []TemplateEntity

	// SourceSORFile is the SOR YAML file used to generate template
	SourceSORFile string

	// DefaultCount is the placeholder count value for each entity
	DefaultCount int

	// GeneratedAt timestamp
	GeneratedAt time.Time
}

// TemplateEntity represents an entity in the template with metadata.
// It includes all the information needed to generate a commented YAML entry.
type TemplateEntity struct {
	// ExternalID is the entity's external_id (YAML key)
	ExternalID string

	// DisplayName is a human-readable name (if available from SOR)
	DisplayName string

	// Description is a brief description (if available from SOR)
	Description string

	// DefaultCount is the suggested row count for this entity
	DefaultCount int
}

// NewTemplate creates a new ConfigurationTemplate with the given entities.
// defaultCount is the suggested row count for all entities.
func NewTemplate(sourceFile string, entities []TemplateEntity, defaultCount int) *ConfigurationTemplate {
	return &ConfigurationTemplate{
		Entities:      entities,
		SourceSORFile: sourceFile,
		DefaultCount:  defaultCount,
		GeneratedAt:   time.Now(),
	}
}

// Render generates YAML output with comments.
// Returns the formatted YAML as bytes.
func (t *ConfigurationTemplate) Render() ([]byte, error) {
	var output string

	// Add header comments
	output += "# Row count configuration for fabricator\n"
	output += fmt.Sprintf("# Generated from: %s\n", t.SourceSORFile)
	output += fmt.Sprintf("# Last updated: %s\n", t.GeneratedAt.Format("2006-01-02 15:04:05"))
	output += "#\n"
	output += "# Edit the numbers below to specify how many rows to generate for each entity.\n"
	output += "# Entities not listed here will use the default count (100).\n"
	output += "\n"

	// Add each entity with comments
	for _, entity := range t.Entities {
		// Add entity comment block
		output += fmt.Sprintf("# Entity: %s\n", entity.ExternalID)
		if entity.DisplayName != "" && entity.DisplayName != entity.ExternalID {
			output += fmt.Sprintf("# Name: %s\n", entity.DisplayName)
		}
		if entity.Description != "" {
			output += fmt.Sprintf("# Description: %s\n", entity.Description)
		}

		// Add the YAML entry
		count := entity.DefaultCount
		if count == 0 {
			count = t.DefaultCount
		}
		output += fmt.Sprintf("%s: %d\n", entity.ExternalID, count)
		output += "\n"
	}

	return []byte(output), nil
}

// WriteToWriter writes the template to an io.Writer (e.g., os.Stdout).
// This is useful for writing the template directly to stdout or a file.
// Note: renamed from WriteTo to avoid confusion with io.WriterTo interface
func (t *ConfigurationTemplate) WriteToWriter(w io.Writer) error {
	rendered, err := t.Render()
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	_, err = w.Write(rendered)
	if err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	return nil
}
