#!/bin/bash
set -e

# Install script for ubik CLI

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY_PATH="$SCRIPT_DIR/bin/ubik-cli"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="ubik"

echo "🔨 ubik CLI Installer"
echo "====================="
echo

# Check if binary exists
if [ ! -f "$BINARY_PATH" ]; then
    echo "❌ Binary not found at $BINARY_PATH"
    echo "📦 Building CLI..."
    cd "$SCRIPT_DIR"
    go build -o bin/ubik-cli cmd/cli/main.go
    echo "✅ Built successfully"
fi

# Check if install dir is writable
if [ -w "$INSTALL_DIR" ]; then
    # No sudo needed
    echo "📦 Installing to $INSTALL_DIR/$BINARY_NAME"
    cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    echo "✅ Installed successfully"
else
    # Need sudo
    echo "📦 Installing to $INSTALL_DIR/$BINARY_NAME (requires sudo)"
    sudo cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
    echo "✅ Installed successfully"
fi

echo
echo "🎉 Installation complete!"
echo
echo "Try it out:"
echo "  $ ubik --version"
echo "  $ ubik --help"
echo "  $ ubik login"
echo
