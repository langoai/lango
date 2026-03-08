package agent

import (
	"testing"
)

func TestRegexDetector_BasicDetection(t *testing.T) {
	det := NewRegexDetector(RegexDetectorConfig{
		RedactEmail: true,
		RedactPhone: true,
	})

	tests := []struct {
		give      string
		wantCount int
		wantNames []string
	}{
		{
			give:      "My email is test@example.com",
			wantCount: 1,
			wantNames: []string{"email"},
		},
		{
			give:      "Call 123-456-7890",
			wantCount: 1,
			wantNames: []string{"us_phone"},
		},
		{
			give:      "Email: a@b.com, Phone: 555-123-4567",
			wantCount: 2,
			wantNames: []string{"email", "us_phone"},
		},
		{
			give:      "No PII here",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			matches := det.Detect(tt.give)
			if len(matches) != tt.wantCount {
				t.Fatalf("want %d matches, got %d: %+v", tt.wantCount, len(matches), matches)
			}
			for i, name := range tt.wantNames {
				if matches[i].PatternName != name {
					t.Errorf("match[%d]: want name %q, got %q", i, name, matches[i].PatternName)
				}
			}
		})
	}
}

func TestRegexDetector_KoreanPatterns(t *testing.T) {
	det := NewRegexDetector(RegexDetectorConfig{
		RedactEmail: true,
		RedactPhone: true,
	})

	tests := []struct {
		give      string
		wantName  string
		wantCount int
	}{
		{
			give:      "주민번호는 900101-1234567 입니다",
			wantName:  "kr_rrn",
			wantCount: 1,
		},
		{
			give:      "전화번호: 010-1234-5678",
			wantName:  "kr_mobile",
			wantCount: 1,
		},
		{
			give:      "집전화: 02-1234-5678",
			wantName:  "kr_landline",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			matches := det.Detect(tt.give)
			if len(matches) != tt.wantCount {
				t.Fatalf("want %d matches, got %d: %+v", tt.wantCount, len(matches), matches)
			}
			if tt.wantCount > 0 && matches[0].PatternName != tt.wantName {
				t.Errorf("want pattern %q, got %q", tt.wantName, matches[0].PatternName)
			}
		})
	}
}

func TestRegexDetector_CreditCardWithLuhn(t *testing.T) {
	det := NewRegexDetector(RegexDetectorConfig{
		RedactEmail: true,
		RedactPhone: true,
	})

	tests := []struct {
		give      string
		wantCount int
	}{
		{give: "Card: 4111111111111111", wantCount: 1},    // Valid Visa
		{give: "Card: 4111111111111112", wantCount: 0},    // Invalid Luhn
		{give: "Card: 5500-0000-0000-0004", wantCount: 1}, // Valid MC
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			matches := det.Detect(tt.give)
			if len(matches) != tt.wantCount {
				t.Fatalf("want %d matches, got %d: %+v", tt.wantCount, len(matches), matches)
			}
		})
	}
}

func TestRegexDetector_DisabledBuiltins(t *testing.T) {
	det := NewRegexDetector(RegexDetectorConfig{
		RedactEmail:      true,
		RedactPhone:      true,
		DisabledBuiltins: []string{"email", "kr_rrn"},
	})

	tests := []struct {
		give      string
		wantCount int
	}{
		{give: "test@example.com", wantCount: 0},  // email disabled
		{give: "900101-1234567", wantCount: 0},    // kr_rrn disabled
		{give: "Call 123-456-7890", wantCount: 1}, // us_phone still active
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			matches := det.Detect(tt.give)
			if len(matches) != tt.wantCount {
				t.Fatalf("want %d matches, got %d: %+v", tt.wantCount, len(matches), matches)
			}
		})
	}
}

func TestRegexDetector_CustomPatterns(t *testing.T) {
	det := NewRegexDetector(RegexDetectorConfig{
		RedactEmail: true,
		RedactPhone: true,
		CustomPatterns: map[string]string{
			"employee_id": `\bEMP-\d{6}\b`,
		},
	})

	matches := det.Detect("Employee ID: EMP-123456")
	if len(matches) != 1 {
		t.Fatalf("want 1 match, got %d", len(matches))
	}
	if matches[0].PatternName != "employee_id" {
		t.Errorf("want name %q, got %q", "employee_id", matches[0].PatternName)
	}
}

func TestRegexDetector_LegacyCustomRegex(t *testing.T) {
	det := NewRegexDetector(RegexDetectorConfig{
		RedactEmail: true,
		RedactPhone: true,
		CustomRegex: []string{`\bSECRET-\d+\b`},
	})

	matches := det.Detect("Code: SECRET-42")
	if len(matches) != 1 {
		t.Fatalf("want 1 match, got %d", len(matches))
	}
}

func TestRegexDetector_LegacyEmailPhoneToggle(t *testing.T) {
	det := NewRegexDetector(RegexDetectorConfig{
		RedactEmail: false,
		RedactPhone: false,
	})

	matches := det.Detect("test@example.com 123-456-7890")
	// Email and phone disabled, but other patterns (kr_rrn etc.) with default enabled still work
	for _, m := range matches {
		if m.PatternName == "email" || m.PatternName == "us_phone" {
			t.Errorf("unexpected match for disabled pattern: %q", m.PatternName)
		}
	}
}

func TestRegexDetector_MatchPositions(t *testing.T) {
	det := NewRegexDetector(RegexDetectorConfig{
		RedactEmail: true,
		RedactPhone: true,
	})

	text := "Email: user@test.com here"
	matches := det.Detect(text)
	if len(matches) != 1 {
		t.Fatalf("want 1 match, got %d", len(matches))
	}

	m := matches[0]
	if text[m.Start:m.End] != "user@test.com" {
		t.Errorf("match text: want %q, got %q", "user@test.com", text[m.Start:m.End])
	}
	if m.Score != 1.0 {
		t.Errorf("score: want 1.0, got %f", m.Score)
	}
}

func TestCompositeDetector_ChainsDetectors(t *testing.T) {
	d1 := NewRegexDetector(RegexDetectorConfig{
		RedactEmail: true,
		RedactPhone: false,
	})
	d2 := NewRegexDetector(RegexDetectorConfig{
		RedactEmail: false,
		RedactPhone: true,
	})

	comp := NewCompositeDetector(d1, d2)

	text := "Email: a@b.com, Phone: 123-456-7890"
	matches := comp.Detect(text)

	// Should find both email and phone
	foundEmail := false
	foundPhone := false
	for _, m := range matches {
		if m.PatternName == "email" {
			foundEmail = true
		}
		if m.PatternName == "us_phone" {
			foundPhone = true
		}
	}

	if !foundEmail {
		t.Error("expected to find email match")
	}
	if !foundPhone {
		t.Error("expected to find us_phone match")
	}
}

func TestCompositeDetector_DeduplicatesOverlapping(t *testing.T) {
	// Create two detectors that will match the same text
	d1 := NewRegexDetector(RegexDetectorConfig{
		RedactEmail: true,
		RedactPhone: true,
	})
	d2 := NewRegexDetector(RegexDetectorConfig{
		RedactEmail: true,
		RedactPhone: true,
	})

	comp := NewCompositeDetector(d1, d2)

	text := "test@example.com"
	matches := comp.Detect(text)

	// Should deduplicate overlapping matches
	if len(matches) != 1 {
		t.Errorf("want 1 deduplicated match, got %d: %+v", len(matches), matches)
	}
}

func TestCompositeDetector_EmptyInput(t *testing.T) {
	d := NewRegexDetector(RegexDetectorConfig{RedactEmail: true})
	comp := NewCompositeDetector(d)

	matches := comp.Detect("")
	if len(matches) != 0 {
		t.Errorf("want 0 matches for empty input, got %d", len(matches))
	}
}

func TestCompositeDetector_NoDetectors(t *testing.T) {
	comp := NewCompositeDetector()

	matches := comp.Detect("test@example.com")
	if matches != nil {
		t.Errorf("want nil matches with no detectors, got %v", matches)
	}
}
