package pipeline

import (
	"github.com/SGNL-ai/fabricator/pkg/generators/model"
)

// CSVWriter handles writing entity data to CSV files
type CSVWriter struct {
	outputDir string
}

// NewCSVWriter creates a new CSV writer
func NewCSVWriter(outputDir string) CSVWriterInterface {
	return &CSVWriter{
		outputDir: outputDir,
	}
}

// WriteFiles writes all entity data to CSV files
func (w *CSVWriter) WriteFiles(graph *model.Graph) error {
	// Implementation will be added later
	return nil
}
