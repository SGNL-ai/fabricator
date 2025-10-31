package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// CountConfiguration maps entity external IDs to their desired row counts.
// It represents a parsed row count configuration from a YAML file.
type CountConfiguration struct {
	// EntityCounts maps entity external_id → row count
	EntityCounts map[string]int

	// SourceFile is the path to the configuration file (for error messages)
	SourceFile string

	// LoadedAt is when the configuration was loaded
	LoadedAt time.Time
}

// LoadConfiguration reads and parses a row count configuration YAML file.
// Returns a CountConfiguration with the parsed entity counts.
//
// The YAML file should have a simple flat structure:
//
//	users: 1000
//	groups: 50
//	permissions: 200
//
// Returns an error if the file cannot be read or parsed.
func LoadConfiguration(path string) (*CountConfiguration, error) {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ValidationError{
			Message:    fmt.Sprintf("Count configuration file not found: %s", path),
			Suggestion: fmt.Sprintf("Generate a template with 'fabricator init-count-config -f <sor.yaml> > %s'", path),
		}
	}

	// Parse YAML
	var entityCounts map[string]int
	err = yaml.Unmarshal(data, &entityCounts)
	if err != nil {
		return nil, &ValidationError{
			Message:    fmt.Sprintf("Invalid YAML syntax in %s: %v", path, err),
			Suggestion: "Validate YAML syntax at yamllint.com or use a YAML validator",
		}
	}

	return &CountConfiguration{
		EntityCounts: entityCounts,
		SourceFile:   path,
		LoadedAt:     time.Now(),
	}, nil
}

// GetCount returns the row count for an entity, or defaultCount if not specified.
// If the entity has a count of 0 in the map, defaultCount is returned.
func (c *CountConfiguration) GetCount(entityExternalID string, defaultCount int) int {
	count, exists := c.EntityCounts[entityExternalID]
	if !exists || count == 0 {
		return defaultCount
	}
	return count
}

// Validate checks the configuration against SOR entities.
// It verifies that:
// - All entities referenced in the config exist in the SOR
// - All count values are positive integers (>0)
//
// Returns a ValidationError if validation fails.
func (c *CountConfiguration) Validate(sorEntities []string) error {
	// Build a set of valid entity IDs for O(1) lookup
	validEntities := make(map[string]bool, len(sorEntities))
	for _, entity := range sorEntities {
		validEntities[entity] = true
	}

	// Validate each entity in the configuration
	for entityID, count := range c.EntityCounts {
		// Check if entity exists in SOR
		if !validEntities[entityID] {
			return &ValidationError{
				EntityID:   entityID,
				Field:      "entity",
				Value:      entityID,
				Message:    fmt.Sprintf("Entity '%s' in count configuration not found in SOR YAML\nAvailable entities: %v", entityID, sorEntities),
				Suggestion: fmt.Sprintf("Remove '%s' or check entity external_id spelling", entityID),
			}
		}

		// Check if count is positive
		if count <= 0 {
			return &ValidationError{
				EntityID:   entityID,
				Field:      "count",
				Value:      count,
				Message:    fmt.Sprintf("Invalid count for entity '%s': %d (expected positive integer)", entityID, count),
				Suggestion: "Use a number like 100, 1000, etc.",
			}
		}
	}

	return nil
}

// HasEntity returns true if entity has an explicit count in the configuration.
func (c *CountConfiguration) HasEntity(entityExternalID string) bool {
	_, exists := c.EntityCounts[entityExternalID]
	return exists
}
