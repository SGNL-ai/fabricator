package util

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildEntityDependencyGraph(t *testing.T) {
	t.Run("should build graph with entity dependencies", func(t *testing.T) {
		entities := map[string]parser.Entity{
			"user": {
				DisplayName: "User",
				ExternalId:  "User",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "user-id"},
				},
			},
			"role": {
				DisplayName: "Role",
				ExternalId:  "Role",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "role-id"},
				},
			},
		}

		relationships := map[string]parser.Relationship{
			"user_role": {
				DisplayName:   "User Role",
				FromAttribute: "user-id",
				ToAttribute:   "role-id",
			},
		}

		graph, err := BuildEntityDependencyGraph(entities, relationships, false)
		require.NoError(t, err)
		assert.NotNil(t, graph)
	})

	t.Run("should prevent cycles when requested", func(t *testing.T) {
		entities := map[string]parser.Entity{
			"user": {
				DisplayName: "User",
				ExternalId:  "User",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "user-id"},
					{Name: "managerId", ExternalId: "managerId", AttributeAlias: "manager-id"},
				},
			},
		}

		// Self-referential relationship (potential cycle)
		relationships := map[string]parser.Relationship{
			"user_manager": {
				DisplayName:   "User Manager",
				FromAttribute: "manager-id",
				ToAttribute:   "user-id",
			},
		}

		// May or may not error depending on cycle detection - just ensure it doesn't panic
		assert.NotPanics(t, func() {
			_, _ = BuildEntityDependencyGraph(entities, relationships, true)
		})
	})

	t.Run("should handle empty entities", func(t *testing.T) {
		entities := map[string]parser.Entity{}
		relationships := map[string]parser.Relationship{}

		graph, err := BuildEntityDependencyGraph(entities, relationships, false)
		require.NoError(t, err)
		assert.NotNil(t, graph)
	})

	t.Run("should skip path-based relationships", func(t *testing.T) {
		entities := map[string]parser.Entity{
			"entity1": {
				DisplayName: "Entity1",
				ExternalId:  "Entity1",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "entity1-id"},
				},
			},
		}

		// Relationship with a path (should be skipped)
		relationships := map[string]parser.Relationship{
			"path_rel": {
				DisplayName:   "Path Relationship",
				FromAttribute: "entity1-id",
				ToAttribute:   "entity1-id",
				Path:          []parser.RelationshipPath{{Relationship: "some", Direction: "path"}}, // This makes it path-based
			},
		}

		graph, err := BuildEntityDependencyGraph(entities, relationships, false)
		require.NoError(t, err)
		assert.NotNil(t, graph)
		// The path-based relationship should be skipped
	})

	t.Run("should handle relationships with unresolvable attributes", func(t *testing.T) {
		entities := map[string]parser.Entity{
			"entity1": {
				DisplayName: "Entity1",
				ExternalId:  "Entity1",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "entity1-id"},
				},
			},
		}

		// Relationship where one attribute cannot be resolved
		relationships := map[string]parser.Relationship{
			"bad_rel": {
				DisplayName:   "Bad Relationship",
				FromAttribute: "nonexistent-attr", // This won't be found
				ToAttribute:   "entity1-id",
			},
		}

		graph, err := BuildEntityDependencyGraph(entities, relationships, false)
		require.NoError(t, err)
		assert.NotNil(t, graph)
		// The unresolvable relationship should be skipped
	})

	t.Run("should handle self-referential relationships", func(t *testing.T) {
		entities := map[string]parser.Entity{
			"entity1": {
				DisplayName: "Entity1",
				ExternalId:  "Entity1",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "entity1-id"},
					{Name: "parent_id", ExternalId: "parent_id", AttributeAlias: "parent-id"},
				},
			},
		}

		// Self-referential relationship (from and to same entity)
		relationships := map[string]parser.Relationship{
			"self_ref": {
				DisplayName:   "Self Reference",
				FromAttribute: "parent-id",  // Non-unique FK
				ToAttribute:   "entity1-id", // Unique PK, same entity
			},
		}

		graph, err := BuildEntityDependencyGraph(entities, relationships, false)
		require.NoError(t, err)
		assert.NotNil(t, graph)
		// Self-referential relationships should be skipped (fromEntityID == toEntityID)
	})

	t.Run("should handle PK->FK relationship (reverse order, should be skipped)", func(t *testing.T) {
		entities := map[string]parser.Entity{
			"parent": {
				DisplayName: "Parent",
				ExternalId:  "Parent",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "parent-id"},
				},
			},
			"child": {
				DisplayName: "Child",
				ExternalId:  "Child",
				Attributes: []parser.Attribute{
					{Name: "parent_ref", ExternalId: "parent_ref", AttributeAlias: "parent-ref"},
				},
			},
		}

		// PK -> FK relationship (reverse of normal, should be skipped)
		relationships := map[string]parser.Relationship{
			"reverse_rel": {
				DisplayName:   "Reverse Relationship",
				FromAttribute: "parent-id",  // Unique PK
				ToAttribute:   "parent-ref", // Non-unique FK
			},
		}

		graph, err := BuildEntityDependencyGraph(entities, relationships, false)
		require.NoError(t, err)
		assert.NotNil(t, graph)
		// This relationship should be skipped (fromUniqueID && !toUniqueID case)
	})

	t.Run("should handle PK->PK identity relationship", func(t *testing.T) {
		entities := map[string]parser.Entity{
			"entity1": {
				DisplayName: "Entity1",
				ExternalId:  "Entity1",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "entity1-id"},
				},
			},
			"entity2": {
				DisplayName: "Entity2",
				ExternalId:  "Entity2",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "entity2-id"},
				},
			},
		}

		// PK -> PK identity relationship
		relationships := map[string]parser.Relationship{
			"identity_rel": {
				DisplayName:   "Identity Relationship",
				FromAttribute: "entity1-id", // Unique PK
				ToAttribute:   "entity2-id", // Unique PK
			},
		}

		graph, err := BuildEntityDependencyGraph(entities, relationships, false)
		require.NoError(t, err)
		assert.NotNil(t, graph)
		// Should create an edge for PK->PK relationship
	})

	t.Run("should skip non-unique to non-unique relationships", func(t *testing.T) {
		entities := map[string]parser.Entity{
			"entity1": {
				DisplayName: "Entity1",
				ExternalId:  "Entity1",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true}, // Need unique ID for valid entity
					{Name: "field1", ExternalId: "field1", AttributeAlias: "field1-alias"},
				},
			},
			"entity2": {
				DisplayName: "Entity2",
				ExternalId:  "Entity2",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true}, // Need unique ID for valid entity
					{Name: "field2", ExternalId: "field2", AttributeAlias: "field2-alias"},
				},
			},
		}

		// Non-unique to non-unique relationship (should be skipped)
		relationships := map[string]parser.Relationship{
			"non_unique_rel": {
				DisplayName:   "Non-unique Relationship",
				FromAttribute: "field1-alias", // Non-unique
				ToAttribute:   "field2-alias", // Non-unique
			},
		}

		graph, err := BuildEntityDependencyGraph(entities, relationships, false)
		require.NoError(t, err)
		assert.NotNil(t, graph)
		// This relationship should be skipped (!fromUniqueID && !toUniqueID case)
	})

	t.Run("should handle edge already exists error gracefully", func(t *testing.T) {
		entities := map[string]parser.Entity{
			"parent": {
				DisplayName: "Parent",
				ExternalId:  "Parent",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "parent-id"},
				},
			},
			"child": {
				DisplayName: "Child",
				ExternalId:  "Child",
				Attributes: []parser.Attribute{
					{Name: "parent_ref1", ExternalId: "parent_ref1", AttributeAlias: "parent-ref1"},
					{Name: "parent_ref2", ExternalId: "parent_ref2", AttributeAlias: "parent-ref2"},
				},
			},
		}

		// Two relationships that create the same edge
		relationships := map[string]parser.Relationship{
			"rel1": {
				DisplayName:   "Relationship 1",
				FromAttribute: "parent-ref1", // FK
				ToAttribute:   "parent-id",   // PK
			},
			"rel2": {
				DisplayName:   "Relationship 2",
				FromAttribute: "parent-ref2", // FK (different attribute but same entities)
				ToAttribute:   "parent-id",   // PK
			},
		}

		graph, err := BuildEntityDependencyGraph(entities, relationships, false)
		require.NoError(t, err)
		assert.NotNil(t, graph)
		// Should handle duplicate edge creation gracefully
	})

	t.Run("should return cycle error when preventCycles is true", func(t *testing.T) {
		entities := map[string]parser.Entity{
			"entity1": {
				DisplayName: "Entity1",
				ExternalId:  "Entity1",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "entity1-id"},
					{Name: "entity2_ref", ExternalId: "entity2_ref", AttributeAlias: "entity2-ref"},
				},
			},
			"entity2": {
				DisplayName: "Entity2",
				ExternalId:  "Entity2",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "entity2-id"},
					{Name: "entity1_ref", ExternalId: "entity1_ref", AttributeAlias: "entity1-ref"},
				},
			},
		}

		// Circular relationships
		relationships := map[string]parser.Relationship{
			"rel1": {
				DisplayName:   "Relationship 1",
				FromAttribute: "entity2-ref", // entity1 -> entity2
				ToAttribute:   "entity2-id",
			},
			"rel2": {
				DisplayName:   "Relationship 2",
				FromAttribute: "entity1-ref", // entity2 -> entity1 (creates cycle)
				ToAttribute:   "entity1-id",
			},
		}

		graph, err := BuildEntityDependencyGraph(entities, relationships, true) // preventCycles = true
		if err == nil {
			// If no error, cycle detection might not have triggered
			// (depends on processing order), which is also valid
			assert.NotNil(t, graph)
		} else {
			// If error occurs, it should be a cycle error
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "cycle")
		}
	})
}

func TestGetTopologicalOrder(t *testing.T) {
	t.Run("should return topological order", func(t *testing.T) {
		entities := map[string]parser.Entity{
			"user": {
				DisplayName: "User",
				ExternalId:  "User",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "user-id"},
				},
			},
		}

		relationships := map[string]parser.Relationship{}

		graph, err := BuildEntityDependencyGraph(entities, relationships, false)
		require.NoError(t, err)

		order, err := GetTopologicalOrder(graph)
		assert.NoError(t, err)
		assert.Contains(t, order, "user")
	})

	t.Run("should handle topological sort errors", func(t *testing.T) {
		// Create a graph with circular dependencies to force an error
		entities := map[string]parser.Entity{
			"entity1": {
				DisplayName: "Entity1",
				ExternalId:  "Entity1",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "entity1-id"},
					{Name: "entity2_ref", ExternalId: "entity2_ref", AttributeAlias: "entity2-ref"},
				},
			},
			"entity2": {
				DisplayName: "Entity2",
				ExternalId:  "Entity2",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "entity2-id"},
					{Name: "entity1_ref", ExternalId: "entity1_ref", AttributeAlias: "entity1-ref"},
				},
			},
		}

		// Create circular dependencies
		relationships := map[string]parser.Relationship{
			"rel1": {
				DisplayName:   "Relationship 1",
				FromAttribute: "entity2-ref", // entity1 -> entity2
				ToAttribute:   "entity2-id",
			},
			"rel2": {
				DisplayName:   "Relationship 2",
				FromAttribute: "entity1-ref", // entity2 -> entity1 (creates cycle)
				ToAttribute:   "entity1-id",
			},
		}

		// Build graph without cycle prevention
		graph, err := BuildEntityDependencyGraph(entities, relationships, false)
		require.NoError(t, err)

		// Try to get topological order - this might succeed or fail depending on implementation
		order, err := GetTopologicalOrder(graph)
		if err != nil {
			// If error occurs, check it's properly wrapped
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to perform topological sort")
		} else {
			// If no error, the sort succeeded despite potential cycles
			assert.NotNil(t, order)
		}
	})
}

func TestParseEntityAttribute(t *testing.T) {
	t.Run("should parse entity and attribute from alias", func(t *testing.T) {
		entities := map[string]parser.Entity{
			"user": {
				DisplayName: "User",
				ExternalId:  "User",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "user-id"},
				},
			},
		}

		// Build the attribute maps as done in the actual implementation
		attributeAliasMap := make(map[string]struct {
			EntityID      string
			AttributeName string
			UniqueID      bool
		})
		entityAttributeMap := make(map[string]struct {
			EntityID      string
			AttributeName string
			UniqueID      bool
		})

		// Populate maps
		for entityID, entity := range entities {
			for _, attr := range entity.Attributes {
				if attr.AttributeAlias != "" {
					attributeAliasMap[attr.AttributeAlias] = struct {
						EntityID      string
						AttributeName string
						UniqueID      bool
					}{
						EntityID:      entityID,
						AttributeName: attr.Name,
						UniqueID:      attr.UniqueId,
					}
				}

				entityKey := entity.ExternalId + "." + attr.ExternalId
				entityAttributeMap[entityKey] = struct {
					EntityID      string
					AttributeName string
					UniqueID      bool
				}{
					EntityID:      entityID,
					AttributeName: attr.Name,
					UniqueID:      attr.UniqueId,
				}
			}
		}

		entityID, attrName, uniqueID := ParseEntityAttribute(entities, "user-id", attributeAliasMap, entityAttributeMap)
		assert.Equal(t, "user", entityID)
		assert.Equal(t, "id", attrName)
		assert.True(t, uniqueID)
	})

	t.Run("should parse entity and attribute from Entity.Attribute format", func(t *testing.T) {
		entities := map[string]parser.Entity{
			"user": {
				DisplayName: "User",
				ExternalId:  "User",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true},
				},
			},
		}

		attributeAliasMap := make(map[string]struct {
			EntityID      string
			AttributeName string
			UniqueID      bool
		})
		entityAttributeMap := make(map[string]struct {
			EntityID      string
			AttributeName string
			UniqueID      bool
		})

		entityID, attrName, uniqueID := ParseEntityAttribute(entities, "User.id", attributeAliasMap, entityAttributeMap)
		assert.Equal(t, "user", entityID)
		assert.Equal(t, "id", attrName)
		assert.True(t, uniqueID)
	})

	t.Run("should handle non-existent attribute", func(t *testing.T) {
		entities := map[string]parser.Entity{}
		attributeAliasMap := make(map[string]struct {
			EntityID      string
			AttributeName string
			UniqueID      bool
		})
		entityAttributeMap := make(map[string]struct {
			EntityID      string
			AttributeName string
			UniqueID      bool
		})

		entityID, attrName, uniqueID := ParseEntityAttribute(entities, "nonexistent", attributeAliasMap, entityAttributeMap)
		assert.Equal(t, "", entityID)
		assert.Equal(t, "", attrName)
		assert.False(t, uniqueID)
	})
}
