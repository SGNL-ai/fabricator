package model

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test fixtures for Graph tests
// A minimal SOR definition with 3 entities and 2 relationships for testing
var testSORDefinition = &models.SORDefinition{
	DisplayName:   "Test SOR",
	Description:   "Test System of Record for unit tests",
	Hostname:      "test.example.com",
	Type:          "Test-1.0.0",
	AdapterConfig: "",
	// Entities with unique IDs as keys
	Entities: map[string]models.Entity{
		"user_entity": {
			DisplayName: "User",
			ExternalId:  "test/User",
			Description: "User entity for testing",
			Attributes: []models.Attribute{
				{
					Name:        "id",
					ExternalId:  "user_id",
					Description: "Primary key",
					Type:        "string",
					UniqueId:    true,
				},
				{
					Name:        "name",
					ExternalId:  "user_name",
					Description: "User's full name",
					Type:        "string",
				},
				{
					Name:        "email",
					ExternalId:  "user_email",
					Description: "User's email address",
					Type:        "string",
				},
			},
		},
		"role_entity": {
			DisplayName: "Role",
			ExternalId:  "test/Role",
			Description: "Role entity for testing",
			Attributes: []models.Attribute{
				{
					Name:        "id",
					ExternalId:  "role_id",
					Description: "Primary key",
					Type:        "string",
					UniqueId:    true,
				},
				{
					Name:        "name",
					ExternalId:  "role_name",
					Description: "Role name",
					Type:        "string",
				},
			},
		},
		"user_role_entity": {
			DisplayName: "UserRole",
			ExternalId:  "test/UserRole",
			Description: "User-Role mapping entity for testing",
			Attributes: []models.Attribute{
				{
					Name:        "id",
					ExternalId:  "mapping_id",
					Description: "Primary key",
					Type:        "string",
					UniqueId:    true,
				},
				{
					Name:        "user_id",
					ExternalId:  "user_ref",
					Description: "Reference to user",
					Type:        "string",
				},
				{
					Name:        "role_id",
					ExternalId:  "role_ref",
					Description: "Reference to role",
					Type:        "string",
				},
			},
		},
	},
	// Relationships with unique IDs as keys
	Relationships: map[string]models.Relationship{
		"user_to_userrole": {
			DisplayName:   "USER_TO_USERROLE",
			Name:          "user_to_userrole",
			FromAttribute: "user_id",
			ToAttribute:   "user_id",
		},
		"role_to_userrole": {
			DisplayName:   "ROLE_TO_USERROLE",
			Name:          "role_to_userrole",
			FromAttribute: "role_id",
			ToAttribute:   "role_id",
		},
	},
}

// TestNewGraph tests the graph constructor
func TestNewGraph(t *testing.T) {
	t.Run("should create a graph from valid YAML definition", func(t *testing.T) {
		// Create a graph from the test SOR definition
		graph, err := NewGraph(testSORDefinition)
		
		// Verify the graph was created successfully
		require.NoError(t, err)
		require.NotNil(t, graph)
		
		// Verify entities were created
		assert.Equal(t, 3, len(graph.GetAllEntities()))
		
		// Verify relationships were created
		assert.Equal(t, 2, len(graph.GetAllRelationships()))
	})

	t.Run("should validate YAML definition is not nil", func(t *testing.T) {
		// Attempt to create a graph with nil YAML definition
		graph, err := NewGraph(nil)
		
		// Verify the creation fails with appropriate error
		assert.Error(t, err)
		assert.Equal(t, ErrNilYAMLModel, err)
		assert.Nil(t, graph)
	})

	t.Run("should validate entities exist in YAML", func(t *testing.T) {
		// Create a YAML definition with no entities
		emptySOR := &models.SORDefinition{
			DisplayName:   "Empty SOR",
			Description:   "SOR with no entities",
			Entities:      make(map[string]models.Entity),
			Relationships: make(map[string]models.Relationship),
		}
		
		// Attempt to create a graph with empty entities
		graph, err := NewGraph(emptySOR)
		
		// Verify the creation fails with appropriate error
		assert.Error(t, err)
		assert.Equal(t, ErrNoEntities, err)
		assert.Nil(t, graph)
	})
}

// TestEntityCreation tests the entity creation from YAML
func TestEntityCreation(t *testing.T) {
	t.Run("should create all entities with correct attributes", func(t *testing.T) {
		// Create a graph from the test SOR definition
		graph, err := NewGraph(testSORDefinition)
		require.NoError(t, err)
		
		// Get specific entity
		userEntity, exists := graph.GetEntity("user_entity")
		require.True(t, exists)
		require.NotNil(t, userEntity)
		
		// Verify entity properties
		assert.Equal(t, "User", userEntity.GetName())
		assert.Equal(t, "test/User", userEntity.GetExternalID())
		
		// Verify attributes were created
		attrs := userEntity.GetAttributes()
		assert.Equal(t, 3, len(attrs))
		
		// Verify primary key was identified
		primaryKey := userEntity.GetPrimaryKey()
		require.NotNil(t, primaryKey)
		assert.Equal(t, "id", primaryKey.GetName())
	})

	t.Run("should validate entity has exactly one unique attribute", func(t *testing.T) {
		// Create a YAML definition with an entity that has multiple unique attributes
		invalidSOR := *testSORDefinition // Clone the test SOR definition
		
		// Make a deep copy of the entities map
		invalidEntities := make(map[string]models.Entity)
		for id, entity := range testSORDefinition.Entities {
			invalidEntities[id] = entity
		}
		
		// Create an invalid entity with multiple unique attributes
		invalidEntity := models.Entity{
			DisplayName: "Invalid",
			ExternalId:  "test/Invalid",
			Description: "Entity with multiple unique attributes",
			Attributes: []models.Attribute{
				{
					Name:       "id1",
					ExternalId: "id1_ext",
					Type:       "string",
					UniqueId:   true, // First unique attribute
				},
				{
					Name:       "id2",
					ExternalId: "id2_ext",
					Type:       "string",
					UniqueId:   true, // Second unique attribute
				},
			},
		}
		
		invalidEntities["invalid_entity"] = invalidEntity
		invalidSOR.Entities = invalidEntities
		
		// Attempt to create a graph with the invalid entity
		graph, err := NewGraph(&invalidSOR)
		
		// Verify creation fails with appropriate error
		assert.Error(t, err)
		assert.Nil(t, graph)
		assert.Contains(t, err.Error(), "unique attribute")
	})
}

// TestRelationshipCreation tests the relationship creation from YAML
func TestRelationshipCreation(t *testing.T) {
	t.Run("should create relationships with correct entities and attributes", func(t *testing.T) {
		// Create a graph from the test SOR definition
		graph, err := NewGraph(testSORDefinition)
		require.NoError(t, err)
		
		// Get specific relationship
		rel, exists := graph.GetRelationship("user_to_userrole")
		require.True(t, exists)
		require.NotNil(t, rel)
		
		// Verify relationship properties
		assert.Equal(t, "user_to_userrole", rel.GetName())
		
		// Verify source and target entities
		sourceEntity := rel.GetSourceEntity()
		targetEntity := rel.GetTargetEntity()
		
		assert.Equal(t, "User", sourceEntity.GetName())
		assert.Equal(t, "UserRole", targetEntity.GetName())
		
		// Verify source and target attributes
		sourceAttr := rel.GetSourceAttribute()
		targetAttr := rel.GetTargetAttribute()
		
		assert.Equal(t, "id", sourceAttr.GetName())
		assert.Equal(t, "user_id", targetAttr.GetName())
		
		// Verify cardinality (should be 1:N as source has unique ID and target doesn't)
		assert.Equal(t, OneToMany, rel.GetCardinality())
		assert.True(t, rel.IsOneToMany())
	})

	t.Run("should validate relationship references existing entities and attributes", func(t *testing.T) {
		// Create a YAML definition with an invalid relationship
		invalidSOR := *testSORDefinition // Clone the test SOR definition
		
		// Make a deep copy of the relationships map
		invalidRelationships := make(map[string]models.Relationship)
		for id, rel := range testSORDefinition.Relationships {
			invalidRelationships[id] = rel
		}
		
		// Add an invalid relationship with non-existent attribute
		invalidRelationships["invalid_relationship"] = models.Relationship{
			DisplayName:   "INVALID_REL",
			Name:          "invalid_rel",
			FromAttribute: "nonexistent_attr", // This attribute doesn't exist
			ToAttribute:   "user_id",
		}
		
		invalidSOR.Relationships = invalidRelationships
		
		// Attempt to create a graph with the invalid relationship
		graph, err := NewGraph(&invalidSOR)
		
		// Verify creation fails with appropriate error
		assert.Error(t, err)
		assert.Nil(t, graph)
		assert.Contains(t, err.Error(), "attribute")
	})
}

// TestTopologicalSorting tests the topological sorting functionality
func TestTopologicalSorting(t *testing.T) {
	t.Run("should return entities in correct dependency order", func(t *testing.T) {
		// Create a graph from the test SOR definition
		graph, err := NewGraph(testSORDefinition)
		require.NoError(t, err)
		
		// Get the topological order
		order, err := graph.GetTopologicalOrder()
		require.NoError(t, err)
		require.NotNil(t, order)
		
		// The order should include all entities
		assert.Equal(t, 3, len(order))
		
		// Check that dependencies are met
		// Find positions of each entity
		userPos := -1
		rolePos := -1
		userRolePos := -1
		
		for i, entityID := range order {
			if entityID == "user_entity" {
				userPos = i
			} else if entityID == "role_entity" {
				rolePos = i
			} else if entityID == "user_role_entity" {
				userRolePos = i
			}
		}
		
		// All entities should be present
		assert.GreaterOrEqual(t, userPos, 0)
		assert.GreaterOrEqual(t, rolePos, 0)
		assert.GreaterOrEqual(t, userRolePos, 0)
		
		// UserRole should come after both User and Role since it depends on both
		assert.Greater(t, userRolePos, userPos)
		assert.Greater(t, userRolePos, rolePos)
	})

	t.Run("should detect circular dependencies", func(t *testing.T) {
		// Create a YAML definition with circular dependencies
		circularSOR := *testSORDefinition // Clone the test SOR definition
		
		// Make a deep copy of the relationships map
		circularRelationships := make(map[string]models.Relationship)
		for id, rel := range testSORDefinition.Relationships {
			circularRelationships[id] = rel
		}
		
		// Add a relationship that creates a circular dependency:
		// User depends on UserRole, which already depends on User
		circularRelationships["circular_dependency"] = models.Relationship{
			DisplayName:   "CIRCULAR_DEP",
			Name:          "circular_dep",
			FromAttribute: "user_id", // From UserRole.user_id
			ToAttribute:   "id",      // To User.id (creates a cycle)
		}
		
		circularSOR.Relationships = circularRelationships
		
		// Create a graph with the circular dependency
		graph, err := NewGraph(&circularSOR)
		require.NoError(t, err) // Graph creation should succeed
		
		// Attempt to get topological order
		order, err := graph.GetTopologicalOrder()
		
		// Verify it detects the circular dependency
		assert.Error(t, err)
		assert.Equal(t, ErrCircularDependency, err)
		assert.Nil(t, order)
	})
}

// TestAccessorMethods tests the graph accessor methods
func TestAccessorMethods(t *testing.T) {
	t.Run("should retrieve entities and relationships by ID", func(t *testing.T) {
		// Create a graph from the test SOR definition
		graph, err := NewGraph(testSORDefinition)
		require.NoError(t, err)
		
		// Test GetEntity
		entity, exists := graph.GetEntity("user_entity")
		assert.True(t, exists)
		assert.NotNil(t, entity)
		assert.Equal(t, "User", entity.GetName())
		
		// Test non-existent entity
		entity, exists = graph.GetEntity("nonexistent")
		assert.False(t, exists)
		assert.Nil(t, entity)
		
		// Test GetRelationship
		rel, exists := graph.GetRelationship("user_to_userrole")
		assert.True(t, exists)
		assert.NotNil(t, rel)
		assert.Equal(t, testSORDefinition.Relationships["user_to_userrole"].Name, rel.GetName())
		
		// Test non-existent relationship
		rel, exists = graph.GetRelationship("nonexistent")
		assert.False(t, exists)
		assert.Nil(t, rel)
	})

	t.Run("should retrieve all relationships for an entity", func(t *testing.T) {
		// Create a graph from the test SOR definition
		graph, err := NewGraph(testSORDefinition)
		require.NoError(t, err)
		
		// Get relationships for User entity
		userRels := graph.GetRelationshipsForEntity("user_entity")
		assert.Equal(t, 1, len(userRels))
		assert.Equal(t, testSORDefinition.Relationships["user_to_userrole"].Name, userRels[0].GetName())
		
		// Get relationships for UserRole entity (should have 2: one from User and one from Role)
		userRoleRels := graph.GetRelationshipsForEntity("user_role_entity")
		assert.Equal(t, 2, len(userRoleRels))
		
		// Test entity with no relationships
		nonExistentRels := graph.GetRelationshipsForEntity("nonexistent")
		assert.Empty(t, nonExistentRels)
	})
}
