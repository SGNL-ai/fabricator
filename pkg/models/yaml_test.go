package models

import (
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSORDefinitionDeserialization(t *testing.T) {
	// Test YAML parsing
	yamlData := `
displayName: Test SOR
description: Test system of record
hostname: localhost:8080
defaultSyncFrequency: DAILY
defaultSyncMinInterval: 1
defaultApiCallFrequency: SECONDLY
defaultApiCallMinInterval: 1
type: Test-1.0.0
adapterConfig: eyJrZXkiOiJ2YWx1ZSJ9
auth:
  - basic:
      username: testuser
entities:
  entity1:
    displayName: Test Entity One
    externalId: Test/EntityOne
    description: A test entity
    pagesOrderedById: false
    attributes:
      - name: id
        externalId: id
        description: unique id
        type: String
        indexed: true
        uniqueId: true
        attributeAlias: alias1
        list: false
      - name: name
        externalId: name
        description: entity name
        type: String
        indexed: false
        uniqueId: false
        attributeAlias: alias2
        list: false
    entityAlias: entity-alias-1
relationships:
  rel1:
    displayName: Test Relationship
    name: test_relationship
    fromAttribute: alias1
    toAttribute: alias2
  rel2:
    displayName: Path Relationship
    name: path_relationship
    path:
      - relationship: rel1
        direction: Forward
`

	var def SORDefinition
	err := yaml.Unmarshal([]byte(yamlData), &def)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	// Check that the fields were parsed correctly
	if def.DisplayName != "Test SOR" {
		t.Errorf("Expected DisplayName to be 'Test SOR', got %s", def.DisplayName)
	}

	if def.Description != "Test system of record" {
		t.Errorf("Expected Description to be 'Test system of record', got %s", def.Description)
	}

	if def.Hostname != "localhost:8080" {
		t.Errorf("Expected Hostname to be 'localhost:8080', got %s", def.Hostname)
	}

	if def.DefaultSyncFrequency != "DAILY" {
		t.Errorf("Expected DefaultSyncFrequency to be 'DAILY', got %s", def.DefaultSyncFrequency)
	}

	if def.DefaultSyncMinInterval != 1 {
		t.Errorf("Expected DefaultSyncMinInterval to be 1, got %d", def.DefaultSyncMinInterval)
	}

	if def.DefaultApiCallFrequency != "SECONDLY" {
		t.Errorf("Expected DefaultApiCallFrequency to be 'SECONDLY', got %s", def.DefaultApiCallFrequency)
	}

	if def.DefaultApiCallMinInterval != 1 {
		t.Errorf("Expected DefaultApiCallMinInterval to be 1, got %d", def.DefaultApiCallMinInterval)
	}

	if def.Type != "Test-1.0.0" {
		t.Errorf("Expected Type to be 'Test-1.0.0', got %s", def.Type)
	}

	if def.AdapterConfig != "eyJrZXkiOiJ2YWx1ZSJ9" {
		t.Errorf("Expected AdapterConfig to be 'eyJrZXkiOiJ2YWx1ZSJ9', got %s", def.AdapterConfig)
	}

	// Check auth
	if len(def.Auth) != 1 {
		t.Errorf("Expected 1 auth entry, got %d", len(def.Auth))
	} else {
		authEntry := def.Auth[0]
		if basicAuth, exists := authEntry["basic"]; exists {
			if basicAuth.Username != "testuser" {
				t.Errorf("Expected auth username to be 'testuser', got %s", basicAuth.Username)
			}
		} else {
			t.Error("Expected 'basic' auth entry, but it doesn't exist")
		}
	}

	// Check entities
	if len(def.Entities) != 1 {
		t.Errorf("Expected 1 entity, got %d", len(def.Entities))
	}

	entity, exists := def.Entities["entity1"]
	if !exists {
		t.Error("Expected entity 'entity1' to exist, but it doesn't")
	} else {
		if entity.DisplayName != "Test Entity One" {
			t.Errorf("Expected DisplayName to be 'Test Entity One', got %s", entity.DisplayName)
		}

		if entity.ExternalId != "Test/EntityOne" {
			t.Errorf("Expected ExternalId to be 'Test/EntityOne', got %s", entity.ExternalId)
		}

		if entity.Description != "A test entity" {
			t.Errorf("Expected Description to be 'A test entity', got %s", entity.Description)
		}

		if entity.PagesOrderedById != false {
			t.Errorf("Expected PagesOrderedById to be false, got %v", entity.PagesOrderedById)
		}

		if entity.EntityAlias != "entity-alias-1" {
			t.Errorf("Expected EntityAlias to be 'entity-alias-1', got %s", entity.EntityAlias)
		}

		// Check attributes
		if len(entity.Attributes) != 2 {
			t.Errorf("Expected 2 attributes, got %d", len(entity.Attributes))
		} else {
			// First attribute
			attr1 := entity.Attributes[0]
			if attr1.Name != "id" {
				t.Errorf("Expected Name to be 'id', got %s", attr1.Name)
			}

			if attr1.ExternalId != "id" {
				t.Errorf("Expected ExternalId to be 'id', got %s", attr1.ExternalId)
			}

			if attr1.Description != "unique id" {
				t.Errorf("Expected Description to be 'unique id', got %s", attr1.Description)
			}

			if attr1.Type != "String" {
				t.Errorf("Expected Type to be 'String', got %s", attr1.Type)
			}

			if !attr1.Indexed {
				t.Error("Expected Indexed to be true, got false")
			}

			if !attr1.UniqueId {
				t.Error("Expected UniqueId to be true, got false")
			}

			if attr1.AttributeAlias != "alias1" {
				t.Errorf("Expected AttributeAlias to be 'alias1', got %s", attr1.AttributeAlias)
			}

			if attr1.List != false {
				t.Errorf("Expected List to be false, got %v", attr1.List)
			}

			// Second attribute
			attr2 := entity.Attributes[1]
			if attr2.Name != "name" {
				t.Errorf("Expected Name to be 'name', got %s", attr2.Name)
			}
		}
	}

	// Check relationships
	if len(def.Relationships) != 2 {
		t.Errorf("Expected 2 relationships, got %d", len(def.Relationships))
	}

	rel1, exists := def.Relationships["rel1"]
	if !exists {
		t.Error("Expected relationship 'rel1' to exist, but it doesn't")
	} else {
		if rel1.DisplayName != "Test Relationship" {
			t.Errorf("Expected DisplayName to be 'Test Relationship', got %s", rel1.DisplayName)
		}

		if rel1.Name != "test_relationship" {
			t.Errorf("Expected Name to be 'test_relationship', got %s", rel1.Name)
		}

		if rel1.FromAttribute != "alias1" {
			t.Errorf("Expected FromAttribute to be 'alias1', got %s", rel1.FromAttribute)
		}

		if rel1.ToAttribute != "alias2" {
			t.Errorf("Expected ToAttribute to be 'alias2', got %s", rel1.ToAttribute)
		}

		if len(rel1.Path) != 0 {
			t.Errorf("Expected Path to be empty, got %d items", len(rel1.Path))
		}
	}

	rel2, exists := def.Relationships["rel2"]
	if !exists {
		t.Error("Expected relationship 'rel2' to exist, but it doesn't")
	} else {
		if rel2.DisplayName != "Path Relationship" {
			t.Errorf("Expected DisplayName to be 'Path Relationship', got %s", rel2.DisplayName)
		}

		if rel2.Name != "path_relationship" {
			t.Errorf("Expected Name to be 'path_relationship', got %s", rel2.Name)
		}

		if len(rel2.Path) != 1 {
			t.Errorf("Expected Path to have 1 item, got %d items", len(rel2.Path))
		} else {
			path := rel2.Path[0]
			if path.Relationship != "rel1" {
				t.Errorf("Expected Relationship to be 'rel1', got %s", path.Relationship)
			}

			if path.Direction != "Forward" {
				t.Errorf("Expected Direction to be 'Forward', got %s", path.Direction)
			}
		}
	}
}

func TestRelationshipLinkSerialization(t *testing.T) {
	// Create a relationship link
	link := RelationshipLink{
		FromEntityID:  "entity1",
		ToEntityID:    "entity2",
		FromAttribute: "attr1",
		ToAttribute:   "attr2",
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(link)
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	// Check that the field names match the JSON tags
	jsonStr := string(jsonData)
	if !strings.Contains(jsonStr, "fromEntityID") {
		t.Errorf("Expected JSON to contain 'fromEntityID', got %s", jsonStr)
	}

	if !strings.Contains(jsonStr, "toEntityID") {
		t.Errorf("Expected JSON to contain 'toEntityID', got %s", jsonStr)
	}

	if !strings.Contains(jsonStr, "fromAttribute") {
		t.Errorf("Expected JSON to contain 'fromAttribute', got %s", jsonStr)
	}

	if !strings.Contains(jsonStr, "toAttribute") {
		t.Errorf("Expected JSON to contain 'toAttribute', got %s", jsonStr)
	}

	// Deserialize from JSON
	var newLink RelationshipLink
	err = json.Unmarshal(jsonData, &newLink)
	if err != nil {
		t.Fatalf("Failed to unmarshal from JSON: %v", err)
	}

	// Check that the values match
	if newLink.FromEntityID != link.FromEntityID {
		t.Errorf("Expected FromEntityID to be '%s', got '%s'", link.FromEntityID, newLink.FromEntityID)
	}

	if newLink.ToEntityID != link.ToEntityID {
		t.Errorf("Expected ToEntityID to be '%s', got '%s'", link.ToEntityID, newLink.ToEntityID)
	}

	if newLink.FromAttribute != link.FromAttribute {
		t.Errorf("Expected FromAttribute to be '%s', got '%s'", link.FromAttribute, newLink.FromAttribute)
	}

	if newLink.ToAttribute != link.ToAttribute {
		t.Errorf("Expected ToAttribute to be '%s', got '%s'", link.ToAttribute, newLink.ToAttribute)
	}
}

func TestCSVData(t *testing.T) {
	// Create CSV data
	data := CSVData{
		ExternalId:  "Test/Entity",
		Headers:     []string{"id", "name"},
		Rows:        [][]string{{"1", "First"}, {"2", "Second"}},
		EntityName:  "TestEntity",
		Description: "A test entity",
	}

	// Check values
	if data.ExternalId != "Test/Entity" {
		t.Errorf("Expected ExternalId to be 'Test/Entity', got '%s'", data.ExternalId)
	}

	if len(data.Headers) != 2 {
		t.Errorf("Expected 2 headers, got %d", len(data.Headers))
	} else {
		if data.Headers[0] != "id" || data.Headers[1] != "name" {
			t.Errorf("Expected headers to be ['id', 'name'], got %v", data.Headers)
		}
	}

	if len(data.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(data.Rows))
	} else {
		if data.Rows[0][0] != "1" || data.Rows[0][1] != "First" {
			t.Errorf("Expected first row to be ['1', 'First'], got %v", data.Rows[0])
		}

		if data.Rows[1][0] != "2" || data.Rows[1][1] != "Second" {
			t.Errorf("Expected second row to be ['2', 'Second'], got %v", data.Rows[1])
		}
	}

	if data.EntityName != "TestEntity" {
		t.Errorf("Expected EntityName to be 'TestEntity', got '%s'", data.EntityName)
	}

	if data.Description != "A test entity" {
		t.Errorf("Expected Description to be 'A test entity', got '%s'", data.Description)
	}
}
