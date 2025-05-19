package generators

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChainedEntityDependencyGraph(t *testing.T) {
	// Create test generator
	g := NewCSVGenerator("test_output", 10, false)

	// Define test entities with a chained dependency:
	// User -> GroupMembership <- Group <- Role
	// This creates a chain of dependencies where:
	// - GroupMembership depends on both User and Group
	// - Role depends on Group
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
		"role1": {
			DisplayName: "Role",
			ExternalId:  "Role",
			Description: "Role entity",
			Attributes: []models.Attribute{
				{
					Name:       "id",
					ExternalId: "id",
					Type:       "String",
					UniqueId:   true,
				},
				{
					Name:       "groupId",
					ExternalId: "groupId",
					Type:       "String",
					UniqueId:   false,
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
		"role_group": {
			DisplayName:   "role_group",
			Name:          "role_group",
			FromAttribute: "Role.groupId",
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

	// We expect the following dependencies:
	// 1. Group should be earlier in the ordering (Role depends on Group)
	// 2. User has no dependencies, so its position is less constrained
	// 3. GroupMembership should be after User and Group
	// 4. Role should be after Group

	require.Len(t, order, 4, "Expected 4 entities in topological order")

	// Find the indices of each entity in the order
	userIndex := -1
	groupIndex := -1
	roleIndex := -1
	membershipIndex := -1

	for i, entityID := range order {
		switch entityID {
		case "user1":
			userIndex = i
		case "group1":
			groupIndex = i
		case "role1":
			roleIndex = i
		case "groupMembership1":
			membershipIndex = i
		}
	}

	// Verify the dependencies:
	// 1. GroupMembership should come after both User and Group
	assert.True(t, membershipIndex > userIndex && membershipIndex > groupIndex,
		"GroupMembership should come after both User and Group in the topological order. "+
			"User index: %d, Group index: %d, GroupMembership index: %d",
		userIndex, groupIndex, membershipIndex)

	// 2. Role should come after Group
	assert.Greater(t, roleIndex, groupIndex,
		"Role should come after Group in the topological order. "+
			"Group index: %d, Role index: %d", groupIndex, roleIndex)

	// Verify that we have exactly three edges
	assert.Len(t, edges, 3, "Expected exactly 3 edges in the graph")
}
