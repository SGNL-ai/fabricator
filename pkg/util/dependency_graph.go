package util

import (
	"errors"
	"fmt"
	"strings"

	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/dominikbraun/graph"
)

// BuildEntityDependencyGraph creates a directed graph of entity dependencies
// based on the relationships in the SOR definition.
// This shared utility is used by both the CSV generator and ER diagram generator.
func BuildEntityDependencyGraph(
	entities map[string]models.Entity,
	relationships map[string]models.Relationship,
	preventCycles bool,
) (graph.Graph[string, string], error) {
	// Create a directed graph where vertices are entity IDs
	var entityGraph graph.Graph[string, string]
	if preventCycles {
		entityGraph = graph.New(graph.StringHash, graph.Directed(), graph.PreventCycles())
	} else {
		entityGraph = graph.New(graph.StringHash, graph.Directed())
	}

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

	// Create maps to find entities/attributes by different identifiers
	// Map of attribute alias to (entityID, attrName, uniqueID)
	attributeAliasMap := make(map[string]struct {
		EntityID      string
		AttributeName string
		UniqueID      bool
	})

	// Map to lookup entities/attributes by "Entity.Attribute" pattern
	entityAttributeMap := make(map[string]struct {
		EntityID      string
		AttributeName string
		UniqueID      bool
	})

	// Build the attribute maps
	for entityID, entity := range entities {
		for _, attr := range entity.Attributes {
			// Handle attributeAlias case (when it exists)
			if attr.AttributeAlias != "" {
				attributeAliasMap[attr.AttributeAlias] = struct {
					EntityID      string
					AttributeName string
					UniqueID      bool
				}{
					EntityID:      entityID,
					AttributeName: attr.Name,
					UniqueID:      attr.UniqueId,
				}
			}

			// Also build Entity.Attribute map for YAMLs without attributeAlias
			entityKey := entity.ExternalId + "." + attr.ExternalId
			entityAttributeMap[entityKey] = struct {
				EntityID      string
				AttributeName string
				UniqueID      bool
			}{
				EntityID:      entityID,
				AttributeName: attr.Name,
				UniqueID:      attr.UniqueId,
			}
		}
	}

	// Process relationships to add edges
	for _, relationship := range relationships {
		// Skip path-based relationships for now
		if len(relationship.Path) > 0 {
			continue
		}

		// Parse the entity-attribute pairs from the relationship
		fromEntityID, fromAttrName, fromUniqueID := ParseEntityAttribute(
			entities, relationship.FromAttribute, attributeAliasMap, entityAttributeMap)
		toEntityID, toAttrName, toUniqueID := ParseEntityAttribute(
			entities, relationship.ToAttribute, attributeAliasMap, entityAttributeMap)

		// If we couldn't identify both ends of the relationship, skip it
		if fromEntityID == "" || toEntityID == "" {
			continue
		}

		// For relationships where the source and target are different entities
		if fromEntityID != toEntityID {
			// In YAML files, relationships are typically:
			// fromAttribute: contains FK (e.g., Role.appId)
			// toAttribute: contains PK (e.g., App.id)
			
			// For topological sort, we need entities with PKs to be processed
			// before entities with FKs that reference them
			
			// Determine the direction of the edge for topological sort
			var sourceEntityID, targetEntityID string
			
			// Check if the reverse edge already exists (to avoid cycles)
			_, errExistingEdge := entityGraph.Edge(fromEntityID, toEntityID)
			hasReverseEdge := errExistingEdge == nil

			if !fromUniqueID && toUniqueID {
				// Case 1: Standard FK->PK relationship
				// Skip if the reverse edge exists to prevent cycles
				if hasReverseEdge && preventCycles {
					continue
				}
				// For topological sort, PK entity should come before FK entity
				sourceEntityID = toEntityID    // Entity with PK
				targetEntityID = fromEntityID  // Entity with FK
			} else if fromUniqueID && !toUniqueID {
				// Case 2: Reverse PK->FK relationship (could create cycles)
				if preventCycles {
					continue
				}
				sourceEntityID = fromEntityID
				targetEntityID = toEntityID
			} else if fromUniqueID && toUniqueID {
				// Case 3: PK->PK identity relationship
				// Skip if the reverse edge exists to prevent cycles
				if hasReverseEdge && preventCycles {
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
			} else {
				// Case 4: Neither attribute is a unique ID
				// Skip if the reverse edge exists to prevent cycles
				if hasReverseEdge && preventCycles {
					continue
				}
				
				// Default to standard relationship direction
				sourceEntityID = toEntityID
				targetEntityID = fromEntityID
			}
			
			// Add the edge: source -> target means "source should be processed before target"
			err := entityGraph.AddEdge(sourceEntityID, targetEntityID)
			if err != nil {
				// Handle the error based on its type
				if errors.Is(err, graph.ErrEdgeCreatesCycle) {
					if preventCycles {
						continue // Skip this edge to prevent cycles
					}
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

// ParseEntityAttribute is a helper function to extract entity and attribute info from a reference
// Public version of parseEntityAttribute that can be used across packages
func ParseEntityAttribute(
	entities map[string]models.Entity,
	attributeRef string,
	attributeAliasMap map[string]struct {
		EntityID      string
		AttributeName string
		UniqueID      bool
	},
	entityAttributeMap map[string]struct {
		EntityID      string
		AttributeName string
		UniqueID      bool
	},
) (entityID string, attrName string, uniqueID bool) {
	// First try attribute alias mapping
	if info, found := attributeAliasMap[attributeRef]; found {
		return info.EntityID, info.AttributeName, info.UniqueID
	}
	
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
	
	// Try entity attribute map as a fallback
	if info, found := entityAttributeMap[attributeRef]; found {
		return info.EntityID, info.AttributeName, info.UniqueID
	}
	
	return "", "", false
}

// GetTopologicalOrder returns a topologically sorted list of entity IDs
// Entities will be sorted such that if entity A depends on entity B,
// entity B will appear earlier in the list.
func GetTopologicalOrder(entityGraph graph.Graph[string, string]) ([]string, error) {
	// Use stable topological sort for deterministic output
	// The comparison function sorts entity IDs alphabetically when there's a choice
	ordering, err := graph.StableTopologicalSort(entityGraph, func(a, b string) bool {
		return strings.Compare(a, b) < 0
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to perform topological sort: %w", err)
	}

	return ordering, nil
}