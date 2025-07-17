# Project name
PROJECT_NAME := ccmd
BINARY_NAME := ccmd

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Build parameters
MAIN_PATH := ./cmd/ccmd
BUILD_DIR := ./dist
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)"

# Target OS and architectures
PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 windows/amd64

.PHONY: all build clean test deps fmt lint vet build-all release help

# Default target
all: clean build

# Build for current platform
build:
	@echo "Building $(BINARY_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
ifeq ($(OS),Windows_NT)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).exe $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME).exe"
else
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"
endif

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf npm_publish/bin
	@rm -rf npm_publish/dist
	@rm -rf npm_publish/node_modules
	$(GOCLEAN)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Please install it from https://golangci-lint.run/"; \
	fi

# Build for all platforms
build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		output_name=$(BINARY_NAME)-$$GOOS-$$GOARCH; \
		if [ $$GOOS = "windows" ]; then \
			output_name=$$output_name.exe; \
		fi; \
		echo "Building for $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$$output_name $(MAIN_PATH); \
	done
	@echo "All builds complete!"

# Compress release artifacts
compress-artifacts:
	@echo "Compressing release artifacts..."
	@cd $(BUILD_DIR) && \
	for file in $(BINARY_NAME)-*; do \
		if [ -f "$$file" ]; then \
			if [[ "$$file" == *.exe ]]; then \
				zip "$${file%.exe}.zip" "$$file"; \
				rm "$$file"; \
			else \
				tar -czf "$$file.tar.gz" "$$file"; \
				rm "$$file"; \
			fi; \
		fi; \
	done
	@echo "Compression complete!"

# Create release
release: build-all compress-artifacts
	@echo "Release artifacts created in $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

# Prepare npm package for publishing
npm-prepare-publish: build-all
	@echo "Preparing npm package..."
	@mkdir -p ./npm_publish/dist/ccmd-darwin-amd64_darwin_amd64
	@mkdir -p ./npm_publish/dist/ccmd-darwin-arm64_darwin_arm64
	@mkdir -p ./npm_publish/dist/ccmd-linux-amd64_linux_amd64
	@mkdir -p ./npm_publish/dist/ccmd-windows-amd64_windows_amd64
	@cp $(BUILD_DIR)/ccmd-darwin-amd64 ./npm_publish/dist/ccmd-darwin-amd64_darwin_amd64/ccmd
	@cp $(BUILD_DIR)/ccmd-darwin-arm64 ./npm_publish/dist/ccmd-darwin-arm64_darwin_arm64/ccmd
	@cp $(BUILD_DIR)/ccmd-windows-amd64.exe ./npm_publish/dist/ccmd-windows-amd64_windows_amd64/ccmd.exe
	@cp $(BUILD_DIR)/ccmd-linux-amd64 ./npm_publish/dist/ccmd-linux-amd64_linux_amd64/ccmd
	@echo "Package ready for publishing"

# Install locally
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME) || sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "Installation complete!"

# Uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(GOPATH)/bin/$(BINARY_NAME) || sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstall complete!"

# Run the binary
run: build
ifeq ($(OS),Windows_NT)
	$(BUILD_DIR)/$(BINARY_NAME).exe
else
	$(BUILD_DIR)/$(BINARY_NAME)
endif

# Development build with race detector
dev-build:
	@echo "Building with race detector..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -race $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-dev $(MAIN_PATH)
	@echo "Development build complete: $(BUILD_DIR)/$(BINARY_NAME)-dev"

# Check for required tools
check-tools:
	@echo "Checking for required tools..."
	@command -v go >/dev/null || (echo "Go is not installed" && exit 1)
	@command -v git >/dev/null || (echo "Git is not installed" && exit 1)
	@echo "All required tools are installed!"

# Show version
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"

# Help
help:
	@echo "Available targets:"
	@echo "  make build         - Build for current platform"
	@echo "  make build-all     - Build for all platforms"
	@echo "  make release       - Build and compress all platforms"
	@echo "  make npm-prepare-publish - Prepare npm package for publishing"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make test          - Run tests"
	@echo "  make deps          - Download dependencies"
	@echo "  make fmt           - Format code"
	@echo "  make lint          - Run linter"
	@echo "  make vet           - Run go vet"
	@echo "  make install       - Install binary locally"
	@echo "  make uninstall     - Uninstall binary"
	@echo "  make run           - Build and run"
	@echo "  make dev-build     - Build with race detector"
	@echo "  make check-tools   - Check for required tools"
	@echo "  make version       - Show version info"
	@echo "  make help          - Show this help"