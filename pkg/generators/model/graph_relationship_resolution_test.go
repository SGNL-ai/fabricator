package model

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRelationshipResolutionPriority tests that relationship attribute resolution follows the correct priority:
// 1. First try attributeAlias
// 2. Then try Entity.ExternalId + "." + Attribute.ExternalId format
// Note: DisplayName format is NOT supported - relationships must use ExternalId or attributeAlias
func TestRelationshipResolutionPriority(t *testing.T) {
	t.Run("should resolve relationships using attributeAlias first", func(t *testing.T) {
		def := &parser.SORDefinition{
			DisplayName: "AttributeAlias Priority Test",
			Description: "Test that attributeAlias has highest priority",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "ons-profile-read/user",
					Attributes: []parser.Attribute{
						{
							Name:           "id",
							ExternalId:     "id",
							Type:           "String",
							UniqueId:       true,
							AttributeAlias: "user-id-alias",
						},
					},
				},
				"profile": {
					DisplayName: "Profile",
					ExternalId:  "ons-profile-read/profile",
					Attributes: []parser.Attribute{
						{
							Name:           "id",
							ExternalId:     "id",
							Type:           "String",
							UniqueId:       true,
							AttributeAlias: "profile-id-alias",
						},
						{
							Name:           "userId",
							ExternalId:     "userId",
							Type:           "String",
							AttributeAlias: "profile-user-id-alias",
						},
					},
				},
			},
			Relationships: map[string]parser.Relationship{
				"profile_to_user": {
					Name:          "ProfileToUser",
					FromAttribute: "profile-user-id-alias", // Using attributeAlias
					ToAttribute:   "user-id-alias",         // Using attributeAlias
				},
			},
		}

		graph, err := NewGraph(def, 100)
		require.NoError(t, err, "Should resolve relationships using attributeAlias")
		assert.NotNil(t, graph)

		relationships := graph.GetAllRelationships()
		assert.Len(t, relationships, 1, "Should create 1 relationship using attributeAlias")
	})

	t.Run("should resolve relationships using ExternalId format when no attributeAlias", func(t *testing.T) {
		def := &parser.SORDefinition{
			DisplayName: "ExternalId Format Test",
			Description: "Test that ExternalId format works when no attributeAlias",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "ons-profile-read/user",
					Attributes: []parser.Attribute{
						{
							Name:       "id",
							ExternalId: "id",
							Type:       "String",
							UniqueId:   true,
							// No attributeAlias
						},
					},
				},
				"profile": {
					DisplayName: "Profile",
					ExternalId:  "ons-profile-read/profile",
					Attributes: []parser.Attribute{
						{
							Name:       "id",
							ExternalId: "id",
							Type:       "String",
							UniqueId:   true,
							// No attributeAlias
						},
						{
							Name:       "userId",
							ExternalId: "userId",
							Type:       "String",
							// No attributeAlias
						},
					},
				},
			},
			Relationships: map[string]parser.Relationship{
				"profile_to_user": {
					Name:          "ProfileToUser",
					FromAttribute: "ons-profile-read/profile.userId", // ExternalId format
					ToAttribute:   "ons-profile-read/user.id",        // ExternalId format
				},
			},
		}

		graph, err := NewGraph(def, 100)
		require.NoError(t, err, "Should resolve relationships using ExternalId format")
		assert.NotNil(t, graph)

		relationships := graph.GetAllRelationships()
		assert.Len(t, relationships, 1, "Should create 1 relationship using ExternalId format")
	})

	t.Run("should NOT support DisplayName format when DisplayName differs from ExternalId", func(t *testing.T) {
		def := &parser.SORDefinition{
			DisplayName: "DisplayName Format Test",
			Description: "Test that DisplayName format is rejected when it differs from ExternalId",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "ons-profile-read/user", // Different from DisplayName
					Attributes: []parser.Attribute{
						{
							Name:       "id",
							ExternalId: "id",
							Type:       "String",
							UniqueId:   true,
							// No attributeAlias
						},
					},
				},
				"profile": {
					DisplayName: "Profile",
					ExternalId:  "ons-profile-read/profile", // Different from DisplayName
					Attributes: []parser.Attribute{
						{
							Name:       "id",
							ExternalId: "id",
							Type:       "String",
							UniqueId:   true,
							// No attributeAlias
						},
						{
							Name:       "userId",
							ExternalId: "userId",
							Type:       "String",
							// No attributeAlias
						},
					},
				},
			},
			Relationships: map[string]parser.Relationship{
				"profile_to_user": {
					Name:          "ProfileToUser",
					FromAttribute: "Profile.userId", // DisplayName format - should FAIL
					ToAttribute:   "User.id",        // DisplayName format - should FAIL
				},
			},
		}

		graph, err := NewGraph(def, 100)
		assert.Error(t, err, "Should reject DisplayName format when ExternalId differs")
		assert.Nil(t, graph)
		assert.Contains(t, err.Error(), "source entity not found", "Should indicate entity not found")
	})

	t.Run("should work with ExternalId format when DisplayName equals ExternalId", func(t *testing.T) {
		def := &parser.SORDefinition{
			DisplayName: "DisplayName Equals ExternalId Test",
			Description: "Test that it works when DisplayName == ExternalId using ExternalId format (like Okta example)",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User", // Same as DisplayName
					Attributes: []parser.Attribute{
						{
							Name:       "id",
							ExternalId: "id",
							Type:       "String",
							UniqueId:   true,
						},
					},
				},
				"profile": {
					DisplayName: "Profile",
					ExternalId:  "Profile", // Same as DisplayName
					Attributes: []parser.Attribute{
						{
							Name:       "id",
							ExternalId: "id",
							Type:       "String",
							UniqueId:   true,
						},
						{
							Name:       "userId",
							ExternalId: "userId",
							Type:       "String",
						},
					},
				},
			},
			Relationships: map[string]parser.Relationship{
				"profile_to_user": {
					Name:          "ProfileToUser",
					FromAttribute: "Profile.userId", // This works because Profile == ExternalId == DisplayName
					ToAttribute:   "User.id",        // This works because User == ExternalId == DisplayName
				},
			},
		}

		graph, err := NewGraph(def, 100)
		require.NoError(t, err, "Should work when DisplayName == ExternalId")
		assert.NotNil(t, graph)

		relationships := graph.GetAllRelationships()
		assert.Len(t, relationships, 1, "Should create 1 relationship")
	})

	t.Run("should fail with helpful error when attribute not found", func(t *testing.T) {
		def := &parser.SORDefinition{
			DisplayName: "Invalid Relationship Test",
			Description: "Test error message when relationship attribute not found",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "ons-profile-read/user",
					Attributes: []parser.Attribute{
						{
							Name:       "id",
							ExternalId: "id",
							Type:       "String",
							UniqueId:   true,
						},
					},
				},
			},
			Relationships: map[string]parser.Relationship{
				"invalid_rel": {
					Name:          "InvalidRel",
					FromAttribute: "ons-profile-read/user.nonexistent", // This doesn't exist
					ToAttribute:   "ons-profile-read/user.id",
				},
			},
		}

		graph, err := NewGraph(def, 100)
		assert.Error(t, err, "Should fail when attribute not found")
		assert.Nil(t, graph)
		assert.Contains(t, err.Error(), "source entity not found", "Error should mention which entity was not found")
	})
}

// TestMixedRelationshipFormats tests that we can use different formats in the same YAML
func TestMixedRelationshipFormats(t *testing.T) {
	def := &parser.SORDefinition{
		DisplayName: "Mixed Format Test",
		Description: "Test using different relationship formats in same YAML",
		Entities: map[string]parser.Entity{
			"user": {
				DisplayName: "User",
				ExternalId:  "ons-profile-read/user",
				Attributes: []parser.Attribute{
					{
						Name:           "id",
						ExternalId:     "id",
						Type:           "String",
						UniqueId:       true,
						AttributeAlias: "user-id",
					},
				},
			},
			"profile": {
				DisplayName: "Profile",
				ExternalId:  "ons-profile-read/profile",
				Attributes: []parser.Attribute{
					{
						Name:       "id",
						ExternalId: "id",
						Type:       "String",
						UniqueId:   true,
						// No alias
					},
					{
						Name:           "userId",
						ExternalId:     "userId",
						Type:           "String",
						AttributeAlias: "profile-user-id",
					},
				},
			},
			"account": {
				DisplayName: "Account",
				ExternalId:  "Account", // Same as DisplayName
				Attributes: []parser.Attribute{
					{
						Name:       "id",
						ExternalId: "id",
						Type:       "String",
						UniqueId:   true,
					},
					{
						Name:       "profileId",
						ExternalId: "profileId",
						Type:       "String",
					},
				},
			},
		},
		Relationships: map[string]parser.Relationship{
			"profile_to_user": {
				Name:          "ProfileToUser",
				FromAttribute: "profile-user-id", // attributeAlias format
				ToAttribute:   "user-id",         // attributeAlias format
			},
			"account_to_profile": {
				Name:          "AccountToProfile",
				FromAttribute: "Account.profileId",           // DisplayName format
				ToAttribute:   "ons-profile-read/profile.id", // ExternalId format
			},
		},
	}

	graph, err := NewGraph(def, 100)
	require.NoError(t, err, "Should handle mixed relationship formats")
	assert.NotNil(t, graph)

	relationships := graph.GetAllRelationships()
	assert.Len(t, relationships, 2, "Should create 2 relationships with different formats")
}
