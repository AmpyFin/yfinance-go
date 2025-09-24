# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

- [Compare v1.0.0...HEAD](https://github.com/yeonlee/yfinance-go/compare/v1.0.0...HEAD)
