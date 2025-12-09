# Release Notes - yfinance-go v1.3.0

**Release Date:** December 8, 2025

## 🎉 Major Improvements

### Enhanced Analysis Scraper
- **Comprehensive Analysis Data Extraction**: The analysis scraper now extracts complete earnings trends, EPS revisions, and revenue estimates from Yahoo Finance analysis pages
- **EPS Trends**: Extracts current estimates and historical trends (7, 30, 60, 90 days ago) for current quarter, next quarter, current year, and next year
- **EPS Revisions**: Tracks up/down revisions in the last 7 and 30 days
- **Revenue Estimates**: Extracts quarterly and annual revenue estimates with proper scaling
- **Growth Estimates**: Includes growth rate estimates for current and future periods

### Improved Data Extraction
- **Robust HTML Parsing**: Updated regex patterns to handle Yahoo Finance's dynamic HTML structure
- **Better Error Handling**: Analysis extraction now continues even if some sections fail, ensuring maximum data extraction
- **MIC Inference**: Fixed hardcoded MIC values - now dynamically infers Market Identifier Codes for all exchanges
- **MIC Caching**: Added caching for MIC inference to improve performance and reduce redundant API calls

### Data Correctness
- **Fixed NYSE Exchange Mapping**: Added proper NYQ → XNYS mapping for NYSE-listed stocks
- **Accurate Price Scaling**: Corrected scaled decimal calculations for precise financial data representation
- **Comprehensive Testing**: Added extensive data correctness tests for prices, historical data, analyst data, and MIC inference

## 🔧 Technical Improvements

### Dependencies
- All dependencies updated to latest stable versions:
  - `ampy-proto/v2`: v2.1.1
  - `ampy-bus`: v1.2.0
  - `ampy-config`: v1.1.4
  - `ampy-observability`: v0.0.3

### Code Quality
- **Production-Ready**: Repository cleaned and verified for open-source use
- **Comprehensive Documentation**: Updated README with analysis scraper capabilities
- **Test Coverage**: All existing tests passing with new improvements

## 📊 New Features

### Analysis Data Structure
The `ScrapeAnalysis()` method now returns comprehensive analysis data including:

```go
// Earnings Estimates
- eps_estimate_current_quarter
- eps_estimate_next_quarter
- eps_estimate_current_year
- eps_estimate_next_year

// EPS Trends
- eps_trend_current_quarter
- eps_trend_current_quarter_7d_ago
- eps_trend_current_quarter_30d_ago
- eps_trend_current_year
- eps_trend_next_year

// EPS Revisions
- eps_revisions_up_7d_current_quarter
- eps_revisions_down_7d_current_quarter
- eps_revisions_up_30d_current_quarter
- eps_revisions_down_30d_current_quarter

// Revenue Estimates
- revenue_estimate_current_quarter
- revenue_estimate_current_year

// Growth Estimates
- growth_estimate_current_year
```

## 🐛 Bug Fixes

- **Fixed Analysis Scraper**: Previously returned empty data - now extracts all available analysis metrics
- **Fixed MIC Inference**: Removed hardcoded XNAS values, now correctly infers MIC for all exchanges
- **Fixed NYSE Mapping**: Added NYQ → XNYS mapping for proper NYSE stock identification
- **Fixed Price Precision**: Corrected scaled decimal calculations in data correctness tests

## 📝 Documentation Updates

- Updated README.md with comprehensive analysis scraper capabilities
- Added examples for `ScrapeAnalysis()` and `ScrapeAnalystInsights()` methods
- Documented all new analysis data fields and their meanings

## ✅ Testing

- All existing tests passing
- New data correctness tests added and passing
- MIC inference tests validate correct exchange mapping
- Price precision tests ensure accurate scaled decimal representation

## 🚀 Migration Guide

No breaking changes. All existing code continues to work. The analysis scraper now returns more comprehensive data than before.

### Before
```go
analysis, err := client.ScrapeAnalysis(ctx, "AAPL", runID)
// Returns empty or minimal data
```

### After
```go
analysis, err := client.ScrapeAnalysis(ctx, "AAPL", runID)
// Returns comprehensive analysis data with 16+ metrics including:
// - EPS estimates for all periods
// - EPS trends with historical data
// - EPS revisions tracking
// - Revenue estimates
// - Growth estimates
```

## 📦 Files Changed

### Core Changes
- `internal/scrape/analysis.go` - Enhanced analysis extraction with robust HTML parsing
- `internal/scrape/regex/analysis.yaml` - Updated regex patterns for current Yahoo Finance HTML structure
- `internal/emit/map_financials.go` - Enhanced MapAnalysisDTO to map all analysis data types
- `client.go` - Fixed MIC inference with caching and dynamic exchange detection
- `internal/norm/security.go` - Added NYQ → XNYS mapping

### Documentation
- `README.md` - Updated with analysis scraper improvements
- `RELEASE_NOTES.md` - This file

### Testing
- `tests/data_correctness_test.go` - Comprehensive data correctness validation

## 🙏 Acknowledgments

This release includes significant improvements to data extraction accuracy and completeness, making yfinance-go more reliable for production financial data pipelines.

---

**Full Changelog:** See [CHANGELOG.md](CHANGELOG.md) for complete version history.

