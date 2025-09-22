# Contributing to Fabricator

Thank you for your interest in contributing to Fabricator! This document provides guidelines and information for contributors.

## 🚀 Quick Start

### Prerequisites

- **Go 1.23+** (tested with 1.23 and 1.24)
- **Git** for version control
- **Make** for build automation
- **golangci-lint** for code quality (installed automatically by CI)

### Development Setup

```bash
# Clone the repository
git clone https://github.com/SGNL-ai/fabricator.git
cd fabricator

# Install dependencies
go mod download

# Run tests to verify setup
make test

# Build the project
make build

# Run the tool
./build/fabricator --help
```

### Optional Development Tools

```bash
# Install pre-commit hooks for code quality
pip install pre-commit
pre-commit install

# Install gosec for security scanning
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Install govulncheck for vulnerability scanning
go install golang.org/x/vuln/cmd/govulncheck@latest
```

## 🏗️ Project Architecture

### Overview

Fabricator uses a clean pipeline architecture for CSV data generation:

```
YAML Input → Schema Validation → Business Logic Validation → Pipeline Processing → CSV Output
```

### Core Components

1. **Parser** (`pkg/fabricator/`): YAML parsing and validation
   - JSON Schema validation for structure/syntax
   - Business logic validation for relationships and entities

2. **Pipeline** (`pkg/generators/pipeline/`): 3-phase data generation
   - **Phase 1**: ID generation in topological order
   - **Phase 2**: Relationship linking with cardinality detection
   - **Phase 3**: Field generation for remaining attributes

3. **Model** (`pkg/generators/model/`): Data structures and graph representation
   - Entity, Attribute, and Relationship abstractions
   - Graph structure with dependency management

### Testing Strategy

- **Unit Tests**: Individual component testing with testify framework
- **Integration Tests**: End-to-end CSV generation workflows
- **Real-world Testing**: Validation against 24+ production SGNL templates
- **Coverage Target**: 80% minimum (excluding main.go)

## 🔄 Development Workflow

### 1. Branching Strategy

- **main**: Production-ready code (protected)
- **feature/**: New functionality (`feature/add-json-export`)
- **bugfix/**: Bug fixes (`bugfix/fix-relationship-parsing`)

### 2. Making Changes

1. **Create a feature branch**:
   ```bash
   git checkout main
   git pull origin main
   git checkout -b feature/your-feature-name
   ```

2. **Follow TDD approach**:
   ```bash
   # Write failing tests first
   go test ./pkg/your-package -v

   # Implement functionality
   # Re-run tests until passing
   make test
   ```

3. **Ensure code quality**:
   ```bash
   # Run all checks
   make ci

   # Check security
   gosec ./...
   govulncheck ./...
   ```

4. **Commit with descriptive messages**:
   ```bash
   git add .
   git commit -m "Add feature: brief description

   - Detailed change 1
   - Detailed change 2

   Fixes #123"
   ```

### 3. Pull Request Process

1. **Push your branch** and create a PR
2. **All CI checks must pass** (tests, linting, security)
3. **Code review** by maintainers
4. **Merge** after approval

## 📝 Coding Standards

### Go Style Guidelines

- **Follow Go idioms**: Use `gofmt`, follow effective Go practices
- **Error handling**: Always check errors, use `fmt.Errorf` with `%w` for wrapping
- **Naming**: Clear, descriptive names; avoid abbreviations
- **Comments**: Document all exported functions with GoDoc format
- **Testing**: Use testify for assertions, table-driven tests preferred

### Example Code Structure

```go
// ProcessEntities processes a list of entities according to the specified rules.
// It returns the number of entities processed and any error encountered.
func ProcessEntities(entities []Entity, rules Rules) (int, error) {
    if len(entities) == 0 {
        return 0, fmt.Errorf("no entities provided")
    }

    processed := 0
    for _, entity := range entities {
        if err := validateEntity(entity); err != nil {
            return processed, fmt.Errorf("failed to validate entity %s: %w", entity.ID, err)
        }
        processed++
    }

    return processed, nil
}
```

### Testing Guidelines

```go
func TestProcessEntities(t *testing.T) {
    tests := []struct {
        name     string
        entities []Entity
        rules    Rules
        want     int
        wantErr  bool
    }{
        {
            name:     "Valid entities",
            entities: []Entity{{ID: "test1"}, {ID: "test2"}},
            rules:    DefaultRules,
            want:     2,
            wantErr:  false,
        },
        {
            name:     "Empty entities",
            entities: []Entity{},
            rules:    DefaultRules,
            want:     0,
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ProcessEntities(tt.entities, tt.rules)

            if tt.wantErr {
                assert.Error(t, err)
                return
            }

            assert.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## 🐛 Bug Reports

When reporting bugs, please include:

1. **Go version** (`go version`)
2. **Operating system** and architecture
3. **Fabricator version** (`./fabricator --version`)
4. **Minimal reproduction case**
5. **Expected vs actual behavior**
6. **Full error output** if applicable

### Example Bug Report

```
**Environment:**
- Go version: 1.24.3
- OS: macOS 14.0 (arm64)
- Fabricator version: v1.2.3

**Issue:**
CSV generation fails with relationship validation error

**Reproduction:**
```bash
./fabricator -f my-template.yaml -n 100 -o output/
```

**Expected:** CSV files generated successfully
**Actual:** Error: "relationship user_role: missing fromAttribute"

**Template:** [attach YAML file]
```

## 🆕 Feature Requests

Feature requests should include:

1. **Use case description** - what problem does this solve?
2. **Proposed solution** - how should it work?
3. **Alternatives considered** - what other approaches were evaluated?
4. **Implementation notes** - any technical considerations

## 📦 Release Process

Releases are automated via GitHub Actions:

1. **Merge to main** triggers automated release
2. **Semantic versioning** (patch increments automatically)
3. **Multi-platform binaries** generated for Linux, macOS, Windows
4. **Release notes** generated automatically from commits

### Manual Release (Maintainers Only)

```bash
# Tag a specific version
git tag v1.2.3
git push origin v1.2.3

# GitHub Actions will handle the rest
```

## 🏗️ Architecture Decisions

### Why This Architecture?

- **Pipeline Pattern**: Clean separation of concerns, testable components
- **Interface-Driven**: Easy mocking and testing
- **JSON Schema**: Industry standard, maintainable validation
- **Graph Model**: Handles complex entity relationships efficiently

### Design Principles

- **Composition over Inheritance**: Use interfaces and dependency injection
- **Explicit Error Handling**: No silent failures, clear error messages
- **Testability**: Every component can be unit tested in isolation
- **Performance**: Generate large datasets efficiently

## 🤝 Community

- **Issues**: Bug reports and feature requests
- **Discussions**: Questions and general discussion
- **Wiki**: Additional documentation and examples

## 📄 License

This project is licensed under the Apache License 2.0. See [LICENSE](LICENSE) for details.

---

**Questions?** Open an issue or start a discussion. We're here to help! 🎉