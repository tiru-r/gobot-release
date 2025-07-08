# Gobot Makefile

# Build variables
BINARY_NAME=gobot
PLATFORMS=linux/amd64 linux/arm64 linux/arm windows/amd64 darwin/amd64 darwin/arm64
DRIVERS=gpio i2c spi onewire ble
EXAMPLES=basic platforms advanced

# Go build flags
BUILD_FLAGS=-ldflags="-s -w"
TEST_FLAGS=-v -race -coverprofile=coverage.out

# Default target
.PHONY: all
all: build

# Build all platforms
.PHONY: build
build:
	go build $(BUILD_FLAGS) ./...

# Build for specific platform  
.PHONY: build-platform
build-platform:
	@if [ -z "$(PLATFORM)" ]; then echo "Usage: make build-platform PLATFORM=linux/amd64"; exit 1; fi
	@echo "Building for $(PLATFORM)..."
	@GOOS=$(shell echo $(PLATFORM) | cut -d'/' -f1) \
	 GOARCH=$(shell echo $(PLATFORM) | cut -d'/' -f2) \
	 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-$(shell echo $(PLATFORM) | tr '/' '-') .

# Build matrix for all platforms
.PHONY: build-matrix  
build-matrix:
	@for platform in $(PLATFORMS); do \
		echo "Building for $$platform..."; \
		GOOS=$$(echo $$platform | cut -d'/' -f1) \
		GOARCH=$$(echo $$platform | cut -d'/' -f2) \
		go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-$$(echo $$platform | tr '/' '-') . || exit 1; \
	done

# Test all packages
.PHONY: test
test:
	go test $(TEST_FLAGS) ./...

# Test specific package
.PHONY: test-package
test-package:
	@if [ -z "$(PACKAGE)" ]; then echo "Usage: make test-package PACKAGE=./pkg/core"; exit 1; fi
	go test $(TEST_FLAGS) $(PACKAGE)

# Test with specific build tags
.PHONY: test-tags
test-tags:
	@if [ -z "$(TAGS)" ]; then echo "Usage: make test-tags TAGS='gpio,i2c'"; exit 1; fi
	go test $(TEST_FLAGS) -tags=$(TAGS) ./...

# Run benchmarks
.PHONY: bench
bench:
	go test -bench=. -benchmem ./...

# Run integration tests
.PHONY: test-integration
test-integration:
	go test -tags=integration $(TEST_FLAGS) ./test/integration/...

# Run end-to-end tests
.PHONY: test-e2e
test-e2e:
	go test -tags=e2e $(TEST_FLAGS) ./test/e2e/...

# Lint code
.PHONY: lint
lint:
	golangci-lint run

# Format code
.PHONY: fmt
fmt:
	go fmt ./...
	goimports -w .

# Tidy dependencies
.PHONY: tidy
tidy:
	go mod tidy

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf bin/
	rm -f coverage.out
	go clean -cache

# Install dependencies
.PHONY: deps
deps:
	go mod download
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run examples
.PHONY: run-example
run-example:
	@if [ -z "$(EXAMPLE)" ]; then echo "Usage: make run-example EXAMPLE=basic/hello"; exit 1; fi
	@if [ ! -d "examples/$(EXAMPLE)" ]; then echo "Example not found: examples/$(EXAMPLE)"; exit 1; fi
	cd examples/$(EXAMPLE) && go run .

# List available examples
.PHONY: list-examples
list-examples:
	@echo "Available examples:"
	@find examples -name "main.go" -type f | sed 's|examples/||' | sed 's|/main.go||' | sort

# Generate documentation
.PHONY: docs
docs:
	@echo "Documentation available in README.md and docs/ directory"

# Generate mocks
.PHONY: mocks
mocks:
	go generate ./...

# Build Docker image
.PHONY: docker-build
docker-build:
	docker build -t gobot:latest .

# Run Docker container
.PHONY: docker-run
docker-run:
	docker run --rm -it gobot:latest

# Security scan
.PHONY: security
security:
	gosec ./...

# Check dependencies for vulnerabilities
.PHONY: vuln-check
vuln-check:
	go run golang.org/x/vuln/cmd/govulncheck ./...

# Install pre-commit hooks
.PHONY: install-hooks
install-hooks:
	@echo "Pre-commit hooks not configured yet"

# CI pipeline
.PHONY: ci
ci: deps lint test build-matrix

# Release preparation
.PHONY: release-prep
release-prep: ci docs
	@echo "Release preparation complete"

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build           - Build all packages"
	@echo "  build-platform  - Build for specific platform (PLATFORM=linux/amd64)"
	@echo "  build-matrix    - Build for all platforms"
	@echo "  test            - Run all tests"
	@echo "  test-package    - Run tests for specific package (PACKAGE=./pkg/core)"
	@echo "  test-tags       - Run tests with build tags (TAGS='gpio,i2c')"
	@echo "  bench           - Run benchmarks"
	@echo "  test-integration - Run integration tests"
	@echo "  test-e2e        - Run end-to-end tests"
	@echo "  lint            - Run linter"
	@echo "  fmt             - Format code"
	@echo "  tidy            - Tidy dependencies"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Install dependencies"
	@echo "  run-example     - Run example (EXAMPLE=basic/hello)"
	@echo "  list-examples   - List available examples"
	@echo "  docs            - Generate documentation"
	@echo "  mocks           - Generate mocks"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-run      - Run Docker container"
	@echo "  security        - Run security scan"
	@echo "  vuln-check      - Check for vulnerabilities"
	@echo "  install-hooks   - Install pre-commit hooks"
	@echo "  ci              - Run CI pipeline"
	@echo "  release-prep    - Prepare for release"
	@echo "  help            - Show this help"