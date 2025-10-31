package subcommands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T051: Test for init-count-config subcommand with valid SOR file
func TestInitCountConfig_ValidSORFile(t *testing.T) {
	// Use an existing example SOR file
	sorPath := "../../examples/okta.sgnl.yaml"
	if _, err := os.Stat(sorPath); os.IsNotExist(err) {
		t.Skip("Skipping test: example SOR file not found")
	}

	// Create a buffer to capture output
	var buf bytes.Buffer

	opts := InitCountConfigOptions{
		SORFile:      sorPath,
		DefaultCount: 100,
		Output:       &buf,
	}

	err := InitCountConfig(opts)
	require.NoError(t, err, "InitCountConfig should succeed with valid SOR file")

	output := buf.String()
	assert.NotEmpty(t, output, "Should generate non-empty output")

	// Verify output contains expected elements
	assert.Contains(t, output, "# Row count configuration for fabricator")
	assert.Contains(t, output, "# Generated from:")
	assert.Contains(t, output, sorPath)

	// Verify entities are present (from okta.sgnl.yaml)
	assert.Contains(t, output, "User:")
	assert.Contains(t, output, "Group:")
	assert.Contains(t, output, "Application:")

	// Verify default counts
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "User:") || strings.HasPrefix(line, "Group:") || strings.HasPrefix(line, "Application:") {
			assert.Contains(t, line, "100", "Should use default count of 100")
		}
	}
}

// T052: Test for init-count-config subcommand with invalid SOR file
func TestInitCountConfig_InvalidSORFile(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "fabricator-test-invalid-sor-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create an invalid YAML file
	invalidYAMLPath := filepath.Join(tempDir, "invalid.yaml")
	invalidContent := `this is not valid YAML for fabricator
entities:
  - this will fail parsing`
	err = os.WriteFile(invalidYAMLPath, []byte(invalidContent), 0644)
	require.NoError(t, err)

	var buf bytes.Buffer

	opts := InitCountConfigOptions{
		SORFile:      invalidYAMLPath,
		DefaultCount: 100,
		Output:       &buf,
	}

	err = InitCountConfig(opts)
	assert.Error(t, err, "InitCountConfig should fail with invalid SOR file")
	assert.Contains(t, err.Error(), "failed to parse SOR file", "Error should indicate parsing failure")
}

// T053: Test for init-count-config subcommand with missing SOR file
func TestInitCountConfig_MissingSORFile(t *testing.T) {
	// Use a non-existent file path
	nonExistentPath := "/tmp/this-file-definitely-does-not-exist-12345.yaml"

	var buf bytes.Buffer

	opts := InitCountConfigOptions{
		SORFile:      nonExistentPath,
		DefaultCount: 100,
		Output:       &buf,
	}

	err := InitCountConfig(opts)
	assert.Error(t, err, "InitCountConfig should fail with missing SOR file")
	assert.Contains(t, err.Error(), "SOR file not found", "Error should indicate file not found")
}

// Additional test: Empty SOR file path
func TestInitCountConfig_EmptySORFilePath(t *testing.T) {
	var buf bytes.Buffer

	opts := InitCountConfigOptions{
		SORFile:      "", // Empty path
		DefaultCount: 100,
		Output:       &buf,
	}

	err := InitCountConfig(opts)
	assert.Error(t, err, "InitCountConfig should fail with empty SOR file path")
	assert.Contains(t, err.Error(), "SOR file path is required")
}

// Additional test: Default count parameter
func TestInitCountConfig_CustomDefaultCount(t *testing.T) {
	// Use an existing example SOR file
	sorPath := "../../examples/okta.sgnl.yaml"
	if _, err := os.Stat(sorPath); os.IsNotExist(err) {
		t.Skip("Skipping test: example SOR file not found")
	}

	var buf bytes.Buffer

	opts := InitCountConfigOptions{
		SORFile:      sorPath,
		DefaultCount: 250, // Custom default
		Output:       &buf,
	}

	err := InitCountConfig(opts)
	require.NoError(t, err)

	output := buf.String()

	// Verify custom default count is used
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "User:") || strings.HasPrefix(line, "Group:") || strings.HasPrefix(line, "Application:") {
			assert.Contains(t, line, "250", "Should use custom default count of 250")
		}
	}
}

// Additional test: Zero default count should use 100
func TestInitCountConfig_ZeroDefaultCountFallback(t *testing.T) {
	// Use an existing example SOR file
	sorPath := "../../examples/okta.sgnl.yaml"
	if _, err := os.Stat(sorPath); os.IsNotExist(err) {
		t.Skip("Skipping test: example SOR file not found")
	}

	var buf bytes.Buffer

	opts := InitCountConfigOptions{
		SORFile:      sorPath,
		DefaultCount: 0, // Zero should fall back to 100
		Output:       &buf,
	}

	err := InitCountConfig(opts)
	require.NoError(t, err)

	output := buf.String()

	// Verify fallback to 100
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "User:") || strings.HasPrefix(line, "Group:") || strings.HasPrefix(line, "Application:") {
			assert.Contains(t, line, "100", "Should fall back to default count of 100 when zero provided")
		}
	}
}
