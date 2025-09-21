package orchestrator

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed test_fixtures/simple_dotted_fk.yaml
var simpleDottedFKYAML string

//go:embed test_fixtures/simple_alias_fk.yaml
var simpleAliasFKYAML string

// TestForeignKeyIntegration tests end-to-end foreign key relationship consistency
// This integration test verifies that generated CSV files maintain proper foreign key relationships
func TestForeignKeyIntegration(t *testing.T) {
	testCases := []struct {
		name             string
		yamlContent      string
		useExternalFile  bool
		externalFilePath string
		expectedFiles    int
		expectedRecords  int
		dataVolume       int
		description      string
	}{
		{
			name:            "dotted notation relationships",
			yamlContent:     simpleDottedFKYAML,
			expectedFiles:   2,
			expectedRecords: 6,
			dataVolume:      3,
			description:     "Simple User->Profile relationship using dotted notation (User.id -> Profile.user_id)",
		},
		{
			name:            "attributeAlias relationships",
			yamlContent:     simpleAliasFKYAML,
			expectedFiles:   2,
			expectedRecords: 6,
			dataVolume:      3,
			description:     "GroupMember->Group relationship using attributeAlias like exported okta (correct FK -> PK direction)",
		},
		{
			name:             "real okta.sgnl.yaml",
			useExternalFile:  true,
			externalFilePath: "../../examples/okta.sgnl.yaml",
			expectedFiles:    4, // User, Group, GroupMember, Application
			expectedRecords:  8, // 2 records per entity * 4 entities
			dataVolume:       2,
			description:      "Real okta.sgnl.yaml with dotted notation relationships (should work)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "fk-test-*")
			require.NoError(t, err)
			defer func() { _ = os.RemoveAll(tempDir) }()

			var yamlPath string
			if tc.useExternalFile {
				yamlPath = tc.externalFilePath
			} else {
				yamlPath = filepath.Join(tempDir, "test.yaml")
				err = os.WriteFile(yamlPath, []byte(tc.yamlContent), 0644)
				require.NoError(t, err)
			}

			// Parse the YAML
			parser := parser.NewParser(yamlPath)
			err = parser.Parse()
			require.NoError(t, err)

			// Generate data
			outputDir := filepath.Join(tempDir, "output")
			options := GenerationOptions{
				DataVolume:      tc.dataVolume,
				AutoCardinality: false,
				GenerateDiagram: false,
				ValidateResults: false,
			}

			_, err = RunGeneration(parser.Definition, outputDir, options)
			require.NoError(t, err)

			// Use validation mode to check relationships
			validationOptions := ValidationOptions{
				GenerateDiagram: false,
			}

			validationResult, err := RunValidation(parser.Definition, outputDir, validationOptions)
			require.NoError(t, err)

			// CRITICAL TEST: Foreign key relationships should be valid regardless of notation style
			assert.Empty(t, validationResult.ValidationErrors,
				"Foreign key relationships should be valid for %s - found %d errors: %v",
				tc.description, len(validationResult.ValidationErrors), validationResult.ValidationErrors)

			// Verify expected file and record counts
			assert.Equal(t, tc.expectedFiles, validationResult.FilesValidated,
				"Should have validated %d CSV files for %s", tc.expectedFiles, tc.description)
			assert.Equal(t, tc.expectedRecords, validationResult.RecordsValidated,
				"Should have validated %d records for %s", tc.expectedRecords, tc.description)
		})
	}
}
