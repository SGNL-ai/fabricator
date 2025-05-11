# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

"Fabricator" is a Go command-line utility that generates CSV files for populating a system-of-record (SOR) based on a YAML definition file. It analyzes entity relationships, and generates consistent test data across CSV files.

## Project Goals

1. Parse YAML files defining system-of-record structures
2. Generate CSV files where each filename (without .csv) matches the entity's "external_id" from the YAML
3. Maintain relationship consistency between entities based on the YAML definition
4. Generate plausible test data for each entity attribute
5. Allow users to specify the amount of test data to create
6. Provide a colorful and informative CLI experience

## Development Commands

### Building the Project
```bash
# Build the project (creates binary in build/ directory)
make build
```

### Testing
```bash
# Run all tests
make test

# Run tests in a specific package
go test ./pkg/<package_name>
```

### Linting and Code Quality
```bash
# Format code
make fmt

# Run static code analysis
make vet

# Run linter
make lint

# Run all checks (format, vet, lint, test, build)
make ci
```

### Cleaning
```bash
# Remove build artifacts
make clean
```

## Development Workflow

### Branching Model

This project follows a simple main-based branching model:

1. **main** - Production-ready code; protected from direct pushes
2. **feature/** - Branches for new functionality or enhancements
3. **bugfix/** - Branches for fixing bugs

This model keeps things straightforward while maintaining good development practices.

### Workflow Guidelines

1. Branch from `main` for new work:
   ```bash
   git checkout main
   git pull
   # For new features
   git checkout -b feature/your-feature-name
   # OR for bug fixes
   git checkout -b bugfix/issue-description
   ```

2. Make changes, commit, and push to GitHub:
   ```bash
   git add .
   git commit -m "Descriptive commit message"
   git push -u origin your-branch-name
   ```

3. Create a Pull Request to merge back to `main`
4. After review and approval, merge the PR

### Dependency Management

This project uses Dependabot to keep dependencies up to date. The configuration is pragmatic:

1. **Go Dependencies**:
   - Updates checked weekly (every Monday)
   - Minor and patch updates are grouped into a single PR
   - Major version updates are ignored to avoid breaking changes
   - PRs are limited to 5 open at a time
   - All PRs target the `main` branch

2. **GitHub Actions**:
   - Updates checked monthly
   - All updates are grouped into a single PR
   - PRs are limited to 3 open at a time
   - All PRs target the `main` branch

This approach balances keeping dependencies current while minimizing maintenance overhead and avoiding potential breaking changes from major version updates.

### Security Best Practices

The project implements security best practices:

1. **Token Permissions**:
   - GitHub token permissions follow the principle of least privilege
   - CI workflow has explicitly limited permissions to only what's needed
   - Dependabot has access limited to required operations

2. **Branch Protection**:
   - Main branch is protected from direct pushes
   - Changes must go through pull requests
   - CI checks must pass before merging

# Project Architecture

The project follows a standard Go project layout:

### Main Components

1. **Command Line Interface** (`cmd/fabricator/main.go`):
   - Handles command-line flag parsing
   - Coordinates the overall process flow
   - Manages stderr/stdout separation
   - Provides colorful informative output

2. **YAML Parser** (`pkg/fabricator/parser.go`):
   - Loads and parses YAML definition files
   - Validates the structure of the YAML
   - Extracts entity and relationship information

3. **CSV Generator** (`pkg/generators/csv_generator.go`):
   - Generates test data for each entity
   - Ensures relationship consistency across entities
   - Creates and writes CSV files to the output directory

4. **Data Models** (`pkg/models/yaml.go`):
   - Defines Go structs matching the YAML structure
   - Includes entities, attributes, relationships
   - Provides structure for CSV data generation

### Configuration

The application accepts the following command-line flags:
- `-f, --file`: Path to the YAML definition file (required)
- `-o, --output`: Directory to store generated CSV files (default: "output")
- `-n, --num-rows`: Number of rows to generate for each entity (default: 100)
- `-v, --version`: Display version information

### Important Testing Notes

- Do NOT create YAML definition files yourself. All SOR YAML files come from another system and should be provided by the user.
- When testing changes, use only the example files that already exist in the project or those explicitly provided by the user.
- Never generate sample YAML files - they have very specific structure requirements and validation rules that must be met.