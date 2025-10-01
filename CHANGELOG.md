# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.2.0] - 2025-01-24

### Added
- Comprehensive API Reference documentation with method capabilities and limitations
- Detailed Data Structures guide with field naming conventions and scaled decimal format
- Complete working examples with data processing, formatting, and error handling
- Method comparison table and use case guidance
- Comprehensive error handling and troubleshooting guide
- Migration guide from Python yfinance with feature comparison
- Data quality and validation guidelines
- Performance optimization and best practices guide
- Documentation improvements summary

### Changed
- Significantly improved developer experience with comprehensive documentation
- Enhanced README.md with organized documentation links
- Better navigation and discoverability of documentation resources

### Fixed
- Addressed all critical documentation gaps identified by user feedback
- Clarified API method capabilities and limitations
- Explained data structure naming conventions and scaled decimal format
- Provided complete working examples for all use cases
- Added comprehensive error handling guidance
- Included migration support from Python yfinance

## [1.0.6] - 2025-01-24

### Fixed
- Resolved Windows PowerShell syntax error in CI workflow
- Replaced CMD syntax with PowerShell for Windows coverage merging
- Fixed Windows package resolution error with proper PowerShell file operations
- All platforms now have reliable coverage file merging

### Changed
- Improved Windows CI workflow reliability using PowerShell instead of CMD
- Enhanced cross-platform compatibility for coverage reporting

## [1.0.5] - 2025-01-24

### Fixed
- Resolved Windows PowerShell syntax error in CI workflow
- Split coverage merging into separate steps for Unix and Windows
- Used proper Windows CMD syntax for Windows runners
- Used bash syntax for Unix runners (Linux, macOS)

## [1.0.4] - 2025-01-24

### Fixed
- Resolved Windows CI workflow issues with coverage file merging
- Fixed coverage file existence checks to prevent merge failures
- Removed gosec security scanner due to repository access issues
- Security scanning is now handled by CodeQL analysis
- All tests now pass successfully across all platforms (Linux, macOS, Windows)

### Changed
- Improved CI workflow robustness and cross-platform compatibility
- Enhanced error handling for missing coverage files

## [1.0.3] - 2025-01-24

### Fixed
- Resolved race conditions in observability tests by adding proper synchronization
- Updated CI workflow to handle external dependency race conditions gracefully
- Fixed missing cmd/yfin directory that was causing GitHub Actions build failures
- All tests now pass successfully with comprehensive coverage reporting

### Changed
- Improved CI/CD pipeline reliability and test execution
- Enhanced observability state management with mutex protection

## [1.0.2] - 2025-01-24

### Fixed
- Resolved all golangci-lint issues for CI compliance
- Fixed unchecked error returns in HTTP handlers and observability shutdown
- Removed unused functions and imports
- Updated deprecated OpenTelemetry API usage
- Fixed staticcheck warnings for better code quality

### Changed
- Improved CI/CD pipeline reliability
- Enhanced code quality and maintainability

## [1.0.1] - 2025-01-24

### Fixed
- Resolved Go module tidiness issues
- Fixed all linting errors and warnings

## [1.0.0] - 2025-01-24

### Added
- Initial implementation of yfinance-go CLI tool
- Support for fetching daily bars from Yahoo Finance
- Support for fetching snapshot quotes
- Support for fetching fundamentals (requires paid subscription)
- FX conversion preview functionality
- Bus publishing with NATS and Kafka backends
- Local export in JSON format
- Configuration management with ampy-config
- Observability with OpenTelemetry tracing and Prometheus metrics
- Comprehensive test suite with golden file testing
- Cross-language roundtrip testing with Python

### Changed
- N/A

### Fixed
- N/A

### Security
- N/A

## [1.0.0] - 2024-01-XX

### Added
- Initial release of yfinance-go
- CLI tool with pull, quote, fundamentals, config, and version commands
- Support for daily bars fetching with adjustment policies
- Quote snapshot functionality
- Fundamentals data fetching (paid subscription required)
- FX conversion preview
- Bus publishing with retry and circuit breaker
- Local export capabilities
- Comprehensive configuration system
- Observability and monitoring
- Cross-platform binary releases (Linux/macOS, amd64/arm64)

### Changed
- N/A

### Fixed
- N/A

### Security
- N/A

---

## Release Notes Format

Each release should include:

- **Added** for new features
- **Changed** for changes in existing functionality
- **Deprecated** for soon-to-be removed features
- **Removed** for now removed features
- **Fixed** for any bug fixes
- **Security** for vulnerability fixes

## Links

- [Compare v1.0.0...HEAD](https://github.com/AmpyFin/yfinance-go/compare/v1.0.0...HEAD)
