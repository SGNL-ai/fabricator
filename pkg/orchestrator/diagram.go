package orchestrator

import (
	"path/filepath"

	"github.com/SGNL-ai/fabricator/pkg/diagrams"
	"github.com/SGNL-ai/fabricator/pkg/parser"
	"github.com/SGNL-ai/fabricator/pkg/util"
)

// DiagramOptions configures diagram generation
type DiagramOptions struct {
	// Additional options can be added here
}

// DiagramResult contains the results of diagram generation
type DiagramResult struct {
	Generated bool
	Path      string
}

// RunDiagramGeneration orchestrates ER diagram generation
func RunDiagramGeneration(def *parser.SORDefinition, outputDir string, options DiagramOptions) (*DiagramResult, error) {
	result := &DiagramResult{}

	// Create diagram filename based on SOR name
	diagramName := util.CleanNameForFilename(def.DisplayName)

	// Determine extension based on Graphviz availability
	extension := ".dot"
	if diagrams.IsGraphvizAvailable() {
		extension = ".svg"
	}

	diagramPath := filepath.Join(outputDir, diagramName+extension)

	// Generate the diagram
	err := diagrams.GenerateERDiagram(def, diagramPath)
	if err != nil {
		return result, err
	}

	result.Generated = true
	result.Path = diagramPath
	return result, nil
}
