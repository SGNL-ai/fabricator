package generators

import (
	"fmt"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/SGNL-ai/fabricator/pkg/parser"
)

// CardinalityWarning represents a detected cardinality violation.
// This occurs when row counts make it impossible to satisfy relationship
// cardinality constraints (e.g., 100 parents need children but only 50 children exist).
type CardinalityWarning struct {
	// RelationshipName identifies the relationship
	RelationshipName string

	// SourceEntity is the "one" side of the relationship
	SourceEntity string

	// SourceCount is how many source entity rows exist
	SourceCount int

	// TargetEntity is the "many" side of the relationship
	TargetEntity string

	// TargetCount is how many target entity rows exist
	TargetCount int

	// Cardinality is the expected cardinality (e.g., "one-to-many")
	Cardinality string

	// Shortfall describes the gap (e.g., "50 departments need 1+ employee but only 30 employees exist")
	Shortfall string
}

// String formats the warning for display.
// Provides a human-readable description of the cardinality violation.
func (w *CardinalityWarning) String() string {
	return fmt.Sprintf(
		"Cardinality warning: Relationship '%s' (%s) - %s has %d rows but %s has %d rows. %s",
		w.RelationshipName,
		w.Cardinality,
		w.SourceEntity,
		w.SourceCount,
		w.TargetEntity,
		w.TargetCount,
		w.Shortfall,
	)
}

// DetectCardinalityViolations analyzes row counts to detect potential imbalances.
// Returns a list of warnings when entity row counts show significant imbalances
// that might affect relationship quality.
//
// Note: This is a best-effort heuristic detection based on configured row counts.
// Actual relationship assignment happens in the pipeline and uses best-effort logic.
// We detect imbalances of >10x ratio between any pair of entities.
func DetectCardinalityViolations(graph *model.Graph, def *parser.SORDefinition, rowCounts map[string]int) []CardinalityWarning {
	warnings := make([]CardinalityWarning, 0)

	// Simple heuristic: Warn if any entity pair has >10x ratio in row counts
	// This indicates potential relationship assignment issues
	entityIDs := make([]string, 0, len(def.Entities))
	for entityID := range def.Entities {
		entityIDs = append(entityIDs, entityID)
	}

	// Check pairs of entities that might have relationships
	for i := 0; i < len(entityIDs); i++ {
		for j := i + 1; j < len(entityIDs); j++ {
			entity1 := entityIDs[i]
			entity2 := entityIDs[j]
			count1 := rowCounts[entity1]
			count2 := rowCounts[entity2]

			// Warn if ratio > 10x in either direction
			if count1 > count2*10 {
				warnings = append(warnings, CardinalityWarning{
					RelationshipName: fmt.Sprintf("%s-%s", entity1, entity2),
					SourceEntity:     entity1,
					SourceCount:      count1,
					TargetEntity:     entity2,
					TargetCount:      count2,
					Cardinality:      "potential",
					Shortfall:        fmt.Sprintf("Significant imbalance: %d %s rows vs %d %s rows (ratio >10x) may affect relationship quality", count1, entity1, count2, entity2),
				})
			} else if count2 > count1*10 {
				warnings = append(warnings, CardinalityWarning{
					RelationshipName: fmt.Sprintf("%s-%s", entity2, entity1),
					SourceEntity:     entity2,
					SourceCount:      count2,
					TargetEntity:     entity1,
					TargetCount:      count1,
					Cardinality:      "potential",
					Shortfall:        fmt.Sprintf("Significant imbalance: %d %s rows vs %d %s rows (ratio >10x) may affect relationship quality", count2, entity2, count1, entity1),
				})
			}
		}
	}

	return warnings
}
