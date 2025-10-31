# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

"Fabricator" is a Go command-line utility that generates CSV files for populating a system-of-record (SOR) based on a YAML definition file. It analyzes entity relationships and generates consistent test data across CSV files.

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

### Dependency Management

This project uses Dependabot to keep dependencies up to date. The configuration is pragmatic:

1. **Go Dependencies**:
   - Updates checked weekly (every Monday)
   - Minor and patch updates are grouped into a single PR
   - Major version updates are ignored to avoid breaking changes

2. **GitHub Actions**:
   - Updates checked monthly
   - All updates are grouped into a single PR

### Security Best Practices

The project implements security best practices:

1. **Token Permissions**:
   - GitHub token permissions follow the principle of least privilege
   - CI workflow has explicitly limited permissions to only what's needed

2. **Branch Protection**:
   - Main branch is protected from direct pushes
   - Changes must go through pull requests
   - CI checks must pass before merging

## Project Architecture

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

### YAML Structure

The YAML files follow a specific structure that defines:

1. **Entities**: Each entity represents a data object (like "user", "group", "permission")
   - Has a unique `external_id` that becomes the CSV filename
   - Contains attributes that become CSV columns
   - May reference other entities through relationships

2. **Attributes**: Define the properties of each entity
   - Have types (string, number, boolean, etc.)
   - May have format constraints (email, URL, date, etc.)
   - Some may be marked as primary keys or required fields

3. **Relationships**: Define connections between entities
   - Can be one-to-one, one-to-many, or many-to-many
   - Specify source and target entities and attributes
   - May include cardinality information

Example YAML structure (simplified):
```yaml
entities:
  - external_id: users
    attributes:
      - name: id
        type: string
        primary_key: true
      - name: email
        type: string
        format: email
    relationships:
      - name: user_groups
        target: groups
        source_attribute: id
        target_attribute: user_id
        cardinality: one_to_many
```

### Configuration

The application accepts the following command-line flags:
- `-f, --file`: Path to the YAML definition file (required)
- `-o, --output`: Directory to store generated CSV files (default: "output")
- `-n, --num-rows`: Number of rows to generate for each entity (default: 100)
- `-a, --auto-cardinality`: Enable automatic cardinality detection for relationships
- `-v, --version`: Display version information

## Implementation Guidelines

### Data Generation Strategy

When generating test data, follow these guidelines:

1. **Type-Appropriate Data**: Generate data that matches the attribute type
   - Strings: Use appropriate patterns (names, emails, etc.)
   - Numbers: Generate within reasonable ranges
   - Dates: Use realistic date ranges
   - Booleans: Generate with appropriate distributions

2. **Format-Aware Generation**: Respect any format constraints
   - Email: Generate plausible email addresses
   - URL: Create valid URLs
   - Phone: Follow standard phone number formats

3. **Relationship Consistency**: When entities are related:
   - Generate IDs that maintain referential integrity
   - Ensure foreign keys reference valid primary keys
   - Respect cardinality constraints (1:1, 1:N, N:1, N:M)

4. **Primary Key Handling**:
   - Ensure primary keys are unique within an entity
   - Use appropriate data types (UUIDs, sequential IDs, etc.)
   - Consider composite keys if defined in the YAML

### Error Handling

Follow these guidelines for error handling:

1. **Early Validation**: Validate inputs before processing
   - Check YAML structure validity before generating data
   - Validate command-line arguments before starting work
   - Ensure output directory exists and is writable

2. **Graceful Failures**: When errors occur:
   - Provide clear, actionable error messages
   - Include context about what was being processed
   - Suggest potential solutions when possible
   - Use appropriate exit codes

3. **Logging Levels**: Use different log levels appropriately
   - Error: For problems that prevent execution
   - Warning: For non-fatal issues that might affect output
   - Info: For normal operation information
   - Debug: For detailed troubleshooting information

4. **Be Explicit**: Never silently fail or ignore errors
   - Always check error return values
   - Return early rather than continuing with invalid state
   - Log the specific error, not just that "an error occurred"

### Code Style Guidelines

1. **Package Organization**:
   - Package names should be single, lowercase words
   - One package per directory
   - Package name should match directory name
   - Keep package interfaces small and focused

2. **Function Design**:
   - Functions should do one thing well
   - Keep functions under 50 lines when possible
   - Return errors rather than using panic
   - Use named return parameters for complex returns

3. **Variable Naming**:
   - Use camelCase for internal variables
   - Use PascalCase for exported identifiers
   - Use short names for loop variables (i, j, k)
   - Use descriptive names for functions and methods

4. **Comments**:
   - All exported functions must have doc comments
   - Comments should explain "why", not "what"
   - Use comments to document non-obvious behavior
   - Follow GoDoc formatting conventions

5. **Error Handling**:
   - Check all error returns
   - Return errors up the call stack rather than handling locally when appropriate
   - Use fmt.Errorf with %w for wrapping errors
   - Create custom error types for domain-specific errors

### Important Design Guidelines

1. **Field Name Analysis**:
   - Do NOT infer relationships from field names
   - Do NOT attempt to parse field names to determine if they are ID fields, primary keys, or foreign keys
   - Field names should ONLY be used to generate appropriate non-relationship sample data (e.g., generating an email address for a field called "email")
   - Relationships between entities must be explicitly defined through the relationship structures in the YAML definition

2. **Relationship Handling**:
   - Relationships defined in the YAML should be treated as the source of truth
   - Cardinality should be determined from the relationship definition or auto-detected when enabled
   - Both ends of a relationship must be properly maintained
   - For one-to-many relationships, ensure uniqueness on the "one" side

3. **Performance Considerations**:
   - Pre-allocate slices and maps when size is known
   - Process entities in dependency order to handle relationships
   - Use buffered I/O for file operations
   - Generate data in memory before writing to files

## Testing and Quality Assurance

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

6. **Test Organization**: Follow this pattern for organizing tests:
   ```go
   func TestFunctionName(t *testing.T) {
       // Common setup if needed
       
       tests := []struct {
           name     string
           input    InputType
           expected OutputType
           wantErr  bool
       }{
           {
               name:     "Valid case description",
               input:    validInput,
               expected: expectedOutput,
               wantErr:  false,
           },
           {
               name:     "Error case description",
               input:    invalidInput,
               expected: OutputType{}, // Zero value
               wantErr:  true,
           },
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               // Test-specific setup
               
               actual, err := FunctionUnderTest(tt.input)
               
               if tt.wantErr {
                   assert.Error(t, err)
                   return
               }
               
               assert.NoError(t, err)
               assert.Equal(t, tt.expected, actual)
               
               // Additional assertions as needed
           })
       }
   }
   ```

### Troubleshooting Guide

When encountering issues, follow this debugging sequence:

1. **YAML Parsing Issues**:
   - Check YAML syntax validity first (use a linter)
   - Verify all required fields are present
   - Check entity and relationship definitions for completeness
   - Ensure external_ids are unique

2. **Relationship Problems**:
   - Verify both ends of the relationship are defined correctly
   - Check cardinality settings match the data model
   - Ensure target entities and attributes exist
   - Verify proper handling of optional vs. required relationships

3. **Data Generation Issues**:
   - Check attribute types and formats are supported
   - Verify primary keys are being handled correctly
   - Ensure relationship constraints are being properly applied
   - Confirm data generators are producing valid values

4. **CSV Output Problems**:
   - Check output directory permissions
   - Verify file naming conventions
   - Ensure CSV headers match attribute names
   - Confirm relationship data is consistent across files

## Working with Claude Code

### Implementation Planning

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
- [ ] Create test for X functionality
- [ ] Update existing tests for Y changes
- [ ] Test end-to-end workflow with new changes

## Documentation Updates
- [ ] Update README with new functionality
- [ ] Add examples of new feature usage
```

### Code Review Checklist

When reviewing implementation:

1. **Functionality**:
   - Does the code accomplish the stated goal?
   - Are edge cases handled appropriately?
   - Does it maintain backward compatibility?

2. **Code Quality**:
   - Is the code clean, readable, and well-organized?
   - Are variable and function names clear and descriptive?
   - Are there appropriate comments for complex logic?

3. **Testing**:
   - Are there tests for the new functionality?
   - Do tests cover happy paths and error cases?
   - Are tests clear and maintainable?

4. **Error Handling**:
   - Are errors properly checked and propagated?
   - Are error messages clear and actionable?
   - Is the system resilient to invalid inputs?

5. **Performance**:
   - Are operations efficient for expected data sizes?
   - Are there any potential bottlenecks?
   - Is memory usage reasonable?

### Development Principles

- **Law of Demeter**: A method of an object should only invoke methods of objects that are closely related to it, promoting loose coupling and better modular design.

- **DRY Principle**: Don't repeat yourself - reduce repetition of software patterns by encouraging code reuse and simplifying maintenance.

- **Composition Over Inheritance**: Favor object composition over class inheritance when designing reusable functionality.

- **Small Interfaces**: Define interfaces with the minimum methods needed for their purpose.

- **Test-Driven Development**:
  - Start with tests that define expected behavior
  - Make small, incremental changes to the code
  - Rerun tests after each change
  - Refactor only after tests are passing

### Working Guidelines

- Act as an expert Go developer; write clean concise code with clear separation of duties.
- Do not try and cheat by hardcoding or bypassing the logical intent.
- Explain your plan before trying to write code. Ask for confirmation or clarification of the plan before suggesting file edits.
- When asked to write a stub test or stub logic, implement only the function signature without implementation logic.

- **IMPORTANT**: NEVER TRY TO FAKE TESTS!
  - Tests must genuinely verify the functionality, not just appear to pass.
  - Each test should have meaningful assertions that validate actual behavior.
  - Tests should fail when the implementation is incorrect.
  - **NEVER use t.Skip() to hide failing tests due to broken implementation**
  - t.Skip() is ONLY acceptable for broken test infrastructure (missing dependencies, external services unavailable, etc.)
  - When implementation is broken, the test MUST fail - that's the entire point of TDD
  - Skipping tests because "the code isn't ready yet" defeats the purpose of Red-Green-Refactor

## Spec-Driven Development with Spec Kit

This project uses [GitHub Spec Kit](https://github.com/github/spec-kit) for structured, AI-assisted development. Spec Kit provides a workflow for defining specifications, creating implementation plans, and executing tasks systematically.

### Spec Kit Workflow

The development process follows six steps:

1. **Establish Principles** (`/speckit.constitution`): Review and update project principles in `.specify/memory/constitution.md`
2. **Create Specification** (`/speckit.specify`): Define what to build, user stories, and acceptance criteria
3. **Create Plan** (`/speckit.plan`): Develop technical implementation approach and architecture
4. **Break Down Tasks** (`/speckit.tasks`): Generate sequenced, actionable tasks with dependencies
5. **Analyze** (`/speckit.analyze`): Optional cross-artifact consistency check
6. **Execute** (`/speckit.implement`): Implement tasks following TDD principles

### Available Commands

Core workflow commands:
- `/speckit.constitution` - Establish or update project governance principles
- `/speckit.specify` - Create functional specifications for new features
- `/speckit.plan` - Develop technical implementation plans
- `/speckit.tasks` - Generate detailed task breakdowns
- `/speckit.implement` - Execute implementation with AI assistance

Enhancement commands (optional):
- `/speckit.clarify` - Ask structured questions to de-risk ambiguous areas (run before `/speckit.plan`)
- `/speckit.analyze` - Generate consistency report across artifacts (after `/speckit.tasks`)
- `/speckit.checklist` - Validate requirements completeness and clarity (after `/speckit.plan`)

### Directory Structure

```
.specify/
├── memory/
│   └── constitution.md     # Project governance principles
├── scripts/                # Utility scripts for workflow automation
├── specs/
│   └── [feature-number]-[feature-name]/
│       ├── spec.md         # Functional specification
│       ├── plan.md         # Technical implementation plan
│       ├── tasks.md        # Detailed task breakdown
│       ├── contracts/      # API and data model specs (optional)
│       ├── research.md     # Tech research details (optional)
│       └── quickstart.md   # Getting started guide (optional)
└── templates/              # Templates for specs and plans
```

### Integration with Existing Workflow

Spec Kit complements the existing development practices:

1. **Constitution**: `.specify/memory/constitution.md` codifies the project principles from this CLAUDE.md file
2. **Specifications**: Feature specs define requirements before implementation begins
3. **TDD Alignment**: Task breakdowns explicitly include test creation steps
4. **Quality Gates**: Analysis commands verify consistency and completeness
5. **Documentation**: Specs serve as living documentation of feature rationale

### When to Use Spec Kit

Use Spec Kit for:
- **New features**: Major functionality additions requiring planning
- **Refactoring**: Significant architectural changes
- **Complex bugs**: Issues requiring investigation and design
- **API changes**: Changes affecting external contracts

Don't use Spec Kit for:
- **Simple bugs**: Single-file fixes with obvious solutions
- **Trivial changes**: Documentation updates, typo fixes
- **Emergency hotfixes**: Critical production issues requiring immediate action

## Active Technologies
- Go 1.25 + gopkg.in/yaml.v3 (existing), github.com/fatih/color (existing), github.com/stretchr/testify (testing) (001-per-entity-row-counts)
- File-based (YAML input, CSV output) (001-per-entity-row-counts)

## Recent Changes
- 001-per-entity-row-counts: Added Go 1.25 + gopkg.in/yaml.v3 (existing), github.com/fatih/color (existing), github.com/stretchr/testify (testing)
