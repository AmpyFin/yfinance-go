package scrape

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

// Executive represents a company executive
type Executive struct {
	Name             string `json:"name,omitempty"`
	Title            string `json:"title,omitempty"`
	YearBorn         *int   `json:"year_born,omitempty"`
	TotalPay         *int64 `json:"total_pay,omitempty"`
	ExercisedValue   *int64 `json:"exercised_value,omitempty"`
	UnexercisedValue *int64 `json:"unexercised_value,omitempty"`
}

// ComprehensiveProfileDTO holds comprehensive profile data
type ComprehensiveProfileDTO struct {
	Symbol string    `json:"symbol"`
	Market string    `json:"market"`
	AsOf   time.Time `json:"as_of"`

	// Company Information
	CompanyName       string `json:"company_name,omitempty"`
	ShortName         string `json:"short_name,omitempty"`
	Address1          string `json:"address1,omitempty"`
	City              string `json:"city,omitempty"`
	State             string `json:"state,omitempty"`
	Zip               string `json:"zip,omitempty"`
	Country           string `json:"country,omitempty"`
	Phone             string `json:"phone,omitempty"`
	Website           string `json:"website,omitempty"`
	Industry          string `json:"industry,omitempty"`
	Sector            string `json:"sector,omitempty"`
	FullTimeEmployees *int64 `json:"full_time_employees,omitempty"`
	BusinessSummary   string `json:"business_summary,omitempty"`

	// Key Executives
	Executives []Executive `json:"executives,omitempty"`

	// Additional Information
	MaxAge                    *int64 `json:"max_age,omitempty"`
	AuditRisk                 *int64 `json:"audit_risk,omitempty"`
	BoardRisk                 *int64 `json:"board_risk,omitempty"`
	CompensationRisk          *int64 `json:"compensation_risk,omitempty"`
	ShareHolderRightsRisk     *int64 `json:"share_holder_rights_risk,omitempty"`
	OverallRisk               *int64 `json:"overall_risk,omitempty"`
	GovernanceEpochDate       *int64 `json:"governance_epoch_date,omitempty"`
	CompensationAsOfEpochDate *int64 `json:"compensation_as_of_epoch_date,omitempty"`
}

// extractCompanyNameFromQuote extracts company name from the quote data in the same script tag
func extractCompanyNameFromQuote(html string, dto *ComprehensiveProfileDTO) {
	// Find the script tag containing quote data for the specific symbol
	// Look for script tags with v7/finance/quote URL that contain the symbol
	scriptPattern := regexp.MustCompile(`(?s)<script type="application/json".*?v7/finance/quote.*?symbols=.*?` + dto.Symbol + `.*?>(.*?)</script>`)
	scriptMatch := scriptPattern.FindStringSubmatch(html)
	if len(scriptMatch) < 2 {
		// Try alternative pattern - look for any script tag with quote data
		scriptPattern = regexp.MustCompile(`(?s)<script type="application/json".*?quoteResponse.*?>(.*?)</script>`)
		scriptMatch = scriptPattern.FindStringSubmatch(html)
		if len(scriptMatch) < 2 {
			return
		}
	}

	scriptContent := scriptMatch[1]

	// Parse the outer JSON structure
	var outerData map[string]interface{}
	if err := json.Unmarshal([]byte(scriptContent), &outerData); err != nil {
		return
	}

	// Extract the body field which contains the inner JSON
	bodyStr, ok := outerData["body"].(string)
	if !ok {
		return
	}

	// Parse the inner JSON structure
	var innerData map[string]interface{}
	if err := json.Unmarshal([]byte(bodyStr), &innerData); err != nil {
		return
	}

	// Navigate to the quote data
	quoteResponse, ok := innerData["quoteResponse"].(map[string]interface{})
	if !ok {
		return
	}

	result, ok := quoteResponse["result"].([]interface{})
	if !ok || len(result) == 0 {
		return
	}

	// Find the quote for our symbol
	for _, quoteData := range result {
		quote, ok := quoteData.(map[string]interface{})
		if !ok {
			continue
		}

		if symbol, ok := quote["symbol"].(string); ok && symbol == dto.Symbol {
			if longName, ok := quote["longName"].(string); ok {
				dto.CompanyName = longName
			}
			if shortName, ok := quote["shortName"].(string); ok {
				dto.ShortName = shortName
			}
			break
		}
	}
}

// ParseComprehensiveProfile extracts comprehensive profile data from HTML using JSON parsing
func ParseComprehensiveProfile(html []byte, symbol, market string) (*ComprehensiveProfileDTO, error) {
	dto := &ComprehensiveProfileDTO{
		Symbol: symbol,
		Market: market,
		AsOf:   time.Now().UTC(),
	}

	htmlStr := string(html)

	// Extract profile data using JSON parsing
	if err := extractProfileFromJSON(htmlStr, dto); err != nil {
		return nil, fmt.Errorf("failed to extract profile from JSON: %w", err)
	}

	// Extract company name from quote data
	extractCompanyNameFromQuote(htmlStr, dto)

	return dto, nil
}

// extractProfileFromJSON extracts profile data from the JSON embedded in HTML
func extractProfileFromJSON(html string, dto *ComprehensiveProfileDTO) error {
	// Find the script tag containing assetProfile data
	scriptPattern := regexp.MustCompile(`(?s)<script type="application/json".*?assetProfile.*?>(.*?)</script>`)
	scriptMatch := scriptPattern.FindStringSubmatch(html)
	if len(scriptMatch) < 2 {
		return fmt.Errorf("no script tag with assetProfile found")
	}

	scriptContent := scriptMatch[1]

	// Parse the outer JSON structure
	var outerData map[string]interface{}
	if err := json.Unmarshal([]byte(scriptContent), &outerData); err != nil {
		return fmt.Errorf("failed to parse outer JSON: %w", err)
	}

	// Extract the body field which contains the inner JSON
	bodyStr, ok := outerData["body"].(string)
	if !ok {
		return fmt.Errorf("body field not found or not a string")
	}

	// Parse the inner JSON structure
	var innerData map[string]interface{}
	if err := json.Unmarshal([]byte(bodyStr), &innerData); err != nil {
		return fmt.Errorf("failed to parse inner JSON: %w", err)
	}

	// Navigate to the assetProfile data
	quoteSummary, ok := innerData["quoteSummary"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("quoteSummary not found")
	}

	result, ok := quoteSummary["result"].([]interface{})
	if !ok || len(result) == 0 {
		return fmt.Errorf("result array not found or empty")
	}

	firstResult, ok := result[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("first result item is not a map")
	}

	assetProfile, ok := firstResult["assetProfile"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("assetProfile not found")
	}

	// Extract company information
	extractCompanyInfoFromJSON(assetProfile, dto)

	// Extract executives information
	extractExecutivesFromJSON(assetProfile, dto)

	// Extract additional information
	extractAdditionalInfoFromJSON(assetProfile, dto)

	return nil
}

// extractCompanyInfoFromJSON extracts company information from the assetProfile JSON
func extractCompanyInfoFromJSON(assetProfile map[string]interface{}, dto *ComprehensiveProfileDTO) {
	// Company Name (from longName in the quote data, not assetProfile)
	// We'll need to get this from the quote data separately

	// Address information
	if val, ok := assetProfile["address1"].(string); ok {
		dto.Address1 = val
	}
	if val, ok := assetProfile["city"].(string); ok {
		dto.City = val
	}
	if val, ok := assetProfile["state"].(string); ok {
		dto.State = val
	}
	if val, ok := assetProfile["zip"].(string); ok {
		dto.Zip = val
	}
	if val, ok := assetProfile["country"].(string); ok {
		dto.Country = val
	}
	if val, ok := assetProfile["phone"].(string); ok {
		dto.Phone = val
	}
	if val, ok := assetProfile["website"].(string); ok {
		dto.Website = val
	}
	if val, ok := assetProfile["industry"].(string); ok {
		dto.Industry = val
	}
	if val, ok := assetProfile["sector"].(string); ok {
		dto.Sector = val
	}
	if val, ok := assetProfile["fullTimeEmployees"].(float64); ok {
		employees := int64(val)
		dto.FullTimeEmployees = &employees
	}
	if val, ok := assetProfile["longBusinessSummary"].(string); ok {
		dto.BusinessSummary = val
	}
}

// extractExecutivesFromJSON extracts executives information from the assetProfile JSON
func extractExecutivesFromJSON(assetProfile map[string]interface{}, dto *ComprehensiveProfileDTO) {
	companyOfficers, ok := assetProfile["companyOfficers"].([]interface{})
	if !ok {
		return
	}

	for _, officerData := range companyOfficers {
		officer, ok := officerData.(map[string]interface{})
		if !ok {
			continue
		}

		executive := Executive{}

		if val, ok := officer["name"].(string); ok {
			executive.Name = val
		}
		if val, ok := officer["title"].(string); ok {
			executive.Title = val
		}
		if val, ok := officer["yearBorn"].(float64); ok {
			yearBorn := int(val)
			executive.YearBorn = &yearBorn
		}

		// Extract total pay from nested structure
		if totalPayData, ok := officer["totalPay"].(map[string]interface{}); ok {
			if val, ok := totalPayData["raw"].(float64); ok {
				totalPay := int64(val)
				executive.TotalPay = &totalPay
			}
		}

		// Extract exercised value from nested structure
		if exercisedValueData, ok := officer["exercisedValue"].(map[string]interface{}); ok {
			if val, ok := exercisedValueData["raw"].(float64); ok {
				exercisedValue := int64(val)
				executive.ExercisedValue = &exercisedValue
			}
		}

		// Extract unexercised value from nested structure
		if unexercisedValueData, ok := officer["unexercisedValue"].(map[string]interface{}); ok {
			if val, ok := unexercisedValueData["raw"].(float64); ok {
				unexercisedValue := int64(val)
				executive.UnexercisedValue = &unexercisedValue
			}
		}

		// Only add executive if we have at least a name or title
		if executive.Name != "" || executive.Title != "" {
			dto.Executives = append(dto.Executives, executive)
		}
	}
}

// extractAdditionalInfoFromJSON extracts additional information from the assetProfile JSON
func extractAdditionalInfoFromJSON(assetProfile map[string]interface{}, dto *ComprehensiveProfileDTO) {
	if val, ok := assetProfile["maxAge"].(float64); ok {
		maxAge := int64(val)
		dto.MaxAge = &maxAge
	}
	if val, ok := assetProfile["auditRisk"].(float64); ok {
		auditRisk := int64(val)
		dto.AuditRisk = &auditRisk
	}
	if val, ok := assetProfile["boardRisk"].(float64); ok {
		boardRisk := int64(val)
		dto.BoardRisk = &boardRisk
	}
	if val, ok := assetProfile["compensationRisk"].(float64); ok {
		compensationRisk := int64(val)
		dto.CompensationRisk = &compensationRisk
	}
	if val, ok := assetProfile["shareHolderRightsRisk"].(float64); ok {
		shareHolderRightsRisk := int64(val)
		dto.ShareHolderRightsRisk = &shareHolderRightsRisk
	}
	if val, ok := assetProfile["overallRisk"].(float64); ok {
		overallRisk := int64(val)
		dto.OverallRisk = &overallRisk
	}
	if val, ok := assetProfile["governanceEpochDate"].(float64); ok {
		governanceEpochDate := int64(val)
		dto.GovernanceEpochDate = &governanceEpochDate
	}
	if val, ok := assetProfile["compensationAsOfEpochDate"].(float64); ok {
		compensationAsOfEpochDate := int64(val)
		dto.CompensationAsOfEpochDate = &compensationAsOfEpochDate
	}
}
