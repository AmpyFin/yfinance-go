# yfinance-go Makefile

# Variables
BINARY_NAME=yfin
MAIN_PATH=./cmd/yfin
DIST_DIR=dist
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE?=$(shell date -u +%Y-%m-%d)
LDFLAGS=-ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for GOOS in linux darwin; do \
		for GOARCH in amd64 arm64; do \
			OUT="$(DIST_DIR)/$(BINARY_NAME)_$${GOOS}_$${GOARCH}"; \
			echo "Building $$OUT..."; \
			GOOS=$$GOOS GOARCH=$$GOARCH CGO_ENABLED=0 go build $(LDFLAGS) -trimpath -o "$$OUT" $(MAIN_PATH); \
			tar -C $(DIST_DIR) -czf "$$OUT.tar.gz" "$$(basename $$OUT)"; \
			rm "$$OUT"; \
		done; \
	done

# Snapshot build (no tag, for testing)
.PHONY: release-snapshot
release-snapshot:
	@echo "Building snapshot release..."
	@mkdir -p $(DIST_DIR)
	GOFLAGS="" go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_snapshot $(MAIN_PATH)

# Tag a release (example: make release-tag v=1.0.0)
.PHONY: release-tag
release-tag:
	@test -n "$(v)" || (echo "Usage: make release-tag v=X.Y.Z"; exit 1)
	@echo "Tagging release v$(v)..."
	git tag v$(v)
	git push origin v$(v)

# Generate checksums for local builds
.PHONY: checksums
checksums:
	@echo "Generating checksums..."
	@if [ ! -d "$(DIST_DIR)" ]; then echo "Error: $(DIST_DIR) directory not found. Run 'make build-all' first."; exit 1; fi
	@if [ -z "$$(ls $(DIST_DIR)/*.tar.gz 2>/dev/null)" ]; then echo "Error: No .tar.gz files found in $(DIST_DIR). Run 'make build-all' first."; exit 1; fi
	(cd $(DIST_DIR) && shasum -a 256 *.tar.gz > checksums.txt)
	@echo "Checksums generated in $(DIST_DIR)/checksums.txt"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linting
.PHONY: lint
lint:
	@echo "Running linters..."
	golangci-lint run

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Tidy modules
.PHONY: tidy
tidy:
	@echo "Tidying modules..."
	go mod tidy
	go mod verify

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(DIST_DIR)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod verify

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build           - Build the binary"
	@echo "  build-all       - Build for all platforms (linux/macos, amd64/arm64)"
	@echo "  release-snapshot- Build snapshot release (no tag)"
	@echo "  release-tag     - Tag a release (usage: make release-tag v=1.0.0)"
	@echo "  checksums       - Generate checksums for built binaries"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  lint            - Run linters"
	@echo "  fmt             - Format code"
	@echo "  tidy            - Tidy and verify modules"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Install dependencies"
	@echo "  help            - Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION         - Version to build (default: dev)"
	@echo "  COMMIT          - Git commit hash (default: auto-detected)"
	@echo "  DATE            - Build date (default: auto-generated)"
	@echo ""
	@echo "Examples:"
	@echo "  make build                    # Build with default version"
	@echo "  make build VERSION=v1.0.0    # Build with specific version"
	@echo "  make build-all               # Build for all platforms"
	@echo "  make release-tag v=1.0.0     # Tag and push release"
	@echo "  make test-coverage           # Run tests with coverage"
