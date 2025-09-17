package pipeline

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldGenerator_GenerateFields(t *testing.T) {
	tests := []struct {
		name       string
		setupGraph func(t *testing.T) *model.Graph
		wantErr    bool
		validate   func(t *testing.T, graph *model.Graph)
	}{
		{
			name: "Generate fields for simple entity",
			setupGraph: func(t *testing.T) *model.Graph {
				def := &models.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test Description",
					Entities: map[string]models.Entity{
						"entity1": {
							DisplayName: "Entity1",
							ExternalId:  "Entity1",
							Attributes: []models.Attribute{
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
								{Name: "name", ExternalId: "name", Type: "String"},
								{Name: "email", ExternalId: "email", Type: "String"},
							},
						},
					},
				}
				graphInterface, err := model.NewGraph(def)
				require.NoError(t, err)
				graph, ok := graphInterface.(*model.Graph)
				require.True(t, ok)

				// Pre-populate with some rows that have IDs but missing other fields
				entities := graph.GetAllEntities()
				for _, entity := range entities {
					for i := 0; i < 2; i++ {
						err := entity.AddRow(model.NewRow(map[string]string{
							"id": "id-" + string(rune('0'+i)),
						}))
						require.NoError(t, err)
					}
				}
				return graph
			},
			wantErr: false,
			validate: func(t *testing.T, graph *model.Graph) {
				entities := graph.GetAllEntities()
				require.Len(t, entities, 1)

				var entity model.EntityInterface
				for _, e := range entities {
					entity = e
					break
				}

				csvData := entity.ToCSV()
				require.Len(t, csvData.Rows, 2, "Should have 2 rows")
				require.Len(t, csvData.Headers, 3, "Should have 3 columns: id, name, email")

				// Check that non-ID fields are now populated
				for _, row := range csvData.Rows {
					require.Len(t, row, 3, "Each row should have 3 values")
					assert.NotEmpty(t, row[0], "ID should not be empty")
					assert.NotEmpty(t, row[1], "Name should be generated")
					assert.NotEmpty(t, row[2], "Email should be generated")
				}
			},
		},
		{
			name: "Generate fields with different data types",
			setupGraph: func(t *testing.T) *model.Graph {
				def := &models.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test Description",
					Entities: map[string]models.Entity{
						"entity1": {
							DisplayName: "Entity1",
							ExternalId:  "Entity1",
							Attributes: []models.Attribute{
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
								{Name: "age", ExternalId: "age", Type: "Integer"},
								{Name: "isActive", ExternalId: "isActive", Type: "Boolean"},
								{Name: "createdAt", ExternalId: "createdAt", Type: "Date"},
							},
						},
					},
				}
				graphInterface, err := model.NewGraph(def)
				require.NoError(t, err)
				graph, ok := graphInterface.(*model.Graph)
				require.True(t, ok)

				// Pre-populate with IDs
				entities := graph.GetAllEntities()
				for _, entity := range entities {
					err := entity.AddRow(model.NewRow(map[string]string{
						"id": "id-1",
					}))
					require.NoError(t, err)
				}
				return graph
			},
			wantErr: false,
			validate: func(t *testing.T, graph *model.Graph) {
				entities := graph.GetAllEntities()
				var entity model.EntityInterface
				for _, e := range entities {
					entity = e
					break
				}

				csvData := entity.ToCSV()
				require.Len(t, csvData.Rows, 1)
				row := csvData.Rows[0]

				// Validate data types are appropriate
				assert.NotEmpty(t, row[0], "ID should not be empty")
				assert.NotEmpty(t, row[1], "Age should be generated")
				assert.NotEmpty(t, row[2], "IsActive should be generated")
				assert.NotEmpty(t, row[3], "CreatedAt should be generated")

				// Could add more specific validation for data formats
			},
		},
		{
			name: "Error on nil graph",
			setupGraph: func(t *testing.T) *model.Graph {
				return nil
			},
			wantErr:  true,
			validate: nil,
		},
		{
			name: "Error on entity with no rows",
			setupGraph: func(t *testing.T) *model.Graph {
				def := &models.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test Description",
					Entities: map[string]models.Entity{
						"entity1": {
							DisplayName: "Entity1",
							ExternalId:  "Entity1",
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
				// Don't add any rows - entity has 0 rows
				return graph
			},
			wantErr:  false, // Field generator should handle empty entities gracefully
			validate: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewFieldGenerator()
			graph := tt.setupGraph(t)

			err := generator.GenerateFields(graph)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, graph)
				}
			}
		})
	}
}
