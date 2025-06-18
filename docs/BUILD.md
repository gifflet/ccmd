# Build System Documentation

This document describes the ccmd build system, which supports cross-platform compilation for macOS, Linux, and Windows.

## Overview

The build system consists of:
- **Makefile**: Primary build orchestration
- **GoReleaser**: Automated release management
- **GitHub Actions**: CI/CD workflows
- **Build Scripts**: Local development helpers

## Quick Start

### Building for Current Platform
```bash
make build
```

### Building for All Platforms
```bash
make build-all
```

### Creating a Release
```bash
make release
```

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Build for current platform |
| `make build-all` | Build for all supported platforms |
| `make release` | Build and compress all platforms |
| `make clean` | Clean build artifacts |
| `make test` | Run tests |
| `make deps` | Download dependencies |
| `make fmt` | Format code |
| `make lint` | Run linter |
| `make vet` | Run go vet |
| `make install` | Install binary locally |
| `make uninstall` | Uninstall binary |
| `make run` | Build and run |
| `make dev-build` | Build with race detector |
| `make check-tools` | Check for required tools |
| `make version` | Show version info |

## Supported Platforms

The build system supports the following platforms:

- **macOS**
  - `darwin/amd64` (Intel)
  - `darwin/arm64` (Apple Silicon)
- **Linux**
  - `linux/amd64` (x86_64)
  - `linux/arm64` (ARM64)
- **Windows**
  - `windows/amd64` (x86_64)
  - `windows/arm64` (ARM64)

## Version Management

Version information is embedded at build time:
- **Version**: From git tags (`git describe --tags`)
- **Commit**: Short commit hash
- **Build Date**: UTC timestamp

Access version info:
```bash
ccmd --version
```

## Release Process

### Local Release
1. Create a git tag:
   ```bash
   git tag -a v0.1.0 -m "Release v0.1.0"
   ```

2. Build release artifacts:
   ```bash
   make release
   ```

### GitHub Release
Releases are automatically created when pushing tags:
```bash
git push origin v0.1.0
```

This triggers the GitHub Actions workflow that:
1. Runs tests
2. Builds for all platforms
3. Creates checksums
4. Signs artifacts with cosign
5. Generates SBOMs
6. Creates GitHub release
7. Uploads all artifacts

## Binary Naming Convention

Binaries follow this naming pattern:
```
ccmd-<os>-<arch>[.exe]
```

Examples:
- `ccmd-darwin-amd64`
- `ccmd-linux-arm64`
- `ccmd-windows-amd64.exe`

## Scripts

### build.sh
Local build helper with options:
```bash
./scripts/build.sh          # Build for current platform
./scripts/build.sh --all    # Build for all platforms
./scripts/build.sh --dev    # Build with race detector
./scripts/build.sh --static # Build static binary
```

### release-local.sh
Create local release builds:
```bash
./scripts/release-local.sh              # Create release (requires tag)
./scripts/release-local.sh --snapshot   # Create snapshot release
```

### test-build.sh
Test the build system:
```bash
./scripts/test-build.sh  # Run build system tests
```

## GoReleaser Configuration

The `.goreleaser.yaml` file configures:
- Cross-platform builds
- Archive formats (tar.gz for Unix, zip for Windows)
- Changelog generation
- Artifact signing
- SBOM generation
- Package formats (deb, rpm, apk)

## GitHub Actions Workflows

### release.yml
Triggered on version tags (`v*`):
- Builds and tests
- Creates GitHub release
- Uploads artifacts
- Signs with cosign

### ci.yml
Triggered on pushes and PRs:
- Tests on multiple Go versions
- Cross-platform build verification
- Code formatting checks
- Race condition testing
- Coverage reporting

## Development Tips

### Quick Iteration
For rapid development:
```bash
go run ./cmd/ccmd [args]
```

### Testing Cross-Compilation
Verify a specific platform build:
```bash
GOOS=linux GOARCH=arm64 go build -o test-binary ./cmd/ccmd
file test-binary
```

### Static Builds
For deployment without dependencies:
```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -o ccmd ./cmd/ccmd
```

## Troubleshooting

### Build Failures
1. Ensure Go 1.23+ is installed
2. Run `go mod download`
3. Check for syntax errors: `go vet ./...`

### Cross-Compilation Issues
- Some packages may require CGO
- Use `CGO_ENABLED=0` for static builds
- Check platform-specific code

### Version Information
If version shows as "dev":
- Ensure you're building from a git repository
- Check that git tags exist: `git tag -l`