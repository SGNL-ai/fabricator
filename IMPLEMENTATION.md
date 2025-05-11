# Fabricator Implementation

This document describes the implementation of the Fabricator tool, a CSV generator for SGNL platform system-of-record data.

## Overview

Fabricator is a command-line tool that:
1. Takes a YAML file defining a SGNL platform system-of-record structure
2. Analyzes entity names, attributes, and relationships
3. Generates CSV files containing test data for each entity
4. Ensures relationship consistency between entities

## Project Structure

The project follows a standard Go project layout:

```
fabricator/
├── cmd/
│   └── fabricator/       # Command-line application code
│       ├── main.go       # Entry point for the application
│       └── main_test.go  # Tests for main package
├── pkg/
│   ├── fabricator/       # Core fabricator logic
│   │   ├── parser.go     # YAML parsing functionality
│   │   └── parser_test.go # Tests for parser
│   ├── generators/       # Data generation packages
│   │   └── csv_generator.go # CSV data generation
│   └── models/           # Data models
│       └── yaml.go       # YAML structure definitions
├── .github/              # GitHub workflow configuration
├── build/                # Build artifacts (generated)
├── example.yaml          # Example YAML definition
├── go.mod                # Go module definition
├── Makefile              # Build automation
├── README.md             # Project documentation
└── TODO.md               # Development tasks
```

## Key Components

### YAML Parser (`pkg/fabricator/parser.go`)

- Loads and parses YAML definition files
- Validates the structure of the YAML
- Extracts entity and relationship information
- Maps external IDs to CSV filenames

### CSV Generator (`pkg/generators/csv_generator.go`)

- Generates test data for each entity
- Ensures relationship consistency across entities
- Creates and writes CSV files
- Handles output directory creation

### Models (`pkg/models/yaml.go`)

- Defines the Go structs that match the YAML structure
- Includes entities, attributes, relationships
- Provides structure for CSV data generation

### Command-line Interface (`cmd/fabricator/main.go`)

- Handles command-line flags and user interaction
- Coordinates the overall process flow
- Provides colorful and informative output

## Data Flow

1. User provides a YAML file via the `-input` flag
2. The parser loads and validates the YAML file
3. The generator creates data for each entity
4. CSV files are written to the output directory
5. The user is provided with a summary of the generated files

## Future Enhancements

- Improved data generation that's more contextually relevant to attribute names
- Configuration file for customizing data generation
- Template-based data generation
- Validation of relationships in generated data
- Web UI for visualizing generated data
- Support for additional output formats (JSON, XML, etc.)

## Building and Running

```bash
# Build the project
make build

# Run the fabricator tool
./build/fabricator -input example.yaml -output data -volume 100
```