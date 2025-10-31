# Feature Specification: Per-Entity Row Count Configuration

**Feature Branch**: `001-per-entity-row-counts`
**Created**: 2025-10-30
**Status**: Draft
**Input**: User description: "We have a feature request to allow differing numbers of rows of data to be generated. Instead of the simple `-n` parameter we should allow the user to pass a yaml file that specifies the count for each entity they want to create. It would also be helpful to allow fabricator to generate a stub file based on the referenced SOR yaml file that the user can edit."

## Clarifications

### Session 2025-10-30

- Q: How should users trigger template generation? → A: Dedicated subcommand `fabricator init-count-config -f sor.yaml`
- Q: How should the system behave when row counts create impossible relationship cardinality? → A: Generate data anyway using best-effort relationship assignment, warn user about cardinality violations
- Q: Where should the generated template be written? → A: Write to stdout by default

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Generate Custom Row Counts per Entity (Priority: P1)

Users need to generate different amounts of test data for different entities based on their testing needs. For example, they may need 1000 users but only 10 departments, reflecting realistic data distributions.

**Why this priority**: This is the core value of the feature - allowing users to specify different row counts per entity. Without this, the feature provides no benefit over the existing `-n` parameter.

**Independent Test**: Can be fully tested by providing a row count configuration file and verifying that each entity CSV contains the exact number of rows specified for that entity.

**Acceptance Scenarios**:

1. **Given** a valid SOR YAML file with 3 entities (users, groups, permissions), **When** user provides a row count configuration specifying 100 users, 20 groups, and 50 permissions, **Then** the generated CSVs contain exactly those row counts
2. **Given** a row count configuration file, **When** user runs fabricator with this configuration, **Then** relationship consistency is maintained across all entities regardless of differing row counts
3. **Given** a row count configuration that specifies counts for only some entities, **When** user runs fabricator, **Then** entities without specified counts use a default row count

---

### User Story 2 - Generate Row Count Configuration Template (Priority: P2)

Users need a starting point for creating their row count configuration file. Rather than manually writing the YAML structure, users should be able to generate a template based on their SOR definition file.

**Why this priority**: This significantly improves user experience by reducing setup time and preventing configuration errors. Users can generate a template and simply edit the numbers rather than learning the configuration format.

**Independent Test**: Can be fully tested by running the stub generation command with a SOR YAML file and verifying the output contains all entities with default placeholder values.

**Acceptance Scenarios**:

1. **Given** a valid SOR YAML file with multiple entities, **When** user runs `fabricator init-count-config -f sor.yaml`, **Then** a row count configuration template is written to stdout containing all entities from the SOR with default row count values
2. **Given** a generated row count configuration template, **When** user opens the file, **Then** the file is properly formatted, human-readable, and includes helpful comments explaining how to customize values
3. **Given** a generated template with default values, **When** user runs fabricator without modifying the template, **Then** data is generated successfully using the default row counts

---

### User Story 3 - Preserve Backward Compatibility with `-n` Parameter (Priority: P1)

Existing users rely on the simple `-n` parameter for uniform row counts across all entities. This workflow must continue to work without requiring migration to the new configuration file approach.

**Why this priority**: Breaking existing workflows would disrupt current users and violate the principle of backward compatibility. This is critical for adoption.

**Independent Test**: Can be fully tested by running fabricator with the existing `-n` parameter (without a row count configuration file) and verifying behavior is unchanged from previous versions.

**Acceptance Scenarios**:

1. **Given** a valid SOR YAML file, **When** user runs fabricator with `-n 100` and no row count configuration file, **Then** all entities generate exactly 100 rows as they did before this feature
2. **Given** both `-n` parameter and a row count configuration file are provided, **When** user runs fabricator, **Then** an error message explains that only one row count method can be used at a time
3. **Given** neither `-n` parameter nor row count configuration file are provided, **When** user runs fabricator, **Then** the default row count (100) is used for all entities

---

### Edge Cases

- What happens when the row count configuration file specifies an entity that doesn't exist in the SOR YAML file?
- What happens when the row count configuration file has invalid YAML syntax?
- What happens when row count values are negative, zero, or non-numeric?
- How does the system handle very large row count values (e.g., millions) that could cause memory issues?
- When row count configuration specifies counts that make relationship cardinality impossible to satisfy (e.g., 1-to-many relationship but "many" side has fewer rows), the system generates data using best-effort relationship assignment and warns the user about cardinality violations
- What happens if the row count configuration file path doesn't exist or isn't readable?
- How does the system behave when the template generation command is given an invalid or malformed SOR YAML file?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept a configuration file path via a new command-line flag that specifies per-entity row counts
- **FR-002**: Row count configuration file MUST use YAML format and map entity external IDs to integer row counts
- **FR-003**: System MUST validate that all entities referenced in the row count configuration exist in the SOR YAML file
- **FR-004**: System MUST validate that all row count values are positive integers
- **FR-005**: System MUST generate CSVs with the exact number of rows specified for each entity in the configuration file
- **FR-006**: System MUST maintain relationship consistency when entities have different row counts using best-effort assignment; when cardinality constraints cannot be satisfied, system MUST generate data and emit warnings describing the violations
- **FR-007**: System MUST continue to support the existing `-n` parameter for uniform row counts across all entities
- **FR-008**: System MUST reject execution when both `-n` parameter and row count configuration file are provided simultaneously
- **FR-009**: System MUST use a default row count for entities not specified in the configuration file
- **FR-010**: System MUST provide a subcommand `init-count-config` to generate a row count configuration template from a SOR YAML file and write it to stdout
- **FR-011**: Generated templates MUST include all entities from the SOR YAML file with placeholder row count values
- **FR-012**: Generated templates MUST be valid YAML that can be immediately used with fabricator (via redirection or copy-paste)
- **FR-013**: System MUST provide clear error messages when row count configuration file has invalid syntax
- **FR-014**: System MUST provide clear error messages when row count configuration references non-existent entities
- **FR-015**: System MUST provide clear error messages when row count values are invalid (negative, zero, non-numeric)

### Key Entities

- **Row Count Configuration**: Maps entity external IDs to desired row counts; stored as YAML file provided by user
- **Entity Row Count**: The number of rows to generate for a specific entity; defaults to global `-n` value or system default (100) if not specified
- **Configuration Template**: A pre-populated row count configuration file generated from a SOR YAML file to help users get started

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can specify different row counts for each entity and receive CSV files with exactly those counts
- **SC-002**: Users can generate a row count configuration template in under 5 seconds for SOR files with up to 50 entities
- **SC-003**: Generated templates are immediately usable without modification, producing valid test data
- **SC-004**: Existing users can continue using the `-n` parameter with identical behavior to previous versions
- **SC-005**: Clear error messages guide users to fix configuration issues on the first attempt in 90% of cases
- **SC-006**: Relationship consistency is maintained across entities regardless of differing row counts

## Assumptions

- The default row count for entities not specified in the configuration file will be 100 (matching the current default for `-n`)
- The row count configuration file will use a simple flat structure mapping entity external_id to count (e.g., `users: 1000`)
- Template generation will include inline YAML comments to help users understand how to customize values
- When `-n` and configuration file conflict, the system will fail fast rather than choosing one silently
- Very large row counts (e.g., >1 million) are the user's responsibility to manage; the system will attempt generation but may run out of memory
- The configuration file will be optional; if not provided, behavior falls back to `-n` parameter or default
- Template generation will be a separate operation from data generation (not an automatic precursor)

## Out of Scope

- Automatic optimization or adjustment of row counts to satisfy relationship constraints
- Built-in validation of whether row counts are "reasonable" for the data model
- Support for dynamic row count expressions or formulas (e.g., "10% of users count")
- Integration with external configuration management systems
- Support for configuration formats other than YAML (JSON, TOML, etc.)
- Automatic detection of "typical" row count ratios based on relationship analysis
