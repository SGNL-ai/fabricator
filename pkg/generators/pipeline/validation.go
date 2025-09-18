package pipeline

import (
	"fmt"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
)

// Validator checks for data integrity and consistency in the generated data
type Validator struct {
	// Configuration options can be added here
}

// NewValidation creates a new validation component
func NewValidation() ValidatorInterface {
	return &Validator{}
}

// ValidateRelationships verifies graph-level relationship consistency
// This is for verification mode and structural validation
func (v *Validator) ValidateRelationships(graph *model.Graph) []string {
	var errors []string

	if graph == nil {
		errors = append(errors, "graph is nil - cannot validate relationships")
		return errors
	}

	// Validate graph structure - ensure relationships are properly defined
	for _, relationship := range graph.GetAllRelationships() {
		// Check that source and target entities exist
		if relationship.GetSourceEntity() == nil {
			errors = append(errors, fmt.Sprintf("relationship %s has nil source entity", relationship.GetID()))
			continue
		}

		if relationship.GetTargetEntity() == nil {
			errors = append(errors, fmt.Sprintf("relationship %s has nil target entity", relationship.GetID()))
			continue
		}

		// Check that source and target attributes exist
		if relationship.GetSourceAttribute() == nil {
			errors = append(errors, fmt.Sprintf("relationship %s has nil source attribute", relationship.GetID()))
			continue
		}

		if relationship.GetTargetAttribute() == nil {
			errors = append(errors, fmt.Sprintf("relationship %s has nil target attribute", relationship.GetID()))
			continue
		}

		// For verification mode: validate cross-entity referential integrity
		// Check that all foreign key values actually exist in target entity
		sourceEntity := relationship.GetSourceEntity()
		targetEntity := relationship.GetTargetEntity()
		sourceAttr := relationship.GetSourceAttribute()
		targetAttr := relationship.GetTargetAttribute()

		// Get all target values for quick lookup
		targetValues := make(map[string]bool)
		targetCSV := targetEntity.ToCSV()
		targetColIndex := -1
		for i, header := range targetCSV.Headers {
			if header == targetAttr.GetName() {
				targetColIndex = i
				break
			}
		}

		if targetColIndex >= 0 {
			for _, row := range targetCSV.Rows {
				if targetColIndex < len(row) {
					targetValues[row[targetColIndex]] = true
				}
			}
		}

		// Check source foreign key values
		sourceCSV := sourceEntity.ToCSV()
		sourceColIndex := -1
		for i, header := range sourceCSV.Headers {
			if header == sourceAttr.GetName() {
				sourceColIndex = i
				break
			}
		}

		if sourceColIndex >= 0 {
			for rowIdx, row := range sourceCSV.Rows {
				if sourceColIndex < len(row) {
					fkValue := row[sourceColIndex]
					if fkValue != "" && !targetValues[fkValue] {
						errors = append(errors, fmt.Sprintf("relationship %s: foreign key '%s' in %s (row %d) does not exist in %s.%s",
							relationship.GetID(), fkValue, sourceEntity.GetExternalID(), rowIdx, targetEntity.GetExternalID(), targetAttr.GetName()))
					}
				}
			}
		}
	}

	return errors
}
