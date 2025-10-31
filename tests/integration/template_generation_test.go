package integration

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/SGNL-ai/fabricator/pkg/config"
	"github.com/SGNL-ai/fabricator/pkg/orchestrator"
	"github.com/SGNL-ai/fabricator/pkg/parser"
	"github.com/SGNL-ai/fabricator/pkg/subcommands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// T060: End-to-end template generation test
func TestTemplateGeneration_EndToEnd(t *testing.T) {
	// Use an existing example SOR file
	sorPath := "../../examples/okta.sgnl.yaml"
	if _, err := os.Stat(sorPath); os.IsNotExist(err) {
		t.Skip("Skipping test: example SOR file not found")
	}

	// Generate template
	var buf bytes.Buffer
	opts := subcommands.InitCountConfigOptions{
		SORFile:      sorPath,
		DefaultCount: 100,
		Output:       &buf,
	}

	err := subcommands.InitCountConfig(opts)
	require.NoError(t, err, "Template generation should succeed")

	output := buf.String()
	assert.NotEmpty(t, output, "Template should not be empty")

	// Verify template structure
	assert.Contains(t, output, "# Row count configuration for fabricator")
	assert.Contains(t, output, "# Generated from:")
	assert.Contains(t, output, sorPath)

	// Verify all entities from the SOR are present
	p := parser.NewParser(sorPath)
	err = p.Parse()
	require.NoError(t, err)

	for _, entity := range p.Definition.Entities {
		assert.Contains(t, output, entity.ExternalId+":", "Template should include entity "+entity.ExternalId)
	}

	// Verify default counts are set
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") && strings.Contains(line, ":") && !strings.HasPrefix(line, " ") {
			assert.Contains(t, line, "100", "All entities should have default count of 100")
		}
	}
}

// T061: Test that generated template is valid YAML
func TestTemplateGeneration_ValidYAML(t *testing.T) {
	// Use an existing example SOR file
	sorPath := "../../examples/okta.sgnl.yaml"
	if _, err := os.Stat(sorPath); os.IsNotExist(err) {
		t.Skip("Skipping test: example SOR file not found")
	}

	// Generate template
	var buf bytes.Buffer
	opts := subcommands.InitCountConfigOptions{
		SORFile:      sorPath,
		DefaultCount: 150,
		Output:       &buf,
	}

	err := subcommands.InitCountConfig(opts)
	require.NoError(t, err)

	output := buf.Bytes()

	// Try to parse the output as YAML
	var counts map[string]int
	err = yaml.Unmarshal(output, &counts)
	require.NoError(t, err, "Generated template should be valid YAML")

	// Verify counts were parsed correctly
	assert.NotEmpty(t, counts, "YAML should contain entity counts")
	for entity, count := range counts {
		assert.Greater(t, count, 0, "Count for %s should be positive", entity)
		assert.Equal(t, 150, count, "Count for %s should be 150", entity)
	}
}

// T062: Test that generated template can be used immediately
func TestTemplateGeneration_TemplateCanBeUsedImmediately(t *testing.T) {
	// Use an existing example SOR file
	sorPath := "../../examples/okta.sgnl.yaml"
	if _, err := os.Stat(sorPath); os.IsNotExist(err) {
		t.Skip("Skipping test: example SOR file not found")
	}

	// Step 1: Generate a template
	var templateBuf bytes.Buffer
	opts := subcommands.InitCountConfigOptions{
		SORFile:      sorPath,
		DefaultCount: 75,
		Output:       &templateBuf,
	}

	err := subcommands.InitCountConfig(opts)
	require.NoError(t, err, "Template generation should succeed")

	// Step 2: Save the template to a temporary file
	tmpDir := t.TempDir()
	countConfigPath := tmpDir + "/counts.yaml"
	err = os.WriteFile(countConfigPath, templateBuf.Bytes(), 0644)
	require.NoError(t, err, "Should be able to save template to file")

	// Step 3: Load the config using the config package
	countConfig, err := config.LoadConfiguration(countConfigPath)
	require.NoError(t, err, "Should be able to load generated template as config")

	// Step 4: Parse the SOR to validate against entities
	p := parser.NewParser(sorPath)
	err = p.Parse()
	require.NoError(t, err)

	var entityIDs []string
	for _, entity := range p.Definition.Entities {
		entityIDs = append(entityIDs, entity.ExternalId)
	}

	// Step 5: Validate the config
	err = countConfig.Validate(entityIDs)
	require.NoError(t, err, "Generated template should be valid config")

	// Step 6: Use the config to generate CSVs
	outputDir := tmpDir + "/output"
	options := orchestrator.GenerationOptions{
		DataVolume:      75,
		CountConfig:     countConfig,
		AutoCardinality: false,
		GenerateDiagram: false,
		ValidateResults: false,
	}

	_, err = orchestrator.RunGeneration(p.Definition, outputDir, options)
	require.NoError(t, err, "Should be able to generate CSVs using the template")

	// Step 7: Verify CSVs were generated with correct row counts
	verifyCSVRowCount(t, outputDir+"/User.csv", 75)
	verifyCSVRowCount(t, outputDir+"/Group.csv", 75)
	verifyCSVRowCount(t, outputDir+"/Application.csv", 75)
}

// Additional test: Template generation with custom default count
func TestTemplateGeneration_CustomDefaultCount(t *testing.T) {
	sorPath := "../../examples/okta.sgnl.yaml"
	if _, err := os.Stat(sorPath); os.IsNotExist(err) {
		t.Skip("Skipping test: example SOR file not found")
	}

	// Generate template with custom default count
	var buf bytes.Buffer
	opts := subcommands.InitCountConfigOptions{
		SORFile:      sorPath,
		DefaultCount: 500,
		Output:       &buf,
	}

	err := subcommands.InitCountConfig(opts)
	require.NoError(t, err)

	// Verify custom default is used
	var counts map[string]int
	err = yaml.Unmarshal(buf.Bytes(), &counts)
	require.NoError(t, err)

	for entity, count := range counts {
		assert.Equal(t, 500, count, "Entity %s should have custom default count of 500", entity)
	}
}
