package model

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dominikbraun/graph"
)

// EntityGraph encapsulates the underlying graph library
// and provides domain-specific operations for entity relationships
type EntityGraph struct {
	graph graph.Graph[string, string]
}

// newEntityGraph creates a new entity graph
// Not exported as it should only be used internally by EntityStore
func newEntityGraph(preventCycles bool) *EntityGraph {
	var g graph.Graph[string, string]
	if preventCycles {
		g = graph.New(graph.StringHash, graph.Directed(), graph.PreventCycles())
	} else {
		g = graph.New(graph.StringHash, graph.Directed())
	}
	
	return &EntityGraph{
		graph: g,
	}
}

// AddEntity adds an entity to the graph
func (eg *EntityGraph) AddEntity(entityID string) error {
	err := eg.graph.AddVertex(entityID)
	if err != nil {
		// If the entity already exists, we can ignore this error
		if errors.Is(err, graph.ErrVertexAlreadyExists) {
			return nil
		}
		return fmt.Errorf("failed to add entity %s to graph: %w", entityID, err)
	}
	return nil
}

// AddDependency adds a dependency relationship between entities
// sourceID is the entity that should be processed first
// targetID is the entity that depends on the source entity
func (eg *EntityGraph) AddDependency(sourceID, targetID string) error {
	// Check for self-dependency
	if sourceID == targetID {
		return nil // Just ignore self-dependencies
	}
	
	// Add the edge: source -> target means "source should be processed before target"
	err := eg.graph.AddEdge(sourceID, targetID)
	if err != nil {
		// Handle specific error cases
		if errors.Is(err, graph.ErrEdgeCreatesCycle) {
			return fmt.Errorf("adding dependency from %s to %s would create a cycle", sourceID, targetID)
		} else if errors.Is(err, graph.ErrEdgeAlreadyExists) {
			// Edge already exists, we can ignore this
			return nil
		}
		// Other unexpected error
		return fmt.Errorf("failed to add dependency from %s to %s: %w", sourceID, targetID, err)
	}
	
	return nil
}

// HasDependency checks if a dependency exists between entities
func (eg *EntityGraph) HasDependency(sourceID, targetID string) bool {
	_, err := eg.graph.Edge(sourceID, targetID)
	return err == nil
}

// GetTopologicalOrder returns entities in dependency order 
// This is the core algorithm for determining generation order
func (eg *EntityGraph) GetTopologicalOrder() ([]string, error) {
	// Use stable topological sort for deterministic output
	// The comparison function sorts entity IDs alphabetically when there's a choice
	ordering, err := graph.StableTopologicalSort(eg.graph, func(a, b string) bool {
		return strings.Compare(a, b) < 0
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to perform topological sort: %w", err)
	}
	
	return ordering, nil
}

// GetEntities returns all entities in the graph
func (eg *EntityGraph) GetEntities() []string {
	vertices, err := eg.graph.VertexSlice()
	if err != nil {
		return []string{} // Return empty slice on error
	}
	return vertices
}

// GetDirectDependencies returns all entities that directly depend on the given entity
// targetID depends on sourceID if there's an edge sourceID -> targetID
func (eg *EntityGraph) GetDirectDependencies(entityID string) []string {
	// Get all adjacent vertices in the forward direction
	// These are the entities that directly depend on the given entity
	adjacentVertices, err := eg.graph.AdjacencyMap()
	if err != nil {
		return []string{} // Return empty slice on error
	}
	
	// Get the adjacency list for the given entity
	dependents, found := adjacentVertices[entityID]
	if !found {
		return []string{} // Entity not found
	}
	
	return dependents
}

// GetDependenciesFor returns all entities that the given entity depends on
// entityID depends on dependencies if there are edges dependencies -> entityID
func (eg *EntityGraph) GetDependenciesFor(entityID string) []string {
	// Get all adjacent vertices in the reverse direction
	// These are the entities that the given entity depends on
	predecessors, err := eg.graph.PredecessorMap()
	if err != nil {
		return []string{} // Return empty slice on error
	}
	
	// Get the predecessor list for the given entity
	dependencies, found := predecessors[entityID]
	if !found {
		return []string{} // Entity not found
	}
	
	return dependencies
}