package generators

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIdentityRelationships tests the handling of relationships between
// primary keys (PKs) of different entities and how we handle bidirectional relationships
func TestIdentityRelationships(t *testing.T) {
	// Disable color output for tests
	color.NoColor = true

	// Create a CSV generator for testing
	g := NewCSVGenerator("output", 10, false)

	// Create test entities with primary keys that reference each other
	entities := map[string]models.Entity{
		"User": {
			DisplayName: "User",
			ExternalId:  "User",
			Description: "User entity for testing",
			Attributes: []models.Attribute{
				{
					Name:       "id",
					ExternalId: "id",
					Type:       "String",
					UniqueId:   true, // This is a primary key
				},
				{
					Name:       "name",
					ExternalId: "name",
					Type:       "String",
					UniqueId:   false,
				},
			},
		},
		"Profile": {
			DisplayName: "Profile",
			ExternalId:  "Profile",
			Description: "Profile entity for testing",
			Attributes: []models.Attribute{
				{
					Name:       "id",
					ExternalId: "id",
					Type:       "String",
					UniqueId:   true, // This is a primary key
				},
				{
					Name:       "userId",
					ExternalId: "userId",
					Type:       "String",
					UniqueId:   true, // This is a foreign key that's also unique (identity relationship)
				},
				{
					Name:       "bio",
					ExternalId: "bio",
					Type:       "String",
					UniqueId:   false,
				},
			},
		},
	}

	// Define a bidirectional identity relationship between User and Profile
	// 1. First direction: Profile.userId -> User.id (FK to PK, pretty common)
	// 2. Reverse direction: User.id -> Profile.userId (PK to FK, can cause cycles)
	relationships := map[string]models.Relationship{
		"user_to_profile": {
			DisplayName:   "user_to_profile",
			Name:          "user_to_profile",
			FromAttribute: "Profile.userId", // FK
			ToAttribute:   "User.id",        // PK
		},
		"profile_to_user": {
			DisplayName:   "profile_to_user",
			Name:          "profile_to_user",
			FromAttribute: "User.id",        // PK
			ToAttribute:   "Profile.userId", // FK
		},
	}

	// Build the dependency graph
	graph, err := g.buildEntityDependencyGraph(entities, relationships)
	require.NoError(t, err, "Failed to build entity dependency graph")
	require.NotNil(t, graph, "Dependency graph should not be nil")

	// Get the graph edges
	edges, _ := graph.Edges()

	// Try to get a topological ordering
	ordering, err := g.getTopologicalOrder(graph)
	require.NoError(t, err, "Failed to get topological order")
	require.NotNil(t, ordering, "Ordering should not be nil")

	// Verify the ordering contains both entities
	require.Len(t, ordering, 2, "Expected 2 entities in topological order")

	// Check that both entities are in the ordering
	userFound := false
	profileFound := false

	for _, entity := range ordering {
		if entity == "User" {
			userFound = true
		}
		if entity == "Profile" {
			profileFound = true
		}
	}

	assert.True(t, userFound, "User should be in the topological order")
	assert.True(t, profileFound, "Profile should be in the topological order")

	// Verify that exactly one edge exists (regardless of direction)
	assert.Len(t, edges, 1, "Expected exactly 1 edge in the graph after filtering")

	// Check the edge direction
	// Note: With our improved logic, the direction could be either way
	// What's important is that we have a deterministic ordering without cycles
	foundEdge := false

	for _, edge := range edges {
		if (edge.Source == "User" && edge.Target == "Profile") ||
			(edge.Source == "Profile" && edge.Target == "User") {
			foundEdge = true
		}
	}

	assert.True(t, foundEdge, "There should be an edge between User and Profile")
}
