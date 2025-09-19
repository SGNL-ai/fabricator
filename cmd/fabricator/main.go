package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/SGNL-ai/fabricator/pkg/diagrams"
	"github.com/SGNL-ai/fabricator/pkg/orchestrator"
	"github.com/SGNL-ai/fabricator/pkg/parser"
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

	// Profiling options
	cpuProfile string
	memProfile string
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

	// Set default for diagram generation based on GraphViz availability
	generateDiagram = diagrams.IsGraphvizAvailable()

	// Add a flag to control diagram generation with appropriate default message
	diagramDesc := "Generate Entity-Relationship diagram"
	if generateDiagram {
		diagramDesc += " (default: false - Graphviz not found)"
	} else {
		diagramDesc += " (default: true)"
	}

	flag.BoolVar(&generateDiagram, "diagram", generateDiagram, diagramDesc)
	flag.BoolVar(&generateDiagram, "d", generateDiagram, diagramDesc)

	// Add profiling flags
	flag.StringVar(&cpuProfile, "cpuprofile", "", "Write CPU profile to file")
	flag.StringVar(&memProfile, "memprofile", "", "Write memory profile to file")

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
	// Start profiling if requested
	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			return fmt.Errorf("could not create CPU profile: %w", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			return fmt.Errorf("could not start CPU profile: %w", err)
		}
		defer pprof.StopCPUProfile()
		color.Yellow("CPU profiling enabled: %s", cpuProfile)
	}

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
	if cpuProfile != "" {
		color.Cyan("CPU profiling: %s", cpuProfile)
	}
	if memProfile != "" {
		color.Cyan("Memory profiling: %s", memProfile)
	}
	color.Cyan("==================")

	// Create a parser and parse the YAML file
	color.Yellow("Parsing YAML definition file...")
	parser := parser.NewParser(inputFile)
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

	// Resolve output directory
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("failed to resolve output directory path: %w", err)
	}

	if !validateOnly {
		// Generation mode
		err := runGenerationMode(def, absOutputDir, dataVolume, autoCardinality)
		if err != nil {
			return err
		}
	} else {
		// Validation-only mode
		err := runValidationMode(def, absOutputDir)
		if err != nil {
			return err
		}
	}

	// Generate ER diagram (common to both modes)
	if generateDiagram {
		diagramResult, err := orchestrator.RunDiagramGeneration(def, absOutputDir, orchestrator.DiagramOptions{})
		if err == nil && diagramResult.Generated {
			color.Green("✓ Generated ER diagram at %s", diagramResult.Path)
		}
	}

	// Write memory profile if requested
	if memProfile != "" {
		f, err := os.Create(memProfile)
		if err != nil {
			return fmt.Errorf("could not create memory profile: %w", err)
		}
		defer f.Close()
		runtime.GC() // Get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			return fmt.Errorf("could not write memory profile: %w", err)
		}
		color.Yellow("Memory profile written: %s", memProfile)
	}

	return nil
}

// runGenerationMode handles data generation workflow
func runGenerationMode(def *parser.SORDefinition, outputDir string, dataVolume int, autoCardinality bool) error {
	// Calculate estimated number of records
	totalRecords := len(def.Entities) * dataVolume
	color.Yellow("Estimated total CSV records to generate: %d", totalRecords)

	// Generate data using orchestrator
	totalEntities := len(def.Entities)
	color.Yellow("Generating data for %d entities...", totalEntities)
	color.Yellow("Writing CSV files to %s...", outputDir)

	options := orchestrator.GenerationOptions{
		DataVolume:      dataVolume,
		AutoCardinality: autoCardinality,
		GenerateDiagram: generateDiagram,
		ValidateResults: validateRelationships,
	}

	result, err := orchestrator.RunGeneration(def, outputDir, options)
	if err != nil {
		return fmt.Errorf("failed to generate CSV data: %w", err)
	}

	// Handle validation results
	if validateRelationships && result.ValidationSummary != nil {
		if len(result.ValidationSummary.Errors) > 0 {
			color.Yellow("Found relationship consistency issues:")
			for _, errMsg := range result.ValidationSummary.Errors {
				color.Red("  • %s", errMsg)
			}
			color.Yellow("\nSome relationships have consistency issues. This might be expected with random data generation.")
		} else {
			color.Green("✓ All relationships are consistent across generated files!")
		}
		color.Green("✓ All unique constraints are respected in generated files!")
	}

	// Print completion summary
	printGenerationSummary(outputDir, result, generateDiagram)
	return nil
}

// runValidationMode handles validation-only workflow
func runValidationMode(def *parser.SORDefinition, outputDir string) error {
	color.Yellow("Validation-only mode: Loading and validating existing CSV files from %s...", outputDir)

	options := orchestrator.ValidationOptions{
		GenerateDiagram: generateDiagram,
	}

	result, err := orchestrator.RunValidation(def, outputDir, options)
	if err != nil {
		return fmt.Errorf("validation-only mode failed: %w", err)
	}

	// Report validation results
	if len(result.ValidationErrors) > 0 {
		color.Yellow("Found %d validation issues:", len(result.ValidationErrors))
		for _, errMsg := range result.ValidationErrors {
			color.Red("  • %s", errMsg)
		}
		color.Yellow("\nSome relationships have consistency issues. This might be expected with existing data.")
	} else {
		color.Green("✓ All CSV files validated successfully - no issues found!")
	}

	// Print validation summary
	printValidationSummary(outputDir, result, generateDiagram)
	return nil
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
	fmt.Println("  --cpuprofile string\n\tWrite CPU profile to file")
	fmt.Println("  --memprofile string\n\tWrite memory profile to file")

	// Build diagram flag description with dynamic default based on Graphviz availability
	diagDesc := "Generate Entity-Relationship diagram"
	if diagrams.IsGraphvizAvailable() {
		diagDesc += " (default true)"
	} else {
		diagDesc += " (default false - Graphviz not found)"
	}
	fmt.Println("  -d, --diagram\n\t" + diagDesc)
}

// SummaryInfo holds common information needed for printing operation summaries
type SummaryInfo struct {
	Title            string
	DirectoryLabel   string
	Directory        string
	DiagramGenerated bool
	DiagramPath      string
	FinalMessage     string
}

// printOperationSummary displays a unified operation completion summary
func printOperationSummary(info SummaryInfo, diagramEnabled bool, printMetrics func()) {
	successColor := color.New(color.FgGreen, color.Bold)
	_, _ = successColor.Printf("\n✓ %s!\n", info.Title)
	color.Green("  %s: %s", info.DirectoryLabel, info.Directory)

	// Print operation-specific metrics
	printMetrics()

	// Print diagram information if enabled
	if diagramEnabled && info.DiagramGenerated {
		color.Green("  Entity-Relationship diagram (SVG): %s", info.DiagramPath)
	} else if diagramEnabled {
		color.Green("  Entity-Relationship diagram (DOT): %s", info.DiagramPath)
	}

	color.Yellow("\n%s", info.FinalMessage)
}

// printGenerationSummary displays the generation completion summary
func printGenerationSummary(outputDir string, result *orchestrator.GenerationResult, diagramGenerated bool) {
	info := SummaryInfo{
		Title:            "CSV Generation Complete",
		DirectoryLabel:   "Output directory",
		Directory:        outputDir,
		DiagramGenerated: result.DiagramGenerated,
		DiagramPath:      result.DiagramPath,
		FinalMessage:     "Use these CSV files to populate your system-of-record.",
	}

	printOperationSummary(info, diagramGenerated, func() {
		color.Green("  CSV files generated: %d", result.CSVFilesGenerated)
		color.Green("  Entities processed: %d", result.EntitiesProcessed)
		color.Green("  Records per entity: %d", result.RecordsPerEntity)
		color.Green("  Total records generated: %d", result.TotalRecords)
	})
}

// printValidationSummary displays the validation completion summary
func printValidationSummary(outputDir string, result *orchestrator.ValidationResult, diagramGenerated bool) {
	info := SummaryInfo{
		Title:            "Validation Complete",
		DirectoryLabel:   "Input directory",
		Directory:        outputDir,
		DiagramGenerated: result.DiagramGenerated,
		DiagramPath:      result.DiagramPath,
		FinalMessage:     "Validation of existing CSV files complete.",
	}

	printOperationSummary(info, diagramGenerated, func() {
		color.Green("  CSV files validated: %d", result.FilesValidated)
		color.Green("  Records validated: %d", result.RecordsValidated)

		if len(result.ValidationErrors) > 0 {
			color.Green("  Validation issues found: %d", len(result.ValidationErrors))
		}
	})
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
