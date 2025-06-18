#!/usr/bin/env bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

echo -e "${BLUE}Testing ccmd build system${NC}"
echo "========================="

# Change to project root
cd "$PROJECT_ROOT"

# Function to test a build
test_build() {
    local os=$1
    local arch=$2
    local expected_binary=$3
    
    echo -e "\n${YELLOW}Testing $os/$arch build...${NC}"
    
    # Build
    GOOS=$os GOARCH=$arch go build -o "dist/test-$expected_binary" ./cmd/ccmd
    
    if [ -f "dist/test-$expected_binary" ]; then
        echo -e "${GREEN}✓ Build successful${NC}"
        file_info=$(file "dist/test-$expected_binary")
        echo "  File info: $file_info"
        rm "dist/test-$expected_binary"
        return 0
    else
        echo -e "${RED}✗ Build failed${NC}"
        return 1
    fi
}

# Create dist directory
mkdir -p dist

# Test builds for different platforms
echo -e "${YELLOW}Testing cross-platform builds...${NC}"

failed=0

# Test Darwin builds
test_build "darwin" "amd64" "ccmd-darwin-amd64" || ((failed++))
test_build "darwin" "arm64" "ccmd-darwin-arm64" || ((failed++))

# Test Linux builds
test_build "linux" "amd64" "ccmd-linux-amd64" || ((failed++))
test_build "linux" "arm64" "ccmd-linux-arm64" || ((failed++))

# Test Windows builds
test_build "windows" "amd64" "ccmd-windows-amd64.exe" || ((failed++))
test_build "windows" "arm64" "ccmd-windows-arm64.exe" || ((failed++))

# Test Makefile targets
echo -e "\n${YELLOW}Testing Makefile targets...${NC}"

# Test make build
echo -e "\n${BLUE}Testing 'make build'...${NC}"
make clean
make build
if [ -f "dist/ccmd" ]; then
    echo -e "${GREEN}✓ make build successful${NC}"
    ./dist/ccmd --version
else
    echo -e "${RED}✗ make build failed${NC}"
    ((failed++))
fi

# Test make build-all
echo -e "\n${BLUE}Testing 'make build-all'...${NC}"
make clean
make build-all
expected_files=(
    "ccmd-darwin-amd64"
    "ccmd-darwin-arm64"
    "ccmd-linux-amd64"
    "ccmd-linux-arm64"
    "ccmd-windows-amd64.exe"
    "ccmd-windows-arm64.exe"
)

for file in "${expected_files[@]}"; do
    if [ -f "dist/$file" ]; then
        echo -e "${GREEN}✓ Found $file${NC}"
    else
        echo -e "${RED}✗ Missing $file${NC}"
        ((failed++))
    fi
done

# Test compression
echo -e "\n${BLUE}Testing 'make compress-artifacts'...${NC}"
make compress-artifacts
compressed_files=(
    "ccmd-darwin-amd64.tar.gz"
    "ccmd-darwin-arm64.tar.gz"
    "ccmd-linux-amd64.tar.gz"
    "ccmd-linux-arm64.tar.gz"
    "ccmd-windows-amd64.zip"
    "ccmd-windows-arm64.zip"
)

for file in "${compressed_files[@]}"; do
    if [ -f "dist/$file" ]; then
        echo -e "${GREEN}✓ Found compressed $file${NC}"
    else
        echo -e "${RED}✗ Missing compressed $file${NC}"
        ((failed++))
    fi
done

# Summary
echo -e "\n${BLUE}Test Summary${NC}"
echo "============"
if [ $failed -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
else
    echo -e "${RED}$failed tests failed${NC}"
    exit 1
fi

# Clean up
make clean