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
		"User": {
			DisplayName: "User",
			ExternalId:  "User",
			Description: "User entity for testing",
			Attributes: []models.Attribute{
				{
					Name:        "id",
					ExternalId:  "id",
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
		"Role": {
			DisplayName: "Role",
			ExternalId:  "Role",
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
		"UserRole": {
			DisplayName: "UserRole",
			ExternalId:  "UserRole",
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
			FromAttribute: "UserRole.user_ref",
			ToAttribute:   "User.id",
		},
		"role_to_userrole": {
			DisplayName:   "ROLE_TO_USERROLE",
			Name:          "role_to_userrole",
			FromAttribute: "User.id",
			ToAttribute:   "UserRole.user_ref",
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
		userEntity, exists := graph.GetEntity("User")
		require.True(t, exists)
		require.NotNil(t, userEntity)

		// Verify entity properties
		assert.Equal(t, "User", userEntity.GetName())
		assert.Equal(t, "User", userEntity.GetExternalID())

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

		// The relationship is defined as UserRole.user_ref -> User.id
		assert.Equal(t, "UserRole", sourceEntity.GetName())
		assert.Equal(t, "User", targetEntity.GetName())

		// Verify source and target attributes
		sourceAttr := rel.GetSourceAttribute()
		targetAttr := rel.GetTargetAttribute()

		assert.Equal(t, "user_id", sourceAttr.GetName())
		assert.Equal(t, "id", targetAttr.GetName())

		// Verify cardinality (FK->PK means N:1)
		assert.Equal(t, ManyToOne, rel.GetCardinality())
		assert.True(t, rel.IsManyToOne())
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
			switch entityID {
			case "User":
				userPos = i
			case "Role":
				rolePos = i
			case "UserRole":
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
		// Create a SOR definition with a circular dependency between three entities
		circularSOR := &models.SORDefinition{
			DisplayName:   "Circular SOR",
			Description:   "Test System of Record with circular dependencies",
			Hostname:      "circular.example.com",
			Type:          "Circular-1.0.0",
			AdapterConfig: "",
			// Entities with circular relationships
			Entities: map[string]models.Entity{
				"EntityA": {
					DisplayName: "EntityA",
					ExternalId:  "EntityA",
					Description: "Entity A for circular dependency test",
					Attributes: []models.Attribute{
						{
							Name:        "id",
							ExternalId:  "id",
							Description: "Primary key",
							Type:        "string",
							UniqueId:    true,
						},
						{
							Name:        "b_id",
							ExternalId:  "b_id",
							Description: "Reference to Entity C",
							Type:        "string",
							UniqueId:    false,
						},
					},
				},
				"EntityB": {
					DisplayName: "EntityB",
					ExternalId:  "EntityB",
					Description: "Entity B for circular dependency test",
					Attributes: []models.Attribute{
						{
							Name:        "id",
							ExternalId:  "id",
							Description: "Primary key",
							Type:        "string",
							UniqueId:    true,
						},
						{
							Name:        "a_id",
							ExternalId:  "a_id",
							Description: "Reference to Entity A",
							Type:        "string",
							UniqueId:    false,
						},
					},
				},
			},
			// Three relationships that create a cycle: A -> B -> A
			Relationships: map[string]models.Relationship{
				"a_to_b": {
					DisplayName:   "a_to_c",
					Name:          "a_to_c",
					FromAttribute: "EntityA.b_id", //FK
					ToAttribute:   "EntityB.id",   //PK
				},
				"b_to_a": {
					DisplayName:   "b_to_a",
					Name:          "b_to_a",
					FromAttribute: "EntityB.a_id", //FK
					ToAttribute:   "EntityA.id",   //PK
				},
			},
		}

		// Create a graph with the circular dependency
		graph, err := NewGraph(circularSOR)
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
		entity, exists := graph.GetEntity("User")
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
		userRels := graph.GetRelationshipsForEntity("User")
		// We should have at least one relationship
		assert.GreaterOrEqual(t, len(userRels), 1)

		// Check that one of the relationships is user_to_userrole
		found := false
		for _, rel := range userRels {
			if rel.GetName() == "user_to_userrole" {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected to find user_to_userrole relationship")

		// Get relationships for UserRole entity (should have relationships)
		userRoleRels := graph.GetRelationshipsForEntity("UserRole")
		assert.NotEmpty(t, userRoleRels, "UserRole entity should have relationships")

		// Test entity with no relationships
		nonExistentRels := graph.GetRelationshipsForEntity("nonexistent")
		assert.Empty(t, nonExistentRels)
	})
}

// Test for attribute reference bug in relationship creation
func TestGraphDottedAttributeReferenceBug(t *testing.T) {
	// This test should fail initially, proving the bug exists
	def := &models.SORDefinition{
		DisplayName: "Test SOR",
		Description: "Test Description",
		Entities: map[string]models.Entity{
			"user": {
				DisplayName: "User",
				ExternalId:  "User",
				Attributes: []models.Attribute{
					{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
					{Name: "profileId", ExternalId: "profileId", Type: "String"}, // externalID is "profileId"
				},
			},
			"profile": {
				DisplayName: "Profile",
				ExternalId:  "Profile",
				Attributes: []models.Attribute{
					{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
				},
			},
		},
		Relationships: map[string]models.Relationship{
			"user_profile": {
				DisplayName:   "User Profile",
				Name:          "user_profile",
				FromAttribute: "user.profileId", // Graph maps this to entity, then passes "user.profileId" to GetAttributeByExternalID
				ToAttribute:   "profile.id",     // But GetAttributeByExternalID expects just "profileId", not "user.profileId"
			},
		},
	}

	// After fix: This should now succeed because dotted notation is properly handled
	graph, err := NewGraph(def)
	assert.NoError(t, err, "Should succeed after dotted notation bug fix")
	assert.NotNil(t, graph)

	// Verify the relationship was created successfully
	relationships := graph.GetAllRelationships()
	assert.Len(t, relationships, 1, "Should have created 1 relationship")
}

// Test UUID attribute reference format (like in real sample.yaml)
func TestGraphUUIDAttributeReferences(t *testing.T) {
	def := &models.SORDefinition{
		DisplayName: "Test SOR",
		Description: "Test Description",
		Entities: map[string]models.Entity{
			"user": {
				DisplayName: "User",
				ExternalId:  "User",
				Attributes: []models.Attribute{
					{Name: "id", ExternalId: "id", Type: "String", UniqueId: true, AttributeAlias: "user-id-uuid"},
					{Name: "profileId", ExternalId: "profileId", Type: "String", AttributeAlias: "user-profileid-uuid"},
				},
			},
			"profile": {
				DisplayName: "Profile",
				ExternalId:  "Profile",
				Attributes: []models.Attribute{
					{Name: "id", ExternalId: "id", Type: "String", UniqueId: true, AttributeAlias: "profile-id-uuid"},
				},
			},
		},
		Relationships: map[string]models.Relationship{
			"user_profile": {
				DisplayName:   "User Profile",
				Name:          "user_profile",
				FromAttribute: "user-profileid-uuid", // UUID format like real sample.yaml
				ToAttribute:   "profile-id-uuid",     // UUID format like real sample.yaml
			},
		},
	}

	graph, err := NewGraph(def)
	assert.NoError(t, err, "Should handle UUID attribute references")
	assert.NotNil(t, graph)

	// Verify the relationship was created successfully
	relationships := graph.GetAllRelationships()
	assert.Len(t, relationships, 1, "Should have created 1 relationship with UUID references")

	// Verify relationship metadata is set correctly
	rel := relationships[0]
	sourceAttr := rel.GetSourceAttribute()
	targetAttr := rel.GetTargetAttribute()

	assert.Equal(t, "profileId", sourceAttr.GetName(), "Source attribute should be profileId")
	assert.Equal(t, "id", targetAttr.GetName(), "Target attribute should be id")

	// Check that source attribute has correct relationship metadata
	assert.True(t, sourceAttr.IsRelationship(), "Source attribute should be marked as relationship")
	assert.Equal(t, "profile", sourceAttr.GetRelatedEntityID(), "Should point to profile entity, not UUID")
	assert.Equal(t, "id", sourceAttr.GetRelatedAttribute(), "Should point to id attribute")
}

// Test with UUID entity IDs (like real sample.yaml)
func TestGraphUUIDEntityIDs(t *testing.T) {
	def := &models.SORDefinition{
		DisplayName: "Test SOR",
		Description: "Test Description",
		Entities: map[string]models.Entity{
			"uuid-entity-1": { // UUID as entity ID
				DisplayName: "Entity1",
				ExternalId:  "Entity1",
				Attributes: []models.Attribute{
					{Name: "id", ExternalId: "id", Type: "String", UniqueId: true, AttributeAlias: "id-alias"},
					{Name: "refField", ExternalId: "refField", Type: "String", AttributeAlias: "ref-alias"},
				},
			},
			"uuid-entity-2": { // UUID as entity ID
				DisplayName: "Entity2",
				ExternalId:  "Entity2",
				Attributes: []models.Attribute{
					{Name: "id", ExternalId: "id", Type: "String", UniqueId: true, AttributeAlias: "target-id-alias"},
				},
			},
		},
		Relationships: map[string]models.Relationship{
			"rel1": {
				DisplayName:   "Test Relationship",
				Name:          "test_rel",
				FromAttribute: "ref-alias",     // UUID alias format
				ToAttribute:   "target-id-alias", // UUID alias format
			},
		},
	}

	graph, err := NewGraph(def)
	require.NoError(t, err)

	// Check relationship metadata
	relationships := graph.GetAllRelationships()
	require.Len(t, relationships, 1)

	rel := relationships[0]
	sourceAttr := rel.GetSourceAttribute()

	// This should show the bug - relatedEntityID might be wrong
	assert.Equal(t, "uuid-entity-2", sourceAttr.GetRelatedEntityID(), "Should be target entity ID, not dotted reference")
	assert.Equal(t, "id", sourceAttr.GetRelatedAttribute(), "Should be target attribute name")

	// CRITICAL: Target ID attributes should NOT be marked as relationships
	targetAttr := rel.GetTargetAttribute()
	assert.False(t, targetAttr.IsRelationship(), "Target ID attribute should NOT be marked as relationship")
	assert.True(t, targetAttr.IsUnique(), "Target should remain as primary key")
}
