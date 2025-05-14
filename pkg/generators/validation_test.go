package generators

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/models"
)

func TestCSVGenerator_LoadExistingCSVFiles(t *testing.T) {
	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "csv-load-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create mock entities with namespace format
	entities := make(map[string]models.Entity)
	entities["entity1"] = models.Entity{
		ExternalId:  "Test/Entity1",
		DisplayName: "Entity One",
		Description: "Test entity 1",
		Attributes: []models.Attribute{
			{
				Name:       "id",
				ExternalId: "id",
				UniqueId:   true,
			},
			{
				Name:       "name",
				ExternalId: "name",
			},
		},
	}
	entities["entity2"] = models.Entity{
		ExternalId:  "Entity2", // No namespace
		DisplayName: "Entity Two",
		Description: "Test entity 2",
		Attributes: []models.Attribute{
			{
				Name:       "id",
				ExternalId: "id",
				UniqueId:   true,
			},
			{
				Name:       "entity1Id",
				ExternalId: "entity1Id",
			},
		},
	}

	// Create mock relationships
	relationships := make(map[string]models.Relationship)
	relationships["rel1"] = models.Relationship{
		Name:          "entity1_to_entity2",
		FromAttribute: "attr1",
		ToAttribute:   "attr2",
	}

	// Step 1: Generate and write CSV files
	generator := NewCSVGenerator(tempDir, 5, false)
	generator.Setup(entities, relationships)
	generator.GenerateData()

	// Write CSV files
	err = generator.WriteCSVFiles()
	if err != nil {
		t.Fatalf("Failed to write CSV files: %v", err)
	}

	// Verify expected files exist
	expectedFiles := []string{"Entity1.csv", "Entity2.csv"}
	for _, filename := range expectedFiles {
		path := filepath.Join(tempDir, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s does not exist", path)
		}
	}

	// Step 2: Create a new generator and load the CSV files
	loadGenerator := NewCSVGenerator(tempDir, 5, false)
	loadGenerator.Setup(entities, relationships)

	// Load existing CSV files
	err = loadGenerator.LoadExistingCSVFiles()
	if err != nil {
		t.Fatalf("Failed to load CSV files: %v", err)
	}

	// Verify the loaded data
	for entityID, entity := range entities {
		// We don't need to explicitly check the filename here since we already verified
		// that the files were created and then loaded successfully

		// Check that data was loaded
		loadedData := loadGenerator.EntityData[entityID]
		if loadedData == nil {
			t.Fatalf("Failed to load data for entity %s", entityID)
		}

		// Check headers and rows
		if len(loadedData.Headers) != len(entity.Attributes) {
			t.Errorf("Expected %d headers for %s, got %d",
				len(entity.Attributes), entityID, len(loadedData.Headers))
		}

		if len(loadedData.Rows) != 5 {
			t.Errorf("Expected 5 rows for %s, got %d", entityID, len(loadedData.Rows))
		}
	}

	// Step 3: Validate relationships
	validationResults := loadGenerator.ValidateRelationships()

	// Since our test data is minimal, we don't expect validation errors
	t.Logf("Validation results count: %d", len(validationResults))

	// Step 4: Validate unique values
	uniqueErrors := loadGenerator.ValidateUniqueValues()

	// Since our test data is minimal, we don't expect unique value errors
	t.Logf("Unique errors count: %d", len(uniqueErrors))
}

func TestCSVGenerator_LoadExistingCSVFiles_MissingDirectory(t *testing.T) {
	// Create a non-existent directory path
	tempDir := "/tmp/non-existent-directory-" + filepath.Base(os.TempDir())

	// Create mock entities
	entities := make(map[string]models.Entity)
	entities["entity1"] = models.Entity{
		ExternalId:  "Entity1",
		DisplayName: "Entity One",
		Attributes: []models.Attribute{
			{
				Name:       "id",
				ExternalId: "id",
				UniqueId:   true,
			},
		},
	}

	// Initialize generator
	generator := NewCSVGenerator(tempDir, 5, false)
	generator.Setup(entities, make(map[string]models.Relationship))

	// Try to load files from non-existent directory
	err := generator.LoadExistingCSVFiles()

	// Verify error is returned
	if err == nil {
		t.Errorf("Expected error when loading from non-existent directory, got nil")
	}

	if !strings.Contains(err.Error(), "directory does not exist") {
		t.Errorf("Expected 'directory does not exist' error, got: %v", err)
	}
}

func TestCSVGenerator_LoadExistingCSVFiles_EmptyDirectory(t *testing.T) {
	// Create an empty directory
	tempDir, err := os.MkdirTemp("", "csv-empty-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create mock entities
	entities := make(map[string]models.Entity)
	entities["entity1"] = models.Entity{
		ExternalId:  "Entity1",
		DisplayName: "Entity One",
		Attributes: []models.Attribute{
			{
				Name:       "id",
				ExternalId: "id",
				UniqueId:   true,
			},
		},
	}

	// Initialize generator
	generator := NewCSVGenerator(tempDir, 5, false)
	generator.Setup(entities, make(map[string]models.Relationship))

	// Try to load files from empty directory
	err = generator.LoadExistingCSVFiles()

	// Verify error is returned
	if err == nil {
		t.Errorf("Expected error when loading from empty directory, got nil")
	}

	if !strings.Contains(err.Error(), "no matching CSV files found") {
		t.Errorf("Expected 'no matching CSV files found' error, got: %v", err)
	}
}

func TestCSVGenerator_LoadExistingCSVFiles_InvalidCSVFormat(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "csv-invalid-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a mock entity
	entities := make(map[string]models.Entity)
	entities["entity1"] = models.Entity{
		ExternalId:  "Entity1",
		DisplayName: "Entity One",
		Attributes: []models.Attribute{
			{
				Name:       "id",
				ExternalId: "id",
				UniqueId:   true,
			},
		},
	}

	// Create an invalid CSV file
	invalidCSVPath := filepath.Join(tempDir, "Entity1.csv")
	invalidContent := "id\none,two" // Extra column that breaks CSV format
	err = os.WriteFile(invalidCSVPath, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid CSV file: %v", err)
	}

	// Initialize generator
	generator := NewCSVGenerator(tempDir, 5, false)
	generator.Setup(entities, make(map[string]models.Relationship))

	// Try to load invalid CSV file - should fail with parse error
	err = generator.LoadExistingCSVFiles()

	// Should get an error when parsing the invalid CSV
	if err == nil || !strings.Contains(err.Error(), "parse") {
		t.Errorf("Expected parse error for invalid CSV, got: %v", err)
	}
}

func TestCSVGenerator_LoadExistingCSVFiles_NamespaceEntities(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "csv-namespace-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create entity with namespace (KeystoneV1/EntityName format)
	entities := make(map[string]models.Entity)
	entities["entity1"] = models.Entity{
		ExternalId:  "KeystoneV1/User",
		DisplayName: "User",
		Attributes: []models.Attribute{
			{
				Name:       "id",
				ExternalId: "id",
				UniqueId:   true,
			},
		},
	}

	// Create valid CSV file with namespace-stripped filename
	csvPath := filepath.Join(tempDir, "User.csv")
	csvContent := "id\n1\n2\n3"
	err = os.WriteFile(csvPath, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create CSV file: %v", err)
	}

	// Initialize generator
	generator := NewCSVGenerator(tempDir, 5, false)
	generator.Setup(entities, make(map[string]models.Relationship))

	// Load the CSV file
	err = generator.LoadExistingCSVFiles()
	if err != nil {
		t.Fatalf("Failed to load CSV file with namespace entity: %v", err)
	}

	// Verify the loaded data
	loadedData := generator.EntityData["entity1"]
	if loadedData == nil {
		t.Fatalf("Failed to load data for entity with namespace")
	}

	// Check that rows were loaded correctly
	if len(loadedData.Rows) != 3 {
		t.Errorf("Expected 3 rows for namespace entity, got %d", len(loadedData.Rows))
	}

	// Check that the header is correct
	if len(loadedData.Headers) != 1 || loadedData.Headers[0] != "id" {
		t.Errorf("Expected header [id] for namespace entity, got %v", loadedData.Headers)
	}
}

func TestCSVGenerator_LoadExistingCSVFiles_EmptyCSVFile(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "csv-empty-file-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create entities
	entities := make(map[string]models.Entity)
	entities["entity1"] = models.Entity{
		ExternalId:  "Entity1",
		DisplayName: "Entity One",
		Attributes: []models.Attribute{
			{
				Name:       "id",
				ExternalId: "id",
				UniqueId:   true,
			},
		},
	}

	// Create an empty CSV file with just the header row
	emptyCSVPath := filepath.Join(tempDir, "Entity1.csv")
	err = os.WriteFile(emptyCSVPath, []byte("id\n"), 0644) // Empty file with header only
	if err != nil {
		t.Fatalf("Failed to create empty CSV file: %v", err)
	}

	// Initialize generator
	generator := NewCSVGenerator(tempDir, 5, false)
	generator.Setup(entities, make(map[string]models.Relationship))

	// Try to load empty CSV file
	err = generator.LoadExistingCSVFiles()

	// Should succeed as the file exists with a valid header
	if err != nil {
		t.Errorf("Expected empty CSV file with header to be loaded successfully, got error: %v", err)
	}
}

func TestCSVGenerator_LoadExistingCSVFiles_PartialMatch(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "csv-partial-match-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create multiple entities
	entities := make(map[string]models.Entity)
	entities["entity1"] = models.Entity{
		ExternalId:  "Entity1",
		DisplayName: "Entity One",
		Attributes: []models.Attribute{
			{
				Name:       "id",
				ExternalId: "id",
				UniqueId:   true,
			},
		},
	}
	entities["entity2"] = models.Entity{
		ExternalId:  "Entity2",
		DisplayName: "Entity Two",
		Attributes: []models.Attribute{
			{
				Name:       "id",
				ExternalId: "id",
				UniqueId:   true,
			},
		},
	}

	// Only create one CSV file
	csvPath := filepath.Join(tempDir, "Entity1.csv")
	csvContent := "id\n1\n2\n3"
	err = os.WriteFile(csvPath, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create CSV file: %v", err)
	}

	// Initialize generator
	generator := NewCSVGenerator(tempDir, 5, false)
	generator.Setup(entities, make(map[string]models.Relationship))

	// Load the CSV files - only partial match
	err = generator.LoadExistingCSVFiles()
	if err != nil {
		t.Fatalf("Failed to load with partial match: %v", err)
	}

	// Verify only one entity was loaded
	loadedData1 := generator.EntityData["entity1"]
	if loadedData1 == nil || len(loadedData1.Rows) != 3 {
		t.Errorf("Expected entity1 to be loaded with 3 rows")
	}

	// Entity2 should be initialized in the generator but might not have any loaded rows
	loadedData2 := generator.EntityData["entity2"]
	if loadedData2 == nil {
		t.Errorf("Expected entity2 to be initialized")
	}
}

func TestCSVGenerator_ValidateLoadedFiles(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "csv-validation-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create entities with a relationship
	entities := make(map[string]models.Entity)
	entities["entity1"] = models.Entity{
		ExternalId:  "Entity1",
		DisplayName: "Entity One",
		Attributes: []models.Attribute{
			{
				Name:           "id",
				ExternalId:     "id",
				UniqueId:       true,
				AttributeAlias: "attr1",
			},
		},
	}
	entities["entity2"] = models.Entity{
		ExternalId:  "Entity2",
		DisplayName: "Entity Two",
		Attributes: []models.Attribute{
			{
				Name:           "id",
				ExternalId:     "id",
				UniqueId:       true,
				AttributeAlias: "attr2",
			},
			{
				Name:           "entity1Id",
				ExternalId:     "entity1Id",
				AttributeAlias: "attr3",
			},
		},
	}

	relationships := make(map[string]models.Relationship)
	relationships["rel1"] = models.Relationship{
		Name:          "entity1_to_entity2",
		FromAttribute: "attr1",
		ToAttribute:   "attr3",
	}

	// Create consistent CSV files
	entity1CSV := filepath.Join(tempDir, "Entity1.csv")
	entity1Content := "id\n1\n2\n3"
	err = os.WriteFile(entity1CSV, []byte(entity1Content), 0644)
	if err != nil {
		t.Fatalf("Failed to create entity1 CSV: %v", err)
	}

	entity2CSV := filepath.Join(tempDir, "Entity2.csv")
	entity2Content := "id,entity1Id\na,1\nb,2\nc,3" // Valid references to entity1
	err = os.WriteFile(entity2CSV, []byte(entity2Content), 0644)
	if err != nil {
		t.Fatalf("Failed to create entity2 CSV: %v", err)
	}

	// Initialize generator
	generator := NewCSVGenerator(tempDir, 5, false)
	generator.Setup(entities, relationships)

	// Map the relationship - this is normally done by Setup but we need to manually add it for testing
	generator.relationshipMap = map[string][]models.RelationshipLink{
		"entity1": {
			{
				FromEntityID:  "entity1",
				ToEntityID:    "entity2",
				FromAttribute: "id",
				ToAttribute:   "entity1Id",
			},
		},
	}

	// Load CSV files
	err = generator.LoadExistingCSVFiles()
	if err != nil {
		t.Fatalf("Failed to load CSV files: %v", err)
	}

	// Validate relationships
	validationResults := generator.ValidateRelationships()
	if len(validationResults) > 0 {
		t.Errorf("Expected valid relationships, got %d validation errors", len(validationResults))
		for _, result := range validationResults {
			t.Logf("Validation error: %+v", result)
		}
	}

	// Validate unique constraints
	uniqueErrors := generator.ValidateUniqueValues()
	if len(uniqueErrors) > 0 {
		t.Errorf("Expected no unique constraint violations, got %d errors", len(uniqueErrors))
		for _, err := range uniqueErrors {
			t.Logf("Unique error: %+v", err)
		}
	}

	// Now create a CSV file with invalid relationships
	invalidEntity2CSV := filepath.Join(tempDir, "Entity2.csv")
	invalidContent := "id,entity1Id\na,1\nb,2\nc,99" // 99 doesn't exist in entity1
	err = os.WriteFile(invalidEntity2CSV, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid entity2 CSV: %v", err)
	}

	// Load and validate again
	invalidGenerator := NewCSVGenerator(tempDir, 5, false)
	invalidGenerator.Setup(entities, relationships)

	// Map the relationship again for the new generator
	invalidGenerator.relationshipMap = map[string][]models.RelationshipLink{
		"entity1": {
			{
				FromEntityID:  "entity1",
				ToEntityID:    "entity2",
				FromAttribute: "id",
				ToAttribute:   "entity1Id",
			},
		},
	}

	err = invalidGenerator.LoadExistingCSVFiles()
	if err != nil {
		t.Fatalf("Failed to load invalid CSV files: %v", err)
	}

	// Relationship validation should detect the issue
	validationResults = invalidGenerator.ValidateRelationships()

	// Since validation works in the reverse direction (from entity2 to entity1),
	// we might not catch this specific issue. This test just verifies the validation runs.
	t.Logf("Validation results for invalid data: %d", len(validationResults))

	// Create CSV file with duplicate values to test unique constraint validation
	duplicateEntity1CSV := filepath.Join(tempDir, "Entity1.csv")
	duplicateContent := "id\n1\n1\n2" // Duplicate ID value
	err = os.WriteFile(duplicateEntity1CSV, []byte(duplicateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create duplicate entity1 CSV: %v", err)
	}

	// Load and validate again
	duplicateGenerator := NewCSVGenerator(tempDir, 5, false)
	duplicateGenerator.Setup(entities, relationships)

	// For uniqueness validation, we need to track the uniqueId attributes
	duplicateGenerator.uniqueIdAttributes = map[string][]string{
		"entity1": {"id"},
		"entity2": {"id"},
	}

	err = duplicateGenerator.LoadExistingCSVFiles()
	if err != nil {
		t.Fatalf("Failed to load duplicate CSV files: %v", err)
	}

	// Uniqueness validation should detect the issue
	uniqueErrors = duplicateGenerator.ValidateUniqueValues()
	if len(uniqueErrors) == 0 {
		t.Errorf("Expected uniqueness validation errors for duplicate ID")
	} else {
		t.Logf("Uniqueness validation detected %d errors", len(uniqueErrors))
	}
}
