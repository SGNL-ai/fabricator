package generators

import (
	"fmt"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
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
	if err != nil {
		t.Fatalf("Failed to build entity dependency graph: %v", err)
	}

	// Print the graph edges for debugging
	fmt.Println("Graph edges:")
	edges, _ := graph.Edges()
	for _, edge := range edges {
		fmt.Printf("Edge: %s -> %s\n", edge.Source, edge.Target)
	}

	// Get the topological order
	order, err := g.getTopologicalOrder(graph)
	if err != nil {
		t.Fatalf("Failed to get topological order: %v", err)
	}
	
	fmt.Println("Topological order:", order)

	// In this case, User should still come before Assignment
	// despite having multiple relationships between them
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

	// Importantly, we should only have one edge in the graph
	// Multiple relationships should not create multiple edges between the same entities
	if len(edges) != 1 {
		t.Errorf("Expected exactly 1 edge in the graph, got %d", len(edges))
	}
}