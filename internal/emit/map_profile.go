package emit

import (
	"fmt"
	"strings"

	"github.com/AmpyFin/yfinance-go/internal/scrape"
	commonv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/common/v1"
)

// ProfileMappingResult represents the result of profile mapping
type ProfileMappingResult struct {
	JSONBytes    []byte                 // JSON fallback when proto schema not available
	Security     *commonv1.SecurityId   // Security identifier
	Meta         *commonv1.Meta         // Metadata
	ContentType  string                 // "application/json" for JSON fallback
	SchemaFQDN   string                 // Schema fully qualified domain name
}

// MapProfileDTO converts ProfileDTO to JSON bytes (fallback since reference schema may not be available)
func MapProfileDTO(dto *scrape.ComprehensiveProfileDTO, runID, producer string) (*ProfileMappingResult, error) {
	if dto == nil {
		return nil, fmt.Errorf("ComprehensiveProfileDTO cannot be nil")
	}

	// Create normalized profile structure for JSON export
	normalizedProfile := normalizeProfileData(dto)

	// Marshal to canonical JSON
	jsonBytes, err := CanonicalMarshaler.Marshal(normalizedProfile)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal profile to canonical JSON: %w", err)
	}

	// Create security identifier
	security := &commonv1.SecurityId{
		Symbol: dto.Symbol,
		Mic:    normalizeMIC(dto.Market),
	}

	// Create metadata
	meta := &commonv1.Meta{
		RunId:         runID,
		Source:        "yfinance-go/scrape",
		Producer:      producer,
		SchemaVersion: "ampy.reference.v1:2.1.0", // Target schema version
	}

	return &ProfileMappingResult{
		JSONBytes:   jsonBytes,
		Security:    security,
		Meta:        meta,
		ContentType: "application/json",
		SchemaFQDN:  "ampy.raw.v1.JsonBlob", // Generic wrapper schema
	}, nil
}

// NormalizedProfile represents a normalized company profile structure
type NormalizedProfile struct {
	Security    SecurityInfo    `json:"security"`
	Company     CompanyInfo     `json:"company"`
	Executives  []ExecutiveInfo `json:"executives,omitempty"`
	Governance  GovernanceInfo  `json:"governance,omitempty"`
	AsOf        string          `json:"as_of"`
}

// SecurityInfo represents security identification
type SecurityInfo struct {
	Symbol string `json:"symbol"`
	Market string `json:"market"`
}

// CompanyInfo represents company details
type CompanyInfo struct {
	Name              string  `json:"name,omitempty"`
	ShortName         string  `json:"short_name,omitempty"`
	Website           string  `json:"website,omitempty"`
	Industry          string  `json:"industry,omitempty"`
	Sector            string  `json:"sector,omitempty"`
	BusinessSummary   string  `json:"business_summary,omitempty"`
	FullTimeEmployees *int64  `json:"full_time_employees,omitempty"`
	Address           Address `json:"address,omitempty"`
}

// Address represents company address
type Address struct {
	Address1 string `json:"address1,omitempty"`
	City     string `json:"city,omitempty"`
	State    string `json:"state,omitempty"`
	Zip      string `json:"zip,omitempty"`
	Country  string `json:"country,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

// ExecutiveInfo represents executive information
type ExecutiveInfo struct {
	Name             string                  `json:"name,omitempty"`
	Title            string                  `json:"title,omitempty"`
	Age              *int                    `json:"age,omitempty"`
	Compensation     *CompensationInfo       `json:"compensation,omitempty"`
}

// CompensationInfo represents executive compensation
type CompensationInfo struct {
	TotalPay         *MonetaryAmount `json:"total_pay,omitempty"`
	ExercisedValue   *MonetaryAmount `json:"exercised_value,omitempty"`
	UnexercisedValue *MonetaryAmount `json:"unexercised_value,omitempty"`
}

// MonetaryAmount represents a monetary value with currency
type MonetaryAmount struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

// GovernanceInfo represents corporate governance information
type GovernanceInfo struct {
	AuditRisk             *int64 `json:"audit_risk,omitempty"`
	BoardRisk             *int64 `json:"board_risk,omitempty"`
	CompensationRisk      *int64 `json:"compensation_risk,omitempty"`
	ShareHolderRightsRisk *int64 `json:"share_holder_rights_risk,omitempty"`
	OverallRisk           *int64 `json:"overall_risk,omitempty"`
	GovernanceEpochDate   *int64 `json:"governance_epoch_date,omitempty"`
}

// normalizeProfileData converts ComprehensiveProfileDTO to normalized structure
func normalizeProfileData(dto *scrape.ComprehensiveProfileDTO) *NormalizedProfile {
	profile := &NormalizedProfile{
		Security: SecurityInfo{
			Symbol: dto.Symbol,
			Market: dto.Market,
		},
		Company: CompanyInfo{
			Name:              cleanString(dto.CompanyName),
			ShortName:         cleanString(dto.ShortName),
			Website:           cleanString(dto.Website),
			Industry:          cleanString(dto.Industry),
			Sector:            cleanString(dto.Sector),
			BusinessSummary:   cleanString(dto.BusinessSummary),
			FullTimeEmployees: dto.FullTimeEmployees,
			Address: Address{
				Address1: cleanString(dto.Address1),
				City:     cleanString(dto.City),
				State:    cleanString(dto.State),
				Zip:      cleanString(dto.Zip),
				Country:  cleanString(dto.Country),
				Phone:    cleanString(dto.Phone),
			},
		},
		AsOf: dto.AsOf.UTC().Format("2006-01-02T15:04:05Z"),
	}

	// Convert executives
	if len(dto.Executives) > 0 {
		profile.Executives = make([]ExecutiveInfo, 0, len(dto.Executives))
		for _, exec := range dto.Executives {
			execInfo := ExecutiveInfo{
				Name:  cleanString(exec.Name),
				Title: cleanString(exec.Title),
			}

			// Convert age (year born to age)
			if exec.YearBorn != nil && *exec.YearBorn > 0 {
				currentYear := dto.AsOf.Year()
				age := currentYear - *exec.YearBorn
				if age > 0 && age < 150 { // Reasonable age bounds
					execInfo.Age = &age
				}
			}

			// Convert compensation
			if exec.TotalPay != nil || exec.ExercisedValue != nil || exec.UnexercisedValue != nil {
				comp := &CompensationInfo{}
				
				if exec.TotalPay != nil {
					comp.TotalPay = &MonetaryAmount{
						Amount:   *exec.TotalPay,
						Currency: "USD", // Assume USD for executive compensation
					}
				}
				
				if exec.ExercisedValue != nil {
					comp.ExercisedValue = &MonetaryAmount{
						Amount:   *exec.ExercisedValue,
						Currency: "USD",
					}
				}
				
				if exec.UnexercisedValue != nil {
					comp.UnexercisedValue = &MonetaryAmount{
						Amount:   *exec.UnexercisedValue,
						Currency: "USD",
					}
				}
				
				execInfo.Compensation = comp
			}

			profile.Executives = append(profile.Executives, execInfo)
		}
	}

	// Convert governance information
	if dto.AuditRisk != nil || dto.BoardRisk != nil || dto.CompensationRisk != nil ||
		dto.ShareHolderRightsRisk != nil || dto.OverallRisk != nil || dto.GovernanceEpochDate != nil {
		
		profile.Governance = GovernanceInfo{
			AuditRisk:             dto.AuditRisk,
			BoardRisk:             dto.BoardRisk,
			CompensationRisk:      dto.CompensationRisk,
			ShareHolderRightsRisk: dto.ShareHolderRightsRisk,
			OverallRisk:           dto.OverallRisk,
			GovernanceEpochDate:   dto.GovernanceEpochDate,
		}
	}

	return profile
}

// cleanString removes extra whitespace and returns empty string for whitespace-only strings
func cleanString(s string) string {
	cleaned := strings.TrimSpace(s)
	if cleaned == "" {
		return ""
	}
	
	// Replace multiple consecutive whitespaces with single space
	words := strings.Fields(cleaned)
	return strings.Join(words, " ")
}

// ValidateProfileData validates the profile data structure
func ValidateProfileData(profile *NormalizedProfile) error {
	if profile == nil {
		return fmt.Errorf("profile cannot be nil")
	}

	if profile.Security.Symbol == "" {
		return fmt.Errorf("security symbol cannot be empty")
	}

	// Validate executive compensation currencies
	for i, exec := range profile.Executives {
		if exec.Compensation != nil {
			if exec.Compensation.TotalPay != nil && exec.Compensation.TotalPay.Currency == "" {
				return fmt.Errorf("executive %d total pay currency cannot be empty", i)
			}
			if exec.Compensation.ExercisedValue != nil && exec.Compensation.ExercisedValue.Currency == "" {
				return fmt.Errorf("executive %d exercised value currency cannot be empty", i)
			}
			if exec.Compensation.UnexercisedValue != nil && exec.Compensation.UnexercisedValue.Currency == "" {
				return fmt.Errorf("executive %d unexercised value currency cannot be empty", i)
			}
		}
	}

	return nil
}

// ProfileSummary provides a concise summary of profile data for preview
type ProfileSummary struct {
	Symbol        string `json:"symbol"`
	CompanyName   string `json:"company_name"`
	Industry      string `json:"industry"`
	Sector        string `json:"sector"`
	Employees     *int64 `json:"employees,omitempty"`
	ExecutiveCount int   `json:"executive_count"`
	HasGovernance bool   `json:"has_governance"`
	JSONSizeBytes int    `json:"json_size_bytes"`
}

// CreateProfileSummary creates a summary of profile data for preview
func CreateProfileSummary(result *ProfileMappingResult, profile *NormalizedProfile) *ProfileSummary {
	summary := &ProfileSummary{
		Symbol:         profile.Security.Symbol,
		CompanyName:    profile.Company.Name,
		Industry:       profile.Company.Industry,
		Sector:         profile.Company.Sector,
		Employees:      profile.Company.FullTimeEmployees,
		ExecutiveCount: len(profile.Executives),
		HasGovernance:  profile.Governance != (GovernanceInfo{}),
		JSONSizeBytes:  len(result.JSONBytes),
	}

	return summary
}
