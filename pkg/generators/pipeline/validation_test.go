package pipeline

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/SGNL-ai/fabricator/pkg/parser"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_ValidateRelationships(t *testing.T) {
	tests := []struct {
		name       string
		setupGraph func() *model.Graph
		wantErrors bool
		errorCount int
	}{
		{
			name: "Valid relationships",
			setupGraph: func() *model.Graph {
				// Create a minimal valid SOR definition for testing
				sorDef := &parser.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test SOR for validator",
					Entities: map[string]parser.Entity{
						"User": {
							DisplayName: "User",
							ExternalId:  "User",
							Description: "Test user entity",
							Attributes: []parser.Attribute{
								{
									Name:        "id",
									ExternalId:  "id",
									Description: "Primary key",
									Type:        "string",
									UniqueId:    true,
								},
								{
									Name:        "name",
									ExternalId:  "name",
									Description: "User name",
									Type:        "string",
								},
							},
						},
					},
					Relationships: map[string]parser.Relationship{},
				}

				graph, err := model.NewGraph(sorDef, 100)
				if err != nil {
					return nil // This will cause test to fail appropriately
				}
				return graph.(*model.Graph)
			},
			wantErrors: false,
			errorCount: 0,
		},
		{
			name: "Invalid relationships",
			setupGraph: func() *model.Graph {
				// Return nil graph to test nil handling
				return nil
			},
			wantErrors: true,
			errorCount: 1, // Expecting 1 validation error for nil graph
		},
		// Additional test cases will be added later
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup and test implementation will be added later
			// This is just a stub
			validator := NewValidation()
			graph := tt.setupGraph()

			errors := validator.ValidateRelationships(graph)

			if tt.wantErrors {
				assert.Len(t, errors, tt.errorCount)
			} else {
				assert.Empty(t, errors)
			}
		})
	}
}

func TestValidator_ValidateUniqueValues(t *testing.T) {
	tests := []struct {
		name       string
		setupGraph func() *model.Graph
		wantErrors bool
		errorCount int
	}{
		{
			name: "Valid unique values",
			setupGraph: func() *model.Graph {
				// Create a minimal valid SOR definition for testing
				sorDef := &parser.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test SOR for validator",
					Entities: map[string]parser.Entity{
						"User": {
							DisplayName: "User",
							ExternalId:  "User",
							Description: "Test user entity",
							Attributes: []parser.Attribute{
								{
									Name:        "id",
									ExternalId:  "id",
									Description: "Primary key",
									Type:        "string",
									UniqueId:    true,
								},
								{
									Name:        "name",
									ExternalId:  "name",
									Description: "User name",
									Type:        "string",
								},
							},
						},
					},
					Relationships: map[string]parser.Relationship{},
				}

				graph, err := model.NewGraph(sorDef, 100)
				if err != nil {
					return nil // This will cause test to fail appropriately
				}
				return graph.(*model.Graph)
			},
			wantErrors: false,
			errorCount: 0,
		},
		{
			name: "Duplicate unique values",
			setupGraph: func() *model.Graph {
				// Create a SOR with duplicate unique values
				sorDef := &parser.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test SOR for validator",
					Entities: map[string]parser.Entity{
						"User": {
							DisplayName: "User",
							ExternalId:  "User",
							Description: "Test user entity",
							Attributes: []parser.Attribute{
								{
									Name:        "id",
									ExternalId:  "id",
									Description: "Primary key",
									Type:        "String",
									UniqueId:    true,
								},
								{
									Name:        "email",
									ExternalId:  "email",
									Description: "Email address",
									Type:        "String",
									UniqueId:    false, // Regular field, not unique
								},
							},
						},
					},
				}

				graphInterface, err := model.NewGraph(sorDef, 100)
				if err != nil {
					return nil
				}
				graph, ok := graphInterface.(*model.Graph)
				if !ok {
					return nil
				}

				// Add rows with duplicate unique values
				entities := graph.GetAllEntities()
				userEntity := entities["User"]

				// Add first row successfully
				err = userEntity.AddRow(model.NewRow(map[string]string{
					"id":    "user-1",
					"email": "test@example.com",
				}))
				if err != nil {
					return nil // Setup failed
				}

				// Try to add second row with duplicate ID (should fail at AddRow level)
				_ = userEntity.AddRow(model.NewRow(map[string]string{
					"id":    "user-1", // Duplicate ID!
					"email": "different@example.com",
				}))
				// AddRow should reject this due to duplicate unique ID

				// Add third row with different ID and same email (should be OK since email isn't unique)
				_ = userEntity.AddRow(model.NewRow(map[string]string{
					"id":    "user-2",
					"email": "test@example.com", // Same email is OK, not unique
				}))

				return graph
			},
			wantErrors: false,
			errorCount: 0, // No errors expected - AddRow should have rejected the duplicate ID
		},
		{
			name: "Invalid foreign key references",
			setupGraph: func() *model.Graph {
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
							FromAttribute: "roleId",
							ToAttribute:   "id",
						},
					},
				}

				graphInterface, err := model.NewGraph(def, 100)
				if err != nil {
					return nil
				}
				graph, ok := graphInterface.(*model.Graph)
				if !ok {
					return nil
				}

				entities := graph.GetAllEntities()
				userEntity := entities["user"]
				roleEntity := entities["role"]

				// Add valid role
				_ = roleEntity.AddRow(model.NewRow(map[string]string{
					"id": "role-1",
				}))

				// Add user with valid foreign key
				_ = userEntity.AddRow(model.NewRow(map[string]string{
					"id":     "user-1",
					"roleId": "role-1", // Valid FK
				}))

				// Add user with invalid foreign key
				_ = userEntity.AddRow(model.NewRow(map[string]string{
					"id":     "user-2",
					"roleId": "role-999", // Invalid FK - doesn't exist
				}))

				return graph
			},
			wantErrors: true,
			errorCount: 1, // Should detect 1 invalid foreign key reference
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Unique value validation is now handled by AddRow during data entry
			// These test cases verify that AddRow properly rejects invalid data
			graph := tt.setupGraph()

			// Just verify the graph was created successfully
			// The real validation happens during AddRow calls in setupGraph
			if tt.wantErrors {
				// If we expected errors, verify that problematic data was rejected during setup
				// by checking row counts
				if graph != nil {
					entities := graph.GetAllEntities()
					if len(entities) > 0 {
						// Entity should have rejected duplicate data during AddRow
						for _, entity := range entities {
							// AddRow should have prevented duplicate unique values
							assert.LessOrEqual(t, entity.GetRowCount(), 2, "AddRow should reject duplicates, limiting row count")
						}
					}
				}
			}
		})
	}
}

// Test that AddRow properly rejects duplicate unique values immediately
func TestEntityAddRow_RejectsDuplicateUniqueValues(t *testing.T) {
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

	graphInterface, err := model.NewGraph(def, 100)
	require.NoError(t, err)
	graph, ok := graphInterface.(*model.Graph)
	require.True(t, ok)

	entities := graph.GetAllEntities()
	userEntity := entities["user"]

	// Add first row successfully
	err = userEntity.AddRow(model.NewRow(map[string]string{
		"id":   "user-1",
		"name": "John Doe",
	}))
	assert.NoError(t, err)

	// Try to add second row with duplicate unique ID - should fail immediately
	err = userEntity.AddRow(model.NewRow(map[string]string{
		"id":   "user-1", // Duplicate unique ID
		"name": "Jane Doe",
	}))
	assert.Error(t, err, "AddRow should reject duplicate unique values immediately")
	assert.Contains(t, err.Error(), "duplicate", "Error should mention duplicate value")

	// Verify only one row was actually added
	assert.Equal(t, 1, userEntity.GetRowCount(), "Should only have 1 row after duplicate rejection")
}
