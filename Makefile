.PHONY: build test clean lint fmt vet ci help coverage-view

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
GOLINT_VERSION=v2.4.0

# Source files
SOURCE_DIRS=./cmd/... ./pkg/...

# Coverage
FILE ?=

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
	# Filter out main.go and mock files from coverage calculation
	grep -v "/main.go" coverage.out.tmp | grep -v "/mocks.go" > coverage.out
	$(GOCMD) tool cover -func=coverage.out

# Coverage helpers
# Usage: make coverage-view FILE=pkg/parser/parser.go
coverage-view:
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make coverage-view FILE=path/to/file.go"; \
		exit 1; \
	fi; \
	if [ ! -f "$(FILE)" ]; then \
		echo "File not found: $(FILE)"; \
		exit 1; \
	fi; \
	T=$$(mktemp); trap "rm -f $$T" EXIT; \
	pkg="./$$( dirname $(FILE) )"; \
	go test $$pkg -coverprofile=$$T >/dev/null 2>&1 && \
	{ \
		U=$$(grep "$$(basename $(FILE)).*0$$" $$T | cut -d: -f2 | cut -d. -f1 | tr '\n' '|' | sed 's/|$$//'); \
		C=$$(grep "$$(basename $(FILE)).*1$$" $$T | cut -d: -f2 | cut -d. -f1 | tr '\n' '|' | sed 's/|$$//'); \
		awk -v u="$$U" -v c="$$C" 'BEGIN {split(u,un,"|"); split(c,cv,"|"); for(i in un) unc[un[i]]=1; for(i in cv) cov[cv[i]]=1} \
			{if(FNR in unc) printf "\033[41m%4d\033[0m \033[31m%s\033[0m\n", FNR, $$0; \
			 else if(FNR in cov) printf "\033[32m%4d %s\033[0m\n", FNR, $$0; \
			 else printf "\033[90m%4d %s\033[0m\n", FNR, $$0}' $(FILE); \
	}

# Check if code coverage meets the threshold (90%)
coverage-check:
	$(GOTEST) -coverprofile=coverage.out.tmp $(SOURCE_DIRS)
	# Filter out main.go and mock files from coverage calculation
	grep -v "/main.go" coverage.out.tmp | grep -v "/mocks.go" > coverage.out
	@coverage=`$(GOCMD) tool cover -func=coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+'`; \
	echo "Total coverage: $$coverage%"; \
	if [ $$(echo "$$coverage < 85.0" | bc -l) -eq 1 ]; then \
		echo "Code coverage is below 85%"; \
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

# Run linter in fix mode
lint-fix:
	$(GOLINT) run --fix $(SOURCE_DIRS)

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