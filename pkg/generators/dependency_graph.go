package generators

import (
	"errors"
	"fmt"
	"strings"
	
	"github.com/SGNL-ai/fabricator/pkg/models"
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
	// Create a directed graph where vertices are entity IDs
	// Use PreventCycles to ensure we don't create cycles in the dependency graph
	entityGraph := graph.New(graph.StringHash, graph.Directed(), graph.PreventCycles())

	// Add all entities as vertices in the graph
	for entityID := range entities {
		err := entityGraph.AddVertex(entityID)
		if err != nil {
			// If the vertex already exists, we can ignore this error
			if !errors.Is(err, graph.ErrVertexAlreadyExists) {
				return nil, fmt.Errorf("failed to add entity %s to graph: %w", entityID, err)
			}
		}
	}

	// Process relationships to add edges
	// We need to identify which entities depend on others
	// Note: The YAML Parser already performs basic validation of these relationships,
	// including checking for attribute existence and potential cycle detection.
	// Here we focus on building the dependency graph with optimal ordering.
	for relName, relationship := range relationships {
		// Skip path-based relationships for now
		if len(relationship.Path) > 0 {
			continue
		}

		// Parse the entity-attribute pairs from the relationship
		fromEntityID, fromAttrName, fromUniqueID := parseEntityAttribute(entities, relationship.FromAttribute)
		toEntityID, toAttrName, toUniqueID := parseEntityAttribute(entities, relationship.ToAttribute)

		// Debug the relationship
		color.Cyan("Relationship: %s -> %s", relationship.FromAttribute, relationship.ToAttribute)
		color.Cyan("  From Entity: %s, Attribute: %s, UniqueID: %t", fromEntityID, fromAttrName, fromUniqueID)
		color.Cyan("  To Entity: %s, Attribute: %s, UniqueID: %t", toEntityID, toAttrName, toUniqueID)

		// If we couldn't identify both ends of the relationship, skip it
		// This should not happen if the YAML validation is working properly
		if fromEntityID == "" || toEntityID == "" {
			color.Yellow("Skipping relationship %s: couldn't identify entities", relName)
			continue
		}

		// For relationships where the source and target are different entities
		if fromEntityID != toEntityID {
			// In YAML files, relationships are typically:
			// fromAttribute: contains FK (e.g., Role.appId)
			// toAttribute: contains PK (e.g., App.id)
			
			// For data generation, we need entities with PKs to be processed
			// before entities with FKs that reference them
			
			// Determine the direction of the edge for topological sort
			var sourceEntityID, targetEntityID string
			
			// Check if the reverse edge already exists (to avoid cycles)
			_, errExistingEdge := entityGraph.Edge(fromEntityID, toEntityID)
			hasReverseEdge := errExistingEdge == nil

			if !fromUniqueID && toUniqueID {
				// Case 1: Standard FK->PK relationship
				// Skip if the reverse edge exists to prevent cycles
				if hasReverseEdge {
					color.Yellow("Skipping edge that would create cycle: %s.%s -> %s.%s (reverse edge exists)",
						fromEntityID, fromAttrName, toEntityID, toAttrName)
					continue
				}
				// For topological sort, PK entity should come before FK entity
				sourceEntityID = toEntityID    // Entity with PK
				targetEntityID = fromEntityID  // Entity with FK
				color.Green("Adding edge for FK->PK: %s -> %s (PK entity first)", 
					sourceEntityID, targetEntityID)
			} else if fromUniqueID && !toUniqueID {
				// Case 2: Reverse PK->FK relationship (could create cycles)
				color.Yellow("Skipping reverse PK->FK relationship from %s.%s to %s.%s to prevent cycles",
					fromEntityID, fromAttrName, toEntityID, toAttrName)
				continue
			} else if fromUniqueID && toUniqueID {
				// Case 3: PK->PK identity relationship
				// Skip if the reverse edge exists to prevent cycles
				if hasReverseEdge {
					color.Yellow("Skipping PK->PK relationship that would create cycle: %s.%s -> %s.%s",
						fromEntityID, fromAttrName, toEntityID, toAttrName)
					continue
				}
				
				// If fromAttr looks like a reference (e.g., accountId), then toEntity should come first
				isReference := strings.Contains(fromAttrName, "Id") || strings.Contains(fromAttrName, "ID")
				if isReference && toAttrName == "id" {
					sourceEntityID = toEntityID
					targetEntityID = fromEntityID
				} else {
					sourceEntityID = toEntityID
					targetEntityID = fromEntityID
				}
				color.Green("Adding edge for identity relationship: %s -> %s", 
					sourceEntityID, targetEntityID)
			} else {
				// Case 4: Neither attribute is a unique ID
				// Skip if the reverse edge exists to prevent cycles
				if hasReverseEdge {
					color.Yellow("Skipping edge that would create cycle: %s.%s -> %s.%s (reverse edge exists)",
						fromEntityID, fromAttrName, toEntityID, toAttrName)
					continue
				}
				
				// Default to standard relationship direction
				sourceEntityID = toEntityID
				targetEntityID = fromEntityID
				color.Green("Adding edge for other relationship: %s -> %s", 
					sourceEntityID, targetEntityID)
			}
			
			// Add the edge: source -> target means "source should be processed before target"
			err := entityGraph.AddEdge(sourceEntityID, targetEntityID)
			if err != nil {
				// Handle the error based on its type
				if errors.Is(err, graph.ErrEdgeCreatesCycle) {
					// Edge would create a cycle
					color.Yellow("Warning: Dependency cycle detected in relationship between %s and %s",
						sourceEntityID, targetEntityID)
					// We now have better relationship validation in the parser, but we still handle this.
					// Log the error but continue building the dependency graph by skipping this edge,
					// which is better than failing the entire process.
					color.Yellow("  Skipping this dependency to prevent a cycle. Some relationships may not be fully consistent.")
					continue
				} else if errors.Is(err, graph.ErrEdgeAlreadyExists) {
					// Edge already exists, we can ignore this
					continue
				} else {
					// Other unexpected error
					return nil, fmt.Errorf("failed to add edge from %s to %s: %w",
						sourceEntityID, targetEntityID, err)
				}
			}
		}
	}

	return entityGraph, nil
}

// Helper function to parse Entity.Attribute format and find the corresponding entity and attribute
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
	
	// For attribute alias format (not implemented in this example)
	// Would need to look up attribute by alias in a separate map
	
	return "", "", false
}

// getTopologicalOrder returns a topologically sorted list of entity IDs
// Entities will be sorted such that if entity A depends on entity B,
// entity B will appear earlier in the list.
func (g *CSVGenerator) getTopologicalOrder(
	entityGraph graph.Graph[string, string],
) ([]string, error) {
	// Use stable topological sort for deterministic output
	// The comparison function sorts entity IDs alphabetically when there's a choice
	ordering, err := graph.StableTopologicalSort(entityGraph, func(a, b string) bool {
		return strings.Compare(a, b) < 0
	})
	
	if err != nil {
		// If topological sort fails, output the error and stop
		color.Red("Error: Cannot generate data in dependency order: %v", err)
		color.Yellow("Data generation cannot safely proceed.")
		color.Yellow("Please review your entity relationships.")
		return nil, fmt.Errorf("failed to perform topological sort: %w", err)
	}

	return ordering, nil
}