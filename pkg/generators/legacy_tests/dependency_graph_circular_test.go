package generators

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCircularRelationshipDetection tests the handling of the circular relationship
// between Role and Assignment from the SW-Assertions-Only-0.1.0.yaml file
func TestCircularRelationshipDetection(t *testing.T) {
	// Disable color output for tests
	color.NoColor = true

	// Create a CSV generator for testing
	g := NewCSVGenerator("output", 10, false)

	// Create test entities that mirror the Role and Assignment entities from SW-Assertions
	entities := map[string]models.Entity{
		"Role": {
			DisplayName: "Role",
			ExternalId:  "Role",
			Description: "Role entity for testing",
			Attributes: []models.Attribute{
				{
					Name:      "id",
					ExternalId: "id",
					Type:      "String",
					UniqueId:  true, // This is a primary key
				},
				{
					Name:      "appId",
					ExternalId: "appId",
					Type:      "String",
					UniqueId:  false,
				},
				{
					Name:      "name",
					ExternalId: "name",
					Type:      "String",
					UniqueId:  false,
				},
			},
		},
		"Assignment": {
			DisplayName: "Assignment",
			ExternalId:  "Assignment",
			Description: "Assignment entity for testing",
			Attributes: []models.Attribute{
				{
					Name:      "id",
					ExternalId: "id",
					Type:      "String",
					UniqueId:  true, // This is a primary key
				},
				{
					Name:      "roleId",
					ExternalId: "roleId",
					Type:      "String",
					UniqueId:  false, // This is a foreign key
				},
				{
					Name:      "uuid",
					ExternalId: "uuid",
					Type:      "String",
					UniqueId:  false,
				},
			},
		},
	}

	// Create relationships that mirror the circular relationship from SW-Assertions
	relationships := map[string]models.Relationship{
		// FK -> PK relationship: Assignment.roleId points to Role.id
		// This represents "Assignment has a reference to Role"
		"assigned_to_role": {
			DisplayName:    "assigned_to_role",
			Name:           "assigned_to_role",
			FromAttribute:  "Assignment.roleId",  // FK
			ToAttribute:    "Role.id",            // PK
		},
		// PK -> FK relationship: Role.id is referenced by Assignment.roleId
		// This creates a circular dependency with the above
		"role_to_assignment": {
			DisplayName:    "role_to_assignment",
			Name:           "role_to_assignment", 
			FromAttribute:  "Role.id",            // PK
			ToAttribute:    "Assignment.roleId",  // FK
		},
	}

	// Build the dependency graph
	graph, err := g.buildEntityDependencyGraph(entities, relationships)
	require.NoError(t, err, "Failed to build entity dependency graph")
	require.NotNil(t, graph, "Dependency graph should not be nil")

	// Try to get a topological ordering
	ordering, err := g.getTopologicalOrder(graph)
	require.NoError(t, err, "Failed to get topological order")
	require.NotNil(t, ordering, "Ordering should not be nil")

	// Verify the ordering contains both entities
	require.Len(t, ordering, 2, "Expected 2 entities in topological order")

	// Check that both entities are in the ordering
	roleFound := false
	assignmentFound := false
	roleIndex := -1
	assignmentIndex := -1
	
	for i, entity := range ordering {
		if entity == "Role" {
			roleFound = true
			roleIndex = i
		}
		if entity == "Assignment" {
			assignmentFound = true
			assignmentIndex = i
		}
	}
	
	assert.True(t, roleFound, "Role should be in the topological order")
	assert.True(t, assignmentFound, "Assignment should be in the topological order")

	// In a correct topological sort with FK->PK filtering,
	// Role should come before Assignment since Assignment depends on Role
	assert.Less(t, roleIndex, assignmentIndex,
		"Role should come before Assignment in the topological order, but found Role at index %d and Assignment at index %d", 
		roleIndex, assignmentIndex)

	// Verify that edges exist in the correct direction
	// In our graph, "Role" should be processed before "Assignment",
	// so we should have an edge from Role to Assignment
	
	// The edge direction is from entity that should be processed first
	// to entity that depends on it (reverse of data flow)
	
	// Check the correct edge: Role -> Assignment
	// This represents "Role must be generated before Assignment"
	_, errRoleToAssignment := graph.Edge("Role", "Assignment")
	hasEdgeRoleToAssignment := errRoleToAssignment == nil
	
	// Check the wrong direction: Assignment -> Role
	// This should not exist as it would create a cycle
	_, errAssignmentToRole := graph.Edge("Assignment", "Role")
	hasEdgeAssignmentToRole := errAssignmentToRole == nil

	assert.True(t, hasEdgeRoleToAssignment,
		"There should be an edge from Role to Assignment (Role should be processed before Assignment)")
	assert.False(t, hasEdgeAssignmentToRole,
		"There should NOT be an edge from Assignment to Role (would create a cycle)")
}