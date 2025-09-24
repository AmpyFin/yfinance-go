# Versioning Policy

This document outlines the versioning strategy for `yfinance-go` following [Semantic Versioning (SemVer)](https://semver.org/) principles.

## Version Format

Versions follow the format: `MAJOR.MINOR.PATCH` (e.g., `1.2.3`)

- **MAJOR**: Breaking changes to public Go API or on-the-wire schemas
- **MINOR**: Backward-compatible features and enhancements
- **PATCH**: Bug fixes and non-behavioral changes

## Versioning Rules

### MAJOR Version (X.0.0)

Increment when you make **breaking changes** to:

- Public Go API (exported functions, types, interfaces)
- Default behavior of CLI commands
- On-the-wire schemas for messages this module emits by default
- Configuration file format (breaking changes only)

**Examples:**
- Removing or renaming exported functions
- Changing function signatures
- Changing default CLI behavior
- Breaking changes to ampy-proto message schemas

### MINOR Version (X.Y.0)

Increment when you add **backward-compatible** features:

- New CLI flags (default-off)
- New exported functions or types
- New configuration options
- Performance improvements
- New endpoints or data sources
- Additional output formats

**Examples:**
- Adding `--new-flag` to existing commands
- Adding new subcommands
- Supporting new output formats
- Adding new configuration options

### PATCH Version (X.Y.Z)

Increment for:

- Bug fixes
- Documentation updates
- Internal refactoring (no API changes)
- Security patches
- Dependency updates (without API changes)

**Examples:**
- Fixing incorrect data parsing
- Updating documentation
- Fixing memory leaks
- Security vulnerability patches

## Schema Compatibility

### ampy-proto Schemas

This module emits messages using `ampy-proto` schemas. Version compatibility:

- **v1 schemas**: Default and stable
- **v2+ schemas**: New versions added alongside v1, not replacing
- **Breaking schema changes**: Require MAJOR version bump

### Configuration Schema

- **Breaking config changes**: MAJOR version bump
- **Additive config changes**: MINOR version bump
- **Config bug fixes**: PATCH version bump

## Deprecation Policy

### CLI Flags and Commands

1. **Announcement**: Deprecate in version N
2. **Warning**: Show deprecation warnings in version N+1
3. **Removal**: Remove in version N+2 (MAJOR bump)

### Go API

1. **Announcement**: Mark as deprecated in version N
2. **Warning**: Add compile-time warnings in version N+1
3. **Removal**: Remove in version N+2 (MAJOR bump)

## Release Process

### Pre-release

1. Update `CHANGELOG.md` with changes
2. Ensure all tests pass
3. Update version in `main.go` (if needed)
4. Create release branch from `main`

### Release

1. Tag the release: `git tag v1.2.3`
2. Push tag: `git push origin v1.2.3`
3. GitHub Actions automatically builds and publishes release
4. Update `CHANGELOG.md` with release date

### Post-release

1. Merge release branch back to `main`
2. Update version to next development version
3. Update documentation if needed

## Version Information

### Build-time Version

Version information is embedded at build time using ldflags:

```bash
go build -ldflags="-X main.version=v1.2.3 -X main.commit=abc123 -X main.date=2024-01-15"
```

### Runtime Version

Check version at runtime:

```bash
yfin version
```

Output:
```
yfin version v1.2.3
commit: abc123
build date: 2024-01-15
```

## Go Module Versioning

### Module Path

```
github.com/yeonlee/yfinance-go
```

### Import Examples

```go
// Latest version
import "github.com/yeonlee/yfinance-go"

// Specific version
import "github.com/yeonlee/yfinance-go@v1.2.3"

// Latest minor version of v1
import "github.com/yeonlee/yfinance-go@v1"
```

### Version Constraints

Use semantic version constraints in `go.mod`:

```go
require github.com/yeonlee/yfinance-go v1.2.3
```

## Compatibility Matrix

| yfinance-go | Go Version | ampy-proto | ampy-config |
|-------------|------------|------------|-------------|
| v1.0.x      | 1.23+      | v2.1.x     | v1.1.x      |
| v1.1.x      | 1.23+      | v2.1.x     | v1.1.x      |
| v2.0.x      | 1.23+      | v3.0.x     | v2.0.x      |

## Migration Guides

### v1.0 to v1.1

No breaking changes. New features available:
- Additional CLI flags
- New configuration options

### v1.x to v2.0

Breaking changes expected:
- Updated ampy-proto schemas
- Modified CLI behavior
- Configuration format changes

Migration guide will be provided in release notes.

## Support Policy

### Supported Versions

- **Current**: Latest release
- **Previous**: Previous major version (security fixes only)
- **LTS**: Long-term support versions (if applicable)

### End-of-Life

- **Announcement**: 6 months before EOL
- **Security fixes**: 3 months after EOL
- **Full support**: Until EOL date

## Examples

### Version Bump Examples

```bash
# Bug fix
v1.2.3 → v1.2.4

# New feature
v1.2.3 → v1.3.0

# Breaking change
v1.2.3 → v2.0.0
```

### Git Tag Examples

```bash
# Create and push tag
git tag v1.2.3
git push origin v1.2.3

# Delete tag (if needed)
git tag -d v1.2.3
git push origin :refs/tags/v1.2.3
```

## References

- [Semantic Versioning](https://semver.org/)
- [Go Modules](https://golang.org/ref/mod)
- [Keep a Changelog](https://keepachangelog.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)
