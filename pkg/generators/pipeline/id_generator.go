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

// GenerateIDs generates unique IDs for all entities in topological order
func (g *IDGenerator) GenerateIDs(graph *model.Graph, dataVolume int) error {
	if graph == nil {
		return fmt.Errorf("graph cannot be nil")
	}

	if dataVolume <= 0 {
		return fmt.Errorf("data volume must be greater than 0")
	}

	// Generate IDs for each entity
	for _, entity := range graph.GetAllEntities() {
		// Find the unique ID attribute (primary key)
		primaryKey := entity.GetPrimaryKey()
		if primaryKey == nil {
			continue // Skip entities without primary keys
		}

		// Show progress for current entity (no newline, will be overwritten)
		fmt.Printf("\r%-80s\r→ Generating %s...", "", entity.GetName())

		// Generate the specified number of rows with unique IDs
		for i := 0; i < dataVolume; i++ {
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
