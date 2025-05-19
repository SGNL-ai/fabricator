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
   - Supports variable cardinality relationships (1:1, 1:N, N:1)
   - Auto-detects relationship cardinality using entity metadata
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
- `-a, --auto-cardinality`: Enable automatic cardinality detection for relationships
- `-v, --version`: Display version information

### Important Testing Notes

- Do NOT create YAML definition files yourself. All SOR YAML files come from another system and should be provided by the user.
- When testing changes, use only the example files that already exist in the project or those explicitly provided by the user.
- Never generate sample YAML files - they have very specific structure requirements and validation rules that must be met.

### Testing Approach

Our testing strategy follows these principles:

1. **Table-Driven Tests**: We use table-driven testing patterns to test multiple scenarios with minimal repetition. Each test case should be a clear, self-contained scenario.

2. **Test Fixtures**: Use the existing YAML fixtures in the examples directory to load realistic data for tests. Do not hardcode test data structures when possible.

3. **Testify Framework**: All tests should use the testify/assert and testify/require packages for consistent assertion styles.

4. **Test Isolation**: Each test should be independent and not rely on state from previous tests. Setup and teardown should be handled within each test function.

5. **Test Hierarchy**:
   - Unit tests: Test individual functions and methods in isolation
   - Component tests: Test interactions between related components
   - Integration tests: Test end-to-end flows from YAML to CSV output

6. **Legacy Tests**: The pkg/generators/legacy_tests directory contains older test implementations for reference. New tests should be created in the main package directories.

### Important Design Guidelines

1. **Field Name Analysis**:
   - Do NOT infer relationships from field names
   - Do NOT attempt to parse field names to determine if they are ID fields, primary keys, or foreign keys
   - Field names should ONLY be used to generate appropriate non-relationship sample data (e.g., generating an email address for a field called "email")
   - Relationships between entities must be explicitly defined through the relationship structures in the YAML definition
   - When relationships exist, they should be processed using the defined relationship links, not by analyzing field name patterns

## Workflow Guidance for Claude

### Getting Started

When first connecting to this project, you should:
1. Review the project structure to understand the codebase organization
2. Examine the main components and their responsibilities
3. Familiarize yourself with the existing test files and examples
4. Check the current implementation state through relevant source files

### Planning and TODO List Format

When planning implementation changes, present a structured TODO list with the following format:

```
# Implementation Plan for [Feature/Fix Name]

## File Changes Needed
1. File: `path/to/file1.go`
   - [ ] Add/modify X functionality
   - [ ] Update Y struct with new fields
   - [ ] Implement Z method

2. File: `path/to/file2.go`
   - [ ] Extend A interface
   - [ ] Fix B function parameters

## Testing Strategy
1. Unit tests
   - [ ] Create test for X functionality
   - [ ] Update existing tests for Y changes

2. Integration tests
   - [ ] Test end-to-end workflow with new changes

## CLI Changes
- [ ] Add new flag(s) to command line interface
- [ ] Update help documentation

## Documentation Updates
- [ ] Update README with new functionality
- [ ] Add examples of new feature usage
```

This structured approach helps with:
1. Clear organization of the implementation plan
2. Systematic tracking of changes across multiple files
3. Ensuring test coverage is maintained
4. Not overlooking documentation updates
5. Better communication about the implementation strategy

## Memories and Guidance

- *important*: Always follow law of Demeter: A design principle that states a method of an object should only invoke methods of objects that are closely related to it, promoting loose coupling and better modular design
- *important*: Always follow the DRY principle (don't repeat yourself): A software development principle aimed at reducing repetition of software patterns, encouraging code reuse and simplifying maintenance
- Act as an expert Go developer; write clean concise code with clear separation of duties.
- Do not try and cheat by hardcoding or bypassing the logical intent
- Explain your plan to me before trying to write code. Ask me to confirm or clarify details of the plan before suggesting file edits.
- when I say to write a stub test or stub logic that means to implement the final to define the func signature. Do not create logic

## Workflow Guidance
- We're practicing TDD. After each change to the code under test, rerun the tests to see if it's working. Make small incremental changes.
```

- **IMPORTANT GUIDELINE**: 
  - NEVER NEVER NEVER NEVER NEVER NEVER TRY TO FAKE TESTS!