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
all: check-deps check-macos-version deps build test

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
	@echo "CGO_ENABLED: $$(go env CGO_ENABLED)"

# Build the package
build: check-deps
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
	@echo "  all              - Build, test, and check everything"
	@echo "  build            - Build the main package"
	@echo "  test             - Run tests"
	@echo "  examples         - Build all examples"
	@echo "  run-basic        - Run basic recording example"
	@echo "  run-audio        - Run audio recording example"
	@echo "  run-hdr          - Run HDR recording example"
	@echo "  run-cropped      - Run cropped recording example"
	@echo "  run-audio-only   - Run audio-only recording example"
	@echo "  run-http-stream  - Run HTTP streaming example"
	@echo "  run-websocket-stream - Run WebSocket streaming example"
	@echo "  run-tcp-stream   - Run TCP streaming example"
	@echo "  run-namedpipe-stream - Run Named Pipe streaming example"
	@echo "  run-namedpipe-ffmpeg-stream - Run FFmpeg-compatible Named Pipe streaming example"
	@echo "  format           - Format code"
	@echo "  lint             - Lint code (requires golangci-lint)"
	@echo "  install          - Install package locally"
	@echo "  clean            - Clean build artifacts"
	@echo "  check-permissions - Check screen recording permissions"
	@echo "  dev-setup        - Set up development environment"
	@echo "  deps             - Install/update dependencies"
	@echo "  help             - Show this help message"
	@echo ""
	@echo "Requirements:"
	@echo "  - macOS $(REQUIRED_VERSION) or later"
	@echo "  - Go 1.21 or later"
	@echo "  - Xcode Command Line Tools"
	@echo "  - Screen Recording permissions"