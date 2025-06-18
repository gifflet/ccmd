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

echo -e "${BLUE}Local Release Builder for ccmd${NC}"
echo "================================="

# Change to project root
cd "$PROJECT_ROOT"

# Check for goreleaser
if ! command -v goreleaser &> /dev/null; then
    echo -e "${YELLOW}GoReleaser not found. Installing...${NC}"
    go install github.com/goreleaser/goreleaser/v2@latest
fi

# Parse command line arguments
SNAPSHOT=false
SKIP_VALIDATE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --snapshot)
            SNAPSHOT=true
            ;;
        --skip-validate)
            SKIP_VALIDATE=true
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  --snapshot       Create snapshot build (no git tag required)"
            echo "  --skip-validate  Skip git validation"
            echo "  --help          Show this help message"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
    shift
done

# Clean previous builds
echo -e "${YELLOW}Cleaning previous builds...${NC}"
rm -rf dist/

# Run goreleaser
if [ "$SNAPSHOT" = true ]; then
    echo -e "${YELLOW}Creating snapshot release...${NC}"
    goreleaser release --snapshot --clean
else
    if [ "$SKIP_VALIDATE" = true ]; then
        echo -e "${YELLOW}Creating release (skip validation)...${NC}"
        goreleaser release --skip-validate --clean
    else
        echo -e "${YELLOW}Creating release...${NC}"
        
        # Check if we have a tag
        if ! git describe --tags --exact-match HEAD &> /dev/null; then
            echo -e "${RED}Error: No tag found at HEAD${NC}"
            echo "Please create a tag first: git tag -a v0.1.0 -m 'Release v0.1.0'"
            exit 1
        fi
        
        goreleaser release --clean
    fi
fi

# Show results
echo -e "\n${GREEN}Release build complete!${NC}"
echo -e "Artifacts created in: ${BLUE}$PROJECT_ROOT/dist/${NC}"
echo -e "\nContents:"
ls -la dist/

# Show checksums if available
if [ -f "dist/checksums.txt" ]; then
    echo -e "\n${YELLOW}Checksums:${NC}"
    cat dist/checksums.txt
fi