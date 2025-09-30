#!/bin/bash

# preview_aapl.sh - Comprehensive CLI examples for AAPL data preview
# This script demonstrates various CLI usage patterns for the yfinance-go scrape system

set -e  # Exit on any error

echo "=== yfinance-go CLI Preview Examples for AAPL ==="
echo "================================================="

# Configuration
CONFIG_FILE="configs/example.dev.yaml"
TICKER="AAPL"
RUN_ID="cli-preview-$(date +%s)"

# Check if config file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Warning: Config file $CONFIG_FILE not found, using defaults"
    CONFIG_FILE=""
fi

# Function to run command with error handling
run_command() {
    local description="$1"
    shift
    echo ""
    echo "--- $description ---"
    echo "Command: $*"
    echo ""
    
    if "$@"; then
        echo "✓ Success"
    else
        echo "✗ Failed with exit code $?"
    fi
}

# Example 1: Basic Key Statistics Preview
run_command "Basic Key Statistics Preview" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --ticker "$TICKER" \
    --endpoint key-statistics \
    --preview

# Example 2: Multiple Endpoints JSON Preview
run_command "Multiple Endpoints JSON Preview" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --ticker "$TICKER" \
    --endpoints key-statistics,financials,analysis,profile \
    --preview-json

# Example 3: News Articles Preview
run_command "News Articles Preview" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --ticker "$TICKER" \
    --endpoint news \
    --preview-news

# Example 4: Proto Summary Preview
run_command "Proto Summary Preview" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --ticker "$TICKER" \
    --endpoints key-statistics,financials,analysis \
    --preview-proto

# Example 5: Endpoint Health Check
run_command "Endpoint Health Check" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --ticker "$TICKER" \
    --endpoint key-statistics \
    --check

# Example 6: Fallback Strategy Testing
echo ""
echo "--- Fallback Strategy Examples ---"

# Test API-only mode (may fail for some endpoints)
run_command "API-Only Mode Test" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --ticker "$TICKER" \
    --endpoint quote \
    --fallback api-only \
    --preview

# Test scrape-only mode
run_command "Scrape-Only Mode Test" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --ticker "$TICKER" \
    --endpoint key-statistics \
    --fallback scrape-only \
    --preview

# Test automatic fallback
run_command "Automatic Fallback Test" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --ticker "$TICKER" \
    --endpoint key-statistics \
    --fallback auto \
    --preview

# Example 7: Rate Limiting and Performance Testing
echo ""
echo "--- Performance and Rate Limiting Examples ---"

# Conservative rate limiting
run_command "Conservative Rate Limiting" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --ticker "$TICKER" \
    --endpoint key-statistics \
    --qps 0.5 \
    --timeout 60s \
    --preview

# With session rotation
run_command "With Session Rotation" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --ticker "$TICKER" \
    --endpoint key-statistics \
    --sessions 5 \
    --preview

# Example 8: Debug and Troubleshooting
echo ""
echo "--- Debug and Troubleshooting Examples ---"

# Debug mode
run_command "Debug Mode" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --ticker "$TICKER" \
    --endpoint key-statistics \
    --log-level debug \
    --preview

# Dry run mode
run_command "Dry Run Mode" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --ticker "$TICKER" \
    --endpoint key-statistics \
    --dry-run

# Example 9: Configuration Testing
echo ""
echo "--- Configuration Examples ---"

# Print effective configuration
run_command "Print Effective Configuration" \
    yfin config \
    ${CONFIG_FILE:+--file "$CONFIG_FILE"} \
    --print-effective

# Validate configuration
run_command "Validate Configuration" \
    yfin config \
    ${CONFIG_FILE:+--file "$CONFIG_FILE"} \
    --validate

# Example 10: Comprehensive Data Collection
echo ""
echo "--- Comprehensive Data Collection ---"

# Collect all available data types
ENDPOINTS="key-statistics,financials,analysis,profile,news"
run_command "All Endpoints Preview" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --ticker "$TICKER" \
    --endpoints "$ENDPOINTS" \
    --preview-json

# Example 11: Error Handling Demonstration
echo ""
echo "--- Error Handling Examples ---"

# Test with invalid ticker (should fail gracefully)
run_command "Invalid Ticker Test" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --ticker "INVALID_TICKER_12345" \
    --endpoint key-statistics \
    --preview

# Test with very short timeout (should timeout)
run_command "Timeout Test" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --ticker "$TICKER" \
    --endpoint key-statistics \
    --timeout 1s \
    --preview

# Example 12: International Ticker Testing
echo ""
echo "--- International Ticker Examples ---"

# Test Hong Kong stock
run_command "Hong Kong Stock (0700.HK)" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --ticker "0700.HK" \
    --endpoint key-statistics \
    --preview

# Test European stock
run_command "European Stock (SAP)" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --ticker "SAP" \
    --endpoint key-statistics \
    --preview

# Summary
echo ""
echo "=== CLI Preview Examples Summary ==="
echo "✓ Basic scraping operations"
echo "✓ Multiple endpoint handling"
echo "✓ Fallback strategy testing"
echo "✓ Rate limiting and performance"
echo "✓ Debug and troubleshooting"
echo "✓ Configuration management"
echo "✓ Error handling"
echo "✓ International ticker support"
echo ""
echo "For more examples, see:"
echo "- examples/cli/soak_smoke.sh - Soak testing examples"
echo "- examples/cli/batch_processing.sh - Batch processing examples"
echo "- docs/scrape/cli.md - Complete CLI documentation"
