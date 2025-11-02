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
			// Detect same_as relationships (both source and target attributes are unique/PKs)
			// These represent bidirectional (0..1)-to-(0..1) identity mappings
			isSameAs := relationship.GetSourceAttribute().IsUnique() && relationship.GetTargetAttribute().IsUnique()
			// For same_as relationships, only assign up to min(source, target) rows
			// Excess rows in larger entity remain unassigned (valid for optional same_as)
			targetRowCount := relationship.GetTargetEntity().GetRowCount()
			// Process all rows for this relationship
			err := entity.ForEachRow(func(row *model.Row, rowIndex int) error {
				// For same_as relationships with source > target, skip excess rows
				if isSameAs && rowIndex >= targetRowCount {
					return nil // Skip - no corresponding target row exists
				}
				// For same_as relationships, always use round-robin (1:1 sequential mapping)
				// Power-law clustering doesn't make sense for identity relationships
				useAutoCardinality := autoCardinality && !isSameAs
				// Ask relationship to provide target PK value for this source row
				targetValue, err := relationship.GetTargetValueForSourceRow(rowIndex, useAutoCardinality)
				if err != nil {
					return fmt.Errorf("failed to get target value for row %d: %w", rowIndex, err)
				}
				// Set the FK value in the source row
				row.SetValue(relationship.GetSourceAttribute().GetName(), targetValue)
				// If this is the last FK for a junction table, check for duplicates
				if isLastRelationship && len(sourceRelationships) > 1 {
					// Check BEFORE registering - is this composite key already seen?
					if entity.IsCompositeKeyRegistered(row) {
						// DEBUG: Log duplicate detection
						// Duplicate - signal ForEachRow to remove this row
						return model.ErrSkipRow
					}
					// Not duplicate - register for future rows to check against
					entity.RegisterCompositeKey(row)
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to link relationship %s: %w", relationship.GetID(), err)
			}
			// Note: Duplicate removal now handled inline via ErrSkipRow
			// Eliminates O(n×m) RemoveRow calls and ~4s of slice copying for large datasets
		}
	}
	// Clear relationship linking progress line
	fmt.Printf("\r%-80s\r", "")
	return nil
}
