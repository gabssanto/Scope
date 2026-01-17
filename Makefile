# Scope - Makefile for building, testing, and releasing

# Variables
BINARY_NAME=scope
VERSION?=0.1.0
BUILD_DIR=./build
MAIN_PATH=./cmd/scope/main.go
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -s -w"

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
NC=\033[0m # No Color

.PHONY: all build clean test test-coverage test-verbose install uninstall run help
.PHONY: build-all release lint fmt vet deps dev-setup ci
.PHONY: test-tag test-untag test-list test-session test-integration

# Default target
all: clean deps test build

## help: Display this help message
help:
	@echo "$(GREEN)Scope Build System$(NC)"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@grep -E '^## ' Makefile | sed 's/## /  /' | column -t -s ':' | sort

## build: Build the binary for current platform
build:
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## build-all: Cross-compile for all platforms
build-all: clean
	@echo "$(GREEN)Cross-compiling for all platforms...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "$(GREEN)✓ All builds complete$(NC)"
	@ls -lh $(BUILD_DIR)

## clean: Remove build artifacts and test data
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -rf /tmp/scope-test-*
	@echo "$(GREEN)✓ Clean complete$(NC)"

## deps: Download and tidy dependencies
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	$(GO) mod download
	$(GO) mod tidy
	@echo "$(GREEN)✓ Dependencies ready$(NC)"

## test: Run all tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	$(GO) test -v -race ./...
	@echo "$(GREEN)✓ Tests passed$(NC)"

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	$(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	$(GO) tool cover -func=coverage.out
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(NC)"

## test-verbose: Run tests with verbose output
test-verbose:
	@echo "$(GREEN)Running tests (verbose)...$(NC)"
	$(GO) test -v -race -count=1 ./...

## test-integration: Run integration tests
test-integration: build
	@echo "$(GREEN)Running integration tests...$(NC)"
	@bash ./scripts/integration-test.sh
	@echo "$(GREEN)✓ Integration tests passed$(NC)"

## test-tag: Quick test of tag functionality
test-tag: build
	@echo "$(GREEN)Testing tag functionality...$(NC)"
	@rm -rf /tmp/scope-test-tag
	@mkdir -p /tmp/scope-test-tag
	$(BUILD_DIR)/$(BINARY_NAME) tag /tmp/scope-test-tag test-tag
	$(BUILD_DIR)/$(BINARY_NAME) list test-tag
	@echo "$(GREEN)✓ Tag test passed$(NC)"

## test-untag: Quick test of untag functionality
test-untag: build
	@echo "$(GREEN)Testing untag functionality...$(NC)"
	@rm -rf /tmp/scope-test-untag
	@mkdir -p /tmp/scope-test-untag
	$(BUILD_DIR)/$(BINARY_NAME) tag /tmp/scope-test-untag test-untag
	$(BUILD_DIR)/$(BINARY_NAME) untag /tmp/scope-test-untag test-untag
	@echo "$(GREEN)✓ Untag test passed$(NC)"

## test-list: Quick test of list functionality
test-list: build
	@echo "$(GREEN)Testing list functionality...$(NC)"
	$(BUILD_DIR)/$(BINARY_NAME) list
	@echo "$(GREEN)✓ List test passed$(NC)"

## fmt: Format all Go code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GO) fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

## vet: Run go vet
vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	$(GO) vet ./...
	@echo "$(GREEN)✓ Vet passed$(NC)"

## lint: Run linters (requires golangci-lint)
lint:
	@echo "$(GREEN)Running linters...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo "$(GREEN)✓ Lint passed$(NC)"; \
	else \
		echo "$(YELLOW)golangci-lint not installed. Install: https://golangci-lint.run/usage/install/$(NC)"; \
	fi

## install: Install binary to /usr/local/bin
install: build
	@echo "$(GREEN)Installing $(BINARY_NAME)...$(NC)"
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)✓ Installed to /usr/local/bin/$(BINARY_NAME)$(NC)"

## uninstall: Remove binary from /usr/local/bin
uninstall:
	@echo "$(YELLOW)Uninstalling $(BINARY_NAME)...$(NC)"
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)✓ Uninstalled$(NC)"

## run: Build and run the binary
run: build
	$(BUILD_DIR)/$(BINARY_NAME)

## dev-setup: Setup development environment
dev-setup:
	@echo "$(GREEN)Setting up development environment...$(NC)"
	$(GO) install golang.org/x/tools/cmd/goimports@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(GREEN)✓ Development environment ready$(NC)"

## release: Create a release (tags, builds, and packages)
release: clean test build-all
	@echo "$(GREEN)Creating release $(VERSION)...$(NC)"
	@mkdir -p $(BUILD_DIR)/release
	@cd $(BUILD_DIR) && \
		tar czf release/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
		tar czf release/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && \
		tar czf release/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
		tar czf release/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && \
		zip release/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	@cd $(BUILD_DIR)/release && sha256sum * > SHA256SUMS
	@echo "$(GREEN)✓ Release $(VERSION) ready in $(BUILD_DIR)/release$(NC)"
	@ls -lh $(BUILD_DIR)/release

## ci: Run CI checks (fmt, vet, lint, test, build)
ci: fmt vet lint test build
	@echo "$(GREEN)✓ All CI checks passed$(NC)"

## benchmark: Run benchmarks
benchmark:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	$(GO) test -bench=. -benchmem ./...

## watch: Watch for changes and rebuild (requires entr)
watch:
	@if command -v entr >/dev/null 2>&1; then \
		find . -name '*.go' | entr -c make build; \
	else \
		echo "$(RED)entr not installed. Install: apt-get install entr or brew install entr$(NC)"; \
	fi

## docker-build: Build Docker image
docker-build:
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t $(BINARY_NAME):$(VERSION) .
	@echo "$(GREEN)✓ Docker image built$(NC)"

## size: Show binary sizes
size: build-all
	@echo "$(GREEN)Binary sizes:$(NC)"
	@ls -lh $(BUILD_DIR) | grep -E '$(BINARY_NAME)-|$(BINARY_NAME).exe'

# Development quick commands
.PHONY: quick qtest qbuild

## quick: Quick build and test
quick: build test
	@echo "$(GREEN)✓ Quick check complete$(NC)"

## qtest: Quick test only
qtest:
	@$(GO) test ./... -short

## qbuild: Quick build without cleaning
qbuild:
	@$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
