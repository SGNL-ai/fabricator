# Fabricator

[![CI](https://github.com/SGNL-ai/fabricator/actions/workflows/ci.yml/badge.svg)](https://github.com/SGNL-ai/fabricator/actions/workflows/ci.yml)
[![Security](https://github.com/SGNL-ai/fabricator/actions/workflows/security.yml/badge.svg)](https://github.com/SGNL-ai/fabricator/actions/workflows/security.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/SGNL-ai/fabricator)](https://goreportcard.com/report/github.com/SGNL-ai/fabricator)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/github/go-mod/go-version/SGNL-ai/fabricator)](https://golang.org/)

> **A modern, enterprise-grade CSV data generator for system-of-record testing**

Fabricator is a powerful command-line tool that generates realistic CSV test data for system-of-record (SOR) platforms. Built with a robust pipeline architecture and comprehensive validation, it transforms YAML definitions into consistent, relationship-aware CSV datasets.

## 🚀 Quick Start

```bash
# Download the latest release for your platform
curl -L https://github.com/SGNL-ai/fabricator/releases/latest/download/fabricator-linux -o fabricator
chmod +x fabricator

# Generate test data from a YAML definition
./fabricator -f examples/sample.yaml -n 1000 -o ./test-data

# Output:
# ✓ Generated 16 CSV files with 1000 rows each
# ✓ All relationships consistent across files
# ✓ Entity-relationship diagram created
```

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   YAML Input    │───▶│   Validation     │───▶│   Pipeline      │
│                 │    │                  │    │                 │
│ • Entities      │    │ • JSON Schema    │    │ • Phase 1: IDs  │
│ • Attributes    │    │ • Business Logic │    │ • Phase 2: Rels │
│ • Relationships │    │ • 96% Template   │    │ • Phase 3: Data │
└─────────────────┘    │   Compatibility  │    └─────────────────┘
                       └──────────────────┘             │
                                                        ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   CSV Output    │◀───│    Validation    │◀───│  Data Model     │
│                 │    │                  │    │                 │
│ • Multi-file    │    │ • Referential    │    │ • Graph         │
│ • Consistent    │    │   Integrity      │    │ • Entities      │
│ • Realistic     │    │ • Uniqueness     │    │ • Relationships │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## ✨ Features

### 🏗️ **Robust Architecture**
- **Pipeline-based processing** with clean separation of concerns
- **Comprehensive validation** with JSON Schema + business logic layers
- **Graph-based dependency resolution** with topological sorting

### 📊 **Data Generation**
- **Realistic test data** with type-aware field generation
- **Relationship consistency** across all CSV files
- **Variable cardinalities** (1:1, 1:N, N:1, N:N) with auto-detection
- **Configurable data volume** from small samples to large datasets

### 🔍 **Validation & Quality**
- **YAML schema validation** using industry-standard JSON Schema
- **Relationship integrity** checking across entities
- **Uniqueness constraint** validation
- **Production template compatibility** (96% of SGNL catalog templates supported)

### 🎨 **User Experience**
- **Colorful CLI output** with progress indicators
- **SVG diagram generation** for entity-relationship visualization
- **Detailed error messages** with actionable guidance
- **Multiple operation modes** (generate, validate-only, diagram-only)

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

## 📈 Performance

Fabricator is designed for efficiency and can handle large datasets:

| Dataset Size | Entities | Time     | Memory   |
|-------------|----------|----------|----------|
| Small       | 5        | <1s      | <50MB    |
| Medium      | 16       | 2-5s     | <100MB   |
| Large       | 50       | 10-30s   | <500MB   |

**Benchmarks** (16 entities, complex relationships):
- **1,000 rows/entity**: ~3 seconds, 16 CSV files
- **10,000 rows/entity**: ~15 seconds, consistent relationships
- **100,000 rows/entity**: ~2 minutes, 1.6M total records

## 🛠️ Development

### Prerequisites for Development

- Go 1.23+ (tested with 1.23 and 1.24)
- golangci-lint for code quality
- Pre-commit hooks (optional but recommended)

### Development Commands

```bash
# Run tests
make test

# Run tests with coverage
make coverage

# Format code
make fmt

# Static analysis
make vet

# Run linter
make lint

# Run all checks (CI pipeline)
make ci

# Security scanning
gosec ./...
govulncheck ./...
```

### Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed development guidelines, architecture documentation, and contribution workflow.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.