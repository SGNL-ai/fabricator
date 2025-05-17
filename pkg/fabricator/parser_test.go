package fabricator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
)

func TestParserValidation(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "parser-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a test YAML file with valid structure
	validYAML := `
displayName: ValidSOR
description: Valid System of Record
hostname: localhost:8080
defaultSyncFrequency: DAILY
defaultSyncMinInterval: 1
defaultApiCallFrequency: SECONDLY
defaultApiCallMinInterval: 1
type: Test-1.0.0
entities:
  test1:
    displayName: TestEntity
    externalId: Test/TestEntity
    description: A test entity
    pagesOrderedById: false
    attributes:
      - name: id
        externalId: id
        description: unique id
        type: String
        indexed: true
        uniqueId: true
        attributeAlias: test-id-alias
        list: false
      - name: otherAttr
        externalId: otherAttr
        description: test attribute
        type: String
        indexed: false
        uniqueId: false
        attributeAlias: test-other-alias
        list: false
    entityAlias: test1-alias
relationships:
  rel1:
    displayName: TestRelationship
    name: test_relationship
    fromAttribute: test-id-alias
    toAttribute: test-other-alias
`
	validFilePath := filepath.Join(tempDir, "valid.yaml")
	err = os.WriteFile(validFilePath, []byte(validYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write valid test file: %v", err)
	}

	// Create a test YAML file with invalid structure (no entities)
	invalidYAML := `
displayName: InvalidSOR
description: Invalid System of Record
hostname: localhost:8080
defaultSyncFrequency: DAILY
defaultSyncMinInterval: 1
defaultApiCallFrequency: SECONDLY
defaultApiCallMinInterval: 1
type: Test-1.0.0
`
	invalidFilePath := filepath.Join(tempDir, "invalid.yaml")
	err = os.WriteFile(invalidFilePath, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid test file: %v", err)
	}

	// Test parsing valid YAML
	t.Run("Parse valid YAML", func(t *testing.T) {
		parser := NewParser(validFilePath)
		err := parser.Parse()
		if err != nil {
			t.Errorf("Parse() failed on valid YAML: %v", err)
		}

		if parser.Definition == nil {
			t.Error("Parse() succeeded but Definition is nil")
		}

		if len(parser.Definition.Entities) != 1 {
			t.Errorf("Expected 1 entity, got %d", len(parser.Definition.Entities))
		}

		if len(parser.Definition.Relationships) != 1 {
			t.Errorf("Expected 1 relationship, got %d", len(parser.Definition.Relationships))
		}
	})

	// Test parsing invalid YAML (no entities)
	t.Run("Parse invalid YAML", func(t *testing.T) {
		parser := NewParser(invalidFilePath)
		err := parser.Parse()
		if err == nil {
			t.Error("Parse() succeeded on invalid YAML with no entities")
		}
	})

	// Test parsing non-existent file
	t.Run("Parse non-existent file", func(t *testing.T) {
		parser := NewParser(filepath.Join(tempDir, "nonexistent.yaml"))
		err := parser.Parse()
		if err == nil {
			t.Error("Parse() succeeded on non-existent file")
		}
	})
}

func TestGetCSVFilenames(t *testing.T) {
	// Create a parser with a mock definition
	parser := &Parser{
		Definition: &models.SORDefinition{
			Entities: map[string]models.Entity{
				"entity1": {
					DisplayName: "Entity One",
					ExternalId:  "Namespace/EntityOne",
				},
				"entity2": {
					DisplayName: "Entity Two",
					ExternalId:  "Namespace/EntityTwo",
				},
			},
		},
	}

	// Get CSV filenames
	filenames := parser.GetCSVFilenames()

	// Check that we have the expected filenames
	if len(filenames) != 2 {
		t.Errorf("Expected 2 filenames, got %d", len(filenames))
	}

	if filenames["Namespace/EntityOne"] != "EntityOne.csv" {
		t.Errorf("Expected EntityOne.csv, got %s", filenames["Namespace/EntityOne"])
	}

	if filenames["Namespace/EntityTwo"] != "EntityTwo.csv" {
		t.Errorf("Expected EntityTwo.csv, got %s", filenames["Namespace/EntityTwo"])
	}
}

// setupTestParser creates a parser with test data for entity relationship testing
func setupTestParser() *Parser {
	// Create entities with attributes and relationships
	entity1 := models.Entity{
		DisplayName: "User",
		ExternalId:  "Test/User",
		Description: "User entity",
		Attributes: []models.Attribute{
			{
				Name:           "id",
				ExternalId:     "id",
				Description:    "User ID",
				Type:           "String",
				Indexed:        true,
				UniqueId:       true,
				AttributeAlias: "user-id-alias",
			},
			{
				Name:           "name",
				ExternalId:     "name",
				Description:    "User name",
				Type:           "String",
				Indexed:        false,
				UniqueId:       false,
				AttributeAlias: "user-name-alias",
			},
			{
				Name:           "roleId",
				ExternalId:     "roleId",
				Description:    "Reference to role",
				Type:           "String",
				Indexed:        true,
				UniqueId:       false,
				AttributeAlias: "user-role-alias",
			},
		},
	}

	entity2 := models.Entity{
		DisplayName: "Role",
		ExternalId:  "Test/Role",
		Description: "Role entity",
		Attributes: []models.Attribute{
			{
				Name:           "id",
				ExternalId:     "id",
				Description:    "Role ID",
				Type:           "String",
				Indexed:        true,
				UniqueId:       true,
				AttributeAlias: "role-id-alias",
			},
			{
				Name:           "name",
				ExternalId:     "name",
				Description:    "Role name",
				Type:           "String",
				Indexed:        false,
				UniqueId:       false,
				AttributeAlias: "role-name-alias",
			},
		},
	}

	entity3 := models.Entity{
		DisplayName: "Permission",
		ExternalId:  "Test/Permission",
		Description: "Permission entity",
		Attributes: []models.Attribute{
			{
				Name:           "id",
				ExternalId:     "id",
				Description:    "Permission ID",
				Type:           "String",
				Indexed:        true,
				UniqueId:       true,
				AttributeAlias: "perm-id-alias",
			},
			{
				Name:           "name",
				ExternalId:     "name",
				Description:    "Permission name",
				Type:           "String",
				Indexed:        false,
				UniqueId:       false,
				AttributeAlias: "perm-name-alias",
			},
			{
				Name:           "roleId",
				ExternalId:     "roleId",
				Description:    "Reference to role",
				Type:           "String",
				Indexed:        true,
				UniqueId:       false,
				AttributeAlias: "perm-role-alias",
			},
		},
	}

	// Create relationships
	relationship1 := models.Relationship{
		DisplayName:   "User to Role",
		Name:          "user_to_role_rel",
		FromAttribute: "user-role-alias",
		ToAttribute:   "role-id-alias",
	}

	relationship2 := models.Relationship{
		DisplayName:   "Role to Permission",
		Name:          "role_to_permission_rel",
		FromAttribute: "role-id-alias",
		ToAttribute:   "perm-role-alias",
	}

	// Path-based relationship (User -> Role -> Permission)
	pathRelationship := models.Relationship{
		DisplayName: "User to Permission",
		Name:        "user_to_permission",
		Path: []models.RelationshipPath{
			{
				Relationship: "rel1", // Reference to the relationship ID, not the name
				Direction:    "Forward",
			},
			{
				Relationship: "rel2", // Reference to the relationship ID, not the name
				Direction:    "Forward",
			},
		},
	}

	// Create a parser with these entities and relationships
	parser := &Parser{
		Definition: &models.SORDefinition{
			DisplayName:   "Test SOR",
			Description:   "Test System of Record",
			Entities:      map[string]models.Entity{"user": entity1, "role": entity2, "permission": entity3},
			Relationships: map[string]models.Relationship{"rel1": relationship1, "rel2": relationship2, "rel3": pathRelationship},
		},
	}

	return parser
}

func TestGetEntityByExternalId(t *testing.T) {
	parser := setupTestParser()

	t.Run("Get existing entity", func(t *testing.T) {
		entity, id, err := parser.GetEntityByExternalId("Test/User")
		if err != nil {
			t.Errorf("GetEntityByExternalId() failed: %v", err)
		}
		if entity == nil {
			t.Fatal("Expected entity to be non-nil")
		}
		if entity.DisplayName != "User" {
			t.Errorf("Expected entity with DisplayName 'User', got '%s'", entity.DisplayName)
		}
		if id != "user" {
			t.Errorf("Expected id 'user', got '%s'", id)
		}
	})

	t.Run("Get non-existent entity", func(t *testing.T) {
		_, _, err := parser.GetEntityByExternalId("Test/NonExistent")
		if err == nil {
			t.Error("Expected error for non-existent entity")
		}
	})
}

func TestGetEntityById(t *testing.T) {
	parser := setupTestParser()

	t.Run("Get existing entity", func(t *testing.T) {
		entity, err := parser.GetEntityById("user")
		if err != nil {
			t.Errorf("GetEntityById() failed: %v", err)
		}
		if entity == nil {
			t.Fatal("Expected entity to be non-nil")
		}
		if entity.DisplayName != "User" {
			t.Errorf("Expected entity with DisplayName 'User', got '%s'", entity.DisplayName)
		}
	})

	t.Run("Get non-existent entity", func(t *testing.T) {
		_, err := parser.GetEntityById("nonexistent")
		if err == nil {
			t.Error("Expected error for non-existent entity")
		}
	})
}

func TestFindRelationshipsForEntity(t *testing.T) {
	parser := setupTestParser()

	t.Run("Find relationships for User entity", func(t *testing.T) {
		relationships := parser.FindRelationshipsForEntity("user")
		if len(relationships) != 1 {
			t.Errorf("Expected 1 relationship for 'user', got %d", len(relationships))
		}

		rel, exists := relationships["rel1"]
		if !exists {
			t.Error("Expected relationship 'rel1' to exist")
		}
		if rel.Name != "user_to_role_rel" {
			t.Errorf("Expected relationship name 'user_to_role_rel', got '%s'", rel.Name)
		}
	})

	t.Run("Find relationships for Role entity", func(t *testing.T) {
		relationships := parser.FindRelationshipsForEntity("role")
		if len(relationships) != 2 {
			t.Errorf("Expected 2 relationships for 'role', got %d", len(relationships))
		}
	})

	t.Run("Find relationships for non-existent entity", func(t *testing.T) {
		relationships := parser.FindRelationshipsForEntity("nonexistent")
		if relationships != nil {
			t.Errorf("Expected nil for non-existent entity, got %v", relationships)
		}
	})
}

func TestFindEntityRelationships(t *testing.T) {
	parser := setupTestParser()

	t.Run("Find relationships for User entity", func(t *testing.T) {
		relationships := parser.FindEntityRelationships("user")
		if len(relationships) != 1 {
			t.Errorf("Expected 1 relationship link for 'user', got %d", len(relationships))
		}

		if relationships[0].FromEntityID != "user" {
			t.Errorf("Expected FromEntityID 'user', got '%s'", relationships[0].FromEntityID)
		}
		if relationships[0].ToEntityID != "role" {
			t.Errorf("Expected ToEntityID 'role', got '%s'", relationships[0].ToEntityID)
		}
	})

	t.Run("Find relationships for Role entity", func(t *testing.T) {
		relationships := parser.FindEntityRelationships("role")
		if len(relationships) != 2 {
			t.Errorf("Expected 2 relationship links for 'role', got %d", len(relationships))
		}
	})

	t.Run("Find relationships for non-existent entity", func(t *testing.T) {
		relationships := parser.FindEntityRelationships("nonexistent")
		if len(relationships) != 0 {
			t.Errorf("Expected empty slice for non-existent entity, got %d items", len(relationships))
		}
	})
}

func TestGetUniqueIdAttributeFor(t *testing.T) {
	parser := setupTestParser()

	t.Run("Get unique ID attribute for User entity", func(t *testing.T) {
		attr, err := parser.GetUniqueIdAttributeFor("user")
		if err != nil {
			t.Errorf("GetUniqueIdAttributeFor() failed: %v", err)
		}
		if attr == nil {
			t.Fatal("Expected attribute to be non-nil")
		}
		if attr.Name != "id" {
			t.Errorf("Expected attribute with Name 'id', got '%s'", attr.Name)
		}
		if !attr.UniqueId {
			t.Error("Expected attribute to have UniqueId=true")
		}
	})

	t.Run("Get unique ID attribute for non-existent entity", func(t *testing.T) {
		_, err := parser.GetUniqueIdAttributeFor("nonexistent")
		if err == nil {
			t.Error("Expected error for non-existent entity")
		}
	})

	t.Run("Entity without unique ID attribute", func(t *testing.T) {
		// Create a test parser with an entity that has no uniqueId attribute
		parser := &Parser{
			Definition: &models.SORDefinition{
				DisplayName: "Test SOR",
				Description: "Test description",
				Entities: map[string]models.Entity{
					"entity1": {
						DisplayName: "Entity One",
						ExternalId:  "Test/EntityOne",
						Description: "Test entity",
						Attributes: []models.Attribute{
							{
								Name:       "name",
								ExternalId: "name",
								UniqueId:   false, // Not a unique ID
							},
							{
								Name:       "description",
								ExternalId: "description",
								UniqueId:   false, // Not a unique ID
							},
						},
					},
				},
			},
		}

		_, err := parser.GetUniqueIdAttributeFor("entity1")
		if err == nil {
			t.Error("Expected error for entity without uniqueId attribute")
		}
	})
}

func TestGetNamespacePrefix(t *testing.T) {
	parser := setupTestParser()

	prefix := parser.GetNamespacePrefix()
	if prefix != "Test" {
		t.Errorf("Expected namespace prefix 'Test', got '%s'", prefix)
	}

	// Test with empty entities
	emptyParser := &Parser{
		Definition: &models.SORDefinition{
			Entities: map[string]models.Entity{},
		},
	}

	emptyPrefix := emptyParser.GetNamespacePrefix()
	if emptyPrefix != "" {
		t.Errorf("Expected empty prefix for empty entities, got '%s'", emptyPrefix)
	}
}

func TestValidate(t *testing.T) {
	// Create a parser with valid definition
	t.Run("Valid definition", func(t *testing.T) {
		parser := setupTestParser()
		err := parser.validate()
		if err != nil {
			t.Errorf("validate() failed on valid definition: %v", err)
		}

		// Test validateRelationships directly
		err = parser.validateRelationships()
		if err != nil {
			t.Errorf("validateRelationships() failed on valid relationships: %v", err)
		}
	})

	// Test with missing required fields
	t.Run("Missing display name", func(t *testing.T) {
		parser := &Parser{
			Definition: &models.SORDefinition{
				// Missing DisplayName
				Description: "Test description",
				Entities:    map[string]models.Entity{},
			},
		}
		err := parser.validate()
		if err == nil {
			t.Error("validate() should fail when DisplayName is missing")
		}
	})

	t.Run("Missing description", func(t *testing.T) {
		parser := &Parser{
			Definition: &models.SORDefinition{
				DisplayName: "Test SOR",
				// Missing Description
				Entities: map[string]models.Entity{},
			},
		}
		err := parser.validate()
		if err == nil {
			t.Error("validate() should fail when Description is missing")
		}
	})

	t.Run("Empty entities", func(t *testing.T) {
		parser := &Parser{
			Definition: &models.SORDefinition{
				DisplayName: "Test SOR",
				Description: "Test description",
				Entities:    map[string]models.Entity{}, // Empty entities
			},
		}
		err := parser.validate()
		if err == nil {
			t.Error("validate() should fail when Entities is empty")
		}
	})

	// Test with invalid entity definitions
	t.Run("Entity missing external ID", func(t *testing.T) {
		parser := &Parser{
			Definition: &models.SORDefinition{
				DisplayName: "Test SOR",
				Description: "Test description",
				Entities: map[string]models.Entity{
					"entity1": {
						DisplayName: "Entity One",
						// Missing ExternalId
						Description: "Test entity",
						Attributes:  []models.Attribute{},
					},
				},
			},
		}
		err := parser.validate()
		if err == nil {
			t.Error("validate() should fail when entity is missing ExternalId")
		}
	})

	t.Run("Entity missing attributes", func(t *testing.T) {
		parser := &Parser{
			Definition: &models.SORDefinition{
				DisplayName: "Test SOR",
				Description: "Test description",
				Entities: map[string]models.Entity{
					"entity1": {
						DisplayName: "Entity One",
						ExternalId:  "Test/EntityOne",
						Description: "Test entity",
						Attributes:  []models.Attribute{}, // Empty attributes
					},
				},
			},
		}
		err := parser.validate()
		if err == nil {
			t.Error("validate() should fail when entity has no attributes")
		}
	})

	// Test with invalid relationship definitions
	t.Run("Relationship with invalid path", func(t *testing.T) {
		// Create a valid parser first
		parser := setupTestParser()

		// Add an invalid relationship - path-based relationship with missing path
		parser.Definition.Relationships["invalid_rel"] = models.Relationship{
			DisplayName: "Invalid Relationship",
			Name:        "invalid_rel",
			// Missing FromAttribute and ToAttribute
			// Also missing Path
		}

		err := parser.validate()
		if err == nil {
			t.Error("validate() should fail with invalid path-based relationship")
		}
	})

	t.Run("Direct relationship missing attributes", func(t *testing.T) {
		// Create a valid parser first
		parser := setupTestParser()

		// Add an invalid relationship - direct relationship with missing attributes
		parser.Definition.Relationships["invalid_rel"] = models.Relationship{
			DisplayName: "Invalid Relationship",
			Name:        "invalid_rel",
			// Missing FromAttribute and ToAttribute
		}

		err := parser.validate()
		if err == nil {
			t.Error("validate() should fail with invalid direct relationship")
		}
	})
}

// TestValidateRelationships tests specific scenarios for relationship validation
func TestValidateRelationships(t *testing.T) {
	t.Run("Non-existent attribute references", func(t *testing.T) {
		// Create a parser with a relationship that references non-existent attributes
		parser := setupTestParser()

		// Add an invalid relationship with non-existent attribute references
		parser.Definition.Relationships["bad_rel"] = models.Relationship{
			DisplayName:   "Bad Relationship",
			Name:          "bad_relationship",
			FromAttribute: "non-existent-from-alias",
			ToAttribute:   "non-existent-to-alias",
		}

		// Test validation should fail
		err := parser.validateRelationships()
		if err == nil {
			t.Error("validateRelationships() should fail with non-existent attribute references")
		}

		// Verify error contains the expected information
		errStr := err.Error()
		if !strings.Contains(errStr, "non-existent-from-alias") {
			t.Errorf("Error message should mention the missing from attribute")
		}
		if !strings.Contains(errStr, "non-existent-to-alias") {
			t.Errorf("Error message should mention the missing to attribute")
		}
	})

	t.Run("Empty relationships", func(t *testing.T) {
		// Test with empty relationships (should pass validation)
		parser := &Parser{
			Definition: &models.SORDefinition{
				DisplayName:   "Test SOR",
				Description:   "Test System of Record",
				Entities:      map[string]models.Entity{},
				Relationships: map[string]models.Relationship{},
			},
		}

		err := parser.validateRelationships()
		if err != nil {
			t.Errorf("validateRelationships() failed on empty relationships: %v", err)
		}
	})
}
