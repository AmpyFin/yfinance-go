# yfinance-go Release

## Installation

### Pre-built Binaries

Download the appropriate binary for your platform:

- **macOS (Intel)**: `yfin_darwin_amd64.tar.gz`
- **macOS (Apple Silicon)**: `yfin_darwin_arm64.tar.gz`
- **Linux (x86_64)**: `yfin_linux_amd64.tar.gz`
- **Linux (ARM64)**: `yfin_linux_arm64.tar.gz`

### Quick Install

```bash
# Download and install (replace with your platform)
curl -L "https://github.com/AmpyFin/yfinance-go/releases/download/${{ github.ref_name }}/yfin_linux_amd64.tar.gz" | tar xz
sudo mv yfin /usr/local/bin/
yfin version
```

### Verify Installation

```bash
# Verify binary integrity
curl -L "https://github.com/AmpyFin/yfinance-go/releases/download/${{ github.ref_name }}/checksums.txt" -o checksums.txt
shasum -a 256 -c checksums.txt
```

## Go Module

```bash
go get github.com/AmpyFin/yfinance-go@${{ github.ref_name }}
```

## Documentation

- [Installation Guide](https://github.com/AmpyFin/yfinance-go/blob/main/docs/install.md)
- [Usage Guide](https://github.com/AmpyFin/yfinance-go/blob/main/docs/usage.md)
- [Versioning Policy](https://github.com/AmpyFin/yfinance-go/blob/main/docs/versioning.md)

## What's New

