package generators

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComplexEntityDependencyGraph(t *testing.T) {
	// Create test generator
	g := NewCSVGenerator("test_output", 10, false)

	// Define test entities with a three-way relationship:
	// User -> GroupMembership <- Group
	// GroupMembership depends on both User and Group
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
		"group1": {
			DisplayName: "Group",
			ExternalId:  "Group",
			Description: "Group entity",
			Attributes: []models.Attribute{
				{
					Name:       "id",
					ExternalId: "id",
					Type:       "String",
					UniqueId:   true,
				},
			},
		},
		"groupMembership1": {
			DisplayName: "GroupMembership",
			ExternalId:  "GroupMembership",
			Description: "Group membership mapping",
			Attributes: []models.Attribute{
				{
					Name:       "id",
					ExternalId: "id",
					Type:       "String",
					UniqueId:   true,
				},
				{
					Name:       "userUuid",
					ExternalId: "userUuid",
					Type:       "String",
					UniqueId:   false,
				},
				{
					Name:       "groupId",
					ExternalId: "groupId",
					Type:       "String",
					UniqueId:   false,
				},
			},
		},
	}

	// Define the relationships
	relationships := map[string]models.Relationship{
		"user_membership": {
			DisplayName:   "user_membership",
			Name:          "user_membership",
			FromAttribute: "GroupMembership.userUuid",
			ToAttribute:   "User.uuid",
		},
		"group_membership": {
			DisplayName:   "group_membership",
			Name:          "group_membership",
			FromAttribute: "GroupMembership.groupId",
			ToAttribute:   "Group.id",
		},
	}

	// Build the dependency graph
	var err error
	g.dependencyGraph, err = g.buildEntityDependencyGraph(entities, relationships)
	require.NoError(t, err, "Failed to build entity dependency graph")

	// Get the graph edges
	edges, _ := g.dependencyGraph.Edges()

	// Get the topological order
	order, err := g.getTopologicalOrder(g.dependencyGraph)
	require.NoError(t, err, "Failed to get topological order")

	// In this case, both User and Group should come before GroupMembership
	require.Len(t, order, 3, "Expected 3 entities in topological order")

	// Check if GroupMembership is the last entity in the order
	// since it depends on both User and Group
	userIndex := -1
	groupIndex := -1
	membershipIndex := -1

	for i, entityID := range order {
		switch entityID {
		case "user1":
			userIndex = i
		case "group1":
			groupIndex = i
		case "groupMembership1":
			membershipIndex = i
		}
	}

	// Verify that GroupMembership comes after both User and Group
	assert.True(t, membershipIndex > userIndex && membershipIndex > groupIndex,
		"GroupMembership should come after both User and Group in the topological order. "+
			"User index: %d, Group index: %d, GroupMembership index: %d",
		userIndex, groupIndex, membershipIndex)

	// Verify that we have exactly two edges
	assert.Len(t, edges, 2, "Expected exactly 2 edges in the graph")
}
