package pipeline

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SGNL-ai/fabricator/pkg/generators/model"
	"github.com/fatih/color"
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
	// Create the output directory if it doesn't exist
	err := os.MkdirAll(w.outputDir, 0750)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write each entity's data to a CSV file
	for _, entity := range graph.GetAllEntities() {
		csvData := entity.ToCSV()

		// Get the filename based on the entity's external ID
		filename := w.getEntityFileName(csvData.ExternalId)
		filePath := filepath.Join(w.outputDir, filename)

		file, err := os.Create(filepath.Clean(filePath))
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", filePath, err)
		}
		defer func() { _ = file.Close() }()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		// Write headers
		err = writer.Write(csvData.Headers)
		if err != nil {
			return fmt.Errorf("failed to write headers to %s: %w", filePath, err)
		}

		// Write data rows
		for _, row := range csvData.Rows {
			err = writer.Write(row)
			if err != nil {
				return fmt.Errorf("failed to write row to %s: %w", filePath, err)
			}
		}

		// Clear progress line and show completion message
		fmt.Printf("\r%50s\r", "") // Clear line with spaces, then return to start
		color.Green("✓ Generated %s with %d rows", filename, len(csvData.Rows))
	}

	return nil
}

// getEntityFileName extracts filename from external ID
func (w *CSVWriter) getEntityFileName(externalID string) string {
	// Handle both formats: with namespace prefix (e.g., "KeystoneV1/Entity") and without
	if len(externalID) == 0 {
		return "unknown.csv"
	}

	// If there's a slash, take the part after the last slash
	if lastSlash := len(externalID) - 1; lastSlash >= 0 {
		for i := lastSlash; i >= 0; i-- {
			if externalID[i] == '/' {
				return externalID[i+1:] + ".csv"
			}
		}
	}

	// No slash found, use the whole external ID
	return externalID + ".csv"
}
