package model

// Attribute represents an entity attribute and its properties
type Attribute struct {
	name           string
	externalID     string
	dataType       string
	isUnique       bool
	isRelationship bool
	description    string
	parentEntity   EntityInterface
	relatedEntity  string
	relatedAttr    string
}

// newAttribute creates a new attribute with the specified properties
// Not exported as only Entity should create attributes
func newAttribute(name, externalID string, dataType string, isUnique bool, description string, parentEntity EntityInterface) AttributeInterface {
	return &Attribute{
		name:         name,
		externalID:   externalID,
		dataType:     dataType,
		isUnique:     isUnique,
		description:  description,
		parentEntity: parentEntity,
	}
}

// GetName returns the attribute name
func (a *Attribute) GetName() string {
	return a.name
}

// GetExternalID returns the attribute's external ID
func (a *Attribute) GetExternalID() string {
	return a.externalID
}

// GetDataType returns the attribute's data type
func (a *Attribute) GetDataType() string {
	return a.dataType
}

// IsUnique returns whether attribute requires unique values
func (a *Attribute) IsUnique() bool {
	return a.isUnique
}

// IsID returns whether attribute is an identifier
func (a *Attribute) IsID() bool {
	return a.isUnique
}

// IsRelationship returns whether attribute is part of a relationship
func (a *Attribute) IsRelationship() bool {
	return a.isRelationship
}

// GetParentEntity returns the parent entity this attribute belongs to
func (a *Attribute) GetParentEntity() EntityInterface {
	return a.parentEntity
}

// GetRelatedEntityID returns the related entity ID if part of a relationship
func (a *Attribute) GetRelatedEntityID() string {
	return a.relatedEntity
}

// GetRelatedAttribute returns the related attribute name if part of a relationship
func (a *Attribute) GetRelatedAttribute() string {
	return a.relatedAttr
}

// setRelationship marks attribute as part of a relationship and sets related entity/attribute
func (a *Attribute) setRelationship(relatedEntityID, relatedAttributeName string) {
	a.isRelationship = true
	a.relatedEntity = relatedEntityID
	a.relatedAttr = relatedAttributeName
}

func (a *Attribute) setParentEntity(entity EntityInterface) {
	a.parentEntity = entity
}
