package generators

import (
	"github.com/fatih/color"
)

// makeRelationshipsConsistentForEntity ensures relationship consistency for a specific entity
// This is used when generating data in topological order
func (g *CSVGenerator) makeRelationshipsConsistentForEntity(entityID string) {
	// Get the relationships where this entity is the "from" entity
	relationships := g.relationshipMap[entityID]
	if len(relationships) == 0 {
		return
	}

	// Get the entity's display name for more readable output
	entityName := "Unknown"
	if csvData, exists := g.EntityData[entityID]; exists && csvData != nil {
		entityName = csvData.EntityName
	}

	// Ensuring consistency (removed from output)
	
	// Process each relationship
	for _, link := range relationships {
		// We only process relationships where this entity is the "from" entity
		if link.FromEntityID == entityID {
			// Get the "to" entity's name for more readable output
			toEntityName := "Unknown"
			if toData, exists := g.EntityData[link.ToEntityID]; exists && toData != nil {
				toEntityName = toData.EntityName
			}
			
			// Display relationship information in a human-readable format
			color.Green("✓ Linking: %s.%s → %s.%s", 
				entityName, link.FromAttribute,
				toEntityName, link.ToAttribute)
				
			g.makeRelationshipsConsistent(entityID, link)
		}
	}
}