#!/bin/bash
# Generate checksums for release artifacts

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if dist directory exists
if [ ! -d "dist" ]; then
    print_error "dist directory not found. Please build the binaries first."
    exit 1
fi

# Check if there are any .tar.gz files
if [ -z "$(ls dist/*.tar.gz 2>/dev/null)" ]; then
    print_error "No .tar.gz files found in dist directory."
    exit 1
fi

print_status "Generating checksums for release artifacts..."

# Generate checksums
cd dist
shasum -a 256 *.tar.gz > checksums.txt

# Display the checksums
print_status "Generated checksums:"
cat checksums.txt

# Verify checksums
print_status "Verifying checksums..."
shasum -a 256 -c checksums.txt

print_status "Checksums generated successfully in dist/checksums.txt"
