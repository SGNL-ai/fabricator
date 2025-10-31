# Implementation Plan: Per-Entity Row Count Configuration

**Branch**: `001-per-entity-row-counts` | **Date**: 2025-10-30 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-per-entity-row-counts/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Enhance fabricator to support per-entity row count configuration via YAML files, replacing the one-size-fits-all `-n` parameter with flexible entity-specific counts. Add `init-count-config` subcommand to generate configuration templates from SOR YAML files. Maintain backward compatibility with existing `-n` parameter.

**Key Capabilities**:
- Accept row count configuration YAML mapping entity external_id to count
- Generate row counts per entity while maintaining relationship consistency
- Provide `init-count-config` subcommand to bootstrap configuration files
- Preserve existing `-n` parameter behavior for uniform row counts
- Best-effort relationship handling with warnings for cardinality violations

## Technical Context

**Language/Version**: Go 1.25
**Primary Dependencies**: gopkg.in/yaml.v3 (existing), github.com/fatih/color (existing), github.com/stretchr/testify (testing)
**Storage**: File-based (YAML input, CSV output)
**Testing**: Go test with testify framework (80% coverage minimum)
**Target Platform**: Cross-platform CLI (macOS, Linux, Windows)
**Project Type**: Single CLI application
**Performance Goals**: Generate templates <5 seconds for 50 entities; maintain current CSV generation performance
**Constraints**: Backward compatible with existing `-n` flag; configuration file optional; stdout for template output
**Scale/Scope**: Support SOR YAML files with up to 50+ entities; handle row counts from 1 to 1M+ per entity

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Phase 0 Evaluation

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. Test-Driven Development** | ✅ PASS | Plan includes test-first approach for all new functionality |
| **II. Go Best Practices** | ✅ PASS | Will follow idiomatic Go patterns for new packages/functions |
| **III. Explicit Relationship Handling** | ✅ PASS | Feature enhances row count flexibility; relationship logic unchanged |
| **IV. Data Integrity & Consistency** | ⚠️ REVIEW | Best-effort cardinality handling may generate incomplete relationships (spec-approved) |
| **V. Performance & Quality** | ✅ PASS | No performance regression expected; template generation <5s |

**Gate Decision**: PASS with review note on Principle IV

**Justification for Principle IV Review**:
The specification explicitly requires best-effort relationship assignment with warnings when cardinality constraints cannot be satisfied (FR-006). This is intentional UX design to allow users flexibility in row count configuration without strict enforcement. The system will:
- Attempt to maintain referential integrity where possible
- Emit clear warnings when constraints violated
- Generate valid CSV files (no corrupt data)
- Document limitations in user-facing messages

This approach prioritizes user flexibility over strict correctness, which is acceptable for a test data generation tool.

### Post-Phase 1 Re-evaluation

**Design Artifacts Completed**:
- ✅ research.md - All technical decisions documented
- ✅ data-model.md - Data structures and flows defined
- ✅ contracts/count-config-schema.yaml - YAML schema specified
- ✅ quickstart.md - User guide created

**Constitution Re-evaluation**:

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. Test-Driven Development** | ✅ PASS | data-model.md includes comprehensive test coverage plan; 80% coverage target maintained |
| **II. Go Best Practices** | ✅ PASS | New packages follow Go conventions (config, subcommands); clean separation of concerns |
| **III. Explicit Relationship Handling** | ✅ PASS | No changes to relationship inference logic; row counts independent of relationship semantics |
| **IV. Data Integrity & Consistency** | ✅ PASS (with noted exception) | Best-effort approach documented and approved in spec; warnings mitigate risk |
| **V. Performance & Quality** | ✅ PASS | Performance analysis shows negligible impact (<1ms for typical operations); no new bottlenecks |

**Final Gate Decision**: ✅ PASS

**Design Quality Assessment**:
- **Simplicity**: Two new packages (config, subcommands) with focused responsibilities
- **Testability**: All components have clear interfaces suitable for unit testing
- **Maintainability**: Minimal changes to existing packages; backward compatible
- **Performance**: Sub-millisecond operations; no regression risk
- **Extensibility**: Subcommands pattern supports future additions (validate, schema, etc.)

**Risk Mitigation**:
- CardinalityWarning system provides clear user feedback for constraint violations
- Validation errors include actionable suggestions (90% fix-on-first-try target)
- Backward compatibility preserved (all existing commands unchanged)
- Template generation prevents manual YAML errors

## Project Structure

### Documentation (this feature)

```text
specs/001-per-entity-row-counts/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── count-config-schema.yaml  # YAML schema for row count configuration
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
# Existing structure (unchanged)
cmd/
└── fabricator/
    └── main.go          # MODIFY: Add init-count-config subcommand, --count-config flag

pkg/
├── fabricator/
│   └── parser.go        # EXISTING: YAML parsing
├── generators/
│   └── csv_generator.go # MODIFY: Accept per-entity row counts
├── orchestrator/
│   └── orchestrator.go  # MODIFY: Pass row count map to generators
├── parser/              # EXISTING: SOR YAML parsing
└── util/                # EXISTING: Utilities

# New structure for this feature
pkg/
├── config/              # NEW PACKAGE
│   ├── count_config.go      # Row count configuration parsing
│   ├── count_config_test.go # Unit tests
│   ├── validator.go         # Configuration validation
│   ├── validator_test.go    # Validation tests
│   ├── template.go          # Template generation logic
│   └── template_test.go     # Template generation tests
└── subcommands/         # NEW PACKAGE
    ├── init_count_config.go      # init-count-config subcommand implementation
    └── init_count_config_test.go # Subcommand tests

tests/
└── integration/         # EXISTING
    └── per_entity_counts_test.go  # NEW: End-to-end integration tests
```

**Structure Decision**: Single project structure maintained. New `config` package handles row count configuration parsing, validation, and template generation. New `subcommands` package provides clean separation for subcommand implementations (future-proofing for additional subcommands). Existing packages modified minimally to accept per-entity counts.

## Complexity Tracking

No constitution violations requiring justification. All changes align with existing principles and patterns.

