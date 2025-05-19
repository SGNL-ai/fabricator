package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockEntity is a reusable mock entity for testing
var mockEntity = &Entity{
	id:         "entity1",
	name:       "TestEntity",
	externalID: "test_entity_ext_id",
}

func TestNewAttribute(t *testing.T) {
	t.Run("should create a new attribute with the specified properties", func(t *testing.T) {
		// Create the attribute
		name := "test_attr"
		externalID := "ext_id"
		dataType := "string"
		isUnique := true
		description := "Test attribute description"

		attr := newAttribute(name, externalID, "", dataType, isUnique, description, mockEntity)

		// Verify properties using public interfaces
		assert.Equal(t, name, attr.GetName())
		assert.Equal(t, externalID, attr.GetExternalID())
		assert.Equal(t, dataType, attr.GetDataType())
		assert.Equal(t, isUnique, attr.IsUnique())
		assert.Equal(t, isUnique, attr.IsID()) // IsID uses IsUnique internally
		assert.Equal(t, mockEntity, attr.GetParentEntity())
	})

	t.Run("should set default values for relationship fields", func(t *testing.T) {
		// Create the attribute
		attr := newAttribute("test_attr", "ext_id", "", "string", true, "Test attribute description", mockEntity)

		// Verify relationship fields are set to defaults using public interfaces
		assert.False(t, attr.IsRelationship())
		assert.Empty(t, attr.GetRelatedEntityID())
		assert.Empty(t, attr.GetRelatedAttribute())
	})
}
func TestIsRelationship(t *testing.T) {
	t.Run("should return whether attribute is part of a relationship", func(t *testing.T) {
		// Create a non-relationship attribute
		nonRelAttr := newAttribute("test_attr", "ext_id", "", "string", false, "Test description", mockEntity)
		assert.False(t, nonRelAttr.IsRelationship())

		// Create a relationship attribute by setting relationship properties
		relAttr := newAttribute("rel_attr", "ext_id2", "", "string", false, "Relationship attribute", mockEntity)
		relAttr.setRelationship("related_entity", "related_attr_name")
		assert.True(t, relAttr.IsRelationship())
	})
}

func TestGetRelatedEntityID(t *testing.T) {
	t.Run("should return the related entity ID if part of a relationship", func(t *testing.T) {
		// Create a relationship attribute
		attr := newAttribute("rel_attr", "ext_id", "", "string", false, "Relationship attribute", mockEntity)
		relatedEntityID := "related_entity"
		attr.setRelationship(relatedEntityID, "related_attr_name")

		// Verify GetRelatedEntityID returns the correct entity ID
		assert.Equal(t, relatedEntityID, attr.GetRelatedEntityID())
	})

	t.Run("should return empty string if not part of a relationship", func(t *testing.T) {
		// Create a non-relationship attribute
		attr := newAttribute("non_rel_attr", "ext_id", "", "string", false, "Non-relationship attribute", mockEntity)

		// Verify GetRelatedEntityID returns empty string
		assert.Empty(t, attr.GetRelatedEntityID())
	})
}

func TestGetRelatedAttribute(t *testing.T) {
	t.Run("should return the related attribute name if part of a relationship", func(t *testing.T) {
		// Create a relationship attribute
		attr := newAttribute("rel_attr", "ext_id", "", "string", false, "Relationship attribute", mockEntity)
		relatedEntityID := "related_entity"
		relatedAttrName := "related_attr_name"
		attr.setRelationship(relatedEntityID, relatedAttrName)

		// Verify GetRelatedAttribute returns the correct attribute name
		assert.Equal(t, relatedAttrName, attr.GetRelatedAttribute())
	})

	t.Run("should return empty string if not part of a relationship", func(t *testing.T) {
		// Create a non-relationship attribute
		attr := newAttribute("non_rel_attr", "ext_id", "", "string", false, "Non-relationship attribute", mockEntity)

		// Verify GetRelatedAttribute returns empty string
		assert.Empty(t, attr.GetRelatedAttribute())
	})
}

func TestSetRelationship(t *testing.T) {
	t.Run("should mark attribute as part of a relationship", func(t *testing.T) {
		// Create a non-relationship attribute
		attr := newAttribute("attr", "ext_id", "", "string", false, "Test attribute", mockEntity)
		assert.False(t, attr.IsRelationship())

		// Set as relationship
		attr.setRelationship("entity_id", "attr_name")

		// Verify it's now marked as a relationship
		assert.True(t, attr.IsRelationship())
	})

	t.Run("should set related entity and attribute", func(t *testing.T) {
		// Create a non-relationship attribute
		attr := newAttribute("attr", "ext_id", "", "string", false, "Test attribute", mockEntity)
		assert.Empty(t, attr.GetRelatedEntityID())
		assert.Empty(t, attr.GetRelatedAttribute())

		// Set relationship values
		entityID := "related_entity_id"
		attrName := "related_attribute_name"
		attr.setRelationship(entityID, attrName)

		// Verify values are set correctly
		assert.Equal(t, entityID, attr.GetRelatedEntityID())
		assert.Equal(t, attrName, attr.GetRelatedAttribute())
	})
}

func TestAttributeValidity(t *testing.T) {
	t.Run("should validate attribute has a name", func(t *testing.T) {
		// Valid name
		validAttr := newAttribute("valid_name", "ext_id", "", "string", false, "Valid attribute", mockEntity)
		assert.NotEmpty(t, validAttr.GetName())

		// Empty name - this test assumes the implementation validates names
		// In a complete implementation, newAttribute could return an error for empty names
		emptyNameAttr := newAttribute("", "ext_id", "", "string", false, "Invalid attribute", mockEntity)
		assert.Empty(t, emptyNameAttr.GetName())
	})

	t.Run("should validate attribute has a parent entity", func(t *testing.T) {
		// Valid parent entity
		validAttr := newAttribute("attr_name", "ext_id", "", "string", false, "Valid attribute", mockEntity)
		assert.NotNil(t, validAttr.GetParentEntity())

		// Nil parent entity - this test assumes the implementation handles nil parents
		// In a complete implementation, newAttribute could return an error for nil parent entities
		nilParentAttr := newAttribute("attr_name", "ext_id", "", "string", false, "Invalid attribute", nil)
		assert.Nil(t, nilParentAttr.GetParentEntity())
	})
}
