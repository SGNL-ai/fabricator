package model

import (
	"errors"
	"fmt"
	"strings"
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
	graph             GraphInterface  // Reference to parent graph for lookups
	usedPKValues      map[string]bool // Track used primary key values for O(1) duplicate detection
	usedCompositeKeys map[string]bool // Track used composite FK keys for junction table duplicate prevention
}

// newEntity creates a new entity with basic properties and attributes
// Not exported as only Graph should create entities
func newEntity(id, externalID, name string, description string, attributes []AttributeInterface, graph GraphInterface) (EntityInterface, error) {
	// Get expected row count from graph for memory optimization
	expectedRows := 0
	if graph != nil {
		expectedRows = graph.GetExpectedDataVolume()
		// Use reasonable default for tests/small datasets
		if expectedRows <= 0 {
			expectedRows = 100
		}
	}

	// Create entity instance with pre-allocated slices
	entity := &Entity{
		id:                id,
		externalID:        externalID,
		name:              name,
		description:       description,
		attributes:        make(map[string]AttributeInterface, len(attributes)),
		attributesByExtID: make(map[string]AttributeInterface, len(attributes)),
		attrList:          make([]AttributeInterface, 0, len(attributes)),
		rows:              make([]*Row, 0, expectedRows), // Pre-allocate with expected capacity
		graph:             graph,
		usedPKValues:      make(map[string]bool, expectedRows), // Pre-allocate hash map
		usedCompositeKeys: make(map[string]bool, expectedRows), // Pre-allocate composite key index
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

	// Track the primary key value in our hash map for future duplicate detection
	if e.primaryKey != nil {
		pkName := e.primaryKey.GetName()
		if pkValue, exists := row.values[pkName]; exists && pkValue != "" {
			e.usedPKValues[pkValue] = true
		}
	}

	// Track composite FK key for junction tables
	e.addCompositeKeyToIndex(row)

	return nil
}

// ForEachRow iterates over all rows and allows modification during iteration
// Any changes made to the row are validated using AddRow validation
// Row order is preserved after iteration completes
func (e *Entity) ForEachRow(fn func(row *Row, index int) error) error {
	originalRowCount := len(e.rows)

	// Process each row by temporarily removing and re-adding for atomic validation
	for i := range originalRowCount {
		// Get the first row (don't pop yet to maintain entity state during fn call)
		row := e.rows[0]

		// Capture original PK value before modification
		var originalPKValue string
		if e.primaryKey != nil {
			pkName := e.primaryKey.GetName()
			if pkValue, exists := row.values[pkName]; exists {
				originalPKValue = pkValue
			}
		}

		// Call the function to potentially modify the row
		if err := fn(row, i); err != nil {
			return fmt.Errorf("error processing row %d in entity %s: %w", i, e.name, err)
		}

		// Now atomically pop the old row and re-add the modified row
		e.rows = e.rows[1:] // Remove the processed row

		// Remove original PK value from hash map (before modification)
		if e.primaryKey != nil && originalPKValue != "" {
			delete(e.usedPKValues, originalPKValue)
		}

		// Re-add the modified row using AddRow for validation
		if err := e.AddRow(row); err != nil {
			return fmt.Errorf("failed to re-add modified row %d in entity %s: %w", i, e.name, err)
		}
	}
	return nil
}

// ToCSV returns CSV representation of the entity
func (e *Entity) ToCSV() *CSVData {
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
	return &CSVData{
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

		// Check uniqueness constraint using O(1) hash map lookup
		if e.usedPKValues[pkValue] {
			return fmt.Errorf("duplicate value '%s' for unique attribute '%s'", pkValue, pkName)
		}
	}

	return nil
}

// PreAllocateRows pre-allocates the rows slice for better memory performance
func (e *Entity) PreAllocateRows(expectedRowCount int) {
	if cap(e.rows) < expectedRowCount {
		// Pre-allocate with exact capacity to avoid slice growth
		e.rows = make([]*Row, 0, expectedRowCount)
		// Also pre-allocate the PK hash map with expected size
		e.usedPKValues = make(map[string]bool, expectedRowCount)
	}
}

// RemoveRow removes a row from the entity and updates hash maps
func (e *Entity) RemoveRow(rowIndex int) error {
	if rowIndex < 0 || rowIndex >= len(e.rows) {
		return fmt.Errorf("row index %d out of range [0, %d)", rowIndex, len(e.rows))
	}

	row := e.rows[rowIndex]

	// Remove PK value from hash map if it exists
	if e.primaryKey != nil {
		pkName := e.primaryKey.GetName()
		if pkValue := row.GetValue(pkName); pkValue != "" {
			delete(e.usedPKValues, pkValue)
		}
	}

	// Remove composite key from hash map if it exists
	fkAttributes := e.GetRelationshipAttributes()
	if len(fkAttributes) > 1 {
		compositeKey := e.buildCompositeKey(row, fkAttributes)
		delete(e.usedCompositeKeys, compositeKey)
	}

	// Remove row from slice
	e.rows = append(e.rows[:rowIndex], e.rows[rowIndex+1:]...)

	return nil
}

// GetRowByIndex returns a row by index for direct access (O(1) operation)
func (e *Entity) GetRowByIndex(index int) *Row {
	if index < 0 || index >= len(e.rows) {
		return nil
	}
	return e.rows[index]
}

// CheckKeyExists checks if a key value exists in the used primary key values (O(1) lookup)
func (e *Entity) CheckKeyExists(keyValue string) bool {
	return e.usedPKValues[keyValue]
}

// IsForeignKeyUnique checks if a row's FK combination is unique
// Returns true if unique, false if duplicate. Caller decides when to use this.
func (e *Entity) IsForeignKeyUnique(row *Row) bool {
	fkAttributes := e.GetRelationshipAttributes()
	if len(fkAttributes) == 0 {
		return true // No FK attributes - always unique
	}

	compositeKey := e.buildCompositeKey(row, fkAttributes)
	return !e.usedCompositeKeys[compositeKey]
}

// addCompositeKeyToIndex adds a row's composite FK key to the internal index
// Should only be called after verifying the key is unique
func (e *Entity) addCompositeKeyToIndex(row *Row) {
	fkAttributes := e.GetRelationshipAttributes()
	if len(fkAttributes) == 0 {
		return // No FK attributes to index
	}

	compositeKey := e.buildCompositeKey(row, fkAttributes)
	e.usedCompositeKeys[compositeKey] = true
}

// buildCompositeKey creates a composite key string from FK attribute values
func (e *Entity) buildCompositeKey(row *Row, fkAttributes []AttributeInterface) string {
	compositeKey := ""
	for i, fkAttr := range fkAttributes {
		if i > 0 {
			compositeKey += "|"
		}
		compositeKey += row.GetValue(fkAttr.GetName())
	}
	return compositeKey
}

// ValidateAllForeignKeys validates all FK relationships for this entity (post-generation validation)
func (e *Entity) ValidateAllForeignKeys() []string {
	var errors []string

	// Check all relationship attributes, including unique ones
	for _, attr := range e.GetRelationshipAttributes() {
		attrName := attr.GetName()

		// Check each row's FK value
		for i, row := range e.rows {
			value, exists := row.values[attrName]
			if exists && value != "" {
				if err := e.validateForeignKeyValue(attrName, value); err != nil {
					errors = append(errors, fmt.Sprintf("row %d: %v", i, err))
				}
			}
		}
	}

	return errors
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

	// Validate that the foreign key value exists in the related entity using O(1) lookup
	// For primary key attributes, use the efficient CheckKeyExists method
	relatedAttr, _ := relatedEntity.GetAttribute(relatedAttributeName)
	if relatedAttr != nil && relatedAttr.IsUnique() {
		// For primary keys, use O(1) hash map lookup
		if !relatedEntity.CheckKeyExists(value) {
			return fmt.Errorf("foreign key value '%s' does not exist in related entity '%s.%s'",
				value, relatedEntityID, relatedAttributeName)
		}
	} else {
		// For non-unique attributes, fall back to linear search (rare case)
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
	sourceAttr.setRelationship(targetEntity.GetID(), targetAttr.GetName())

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
