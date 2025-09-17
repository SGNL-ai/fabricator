package pipeline

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestValidator_ValidateRelationships(t *testing.T) {
	tests := []struct {
		name        string
		setupGraph  func() *model.Graph
		wantErrors  bool
		errorCount  int
	}{
		{
			name: "Valid relationships",
			setupGraph: func() *model.Graph {
				// Create a minimal valid SOR definition for testing
				sorDef := &models.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test SOR for validator",
					Entities: map[string]models.Entity{
						"User": {
							DisplayName: "User",
							ExternalId:  "User",
							Description: "Test user entity",
							Attributes: []models.Attribute{
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
					Relationships: map[string]models.Relationship{},
				}

				graph, err := model.NewGraph(sorDef)
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
				// Will create a graph with invalid relationships
				return nil
			},
			wantErrors: true,
			errorCount: 2, // Expecting 2 validation errors
		},
		// Additional test cases will be added later
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup and test implementation will be added later
			// This is just a stub
			validator := NewValidator()
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
		name        string
		setupGraph  func() *model.Graph
		wantErrors  bool
		errorCount  int
	}{
		{
			name: "Valid unique values",
			setupGraph: func() *model.Graph {
				// Create a minimal valid SOR definition for testing
				sorDef := &models.SORDefinition{
					DisplayName: "Test SOR",
					Description: "Test SOR for validator",
					Entities: map[string]models.Entity{
						"User": {
							DisplayName: "User",
							ExternalId:  "User",
							Description: "Test user entity",
							Attributes: []models.Attribute{
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
					Relationships: map[string]models.Relationship{},
				}

				graph, err := model.NewGraph(sorDef)
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
				// Will create a graph with duplicate values in unique fields
				return nil
			},
			wantErrors: true,
			errorCount: 1, // Expecting 1 validation error
		},
		// Additional test cases will be added later
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup and test implementation will be added later
			// This is just a stub
			validator := NewValidator()
			graph := tt.setupGraph()
			
			errors := validator.ValidateUniqueValues(graph)
			
			if tt.wantErrors {
				assert.Len(t, errors, tt.errorCount)
			} else {
				assert.Empty(t, errors)
			}
		})
	}
}