package generators

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComplexRelationshipsTopologicalSort tests that a complex set of relationships
// with potential circular references can still be sorted topologically.
func TestComplexRelationshipsTopologicalSort(t *testing.T) {
	// Create a generator with a small data volume for testing
	dataVolume := 10
	g := NewCSVGenerator("test_output", dataVolume, false)

	// Define entities for a complex system:
	// - User: has many tickets, part of a team
	// - Team: has many users, has manager (a user)
	// - Ticket: assigned to a user, created by a user, belongs to a team
	entities := map[string]models.Entity{
		"user": {
			DisplayName: "User",
			ExternalId:  "User",
			Description: "User entity",
			Attributes: []models.Attribute{
				{
					Name:       "id",
					ExternalId: "id",
					Type:       "String",
					UniqueId:   true,
				},
				{
					Name:       "name",
					ExternalId: "name",
					Type:       "String",
				},
				{
					Name:       "teamId",
					ExternalId: "teamId",
					Type:       "String",
				},
			},
		},
		"team": {
			DisplayName: "Team",
			ExternalId:  "Team",
			Description: "Team entity",
			Attributes: []models.Attribute{
				{
					Name:       "id",
					ExternalId: "id",
					Type:       "String",
					UniqueId:   true,
				},
				{
					Name:       "name",
					ExternalId: "name",
					Type:       "String",
				},
				{
					Name:       "managerId",
					ExternalId: "managerId",
					Type:       "String",
				},
			},
		},
		"ticket": {
			DisplayName: "Ticket",
			ExternalId:  "Ticket",
			Description: "Ticket entity",
			Attributes: []models.Attribute{
				{
					Name:       "id",
					ExternalId: "id",
					Type:       "String",
					UniqueId:   true,
				},
				{
					Name:       "title",
					ExternalId: "title",
					Type:       "String",
				},
				{
					Name:       "assignedToId",
					ExternalId: "assignedToId",
					Type:       "String",
				},
				{
					Name:       "createdById",
					ExternalId: "createdById",
					Type:       "String",
				},
				{
					Name:       "teamId",
					ExternalId: "teamId",
					Type:       "String",
				},
			},
		},
	}

	// Define relationships with potential circular references
	relationships := map[string]models.Relationship{
		"user_team": {
			DisplayName:   "user_team",
			Name:          "user_team",
			FromAttribute: "User.teamId",
			ToAttribute:   "Team.id",
		},
		"team_manager": {
			DisplayName:   "team_manager",
			Name:          "team_manager",
			FromAttribute: "Team.managerId",
			ToAttribute:   "User.id",
		},
		"ticket_assigned_to": {
			DisplayName:   "ticket_assigned_to",
			Name:          "ticket_assigned_to",
			FromAttribute: "Ticket.assignedToId",
			ToAttribute:   "User.id",
		},
		"ticket_created_by": {
			DisplayName:   "ticket_created_by",
			Name:          "ticket_created_by",
			FromAttribute: "Ticket.createdById",
			ToAttribute:   "User.id",
		},
		"ticket_team": {
			DisplayName:   "ticket_team",
			Name:          "ticket_team",
			FromAttribute: "Ticket.teamId",
			ToAttribute:   "Team.id",
		},
	}

	// Setup should handle these circular dependencies gracefully
	err := g.Setup(entities, relationships)
	require.NoError(t, err, "Failed to set up complex relationships")
	
	// Verify that we can get a valid topological order
	entityOrder, err := g.getTopologicalOrder(g.dependencyGraph)
	require.NoError(t, err, "Failed to get topological order")
	
	// Verify that all entities are included in the ordering
	assert.Len(t, entityOrder, 3, "Expected all 3 entities in order")
	
	// Verify each entity is in the ordering
	expectedEntities := map[string]bool{"user": true, "team": true, "ticket": true}
	for _, entity := range entityOrder {
		delete(expectedEntities, entity)
	}
	
	assert.Empty(t, expectedEntities, "All entities should be present in topological order")
}