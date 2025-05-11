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

## Documentation
- [x] Update README.md with usage instructions
- [x] Add examples
- [x] Document code
- [x] Create IMPLEMENTATION.md with details
- [x] Update CLAUDE.md with project information

## Potential Future Enhancements
- [ ] Add support for Claude API integration for more realistic test data
- [ ] Add support for custom data generation rules
- [ ] Add a web UI for visualizing the generated data
- [ ] Support for additional output formats (JSON, XML, etc.)
- [ ] Create a validation mode to verify existing CSV files
- [ ] Add schema documentation generator from YAML