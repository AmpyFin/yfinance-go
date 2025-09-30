package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePullFlags(t *testing.T) {
	tests := []struct {
		name    string
		config  PullConfig
		wantErr bool
	}{
		{
			name: "valid with ticker",
			config: PullConfig{
				Ticker:   "AAPL",
				Start:    "2024-01-01",
				End:      "2024-01-31",
				Adjusted: "split_dividend",
			},
			wantErr: false,
		},
		{
			name: "invalid - no ticker or universe",
			config: PullConfig{
				Start:    "2024-01-01",
				End:      "2024-01-31",
				Adjusted: "split_dividend",
			},
			wantErr: true,
		},
		{
			name: "invalid - bad adjusted value",
			config: PullConfig{
				Ticker:   "AAPL",
				Start:    "2024-01-01",
				End:      "2024-01-31",
				Adjusted: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pullConfig = tt.config
			err := validatePullFlags()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseDates(t *testing.T) {
	tests := []struct {
		name      string
		startStr  string
		endStr    string
		wantErr   bool
		wantStart time.Time
		wantEnd   time.Time
	}{
		{
			name:      "valid dates",
			startStr:  "2024-01-01",
			endStr:    "2024-01-31",
			wantErr:   false,
			wantStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "invalid start date",
			startStr: "invalid",
			endStr:   "2024-01-31",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := parseDates(tt.startStr, tt.endStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantStart, start)
				assert.Equal(t, tt.wantEnd, end)
			}
		})
	}
}

func TestParseAdjusted(t *testing.T) {
	tests := []struct {
		name     string
		adjusted string
		wantErr  bool
		want     bool
	}{
		{
			name:     "raw",
			adjusted: "raw",
			wantErr:  false,
			want:     false,
		},
		{
			name:     "split_dividend",
			adjusted: "split_dividend",
			wantErr:  false,
			want:     true,
		},
		{
			name:     "invalid",
			adjusted: "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseAdjusted(tt.adjusted)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestGetSymbols(t *testing.T) {
	// Create a temporary universe file
	tempDir := t.TempDir()
	universeFile := filepath.Join(tempDir, "symbols.txt")

	content := `AAPL
MSFT
# This is a comment
TSLA
`
	err := os.WriteFile(universeFile, []byte(content), 0644)
	require.NoError(t, err)

	tests := []struct {
		name         string
		ticker       string
		universeFile string
		wantErr      bool
		wantSymbols  []string
	}{
		{
			name:         "single ticker",
			ticker:       "AAPL",
			universeFile: "",
			wantErr:      false,
			wantSymbols:  []string{"AAPL"},
		},
		{
			name:         "universe file",
			ticker:       "",
			universeFile: universeFile,
			wantErr:      false,
			wantSymbols:  []string{"AAPL", "MSFT", "TSLA"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symbols, err := getSymbols(tt.ticker, tt.universeFile)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantSymbols, symbols)
			}
		})
	}
}

func TestWriteJSONFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")

	testData := map[string]interface{}{
		"test":   "data",
		"number": 42,
	}

	err := writeJSONFile(filePath, testData)
	require.NoError(t, err)

	// Check that file exists and has content
	info, err := os.Stat(filePath)
	require.NoError(t, err)
	assert.True(t, info.Size() > 0)

	// Check that file contains JSON
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), `"test": "data"`)
	assert.Contains(t, string(content), `"number": 42`)
}

func TestExitCodes(t *testing.T) {
	assert.Equal(t, 0, ExitSuccess)
	assert.Equal(t, 1, ExitGeneral)
	assert.Equal(t, 2, ExitPaidFeature)
	assert.Equal(t, 3, ExitConfigError)
	assert.Equal(t, 4, ExitPublishError)
}
