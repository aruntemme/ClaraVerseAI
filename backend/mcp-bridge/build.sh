#!/bin/bash

# Cross-platform build script for MCP Client

echo "Building ClaraVerse MCP Client..."
echo ""

# Create bin directory
mkdir -p bin

# Windows
echo "[1/3] Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o bin/mcp-client-windows-amd64.exe ./cmd/mcp-client
if [ $? -ne 0 ]; then
    echo "❌ Build failed for Windows"
    exit 1
fi
echo "✓ Windows build complete"

# Linux
echo "[2/3] Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o bin/mcp-client-linux-amd64 ./cmd/mcp-client
if [ $? -ne 0 ]; then
    echo "❌ Build failed for Linux"
    exit 1
fi
echo "✓ Linux build complete"

# macOS Intel
echo "[3/3] Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -o bin/mcp-client-darwin-amd64 ./cmd/mcp-client
if [ $? -ne 0 ]; then
    echo "❌ Build failed for macOS"
    exit 1
fi
echo "✓ macOS build complete"

# macOS Apple Silicon (bonus)
echo "[4/4] Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -o bin/mcp-client-darwin-arm64 ./cmd/mcp-client
if [ $? -ne 0 ]; then
    echo "⚠️  Build failed for macOS ARM (non-critical)"
else
    echo "✓ macOS ARM build complete"
fi

echo ""
echo "========================================"
echo "✓ All builds completed successfully!"
echo "========================================"
echo ""
echo "Binaries created in bin/:"
ls -lh bin/mcp-client-*
echo ""
echo "To run:"
echo "  Windows: bin/mcp-client-windows-amd64.exe"
echo "  Linux:   bin/mcp-client-linux-amd64"
echo "  macOS:   bin/mcp-client-darwin-amd64"
echo "  macOS (M1/M2): bin/mcp-client-darwin-arm64"
echo ""
