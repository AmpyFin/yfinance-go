package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/AmpyFin/yfinance-go"
	"github.com/AmpyFin/yfinance-go/internal/bus"
	"github.com/AmpyFin/yfinance-go/internal/config"
	"github.com/AmpyFin/yfinance-go/internal/emit"
	"github.com/AmpyFin/yfinance-go/internal/httpx"
	"github.com/AmpyFin/yfinance-go/internal/norm"
	"github.com/AmpyFin/yfinance-go/internal/obsv"
	"github.com/AmpyFin/yfinance-go/internal/scrape"
	fundamentalsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/fundamentals/v1"
	newsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/news/v1"
)

// Version information set via ldflags during build
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// Exit codes as specified in the requirements
const (
	ExitSuccess = 0
	ExitGeneral = 1
	ExitPaidFeature = 2
	ExitConfigError = 3
	ExitPublishError = 4
)

// Global configuration
type GlobalConfig struct {
	ConfigFile   string
	LogLevel     string
	RunID        string
	Concurrency  int
	QPS          float64
	RetryMax     int
	Sessions     int
	Timeout      time.Duration
}

// Pull command configuration
type PullConfig struct {
	Ticker        string
	UniverseFile  string
	Start         string
	End           string
	Adjusted      string
	Market        string
	FXTarget      string
	Preview       bool
	Publish       bool
	Env           string
	TopicPrefix   string
	Out           string
	OutDir        string
	DryRunPublish bool
}

// Quote command configuration
type QuoteConfig struct {
	Tickers     string
	Preview     bool
	Publish     bool
	Env         string
	TopicPrefix string
	Out         string
	OutDir      string
}

// Fundamentals command configuration
type FundamentalsConfig struct {
	Ticker  string
	Preview bool
}

// Scrape command configuration
type ScrapeConfig struct {
	Check        bool
	Ticker       string
	Endpoint     string
	Endpoints    string // Comma-separated list of endpoints for preview-json
	Preview      bool
	PreviewJSON  bool
	PreviewNews  bool   // Preview news articles without emitting proto
	PreviewProto bool   // Preview proto summaries without full output
	Force        bool
}

// ComprehensiveStatsConfig holds configuration for comprehensive statistics command
type ComprehensiveStatsConfig struct {
	Ticker  string
	Preview bool
}

// ComprehensiveProfileConfig holds configuration for comprehensive profile command
type ComprehensiveProfileConfig struct {
	Ticker  string
	Preview bool
}

// Config command configuration
type ConfigConfig struct {
	PrintEffective bool
	JSON           bool
}

var (
	globalConfig GlobalConfig
	pullConfig   PullConfig
	quoteConfig  QuoteConfig
	fundConfig   FundamentalsConfig
	scrapeConfig ScrapeConfig
	comprehensiveStatsConfig ComprehensiveStatsConfig
	comprehensiveProfileConfig ComprehensiveProfileConfig
	configConfig ConfigConfig
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "yfin",
	Short: "Yahoo Finance data fetcher and publisher",
	Long: `yfin is a command-line tool for fetching Yahoo Finance data including:
- Daily bars (adjusted/raw) - daily-only scope
- Snapshot quotes
- Fundamentals (requires paid subscription)

The tool supports FX conversion preview, bus publishing, and local export.`,
}

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Fetch daily bars for a symbol or universe",
	Long: `Fetch daily bars for a single symbol or multiple symbols from a universe file.
Only daily bars are supported by design.

Examples:
  yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --adjusted split_dividend --preview
  yfin pull --universe-file ./nasdaq100.txt --start 2024-01-01 --end 2024-12-31 --preview --concurrency 32
  yfin pull --ticker SAP --start 2024-01-01 --end 2024-12-31 --out json --out-dir ./out --preview`,
	RunE: runPull,
}

// quoteCmd represents the quote command
var quoteCmd = &cobra.Command{
	Use:   "quote",
	Short: "Fetch snapshot quotes",
	Long: `Fetch snapshot quotes for one or more symbols.

Examples:
  yfin quote --tickers AAPL,MSFT,TSLA --preview
  yfin quote --tickers AAPL --publish --env prod --topic-prefix ampy`,
	RunE: runQuote,
}

// fundamentalsCmd represents the fundamentals command
var fundamentalsCmd = &cobra.Command{
	Use:   "fundamentals",
	Short: "Fetch fundamentals (requires paid subscription)",
	Long: `Fetch fundamentals data for a symbol.
Note: This endpoint requires Yahoo Finance paid subscription.

Examples:
  yfin fundamentals --ticker AAPL --preview`,
	RunE: runFundamentals,
}

// scrapeCmd represents the scrape command
var scrapeCmd = &cobra.Command{
	Use:   "scrape",
	Short: "Web scraping operations",
	Long: `Web scraping operations for Yahoo Finance data.
This command provides access to scraping functionality when API endpoints are unavailable.

Examples:
  yfin scrape --check --ticker AAPL --endpoint profile --preview
  yfin scrape --check --ticker MSFT --endpoint key-statistics --preview
  yfin scrape --preview-json --ticker AAPL --endpoints key-statistics,financials,analysis,profile
  yfin scrape --preview-news --ticker AAPL
  yfin scrape --preview-proto --ticker AAPL --endpoints financials,analysis,profile,news`,
	RunE: runScrape,
}

// comprehensiveStatsCmd represents the comprehensive statistics command
var comprehensiveStatsCmd = &cobra.Command{
	Use:   "comprehensive-stats",
	Short: "Extract comprehensive key statistics with historical data",
	Long: `Extract comprehensive key statistics including current values and 5-year historical data.
This command uses YAML-configured regex patterns to extract all key statistics from Yahoo Finance.

Examples:
  yfin comprehensive-stats --ticker AAPL
  yfin comprehensive-stats --ticker MSFT --preview`,
	RunE: runComprehensiveStats,
}

// comprehensiveProfileCmd represents the comprehensive profile command
var comprehensiveProfileCmd = &cobra.Command{
	Use:   "comprehensive-profile",
	Short: "Extract comprehensive company profile information",
	Long: `Extract comprehensive company profile information including company details, 
key executives, and business summary from Yahoo Finance.

Examples:
  yfin comprehensive-profile --ticker AAPL
  yfin comprehensive-profile --ticker MSFT --preview`,
	RunE: runComprehensiveProfile,
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long: `Configuration management for yfinance-go.
Loads and validates configuration from ampy-config files.

Examples:
  yfin config --file ./configs/example.dev.yaml --print-effective
  yfin config --print-effective --json`,
	RunE: runConfig,
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print version information including build details.`,
	RunE:  runVersion,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&globalConfig.ConfigFile, "config", "", "ampy-config file (optional)")
	rootCmd.PersistentFlags().StringVar(&globalConfig.LogLevel, "log-level", "info", "Log level (info|debug|warn|error)")
	rootCmd.PersistentFlags().StringVar(&globalConfig.RunID, "run-id", "", "Run ID for tracking (if empty, autogenerated)")
	rootCmd.PersistentFlags().IntVar(&globalConfig.Concurrency, "concurrency", 0, "Worker pool size (default from config)")
	rootCmd.PersistentFlags().Float64Var(&globalConfig.QPS, "qps", 0, "Per-host QPS (default from config)")
	rootCmd.PersistentFlags().IntVar(&globalConfig.RetryMax, "retry-max", 0, "HTTP retry attempts")
	rootCmd.PersistentFlags().IntVar(&globalConfig.Sessions, "sessions", 0, "Session rotation pool size")
	rootCmd.PersistentFlags().DurationVar(&globalConfig.Timeout, "timeout", 0, "HTTP timeout (e.g., 6s)")
	
	// Observability flags
	rootCmd.PersistentFlags().Bool("observability-disable-tracing", false, "Disable OpenTelemetry tracing")
	rootCmd.PersistentFlags().Bool("observability-disable-metrics", false, "Disable Prometheus metrics")

	// Pull command flags
	pullCmd.Flags().StringVar(&pullConfig.Ticker, "ticker", "", "Stock symbol to fetch (e.g., AAPL)")
	pullCmd.Flags().StringVar(&pullConfig.UniverseFile, "universe-file", "", "Newline-delimited list of symbols")
	pullCmd.Flags().StringVar(&pullConfig.Start, "start", "", "Start date (YYYY-MM-DD, UTC)")
	pullCmd.Flags().StringVar(&pullConfig.End, "end", "", "End date (YYYY-MM-DD, UTC)")
	pullCmd.Flags().StringVar(&pullConfig.Adjusted, "adjusted", "split_dividend", "Adjustment policy (raw|split_dividend)")
	pullCmd.Flags().StringVar(&pullConfig.Market, "market", "", "Market MIC (optional hint for MIC inference)")
	pullCmd.Flags().StringVar(&pullConfig.FXTarget, "fx-target", "", "Target currency for FX conversion preview (e.g., USD)")
	pullCmd.Flags().BoolVar(&pullConfig.Preview, "preview", false, "Show preview without publishing")
	pullCmd.Flags().BoolVar(&pullConfig.Publish, "publish", false, "Enable bus publishing")
	pullCmd.Flags().StringVar(&pullConfig.Env, "env", "dev", "Environment (dev, staging, prod)")
	pullCmd.Flags().StringVar(&pullConfig.TopicPrefix, "topic-prefix", "ampy", "Topic prefix for bus publishing")
	pullCmd.Flags().StringVar(&pullConfig.Out, "out", "", "Output format (json|parquet)")
	pullCmd.Flags().StringVar(&pullConfig.OutDir, "out-dir", "", "Output directory")
	pullCmd.Flags().BoolVar(&pullConfig.DryRunPublish, "dry-run-publish", false, "Alias for --preview; no network send but compute payload sizes")

	// Quote command flags
	quoteCmd.Flags().StringVar(&quoteConfig.Tickers, "tickers", "", "Comma-separated list of symbols (e.g., AAPL,MSFT,TSLA)")
	quoteCmd.Flags().BoolVar(&quoteConfig.Preview, "preview", false, "Show preview without publishing")
	quoteCmd.Flags().BoolVar(&quoteConfig.Publish, "publish", false, "Enable bus publishing")
	quoteCmd.Flags().StringVar(&quoteConfig.Env, "env", "dev", "Environment (dev, staging, prod)")
	quoteCmd.Flags().StringVar(&quoteConfig.TopicPrefix, "topic-prefix", "ampy", "Topic prefix for bus publishing")
	quoteCmd.Flags().StringVar(&quoteConfig.Out, "out", "", "Output format (json)")
	quoteCmd.Flags().StringVar(&quoteConfig.OutDir, "out-dir", "", "Output directory")

	// Fundamentals command flags
	fundamentalsCmd.Flags().StringVar(&fundConfig.Ticker, "ticker", "", "Stock symbol to fetch (e.g., AAPL)")
	fundamentalsCmd.Flags().BoolVar(&fundConfig.Preview, "preview", false, "Show preview")

	// Scrape command flags
	scrapeCmd.Flags().BoolVar(&scrapeConfig.Check, "check", false, "Check scraping connectivity (no parsing)")
	scrapeCmd.Flags().StringVar(&scrapeConfig.Ticker, "ticker", "", "Stock symbol to scrape (e.g., AAPL)")
	scrapeCmd.Flags().StringVar(&scrapeConfig.Endpoint, "endpoint", "", "Endpoint to scrape (profile, key-statistics, financials, balance-sheet, cash-flow, analysis, analyst-insights, news)")
	scrapeCmd.Flags().StringVar(&scrapeConfig.Endpoints, "endpoints", "", "Comma-separated list of endpoints for preview-json (e.g., key-statistics,financials,analysis,profile)")
	scrapeCmd.Flags().BoolVar(&scrapeConfig.Preview, "preview", false, "Show preview without parsing")
	scrapeCmd.Flags().BoolVar(&scrapeConfig.PreviewJSON, "preview-json", false, "Preview JSON extraction without emitting proto")
	scrapeCmd.Flags().BoolVar(&scrapeConfig.PreviewNews, "preview-news", false, "Preview news articles without emitting proto")
	scrapeCmd.Flags().BoolVar(&scrapeConfig.PreviewProto, "preview-proto", false, "Preview proto summaries with counts, periods, and metadata")
	scrapeCmd.Flags().BoolVar(&scrapeConfig.Force, "force", false, "Force scraping even if API is available")

	// Comprehensive stats command flags
	comprehensiveStatsCmd.Flags().StringVar(&comprehensiveStatsConfig.Ticker, "ticker", "", "Stock symbol to analyze (e.g., AAPL)")
	comprehensiveStatsCmd.Flags().BoolVar(&comprehensiveStatsConfig.Preview, "preview", false, "Show preview of extracted data")

	// Comprehensive profile command flags
	comprehensiveProfileCmd.Flags().StringVar(&comprehensiveProfileConfig.Ticker, "ticker", "", "Stock symbol to analyze (e.g., AAPL)")
	comprehensiveProfileCmd.Flags().BoolVar(&comprehensiveProfileConfig.Preview, "preview", false, "Show preview of extracted data")

	// Config command flags
	configCmd.Flags().BoolVar(&configConfig.PrintEffective, "print-effective", false, "Print effective configuration")
	configCmd.Flags().BoolVar(&configConfig.JSON, "json", false, "Output in JSON format")

	// Add subcommands
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(quoteCmd)
	rootCmd.AddCommand(fundamentalsCmd)
	rootCmd.AddCommand(scrapeCmd)
	rootCmd.AddCommand(comprehensiveStatsCmd)
	rootCmd.AddCommand(comprehensiveProfileCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(ExitGeneral)
	}
}

// runPull executes the pull command
func runPull(cmd *cobra.Command, args []string) error {
	// Validate flags
	if err := validatePullFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Generate run ID if not provided
	runID := globalConfig.RunID
	if runID == "" {
		runID = fmt.Sprintf("yfin_%d", time.Now().Unix())
	}

	// Parse dates
	startTime, endTime, err := parseDates(pullConfig.Start, pullConfig.End)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Invalid date format: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Parse adjustment policy
	adjusted, err := parseAdjusted(pullConfig.Adjusted)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Invalid adjusted value: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Validate interval (daily-only enforcement)
	loader := config.NewLoader(globalConfig.ConfigFile)
	cfg, err := loader.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to load configuration: %v\n", err)
		os.Exit(ExitConfigError)
	}
	
	// For yfinance-go, we only support daily intervals
	if err := cfg.ValidateInterval("1d"); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Initialize observability
	ctx := context.Background()
	disableTracing, _ := cmd.Flags().GetBool("observability-disable-tracing")
	disableMetrics, _ := cmd.Flags().GetBool("observability-disable-metrics")
	
	obsvConfig := &obsv.Config{
		ServiceName:       "yfinance-go",
		ServiceVersion:    version,
		Environment:       cfg.App.Env,
		CollectorEndpoint: cfg.Observability.Tracing.OTLP.Endpoint,
		TraceProtocol:     "grpc",
		SampleRatio:       cfg.Observability.Tracing.OTLP.SampleRatio,
		LogLevel:          cfg.Observability.Logs.Level,
		MetricsAddr:       cfg.Observability.Metrics.Prometheus.Addr,
		MetricsEnabled:    cfg.Observability.Metrics.Prometheus.Enabled && !disableMetrics,
		TracingEnabled:    cfg.Observability.Tracing.OTLP.Enabled && !disableTracing,
	}
	
	if err := obsv.Init(ctx, obsvConfig); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to initialize observability: %v\n", err)
		os.Exit(ExitConfigError)
	}
	defer func() { _ = obsv.Shutdown(ctx) }()

	// Get symbols to process
	symbols, err := getSymbols(pullConfig.Ticker, pullConfig.UniverseFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to get symbols: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Create client
	client, err := createClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create client: %v\n", err)
		os.Exit(ExitGeneral)
	}

	// Create bus if publishing or previewing
	var busInstance *bus.Bus
	var busConfig *bus.Config
	if pullConfig.Publish || pullConfig.Preview || pullConfig.DryRunPublish {
		busConfig = createBusConfig(pullConfig.Env, pullConfig.TopicPrefix)
		busInstance, err = bus.NewBus(busConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to create bus: %v\n", err)
			os.Exit(ExitGeneral)
		}
		defer busInstance.Close(context.Background())
	}

	// Process symbols
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	successCount := 0
	for _, symbol := range symbols {
		if err := processSymbol(ctx, client, symbol, startTime, endTime, adjusted, runID, busInstance, busConfig); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to process %s: %v\n", symbol, err)
			continue
		}
		successCount++
	}

	if successCount == 0 {
		fmt.Fprintf(os.Stderr, "ERROR: No symbols processed successfully\n")
		os.Exit(ExitGeneral)
	}

	fmt.Printf("Successfully processed %d/%d symbols\n", successCount, len(symbols))
	return nil
}

// runQuote executes the quote command
func runQuote(cmd *cobra.Command, args []string) error {
	// Validate flags
	if err := validateQuoteFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Generate run ID if not provided
	runID := globalConfig.RunID
	if runID == "" {
		runID = fmt.Sprintf("yfin_%d", time.Now().Unix())
	}

	// Parse tickers
	tickers := strings.Split(quoteConfig.Tickers, ",")
	for i, ticker := range tickers {
		tickers[i] = strings.TrimSpace(ticker)
	}

	// Create client
	client, err := createClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create client: %v\n", err)
		os.Exit(ExitGeneral)
	}

	// Create bus if publishing
	var busInstance *bus.Bus
	var busConfig *bus.Config
	if quoteConfig.Publish || quoteConfig.Preview {
		busConfig = createBusConfig(quoteConfig.Env, quoteConfig.TopicPrefix)
		busInstance, err = bus.NewBus(busConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to create bus: %v\n", err)
			os.Exit(ExitGeneral)
		}
		defer busInstance.Close(context.Background())
	}

	// Process quotes
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	successCount := 0
	for _, ticker := range tickers {
		if err := processQuote(ctx, client, ticker, runID, busInstance, busConfig); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to process quote for %s: %v\n", ticker, err)
			continue
		}
		successCount++
	}

	if successCount == 0 {
		fmt.Fprintf(os.Stderr, "ERROR: No quotes processed successfully\n")
		os.Exit(ExitGeneral)
	}

	fmt.Printf("Successfully processed %d/%d quotes\n", successCount, len(tickers))
	return nil
}

// runFundamentals executes the fundamentals command
func runFundamentals(cmd *cobra.Command, args []string) error {
	// Validate flags
	if err := validateFundamentalsFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Generate run ID if not provided
	runID := globalConfig.RunID
	if runID == "" {
		runID = fmt.Sprintf("yfin_%d", time.Now().Unix())
	}

	// Create client
	client, err := createClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create client: %v\n", err)
		os.Exit(ExitGeneral)
	}

	// Process fundamentals
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := processFundamentals(ctx, client, fundConfig.Ticker, runID); err != nil {
		// Check if it's a paid feature error
		if isPaidFeatureError(err) {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(ExitPaidFeature)
		}
		fmt.Fprintf(os.Stderr, "ERROR: Failed to process fundamentals for %s: %v\n", fundConfig.Ticker, err)
		os.Exit(ExitGeneral)
	}

	return nil
}

// runScrape executes the scrape command
func runScrape(cmd *cobra.Command, args []string) error {
	// Validate flags
	if err := validateScrapeFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Generate run ID if not provided
	runID := globalConfig.RunID
	if runID == "" {
		runID = fmt.Sprintf("yfin_scrape_%d", time.Now().Unix())
	}

	// Load configuration
	loader := config.NewLoader(globalConfig.ConfigFile)
	cfg, err := loader.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to load configuration: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Get scrape configuration
	scrapeCfg := cfg.GetScrapeConfig()
	if !scrapeCfg.Enabled {
		fmt.Fprintf(os.Stderr, "ERROR: Scraping is disabled in configuration\n")
		os.Exit(ExitConfigError)
	}

	// Initialize observability
	ctx := context.Background()
	disableTracing, _ := cmd.Flags().GetBool("observability-disable-tracing")
	disableMetrics, _ := cmd.Flags().GetBool("observability-disable-metrics")
	
	obsvConfig := &obsv.Config{
		ServiceName:       "yfinance-go",
		ServiceVersion:    version,
		Environment:       cfg.App.Env,
		CollectorEndpoint: cfg.Observability.Tracing.OTLP.Endpoint,
		TraceProtocol:     "grpc",
		SampleRatio:       cfg.Observability.Tracing.OTLP.SampleRatio,
		LogLevel:          cfg.Observability.Logs.Level,
		MetricsAddr:       cfg.Observability.Metrics.Prometheus.Addr,
		MetricsEnabled:    cfg.Observability.Metrics.Prometheus.Enabled && !disableMetrics,
		TracingEnabled:    cfg.Observability.Tracing.OTLP.Enabled && !disableTracing,
	}
	
	if err := obsv.Init(ctx, obsvConfig); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to initialize observability: %v\n", err)
		os.Exit(ExitConfigError)
	}
	defer func() { _ = obsv.Shutdown(ctx) }()

	// Create scrape client
	scrapeClient, err := createScrapeClient(scrapeCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create scrape client: %v\n", err)
		os.Exit(ExitGeneral)
	}

	// Execute scrape check
	if scrapeConfig.Check {
		return runScrapeCheck(ctx, scrapeClient, scrapeConfig.Ticker, scrapeConfig.Endpoint, runID)
	}

	// Execute preview-json mode
	if scrapeConfig.PreviewJSON {
		return runScrapePreviewJSON(ctx, scrapeClient, scrapeConfig.Ticker, scrapeConfig.Endpoints, runID)
	}

	// Execute preview-news mode
	if scrapeConfig.PreviewNews {
		return runScrapePreviewNews(ctx, scrapeClient, scrapeConfig.Ticker, runID)
	}

	// Execute preview-proto mode
	if scrapeConfig.PreviewProto {
		return runScrapePreviewProto(ctx, scrapeClient, scrapeConfig.Ticker, scrapeConfig.Endpoints, runID)
	}

	fmt.Fprintf(os.Stderr, "ERROR: Either --check, --preview-json, --preview-news, or --preview-proto mode is required\n")
	os.Exit(ExitGeneral)
	return nil
}

// runComprehensiveStats executes the comprehensive statistics command
func runComprehensiveStats(cmd *cobra.Command, args []string) error {
	// Validate flags
	if comprehensiveStatsConfig.Ticker == "" {
		return fmt.Errorf("--ticker is required")
	}

	// Generate run ID if not provided
	runID := globalConfig.RunID
	if runID == "" {
		runID = fmt.Sprintf("yfin_comprehensive_stats_%d", time.Now().Unix())
	}

	// Load configuration
	loader := config.NewLoader(globalConfig.ConfigFile)
	cfg, err := loader.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to load configuration: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Get scrape configuration
	scrapeCfg := cfg.GetScrapeConfig()
	if !scrapeCfg.Enabled {
		fmt.Fprintf(os.Stderr, "ERROR: Scraping is disabled in configuration\n")
		os.Exit(ExitConfigError)
	}

	// Initialize observability
	ctx := context.Background()
	disableTracing, _ := cmd.Flags().GetBool("observability-disable-tracing")
	disableMetrics, _ := cmd.Flags().GetBool("observability-disable-metrics")
	
	obsvConfig := &obsv.Config{
		ServiceName:       "yfinance-go",
		ServiceVersion:    version,
		Environment:       cfg.App.Env,
		CollectorEndpoint: cfg.Observability.Tracing.OTLP.Endpoint,
		TraceProtocol:     "grpc",
		SampleRatio:       cfg.Observability.Tracing.OTLP.SampleRatio,
		LogLevel:          cfg.Observability.Logs.Level,
		MetricsAddr:       cfg.Observability.Metrics.Prometheus.Addr,
		MetricsEnabled:    cfg.Observability.Metrics.Prometheus.Enabled && !disableMetrics,
		TracingEnabled:    cfg.Observability.Tracing.OTLP.Enabled && !disableTracing,
	}
	
	if err := obsv.Init(ctx, obsvConfig); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to initialize observability: %v\n", err)
		os.Exit(ExitConfigError)
	}
	defer func() { _ = obsv.Shutdown(ctx) }()

	// Create scrape client
	scrapeClient, err := createScrapeClient(scrapeCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create scrape client: %v\n", err)
		os.Exit(ExitGeneral)
	}

	// Execute comprehensive statistics extraction
	return runComprehensiveStatsExtraction(ctx, scrapeClient, comprehensiveStatsConfig.Ticker, runID)
}

// runConfig executes the config command
func runConfig(cmd *cobra.Command, args []string) error {
	if !configConfig.PrintEffective {
		return fmt.Errorf("--print-effective flag is required")
	}

	// Determine effective config path
	effectivePath := globalConfig.ConfigFile
	if effectivePath == "" {
		// Default to a standard effective config path
		effectivePath = "configs/effective.yaml"
	}

	// Load configuration using ampy-config
	loader := config.NewLoader(effectivePath)
	_, err := loader.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to load configuration: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Get effective configuration
	effectiveConfig, err := loader.GetEffectiveConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to get effective configuration: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Print configuration
	if configConfig.JSON {
		// Print as JSON
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(effectiveConfig); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to encode configuration as JSON: %v\n", err)
			os.Exit(ExitConfigError)
		}
	} else {
		// Print as key=value pairs
		printEffectiveConfig(effectiveConfig)
	}

	return nil
}

// printEffectiveConfig prints the effective configuration in key=value format
func printEffectiveConfig(configMap map[string]interface{}) {
	fmt.Println("EFFECTIVE CONFIG (redacted)")
	
	// Flatten the configuration map
	flattened := flattenConfigMap(configMap, "")
	
	// Sort keys for consistent output
	var keys []string
	for key := range flattened {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	
	// Print sorted key-value pairs
	for _, key := range keys {
		value := flattened[key]
		fmt.Printf("%s=%v\n", key, value)
	}
}

// flattenConfigMap flattens a nested configuration map into dot-notation keys
func flattenConfigMap(configMap map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})
	
	for key, value := range configMap {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		
		if nestedMap, ok := value.(map[string]interface{}); ok {
			// Recursively flatten nested maps
			nested := flattenConfigMap(nestedMap, fullKey)
			for k, v := range nested {
				result[k] = v
			}
		} else if slice, ok := value.([]interface{}); ok {
			// Handle slices (like allowed_intervals)
			var strSlice []string
			for _, item := range slice {
				if str, ok := item.(string); ok {
					strSlice = append(strSlice, str)
				}
			}
			if len(strSlice) > 0 {
				result[fullKey] = fmt.Sprintf("[%s]", strings.Join(strSlice, ","))
			}
		} else {
			result[fullKey] = value
		}
	}
	
	return result
}

// validatePullFlags validates pull command flags
func validatePullFlags() error {
	if pullConfig.Ticker == "" && pullConfig.UniverseFile == "" {
		return fmt.Errorf("either --ticker or --universe-file must be specified")
	}
	if pullConfig.Ticker != "" && pullConfig.UniverseFile != "" {
		return fmt.Errorf("cannot specify both --ticker and --universe-file")
	}
	if pullConfig.Start == "" || pullConfig.End == "" {
		return fmt.Errorf("--start and --end are required")
	}
	if pullConfig.Adjusted != "raw" && pullConfig.Adjusted != "split_dividend" {
		return fmt.Errorf("--adjusted must be 'raw' or 'split_dividend'")
	}
	if pullConfig.Out != "" && pullConfig.Out != "json" && pullConfig.Out != "parquet" {
		return fmt.Errorf("--out must be 'json' or 'parquet'")
	}
	return nil
}

// validateQuoteFlags validates quote command flags
func validateQuoteFlags() error {
	if quoteConfig.Tickers == "" {
		return fmt.Errorf("--tickers is required")
	}
	if quoteConfig.Out != "" && quoteConfig.Out != "json" {
		return fmt.Errorf("--out must be 'json' for quotes")
	}
	return nil
}

// validateFundamentalsFlags validates fundamentals command flags
func validateFundamentalsFlags() error {
	if fundConfig.Ticker == "" {
		return fmt.Errorf("--ticker is required")
	}
	return nil
}

// validateScrapeFlags validates scrape command flags
func validateScrapeFlags() error {
	// Check that either --check, --preview-json, --preview-news, or --preview-proto is specified
	if !scrapeConfig.Check && !scrapeConfig.PreviewJSON && !scrapeConfig.PreviewNews && !scrapeConfig.PreviewProto {
		return fmt.Errorf("either --check, --preview-json, --preview-news, or --preview-proto flag is required")
	}
	
	// All modes require ticker
	if scrapeConfig.Ticker == "" {
		return fmt.Errorf("--ticker is required")
	}
	
	// Check mode requires endpoint
	if scrapeConfig.Check {
		if scrapeConfig.Endpoint == "" {
			return fmt.Errorf("--endpoint is required for --check mode")
		}
		
		// Validate endpoint
		validEndpoints := []string{"profile", "key-statistics", "financials", "balance-sheet", "cash-flow", "analysis", "analyst-insights", "news"}
		valid := false
		for _, ep := range validEndpoints {
			if scrapeConfig.Endpoint == ep {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("--endpoint must be one of: %v", validEndpoints)
		}
	}
	
	// Preview-json mode requires endpoints
	if scrapeConfig.PreviewJSON {
		if scrapeConfig.Endpoints == "" {
			return fmt.Errorf("--endpoints is required for --preview-json mode")
		}
		
		// Validate endpoints
		endpointList := strings.Split(scrapeConfig.Endpoints, ",")
		validEndpoints := []string{"profile", "key-statistics", "financials", "balance-sheet", "cash-flow", "analysis", "analyst-insights", "news"}
		for _, ep := range endpointList {
			ep = strings.TrimSpace(ep)
			if ep == "" {
				continue
			}
			valid := false
			for _, validEp := range validEndpoints {
				if ep == validEp {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid endpoint '%s' in --endpoints", ep)
			}
		}
	}
	
	// Preview-proto mode requires endpoints
	if scrapeConfig.PreviewProto {
		if scrapeConfig.Endpoints == "" {
			return fmt.Errorf("--endpoints is required for --preview-proto mode")
		}
		
		// Validate endpoints
		endpointList := strings.Split(scrapeConfig.Endpoints, ",")
		validEndpoints := []string{"profile", "key-statistics", "financials", "balance-sheet", "cash-flow", "analysis", "analyst-insights", "news"}
		for _, ep := range endpointList {
			ep = strings.TrimSpace(ep)
			if ep == "" {
				continue
			}
			valid := false
			for _, validEp := range validEndpoints {
				if ep == validEp {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid endpoint '%s' in --endpoints", ep)
			}
		}
	}
	
	return nil
}

// parseDates parses start and end date strings
func parseDates(startStr, endStr string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date: %v", err)
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date: %v", err)
	}
	return start, end, nil
}

// parseAdjusted parses the adjusted flag
func parseAdjusted(adjusted string) (bool, error) {
	switch adjusted {
	case "raw":
		return false, nil
	case "split_dividend":
		return true, nil
	default:
		return false, fmt.Errorf("invalid adjusted value: %s", adjusted)
	}
}

// getSymbols returns the list of symbols to process
func getSymbols(ticker, universeFile string) ([]string, error) {
	if ticker != "" {
		return []string{ticker}, nil
	}
	
	// Read universe file
	content, err := os.ReadFile(universeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read universe file: %v", err)
	}
	
	lines := strings.Split(string(content), "\n")
	var symbols []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			symbols = append(symbols, line)
		}
	}
	
	if len(symbols) == 0 {
		return nil, fmt.Errorf("no symbols found in universe file")
	}
	
	return symbols, nil
}

// createClient creates a yfinance client with configuration
func createClient() (*yfinance.Client, error) {
	// Determine effective config path
	effectivePath := globalConfig.ConfigFile
	if effectivePath == "" {
		// Default to a standard effective config path
		effectivePath = "configs/effective.yaml"
	}
	
	// Load configuration using ampy-config
	loader := config.NewLoader(effectivePath)
	cfg, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	
	// Convert to HTTP config
	httpConfig := cfg.GetHTTPConfig()
	
	// Apply global flags if set (CLI flags override config)
	if globalConfig.QPS > 0 {
		httpConfig.QPS = globalConfig.QPS
	}
	if globalConfig.RetryMax > 0 {
		httpConfig.MaxAttempts = globalConfig.RetryMax
	}
	if globalConfig.Sessions > 0 {
		httpConfig.EnableSessionRotation = true
		httpConfig.NumSessions = globalConfig.Sessions
	}
	if globalConfig.Timeout > 0 {
		httpConfig.Timeout = globalConfig.Timeout
	}
	
	// Create httpx config from our config
	httpxConfig := &httpx.Config{
		BaseURL:            httpConfig.BaseURL,
		Timeout:            httpConfig.Timeout,
		IdleTimeout:        httpConfig.IdleTimeout,
		MaxConnsPerHost:    httpConfig.MaxConnsPerHost,
		UserAgent:          httpConfig.UserAgent,
		MaxAttempts:        httpConfig.MaxAttempts,
		BackoffBaseMs:      httpConfig.BackoffBaseMs,
		BackoffJitterMs:    httpConfig.BackoffJitterMs,
		MaxDelayMs:         httpConfig.MaxDelayMs,
		QPS:                httpConfig.QPS,
		Burst:              httpConfig.Burst,
		CircuitWindow:      httpConfig.CircuitWindow,
		FailureThreshold:   int(httpConfig.FailureThreshold * 100), // Convert to percentage
		ResetTimeout:       httpConfig.ResetTimeout,
		EnableSessionRotation: httpConfig.EnableSessionRotation,
		NumSessions:        httpConfig.NumSessions,
	}
	
	// Create client
	if httpConfig.EnableSessionRotation {
		return yfinance.NewClientWithSessionRotation(), nil
	}
	return yfinance.NewClientWithConfig(httpxConfig), nil
}

// createBusConfig creates bus configuration
func createBusConfig(env, topicPrefix string) *bus.Config {
	// Determine effective config path
	effectivePath := globalConfig.ConfigFile
	if effectivePath == "" {
		// Default to a standard effective config path
		effectivePath = "configs/effective.yaml"
	}
	
	// Load configuration using ampy-config
	loader := config.NewLoader(effectivePath)
	cfg, err := loader.Load()
	if err != nil {
		// Fallback to default config if loading fails
		return &bus.Config{
			Enabled:         true,
			Env:             env,
			TopicPrefix:     topicPrefix,
			MaxPayloadBytes: 1024 * 1024, // 1 MiB
			Publisher: bus.PublisherConfig{
				Backend: "nats",
				NATS: bus.NATSConfig{
					URL:          "nats://localhost:4222",
					SubjectStyle: "topic",
					AckWaitMs:    5000,
				},
			},
			Retry: bus.RetryConfig{
				Attempts:   5,
				BaseMs:     250,
				MaxDelayMs: 8000,
			},
			CircuitBreaker: bus.CircuitBreakerConfig{
				Window:           50,
				FailureThreshold: 0.30,
				ResetTimeoutMs:   30000,
				HalfOpenProbes:   3,
			},
		}
	}
	
	// Get bus config from loaded configuration
	busConfig := cfg.GetBusConfig()
	
	// Override with CLI parameters
	busConfig.Enabled = true
	busConfig.Env = env
	busConfig.TopicPrefix = topicPrefix
	
	// Convert to bus.Config
	return &bus.Config{
		Enabled:         busConfig.Enabled,
		Env:             busConfig.Env,
		TopicPrefix:     busConfig.TopicPrefix,
		MaxPayloadBytes: busConfig.MaxPayloadBytes,
		Publisher: bus.PublisherConfig{
			Backend: busConfig.Publisher.Backend,
			NATS: bus.NATSConfig{
				URL:          busConfig.Publisher.NATS.URL,
				SubjectStyle: busConfig.Publisher.NATS.SubjectStyle,
				AckWaitMs:    busConfig.Publisher.NATS.AckWaitMs,
			},
			Kafka: bus.KafkaConfig{
				Brokers:     busConfig.Publisher.Kafka.Brokers,
				Acks:        busConfig.Publisher.Kafka.Acks,
				Compression: busConfig.Publisher.Kafka.Compression,
			},
		},
		Retry: bus.RetryConfig{
			Attempts:   busConfig.Retry.Attempts,
			BaseMs:     busConfig.Retry.BaseMs,
			MaxDelayMs: busConfig.Retry.MaxDelayMs,
		},
		CircuitBreaker: bus.CircuitBreakerConfig{
			Window:           busConfig.CircuitBreaker.Window,
			FailureThreshold: busConfig.CircuitBreaker.FailureThreshold,
			ResetTimeoutMs:   busConfig.CircuitBreaker.ResetTimeoutMs,
			HalfOpenProbes:   busConfig.CircuitBreaker.HalfOpenProbes,
		},
	}
}

// processSymbol processes a single symbol for bars
func processSymbol(ctx context.Context, client *yfinance.Client, symbol string, start, end time.Time, adjusted bool, runID string, busInstance *bus.Bus, busConfig *bus.Config) error {
	// Fetch bars
	bars, err := client.FetchDailyBars(ctx, symbol, start, end, adjusted, runID)
	if err != nil {
		return err
	}
	
	if len(bars.Bars) == 0 {
		fmt.Printf("No bars found for %s in the specified period\n", symbol)
		return nil
	}
	
	// Print preview
	printBarsPreview(bars, runID, pullConfig.Env, pullConfig.TopicPrefix)
	
	// Handle FX preview if requested
	if pullConfig.FXTarget != "" {
		if err := handleFXPreview(ctx, client, bars, pullConfig.FXTarget); err != nil {
			fmt.Printf("FX preview failed: %v\n", err)
		}
	}
	
	// Handle bus publishing
	if busInstance != nil {
		preview := pullConfig.Preview || pullConfig.DryRunPublish
		if err := handleBusPublishing(ctx, bars, busInstance, busConfig, runID, preview); err != nil {
			return fmt.Errorf("bus publishing failed: %v", err)
		}
	}
	
	// Handle local export
	if pullConfig.Out != "" && pullConfig.OutDir != "" {
		if err := handleLocalExport(bars, symbol, start, end, adjusted, pullConfig.Out, pullConfig.OutDir); err != nil {
			return fmt.Errorf("local export failed: %v", err)
		}
	}
	
	return nil
}

// processQuote processes a single quote
func processQuote(ctx context.Context, client *yfinance.Client, ticker string, runID string, busInstance *bus.Bus, busConfig *bus.Config) error {
	// Fetch quote
	quote, err := client.FetchQuote(ctx, ticker, runID)
	if err != nil {
		return err
	}
	
	// Print preview
	printQuotePreview(quote)
	
	// Handle bus publishing
	if busInstance != nil {
		if err := handleQuoteBusPublishing(ctx, quote, busInstance, busConfig, runID, quoteConfig.Preview); err != nil {
			return fmt.Errorf("bus publishing failed: %v", err)
		}
	}
	
	// Handle local export
	if quoteConfig.Out != "" && quoteConfig.OutDir != "" {
		if err := handleQuoteLocalExport(quote, ticker, quoteConfig.Out, quoteConfig.OutDir); err != nil {
			return fmt.Errorf("local export failed: %v", err)
		}
	}
	
	return nil
}

// processFundamentals processes fundamentals
func processFundamentals(ctx context.Context, client *yfinance.Client, ticker string, runID string) error {
	// Fetch fundamentals
	fundamentals, err := client.FetchFundamentalsQuarterly(ctx, ticker, runID)
	if err != nil {
		return err
	}
	
	// Print preview
	printFundamentalsPreview(fundamentals)
	
	return nil
}

// printBarsPreview prints the bars preview according to specification
func printBarsPreview(bars *norm.NormalizedBarBatch, runID, env, topicPrefix string) {
	firstBar := bars.Bars[0]
	lastBar := bars.Bars[len(bars.Bars)-1]
	
	fmt.Printf("RUN %s  (env=%s, topic_prefix=%s)\n", runID, env, topicPrefix)
	fmt.Printf("SYMBOL %s (MIC=%s, CCY=%s)  range=%s..%s  bars=%d  adjusted=%s\n",
		bars.Security.Symbol,
		bars.Security.MIC,
		firstBar.CurrencyCode,
		firstBar.Start.Format("2006-01-02"),
		lastBar.End.Format("2006-01-02"),
		len(bars.Bars),
		firstBar.AdjustmentPolicyID)
	fmt.Printf("first=%s  last=%s  last_close=%.4f %s\n",
		firstBar.Start.Format("2006-01-02T15:04:05Z"),
		lastBar.End.Format("2006-01-02T15:04:05Z"),
		float64(lastBar.Close.Scaled)/float64(lastBar.Close.Scale),
		lastBar.CurrencyCode)
}

// printQuotePreview prints the quote preview according to specification
func printQuotePreview(quote *norm.NormalizedQuote) {
	price := "N/A"
	if quote.RegularMarketPrice != nil {
		price = fmt.Sprintf("%.4f", norm.FromScaledDecimal(*quote.RegularMarketPrice))
	}
	
	high := "N/A"
	if quote.RegularMarketHigh != nil {
		high = fmt.Sprintf("%.4f", norm.FromScaledDecimal(*quote.RegularMarketHigh))
	}
	
	low := "N/A"
	if quote.RegularMarketLow != nil {
		low = fmt.Sprintf("%.4f", norm.FromScaledDecimal(*quote.RegularMarketLow))
	}
	
	fmt.Printf("SYMBOL %s quote  price=%s %s  high=%s  low=%s  venue=%s\n",
		quote.Security.Symbol, price, quote.CurrencyCode, high, low, quote.Venue)
}

// printFundamentalsPreview prints the fundamentals preview
func printFundamentalsPreview(fundamentals *norm.NormalizedFundamentalsSnapshot) {
	fmt.Printf("SYMBOL %s fundamentals  lines=%d  source=%s\n",
		fundamentals.Security.Symbol, len(fundamentals.Lines), fundamentals.Source)
	
	// Show first few lines
	for i, line := range fundamentals.Lines {
		if i >= 5 {
			break
		}
		fmt.Printf("  %s: %.2f %s\n", line.Key, float64(line.Value.Scaled)/float64(line.Value.Scale), line.CurrencyCode)
	}
}

// handleFXPreview handles FX conversion preview
func handleFXPreview(ctx context.Context, client *yfinance.Client, bars *norm.NormalizedBarBatch, targetCurrency string) error {
	// Check if FX conversion is needed
	firstBar := bars.Bars[0]
	if firstBar.CurrencyCode == targetCurrency {
		fmt.Printf("fx_preview target=%s (no conversion needed)\n", targetCurrency)
		return nil
	}
	
	// For now, just show that FX preview is requested
	// In a full implementation, this would use the FX manager
	fmt.Printf("fx_preview target=%s as_of=%s rate_scale=8 rounding=half_up  (provider=yahoo-web, cache_hit=true)\n",
		targetCurrency, time.Now().Format("2006-01-02T15:04:05Z"))
	
	return nil
}

// handleBusPublishing handles bus publishing for bars
func handleBusPublishing(ctx context.Context, bars *norm.NormalizedBarBatch, busInstance *bus.Bus, busConfig *bus.Config, runID string, preview bool) error {
	// Emit to ampy-proto format
	ampyBatch, err := emit.EmitBarBatch(bars)
	if err != nil {
		return fmt.Errorf("failed to emit bar batch: %v", err)
	}
	
	// Create bus message
	busMessage := &bus.BarBatchMessage{
		Batch: ampyBatch,
		Key: &bus.Key{
			Symbol: bars.Security.Symbol,
			MIC:    bars.Security.MIC,
		},
		RunID: runID,
		Env:   busConfig.Env,
	}
	
	if preview {
		// Estimate payload size
		payloadSize := estimateBarBatchSize(ampyBatch)
		previewSummary, err := busInstance.PreviewBars(busMessage, payloadSize)
		if err != nil {
			return fmt.Errorf("failed to generate preview: %v", err)
		}
		bus.PrintPreview(previewSummary)
	} else {
		// Actually publish
		if err := busInstance.PublishBars(ctx, busMessage); err != nil {
			return fmt.Errorf("failed to publish bars: %v", err)
		}
		fmt.Printf("Published %d bars to bus\n", len(bars.Bars))
	}
	
	return nil
}

// handleQuoteBusPublishing handles bus publishing for quotes
func handleQuoteBusPublishing(ctx context.Context, quote *norm.NormalizedQuote, busInstance *bus.Bus, busConfig *bus.Config, runID string, preview bool) error {
	// Emit to ampy-proto format
	ampyQuote, err := emit.EmitQuote(quote)
	if err != nil {
		return fmt.Errorf("failed to emit quote: %v", err)
	}
	
	// Create bus message
	busMessage := &bus.QuoteMessage{
		Quote: ampyQuote,
		Key: &bus.Key{
			Symbol: quote.Security.Symbol,
			MIC:    quote.Security.MIC,
		},
		RunID: runID,
		Env:   busConfig.Env,
	}
	
	if preview {
		// Estimate payload size
		payloadSize := estimateQuoteSize(ampyQuote)
		previewSummary, err := busInstance.PreviewQuote(busMessage, payloadSize)
		if err != nil {
			return fmt.Errorf("failed to generate preview: %v", err)
		}
		bus.PrintPreview(previewSummary)
	} else {
		// Actually publish
		if err := busInstance.PublishQuote(ctx, busMessage); err != nil {
			return fmt.Errorf("failed to publish quote: %v", err)
		}
		fmt.Printf("Published quote to bus\n")
	}
	
	return nil
}

// handleLocalExport handles local export for bars
func handleLocalExport(bars *norm.NormalizedBarBatch, symbol string, start, end time.Time, adjusted bool, outFormat, outDir string) error {
	// Create output directory
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}
	
	// Generate filename
	adjustedStr := "raw"
	if adjusted {
		adjustedStr = "adjusted"
	}
	filename := fmt.Sprintf("%s_1d_%s_%s_%s.%s",
		symbol,
		start.Format("20060102"),
		end.Format("20060102"),
		adjustedStr,
		outFormat)
	
	filePath := filepath.Join(outDir, "bars", filename)
	
	// Create bars subdirectory
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create bars directory: %v", err)
	}
	
	// Write file
	switch outFormat {
	case "json":
		return writeJSONFile(filePath, bars)
	case "parquet":
		return fmt.Errorf("parquet export not implemented yet")
	default:
		return fmt.Errorf("unsupported output format: %s", outFormat)
	}
}

// handleQuoteLocalExport handles local export for quotes
func handleQuoteLocalExport(quote *norm.NormalizedQuote, ticker, outFormat, outDir string) error {
	// Create output directory
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}
	
	// Generate filename
	filename := fmt.Sprintf("%s_snapshot_quote.%s", ticker, outFormat)
	filePath := filepath.Join(outDir, "quotes", filename)
	
	// Create quotes subdirectory
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create quotes directory: %v", err)
	}
	
	// Write file
	switch outFormat {
	case "json":
		return writeJSONFile(filePath, quote)
	default:
		return fmt.Errorf("unsupported output format: %s", outFormat)
	}
}

// writeJSONFile writes data to a JSON file
func writeJSONFile(filepath string, data interface{}) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// estimateBarBatchSize estimates the size of a bar batch payload
func estimateBarBatchSize(batch interface{}) int {
	// This is a rough estimate - in a real implementation you would marshal to get exact size
	// For now, estimate based on typical bar size
	return 200 * 10 // Assume 200 bytes per bar, 10 bars average
}

// estimateQuoteSize estimates the size of a quote payload
func estimateQuoteSize(quote interface{}) int {
	// Quote messages are typically small
	return 150
}

// isPaidFeatureError checks if an error indicates a paid feature is required
func isPaidFeatureError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "paid subscription") || strings.Contains(errStr, "401") || strings.Contains(errStr, "Unauthorized")
}

// createScrapeClient creates a scrape client with configuration
func createScrapeClient(cfg *config.ScrapeConfig) (scrape.Client, error) {
	// Convert config to scrape.Config
	scrapeCfg := &scrape.Config{
		Enabled:   cfg.Enabled,
		UserAgent: cfg.UserAgent,
		TimeoutMs: cfg.TimeoutMs,
		QPS:       cfg.QPS,
		Burst:     cfg.Burst,
		Retry: scrape.RetryConfig{
			Attempts:   cfg.Retry.Attempts,
			BaseMs:     cfg.Retry.BaseMs,
			MaxDelayMs: cfg.Retry.MaxDelayMs,
		},
		RobotsPolicy: cfg.RobotsPolicy,
		CacheTTLMs:   cfg.CacheTTLMs,
		Endpoints: scrape.EndpointConfig{
			KeyStatistics: cfg.Endpoints.KeyStatistics,
			Financials:    cfg.Endpoints.Financials,
			Analysis:      cfg.Endpoints.Analysis,
			Profile:       cfg.Endpoints.Profile,
			News:          cfg.Endpoints.News,
		},
	}

	// Create scrape client
	return scrape.NewClient(scrapeCfg, nil), nil
}

// runScrapeCheck runs a scrape connectivity check
func runScrapeCheck(ctx context.Context, client scrape.Client, ticker, endpoint, runID string) error {
	// Build URL for the endpoint
	url := buildScrapeURL(ticker, endpoint)
	
	// Fetch the page
	body, meta, err := client.Fetch(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %v", url, err)
	}

	// Print results
	fmt.Printf("SCRAPE CHECK host=%s url=%s status=%d bytes=%d gzip=%t redirects=%d latency_p50≈%dms\n",
		meta.Host,
		meta.URL,
		meta.Status,
		meta.Bytes,
		meta.Gzip,
		meta.Redirects,
		meta.Duration.Milliseconds())

	// Show the full content (no truncation)
	fmt.Printf("CONTENT PREVIEW: %s\n", string(body))

	return nil
}

// runScrapePreviewNews executes the preview-news mode for testing news parser
func runScrapePreviewNews(ctx context.Context, client scrape.Client, ticker, runID string) error {
	if ticker == "" {
		return fmt.Errorf("ticker is required for preview-news mode")
	}

	fmt.Printf("PREVIEW NEWS ticker=%s\n", ticker)

	// Create a timeout context (30 seconds max)
	newsCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Build URL and fetch
	url := buildScrapeURL(ticker, "news")
	body, meta, err := client.Fetch(newsCtx, url)
	if err != nil {
		return fmt.Errorf("failed to fetch news for %s: %v", ticker, err)
	}

	fmt.Printf("FETCH META: host=%s status=%d bytes=%d gzip=%t redirects=%d latency=%dms\n",
		meta.Host, meta.Status, meta.Bytes, meta.Gzip, meta.Redirects, meta.Duration.Milliseconds())

	// Parse news
	now := time.Now()
	baseURL := fmt.Sprintf("https://%s", meta.Host)
	articles, stats, err := scrape.ParseNews(body, baseURL, now)
	if err != nil {
		return fmt.Errorf("failed to parse news: %v", err)
	}

	// Print summary
	fmt.Printf("\n%s news: found=%d deduped=%d returned=%d as_of=%s\n",
		ticker, stats.TotalFound, stats.Deduped, stats.TotalReturned, stats.AsOf.Format(time.RFC3339))

	if stats.NextPageHint != "" {
		fmt.Printf("Next page hint: %s\n", stats.NextPageHint)
	}

	// Print articles in table format
	if len(articles) > 0 {
		fmt.Printf("\nARTICLES:\n")
		for i, article := range articles {
			timeStr := "unknown"
			if article.PublishedAt != nil {
				timeStr = formatRelativeTime(*article.PublishedAt, now)
			}

			// Truncate title for display
			title := article.Title
			if len(title) > 50 {
				title = title[:47] + "..."
			}

			fmt.Printf("%2d) %-8s | %-15s | %s\n", i+1, timeStr, truncateString(article.Source, 15), title)

			// Show related tickers if any
			if len(article.RelatedTickers) > 0 {
				fmt.Printf("    Tickers: %s\n", strings.Join(article.RelatedTickers, ", "))
			}
		}
	}

	return nil
}

// formatRelativeTime formats a time relative to now for display
func formatRelativeTime(t, now time.Time) string {
	diff := now.Sub(t)
	
	if diff < time.Minute {
		return "now"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%dm ago", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else {
		weeks := int(diff.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// runScrapePreviewJSON executes the preview-json mode for testing extractors
func runScrapePreviewJSON(ctx context.Context, client scrape.Client, ticker, endpoints, runID string) error {
	if ticker == "" {
		return fmt.Errorf("ticker is required for preview-json mode")
	}
	
	if endpoints == "" {
		return fmt.Errorf("endpoints is required for preview-json mode")
	}

	// Parse endpoints
	endpointList := strings.Split(endpoints, ",")
	for i, ep := range endpointList {
		endpointList[i] = strings.TrimSpace(ep)
	}

	fmt.Printf("PREVIEW JSON EXTRACTION ticker=%s endpoints=%s\n", ticker, endpoints)

	// Process each endpoint with individual timeouts
	for _, endpoint := range endpointList {
		if endpoint == "" {
			continue
		}

		fmt.Printf("\n--- %s ---\n", strings.ToUpper(endpoint))
		
		// Create a timeout context for each endpoint (15 seconds max)
		endpointCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		
		// Build URL and fetch
		url := buildScrapeURL(ticker, endpoint)
		body, meta, err := client.Fetch(endpointCtx, url)
		cancel() // Always cancel the context
		
		if err != nil {
			fmt.Printf("ERROR: Failed to fetch %s: %v\n", url, err)
			continue
		}

		fmt.Printf("FETCHED: host=%s status=%d bytes=%d gzip=%t\n", 
			meta.Host, meta.Status, meta.Bytes, meta.Gzip)

		// Parse based on endpoint type
		switch endpoint {
		case "key-statistics":
			if dto, err := scrape.ParseComprehensiveKeyStatistics(body, ticker, "NMS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				printComprehensiveStatisticsSummary(dto)
			}
		case "profile":
			if dto, err := scrape.ParseComprehensiveProfile(body, ticker, "NMS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				printComprehensiveProfileSummary(dto)
			}
		case "financials":
			if dto, err := scrape.ParseComprehensiveFinancials(body, ticker, "NMS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				printComprehensiveFinancialsSummary(dto)
			}
		case "balance-sheet", "cash-flow":
			// For balance sheet and cash flow, we need to fetch financials page to get currency
			financialsURL := buildScrapeURL(ticker, "financials")
			fmt.Printf("FETCHING CURRENCY: %s\n", financialsURL)
			
			financialsBody, financialsMeta, err := client.Fetch(ctx, financialsURL)
			if err != nil {
				fmt.Printf("CURRENCY FETCH ERROR: %v\n", err)
				// Continue with original parsing but currency will default to USD
				if dto, err := scrape.ParseComprehensiveFinancials(body, ticker, "NMS"); err != nil {
					fmt.Printf("PARSE ERROR: %v\n", err)
				} else {
					printComprehensiveFinancialsSummary(dto)
				}
			} else {
				fmt.Printf("CURRENCY FETCHED: host=%s status=%d bytes=%d gzip=%t\n", 
					financialsMeta.Host, financialsMeta.Status, financialsMeta.Bytes, financialsMeta.Gzip)
				
				// Parse the current endpoint (balance-sheet or cash-flow) with currency from financials
				if dto, err := scrape.ParseComprehensiveFinancialsWithCurrency(body, financialsBody, ticker, "NMS"); err != nil {
					fmt.Printf("PARSE ERROR: %v\n", err)
				} else {
					printComprehensiveFinancialsSummary(dto)
				}
			}
		case "analysis":
			if dto, err := scrape.ParseAnalysis(body, ticker, "NMS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				printAnalysisSummary(dto)
			}
		case "analyst-insights":
			if dto, err := scrape.ParseAnalystInsights(body, ticker, "NMS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				printAnalystInsightsSummary(dto)
			}
		default:
			fmt.Printf("UNSUPPORTED ENDPOINT: %s (only key-statistics, profile, financials, balance-sheet, cash-flow, analysis, and analyst-insights are supported)\n", endpoint)
		}
	}

	return nil
}

// printAnalysisSummary prints a comprehensive summary of analysis data
func printAnalysisSummary(dto *scrape.ComprehensiveAnalysisDTO) {
	fmt.Printf("ANALYSIS SUMMARY: symbol=%s\n", dto.Symbol)
	
	// Earnings Estimate
	fmt.Printf("\nEARNINGS ESTIMATE (Currency: %s):\n", dto.EarningsEstimate.Currency)
	fmt.Printf("                     Current Qtr    Next Qtr    Current Year    Next Year\n")
	fmt.Printf("No. of Analysts      ")
	printAnalysisRow(dto.EarningsEstimate.CurrentQtr.NoOfAnalysts, dto.EarningsEstimate.NextQtr.NoOfAnalysts, 
		dto.EarningsEstimate.CurrentYear.NoOfAnalysts, dto.EarningsEstimate.NextYear.NoOfAnalysts, "int")
	fmt.Printf("Avg. Estimate        ")
	printAnalysisRow(dto.EarningsEstimate.CurrentQtr.AvgEstimate, dto.EarningsEstimate.NextQtr.AvgEstimate,
		dto.EarningsEstimate.CurrentYear.AvgEstimate, dto.EarningsEstimate.NextYear.AvgEstimate, "float")
	fmt.Printf("Low Estimate         ")
	printAnalysisRow(dto.EarningsEstimate.CurrentQtr.LowEstimate, dto.EarningsEstimate.NextQtr.LowEstimate,
		dto.EarningsEstimate.CurrentYear.LowEstimate, dto.EarningsEstimate.NextYear.LowEstimate, "float")
	fmt.Printf("High Estimate        ")
	printAnalysisRow(dto.EarningsEstimate.CurrentQtr.HighEstimate, dto.EarningsEstimate.NextQtr.HighEstimate,
		dto.EarningsEstimate.CurrentYear.HighEstimate, dto.EarningsEstimate.NextYear.HighEstimate, "float")
	fmt.Printf("Year Ago EPS         ")
	printAnalysisRow(dto.EarningsEstimate.CurrentQtr.YearAgoEPS, dto.EarningsEstimate.NextQtr.YearAgoEPS,
		dto.EarningsEstimate.CurrentYear.YearAgoEPS, dto.EarningsEstimate.NextYear.YearAgoEPS, "float")

	// Revenue Estimate
	fmt.Printf("\nREVENUE ESTIMATE (Currency: %s):\n", dto.RevenueEstimate.Currency)
	fmt.Printf("                     Current Qtr    Next Qtr    Current Year    Next Year\n")
	fmt.Printf("No. of Analysts      ")
	printAnalysisRow(dto.RevenueEstimate.CurrentQtr.NoOfAnalysts, dto.RevenueEstimate.NextQtr.NoOfAnalysts,
		dto.RevenueEstimate.CurrentYear.NoOfAnalysts, dto.RevenueEstimate.NextYear.NoOfAnalysts, "int")
	fmt.Printf("Avg. Estimate        ")
	printAnalysisRow(dto.RevenueEstimate.CurrentQtr.AvgEstimate, dto.RevenueEstimate.NextQtr.AvgEstimate,
		dto.RevenueEstimate.CurrentYear.AvgEstimate, dto.RevenueEstimate.NextYear.AvgEstimate, "string")
	fmt.Printf("Low Estimate         ")
	printAnalysisRow(dto.RevenueEstimate.CurrentQtr.LowEstimate, dto.RevenueEstimate.NextQtr.LowEstimate,
		dto.RevenueEstimate.CurrentYear.LowEstimate, dto.RevenueEstimate.NextYear.LowEstimate, "string")
	fmt.Printf("High Estimate        ")
	printAnalysisRow(dto.RevenueEstimate.CurrentQtr.HighEstimate, dto.RevenueEstimate.NextQtr.HighEstimate,
		dto.RevenueEstimate.CurrentYear.HighEstimate, dto.RevenueEstimate.NextYear.HighEstimate, "string")
	fmt.Printf("Year Ago Sales       ")
	printAnalysisRow(dto.RevenueEstimate.CurrentQtr.YearAgoSales, dto.RevenueEstimate.NextQtr.YearAgoSales,
		dto.RevenueEstimate.CurrentYear.YearAgoSales, dto.RevenueEstimate.NextYear.YearAgoSales, "string")
	fmt.Printf("Sales Growth         ")
	printAnalysisRow(dto.RevenueEstimate.CurrentQtr.SalesGrowthYearEst, dto.RevenueEstimate.NextQtr.SalesGrowthYearEst,
		dto.RevenueEstimate.CurrentYear.SalesGrowthYearEst, dto.RevenueEstimate.NextYear.SalesGrowthYearEst, "string")

	// Earnings History
	fmt.Printf("\nEARNINGS HISTORY (Currency: %s):\n", dto.EarningsHistory.Currency)
	if len(dto.EarningsHistory.Data) > 0 {
		fmt.Printf("Date              EPS Est.    EPS Actual    Difference    Surprise %%\n")
		for _, entry := range dto.EarningsHistory.Data {
			fmt.Printf("%-16s  ", entry.Date)
			if entry.EPSEst != nil {
				fmt.Printf("%-10.2f  ", *entry.EPSEst)
			} else {
				fmt.Printf("%-10s  ", "--")
			}
			if entry.EPSActual != nil {
				fmt.Printf("%-10.2f  ", *entry.EPSActual)
			} else {
				fmt.Printf("%-10s  ", "--")
			}
			if entry.Difference != nil {
				fmt.Printf("%-10.2f  ", *entry.Difference)
			} else {
				fmt.Printf("%-10s  ", "--")
			}
			if entry.SurprisePercent != nil {
				fmt.Printf("%-10s", *entry.SurprisePercent)
			} else {
				fmt.Printf("%-10s", "--")
			}
			fmt.Printf("\n")
		}
	}

	// EPS Trend
	fmt.Printf("\nEPS TREND (Currency: %s):\n", dto.EPSTrend.Currency)
	fmt.Printf("                     Current Qtr    Next Qtr    Current Year    Next Year\n")
	fmt.Printf("Current Estimate     ")
	printAnalysisRow(dto.EPSTrend.CurrentQtr.CurrentEstimate, dto.EPSTrend.NextQtr.CurrentEstimate,
		dto.EPSTrend.CurrentYear.CurrentEstimate, dto.EPSTrend.NextYear.CurrentEstimate, "float")
	fmt.Printf("7 Days Ago          ")
	printAnalysisRow(dto.EPSTrend.CurrentQtr.Days7Ago, dto.EPSTrend.NextQtr.Days7Ago,
		dto.EPSTrend.CurrentYear.Days7Ago, dto.EPSTrend.NextYear.Days7Ago, "float")
	fmt.Printf("30 Days Ago         ")
	printAnalysisRow(dto.EPSTrend.CurrentQtr.Days30Ago, dto.EPSTrend.NextQtr.Days30Ago,
		dto.EPSTrend.CurrentYear.Days30Ago, dto.EPSTrend.NextYear.Days30Ago, "float")
	fmt.Printf("60 Days Ago         ")
	printAnalysisRow(dto.EPSTrend.CurrentQtr.Days60Ago, dto.EPSTrend.NextQtr.Days60Ago,
		dto.EPSTrend.CurrentYear.Days60Ago, dto.EPSTrend.NextYear.Days60Ago, "float")
	fmt.Printf("90 Days Ago         ")
	printAnalysisRow(dto.EPSTrend.CurrentQtr.Days90Ago, dto.EPSTrend.NextQtr.Days90Ago,
		dto.EPSTrend.CurrentYear.Days90Ago, dto.EPSTrend.NextYear.Days90Ago, "float")

	// EPS Revisions
	fmt.Printf("\nEPS REVISIONS (Currency: %s):\n", dto.EPSRevisions.Currency)
	fmt.Printf("                     Current Qtr    Next Qtr    Current Year    Next Year\n")
	fmt.Printf("Up Last 7 Days      ")
	printAnalysisRow(dto.EPSRevisions.CurrentQtr.UpLast7Days, dto.EPSRevisions.NextQtr.UpLast7Days,
		dto.EPSRevisions.CurrentYear.UpLast7Days, dto.EPSRevisions.NextYear.UpLast7Days, "int")
	fmt.Printf("Up Last 30 Days     ")
	printAnalysisRow(dto.EPSRevisions.CurrentQtr.UpLast30Days, dto.EPSRevisions.NextQtr.UpLast30Days,
		dto.EPSRevisions.CurrentYear.UpLast30Days, dto.EPSRevisions.NextYear.UpLast30Days, "int")
	fmt.Printf("Down Last 7 Days    ")
	printAnalysisRow(dto.EPSRevisions.CurrentQtr.DownLast7Days, dto.EPSRevisions.NextQtr.DownLast7Days,
		dto.EPSRevisions.CurrentYear.DownLast7Days, dto.EPSRevisions.NextYear.DownLast7Days, "int")
	fmt.Printf("Down Last 30 Days   ")
	printAnalysisRow(dto.EPSRevisions.CurrentQtr.DownLast30Days, dto.EPSRevisions.NextQtr.DownLast30Days,
		dto.EPSRevisions.CurrentYear.DownLast30Days, dto.EPSRevisions.NextYear.DownLast30Days, "int")

	// Growth Estimate
	fmt.Printf("\nGROWTH ESTIMATE:\n")
	fmt.Printf("                     Current Qtr    Next Qtr    Current Year    Next Year\n")
	fmt.Printf("Growth Rate          ")
	printAnalysisRow(dto.GrowthEstimate.CurrentQtr, dto.GrowthEstimate.NextQtr,
		dto.GrowthEstimate.CurrentYear, dto.GrowthEstimate.NextYear, "string")
}

// printAnalysisRow prints a formatted row for analysis tables
func printAnalysisRow(currentQtr, nextQtr, currentYear, nextYear interface{}, dataType string) {
	switch dataType {
	case "int":
		printAnalysisCell(currentQtr, "int")
		printAnalysisCell(nextQtr, "int")
		printAnalysisCell(currentYear, "int")
		printAnalysisCell(nextYear, "int")
	case "float":
		printAnalysisCell(currentQtr, "float")
		printAnalysisCell(nextQtr, "float")
		printAnalysisCell(currentYear, "float")
		printAnalysisCell(nextYear, "float")
	case "string":
		printAnalysisCell(currentQtr, "string")
		printAnalysisCell(nextQtr, "string")
		printAnalysisCell(currentYear, "string")
		printAnalysisCell(nextYear, "string")
	}
	fmt.Printf("\n")
}

// printAnalysisCell prints a single cell value with proper formatting
func printAnalysisCell(value interface{}, dataType string) {
	switch dataType {
	case "int":
		if v, ok := value.(*int); ok && v != nil {
			fmt.Printf("%-15d", *v)
		} else {
			fmt.Printf("%-15s", "--")
		}
	case "float":
		if v, ok := value.(*float64); ok && v != nil {
			fmt.Printf("%-15.2f", *v)
		} else {
			fmt.Printf("%-15s", "--")
		}
	case "string":
		if v, ok := value.(*string); ok && v != nil {
			fmt.Printf("%-15s", *v)
		} else {
			fmt.Printf("%-15s", "--")
		}
	}
}

// printAnalystInsightsSummary prints a comprehensive summary of analyst insights
func printAnalystInsightsSummary(dto *scrape.AnalystInsightsDTO) {
	fmt.Printf("ANALYST INSIGHTS: symbol=%s\n", dto.Symbol)
	
	// Current Price
	if dto.CurrentPrice != nil {
		fmt.Printf("Current Price: %.2f\n", *dto.CurrentPrice)
	}
	
	// Price Targets
	fmt.Printf("\nPRICE TARGETS:\n")
	if dto.TargetMeanPrice != nil {
		fmt.Printf("  Average Target: %.2f\n", *dto.TargetMeanPrice)
	}
	if dto.TargetMedianPrice != nil {
		fmt.Printf("  Median Target: %.2f\n", *dto.TargetMedianPrice)
	}
	if dto.TargetHighPrice != nil {
		fmt.Printf("  High Target: %.2f\n", *dto.TargetHighPrice)
	}
	if dto.TargetLowPrice != nil {
		fmt.Printf("  Low Target: %.2f\n", *dto.TargetLowPrice)
	}
	
	// Analyst Recommendations
	fmt.Printf("\nANALYST RECOMMENDATIONS:\n")
	if dto.NumberOfAnalysts != nil {
		fmt.Printf("  Number of Analysts: %d\n", *dto.NumberOfAnalysts)
	}
	if dto.RecommendationMean != nil {
		fmt.Printf("  Recommendation Score: %.2f\n", *dto.RecommendationMean)
	}
	if dto.RecommendationKey != nil {
		fmt.Printf("  Recommendation: %s\n", *dto.RecommendationKey)
	}
	
	// Calculate upside/downside potential
	if dto.CurrentPrice != nil && dto.TargetMeanPrice != nil {
		upside := ((*dto.TargetMeanPrice - *dto.CurrentPrice) / *dto.CurrentPrice) * 100
		fmt.Printf("\nPOTENTIAL:\n")
		if upside > 0 {
			fmt.Printf("  Upside Potential: +%.1f%%\n", upside)
		} else {
			fmt.Printf("  Downside Risk: %.1f%%\n", upside)
		}
	}
}

// printKeyStatisticsSummary prints a summary of key statistics
func printKeyStatisticsSummary(dto *scrape.KeyStatisticsDTO) {
	fmt.Printf("KEY STATISTICS: ok fields={")
	fields := []string{}
	
	if dto.MarketCap != nil {
		fields = append(fields, "market_cap")
	}
	if dto.EnterpriseValue != nil {
		fields = append(fields, "enterprise_value")
	}
	if dto.ForwardPE != nil {
		fields = append(fields, "forward_pe")
	}
	if dto.TrailingPE != nil {
		fields = append(fields, "trailing_pe")
	}
	if dto.SharesOutstanding != nil {
		fields = append(fields, "shares_outstanding")
	}
	if dto.Beta != nil {
		fields = append(fields, "beta")
	}
	
	fmt.Printf("%s} currency=%s", strings.Join(fields, ","), dto.Currency)
	
	// Show some key numeric values (redacted format)
	if dto.MarketCap != nil {
		// Calculate the actual value correctly
		multiplier := float64(1)
		for i := 0; i < dto.MarketCap.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.MarketCap.Scaled) / multiplier
		fmt.Printf(" market_cap=~%.1fB", actualValue/1e9)
	}
	if dto.ForwardPE != nil {
		// Calculate the actual value correctly
		multiplier := float64(1)
		for i := 0; i < dto.ForwardPE.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.ForwardPE.Scaled) / multiplier
		fmt.Printf(" forward_pe=%.2f", actualValue)
	}
	if dto.SharesOutstanding != nil {
		fmt.Printf(" shares=%.1fB", float64(*dto.SharesOutstanding)/1e9)
	}
	fmt.Printf("\n")
}

// printFinancialsSummary prints a summary of financials
func printFinancialsSummary(dto *scrape.FinancialsDTO) {
	// Group lines by year
	years := make(map[int]int)
	for _, line := range dto.Lines {
		years[line.PeriodEnd.Year()]++
	}
	
	yearList := make([]int, 0, len(years))
	for year := range years {
		yearList = append(yearList, year)
	}
	sort.Ints(yearList)
	
	// Get currency from first line or default to USD
	currency := "USD"
	if len(dto.Lines) > 0 {
		currency = dto.Lines[0].Currency
	}
	fmt.Printf("FINANCIALS: lines=%d currency=%s", len(dto.Lines), currency)
	
	if len(yearList) > 0 {
		fmt.Printf(" periods=[%d..%d]", yearList[0], yearList[len(yearList)-1])
	}
	
	// Show some key financial metrics (redacted format)
	var revenue, netIncome *scrape.Scaled
	for _, line := range dto.Lines {
		if line.Key == "total_revenue" && revenue == nil {
			revenue = &line.Value
		}
		if line.Key == "net_income" && netIncome == nil {
			netIncome = &line.Value
		}
	}
	
	if revenue != nil {
		// Calculate the actual value correctly
		multiplier := float64(1)
		for i := 0; i < revenue.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(revenue.Scaled) / multiplier
		fmt.Printf(" revenue=~%.1fB", actualValue/1e9)
	}
	if netIncome != nil {
		// Calculate the actual value correctly
		multiplier := float64(1)
		for i := 0; i < netIncome.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(netIncome.Scaled) / multiplier
		fmt.Printf(" net_income=~%.1fB", actualValue/1e9)
	}
	fmt.Printf("\n")
}


// printProfileSummary prints a summary of profile
func printProfileSummary(dto *scrape.ProfileDTO) {
	employees := "unknown"
	if dto.Employees != nil {
		employees = fmt.Sprintf("~%d", *dto.Employees)
	}
	
	fmt.Printf("PROFILE: officers=%d employees=%s industry=\"%s\" sector=\"%s\"\n", 
		len(dto.Officers), employees, dto.Industry, dto.Sector)
}

// buildScrapeURL builds the URL for a given ticker and endpoint
func buildScrapeURL(ticker, endpoint string) string {
	baseURL := "https://finance.yahoo.com"
	
	switch endpoint {
	case "profile":
		return fmt.Sprintf("%s/quote/%s/profile", baseURL, ticker)
	case "key-statistics":
		return fmt.Sprintf("%s/quote/%s/key-statistics", baseURL, ticker)
	case "financials":
		return fmt.Sprintf("%s/quote/%s/financials", baseURL, ticker)
	case "balance-sheet":
		return fmt.Sprintf("%s/quote/%s/balance-sheet", baseURL, ticker)
	case "cash-flow":
		return fmt.Sprintf("%s/quote/%s/cash-flow", baseURL, ticker)
	case "analysis":
		return fmt.Sprintf("%s/quote/%s/analysis", baseURL, ticker)
	case "analyst-insights":
		return fmt.Sprintf("%s/quote/%s/analyst-insights", baseURL, ticker)
	case "news":
		return fmt.Sprintf("%s/quote/%s/news", baseURL, ticker)
	default:
		return fmt.Sprintf("%s/quote/%s", baseURL, ticker)
	}
}

// runComprehensiveStatsExtraction executes comprehensive statistics extraction
func runComprehensiveStatsExtraction(ctx context.Context, client scrape.Client, ticker, runID string) error {
	if ticker == "" {
		return fmt.Errorf("ticker is required for comprehensive stats extraction")
	}

	fmt.Printf("COMPREHENSIVE STATISTICS EXTRACTION ticker=%s\n", ticker)

	// Create a timeout context (30 seconds max)
	extractionCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Build URL for key-statistics endpoint
	url := buildScrapeURL(ticker, "key-statistics")
	body, meta, err := client.Fetch(extractionCtx, url)
	
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %w", url, err)
	}

	fmt.Printf("FETCHED: host=%s status=%d bytes=%d gzip=%t\n", 
		meta.Host, meta.Status, meta.Bytes, meta.Gzip)

	// Parse comprehensive statistics
	comprehensiveDTO, err := scrape.ParseComprehensiveKeyStatistics(body, ticker, "NMS")
	if err != nil {
		return fmt.Errorf("failed to parse comprehensive statistics: %w", err)
	}

	// Print comprehensive statistics summary
	printComprehensiveStatisticsSummary(comprehensiveDTO)

	return nil
}

// printComprehensiveStatisticsSummary prints a summary of comprehensive statistics
func printComprehensiveStatisticsSummary(dto *scrape.ComprehensiveKeyStatisticsDTO) {
	fmt.Printf("COMPREHENSIVE STATISTICS: symbol=%s currency=%s\n", dto.Symbol, dto.Currency)
	
	// Current values
	fmt.Printf("CURRENT VALUES:\n")
	if dto.Current.MarketCap != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.MarketCap.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.MarketCap.Scaled) / multiplier
		fmt.Printf("  Market Cap: %.2fB\n", actualValue/1e9)
	}
	if dto.Current.EnterpriseValue != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.EnterpriseValue.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.EnterpriseValue.Scaled) / multiplier
		fmt.Printf("  Enterprise Value: %.2fB\n", actualValue/1e9)
	}
	if dto.Current.ForwardPE != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.ForwardPE.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.ForwardPE.Scaled) / multiplier
		fmt.Printf("  Forward P/E: %.2f\n", actualValue)
	}
	if dto.Current.TrailingPE != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.TrailingPE.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.TrailingPE.Scaled) / multiplier
		fmt.Printf("  Trailing P/E: %.2f\n", actualValue)
	}
	if dto.Current.PEGRatio != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.PEGRatio.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.PEGRatio.Scaled) / multiplier
		fmt.Printf("  PEG Ratio: %.2f\n", actualValue)
	}
	if dto.Current.PriceSales != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.PriceSales.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.PriceSales.Scaled) / multiplier
		fmt.Printf("  Price/Sales: %.2f\n", actualValue)
	}
	if dto.Current.PriceBook != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.PriceBook.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.PriceBook.Scaled) / multiplier
		fmt.Printf("  Price/Book: %.2f\n", actualValue)
	}
	if dto.Current.EnterpriseValueRevenue != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.EnterpriseValueRevenue.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.EnterpriseValueRevenue.Scaled) / multiplier
		fmt.Printf("  Enterprise Value/Revenue: %.2f\n", actualValue)
	}
	if dto.Current.EnterpriseValueEBITDA != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.EnterpriseValueEBITDA.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.EnterpriseValueEBITDA.Scaled) / multiplier
		fmt.Printf("  Enterprise Value/EBITDA: %.2f\n", actualValue)
	}

	// Additional statistics
	fmt.Printf("ADDITIONAL STATISTICS:\n")
	if dto.Additional.Beta != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Additional.Beta.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Additional.Beta.Scaled) / multiplier
		fmt.Printf("  Beta: %.2f\n", actualValue)
	}
	if dto.Additional.SharesOutstanding != nil {
		fmt.Printf("  Shares Outstanding: %.2fB\n", float64(*dto.Additional.SharesOutstanding)/1e9)
	}
	if dto.Additional.ProfitMargin != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Additional.ProfitMargin.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Additional.ProfitMargin.Scaled) / multiplier
		fmt.Printf("  Profit Margin: %.2f%%\n", actualValue)
	}
	if dto.Additional.OperatingMargin != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Additional.OperatingMargin.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Additional.OperatingMargin.Scaled) / multiplier
		fmt.Printf("  Operating Margin: %.2f%%\n", actualValue)
	}
	if dto.Additional.ReturnOnAssets != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Additional.ReturnOnAssets.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Additional.ReturnOnAssets.Scaled) / multiplier
		fmt.Printf("  Return on Assets: %.2f%%\n", actualValue)
	}
	if dto.Additional.ReturnOnEquity != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Additional.ReturnOnEquity.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Additional.ReturnOnEquity.Scaled) / multiplier
		fmt.Printf("  Return on Equity: %.2f%%\n", actualValue)
	}
	
	// Historical values
	if len(dto.Historical) > 0 {
		fmt.Printf("HISTORICAL VALUES:\n")
		for _, quarter := range dto.Historical {
			fmt.Printf("  %s:\n", quarter.Date)
			if quarter.MarketCap != nil {
				multiplier := float64(1)
				for i := 0; i < quarter.MarketCap.Scale; i++ {
					multiplier *= 10
				}
				actualValue := float64(quarter.MarketCap.Scaled) / multiplier
				fmt.Printf("    Market Cap: %.2fB\n", actualValue/1e9)
			}
			if quarter.ForwardPE != nil {
				multiplier := float64(1)
				for i := 0; i < quarter.ForwardPE.Scale; i++ {
					multiplier *= 10
				}
				actualValue := float64(quarter.ForwardPE.Scaled) / multiplier
				fmt.Printf("    Forward P/E: %.2f\n", actualValue)
			}
			if quarter.TrailingPE != nil {
				multiplier := float64(1)
				for i := 0; i < quarter.TrailingPE.Scale; i++ {
					multiplier *= 10
				}
				actualValue := float64(quarter.TrailingPE.Scaled) / multiplier
				fmt.Printf("    Trailing P/E: %.2f\n", actualValue)
			}
		}
	}
}

// runComprehensiveProfile executes the comprehensive profile command
func runComprehensiveProfile(cmd *cobra.Command, args []string) error {
	// Validate flags
	if comprehensiveProfileConfig.Ticker == "" {
		return fmt.Errorf("--ticker is required")
	}

	// Generate run ID if not provided
	runID := globalConfig.RunID
	if runID == "" {
		runID = fmt.Sprintf("yfin_comprehensive_profile_%d", time.Now().Unix())
	}

	// Load configuration
	loader := config.NewLoader(globalConfig.ConfigFile)
	cfg, err := loader.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to load configuration: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Get scrape configuration
	scrapeCfg := cfg.GetScrapeConfig()
	if !scrapeCfg.Enabled {
		fmt.Fprintf(os.Stderr, "ERROR: Scraping is disabled in configuration\n")
		os.Exit(ExitConfigError)
	}

	// Initialize observability
	ctx := context.Background()
	disableTracing, _ := cmd.Flags().GetBool("observability-disable-tracing")
	disableMetrics, _ := cmd.Flags().GetBool("observability-disable-metrics")
	
	obsvConfig := &obsv.Config{
		ServiceName:       "yfinance-go",
		ServiceVersion:    version,
		Environment:       cfg.App.Env,
		CollectorEndpoint: cfg.Observability.Tracing.OTLP.Endpoint,
		TraceProtocol:     "grpc",
		SampleRatio:       cfg.Observability.Tracing.OTLP.SampleRatio,
		LogLevel:          cfg.Observability.Logs.Level,
		MetricsAddr:       cfg.Observability.Metrics.Prometheus.Addr,
		MetricsEnabled:    cfg.Observability.Metrics.Prometheus.Enabled && !disableMetrics,
		TracingEnabled:    cfg.Observability.Tracing.OTLP.Enabled && !disableTracing,
	}
	
	if err := obsv.Init(ctx, obsvConfig); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to initialize observability: %v\n", err)
		os.Exit(ExitConfigError)
	}
	defer func() { _ = obsv.Shutdown(ctx) }()

	// Create scrape client
	scrapeClient, err := createScrapeClient(scrapeCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create scrape client: %v\n", err)
		os.Exit(ExitGeneral)
	}

	// Execute comprehensive profile extraction
	return runComprehensiveProfileExtraction(ctx, scrapeClient, comprehensiveProfileConfig.Ticker, runID)
}

// runComprehensiveProfileExtraction executes comprehensive profile extraction
func runComprehensiveProfileExtraction(ctx context.Context, client scrape.Client, ticker, runID string) error {
	if ticker == "" {
		return fmt.Errorf("ticker is required for comprehensive profile extraction")
	}

	fmt.Printf("COMPREHENSIVE PROFILE EXTRACTION ticker=%s\n", ticker)

	// Create a timeout context (30 seconds max)
	extractionCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Build URL for profile endpoint
	url := buildScrapeURL(ticker, "profile")
	body, meta, err := client.Fetch(extractionCtx, url)
	
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %w", url, err)
	}

	fmt.Printf("FETCHED: host=%s status=%d bytes=%d gzip=%t\n", 
		meta.Host, meta.Status, meta.Bytes, meta.Gzip)

	// Parse comprehensive profile
	comprehensiveDTO, err := scrape.ParseComprehensiveProfile(body, ticker, "NMS")
	if err != nil {
		return fmt.Errorf("failed to parse comprehensive profile: %w", err)
	}

	// Print comprehensive profile summary
	printComprehensiveProfileSummary(comprehensiveDTO)

	return nil
}

// printComprehensiveProfileSummary prints a summary of comprehensive profile
func printComprehensiveProfileSummary(dto *scrape.ComprehensiveProfileDTO) {
	fmt.Printf("COMPREHENSIVE PROFILE: symbol=%s\n", dto.Symbol)
	
	// Company Information
	fmt.Printf("COMPANY INFORMATION:\n")
	if dto.CompanyName != "" {
		fmt.Printf("  Company Name: %s\n", dto.CompanyName)
	}
	if dto.ShortName != "" {
		fmt.Printf("  Short Name: %s\n", dto.ShortName)
	}
	if dto.Address1 != "" {
		fmt.Printf("  Address: %s\n", dto.Address1)
	}
	if dto.City != "" && dto.State != "" {
		fmt.Printf("  City, State: %s, %s\n", dto.City, dto.State)
	}
	if dto.Zip != "" {
		fmt.Printf("  ZIP: %s\n", dto.Zip)
	}
	if dto.Country != "" {
		fmt.Printf("  Country: %s\n", dto.Country)
	}
	if dto.Phone != "" {
		fmt.Printf("  Phone: %s\n", dto.Phone)
	}
	if dto.Website != "" {
		fmt.Printf("  Website: %s\n", dto.Website)
	}
	if dto.Industry != "" {
		fmt.Printf("  Industry: %s\n", dto.Industry)
	}
	if dto.Sector != "" {
		fmt.Printf("  Sector: %s\n", dto.Sector)
	}
	if dto.FullTimeEmployees != nil {
		fmt.Printf("  Full Time Employees: %d\n", *dto.FullTimeEmployees)
	}
	if dto.BusinessSummary != "" {
		// Truncate business summary if too long
		summary := dto.BusinessSummary
		if len(summary) > 200 {
			summary = summary[:200] + "..."
		}
		fmt.Printf("  Business Summary: %s\n", summary)
	}

	// Key Executives
	if len(dto.Executives) > 0 {
		fmt.Printf("KEY EXECUTIVES:\n")
		for i, exec := range dto.Executives {
			if i >= 5 { // Limit to top 5 executives
				break
			}
			fmt.Printf("  %d. %s", i+1, exec.Name)
			if exec.Title != "" {
				fmt.Printf(" - %s", exec.Title)
			}
			if exec.YearBorn != nil {
				fmt.Printf(" (Born: %d)", *exec.YearBorn)
			}
			if exec.TotalPay != nil {
				fmt.Printf(" - Total Pay: $%.2fM", float64(*exec.TotalPay)/1e6)
			}
			fmt.Printf("\n")
		}
	}

	// Additional Information
	fmt.Printf("ADDITIONAL INFORMATION:\n")
	if dto.MaxAge != nil {
		fmt.Printf("  Max Age: %d\n", *dto.MaxAge)
	}
	if dto.AuditRisk != nil {
		fmt.Printf("  Audit Risk: %d\n", *dto.AuditRisk)
	}
	if dto.BoardRisk != nil {
		fmt.Printf("  Board Risk: %d\n", *dto.BoardRisk)
	}
	if dto.CompensationRisk != nil {
		fmt.Printf("  Compensation Risk: %d\n", *dto.CompensationRisk)
	}
	if dto.ShareHolderRightsRisk != nil {
		fmt.Printf("  Share Holder Rights Risk: %d\n", *dto.ShareHolderRightsRisk)
	}
	if dto.OverallRisk != nil {
		fmt.Printf("  Overall Risk: %d\n", *dto.OverallRisk)
	}
}

// printComprehensiveFinancialsSummary prints a summary of comprehensive financials
func printComprehensiveFinancialsSummary(dto *scrape.ComprehensiveFinancialsDTO) {
	fmt.Printf("COMPREHENSIVE FINANCIALS: symbol=%s currency=%s\n", dto.Symbol, dto.Currency)
	
	// Current values
	fmt.Printf("CURRENT VALUES:\n")
	if dto.Current.TotalRevenue != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.TotalRevenue.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.TotalRevenue.Scaled) / multiplier
		fmt.Printf("  Total Revenue: %.0f\n", actualValue)
	}
	if dto.Current.CostOfRevenue != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.CostOfRevenue.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.CostOfRevenue.Scaled) / multiplier
		fmt.Printf("  Cost of Revenue: %.0f\n", actualValue)
	}
	if dto.Current.GrossProfit != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.GrossProfit.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.GrossProfit.Scaled) / multiplier
		fmt.Printf("  Gross Profit: %.0f\n", actualValue)
	}
	if dto.Current.OperatingIncome != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.OperatingIncome.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.OperatingIncome.Scaled) / multiplier
		fmt.Printf("  Operating Income: %.0f\n", actualValue)
	}
	if dto.Current.NetIncomeCommonStockholders != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.NetIncomeCommonStockholders.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.NetIncomeCommonStockholders.Scaled) / multiplier
		fmt.Printf("  Net Income: %.0f\n", actualValue)
	}
	if dto.Current.BasicEPS != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.BasicEPS.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.BasicEPS.Scaled) / multiplier
		fmt.Printf("  Basic EPS: %.2f %s\n", actualValue, dto.Currency)
	}
	if dto.Current.DilutedEPS != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.DilutedEPS.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.DilutedEPS.Scaled) / multiplier
		fmt.Printf("  Diluted EPS: %.2f %s\n", actualValue, dto.Currency)
	}
	if dto.Current.BasicAverageShares != nil {
		fmt.Printf("  Basic Average Shares: %d\n", *dto.Current.BasicAverageShares)
	}
	if dto.Current.DilutedAverageShares != nil {
		fmt.Printf("  Diluted Average Shares: %d\n", *dto.Current.DilutedAverageShares)
	}
	if dto.Current.TotalExpenses != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.TotalExpenses.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.TotalExpenses.Scaled) / multiplier
		fmt.Printf("  Total Expenses: %.0f\n", actualValue)
	}
	if dto.Current.EBIT != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.EBIT.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.EBIT.Scaled) / multiplier
		fmt.Printf("  EBIT: %.0f\n", actualValue)
	}
	if dto.Current.EBITDA != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.EBITDA.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.EBITDA.Scaled) / multiplier
		fmt.Printf("  EBITDA: %.0f\n", actualValue)
	}
	if dto.Current.NormalizedEBITDA != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.NormalizedEBITDA.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.NormalizedEBITDA.Scaled) / multiplier
		fmt.Printf("  Normalized EBITDA: %.0f\n", actualValue)
	}
	
	// Balance Sheet values
	fmt.Printf("\nBALANCE SHEET:\n")
	if dto.Current.TotalAssets != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.TotalAssets.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.TotalAssets.Scaled) / multiplier
		fmt.Printf("  Total Assets: %.0f\n", actualValue)
	}
	if dto.Current.TotalCapitalization != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.TotalCapitalization.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.TotalCapitalization.Scaled) / multiplier
		fmt.Printf("  Total Capitalization: %.0f\n", actualValue)
	}
	if dto.Current.CommonStockEquity != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.CommonStockEquity.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.CommonStockEquity.Scaled) / multiplier
		fmt.Printf("  Common Stock Equity: %.0f\n", actualValue)
	}
	if dto.Current.CapitalLeaseObligations != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.CapitalLeaseObligations.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.CapitalLeaseObligations.Scaled) / multiplier
		fmt.Printf("  Capital Lease Obligations: %.0f\n", actualValue)
	}
	if dto.Current.NetTangibleAssets != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.NetTangibleAssets.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.NetTangibleAssets.Scaled) / multiplier
		fmt.Printf("  Net Tangible Assets: %.0f\n", actualValue)
	}
	if dto.Current.WorkingCapital != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.WorkingCapital.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.WorkingCapital.Scaled) / multiplier
		fmt.Printf("  Working Capital: %.0f\n", actualValue)
	}
	if dto.Current.InvestedCapital != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.InvestedCapital.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.InvestedCapital.Scaled) / multiplier
		fmt.Printf("  Invested Capital: %.0f\n", actualValue)
	}
	if dto.Current.TangibleBookValue != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.TangibleBookValue.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.TangibleBookValue.Scaled) / multiplier
		fmt.Printf("  Tangible Book Value: %.0f\n", actualValue)
	}
	if dto.Current.TotalDebt != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.TotalDebt.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.TotalDebt.Scaled) / multiplier
		fmt.Printf("  Total Debt: %.0f\n", actualValue)
	}
	if dto.Current.ShareIssued != nil {
		fmt.Printf("  Share Issued: %d\n", *dto.Current.ShareIssued)
	}
	
	// Cash Flow values
	fmt.Printf("\nCASH FLOW:\n")
	if dto.Current.OperatingCashFlow != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.OperatingCashFlow.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.OperatingCashFlow.Scaled) / multiplier
		fmt.Printf("  Operating Cash Flow: %.0f\n", actualValue)
	}
	if dto.Current.InvestingCashFlow != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.InvestingCashFlow.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.InvestingCashFlow.Scaled) / multiplier
		fmt.Printf("  Investing Cash Flow: %.0f\n", actualValue)
	}
	if dto.Current.FinancingCashFlow != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.FinancingCashFlow.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.FinancingCashFlow.Scaled) / multiplier
		fmt.Printf("  Financing Cash Flow: %.0f\n", actualValue)
	}
	if dto.Current.EndCashPosition != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.EndCashPosition.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.EndCashPosition.Scaled) / multiplier
		fmt.Printf("  End Cash Position: %.0f\n", actualValue)
	}
	if dto.Current.CapitalExpenditure != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.CapitalExpenditure.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.CapitalExpenditure.Scaled) / multiplier
		fmt.Printf("  Capital Expenditure: %.0f\n", actualValue)
	}
	if dto.Current.IssuanceOfDebt != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.IssuanceOfDebt.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.IssuanceOfDebt.Scaled) / multiplier
		fmt.Printf("  Issuance of Debt: %.0f\n", actualValue)
	}
	if dto.Current.RepaymentOfDebt != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.RepaymentOfDebt.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.RepaymentOfDebt.Scaled) / multiplier
		fmt.Printf("  Repayment of Debt: %.0f\n", actualValue)
	}
	if dto.Current.RepurchaseOfCapitalStock != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.RepurchaseOfCapitalStock.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.RepurchaseOfCapitalStock.Scaled) / multiplier
		fmt.Printf("  Repurchase of Capital Stock: %.0f\n", actualValue)
	}
	if dto.Current.FreeCashFlow != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.FreeCashFlow.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.FreeCashFlow.Scaled) / multiplier
		fmt.Printf("  Free Cash Flow: %.0f\n", actualValue)
	}
	
	// Historical values
	fmt.Printf("HISTORICAL VALUES:\n")
	if dto.Historical.Q2_2025.TotalRevenue != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Historical.Q2_2025.TotalRevenue.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Historical.Q2_2025.TotalRevenue.Scaled) / multiplier
		fmt.Printf("  Q2 2025 Revenue: %.0f\n", actualValue)
	}
	if dto.Historical.Q1_2025.TotalRevenue != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Historical.Q1_2025.TotalRevenue.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Historical.Q1_2025.TotalRevenue.Scaled) / multiplier
		fmt.Printf("  Q1 2025 Revenue: %.0f\n", actualValue)
	}
	if dto.Historical.Q4_2024.TotalRevenue != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Historical.Q4_2024.TotalRevenue.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Historical.Q4_2024.TotalRevenue.Scaled) / multiplier
		fmt.Printf("  Q4 2024 Revenue: %.0f\n", actualValue)
	}
	
	fmt.Printf("EXTRACTED: %d fields\n", countFinancialsFields(dto))
}

// countFinancialsFields counts the number of extracted fields in financials data
func countFinancialsFields(dto *scrape.ComprehensiveFinancialsDTO) int {
	count := 0
	
	// Count current fields
	if dto.Current.TotalRevenue != nil { count++ }
	if dto.Current.CostOfRevenue != nil { count++ }
	if dto.Current.GrossProfit != nil { count++ }
	if dto.Current.OperatingIncome != nil { count++ }
	if dto.Current.NetIncomeCommonStockholders != nil { count++ }
	if dto.Current.BasicEPS != nil { count++ }
	if dto.Current.DilutedEPS != nil { count++ }
	if dto.Current.EBITDA != nil { count++ }
	
	// Count historical fields
	if dto.Historical.Q2_2025.TotalRevenue != nil { count++ }
	if dto.Historical.Q1_2025.TotalRevenue != nil { count++ }
	if dto.Historical.Q4_2024.TotalRevenue != nil { count++ }
	
	return count
}

// runScrapePreviewProto executes the preview-proto mode for testing proto emission
func runScrapePreviewProto(ctx context.Context, client scrape.Client, ticker, endpoints, runID string) error {
	if ticker == "" {
		return fmt.Errorf("ticker is required for preview-proto mode")
	}
	
	if endpoints == "" {
		return fmt.Errorf("endpoints is required for preview-proto mode")
	}

	// Parse endpoints
	endpointList := strings.Split(endpoints, ",")
	for i, ep := range endpointList {
		endpointList[i] = strings.TrimSpace(ep)
	}

	fmt.Printf("PREVIEW PROTO EMISSION ticker=%s endpoints=%s\n", ticker, endpoints)

	// Create mapper configuration
	mapperConfig := emit.ScrapeMapperConfig{
		RunID:    runID,
		Producer: fmt.Sprintf("yfin-%s", version),
		Source:   "yfinance-go/scrape",
		TraceID:  "", // Could be extracted from context if available
	}
	
	// mapper := emit.NewScrapeMapper(mapperConfig) // Not used in this function

	// Process each endpoint
	for _, endpoint := range endpointList {
		if endpoint == "" {
			continue
		}

		fmt.Printf("\n--- %s ---\n", strings.ToUpper(endpoint))
		
		// Create a timeout context for each endpoint (15 seconds max)
		endpointCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		
		// Build URL and fetch
		url := buildScrapeURL(ticker, endpoint)
		body, meta, err := client.Fetch(endpointCtx, url)
		cancel() // Always cancel the context
		
		if err != nil {
			fmt.Printf("ERROR: Failed to fetch %s: %v\n", url, err)
			continue
		}

		fmt.Printf("FETCH META: host=%s status=%d bytes=%d gzip=%t redirects=%d latency=%dms\n", 
			meta.Host, meta.Status, meta.Bytes, meta.Gzip, meta.Redirects, meta.Duration.Milliseconds())

		// Parse and map based on endpoint type
		switch endpoint {
		case "financials":
			if dto, err := scrape.ParseComprehensiveFinancials(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				// Use the comprehensive mapping for more complete data
				if snapshots, err := emit.MapComprehensiveFinancialsDTO(dto, runID, mapperConfig.Producer); err != nil {
					fmt.Printf("MAPPING ERROR: %v\n", err)
				} else {
					for _, snapshot := range snapshots {
						printFundamentalsSnapshot(snapshot)
					}
				}
			}
		
		case "profile":
			if dto, err := scrape.ParseComprehensiveProfile(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				if result, err := emit.MapProfileDTO(dto, runID, mapperConfig.Producer); err != nil {
					fmt.Printf("MAPPING ERROR: %v\n", err)
				} else {
					printProfileResult(result)
				}
			}
		
		case "news":
			if articles, stats, err := scrape.ParseNews(body, "https://finance.yahoo.com", time.Now()); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				if protoArticles, err := emit.MapNewsItems(articles, ticker, runID, mapperConfig.Producer); err != nil {
					fmt.Printf("MAPPING ERROR: %v\n", err)
				} else {
					printNewsArticles(protoArticles, stats)
				}
			}
		
		case "balance-sheet":
			if dto, err := scrape.ParseComprehensiveFinancials(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				// Balance sheet data is included in comprehensive financials
				if snapshots, err := emit.MapComprehensiveFinancialsDTO(dto, runID, mapperConfig.Producer); err != nil {
					fmt.Printf("MAPPING ERROR: %v\n", err)
				} else {
					for _, snapshot := range snapshots {
						printFundamentalsSnapshot(snapshot)
					}
				}
			}
		
		case "cash-flow":
			if dto, err := scrape.ParseComprehensiveFinancials(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				// Cash flow data is included in comprehensive financials
				if snapshots, err := emit.MapComprehensiveFinancialsDTO(dto, runID, mapperConfig.Producer); err != nil {
					fmt.Printf("MAPPING ERROR: %v\n", err)
				} else {
					for _, snapshot := range snapshots {
						printFundamentalsSnapshot(snapshot)
					}
				}
			}
		
		case "key-statistics":
			if dto, err := scrape.ParseComprehensiveKeyStatistics(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				if snapshot, err := emit.MapKeyStatisticsDTO(dto, runID, mapperConfig.Producer); err != nil {
					fmt.Printf("MAPPING ERROR: %v\n", err)
				} else {
					printFundamentalsSnapshot(snapshot)
				}
			}
		
		case "analysis":
			if dto, err := scrape.ParseAnalysis(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				if snapshot, err := emit.MapAnalysisDTO(dto, runID, mapperConfig.Producer); err != nil {
					fmt.Printf("MAPPING ERROR: %v\n", err)
				} else {
					printFundamentalsSnapshot(snapshot)
				}
			}
		
		case "analyst-insights":
			if dto, err := scrape.ParseAnalystInsights(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				if snapshot, err := emit.MapAnalystInsightsDTO(dto, runID, mapperConfig.Producer); err != nil {
					fmt.Printf("MAPPING ERROR: %v\n", err)
				} else {
					printFundamentalsSnapshot(snapshot)
				}
			}
		
		default:
			fmt.Printf("PROTO MAPPING: endpoint '%s' not yet supported for proto emission\n", endpoint)
			fmt.Printf("Supported endpoints: financials, balance-sheet, cash-flow, key-statistics, analysis, analyst-insights, profile, news\n")
		}
	}

	return nil
}

// convertToFinancialsDTO converts ComprehensiveFinancialsDTO to simple FinancialsDTO
func convertToFinancialsDTO(comprehensive *scrape.ComprehensiveFinancialsDTO) *scrape.FinancialsDTO {
	dto := &scrape.FinancialsDTO{
		Symbol: comprehensive.Symbol,
		Market: comprehensive.Market,
		AsOf:   comprehensive.AsOf,
		Lines:  []scrape.PeriodLine{},
	}

	// Use a recent quarter for period (approximate)
	now := comprehensive.AsOf
	quarterStart := time.Date(now.Year(), ((now.Month()-1)/3)*3+1, 1, 0, 0, 0, 0, time.UTC)
	quarterEnd := quarterStart.AddDate(0, 3, -1)

	// Convert current values to period lines
	if comprehensive.Current.TotalRevenue != nil {
		dto.Lines = append(dto.Lines, scrape.PeriodLine{
			PeriodStart: quarterStart,
			PeriodEnd:   quarterEnd,
			Key:         "total_revenue",
			Value:       *comprehensive.Current.TotalRevenue,
			Currency:    scrape.Currency(comprehensive.Currency),
		})
	}

	if comprehensive.Current.OperatingIncome != nil {
		dto.Lines = append(dto.Lines, scrape.PeriodLine{
			PeriodStart: quarterStart,
			PeriodEnd:   quarterEnd,
			Key:         "operating_income",
			Value:       *comprehensive.Current.OperatingIncome,
			Currency:    scrape.Currency(comprehensive.Currency),
		})
	}

	if comprehensive.Current.NetIncomeCommonStockholders != nil {
		dto.Lines = append(dto.Lines, scrape.PeriodLine{
			PeriodStart: quarterStart,
			PeriodEnd:   quarterEnd,
			Key:         "net_income",
			Value:       *comprehensive.Current.NetIncomeCommonStockholders,
			Currency:    scrape.Currency(comprehensive.Currency),
		})
	}

	if comprehensive.Current.BasicEPS != nil {
		dto.Lines = append(dto.Lines, scrape.PeriodLine{
			PeriodStart: quarterStart,
			PeriodEnd:   quarterEnd,
			Key:         "eps_basic",
			Value:       *comprehensive.Current.BasicEPS,
			Currency:    scrape.Currency(comprehensive.Currency),
		})
	}

	return dto
}

// printFundamentalsSnapshot prints a summary of fundamentals snapshot
func printFundamentalsSnapshot(snapshot *fundamentalsv1.FundamentalsSnapshot) {
	fmt.Printf("%s fundamentals: lines=%d currency=%s source=%s ok\n", 
		snapshot.Security.Symbol, 
		len(snapshot.Lines),
		getCurrencyFromLines(snapshot.Lines),
		snapshot.Source)
	
	if len(snapshot.Lines) > 0 {
		earliest, latest := getTimeBounds(snapshot.Lines)
		fmt.Printf("Period range: %s to %s\n", 
			earliest.Format("2006-01-02"), 
			latest.Format("2006-01-02"))
	}
	
	fmt.Printf("Schema version: %s\n", snapshot.Meta.SchemaVersion)
	fmt.Printf("Run ID: %s\n", snapshot.Meta.RunId)
}

// printProfileResult prints a summary of profile mapping result
func printProfileResult(result *emit.ProfileMappingResult) {
	fmt.Printf("%s profile: content_type=%s bytes=%d schema=%s\n", 
		result.Security.Symbol, 
		result.ContentType,
		len(result.JSONBytes),
		result.SchemaFQDN)
	
	fmt.Printf("Schema version: %s\n", result.Meta.SchemaVersion)
	fmt.Printf("Run ID: %s\n", result.Meta.RunId)
}

// printNewsArticles prints a summary of news articles
func printNewsArticles(articles []*newsv1.NewsItem, stats *scrape.NewsStats) {
	if len(articles) == 0 {
		fmt.Printf("No news articles found\n")
		return
	}

	summary := emit.CreateNewsSummary(articles)
	
	fmt.Printf("News articles: total=%d unique_sources=%d has_images=%d\n", 
		summary.TotalArticles,
		summary.UniqueSources,
		summary.HasImages)
	
	if summary.EarliestTime != nil && summary.LatestTime != nil {
		fmt.Printf("Time range: %s to %s\n", 
			summary.EarliestTime.Format("2006-01-02T15:04:05Z"),
			summary.LatestTime.Format("2006-01-02T15:04:05Z"))
	}
	
	if len(summary.TopSources) > 0 {
		fmt.Printf("Top sources: %s\n", strings.Join(summary.TopSources, ", "))
	}
	
	if len(summary.RelatedTickers) > 0 {
		fmt.Printf("Related tickers: %s\n", strings.Join(summary.RelatedTickers, ", "))
	}

	if len(articles) > 0 {
		fmt.Printf("Schema version: %s\n", articles[0].Meta.SchemaVersion)
		fmt.Printf("Run ID: %s\n", articles[0].Meta.RunId)
		
		// Print actual ampy-proto messages
		fmt.Printf("\n--- AMPY-PROTO NEWS MESSAGES ---\n")
		for i, article := range articles {
			if i >= 3 { // Limit to first 3 articles for readability
				fmt.Printf("... and %d more articles\n", len(articles)-3)
				break
			}
			
			// Convert to JSON for display
			jsonData, err := json.MarshalIndent(article, "", "  ")
			if err != nil {
				fmt.Printf("Error marshaling article %d: %v\n", i+1, err)
				continue
			}
			
			fmt.Printf("\nArticle %d:\n%s\n", i+1, string(jsonData))
		}
	}
}

// getCurrencyFromLines extracts currency from the first line that has one
func getCurrencyFromLines(lines []*fundamentalsv1.LineItem) string {
	for _, line := range lines {
		if line.CurrencyCode != "" {
			return line.CurrencyCode
		}
	}
	return "unknown"
}

// getTimeBounds returns the earliest and latest period bounds
func getTimeBounds(lines []*fundamentalsv1.LineItem) (time.Time, time.Time) {
	if len(lines) == 0 {
		now := time.Now()
		return now, now
	}
	
	earliest := lines[0].PeriodStart.AsTime()
	latest := lines[0].PeriodEnd.AsTime()
	
	for _, line := range lines {
		if line.PeriodStart.AsTime().Before(earliest) {
			earliest = line.PeriodStart.AsTime()
		}
		if line.PeriodEnd.AsTime().After(latest) {
			latest = line.PeriodEnd.AsTime()
		}
	}
	
	return earliest, latest
}

// runVersion executes the version command
func runVersion(cmd *cobra.Command, args []string) error {
	fmt.Printf("yfin version %s\n", version)
	fmt.Printf("commit: %s\n", commit)
	fmt.Printf("build date: %s\n", date)
	return nil
}