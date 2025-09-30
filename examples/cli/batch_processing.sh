#!/bin/bash

# batch_processing.sh - Batch processing examples for yfinance-go
# This script demonstrates various batch processing patterns and best practices

set -e  # Exit on any error

echo "=== yfinance-go Batch Processing Examples ==="
echo "============================================="

# Configuration
CONFIG_FILE="configs/example.dev.yaml"
OUTPUT_DIR="./batch_output"
LOG_DIR="./batch_logs"

# Create output directories
mkdir -p "$OUTPUT_DIR" "$LOG_DIR"

# Check if config file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Warning: Config file $CONFIG_FILE not found, using defaults"
    CONFIG_FILE=""
fi

# Function to create universe files
create_universe_file() {
    local filename="$1"
    local description="$2"
    shift 2
    local tickers=("$@")
    
    echo "Creating $description universe: $filename"
    cat > "$filename" << EOF
# $description
# Generated on $(date)
EOF
    
    for ticker in "${tickers[@]}"; do
        echo "$ticker" >> "$filename"
    done
    
    echo "Created universe with ${#tickers[@]} tickers"
}

# Function to run batch command with logging
run_batch_command() {
    local description="$1"
    local log_file="$LOG_DIR/$(echo "$description" | tr ' ' '_' | tr '[:upper:]' '[:lower:]').log"
    shift
    
    echo ""
    echo "--- $description ---"
    echo "Command: $*"
    echo "Log file: $log_file"
    echo ""
    
    if "$@" 2>&1 | tee "$log_file"; then
        echo "✓ Batch processing completed successfully"
        return 0
    else
        local exit_code=$?
        echo "✗ Batch processing failed with exit code $exit_code"
        echo "Check log file: $log_file"
        return $exit_code
    fi
}

# Example 1: Small Universe Processing
echo ""
echo "=== Example 1: Small Universe Processing ==="

SMALL_UNIVERSE="$OUTPUT_DIR/small_universe.txt"
create_universe_file "$SMALL_UNIVERSE" "Small Test Universe" \
    "AAPL" "MSFT" "GOOGL" "AMZN" "TSLA"

run_batch_command "Small Universe Key Statistics" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$SMALL_UNIVERSE" \
    --endpoint key-statistics \
    --fallback auto \
    --qps 1.0 \
    --concurrency 3 \
    --preview

# Example 2: Large Universe Processing with Rate Limiting
echo ""
echo "=== Example 2: Large Universe Processing ==="

LARGE_UNIVERSE="$OUTPUT_DIR/large_universe.txt"
create_universe_file "$LARGE_UNIVERSE" "Large Test Universe" \
    "AAPL" "MSFT" "GOOGL" "AMZN" "TSLA" "META" "NFLX" "NVDA" "CRM" "ADBE" \
    "PYPL" "INTC" "CSCO" "PEP" "KO" "WMT" "HD" "DIS" "VZ" "T"

run_batch_command "Large Universe Conservative Processing" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$LARGE_UNIVERSE" \
    --endpoint key-statistics \
    --fallback auto \
    --qps 0.5 \
    --concurrency 4 \
    --timeout 60s \
    --preview

# Example 3: Multi-Endpoint Batch Processing
echo ""
echo "=== Example 3: Multi-Endpoint Batch Processing ==="

MULTI_ENDPOINT_UNIVERSE="$OUTPUT_DIR/multi_endpoint_universe.txt"
create_universe_file "$MULTI_ENDPOINT_UNIVERSE" "Multi-Endpoint Universe" \
    "AAPL" "MSFT" "GOOGL" "TSLA" "AMZN"

ENDPOINTS="key-statistics,financials,analysis,profile"

run_batch_command "Multi-Endpoint Batch Processing" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$MULTI_ENDPOINT_UNIVERSE" \
    --endpoints "$ENDPOINTS" \
    --fallback auto \
    --qps 1.0 \
    --concurrency 2 \
    --preview-json

# Example 4: International Universe Processing
echo ""
echo "=== Example 4: International Universe Processing ==="

INTERNATIONAL_UNIVERSE="$OUTPUT_DIR/international_universe.txt"
create_universe_file "$INTERNATIONAL_UNIVERSE" "International Universe" \
    "AAPL" "MSFT" "0700.HK" "BABA" "TSM" "ASML" "SAP" "NESN.SW" "005930.KS" "7203.T"

run_batch_command "International Universe Processing" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$INTERNATIONAL_UNIVERSE" \
    --endpoint key-statistics \
    --fallback auto \
    --qps 0.8 \
    --concurrency 3 \
    --timeout 45s \
    --preview

# Example 5: Sector-Based Processing
echo ""
echo "=== Example 5: Sector-Based Processing ==="

# Technology sector
TECH_UNIVERSE="$OUTPUT_DIR/tech_universe.txt"
create_universe_file "$TECH_UNIVERSE" "Technology Sector" \
    "AAPL" "MSFT" "GOOGL" "AMZN" "META" "NFLX" "NVDA" "CRM" "ADBE" "INTC"

# Healthcare sector
HEALTHCARE_UNIVERSE="$OUTPUT_DIR/healthcare_universe.txt"
create_universe_file "$HEALTHCARE_UNIVERSE" "Healthcare Sector" \
    "JNJ" "PFE" "UNH" "ABBV" "TMO" "DHR" "BMY" "AMGN" "GILD" "BIIB"

# Process each sector
for sector_file in "$TECH_UNIVERSE" "$HEALTHCARE_UNIVERSE"; do
    sector_name=$(basename "$sector_file" .txt | sed 's/_/ /g' | sed 's/\b\w/\U&/g')
    
    run_batch_command "$sector_name Sector Processing" \
        yfin scrape \
        ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
        --universe-file "$sector_file" \
        --endpoint key-statistics \
        --fallback auto \
        --qps 1.0 \
        --concurrency 4 \
        --preview
done

# Example 6: Error Handling and Recovery
echo ""
echo "=== Example 6: Error Handling and Recovery ==="

# Create universe with some invalid tickers
ERROR_TEST_UNIVERSE="$OUTPUT_DIR/error_test_universe.txt"
create_universe_file "$ERROR_TEST_UNIVERSE" "Error Testing Universe" \
    "AAPL" "INVALID_TICKER_1" "MSFT" "INVALID_TICKER_2" "GOOGL"

run_batch_command "Error Handling Test" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$ERROR_TEST_UNIVERSE" \
    --endpoint key-statistics \
    --fallback auto \
    --qps 1.0 \
    --concurrency 2 \
    --timeout 30s \
    --preview || echo "Expected some failures due to invalid tickers"

# Example 7: News Batch Processing
echo ""
echo "=== Example 7: News Batch Processing ==="

NEWS_UNIVERSE="$OUTPUT_DIR/news_universe.txt"
create_universe_file "$NEWS_UNIVERSE" "News Universe" \
    "AAPL" "TSLA" "GME" "AMC" "NVDA"  # Stocks that typically have lots of news

run_batch_command "News Batch Processing" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$NEWS_UNIVERSE" \
    --endpoint news \
    --fallback scrape-only \
    --qps 0.5 \
    --concurrency 2 \
    --timeout 60s \
    --preview-news

# Example 8: Parallel Processing with Custom Script
echo ""
echo "=== Example 8: Parallel Processing Script ==="

# Create a custom parallel processing script
PARALLEL_SCRIPT="$OUTPUT_DIR/parallel_process.sh"
cat > "$PARALLEL_SCRIPT" << 'EOF'
#!/bin/bash

# parallel_process.sh - Process tickers in parallel with proper rate limiting

process_ticker() {
    local ticker="$1"
    local config_file="$2"
    local output_dir="$3"
    
    echo "Processing $ticker..."
    
    local output_file="$output_dir/${ticker}_results.json"
    local log_file="$output_dir/${ticker}_process.log"
    
    if yfin scrape \
        ${config_file:+--config "$config_file"} \
        --ticker "$ticker" \
        --endpoint key-statistics \
        --fallback auto \
        --qps 1.0 \
        --timeout 45s \
        --preview > "$output_file" 2> "$log_file"; then
        echo "✓ $ticker completed successfully"
        return 0
    else
        echo "✗ $ticker failed"
        return 1
    fi
}

# Export function for parallel execution
export -f process_ticker

# Process tickers in parallel (max 3 at a time)
TICKERS=("$@")
CONFIG_FILE="${CONFIG_FILE:-}"
OUTPUT_DIR="${OUTPUT_DIR:-./parallel_output}"

mkdir -p "$OUTPUT_DIR"

echo "Processing ${#TICKERS[@]} tickers in parallel (max 3 concurrent)..."

printf '%s\n' "${TICKERS[@]}" | \
    xargs -n 1 -P 3 -I {} bash -c 'process_ticker "$@"' _ {} "$CONFIG_FILE" "$OUTPUT_DIR"

echo "Parallel processing completed"
EOF

chmod +x "$PARALLEL_SCRIPT"

# Run parallel processing
PARALLEL_OUTPUT="$OUTPUT_DIR/parallel_results"
mkdir -p "$PARALLEL_OUTPUT"

run_batch_command "Parallel Processing" \
    env CONFIG_FILE="$CONFIG_FILE" OUTPUT_DIR="$PARALLEL_OUTPUT" \
    "$PARALLEL_SCRIPT" "AAPL" "MSFT" "GOOGL" "AMZN" "TSLA" "META"

# Example 9: Batch Processing with Data Export
echo ""
echo "=== Example 9: Batch Processing with Data Export ==="

EXPORT_UNIVERSE="$OUTPUT_DIR/export_universe.txt"
create_universe_file "$EXPORT_UNIVERSE" "Export Universe" \
    "AAPL" "MSFT" "GOOGL"

EXPORT_DIR="$OUTPUT_DIR/exported_data"
mkdir -p "$EXPORT_DIR"

run_batch_command "Batch Processing with Export" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$EXPORT_UNIVERSE" \
    --endpoint key-statistics \
    --fallback auto \
    --qps 1.0 \
    --concurrency 2 \
    --out json \
    --out-dir "$EXPORT_DIR" \
    --preview

# List exported files
echo "Exported files:"
ls -la "$EXPORT_DIR"/ || echo "No files exported"

# Example 10: Monitoring and Health Checks During Batch Processing
echo ""
echo "=== Example 10: Monitoring During Batch Processing ==="

# Create monitoring script
MONITOR_SCRIPT="$OUTPUT_DIR/monitor_batch.sh"
cat > "$MONITOR_SCRIPT" << 'EOF'
#!/bin/bash

# monitor_batch.sh - Monitor batch processing health

echo "=== Batch Processing Health Monitor ==="
echo "Timestamp: $(date)"

# Check application health
echo ""
echo "Application Health:"
if curl -s -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "✓ Application health check passed"
    curl -s http://localhost:8080/health | jq . 2>/dev/null || echo "Health endpoint responded"
else
    echo "✗ Application health check failed"
fi

# Check recent metrics (if Prometheus is available)
echo ""
echo "Recent Metrics:"
if curl -s -f "http://localhost:9090/api/v1/query?query=rate(yfin_scrape_requests_total[5m])" > /dev/null 2>&1; then
    echo "✓ Metrics endpoint available"
    # Add specific metric queries here
else
    echo "ℹ Metrics endpoint not available (this is normal for local development)"
fi

# Check system resources
echo ""
echo "System Resources:"
echo "Memory usage: $(free -h | grep '^Mem:' | awk '{print $3 "/" $2}')"
echo "CPU load: $(uptime | awk -F'load average:' '{print $2}')"
echo "Disk usage: $(df -h . | tail -1 | awk '{print $5}')"

echo ""
echo "=== End Health Monitor ==="
EOF

chmod +x "$MONITOR_SCRIPT"

# Run monitoring during a batch job
MONITOR_UNIVERSE="$OUTPUT_DIR/monitor_universe.txt"
create_universe_file "$MONITOR_UNIVERSE" "Monitor Universe" \
    "AAPL" "MSFT" "GOOGL" "AMZN"

echo "Starting batch processing with monitoring..."

# Start monitoring in background
"$MONITOR_SCRIPT" > "$LOG_DIR/health_monitor.log" 2>&1 &
MONITOR_PID=$!

# Run batch processing
run_batch_command "Monitored Batch Processing" \
    yfin scrape \
    ${CONFIG_FILE:+--config "$CONFIG_FILE"} \
    --universe-file "$MONITOR_UNIVERSE" \
    --endpoint key-statistics \
    --fallback auto \
    --qps 1.0 \
    --concurrency 3 \
    --preview

# Stop monitoring
kill $MONITOR_PID 2>/dev/null || true
wait $MONITOR_PID 2>/dev/null || true

echo "Health monitoring log:"
cat "$LOG_DIR/health_monitor.log"

# Summary and Cleanup
echo ""
echo "=== Batch Processing Examples Summary ==="
echo "✓ Small universe processing"
echo "✓ Large universe with rate limiting"
echo "✓ Multi-endpoint processing"
echo "✓ International ticker handling"
echo "✓ Sector-based processing"
echo "✓ Error handling and recovery"
echo "✓ News batch processing"
echo "✓ Parallel processing"
echo "✓ Data export functionality"
echo "✓ Health monitoring during processing"
echo ""
echo "Generated Files:"
echo "- Universe files: $OUTPUT_DIR/*.txt"
echo "- Log files: $LOG_DIR/*.log"
echo "- Exported data: $EXPORT_DIR/"
echo "- Scripts: $OUTPUT_DIR/*.sh"
echo ""
echo "Best Practices Demonstrated:"
echo "1. Conservative rate limiting (0.5-1.0 QPS)"
echo "2. Proper error handling and logging"
echo "3. Timeout configuration for reliability"
echo "4. Parallel processing with concurrency limits"
echo "5. Health monitoring during batch jobs"
echo "6. Sector and geographic diversification"
echo ""
echo "Production Recommendations:"
echo "1. Use even lower QPS (0.2-0.5) for large batches"
echo "2. Implement retry logic for failed tickers"
echo "3. Add comprehensive monitoring and alerting"
echo "4. Use persistent storage for results"
echo "5. Schedule batch jobs during off-peak hours"
echo ""
echo "For more information:"
echo "- docs/scrape/cli.md - CLI documentation"
echo "- docs/observability.md - Monitoring guide"
echo "- runbooks/scrape-fallback.md - Operational procedures"

# Optional cleanup
read -p "Clean up generated files? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Cleaning up..."
    rm -rf "$OUTPUT_DIR" "$LOG_DIR"
    echo "Cleanup completed"
else
    echo "Files preserved in $OUTPUT_DIR and $LOG_DIR"
fi
