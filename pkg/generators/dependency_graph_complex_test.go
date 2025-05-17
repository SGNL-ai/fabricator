package generators

import (
	"fmt"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
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
	if err != nil {
		t.Fatalf("Failed to build entity dependency graph: %v", err)
	}

	// Print the graph edges for debugging
	fmt.Println("Graph edges:")
	edges, _ := g.dependencyGraph.Edges()
	for _, edge := range edges {
		fmt.Printf("Edge: %s -> %s\n", edge.Source, edge.Target)
	}

	// Get the topological order
	order, err := g.getTopologicalOrder(g.dependencyGraph)
	if err != nil {
		t.Fatalf("Failed to get topological order: %v", err)
	}
	
	fmt.Println("Topological order:", order)

	// In this case, both User and Group should come before GroupMembership
	if len(order) != 3 {
		t.Fatalf("Expected 3 entities in topological order, got %d", len(order))
	}

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
	if membershipIndex < userIndex || membershipIndex < groupIndex {
		t.Errorf("GroupMembership should come after both User and Group in the topological order")
		t.Errorf("User index: %d, Group index: %d, GroupMembership index: %d", 
			userIndex, groupIndex, membershipIndex)
	}

	// Verify that we have exactly two edges
	if len(edges) != 2 {
		t.Errorf("Expected exactly 2 edges in the graph, got %d", len(edges))
	}
}