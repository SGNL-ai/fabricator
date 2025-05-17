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

	color.Green("  Making relationships consistent for %s", entityID)
	
	// Process each relationship
	for _, link := range relationships {
		// We only process relationships where this entity is the "from" entity
		if link.FromEntityID == entityID {
			g.makeRelationshipsConsistent(entityID, link)
		}
	}
}