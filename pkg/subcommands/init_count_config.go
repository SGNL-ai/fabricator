package subcommands

import (
	"fmt"
	"io"
	"os"

	"github.com/SGNL-ai/fabricator/pkg/config"
	"github.com/SGNL-ai/fabricator/pkg/parser"
	"github.com/fatih/color"
)

// InitCountConfigOptions holds the options for the init-count-config subcommand
type InitCountConfigOptions struct {
	// SORFile is the path to the SOR YAML definition file
	SORFile string

	// DefaultCount is the default row count to use for all entities
	DefaultCount int

	// Output is where to write the template (defaults to stdout)
	Output io.Writer
}

// InitCountConfig generates a row count configuration template from a SOR YAML file.
// It writes the template to the specified output (typically stdout).
func InitCountConfig(opts InitCountConfigOptions) error {
	// Validate input file
	if opts.SORFile == "" {
		return fmt.Errorf("SOR file path is required")
	}

	// Check if file exists
	if _, err := os.Stat(opts.SORFile); os.IsNotExist(err) {
		return fmt.Errorf("SOR file not found: %s", opts.SORFile)
	}

	// Set defaults
	if opts.DefaultCount == 0 {
		opts.DefaultCount = 100
	}
	if opts.Output == nil {
		opts.Output = os.Stdout
	}

	// Parse the SOR YAML file
	p := parser.NewParser(opts.SORFile)
	if err := p.Parse(); err != nil {
		return fmt.Errorf("failed to parse SOR file: %w", err)
	}

	// Extract entities from the parsed definition
	entities := make([]config.TemplateEntity, 0, len(p.Definition.Entities))
	for _, entity := range p.Definition.Entities {
		entities = append(entities, config.TemplateEntity{
			ExternalID:   entity.ExternalId,
			DisplayName:  entity.DisplayName,
			Description:  entity.Description,
			DefaultCount: opts.DefaultCount,
		})
	}

	// Create the template
	template := config.NewTemplate(opts.SORFile, entities, opts.DefaultCount)

	// Write the template to output
	if err := template.WriteToWriter(opts.Output); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	// Write a success message to stderr (so it doesn't interfere with the YAML output)
	_, _ = color.New(color.FgGreen).Fprintf(os.Stderr, "✓ Generated row count configuration template with %d entities\n", len(entities))
	_, _ = color.New(color.FgCyan).Fprintf(os.Stderr, "  Redirect output to a file: fabricator init-count-config -f %s > counts.yaml\n", opts.SORFile)

	return nil
}
