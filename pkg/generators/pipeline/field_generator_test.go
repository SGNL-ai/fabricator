package pipeline

import (
	"fmt"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/SGNL-ai/fabricator/pkg/parser"

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
				def := &parser.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test Description",
					Entities: map[string]parser.Entity{
						"entity1": {
							DisplayName: "Entity1",
							ExternalId:  "Entity1",
							Attributes: []parser.Attribute{
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
								{Name: "name", ExternalId: "name", Type: "String"},
								{Name: "email", ExternalId: "email", Type: "String"},
							},
						},
					},
				}
				graphInterface, err := model.NewGraph(def, 100)
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
				def := &parser.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test Description",
					Entities: map[string]parser.Entity{
						"entity1": {
							DisplayName: "Entity1",
							ExternalId:  "Entity1",
							Attributes: []parser.Attribute{
								{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
								{Name: "age", ExternalId: "age", Type: "Integer"},
								{Name: "isActive", ExternalId: "isActive", Type: "Boolean"},
								{Name: "createdAt", ExternalId: "createdAt", Type: "Date"},
							},
						},
					},
				}
				graphInterface, err := model.NewGraph(def, 100)
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
				def := &parser.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test Description",
					Entities: map[string]parser.Entity{
						"entity1": {
							DisplayName: "Entity1",
							ExternalId:  "Entity1",
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
				// Don't add any rows - entity has 0 rows
				return graph
			},
			wantErr:  false, // Field generator should handle empty entities gracefully
			validate: nil,
		},
		{
			name: "Should not generate fields for relationship attributes",
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
								{Name: "name", ExternalId: "name", Type: "String"},
								{Name: "profile_id", ExternalId: "profile_id", Type: "String"}, // This should become a relationship attribute
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
							FromAttribute: "user.profile_id", // This should mark profile_id as a relationship attribute
							ToAttribute:   "profile.id",
						},
					},
				}
				graphInterface, err := model.NewGraph(def, 100)
				require.NoError(t, err)
				graph, ok := graphInterface.(*model.Graph)
				require.True(t, ok)

				// Pre-populate with rows
				entities := graph.GetAllEntities()
				userEntity := entities["user"]
				profileEntity := entities["profile"]

				// Add profile first
				err = profileEntity.AddRow(model.NewRow(map[string]string{
					"id": "profile-1",
				}))
				require.NoError(t, err)

				// Add user with FK value already set by relationship linker (simulate)
				err = userEntity.AddRow(model.NewRow(map[string]string{
					"id":         "user-1",
					"profile_id": "profile-1", // FK value set by relationship linker
					// name field should be empty and filled by field generator
				}))
				require.NoError(t, err)

				return graph
			},
			wantErr: false,
			validate: func(t *testing.T, graph *model.Graph) {
				entities := graph.GetAllEntities()
				userEntity := entities["user"]

				// Verify that field generator filled non-relationship fields but left FK alone
				csvData := userEntity.ToCSV()
				require.Len(t, csvData.Rows, 1)
				row := csvData.Rows[0]

				// Find column indices
				idCol, nameCol, profileIdCol := -1, -1, -1
				for i, header := range csvData.Headers {
					switch header {
					case "id":
						idCol = i
					case "name":
						nameCol = i
					case "profile_id":
						profileIdCol = i
					}
				}

				require.NotEqual(t, -1, idCol, "Should have id column")
				require.NotEqual(t, -1, nameCol, "Should have name column")
				require.NotEqual(t, -1, profileIdCol, "Should have profile_id column")

				// ID should remain as set
				assert.Equal(t, "user-1", row[idCol], "ID should remain unchanged")

				// FK should remain as set by relationship linker (NOT overwritten by field generator)
				assert.Equal(t, "profile-1", row[profileIdCol], "FK should NOT be overwritten by field generator")

				// Regular field should be generated
				assert.NotEmpty(t, row[nameCol], "Name should be generated by field generator")
				assert.NotEqual(t, "profile-1", row[nameCol], "Name should not be FK value")
			},
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

func TestFieldGenerator_generateFieldValue(t *testing.T) {
	generator := &FieldGenerator{}

	tests := []struct {
		name             string
		attrName         string
		dataType         string
		expectedContains string // What the result should contain or match
		validateFunc     func(t *testing.T, result string)
	}{
		{
			name:         "email field by name",
			attrName:     "email",
			dataType:     "String",
			validateFunc: func(t *testing.T, result string) {
				assert.Contains(t, result, "@", "Email should contain @")
			},
		},
		{
			name:         "user_email field by name",
			attrName:     "user_email",
			dataType:     "String",
			validateFunc: func(t *testing.T, result string) {
				assert.Contains(t, result, "@", "Email field should contain @")
			},
		},
		{
			name:         "name field by name",
			attrName:     "name",
			dataType:     "String",
			validateFunc: func(t *testing.T, result string) {
				assert.NotEmpty(t, result, "Name should not be empty")
				assert.True(t, len(result) > 1, "Name should have reasonable length")
			},
		},
		{
			name:         "full_name field by name",
			attrName:     "full_name",
			dataType:     "String",
			validateFunc: func(t *testing.T, result string) {
				assert.NotEmpty(t, result, "Name should not be empty")
			},
		},
		{
			name:         "phone field by name",
			attrName:     "phone",
			dataType:     "String",
			validateFunc: func(t *testing.T, result string) {
				assert.NotEmpty(t, result, "Phone should not be empty")
				// Phone numbers typically contain digits or common characters
				assert.True(t, len(result) >= 10, "Phone should have reasonable length")
			},
		},
		{
			name:         "phone_number field by name",
			attrName:     "phone_number",
			dataType:     "String",
			validateFunc: func(t *testing.T, result string) {
				assert.NotEmpty(t, result, "Phone should not be empty")
			},
		},
		{
			name:         "address field by name",
			attrName:     "address",
			dataType:     "String",
			validateFunc: func(t *testing.T, result string) {
				assert.NotEmpty(t, result, "Address should not be empty")
			},
		},
		{
			name:         "home_address field by name",
			attrName:     "home_address",
			dataType:     "String",
			validateFunc: func(t *testing.T, result string) {
				assert.NotEmpty(t, result, "Address should not be empty")
			},
		},
		{
			name:         "status field by name",
			attrName:     "status",
			dataType:     "String",
			validateFunc: func(t *testing.T, result string) {
				validStatuses := []string{"active", "inactive", "pending"}
				assert.Contains(t, validStatuses, result, "Status should be one of the predefined values")
			},
		},
		{
			name:         "user_status field by name",
			attrName:     "user_status",
			dataType:     "String",
			validateFunc: func(t *testing.T, result string) {
				validStatuses := []string{"active", "inactive", "pending"}
				assert.Contains(t, validStatuses, result, "Status should be one of the predefined values")
			},
		},
		{
			name:         "date field by name",
			attrName:     "date",
			dataType:     "String",
			validateFunc: func(t *testing.T, result string) {
				assert.NotEmpty(t, result, "Date should not be empty")
				// Should be in RFC3339 format
				assert.Contains(t, result, "T", "Date should contain time separator")
			},
		},
		{
			name:         "created_date field by name",
			attrName:     "created_date",
			dataType:     "String",
			validateFunc: func(t *testing.T, result string) {
				assert.NotEmpty(t, result, "Date should not be empty")
			},
		},
		{
			name:         "time field by name",
			attrName:     "time",
			dataType:     "String",
			validateFunc: func(t *testing.T, result string) {
				assert.NotEmpty(t, result, "Time should not be empty")
			},
		},
		{
			name:         "created_time field by name",
			attrName:     "created_time",
			dataType:     "String",
			validateFunc: func(t *testing.T, result string) {
				assert.NotEmpty(t, result, "Time should not be empty")
			},
		},
		{
			name:         "Integer data type",
			attrName:     "count",
			dataType:     "Integer",
			validateFunc: func(t *testing.T, result string) {
				// Should be a valid integer
				num := 0
				_, err := fmt.Sscanf(result, "%d", &num)
				assert.NoError(t, err, "Should be a valid integer")
				assert.True(t, num >= 1 && num <= 1000, "Integer should be in expected range")
			},
		},
		{
			name:         "Int64 data type",
			attrName:     "big_number",
			dataType:     "Int64",
			validateFunc: func(t *testing.T, result string) {
				num := 0
				_, err := fmt.Sscanf(result, "%d", &num)
				assert.NoError(t, err, "Should be a valid integer")
			},
		},
		{
			name:         "Boolean data type",
			attrName:     "flag",
			dataType:     "Boolean",
			validateFunc: func(t *testing.T, result string) {
				assert.True(t, result == "true" || result == "false", "Should be valid boolean string")
			},
		},
		{
			name:         "Bool data type",
			attrName:     "enabled",
			dataType:     "Bool",
			validateFunc: func(t *testing.T, result string) {
				assert.True(t, result == "true" || result == "false", "Should be valid boolean string")
			},
		},
		{
			name:         "Date data type",
			attrName:     "birth_date",
			dataType:     "Date",
			validateFunc: func(t *testing.T, result string) {
				assert.NotEmpty(t, result, "Date should not be empty")
				// Should be in YYYY-MM-DD format
				assert.Regexp(t, `\d{4}-\d{2}-\d{2}`, result, "Date should be in YYYY-MM-DD format")
			},
		},
		{
			name:         "DateTime data type",
			attrName:     "timestamp",
			dataType:     "DateTime",
			validateFunc: func(t *testing.T, result string) {
				assert.NotEmpty(t, result, "DateTime should not be empty")
				assert.Contains(t, result, "T", "DateTime should contain time separator")
			},
		},
		{
			name:         "Float data type",
			attrName:     "rate",
			dataType:     "Float",
			validateFunc: func(t *testing.T, result string) {
				var f float64
				_, err := fmt.Sscanf(result, "%f", &f)
				assert.NoError(t, err, "Should be a valid float")
				assert.True(t, f >= 1.0 && f <= 100.0, "Float should be in expected range")
			},
		},
		{
			name:         "Double data type",
			attrName:     "precision",
			dataType:     "Double",
			validateFunc: func(t *testing.T, result string) {
				var f float64
				_, err := fmt.Sscanf(result, "%f", &f)
				assert.NoError(t, err, "Should be a valid double")
			},
		},
		{
			name:         "Unknown data type defaults to word",
			attrName:     "random_field",
			dataType:     "UnknownType",
			validateFunc: func(t *testing.T, result string) {
				assert.NotEmpty(t, result, "Should generate some word")
				assert.True(t, len(result) > 0, "Should have some content")
			},
		},
		{
			name:         "String data type defaults to word",
			attrName:     "description",
			dataType:     "String",
			validateFunc: func(t *testing.T, result string) {
				assert.NotEmpty(t, result, "Should generate some word")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test attribute by creating a full entity and extracting an attribute
			// This is the proper way since newAttribute is not exported
			def := &parser.SORDefinition{
				DisplayName: "Test SOR",
				Description: "Test Description",
				Entities: map[string]parser.Entity{
					"test_entity": {
						DisplayName: "TestEntity",
						ExternalId:  "TestEntity",
						Attributes: []parser.Attribute{
							{
								Name:       "id",
								ExternalId: "id",
								Type:       "String",
								UniqueId:   true, // Every entity needs a unique ID
							},
							{
								Name:       tt.attrName,
								ExternalId: tt.attrName,
								Type:       tt.dataType,
							},
						},
					},
				},
			}

			graphInterface, err := model.NewGraph(def, 100)
			require.NoError(t, err)
			graph, ok := graphInterface.(*model.Graph)
			require.True(t, ok)

			// Get the entity and its attribute
			entities := graph.GetAllEntities()
			require.Len(t, entities, 1)

			var entity model.EntityInterface
			for _, e := range entities {
				entity = e
				break
			}

			attrs := entity.GetAttributes()
			require.Len(t, attrs, 2) // id + the test attribute
			testAttr := attrs[1]     // Get the test attribute, not the ID

			result := generator.generateFieldValue(testAttr)
			tt.validateFunc(t, result)
		})
	}
}

func TestFieldGenerator_contains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"exact match", "email", "email", true},
		{"prefix match", "email_address", "email", true},
		{"suffix match", "user_email", "email", true},
		{"no match", "username", "email", false},
		{"empty string", "", "email", false},
		{"empty substring", "email", "", true},
		{"substring longer than string", "em", "email", false},
		{"case sensitive - no match", "Email", "email", false},
		{"middle match should not match", "myemailfield", "email", false}, // This function only checks prefix/suffix
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}
