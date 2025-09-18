package diagrams

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/parser"
)

// TestIsGraphvizAvailable tests both success and failure paths of IsGraphvizAvailable
func TestIsGraphvizAvailable(t *testing.T) {
	// Store the original exec.Command function
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	// Success case - mock a successful command execution
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo", "")
	}

	if !IsGraphvizAvailable() {
		t.Error("Expected IsGraphvizAvailable to return true for mocked success, got false")
	}

	// Failure case - mock a failed command execution
	execCommand = func(name string, args ...string) *exec.Cmd {
		cmd := exec.Command("false")
		return cmd
	}

	if IsGraphvizAvailable() {
		t.Error("Expected IsGraphvizAvailable to return false for mocked failure, got true")
	}
}

// We'll use the execCommand variable from er_generator.go

// TestGenerateDOTOnly tests the generation of just a DOT file (no SVG)
func TestGenerateDOTOnly(t *testing.T) {
	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "er-diagram-test-dot-only-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }() // Clean up after the test

	// Force DOT-only output by using .dot extension
	testOutputPath := filepath.Join(tempDir, "test_diagram.dot")

	// Create a minimal mock SORDefinition
	mockDefinition := &parser.SORDefinition{
		DisplayName: "Test SOR DOT Only",
		Description: "Test System of Record for DOT-only generation",
		Entities: map[string]parser.Entity{
			"user": {
				DisplayName: "User",
				ExternalId:  "Test/User",
				Attributes: []parser.Attribute{
					{
						Name:           "id",
						ExternalId:     "id",
						Type:           "string",
						UniqueId:       true,
						AttributeAlias: "userID",
					},
				},
			},
		},
	}

	// Temporarily modify IsGraphvizAvailable to return false
	originalFunc := IsGraphvizAvailable
	defer func() { IsGraphvizAvailable = originalFunc }()
	IsGraphvizAvailable = func() bool { return false }

	// Call the function being tested
	err = GenerateERDiagram(mockDefinition, testOutputPath)
	if err != nil {
		t.Fatalf("GenerateERDiagram failed: %v", err)
	}

	// Verify that the DOT file was created
	fileInfo, err := os.Stat(testOutputPath)
	if err != nil {
		t.Errorf("Failed to stat output file: %v", err)
	}

	// Verify that the file has content
	if fileInfo.Size() == 0 {
		t.Error("Generated DOT file has zero size")
	}
}

// TestGenerateWithInvalidTempFile tests error handling when temporary file creation fails
func TestGenerateWithInvalidTempFile(t *testing.T) {
	// Create a minimal mock SORDefinition
	mockDefinition := &parser.SORDefinition{
		DisplayName: "Test SOR",
		Entities:    map[string]parser.Entity{},
	}

	// Create a generator
	generator := NewERDiagramGenerator(mockDefinition)

	// Mock os.CreateTemp to return an error
	originalCreateTemp := createTemp
	defer func() { createTemp = originalCreateTemp }()
	createTemp = func(dir, pattern string) (*os.File, error) {
		return nil, os.ErrPermission
	}

	// Attempt to generate diagram - should fail
	err := generator.Generate("/tmp/invalid_diagram.svg")
	if err == nil {
		t.Error("Expected error when temporary file creation fails, got nil")
	}
}

// We'll use the createTemp variable from er_generator.go

// TestExtractEntitiesWithNamespaces tests entity extraction with various namespace patterns
func TestExtractEntitiesWithNamespaces(t *testing.T) {
	// Create a SORDefinition with entities having different namespace patterns
	mockDefinition := &parser.SORDefinition{
		DisplayName: "Test SOR",
		Entities: map[string]parser.Entity{
			"entity1": {
				DisplayName: "", // No display name, should use ExternalId
				ExternalId:  "Namespace/Entity1",
			},
			"entity2": {
				DisplayName: "Entity Two", // Has display name, should use it
				ExternalId:  "Namespace/Entity2",
			},
			"entity3": {
				DisplayName: "", // No display name, no namespace
				ExternalId:  "EntityWithoutNamespace",
			},
			"entity4": {
				DisplayName: "", // No display name, complex namespace
				ExternalId:  "Multi/Level/Namespace/Entity4",
			},
		},
	}

	// Create a generator and extract entities
	generator := NewERDiagramGenerator(mockDefinition)
	generator.extractEntities()

	// Verify entity count
	if len(generator.Entities) != 4 {
		t.Errorf("Expected 4 entities, got %d", len(generator.Entities))
	}

	// Check specific entity names
	testCases := []struct {
		entityID      string
		expectedName  string
		expectedExtID string
	}{
		{"entity1", "Entity1", "Namespace/Entity1"},                     // Should use last part of ExternalId
		{"entity2", "Entity Two", "Namespace/Entity2"},                  // Should use DisplayName
		{"entity3", "EntityWithoutNamespace", "EntityWithoutNamespace"}, // No namespace
		{"entity4", "Entity4", "Multi/Level/Namespace/Entity4"},         // Complex namespace
	}

	for _, tc := range testCases {
		entity, exists := generator.Entities[tc.entityID]
		if !exists {
			t.Errorf("Expected entity '%s' to exist", tc.entityID)
			continue
		}
		if entity.Name != tc.expectedName {
			t.Errorf("Entity '%s' name - expected '%s', got '%s'", tc.entityID, tc.expectedName, entity.Name)
		}
		if entity.ExternalID != tc.expectedExtID {
			t.Errorf("Entity '%s' external ID - expected '%s', got '%s'", tc.entityID, tc.expectedExtID, entity.ExternalID)
		}
	}
}

// TestExtractRelationshipsComprehensive tests all relationship extraction paths
func TestExtractRelationshipsComprehensive(t *testing.T) {
	// Create a SORDefinition with various relationship patterns
	mockDefinition := &parser.SORDefinition{
		DisplayName: "Test SOR",
		Entities: map[string]parser.Entity{
			"user": {
				DisplayName: "User",
				ExternalId:  "Test/User",
				Attributes: []parser.Attribute{
					{
						Name:           "id",
						ExternalId:     "id",
						Type:           "string",
						UniqueId:       true,
						AttributeAlias: "userID",
					},
					{
						Name:           "role_id",
						ExternalId:     "role_id",
						Type:           "string",
						AttributeAlias: "userRoleId",
					},
					{
						Name:           "self_ref",
						ExternalId:     "self_ref",
						Type:           "string",
						AttributeAlias: "userSelfRef",
					},
				},
			},
			"role": {
				DisplayName: "Role",
				ExternalId:  "Test/Role",
				Attributes: []parser.Attribute{
					{
						Name:           "id",
						ExternalId:     "id",
						Type:           "string",
						UniqueId:       true,
						AttributeAlias: "roleID",
					},
					{
						Name:           "perm_ids",
						ExternalId:     "perm_ids",
						Type:           "string",
						List:           true,
						AttributeAlias: "rolePermIds",
					},
				},
			},
			"permission": {
				DisplayName: "Permission",
				ExternalId:  "Test/Permission",
				Attributes: []parser.Attribute{
					{
						Name:           "id",
						ExternalId:     "id",
						Type:           "string",
						UniqueId:       true,
						AttributeAlias: "permissionID",
					},
				},
			},
		},
		Relationships: map[string]parser.Relationship{
			"rel1": {
				// Regular direct relationship with DisplayName
				DisplayName:   "User to Role",
				Name:          "user_to_role",
				FromAttribute: "userRoleId",
				ToAttribute:   "roleID",
			},
			"rel2": {
				// Direct relationship without DisplayName (should use Name)
				Name:          "role_to_permission",
				FromAttribute: "rolePermIds",
				ToAttribute:   "permissionID",
			},
			"rel3": {
				// Self-referential relationship (should be skipped)
				DisplayName:   "User Self Ref",
				Name:          "user_self_ref",
				FromAttribute: "userSelfRef",
				ToAttribute:   "userID",
			},
			"rel4": {
				// Missing from attribute (should be skipped)
				DisplayName:   "Invalid Relationship",
				Name:          "invalid_rel",
				FromAttribute: "",
				ToAttribute:   "roleID",
			},
			"rel5": {
				// Non-existent from attribute alias (should be skipped)
				DisplayName:   "Invalid Relationship 2",
				Name:          "invalid_rel2",
				FromAttribute: "nonexistent",
				ToAttribute:   "roleID",
			},
			"rel6": {
				// Simple path-based relationship
				DisplayName: "Path Based Relationship",
				Name:        "path_rel",
				Path: []parser.RelationshipPath{
					{Relationship: "rel1", Direction: "outbound"},
					{Relationship: "rel2", Direction: "outbound"},
				},
			},
			"rel7": {
				// Path with non-existent relationship (should be skipped)
				DisplayName: "Invalid Path",
				Name:        "invalid_path",
				Path: []parser.RelationshipPath{
					{Relationship: "nonexistent", Direction: "outbound"},
				},
			},
			"rel8": {
				// Self-referential path (should be skipped)
				DisplayName: "Self Path",
				Name:        "self_path",
				Path: []parser.RelationshipPath{
					{Relationship: "rel3", Direction: "outbound"},
				},
			},
		},
	}

	// Create a generator and extract relationships
	generator := NewERDiagramGenerator(mockDefinition)
	generator.extractEntities()
	generator.extractRelationships()

	// Debug output
	t.Logf("Found %d relationships:", len(generator.Relationships))
	for i, rel := range generator.Relationships {
		t.Logf("  Rel #%d: %s -> %s (pathBased: %v, displayName: %s)",
			i, rel.FromEntity, rel.ToEntity, rel.PathBased, rel.DisplayName)
	}

	// We expect the following relationships to be extracted:
	// 1. user -> role (direct, rel1)
	// 2. role -> permission (direct, rel2)
	// 3. user -> permission (path-based, rel6)
	// Relationships rel3, rel4, rel5, rel7, rel8 should be skipped

	// Check expected relationship count
	expectedCount := 3 // 2 direct + 1 path-based
	if len(generator.Relationships) != expectedCount {
		t.Errorf("Expected %d relationships, got %d", expectedCount, len(generator.Relationships))
	}

	// Check for specific relationships
	foundRel1 := false
	foundRel2 := false
	foundRel6 := false

	for _, rel := range generator.Relationships {
		if rel.FromEntity == "user" && rel.ToEntity == "role" && !rel.PathBased {
			foundRel1 = true
			if rel.DisplayName != "User to Role" {
				t.Errorf("Expected rel1 display name to be 'User to Role', got '%s'", rel.DisplayName)
			}
		} else if rel.FromEntity == "role" && rel.ToEntity == "permission" && !rel.PathBased {
			foundRel2 = true
			if rel.DisplayName != "role_to_permission" {
				t.Errorf("Expected rel2 display name to be 'role_to_permission', got '%s'", rel.DisplayName)
			}
		} else if rel.PathBased {
			foundRel6 = true
			if rel.DisplayName != "Path Based Relationship" {
				t.Errorf("Expected rel6 display name to be 'Path Based Relationship', got '%s'", rel.DisplayName)
			}
		}
	}

	if !foundRel1 {
		t.Error("Expected relationship 'user -> role' not found")
	}
	if !foundRel2 {
		t.Error("Expected relationship 'role -> permission' not found")
	}
	if !foundRel6 {
		t.Error("Expected path-based relationship not found")
	}
}

// TestGenerateWithEmptyEntities tests generating a diagram with empty entities
func TestGenerateWithEmptyEntities(t *testing.T) {
	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "er-diagram-empty-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }() // Clean up after the test

	// Create a SORDefinition with empty entities
	mockDefinition := &parser.SORDefinition{
		DisplayName:   "Empty SOR",
		Description:   "SOR with no entities",
		Entities:      map[string]parser.Entity{},
		Relationships: map[string]parser.Relationship{},
	}

	// Create output path
	testOutputPath := filepath.Join(tempDir, "empty_diagram.dot")

	// Generate the diagram
	err = GenerateERDiagram(mockDefinition, testOutputPath)
	if err != nil {
		t.Fatalf("GenerateERDiagram failed with empty entities: %v", err)
	}

	// Verify that the DOT file was created
	fileInfo, err := os.Stat(testOutputPath)
	if err != nil {
		t.Errorf("Failed to stat output file: %v", err)
	}

	// Verify that the file has content
	if fileInfo.Size() == 0 {
		t.Error("Generated DOT file has zero size")
	}
}

// TestGenerateGraphvizExecError tests error handling when graphviz exec fails
func TestGenerateGraphvizExecError(t *testing.T) {
	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "er-diagram-exec-error-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }() // Clean up after the test

	// Create a simple SORDefinition
	mockDefinition := &parser.SORDefinition{
		DisplayName: "Test SOR",
		Entities: map[string]parser.Entity{
			"user": {
				DisplayName: "User",
				ExternalId:  "Test/User",
				Attributes: []parser.Attribute{
					{
						Name:           "id",
						ExternalId:     "id",
						AttributeAlias: "userID",
					},
				},
			},
		},
	}

	// Set output path with SVG extension to trigger graphviz
	testOutputPath := filepath.Join(tempDir, "test_diagram.svg")

	// Force IsGraphvizAvailable to return true
	originalIsGraphvizAvailable := IsGraphvizAvailable
	defer func() { IsGraphvizAvailable = originalIsGraphvizAvailable }()
	IsGraphvizAvailable = func() bool { return true }

	// Mock exec.Command to force an error when executing dot
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()
	execCommand = func(name string, args ...string) *exec.Cmd {
		// Only mock 'dot' command, leave 'which' working normally
		if name == "dot" {
			return exec.Command("nonexistent-command")
		}
		return originalExecCommand(name, args...)
	}

	// Call the function
	err = GenerateERDiagram(mockDefinition, testOutputPath)
	if err == nil {
		t.Error("Expected error when graphviz execution fails, got nil")
	}
}
