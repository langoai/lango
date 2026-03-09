package agent

import (
	"regexp"
	"testing"
)

func TestBuiltinPatterns_ValidRegex(t *testing.T) {
	for _, p := range BuiltinPatterns {
		t.Run(p.Name, func(t *testing.T) {
			_, err := regexp.Compile(p.Pattern)
			if err != nil {
				t.Errorf("pattern %q has invalid regex: %v", p.Name, err)
			}
		})
	}
}

func TestBuiltinPatterns_UniqueNames(t *testing.T) {
	seen := make(map[string]bool)
	for _, p := range BuiltinPatterns {
		if seen[p.Name] {
			t.Errorf("duplicate pattern name: %q", p.Name)
		}
		seen[p.Name] = true
	}
}

func TestBuiltinPatterns_HaveCategories(t *testing.T) {
	for _, p := range BuiltinPatterns {
		if p.Category == "" {
			t.Errorf("pattern %q has no category", p.Name)
		}
	}
}

func TestLookupBuiltinPattern(t *testing.T) {
	tests := []struct {
		give   string
		wantOK bool
	}{
		{give: "email", wantOK: true},
		{give: "kr_rrn", wantOK: true},
		{give: "nonexistent", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			_, ok := LookupBuiltinPattern(tt.give)
			if ok != tt.wantOK {
				t.Errorf("LookupBuiltinPattern(%q): want ok=%v, got %v", tt.give, tt.wantOK, ok)
			}
		})
	}
}

func TestBuiltinPattern_Email(t *testing.T) {
	re := regexp.MustCompile(builtinPatternMap["email"].Pattern)

	tests := []struct {
		give      string
		wantMatch bool
	}{
		{give: "user@example.com", wantMatch: true},
		{give: "test.user+tag@domain.co.kr", wantMatch: true},
		{give: "notanemail", wantMatch: false},
		{give: "@nodomain", wantMatch: false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			matched := re.MatchString(tt.give)
			if matched != tt.wantMatch {
				t.Errorf("email pattern match(%q): want %v, got %v", tt.give, tt.wantMatch, matched)
			}
		})
	}
}

func TestBuiltinPattern_KRMobile(t *testing.T) {
	re := regexp.MustCompile(builtinPatternMap["kr_mobile"].Pattern)

	tests := []struct {
		give      string
		wantMatch bool
	}{
		{give: "010-1234-5678", wantMatch: true},
		{give: "01012345678", wantMatch: true},
		{give: "011-123-4567", wantMatch: true},
		{give: "016-1234-5678", wantMatch: true},
		{give: "019-1234-5678", wantMatch: true},
		{give: "020-1234-5678", wantMatch: false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			matched := re.MatchString(tt.give)
			if matched != tt.wantMatch {
				t.Errorf("kr_mobile pattern match(%q): want %v, got %v", tt.give, tt.wantMatch, matched)
			}
		})
	}
}

func TestBuiltinPattern_KRRN(t *testing.T) {
	re := regexp.MustCompile(builtinPatternMap["kr_rrn"].Pattern)

	tests := []struct {
		give      string
		wantMatch bool
	}{
		{give: "900101-1234567", wantMatch: true},
		{give: "9001011234567", wantMatch: true},
		{give: "900101-2345678", wantMatch: true},
		{give: "900101-5234567", wantMatch: false}, // digit after dash must be 1-4
		{give: "12345-1234567", wantMatch: false},  // too short
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			matched := re.MatchString(tt.give)
			if matched != tt.wantMatch {
				t.Errorf("kr_rrn pattern match(%q): want %v, got %v", tt.give, tt.wantMatch, matched)
			}
		})
	}
}

func TestBuiltinPattern_USSSN(t *testing.T) {
	re := regexp.MustCompile(builtinPatternMap["us_ssn"].Pattern)

	tests := []struct {
		give      string
		wantMatch bool
	}{
		{give: "123-45-6789", wantMatch: true},
		{give: "000-00-0000", wantMatch: true},
		{give: "123456789", wantMatch: false},
		{give: "12-345-6789", wantMatch: false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			matched := re.MatchString(tt.give)
			if matched != tt.wantMatch {
				t.Errorf("us_ssn pattern match(%q): want %v, got %v", tt.give, tt.wantMatch, matched)
			}
		})
	}
}

func TestBuiltinPattern_CreditCard(t *testing.T) {
	re := regexp.MustCompile(builtinPatternMap["credit_card"].Pattern)

	tests := []struct {
		give      string
		wantMatch bool
	}{
		{give: "4111111111111111", wantMatch: true},    // Visa
		{give: "4111-1111-1111-1111", wantMatch: true}, // Visa with dashes
		{give: "5500000000000004", wantMatch: true},    // Mastercard
		{give: "371449635398431", wantMatch: true},     // AMEX
		{give: "6011111111111117", wantMatch: true},    // Discover
		{give: "1234567890123456", wantMatch: false},   // Invalid prefix
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			matched := re.MatchString(tt.give)
			if matched != tt.wantMatch {
				t.Errorf("credit_card pattern match(%q): want %v, got %v", tt.give, tt.wantMatch, matched)
			}
		})
	}
}

func TestValidateLuhn(t *testing.T) {
	tests := []struct {
		give   string
		wantOK bool
	}{
		{give: "4111111111111111", wantOK: true},
		{give: "4111-1111-1111-1111", wantOK: true},
		{give: "5500000000000004", wantOK: true},
		{give: "1234567890123456", wantOK: false},
		{give: "0000000000000000", wantOK: true}, // valid Luhn
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			ok := validateLuhn(tt.give)
			if ok != tt.wantOK {
				t.Errorf("validateLuhn(%q): want %v, got %v", tt.give, tt.wantOK, ok)
			}
		})
	}
}

func TestBuiltinPattern_IPv4(t *testing.T) {
	re := regexp.MustCompile(builtinPatternMap["ipv4"].Pattern)

	tests := []struct {
		give      string
		wantMatch bool
	}{
		{give: "192.168.1.1", wantMatch: true},
		{give: "10.0.0.1", wantMatch: true},
		{give: "255.255.255.255", wantMatch: true},
		{give: "999.999.999.999", wantMatch: false},
		{give: "256.1.1.1", wantMatch: false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			matched := re.MatchString(tt.give)
			if matched != tt.wantMatch {
				t.Errorf("ipv4 pattern match(%q): want %v, got %v", tt.give, tt.wantMatch, matched)
			}
		})
	}
}

func TestBuiltinPattern_KRLandline(t *testing.T) {
	re := regexp.MustCompile(builtinPatternMap["kr_landline"].Pattern)

	tests := []struct {
		give      string
		wantMatch bool
	}{
		{give: "02-1234-5678", wantMatch: true},  // Seoul
		{give: "031-123-4567", wantMatch: true},  // Gyeonggi
		{give: "051-1234-5678", wantMatch: true}, // Busan
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			matched := re.MatchString(tt.give)
			if matched != tt.wantMatch {
				t.Errorf("kr_landline pattern match(%q): want %v, got %v", tt.give, tt.wantMatch, matched)
			}
		})
	}
}
