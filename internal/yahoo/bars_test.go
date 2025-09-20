package yahoo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeBarsResponse(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "AAPL daily bars",
			filename: "AAPL_1d_sample.json",
			wantErr:  false,
		},
		{
			name:     "AAPL raw bars",
			filename: "AAPL_1d_raw_sample.json",
			wantErr:  false,
		},
		{
			name:     "SAP EUR bars",
			filename: "SAP_XETR_1d_eur.json",
			wantErr:  false,
		},
		{
			name:     "TM JPY bars",
			filename: "TM_XTKS_1d_jpy.json",
			wantErr:  false,
		},
		{
			name:     "TSLA split window",
			filename: "TSLA_1d_split_window.json",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read test data
			data, err := os.ReadFile(filepath.Join("../../testdata/source/yahoo/bars", tt.filename))
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			// Decode response
			response, err := DecodeBarsResponse(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeBarsResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Validate response structure
				if response == nil {
					t.Fatal("Response is nil")
				}

				// Check that we have results
				if len(response.Chart.Result) == 0 {
					t.Fatal("No chart results found")
				}

				result := response.Chart.Result[0]

				// Validate metadata
				if result.Meta.Symbol == "" {
					t.Error("Missing symbol in metadata")
				}
				if result.Meta.Currency == "" {
					t.Error("Missing currency in metadata")
				}

				// Validate bars data
				bars, err := response.GetBars()
				if err != nil {
					t.Errorf("Failed to get bars: %v", err)
				}

				if len(bars) == 0 {
					t.Error("No bars found")
				}

				// Validate first bar
				bar := bars[0]
				if bar.Timestamp == 0 {
					t.Error("Invalid timestamp")
				}
				if bar.Open <= 0 {
					t.Error("Invalid open price")
				}
				if bar.High <= 0 {
					t.Error("Invalid high price")
				}
				if bar.Low <= 0 {
					t.Error("Invalid low price")
				}
				if bar.Close <= 0 {
					t.Error("Invalid close price")
				}
				if bar.Volume < 0 {
					t.Error("Invalid volume")
				}

				// Check OHLC relationships
				if bar.High < bar.Low {
					t.Error("High < Low")
				}
				if bar.High < bar.Open {
					t.Error("High < Open")
				}
				if bar.High < bar.Close {
					t.Error("High < Close")
				}
				if bar.Low > bar.Open {
					t.Error("Low > Open")
				}
				if bar.Low > bar.Close {
					t.Error("Low > Close")
				}
			}
		})
	}
}

func TestBarsResponseValidation(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{
			name: "valid response",
			data: `{
				"chart": {
					"result": [{
						"meta": {
							"currency": "USD",
							"symbol": "AAPL",
							"exchangeName": "NASDAQ"
						},
						"timestamp": [1704326400],
						"indicators": {
							"quote": [{
								"open": [189.23],
								"high": [191.0],
								"low": [188.9],
								"close": [190.45],
								"volume": [43210000]
							}]
						}
					}],
					"error": null
				}
			}`,
			wantErr: false,
		},
		{
			name: "missing symbol",
			data: `{
				"chart": {
					"result": [{
						"meta": {
							"currency": "USD",
							"symbol": "",
							"exchangeName": "NASDAQ"
						},
						"timestamp": [1704326400],
						"indicators": {
							"quote": [{
								"open": [189.23],
								"high": [191.0],
								"low": [188.9],
								"close": [190.45],
								"volume": [43210000]
							}]
						}
					}],
					"error": null
				}
			}`,
			wantErr: true,
		},
		{
			name: "missing currency",
			data: `{
				"chart": {
					"result": [{
						"meta": {
							"currency": "",
							"symbol": "AAPL",
							"exchangeName": "NASDAQ"
						},
						"timestamp": [1704326400],
						"indicators": {
							"quote": [{
								"open": [189.23],
								"high": [191.0],
								"low": [188.9],
								"close": [190.45],
								"volume": [43210000]
							}]
						}
					}],
					"error": null
				}
			}`,
			wantErr: true,
		},
		{
			name: "invalid OHLC data",
			data: `{
				"chart": {
					"result": [{
						"meta": {
							"currency": "USD",
							"symbol": "AAPL",
							"exchangeName": "NASDAQ"
						},
						"timestamp": [1704326400],
						"indicators": {
							"quote": [{
								"open": [189.23],
								"high": [188.0],
								"low": [188.9],
								"close": [190.45],
								"volume": [43210000]
							}]
						}
					}],
					"error": null
				}
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response BarsResponse
			err := json.Unmarshal([]byte(tt.data), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal test data: %v", err)
			}

			err = response.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
