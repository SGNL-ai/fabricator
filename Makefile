.PHONY: build test clean lint fmt vet ci help

# Binary name
BINARY_NAME=fabricator
# Build directory
BUILD_DIR=build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt

# Lint parameters
GOLINT=golangci-lint
GOLINT_VERSION=v2.1.5

# Source files
SOURCE_DIRS=./cmd/... ./pkg/...

all: help

# Build the project
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/fabricator

# Run tests
test:
	$(GOTEST) -v $(SOURCE_DIRS)

# Run tests with coverage
coverage:
	$(GOTEST) -v -coverprofile=coverage.out.tmp $(SOURCE_DIRS)
	# Filter out main.go from coverage calculation
	grep -v "/main.go" coverage.out.tmp > coverage.out
	$(GOCMD) tool cover -func=coverage.out

# Check if code coverage meets the threshold (90%)
coverage-check:
	$(GOTEST) -coverprofile=coverage.out.tmp $(SOURCE_DIRS)
	# Filter out main.go from coverage calculation
	grep -v "/main.go" coverage.out.tmp > coverage.out
	@coverage=`$(GOCMD) tool cover -func=coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+'`; \
	echo "Total coverage: $$coverage%"; \
	if [ $$(echo "$$coverage < 90.0" | bc -l) -eq 1 ]; then \
		echo "Code coverage is below 90%"; \
		exit 1; \
	fi

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out*

# Code formatting
fmt:
	$(GOFMT) $(SOURCE_DIRS)

# Static code analysis
vet:
	$(GOVET) $(SOURCE_DIRS)

# Run linter
lint:
	$(GOLINT) run $(SOURCE_DIRS)

# Run all CI checks
ci: fmt vet lint coverage-check build

# Display help information
help:
	@echo "Available targets:"
	@echo "  build         - Build the fabricator binary"
	@echo "  test          - Run tests"
	@echo "  coverage      - Run tests with coverage and display the result"
	@echo "  coverage-check - Run tests with coverage and ensure it's at least 90% (excluding main.go)"
	@echo "  clean         - Remove build artifacts"
	@echo "  fmt           - Run go fmt"
	@echo "  vet           - Run go vet"
	@echo "  lint          - Run golangci-lint"
	@echo "  ci            - Run all CI checks: fmt, vet, lint, coverage-check, build"
	@echo "  help          - Display this help message"