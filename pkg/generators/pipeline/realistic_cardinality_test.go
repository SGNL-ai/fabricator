package pipeline

import (
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/SGNL-ai/fabricator/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRealisticCardinalityDistribution tests that auto-cardinality creates realistic data patterns
func TestRealisticCardinalityDistribution(t *testing.T) {
	t.Run("1:1 relationship should create unique pairings", func(t *testing.T) {
		// Employee ↔ User: Each employee gets exactly one unique user (true 1:1)
		def := &parser.SORDefinition{
			DisplayName: "1:1 Test",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true}, // PK
					},
				},
				"employee": {
					DisplayName: "Employee",
					ExternalId:  "Employee",
					Attributes: []parser.Attribute{
						{Name: "user_id", ExternalId: "user_id", Type: "String", UniqueId: true}, // Unique FK for true 1:1
					},
				},
			},
			Relationships: map[string]parser.Relationship{
				"employee_user": {
					Name:          "employee_user",
					FromAttribute: "employee.user_id", // Unique FK
					ToAttribute:   "user.id",          // Unique PK
				},
			},
		}

		graph, err := model.NewGraph(def, 5)
		require.NoError(t, err)

		// Generate data (no CSV output needed for this test)
		tempDir := t.TempDir()
		generator := NewDataGenerator(tempDir, 5, true) // Enable auto-cardinality
		err = generator.Generate(graph.(*model.Graph))
		require.NoError(t, err)

		// Verify 1:1 relationship integrity
		entities := graph.GetAllEntities()
		userEntity := entities["user"]
		employeeEntity := entities["employee"]

		userCSV := userEntity.ToCSV()
		employeeCSV := employeeEntity.ToCSV()

		// Find user_id column in employee CSV
		userIdCol := -1
		for i, header := range employeeCSV.Headers {
			if header == "user_id" {
				userIdCol = i
				break
			}
		}
		require.NotEqual(t, -1, userIdCol, "Should have user_id column")

		// Collect all employee user_id values
		employeeUserIds := make([]string, 0)
		for _, row := range employeeCSV.Rows {
			employeeUserIds = append(employeeUserIds, row[userIdCol])
		}

		// Collect all user id values
		userIds := make([]string, 0)
		for _, row := range userCSV.Rows {
			userIds = append(userIds, row[0]) // First column is ID
		}

		// CRITICAL: 1:1 relationship tests
		// 1. Every employee user_id should reference a valid user
		for i, employeeUserId := range employeeUserIds {
			assert.Contains(t, userIds, employeeUserId,
				"Employee row %d user_id '%s' should reference valid user", i, employeeUserId)
		}

		// 2. Each user should be referenced exactly once (1:1 constraint)
		usedUsers := make(map[string]int)
		for _, employeeUserId := range employeeUserIds {
			usedUsers[employeeUserId]++
		}

		for userId, useCount := range usedUsers {
			assert.Equal(t, 1, useCount,
				"User '%s' should be referenced exactly once in 1:1 relationship, was referenced %d times",
				userId, useCount)
		}

		// 3. No user should be unused (assuming equal entity counts)
		if len(employeeUserIds) == len(userIds) {
			assert.Equal(t, len(userIds), len(usedUsers),
				"All users should be used in 1:1 relationship")
		}
	})

	t.Run("N:1 relationship should create realistic clustering distribution", func(t *testing.T) {
		// Users → Department: Should create varied department sizes (realistic clustering)
		def := &parser.SORDefinition{
			DisplayName: "N:1 Test",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},            // PK
						{Name: "dept_id", ExternalId: "dept_id", Type: "String", UniqueId: false}, // FK non-unique for N:1
					},
				},
				"department": {
					DisplayName: "Department",
					ExternalId:  "Department",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true}, // PK
					},
				},
			},
			Relationships: map[string]parser.Relationship{
				"user_dept": {
					Name:          "user_department",
					FromAttribute: "user.dept_id",  // Non-unique FK
					ToAttribute:   "department.id", // Unique PK
				},
			},
		}

		graph, err := model.NewGraph(def, 25) // 25 users, 25 departments
		require.NoError(t, err)

		// Generate data with auto-cardinality enabled
		tempDir := t.TempDir()
		generator := NewDataGenerator(tempDir, 25, true) // Enable auto-cardinality
		err = generator.Generate(graph.(*model.Graph))
		require.NoError(t, err)

		// Analyze clustering distribution
		userEntity := graph.GetAllEntities()["user"]
		userCSV := userEntity.ToCSV()

		// Find dept_id column
		deptIdCol := -1
		for i, header := range userCSV.Headers {
			if header == "dept_id" {
				deptIdCol = i
				break
			}
		}
		require.NotEqual(t, -1, deptIdCol)

		// Count how many users are in each department
		deptCounts := make(map[string]int)
		for _, row := range userCSV.Rows {
			deptId := row[deptIdCol]
			deptCounts[deptId]++
		}

		// CRITICAL: Test realistic clustering distribution
		// With realistic N:1, we should see VARIED department sizes:
		// - Some departments with 1 user
		// - Some departments with multiple users
		// - NOT uniform distribution (that's unrealistic)

		// Count how many departments have each user count
		userCountDistribution := make(map[int]int) // map[usersInDept]numberOfDepts
		for _, userCount := range deptCounts {
			userCountDistribution[userCount]++
		}

		t.Logf("Department clustering distribution: %v", deptCounts)
		t.Logf("User count distribution: %v", userCountDistribution)

		// Should have variation in department sizes (not all departments same size)
		assert.Greater(t, len(userCountDistribution), 1,
			"Realistic N:1 should create varied department sizes, not uniform distribution")

		// At least some departments should have multiple users (clustering)
		hasMultiUserDepts := false
		for userCount, deptCount := range userCountDistribution {
			if userCount > 1 && deptCount > 0 {
				hasMultiUserDepts = true
				break
			}
		}
		assert.True(t, hasMultiUserDepts, "N:1 should create clustering - some departments should have multiple users")
	})

}
