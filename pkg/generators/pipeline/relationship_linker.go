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

	// Process relationships to establish foreign key links
	for _, relationship := range graph.GetAllRelationships() {
		sourceEntity := relationship.GetSourceEntity()
		targetEntity := relationship.GetTargetEntity()
		sourceAttr := relationship.GetSourceAttribute()
		targetAttr := relationship.GetTargetAttribute()

		if sourceEntity == nil || targetEntity == nil || sourceAttr == nil || targetAttr == nil {
			continue // Skip invalid relationships
		}

		// Establish FK relationship between entities
		err := l.linkEntityRelationship(sourceEntity, targetEntity, sourceAttr, targetAttr, relationship, autoCardinality)
		if err != nil {
			return fmt.Errorf("failed to link relationship %s: %w", relationship.GetID(), err)
		}
	}

	return nil
}

// linkEntityRelationship establishes foreign key relationships between two entities
func (l *RelationshipLinker) linkEntityRelationship(source, target model.EntityInterface, sourceAttr, targetAttr model.AttributeInterface, relationship model.RelationshipInterface, autoCardinality bool) error {
	// Get target entity's values for linking
	targetCSV := target.ToCSV()
	targetValues := make([]string, 0)

	// Find target attribute column
	targetColIndex := -1
	for i, header := range targetCSV.Headers {
		if header == targetAttr.GetName() {
			targetColIndex = i
			break
		}
	}

	// Collect available target values
	if targetColIndex >= 0 {
		for _, row := range targetCSV.Rows {
			if targetColIndex < len(row) && row[targetColIndex] != "" {
				targetValues = append(targetValues, row[targetColIndex])
			}
		}
	}

	// If no target values available, skip linking (leave FK fields empty)
	if len(targetValues) == 0 {
		return nil
	}

	// Determine FK distribution strategy based on cardinality
	var getTargetValue func(rowIndex int) string

	if autoCardinality {
		// Use relationship cardinality to determine distribution
		if relationship.IsOneToOne() {
			// 1:1 - each source row gets unique target value
			getTargetValue = func(rowIndex int) string {
				if rowIndex < len(targetValues) {
					return targetValues[rowIndex]
				}
				return targetValues[rowIndex%len(targetValues)] // fallback to round-robin
			}
		} else if relationship.IsOneToMany() {
			// 1:N - multiple source rows can reference same target
			getTargetValue = func(rowIndex int) string {
				return targetValues[rowIndex%len(targetValues)] // round-robin
			}
		} else {
			// Default: round-robin distribution
			getTargetValue = func(rowIndex int) string {
				return targetValues[rowIndex%len(targetValues)]
			}
		}
	} else {
		// Simple round-robin distribution when autoCardinality is disabled
		getTargetValue = func(rowIndex int) string {
			return targetValues[rowIndex%len(targetValues)]
		}
	}

	// Use iterator to set FK values in source entity rows
	rowIndex := 0
	return source.ForEachRow(func(row *model.Row) error {
		// Pick target value based on cardinality strategy
		targetValue := getTargetValue(rowIndex)

		// Set the foreign key value
		row.SetValue(sourceAttr.GetName(), targetValue)

		rowIndex++
		return nil
	})
}
