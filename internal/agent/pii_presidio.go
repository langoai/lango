package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var _ PIIDetector = (*PresidioDetector)(nil)

// PresidioDetector detects PII by calling a Microsoft Presidio analyzer endpoint.
type PresidioDetector struct {
	baseURL    string
	httpClient *http.Client
	threshold  float64
	language   string
}

// PresidioOption configures a PresidioDetector.
type PresidioOption func(*PresidioDetector)

// WithPresidioThreshold sets the minimum confidence score for Presidio results.
func WithPresidioThreshold(t float64) PresidioOption {
	return func(d *PresidioDetector) { d.threshold = t }
}

// WithPresidioLanguage sets the language hint for Presidio analysis.
func WithPresidioLanguage(lang string) PresidioOption {
	return func(d *PresidioDetector) { d.language = lang }
}

// WithPresidioTimeout sets the HTTP client timeout for Presidio requests.
func WithPresidioTimeout(t time.Duration) PresidioOption {
	return func(d *PresidioDetector) { d.httpClient.Timeout = t }
}

// NewPresidioDetector creates a new Presidio-based PII detector.
func NewPresidioDetector(baseURL string, opts ...PresidioOption) *PresidioDetector {
	d := &PresidioDetector{
		baseURL:   baseURL,
		threshold: 0.7,
		language:  "en",
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// presidioRequest is the request body for the Presidio /analyze endpoint.
type presidioRequest struct {
	Text           string  `json:"text"`
	Language       string  `json:"language"`
	ScoreThreshold float64 `json:"score_threshold"`
}

// presidioResult is a single entity returned by Presidio /analyze.
type presidioResult struct {
	EntityType string  `json:"entity_type"`
	Start      int     `json:"start"`
	End        int     `json:"end"`
	Score      float64 `json:"score"`
}

// presidioEntityCategory maps Presidio entity types to PIICategory.
var presidioEntityCategory = map[string]PIICategory{
	"EMAIL_ADDRESS":           PIICategoryContact,
	"PHONE_NUMBER":            PIICategoryContact,
	"PERSON":                  PIICategoryIdentity,
	"CREDIT_CARD":             PIICategoryFinancial,
	"IBAN_CODE":               PIICategoryFinancial,
	"US_SSN":                  PIICategoryIdentity,
	"US_PASSPORT":             PIICategoryIdentity,
	"US_DRIVER_LICENSE":       PIICategoryIdentity,
	"IP_ADDRESS":              PIICategoryNetwork,
	"LOCATION":                PIICategoryIdentity,
	"DATE_TIME":               PIICategoryIdentity,
	"NRP":                     PIICategoryIdentity,
	"MEDICAL_LICENSE":         PIICategoryIdentity,
	"URL":                     PIICategoryNetwork,
	"US_BANK_NUMBER":          PIICategoryFinancial,
	"UK_NHS":                  PIICategoryIdentity,
	"SG_NRIC_FIN":             PIICategoryIdentity,
	"AU_ABN":                  PIICategoryIdentity,
	"AU_ACN":                  PIICategoryIdentity,
	"AU_TFN":                  PIICategoryIdentity,
	"AU_MEDICARE":             PIICategoryIdentity,
	"IN_PAN":                  PIICategoryIdentity,
	"IN_AADHAAR":              PIICategoryIdentity,
	"IN_VEHICLE_REGISTRATION": PIICategoryIdentity,
}

// Detect calls the Presidio /analyze endpoint and returns matches.
// On error, it returns nil (graceful degradation).
func (d *PresidioDetector) Detect(text string) []PIIMatch {
	body, err := json.Marshal(presidioRequest{
		Text:           text,
		Language:       d.language,
		ScoreThreshold: d.threshold,
	})
	if err != nil {
		piiLogger.Warnw("presidio: marshal request", "error", err)
		return nil
	}

	resp, err := d.httpClient.Post(
		d.baseURL+"/analyze",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		piiLogger.Debugw("presidio: request failed (graceful degradation)", "error", err)
		return nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		piiLogger.Warnw("presidio: non-200 status", "status", resp.StatusCode)
		return nil
	}

	var results []presidioResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		piiLogger.Warnw("presidio: decode response", "error", err)
		return nil
	}

	matches := make([]PIIMatch, 0, len(results))
	for _, r := range results {
		cat, ok := presidioEntityCategory[r.EntityType]
		if !ok {
			cat = PIICategoryIdentity // default fallback
		}
		matches = append(matches, PIIMatch{
			PatternName: "presidio:" + r.EntityType,
			Category:    cat,
			Start:       r.Start,
			End:         r.End,
			Score:       r.Score,
		})
	}
	return matches
}

// HealthCheck verifies that the Presidio analyzer service is reachable.
func (d *PresidioDetector) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("create health request: %w", err)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("presidio health check: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("presidio unhealthy: status %d", resp.StatusCode)
	}
	return nil
}
