# Tasks: Per-Entity Row Count Configuration

**Input**: Design documents from `/specs/001-per-entity-row-counts/`
**Prerequisites**: plan.md, spec.md, data-model.md, research.md, contracts/count-config-schema.yaml

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

**Project Constitution**: This feature follows TDD (Test-Driven Development) as mandated by the project constitution. All tests MUST be written and approved BEFORE implementation begins.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

This is a single Go CLI project with structure:
- `cmd/fabricator/` - Main CLI entry point
- `pkg/` - Reusable packages
- `tests/integration/` - Integration tests

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create new package structure for configuration management

- [x] T001 [P] Create `pkg/config/` directory structure
- [x] T002 [P] Create `pkg/subcommands/` directory structure
- [x] T003 [P] Create `tests/fixtures/` directory for test configuration files

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Define ValidationError struct in `pkg/config/validator.go`
- [x] T005 Write tests for ValidationError in `pkg/config/validator_test.go`
- [x] T006 Implement ValidationError.Error() method in `pkg/config/validator.go`
- [x] T007 Define CountConfiguration struct in `pkg/config/count_config.go`
- [x] T008 Define ConfigurationTemplate and TemplateEntity structs in `pkg/config/template.go`

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Generate Custom Row Counts per Entity (Priority: P1) 🎯 MVP

**Goal**: Users can provide a YAML configuration file specifying different row counts per entity, and fabricator generates CSV files with exactly those row counts while maintaining relationship consistency.

**Independent Test**: Provide a row count configuration file and verify that each entity CSV contains the exact number of rows specified for that entity.

**Acceptance Criteria**:
1. Given a valid SOR YAML file with 3 entities, when user provides a row count configuration specifying 100 users, 20 groups, and 50 permissions, then the generated CSVs contain exactly those row counts
2. Given a row count configuration file, when user runs fabricator with this configuration, then relationship consistency is maintained across all entities regardless of differing row counts
3. Given a row count configuration that specifies counts for only some entities, when user runs fabricator, then entities without specified counts use a default row count

### Tests for User Story 1 (TDD - WRITE FIRST)

> **CONSTITUTION REQUIREMENT**: Write these tests FIRST, ensure they FAIL before implementation

- [x] T009 [P] [US1] Write test for LoadConfiguration with valid YAML in `pkg/config/count_config_test.go`
- [x] T010 [P] [US1] Write test for LoadConfiguration with invalid YAML in `pkg/config/count_config_test.go`
- [x] T011 [P] [US1] Write test for LoadConfiguration with non-existent file in `pkg/config/count_config_test.go`
- [x] T012 [P] [US1] Write test for GetCount with existing entity in `pkg/config/count_config_test.go`
- [x] T013 [P] [US1] Write test for GetCount with missing entity (should return default) in `pkg/config/count_config_test.go`
- [x] T014 [P] [US1] Write test for GetCount with zero value in map (should return default) in `pkg/config/count_config_test.go`
- [x] T015 [P] [US1] Write test for Validate against matching entities (success) in `pkg/config/count_config_test.go`
- [x] T016 [P] [US1] Write test for Validate against mismatched entities (error) in `pkg/config/count_config_test.go`
- [x] T017 [P] [US1] Write test for Validate with negative count values in `pkg/config/count_config_test.go`
- [x] T018 [P] [US1] Write test for Validate with zero count values in `pkg/config/count_config_test.go`
- [x] T019 [P] [US1] Write test for Validate with non-integer count values in `pkg/config/count_config_test.go`

### Implementation for User Story 1

- [x] T020 [P] [US1] Implement LoadConfiguration function in `pkg/config/count_config.go`
- [x] T021 [P] [US1] Implement GetCount method in `pkg/config/count_config.go`
- [x] T022 [P] [US1] Implement HasEntity method in `pkg/config/count_config.go`
- [x] T023 [US1] Implement Validate method in `pkg/config/count_config.go`
- [x] T024 [US1] Implement validation helper functions in `pkg/config/validator.go`
- [x] T025 [US1] Add `--count-config` flag to main.go in `cmd/fabricator/main.go`
- [x] T026 [US1] Add flag conflict validation (reject both -n and --count-config) in `cmd/fabricator/main.go`
- [x] T027 [US1] Modify orchestrator to accept CountConfiguration in `pkg/orchestrator/generation.go`
- [x] T028 [US1] Implement BuildRowCountsMap method in `pkg/orchestrator/generation.go`
- [x] T029 [US1] Update CSV generator signature to accept `map[string]int` in `pkg/generators/pipeline/generator.go`
- [x] T030 [US1] Modify CSV generator to lookup row counts per entity in `pkg/generators/pipeline/id_generator.go`
- [x] T031 [US1] Add cardinality violation detection logic in `pkg/generators/warnings.go`
- [x] T032 [US1] Implement CardinalityWarning struct and String() method in `pkg/generators/warnings.go`
- [x] T033 [US1] Add warning emission to stderr with color in orchestrator after generation in `pkg/orchestrator/generation.go`

### Integration Tests for User Story 1

- [x] T034 [US1] Write integration test for end-to-end with config file in `tests/integration/per_entity_counts_test.go`
- [x] T035 [US1] Write integration test for backward compatibility (-n flag only) in `tests/integration/per_entity_counts_test.go`
- [x] T036 [US1] Write integration test for mixed defaults (partial config) in `tests/integration/per_entity_counts_test.go`
- [x] T037 [US1] Write integration test for cardinality warnings in `tests/integration/per_entity_counts_test.go`
- [x] T038 [US1] Write integration test for conflict detection (CLI level - deferred to manual testing)
- [x] T039 [US1] Create test fixtures (sample count config files) in `tests/fixtures/`

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently. Users can generate CSV files with per-entity row counts via config file.

---

## Phase 4: User Story 3 - Preserve Backward Compatibility with `-n` Parameter (Priority: P1)

**Goal**: Existing users can continue to use the `-n` parameter for uniform row counts across all entities without any workflow changes.

**Independent Test**: Run fabricator with the existing `-n` parameter (without a row count configuration file) and verify behavior is unchanged from previous versions.

**Acceptance Criteria**:
1. Given a valid SOR YAML file, when user runs fabricator with `-n 100` and no row count configuration file, then all entities generate exactly 100 rows as they did before this feature
2. Given both `-n` parameter and a row count configuration file are provided, when user runs fabricator, then an error message explains that only one row count method can be used at a time
3. Given neither `-n` parameter nor row count configuration file are provided, when user runs fabricator, then the default row count (100) is used for all entities

**Note**: This user story is largely implemented during User Story 1 (T026, T028, T029, T030 ensure backward compatibility). This phase adds explicit tests to verify the behavior.

### Tests for User Story 3 (TDD - WRITE FIRST)

- [x] T040 [P] [US3] Write integration test for `-n 100` with no config file in `tests/integration/backward_compatibility_test.go`
- [x] T041 [P] [US3] Write integration test for default behavior (no -n, no config) in `tests/integration/backward_compatibility_test.go`
- [x] T042 [P] [US3] Write integration test verifying error message format for conflicting flags in `tests/integration/backward_compatibility_test.go`

### Implementation for User Story 3

- [x] T043 [US3] Verify BuildRowCountsMap handles -n flag correctly (review T028)
- [x] T044 [US3] Verify error message quality for conflicting flags (review T026)
- [x] T045 [US3] Add comprehensive error message tests in `cmd/fabricator/main_test.go` if not already covered

**Checkpoint**: At this point, User Stories 1 AND 3 should both work. Existing users can continue using `-n`, new users can use `--count-config`.

---

## Phase 5: User Story 2 - Generate Row Count Configuration Template (Priority: P2)

**Goal**: Users can generate a row count configuration template from their SOR YAML file to bootstrap their configuration without manual writing.

**Independent Test**: Run the stub generation command with a SOR YAML file and verify the output contains all entities with default placeholder values.

**Acceptance Criteria**:
1. Given a valid SOR YAML file with multiple entities, when user runs `fabricator init-count-config -f sor.yaml`, then a row count configuration template is written to stdout containing all entities from the SOR with default row count values
2. Given a generated row count configuration template, when user opens the file, then the file is properly formatted, human-readable, and includes helpful comments explaining how to customize values
3. Given a generated template with default values, when user runs fabricator without modifying the template, then data is generated successfully using the default row counts

### Tests for User Story 2 (TDD - WRITE FIRST)

- [x] T046 [P] [US2] Write test for ConfigurationTemplate.Render with entities in `pkg/config/template_test.go`
- [x] T047 [P] [US2] Write test for ConfigurationTemplate.Render with empty entities in `pkg/config/template_test.go`
- [x] T048 [P] [US2] Write test for ConfigurationTemplate.Render with descriptions in `pkg/config/template_test.go`
- [x] T049 [P] [US2] Write test for ConfigurationTemplate.WriteTo stdout in `pkg/config/template_test.go`
- [x] T050 [P] [US2] Write test for NewTemplate factory function in `pkg/config/template_test.go`
- [x] T051 [P] [US2] Write test for init-count-config subcommand with valid SOR file in `pkg/subcommands/init_count_config_test.go`
- [x] T052 [P] [US2] Write test for init-count-config subcommand with invalid SOR file in `pkg/subcommands/init_count_config_test.go`
- [x] T053 [P] [US2] Write test for init-count-config subcommand with missing SOR file in `pkg/subcommands/init_count_config_test.go`

### Implementation for User Story 2

- [x] T054 [P] [US2] Implement NewTemplate factory function in `pkg/config/template.go`
- [x] T055 [P] [US2] Implement ConfigurationTemplate.Render method in `pkg/config/template.go`
- [x] T056 [P] [US2] Implement ConfigurationTemplate.WriteTo method in `pkg/config/template.go`
- [x] T057 [US2] Implement InitCountConfig subcommand handler in `pkg/subcommands/init_count_config.go`
- [x] T058 [US2] Add subcommand routing to main.go (check os.Args[1] for "init-count-config") in `cmd/fabricator/main.go`
- [x] T059 [US2] Add help text for init-count-config subcommand to usage output in `cmd/fabricator/main.go`

### Integration Tests for User Story 2

- [x] T060 [US2] Write integration test for end-to-end template generation in `tests/integration/template_generation_test.go`
- [x] T061 [US2] Write integration test for generated template is valid YAML in `tests/integration/template_generation_test.go`
- [x] T062 [US2] Write integration test for generated template can be used immediately in `tests/integration/template_generation_test.go`

**Checkpoint**: All three user stories should now be independently functional. Users can generate templates, customize row counts, and maintain backward compatibility.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories and final quality checks

- [x] T063 [P] Update main README.md with `--count-config` flag documentation
- [x] T064 [P] Update main README.md with `init-count-config` subcommand documentation
- [x] T065 [P] Add examples to `examples/` directory showing count configuration usage
- [x] T066 [P] Verify test coverage meets 80% threshold using `make test` and `go tool cover`
- [x] T067 [P] Run `make lint` and fix any linter warnings
- [x] T068 [P] Run `make vet` and fix any static analysis issues
- [x] T069 Run full CI checks with `make ci`
- [x] T070 Validate quickstart.md examples manually (generate template, customize, generate CSVs)
- [x] T071 Performance test: Generate template for 50-entity SOR file (target: <5 seconds per SC-002)
- [x] T072 Performance test: Verify no regression in CSV generation with row count map vs. uniform int

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational phase completion
- **User Story 3 (Phase 4)**: Depends on User Story 1 (tests backward compatibility of US1 implementation)
- **User Story 2 (Phase 5)**: Can start after Foundational phase - independent of US1/US3
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1) - Per-Entity Counts**: Can start after Foundational (Phase 2) - No dependencies on other stories - **MVP CORE**
- **User Story 3 (P1) - Backward Compatibility**: Depends on User Story 1 - Tests/verifies backward compatibility
- **User Story 2 (P2) - Template Generation**: Can start after Foundational (Phase 2) - Independent of US1/US3 (but requires US1 config parsing to be useful)

### Within Each User Story

**Constitution Requirement: TDD Mandatory**
- Tests MUST be written FIRST and approved before implementation
- Follow Red-Green-Refactor cycle strictly
- All tests MUST fail initially (no implementation yet)
- Implementation proceeds only after tests are written and failing
- Tests pass → Refactor → Move to next task

**Execution Order Within Story**:
1. Write ALL tests for the story (T009-T019, T034-T039 for US1)
2. Verify ALL tests FAIL (expected - no implementation yet)
3. Get test approval/review
4. Implement configuration loading (T020-T024)
5. Implement CLI integration (T025-T026)
6. Implement orchestrator changes (T027-T028)
7. Implement generator changes (T029-T033)
8. Verify ALL tests PASS
9. Refactor if needed
10. Story checkpoint

### Parallel Opportunities

**Phase 1 (Setup)**:
- All three tasks can run in parallel (different directories)

**Phase 2 (Foundational)**:
- T004-T006 can run in parallel with T007-T008 (different files)

**Phase 3 (User Story 1) - Tests**:
- T009-T019 can all run in parallel (different test functions, same or different files)
- T034-T039 can all run in parallel (different test functions)

**Phase 3 (User Story 1) - Implementation**:
- T020-T022 can run in parallel (same file, different methods)
- After T020-T024 complete: T025-T026, T027-T028, T029-T030 can run in parallel (different packages)
- T031-T033 must run after T029-T030 (same files)

**Phase 4 (User Story 3) - Tests**:
- T040-T042 can all run in parallel

**Phase 5 (User Story 2) - Tests**:
- T046-T050 can run in parallel
- T051-T053 can run in parallel

**Phase 5 (User Story 2) - Implementation**:
- T054-T056 can run in parallel (same file, different methods)
- T057-T059 must run sequentially (integration)

**Phase 6 (Polish)**:
- T063-T068 can all run in parallel (different files/concerns)

---

## Parallel Example: User Story 1 Tests

```bash
# Launch all CountConfiguration tests together (TDD - write first):
Task: "Write test for LoadConfiguration with valid YAML in pkg/config/count_config_test.go"
Task: "Write test for LoadConfiguration with invalid YAML in pkg/config/count_config_test.go"
Task: "Write test for LoadConfiguration with non-existent file in pkg/config/count_config_test.go"
Task: "Write test for GetCount with existing entity in pkg/config/count_config_test.go"
Task: "Write test for GetCount with missing entity in pkg/config/count_config_test.go"
Task: "Write test for GetCount with zero value in map in pkg/config/count_config_test.go"

# Launch all Validator tests together (TDD - write first):
Task: "Write test for Validate against matching entities in pkg/config/validator_test.go"
Task: "Write test for Validate against mismatched entities in pkg/config/validator_test.go"
Task: "Write test for Validate with negative count values in pkg/config/validator_test.go"
Task: "Write test for Validate with zero count values in pkg/config/validator_test.go"
Task: "Write test for Validate with non-integer count values in pkg/config/validator_test.go"
```

---

## Parallel Example: User Story 1 Implementation (After Tests Pass)

```bash
# After ALL tests written and failing, launch core implementations:
Task: "Implement LoadConfiguration function in pkg/config/count_config.go"
Task: "Implement GetCount method in pkg/config/count_config.go"
Task: "Implement HasEntity method in pkg/config/count_config.go"

# Then launch package integrations in parallel:
Task: "Add --count-config flag to main.go in cmd/fabricator/main.go"
Task: "Modify orchestrator to accept CountConfiguration in pkg/orchestrator/orchestrator.go"
Task: "Update CSV generator signature in pkg/generators/csv_generator.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 + User Story 3)

**Goal**: Deliver per-entity row count functionality with backward compatibility

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T008) - **CRITICAL BLOCKER**
3. Write ALL tests for User Story 1 (T009-T019, T034-T039) - **TDD REQUIREMENT**
4. Verify tests FAIL - **TDD CHECKPOINT**
5. Complete Phase 3: User Story 1 Implementation (T020-T039)
6. Verify ALL tests PASS - **TDD CHECKPOINT**
7. Complete Phase 4: User Story 3 Tests + Verification (T040-T045)
8. **STOP and VALIDATE**: Test User Story 1 + 3 independently
   - Run with config file → verify per-entity counts
   - Run with `-n` flag → verify uniform counts (backward compatible)
   - Run with both → verify error
   - Run with neither → verify default 100
9. Deploy/demo if ready (MVP complete!)

### Incremental Delivery

1. **Foundation** (Phases 1-2) → Infrastructure ready
2. **MVP** (Phases 3-4: US1 + US3) → Core feature + backward compatibility
   - Test independently → Deploy/Demo
   - **Value delivered**: Users can specify per-entity counts via config file, existing `-n` workflows unchanged
3. **Enhancement** (Phase 5: US2) → Template generation
   - Test independently → Deploy/Demo
   - **Value delivered**: Users can bootstrap config files without manual writing
4. **Polish** (Phase 6) → Documentation, performance, quality
   - Final validation → Deploy/Demo

### Parallel Team Strategy

With multiple developers:

1. **Team completes Setup + Foundational together** (Phases 1-2)
2. **Once Foundational is done, split work**:
   - Developer A: User Story 1 (Phase 3) - Core feature
   - Developer B: User Story 2 (Phase 5) - Template generation (independent)
   - User Story 3 (Phase 4) can be done by Developer A after US1 (tests backward compatibility of A's work)
3. Stories complete and integrate independently
4. Team reconvenes for Polish (Phase 6)

---

## Task Count Summary

- **Phase 1 (Setup)**: 3 tasks
- **Phase 2 (Foundational)**: 5 tasks
- **Phase 3 (US1)**: 31 tasks (11 tests + 14 implementation + 6 integration tests)
- **Phase 4 (US3)**: 6 tasks (3 tests + 3 verification)
- **Phase 5 (US2)**: 17 tasks (8 tests + 6 implementation + 3 integration tests)
- **Phase 6 (Polish)**: 10 tasks

**Total**: 72 tasks

**Tests**: 28 test tasks (following TDD)
**Implementation**: 44 implementation + verification tasks
**Test-to-Implementation Ratio**: ~39% (healthy TDD ratio)

---

## Notes

- **[P] tasks** = different files or independent test functions, no dependencies
- **[Story] label** maps task to specific user story for traceability
- **Each user story** should be independently completable and testable
- **TDD Requirement**: Verify tests FAIL before implementing (Red phase)
- **TDD Requirement**: Verify tests PASS after implementing (Green phase)
- **Constitution**: 80% test coverage minimum - verify with `make test` and `go tool cover`
- **Commit strategy**: Commit after logical groups (e.g., all tests for one component, one complete implementation unit)
- **Stop at any checkpoint** to validate story independently
- **Avoid**: Vague tasks, same file conflicts, cross-story dependencies that break independence
- **Performance targets**: Template generation <5 seconds for 50 entities, no CSV generation regression

---

## Validation Checklist

Before marking feature complete:

- [ ] All 72 tasks completed
- [ ] All tests passing (`make test`)
- [ ] Test coverage ≥ 80% (`go tool cover -html=coverage.out`)
- [ ] No linter warnings (`make lint`)
- [ ] No vet issues (`make vet`)
- [ ] CI passing (`make ci`)
- [ ] User Story 1 independently testable (per-entity counts work)
- [ ] User Story 3 independently testable (backward compatibility preserved)
- [ ] User Story 2 independently testable (template generation works)
- [ ] Quickstart.md examples validated manually
- [ ] Performance targets met (SC-002: template generation <5 seconds for 50 entities)
- [ ] No performance regression in CSV generation
- [ ] README.md updated with new flags and subcommand
- [ ] Examples directory includes count configuration samples
