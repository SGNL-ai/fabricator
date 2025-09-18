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
	ForEachRow(fn func(row *Row) error) error
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
