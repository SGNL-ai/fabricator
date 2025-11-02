# Fabricator Constitution

## Core Principles

### I. Test-Driven Development (NON-NEGOTIABLE)
TDD is mandatory for all code changes:
- Tests must be written and user-approved BEFORE implementation
- Follow strict Red-Green-Refactor cycle
- Acceptable test coverage is 80% or better using testify framework
- Never skip work or justify incomplete functionality with "it likely works"
- Act as a professional engineer - no shortcuts
- **NEVER use t.Skip() to hide failing tests when implementation is broken**
  - t.Skip() is ONLY acceptable when the test infrastructure itself is broken (missing external dependencies, unavailable services, etc.)
  - When implementation is broken, tests MUST fail - that is the entire point of TDD
  - Skipping tests because "code isn't ready yet" defeats Red-Green-Refactor
  - Failing tests drive fixes; skipped tests hide problems

Testing hierarchy:
- **Unit tests**: Test individual functions and methods in isolation
- **Component tests**: Test interactions between related components
- **Integration tests**: Test end-to-end flows from YAML to CSV output

### II. Go Best Practices
Follow idiomatic Go development standards:
- **Package Organization**: Single lowercase package names, one per directory
- **Function Design**: Functions do one thing well, under 50 lines when possible
- **Error Handling**: Always check error returns, never silently fail
- **Variable Naming**: camelCase for internal, PascalCase for exported
- **Comments**: All exported functions require doc comments explaining "why"

### III. Explicit Relationship Handling
Relationships between entities must be explicitly defined:
- **Never infer relationships** from field names
- Do NOT parse field names to determine ID fields or foreign keys
- Field names only guide sample data generation (e.g., "email" → email format)
- YAML relationship structures are the single source of truth
- Cardinality determined from definition or auto-detected when enabled

### IV. Data Integrity & Consistency
Generated data must maintain referential integrity:
- Primary keys must be unique within entities
- Foreign keys must reference valid primary keys in target entities
- Respect cardinality constraints (1:1, 1:N, N:1, N:M)
- Process entities in dependency order
- Validate all relationships post-generation

### V. Performance & Quality
Optimize for efficiency without sacrificing correctness:
- Pre-allocate slices and maps when size is known
- Use buffered I/O for file operations
- Generate data in memory before writing
- Table-driven tests for comprehensive coverage
- Use testify framework for assertions

## Development Workflow

### Branching Model
- **main**: Production-ready code, protected from direct pushes
- **feature/**: New functionality or enhancements
- **bugfix/**: Bug fixes

All changes require:
- Pull request with passing CI checks
- Test coverage meeting 80% threshold
- Code review approval

### Error Handling Standards
1. **Early Validation**: Validate inputs before processing
2. **Graceful Failures**: Provide clear, actionable error messages with context
3. **Appropriate Logging**: Error/Warning/Info/Debug levels used correctly
4. **Explicit Errors**: Always check returns, log specific errors

### Code Review Requirements
All changes must verify:
- Functionality accomplishes stated goals
- Edge cases handled appropriately
- Tests cover happy paths and error cases
- Error messages are clear and actionable
- Performance is acceptable for expected data sizes

## Testing Standards

### Test Requirements
- Use testify/assert and testify/require for all assertions
- Table-driven tests for multiple scenarios
- Test fixtures from examples directory (never generate YAML)
- Each test must be independent and self-contained
- Tests must genuinely verify functionality, not fake passing

### Test Organization Pattern
```go
func TestFunctionName(t *testing.T) {
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
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual, err := FunctionUnderTest(tt.input)

            if tt.wantErr {
                assert.Error(t, err)
                return
            }

            assert.NoError(t, err)
            assert.Equal(t, tt.expected, actual)
        })
    }
}
```

## Security & Compliance

### Security Best Practices
- GitHub token permissions follow principle of least privilege
- Never commit credentials or sensitive data
- Branch protection enforced on main
- CI workflow permissions explicitly limited

### Dependency Management
- Go dependencies: weekly updates (Monday), minor/patch grouped
- GitHub Actions: monthly updates, all grouped
- Major version updates ignored to avoid breaking changes
- Dependabot manages automated updates

## Governance

This constitution supersedes all other development practices and guidelines. All pull requests and code reviews must verify compliance with these principles.

### Amendment Process
- Constitution changes require documentation and approval
- Migration plan required for breaking changes
- Complexity must be justified with clear rationale
- Refer to CLAUDE.md for runtime development guidance

### Quality Gates
- 80% minimum test coverage
- All CI checks must pass
- Code review approval required
- No direct commits to main branch

<!--
Sync Impact Report - Constitution v1.1.0
========================================
Version Change: 1.0.0 → 1.1.0 (MINOR - added explicit test skipping prohibition to Principle I)

Modified Principles:
- Principle I (Test-Driven Development): Added explicit prohibition on t.Skip() for broken implementations

Added Sections: None

Removed Sections: None

Templates Requiring Updates:
- ✅ .specify/templates/plan-template.md: Reviewed - no changes needed (constitution check already enforces TDD)
- ✅ .specify/templates/spec-template.md: Reviewed - no changes needed (spec focuses on requirements, not testing)
- ✅ .specify/templates/tasks-template.md: Reviewed - no changes needed (tasks already mandate TDD test-first approach)
- ✅ .specify/templates/commands/*.md: Reviewed - no references to test skipping practices

Runtime Guidance:
- ✅ CLAUDE.md: Already updated with identical test skipping prohibition (lines 456-459)

Follow-up TODOs: None

Rationale for MINOR bump:
- Added material new guidance (prohibition on t.Skip() misuse)
- Expands existing Principle I with explicit rule
- Does not change existing principles or remove guidance
- Backward compatible (doesn't break existing workflows)
-->

**Version**: 1.1.0 | **Ratified**: 2025-10-30 | **Last Amended**: 2025-10-31
