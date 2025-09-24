# Installation Guide

This guide covers different ways to install `yfinance-go` on your system.

## Pre-built Binaries (Recommended)

### Download from GitHub Releases

1. Go to the [Releases page](https://github.com/yeonlee/yfinance-go/releases)
2. Download the appropriate binary for your platform:
   - `yfin_darwin_amd64.tar.gz` - macOS (Intel)
   - `yfin_darwin_arm64.tar.gz` - macOS (Apple Silicon)
   - `yfin_linux_amd64.tar.gz` - Linux (x86_64)
   - `yfin_linux_arm64.tar.gz` - Linux (ARM64)

3. Extract and install:

```bash
# Extract the archive
tar xzf yfin_darwin_arm64.tar.gz

# Move to a directory in your PATH
sudo mv yfin /usr/local/bin/

# Verify installation
yfin version
```

### Using curl (macOS/Linux)

```bash
# Set variables
VERSION="v1.0.0"
OS="darwin"  # or "linux"
ARCH="arm64" # or "amd64"

# Download and install
curl -L "https://github.com/yeonlee/yfinance-go/releases/download/${VERSION}/yfin_${OS}_${ARCH}.tar.gz" | tar xz
sudo mv yfin /usr/local/bin/
yfin version
```

### Verify Installation

```bash
yfin version
```

Expected output:
```
yfin version v1.0.0
commit: abc123
build date: 2024-01-15
```

## Build from Source

### Prerequisites

- Go 1.23 or later
- Git

### Build Steps

```bash
# Clone the repository
git clone https://github.com/yeonlee/yfinance-go.git
cd yfinance-go

# Build the binary
go build -o yfin ./cmd/yfin

# Install to system
sudo mv yfin /usr/local/bin/

# Verify installation
yfin version
```

### Build with Version Information

```bash
# Build with version details
go build -ldflags="-X main.version=v1.0.0 -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%d)" -o yfin ./cmd/yfin
```

## Go Module Installation

If you want to use `yfinance-go` as a Go library:

```bash
go get github.com/yeonlee/yfinance-go@v1.0.0
```

Then import in your Go code:

```go
import "github.com/yeonlee/yfinance-go"
```

## Package Managers

### Homebrew (macOS) - Optional

If a Homebrew formula is available:

```bash
# Add the tap
brew tap yeonlee/yfinance-go

# Install
brew install yfin

# Verify
yfin version
```

### APT (Ubuntu/Debian) - Future

Package for APT repositories may be available in the future.

### YUM/DNF (RHEL/CentOS/Fedora) - Future

Package for YUM/DNF repositories may be available in the future.

## Docker

### Using Pre-built Image

```bash
# Pull the image
docker pull ghcr.io/yeonlee/yfinance-go:latest

# Run a command
docker run --rm ghcr.io/yeonlee/yfinance-go:latest yfin version
```

### Build Custom Image

```dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o yfin ./cmd/yfin

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/yfin .
CMD ["./yfin"]
```

Build and run:

```bash
docker build -t yfinance-go .
docker run --rm yfinance-go yfin version
```

## Configuration

### Default Configuration

`yfinance-go` uses configuration files in the `configs/` directory:

- `configs/effective.yaml` - Default configuration
- `configs/example.dev.yaml` - Development environment
- `configs/example.staging.yaml` - Staging environment
- `configs/example.prod.yaml` - Production environment

### Custom Configuration

```bash
# Use custom config file
yfin pull --config ./my-config.yaml --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview
```

### Environment Variables

Some configuration can be overridden with environment variables:

```bash
export YFINANCE_LOG_LEVEL=debug
export YFINANCE_CONCURRENCY=32
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview
```

## Verification

### Test Installation

```bash
# Check version
yfin version

# Test basic functionality
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-01-02 --preview

# Test configuration
yfin config --print-effective
```

### Expected Output

```bash
$ yfin version
yfin version v1.0.0
commit: abc123
build date: 2024-01-15

$ yfin pull --ticker AAPL --start 2024-01-01 --end 2024-01-02 --preview
RUN yfin_1704067200  (env=dev, topic_prefix=ampy)
SYMBOL AAPL (MIC=XNAS, CCY=USD)  range=2024-01-01..2024-01-02  bars=1  adjusted=split_dividend
first=2024-01-01T00:00:00Z  last=2024-01-02T00:00:00Z  last_close=192.5300 USD
```

## Troubleshooting

### Common Issues

#### Permission Denied

```bash
# Make sure the binary is executable
chmod +x yfin

# Check PATH
echo $PATH
which yfin
```

#### Go Module Issues

```bash
# Clean module cache
go clean -modcache

# Update dependencies
go mod tidy
go mod download
```

#### Configuration Issues

```bash
# Check configuration
yfin config --print-effective

# Validate configuration file
yfin config --config ./configs/example.dev.yaml --print-effective
```

### Getting Help

```bash
# Show help
yfin --help

# Show command-specific help
yfin pull --help
yfin quote --help
yfin fundamentals --help
yfin config --help
yfin version --help
```

### Logs and Debugging

```bash
# Enable debug logging
yfin --log-level debug pull --ticker AAPL --start 2024-01-01 --end 2024-01-02 --preview

# Disable observability for testing
yfin --observability-disable-tracing --observability-disable-metrics pull --ticker AAPL --start 2024-01-01 --end 2024-01-02 --preview
```

## Uninstallation

### Remove Binary

```bash
# Remove from system
sudo rm /usr/local/bin/yfin

# Or if installed in user directory
rm ~/bin/yfin
```

### Remove Go Module

```bash
# Remove from go.mod
go mod edit -droprequire github.com/yeonlee/yfinance-go

# Clean up
go mod tidy
```

### Remove Homebrew Package

```bash
# Uninstall
brew uninstall yfin

# Remove tap (if no other packages from tap)
brew untap yeonlee/yfinance-go
```

## Security

### Verify Binary Integrity

```bash
# Download checksums
curl -L "https://github.com/yeonlee/yfinance-go/releases/download/v1.0.0/checksums.txt" -o checksums.txt

# Verify downloaded binary
shasum -a 256 -c checksums.txt
```

### GPG Signatures

If GPG signatures are available:

```bash
# Import public key
gpg --keyserver keyserver.ubuntu.com --recv-keys <KEY_ID>

# Verify signature
gpg --verify yfin_darwin_arm64.tar.gz.asc yfin_darwin_arm64.tar.gz
```

## Next Steps

After installation, see:

- [Usage Guide](usage.md) - Learn how to use yfinance-go
- [Versioning Policy](versioning.md) - Understand versioning and compatibility
- [Configuration](https://github.com/yeonlee/yfinance-go/tree/main/configs) - Configure the tool for your environment
