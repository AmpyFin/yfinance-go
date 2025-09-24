# Usage Guide

This guide provides comprehensive examples of how to use `yfinance-go` for fetching Yahoo Finance data.

## Overview

`yfinance-go` is a command-line tool for fetching financial data from Yahoo Finance, including:

- **Daily bars** (OHLCV data with adjustment policies)
- **Snapshot quotes** (real-time price data)
- **Fundamentals** (financial metrics, requires paid subscription)

## Basic Usage

### Check Version

```bash
yfin version
```

### Get Help

```bash
# General help
yfin --help

# Command-specific help
yfin pull --help
yfin quote --help
yfin fundamentals --help
yfin config --help
```

## Daily Bars (pull command)

### Basic Bar Fetching

```bash
# Fetch daily bars for a single symbol
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview

# Fetch with specific adjustment policy
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --adjusted raw --preview
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --adjusted split_dividend --preview
```

### Multiple Symbols

```bash
# Create a universe file
echo -e "AAPL\nMSFT\nGOOGL\nTSLA" > nasdaq100.txt

# Fetch bars for multiple symbols
yfin pull --universe-file nasdaq100.txt --start 2024-01-01 --end 2024-12-31 --preview
```

### International Markets

```bash
# German market (SAP)
yfin pull --ticker SAP --market XETR --start 2024-01-01 --end 2024-12-31 --preview

# Japanese market (Toyota)
yfin pull --ticker TM --market XTKS --start 2024-01-01 --end 2024-12-31 --preview
```

### FX Conversion Preview

```bash
# Convert to EUR
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --fx-target EUR --preview

# Convert to JPY
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --fx-target JPY --preview
```

### Local Export

```bash
# Export to JSON
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --out json --out-dir ./data --preview

# Export multiple symbols
yfin pull --universe-file nasdaq100.txt --start 2024-01-01 --end 2024-12-31 --out json --out-dir ./data --preview
```

### Bus Publishing

```bash
# Preview bus publishing
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --publish --env dev --topic-prefix ampy --preview

# Actually publish to bus
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --publish --env prod --topic-prefix ampy
```

### Performance Tuning

```bash
# Increase concurrency
yfin pull --universe-file nasdaq100.txt --start 2024-01-01 --end 2024-12-31 --concurrency 32 --preview

# Adjust QPS (queries per second)
yfin pull --universe-file nasdaq100.txt --start 2024-01-01 --end 2024-12-31 --qps 10 --preview

# Session rotation
yfin pull --universe-file nasdaq100.txt --start 2024-01-01 --end 2024-12-31 --sessions 5 --preview
```

## Snapshot Quotes (quote command)

### Single Quote

```bash
# Get quote for a single symbol
yfin quote --tickers AAPL --preview
```

### Multiple Quotes

```bash
# Get quotes for multiple symbols
yfin quote --tickers AAPL,MSFT,GOOGL,TSLA --preview
```

### Export Quotes

```bash
# Export quotes to JSON
yfin quote --tickers AAPL,MSFT,GOOGL --out json --out-dir ./quotes --preview
```

### Publish Quotes

```bash
# Preview quote publishing
yfin quote --tickers AAPL,MSFT --publish --env dev --topic-prefix ampy --preview

# Actually publish quotes
yfin quote --tickers AAPL,MSFT --publish --env prod --topic-prefix ampy
```

## Fundamentals (fundamentals command)

**Note**: Fundamentals require a Yahoo Finance paid subscription.

### Basic Fundamentals

```bash
# Fetch fundamentals for a symbol
yfin fundamentals --ticker AAPL --preview
```

### Error Handling

Fundamentals commands return exit code 2 for paid subscription errors:

```bash
yfin fundamentals --ticker AAPL --preview
echo $?  # Will be 2 if paid subscription required
```

## Configuration Management

### View Effective Configuration

```bash
# Print effective configuration
yfin config --print-effective

# Print as JSON
yfin config --print-effective --json
```

### Use Custom Configuration

```bash
# Use specific config file
yfin --config ./my-config.yaml pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview
```

## Advanced Usage

### Observability Control

```bash
# Disable tracing
yfin --observability-disable-tracing pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview

# Disable metrics
yfin --observability-disable-metrics pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview

# Disable both
yfin --observability-disable-tracing --observability-disable-metrics pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview
```

### Logging Control

```bash
# Debug logging
yfin --log-level debug pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview

# Warning level
yfin --log-level warn pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview
```

### Custom Run ID

```bash
# Specify custom run ID for tracking
yfin --run-id my-daily-job pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview
```

### HTTP Configuration

```bash
# Custom timeout
yfin --timeout 30s pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview

# Custom retry attempts
yfin --retry-max 5 pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview
```

## Output Examples

### Bar Preview Output

```
RUN yfin_1704067200  (env=dev, topic_prefix=ampy)
SYMBOL AAPL (MIC=XNAS, CCY=USD)  range=2024-01-01..2024-12-31  bars=252  adjusted=split_dividend
first=2024-01-01T00:00:00Z  last=2024-12-31T00:00:00Z  last_close=192.5300 USD
```

### Quote Preview Output

```
SYMBOL AAPL quote  price=192.5300 USD  high=195.0000  low=190.0000  venue=XNAS
```

### Fundamentals Preview Output

```
SYMBOL AAPL fundamentals  lines=45  source=yahoo-finance
  market_cap: 3000000000000.00 USD
  revenue: 383285000000.00 USD
  net_income: 99803000000.00 USD
  eps: 6.13 USD
  pe_ratio: 31.40
```

### Bus Preview Output

```
BUS PREVIEW (env=dev, topic_prefix=ampy)
  topic: ampy.dev.bars.v1.AAPL.XNAS
  payload_size: 2048 bytes
  estimated_messages: 1
  retry_config: attempts=5, base_ms=250, max_delay_ms=8000
  circuit_breaker: window=50, threshold=0.30, reset_timeout_ms=30000
```

## Error Handling

### Exit Codes

- `0` - Success
- `1` - General error
- `2` - Paid feature required (fundamentals)
- `3` - Configuration error
- `4` - Publishing error

### Common Error Scenarios

```bash
# Invalid date format
yfin pull --ticker AAPL --start 2024-13-01 --end 2024-12-31 --preview
# ERROR: Invalid date format: parsing time "2024-13-01": month out of range

# Missing required flags
yfin pull --ticker AAPL --start 2024-01-01 --preview
# ERROR: --start and --end are required

# Invalid adjustment policy
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --adjusted invalid --preview
# ERROR: --adjusted must be 'raw' or 'split_dividend'

# Paid subscription required
yfin fundamentals --ticker AAPL --preview
# ERROR: This endpoint requires Yahoo Finance paid subscription
```

## Best Practices

### Performance

1. **Use appropriate concurrency**: Start with default, increase for large universes
2. **Batch operations**: Use universe files for multiple symbols
3. **Monitor QPS**: Don't exceed Yahoo Finance rate limits
4. **Use session rotation**: For high-volume operations

### Data Quality

1. **Always preview first**: Use `--preview` flag before publishing
2. **Validate date ranges**: Ensure start/end dates are reasonable
3. **Check adjustment policies**: Understand raw vs split_dividend
4. **Verify symbols**: Ensure ticker symbols are correct

### Production Usage

1. **Use configuration files**: Don't rely on CLI flags for production
2. **Monitor observability**: Enable tracing and metrics
3. **Handle errors gracefully**: Check exit codes and error messages
4. **Use appropriate environments**: dev/staging/prod configurations

### Security

1. **Protect configuration**: Keep sensitive config files secure
2. **Use environment variables**: For sensitive configuration
3. **Monitor access**: Track who has access to production configs
4. **Regular updates**: Keep the tool updated for security patches

## Integration Examples

### Shell Scripts

```bash
#!/bin/bash
# Daily data fetch script

DATE=$(date -d "yesterday" +%Y-%m-%d)
OUTPUT_DIR="./data/$(date +%Y/%m)"

yfin pull \
  --universe-file ./symbols/nasdaq100.txt \
  --start "$DATE" \
  --end "$DATE" \
  --out json \
  --out-dir "$OUTPUT_DIR" \
  --concurrency 16 \
  --log-level info

if [ $? -eq 0 ]; then
  echo "Data fetch completed successfully"
else
  echo "Data fetch failed"
  exit 1
fi
```

### Cron Jobs

```bash
# Crontab entry for daily data fetch at 6 AM
0 6 * * * /usr/local/bin/yfin pull --universe-file /path/to/symbols.txt --start $(date -d "yesterday" +\%Y-\%m-\%d) --end $(date -d "yesterday" +\%Y-\%m-\%d) --out json --out-dir /data/$(date +\%Y/\%m) --config /etc/yfinance/config.yaml
```

### Docker Usage

```bash
# Run in Docker container
docker run --rm \
  -v $(pwd)/config:/config \
  -v $(pwd)/data:/data \
  ghcr.io/AmpyFin/yfinance-go:latest \
  yfin pull \
  --config /config/prod.yaml \
  --ticker AAPL \
  --start 2024-01-01 \
  --end 2024-12-31 \
  --out json \
  --out-dir /data
```

## Troubleshooting

### Common Issues

1. **Network timeouts**: Increase `--timeout` or check network connectivity
2. **Rate limiting**: Reduce `--qps` or increase `--sessions`
3. **Memory issues**: Reduce `--concurrency` for large universes
4. **Configuration errors**: Use `yfin config --print-effective` to debug

### Debug Mode

```bash
# Enable debug logging
yfin --log-level debug pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview

# Disable observability for debugging
yfin --observability-disable-tracing --observability-disable-metrics --log-level debug pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview
```

### Getting Help

```bash
# Check version and build info
yfin version

# Get detailed help
yfin --help
yfin pull --help

# Validate configuration
yfin config --print-effective
```

## Next Steps

- [Installation Guide](install.md) - How to install yfinance-go
- [Versioning Policy](versioning.md) - Understanding versions and compatibility
- [Configuration](https://github.com/AmpyFin/yfinance-go/tree/main/configs) - Configuration examples and options
