#!/usr/bin/env bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

echo -e "${GREEN}Building ccmd...${NC}"

# Change to project root
cd "$PROJECT_ROOT"

# Check for required tools
echo "Checking required tools..."
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

# Get version information
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

echo "Version: $VERSION"
echo "Commit: $COMMIT"
echo "Build Date: $BUILD_DATE"

# Build options
LDFLAGS="-X main.version=$VERSION -X main.commit=$COMMIT -X main.buildDate=$BUILD_DATE"
BUILD_FLAGS=""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --all)
            echo -e "${YELLOW}Building for all platforms...${NC}"
            make build-all
            exit 0
            ;;
        --release)
            echo -e "${YELLOW}Creating release build...${NC}"
            make release
            exit 0
            ;;
        --dev)
            echo -e "${YELLOW}Building with race detector...${NC}"
            BUILD_FLAGS="-race"
            ;;
        --static)
            echo -e "${YELLOW}Building static binary...${NC}"
            export CGO_ENABLED=0
            LDFLAGS="$LDFLAGS -s -w"
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  --all      Build for all platforms"
            echo "  --release  Create release build with compression"
            echo "  --dev      Build with race detector"
            echo "  --static   Build static binary"
            echo "  --help     Show this help message"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
    shift
done

# Create build directory
mkdir -p dist

# Build binary
echo "Building binary..."
go build $BUILD_FLAGS -ldflags "$LDFLAGS" -o dist/ccmd ./cmd/ccmd

if [ $? -eq 0 ]; then
    echo -e "${GREEN}Build successful!${NC}"
    echo "Binary location: $PROJECT_ROOT/dist/ccmd"
    
    # Show binary info
    echo -e "\nBinary info:"
    file dist/ccmd
    ls -lh dist/ccmd
    
    # Test version output
    echo -e "\nVersion output:"
    ./dist/ccmd --version
else
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi