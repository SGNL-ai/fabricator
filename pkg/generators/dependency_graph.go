package generators

import (
	"fmt"
	"strings"
	
	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/SGNL-ai/fabricator/pkg/util"
	"github.com/dominikbraun/graph"
	"github.com/fatih/color"
)

// TODO: Review this dependency graph implementation for edge cases and improvements
// Possible enhancements:
// 1. Handle more complex relationship types and paths
// 2. Consider additional metrics for entity generation order
// 3. Add support for weighted edges based on relationship types
// 
// Note: Cyclic dependency detection is now handled in two places:
// 1. Basic detection in the YAML parser (validateRelationships)
// 2. Detailed handling here during dependency graph construction

// buildEntityDependencyGraph creates a directed graph of entity dependencies
// based on the relationships in the SOR definition.
func (g *CSVGenerator) buildEntityDependencyGraph(
	entities map[string]models.Entity,
	relationships map[string]models.Relationship,
) (graph.Graph[string, string], error) {
	// Use shared utility to build the dependency graph
	// We want to prevent cycles for the CSV generator
	entityGraph, err := util.BuildEntityDependencyGraph(entities, relationships, true)
	if err != nil {
		return nil, err
	}

	// Debug relationship information for CSV generator
	for relName, relationship := range relationships {
		// Skip path-based relationships for now
		if len(relationship.Path) > 0 {
			continue
		}

		// Parse the entity-attribute pairs from the relationship using our implementation
		// We'll create a local helper function for debugging output
		fromEntityID, fromAttrName, fromUniqueID := parseEntityAttribute(entities, relationship.FromAttribute)
		toEntityID, toAttrName, toUniqueID := parseEntityAttribute(entities, relationship.ToAttribute)

		// Debug the relationship
		color.Cyan("Relationship: %s -> %s", relationship.FromAttribute, relationship.ToAttribute)
		color.Cyan("  From Entity: %s, Attribute: %s, UniqueID: %t", fromEntityID, fromAttrName, fromUniqueID)
		color.Cyan("  To Entity: %s, Attribute: %s, UniqueID: %t", toEntityID, toAttrName, toUniqueID)

		// If we couldn't identify both ends of the relationship, log it
		if fromEntityID == "" || toEntityID == "" {
			color.Yellow("Skipping relationship %s: couldn't identify entities", relName)
		}
	}

	return entityGraph, nil
}

// getTopologicalOrder returns a topologically sorted list of entity IDs
// Entities will be sorted such that if entity A depends on entity B,
// entity B will appear earlier in the list.
func (g *CSVGenerator) getTopologicalOrder(
	entityGraph graph.Graph[string, string],
) ([]string, error) {
	// Use shared utility to get topological ordering
	ordering, err := util.GetTopologicalOrder(entityGraph)
	
	if err != nil {
		// If topological sort fails, output the error and stop
		color.Red("Error: Cannot generate data in dependency order: %v", err)
		color.Yellow("Data generation cannot safely proceed.")
		color.Yellow("Please review your entity relationships.")
		return nil, fmt.Errorf("failed to perform topological sort: %w", err)
	}

	return ordering, nil
}

// Helper function to parse Entity.Attribute format and find the corresponding entity and attribute
// This is used for debug output in the CSV generator
func parseEntityAttribute(
	entities map[string]models.Entity,
	attributeRef string,
) (entityID string, attrName string, uniqueID bool) {
	// Check if it's in Entity.Attribute format
	if strings.Contains(attributeRef, ".") {
		parts := strings.Split(attributeRef, ".")
		if len(parts) == 2 {
			entityName := parts[0]
			attributeName := parts[1]
			
			// Find the entity by external ID
			for id, entity := range entities {
				if entity.ExternalId == entityName {
					// Find the attribute
					for _, attr := range entity.Attributes {
						if attr.ExternalId == attributeName {
							return id, attr.Name, attr.UniqueId
						}
					}
				}
			}
		}
	}
	
	return "", "", false
}