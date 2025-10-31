package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/SGNL-ai/fabricator/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDataGeneratorPipelineIntegration tests the complete 3-phase pipeline orchestration
// This should expose any gaps between individual component unit tests and full pipeline execution
func TestDataGeneratorPipelineIntegration(t *testing.T) {
	t.Run("should execute complete pipeline with relationship attributes", func(t *testing.T) {
		// Create a YAML definition with clear FK relationship
		def := &parser.SORDefinition{
			DisplayName: "Pipeline Test SOR",
			Description: "Test complete pipeline orchestration",
			Entities: map[string]parser.Entity{
				"user": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
						{Name: "name", ExternalId: "name", Type: "String"},
						{Name: "profile_id", ExternalId: "profile_id", Type: "String"}, // FK field
					},
				},
				"profile": {
					DisplayName: "Profile",
					ExternalId:  "Profile",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
						{Name: "bio", ExternalId: "bio", Type: "String"},
					},
				},
			},
			Relationships: map[string]parser.Relationship{
				"user_profile": {
					DisplayName:   "User Profile",
					Name:          "user_profile",
					FromAttribute: "User.profile_id", // Use DisplayName "User"
					ToAttribute:   "Profile.id",      // Use DisplayName "Profile"
				},
			},
		}

		// Create graph from definition (this should mark FK attributes as relationships)
		graphInterface, err := model.NewGraph(def, 100)
		require.NoError(t, err)
		graph, ok := graphInterface.(*model.Graph)
		require.True(t, ok)

		// Create temporary directory for CSV output
		tempDir, err := os.MkdirTemp("", "pipeline-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Execute the complete 3-phase pipeline
		generator := NewDataGenerator(tempDir, map[string]int{"User": 2, "Profile": 2}, false)
		err = generator.Generate(graph)
		require.NoError(t, err)

		// Verify files were created
		userCSVPath := filepath.Join(tempDir, "User.csv")
		profileCSVPath := filepath.Join(tempDir, "Profile.csv")

		userContent, err := os.ReadFile(userCSVPath)
		require.NoError(t, err)
		profileContent, err := os.ReadFile(profileCSVPath)
		require.NoError(t, err)

		t.Logf("User.csv content:\n%s", string(userContent))
		t.Logf("Profile.csv content:\n%s", string(profileContent))

		// Verify the graph state after pipeline execution
		entities := graph.GetAllEntities()
		userEntity := entities["User"]
		profileEntity := entities["Profile"]

		// Check that relationship attributes are properly marked
		userRelAttrs := userEntity.GetRelationshipAttributes()
		userNonRelAttrs := userEntity.GetNonRelationshipAttributes()

		t.Logf("User relationship attributes: %d", len(userRelAttrs))
		for i, attr := range userRelAttrs {
			t.Logf("  Rel attr %d: %s (isRelationship: %v)", i, attr.GetName(), attr.IsRelationship())
		}

		t.Logf("User non-relationship attributes: %d", len(userNonRelAttrs))
		for i, attr := range userNonRelAttrs {
			t.Logf("  Non-rel attr %d: %s (isRelationship: %v)", i, attr.GetName(), attr.IsRelationship())
		}

		// CRITICAL TEST: profile_id should be marked as a relationship attribute
		assert.Len(t, userRelAttrs, 1, "User should have 1 relationship attribute (profile_id)")
		if len(userRelAttrs) > 0 {
			assert.Equal(t, "profile_id", userRelAttrs[0].GetName(), "Relationship attribute should be profile_id")
			assert.True(t, userRelAttrs[0].IsRelationship(), "profile_id should be marked as relationship")
		}

		// CRITICAL TEST: Non-relationship attributes should NOT include profile_id
		assert.Len(t, userNonRelAttrs, 2, "User should have 2 non-relationship attributes (id, name)")
		nonRelNames := make([]string, len(userNonRelAttrs))
		for i, attr := range userNonRelAttrs {
			nonRelNames[i] = attr.GetName()
		}
		assert.Contains(t, nonRelNames, "id", "Non-relationship attributes should include id")
		assert.Contains(t, nonRelNames, "name", "Non-relationship attributes should include name")
		assert.NotContains(t, nonRelNames, "profile_id", "Non-relationship attributes should NOT include profile_id")

		// Check actual data in entities
		userCSVData := userEntity.ToCSV()
		profileCSVData := profileEntity.ToCSV()

		require.Len(t, userCSVData.Rows, 2, "Should have 2 user rows")
		require.Len(t, profileCSVData.Rows, 2, "Should have 2 profile rows")

		// Find column indices
		var userProfileIdCol = -1
		for i, header := range userCSVData.Headers {
			if header == "profile_id" {
				userProfileIdCol = i
				break
			}
		}
		require.NotEqual(t, -1, userProfileIdCol, "User CSV should have profile_id column")

		// Get profile IDs for validation
		profileIDs := make([]string, len(profileCSVData.Rows))
		for i, row := range profileCSVData.Rows {
			profileIDs[i] = row[0] // First column is ID
		}

		// CRITICAL TEST: profile_id values should reference valid profile IDs
		for i, userRow := range userCSVData.Rows {
			profileIdValue := userRow[userProfileIdCol]
			assert.Contains(t, profileIDs, profileIdValue,
				"User row %d: profile_id '%s' should reference valid profile ID from %v",
				i, profileIdValue, profileIDs)
		}
	})

	t.Run("should handle attributeAlias relationships", func(t *testing.T) {
		// Test pipeline with attributeAlias format like sample.yaml
		def := &parser.SORDefinition{
			DisplayName: "AttributeAlias Pipeline Test",
			Description: "Test pipeline with attributeAlias relationships",
			Entities: map[string]parser.Entity{
				"user-entity": {
					DisplayName: "User",
					ExternalId:  "User",
					Attributes: []parser.Attribute{
						{Name: "userId", ExternalId: "uuid", Type: "String", UniqueId: true, AttributeAlias: "user-pk-alias"},
						{Name: "name", ExternalId: "name", Type: "String", AttributeAlias: "user-name-alias"},
					},
				},
				"profile-entity": {
					DisplayName: "Profile",
					ExternalId:  "Profile",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true, AttributeAlias: "profile-pk-alias"},
						{Name: "userId", ExternalId: "uuid", Type: "String", AttributeAlias: "profile-user-alias"},
					},
				},
			},
			Relationships: map[string]parser.Relationship{
				"profile-to-user": {
					DisplayName:   "Profile to User",
					Name:          "profile_to_user",
					FromAttribute: "profile-user-alias",
					ToAttribute:   "user-pk-alias",
				},
			},
		}

		graphInterface, err := model.NewGraph(def, 100)
		require.NoError(t, err)
		graph, ok := graphInterface.(*model.Graph)
		require.True(t, ok)

		tempDir, err := os.MkdirTemp("", "pipeline-alias-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Execute pipeline directly
		generator := NewDataGenerator(tempDir, map[string]int{"User": 2, "Profile": 2}, false)
		err = generator.Generate(graph)
		require.NoError(t, err)

		// Check relationship attribute marking
		entities := graph.GetAllEntities()
		userEntity := entities["User"]

		userRelAttrs := userEntity.GetRelationshipAttributes()
		t.Logf("User relationship attributes with attributeAlias: %d", len(userRelAttrs))
		for i, attr := range userRelAttrs {
			t.Logf("  Attr %d: %s (isRelationship: %v)", i, attr.GetName(), attr.IsRelationship())
		}

		// CRITICAL TEST: With attributeAlias, is the FK attribute properly marked as relationship?
		assert.Len(t, userRelAttrs, 0, "Expected 0 relationship attributes - but should this be 1?")

		// Read generated CSV to see what happened
		userCSVPath := filepath.Join(tempDir, "User.csv")
		profileCSVPath := filepath.Join(tempDir, "Profile.csv")

		userContent, err := os.ReadFile(userCSVPath)
		require.NoError(t, err)
		profileContent, err := os.ReadFile(profileCSVPath)
		require.NoError(t, err)

		t.Logf("AttributeAlias User.csv:\n%s", string(userContent))
		t.Logf("AttributeAlias Profile.csv:\n%s", string(profileContent))
	})

	t.Run("should verify pipeline phase execution order", func(t *testing.T) {
		// Test that verifies the 3-phase pipeline executes in correct order
		def := &parser.SORDefinition{
			DisplayName: "Phase Order Test",
			Description: "Test pipeline phase execution order",
			Entities: map[string]parser.Entity{
				"entity": {
					DisplayName: "Entity",
					ExternalId:  "Entity",
					Attributes: []parser.Attribute{
						{Name: "id", ExternalId: "id", Type: "String", UniqueId: true},
						{Name: "regular_field", ExternalId: "regular_field", Type: "String"},
					},
				},
			},
		}

		graphInterface, err := model.NewGraph(def, 100)
		require.NoError(t, err)
		graph, ok := graphInterface.(*model.Graph)
		require.True(t, ok)

		tempDir, err := os.MkdirTemp("", "phase-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Execute pipeline
		generator := NewDataGenerator(tempDir, map[string]int{"Entity": 1}, false)
		err = generator.Generate(graph)
		require.NoError(t, err)

		// Verify entity has data after pipeline execution
		entities := graph.GetAllEntities()
		require.Len(t, entities, 1)

		var entity model.EntityInterface
		for _, e := range entities {
			entity = e
			break
		}

		assert.Equal(t, 1, entity.GetRowCount(), "Should have 1 row after pipeline")

		// Verify CSV file was written
		csvPath := filepath.Join(tempDir, "Entity.csv")
		_, err = os.Stat(csvPath)
		assert.NoError(t, err, "CSV file should be written")

		csvData := entity.ToCSV()
		require.Len(t, csvData.Rows, 1)
		row := csvData.Rows[0]

		// All fields should be populated (ID by phase 1, regular_field by phase 3)
		assert.NotEmpty(t, row[0], "ID should be populated by ID generator")
		assert.NotEmpty(t, row[1], "Regular field should be populated by field generator")
	})
}
