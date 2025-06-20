# Makefile for Liberty Unleashed Server Directory

APP_NAME := lusd
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS := -s -w
LDFLAGS += -X main.Version=$(VERSION)
LDFLAGS += -X main.BuildTime=$(BUILD_TIME)
LDFLAGS += -X main.CommitHash=$(COMMIT_HASH)

# Go build flags
GOFLAGS := -ldflags="$(LDFLAGS)"

.PHONY: all build clean test lint docker docker-build docker-run install deploy

# Default target
all: build

# Build for current platform
build:
	@echo "Building $(APP_NAME) version $(VERSION)"
	@go build $(GOFLAGS) -o $(APP_NAME) cmd/lusd/main.go

# Build for all platforms
build-all: clean
	@echo "Building $(APP_NAME) version $(VERSION) for all platforms"
	@mkdir -p build	@echo "Building for Linux amd64..."
	@GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o build/$(APP_NAME)-linux-amd64 cmd/lusd/main.go
	@echo "Building for Linux arm64..."
	@GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -o build/$(APP_NAME)-linux-arm64 cmd/lusd/main.go
	@echo "Building for Windows amd64..."  
	@GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -o build/$(APP_NAME)-windows-amd64.exe cmd/lusd/main.go
	@echo "Building for macOS amd64..."
	@GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) -o build/$(APP_NAME)-darwin-amd64 cmd/lusd/main.go
	@echo "Building for macOS arm64..."
	@GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) -o build/$(APP_NAME)-darwin-arm64 cmd/lusd/main.go
	@chmod +x build/$(APP_NAME)-linux-* build/$(APP_NAME)-darwin-*
	@echo "Build completed successfully!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf build/
	@rm -f $(APP_NAME) $(APP_NAME).exe
	@echo "Clean completed!"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(APP_NAME):$(VERSION) .
	@docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

# Run Docker container
docker-run: docker-build
	@echo "Running Docker container..."
	@docker run -p 80:80 --rm $(APP_NAME):$(VERSION)

# Install locally (Linux/macOS)
install: build
	@echo "Installing $(APP_NAME)..."
	@sudo cp $(APP_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(APP_NAME)
	@echo "Installation completed!"

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build for current platform"
	@echo "  build-all  - Build for all platforms"
	@echo "  clean      - Clean build artifacts"
	@echo "  test       - Run tests"
	@echo "  lint       - Run linter"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run - Run Docker container"
	@echo "  install    - Install locally"
	@echo "  help       - Show this help"
