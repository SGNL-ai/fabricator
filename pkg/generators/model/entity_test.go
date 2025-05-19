package model

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// We'll stop using global mocks in this test suite and create all objects within the test functions

// Helper no longer needed - we'll create test entities directly in each test

// Tests for entity constructor
func TestNewEntity(t *testing.T) {
	// Create a gomock controller for each test
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("should create a new entity with the specified properties", func(t *testing.T) {
		// Create entity
		id := "test_entity"
		externalID := "test_ext_id"
		name := "Test Entity"
		description := "Test entity description"

		// Create a mock graph
		mockGraph := NewMockGraphInterface(ctrl)

		// Create with empty attributes list
		entity, err := newEntity(id, externalID, name, description, []AttributeInterface{}, mockGraph)

		// Verify no error was returned
		require.NoError(t, err)
		require.NotNil(t, entity)

		// Verify properties were set correctly
		assert.Equal(t, id, entity.GetID())
		assert.Equal(t, externalID, entity.GetExternalID())
		assert.Equal(t, name, entity.GetName())
		assert.Equal(t, description, entity.GetDescription())

		// Verify internal state
		assert.NotNil(t, entity.GetAttributes())
		assert.Empty(t, entity.GetAttributes())
		assert.Equal(t, 0, entity.GetRowCount())
		assert.Nil(t, entity.GetPrimaryKey())
	})

	t.Run("should validate entity has valid ID and name", func(t *testing.T) {
		// Create a mock graph
		mockGraph := NewMockGraphInterface(ctrl)

		// Test with empty ID
		entity, err := newEntity("", "ext_id", "Entity Name", "Description", []AttributeInterface{}, mockGraph)
		assert.Error(t, err)
		assert.Nil(t, entity)

		// Test with empty name
		entity, err = newEntity("entity_id", "ext_id", "", "Description", []AttributeInterface{}, mockGraph)
		assert.Error(t, err)
		assert.Nil(t, entity)

		// Test with empty external ID
		entity, err = newEntity("entity_id", "", "Entity Name", "Description", []AttributeInterface{}, mockGraph)
		assert.Error(t, err)
		assert.Nil(t, entity)
	})

	t.Run("should add attributes during entity creation", func(t *testing.T) {
		// Set up test attributes
		// For now, we'll continue to use concrete Attribute type since we're in the same package
		// This would typically be replaced with a mock, but since we're in the same package
		// it's simpler to use the real implementation
		idAttr := &Attribute{
			name:        "id",
			externalID:  "id_ext",
			dataType:    "string",
			isUnique:    true,
			description: "Primary key",
		}

		nameAttr := &Attribute{
			name:        "name",
			externalID:  "name_ext",
			dataType:    "string",
			isUnique:    false,
			description: "Name field",
		}

		// Create a mock graph
		mockGraph := NewMockGraphInterface(ctrl)

		// Create entity with attributes
		entity, err := newEntity("test_entity", "test_ext_id", "Test Entity", "Description", []AttributeInterface{idAttr, nameAttr}, mockGraph)
		require.NoError(t, err)

		// Verify attributes were added
		attrs := entity.GetAttributes()
		require.Len(t, attrs, 2)
		assert.Equal(t, "id", attrs[0].GetName())
		assert.Equal(t, "name", attrs[1].GetName())

		// Verify primary key was set
		assert.NotNil(t, entity.GetPrimaryKey())
		assert.Equal(t, "id", entity.GetPrimaryKey().GetName())
	})

	t.Run("should validate there is exactly one unique attribute", func(t *testing.T) {
		// Create attributes with multiple unique attributes
		idAttr := &Attribute{
			name:        "id",
			externalID:  "id_ext",
			dataType:    "string",
			isUnique:    true,
			description: "First unique attribute",
		}

		secondIDAttr := &Attribute{
			name:        "second_id",
			externalID:  "second_id_ext",
			dataType:    "string",
			isUnique:    true,
			description: "Second unique attribute",
		}

		// Create a mock graph
		mockGraph := NewMockGraphInterface(ctrl)

		// Create entity with multiple unique attributes - should fail
		entity, err := newEntity("test_entity", "test_ext_id", "Test Entity", "Description",
			[]AttributeInterface{idAttr, secondIDAttr}, mockGraph)
		assert.Error(t, err)
		assert.Nil(t, entity)
	})

	t.Run("should validate attribute names are unique", func(t *testing.T) {
		// Create attributes with duplicate names
		firstAttr := &Attribute{
			name:        "duplicate",
			externalID:  "first_ext",
			dataType:    "string",
			isUnique:    true,
			description: "First attribute",
		}

		duplicateAttr := &Attribute{
			name:        "duplicate",
			externalID:  "second_ext",
			dataType:    "string",
			isUnique:    false,
			description: "Duplicate attribute",
		}

		// Create a mock graph
		mockGraph := NewMockGraphInterface(ctrl)

		// Create entity with duplicate attribute names - should fail
		entity, err := newEntity("test_entity", "test_ext_id", "Test Entity", "Description",
			[]AttributeInterface{firstAttr, duplicateAttr}, mockGraph)
		assert.Error(t, err)
		assert.Nil(t, entity)
	})
}

// Tests for attribute retrieval methods
func TestGetAttributes(t *testing.T) {
	// Create a gomock controller for each test
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("should return all attributes in order", func(t *testing.T) {
		// Create attributes in specific order
		idAttr := &Attribute{
			name:        "id",
			externalID:  "id_ext",
			dataType:    "string",
			isUnique:    true,
			description: "Primary key",
		}

		nameAttr := &Attribute{
			name:        "name",
			externalID:  "name_ext",
			dataType:    "string",
			isUnique:    false,
			description: "Name attribute",
		}

		parentAttr := &Attribute{
			name:        "parent_id",
			externalID:  "parent_ext",
			dataType:    "string",
			isUnique:    false,
			description: "Foreign key",
		}

		// Create a mock graph
		mockGraph := NewMockGraphInterface(ctrl)

		// Create entity with attributes
		entity, err := newEntity("test_entity", "test_ext_id", "Test Entity", "Description",
			[]AttributeInterface{idAttr, nameAttr, parentAttr}, mockGraph)
		require.NoError(t, err)

		// Verify attributes are returned in order they were added
		attrs := entity.GetAttributes()
		require.Len(t, attrs, 3)
		assert.Equal(t, "id", attrs[0].GetName())
		assert.Equal(t, "name", attrs[1].GetName())
		assert.Equal(t, "parent_id", attrs[2].GetName())
	})

	t.Run("should get attribute by name", func(t *testing.T) {
		// Create attribute
		attrName := "test_attr"
		testAttr := &Attribute{
			name:        attrName,
			externalID:  "ext_id",
			dataType:    "string",
			isUnique:    true,
			description: "Test attribute",
		}

		// Create a mock graph
		mockGraph := NewMockGraphInterface(ctrl)

		// Create entity with attribute
		entity, err := newEntity("test_entity", "test_ext_id", "Test Entity", "Description",
			[]AttributeInterface{testAttr}, mockGraph)
		require.NoError(t, err)

		// Get attribute by name
		attr, exists := entity.GetAttribute(attrName)
		assert.True(t, exists)
		assert.NotNil(t, attr)
		assert.Equal(t, attrName, attr.GetName())

		// Try to get non-existent attribute
		attr, exists = entity.GetAttribute("non_existent")
		assert.False(t, exists)
		assert.Nil(t, attr)
	})

	t.Run("should get primary key attribute", func(t *testing.T) {
		// Create a mock graph for the empty entity
		mockGraphEmpty := NewMockGraphInterface(ctrl)

		// Create entity with no attributes first
		emptyEntity, err := newEntity("test_entity", "test_ext_id", "Test Entity", "Description", []AttributeInterface{}, mockGraphEmpty)
		require.NoError(t, err)

		// No primary key initially
		assert.Nil(t, emptyEntity.GetPrimaryKey())

		// Create unique attribute to serve as primary key
		pkName := "id"
		pkAttr := &Attribute{
			name:        pkName,
			externalID:  "id_ext",
			dataType:    "string",
			isUnique:    true,
			description: "Primary key",
		}

		// Create a new mock graph for the entity with primary key
		mockGraph := NewMockGraphInterface(ctrl)

		// Create entity with primary key
		entity, err := newEntity("test_entity", "test_ext_id", "Test Entity", "Description",
			[]AttributeInterface{pkAttr}, mockGraph)
		require.NoError(t, err)

		// Verify primary key is set
		pk := entity.GetPrimaryKey()
		assert.NotNil(t, pk)
		assert.Equal(t, pkName, pk.GetName())
		assert.True(t, pk.IsUnique())
	})

	t.Run("should filter attributes by properties", func(t *testing.T) {
		// Create different types of attributes
		idAttr := &Attribute{
			name:        "id",
			externalID:  "id_ext",
			dataType:    "string",
			isUnique:    true,
			description: "Primary key",
		}

		nameAttr := &Attribute{
			name:        "name",
			externalID:  "name_ext",
			dataType:    "string",
			isUnique:    false,
			description: "Regular attribute",
		}

		// Create relationship attribute
		parentAttr := &Attribute{
			name:           "parent_id",
			externalID:     "parent_ext",
			dataType:       "string",
			isUnique:       false,
			isRelationship: true,
			relatedEntity:  "parent_entity",
			relatedAttr:    "id",
			description:    "Relationship attribute",
		}

		// Create a mock graph
		mockGraph := NewMockGraphInterface(ctrl)

		// Create entity with various attributes
		entity, err := newEntity("test_entity", "test_ext_id", "Test Entity", "Description",
			[]AttributeInterface{idAttr, nameAttr, parentAttr}, mockGraph)
		require.NoError(t, err)

		// Test GetNonUniqueAttributes
		nonUniqueAttrs := entity.GetNonUniqueAttributes()
		assert.Len(t, nonUniqueAttrs, 2)
		assert.Equal(t, "name", nonUniqueAttrs[0].GetName())
		assert.Equal(t, "parent_id", nonUniqueAttrs[1].GetName())

		// Test GetRelationshipAttributes
		relAttrs := entity.GetRelationshipAttributes()
		assert.Len(t, relAttrs, 1)
		assert.Equal(t, "parent_id", relAttrs[0].GetName())

		// Test GetNonRelationshipAttributes
		nonRelAttrs := entity.GetNonRelationshipAttributes()
		assert.Len(t, nonRelAttrs, 2)
		assert.Equal(t, "id", nonRelAttrs[0].GetName())
		assert.Equal(t, "name", nonRelAttrs[1].GetName())
	})
}

// Tests for row management
func TestEntityRows(t *testing.T) {
	// Create a gomock controller for each test
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("should add row with valid values", func(t *testing.T) {
		// Create attributes
		idAttr := &Attribute{
			name:        "id",
			externalID:  "id_ext",
			dataType:    "string",
			isUnique:    true,
			description: "Primary key",
		}

		nameAttr := &Attribute{
			name:        "name",
			externalID:  "name_ext",
			dataType:    "string",
			isUnique:    false,
			description: "Name attribute",
		}

		// Create a mock graph
		mockGraph := NewMockGraphInterface(ctrl)

		// Create entity with attributes
		entity, err := newEntity("test_entity", "test_ext_id", "Test Entity", "Description",
			[]AttributeInterface{idAttr, nameAttr}, mockGraph)
		require.NoError(t, err)

		// Add row with valid values
		values := &Row{values: map[string]string{
			"id":   "1",
			"name": "Test Row",
		}}
		err = entity.AddRow(values)
		require.NoError(t, err)

		// Verify row count
		assert.Equal(t, 1, entity.GetRowCount())
	})

	t.Run("should enforce primary key uniqueness", func(t *testing.T) {
		// Create attribute for primary key
		idAttr := &Attribute{
			name:        "id",
			externalID:  "id_ext",
			dataType:    "string",
			isUnique:    true,
			description: "Primary key",
		}

		// Create a mock graph
		mockGraph := NewMockGraphInterface(ctrl)

		// Create entity with primary key
		entity, err := newEntity("test_entity", "test_ext_id", "Test Entity", "Description",
			[]AttributeInterface{idAttr}, mockGraph)
		require.NoError(t, err)

		// Add first row
		err = entity.AddRow(&Row{values: map[string]string{"id": "1"}})
		require.NoError(t, err)

		// Add row with duplicate primary key - should fail
		err = entity.AddRow(&Row{values: map[string]string{"id": "1"}})
		assert.Error(t, err)

		// Verify row count remains 1
		assert.Equal(t, 1, entity.GetRowCount())
	})

	t.Run("should validate required attributes", func(t *testing.T) {
		// Create attributes
		idAttr := &Attribute{
			name:        "id",
			externalID:  "id_ext",
			dataType:    "string",
			isUnique:    true,
			description: "Primary key",
		}

		nameAttr := &Attribute{
			name:        "name",
			externalID:  "name_ext",
			dataType:    "string",
			isUnique:    false,
			description: "Name attribute",
		}

		// Create a mock graph
		mockGraph := NewMockGraphInterface(ctrl)

		// Create entity with attributes
		entity, err := newEntity("test_entity", "test_ext_id", "Test Entity", "Description",
			[]AttributeInterface{idAttr, nameAttr}, mockGraph)
		require.NoError(t, err)

		// Add row missing primary key - should fail
		err = entity.AddRow(&Row{values: map[string]string{"name": "Test Row"}})
		assert.Error(t, err)

		// Verify no rows were added
		assert.Equal(t, 0, entity.GetRowCount())
	})

	t.Run("should validate foreign key references", func(t *testing.T) {
		// Create a mock graph
		mockGraph := NewMockGraphInterface(ctrl)

		// Create parent entity attributes
		parentIDAttr := &Attribute{
			name:        "id",
			externalID:  "id_ext",
			dataType:    "string",
			isUnique:    true,
			description: "Parent PK",
		}

		// Create parent entity
		parentEntity, err := newEntity("parent", "parent_ext", "Parent", "Parent entity",
			[]AttributeInterface{parentIDAttr}, mockGraph)
		require.NoError(t, err)

		// Add row to parent
		err = parentEntity.AddRow(&Row{values: map[string]string{"id": "parent1"}})
		require.NoError(t, err)

		// Set up mockGraph to return the parent entity when GetEntity("parent") is called
		mockGraph.EXPECT().GetEntity("parent").Return(parentEntity, true).AnyTimes()

		// Create child entity attributes
		childIDAttr := &Attribute{
			name:        "id",
			externalID:  "id_ext",
			dataType:    "string",
			isUnique:    true,
			description: "Child PK",
		}

		// Create relationship attribute
		parentFKAttr := &Attribute{
			name:           "parent_id",
			externalID:     "parent_id_ext",
			dataType:       "string",
			isUnique:       false,
			isRelationship: true,
			relatedEntity:  "parent",
			relatedAttr:    "id",
			description:    "Parent FK",
		}

		// Create child entity
		childEntity, err := newEntity("child", "child_ext", "Child", "Child entity",
			[]AttributeInterface{childIDAttr, parentFKAttr}, mockGraph)
		require.NoError(t, err)

		// Add row with valid foreign key reference - should succeed
		err = childEntity.AddRow(&Row{
			values: map[string]string{
				"id":        "child1",
				"parent_id": "parent1",
			},
		})
		assert.NoError(t, err)

		// Add row with invalid foreign key reference - should fail
		err = childEntity.AddRow(&Row{
			values: map[string]string{
				"id":        "child2",
				"parent_id": "nonexistent",
			},
		})
		assert.Error(t, err)
	})
}

// Tests for CSV generation
func TestEntityToCSV(t *testing.T) {
	// Create a gomock controller for each test
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("should convert entity to CSV data", func(t *testing.T) {
		// Create attributes
		idAttr := &Attribute{
			name:        "id",
			externalID:  "id_ext",
			dataType:    "string",
			isUnique:    true,
			description: "Primary key",
		}

		nameAttr := &Attribute{
			name:        "name",
			externalID:  "name_ext",
			dataType:    "string",
			isUnique:    false,
			description: "Name attribute",
		}

		// Create a mock graph
		mockGraph := NewMockGraphInterface(ctrl)

		// Create entity with attributes
		entity, err := newEntity("test_entity", "test_ext_id", "Test Entity", "Description",
			[]AttributeInterface{idAttr, nameAttr}, mockGraph)
		require.NoError(t, err)

		// Add rows
		err = entity.AddRow(&Row{values: map[string]string{"id": "1", "name": "Row 1"}})
		require.NoError(t, err)

		err = entity.AddRow(&Row{values: map[string]string{"id": "2", "name": "Row 2"}})
		require.NoError(t, err)

		// Convert to CSV
		csvData := entity.ToCSV()
		require.NotNil(t, csvData)

		// Verify CSV structure
		assert.Equal(t, entity.GetExternalID(), csvData.ExternalId)
		assert.Equal(t, 2, len(csvData.Rows))
		assert.Equal(t, []string{"id", "name"}, csvData.Headers)

		// Verify rows contain expected data
		assert.Len(t, csvData.Rows, 2)

		// The expected format is a 2D array where rows[0][0] is first row, first cell
		// Since we don't know the exact order, we'll verify there's a row with each ID
		var foundRow1, foundRow2 bool

		for _, row := range csvData.Rows {
			assert.Len(t, row, 2) // Each row should have 2 values

			if row[0] == "1" && row[1] == "Row 1" {
				foundRow1 = true
			}

			if row[0] == "2" && row[1] == "Row 2" {
				foundRow2 = true
			}
		}

		assert.True(t, foundRow1, "Should find row with id=1, name=Row 1")
		assert.True(t, foundRow2, "Should find row with id=2, name=Row 2")
	})
}

// Tests for relationship creation
func TestAddRelationship(t *testing.T) {
	// Create a gomock controller for each test
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("should create relationship between two entities", func(t *testing.T) {
		// Create a mock graph
		mockGraph := NewMockGraphInterface(ctrl)

		// Create source entity with unique attribute
		sourceID := "source_entity"
		sourceExternalID := "source_ext"
		sourceIdAttr := &Attribute{
			name:        "id",
			externalID:  "source_id",
			dataType:    "string",
			isUnique:    true,
			description: "Source primary key",
		}

		sourceEntity, err := newEntity(sourceID, sourceExternalID, "Source Entity", "Source entity description",
			[]AttributeInterface{sourceIdAttr}, mockGraph)
		require.NoError(t, err)

		// Create target entity with unique attribute
		targetID := "target_entity"
		targetExternalID := "target_ext"
		targetIdAttr := &Attribute{
			name:        "id",
			externalID:  "target_id",
			dataType:    "string",
			isUnique:    true,
			description: "Target primary key",
		}

		targetEntity, err := newEntity(targetID, targetExternalID, "Target Entity", "Target entity description",
			[]AttributeInterface{targetIdAttr}, mockGraph)
		require.NoError(t, err)

		// Create relationship between entities
		relationshipID := "test_relationship"
		relationshipName := "TEST_REL"
		relationship, err := sourceEntity.addRelationship(
			relationshipID,
			relationshipName,
			targetEntity,
			"source_id", // source attribute external ID
			"target_id", // target attribute external ID
		)

		// Verify relationship creation
		require.NoError(t, err)
		require.NotNil(t, relationship)

		// Verify relationship properties
		assert.Equal(t, relationshipID, relationship.GetID())
		assert.Equal(t, relationshipName, relationship.GetName())
		assert.Equal(t, sourceEntity, relationship.GetSourceEntity())
		assert.Equal(t, targetEntity, relationship.GetTargetEntity())
		assert.Equal(t, sourceIdAttr, relationship.GetSourceAttribute())
		assert.Equal(t, targetIdAttr, relationship.GetTargetAttribute())

		// Verify cardinality was determined (should be 1:1 since both attributes are unique)
		assert.Equal(t, OneToOne, relationship.GetCardinality())
		assert.True(t, relationship.IsOneToOne())

		// Verify attributes were marked as relationship attributes
		assert.True(t, sourceIdAttr.IsRelationship())
		assert.True(t, targetIdAttr.IsRelationship())

		// Verify related entity and attribute references were set
		assert.Equal(t, targetID, sourceIdAttr.GetRelatedEntityID())
		assert.Equal(t, targetIdAttr.GetName(), sourceIdAttr.GetRelatedAttribute())
		assert.Equal(t, sourceID, targetIdAttr.GetRelatedEntityID())
		assert.Equal(t, sourceIdAttr.GetName(), targetIdAttr.GetRelatedAttribute())
	})

	t.Run("should handle one-to-many relationship", func(t *testing.T) {
		// Create a mock graph
		mockGraph := NewMockGraphInterface(ctrl)

		// Create source entity with unique attribute (one side)
		sourceIDAttr := &Attribute{
			name:        "id",
			externalID:  "parent_id",
			dataType:    "string",
			isUnique:    true,
			description: "Parent primary key",
		}

		sourceEntity, err := newEntity("parent", "parent_ext", "Parent", "Parent entity",
			[]AttributeInterface{sourceIDAttr}, mockGraph)
		require.NoError(t, err)

		// Create target entity with non-unique attribute (many side)
		targetIDAttr := &Attribute{
			name:        "id",
			externalID:  "child_id",
			dataType:    "string",
			isUnique:    true,
			description: "Child primary key",
		}

		targetFKAttr := &Attribute{
			name:        "parent_id",
			externalID:  "parent_ref",
			dataType:    "string",
			isUnique:    false, // Non-unique for many side
			description: "Reference to parent",
		}

		targetEntity, err := newEntity("child", "child_ext", "Child", "Child entity",
			[]AttributeInterface{targetIDAttr, targetFKAttr}, mockGraph)
		require.NoError(t, err)

		// Create relationship from parent to child
		relationship, err := sourceEntity.addRelationship(
			"parent_to_child",
			"PARENT_CHILD",
			targetEntity,
			"parent_id",  // source attribute external ID
			"parent_ref", // target attribute external ID
		)

		// Verify relationship creation
		require.NoError(t, err)
		require.NotNil(t, relationship)

		// Verify relationship properties
		assert.Equal(t, sourceEntity, relationship.GetSourceEntity())
		assert.Equal(t, targetEntity, relationship.GetTargetEntity())
		assert.Equal(t, sourceIDAttr, relationship.GetSourceAttribute())
		assert.Equal(t, targetFKAttr, relationship.GetTargetAttribute())

		// Verify cardinality was determined (should be 1:N)
		assert.Equal(t, OneToMany, relationship.GetCardinality())
		assert.True(t, relationship.IsOneToMany())
	})

	t.Run("should fail when source attribute not found", func(t *testing.T) {
		// Create a mock graph
		mockGraph := NewMockGraphInterface(ctrl)

		// Create source entity
		sourceIDAttr := &Attribute{
			name:        "id",
			externalID:  "source_id",
			dataType:    "string",
			isUnique:    true,
			description: "Source primary key",
		}

		sourceEntity, err := newEntity("source", "source_ext", "Source", "Source entity",
			[]AttributeInterface{sourceIDAttr}, mockGraph)
		require.NoError(t, err)

		// Create target entity
		targetIDAttr := &Attribute{
			name:        "id",
			externalID:  "target_id",
			dataType:    "string",
			isUnique:    true,
			description: "Target primary key",
		}

		targetEntity, err := newEntity("target", "target_ext", "Target", "Target entity",
			[]AttributeInterface{targetIDAttr}, mockGraph)
		require.NoError(t, err)

		// Try to create relationship with non-existent source attribute
		relationship, err := sourceEntity.addRelationship(
			"test_relationship",
			"TEST_REL",
			targetEntity,
			"nonexistent_id", // non-existent source attribute
			"target_id",
		)

		// Verify relationship creation failed
		assert.Error(t, err)
		assert.Nil(t, relationship)
		assert.Contains(t, err.Error(), "source attribute")
	})

	t.Run("should fail when target attribute not found", func(t *testing.T) {
		// Create a mock graph
		mockGraph := NewMockGraphInterface(ctrl)

		// Create source entity
		sourceIDAttr := &Attribute{
			name:        "id",
			externalID:  "source_id",
			dataType:    "string",
			isUnique:    true,
			description: "Source primary key",
		}

		sourceEntity, err := newEntity("source", "source_ext", "Source", "Source entity",
			[]AttributeInterface{sourceIDAttr}, mockGraph)
		require.NoError(t, err)

		// Create target entity
		targetIDAttr := &Attribute{
			name:        "id",
			externalID:  "target_id",
			dataType:    "string",
			isUnique:    true,
			description: "Target primary key",
		}

		targetEntity, err := newEntity("target", "target_ext", "Target", "Target entity",
			[]AttributeInterface{targetIDAttr}, mockGraph)
		require.NoError(t, err)

		// Try to create relationship with non-existent target attribute
		relationship, err := sourceEntity.addRelationship(
			"test_relationship",
			"TEST_REL",
			targetEntity,
			"source_id",
			"nonexistent_id", // non-existent target attribute
		)

		// Verify relationship creation failed
		assert.Error(t, err)
		assert.Nil(t, relationship)
		assert.Contains(t, err.Error(), "target attribute")
	})
}
