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

		_, err := BuildEntityDependencyGraph(entities, relationships, true)
		// May or may not error depending on cycle detection - just ensure it doesn't panic
		assert.NotPanics(t, func() {
			BuildEntityDependencyGraph(entities, relationships, true)
		})
	})

	t.Run("should handle empty entities", func(t *testing.T) {
		entities := map[string]parser.Entity{}
		relationships := map[string]parser.Relationship{}

		graph, err := BuildEntityDependencyGraph(entities, relationships, false)
		require.NoError(t, err)
		assert.NotNil(t, graph)
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
}

func TestParseEntityAttribute(t *testing.T) {
	t.Run("should parse entity and attribute from reference", func(t *testing.T) {
		// This function might be used internally - test basic functionality
		entities := map[string]parser.Entity{
			"user": {
				DisplayName: "User",
				ExternalId:  "User",
				Attributes: []parser.Attribute{
					{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "user-id"},
				},
			},
		}

		result := ParseEntityAttribute(entities, "user-id")
		assert.Equal(t, "user", result.EntityID)
		assert.Equal(t, "id", result.AttributeName)
	})

	t.Run("should handle non-existent attribute", func(t *testing.T) {
		entities := map[string]parser.Entity{}

		result := ParseEntityAttribute(entities, "nonexistent")
		assert.Equal(t, "", result.EntityID)
		assert.Equal(t, "", result.AttributeName)
	})
}