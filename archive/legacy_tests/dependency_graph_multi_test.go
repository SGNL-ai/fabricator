package generators

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityDependencyWithMultipleRelationships(t *testing.T) {
	// Create test generator
	g := NewCSVGenerator("test_output", 10, false)

	// Define test entities with multiple relationships
	entities := map[string]models.Entity{
		"user1": {
			DisplayName: "User",
			ExternalId:  "User",
			Description: "User entity",
			Attributes: []models.Attribute{
				{
					Name:       "uuid",
					ExternalId: "uuid",
					Type:       "String",
					UniqueId:   true,
				},
			},
		},
		"assignment1": {
			DisplayName: "Assignment",
			ExternalId:  "Assignment",
			Description: "Assignment object",
			Attributes: []models.Attribute{
				{
					Name:       "id",
					ExternalId: "id",
					Type:       "String",
					UniqueId:   true,
				},
				{
					Name:       "assignedToUUID",
					ExternalId: "assignedToUUID",
					Type:       "String",
					UniqueId:   false,
				},
				{
					Name:       "createdByUUID",
					ExternalId: "createdByUUID",
					Type:       "String",
					UniqueId:   false,
				},
			},
		},
	}

	// Define two relationships from Assignment to User
	relationships := map[string]models.Relationship{
		"assigned_to_user": {
			DisplayName:   "assigned_to_user",
			Name:          "assigned_to_user",
			FromAttribute: "Assignment.assignedToUUID",
			ToAttribute:   "User.uuid",
		},
		"created_by_user": {
			DisplayName:   "created_by_user",
			Name:          "created_by_user",
			FromAttribute: "Assignment.createdByUUID",
			ToAttribute:   "User.uuid",
		},
	}

	// Build the dependency graph
	graph, err := g.buildEntityDependencyGraph(entities, relationships)
	require.NoError(t, err, "Failed to build entity dependency graph")

	// Get the graph edges
	edges, _ := graph.Edges()

	// Get the topological order
	order, err := g.getTopologicalOrder(graph)
	require.NoError(t, err, "Failed to get topological order")

	// In this case, User should still come before Assignment
	// despite having multiple relationships between them
	require.Len(t, order, 2, "Expected 2 entities in topological order")

	// The first entity should be User and the second should be Assignment
	assert.Equal(t, "user1", order[0], "First entity should be 'user1'")
	assert.Equal(t, "assignment1", order[1], "Second entity should be 'assignment1'")

	// Importantly, we should only have one edge in the graph
	// Multiple relationships should not create multiple edges between the same entities
	assert.Len(t, edges, 1, "Expected exactly 1 edge in the graph")
}
