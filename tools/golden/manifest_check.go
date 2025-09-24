package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Manifest represents the structure of MANIFEST.yaml
type Manifest struct {
	Version int `yaml:"version"`
	Items   []ManifestItem `yaml:"items"`
}

type ManifestItem struct {
	Path       string   `yaml:"path"`
	SchemaFQDN string   `yaml:"schema_fqdn"`
	SHA256     string   `yaml:"sha256"`
	Notes      []string `yaml:"notes"`
}

// Basic structures for validation (simplified versions of ampy-proto)
type Security struct {
	Symbol string `json:"symbol"`
	MIC    string `json:"mic"`
}

type ScaledDecimal struct {
	Scaled int `json:"scaled"`
	Scale  int `json:"scale"`
}

type Bar struct {
	Start              string        `json:"start"`
	End                string        `json:"end"`
	Open               ScaledDecimal `json:"open"`
	High               ScaledDecimal `json:"high"`
	Low                ScaledDecimal `json:"low"`
	Close              ScaledDecimal `json:"close"`
	Volume             int64         `json:"volume"`
	Adjusted           bool          `json:"adjusted"`
	AdjustmentPolicyID string        `json:"adjustment_policy_id"`
	EventTime          string        `json:"event_time"`
	IngestTime         string        `json:"ingest_time"`
	AsOf               string        `json:"as_of"`
}

type BarBatch struct {
	Security Security `json:"security"`
	Bars     []Bar    `json:"bars"`
	Meta     Meta     `json:"meta"`
}

type Quote struct {
	Security   Security `json:"security"`
	Type       string   `json:"type"`
	Bid        ScaledDecimal `json:"bid"`
	BidSize    int      `json:"bid_size"`
	Ask        ScaledDecimal `json:"ask"`
	AskSize    int      `json:"ask_size"`
	Venue      string   `json:"venue"`
	EventTime  string   `json:"event_time"`
	IngestTime string   `json:"ingest_time"`
	Meta       Meta     `json:"meta"`
}

type FundamentalLine struct {
	Key          string        `json:"key"`
	Value        ScaledDecimal `json:"value"`
	CurrencyCode string        `json:"currency_code"`
	PeriodStart  string        `json:"period_start"`
	PeriodEnd    string        `json:"period_end"`
}

type FundamentalSnapshot struct {
	Security Security           `json:"security"`
	Lines    []FundamentalLine  `json:"lines"`
	Source   string             `json:"source"`
	AsOf     string             `json:"as_of"`
	Meta     Meta               `json:"meta"`
}

type Meta struct {
	RunID         string `json:"run_id"`
	Source        string `json:"source"`
	Producer      string `json:"producer"`
	SchemaVersion string `json:"schema_version"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <manifest-file>\n", os.Args[0])
		os.Exit(1)
	}

	manifestPath := os.Args[1]
	
	// Load manifest
	manifest, err := loadManifest(manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load manifest: %v\n", err)
		os.Exit(1)
	}

	// Validate each item
	allValid := true
	for _, item := range manifest.Items {
		if !validateItem(item) {
			allValid = false
		}
	}

	if !allValid {
		os.Exit(1)
	}

	fmt.Println("All golden files validated successfully!")
}

func loadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	err = yaml.Unmarshal(data, &manifest)
	if err != nil {
		return nil, err
	}

	return &manifest, nil
}

func validateItem(item ManifestItem) bool {
	fmt.Printf("Validating %s...\n", item.Path)
	
	// Check if file exists
	if _, err := os.Stat(item.Path); os.IsNotExist(err) {
		fmt.Printf("  ERROR: File does not exist\n")
		return false
	}

	// Compute SHA256
	computedHash, err := computeSHA256(item.Path)
	if err != nil {
		fmt.Printf("  ERROR: Failed to compute SHA256: %v\n", err)
		return false
	}

	// Check SHA256 if provided
	if item.SHA256 != "" && item.SHA256 != "<fill-after-generate>" {
		if computedHash != item.SHA256 {
			fmt.Printf("  ERROR: SHA256 mismatch. Expected: %s, Got: %s\n", item.SHA256, computedHash)
			return false
		}
		fmt.Printf("  OK sha256 %s\n", computedHash)
	} else {
		fmt.Printf("  INFO: SHA256 not set in manifest: %s\n", computedHash)
	}

	// Validate schema based on type
	switch item.SchemaFQDN {
	case "ampy.bars.v1.BarBatch":
		if !validateBarBatch(item.Path) {
			return false
		}
	case "ampy.ticks.v1.Quote":
		if !validateQuote(item.Path) {
			return false
		}
	case "ampy.fundamentals.v1.Snapshot":
		if !validateFundamentalSnapshot(item.Path) {
			return false
		}
	default:
		fmt.Printf("  ERROR: Unknown schema FQDN: %s\n", item.SchemaFQDN)
		return false
	}

	fmt.Printf("  OK schema %s\n", item.SchemaFQDN)
	return true
}

func computeSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func validateBarBatch(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("  ERROR: Failed to read file: %v\n", err)
		return false
	}

	var batch BarBatch
	if err := json.Unmarshal(data, &batch); err != nil {
		fmt.Printf("  ERROR: Failed to parse JSON: %v\n", err)
		return false
	}

	// Validate required fields
	if batch.Security.Symbol == "" {
		fmt.Printf("  ERROR: Missing security.symbol\n")
		return false
	}

	if len(batch.Bars) == 0 {
		fmt.Printf("  ERROR: No bars found\n")
		return false
	}

	// Validate meta
	if batch.Meta.Source != "yfinance-go" {
		fmt.Printf("  ERROR: Invalid meta.source: %s\n", batch.Meta.Source)
		return false
	}

	// Validate time semantics for daily bars
	for i, bar := range batch.Bars {
		if !validateDailyBarTime(bar) {
			fmt.Printf("  ERROR: Invalid time semantics for bar %d\n", i)
			return false
		}
	}

	fmt.Printf("  OK time semantics (1d)\n")
	return true
}

func validateQuote(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("  ERROR: Failed to read file: %v\n", err)
		return false
	}

	var quote Quote
	if err := json.Unmarshal(data, &quote); err != nil {
		fmt.Printf("  ERROR: Failed to parse JSON: %v\n", err)
		return false
	}

	// Validate required fields
	if quote.Security.Symbol == "" {
		fmt.Printf("  ERROR: Missing security.symbol\n")
		return false
	}

	if quote.Type != "QUOTE" {
		fmt.Printf("  ERROR: Invalid type: %s\n", quote.Type)
		return false
	}

	// Validate meta
	if quote.Meta.Source != "yfinance-go" {
		fmt.Printf("  ERROR: Invalid meta.source: %s\n", quote.Meta.Source)
		return false
	}

	return true
}

func validateFundamentalSnapshot(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("  ERROR: Failed to read file: %v\n", err)
		return false
	}

	var snapshot FundamentalSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		fmt.Printf("  ERROR: Failed to parse JSON: %v\n", err)
		return false
	}

	// Validate required fields
	if snapshot.Security.Symbol == "" {
		fmt.Printf("  ERROR: Missing security.symbol\n")
		return false
	}

	if len(snapshot.Lines) == 0 {
		fmt.Printf("  ERROR: No fundamental lines found\n")
		return false
	}

	// Validate meta
	if snapshot.Meta.Source != "yfinance-go" {
		fmt.Printf("  ERROR: Invalid meta.source: %s\n", snapshot.Meta.Source)
		return false
	}

	// Validate currency codes
	for i, line := range snapshot.Lines {
		if line.CurrencyCode == "" {
			fmt.Printf("  ERROR: Missing currency_code for line %d\n", i)
			return false
		}
	}

	return true
}

func validateDailyBarTime(bar Bar) bool {
	// Parse times
	start, err := time.Parse(time.RFC3339, bar.Start)
	if err != nil {
		return false
	}
	
	end, err := time.Parse(time.RFC3339, bar.End)
	if err != nil {
		return false
	}
	
	eventTime, err := time.Parse(time.RFC3339, bar.EventTime)
	if err != nil {
		return false
	}

	// For daily bars:
	// - start should be 00:00:00Z
	// - end should be next day 00:00:00Z
	// - event_time should equal end
	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		return false
	}
	
	if end.Hour() != 0 || end.Minute() != 0 || end.Second() != 0 {
		return false
	}
	
	if !eventTime.Equal(end) {
		return false
	}
	
	// end should be start + 24 hours
	expectedEnd := start.Add(24 * time.Hour)
	return end.Equal(expectedEnd)
}
