# Release Guide

This guide explains how to use the automated release system for `yfinance-go`.

## Overview

The release system provides:
- **SemVer versioning** with automated version detection
- **Cross-platform binaries** (Linux/macOS, amd64/arm64)
- **Automated GitHub releases** with checksums and release notes
- **Go module versioning** for library usage
- **Comprehensive documentation** and changelog management

## Quick Start

### 1. Create a Release

```bash
# Tag a new version (this triggers the release workflow)
make release-tag v=1.0.0

# Or manually:
git tag v1.0.0
git push origin v1.0.0
```

### 2. Monitor the Release

The GitHub Actions workflow will:
1. Build binaries for all platforms
2. Generate checksums
3. Create a GitHub release with assets
4. Test the release binaries

### 3. Verify the Release

Check the [Releases page](https://github.com/AmpyFin/yfinance-go/releases) for:
- Binary downloads for all platforms
- Checksums file
- Release notes
- Go module availability

## Release Workflow

### Automatic Release Process

When you push a tag matching `v*.*.*`:

1. **Build Phase**:
   - Builds binaries for Linux (amd64/arm64) and macOS (amd64/arm64)
   - Embeds version information via ldflags
   - Creates compressed archives (.tar.gz)

2. **Release Phase**:
   - Generates SHA256 checksums
   - Creates GitHub release with all assets
   - Generates release notes using release-drafter
   - Tests the released binaries

3. **Verification Phase**:
   - Downloads and tests the release binaries
   - Verifies version information
   - Ensures all commands work correctly

### Manual Release Process

For testing or manual releases:

```bash
# Build all platforms locally
make build-all VERSION=v1.0.0

# Generate checksums
make checksums

# Test the binaries
tar -xzf dist/yfin_darwin_arm64.tar.gz
./yfin_darwin_arm64 version
```

## Version Management

### SemVer Rules

- **MAJOR** (X.0.0): Breaking API changes
- **MINOR** (X.Y.0): New features, backward compatible
- **PATCH** (X.Y.Z): Bug fixes, documentation

### Version Information

Version information is embedded at build time:

```bash
# Check version
yfin version

# Output:
# yfin version v1.0.0
# commit: abc123
# build date: 2024-01-15
```

### Go Module Usage

```bash
# Import specific version
go get github.com/AmpyFin/yfinance-go@v1.0.0

# Import latest version
go get github.com/AmpyFin/yfinance-go@latest
```

## Release Assets

### Binary Downloads

Each release includes:
- `yfin_darwin_amd64.tar.gz` - macOS (Intel)
- `yfin_darwin_arm64.tar.gz` - macOS (Apple Silicon)
- `yfin_linux_amd64.tar.gz` - Linux (x86_64)
- `yfin_linux_arm64.tar.gz` - Linux (ARM64)
- `checksums.txt` - SHA256 checksums

### Installation

```bash
# Download and install
curl -L "https://github.com/AmpyFin/yfinance-go/releases/download/v1.0.0/yfin_linux_amd64.tar.gz" | tar xz
sudo mv yfin /usr/local/bin/

# Verify installation
yfin version
```

### Verify Integrity

```bash
# Download checksums
curl -L "https://github.com/AmpyFin/yfinance-go/releases/download/v1.0.0/checksums.txt" -o checksums.txt

# Verify binary
shasum -a 256 -c checksums.txt
```

## Documentation

### Release Notes

Release notes are automatically generated from:
- Conventional commit messages
- Pull request labels
- CHANGELOG.md entries

### Documentation Updates

Each release should include:
- Updated `CHANGELOG.md`
- Updated documentation in `docs/`
- Updated version references

## Development Workflow

### Pre-Release Testing

```bash
# Build snapshot for testing
make release-snapshot

# Test locally
./yfin_snapshot version
./yfin_snapshot --help
```

### Release Preparation

1. **Update CHANGELOG.md**:
   ```markdown
   ## [1.0.0] - 2024-01-15
   
   ### Added
   - Initial release
   - Daily bars fetching
   - Quote snapshots
   - Fundamentals support
   ```

2. **Update version references**:
   - Update any hardcoded version strings
   - Update documentation examples

3. **Run tests**:
   ```bash
   make test
   make test-coverage
   ```

4. **Create release**:
   ```bash
   make release-tag v=1.0.0
   ```

### Post-Release

1. **Verify release**:
   - Check GitHub releases page
   - Test binary downloads
   - Verify Go module availability

2. **Update development**:
   - Merge any release branches
   - Update version to next development version
   - Update documentation

## Troubleshooting

### Common Issues

#### Build Failures

```bash
# Check Go version
go version

# Clean and rebuild
make clean
make build-all VERSION=v1.0.0
```

#### Release Failures

```bash
# Check tag format
git tag -l | grep v

# Verify tag push
git push origin v1.0.0
```

#### Binary Issues

```bash
# Test binary locally
make build VERSION=v1.0.0
./yfin version

# Check binary size
ls -la yfin
```

### GitHub Actions Issues

1. **Check workflow runs**: GitHub Actions tab
2. **Review logs**: Click on failed job
3. **Common fixes**:
   - Update Go version in workflow
   - Fix permission issues
   - Resolve dependency conflicts

### Go Module Issues

```bash
# Clean module cache
go clean -modcache

# Update dependencies
go mod tidy
go mod download
```

## Security

### Binary Verification

All releases include:
- SHA256 checksums
- GitHub's built-in provenance (if enabled)
- Signed commits (if GPG is configured)

### Best Practices

1. **Always verify checksums** before installation
2. **Use official releases** from GitHub
3. **Keep dependencies updated**
4. **Review release notes** for security updates

## Advanced Usage

### Custom Builds

```bash
# Build with custom version
make build VERSION=v1.0.0-rc1 COMMIT=abc123

# Build specific platform
GOOS=linux GOARCH=amd64 go build -o yfin ./cmd/yfin
```

### Release Automation

```bash
# Script for automated releases
#!/bin/bash
VERSION=$1
if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    exit 1
fi

# Update changelog
# ... (automated changelog update)

# Create release
make release-tag v=$VERSION
```

### Integration with CI/CD

The release workflow can be integrated with:
- Automated testing pipelines
- Security scanning
- Dependency updates
- Documentation generation

## Support

For issues with the release system:
1. Check this guide first
2. Review GitHub Actions logs
3. Test locally with `make build-all`
4. Open an issue with details

## References

- [Semantic Versioning](https://semver.org/)
- [Go Modules](https://golang.org/ref/mod)
- [GitHub Actions](https://docs.github.com/en/actions)
- [Release Drafter](https://github.com/release-drafter/release-drafter)
