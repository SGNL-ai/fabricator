package model

import (
	"github.com/SGNL-ai/fabricator/pkg/parser"
)

type Row struct {
	values map[string]string
}

// NewRow creates a new Row with the given values
func NewRow(values map[string]string) *Row {
	return &Row{values: values}
}

// SetValue updates a field value in the row
func (r *Row) SetValue(fieldName, value string) {
	if r.values == nil {
		r.values = make(map[string]string)
	}
	r.values[fieldName] = value
}

// GetValue gets a field value from the row
func (r *Row) GetValue(fieldName string) string {
	if r.values == nil {
		return ""
	}
	return r.values[fieldName]
}

// GraphInterface defines the operations that can be performed on a Graph
type GraphInterface interface {
	GetEntity(id string) (EntityInterface, bool)
	GetAllEntities() map[string]EntityInterface
	GetEntitiesList() []EntityInterface
	GetRelationship(id string) (RelationshipInterface, bool)
	GetAllRelationships() []RelationshipInterface
	GetRelationshipsForEntity(entityID string) []RelationshipInterface
	GetTopologicalOrder() ([]string, error)
	GetExpectedDataVolume() int // For memory optimization

	createEntitiesFromYAML(yamlEntities map[string]parser.Entity) error
}

// EntityInterface defines the operations that can be performed on an Entity
type EntityInterface interface {
	GetID() string
	GetExternalID() string
	GetName() string
	GetDescription() string
	GetAttributes() []AttributeInterface
	GetAttribute(name string) (AttributeInterface, bool)
	GetAttributeByExternalID(externalID string) (AttributeInterface, bool)
	GetPrimaryKey() AttributeInterface
	GetNonUniqueAttributes() []AttributeInterface
	GetRelationshipAttributes() []AttributeInterface
	GetNonRelationshipAttributes() []AttributeInterface
	GetRowCount() int
	AddRow(row *Row) error
	ForEachRow(fn func(row *Row, index int) error) error
	ToCSV() *CSVData

	// Internal method for relationships
	addRelationship(relationshipID, relationshipName string,
		targetEntity EntityInterface, sourceExternalID, targetExternalID string) (RelationshipInterface, error)

	// Post-generation validation
	ValidateAllForeignKeys() []string

	// Helper for foreign key validation
	validateForeignKeyValue(attributeName string, value string) error

	// Helper for attribute lookup by reference (UUID alias or dotted notation)
	findAttributeByReference(reference string) (AttributeInterface, bool)

	// Returns rows of data
	getRows() []*Row

	// Performance optimizations
	GetRowByIndex(index int) *Row        // Direct row access by index (O(1))
	CheckKeyExists(keyValue string) bool // O(1) primary key existence check

	// Junction table duplicate prevention
	IsForeignKeyUnique(row *Row) bool // Check if row's FK combination is unique for junction tables
	RemoveRow(rowIndex int) error     // Remove row and update hash maps
}

// RelationshipInterface defines operations for relationships
type RelationshipInterface interface {
	GetID() string
	GetName() string
	GetSourceEntity() EntityInterface
	GetTargetEntity() EntityInterface
	GetSourceAttribute() AttributeInterface
	GetTargetAttribute() AttributeInterface
	GetCardinality() string
	IsOneToOne() bool
	IsOneToMany() bool
	IsManyToOne() bool

	// Target value selection for FK population
	GetTargetValueForSourceRow(sourceRowIndex int, autoCardinality bool) (string, error)
}

// AttributeInterface defines operations for attributes
type AttributeInterface interface {
	GetName() string
	GetExternalID() string
	GetAttributeAlias() string
	GetDataType() string
	IsUnique() bool
	IsRelationship() bool
	GetParentEntity() EntityInterface
	GetRelatedEntityID() string
	GetRelatedAttribute() string

	// Required for relationship handling
	setRelationship(relatedEntityID, relatedAttributeName string)

	//
	setParentEntity(entity EntityInterface)
}

// Ensure types implement their interfaces
var (
	_ GraphInterface        = (*Graph)(nil)
	_ EntityInterface       = (*Entity)(nil)
	_ RelationshipInterface = (*Relationship)(nil)
	_ AttributeInterface    = (*Attribute)(nil)
)
