package orchestrator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiagramOrchestrator(t *testing.T) {
	t.Run("should generate diagram successfully", func(t *testing.T) {
		def := &parser.SORDefinition{
			DisplayName: "Test SOR",
			Description: "Test Description",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
						{Name: "name", ExternalId: "name", Type: "String"},
					},
				},
			},
		}

		tempDir, err := os.MkdirTemp("", "diagram_test_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		options := DiagramOptions{}

		result, err := RunDiagramGeneration(def, tempDir, options)

		// Should succeed
		assert.NoError(t, err, "Diagram generation should succeed")
		assert.NotNil(t, result, "Should return diagram result")
		assert.True(t, result.Generated, "Should indicate diagram was generated")
		assert.NotEmpty(t, result.Path, "Should provide diagram path")

		// Verify diagram file was created
		assert.FileExists(t, result.Path, "Diagram file should exist")

		// Check filename is based on SOR name
		expectedFilename := "Test_SOR" // Cleaned version of "Test SOR"
		filename := filepath.Base(result.Path)
		assert.Contains(t, filename, expectedFilename, "Filename should be based on SOR name")
	})

	t.Run("should handle special characters in SOR name", func(t *testing.T) {
		def := &parser.SORDefinition{
			DisplayName: "Complex SOR/Name: With*Special?Chars",
			Description: "Test Description",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
					},
				},
			},
		}

		tempDir, err := os.MkdirTemp("", "diagram_test_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		options := DiagramOptions{}

		result, err := RunDiagramGeneration(def, tempDir, options)

		assert.NoError(t, err, "Should handle special characters gracefully")
		assert.True(t, result.Generated, "Should generate diagram")

		// Verify filename is filesystem-safe
		filename := filepath.Base(result.Path)
		assert.NotContains(t, filename, "/", "Filename should not contain slashes")
		assert.NotContains(t, filename, ":", "Filename should not contain colons")
		assert.NotContains(t, filename, "*", "Filename should not contain asterisks")
		assert.NotContains(t, filename, "?", "Filename should not contain question marks")
	})

	t.Run("should handle empty SOR name", func(t *testing.T) {
		def := &parser.SORDefinition{
			DisplayName: "", // Empty name
			Description: "Test Description",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
					},
				},
			},
		}

		tempDir, err := os.MkdirTemp("", "diagram_test_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		options := DiagramOptions{}

		result, err := RunDiagramGeneration(def, tempDir, options)

		assert.NoError(t, err, "Should handle empty SOR name gracefully")
		assert.True(t, result.Generated, "Should generate diagram with default name")

		// Should use default filename
		filename := filepath.Base(result.Path)
		assert.Contains(t, filename, "entity_relationship_diagram", "Should use default filename for empty name")
	})

	t.Run("should detect Graphviz availability and choose extension", func(t *testing.T) {
		def := &parser.SORDefinition{
			DisplayName: "Test SOR",
			Description: "Test Description",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
					},
				},
			},
		}

		tempDir, err := os.MkdirTemp("", "diagram_test_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		options := DiagramOptions{}

		result, err := RunDiagramGeneration(def, tempDir, options)

		assert.NoError(t, err, "Should succeed regardless of Graphviz availability")
		assert.True(t, result.Generated, "Should generate diagram")

		// Verify file extension is appropriate (.svg if Graphviz available, .dot otherwise)
		ext := filepath.Ext(result.Path)
		assert.Contains(t, []string{".svg", ".dot"}, ext, "Should have appropriate file extension")
	})
}
