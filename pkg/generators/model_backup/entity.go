package model

import (
	"fmt"
	"strings"

	"github.com/SGNL-ai/fabricator/pkg/models"
)

// Entity represents a data entity during the generation process
type Entity struct {
	// Core identity
	id         string // Internal ID used for graph operations
	externalID string // External ID from YAML (used for CSV file names)
	name       string // Display name

	// Structure
	attributes     map[string]*Attribute // Map of attribute name to attribute
	attributeOrder []string              // Preserves attribute order
	rows           []*Row                // Generated rows of data

	// Relationships
	incomingRelationships []*Relationship // Relationships where this entity is the target
	outgoingRelationships []*Relationship // Relationships where this entity is the source

	// Metadata
	description string            // Entity description
	properties  map[string]string // Additional properties
}

// Row represents a single row of data for an entity
type Row struct {
	entityID string             // Reference to parent entity
	index    int                // Row index
	values   map[string]string  // Map of attribute name to value
}

// Attribute represents an entity attribute
type Attribute struct {
	// Core identity 
	name       string // Attribute name
	externalID string // External ID from YAML

	// Properties
	dataType    string // Data type
	description string // Attribute description
	isUnique    bool   // Whether the attribute requires unique values
	isID        bool   // Whether the attribute is an identifier
	
	// Generation metadata
	isRelationship   bool   // Whether this attribute is part of a relationship
	relatedEntityID  string // If part of a relationship, the related entity
	relatedAttribute string // If part of a relationship, the related attribute
}

// Relationship represents a relationship between entities
type Relationship struct {
	// Core identity
	id         string    // Relationship ID
	name       string    // Relationship name
	
	// Endpoints
	sourceEntity    *Entity    // Source entity
	targetEntity    *Entity    // Target entity
	sourceAttribute *Attribute // Source attribute
	targetAttribute *Attribute // Target attribute
	
	// Properties
	cardinality string // Relationship cardinality (1:1, 1:N, N:1)
	description string // Relationship description
}

// GetID returns the entity's ID
func (e *Entity) GetID() string {
	return e.id
}

// GetExternalID returns the entity's external ID
func (e *Entity) GetExternalID() string {
	return e.externalID
}

// GetName returns the entity's name
func (e *Entity) GetName() string {
	return e.name
}

// GetDescription returns the entity's description
func (e *Entity) GetDescription() string {
	return e.description
}

// GetAttribute returns the attribute with the given name
func (e *Entity) GetAttribute(name string) *Attribute {
	return e.attributes[name]
}

// GetAttributes returns all attributes in order
func (e *Entity) GetAttributes() []*Attribute {
	result := make([]*Attribute, len(e.attributeOrder))
	for i, name := range e.attributeOrder {
		result[i] = e.attributes[name]
	}
	return result
}

// GetAttributesByType returns attributes filtered by a predicate function
func (e *Entity) GetAttributesByType(filter func(*Attribute) bool) []*Attribute {
	result := make([]*Attribute, 0)
	for _, name := range e.attributeOrder {
		attr := e.attributes[name]
		if filter(attr) {
			result = append(result, attr)
		}
	}
	return result
}

// GetIDAttributes returns all attributes that are identifiers
func (e *Entity) GetIDAttributes() []*Attribute {
	return e.GetAttributesByType(func(a *Attribute) bool { return a.isID })
}

// GetRelationshipAttributes returns all attributes that are part of relationships
func (e *Entity) GetRelationshipAttributes() []*Attribute {
	return e.GetAttributesByType(func(a *Attribute) bool { return a.isRelationship })
}

// GetNonRelationshipAttributes returns all attributes that are not part of relationships
func (e *Entity) GetNonRelationshipAttributes() []*Attribute {
	return e.GetAttributesByType(func(a *Attribute) bool { return !a.isRelationship })
}

// GetUniqueAttributes returns all attributes that require unique values
func (e *Entity) GetUniqueAttributes() []*Attribute {
	return e.GetAttributesByType(func(a *Attribute) bool { return a.isUnique })
}

// GetHeaderNames returns attribute names in order (for CSV headers)
func (e *Entity) GetHeaderNames() []string {
	return e.attributeOrder
}

// CreateEmptyRows creates the specified number of empty rows
func (e *Entity) CreateEmptyRows(count int) {
	e.rows = make([]*Row, count)
	for i := 0; i < count; i++ {
		e.rows[i] = &Row{
			entityID: e.id,
			index:    i,
			values:   make(map[string]string),
		}
	}
}

// GetRow returns the row at the specified index
func (e *Entity) GetRow(index int) *Row {
	if index < 0 || index >= len(e.rows) {
		return nil
	}
	return e.rows[index]
}

// GetRows returns all rows
func (e *Entity) GetRows() []*Row {
	return e.rows
}

// GetRowCount returns the number of rows
func (e *Entity) GetRowCount() int {
	return len(e.rows)
}

// GetRelationships returns all relationships for this entity
func (e *Entity) GetRelationships() []*Relationship {
	result := make([]*Relationship, 0, len(e.incomingRelationships)+len(e.outgoingRelationships))
	result = append(result, e.incomingRelationships...)
	result = append(result, e.outgoingRelationships...)
	return result
}

// GetIncomingRelationships returns relationships where this entity is the target
func (e *Entity) GetIncomingRelationships() []*Relationship {
	return e.incomingRelationships
}

// GetOutgoingRelationships returns relationships where this entity is the source
func (e *Entity) GetOutgoingRelationships() []*Relationship {
	return e.outgoingRelationships
}

// ToCSVData converts the entity to the CSVData model format
func (e *Entity) ToCSVData() *models.CSVData {
	// Extract headers - use external IDs for CSV
	headers := make([]string, len(e.attributeOrder))
	for i, name := range e.attributeOrder {
		headers[i] = e.attributes[name].externalID
	}
	
	// Convert rows to string slices for CSV
	csvRows := make([][]string, len(e.rows))
	for i, row := range e.rows {
		rowData := make([]string, len(headers))
		for j, attrName := range e.attributeOrder {
			if value, ok := row.values[attrName]; ok {
				rowData[j] = value
			} else {
				rowData[j] = "" // Empty string for missing values
			}
		}
		csvRows[i] = rowData
	}
	
	return &models.CSVData{
		ExternalId:  e.externalID,
		EntityName:  e.name,
		Description: e.description,
		Headers:     headers,
		Rows:        csvRows,
	}
}

// GetValue returns a row value by attribute name
func (r *Row) GetValue(attrName string) string {
	return r.values[attrName]
}

// SetValue sets a row value by attribute name
func (r *Row) SetValue(attrName, value string) {
	r.values[attrName] = value
}

// GetIndex returns the row index
func (r *Row) GetIndex() int {
	return r.index
}

// GetValues returns a copy of all values
func (r *Row) GetValues() map[string]string {
	result := make(map[string]string)
	for k, v := range r.values {
		result[k] = v
	}
	return result
}

// GetName returns the attribute name
func (a *Attribute) GetName() string {
	return a.name
}

// GetExternalID returns the attribute external ID
func (a *Attribute) GetExternalID() string {
	return a.externalID
}

// GetDataType returns the attribute data type
func (a *Attribute) GetDataType() string {
	return a.dataType
}

// IsUnique returns whether values must be unique
func (a *Attribute) IsUnique() bool {
	return a.isUnique
}

// IsID returns whether the attribute is an identifier
func (a *Attribute) IsID() bool {
	return a.isID
}

// IsRelationship returns whether the attribute is part of a relationship
func (a *Attribute) IsRelationship() bool {
	return a.isRelationship
}

// GetRelatedEntityID returns the related entity ID (if part of a relationship)
func (a *Attribute) GetRelatedEntityID() string {
	return a.relatedEntityID
}

// GetRelatedAttribute returns the related attribute name (if part of a relationship)
func (a *Attribute) GetRelatedAttribute() string {
	return a.relatedAttribute
}

// GetIsIdentifier returns true if the attribute is an identifier
func (a *Attribute) GetIsIdentifier() bool {
	return a.isID || a.isUnique
}

// GetID returns the relationship ID
func (r *Relationship) GetID() string {
	return r.id
}

// GetName returns the relationship name
func (r *Relationship) GetName() string {
	return r.name
}

// GetSourceEntity returns the source entity
func (r *Relationship) GetSourceEntity() *Entity {
	return r.sourceEntity
}

// GetTargetEntity returns the target entity
func (r *Relationship) GetTargetEntity() *Entity {
	return r.targetEntity
}

// GetSourceAttribute returns the source attribute
func (r *Relationship) GetSourceAttribute() *Attribute {
	return r.sourceAttribute
}

// GetTargetAttribute returns the target attribute
func (r *Relationship) GetTargetAttribute() *Attribute {
	return r.targetAttribute
}

// GetCardinality returns the relationship cardinality
func (r *Relationship) GetCardinality() string {
	return r.cardinality
}

// SetCardinality sets the relationship cardinality
func (r *Relationship) SetCardinality(cardinality string) {
	r.cardinality = cardinality
}

// GetDescription returns the relationship description
func (r *Relationship) GetDescription() string {
	return r.description
}

// IsOneToOne returns true if the relationship is 1:1
func (r *Relationship) IsOneToOne() bool {
	return r.cardinality == "1:1"
}

// IsOneToMany returns true if the relationship is 1:N
func (r *Relationship) IsOneToMany() bool {
	return r.cardinality == "1:N"
}

// IsManyToOne returns true if the relationship is N:1
func (r *Relationship) IsManyToOne() bool {
	return r.cardinality == "N:1"
}

// String helper methods for debugging

// String returns a string representation of the entity
func (e *Entity) String() string {
	return fmt.Sprintf("Entity[%s] (%s) with %d attributes, %d rows", 
		e.id, e.externalID, len(e.attributes), len(e.rows))
}

// String returns a string representation of the attribute
func (a *Attribute) String() string {
	return fmt.Sprintf("Attribute[%s] (%s) type=%s, isID=%v, isUnique=%v, isRelationship=%v", 
		a.name, a.externalID, a.dataType, a.isID, a.isUnique, a.isRelationship)
}

// String returns a string representation of the relationship
func (r *Relationship) String() string {
	return fmt.Sprintf("Relationship[%s] %s.%s (%s) %s.%s", 
		r.id, r.sourceEntity.id, r.sourceAttribute.name, 
		r.cardinality, r.targetEntity.id, r.targetAttribute.name)
}