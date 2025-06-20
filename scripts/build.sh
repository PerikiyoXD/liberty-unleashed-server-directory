#!/bin/bash

# Build script for Liberty Unleashed Server Directory
# This script builds the application for multiple platforms

set -e

APP_NAME="lusd"
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "Building $APP_NAME version $VERSION"
echo "Build time: $BUILD_TIME"
echo "Commit: $COMMIT_HASH"

# Create build directory
mkdir -p build

# Build flags
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X main.Version=$VERSION"
LDFLAGS="$LDFLAGS -X main.BuildTime=$BUILD_TIME"
LDFLAGS="$LDFLAGS -X main.CommitHash=$COMMIT_HASH"

# Build for different platforms
echo "Building for Linux amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o build/${APP_NAME}-linux-amd64 cmd/lusd/main.go

echo "Building for Linux arm64..."
GOOS=linux GOARCH=arm64 go build -ldflags="$LDFLAGS" -o build/${APP_NAME}-linux-arm64 cmd/lusd/main.go

echo "Building for Windows amd64..."
GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o build/${APP_NAME}-windows-amd64.exe cmd/lusd/main.go

echo "Building for macOS amd64..."
GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -o build/${APP_NAME}-darwin-amd64 cmd/lusd/main.go

echo "Building for macOS arm64..."
GOOS=darwin GOARCH=arm64 go build -ldflags="$LDFLAGS" -o build/${APP_NAME}-darwin-arm64 cmd/lusd/main.go

# Make Linux/macOS binaries executable
chmod +x build/${APP_NAME}-linux-* build/${APP_NAME}-darwin-*

echo "Build completed successfully!"
echo "Binaries available in build/ directory:"
ls -la build/
