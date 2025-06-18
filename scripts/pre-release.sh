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

echo -e "${BLUE}Pre-release checks for ccmd${NC}"
echo "============================"

# Change to project root
cd "$PROJECT_ROOT"

# Function to check if command exists
command_exists() {
    command -v "$1" &> /dev/null
}

# Check required tools
echo -e "\n${YELLOW}Checking required tools...${NC}"
MISSING_TOOLS=()

if ! command_exists go; then
    MISSING_TOOLS+=("go")
fi

if ! command_exists git; then
    MISSING_TOOLS+=("git")
fi

if ! command_exists golangci-lint; then
    MISSING_TOOLS+=("golangci-lint (optional but recommended)")
fi

if [ ${#MISSING_TOOLS[@]} -gt 0 ]; then
    echo -e "${RED}Missing tools:${NC}"
    for tool in "${MISSING_TOOLS[@]}"; do
        echo "  - $tool"
    done
    if [[ ! " ${MISSING_TOOLS[@]} " =~ "golangci-lint" ]] || [ ${#MISSING_TOOLS[@]} -gt 1 ]; then
        exit 1
    fi
fi

echo -e "${GREEN}✓ All required tools are installed${NC}"

# Check git status
echo -e "\n${YELLOW}Checking git status...${NC}"
if [ -n "$(git status --porcelain)" ]; then
    echo -e "${RED}✗ Working directory is not clean${NC}"
    echo "Please commit or stash your changes before releasing."
    git status --short
    exit 1
fi
echo -e "${GREEN}✓ Working directory is clean${NC}"

# Check if we're on a branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
echo -e "${GREEN}✓ Current branch: $CURRENT_BRANCH${NC}"

# Run formatting
echo -e "\n${YELLOW}Running code formatting...${NC}"
go fmt ./...
if [ -n "$(git status --porcelain)" ]; then
    echo -e "${RED}✗ Code formatting made changes${NC}"
    echo "Please review and commit formatting changes."
    git status --short
    exit 1
fi
echo -e "${GREEN}✓ Code is properly formatted${NC}"

# Run tests
echo -e "\n${YELLOW}Running tests...${NC}"
if ! go test ./...; then
    echo -e "${RED}✗ Tests failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ All tests passed${NC}"

# Run linter if available
if command_exists golangci-lint; then
    echo -e "\n${YELLOW}Running linter...${NC}"
    if ! golangci-lint run; then
        echo -e "${RED}✗ Linter found issues${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Linter passed${NC}"
fi

# Build test
echo -e "\n${YELLOW}Testing build...${NC}"
if ! make build; then
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Build successful${NC}"

# Test version output
echo -e "\n${YELLOW}Testing version output...${NC}"
./dist/ccmd --version
echo -e "${GREEN}✓ Version command works${NC}"

# Check for existing tag
echo -e "\n${YELLOW}Checking git tags...${NC}"
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "none")
echo "Latest tag: $LATEST_TAG"

# Suggest next version
if [ "$LATEST_TAG" != "none" ]; then
    # Extract version numbers
    VERSION_REGEX="^v?([0-9]+)\.([0-9]+)\.([0-9]+)"
    if [[ $LATEST_TAG =~ $VERSION_REGEX ]]; then
        MAJOR="${BASH_REMATCH[1]}"
        MINOR="${BASH_REMATCH[2]}"
        PATCH="${BASH_REMATCH[3]}"
        
        # Suggest next versions
        NEXT_PATCH="v$MAJOR.$MINOR.$((PATCH + 1))"
        NEXT_MINOR="v$MAJOR.$((MINOR + 1)).0"
        NEXT_MAJOR="v$((MAJOR + 1)).0.0"
        
        echo -e "\n${BLUE}Suggested next versions:${NC}"
        echo "  Patch release: $NEXT_PATCH (bug fixes)"
        echo "  Minor release: $NEXT_MINOR (new features)"
        echo "  Major release: $NEXT_MAJOR (breaking changes)"
    fi
else
    echo -e "\n${BLUE}No tags found. Suggested first version: v0.1.0${NC}"
fi

# Summary
echo -e "\n${GREEN}✅ All pre-release checks passed!${NC}"
echo -e "\n${BLUE}Next steps:${NC}"
echo "1. Create and push a new tag:"
echo "   git tag -a v0.1.0 -m 'Initial release'"
echo "   git push origin v0.1.0"
echo ""
echo "2. The GitHub Actions workflow will automatically:"
echo "   - Build binaries for all platforms"
echo "   - Create release artifacts"
echo "   - Generate changelog"
echo "   - Create GitHub release"
echo ""
echo "3. For local release testing:"
echo "   ./scripts/release-local.sh --snapshot"

# Clean up
make clean