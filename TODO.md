# Fabricator CSV Generator TODO

## Project Structure
- [x] Create directory structure
- [x] Define package structure
- [x] Create basic CLI framework

## Core Functionality
- [x] YAML Parsing
  - [x] Create YAML model structures
  - [x] Load and parse YAML file
  - [x] Validate YAML structure

- [x] Entity Analysis
  - [x] Extract entity names and attributes 
  - [x] Identify relationships between entities
  - [x] Map externalId to CSV file names

- [x] CSV Generation
  - [x] Create generators for different data types
  - [x] Ensure relationship consistency between entities
  - [x] Add support for test data volume parameter

- [x] Output Management
  - [x] Create output directory structure
  - [x] Write CSV files
  - [x] Implement colorful and informative user output

## CLI Improvements
- [x] Add short and long-form command-line options (-f/--file)
- [x] Update README with improved usage examples
- [x] Make CLI output more informative
- [x] Ensure error handling is comprehensive

## Generalization
- [x] Make data generation generic (not specific to SGNL)
- [x] Improve field type detection for better data generation
- [x] Support both namespaced and non-namespaced entity IDs
- [x] Generate appropriate data based on field name patterns

## Testing
- [x] Write unit tests for YAML parsing
- [x] Write unit tests for CLI
- [ ] Create integration tests

## Test Refactoring
- [ ] Create new test framework and structure using testify
- [ ] Design a table-driven testing approach for field type detection
- [ ] Implement clean focused tests for dependency graph functionality
- [ ] Create comprehensive tests for relationship validation
- [ ] Develop integration tests for end-to-end CSV generation
- [ ] Create proper test fixtures using YAML models
- [ ] Build helper utilities that abstract away implementation details
- [ ] Ensure test coverage for all edge cases
- [ ] Add documentation for the test approach and structure

## Documentation
- [x] Update README.md with usage instructions
- [x] Add examples
- [x] Document code
- [x] Create IMPLEMENTATION.md with details
- [x] Update CLAUDE.md with project information

## Architecture Refactoring
- [x] Create stub model structure for new domain model
  - [x] Design models with appropriate functions and relationships
  - [x] Implement Attribute model stub
  - [x] Implement Entity model stub
  - [x] Implement Relationship model stub
  - [x] Implement Graph model stub 
- [ ] Implement full model functionality
  - [x] Implement Attribute model functionality
  - [x] Implement Entity model functionality with validation
  - [x] Implement Relationship model with proper cardinality detection
  - [ ] Implement Graph model with YAML parsing and validation
    - [x] Create graph_test.go file with tests for Graph model
    - [x] Ensure all Relationship tests pass
    - [x] Add Graph constructor with validation
    - [x] Implement entity creation from YAML in Graph model
    - [x] Implement relationship creation from YAML in Graph model
    - [ ] Implement entity.addRelationship for proper relationship creation
    - [ ] Complete foreign key validation when Graph model is available
    - [ ] Add topological sorting functionality to Graph model
    - [x] Run Graph constructor tests to verify implementation
- [ ] Integrate new model with CSV generation
  - [ ] Create adapter between model and CSV generator
  - [ ] Update relationship handling to use new models
  - [ ] Update validation to use new models
- [ ] Refactor CSVGenerator to be a coordinator with steps in sub-packages
- [ ] Design pipeline architecture for generation steps
- [ ] Extract core generation steps from CSVGenerator into dedicated modules
- [ ] Rename relationship_handler.go to relationship_builder.go for clarity
- [ ] Rename relationship_validator.go to data_validator.go
- [ ] Evaluate need for entity_relationships.go and possibly consolidate
- [ ] Redesign data generation flow into a three-phase process:
  - [ ] Phase 1: Generate all identifier fields in topological order
  - [ ] Phase 2: Establish relationship structure between entities
  - [ ] Phase 3: Fill in remaining non-relationship fields
- [ ] Add explicit support for dependencies in field generation
- [ ] Create clear interfaces between generation phases

## Potential Future Enhancements
- [ ] Add support for Claude API integration for more realistic test data
- [ ] Add support for custom data generation rules
- [ ] Add a web UI for visualizing the generated data
- [ ] Support for additional output formats (JSON, XML, etc.)
- [x] Create a validation mode to verify existing CSV files
- [ ] Add schema documentation generator from YAML