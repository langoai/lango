package agent

import "regexp"

// PIICategory classifies the type of personal information.
type PIICategory string

const (
	PIICategoryContact   PIICategory = "contact"
	PIICategoryIdentity  PIICategory = "identity"
	PIICategoryFinancial PIICategory = "financial"
	PIICategoryNetwork   PIICategory = "network"
)

// Valid reports whether c is a known PII category.
func (c PIICategory) Valid() bool {
	switch c {
	case PIICategoryContact, PIICategoryIdentity, PIICategoryFinancial, PIICategoryNetwork:
		return true
	}
	return false
}

// Values returns all known PII categories.
func (c PIICategory) Values() []PIICategory {
	return []PIICategory{PIICategoryContact, PIICategoryIdentity, PIICategoryFinancial, PIICategoryNetwork}
}

// PIIPatternDef defines a single PII detection pattern.
type PIIPatternDef struct {
	Name           string
	Label          string
	Category       PIICategory
	Pattern        string
	EnabledDefault bool
	Validate       func(match string) bool // optional post-match validation
}

// PIIMatch represents a single PII detection result.
type PIIMatch struct {
	PatternName string
	Category    PIICategory
	Start       int
	End         int
	Score       float64 // 1.0 for regex, variable for Presidio
}

// BuiltinPatterns defines the default PII detection patterns.
var BuiltinPatterns = []PIIPatternDef{
	// Contact patterns
	{
		Name:           "email",
		Label:          "Email Address",
		Category:       PIICategoryContact,
		Pattern:        `\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}\b`,
		EnabledDefault: true,
	},
	{
		Name:           "us_phone",
		Label:          "US Phone Number",
		Category:       PIICategoryContact,
		Pattern:        `\b\d{3}-\d{3}-\d{4}\b`,
		EnabledDefault: true,
	},
	{
		Name:           "kr_mobile",
		Label:          "Korean Mobile Number",
		Category:       PIICategoryContact,
		Pattern:        `\b01[016789]-?\d{3,4}-?\d{4}\b`,
		EnabledDefault: true,
	},
	{
		Name:           "kr_landline",
		Label:          "Korean Landline Number",
		Category:       PIICategoryContact,
		Pattern:        `\b0[2-6][1-5]?-?\d{3,4}-?\d{4}\b`,
		EnabledDefault: true,
	},
	{
		Name:           "intl_phone",
		Label:          "International Phone Number",
		Category:       PIICategoryContact,
		Pattern:        `\+\d{1,3}[-.\s]?\d{1,4}[-.\s]?\d{3,4}[-.\s]?\d{3,4}\b`,
		EnabledDefault: false,
	},

	// Identity patterns
	{
		Name:           "kr_rrn",
		Label:          "Korean Resident Registration Number",
		Category:       PIICategoryIdentity,
		Pattern:        `\b\d{6}-?[1-4]\d{6}\b`,
		EnabledDefault: true,
	},
	{
		Name:           "us_ssn",
		Label:          "US Social Security Number",
		Category:       PIICategoryIdentity,
		Pattern:        `\b\d{3}-\d{2}-\d{4}\b`,
		EnabledDefault: true,
	},
	{
		Name:           "kr_driver",
		Label:          "Korean Driver License Number",
		Category:       PIICategoryIdentity,
		Pattern:        `\b\d{2}-\d{2}-\d{6}-\d{2}\b`,
		EnabledDefault: false,
	},
	{
		Name:           "passport",
		Label:          "Passport Number",
		Category:       PIICategoryIdentity,
		Pattern:        `\b[A-Z]{1,2}\d{7,8}\b`,
		EnabledDefault: false,
	},

	// Financial patterns
	{
		Name:           "credit_card",
		Label:          "Credit Card Number",
		Category:       PIICategoryFinancial,
		Pattern:        `\b(?:4\d{3}|5[1-5]\d{2}|3[47]\d{2}|6(?:011|5\d{2}))[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{3,4}\b`,
		EnabledDefault: true,
		Validate:       validateLuhn,
	},
	{
		Name:           "kr_bank_account",
		Label:          "Korean Bank Account Number",
		Category:       PIICategoryFinancial,
		Pattern:        `\b\d{3,4}-\d{2,6}-\d{2,6}\b`,
		EnabledDefault: false,
	},
	{
		Name:           "iban",
		Label:          "IBAN",
		Category:       PIICategoryFinancial,
		Pattern:        `\b[A-Z]{2}\d{2}[A-Z0-9]{4}\d{7}([A-Z0-9]?){0,16}\b`,
		EnabledDefault: false,
	},

	// Network patterns
	{
		Name:           "ipv4",
		Label:          "IPv4 Address",
		Category:       PIICategoryNetwork,
		Pattern:        `\b(?:(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\b`,
		EnabledDefault: false,
	},
}

// builtinPatternMap is a lookup map keyed by pattern name.
var builtinPatternMap map[string]PIIPatternDef

func init() {
	builtinPatternMap = make(map[string]PIIPatternDef, len(BuiltinPatterns))
	for _, p := range BuiltinPatterns {
		builtinPatternMap[p.Name] = p
	}
}

// LookupBuiltinPattern returns a builtin pattern by name and whether it exists.
func LookupBuiltinPattern(name string) (PIIPatternDef, bool) {
	p, ok := builtinPatternMap[name]
	return p, ok
}

// validateLuhn checks whether a credit card number passes the Luhn algorithm.
func validateLuhn(s string) bool {
	// Strip separators
	var digits []int
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			digits = append(digits, int(ch-'0'))
		}
	}
	if len(digits) < 13 || len(digits) > 19 {
		return false
	}

	sum := 0
	alt := false
	for i := len(digits) - 1; i >= 0; i-- {
		d := digits[i]
		if alt {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		alt = !alt
	}
	return sum%10 == 0
}

// compilePattern compiles a PIIPatternDef into a *regexp.Regexp.
// Returns nil if the pattern is invalid.
func compilePattern(pattern string) *regexp.Regexp {
	re, err := regexp.Compile(pattern)
	if err != nil {
		piiLogger.Warnw("invalid PII pattern", "pattern", pattern, "error", err)
		return nil
	}
	return re
}
