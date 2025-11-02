package pipeline

import (
	"fmt"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/google/uuid"
)

// IDGenerator handles the generation of entity IDs in topological order
type IDGenerator struct {
	// Configuration options can be added here
}

// NewIDGenerator creates a new ID generator
func NewIDGenerator() IDGeneratorInterface {
	return &IDGenerator{}
}

// GenerateIDs generates unique IDs for all entities in topological order.
// rowCounts maps entity external_id to the number of rows to generate.
func (g *IDGenerator) GenerateIDs(graph *model.Graph, rowCounts map[string]int) error {
	if graph == nil {
		return fmt.Errorf("graph cannot be nil")
	}

	if len(rowCounts) == 0 {
		return fmt.Errorf("row counts map cannot be nil or empty")
	}

	// Generate IDs for each entity
	for _, entity := range graph.GetAllEntities() {
		// Find the unique ID attribute (primary key)
		primaryKey := entity.GetPrimaryKey()
		if primaryKey == nil {
			continue // Skip entities without primary keys
		}

		// Get row count for this entity
		entityID := entity.GetExternalID()
		count, exists := rowCounts[entityID]
		if !exists {
			return fmt.Errorf("no row count specified for entity %s", entityID)
		}

		if count <= 0 {
			return fmt.Errorf("row count for entity %s must be greater than 0, got %d", entityID, count)
		}

		// Show progress for current entity (no newline, will be overwritten)
		fmt.Printf("\r%-80s\r→ Generating %s (%d rows)...", "", entity.GetName(), count)

		// Generate the specified number of rows with unique IDs
		for i := 0; i < count; i++ {
			// Create row with just the primary key
			rowData := map[string]string{
				primaryKey.GetName(): uuid.New().String(),
			}

			// Add row to entity (AddRow will validate uniqueness)
			err := entity.AddRow(model.NewRow(rowData))
			if err != nil {
				return fmt.Errorf("failed to add row to entity %s: %w", entity.GetExternalID(), err)
			}
		}
	}

	return nil
}
