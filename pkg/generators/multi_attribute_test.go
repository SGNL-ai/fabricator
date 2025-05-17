package generators

import (
	"fmt"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
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
	if err != nil {
		t.Fatalf("Failed to setup generator: %v", err)
	}

	// Generate data
	err = g.GenerateData()
	if err != nil {
		t.Fatalf("Failed to generate data: %v", err)
	}

	// Verify the generated data
	userData := g.EntityData["user1"]
	ticketData := g.EntityData["ticket1"]

	// Find indices of relevant columns
	userIdIndex := findColumnIndex(userData.Headers, "id")
	ticketAssignedToIndex := findColumnIndex(ticketData.Headers, "assignedTo")
	ticketCreatedByIndex := findColumnIndex(ticketData.Headers, "createdBy")

	if userIdIndex == -1 || ticketAssignedToIndex == -1 || ticketCreatedByIndex == -1 {
		t.Fatalf("Required columns not found. User.id: %d, Ticket.assignedTo: %d, Ticket.createdBy: %d",
			userIdIndex, ticketAssignedToIndex, ticketCreatedByIndex)
	}

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
	if invalidAssignedTo > 0 {
		t.Errorf("%d tickets have invalid assignedTo values", invalidAssignedTo)
	}

	if invalidCreatedBy > 0 {
		t.Errorf("%d tickets have invalid createdBy values", invalidCreatedBy)
	}

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

	if duplicateUserIds > 0 {
		t.Errorf("Found %d duplicate User IDs", duplicateUserIds)
	}

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