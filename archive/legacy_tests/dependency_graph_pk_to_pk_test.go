package generators

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
					Name:       "id",
					ExternalId: "id",
					Type:       "String",
					UniqueId:   true, // Primary key
				},
				{
					Name:       "name",
					ExternalId: "name",
					Type:       "String",
					UniqueId:   false,
				},
			},
		},
		"Settings": {
			DisplayName: "Settings",
			ExternalId:  "Settings",
			Description: "Settings entity for testing",
			Attributes: []models.Attribute{
				{
					Name:       "id",
					ExternalId: "id",
					Type:       "String",
					UniqueId:   true, // Primary key
				},
				{
					Name:       "accountId",
					ExternalId: "accountId",
					Type:       "String",
					UniqueId:   true, // This is a foreign key that's also unique (PK-PK relationship)
				},
				{
					Name:       "theme",
					ExternalId: "theme",
					Type:       "String",
					UniqueId:   false,
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

	assert.True(t, accountFound, "Account should be in the topological order")
	assert.True(t, settingsFound, "Settings should be in the topological order")

	// Account should come before Settings since Settings depends on Account's id
	assert.Less(t, accountIndex, settingsIndex,
		"Account should come before Settings in the topological order")

	// Verify edges
	assert.Len(t, edges, 1, "Expected exactly 1 edge in the graph")

	// Verify the direction of the edge (should be Account -> Settings)
	accountToSettingsFound := false
	for _, edge := range edges {
		if edge.Source == "Account" && edge.Target == "Settings" {
			accountToSettingsFound = true
			break
		}
	}

	assert.True(t, accountToSettingsFound, "There should be an edge from Account to Settings")
}
