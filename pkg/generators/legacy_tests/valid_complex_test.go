package generators

import (
	"fmt"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidComplexRelationships(t *testing.T) {
	// Create a generator with a small data volume for testing
	dataVolume := 10
	g := NewCSVGenerator("test_output", dataVolume, false)

	// Define entities for a DAG (directed acyclic graph) relationship:
	// - Organization: top level entity
	// - Department: belongs to an organization
	// - Project: belongs to a department
	// - Task: belongs to a project, assigned to a user
	// - User: belongs to a department
	entities := map[string]models.Entity{
		"org1": {
			DisplayName: "Organization",
			ExternalId:  "Organization",
			Description: "Organization entity",
			Attributes: []models.Attribute{
				{
					Name:       "id",
					ExternalId: "id",
					Type:       "String",
					UniqueId:   true,
				},
				{
					Name:       "name",
					ExternalId: "name",
					Type:       "String",
				},
			},
		},
		"dept1": {
			DisplayName: "Department",
			ExternalId:  "Department",
			Description: "Department entity",
			Attributes: []models.Attribute{
				{
					Name:       "id",
					ExternalId: "id",
					Type:       "String",
					UniqueId:   true,
				},
				{
					Name:       "name",
					ExternalId: "name",
					Type:       "String",
				},
				{
					Name:       "organizationId",
					ExternalId: "organizationId",
					Type:       "String",
				},
			},
		},
		"user1": {
			DisplayName: "User",
			ExternalId:  "User",
			Description: "User entity",
			Attributes: []models.Attribute{
				{
					Name:       "id",
					ExternalId: "id",
					Type:       "String",
					UniqueId:   true,
				},
				{
					Name:       "name",
					ExternalId: "name",
					Type:       "String",
				},
				{
					Name:       "departmentId",
					ExternalId: "departmentId",
					Type:       "String",
				},
			},
		},
		"project1": {
			DisplayName: "Project",
			ExternalId:  "Project",
			Description: "Project entity",
			Attributes: []models.Attribute{
				{
					Name:       "id",
					ExternalId: "id",
					Type:       "String",
					UniqueId:   true,
				},
				{
					Name:       "name",
					ExternalId: "name",
					Type:       "String",
				},
				{
					Name:       "departmentId",
					ExternalId: "departmentId",
					Type:       "String",
				},
			},
		},
		"task1": {
			DisplayName: "Task",
			ExternalId:  "Task",
			Description: "Task entity",
			Attributes: []models.Attribute{
				{
					Name:       "id",
					ExternalId: "id",
					Type:       "String",
					UniqueId:   true,
				},
				{
					Name:       "title",
					ExternalId: "title",
					Type:       "String",
				},
				{
					Name:       "projectId",
					ExternalId: "projectId",
					Type:       "String",
				},
				{
					Name:       "assignedToId",
					ExternalId: "assignedToId",
					Type:       "String",
				},
				{
					Name:       "createdById",
					ExternalId: "createdById",
					Type:       "String",
				},
			},
		},
	}

	// Define relationships
	relationships := map[string]models.Relationship{
		"department_organization": {
			DisplayName:   "department_organization",
			Name:          "department_organization",
			FromAttribute: "Department.organizationId",
			ToAttribute:   "Organization.id",
		},
		"user_department": {
			DisplayName:   "user_department",
			Name:          "user_department",
			FromAttribute: "User.departmentId",
			ToAttribute:   "Department.id",
		},
		"project_department": {
			DisplayName:   "project_department",
			Name:          "project_department",
			FromAttribute: "Project.departmentId",
			ToAttribute:   "Department.id",
		},
		"task_project": {
			DisplayName:   "task_project",
			Name:          "task_project",
			FromAttribute: "Task.projectId",
			ToAttribute:   "Project.id",
		},
		"task_assigned_to": {
			DisplayName:   "task_assigned_to",
			Name:          "task_assigned_to",
			FromAttribute: "Task.assignedToId",
			ToAttribute:   "User.id",
		},
		"task_created_by": {
			DisplayName:   "task_created_by",
			Name:          "task_created_by",
			FromAttribute: "Task.createdById",
			ToAttribute:   "User.id",
		},
	}

	// Set up the generator
	var err error
	err = g.Setup(entities, relationships)
	require.NoError(t, err, "Failed to setup generator")

	// Generate data
	err = g.GenerateData()
	require.NoError(t, err, "Failed to generate data")

	// Verify dependency order
	fmt.Println("Generation was successful. Verifying dependencies...")

	// Verify entities are generated in dependency order
	// Organization should be first, Task should be last
	// These relationships would have been established in the topological sort

	// Verify references are valid
	orgData := g.EntityData["org1"]
	deptData := g.EntityData["dept1"]
	userData := g.EntityData["user1"]
	projectData := g.EntityData["project1"]
	taskData := g.EntityData["task1"]

	// Find indices of relevant columns
	orgIdIndex := findColumnIndex(orgData.Headers, "id")
	deptOrgIdIndex := findColumnIndex(deptData.Headers, "organizationId")
	deptIdIndex := findColumnIndex(deptData.Headers, "id")
	userDeptIdIndex := findColumnIndex(userData.Headers, "departmentId")
	userId := findColumnIndex(userData.Headers, "id")
	projectDeptIdIndex := findColumnIndex(projectData.Headers, "departmentId")
	projectIdIndex := findColumnIndex(projectData.Headers, "id")
	taskProjectIdIndex := findColumnIndex(taskData.Headers, "projectId")
	taskAssignedToIdIndex := findColumnIndex(taskData.Headers, "assignedToId")
	taskCreatedByIdIndex := findColumnIndex(taskData.Headers, "createdById")

	// Collect valid IDs for each entity
	validOrgIds := extractValues(orgData.Rows, orgIdIndex)
	validDeptIds := extractValues(deptData.Rows, deptIdIndex)
	validUserIds := extractValues(userData.Rows, userId)
	validProjectIds := extractValues(projectData.Rows, projectIdIndex)

	// Check Department->Organization relationships
	validateReferences(t, "Department", "organizationId", deptData.Rows, deptOrgIdIndex, validOrgIds)

	// Check User->Department relationships
	validateReferences(t, "User", "departmentId", userData.Rows, userDeptIdIndex, validDeptIds)

	// Check Project->Department relationships
	validateReferences(t, "Project", "departmentId", projectData.Rows, projectDeptIdIndex, validDeptIds)

	// Check Task->Project relationships
	validateReferences(t, "Task", "projectId", taskData.Rows, taskProjectIdIndex, validProjectIds)

	// Check Task->User (assignedTo) relationships
	validateReferences(t, "Task", "assignedToId", taskData.Rows, taskAssignedToIdIndex, validUserIds)

	// Check Task->User (createdBy) relationships
	validateReferences(t, "Task", "createdById", taskData.Rows, taskCreatedByIdIndex, validUserIds)

	// Verify uniqueness of IDs across entities
	validateUniqueValues(t, "Organization", "id", orgData.Rows, orgIdIndex)
	validateUniqueValues(t, "Department", "id", deptData.Rows, deptIdIndex)
	validateUniqueValues(t, "User", "id", userData.Rows, userId)
	validateUniqueValues(t, "Project", "id", projectData.Rows, projectIdIndex)
	validateUniqueValues(t, "Task", "id", taskData.Rows, findColumnIndex(taskData.Headers, "id"))
}

// Helper function to extract values from a specific column
func extractValues(rows [][]string, colIndex int) map[string]bool {
	values := make(map[string]bool)
	for _, row := range rows {
		values[row[colIndex]] = true
	}
	return values
}

// Helper function to validate that references are valid
func validateReferences(t *testing.T, entityName, fieldName string, rows [][]string, colIndex int, validValues map[string]bool) {
	invalid := 0
	for i, row := range rows {
		if !validValues[row[colIndex]] {
			invalid++
			fmt.Printf("Warning: %s %d has invalid %s: %s\n", entityName, i, fieldName, row[colIndex])
		}
	}
	
	assert.Equal(t, 0, invalid, "%d %s entries have invalid %s values", invalid, entityName, fieldName)
	if invalid == 0 {
		fmt.Printf("✓ All %s.%s references are valid\n", entityName, fieldName)
	}
}

// Helper function to validate that values are unique
func validateUniqueValues(t *testing.T, entityName, fieldName string, rows [][]string, colIndex int) {
	values := make(map[string]int)
	duplicates := 0
	
	for i, row := range rows {
		value := row[colIndex]
		if prevIndex, exists := values[value]; exists {
			duplicates++
			fmt.Printf("Warning: Duplicate %s.%s value '%s' at rows %d and %d\n",
				entityName, fieldName, value, prevIndex, i)
		} else {
			values[value] = i
		}
	}
	
	assert.Equal(t, 0, duplicates, "Found %d duplicate %s.%s values", duplicates, entityName, fieldName)
	if duplicates == 0 {
		fmt.Printf("✓ All %s.%s values are unique\n", entityName, fieldName)
	}
}