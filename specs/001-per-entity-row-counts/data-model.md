# Data Model: Per-Entity Row Count Configuration

**Feature**: 001-per-entity-row-counts
**Date**: 2025-10-30
**Status**: Complete

## Overview

This document defines the data structures and their relationships for per-entity row count configuration feature. All structures follow Go conventions and integrate with existing fabricator data models.

## Core Data Structures

### 1. CountConfiguration

**Purpose**: Represents parsed row count configuration from YAML file

**Location**: `pkg/config/count_config.go`

**Structure**:
```go
// CountConfiguration maps entity external IDs to their desired row counts
type CountConfiguration struct {
    // EntityCounts maps entity external_id → row count
    EntityCounts map[string]int

    // SourceFile is the path to the configuration file (for error messages)
    SourceFile string

    // LoadedAt is when the configuration was loaded
    LoadedAt time.Time
}
```

**Fields**:
- `EntityCounts`: Key-value map where keys are entity external_ids from SOR YAML
- `SourceFile`: Original file path for better error messages and debugging
- `LoadedAt`: Timestamp for cache invalidation or logging

**Validation Rules**:
- Map keys must match entity external_ids in SOR definition
- Map values must be positive integers (>0)
- Empty map is valid (falls back to defaults)

**Methods**:
```go
// GetCount returns the row count for an entity, or defaultCount if not specified
func (c *CountConfiguration) GetCount(entityExternalID string, defaultCount int) int

// Validate checks the configuration against SOR entities
func (c *CountConfiguration) Validate(sorEntities []string) error

// HasEntity returns true if entity has explicit count in configuration
func (c *CountConfiguration) HasEntity(entityExternalID string) bool
```

### 2. ConfigurationTemplate

**Purpose**: Represents a generated row count configuration template

**Location**: `pkg/config/template.go`

**Structure**:
```go
// ConfigurationTemplate holds data for generating a count config template
type ConfigurationTemplate struct {
    // Entities to include in template
    Entities []TemplateEntity

    // SourceSORFile is the SOR YAML file used to generate template
    SourceSORFile string

    // DefaultCount is the placeholder count value for each entity
    DefaultCount int

    // GeneratedAt timestamp
    GeneratedAt time.Time
}

// TemplateEntity represents an entity in the template with metadata
type TemplateEntity struct {
    // ExternalID is the entity's external_id (YAML key)
    ExternalID string

    // DisplayName is a human-readable name (if available from SOR)
    DisplayName string

    // Description is a brief description (if available from SOR)
    Description string

    // DefaultCount is the suggested row count for this entity
    DefaultCount int
}
```

**Methods**:
```go
// Render generates YAML output with comments
func (t *ConfigurationTemplate) Render() ([]byte, error)

// WriteTo writes the template to an io.Writer (e.g., os.Stdout)
func (t *ConfigurationTemplate) WriteTo(w io.Writer) error
```

### 3. ValidationError

**Purpose**: Structured error type for configuration validation failures

**Location**: `pkg/config/validator.go`

**Structure**:
```go
// ValidationError represents a configuration validation failure
type ValidationError struct {
    // EntityID is the problematic entity external_id (if applicable)
    EntityID string

    // Field is the problematic field name (if applicable)
    Field string

    // Value is the invalid value
    Value interface{}

    // Message is a human-readable description
    Message string

    // Suggestion is an actionable fix suggestion
    Suggestion string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s\nSuggestion: %s", e.Message, e.Suggestion)
}
```

**Usage Pattern**:
```go
if count <= 0 {
    return &ValidationError{
        EntityID:   entityID,
        Field:      "count",
        Value:      count,
        Message:    fmt.Sprintf("Invalid count for entity '%s': %d (expected positive integer)", entityID, count),
        Suggestion: "Use a number like 100, 1000, etc.",
    }
}
```

### 4. CardinalityWarning

**Purpose**: Represents a cardinality constraint violation warning

**Location**: `pkg/generators/warnings.go` (extend existing package)

**Structure**:
```go
// CardinalityWarning represents a detected cardinality violation
type CardinalityWarning struct {
    // RelationshipName identifies the relationship
    RelationshipName string

    // SourceEntity is the "one" side of the relationship
    SourceEntity string

    // SourceCount is how many source entity rows exist
    SourceCount int

    // TargetEntity is the "many" side of the relationship
    TargetEntity string

    // TargetCount is how many target entity rows exist
    TargetCount int

    // Cardinality is the expected cardinality (e.g., "one-to-many")
    Cardinality string

    // Shortfall describes the gap (e.g., "50 departments need 1+ employee but only 30 employees exist")
    Shortfall string
}

// String formats the warning for display
func (w *CardinalityWarning) String() string {
    return fmt.Sprintf(
        "Cardinality warning: Relationship '%s' (%s) - %s has %d rows but %s has %d rows. %s",
        w.RelationshipName, w.Cardinality, w.SourceEntity, w.SourceCount,
        w.TargetEntity, w.TargetCount, w.Shortfall,
    )
}
```

## Data Flows

### Flow 1: Load and Apply Count Configuration

```
User provides --count-config file.yaml
    ↓
pkg/config.LoadConfiguration(path) → CountConfiguration
    ↓
CountConfiguration.Validate(sorEntities) → error or nil
    ↓
orchestrator builds rowCounts map[string]int
    ↓
generators.Generate(..., rowCounts, defaultCount)
    ↓
For each entity: count = rowCounts[entity.ExternalID] or defaultCount
    ↓
Generate CSV with specified row count
```

### Flow 2: Generate Template

```
User runs: fabricator init-count-config -f sor.yaml
    ↓
parser.ParseSOR(sor.yaml) → SOR definition
    ↓
subcommands.InitCountConfig() extracts entities
    ↓
config.NewTemplate(entities, defaultCount=100) → ConfigurationTemplate
    ↓
template.WriteTo(os.Stdout)
    ↓
YAML output to stdout
```

### Flow 3: Backward Compatible Default

```
User provides -n 50 (no config file)
    ↓
orchestrator builds uniform map: map[entity.ExternalID]50
    ↓
generators.Generate(..., rowCounts, defaultCount=50)
    ↓
Behavior identical to previous version
```

### Flow 4: Cardinality Violation Detection

```
generators.Generate() assigns relationships
    ↓
For each relationship: check if sufficient target rows exist
    ↓
If insufficient: create CardinalityWarning
    ↓
Append warning to []CardinalityWarning slice
    ↓
After generation: emit warnings to stderr with color
    ↓
CSV files still written (best-effort relationships)
```

## State Transitions

### CountConfiguration States

```
[Not Loaded]
    ↓ LoadConfiguration()
[Loaded - Unvalidated]
    ↓ Validate()
[Validated - Ready] ←→ [Validation Failed]
    ↓ GetCount()
[In Use]
```

**Transitions**:
- **Not Loaded → Loaded**: File read and YAML parsed
- **Loaded → Validated**: SOR entities cross-checked, counts verified
- **Loaded → Failed**: Validation errors found (unknown entities, invalid counts)
- **Validated → In Use**: Orchestrator queries counts during generation

## Relationships to Existing Models

### Integration with SOR Parser

**Existing**: `pkg/parser` parses SOR YAML → SOR definition with entities

**Integration Point**: CountConfiguration validation requires entity list from SOR

```go
// In orchestrator
sor := parser.ParseSOR(sorFile)
config := config.LoadConfiguration(configFile)
if err := config.Validate(sor.EntityExternalIDs()); err != nil {
    return err
}
```

### Integration with CSV Generator

**Existing**: `pkg/generators` generates CSV rows with uniform `dataVolume int`

**Change**: Replace `dataVolume int` with `rowCounts map[string]int, defaultCount int`

```go
// Before
func Generate(entities []Entity, dataVolume int) error

// After
func Generate(entities []Entity, rowCounts map[string]int, defaultCount int) error

// Usage inside generator
for _, entity := range entities {
    count := rowCounts[entity.ExternalID]
    if count == 0 {
        count = defaultCount
    }
    // Generate 'count' rows for this entity
}
```

### Integration with Orchestrator

**Existing**: `pkg/orchestrator` coordinates parsing → generation flow

**Changes**:
1. Accept `--count-config` flag from main
2. Load CountConfiguration if flag provided
3. Build `rowCounts` map (either from config or uniform from `-n`)
4. Pass map to generators
5. Collect and emit CardinalityWarnings after generation

```go
// Orchestrator modifications
type Orchestrator struct {
    // ... existing fields
    countConfig *config.CountConfiguration
}

func (o *Orchestrator) BuildRowCountsMap(sorEntities []Entity, nFlag int) map[string]int {
    if o.countConfig != nil {
        // Use config file
        return o.countConfig.EntityCounts
    }
    // Use uniform -n flag
    counts := make(map[string]int, len(sorEntities))
    for _, entity := range sorEntities {
        counts[entity.ExternalID] = nFlag
    }
    return counts
}
```

## Validation Rules Summary

### Configuration File Validation

| Rule | Check | Error Message |
|------|-------|---------------|
| **File exists** | os.Stat(path) | "Count configuration file not found: {path}" |
| **Valid YAML** | yaml.Unmarshal() | "Invalid YAML syntax in {file}:{line}:{col}: {details}" |
| **Entity exists** | entity in SOR | "Entity '{id}' in count configuration not found in SOR" |
| **Positive count** | count > 0 | "Invalid count for entity '{id}': {value} (expected positive integer)" |
| **Integer type** | Type assertion | "Invalid count type for entity '{id}': {type} (expected integer)" |

### Flag Validation

| Rule | Check | Error Message |
|------|-------|---------------|
| **Mutual exclusion** | -n XOR --count-config | "Cannot use both -n flag and --count-config file" |
| **Config readable** | File permissions | "Cannot read count configuration file: {path}" |

## Performance Characteristics

### Memory Usage

**CountConfiguration**:
- Map overhead: ~48 bytes per entry (Go map baseline)
- 50 entities: ~2.4 KB
- 1000 entities: ~48 KB
- **Negligible impact**

**ConfigurationTemplate**:
- Transient (exists only during template generation)
- 50 entities with descriptions: ~10 KB
- Written to stdout, immediately GC'd
- **No sustained memory impact**

### Time Complexity

**LoadConfiguration**: O(n) where n = entities in config file
**Validate**: O(m) where m = entities in SOR (map lookup)
**GetCount**: O(1) (map access)
**Template.Render**: O(n) where n = entities in SOR

**All operations sub-millisecond for expected scales (<1000 entities)**

## Testing Considerations

### Unit Test Coverage

**CountConfiguration**:
- Load valid YAML → success
- Load invalid YAML → parse error
- Load non-existent file → file not found error
- Validate against matching entities → success
- Validate against mismatched entities → validation error
- GetCount with existing entity → configured value
- GetCount with missing entity → default value
- GetCount with zero value in map → default value

**ConfigurationTemplate**:
- Render with entities → valid YAML output
- Render with empty entities → valid empty YAML
- Render with descriptions → comments present
- WriteTo stdout → output captured and verified

**ValidationError**:
- Error() format includes message and suggestion
- Different error types format correctly

### Integration Test Scenarios

1. **End-to-end with config file**:
   - Load SOR + config → generate CSVs → verify row counts match config

2. **Backward compatibility**:
   - Run with -n 50 (no config) → verify uniform 50 rows

3. **Mixed defaults**:
   - Config specifies 2 entities, SOR has 5 → verify 2 custom, 3 default

4. **Cardinality warnings**:
   - Config creates impossible cardinality → verify warning emitted, CSV generated

## Next Steps

- Create contract/schema file (count-config-schema.yaml)
- Generate quickstart guide
- Update agent context with new packages
