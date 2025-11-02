package config

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T046: Test for ConfigurationTemplate.Render with entities
func TestConfigurationTemplate_Render_WithEntities(t *testing.T) {
	entities := []TemplateEntity{
		{
			ExternalID:   "users",
			DisplayName:  "Users",
			Description:  "User accounts",
			DefaultCount: 100,
		},
		{
			ExternalID:   "groups",
			DisplayName:  "Groups",
			Description:  "User groups",
			DefaultCount: 50,
		},
	}

	template := NewTemplate("test.yaml", entities, 100)
	require.NotNil(t, template, "NewTemplate should return non-nil template")

	rendered, err := template.Render()
	require.NoError(t, err, "Render should not return error")
	require.NotEmpty(t, rendered, "Rendered output should not be empty")

	output := string(rendered)

	// Verify header comments
	assert.Contains(t, output, "# Row count configuration for fabricator")
	assert.Contains(t, output, "# Generated from: test.yaml")
	assert.Contains(t, output, "# Last updated:")

	// Verify first entity
	assert.Contains(t, output, "# Entity: users")
	assert.Contains(t, output, "# Name: Users")
	assert.Contains(t, output, "# Description: User accounts")
	assert.Contains(t, output, "users: 100")

	// Verify second entity
	assert.Contains(t, output, "# Entity: groups")
	assert.Contains(t, output, "# Name: Groups")
	assert.Contains(t, output, "# Description: User groups")
	assert.Contains(t, output, "groups: 50")
}

// T047: Test for ConfigurationTemplate.Render with empty entities
func TestConfigurationTemplate_Render_WithEmptyEntities(t *testing.T) {
	template := NewTemplate("empty.yaml", []TemplateEntity{}, 100)
	require.NotNil(t, template, "NewTemplate should return non-nil template even with empty entities")

	rendered, err := template.Render()
	require.NoError(t, err, "Render should not return error with empty entities")
	require.NotEmpty(t, rendered, "Rendered output should contain header even with no entities")

	output := string(rendered)

	// Should still have header
	assert.Contains(t, output, "# Row count configuration for fabricator")
	assert.Contains(t, output, "# Generated from: empty.yaml")

	// Should not have any entity entries
	assert.NotContains(t, output, "# Entity:")
}

// T048: Test for ConfigurationTemplate.Render with descriptions
func TestConfigurationTemplate_Render_WithDescriptions(t *testing.T) {
	entities := []TemplateEntity{
		{
			ExternalID:   "products",
			DisplayName:  "Products",
			Description:  "Product catalog items with detailed information",
			DefaultCount: 200,
		},
		{
			ExternalID:   "orders",
			DisplayName:  "Orders",
			Description:  "Customer orders",
			DefaultCount: 500,
		},
	}

	template := NewTemplate("catalog.yaml", entities, 100)
	rendered, err := template.Render()
	require.NoError(t, err)

	output := string(rendered)

	// Verify descriptions are included
	assert.Contains(t, output, "# Description: Product catalog items with detailed information")
	assert.Contains(t, output, "# Description: Customer orders")

	// Verify entity values
	assert.Contains(t, output, "products: 200")
	assert.Contains(t, output, "orders: 500")
}

// T049: Test for ConfigurationTemplate.WriteToWriter stdout
func TestConfigurationTemplate_WriteToWriter(t *testing.T) {
	entities := []TemplateEntity{
		{
			ExternalID:   "test_entity",
			DisplayName:  "Test Entity",
			Description:  "A test entity for WriteToWriter",
			DefaultCount: 42,
		},
	}

	template := NewTemplate("test.yaml", entities, 100)

	// Create a buffer to capture output
	var buf bytes.Buffer
	err := template.WriteToWriter(&buf)
	require.NoError(t, err, "WriteToWriter should not return error")

	output := buf.String()
	assert.NotEmpty(t, output, "WriteToWriter should write non-empty output")

	// Verify content was written
	assert.Contains(t, output, "# Entity: test_entity")
	assert.Contains(t, output, "test_entity: 42")
}

// T050: Test for NewTemplate factory function
func TestNewTemplate(t *testing.T) {
	tests := []struct {
		name         string
		sourceFile   string
		entities     []TemplateEntity
		defaultCount int
		wantNil      bool
	}{
		{
			name:       "valid template with entities",
			sourceFile: "sor.yaml",
			entities: []TemplateEntity{
				{ExternalID: "users", DisplayName: "Users", DefaultCount: 100},
			},
			defaultCount: 100,
			wantNil:      false,
		},
		{
			name:         "valid template with empty entities",
			sourceFile:   "empty.yaml",
			entities:     []TemplateEntity{},
			defaultCount: 50,
			wantNil:      false,
		},
		{
			name:         "valid template with zero default count",
			sourceFile:   "zero.yaml",
			entities:     []TemplateEntity{{ExternalID: "test"}},
			defaultCount: 0,
			wantNil:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := NewTemplate(tt.sourceFile, tt.entities, tt.defaultCount)

			if tt.wantNil {
				assert.Nil(t, template, "Expected nil template")
			} else {
				require.NotNil(t, template, "Expected non-nil template")
				assert.Equal(t, tt.sourceFile, template.SourceSORFile)
				assert.Equal(t, len(tt.entities), len(template.Entities))
				assert.Equal(t, tt.defaultCount, template.DefaultCount)
				assert.False(t, template.GeneratedAt.IsZero(), "GeneratedAt should be set")
			}
		})
	}
}

// Additional test: Verify timestamp format
func TestConfigurationTemplate_Render_TimestampFormat(t *testing.T) {
	entities := []TemplateEntity{
		{ExternalID: "test", DisplayName: "Test", DefaultCount: 10},
	}

	template := NewTemplate("test.yaml", entities, 100)
	// Set a known time for testing
	knownTime := time.Date(2025, 10, 30, 15, 30, 45, 0, time.UTC)
	template.GeneratedAt = knownTime

	rendered, err := template.Render()
	require.NoError(t, err)

	output := string(rendered)
	assert.Contains(t, output, "# Last updated: 2025-10-30 15:30:45")
}

// Additional test: Verify entity without display name
func TestConfigurationTemplate_Render_EntityWithoutDisplayName(t *testing.T) {
	entities := []TemplateEntity{
		{
			ExternalID:   "simple_entity",
			DisplayName:  "", // Empty display name
			Description:  "An entity without display name",
			DefaultCount: 25,
		},
	}

	template := NewTemplate("test.yaml", entities, 100)
	rendered, err := template.Render()
	require.NoError(t, err)

	output := string(rendered)

	// Should have entity ID
	assert.Contains(t, output, "# Entity: simple_entity")
	// Should NOT have "# Name:" line since DisplayName is empty
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "simple_entity") {
			// Check that there's no "# Name:" line immediately after
			assert.NotContains(t, line, "# Name:")
		}
	}
	// Should still have description
	assert.Contains(t, output, "# Description: An entity without display name")
	assert.Contains(t, output, "simple_entity: 25")
}

// Additional test: Verify entity with same ExternalID and DisplayName
func TestConfigurationTemplate_Render_EntityWithSameIDAndDisplayName(t *testing.T) {
	entities := []TemplateEntity{
		{
			ExternalID:   "User",
			DisplayName:  "User", // Same as ExternalID
			Description:  "User entity",
			DefaultCount: 100,
		},
	}

	template := NewTemplate("test.yaml", entities, 100)
	rendered, err := template.Render()
	require.NoError(t, err)

	output := string(rendered)

	// Should have entity ID
	assert.Contains(t, output, "# Entity: User")
	// Should NOT have "# Name: User" since it's the same as ExternalID
	lines := strings.Split(output, "\n")
	nameLineCount := 0
	for _, line := range lines {
		if strings.Contains(line, "# Name: User") {
			nameLineCount++
		}
	}
	assert.Equal(t, 0, nameLineCount, "Should not include Name comment when DisplayName equals ExternalID")
}

// Additional test: Verify default count fallback
func TestConfigurationTemplate_Render_DefaultCountFallback(t *testing.T) {
	entities := []TemplateEntity{
		{
			ExternalID:   "entity_with_zero_count",
			DisplayName:  "Entity",
			DefaultCount: 0, // Zero count should fall back to template default
		},
	}

	template := NewTemplate("test.yaml", entities, 200)
	rendered, err := template.Render()
	require.NoError(t, err)

	output := string(rendered)

	// Should use template's default count (200) instead of entity's (0)
	assert.Contains(t, output, "entity_with_zero_count: 200")
}

// Additional test: Verify WriteToWriter error handling (write failure)
func TestConfigurationTemplate_WriteToWriter_WriteError(t *testing.T) {
	entities := []TemplateEntity{
		{ExternalID: "test", DefaultCount: 10},
	}

	template := NewTemplate("test.yaml", entities, 100)

	// Use a writer that always fails
	writer := &failingWriter{}
	err := template.WriteToWriter(writer)

	assert.Error(t, err, "WriteToWriter should return error when write fails")
	assert.Contains(t, err.Error(), "failed to write template")
}

// Helper: A writer that always fails
type failingWriter struct{}

func (w *failingWriter) Write(p []byte) (n int, err error) {
	return 0, assert.AnError // Always return an error
}
