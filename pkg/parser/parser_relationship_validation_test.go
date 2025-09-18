package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRelationshipValidation tests the validateRelationships function
func TestRelationshipValidation(t *testing.T) {
	tests := []struct {
		name           string
		entities       map[string]Entity
		relationships  map[string]Relationship
		shouldValidate bool
		errorMessages  []string
	}{
		{
			name: "valid direct relationship",
			entities: map[string]Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []Attribute{
						{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "User.id"},
						{Name: "name", ExternalId: "name", AttributeAlias: "User.name"},
					},
				},
				"assignment": {
					DisplayName: "Assignment",
					ExternalId:  "Assignment",
					Attributes: []Attribute{
						{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "Assignment.id"},
						{Name: "userId", ExternalId: "userId", AttributeAlias: "Assignment.userId"},
					},
				},
			},
			relationships: map[string]Relationship{
				"user_assignment": {
					DisplayName:   "User Assignment",
					FromAttribute: "Assignment.userId",
					ToAttribute:   "User.id",
				},
			},
			shouldValidate: true,
			errorMessages:  []string{},
		},
		{
			name: "missing attribute",
			entities: map[string]Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []Attribute{
						{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "User.id"},
					},
				},
				"assignment": {
					DisplayName: "Assignment",
					ExternalId:  "Assignment",
					Attributes: []Attribute{
						{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "Assignment.id"},
					},
				},
			},
			relationships: map[string]Relationship{
				"user_assignment": {
					DisplayName:   "User Assignment",
					FromAttribute: "Assignment.userId", // This doesn't exist
					ToAttribute:   "User.id",
				},
			},
			shouldValidate: false,
			errorMessages:  []string{"fromAttribute 'Assignment.userId' does not match any entity attribute"},
		},
		{
			name: "bidirectional relationship (potential cycle)",
			entities: map[string]Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []Attribute{
						{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "User.id"},
						{Name: "groupId", ExternalId: "groupId", AttributeAlias: "User.groupId"},
					},
				},
				"group": {
					DisplayName: "Group",
					ExternalId:  "Group",
					Attributes: []Attribute{
						{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "Group.id"},
						{Name: "ownerId", ExternalId: "ownerId", AttributeAlias: "Group.ownerId"},
					},
				},
			},
			relationships: map[string]Relationship{
				"user_group": {
					DisplayName:   "User Group",
					FromAttribute: "User.groupId",
					ToAttribute:   "Group.id",
				},
				"group_owner": {
					DisplayName:   "Group Owner",
					FromAttribute: "Group.ownerId",
					ToAttribute:   "User.id",
				},
			},
			shouldValidate: true,
			errorMessages:  []string{}, // Bidirectional relationships are now allowed
		},
		{
			name: "self-referential relationship with uniqueId attributes",
			entities: map[string]Entity{
				"employee": {
					DisplayName: "Employee",
					ExternalId:  "Employee",
					Attributes: []Attribute{
						{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "Employee.id"},
						{Name: "managerId", ExternalId: "managerId", UniqueId: true, AttributeAlias: "Employee.managerId"},
					},
				},
			},
			relationships: map[string]Relationship{
				"employee_manager": {
					DisplayName:   "Employee Manager",
					FromAttribute: "Employee.managerId",
					ToAttribute:   "Employee.id",
				},
			},
			shouldValidate: false,
			errorMessages:  []string{"potential self-referential issue between uniqueId attributes"},
		},
		{
			name: "non-uniqueId relationship attributes",
			entities: map[string]Entity{
				"product": {
					DisplayName: "Product",
					ExternalId:  "Product",
					Attributes: []Attribute{
						{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "Product.id"},
						{Name: "name", ExternalId: "name", AttributeAlias: "Product.name"},
					},
				},
				"comment": {
					DisplayName: "Comment",
					ExternalId:  "Comment",
					Attributes: []Attribute{
						{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "Comment.id"},
						{Name: "text", ExternalId: "text", AttributeAlias: "Comment.text"},                      // Not a uniqueId
						{Name: "productName", ExternalId: "productName", AttributeAlias: "Comment.productName"}, // Not a uniqueId
					},
				},
			},
			relationships: map[string]Relationship{
				"product_comment": {
					DisplayName:   "Product Comment",
					FromAttribute: "Comment.productName", // Not a uniqueId
					ToAttribute:   "Product.name",        // Not a uniqueId
				},
			},
			shouldValidate: true,
			errorMessages:  []string{}, // Non-uniqueId relationships now generate warnings, not errors
		},
		{
			name: "invalid path-based relationship",
			entities: map[string]Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []Attribute{
						{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "User.id"},
					},
				},
			},
			relationships: map[string]Relationship{
				"user_to_role": {
					DisplayName: "User To Role",
					Path: []RelationshipPath{
						{Relationship: "non_existent", Direction: "to"},
					},
				},
			},
			shouldValidate: false,
			errorMessages:  []string{"references non-existent relationship non_existent"},
		},
		{
			name: "nested path-based relationship",
			entities: map[string]Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []Attribute{
						{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "User.id"},
					},
				},
				"group": {
					DisplayName: "Group",
					ExternalId:  "Group",
					Attributes: []Attribute{
						{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "Group.id"},
					},
				},
			},
			relationships: map[string]Relationship{
				"user_to_group": {
					DisplayName:   "User To Group",
					FromAttribute: "User.id",
					ToAttribute:   "Group.id",
				},
				"complex_path": {
					DisplayName: "Complex Path",
					Path: []RelationshipPath{
						{Relationship: "nested_path", Direction: "to"},
					},
				},
				"nested_path": {
					DisplayName: "Nested Path",
					Path: []RelationshipPath{
						{Relationship: "user_to_group", Direction: "to"},
					},
				},
			},
			shouldValidate: false,
			errorMessages:  []string{"references path-based relationship nested_path (nested paths not supported)"},
		},
		{
			name: "path with any direction value is valid",
			entities: map[string]Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []Attribute{
						{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "User.id"},
					},
				},
				"group": {
					DisplayName: "Group",
					ExternalId:  "Group",
					Attributes: []Attribute{
						{Name: "id", ExternalId: "id", UniqueId: true, AttributeAlias: "Group.id"},
					},
				},
			},
			relationships: map[string]Relationship{
				"user_to_group": {
					DisplayName:   "User To Group",
					FromAttribute: "User.id",
					ToAttribute:   "Group.id",
				},
				"path_with_any_direction": {
					DisplayName: "Path With Any Direction Value",
					Path: []RelationshipPath{
						{Relationship: "user_to_group", Direction: "Forward"}, // Any direction value is accepted
					},
				},
			},
			shouldValidate: true,
			errorMessages:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				Definition: &SORDefinition{
					Entities:      tt.entities,
					Relationships: tt.relationships,
				},
			}

			err := p.validateRelationships()

			if tt.shouldValidate {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				errorMsg := err.Error()

				// Check if all expected error messages are present
				for _, expectedMsg := range tt.errorMessages {
					assert.True(t, strings.Contains(errorMsg, expectedMsg),
						"Expected error message to contain '%s', but got: %s", expectedMsg, errorMsg)
				}
			}
		})
	}
}
