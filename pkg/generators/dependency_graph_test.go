package generators

import (
	"fmt"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
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

	// In this case, User should come before Assignment
	// because Assignment depends on User (Assignment.uuid -> User.uuid)
	if len(order) != 2 {
		t.Fatalf("Expected 2 entities in topological order, got %d", len(order))
	}

	// The first entity should be User and the second should be Assignment
	if order[0] != "user1" {
		t.Errorf("Expected first entity to be 'user1', got '%s'", order[0])
	}

	if order[1] != "assignment1" {
		t.Errorf("Expected second entity to be 'assignment1', got '%s'", order[1])
	}
}