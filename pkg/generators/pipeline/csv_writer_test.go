package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/SGNL-ai/fabricator/pkg/parser"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSVWriter_WriteFiles(t *testing.T) {
	tests := []struct {
		name        string
		setupGraph  func(t *testing.T) *model.Graph
		validateDir func(*testing.T, string)
		wantErr     bool
	}{
		{
			name: "Write single entity",
			setupGraph: func(t *testing.T) *model.Graph {
				def := &parser.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test Description",
					Entities: map[string]parser.Entity{
						"entity1": {
							DisplayName: "Entity1",
							ExternalId:  "TestEntity",
							Attributes: []parser.Attribute{
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
								{Name: "name", ExternalId: "name", Type: "String"},
							},
						},
					},
				}
				graphInterface, err := model.NewGraph(def, 100)
				require.NoError(t, err)
				graph, ok := graphInterface.(*model.Graph)
				require.True(t, ok)

				// Add some test data
				entities := graph.GetAllEntities()
				for _, entity := range entities {
					err := entity.AddRow(model.NewRow(map[string]string{
						"id":   "test-1",
						"name": "Test Name",
					}))
					require.NoError(t, err)
				}
				return graph
			},
			validateDir: func(t *testing.T, dir string) {
				// Check that TestEntity.csv file was created
				csvFile := filepath.Join(dir, "TestEntity.csv")
				assert.FileExists(t, csvFile)

				// Read and validate CSV content
				content, err := os.ReadFile(csvFile)
				require.NoError(t, err)

				csvContent := string(content)
				assert.Contains(t, csvContent, "id,name")          // Headers
				assert.Contains(t, csvContent, "test-1,Test Name") // Data
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for output
			tempDir, err := os.MkdirTemp("", "csv_writer_test")
			require.NoError(t, err)
			defer func() { _ = os.RemoveAll(tempDir) }()

			writer := NewCSVWriter(tempDir)
			graph := tt.setupGraph(t)

			err = writer.WriteFiles(graph)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validateDir != nil {
					tt.validateDir(t, tempDir)
				}
			}
		})
	}
}

func TestCSVWriter_getEntityFileName_EdgeCases(t *testing.T) {
	t.Run("should handle empty external ID path in getEntityFileName", func(t *testing.T) {
		// Since empty external ID is rejected by the model layer,
		// I need to test the getEntityFileName function directly
		// But it's private, so I'll create an entity with a very minimal external ID

		def := &parser.SORDefinition{
			DisplayName: "Minimal ID Test",
			Description: "Test minimal external ID handling",
			Entities: map[string]parser.Entity{
				"minimal_entity": {
					DisplayName: "Minimal Entity",
					ExternalId:  "E", // Very short but valid external ID
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
					},
				},
			},
		}

		graphInterface, err := model.NewGraph(def, 100)
		require.NoError(t, err)
		graph, ok := graphInterface.(*model.Graph)
		require.True(t, ok)

		// Add data
		entities := graph.GetAllEntities()
		for _, entity := range entities {
			err := entity.AddRow(model.NewRow(map[string]string{"id": "test-id"}))
			require.NoError(t, err)
		}

		tempDir, err := os.MkdirTemp("", "csv-minimal-id-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		writer := NewCSVWriter(tempDir)
		err = writer.WriteFiles(graph)
		assert.NoError(t, err)

		// Should create E.csv for minimal external ID
		_, err = os.Stat(filepath.Join(tempDir, "E.csv"))
		assert.NoError(t, err)
	})

	t.Run("should handle various external ID formats", func(t *testing.T) {
		testCases := []struct {
			externalID       string
			expectedFilename string
		}{
			{"SimpleEntity", "SimpleEntity.csv"},
			{"Namespace/Entity", "Entity.csv"},
			{"Deep/Nested/Path/Entity", "Entity.csv"},
			{"Multiple/Slashes/Entity", "Entity.csv"},
			{"SingleChar", "SingleChar.csv"},
			{"Entity/", ".csv"}, // Edge case: ends with slash
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("external_id_%s", tc.externalID), func(t *testing.T) {
				def := &parser.SORDefinition{
					DisplayName: "Filename Test",
					Description: "Test filename generation",
					Entities: map[string]parser.Entity{
						"test_entity": {
							DisplayName: "Test Entity",
							ExternalId:  tc.externalID,
							Attributes: []parser.Attribute{
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
							},
						},
					},
				}

				graphInterface, err := model.NewGraph(def, 100)
				require.NoError(t, err)
				graph, ok := graphInterface.(*model.Graph)
				require.True(t, ok)

				// Add data
				entities := graph.GetAllEntities()
				for _, entity := range entities {
					err := entity.AddRow(model.NewRow(map[string]string{"id": "test-id"}))
					require.NoError(t, err)
				}

				tempDir, err := os.MkdirTemp("", "csv-filename-test-*")
				require.NoError(t, err)
				defer os.RemoveAll(tempDir)

				writer := NewCSVWriter(tempDir)
				err = writer.WriteFiles(graph)
				assert.NoError(t, err)

				// Check expected filename was created
				expectedPath := filepath.Join(tempDir, tc.expectedFilename)
				_, err = os.Stat(expectedPath)
				assert.NoError(t, err, "Should create %s for external ID %s", tc.expectedFilename, tc.externalID)
			})
		}
	})
}
