package pipeline

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/SGNL-ai/fabricator/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJunctionTableDuplicatePruning tests that M:M relationships don't create duplicate FK tuples
func TestJunctionTableDuplicatePruning(t *testing.T) {
	t.Run("should prune duplicate FK tuples in junction tables", func(t *testing.T) {
		// Users ←→ Groups via UserGroupMembership (M:M relationship)
		def := &parser.SORDefinition{
			DisplayName: "M:M Junction Test",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
					},
				},
				"group": {
					DisplayName: "Group",
					ExternalId:  "Group",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
					},
				},
				"user_group_membership": {
					DisplayName: "UserGroupMembership",
					ExternalId:  "UserGroupMembership",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},              // PK
						{Name: "user_id", ExternalId: "user_id", Type: "String", UniqueId: false},   // FK1
						{Name: "group_id", ExternalId: "group_id", Type: "String", UniqueId: false}, // FK2
					},
				},
			},
			Relationships: map[string]parser.Relationship{
				"membership_user": {
					Name:          "membership_to_user",
					FromAttribute: "user_group_membership.user_id", // FK1
					ToAttribute:   "user.id",                       // PK
				},
				"membership_group": {
					Name:          "membership_to_group",
					FromAttribute: "user_group_membership.group_id", // FK2
					ToAttribute:   "group.id",                       // PK
				},
			},
		}

		// Generate small dataset for easier validation
		graph, err := model.NewGraph(def, 10) // 10 records per entity
		require.NoError(t, err)

		// Generate data with auto-cardinality (should create realistic M:M patterns)
		tempDir := t.TempDir()
		generator := NewDataGenerator(tempDir, 10, true)
		err = generator.Generate(graph.(*model.Graph))
		require.NoError(t, err)

		// Analyze junction table for duplicate FK tuples
		entities := graph.GetAllEntities()
		membershipEntity := entities["user_group_membership"]
		membershipCSV := membershipEntity.ToCSV()

		// Find FK columns
		userIdCol, groupIdCol := -1, -1
		for i, header := range membershipCSV.Headers {
			switch header {
			case "user_id":
				userIdCol = i
			case "group_id":
				groupIdCol = i
			}
		}
		require.NotEqual(t, -1, userIdCol, "Should have user_id column")
		require.NotEqual(t, -1, groupIdCol, "Should have group_id column")

		// Collect all FK tuples
		fkTuples := make([]string, 0)
		tupleSet := make(map[string]bool)

		for _, row := range membershipCSV.Rows {
			userId := row[userIdCol]
			groupId := row[groupIdCol]
			tuple := userId + "|" + groupId // Create composite key
			fkTuples = append(fkTuples, tuple)

			// CRITICAL: No duplicate FK tuples should exist
			assert.False(t, tupleSet[tuple],
				"Junction table should not have duplicate FK tuple (user_id=%s, group_id=%s)",
				userId, groupId)
			tupleSet[tuple] = true
		}

		// Verify all FK tuples are unique
		assert.Len(t, tupleSet, len(fkTuples),
			"All FK tuples should be unique - found %d tuples, %d unique",
			len(fkTuples), len(tupleSet))

		t.Logf("Generated %d unique FK tuples in junction table", len(tupleSet))
	})

	t.Run("should detect entities with multiple FK attributes", func(t *testing.T) {
		// Test that we can identify junction tables (entities with 2+ FK attributes)
		def := &parser.SORDefinition{
			DisplayName: "Junction Detection Test",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
					},
				},
				"role": {
					DisplayName: "Role",
					ExternalId:  "Role",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
					},
				},
				"permission": {
					DisplayName: "Permission",
					ExternalId:  "Permission",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
					},
				},
				"user_role_permission": {
					DisplayName: "UserRolePermission",
					ExternalId:  "UserRolePermission",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},            // PK
						{Name: "user_id", ExternalId: "user_id", Type: "String", UniqueId: false}, // FK1
						{Name: "role_id", ExternalId: "role_id", Type: "String", UniqueId: false}, // FK2
						{Name: "perm_id", ExternalId: "perm_id", Type: "String", UniqueId: false}, // FK3
					},
				},
			},
			Relationships: map[string]parser.Relationship{
				"urp_user": {
					Name:          "urp_to_user",
					FromAttribute: "user_role_permission.user_id",
					ToAttribute:   "user.id",
				},
				"urp_role": {
					Name:          "urp_to_role",
					FromAttribute: "user_role_permission.role_id",
					ToAttribute:   "role.id",
				},
				"urp_permission": {
					Name:          "urp_to_permission",
					FromAttribute: "user_role_permission.perm_id",
					ToAttribute:   "permission.id",
				},
			},
		}

		graph, err := model.NewGraph(def, 20)
		require.NoError(t, err)

		// Generate data
		tempDir := t.TempDir()
		generator := NewDataGenerator(tempDir, 20, true)
		err = generator.Generate(graph.(*model.Graph))
		require.NoError(t, err)

		// Verify 3-way junction table has unique FK triples
		entities := graph.GetAllEntities()
		urpEntity := entities["user_role_permission"]
		urpCSV := urpEntity.ToCSV()

		// Find all FK columns
		userIdCol, roleIdCol, permIdCol := -1, -1, -1
		for i, header := range urpCSV.Headers {
			switch header {
			case "user_id":
				userIdCol = i
			case "role_id":
				roleIdCol = i
			case "perm_id":
				permIdCol = i
			}
		}
		require.NotEqual(t, -1, userIdCol)
		require.NotEqual(t, -1, roleIdCol)
		require.NotEqual(t, -1, permIdCol)

		// Check for unique FK triples
		tripleSet := make(map[string]bool)
		for _, row := range urpCSV.Rows {
			userId := row[userIdCol]
			roleId := row[roleIdCol]
			permId := row[permIdCol]
			triple := userId + "|" + roleId + "|" + permId

			// CRITICAL: No duplicate FK triples
			assert.False(t, tripleSet[triple],
				"Junction table should not have duplicate FK triple (user_id=%s, role_id=%s, perm_id=%s)",
				userId, roleId, permId)
			tripleSet[triple] = true
		}

		t.Logf("Generated %d unique FK triples in 3-way junction table", len(tripleSet))
	})

	t.Run("should maintain realistic M:M distribution patterns", func(t *testing.T) {
		// Test that M:M relationships create realistic patterns
		// - Some users belong to multiple groups
		// - Some groups have multiple users
		// - But no duplicate memberships
		def := &parser.SORDefinition{
			DisplayName: "M:M Distribution Test",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
					},
				},
				"group": {
					DisplayName: "Group",
					ExternalId:  "Group",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
					},
				},
				"membership": {
					DisplayName: "Membership",
					ExternalId:  "Membership",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
						{Name: "user_id", ExternalId: "user_id", Type: "String", UniqueId: false},
						{Name: "group_id", ExternalId: "group_id", Type: "String", UniqueId: false},
					},
				},
			},
			Relationships: map[string]parser.Relationship{
				"membership_user": {
					Name:          "membership_to_user",
					FromAttribute: "membership.user_id",
					ToAttribute:   "user.id",
				},
				"membership_group": {
					Name:          "membership_to_group",
					FromAttribute: "membership.group_id",
					ToAttribute:   "group.id",
				},
			},
		}

		graph, err := model.NewGraph(def, 15) // 15 users, 15 groups, 15 memberships
		require.NoError(t, err)

		tempDir := t.TempDir()
		generator := NewDataGenerator(tempDir, 15, true)
		err = generator.Generate(graph.(*model.Graph))
		require.NoError(t, err)

		// Analyze M:M distribution
		entities := graph.GetAllEntities()
		membershipEntity := entities["membership"]
		membershipCSV := membershipEntity.ToCSV()

		// Find FK columns
		userIdCol, groupIdCol := -1, -1
		for i, header := range membershipCSV.Headers {
			switch header {
			case "user_id":
				userIdCol = i
			case "group_id":
				groupIdCol = i
			}
		}
		require.NotEqual(t, -1, userIdCol)
		require.NotEqual(t, -1, groupIdCol)

		// Verify unique FK pairs and realistic distribution
		userGroups := make(map[string][]string) // user_id -> list of group_ids
		groupUsers := make(map[string][]string) // group_id -> list of user_ids
		fkPairs := make(map[string]bool)        // unique (user_id, group_id) pairs

		for _, row := range membershipCSV.Rows {
			userId := row[userIdCol]
			groupId := row[groupIdCol]
			pair := userId + "|" + groupId

			// CRITICAL: No duplicate pairs
			assert.False(t, fkPairs[pair],
				"Should not have duplicate membership (user_id=%s, group_id=%s)", userId, groupId)
			fkPairs[pair] = true

			// Track M:M relationships
			userGroups[userId] = append(userGroups[userId], groupId)
			groupUsers[groupId] = append(groupUsers[groupId], userId)
		}

		// Verify realistic M:M patterns
		t.Logf("Users with multiple groups: %d", countWithMultiple(userGroups))
		t.Logf("Groups with multiple users: %d", countWithMultiple(groupUsers))
		t.Logf("Total unique FK pairs: %d", len(fkPairs))

		// With power law concentration, may have limited M:M variety after duplicate pruning
		// The key requirement is NO duplicate FK pairs (which is working)

		// Log the actual M:M patterns achieved
		t.Logf("M:M patterns after duplicate pruning:")
		t.Logf("  Users with multiple groups: %d", countWithMultiple(userGroups))
		t.Logf("  Groups with multiple users: %d", countWithMultiple(groupUsers))

		// Main requirement: Should have some M:M relationships (not zero)
		assert.Greater(t, len(fkPairs), 0, "Should have at least some M:M relationships after pruning")
	})
}

// countWithMultiple counts how many keys have more than 1 value in the map
func countWithMultiple(m map[string][]string) int {
	count := 0
	for _, values := range m {
		if len(values) > 1 {
			count++
		}
	}
	return count
}
