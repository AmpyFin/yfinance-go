#!/bin/bash

# soak_smoke.sh - Soak testing examples for yfinance-go
# This script demonstrates various soak testing scenarios

set -e  # Exit on any error

echo "=== yfinance-go Soak Testing Examples ==="
echo "========================================"

# Configuration
CONFIG_FILE="configs/example.dev.yaml"
UNIVERSE_FILE="testdata/universe/soak.txt"

# Check if files exist
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Warning: Config file $CONFIG_FILE not found, using defaults"
    CONFIG_FILE=""
fi

if [ ! -f "$UNIVERSE_FILE" ]; then
    echo "Error: Universe file $UNIVERSE_FILE not found"
    echo "Creating a small test universe..."
    mkdir -p "$(dirname "$UNIVERSE_FILE")"
    cat > "$UNIVERSE_FILE" << EOF
# Test universe for soak testing
AAPL
MSFT
GOOGL
TSLA
AMZN
EOF
    echo "Created test universe with 5 tickers"
fi

# Function to run soak test with error handling
run_soak_test() {
    local description="$1"
    shift
    echo ""
    echo "--- $description ---"
    echo "Command: yfin soak $*"
    echo ""
    
    if yfin soak "$@"; then
        echo "✓ Soak test completed successfully"
    else
        local exit_code=$?
        echo "✗ Soak test failed with exit code $exit_code"
        return $exit_code
    fi
}

# Example 1: Quick Smoke Test (30 seconds)
run_soak_test "Quick Smoke Test (30 seconds)" \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$UNIVERSE_FILE" \
    --endpoints key-statistics \
    --fallback auto \
    --duration 30s \
    --concurrency 2 \
    --qps 1.0 \
    --preview

# Example 2: Multi-Endpoint Test (2 minutes)
run_soak_test "Multi-Endpoint Test (2 minutes)" \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$UNIVERSE_FILE" \
    --endpoints key-statistics,financials,news \
    --fallback auto \
    --duration 2m \
    --concurrency 4 \
    --qps 2.0 \
    --preview

# Example 3: Memory Leak Detection Test (3 minutes)
run_soak_test "Memory Leak Detection Test (3 minutes)" \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$UNIVERSE_FILE" \
    --endpoints key-statistics,financials \
    --fallback auto \
    --duration 3m \
    --concurrency 6 \
    --qps 3.0 \
    --memory-check \
    --preview

# Example 4: Correctness Probe Test (5 minutes)
run_soak_test "Correctness Probe Test (5 minutes)" \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$UNIVERSE_FILE" \
    --endpoints key-statistics,financials,analysis \
    --fallback auto \
    --duration 5m \
    --concurrency 4 \
    --qps 2.0 \
    --probe-interval 2m \
    --preview

# Example 5: High Concurrency Test (2 minutes)
run_soak_test "High Concurrency Test (2 minutes)" \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$UNIVERSE_FILE" \
    --endpoints key-statistics \
    --fallback auto \
    --duration 2m \
    --concurrency 10 \
    --qps 5.0 \
    --preview

# Example 6: Fallback Strategy Testing
echo ""
echo "--- Fallback Strategy Testing ---"

# API-only soak test (may have limited data)
run_soak_test "API-Only Soak Test (1 minute)" \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$UNIVERSE_FILE" \
    --endpoints quote \
    --fallback api-only \
    --duration 1m \
    --concurrency 3 \
    --qps 2.0 \
    --preview

# Scrape-only soak test
run_soak_test "Scrape-Only Soak Test (2 minutes)" \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$UNIVERSE_FILE" \
    --endpoints key-statistics,financials \
    --fallback scrape-only \
    --duration 2m \
    --concurrency 4 \
    --qps 1.5 \
    --preview

# Example 7: Failure Injection Testing
echo ""
echo "--- Failure Injection Testing ---"

# Test with simulated failures
run_soak_test "Failure Injection Test (2 minutes)" \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$UNIVERSE_FILE" \
    --endpoints key-statistics,financials \
    --fallback auto \
    --duration 2m \
    --concurrency 4 \
    --qps 2.0 \
    --failure-rate 0.1 \
    --preview

# Example 8: Conservative Production-Like Test
echo ""
echo "--- Production-Like Testing ---"

# Conservative settings similar to production
run_soak_test "Conservative Production-Like Test (5 minutes)" \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$UNIVERSE_FILE" \
    --endpoints key-statistics,financials,analysis,profile,news \
    --fallback auto \
    --duration 5m \
    --concurrency 6 \
    --qps 1.0 \
    --memory-check \
    --probe-interval 2m \
    --preview

# Example 9: News-Focused Soak Test
echo ""
echo "--- News-Focused Testing ---"

# Focus on news endpoint (good for testing parsing variety)
run_soak_test "News-Focused Soak Test (3 minutes)" \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$UNIVERSE_FILE" \
    --endpoints news \
    --fallback scrape-only \
    --duration 3m \
    --concurrency 3 \
    --qps 1.0 \
    --preview

# Example 10: Stress Test (Short Duration, High Load)
echo ""
echo "--- Stress Testing ---"

# Short but intense test
run_soak_test "Stress Test (1 minute, high load)" \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$UNIVERSE_FILE" \
    --endpoints key-statistics \
    --fallback auto \
    --duration 1m \
    --concurrency 15 \
    --qps 8.0 \
    --preview

# Example 11: Custom Universe Test
echo ""
echo "--- Custom Universe Testing ---"

# Create a custom universe for specific testing
CUSTOM_UNIVERSE="/tmp/custom-soak-universe.txt"
cat > "$CUSTOM_UNIVERSE" << EOF
# Custom universe for specific testing
AAPL
MSFT
GOOGL
0700.HK
SAP
NESN.SW
EOF

run_soak_test "Custom Universe Test (2 minutes)" \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$CUSTOM_UNIVERSE" \
    --endpoints key-statistics,financials \
    --fallback auto \
    --duration 2m \
    --concurrency 4 \
    --qps 2.0 \
    --preview

# Clean up custom universe
rm -f "$CUSTOM_UNIVERSE"

# Example 12: Publishing Test (Dry Run)
echo ""
echo "--- Publishing Testing ---"

# Test publishing pipeline (dry run mode)
run_soak_test "Publishing Pipeline Test (1 minute, dry run)" \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$UNIVERSE_FILE" \
    --endpoints news \
    --fallback auto \
    --duration 1m \
    --concurrency 3 \
    --qps 1.0 \
    --publish \
    --env dev \
    --topic-prefix ampy.dev \
    --preview

# Summary and Analysis
echo ""
echo "=== Soak Testing Summary ==="
echo "✓ Quick smoke tests completed"
echo "✓ Multi-endpoint testing verified"
echo "✓ Memory leak detection tested"
echo "✓ Correctness probes validated"
echo "✓ Fallback strategies tested"
echo "✓ Failure injection verified"
echo "✓ Production-like scenarios tested"
echo "✓ Stress testing completed"
echo ""
echo "Key Observations:"
echo "- Monitor memory usage during longer tests"
echo "- Adjust QPS based on error rates"
echo "- Use session rotation for better reliability"
echo "- Enable correctness probes for data quality"
echo ""
echo "Next Steps:"
echo "1. Run longer soak tests (30m-2h) in staging"
echo "2. Monitor metrics during soak tests"
echo "3. Adjust configuration based on results"
echo "4. Set up automated soak testing in CI/CD"
echo ""
echo "For more information:"
echo "- docs/soak-testing.md - Complete soak testing guide"
echo "- docs/observability.md - Monitoring and metrics"
echo "- runbooks/incident-playbook.md - Incident response"
