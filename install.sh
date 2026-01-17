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
ARCHIVE="${BINARY}.tar.gz"
REPO="gabssanto/Scope"
URL="https://github.com/${REPO}/releases/latest/download/${ARCHIVE}"

echo "Installing Scope for ${OS}/${ARCH}..."
echo "Downloading from: $URL"

# Create temp directory
TMP_DIR=$(mktemp -d)
trap "rm -rf ${TMP_DIR}" EXIT

cd "${TMP_DIR}"

# Download archive
if command -v curl > /dev/null 2>&1; then
    curl -L -o "${ARCHIVE}" "$URL"
elif command -v wget > /dev/null 2>&1; then
    wget -O "${ARCHIVE}" "$URL"
else
    echo "Error: curl or wget is required"
    exit 1
fi

# Extract binary
tar xzf "${ARCHIVE}"

# Make executable
chmod +x "${BINARY}"

# Install to /usr/local/bin
if [ -w /usr/local/bin ]; then
    mv "${BINARY}" /usr/local/bin/scope
    echo "Scope installed to /usr/local/bin/scope"
else
    echo "Installing to /usr/local/bin requires sudo..."
    sudo mv "${BINARY}" /usr/local/bin/scope
    echo "Scope installed to /usr/local/bin/scope"
fi

# Verify installation
if command -v scope > /dev/null 2>&1; then
    echo ""
    echo "Installation successful!"
    echo ""
    scope help
else
    echo "Installation may have failed. Please check /usr/local/bin/scope"
    exit 1
fi
