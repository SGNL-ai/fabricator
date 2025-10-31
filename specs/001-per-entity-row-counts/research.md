# Technical Research: Per-Entity Row Count Configuration

**Feature**: 001-per-entity-row-counts
**Date**: 2025-10-30
**Status**: Complete

## Overview

This document captures technical research and decisions for implementing per-entity row count configuration in fabricator. The research focuses on CLI subcommand patterns, YAML configuration design, and integration with existing CSV generation logic.

## Research Areas

### 1. CLI Subcommand Implementation in Go

**Decision**: Use explicit subcommand routing in main.go with dedicated handler functions

**Rationale**:
- Fabricator currently uses flag-based commands (no subcommands)
- Adding subcommands requires checking `os.Args[1]` before flag parsing
- Pattern: `if len(os.Args) > 1 && os.Args[1] == "init-count-config" { ... }`
- Keeps main.go as routing layer, delegates logic to `pkg/subcommands`
- Future-proof for additional subcommands (e.g., `validate`, `schema`, etc.)

**Alternatives Considered**:
- **cobra/viper framework**: Rejected - adds heavyweight dependency for single subcommand
- **Flag-based mode switch**: Rejected - violates Unix command design (actions should be verbs)
- **Separate binary**: Rejected - increases distribution complexity

**Implementation Approach**:
```go
// In cmd/fabricator/main.go
if len(os.Args) > 1 {
    switch os.Args[1] {
    case "init-count-config":
        subcommands.InitCountConfig(os.Args[2:])
        return
    }
}
// Continue with existing flag parsing for generate command
```

### 2. Row Count Configuration YAML Schema

**Decision**: Flat map structure with entity external_id as key, count as integer value

**Rationale**:
- Simple, intuitive schema matching user mental model
- Easy to manually edit without complex nesting
- Supports inline YAML comments for guidance
- Minimal parsing logic required
- Aligns with spec assumption (line 108)

**Schema**:
```yaml
# Row count configuration for fabricator
# Generated from: <sor-file-name>.yaml
# Last updated: <timestamp>

# Entity: users
# Description: User accounts
users: 1000

# Entity: groups
# Description: User groups
groups: 50

# Entity: permissions
# Description: Access permissions
permissions: 200
```

**Alternatives Considered**:
- **Nested structure with metadata**: Rejected - over-engineered for simple count mapping
- **Array of objects**: Rejected - harder to edit, more verbose
- **JSON format**: Rejected - spec specifies YAML, less human-friendly

**Validation Rules**:
- Keys must match entity external_id from SOR YAML
- Values must be positive integers (>0)
- Unknown entity keys trigger clear error message
- Missing entities default to `-n` value or 100

### 3. Integration with Existing CSV Generator

**Decision**: Modify generators to accept `map[string]int` (entity external_id → count) instead of single `int`

**Rationale**:
- Current generator uses `dataVolume int` parameter uniformly
- Changing to map requires minimal refactoring
- Orchestrator constructs map from either config file or `-n` flag
- Generator looks up count per entity: `count := rowCounts[entity.ExternalID]`
- Fallback logic: `if count == 0 { count = defaultCount }`

**Implementation Pattern**:
```go
// Current signature
func Generate(entities []Entity, dataVolume int) error

// New signature
func Generate(entities []Entity, rowCounts map[string]int, defaultCount int) error
```

**Migration Path**:
1. Add new parameter to generator functions
2. Update orchestrator to build map from either source
3. Existing `-n` flag constructs uniform map: `map[entity.ExternalID]dataVolume`
4. Config file provides non-uniform map directly
5. Backward compatibility maintained

### 4. Cardinality Violation Handling

**Decision**: Best-effort assignment with structured warnings logged to stderr

**Rationale**:
- Spec explicitly requires generation even when cardinality impossible (FR-006)
- Users value flexibility over strict enforcement for test data
- Warnings must be actionable and clear
- Example: "Warning: Entity 'employees' has 50 rows but 100 departments require 1+ employee each"

**Warning Strategy**:
- Detect violations during relationship assignment phase
- Calculate shortfall: required relationships vs. available target rows
- Emit warning with entity names, counts, and relationship name
- Continue generation (no errors thrown)
- Use color coding: yellow for warnings (existing fatih/color dependency)

**Detection Logic**:
```go
// For one-to-many: parent entity → many children
// If parent count > child count, some parents cannot get children
if relationship.Cardinality == OneToMany {
    if parentCount > childCount {
        warn("Cardinality violation: %d %s require children but only %d %s available",
             parentCount, parentEntity, childCount, childEntity)
    }
}
```

### 5. Template Generation from SOR YAML

**Decision**: Parse SOR YAML, extract entities, write formatted YAML to stdout with comments

**Rationale**:
- Spec requires stdout output for Unix flexibility (clarification Q3)
- Template must be immediately usable (FR-012)
- Comments improve UX by explaining each entity
- Default values aid user understanding (suggest 100)

**Template Generator Logic**:
1. Parse SOR YAML file
2. Extract all entity external_ids
3. Optionally extract entity display names/descriptions
4. Generate YAML output with:
   - Header comment explaining file purpose
   - Per-entity comment blocks
   - Default count value (100)
   - Footer comment with usage instructions
5. Write to stdout (os.Stdout)

**Comment Format**:
```yaml
# Entity: <external_id>
# Description: <description if available from SOR>
<external_id>: 100
```

### 6. Error Handling Strategy

**Decision**: Fail fast with actionable error messages for invalid configurations

**Rationale**:
- Spec requires clear error messages (FR-013, FR-014, FR-015)
- 90% of users should fix issues on first attempt (SC-005)
- Errors should include: problem description, location, suggested fix

**Error Categories & Messages**:

1. **File not found**: `Error: Count configuration file not found: <path>\nSuggestion: Generate a template with 'fabricator init-count-config -f <sor.yaml> > counts.yaml'`

2. **Invalid YAML syntax**: `Error: Invalid YAML syntax in <file>:<line>:<col>\nDetails: <parse error>\nSuggestion: Validate YAML syntax at yamllint.com`

3. **Unknown entity**: `Error: Entity 'foo' in count configuration not found in SOR YAML\nAvailable entities: users, groups, permissions\nSuggestion: Remove 'foo' or check entity external_id spelling`

4. **Invalid count value**: `Error: Invalid count for entity 'users': <value>\nExpected: Positive integer (>0)\nGiven: <value>\nSuggestion: Use a number like 100, 1000, etc.`

5. **Conflicting flags**: `Error: Cannot use both -n flag and --count-config file\nSuggestion: Choose one: -n for uniform counts OR --count-config for per-entity counts`

### 7. Testing Strategy

**Decision**: Three-tier testing approach (unit, component, integration)

**Rationale**:
- Constitution requires 80% coverage minimum
- Table-driven tests with testify framework
- Test fixtures from existing examples directory
- Mock SOR YAML files for edge cases

**Test Coverage Plan**:

**Unit Tests** (pkg/config/*_test.go):
- Row count configuration parsing (valid/invalid YAML)
- Validation logic (unknown entities, invalid counts)
- Template generation formatting
- Error message generation

**Component Tests** (pkg/subcommands/*_test.go):
- init-count-config subcommand with various SOR files
- Flag parsing for --count-config
- Integration between config and orchestrator

**Integration Tests** (tests/integration/per_entity_counts_test.go):
- End-to-end: config file → CSV generation → row count verification
- Backward compatibility: -n flag still works
- Conflict detection: both flags provided
- Cardinality warning scenarios

**Test Fixtures**:
- Use existing examples/*.yaml as base SOR files
- Create test count configurations in tests/fixtures/
- Small datasets (10-50 rows) for fast execution

### 8. Backward Compatibility Strategy

**Decision**: Preserve all existing flag behaviors; treat config file as alternative to -n

**Rationale**:
- Spec requires existing workflows unchanged (User Story 3, FR-007)
- No migration required for current users
- Clear mental model: uniform (-n) vs. per-entity (config) counts

**Compatibility Matrix**:

| Scenario | -n flag | --count-config | Behavior |
|----------|---------|----------------|----------|
| Legacy user | ✅ 100 | ❌ (not provided) | Generate 100 rows per entity (current behavior) |
| New user | ❌ (not provided) | ✅ file.yaml | Use per-entity counts from file |
| Power user | ❌ (not provided) | ❌ (not provided) | Default to 100 rows per entity |
| Error case | ✅ 100 | ✅ file.yaml | Error: conflicting options (FR-008) |

**Flag Validation Logic**:
```go
if dataVolume != defaultDataVolume && countConfigFile != "" {
    return fmt.Errorf("cannot use both -n and --count-config flags")
}
```

## Dependencies

### Existing (No Changes)
- `gopkg.in/yaml.v3`: YAML parsing (already used for SOR files)
- `github.com/fatih/color`: Terminal color output (warnings)
- `github.com/stretchr/testify`: Testing framework

### New (None Required)
No new external dependencies needed. Feature uses existing dependencies.

## Performance Considerations

### Template Generation
- **Target**: <5 seconds for 50 entities (SC-002)
- **Expected**: <1 second for typical 10-20 entity SOR files
- **Bottleneck**: None (simple YAML parse + format + write)
- **Optimization**: Not needed at this scale

### CSV Generation
- **Target**: No regression from current performance
- **Impact**: Minimal - map lookup O(1) vs. single int read
- **Memory**: Map overhead negligible (<1KB for 50 entities)
- **Optimization**: Not needed

### Large Row Counts
- **Scenario**: User specifies 1M+ rows for entity
- **Behavior**: System attempts generation (spec assumption line 111)
- **Risk**: Out of memory errors
- **Mitigation**: Document limitation; user responsibility per spec

## Open Questions

None. All technical decisions resolved through research phase.

## References

- Feature Specification: [spec.md](./spec.md)
- Clarifications (Session 2025-10-30): [spec.md#clarifications](./spec.md#clarifications)
- Constitution: [/.specify/memory/constitution.md](../../.specify/memory/constitution.md)
- Existing YAML Parser: [/pkg/parser/parser.go](../../pkg/parser/parser.go)
- Existing CSV Generator: [/pkg/generators/csv_generator.go](../../pkg/generators/csv_generator.go)

## Next Steps

Proceed to Phase 1:
1. Generate data-model.md defining configuration structures
2. Create contracts/count-config-schema.yaml
3. Generate quickstart.md for users
4. Update agent context with new packages
