package generators

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityDependencySort(t *testing.T) {
	// Create test generator
	g := NewCSVGenerator("test_output", 10, false)

	// Define test entities based on the given YAML example
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
					Name:       "uuid",
					ExternalId: "uuid",
					Type:       "String",
					UniqueId:   false,
				},
			},
		},
	}

	// Define the relationship from Assignment.uuid to User.uuid
	relationships := map[string]models.Relationship{
		"assigned_to_user": {
			DisplayName:   "assigned_to_user",
			Name:          "assigned_to_user",
			FromAttribute: "Assignment.uuid",
			ToAttribute:   "User.uuid",
		},
	}

	// Build the dependency graph
	var err error
	g.dependencyGraph, err = g.buildEntityDependencyGraph(entities, relationships)
	require.NoError(t, err, "Failed to build entity dependency graph")

	// We don't need edges for this test
	// Uncomment if needed for edge verification:
	// edges, _ := g.dependencyGraph.Edges()

	// Get the topological order
	order, err := g.getTopologicalOrder(g.dependencyGraph)
	require.NoError(t, err, "Failed to get topological order")

	// In this case, User should come before Assignment
	// because Assignment depends on User (Assignment.uuid -> User.uuid)
	require.Len(t, order, 2, "Expected 2 entities in topological order")

	// The first entity should be User and the second should be Assignment
	assert.Equal(t, "user1", order[0], "First entity should be 'user1'")
	assert.Equal(t, "assignment1", order[1], "Second entity should be 'assignment1'")
}
