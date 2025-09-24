package norm

import (
	"fmt"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/yahoo"
)

// NormalizeCompanyInfo normalizes company information from chart metadata
func NormalizeCompanyInfo(meta *yahoo.ChartMeta, runID string) (*NormalizedCompanyInfo, error) {
	if meta == nil {
		return nil, fmt.Errorf("metadata is nil")
	}

	// Create security
	security := Security{
		Symbol: meta.Symbol,
		MIC:    InferMIC(meta.ExchangeName, ""),
	}

	// Convert first trade date if available
	var firstTradeDate *time.Time
	if meta.FirstTradeDate != 0 {
		ftd := time.Unix(meta.FirstTradeDate, 0).UTC()
		firstTradeDate = &ftd
	}

	// Create normalized company info
	companyInfo := &NormalizedCompanyInfo{
		Security:           security,
		LongName:           meta.LongName,
		ShortName:          meta.ShortName,
		Exchange:           meta.ExchangeName,
		FullExchangeName:   meta.FullExchangeName,
		Currency:           meta.Currency,
		InstrumentType:     meta.InstrumentType,
		FirstTradeDate:     firstTradeDate,
		Timezone:           meta.Timezone,
		ExchangeTimezone:   meta.ExchangeTimezoneName,
		EventTime:          time.Now().UTC(),
		IngestTime:         time.Now().UTC(),
		Meta: Meta{
			RunID:         runID,
			Source:        "yahoo",
			Producer:      "yfinance-go",
			SchemaVersion: "1.0",
		},
	}

	return companyInfo, nil
}
