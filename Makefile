# ScreenCaptureKit Go Makefile

.PHONY: build test clean examples install deps check-deps format lint

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOCLEAN=$(GOCMD) clean
GOFORMAT=gofmt
GOLINT=golangci-lint

# Build output directory
BUILD_DIR=build

# Check if we're on macOS
UNAME_S := $(shell uname -s)
ifneq ($(UNAME_S),Darwin)
$(error This project requires macOS to build and run)
endif

# Check macOS version (requires 12.3+)
MACOS_VERSION := $(shell sw_vers -productVersion | cut -d '.' -f 1,2)
REQUIRED_VERSION := 12.3

check-macos-version:
	@echo "Checking macOS version..."
	@echo "Current version: $(MACOS_VERSION)"
	@echo "Required version: $(REQUIRED_VERSION) or later"

# Default target
all: check-deps check-macos-version deps build-swift build test

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) tidy
	$(GOMOD) download

# Check if required tools are installed
check-deps:
	@echo "Checking dependencies..."
	@which xcode-select > /dev/null || (echo "Error: Xcode Command Line Tools not installed. Run: xcode-select --install" && exit 1)
	@$(GOCMD) version > /dev/null || (echo "Error: Go not installed" && exit 1)
	@which swift > /dev/null || (echo "Error: Swift not found. Install with: xcode-select --install" && exit 1)
	@echo "✅ Go version: $$(go version | cut -d' ' -f3-4)"
	@echo "✅ Swift version: $$(swift --version | head -1)"
	@echo "✅ CGO_ENABLED: $$(go env CGO_ENABLED)"

# Build Swift CLI binary
build-swift: check-deps
	@echo "Building ScreenCaptureKit Swift CLI..."
	@if [ ! -f ".build/release/screencapturekit" ] && [ ! -f ".build/apple/Products/Release/screencapturekit" ]; then \
		echo "🔨 Building Swift binary (this may take a moment)..."; \
		swift build --configuration=release; \
		echo "✅ Swift binary built successfully"; \
	else \
		echo "✅ Swift binary already exists"; \
	fi

# Build the Go package
build: check-deps build-swift
	@echo "Building ScreenCaptureKit Go package..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -v -x -o $(BUILD_DIR)/screencapturekit .

# Test the package
test: check-deps
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Build examples
examples: check-deps
	@echo "Building examples..."
	@mkdir -p $(BUILD_DIR)/examples
	@cd cmd/examples && \
	$(GOBUILD) -o ../../$(BUILD_DIR)/examples/basic_recording basic_recording.go && \
	$(GOBUILD) -o ../../$(BUILD_DIR)/examples/audio_recording audio_recording.go && \
	$(GOBUILD) -o ../../$(BUILD_DIR)/examples/hdr_recording hdr_recording.go && \
	$(GOBUILD) -o ../../$(BUILD_DIR)/examples/cropped_recording cropped_recording.go && \
	$(GOBUILD) -o ../../$(BUILD_DIR)/examples/audio_only audio_only.go && \
	$(GOBUILD) -o ../../$(BUILD_DIR)/examples/http_streaming http_streaming.go && \
	$(GOBUILD) -o ../../$(BUILD_DIR)/examples/websocket_streaming websocket_streaming.go && \
	$(GOBUILD) -o ../../$(BUILD_DIR)/examples/tcp_streaming tcp_streaming.go && \
	$(GOBUILD) -o ../../$(BUILD_DIR)/examples/namedpipe_streaming namedpipe_streaming.go && \
	$(GOBUILD) -o ../../$(BUILD_DIR)/examples/namedpipe_ffmpeg_streaming namedpipe_ffmpeg_streaming.go
	@echo "Examples built in $(BUILD_DIR)/examples/"

# Run a specific example
run-basic: examples
	@echo "Running basic recording example..."
	$(BUILD_DIR)/examples/basic_recording

run-audio: examples
	@echo "Running audio recording example..."
	$(BUILD_DIR)/examples/audio_recording

run-hdr: examples
	@echo "Running HDR recording example..."
	$(BUILD_DIR)/examples/hdr_recording

run-cropped: examples
	@echo "Running cropped recording example..."
	$(BUILD_DIR)/examples/cropped_recording

run-audio-only: examples
	@echo "Running audio-only recording example..."
	$(BUILD_DIR)/examples/audio_only

# Streaming examples
run-http-stream: examples
	@echo "Running HTTP streaming example..."
	$(BUILD_DIR)/examples/http_streaming

run-websocket-stream: examples
	@echo "Running WebSocket streaming example..."
	$(BUILD_DIR)/examples/websocket_streaming

run-tcp-stream: examples
	@echo "Running TCP streaming example..."
	$(BUILD_DIR)/examples/tcp_streaming

run-namedpipe-stream: examples
	@echo "Running Named Pipe streaming example..."
	$(BUILD_DIR)/examples/namedpipe_streaming

run-namedpipe-ffmpeg-stream: examples
	@echo "Running FFmpeg-compatible Named Pipe streaming example..."
	$(BUILD_DIR)/examples/namedpipe_ffmpeg_streaming

# Format code
format:
	@echo "Formatting code..."
	$(GOFORMAT) -s -w .

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@which $(GOLINT) > /dev/null || (echo "Error: golangci-lint not installed. Run: brew install golangci-lint" && exit 1)
	$(GOLINT) run

# Install the package locally
install: check-deps
	@echo "Installing ScreenCaptureKit Go package..."
	$(GOCMD) install .

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Clean Swift build artifacts
clean-swift:
	@echo "Cleaning Swift build artifacts..."
	rm -rf .build

# Full clean (Go + Swift)
clean-all: clean clean-swift
	@echo "All build artifacts cleaned"

# Verify Swift binary installation
verify-binary:
	@echo "Verifying ScreenCaptureKit binary..."
	@if [ -f ".build/release/screencapturekit" ]; then \
		echo "✅ Binary found: .build/release/screencapturekit"; \
		./.build/release/screencapturekit list screens > /dev/null 2>&1 && echo "✅ Binary functional" || echo "⚠️  Binary needs screen recording permission"; \
	elif [ -f ".build/apple/Products/Release/screencapturekit" ]; then \
		echo "✅ Binary found: .build/apple/Products/Release/screencapturekit"; \
		./.build/apple/Products/Release/screencapturekit list screens > /dev/null 2>&1 && echo "✅ Binary functional" || echo "⚠️  Binary needs screen recording permission"; \
	elif [ -f "/usr/local/bin/screencapturekit" ]; then \
		echo "✅ Binary found: /usr/local/bin/screencapturekit"; \
		screencapturekit list screens > /dev/null 2>&1 && echo "✅ Binary functional" || echo "⚠️  Binary needs screen recording permission"; \
	else \
		echo "❌ Binary not found. Run 'make build-swift' to build it"; \
	fi

# Install Swift binary globally
install-binary: build-swift
	@echo "Installing ScreenCaptureKit binary globally..."
	@if [ -f ".build/release/screencapturekit" ]; then \
		sudo cp .build/release/screencapturekit /usr/local/bin/; \
		echo "✅ Binary installed to /usr/local/bin/screencapturekit"; \
	elif [ -f ".build/apple/Products/Release/screencapturekit" ]; then \
		sudo cp .build/apple/Products/Release/screencapturekit /usr/local/bin/; \
		echo "✅ Binary installed to /usr/local/bin/screencapturekit"; \
	else \
		echo "❌ Binary not found. Build failed?"; \
		exit 1; \
	fi

# Rebuild Swift binary from scratch
rebuild-swift: clean-swift build-swift
	@echo "Swift binary rebuilt successfully"

# Check screen recording permissions
check-permissions:
	@echo "Checking screen recording permissions..."
	@echo "If the examples fail, you may need to grant Screen Recording permission:"
	@echo "1. Open System Preferences > Security & Privacy > Privacy > Screen Recording"
	@echo "2. Add Terminal (or your IDE) to the list"
	@echo "3. Restart your terminal/IDE"

# Development setup
dev-setup: check-deps deps
	@echo "Setting up development environment..."
	@which $(GOLINT) > /dev/null || echo "Consider installing golangci-lint: brew install golangci-lint"
	@echo "Development setup complete!"

# Show help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Building:"
	@echo "  all              - Build everything (Swift + Go + examples + test)"
	@echo "  build-swift      - Build Swift CLI binary"
	@echo "  build            - Build Go package (includes Swift)"
	@echo "  examples         - Build all examples"
	@echo "  rebuild-swift    - Clean and rebuild Swift binary"
	@echo ""
	@echo "Binary Management:"
	@echo "  verify-binary    - Verify Swift binary installation"
	@echo "  install-binary   - Install Swift binary globally (/usr/local/bin)"
	@echo "  clean-swift      - Clean Swift build artifacts"
	@echo "  clean-all        - Clean all build artifacts (Go + Swift)"
	@echo ""
	@echo "Examples:"
	@echo "  run-basic        - Run basic recording example"
	@echo "  run-audio        - Run audio recording example"
	@echo "  run-hdr          - Run HDR recording example"
	@echo "  run-cropped      - Run cropped recording example"
	@echo "  run-audio-only   - Run audio-only recording example"
	@echo "  run-http-stream  - Run HTTP streaming example"
	@echo "  run-websocket-stream - Run WebSocket streaming example"
	@echo "  run-tcp-stream   - Run TCP streaming example"
	@echo "  run-namedpipe-stream - Run Named Pipe streaming example"
	@echo "  run-namedpipe-ffmpeg-stream - Run FFmpeg-compatible streaming example"
	@echo ""
	@echo "Development:"
	@echo "  test             - Run tests"
	@echo "  format           - Format code"
	@echo "  lint             - Lint code (requires golangci-lint)"
	@echo "  clean            - Clean Go build artifacts"
	@echo "  dev-setup        - Set up development environment"
	@echo ""
	@echo "System:"
	@echo "  check-deps       - Check system dependencies"
	@echo "  check-permissions - Check screen recording permissions"
	@echo "  install          - Install Go package locally"
	@echo "  deps             - Install/update Go dependencies"
	@echo "  help             - Show this help message"
	@echo ""
	@echo "Requirements:"
	@echo "  - macOS $(REQUIRED_VERSION) or later"
	@echo "  - Go 1.21 or later"
	@echo "  - Xcode Command Line Tools (for Swift)"
	@echo "  - Screen Recording permissions"