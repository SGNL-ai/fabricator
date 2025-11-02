package integration

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/config"
	"github.com/SGNL-ai/fabricator/pkg/orchestrator"
	"github.com/SGNL-ai/fabricator/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T034: Integration test for end-to-end with config file
func TestEndToEndWithConfigFile(t *testing.T) {
	// Use an existing example SOR file
	sorPath := "../../examples/okta.sgnl.yaml"
	if _, err := os.Stat(sorPath); os.IsNotExist(err) {
		t.Skip("Skipping test: example SOR file not found")
	}

	// Create a temporary output directory
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")

	// Create a test count configuration
	countConfigPath := filepath.Join(tmpDir, "counts.yaml")
	countConfigContent := `User: 15
Group: 7
Application: 12
`
	err := os.WriteFile(countConfigPath, []byte(countConfigContent), 0644)
	require.NoError(t, err, "Failed to create test count config")

	// Parse the SOR file
	p := parser.NewParser(sorPath)
	err = p.Parse()
	require.NoError(t, err, "Failed to parse SOR file")

	// Load and validate count configuration
	countConfig, err := config.LoadConfiguration(countConfigPath)
	require.NoError(t, err, "Failed to load count configuration")

	// Validate configuration against SOR entities
	var entityIDs []string
	for _, entity := range p.Definition.Entities {
		entityIDs = append(entityIDs, entity.ExternalId)
	}
	err = countConfig.Validate(entityIDs)
	require.NoError(t, err, "Count configuration validation failed")

	// Build row counts map
	rowCounts := orchestrator.BuildRowCountsMap(p.Definition, countConfig, 100)

	// Verify row counts map has correct values
	assert.Equal(t, 15, rowCounts["User"], "User should have 15 rows")
	assert.Equal(t, 7, rowCounts["Group"], "Group should have 7 rows")
	assert.Equal(t, 12, rowCounts["Application"], "Application should have 12 rows")

	// Run generation
	options := orchestrator.GenerationOptions{
		DataVolume:      100, // Default for unspecified entities
		CountConfig:     countConfig,
		AutoCardinality: false,
		GenerateDiagram: false,
		ValidateResults: false,
	}

	_, err = orchestrator.RunGeneration(p.Definition, outputDir, options)
	require.NoError(t, err, "Generation should succeed")

	// Verify CSV files have correct row counts
	verifyCSVRowCount(t, filepath.Join(outputDir, "User.csv"), 15)
	verifyCSVRowCount(t, filepath.Join(outputDir, "Group.csv"), 7)
	verifyCSVRowCount(t, filepath.Join(outputDir, "Application.csv"), 12)
}

// T035: Integration test for backward compatibility (-n flag only)
func TestBackwardCompatibilityWithNFlag(t *testing.T) {
	// Use an existing example SOR file
	sorPath := "../../examples/okta.sgnl.yaml"
	if _, err := os.Stat(sorPath); os.IsNotExist(err) {
		t.Skip("Skipping test: example SOR file not found")
	}

	// Create a temporary output directory
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")

	// Parse the SOR file
	p := parser.NewParser(sorPath)
	err := p.Parse()
	require.NoError(t, err, "Failed to parse SOR file")

	// Build row counts map WITHOUT count config (uniform -n behavior)
	rowCounts := orchestrator.BuildRowCountsMap(p.Definition, nil, 50)

	// Verify all entities get uniform count
	for _, entity := range p.Definition.Entities {
		assert.Equal(t, 50, rowCounts[entity.ExternalId], "All entities should have 50 rows")
	}

	// Run generation
	options := orchestrator.GenerationOptions{
		DataVolume:      50,
		CountConfig:     nil, // No count config - backward compatible mode
		AutoCardinality: false,
		GenerateDiagram: false,
		ValidateResults: false,
	}

	_, err = orchestrator.RunGeneration(p.Definition, outputDir, options)
	require.NoError(t, err, "Generation should succeed")

	// Verify all CSV files have uniform row count (50)
	verifyCSVRowCount(t, filepath.Join(outputDir, "User.csv"), 50)
	verifyCSVRowCount(t, filepath.Join(outputDir, "Group.csv"), 50)
	verifyCSVRowCount(t, filepath.Join(outputDir, "Application.csv"), 50)
}

// T036: Integration test for mixed defaults (partial config)
func TestMixedDefaultsPartialConfig(t *testing.T) {
	// Use an existing example SOR file
	sorPath := "../../examples/okta.sgnl.yaml"
	if _, err := os.Stat(sorPath); os.IsNotExist(err) {
		t.Skip("Skipping test: example SOR file not found")
	}

	// Create a temporary output directory
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")

	// Create a partial count configuration (only some entities specified)
	countConfigPath := filepath.Join(tmpDir, "partial_counts.yaml")
	partialConfigContent := `User: 25
# Group and Application not specified
`
	err := os.WriteFile(countConfigPath, []byte(partialConfigContent), 0644)
	require.NoError(t, err, "Failed to create partial count config")

	// Parse the SOR file
	p := parser.NewParser(sorPath)
	err = p.Parse()
	require.NoError(t, err, "Failed to parse SOR file")

	// Load count configuration
	countConfig, err := config.LoadConfiguration(countConfigPath)
	require.NoError(t, err, "Failed to load count configuration")

	// Build row counts map with default of 100 for unspecified entities
	rowCounts := orchestrator.BuildRowCountsMap(p.Definition, countConfig, 100)

	// Verify mixed counts
	assert.Equal(t, 25, rowCounts["User"], "User should have custom count (25)")
	// Other entities should have default (100)
	assert.Equal(t, 100, rowCounts["Group"], "Group should have default count (100)")
	assert.Equal(t, 100, rowCounts["Application"], "Application should have default count (100)")

	// Run generation
	options := orchestrator.GenerationOptions{
		DataVolume:      100,
		CountConfig:     countConfig,
		AutoCardinality: false,
		GenerateDiagram: false,
		ValidateResults: false,
	}

	_, err = orchestrator.RunGeneration(p.Definition, outputDir, options)
	require.NoError(t, err, "Generation should succeed")

	// Verify CSV files have correct row counts
	verifyCSVRowCount(t, filepath.Join(outputDir, "User.csv"), 25)
	verifyCSVRowCount(t, filepath.Join(outputDir, "Group.csv"), 100)
	verifyCSVRowCount(t, filepath.Join(outputDir, "Application.csv"), 100)
}

// T037: Integration test for cardinality warnings
func TestCardinalityWarnings(t *testing.T) {
	// Use an existing example SOR file
	sorPath := "../../examples/okta.sgnl.yaml"
	if _, err := os.Stat(sorPath); os.IsNotExist(err) {
		t.Skip("Skipping test: example SOR file not found")
	}

	// Create a temporary output directory
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")

	// Create count configuration with severe imbalance
	countConfigPath := filepath.Join(tmpDir, "imbalanced_counts.yaml")
	imbalancedConfigContent := `User: 1000
Group: 10
Application: 5
`
	err := os.WriteFile(countConfigPath, []byte(imbalancedConfigContent), 0644)
	require.NoError(t, err, "Failed to create imbalanced count config")

	// Parse the SOR file
	p := parser.NewParser(sorPath)
	err = p.Parse()
	require.NoError(t, err, "Failed to parse SOR file")

	// Load count configuration
	countConfig, err := config.LoadConfiguration(countConfigPath)
	require.NoError(t, err, "Failed to load count configuration")

	// Run generation
	options := orchestrator.GenerationOptions{
		DataVolume:      100,
		CountConfig:     countConfig,
		AutoCardinality: false,
		GenerateDiagram: false,
		ValidateResults: false,
	}

	_, err = orchestrator.RunGeneration(p.Definition, outputDir, options)
	require.NoError(t, err, "Generation should succeed despite imbalances")

	// Verify CSVs were generated (best-effort)
	verifyCSVRowCount(t, filepath.Join(outputDir, "User.csv"), 1000)
	verifyCSVRowCount(t, filepath.Join(outputDir, "Group.csv"), 10)
	verifyCSVRowCount(t, filepath.Join(outputDir, "Application.csv"), 5)

	// Note: Warnings would be emitted to stderr during execution
	// Testing warning emission would require capturing stderr, which is complex
	// For now, we verify generation succeeds (best-effort behavior)
}

// T038: Integration test for conflict detection (both flags)
// NOTE: This test is at the CLI level and would need to test main() directly
// For now, we rely on unit tests of the validation logic in main_test.go

// T039: Helper function to create test fixtures
func TestFixtureCreation(t *testing.T) {
	// Verify test fixtures exist or can be created
	fixturesDir := "../fixtures"
	err := os.MkdirAll(fixturesDir, 0755)
	require.NoError(t, err, "Should be able to create fixtures directory")

	// Verify test counts file exists
	testCountsPath := filepath.Join(fixturesDir, "test_counts.yaml")
	_, err = os.Stat(testCountsPath)
	if os.IsNotExist(err) {
		t.Logf("Test fixture test_counts.yaml not found, test will create temporary ones")
	}
}

// Helper function to verify CSV row count
func verifyCSVRowCount(t *testing.T, csvPath string, expectedRows int) {
	t.Helper()

	// Open CSV file
	file, err := os.Open(csvPath)
	require.NoError(t, err, "CSV file should exist: %s", csvPath)
	defer func() { _ = file.Close() }()

	// Read CSV
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	require.NoError(t, err, "Should be able to read CSV")

	// Count rows (subtract 1 for header)
	actualRows := len(records) - 1
	assert.Equal(t, expectedRows, actualRows, "CSV %s should have exactly %d data rows (got %d)", filepath.Base(csvPath), expectedRows, actualRows)
}
