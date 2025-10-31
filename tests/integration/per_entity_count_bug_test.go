package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/config"
	"github.com/SGNL-ai/fabricator/pkg/orchestrator"
	"github.com/SGNL-ai/fabricator/pkg/parser"
	"github.com/stretchr/testify/require"
)

// TestPerEntityCountBug reproduces the duplicate UUID bug when using per-entity counts
// where source entity has more rows than target entity.
//
// Bug: When Group (107 rows) links to EntraIdGroup (105 rows), ForEachRow fails
// with "duplicate value for unique attribute 'id'" at row 105.
//
// This test should FAIL until the bug is fixed.
func TestPerEntityCountBug_SourceHasMoreRowsThanTarget(t *testing.T) {
	// Create minimal YAML with 2 entities and 1 relationship
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "test.yaml")
	yamlContent := `displayName: Minimal Bug Repro
description: Reproduce duplicate UUID bug

entities:
  Group:
    displayName: Group
    externalId: Group
    attributes:
      - name: id
        externalId: id
        type: String
        uniqueId: true

  EntraIdGroup:
    displayName: EntraIdGroup
    externalId: EntraIdGroup
    attributes:
      - name: id
        externalId: id
        type: String
        uniqueId: true

relationships:
  GroupToEntraIdGroup:
    name: GroupToEntraIdGroup
    fromAttribute: Group.id
    toAttribute: EntraIdGroup.id
`
	err := os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Create count config: Group (107) > EntraIdGroup (105)
	countConfigPath := filepath.Join(tmpDir, "counts.yaml")
	countConfigContent := `Group: 107
EntraIdGroup: 105
`
	err = os.WriteFile(countConfigPath, []byte(countConfigContent), 0644)
	require.NoError(t, err)

	// Parse SOR
	p := parser.NewParser(yamlPath)
	err = p.Parse()
	require.NoError(t, err)

	// Load count config
	countConfig, err := config.LoadConfiguration(countConfigPath)
	require.NoError(t, err)

	// Test with both round-robin and power-law distribution
	testCases := []struct {
		name            string
		autoCardinality bool
		description     string
	}{
		{
			name:            "round-robin",
			autoCardinality: false,
			description:     "Round-robin causes modulo wrapping: row 105 % 105 = 0, duplicating row 0's ID",
		},
		{
			name:            "power-law",
			autoCardinality: true,
			description:     "Power-law uses GetRowCount() which shrinks during ForEachRow, causing wrong calculations",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outputDir := filepath.Join(tmpDir, "output-"+tc.name)

			options := orchestrator.GenerationOptions{
				DataVolume:      100,
				CountConfig:     countConfig,
				AutoCardinality: tc.autoCardinality,
				GenerateDiagram: false,
				ValidateResults: false,
			}

			_, err = orchestrator.RunGeneration(p.Definition, outputDir, options)

			// BUG: This currently fails with "duplicate value for unique attribute 'id'"
			// Reason: tc.description
			// After fix, this should succeed
			require.NoError(t, err, "Should generate successfully with source > target counts (%s)", tc.description)

			// Verify output
			verifyCSVRowCount(t, filepath.Join(outputDir, "Group.csv"), 107)
			verifyCSVRowCount(t, filepath.Join(outputDir, "EntraIdGroup.csv"), 105)
		})
	}
}
