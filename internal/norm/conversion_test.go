package norm

import (
	"context"
	"testing"
	"time"
)

func TestNormalizedBarBatchConvertTo(t *testing.T) {
	// Create a mock FX converter
	fxConverter := &MockFXConverter{}

	// Create test data
	security := Security{Symbol: "AAPL", MIC: "XNAS"}
	bars := []NormalizedBar{
		{
			Start:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			End:          time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
			Open:         ScaledDecimal{Scaled: 10000, Scale: 4}, // 1.0000
			High:         ScaledDecimal{Scaled: 10500, Scale: 4}, // 1.0500
			Low:          ScaledDecimal{Scaled: 9500, Scale: 4},  // 0.9500
			Close:        ScaledDecimal{Scaled: 10200, Scale: 4}, // 1.0200
			CurrencyCode: "USD",
			Volume:       1000000,
			Adjusted:     false,
			EventTime:    time.Now().UTC(),
			IngestTime:   time.Now().UTC(),
			AsOf:         time.Now().UTC(),
		},
	}
	meta := Meta{
		RunID:         "test-run",
		Source:        "test",
		Producer:      "test",
		SchemaVersion: "1.0",
	}

	batch := &NormalizedBarBatch{
		Security: security,
		Bars:     bars,
		Meta:     meta,
	}

	ctx := context.Background()

	// Test same currency conversion (should succeed)
	converted, fxMeta, err := batch.ConvertTo(ctx, "USD", fxConverter)
	if err != nil {
		t.Errorf("Unexpected error for same currency: %v", err)
	}

	if converted == nil {
		t.Fatal("Converted batch is nil")
	}

	if len(converted.Bars) != len(bars) {
		t.Errorf("Expected %d bars, got %d", len(bars), len(converted.Bars))
	}

	// Verify first bar
	convertedBar := converted.Bars[0]
	if convertedBar.OriginalCurrency != "USD" {
		t.Errorf("Expected original currency USD, got %s", convertedBar.OriginalCurrency)
	}
	if convertedBar.ConvertedCurrency != "USD" {
		t.Errorf("Expected converted currency USD, got %s", convertedBar.ConvertedCurrency)
	}

	// Verify prices are unchanged for same currency
	if convertedBar.Open.Scaled != bars[0].Open.Scaled || convertedBar.Open.Scale != bars[0].Open.Scale {
		t.Errorf("Open price changed: expected %v, got %v", bars[0].Open, convertedBar.Open)
	}

	// Verify FX metadata
	if fxMeta == nil {
		t.Error("Expected FX metadata")
	} else if fxMeta.Provider != "none" {
		t.Errorf("Expected provider 'none', got '%s'", fxMeta.Provider)
	}

	// Test different currency conversion (should fail with mock provider)
	_, _, err = batch.ConvertTo(ctx, "EUR", fxConverter)
	if err == nil {
		t.Error("Expected error for different currency with mock provider")
	}
}

func TestNormalizedQuoteConvertTo(t *testing.T) {
	// Create a mock FX converter
	fxConverter := &MockFXConverter{}

	// Create test data
	security := Security{Symbol: "AAPL", MIC: "XNAS"}
	bid := ScaledDecimal{Scaled: 10000, Scale: 4} // 1.0000
	ask := ScaledDecimal{Scaled: 10050, Scale: 4} // 1.0050
	regularMarketPrice := ScaledDecimal{Scaled: 10025, Scale: 4} // 1.0025

	quote := &NormalizedQuote{
		Security:            security,
		Type:                "QUOTE",
		Bid:                 &bid,
		Ask:                 &ask,
		RegularMarketPrice:  &regularMarketPrice,
		CurrencyCode:        "USD",
		EventTime:           time.Now().UTC(),
		IngestTime:          time.Now().UTC(),
		Meta: Meta{
			RunID:         "test-run",
			Source:        "test",
			Producer:      "test",
			SchemaVersion: "1.0",
		},
	}

	ctx := context.Background()

	// Test same currency conversion (should succeed)
	converted, fxMeta, err := quote.ConvertTo(ctx, "USD", fxConverter)
	if err != nil {
		t.Errorf("Unexpected error for same currency: %v", err)
	}

	if converted == nil {
		t.Fatal("Converted quote is nil")
	}

	if converted.OriginalCurrency != "USD" {
		t.Errorf("Expected original currency USD, got %s", converted.OriginalCurrency)
	}
	if converted.ConvertedCurrency != "USD" {
		t.Errorf("Expected converted currency USD, got %s", converted.ConvertedCurrency)
	}

	// Verify prices are unchanged for same currency
	if converted.Bid.Scaled != bid.Scaled || converted.Bid.Scale != bid.Scale {
		t.Errorf("Bid price changed: expected %v, got %v", bid, *converted.Bid)
	}

	// Verify FX metadata
	if fxMeta == nil {
		t.Error("Expected FX metadata")
	} else if fxMeta.Provider != "none" {
		t.Errorf("Expected provider 'none', got '%s'", fxMeta.Provider)
	}

	// Test different currency conversion (should fail with mock provider)
	_, _, err = quote.ConvertTo(ctx, "EUR", fxConverter)
	if err == nil {
		t.Error("Expected error for different currency with mock provider")
	}
}

func TestNormalizedFundamentalsSnapshotConvertTo(t *testing.T) {
	// Create a mock FX converter
	fxConverter := &MockFXConverter{}

	// Create test data
	security := Security{Symbol: "AAPL", MIC: "XNAS"}
	lines := []NormalizedFundamentalsLine{
		{
			Key:          "revenue",
			Value:        ScaledDecimal{Scaled: 100000000, Scale: 2}, // 1,000,000.00
			CurrencyCode: "USD",
			PeriodStart:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			PeriodEnd:    time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC),
		},
	}

	snapshot := &NormalizedFundamentalsSnapshot{
		Security: security,
		Lines:    lines,
		Source:   "test",
		AsOf:     time.Now().UTC(),
		Meta: Meta{
			RunID:         "test-run",
			Source:        "test",
			Producer:      "test",
			SchemaVersion: "1.0",
		},
	}

	ctx := context.Background()

	// Test same currency conversion (should succeed)
	converted, fxMeta, err := snapshot.ConvertTo(ctx, "USD", fxConverter)
	if err != nil {
		t.Errorf("Unexpected error for same currency: %v", err)
	}

	if converted == nil {
		t.Fatal("Converted snapshot is nil")
	}

	if len(converted.Lines) != len(lines) {
		t.Errorf("Expected %d lines, got %d", len(lines), len(converted.Lines))
	}

	// Verify first line
	convertedLine := converted.Lines[0]
	if convertedLine.OriginalCurrency != "USD" {
		t.Errorf("Expected original currency USD, got %s", convertedLine.OriginalCurrency)
	}
	if convertedLine.ConvertedCurrency != "USD" {
		t.Errorf("Expected converted currency USD, got %s", convertedLine.ConvertedCurrency)
	}

	// Verify value is unchanged for same currency
	if convertedLine.Value.Scaled != lines[0].Value.Scaled || convertedLine.Value.Scale != lines[0].Value.Scale {
		t.Errorf("Value changed: expected %v, got %v", lines[0].Value, convertedLine.Value)
	}

	// Verify FX metadata
	if fxMeta == nil {
		t.Error("Expected FX metadata")
	} else if fxMeta.Provider != "none" {
		t.Errorf("Expected provider 'none', got '%s'", fxMeta.Provider)
	}

	// Test different currency conversion (should fail with mock provider)
	_, _, err = snapshot.ConvertTo(ctx, "EUR", fxConverter)
	if err == nil {
		t.Error("Expected error for different currency with mock provider")
	}
}

func TestNormalizedMarketDataConvertTo(t *testing.T) {
	// Create a mock FX converter
	fxConverter := &MockFXConverter{}

	// Create test data
	security := Security{Symbol: "AAPL", MIC: "XNAS"}
	regularMarketPrice := ScaledDecimal{Scaled: 10000, Scale: 4} // 1.0000
	fiftyTwoWeekHigh := ScaledDecimal{Scaled: 12000, Scale: 4}   // 1.2000
	fiftyTwoWeekLow := ScaledDecimal{Scaled: 8000, Scale: 4}     // 0.8000

	marketData := &NormalizedMarketData{
		Security:               security,
		RegularMarketPrice:     &regularMarketPrice,
		FiftyTwoWeekHigh:       &fiftyTwoWeekHigh,
		FiftyTwoWeekLow:        &fiftyTwoWeekLow,
		CurrencyCode:           "USD",
		EventTime:              time.Now().UTC(),
		IngestTime:             time.Now().UTC(),
		Meta: Meta{
			RunID:         "test-run",
			Source:        "test",
			Producer:      "test",
			SchemaVersion: "1.0",
		},
	}

	ctx := context.Background()

	// Test same currency conversion (should succeed)
	converted, fxMeta, err := marketData.ConvertTo(ctx, "USD", fxConverter)
	if err != nil {
		t.Errorf("Unexpected error for same currency: %v", err)
	}

	if converted == nil {
		t.Fatal("Converted market data is nil")
	}

	if converted.OriginalCurrency != "USD" {
		t.Errorf("Expected original currency USD, got %s", converted.OriginalCurrency)
	}
	if converted.ConvertedCurrency != "USD" {
		t.Errorf("Expected converted currency USD, got %s", converted.ConvertedCurrency)
	}

	// Verify prices are unchanged for same currency
	if converted.RegularMarketPrice.Scaled != regularMarketPrice.Scaled || converted.RegularMarketPrice.Scale != regularMarketPrice.Scale {
		t.Errorf("Regular market price changed: expected %v, got %v", regularMarketPrice, *converted.RegularMarketPrice)
	}

	// Verify FX metadata
	if fxMeta == nil {
		t.Error("Expected FX metadata")
	} else if fxMeta.Provider != "none" {
		t.Errorf("Expected provider 'none', got '%s'", fxMeta.Provider)
	}

	// Test different currency conversion (should fail with mock provider)
	_, _, err = marketData.ConvertTo(ctx, "EUR", fxConverter)
	if err == nil {
		t.Error("Expected error for different currency with mock provider")
	}
}

func TestConvertToEmptyData(t *testing.T) {
	// Create a mock FX converter
	fxConverter := &MockFXConverter{}

	ctx := context.Background()

	// Test empty bar batch
	emptyBatch := &NormalizedBarBatch{
		Security: Security{Symbol: "TEST"},
		Bars:     []NormalizedBar{},
		Meta:     Meta{RunID: "test"},
	}

	_, _, err := emptyBatch.ConvertTo(ctx, "USD", fxConverter)
	if err == nil {
		t.Error("Expected error for empty bar batch")
	}

	// Test empty fundamentals snapshot
	emptySnapshot := &NormalizedFundamentalsSnapshot{
		Security: Security{Symbol: "TEST"},
		Lines:    []NormalizedFundamentalsLine{},
		Meta:     Meta{RunID: "test"},
	}

	_, _, err = emptySnapshot.ConvertTo(ctx, "USD", fxConverter)
	if err == nil {
		t.Error("Expected error for empty fundamentals snapshot")
	}
}