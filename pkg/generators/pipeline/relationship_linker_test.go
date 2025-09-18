package pipeline

import (
	"fmt"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/SGNL-ai/fabricator/pkg/parser"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelationshipLinker_LinkRelationships(t *testing.T) {
	tests := []struct {
		name            string
		setupGraph      func(t *testing.T) *model.Graph
		autoCardinality bool
		wantErr         bool
		validate        func(t *testing.T, graph *model.Graph)
	}{
		{
			name: "Link one-to-one relationship",
			setupGraph: func(t *testing.T) *model.Graph {
				def := &parser.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test Description",
					Entities: map[string]parser.Entity{
						"user": {
							DisplayName: "User",
							ExternalId:  "User",
							Attributes: []parser.Attribute{
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
								{Name: "profileId", ExternalId: "profileId", Type: "String"},
							},
						},
						"profile": {
							DisplayName: "Profile",
							ExternalId:  "Profile",
							Attributes: []parser.Attribute{
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
								{Name: "userId", ExternalId: "userId", Type: "String"},
							},
						},
					},
					Relationships: map[string]parser.Relationship{
						"user_profile": {
							DisplayName:   "User Profile",
							Name:          "user_profile",
							FromAttribute: "user.profileId", // Use entityID.externalId format
							ToAttribute:   "profile.id",     // Use entityID.externalId format
						},
					},
				}
				graphInterface, err := model.NewGraph(def)
				require.NoError(t, err)
				graph, ok := graphInterface.(*model.Graph)
				require.True(t, ok)

				// Pre-populate with proper relationship data
				entities := graph.GetAllEntities()
				userEntity := entities["user"]
				profileEntity := entities["profile"]

				// Add profiles first (target entities)
				for i := 0; i < 2; i++ {
					err := profileEntity.AddRow(model.NewRow(map[string]string{
						"id": "profile-" + string(rune('0'+i)),
					}))
					require.NoError(t, err)
				}

				// Add users with only IDs (no FK values yet - that's what the linker will add)
				for i := 0; i < 2; i++ {
					err := userEntity.AddRow(model.NewRow(map[string]string{
						"id": "user-" + string(rune('0'+i)),
						// profileId will be set by the relationship linker
					}))
					require.NoError(t, err)
				}
				return graph
			},
			autoCardinality: false,
			wantErr:         false,
			validate: func(t *testing.T, graph *model.Graph) {
				entities := graph.GetAllEntities()
				userEntity := entities["user"]

				// Check that FK values were set by the relationship linker
				userCSV := userEntity.ToCSV()
				require.Len(t, userCSV.Rows, 2, "Should have 2 user rows")

				// Find profileId column
				profileIdCol := -1
				for i, header := range userCSV.Headers {
					if header == "profileId" {
						profileIdCol = i
						break
					}
				}
				require.NotEqual(t, -1, profileIdCol, "Should find profileId column")

				// Verify FK values were set to valid profile IDs
				for _, row := range userCSV.Rows {
					fkValue := row[profileIdCol]
					assert.NotEmpty(t, fkValue, "Foreign key should be set")
					assert.Contains(t, []string{"profile-0", "profile-1"}, fkValue, "FK should reference valid profile ID")
				}
			},
		},
		{
			name: "Link with auto cardinality detection",
			setupGraph: func(t *testing.T) *model.Graph {
				def := &parser.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test Description",
					Entities: map[string]parser.Entity{
						"user": {
							DisplayName: "User",
							ExternalId:  "User",
							Attributes: []parser.Attribute{
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
								{Name: "profileId", ExternalId: "profileId", Type: "String", UniqueId: false}, // FK attribute, not unique
							},
						},
						"profile": {
							DisplayName: "Profile",
							ExternalId:  "Profile",
							Attributes: []parser.Attribute{
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
							},
						},
					},
					Relationships: map[string]parser.Relationship{
						"user_profile": {
							DisplayName:   "User Profile",
							Name:          "user_profile",
							FromAttribute: "user.profileId",
							ToAttribute:   "profile.id",
						},
					},
				}
				graphInterface, err := model.NewGraph(def)
				require.NoError(t, err)
				graph, ok := graphInterface.(*model.Graph)
				require.True(t, ok)

				entities := graph.GetAllEntities()
				userEntity := entities["user"]
				profileEntity := entities["profile"]

				// Add 3 profiles
				for i := 0; i < 3; i++ {
					err := profileEntity.AddRow(model.NewRow(map[string]string{
						"id": fmt.Sprintf("profile-%d", i),
					}))
					require.NoError(t, err)
				}

				// Add 3 users (should get unique profile assignments in 1:1 relationship)
				for i := 0; i < 3; i++ {
					err := userEntity.AddRow(model.NewRow(map[string]string{
						"id": fmt.Sprintf("user-%d", i),
					}))
					require.NoError(t, err)
				}

				return graph
			},
			autoCardinality: true, // Enable auto cardinality detection
			wantErr:         false,
			validate: func(t *testing.T, graph *model.Graph) {
				entities := graph.GetAllEntities()
				userEntity := entities["user"]

				// For 1:1 relationship with autoCardinality, each user should get unique profile
				csvData := userEntity.ToCSV()
				profileIdCol := -1
				for i, header := range csvData.Headers {
					if header == "profileId" {
						profileIdCol = i
						break
					}
				}
				require.NotEqual(t, -1, profileIdCol)

				// Check that all FK values are unique (1:1 cardinality)
				usedValues := make(map[string]bool)
				for _, row := range csvData.Rows {
					fkValue := row[profileIdCol]
					assert.False(t, usedValues[fkValue], "1:1 relationship should have unique FK values")
					usedValues[fkValue] = true
				}
			},
		},
		{
			name: "Handle valid relationship linking",
			setupGraph: func(t *testing.T) *model.Graph {
				def := &parser.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test Description",
					Entities: map[string]parser.Entity{
						"user": {
							DisplayName: "User",
							ExternalId:  "User",
							Attributes: []parser.Attribute{
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
								{Name: "roleId", ExternalId: "roleId", Type: "String"},
							},
						},
						"role": {
							DisplayName: "Role",
							ExternalId:  "Role",
							Attributes: []parser.Attribute{
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
							},
						},
					},
					Relationships: map[string]parser.Relationship{
						"user_role": {
							DisplayName:   "User Role",
							Name:          "user_role",
							FromAttribute: "user.roleId",
							ToAttribute:   "role.id",
						},
					},
				}
				graphInterface, err := model.NewGraph(def)
				require.NoError(t, err)
				graph, ok := graphInterface.(*model.Graph)
				require.True(t, ok)

				entities := graph.GetAllEntities()
				userEntity := entities["user"]
				roleEntity := entities["role"]

				// Add valid roles first
				err = roleEntity.AddRow(model.NewRow(map[string]string{
					"id": "role-1",
				}))
				require.NoError(t, err)

				// Add user with no FK (linker will set it)
				err = userEntity.AddRow(model.NewRow(map[string]string{
					"id": "user-1",
					// roleId will be set by relationship linker
				}))
				require.NoError(t, err)

				return graph
			},
			autoCardinality: false,
			wantErr:         false,
			validate: func(t *testing.T, graph *model.Graph) {
				entities := graph.GetAllEntities()
				userEntity := entities["user"]

				// Verify FK was set correctly
				csvData := userEntity.ToCSV()
				require.Len(t, csvData.Rows, 1)

				// Find roleId column
				roleIdCol := -1
				for i, header := range csvData.Headers {
					if header == "roleId" {
						roleIdCol = i
						break
					}
				}
				require.NotEqual(t, -1, roleIdCol)

				// Verify FK points to valid role
				fkValue := csvData.Rows[0][roleIdCol]
				assert.Equal(t, "role-1", fkValue, "Should link to valid role ID")
			},
		},
		{
			name: "Error on nil graph",
			setupGraph: func(t *testing.T) *model.Graph {
				return nil
			},
			autoCardinality: false,
			wantErr:         true,
			validate:        nil,
		},
		{
			name: "Handle entity with no target data for relationships",
			setupGraph: func(t *testing.T) *model.Graph {
				def := &parser.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test Description",
					Entities: map[string]parser.Entity{
						"user": {
							DisplayName: "User",
							ExternalId:  "User",
							Attributes: []parser.Attribute{
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
								{Name: "roleId", ExternalId: "roleId", Type: "String"},
							},
						},
						"role": {
							DisplayName: "Role",
							ExternalId:  "Role",
							Attributes: []parser.Attribute{
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
							},
						},
					},
					Relationships: map[string]parser.Relationship{
						"user_role": {
							DisplayName:   "User Role",
							Name:          "user_role",
							FromAttribute: "user.roleId",
							ToAttribute:   "role.id",
						},
					},
				}
				graphInterface, err := model.NewGraph(def)
				require.NoError(t, err)
				graph, ok := graphInterface.(*model.Graph)
				require.True(t, ok)

				// Add user but NO roles (empty target entity)
				entities := graph.GetAllEntities()
				userEntity := entities["user"]

				err = userEntity.AddRow(model.NewRow(map[string]string{
					"id": "user-1",
				}))
				require.NoError(t, err)

				// Role entity has no data - relationship linker should handle gracefully
				return graph
			},
			autoCardinality: false,
			wantErr:         false, // Should handle gracefully, not error
			validate: func(t *testing.T, graph *model.Graph) {
				entities := graph.GetAllEntities()
				userEntity := entities["user"]
				roleEntity := entities["role"]

				// User should still have its row
				assert.Equal(t, 1, userEntity.GetRowCount())
				// Role should be empty
				assert.Equal(t, 0, roleEntity.GetRowCount())

				// User's roleId should remain empty (no target data to link to)
				csvData := userEntity.ToCSV()
				assert.Equal(t, "", csvData.Rows[0][1], "roleId should remain empty when no target data")
			},
		},
		{
			name: "Handle missing relationship attributes",
			setupGraph: func(t *testing.T) *model.Graph {
				def := &parser.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test Description",
					Entities: map[string]parser.Entity{
						"user": {
							DisplayName: "User",
							ExternalId:  "User",
							Attributes: []parser.Attribute{
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
								// Missing roleId attribute that relationship expects
							},
						},
						"role": {
							DisplayName: "Role",
							ExternalId:  "Role",
							Attributes: []parser.Attribute{
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
							},
						},
					},
					Relationships: map[string]parser.Relationship{
						"user_role": {
							DisplayName:   "User Role",
							Name:          "user_role",
							FromAttribute: "roleId", // This attribute doesn't exist in user entity
							ToAttribute:   "id",
						},
					},
				}
				graphInterface, err := model.NewGraph(def)
				if err != nil {
					// Graph creation should fail with missing attribute
					return nil
				}
				graph, ok := graphInterface.(*model.Graph)
				if !ok {
					return nil
				}
				return graph
			},
			autoCardinality: false,
			wantErr:         true, // Should error because attribute doesn't exist
			validate: func(t *testing.T, graph *model.Graph) {
				// Should handle missing attributes gracefully
				entities := graph.GetAllEntities()
				require.Len(t, entities, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linker := NewRelationshipLinker()
			graph := tt.setupGraph(t)

			err := linker.LinkRelationships(graph, tt.autoCardinality)

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
