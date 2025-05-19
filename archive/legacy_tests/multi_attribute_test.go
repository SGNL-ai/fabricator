package generators

import (
	"fmt"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiAttributeRelationships(t *testing.T) {
	// Create a generator with a small data volume for testing
	dataVolume := 10
	g := NewCSVGenerator("test_output", dataVolume, false)

	// Define entities: User and Ticket
	// Ticket has two attributes that reference User: assignedTo and createdBy
	entities := map[string]models.Entity{
		"user1": {
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
			},
		},
		"ticket1": {
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
					Name:       "assignedTo",
					ExternalId: "assignedTo",
					Type:       "String",
				},
				{
					Name:       "createdBy",
					ExternalId: "createdBy",
					Type:       "String",
				},
			},
		},
	}

	// Define relationships between Ticket attributes and User
	relationships := map[string]models.Relationship{
		"ticket_assigned_to": {
			DisplayName:   "ticket_assigned_to",
			Name:          "ticket_assigned_to",
			FromAttribute: "Ticket.assignedTo",
			ToAttribute:   "User.id",
		},
		"ticket_created_by": {
			DisplayName:   "ticket_created_by",
			Name:          "ticket_created_by",
			FromAttribute: "Ticket.createdBy",
			ToAttribute:   "User.id",
		},
	}

	// Set up the generator
	var err error
	err = g.Setup(entities, relationships)
	require.NoError(t, err, "Failed to setup generator")

	// Generate data
	err = g.GenerateData()
	require.NoError(t, err, "Failed to generate data")

	// Verify the generated data
	userData := g.EntityData["user1"]
	ticketData := g.EntityData["ticket1"]

	// Find indices of relevant columns
	userIdIndex := findColumnIndex(userData.Headers, "id")
	ticketAssignedToIndex := findColumnIndex(ticketData.Headers, "assignedTo")
	ticketCreatedByIndex := findColumnIndex(ticketData.Headers, "createdBy")

	require.NotEqual(t, -1, userIdIndex, "User.id column not found")
	require.NotEqual(t, -1, ticketAssignedToIndex, "Ticket.assignedTo column not found")
	require.NotEqual(t, -1, ticketCreatedByIndex, "Ticket.createdBy column not found")

	// Collect all valid user IDs
	validUserIds := make(map[string]bool)
	for _, row := range userData.Rows {
		userId := row[userIdIndex]
		validUserIds[userId] = true
	}
	fmt.Printf("Valid User IDs (%d): %v\n", len(validUserIds), validUserIds)

	// Check that all tickets have valid assignedTo and createdBy values
	invalidAssignedTo := 0
	invalidCreatedBy := 0

	for i, row := range ticketData.Rows {
		assignedToId := row[ticketAssignedToIndex]
		createdById := row[ticketCreatedByIndex]

		if !validUserIds[assignedToId] {
			invalidAssignedTo++
			fmt.Printf("Warning: Ticket %d has invalid assignedTo: %s\n", i, assignedToId)
		}

		if !validUserIds[createdById] {
			invalidCreatedBy++
			fmt.Printf("Warning: Ticket %d has invalid createdBy: %s\n", i, createdById)
		}
	}

	// Assert that all references are valid
	assert.Zero(t, invalidAssignedTo, "All tickets should have valid assignedTo values")
	assert.Zero(t, invalidCreatedBy, "All tickets should have valid createdBy values")

	// Check for duplicated IDs in the User data
	userIds := make(map[string]int)
	duplicateUserIds := 0
	for i, row := range userData.Rows {
		userId := row[userIdIndex]
		if _, exists := userIds[userId]; exists {
			duplicateUserIds++
			fmt.Printf("Warning: Duplicate User ID %s at row %d\n", userId, i)
		}
		userIds[userId] = i
	}

	assert.Zero(t, duplicateUserIds, "User IDs should be unique")

	// Check if assignedTo and createdBy are ever the same for a ticket
	sameUserCount := 0
	for i, row := range ticketData.Rows {
		if row[ticketAssignedToIndex] == row[ticketCreatedByIndex] {
			sameUserCount++
			fmt.Printf("Note: Ticket %d has same user for assignedTo and createdBy: %s\n",
				i, row[ticketAssignedToIndex])
		}
	}

	fmt.Printf("Number of tickets with same assignedTo and createdBy: %d out of %d\n",
		sameUserCount, len(ticketData.Rows))
}

// Helper function to find the index of a column by name
func findColumnIndex(headers []string, name string) int {
	for i, header := range headers {
		if header == name {
			return i
		}
	}
	return -1
}
