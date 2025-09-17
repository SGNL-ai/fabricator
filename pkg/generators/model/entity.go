package model

import (
	"errors"
	"fmt"
	"strings"

	"github.com/SGNL-ai/fabricator/pkg/models"
)

// TODO make interface
// Row represents a single row of data for an entity

// Entity represents a data entity and manages its attributes and row data
type Entity struct {
	id                string
	externalID        string
	name              string
	description       string
	attributes        map[string]AttributeInterface // Map attribute name to attribute object
	attributesByExtID map[string]AttributeInterface // Map attribute external ID to attribute object
	attrList          []AttributeInterface          // Ordered list of attributes
	rows              []*Row
	primaryKey        AttributeInterface
	graph             GraphInterface // Reference to parent graph for lookups
}

// newEntity creates a new entity with basic properties and attributes
// Not exported as only Graph should create entities
func newEntity(id, externalID, name string, description string, attributes []AttributeInterface, graph GraphInterface) (EntityInterface, error) {
	// Create entity instance
	entity := &Entity{
		id:                id,
		externalID:        externalID,
		name:              name,
		description:       description,
		attributes:        make(map[string]AttributeInterface, len(attributes)),
		attributesByExtID: make(map[string]AttributeInterface, len(attributes)),
		attrList:          make([]AttributeInterface, 0, len(attributes)),
		rows:              make([]*Row, 0),
		graph:             graph,
	}

	// Add attributes to entity
	for _, attr := range attributes {
		if attr != nil {
			// Set parent entity reference
			attr.setParentEntity(entity)

			// Track primary key
			if attr.IsUnique() {
				entity.primaryKey = attr
			}

			// Add to attribute maps and list
			entity.attributes[attr.GetName()] = attr
			entity.attributesByExtID[attr.GetExternalID()] = attr
			if attr.GetAttributeAlias() != "" {
				entity.attributesByExtID[attr.GetAttributeAlias()] = attr
			}
			entity.attrList = append(entity.attrList, attr)
		}
	}

	// Validate the entity
	if err := entity.validate(); err != nil {
		return nil, err
	}

	return entity, nil
}

// validate checks that the entity and its attributes are valid
func (e *Entity) validate() error {
	// Check entity ID
	if e.id == "" {
		return errors.New("entity ID cannot be empty")
	}

	// Check external ID
	if e.externalID == "" {
		return errors.New("entity external ID cannot be empty")
	}

	// Check name
	if e.name == "" {
		return errors.New("entity name cannot be empty")
	}

	// If we have no attributes, that's valid
	if len(e.attributes) == 0 {
		return nil
	}

	// Check for duplicate attribute names
	attrNames := make(map[string]bool)
	for _, attr := range e.attrList {
		if attr == nil {
			return errors.New("attribute cannot be nil")
		}

		if _, exists := attrNames[attr.GetName()]; exists {
			return fmt.Errorf("duplicate attribute name: %s", attr.GetName())
		}
		attrNames[attr.GetName()] = true
	}

	// Count unique attributes for primary key check
	uniqueAttrs := 0
	for _, attr := range e.attrList {
		if attr.IsUnique() {
			uniqueAttrs++
		}
	}

	// Verify we have exactly one unique attribute
	if uniqueAttrs != 1 {
		return fmt.Errorf("entity must have exactly one unique attribute, found %d", uniqueAttrs)
	}

	// Verify primary key is set
	if e.primaryKey == nil {
		return errors.New("primary key attribute not set")
	}

	return nil
}

// GetID returns entity's internal ID
func (e *Entity) GetID() string {
	return e.id
}

// GetExternalID returns entity's external ID (used for CSV filenames)
func (e *Entity) GetExternalID() string {
	return e.externalID
}

// GetName returns entity's display name
func (e *Entity) GetName() string {
	return e.name
}

// GetDescription returns entity's description
func (e *Entity) GetDescription() string {
	return e.description
}

// GetAttributes returns all attributes in order
func (e *Entity) GetAttributes() []AttributeInterface {
	return e.attrList
}

// GetAttribute gets an attribute by name with existence check
func (e *Entity) GetAttribute(name string) (AttributeInterface, bool) {
	attr, exists := e.attributes[name]
	return attr, exists
}

// GetAttributeByExternalID gets an attribute by its external ID
func (e *Entity) GetAttributeByExternalID(externalID string) (AttributeInterface, bool) {

	// If attribute is prefixed by entity name, strip that
	if strings.Index(externalID, ".") > 0 {
		externalID = strings.TrimPrefix(externalID, fmt.Sprintf("%s.", e.externalID))
	}

	attr, exists := e.attributesByExtID[externalID]

	return attr, exists
}

// GetPrimaryKey returns the single unique attribute that serves as primary key
func (e *Entity) GetPrimaryKey() AttributeInterface {
	return e.primaryKey
}

// GetNonUniqueAttributes returns attributes not marked as unique
func (e *Entity) GetNonUniqueAttributes() []AttributeInterface {
	var result []AttributeInterface
	for _, attr := range e.attrList {
		if !attr.IsUnique() {
			result = append(result, attr)
		}
	}
	return result
}

// GetRelationshipAttributes returns attributes involved in relationships
func (e *Entity) GetRelationshipAttributes() []AttributeInterface {
	var result []AttributeInterface
	for _, attr := range e.attrList {
		if attr.IsRelationship() {
			result = append(result, attr)
		}
	}
	return result
}

// GetNonRelationshipAttributes returns attributes not involved in relationships
func (e *Entity) GetNonRelationshipAttributes() []AttributeInterface {
	var result []AttributeInterface
	for _, attr := range e.attrList {
		if !attr.IsRelationship() {
			result = append(result, attr)
		}
	}
	return result
}

// GetRowCount returns the number of rows
func (e *Entity) GetRowCount() int {
	return len(e.rows)
}

// AddRow adds a new row with provided values
// Validates uniqueness constraint for primary key
// Validates foreign key values exist in related entities
func (e *Entity) AddRow(row *Row) error {
	// Validate row values against entity constraints
	if err := e.validateRow(row); err != nil {
		return err
	}

	// Add row to entity
	e.rows = append(e.rows, row)

	return nil
}

// ForEachRow iterates over all rows and allows modification during iteration
// Any changes made to the row are validated using AddRow validation
// Row order is preserved after iteration completes
func (e *Entity) ForEachRow(fn func(row *Row) error) error {
	originalRowCount := len(e.rows)

	// Process each row by popping from front and re-adding at end
	for i := range originalRowCount {
		// Pop the first row
		row := e.rows[0]
		e.rows = e.rows[1:]

		// Call the function to potentially modify the row
		if err := fn(row); err != nil {
			return fmt.Errorf("error processing row %d in entity %s: %w", i, e.name, err)
		}

		// Re-add the row using AddRow for validation
		if err := e.AddRow(row); err != nil {
			return err
		}
	}
	return nil
}


// ToCSV returns CSV representation of the entity
func (e *Entity) ToCSV() *models.CSVData {
	// Create headers from attribute external IDs
	headers := make([]string, 0, len(e.attrList))
	for _, attr := range e.attrList {
		headers = append(headers, attr.GetName())
	}

	// Create rows from entity data
	csvRows := make([][]string, 0, len(e.rows))
	for _, row := range e.rows {
		csvRow := make([]string, 0, len(headers))
		for _, attrName := range headers {
			csvRow = append(csvRow, row.values[attrName])
		}
		csvRows = append(csvRows, csvRow)
	}

	// Create CSV data
	return &models.CSVData{
		ExternalId:  e.externalID,
		Headers:     headers,
		Rows:        csvRows,
		EntityName:  e.name,
		Description: e.description,
	}
}

// validateRow validates a row against entity constraints
func (e *Entity) validateRow(row *Row) error {
	// If we have no primary key, we can't validate rows
	if e.primaryKey == nil && len(e.attributes) > 0 {
		return errors.New("entity has no primary key attribute defined")
	}

	// Check that all required values are provided
	if e.primaryKey != nil {
		pkName := e.primaryKey.GetName()
		pkValue, exists := row.values[pkName]

		// Primary key is required
		if !exists || pkValue == "" {
			return fmt.Errorf("missing required primary key value for attribute '%s'", pkName)
		}

		// Check uniqueness constraint (simple check since ForEachRow pops the row)
		for _, existingRow := range e.rows {
			if existingRow.values[pkName] == pkValue {
				return fmt.Errorf("duplicate value '%s' for unique attribute '%s'", pkValue, pkName)
			}
		}
	}

	// Validate foreign key references (only if values are provided)
	for _, attr := range e.GetRelationshipAttributes() {
		attrName := attr.GetName()
		value, exists := row.values[attrName]

		// Foreign key values are optional - only validate if provided
		if exists && value != "" {
			// Validate foreign key references
			if err := e.validateForeignKeyValue(attrName, value); err != nil {
				return err
			}
		}
	}

	return nil
}



// validateForeignKeyValue verifies that a foreign key value exists in the related entity
func (e *Entity) validateForeignKeyValue(attributeName string, value string) error {
	// Get the attribute
	attr, exists := e.GetAttribute(attributeName)
	if !exists {
		return fmt.Errorf("attribute '%s' not found", attributeName)
	}

	// Check if it's a relationship attribute
	if !attr.IsRelationship() {
		return nil
	}

	// Get information about the related entity and attribute
	relatedEntityID := attr.GetRelatedEntityID()
	relatedAttributeName := attr.GetRelatedAttribute()

	// Find the related entity through the graph
	relatedEntity, exists := e.graph.GetEntity(relatedEntityID)
	if !exists {
		return fmt.Errorf("related entity '%s' not found for foreign key validation", relatedEntityID)
	}

	// Check if the target attribute exists in the related entity
	_, exists = relatedEntity.GetAttribute(relatedAttributeName)
	if !exists {
		return fmt.Errorf("related attribute '%s' not found in entity '%s'",
			relatedAttributeName, relatedEntityID)
	}

	// Validate that the foreign key value exists in the related entity's rows
	// For a valid foreign key reference, the value should exist in the related entity's attribute
	valueFound := false
	for _, row := range relatedEntity.getRows() {
		if attrValue, exists := row.values[relatedAttributeName]; exists && attrValue == value {
			valueFound = true
			break
		}
	}

	if !valueFound {
		return fmt.Errorf("foreign key value '%s' does not exist in related entity '%s.%s'",
			value, relatedEntityID, relatedAttributeName)
	}

	return nil
}

// addRelationship creates a relationship between this entity and another entity
// This follows the architecture described in the refactoring plan
// Not exported as it should only be called by the Graph when building relationships
func (e *Entity) addRelationship(
	relationshipID, relationshipName string,
	targetEntity EntityInterface,
	sourceExternalID, targetExternalID string) (RelationshipInterface, error) {

	// Find source attribute - try both UUID alias and dotted notation
	sourceAttr, sourceExists := e.findAttributeByReference(sourceExternalID)
	if !sourceExists {
		return nil, fmt.Errorf("source attribute '%s' not found in entity '%s'",
			sourceExternalID, e.GetName())
	}

	// Find target attribute - try both UUID alias and dotted notation
	targetAttr, targetExists := targetEntity.findAttributeByReference(targetExternalID)
	if !targetExists {
		return nil, fmt.Errorf("target attribute '%s' not found in entity '%s'",
			targetExternalID, targetEntity.GetName())
	}

	// Create relationship using the constructor
	relationship, err := newRelationship(
		relationshipID,
		relationshipName,
		e,                    // source entity
		targetEntity,         // target entity
		sourceAttr.GetName(), // source attribute name
		targetAttr.GetName(), // target attribute name
	)

	if err != nil {
		return nil, err
	}

	// Set only the source attribute as a foreign key relationship
	// The target attribute remains as a regular attribute (likely a primary key)
	// CRITICAL: Never mark unique attributes as relationship sources
	if !sourceAttr.IsUnique() {
		sourceAttr.setRelationship(targetEntity.GetID(), targetAttr.GetName())
	}

	// Don't set bidirectional relationship - target attribute is not a FK

	return relationship, nil
}

func (e *Entity) getRows() []*Row {
	return e.rows
}

// findAttributeByReference finds an attribute using either UUID alias or dotted notation
func (e *Entity) findAttributeByReference(reference string) (AttributeInterface, bool) {
	// First try: look up by attributeAlias (UUID format)
	for _, attr := range e.attrList {
		if attr.GetAttributeAlias() == reference {
			return attr, true
		}
	}

	// Second try: extract from dotted notation and look up by externalID
	if strings.Contains(reference, ".") {
		parts := strings.Split(reference, ".")
		attrName := parts[len(parts)-1] // Get the part after the last dot
		return e.GetAttributeByExternalID(attrName)
	}

	// Third try: direct lookup by externalID (for simple names)
	return e.GetAttributeByExternalID(reference)
}
