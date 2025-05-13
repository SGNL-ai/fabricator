# Fabricator

[![CI](https://github.com/SGNL-ai/fabricator/actions/workflows/ci.yml/badge.svg)](https://github.com/SGNL-ai/fabricator/actions/workflows/ci.yml)

A command-line tool that generates CSV files for populating a system-of-record (SOR) based on a YAML definition file.

## Features

- Parses YAML files that define system-of-record (SOR) structures
- Analyzes entity names, attributes, and relationships
- Creates consistent test data across related entities
- Supports variable relationship cardinalities (1:1, 1:N, N:1)
- Automatically detects cardinality based on entity metadata
- Generates a set of CSV files with realistic test data
- Validates existing CSV files against a YAML definition
- Performs relationship consistency and uniqueness constraint checks
- Creates an SVG entity-relationship diagram visualization
- Outputs colorful and informative progress messages

## Installation

Requires Go 1.24.3 or higher.

### From Source

```bash
# Clone the repository
git clone https://github.com/SGNL-ai/fabricator.git
cd fabricator

# Build the project
make build
```

The binary will be built to `build/fabricator`.

### From GitHub Releases

Pre-built binaries for Linux, macOS (Intel and Apple Silicon), and Windows are automatically generated for each release and can be downloaded from the [GitHub Releases page](https://github.com/SGNL-ai/fabricator/releases).

```bash
# For macOS Intel
curl -L https://github.com/SGNL-ai/fabricator/releases/latest/download/fabricator-macos-intel -o fabricator
chmod +x fabricator
./fabricator --version

# For macOS Apple Silicon (M1/M2/M3)
curl -L https://github.com/SGNL-ai/fabricator/releases/latest/download/fabricator-macos-apple-silicon -o fabricator
chmod +x fabricator
./fabricator --version

# For Linux
curl -L https://github.com/SGNL-ai/fabricator/releases/latest/download/fabricator-linux -o fabricator
chmod +x fabricator
./fabricator --version
```

For Windows users, download the `fabricator-windows.exe` file from the releases page.

## Usage

```bash
# Basic usage (short options)
./build/fabricator -f <yaml-file> [-o <dir>] [-n <count>] [-a]

# Basic usage (long options)
./build/fabricator --file <yaml-file> [--output <dir>] [--num-rows <count>] [--auto-cardinality]

# View version information
./build/fabricator -v
```

### Command Line Options

| Short Flag | Long Flag            | Description                                      | Default   |
|------------|----------------------|--------------------------------------------------|-----------|
| `-f`       | `--file`             | Path to the YAML definition file (required)      | -         |
| `-o`       | `--output`           | Directory to store generated CSV files           | "output"  |
| `-n`       | `--num-rows`         | Number of rows to generate for each entity       | 100       |
| `-a`       | `--auto-cardinality` | Enable automatic cardinality detection           | false     |
| `-d`       | `--diagram`          | Generate Entity-Relationship diagram             | true      |
|            | `--validate`         | Validate relationships in CSV files              | true      |
|            | `--validate-only`    | Validate existing CSV files without generation   | false     |
| `-v`       | `--version`          | Display version information                      | -         |

### Examples

```bash
# Generate CSV files from example.yaml with 500 rows per entity
./build/fabricator -f example.yaml -n 500 -o data/sgnl

# Using long-form options
./build/fabricator --file example.yaml --num-rows 1000 --output export/data

# Generate CSV files with automatic cardinality detection for relationships
./build/fabricator -f example.yaml -n 200 -a

# Using long-form options with auto-cardinality
./build/fabricator --file example.yaml --num-rows 500 --auto-cardinality --output data/variable-cardinality

# Generate CSV files but disable ER diagram generation
./build/fabricator -f example.yaml --diagram=false

# Validate existing CSV files without generating new data
./build/fabricator -f example.yaml -o existing/csv/data --validate-only

# Validate existing CSV files and generate an ER diagram
./build/fabricator -f example.yaml -o existing/csv/data --validate-only --diagram
```

## YAML Format

The YAML file should define a system-of-record structure, including:

- Entities with attributes
- Relationships between entities
- External IDs that will be used for CSV filenames

Each entity in the YAML file will result in a corresponding CSV file, with the filename derived from the entity's `externalId`.

## Generated Data & Validation

The tool provides the following functionality:

1. CSV Generation:
   - CSV files named after each entity's external ID (without the namespace prefix)
   - Headers matching the entity's attribute external IDs
   - Consistent data across relationships between entities
   - Variable cardinality relationships (with the `-a` flag)
   - Realistic test data based on attribute names and types

2. CSV Validation (via `--validate-only`):
   - Checks existing CSV files against a YAML definition
   - Validates relationship consistency across entities
   - Verifies unique constraint requirements are met
   - Helpful for validating production or manually-created data exports
   - Use with the existing output directory containing CSV files

3. Entity-Relationship Diagram (enabled by default):
   - SVG visualization of all entities and their relationships
   - Color-coded entities with attributes listed
   - Primary keys (uniqueId attributes) highlighted
   - Relationship cardinality indicators (1:1, 1:N, N:1, N:M)
   - Can be disabled with `--diagram=false`
   - Works in both generation and validation-only modes

The data generator intelligently creates appropriate values based on field names:
- ID fields get unique identifiers
- Name fields get contextual names based on entity types (e.g., person names for users, company names for organizations)
- Date fields get properly formatted dates
- Email fields get valid email addresses
- Boolean fields get true/false values
- Numeric fields get appropriate numbers

## Relationship Cardinality

When the auto-cardinality feature is enabled (`-a` flag), Fabricator automatically detects and generates appropriate cardinality for entity relationships:

- **1:1 relationships** - Simple one-to-one mappings between entities
- **1:N relationships** - One entity related to multiple instances of another entity
- **N:1 relationships** - Multiple entities related to a single instance of another entity

Cardinality detection is based on:

1. Entity metadata (primary detection method)
   - Fields with `uniqueId: true` are used to identify key relationships
   - When a relationship links a unique ID to a non-unique field, cardinality is automatically determined

2. Field naming patterns (fallback method)
   - Field names ending with "Id" typically indicate N:1 relationships
   - Plural field names or names ending with "Ids" suggest 1:N relationships

Without the `-a` flag, all relationships default to 1:1 cardinality.

## Development

```bash
# Run tests
make test

# Format code
make fmt

# Run linter
make lint

# Run all checks (CI)
make ci
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.