package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SGNL-ai/fabricator/pkg/diagrams"
	"github.com/SGNL-ai/fabricator/pkg/fabricator"
	"github.com/SGNL-ai/fabricator/pkg/generators"
	"github.com/SGNL-ai/fabricator/pkg/models"
	"github.com/fatih/color"
)

// Version information (will be set during build)
var (
	version = "dev"
)

// Command line flags
var (
	// Show version
	showVersion bool

	// Input file
	inputFile string

	// Output directory
	outputDir string

	// Data volume
	dataVolume int

	// Auto-cardinality for relationships
	autoCardinality bool

	// Validate relationships
	validateRelationships bool

	// Generate ER diagram
	generateDiagram bool

	// Validation-only mode (skip CSV generation)
	validateOnly bool
)

func init() {
	// Define flags with both short and long forms
	flag.BoolVar(&showVersion, "v", false, "Display version information")
	flag.BoolVar(&showVersion, "version", false, "Display version information")

	flag.StringVar(&inputFile, "f", "", "Path to the YAML definition file (required)")
	flag.StringVar(&inputFile, "file", "", "Path to the YAML definition file (required)")

	flag.StringVar(&outputDir, "o", "output", "Directory to store generated CSV files")
	flag.StringVar(&outputDir, "output", "output", "Directory to store generated CSV files")

	flag.IntVar(&dataVolume, "n", 100, "Number of rows to generate for each entity")
	flag.IntVar(&dataVolume, "num-rows", 100, "Number of rows to generate for each entity")

	flag.BoolVar(&autoCardinality, "a", false, "Enable automatic cardinality detection for relationships")
	flag.BoolVar(&autoCardinality, "auto-cardinality", false, "Enable automatic cardinality detection for relationships")

	flag.BoolVar(&validateOnly, "validate-only", false, "Validate existing CSV files without generating new data")

	// Set default for validation to true
	validateRelationships = true

	// Add a standard boolean flag for validation
	flag.BoolVar(&validateRelationships, "validate", true, "Validate relationships consistency in output CSV files")

	// Check if GraphViz is available to determine default for diagram generation
	graphvizAvailable := diagrams.IsGraphvizAvailable()

	// Set default for diagram generation based on GraphViz availability
	generateDiagram = graphvizAvailable

	// Add a flag to control diagram generation with appropriate default message
	diagramDesc := "Generate Entity-Relationship diagram"
	if !graphvizAvailable {
		diagramDesc += " (default: false - Graphviz not found)"
	} else {
		diagramDesc += " (default: true)"
	}

	flag.BoolVar(&generateDiagram, "diagram", generateDiagram, diagramDesc)
	flag.BoolVar(&generateDiagram, "d", generateDiagram, diagramDesc)

	// Override default usage output
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		printUsage()
	}
}

func main() {
	// Parse command-line flags
	flag.Parse()

	// Display version information if requested
	if showVersion {
		fmt.Fprintf(os.Stderr, "Fabricator %s\n", version)
		os.Exit(0)
	}

	// Validate required flags
	if inputFile == "" {
		color.Red("Error: Input file is required. Use -f/--file flag to specify a YAML file.")
		fmt.Println("\nUsage:")
		printUsage()
		os.Exit(1)
	}

	// Main application logic
	if err := run(inputFile, outputDir, dataVolume, autoCardinality); err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}
}

// run performs the main application logic
func run(inputFile, outputDir string, dataVolume int, autoCardinality bool) error {
	// Print start message
	printHeader()
	color.Cyan("Input file: %s", inputFile)
	color.Cyan("Output directory: %s", outputDir)
	if !validateOnly {
		color.Cyan("Data volume: %d rows per entity", dataVolume)
		color.Cyan("Auto-cardinality: %t", autoCardinality)
	}
	color.Cyan("Validation-only mode: %t", validateOnly)
	color.Cyan("Validate relationships: %t", validateRelationships)
	color.Cyan("Generate ER diagram: %t", generateDiagram)
	color.Cyan("==================")

	// Create a parser and parse the YAML file
	color.Yellow("Parsing YAML definition file...")
	parser := fabricator.NewParser(inputFile)
	err := parser.Parse()
	if err != nil {
		// Extract details about relationship validation issues for better reporting
		if strings.Contains(err.Error(), "relationship issues") {
			// The full error message has detailed info, let's keep it
			return fmt.Errorf("failed to parse YAML file due to relationship validation issues:\n%w", err)
		}
		return fmt.Errorf("failed to parse YAML file: %w", err)
	}

	// Extract definition from parser
	def := parser.Definition

	// Display entity and relationship statistics
	printParsingStatistics(def)

	// Resolve output directory
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("failed to resolve output directory path: %w", err)
	}

	var generator *generators.CSVGenerator

	if !validateOnly {
		// Calculate estimated number of records
		totalRecords := len(def.Entities) * dataVolume
		color.Yellow("Estimated total CSV records to generate: %d", totalRecords)

		// Initialize CSV generator
		color.Yellow("Initializing CSV generator...")
		generator = generators.NewCSVGenerator(absOutputDir, dataVolume, autoCardinality)
		err = generator.Setup(def.Entities, def.Relationships)
			if err != nil {
				return fmt.Errorf("failed to setup CSV generator: %w", err)
			}

		// Generate data
		totalEntities := len(def.Entities)
		color.Yellow("Generating data for %d entities...", totalEntities)
		
		// For simplicity and avoid adding a callback function in this PR, we'll keep the current approach
		// In the future, we could add progress tracking for large entity sets
		err = generator.GenerateData()
		if err != nil {
			return fmt.Errorf("failed to generate data: %w", err)
		}

		// Write CSV files
		color.Yellow("Writing CSV files to %s...", absOutputDir)
		err = generator.WriteCSVFiles()
		if err != nil {
			return fmt.Errorf("failed to write CSV files: %w", err)
		}
	} else {
		// In validation-only mode, initialize without generating
		color.Yellow("Validation-only mode: Loading existing CSV files from %s...", absOutputDir)
		generator = generators.NewCSVGenerator(absOutputDir, dataVolume, autoCardinality)
		err = generator.Setup(def.Entities, def.Relationships)
			if err != nil {
				return fmt.Errorf("failed to setup CSV generator: %w", err)
			}

		// Load existing CSV files for validation
		err = generator.LoadExistingCSVFiles()
		if err != nil {
			return fmt.Errorf("failed to load existing CSV files: %w", err)
		}
	}

	// Generate ER diagram if requested
	if generateDiagram {
		// Check if Graphviz is available (this is a double-check since flag might be manually set)
		graphvizAvailable := diagrams.IsGraphvizAvailable()
		outputFormat := "DOT"
		extension := ".dot"

		if graphvizAvailable {
			outputFormat = "SVG"
			extension = ".svg"
		}

		color.Yellow("Generating Entity-Relationship diagram...")
		color.Cyan("  - Format: %s", outputFormat)

		// Create diagram filename based on SOR name
		diagramName := cleanNameForFilename(def.DisplayName)

		// Set diagram output path
		diagramPath := filepath.Join(absOutputDir, diagramName+extension)

		// Generate the diagram with panic recovery
		func() {
			defer func() {
				if r := recover(); r != nil {
					color.Red("Warning: ER diagram generation failed with panic: %v", r)
				}
			}()

			// Generate the diagram
			err := diagrams.GenerateERDiagram(def, diagramPath)
			if err != nil {
				color.Red("Warning: Could not generate ER diagram: %v", err)
			} else {
				color.Green("✓ Generated ER diagram at %s", diagramPath)

				// If Graphviz isn't available but diagram generation was requested, show a helpful message
				if !graphvizAvailable {
					color.Yellow("  Note: Generated DOT file only. To convert to SVG:")
					color.Yellow("  1. Install Graphviz (https://graphviz.org)")
					color.Yellow("  2. Run: dot -Tsvg %s -o %s",
						diagramPath,
						strings.TrimSuffix(diagramPath, extension)+".svg")
				}
			}
		}()
	}

	// Validate relationships if requested
	if validateRelationships {
		color.Yellow("Validating relationship consistency in generated files...")

		// Validate relationship consistency
		color.Cyan("  - Checking relationship consistency...")
		validationResults := generator.ValidateRelationships()

		// Check if there are validation errors
		validIssues := false
		// First pass to determine if there are any real issues to report
		for _, result := range validationResults {
			if result.InvalidRows > 0 {
				validIssues = true
				break
			}
		}
		
		if validIssues {
			color.Yellow("Found relationship consistency issues:")
			for _, result := range validationResults {
				if result.InvalidRows > 0 {
					color.Red("  • %s (%s) → %s (%s): %d invalid references out of %d rows",
						result.FromEntity, result.FromEntityFile,
						result.ToEntity, result.ToEntityFile,
						result.InvalidRows, result.TotalRows)

					// Show a limited number of detailed errors to avoid flooding the console
					maxErrorsToShow := 5
					errorsShown := 0
					for _, errMsg := range result.Errors {
						if errorsShown < maxErrorsToShow {
							color.Yellow("    - %s", errMsg)
							errorsShown++
						} else if errorsShown == maxErrorsToShow {
							color.Yellow("    - ... and %d more errors", len(result.Errors)-maxErrorsToShow)
							break
						}
					}
				}
			}
			color.Yellow("\nSome relationships have consistency issues. This might be expected with random data generation.")
		} else {
			color.Green("✓ All relationships are consistent across generated files!")
		}

		// Validate unique values
		color.Cyan("  - Checking uniqueness constraints...")
		uniqueValueErrors := generator.ValidateUniqueValues()
		if len(uniqueValueErrors) > 0 {
			color.Yellow("\nFound uniqueness constraint violations:")
			for _, entityError := range uniqueValueErrors {
				color.Red("  • Entity %s (%s):", entityError.EntityID, entityError.EntityFile)
				for _, errMsg := range entityError.Messages {
					color.Yellow("    - %s", errMsg)
				}
			}
			color.Yellow("\nSome unique attributes have duplicate values. This may cause issues in the SOR.")
		} else {
			color.Green("✓ All unique constraints are respected in generated files!")
		}
	}

	// Print completion message
	printCompletionSummary(absOutputDir, def, dataVolume, generateDiagram)

	return nil
}

// printHeader displays the app header
func printHeader() {
	headerColor := color.New(color.FgCyan, color.Bold)

	logo := `
  _____     _          _           _
 |  ___|_ _| |__  _ __(_) ___ __ _| |_ ___  _ __
 | |_ / _` + "`" + ` | '_ \| '__| |/ __/ _` + "`" + ` | __/ _ \| '__|
 |  _| (_| | |_) | |  | | (_| (_| | || (_) | |
 |_|  \__,_|_.__/|_|  |_|\___\__,_|\__\___/|_|
 CSV Generator %s
`

	// Print the logo with version information
	_, _ = headerColor.Printf(logo, version)
	fmt.Println() // Add an extra newline
}

// printParsingStatistics displays detailed statistics about the parsed YAML
func printParsingStatistics(def *models.SORDefinition) {
	// Count attributes
	totalAttributes := 0
	for _, entity := range def.Entities {
		totalAttributes += len(entity.Attributes)
	}

	// Find the namespace prefix pattern
	var namespacePrefix string
	var namespaceCount int
	for _, entity := range def.Entities {
		if strings.Contains(entity.ExternalId, "/") {
			parts := strings.Split(entity.ExternalId, "/")
			if len(parts) > 0 {
				prefix := parts[0]
				switch prefix {
				case "":
					// This shouldn't happen as we're checking for non-empty parts[0]
					continue
				case namespacePrefix:
					namespaceCount++
				default:
					if namespacePrefix == "" {
						namespacePrefix = prefix
						namespaceCount = 1
					}
				}
			}
		}
	}

	// Count direct vs path-based relationships
	directRelationships := 0
	pathRelationships := 0
	for _, rel := range def.Relationships {
		if len(rel.Path) > 0 {
			pathRelationships++
		} else {
			directRelationships++
		}
	}

	// Count attribute types
	uniqueIdAttributes := 0
	indexedAttributes := 0
	listAttributes := 0
	for _, entity := range def.Entities {
		for _, attr := range entity.Attributes {
			if attr.UniqueId {
				uniqueIdAttributes++
			}
			if attr.Indexed {
				indexedAttributes++
			}
			if attr.List {
				listAttributes++
			}
		}
	}

	// Print statistics
	color.Green("✓ Successfully parsed YAML definition")
	color.Cyan("  SOR name: %s", def.DisplayName)
	color.Cyan("  Description: %s", def.Description)

	if namespacePrefix != "" && namespaceCount > 0 {
		color.Cyan("  Namespace format detected: %s/EntityName (%d entities)",
			namespacePrefix, namespaceCount)
	}

	color.Cyan("  Entities: %d", len(def.Entities))
	color.Cyan("  Total attributes: %d", totalAttributes)
	color.Cyan("     - Unique ID attributes: %d", uniqueIdAttributes)
	color.Cyan("     - Indexed attributes: %d", indexedAttributes)
	color.Cyan("     - List attributes: %d", listAttributes)
	color.Cyan("  Relationships: %d total (%d direct, %d path-based)",
		len(def.Relationships), directRelationships, pathRelationships)
}

// printUsage displays the usage information with proper double-dash syntax for long options
func printUsage() {
	fmt.Println("  -v, --version\n\tDisplay version information")
	fmt.Println("  -f, --file string\n\tPath to the YAML definition file (required)")
	fmt.Println("  -o, --output string\n\tDirectory to store generated CSV files (default \"output\")")
	fmt.Println("  -n, --num-rows int\n\tNumber of rows to generate for each entity (default 100)")
	fmt.Println("  -a, --auto-cardinality\n\tEnable automatic cardinality detection for relationships")
	fmt.Println("  --validate\n\tValidate relationships consistency in output CSV files (default true)")
	fmt.Println("  --validate-only\n\tValidate existing CSV files without generating new data")

	// Build diagram flag description with dynamic default based on Graphviz availability
	diagDesc := "Generate Entity-Relationship diagram"
	if diagrams.IsGraphvizAvailable() {
		diagDesc += " (default true)"
	} else {
		diagDesc += " (default false - Graphviz not found)"
	}
	fmt.Println("  -d, --diagram\n\t" + diagDesc)
}

// printCompletionSummary displays the summary of generated files
func printCompletionSummary(outputDir string, def *models.SORDefinition, volume int, diagramGenerated bool) {
	// Extract entities for backward compatibility
	entities := def.Entities
	// List generated files
	files, err := os.ReadDir(outputDir)
	if err != nil {
		color.Yellow("Warning: Could not list generated files: %v", err)
		return
	}

	// Count CSV files
	csvFiles := 0
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".csv" {
			csvFiles++
		}
	}

	// Print summary
	successColor := color.New(color.FgGreen, color.Bold)

	if validateOnly {
		_, _ = successColor.Println("\n✓ Validation Complete!")
		color.Green("  Input directory: %s", outputDir)
		color.Green("  CSV files validated: %d", csvFiles)
		color.Green("  Entities in definition: %d", len(entities))
	} else {
		_, _ = successColor.Println("\n✓ CSV Generation Complete!")
		color.Green("  Output directory: %s", outputDir)
		color.Green("  CSV files generated: %d", csvFiles)
		color.Green("  Entities processed: %d", len(entities))
		color.Green("  Records per entity: %d", volume)
		color.Green("  Total records generated: %d", csvFiles*volume)
	}

	if diagramGenerated {
		// Get a clean filename from the SOR display name
		diagramName := cleanNameForFilename(def.DisplayName)

		// Look for specific diagram files based on SOR name
		dotFile := filepath.Join(outputDir, diagramName+".dot")
		svgFile := filepath.Join(outputDir, diagramName+".svg")

		// Also check for default names in case they were generated earlier
		defaultDotFile := filepath.Join(outputDir, "entity_relationship_diagram.dot")
		defaultSvgFile := filepath.Join(outputDir, "entity_relationship_diagram.svg")

		// Check if the files exist and print them in output
		if _, err := os.Stat(dotFile); err == nil {
			color.Green("  Entity-Relationship diagram (DOT): %s", dotFile)
		} else if _, err := os.Stat(defaultDotFile); err == nil {
			color.Green("  Entity-Relationship diagram (DOT): %s", defaultDotFile)
		}

		if _, err := os.Stat(svgFile); err == nil {
			color.Green("  Entity-Relationship diagram (SVG): %s", svgFile)
		} else if _, err := os.Stat(defaultSvgFile); err == nil {
			color.Green("  Entity-Relationship diagram (SVG): %s", defaultSvgFile)
		}
	}

	if validateOnly {
		color.Green("\nValidation of existing CSV files complete.\n")
	} else {
		color.Green("\nUse these CSV files to populate your system-of-record.\n")
	}
}

// cleanNameForFilename creates a filesystem-safe name from a display name
func cleanNameForFilename(name string) string {
	// Replace spaces and slashes with underscores
	cleaned := strings.ReplaceAll(name, " ", "_")
	cleaned = strings.ReplaceAll(cleaned, "/", "_")
	// If the name is empty, use a default
	if cleaned == "" {
		cleaned = "entity_relationship_diagram"
	}
	return cleaned
}
