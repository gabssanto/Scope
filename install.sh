#!/bin/sh
set -e

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    armv7l) ARCH="arm" ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Check OS
case $OS in
    linux|darwin) ;;
    *)
        echo "Unsupported operating system: $OS"
        echo "Supported: Linux, macOS (darwin)"
        exit 1
        ;;
esac

BINARY="scope-${OS}-${ARCH}"
REPO="gabssanto/Scope"
URL="https://github.com/${REPO}/releases/latest/download/${BINARY}"

echo "Installing Scope for ${OS}/${ARCH}..."
echo "Downloading from: $URL"

# Create temp directory
TMP_DIR=$(mktemp -d)
trap "rm -rf ${TMP_DIR}" EXIT

cd "${TMP_DIR}"

# Download binary
if command -v curl > /dev/null 2>&1; then
    curl -fSL -o scope "$URL"
elif command -v wget > /dev/null 2>&1; then
    wget -O scope "$URL"
else
    echo "Error: curl or wget is required"
    exit 1
fi

# Make executable
chmod +x scope

# Install to /usr/local/bin
if [ -w /usr/local/bin ]; then
    mv scope /usr/local/bin/scope
    echo "Scope installed to /usr/local/bin/scope"
else
    echo "Installing to /usr/local/bin requires sudo..."
    sudo mv scope /usr/local/bin/scope
    echo "Scope installed to /usr/local/bin/scope"
fi

# Verify installation
echo ""
echo "Installation successful!"
echo ""

# Check for PATH conflicts
INSTALLED_PATH=$(which scope 2>/dev/null || true)
if [ -n "$INSTALLED_PATH" ] && [ "$INSTALLED_PATH" != "/usr/local/bin/scope" ]; then
    echo "⚠️  Warning: Another 'scope' binary found at: $INSTALLED_PATH"
    echo "   This may take precedence over the newly installed version."
    echo ""
    echo "   To fix this, either:"
    echo "   1. Remove the old binary: rm $INSTALLED_PATH"
    echo "   2. Or run directly: /usr/local/bin/scope"
    echo ""
fi

/usr/local/bin/scope version
echo ""
echo "Run 'scope help' for usage information."
