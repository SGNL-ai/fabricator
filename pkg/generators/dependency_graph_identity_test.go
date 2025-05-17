package generators

import (
	"fmt"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/fatih/color"
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
					Name:      "id",
					ExternalId: "id",
					Type:      "String",
					UniqueId:  true, // This is a primary key
				},
				{
					Name:      "name",
					ExternalId: "name",
					Type:      "String",
					UniqueId:  false,
				},
			},
		},
		"Profile": {
			DisplayName: "Profile",
			ExternalId:  "Profile",
			Description: "Profile entity for testing",
			Attributes: []models.Attribute{
				{
					Name:      "id",
					ExternalId: "id",
					Type:      "String",
					UniqueId:  true, // This is a primary key
				},
				{
					Name:      "userId",
					ExternalId: "userId",
					Type:      "String",
					UniqueId:  true, // This is a foreign key that's also unique (identity relationship)
				},
				{
					Name:      "bio",
					ExternalId: "bio",
					Type:      "String",
					UniqueId:  false,
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
			FromAttribute: "Profile.userId",  // FK
			ToAttribute:   "User.id",         // PK
		},
		"profile_to_user": {
			DisplayName:   "profile_to_user",
			Name:          "profile_to_user", 
			FromAttribute: "User.id",         // PK
			ToAttribute:   "Profile.userId",  // FK
		},
	}

	// Build the dependency graph
	graph, err := g.buildEntityDependencyGraph(entities, relationships)
	if err != nil {
		t.Fatalf("Failed to build entity dependency graph: %v", err)
	}
	if graph == nil {
		t.Fatalf("Dependency graph should not be nil")
	}

	// Print the graph edges for debugging
	fmt.Println("Graph edges:")
	edges, _ := graph.Edges()
	for _, edge := range edges {
		fmt.Printf("Edge: %s -> %s\n", edge.Source, edge.Target)
	}

	// Try to get a topological ordering
	ordering, err := g.getTopologicalOrder(graph)
	if err != nil {
		t.Fatalf("Failed to get topological order: %v", err)
	}
	if ordering == nil {
		t.Fatalf("Ordering should not be nil")
	}

	// Print the topological order for debugging
	fmt.Println("Topological order:", ordering)

	// Verify the ordering contains both entities
	if len(ordering) != 2 {
		t.Fatalf("Expected 2 entities in topological order, got %d", len(ordering))
	}

	// Check that both entities are in the ordering
	userFound := false
	profileFound := false
	userIndex := -1
	profileIndex := -1
	
	for i, entity := range ordering {
		if entity == "User" {
			userFound = true
			userIndex = i
		}
		if entity == "Profile" {
			profileFound = true
			profileIndex = i
		}
	}
	
	if !userFound {
		t.Errorf("User should be in the topological order but was not found")
	}
	if !profileFound {
		t.Errorf("Profile should be in the topological order but was not found")
	}

	// With identity relationships, either entity could reasonably come first
	// The important thing is that we have a valid topological order without cycles
	// Check that we have both entities in the order
	if !userFound || !profileFound {
		t.Errorf("Both User and Profile should be in the topological order")
	}
	
	// Log the actual order for information
	fmt.Printf("Actual topological order: %s index=%d, %s index=%d\n", 
		"User", userIndex, "Profile", profileIndex)

	// Verify that exactly one edge exists (regardless of direction)
	if len(edges) != 1 {
		t.Errorf("Expected exactly 1 edge in the graph after filtering, got %d", len(edges))
	}

	// Check the edge direction
	// Note: With our improved logic, the direction could be either way
	// What's important is that we have a deterministic ordering without cycles
	foundEdge := false
	var fromEntity, toEntity string

	for _, edge := range edges {
		if (edge.Source == "User" && edge.Target == "Profile") || 
		   (edge.Source == "Profile" && edge.Target == "User") {
			foundEdge = true
			fromEntity = edge.Source
			toEntity = edge.Target
		}
	}

	if !foundEdge {
		t.Errorf("There should be an edge between User and Profile")
	} else {
		// Log the actual edge direction for information
		fmt.Printf("Found edge direction: %s -> %s\n", fromEntity, toEntity)
	}
}

// TestPrimaryKeyToPrimaryKeyRelationships tests relationships between
// primary keys of different entities (PK-to-PK)
func TestPrimaryKeyToPrimaryKeyRelationships(t *testing.T) {
	// Disable color output for tests
	color.NoColor = true

	// Create a CSV generator for testing
	g := NewCSVGenerator("output", 10, false)

	// Create test entities with primary keys that reference each other
	entities := map[string]models.Entity{
		"Account": {
			DisplayName: "Account",
			ExternalId:  "Account",
			Description: "Account entity for testing",
			Attributes: []models.Attribute{
				{
					Name:      "id",
					ExternalId: "id",
					Type:      "String",
					UniqueId:  true, // Primary key
				},
				{
					Name:      "name",
					ExternalId: "name",
					Type:      "String",
					UniqueId:  false,
				},
			},
		},
		"Settings": {
			DisplayName: "Settings",
			ExternalId:  "Settings",
			Description: "Settings entity for testing",
			Attributes: []models.Attribute{
				{
					Name:      "id",
					ExternalId: "id",
					Type:      "String",
					UniqueId:  true, // Primary key
				},
				{
					Name:      "accountId",
					ExternalId: "accountId",
					Type:      "String",
					UniqueId:  true, // This is a foreign key that's also unique (PK-PK relationship)
				},
				{
					Name:      "theme",
					ExternalId: "theme",
					Type:      "String",
					UniqueId:  false,
				},
			},
		},
	}

	// Define relationships where PKs reference each other
	relationships := map[string]models.Relationship{
		"settings_to_account": {
			DisplayName:   "settings_to_account",
			Name:          "settings_to_account",
			FromAttribute: "Settings.accountId", // PK
			ToAttribute:   "Account.id",         // PK
		},
	}

	// Build the dependency graph
	graph, err := g.buildEntityDependencyGraph(entities, relationships)
	if err != nil {
		t.Fatalf("Failed to build entity dependency graph: %v", err)
	}
	if graph == nil {
		t.Fatalf("Dependency graph should not be nil")
	}

	// Print the graph edges for debugging
	fmt.Println("Graph edges for PK-PK test:")
	edges, _ := graph.Edges()
	for _, edge := range edges {
		fmt.Printf("Edge: %s -> %s\n", edge.Source, edge.Target)
	}

	// Try to get a topological ordering
	ordering, err := g.getTopologicalOrder(graph)
	if err != nil {
		t.Fatalf("Failed to get topological order: %v", err)
	}
	if ordering == nil {
		t.Fatalf("Ordering should not be nil")
	}

	// Print the topological order for debugging
	fmt.Println("Topological order for PK-PK test:", ordering)

	// Verify the ordering contains both entities
	if len(ordering) != 2 {
		t.Fatalf("Expected 2 entities in topological order, got %d", len(ordering))
	}

	// Check that both entities are in the ordering
	accountFound := false
	settingsFound := false
	accountIndex := -1
	settingsIndex := -1
	
	for i, entity := range ordering {
		if entity == "Account" {
			accountFound = true
			accountIndex = i
		}
		if entity == "Settings" {
			settingsFound = true
			settingsIndex = i
		}
	}
	
	if !accountFound {
		t.Errorf("Account should be in the topological order but was not found")
	}
	if !settingsFound {
		t.Errorf("Settings should be in the topological order but was not found")
	}

	// Account should come before Settings since Settings depends on Account's id
	if accountIndex >= settingsIndex {
		t.Errorf("Account should come before Settings in the topological order, but found Account at index %d and Settings at index %d", 
			accountIndex, settingsIndex)
	}

	// Verify edges
	if len(edges) != 1 {
		t.Errorf("Expected exactly 1 edge in the graph, got %d", len(edges))
	}

	// Verify the direction of the edge (should be Account -> Settings)
	accountToSettingsFound := false
	for _, edge := range edges {
		if edge.Source == "Account" && edge.Target == "Settings" {
			accountToSettingsFound = true
			break
		}
	}

	if !accountToSettingsFound {
		t.Errorf("There should be an edge from Account to Settings")
	}
}