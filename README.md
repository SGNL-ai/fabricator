# Fabricator

[![CI](https://github.com/SGNL-ai/fabricator/actions/workflows/ci.yml/badge.svg)](https://github.com/SGNL-ai/fabricator/actions/workflows/ci.yml)

A command-line tool that generates CSV files for populating a system-of-record (SOR) based on a YAML definition file.

## Features

- Parses YAML files that define system-of-record (SOR) structures
- Analyzes entity names, attributes, and relationships
- Creates consistent test data across related entities
- Generates a set of CSV files with realistic test data
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

You can also download pre-built binaries from the [GitHub Releases page](https://github.com/SGNL-ai/fabricator/releases) if available.

## Usage

```bash
# Basic usage (short options)
./build/fabricator -f <yaml-file> [-o <dir>] [-n <count>]

# Basic usage (long options)
./build/fabricator --file <yaml-file> [--output <dir>] [--num-rows <count>]

# View version information
./build/fabricator -v
```

### Command Line Options

| Short Flag | Long Flag   | Description                                      | Default   |
|------------|-------------|--------------------------------------------------|-----------|
| `-f`       | `--file`    | Path to the YAML definition file (required)      | -         |
| `-o`       | `--output`  | Directory to store generated CSV files           | "output"  |
| `-n`       | `--num-rows`| Number of rows to generate for each entity       | 100       |
| `-v`       | `--version` | Display version information                      | -         |

### Examples

```bash
# Generate CSV files from example.yaml with 500 rows per entity
./build/fabricator -f example.yaml -n 500 -o data/sgnl

# Using long-form options
./build/fabricator --file example.yaml --num-rows 1000 --output export/data
```

## YAML Format

The YAML file should define a system-of-record structure, including:

- Entities with attributes
- Relationships between entities
- External IDs that will be used for CSV filenames

Each entity in the YAML file will result in a corresponding CSV file, with the filename derived from the entity's `externalId`.

## Generated Data

The tool generates the following for each entity:

1. A CSV file named after the entity's external ID (without the namespace prefix)
2. Headers matching the entity's attribute external IDs
3. Consistent data across relationships between entities
4. Realistic test data based on attribute names and types

The data generator intelligently creates appropriate values based on field names:
- ID fields get unique identifiers
- Name fields get contextual names based on entity types (e.g., person names for users, company names for organizations)
- Date fields get properly formatted dates
- Email fields get valid email addresses
- Boolean fields get true/false values
- Numeric fields get appropriate numbers

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