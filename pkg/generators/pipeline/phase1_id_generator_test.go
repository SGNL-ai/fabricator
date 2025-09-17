package pipeline

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIDGenerator_GenerateIDs(t *testing.T) {
	tests := []struct {
		name       string
		setupGraph func() *model.Graph
		dataVolume int
		wantErr    bool
		validate   func(t *testing.T, graph *model.Graph)
	}{
		{
			name: "Generate IDs for single entity",
			setupGraph: func() *model.Graph {
				def := &models.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test Description",
					Entities: map[string]models.Entity{
						"entity1": {
							DisplayName: "Entity1",
							ExternalId:  "Entity1",
							Description: "Test entity",
							Attributes: []models.Attribute{
								{
									Name:       "id",
									ExternalId: "id",
									Type:       "String",
									UniqueId:   true,
								},
								{
									Name:       "name",
									ExternalId: "name",
									Type:       "String",
								},
							},
						},
					},
				}
				graphInterface, err := model.NewGraph(def)
				require.NoError(t, err)
				graph, ok := graphInterface.(*model.Graph)
				require.True(t, ok)
				return graph
			},
			dataVolume: 3,
			wantErr:    false,
			validate: func(t *testing.T, graph *model.Graph) {
				entities := graph.GetAllEntities()
				require.Len(t, entities, 1)

				// Get the entity (we know there's only one)
				var entity model.EntityInterface
				for _, e := range entities {
					entity = e
					break
				}
				require.NotNil(t, entity)

				// Check that the entity has the expected number of rows
				assert.Equal(t, 3, entity.GetRowCount(), "Entity should have 3 rows after ID generation")

				// Check that unique IDs are actually unique
				csvData := entity.ToCSV()
				require.NotEmpty(t, csvData.Headers, "CSV should have headers")
				require.NotEmpty(t, csvData.Rows, "CSV should have data rows")

				idValues := make(map[string]bool)
				idColumnIndex := -1

				// Find the ID column
				for i, header := range csvData.Headers {
					if header == "id" {
						idColumnIndex = i
						break
					}
				}
				require.NotEqual(t, -1, idColumnIndex, "Should find 'id' column in headers")

				// Check that all ID values are unique
				for _, row := range csvData.Rows {
					require.Greater(t, len(row), idColumnIndex, "Row should have enough columns")
					idValue := row[idColumnIndex]
					assert.NotEmpty(t, idValue, "ID value should not be empty")
					assert.False(t, idValues[idValue], "ID values must be unique, found duplicate: %s", idValue)
					idValues[idValue] = true
				}
			},
		},
		{
			name: "Generate IDs with multiple entities",
			setupGraph: func() *model.Graph {
				def := &models.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test Description",
					Entities: map[string]models.Entity{
						"user": {
							DisplayName: "User",
							ExternalId:  "User",
							Attributes: []models.Attribute{
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
								{Name: "name", ExternalId: "name", Type: "String"},
							},
						},
						"role": {
							DisplayName: "Role",
							ExternalId:  "Role",
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
				return graph
			},
			dataVolume: 5,
			wantErr:    false,
			validate: func(t *testing.T, graph *model.Graph) {
				entities := graph.GetAllEntities()
				require.Len(t, entities, 2)

				// Check both entities have the correct number of rows
				for entityID, entity := range entities {
					assert.Equal(t, 5, entity.GetRowCount(), "Entity %s should have 5 rows", entityID)

					// Verify unique IDs across all rows in this entity
					csvData := entity.ToCSV()
					idValues := make(map[string]bool)
					for _, row := range csvData.Rows {
						idValue := row[0] // First column is ID
						assert.False(t, idValues[idValue], "Entity %s should have unique IDs, found duplicate: %s", entityID, idValue)
						idValues[idValue] = true
					}
				}
			},
		},
		{
			name: "Error on zero data volume",
			setupGraph: func() *model.Graph {
				def := &models.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test Description",
					Entities: map[string]models.Entity{
						"entity1": {
							DisplayName: "Entity1",
							ExternalId:  "Entity1",
							Attributes: []models.Attribute{
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
							},
						},
					},
				}
				graphInterface, err := model.NewGraph(def)
				require.NoError(t, err)
				graph, ok := graphInterface.(*model.Graph)
				require.True(t, ok)
				return graph
			},
			dataVolume: 0,
			wantErr:    true,
			validate:   nil,
		},
		{
			name: "Error on nil graph",
			setupGraph: func() *model.Graph {
				return nil
			},
			dataVolume: 5,
			wantErr:    true,
			validate:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewIDGenerator()
			graph := tt.setupGraph()

			err := generator.GenerateIDs(graph, tt.dataVolume)

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
