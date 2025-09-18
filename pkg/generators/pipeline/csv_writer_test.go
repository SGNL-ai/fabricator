package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/SGNL-ai/fabricator/pkg/models"
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
				def := &models.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test Description",
					Entities: map[string]models.Entity{
						"entity1": {
							DisplayName: "Entity1",
							ExternalId:  "TestEntity",
							Attributes: []models.Attribute{
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
								{Name: "name", ExternalId: "name", Type: "String"},
							},
						},
					},
				}
				graphInterface, err := model.NewGraph(def)
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
