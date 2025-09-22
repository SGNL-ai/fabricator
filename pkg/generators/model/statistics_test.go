package model

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphStatistics(t *testing.T) {
	t.Run("should calculate basic statistics correctly", func(t *testing.T) {
		def := &parser.SORDefinition{
			DisplayName: "Test SOR",
			Description: "Test Description",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "TestApp/User",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true, Indexed: true},
						{Name: "name", ExternalId: "name", Type: "String", Indexed: false},
						{Name: "emails", ExternalId: "emails", Type: "String", List: true},
						{Name: "roleId", ExternalId: "roleId", Type: "String"}, // Add missing FK attribute
					},
				},
				"role": {
					DisplayName: "Role",
					ExternalId:  "TestApp/Role",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true, Indexed: true},
						{Name: "name", ExternalId: "name", Type: "String", Indexed: false},
					},
				},
			},
			Relationships: map[string]parser.Relationship{
				"user_role": {
					DisplayName:   "User Role",
					Name:          "user_role",
					FromAttribute: "User.roleId",
					ToAttribute:   "Role.id",
				},
				"role_path": {
					DisplayName: "Role Path",
					Name:        "role_path",
					Path: []parser.RelationshipPath{
						{Relationship: "user_role", Direction: "outbound"},
					},
				},
			},
		}

		graphInterface, err := NewGraph(def, 100)
		require.NoError(t, err)
		graph, ok := graphInterface.(*Graph)
		require.True(t, ok)

		stats := graph.GetStatistics()

		// Basic info
		assert.Equal(t, "Test SOR", stats.SORName)
		assert.Equal(t, "Test Description", stats.Description)

		// Entity and attribute counts
		assert.Equal(t, 2, stats.EntityCount)
		assert.Equal(t, 6, stats.TotalAttributes)   // 4 + 2 attributes
		assert.Equal(t, 2, stats.UniqueAttributes)  // 2 id attributes
		assert.Equal(t, 2, stats.IndexedAttributes) // 2 indexed attributes
		assert.Equal(t, 1, stats.ListAttributes)    // 1 list attribute

		// Relationship counts
		assert.Equal(t, 1, stats.RelationshipCount)      // Only direct relationships in graph
		assert.Equal(t, 1, stats.DirectRelationships)    // 1 direct
		assert.Equal(t, 1, stats.PathBasedRelationships) // 1 path-based (from YAML)

		// Namespace format
		assert.Len(t, stats.NamespaceFormats, 1)
		assert.Equal(t, 2, stats.NamespaceFormats["TestApp"])
	})

	t.Run("should handle entities without namespace", func(t *testing.T) {
		def := &parser.SORDefinition{
			DisplayName: "Simple SOR",
			Description: "Simple Description",
			Entities: map[string]parser.Entity{
				"entity1": {
					DisplayName: "Entity1",
					ExternalId:  "Entity1", // No namespace
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
					},
				},
				"entity2": {
					DisplayName: "Entity2",
					ExternalId:  "Entity2", // No namespace
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
					},
				},
			},
		}

		graphInterface, err := NewGraph(def, 100)
		require.NoError(t, err)
		graph, ok := graphInterface.(*Graph)
		require.True(t, ok)

		stats := graph.GetStatistics()

		// Should detect no namespace format
		assert.Len(t, stats.NamespaceFormats, 1)
		assert.Equal(t, 2, stats.NamespaceFormats["(no namespace)"])
	})

	t.Run("should handle mixed namespace formats", func(t *testing.T) {
		def := &parser.SORDefinition{
			DisplayName: "Mixed SOR",
			Description: "Mixed Description",
			Entities: map[string]parser.Entity{
				"app1_entity": {
					DisplayName: "App1 Entity",
					ExternalId:  "App1/Entity",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
					},
				},
				"app2_entity": {
					DisplayName: "App2 Entity",
					ExternalId:  "App2/Entity",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
					},
				},
				"plain_entity": {
					DisplayName: "Plain Entity",
					ExternalId:  "PlainEntity", // No namespace
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
					},
				},
			},
		}

		graphInterface, err := NewGraph(def, 100)
		require.NoError(t, err)
		graph, ok := graphInterface.(*Graph)
		require.True(t, ok)

		stats := graph.GetStatistics()

		// Should detect multiple namespace formats
		assert.Len(t, stats.NamespaceFormats, 3)
		assert.Equal(t, 1, stats.NamespaceFormats["App1"])
		assert.Equal(t, 1, stats.NamespaceFormats["App2"])
		assert.Equal(t, 1, stats.NamespaceFormats["(no namespace)"])
	})

	t.Run("should handle empty SOR gracefully", func(t *testing.T) {
		def := &parser.SORDefinition{
			DisplayName:   "Empty SOR",
			Description:   "Empty Description",
			Entities:      map[string]parser.Entity{},
			Relationships: map[string]parser.Relationship{},
		}

		// This should fail during graph creation due to validation
		_, err := NewGraph(def, 100)
		assert.Error(t, err, "Should fail for empty SOR")
		assert.Contains(t, err.Error(), "at least one entity", "Should mention entity requirement")
	})
}
