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
							FromAttribute: "User.profileId", // Use entityID.externalId format
							ToAttribute:   "Profile.id",     // Use entityID.externalId format
						},
					},
				}
				graphInterface, err := model.NewGraph(def, 100)
				require.NoError(t, err)
				graph, ok := graphInterface.(*model.Graph)
				require.True(t, ok)

				// Pre-populate with proper relationship data
				entities := graph.GetAllEntities()
				userEntity := entities["User"]
				profileEntity := entities["Profile"]

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
				userEntity := entities["User"]
				profileEntity := entities["Profile"]

				// Verify row counts after relationship linking
				assert.Equal(t, 2, userEntity.GetRowCount(), "User entity should still have 2 rows after linking")
				assert.Equal(t, 2, profileEntity.GetRowCount(), "Profile entity should still have 2 rows after linking")

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
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true}, // PK
							},
						},
						"employee": {
							DisplayName: "Employee",
							ExternalId:  "Employee",
							Attributes: []parser.Attribute{
								{Name: "user_id", ExternalId: "user_id", Type: "String", UniqueId: true}, // Unique FK for true 1:1
							},
						},
					},
					Relationships: map[string]parser.Relationship{
						"employee_user": {
							DisplayName:   "Employee User",
							Name:          "employee_user",
							FromAttribute: "Employee.user_id", // Unique FK
							ToAttribute:   "User.id",          // Unique PK
						},
					},
				}
				graphInterface, err := model.NewGraph(def, 100)
				require.NoError(t, err)
				graph, ok := graphInterface.(*model.Graph)
				require.True(t, ok)

				entities := graph.GetAllEntities()
				userEntity := entities["User"]
				employeeEntity := entities["Employee"]

				// Add 3 users
				for i := 0; i < 3; i++ {
					err := userEntity.AddRow(model.NewRow(map[string]string{
						"id": fmt.Sprintf("user-%d", i),
					}))
					require.NoError(t, err)
				}

				// Add 3 employees (will get unique user assignments in 1:1 relationship)
				for i := 0; i < 3; i++ {
					err := employeeEntity.AddRow(model.NewRow(map[string]string{
						"user_id": fmt.Sprintf("temp-emp-%d", i), // Temporary value, will be overwritten by relationship linker
					}))
					require.NoError(t, err)
				}

				return graph
			},
			autoCardinality: true, // Enable auto cardinality detection
			wantErr:         false,
			validate: func(t *testing.T, graph *model.Graph) {
				entities := graph.GetAllEntities()
				userEntity := entities["User"]
				employeeEntity := entities["Employee"]

				// Verify row counts after relationship linking
				assert.Equal(t, 3, userEntity.GetRowCount(), "User entity should still have 3 rows after linking")
				assert.Equal(t, 3, employeeEntity.GetRowCount(), "Employee entity should still have 3 rows after linking")

				// Directly inspect entity rows for FK values
				require.Equal(t, 3, employeeEntity.GetRowCount(), "Should have 3 employee rows")

				// Check FK values directly from entity rows
				usedValues := make(map[string]bool)
				for i := 0; i < employeeEntity.GetRowCount(); i++ {
					row := employeeEntity.GetRowByIndex(i)
					require.NotNil(t, row, "Row %d should exist", i)

					fkValue := row.GetValue("user_id")
					assert.NotEmpty(t, fkValue, "FK should be set for employee %d", i)
					assert.False(t, usedValues[fkValue], "1:1 relationship should have unique FK values")
					usedValues[fkValue] = true
				}

				// Should have used 3 unique user IDs
				assert.Len(t, usedValues, 3, "Should have 3 unique FK values for 1:1 relationship")
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
							FromAttribute: "User.roleId",
							ToAttribute:   "Role.id",
						},
					},
				}
				graphInterface, err := model.NewGraph(def, 100)
				require.NoError(t, err)
				graph, ok := graphInterface.(*model.Graph)
				require.True(t, ok)

				entities := graph.GetAllEntities()
				userEntity := entities["User"]
				roleEntity := entities["Role"]

				// Add valid roles first
				err = roleEntity.AddRow(model.NewRow(map[string]string{
					"id": "role-1",
				}))
				require.NoError(t, err)

				// Add user with empty FK field (linker will set it)
				err = userEntity.AddRow(model.NewRow(map[string]string{
					"id":     "user-1",
					"roleId": "xxx", // Empty FK field that linker will populate
				}))
				require.NoError(t, err)

				// Verify relationship exists in graph
				relationships := graph.GetAllRelationships()
				assert.Len(t, relationships, 1, "Graph should have 1 relationship")

				// Verify User entity has the relationship
				userRelationships := graph.GetRelationshipsForEntity("User")
				assert.Len(t, userRelationships, 1, "User entity should have 1 relationship")

				// Verify relationship details
				rel := relationships[0]
				assert.Equal(t, "User", rel.GetSourceEntity().GetID(), "Source entity should be user")
				assert.Equal(t, "Role", rel.GetTargetEntity().GetID(), "Target entity should be role")
				assert.NotNil(t, rel.GetSourceAttribute(), "Source attribute should not be nil")
				assert.NotNil(t, rel.GetTargetAttribute(), "Target attribute should not be nil")
				assert.Equal(t, "roleId", rel.GetSourceAttribute().GetName(), "Source attribute should be roleId")
				assert.Equal(t, "id", rel.GetTargetAttribute().GetName(), "Target attribute should be id")

				// Verify initial FK value by directly inspecting row
				userRow := userEntity.GetRowByIndex(0)
				require.NotNil(t, userRow, "User row should exist")
				initialFK := userRow.GetValue("roleId")
				assert.Equal(t, "xxx", initialFK, "Initial FK should be xxx before linking")

				// CRITICAL: Final state verification before relationship linking
				assert.Equal(t, 1, userEntity.GetRowCount(), "User entity should have 1 row")
				assert.Equal(t, 1, roleEntity.GetRowCount(), "Role entity should have 1 row")
				assert.Equal(t, 1, rel.GetSourceEntity().GetRowCount(), "Relationship's source entity should have 1 row")
				assert.Equal(t, 1, rel.GetTargetEntity().GetRowCount(), "Relationship's target entity should have 1 row")

				// CRITICAL: Verify relationship points to same entity instances
				assert.Equal(t, userEntity, rel.GetSourceEntity(), "Relationship should point to same User entity instance")
				assert.Equal(t, roleEntity, rel.GetTargetEntity(), "Relationship should point to same Role entity instance")

				return graph
			},
			autoCardinality: false,
			wantErr:         false,
			validate: func(t *testing.T, graph *model.Graph) {
				entities := graph.GetAllEntities()
				userEntity := entities["User"]

				// Directly inspect entity row for FK value
				require.Equal(t, 1, userEntity.GetRowCount(), "Should have 1 user row")

				row := userEntity.GetRowByIndex(0)
				require.NotNil(t, row, "User row should exist")

				fkValue := row.GetValue("roleId")
				assert.NotEmpty(t, fkValue, "FK should be set to some role ID")
				assert.Equal(t, "role-1", fkValue, "Should link to role-1 (only available role)")
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
							FromAttribute: "User.roleId",
							ToAttribute:   "Role.id",
						},
					},
				}
				graphInterface, err := model.NewGraph(def, 100)
				require.NoError(t, err)
				graph, ok := graphInterface.(*model.Graph)
				require.True(t, ok)

				// Add user but NO roles (empty target entity)
				entities := graph.GetAllEntities()
				userEntity := entities["User"]

				err = userEntity.AddRow(model.NewRow(map[string]string{
					"id": "user-1",
				}))
				require.NoError(t, err)

				// Role entity has no data - relationship linker should handle gracefully
				return graph
			},
			autoCardinality: false,
			wantErr:         true, // Should error with explicit empty target entity message
			validate: func(t *testing.T, graph *model.Graph) {
				entities := graph.GetAllEntities()
				userEntity := entities["User"]
				roleEntity := entities["Role"]

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
				graphInterface, err := model.NewGraph(def, 100)
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
