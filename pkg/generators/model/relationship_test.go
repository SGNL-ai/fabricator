package model

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test fixtures for Relationship tests
var (
	// Mock source entity with primary key
	sourceEntity = &Entity{
		id:          "source_entity",
		externalID:  "source_ext",
		name:        "Source Entity",
		description: "Source entity for testing",
		attributes:  make(map[string]AttributeInterface),
		attrList:    []AttributeInterface{},
	}

	// Mock target entity with primary key
	targetEntity = &Entity{
		id:          "target_entity",
		externalID:  "target_ext",
		name:        "Target Entity",
		description: "Target entity for testing",
		attributes:  make(map[string]AttributeInterface),
		attrList:    []AttributeInterface{},
	}

	// Mock source attribute (primary key)
	sourcePKAttr = &Attribute{
		name:         "id",
		externalID:   "source_id_ext",
		dataType:     "string",
		isUnique:     true,
		description:  "Source primary key",
		parentEntity: sourceEntity,
	}

	// Mock source attribute (foreign key)
	sourceFKAttr = &Attribute{
		name:           "target_id",
		externalID:     "target_id_ext",
		dataType:       "string",
		isUnique:       false,
		isRelationship: true,
		description:    "Source foreign key to target",
		parentEntity:   sourceEntity,
		relatedEntity:  "target_entity",
		relatedAttr:    "id",
	}

	// Mock target attribute (primary key)
	targetPKAttr = &Attribute{
		name:         "id",
		externalID:   "target_id_ext",
		dataType:     "string",
		isUnique:     true,
		description:  "Target primary key",
		parentEntity: targetEntity,
	}
)

// setupEntities configures test entity fixtures with appropriate attributes
func setupEntities() {
	// Reset entities' attributes
	sourceEntity.attributes = make(map[string]AttributeInterface)
	sourceEntity.attrList = make([]AttributeInterface, 0)
	sourceEntity.primaryKey = nil

	targetEntity.attributes = make(map[string]AttributeInterface)
	targetEntity.attrList = make([]AttributeInterface, 0)
	targetEntity.primaryKey = nil

	// Add primary keys to entities
	sourceEntity.attributes[sourcePKAttr.name] = sourcePKAttr
	sourceEntity.attrList = append(sourceEntity.attrList, sourcePKAttr)
	sourceEntity.primaryKey = sourcePKAttr

	targetEntity.attributes[targetPKAttr.name] = targetPKAttr
	targetEntity.attrList = append(targetEntity.attrList, targetPKAttr)
	targetEntity.primaryKey = targetPKAttr
}

// Tests for relationship constructor
func TestNewRelationship(t *testing.T) {
	// Set up test entities
	setupEntities()

	t.Run("should create relationship with valid entities and attributes", func(t *testing.T) {
		// Add foreign key to source entity
		sourceEntity.attributes[sourceFKAttr.name] = sourceFKAttr
		sourceEntity.attrList = append(sourceEntity.attrList, sourceFKAttr)

		// Create relationship
		rel, err := newRelationship(
			"test_rel",
			"Test Relationship",
			sourceEntity,
			targetEntity,
			sourceFKAttr.name,
			targetPKAttr.name,
		)

		// Verify relationship was created successfully
		require.NoError(t, err)
		require.NotNil(t, rel)

		// Verify basic properties were set correctly
		assert.Equal(t, "test_rel", rel.GetID())
		assert.Equal(t, "Test Relationship", rel.GetName())
		assert.Equal(t, sourceEntity, rel.GetSourceEntity())
		assert.Equal(t, targetEntity, rel.GetTargetEntity())
		assert.Equal(t, sourceFKAttr, rel.GetSourceAttribute())
		assert.Equal(t, targetPKAttr, rel.GetTargetAttribute())

		// Verify cardinality was determined correctly (many-to-one)
		assert.Equal(t, "N:1", rel.GetCardinality())
	})

	t.Run("should validate source entity is not nil", func(t *testing.T) {
		rel, err := newRelationship(
			"test_rel",
			"Test Relationship",
			nil, // nil source entity
			targetEntity,
			"id",
			"id",
		)

		// Verify relationship creation failed with appropriate error
		assert.Error(t, err)
		assert.Nil(t, rel)
		assert.True(t, strings.Contains(strings.ToLower(err.Error()), "source entity"))
	})

	t.Run("should validate target entity is not nil", func(t *testing.T) {
		rel, err := newRelationship(
			"test_rel",
			"Test Relationship",
			sourceEntity,
			nil, // nil target entity
			"id",
			"id",
		)

		// Verify relationship creation failed with appropriate error
		assert.Error(t, err)
		assert.Nil(t, rel)
		assert.True(t, strings.Contains(strings.ToLower(err.Error()), "target entity"))
	})

	t.Run("should validate source attribute exists", func(t *testing.T) {
		rel, err := newRelationship(
			"test_rel",
			"Test Relationship",
			sourceEntity,
			targetEntity,
			"nonexistent_attr", // nonexistent source attribute
			"id",
		)

		// Verify relationship creation failed with appropriate error
		assert.Error(t, err)
		assert.Nil(t, rel)
		assert.True(t, strings.Contains(strings.ToLower(err.Error()), "source attribute"))
	})

	t.Run("should validate target attribute exists", func(t *testing.T) {
		rel, err := newRelationship(
			"test_rel",
			"Test Relationship",
			sourceEntity,
			targetEntity,
			"id",
			"nonexistent_attr", // nonexistent target attribute
		)

		// Verify relationship creation failed with appropriate error
		assert.Error(t, err)
		assert.Nil(t, rel)
		assert.True(t, strings.Contains(strings.ToLower(err.Error()), "target attribute"))
	})

	t.Run("should validate relationship ID is not empty", func(t *testing.T) {
		rel, err := newRelationship(
			"", // empty ID
			"Test Relationship",
			sourceEntity,
			targetEntity,
			"id",
			"id",
		)

		// Verify relationship creation failed with appropriate error
		assert.Error(t, err)
		assert.Nil(t, rel)
		assert.True(t, strings.Contains(strings.ToLower(err.Error()), "id"))
	})

	t.Run("should validate relationship name is not empty", func(t *testing.T) {
		rel, err := newRelationship(
			"test_rel",
			"", // empty name
			sourceEntity,
			targetEntity,
			"id",
			"id",
		)

		// Verify relationship creation failed with appropriate error
		assert.Error(t, err)
		assert.Nil(t, rel)
		assert.True(t, strings.Contains(strings.ToLower(err.Error()), "name"))
	})

	t.Run("should validate at least one attribute is unique", func(t *testing.T) {
		// Create non-unique attributes for testing
		nonUniqueSourceAttr := &Attribute{
			name:         "regular_attr",
			externalID:   "regular_ext",
			dataType:     "string",
			isUnique:     false,
			description:  "Non-unique attribute",
			parentEntity: sourceEntity,
		}

		nonUniqueTargetAttr := &Attribute{
			name:         "regular_attr",
			externalID:   "regular_ext",
			dataType:     "string",
			isUnique:     false,
			description:  "Non-unique attribute",
			parentEntity: targetEntity,
		}

		// Add non-unique attributes to entities
		sourceEntity.attributes[nonUniqueSourceAttr.name] = nonUniqueSourceAttr
		sourceEntity.attrList = append(sourceEntity.attrList, nonUniqueSourceAttr)

		targetEntity.attributes[nonUniqueTargetAttr.name] = nonUniqueTargetAttr
		targetEntity.attrList = append(targetEntity.attrList, nonUniqueTargetAttr)

		// Try to create relationship with non-unique attributes on both sides
		rel, err := newRelationship(
			"test_rel",
			"Test Relationship",
			sourceEntity,
			targetEntity,
			nonUniqueSourceAttr.name,
			nonUniqueTargetAttr.name,
		)

		// Verify relationship creation failed with appropriate error
		assert.Error(t, err)
		assert.Nil(t, rel)
		assert.True(t, strings.Contains(strings.ToLower(err.Error()), "unique"))
	})
}

// Tests for cardinality detection
func TestRelationshipCardinality(t *testing.T) {
	// Set up test entities
	setupEntities()

	t.Run("should detect many-to-one relationship (non-unique FK to unique PK)", func(t *testing.T) {
		// Configure entities with appropriate attributes
		sourceEntity.attributes[sourceFKAttr.name] = sourceFKAttr
		sourceEntity.attrList = append(sourceEntity.attrList, sourceFKAttr)

		// Create relationship
		rel, err := newRelationship(
			"many_to_one",
			"Many to One",
			sourceEntity,
			targetEntity,
			sourceFKAttr.name,
			targetPKAttr.name,
		)

		// Verify relationship cardinality
		require.NoError(t, err)
		require.NotNil(t, rel)
		assert.Equal(t, "N:1", rel.GetCardinality())
		assert.False(t, rel.IsOneToOne())
		assert.False(t, rel.IsOneToMany())
		assert.True(t, rel.IsManyToOne())
	})

	t.Run("should detect one-to-many relationship (unique PK to non-unique FK)", func(t *testing.T) {
		// Create a one-to-many relationship
		rel, err := newRelationship(
			"one_to_many",
			"One to Many",
			targetEntity,
			sourceEntity,
			targetPKAttr.name,
			sourceFKAttr.name,
		)

		// Verify relationship cardinality
		require.NoError(t, err)
		require.NotNil(t, rel)
		assert.Equal(t, "1:N", rel.GetCardinality())
		assert.False(t, rel.IsOneToOne())
		assert.True(t, rel.IsOneToMany())
		assert.False(t, rel.IsManyToOne())
	})

	t.Run("should detect one-to-one relationship (unique to unique)", func(t *testing.T) {
		// Create a unique foreign key attribute
		uniqueFKAttr := &Attribute{
			name:           "unique_fk",
			externalID:     "unique_fk_ext",
			dataType:       "string",
			isUnique:       true,
			isRelationship: true,
			description:    "Unique foreign key",
			parentEntity:   sourceEntity,
			relatedEntity:  "target_entity",
			relatedAttr:    "id",
		}

		// Add unique FK to source entity
		sourceEntity.attributes[uniqueFKAttr.name] = uniqueFKAttr
		sourceEntity.attrList = append(sourceEntity.attrList, uniqueFKAttr)

		// Create relationship with unique attributes on both sides
		rel, err := newRelationship(
			"one_to_one",
			"One to One",
			sourceEntity,
			targetEntity,
			uniqueFKAttr.name,
			targetPKAttr.name,
		)

		// Verify relationship cardinality
		require.NoError(t, err)
		require.NotNil(t, rel)
		assert.Equal(t, "1:1", rel.GetCardinality())
		assert.True(t, rel.IsOneToOne())
		assert.False(t, rel.IsOneToMany())
		assert.False(t, rel.IsManyToOne())
	})
}
