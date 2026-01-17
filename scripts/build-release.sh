#!/bin/bash
# Build script for creating release binaries

set -e

VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo '0.1.0')}"
BUILD_DIR="${BUILD_DIR:-./build}"
RELEASE_DIR="${BUILD_DIR}/release"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}Building Scope Release ${VERSION}${NC}"
echo ""

# Clean previous builds
echo -e "${YELLOW}Cleaning previous builds...${NC}"
rm -rf "${BUILD_DIR}"
mkdir -p "${RELEASE_DIR}"

# Build matrix
platforms=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for platform in "${platforms[@]}"; do
    IFS='/' read -r -a parts <<< "$platform"
    GOOS="${parts[0]}"
    GOARCH="${parts[1]}"

    output_name="scope-${GOOS}-${GOARCH}"
    if [ "$GOOS" == "windows" ]; then
        output_name="${output_name}.exe"
    fi

    echo -e "${GREEN}Building ${GOOS}/${GOARCH}...${NC}"

    GOOS="${GOOS}" GOARCH="${GOARCH}" go build \
        -ldflags="-s -w -X main.Version=${VERSION}" \
        -o "${BUILD_DIR}/${output_name}" \
        ./cmd/scope/main.go

    # Create archives
    cd "${BUILD_DIR}"
    if [ "$GOOS" == "windows" ]; then
        zip "${output_name}.zip" "${output_name}"
        mv "${output_name}.zip" release/
    else
        tar czf "${output_name}.tar.gz" "${output_name}"
        mv "${output_name}.tar.gz" release/
    fi

    # Move binary to release dir
    mv "${output_name}" release/

    cd - > /dev/null
done

# Generate checksums
echo ""
echo -e "${YELLOW}Generating checksums...${NC}"
cd "${RELEASE_DIR}"
sha256sum * > SHA256SUMS
cd - > /dev/null

echo ""
echo -e "${GREEN}Build complete!${NC}"
echo -e "${BLUE}Release artifacts in: ${RELEASE_DIR}${NC}"
echo ""
ls -lh "${RELEASE_DIR}"
