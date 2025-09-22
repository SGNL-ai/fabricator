package pipeline

import (
	"fmt"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
)

// RelationshipLinker handles establishing relationships between entities
type RelationshipLinker struct {
	// Configuration options can be added here
}

// NewRelationshipLinker creates a new relationship linker
func NewRelationshipLinker() RelationshipLinkerInterface {
	return &RelationshipLinker{}
}

// LinkRelationships establishes relationships between entities
func (l *RelationshipLinker) LinkRelationships(graph *model.Graph, autoCardinality bool) error {
	if graph == nil {
		return fmt.Errorf("graph cannot be nil")
	}

	// Process entities in order for optimal FK assignment
	for _, entity := range graph.GetAllEntities() {
		// Get relationships where this entity is the source (has FK attributes)
		entityRelationships := graph.GetRelationshipsForEntity(entity.GetID())

		// Filter to only relationships where this entity is the source
		sourceRelationships := make([]model.RelationshipInterface, 0)
		for _, relationship := range entityRelationships {
			if relationship.GetSourceEntity().GetID() == entity.GetID() {
				sourceRelationships = append(sourceRelationships, relationship)
			}
		}

		if len(sourceRelationships) == 0 {
			continue // No FK relationships for this entity
		}

		// Show progress for current entity relationships
		fmt.Printf("\r%-80s\r→ Linking %s relationships...", "", entity.GetName())

		// Process FK relationships for this entity
		for i, relationship := range sourceRelationships {
			isLastRelationship := (i == len(sourceRelationships)-1)

			// Process all rows for this relationship
			duplicateIndices := make([]int, 0)
			err := entity.ForEachRow(func(row *model.Row, rowIndex int) error {
				// Ask relationship to provide target PK value for this source row
				targetValue, err := relationship.GetTargetValueForSourceRow(rowIndex, autoCardinality)
				if err != nil {
					return fmt.Errorf("failed to get target value for row %d: %w", rowIndex, err)
				}

				// Set the FK value in the source row
				row.SetValue(relationship.GetSourceAttribute().GetName(), targetValue)

				// If this is the last FK relationship for a junction table, check for duplicates
				if isLastRelationship && len(sourceRelationships) > 1 {
					if !entity.IsForeignKeyUnique(row) {
						// Duplicate composite key found - mark for removal
						duplicateIndices = append(duplicateIndices, rowIndex)
					}
				}

				return nil
			})

			if err != nil {
				return fmt.Errorf("failed to link relationship %s: %w", relationship.GetID(), err)
			}

			// Remove duplicate rows (in reverse order to maintain indices)
			for i := len(duplicateIndices) - 1; i >= 0; i-- {
				err := entity.RemoveRow(duplicateIndices[i])
				if err != nil {
					return fmt.Errorf("failed to remove duplicate row %d from entity %s: %w",
						duplicateIndices[i], entity.GetName(), err)
				}
			}
		}
	}

	// Clear relationship linking progress line
	fmt.Printf("\r%-80s\r", "")

	return nil
}
