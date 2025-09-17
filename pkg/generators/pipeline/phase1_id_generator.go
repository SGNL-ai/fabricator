package pipeline

import (
	"github.com/SGNL-ai/fabricator/pkg/generators/model"
)

// IDGenerator handles the generation of entity IDs in topological order
type IDGenerator struct {
	// Configuration options can be added here
}

// NewIDGenerator creates a new ID generator
func NewIDGenerator() IDGeneratorInterface {
	return &IDGenerator{}
}

// GenerateIDs generates unique IDs for all entities in topological order
func (g *IDGenerator) GenerateIDs(graph *model.Graph, dataVolume int) error {
	// Implementation will be added later
	return nil
}
