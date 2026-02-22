package agent

import (
	"sort"

	"github.com/langoai/lango/internal/logging"
)

var piiLogger = logging.SubsystemSugar("pii-redactor")

// PIIConfig defines configuration for PII redaction.
type PIIConfig struct {
	// Legacy fields (backward compatibility).
	RedactEmail bool
	RedactPhone bool
	CustomRegex []string

	// New pattern management.
	DisabledBuiltins []string
	CustomPatterns   map[string]string // name -> regex

	// Presidio integration.
	PresidioEnabled   bool
	PresidioURL       string
	PresidioThreshold float64
	PresidioLanguage  string
}

// PIIRedactor redacts PII from input strings using a PIIDetector.
type PIIRedactor struct {
	config   PIIConfig
	detector PIIDetector
}

// NewPIIRedactor creates a new PIIRedactor from the given configuration.
func NewPIIRedactor(cfg PIIConfig) *PIIRedactor {
	regexDet := NewRegexDetector(RegexDetectorConfig{
		DisabledBuiltins: cfg.DisabledBuiltins,
		CustomPatterns:   cfg.CustomPatterns,
		CustomRegex:      cfg.CustomRegex,
		RedactEmail:      cfg.RedactEmail,
		RedactPhone:      cfg.RedactPhone,
	})

	var detector PIIDetector = regexDet

	// If Presidio is enabled, create a CompositeDetector.
	if cfg.PresidioEnabled && cfg.PresidioURL != "" {
		var opts []PresidioOption
		if cfg.PresidioThreshold > 0 {
			opts = append(opts, WithPresidioThreshold(cfg.PresidioThreshold))
		}
		if cfg.PresidioLanguage != "" {
			opts = append(opts, WithPresidioLanguage(cfg.PresidioLanguage))
		}
		presidioDet := NewPresidioDetector(cfg.PresidioURL, opts...)
		detector = NewCompositeDetector(regexDet, presidioDet)
		piiLogger.Infow("PII redactor initialized with Presidio", "url", cfg.PresidioURL)
	}

	return &PIIRedactor{
		config:   cfg,
		detector: detector,
	}
}

// RedactInput applies PII redaction patterns to an input string.
// Detected PII is replaced with [REDACTED].
func (r *PIIRedactor) RedactInput(input string) string {
	matches := r.detector.Detect(input)
	if len(matches) == 0 {
		return input
	}

	// Sort matches by start position.
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Start < matches[j].Start
	})

	// Merge overlapping matches.
	merged := make([]PIIMatch, 0, len(matches))
	for _, m := range matches {
		if len(merged) > 0 && m.Start < merged[len(merged)-1].End {
			// Extend the previous match if this one overlaps.
			if m.End > merged[len(merged)-1].End {
				merged[len(merged)-1].End = m.End
			}
			continue
		}
		merged = append(merged, m)
	}

	// Build result by replacing matched ranges with [REDACTED].
	var result []byte
	lastPos := 0
	for _, m := range merged {
		result = append(result, input[lastPos:m.Start]...)
		result = append(result, "[REDACTED]"...)
		lastPos = m.End
	}
	result = append(result, input[lastPos:]...)

	return string(result)
}
