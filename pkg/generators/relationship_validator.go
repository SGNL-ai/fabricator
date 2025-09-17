package generators

import (
	"fmt"
	"strings"

	"github.com/SGNL-ai/fabricator/pkg/models"
)

// RelationshipValidationResult contains information about relationship validation
type RelationshipValidationResult struct {
	FromEntity     string
	ToEntity       string
	FromEntityFile string
	ToEntityFile   string
	Errors         []string
	TotalRows      int
	InvalidRows    int
}

// ValidateRelationships checks if the generated data maintains relationship integrity
func (g *CSVGenerator) ValidateRelationships() []RelationshipValidationResult {
	results := []RelationshipValidationResult{}

	// Validate each relationship
	for fromEntityID, links := range g.relationshipMap {
		for _, link := range links {
			result := g.validateRelationship(fromEntityID, link)
			if len(result.Errors) > 0 {
				results = append(results, result)
			}
		}
	}

	return results
}

// validateRelationship checks a single relationship for integrity
func (g *CSVGenerator) validateRelationship(fromEntityID string, link models.RelationshipLink) RelationshipValidationResult {
	// Get file names for both entities
	fromFileName := GetEntityFileName(g.EntityData[fromEntityID])
	toFileName := GetEntityFileName(g.EntityData[link.ToEntityID])

	result := RelationshipValidationResult{
		FromEntity:     fromEntityID,
		ToEntity:       link.ToEntityID,
		FromEntityFile: fromFileName,
		ToEntityFile:   toFileName,
		Errors:         []string{},
		TotalRows:      0,
		InvalidRows:    0,
	}

	// Get the data for both entities
	fromData := g.EntityData[fromEntityID]
	toData := g.EntityData[link.ToEntityID]

	if fromData == nil || toData == nil {
		result.Errors = append(result.Errors,
			fmt.Sprintf("Missing entity data (from: %s, to: %s)", fromEntityID, link.ToEntityID))
		return result
	}

	// Find the column indices for the attributes
	fromAttrIndex := -1
	fromAttrName := ""
	for i, header := range fromData.Headers {
		if strings.EqualFold(header, link.FromAttribute) ||
			strings.EqualFold(header, link.FromAttribute+"Id") {
			fromAttrIndex = i
			fromAttrName = header
			break
		}
	}

	toAttrIndex := -1
	for i, header := range toData.Headers {
		if strings.EqualFold(header, link.ToAttribute) ||
			strings.EqualFold(header, link.ToAttribute+"Id") {
			toAttrIndex = i
			break
		}
	}

	if fromAttrIndex == -1 || toAttrIndex == -1 {
		result.Errors = append(result.Errors,
			fmt.Sprintf("Could not find attribute columns (from: %s, to: %s)", link.FromAttribute, link.ToAttribute))
		return result
	}

	// Build a set of all valid target values
	validTargetValues := make(map[string]bool)
	for _, row := range toData.Rows {
		validTargetValues[row[toAttrIndex]] = true
	}

	// Determine the relationship direction based on field names
	isPrimaryKey := func(name string) bool {
		return name == "id" || strings.HasSuffix(strings.ToLower(name), "uuid") ||
			strings.HasSuffix(strings.ToLower(name), "guid")
	}

	isForeignKey := func(name string) bool {
		// Foreign keys usually contain "id" but are not just "id"
		return name != "id" && strings.Contains(strings.ToLower(name), "id")
	}

	// Check if this is a primary key to foreign key relationship
	fromIsPrimary := isPrimaryKey(link.FromAttribute)
	toIsForeign := isForeignKey(link.ToAttribute)

	// Case 1: Primary key in source entity, foreign key in target entity
	if fromIsPrimary && toIsForeign {
		// Collect all valid primary keys from source
		validSourceIds := make(map[string]bool)
		for _, row := range fromData.Rows {
			validSourceIds[row[fromAttrIndex]] = true
		}

		// Check all foreign keys in target reference valid source keys
		result.TotalRows = len(toData.Rows)
		for i, row := range toData.Rows {
			targetValue := row[toAttrIndex]
			if targetValue == "" {
				result.Errors = append(result.Errors,
					fmt.Sprintf("Row %d has empty foreign key in %s", i, link.ToAttribute))
				result.InvalidRows++
				continue
			}

			if !validSourceIds[targetValue] {
				result.Errors = append(result.Errors,
					fmt.Sprintf("Row %d has invalid reference: %s = %s", i, link.ToAttribute, targetValue))
				result.InvalidRows++
			}
		}
		return result
	}

	// Default case: Validate from source to target
	// Check that all source rows reference valid target values
	result.TotalRows = len(fromData.Rows)
	for i, row := range fromData.Rows {
		sourceValue := row[fromAttrIndex]
		if sourceValue == "" {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Row %d has empty value for attribute %s", i, fromAttrName))
			result.InvalidRows++
			continue
		}

		if !validTargetValues[sourceValue] {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Row %d has invalid reference: %s = %s", i, fromAttrName, sourceValue))
			result.InvalidRows++
		}
	}

	return result
}

// UniqueValueError represents an error with unique values in an entity
type UniqueValueError struct {
	EntityID   string
	EntityFile string
	Messages   []string
}

// ValidateUniqueValues ensures that attributes marked as uniqueId have unique values
func (g *CSVGenerator) ValidateUniqueValues() []UniqueValueError {
	results := []UniqueValueError{}

	// For each entity
	for entityID, csvData := range g.EntityData {
		// Get the file name for this entity
		entityFile := GetEntityFileName(csvData)

		// Get the unique attributes for this entity
		uniqueAttrs := g.uniqueIdAttributes[entityID]

		// Skip if there are no unique attributes
		if len(uniqueAttrs) == 0 {
			continue
		}

		// Initialize error structure for this entity
		entityErrors := UniqueValueError{
			EntityID:   entityID,
			EntityFile: entityFile,
			Messages:   []string{},
		}

		// For each unique attribute
		for _, uniqueAttr := range uniqueAttrs {
			// Find the index of this attribute in the headers
			attrIndex := -1
			for i, header := range csvData.Headers {
				if header == uniqueAttr {
					attrIndex = i
					break
				}
			}

			// Skip if we couldn't find the attribute
			if attrIndex == -1 {
				entityErrors.Messages = append(entityErrors.Messages,
					fmt.Sprintf("Could not find unique attribute %s in headers", uniqueAttr))
				continue
			}

			// Check that all values are unique
			usedValues := make(map[string]int)        // Map value to its row count
			duplicateValues := make(map[string][]int) // Map of value to list of row indices where it appears

			for i, row := range csvData.Rows {
				value := row[attrIndex]

				// Empty values are always a problem for uniqueId attributes
				if value == "" {
					entityErrors.Messages = append(entityErrors.Messages,
						fmt.Sprintf("Row %d has empty value for unique attribute %s", i+1, uniqueAttr))
					continue
				}

				// Track occurrences of this value
				usedValues[value]++

				// If we've seen this value before, add it to duplicates
				if usedValues[value] > 1 {
					if duplicateValues[value] == nil {
						// First find the initial occurrence
						for j, prevRow := range csvData.Rows[:i] {
							if prevRow[attrIndex] == value {
								duplicateValues[value] = []int{j + 1} // +1 for 1-based row numbering
								break
							}
						}
					}
					// Add current occurrence
					duplicateValues[value] = append(duplicateValues[value], i+1) // +1 for 1-based row numbering
				}
			}

			// Report duplicate values with detailed information
			if len(duplicateValues) > 0 {
				// Count total duplicates
				totalDuplicates := 0
				for _, rowIndices := range duplicateValues {
					totalDuplicates += len(rowIndices)
				}

				// Base message about duplicates
				entityErrors.Messages = append(entityErrors.Messages,
					fmt.Sprintf("Attribute %s has %d duplicate values", uniqueAttr, len(duplicateValues)))

				// Add detailed information for each duplicate value
				if len(duplicateValues) <= 5 { // Limit detail to avoid overwhelming output
					for value, rowIndices := range duplicateValues {
						// Show up to 5 rows where this value appears
						rowDisplay := ""
						if len(rowIndices) <= 5 {
							rowNumbers := make([]string, len(rowIndices))
							for i, rowIdx := range rowIndices {
								rowNumbers[i] = fmt.Sprintf("%d", rowIdx)
							}
							rowDisplay = strings.Join(rowNumbers, ", ")
						} else {
							rowNumbers := make([]string, 5)
							for i, rowIdx := range rowIndices[:5] {
								rowNumbers[i] = fmt.Sprintf("%d", rowIdx)
							}
							rowDisplay = strings.Join(rowNumbers, ", ") + "... (and " +
								fmt.Sprintf("%d", len(rowIndices)-5) + " more)"
						}

						// Truncate very long values
						displayValue := value
						if len(displayValue) > 30 {
							displayValue = displayValue[:27] + "..."
						}

						entityErrors.Messages = append(entityErrors.Messages,
							fmt.Sprintf("  - Value '%s' appears in rows: %s", displayValue, rowDisplay))
					}
				}
			}
		}

		// Add entity errors to results if we found any
		if len(entityErrors.Messages) > 0 {
			results = append(results, entityErrors)
		}
	}

	return results
}

// Helper function to get entity file paths
