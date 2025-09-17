package pipeline

import (
	"github.com/SGNL-ai/fabricator/pkg/generators/model"
)

// FieldGenerator handles generation of non-ID and non-relationship fields
type FieldGenerator struct {
	// Configuration options can be added here
}

// NewFieldGenerator creates a new field generator
func NewFieldGenerator() FieldGeneratorInterface {
	return &FieldGenerator{}
}

// GenerateFields generates values for all non-ID and non-relationship fields
func (g *FieldGenerator) GenerateFields(graph *model.Graph) error {
	// Implementation will be added later
	return nil
}