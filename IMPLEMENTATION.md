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

### Generators Package Architecture

The `pkg/generators` package follows a modular architecture with clear separation of concerns:

#### Core Components

1. **CSVGenerator** (`pkg/generators/csv_generator.go`)
   - Central orchestrator that coordinates the entire generation process
   - Holds state including entity data, relationships, and configuration
   - Manages the overall workflow from initialization to CSV output
   - Provides public API for other packages to use

2. **Dependency Management** (`pkg/generators/dependency_graph.go`)
   - Builds directed graphs representing entity dependencies
   - Performs topological sorting to determine optimal generation order
   - Handles complex relationship scenarios including circular references
   - Ensures dependent entities are generated in the correct sequence

3. **Field Generation** (`pkg/generators/field_generator.go`)
   - Analyzes field names to determine appropriate data types
   - Generates contextually appropriate values for each field type
   - Ensures uniqueness for fields requiring unique values
   - Provides specialized generators for different field types (names, emails, dates, etc.)

4. **Relationship Handling** (`pkg/generators/relationship_handler.go`)
   - Actively creates and modifies data to enforce relationships during generation
   - Establishes connections between entities by writing consistent values
   - Supports different relationship types (one-to-one, one-to-many, many-to-one)
   - Performs automated cardinality detection when configured
   - Acts as a "writer" component in the system

5. **Data Validation** (`pkg/generators/relationship_validator.go`)
   - Passively checks data integrity after generation is complete
   - Verifies that all relationships are valid and consistent
   - Ensures uniqueness constraints are maintained
   - Provides detailed validation reports without modifying data
   - Acts as a "reader/verification" component in the system
   - Can be used independently in validation-only mode

6. **Entity Relationships** (`pkg/generators/entity_relationships.go`)
   - Extends CSVGenerator with specific relationship consistency functionality
   - Processes relationships for specific entities
   - Enforces consistency rules based on relationship type

7. **Utilities** (`pkg/generators/utils.go`)
   - Provides shared helper functions used across the package
   - Handles common operations like file naming and attribute checking

#### Data Generation Flow

1. **Initialization**
   ```
   User Input → Create CSVGenerator with configuration
   ```

2. **Setup Phase**
   ```
   Parse YAML Models → Build Dependency Graph → Calculate Generation Order
   ```

3. **Data Generation Phase**
   ```
   For each entity (in dependency order):
     For each attribute:
       Generate appropriate field value
       Store in memory
   ```

4. **Relationship Consistency Phase**
   ```
   For each relationship:
     Determine relationship type
     Apply appropriate consistency rules
     Update entity data accordingly
   ```

5. **Validation Phase** (optional)
   ```
   For each relationship:
     Check referential integrity
     Verify uniqueness constraints
     Generate validation results
   ```

6. **Output Generation Phase**
   ```
   For each entity:
     Create CSV file
     Write headers and data rows
     Place in output directory
   ```

This modular architecture enables clean separation of concerns while maintaining efficient coordination between components. Each module focuses on a specific aspect of the generation process, making the system flexible and maintainable.

#### Planned Architectural Improvements

The current architecture has room for further refinements:

1. **Three-Phase Generation Process**
   - Phase 1: Generate all identifier fields in topological order
   - Phase 2: Establish relationship structure between entities
   - Phase 3: Fill in remaining non-relationship fields
   - Replace the current "generate then replace" approach

2. **Purpose-Built Domain Model**
   - Create proper domain models for generated data instead of maps and arrays
   - Create clean abstractions that decouple from the YAML structure
   - Encapsulate Graph library internally for dependency management
   - Add type-safe interfaces for entity and relationship operations

3. **Pipeline-Based Processing**
   - Refactor CSVGenerator to be a lightweight coordinator
   - Move generation steps into dedicated sub-packages
   - Create a pipeline architecture for generation steps
   - Improve testability by isolating individual pipeline stages

4. **Component Improvements**
   - Rename relationship_handler.go to relationship_builder.go
   - Rename relationship_validator.go to data_validator.go
   - Evaluate and potentially consolidate entity_relationships.go
   - Create explicit interfaces between generation phases
   
5. **Interface-Driven Design**
   - Define clear interfaces between components
   - Reduce direct dependencies between modules
   - Enable easier mocking for tests
   - Support metadata needed during the generation process

These improvements will enhance the system's maintainability, testability, and separation of concerns while providing a more efficient and type-safe generation process.

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

## Testing Strategy

Our testing approach is focused on creating clean, maintainable tests that provide good coverage while being easy to understand and extend. We follow these key principles:

### Test Organization

1. **Unit Tests**: Focused tests for individual functions and methods
   - Test a single functionality in isolation
   - Mock external dependencies
   - Quick to run and diagnose

2. **Component Tests**: Tests for interactions between related components
   - Test how different parts of the system work together
   - Focus on boundaries between components
   - Verify correct integration of units

3. **Integration Tests**: End-to-end flow tests
   - Test complete workflows from YAML input to CSV output
   - Verify file creation and data consistency
   - Test realistic scenarios with example YAML files

### Test Implementation

1. **Table-Driven Testing**: We use Go's table-driven testing pattern for all suitable tests
   - Define test cases as data structures
   - Loop through test cases with the same test logic
   - Clearly separate test data from test logic
   - Use descriptive test case names that explain the scenario

2. **TestHelper Framework**: We use a comprehensive test helper framework (`pkg/generators/test_helpers.go`)
   - Provides utilities for creating test fixtures and validating outputs
   - Abstracts away common test setup and verification logic
   - Creates temporary directories for test isolation
   - Manages test resource cleanup
   - Supports table-driven test execution
   - Provides relationship validation utilities

3. **Testify Framework**: All tests use the testify assertion library
   - `assert` for non-fatal assertions that continue test execution
   - `require` for fatal assertions that stop test execution on failure
   - Consistent assertion style across all tests
   - Clear failure messages that explain the expected vs. actual values

4. **Test File Organization**: Tests are organized following these conventions
   - Tests for a file `example.go` should be in `example_test.go`
   - Tests should be in the same package as the code they're testing
   - Helper functions are centralized in the `test_helpers.go` file
   - Integration tests use a dedicated `csv_generator_integration_test.go` file
   - Component-specific tests focus on isolated functionality

5. **Test Coverage**: We aim for high test coverage with these priorities
   - Core logic and data transformation should have 90%+ coverage
   - Error handling paths should be explicitly tested
   - Edge cases should be identified and tested
   - Complex algorithms should have test cases that verify correct behavior
   - Relationship validation and consistency should be thoroughly tested

6. **Test Categories**: Tests are organized into several categories
   - **Unit Tests**: Test individual components like field generation, dependency graph
   - **Validation Tests**: Test relationship validation and consistency checks
   - **Integration Tests**: Test end-to-end generation process
   - **Performance Tests**: Test system behavior with large datasets
   - **Edge Case Tests**: Test specific boundary conditions and error scenarios

7. **Legacy Tests Reference**: Old test implementations are preserved in `pkg/generators/legacy_tests/` for reference
   - New tests should be created in the main package directories
   - Legacy tests will eventually be removed after refactoring is complete

### Test Helper Structure

The `TestHelper` framework in `pkg/generators/test_helpers.go` provides several categories of functionality:

#### 1. Test Setup and Cleanup

```go
// Create a new test helper with isolated environment
helper := NewTestHelper(t)
defer helper.Cleanup() // Automatically cleanup temp files
```

- Creates isolated temporary directories for each test
- Manages resource cleanup after test completion
- Prevents test interference and ensures repeatability

#### 2. Test Fixture Creation

```go
// Create basic entity fixtures
userEntity := helper.CreateBasicEntity(
    "User",
    "User",
    []string{"id", "name", "email"},
    [][]string{},
)

// Create relationships between entities
entities, relationships := helper.CreateBasicEntitiesWithRelationship(
    "User",       // Source entity
    "Department", // Target entity
    "dept_id",    // Source to target attribute
    "id",         // Target to source attribute
    "N:1",        // Cardinality
)
```

- Creates entity fixtures with specified attributes
- Builds relationship structures between entities
- Supports common relationship patterns (1:1, N:1, etc.)
- Creates complete YAML models for testing

#### 3. Generator Setup

```go
// Create generator with standard configuration
generator := helper.SetupBasicGenerator(10, false)
```

- Creates preconfigured generators with standard settings
- Manages output directories and cleanup
- Sets appropriate configuration for test contexts

#### 4. Output Validation

```go
// Verify CSV existence and structure
helper.VerifyCSVFileExists("User")
helper.VerifyCSVFileContents("User", []string{"id", "name", "email"}, 10)

// Verify relationship consistency
helper.VerifyRelationshipConsistency(
    "User", "Department",
    "dept_id", "id",
    "N:1",
)
```

- Validates CSV file existence and structure
- Checks headers and row counts match expectations
- Verifies relationship consistency between entities
- Provides detailed feedback on validation failures

#### 5. Table-Driven Test Support

```go
// Run multiple test cases with the same logic
testCases := []struct {
    Name           string
    EntitiesSetup  func() map[string]*models.Entity
    Relationships  func() []*models.Relationship
    NumRows        int
    AutoCardinality bool
    Validator      func(*CSVGenerator) bool
}{
    // Test cases defined here
}

helper.RunTableDrivenTest(testCases)
```

- Supports table-driven test execution
- Manages test case setup and teardown
- Provides per-test isolation
- Allows custom validation logic for each test case

This comprehensive test helper structure enables clean, focused tests without duplicating setup and validation code across test files.

### Running Tests

```bash
# Run all tests
make test

# Run tests in a specific package
go test ./pkg/generators

# Run tests with verbose output
go test -v ./pkg/generators

# Run a specific test
go test ./pkg/generators -run TestFieldGeneration
```

## Building and Running

```bash
# Build the project
make build

# Run the fabricator tool
./build/fabricator -input example.yaml -output data -volume 100
```