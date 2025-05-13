package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	// Set default for validation to true
	validateRelationships = true

	// Add a standard boolean flag for validation
	flag.BoolVar(&validateRelationships, "validate", true, "Validate relationships consistency in output CSV files")

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
	color.Cyan("Data volume: %d rows per entity", dataVolume)
	color.Cyan("Auto-cardinality: %t", autoCardinality)
	color.Cyan("Validate relationships: %t", validateRelationships)
	color.Cyan("==================")

	// Create a parser and parse the YAML file
	color.Yellow("Parsing YAML definition file...")
	parser := fabricator.NewParser(inputFile)
	err := parser.Parse()
	if err != nil {
		return fmt.Errorf("failed to parse YAML file: %w", err)
	}

	// Extract definition from parser
	def := parser.Definition

	// Display entity and relationship statistics
	printParsingStatistics(def)

	// Calculate estimated number of records
	totalRecords := len(def.Entities) * dataVolume
	color.Yellow("Estimated total CSV records to generate: %d", totalRecords)

	// Create output directory
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("failed to resolve output directory path: %w", err)
	}

	// Initialize CSV generator
	color.Yellow("Initializing CSV generator...")
	generator := generators.NewCSVGenerator(absOutputDir, dataVolume, autoCardinality)
	generator.Setup(def.Entities, def.Relationships)

	// Generate data
	color.Yellow("Generating data for %d entities...", len(def.Entities))
	generator.GenerateData()

	// Write CSV files
	color.Yellow("Writing CSV files to %s...", absOutputDir)
	err = generator.WriteCSVFiles()
	if err != nil {
		return fmt.Errorf("failed to write CSV files: %w", err)
	}

	// Validate relationships if requested
	if validateRelationships {
		color.Yellow("Validating relationship consistency in generated files...")

		// Validate relationship consistency
		validationResults := generator.ValidateRelationships()

		// Check if there are validation errors
		if len(validationResults) > 0 {
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
	printCompletionSummary(absOutputDir, def.Entities, dataVolume)

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
}

// printCompletionSummary displays the summary of generated files
func printCompletionSummary(outputDir string, entities map[string]models.Entity, volume int) {
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
	_, _ = successColor.Println("\n✓ CSV Generation Complete!")
	color.Green("  Output directory: %s", outputDir)
	color.Green("  CSV files generated: %d", csvFiles)
	color.Green("  Entities processed: %d", len(entities))
	color.Green("  Records per entity: %d", volume)
	color.Green("  Total records generated: %d", csvFiles*volume)

	color.Green("\nUse these CSV files to populate your system-of-record.\n")
}
