package inspect

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/emit"
	"github.com/AmpyFin/yfinance-go/internal/scrape"
	"google.golang.org/protobuf/encoding/protojson"
)

func InspectProtoData() {
	// Tickers to analyze
	tickers := []string{"AAPL", "TSM", "BABA", "000660.KS"}

	// Create scrape client
	config := scrape.DefaultConfig()
	config.UserAgent = "yfinance-go-inspector/1.0"
	config.TimeoutMs = 30000
	client := scrape.NewClient(config, nil)

	// Create mapper config
	runID := fmt.Sprintf("inspect_%d", time.Now().Unix())
	producer := "yfinance-go-inspector"

	fmt.Printf("=== AMPY-PROTO DATA INSPECTION ===\n")
	fmt.Printf("Run ID: %s\n", runID)
	fmt.Printf("Tickers: %s\n\n", strings.Join(tickers, ", "))

	for _, ticker := range tickers {
		fmt.Printf("🔍 ANALYZING %s\n", ticker)
		fmt.Printf("%s\n\n", "="+strings.Repeat("=", len(ticker)+11))

		// Analyze financials
		if err := analyzeFinancials(client, ticker, runID, producer); err != nil {
			fmt.Printf("❌ Financials error: %v\n\n", err)
		}

		// Analyze profile
		if err := analyzeProfile(client, ticker, runID, producer); err != nil {
			fmt.Printf("❌ Profile error: %v\n\n", err)
		}

		// Analyze news
		if err := analyzeNews(client, ticker, runID, producer); err != nil {
			fmt.Printf("❌ News error: %v\n\n", err)
		}

		fmt.Printf("\n%s\n\n", strings.Repeat("-", 80))
	}
}

func analyzeFinancials(client scrape.Client, ticker, runID, producer string) error {
	fmt.Printf("📊 FINANCIALS ANALYSIS\n")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Fetch raw data
	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s/financials", ticker)
	body, meta, err := client.Fetch(ctx, url)
	if err != nil {
		return fmt.Errorf("fetch failed: %w", err)
	}

	fmt.Printf("📡 Fetch: %d bytes, %dms, %s\n", meta.Bytes, meta.Duration.Milliseconds(), meta.Host)

	// Parse to DTO
	dto, err := scrape.ParseComprehensiveFinancials(body, ticker, "XNAS")
	if err != nil {
		return fmt.Errorf("parse failed: %w", err)
	}

	// Convert to simple DTO for mapping
	simpleDTO := convertToFinancialsDTO(dto)

	// Map to ampy-proto
	snapshot, err := emit.MapFinancialsDTO(simpleDTO, runID, producer)
	if err != nil {
		return fmt.Errorf("mapping failed: %w", err)
	}

	// Display results
	fmt.Printf("🏢 Security: %s (MIC: %s)\n", snapshot.Security.Symbol, snapshot.Security.Mic)
	fmt.Printf("📅 As Of: %s\n", snapshot.AsOf.AsTime().Format("2006-01-02 15:04:05 UTC"))
	fmt.Printf("🔗 Source: %s\n", snapshot.Source)
	fmt.Printf("📋 Lines: %d\n", len(snapshot.Lines))

	// Display line items with actual values
	fmt.Printf("\n💰 FINANCIAL LINE ITEMS:\n")
	for _, line := range snapshot.Lines {
		value := float64(line.Value.Scaled) / float64(pow10(int(line.Value.Scale)))
		fmt.Printf("  • %s: %s %.2f (scale=%d, scaled=%d)\n",
			line.Key,
			line.CurrencyCode,
			value,
			line.Value.Scale,
			line.Value.Scaled)
		fmt.Printf("    Period: %s to %s\n",
			line.PeriodStart.AsTime().Format("2006-01-02"),
			line.PeriodEnd.AsTime().Format("2006-01-02"))
	}

	// Display underlying DTO data for comparison
	fmt.Printf("\n🔍 UNDERLYING DTO DATA:\n")
	if dto.Current.TotalRevenue != nil {
		fmt.Printf("  • Total Revenue: %.2f (scale=%d)\n",
			float64(dto.Current.TotalRevenue.Scaled)/float64(pow10(dto.Current.TotalRevenue.Scale)),
			dto.Current.TotalRevenue.Scale)
	}
	if dto.Current.OperatingIncome != nil {
		fmt.Printf("  • Operating Income: %.2f (scale=%d)\n",
			float64(dto.Current.OperatingIncome.Scaled)/float64(pow10(dto.Current.OperatingIncome.Scale)),
			dto.Current.OperatingIncome.Scale)
	}
	if dto.Current.NetIncomeCommonStockholders != nil {
		fmt.Printf("  • Net Income: %.2f (scale=%d)\n",
			float64(dto.Current.NetIncomeCommonStockholders.Scaled)/float64(pow10(dto.Current.NetIncomeCommonStockholders.Scale)),
			dto.Current.NetIncomeCommonStockholders.Scale)
	}
	if dto.Current.TotalDebt != nil {
		fmt.Printf("  • Total Debt: %.2f (scale=%d)\n",
			float64(dto.Current.TotalDebt.Scaled)/float64(pow10(dto.Current.TotalDebt.Scale)),
			dto.Current.TotalDebt.Scale)
	}
	if dto.Current.TotalAssets != nil {
		fmt.Printf("  • Total Assets: %.2f (scale=%d)\n",
			float64(dto.Current.TotalAssets.Scaled)/float64(pow10(dto.Current.TotalAssets.Scale)),
			dto.Current.TotalAssets.Scale)
	}
	if dto.Current.BasicEPS != nil {
		fmt.Printf("  • Basic EPS: %.4f (scale=%d)\n",
			float64(dto.Current.BasicEPS.Scaled)/float64(pow10(dto.Current.BasicEPS.Scale)),
			dto.Current.BasicEPS.Scale)
	}

	// Display proto message in JSON format
	fmt.Printf("\n📄 AMPY-PROTO MESSAGE (JSON):\n")
	jsonBytes, err := protojson.MarshalOptions{
		Multiline: true,
		Indent:    "  ",
	}.Marshal(snapshot)
	if err != nil {
		fmt.Printf("Error marshaling to JSON: %v\n", err)
	} else {
		fmt.Printf("%s\n", jsonBytes)
	}

	fmt.Printf("\n")
	return nil
}

func analyzeProfile(client scrape.Client, ticker, runID, producer string) error {
	fmt.Printf("🏢 PROFILE ANALYSIS\n")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Fetch raw data
	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s/profile", ticker)
	body, meta, err := client.Fetch(ctx, url)
	if err != nil {
		return fmt.Errorf("fetch failed: %w", err)
	}

	fmt.Printf("📡 Fetch: %d bytes, %dms, %s\n", meta.Bytes, meta.Duration.Milliseconds(), meta.Host)

	// Parse to DTO
	dto, err := scrape.ParseComprehensiveProfile(body, ticker, "XNAS")
	if err != nil {
		return fmt.Errorf("parse failed: %w", err)
	}

	// Map to result
	result, err := emit.MapProfileDTO(dto, runID, producer)
	if err != nil {
		return fmt.Errorf("mapping failed: %w", err)
	}

	// Display results
	fmt.Printf("🏢 Security: %s (MIC: %s)\n", result.Security.Symbol, result.Security.Mic)
	fmt.Printf("📄 Content Type: %s\n", result.ContentType)
	fmt.Printf("🔗 Schema: %s\n", result.SchemaFQDN)
	fmt.Printf("📊 JSON Size: %d bytes\n", len(result.JSONBytes))

	// Display underlying DTO data
	fmt.Printf("\n🔍 UNDERLYING PROFILE DATA:\n")
	fmt.Printf("  • Company Name: %s\n", dto.CompanyName)
	fmt.Printf("  • Industry: %s\n", dto.Industry)
	fmt.Printf("  • Sector: %s\n", dto.Sector)
	fmt.Printf("  • Country: %s\n", dto.Country)
	fmt.Printf("  • Website: %s\n", dto.Website)
	fmt.Printf("  • Employees: %d\n", dto.FullTimeEmployees)
	if len(dto.BusinessSummary) > 0 {
		summary := dto.BusinessSummary
		if len(summary) > 200 {
			summary = summary[:200] + "..."
		}
		fmt.Printf("  • Business Summary: %s\n", summary)
	}
	fmt.Printf("  • Executives: %d\n", len(dto.Executives))
	for i, exec := range dto.Executives {
		if i >= 3 { // Show only first 3 executives
			fmt.Printf("    ... and %d more\n", len(dto.Executives)-3)
			break
		}
		fmt.Printf("    - %s: %s\n", exec.Title, exec.Name)
	}

	// Display JSON payload (truncated)
	fmt.Printf("\n📄 JSON PAYLOAD (first 500 chars):\n")
	jsonStr := string(result.JSONBytes)
	if len(jsonStr) > 500 {
		jsonStr = jsonStr[:500] + "..."
	}
	fmt.Printf("%s\n", jsonStr)

	fmt.Printf("\n")
	return nil
}

func analyzeNews(client scrape.Client, ticker, runID, producer string) error {
	fmt.Printf("📰 NEWS ANALYSIS\n")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Fetch raw data
	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s/news", ticker)
	body, meta, err := client.Fetch(ctx, url)
	if err != nil {
		return fmt.Errorf("fetch failed: %w", err)
	}

	fmt.Printf("📡 Fetch: %d bytes, %dms, %s\n", meta.Bytes, meta.Duration.Milliseconds(), meta.Host)

	// Parse to DTO
	articles, stats, err := scrape.ParseNews(body, "https://finance.yahoo.com", time.Now())
	if err != nil {
		return fmt.Errorf("parse failed: %w", err)
	}

	// Map to ampy-proto
	protoArticles, err := emit.MapNewsItems(articles, ticker, runID, producer)
	if err != nil {
		return fmt.Errorf("mapping failed: %w", err)
	}

	// Display results
	fmt.Printf("📊 Articles Found: %d\n", len(protoArticles))
	fmt.Printf("📊 Parse Stats: %+v\n", stats)

	// Display first few articles with details
	fmt.Printf("\n📰 NEWS ARTICLES:\n")
	for i, article := range protoArticles {
		if i >= 3 { // Show only first 3 articles
			fmt.Printf("  ... and %d more articles\n", len(protoArticles)-3)
			break
		}

		fmt.Printf("  [%d] %s\n", i+1, article.Headline)
		fmt.Printf("      Source: %s\n", article.Source)
		fmt.Printf("      URL: %s\n", article.Url)
		if article.PublishedAt != nil {
			fmt.Printf("      Published: %s\n", article.PublishedAt.AsTime().Format("2006-01-02 15:04:05 UTC"))
		}
		fmt.Printf("      Tickers: %v\n", article.Tickers)
		if article.Body != "" {
			body := article.Body
			if len(body) > 150 {
				body = body[:150] + "..."
			}
			fmt.Printf("      Body: %s\n", body)
		}

		// Show underlying DTO data for comparison
		if i < len(articles) {
			srcArticle := articles[i]
			fmt.Printf("      [DTO] Title: %s\n", srcArticle.Title)
			fmt.Printf("      [DTO] Related Tickers: %v\n", srcArticle.RelatedTickers)
			if srcArticle.PublishedAt != nil {
				fmt.Printf("      [DTO] Published: %s\n", srcArticle.PublishedAt.Format("2006-01-02 15:04:05 UTC"))
			}
		}
		fmt.Printf("\n")
	}

	// Display one full proto message in JSON
	if len(protoArticles) > 0 {
		fmt.Printf("📄 SAMPLE AMPY-PROTO NEWS MESSAGE (JSON):\n")
		jsonBytes, err := protojson.MarshalOptions{
			Multiline: true,
			Indent:    "  ",
		}.Marshal(protoArticles[0])
		if err != nil {
			fmt.Printf("Error marshaling to JSON: %v\n", err)
		} else {
			fmt.Printf("%s\n", jsonBytes)
		}
	}

	fmt.Printf("\n")
	return nil
}

// Helper function to convert ComprehensiveFinancialsDTO to simple FinancialsDTO
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

	// Add more financial metrics if available
	if comprehensive.Current.TotalDebt != nil {
		dto.Lines = append(dto.Lines, scrape.PeriodLine{
			PeriodStart: quarterStart,
			PeriodEnd:   quarterEnd,
			Key:         "total_debt",
			Value:       *comprehensive.Current.TotalDebt,
			Currency:    scrape.Currency(comprehensive.Currency),
		})
	}

	if comprehensive.Current.TotalAssets != nil {
		dto.Lines = append(dto.Lines, scrape.PeriodLine{
			PeriodStart: quarterStart,
			PeriodEnd:   quarterEnd,
			Key:         "total_assets",
			Value:       *comprehensive.Current.TotalAssets,
			Currency:    scrape.Currency(comprehensive.Currency),
		})
	}

	return dto
}

// Helper function to calculate power of 10
func pow10(n int) int64 {
	if n == 0 {
		return 1
	}
	result := int64(1)
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
}
