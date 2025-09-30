package norm

import (
	"fmt"
	"strings"
)

// ExchangeToMIC maps Yahoo Finance exchange names to MIC codes
var ExchangeToMIC = map[string]string{
	"NASDAQ":   "XNAS",
	"NMS":      "XNMS", // Nasdaq Market System
	"NasdaqGS": "XNAS", // Nasdaq Global Select Market
	"NYSE":     "XNYS",
	"AMEX":     "XASE",
	"OTC":      "OTC",
	"OTCBB":    "OTC",
	"PINK":     "OTC",
	"BATS":     "BATS",
	"EDGX":     "EDGX",
	"EDGA":     "EDGA",
	"XETR":     "XETR", // Frankfurt
	"XTKS":     "XTKS", // Tokyo
	"LSE":      "XLON", // London
	"TSE":      "XTKS", // Tokyo Stock Exchange
	"ASX":      "XASX", // Sydney
	"HKEX":     "XHKG", // Hong Kong
	"SGX":      "XSES", // Singapore
	"BSE":      "XBOM", // Mumbai
	"NSE":      "XNSE", // National Stock Exchange of India
}

// InferMIC attempts to infer the MIC code from exchange information
func InferMIC(exchangeName, fullExchangeName string) string {
	// Try full exchange name first
	if mic, ok := ExchangeToMIC[fullExchangeName]; ok {
		return mic
	}

	// Try exchange name
	if mic, ok := ExchangeToMIC[exchangeName]; ok {
		return mic
	}

	// Try case-insensitive match
	exchangeUpper := strings.ToUpper(exchangeName)
	if mic, ok := ExchangeToMIC[exchangeUpper]; ok {
		return mic
	}

	fullExchangeUpper := strings.ToUpper(fullExchangeName)
	if mic, ok := ExchangeToMIC[fullExchangeUpper]; ok {
		return mic
	}

	// Return empty string if no match found
	return ""
}

// CreateSecurity creates a Security with best-effort MIC inference
func CreateSecurity(symbol, exchangeName, fullExchangeName string) Security {
	mic := InferMIC(exchangeName, fullExchangeName)

	// Clean up symbol for specific exchanges
	cleanSymbol := cleanSymbol(symbol, mic)

	return Security{
		Symbol: cleanSymbol,
		MIC:    mic,
	}
}

// cleanSymbol cleans up symbol names for specific exchanges
func cleanSymbol(symbol, mic string) string {
	// For Tokyo Stock Exchange, remove .T suffix
	if mic == "XTKS" && strings.HasSuffix(symbol, ".T") {
		return strings.TrimSuffix(symbol, ".T")
	}

	return symbol
}

// ValidateSecurity validates a Security
func ValidateSecurity(security Security) error {
	if security.Symbol == "" {
		return fmt.Errorf("missing symbol")
	}
	return nil
}
