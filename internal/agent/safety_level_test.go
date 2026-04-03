package agent

import (
	"testing"
)

func TestSafetyLevel_String(t *testing.T) {
	tests := []struct {
		give SafetyLevel
		want string
	}{
		{give: SafetyLevelSafe, want: "safe"},
		{give: SafetyLevelModerate, want: "moderate"},
		{give: SafetyLevelDangerous, want: "dangerous"},
		{give: 0, want: "dangerous"},  // zero value → fail-safe
		{give: 99, want: "dangerous"}, // unknown → fail-safe
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.give.String()
			if got != tt.want {
				t.Errorf("SafetyLevel(%d).String() = %q, want %q", tt.give, got, tt.want)
			}
		})
	}
}

func TestSafetyLevel_IsDangerous(t *testing.T) {
	tests := []struct {
		give SafetyLevel
		want bool
	}{
		{give: SafetyLevelSafe, want: false},
		{give: SafetyLevelModerate, want: false},
		{give: SafetyLevelDangerous, want: true},
		{give: 0, want: true}, // zero value → dangerous
	}

	for _, tt := range tests {
		t.Run(tt.give.String(), func(t *testing.T) {
			got := tt.give.IsDangerous()
			if got != tt.want {
				t.Errorf("SafetyLevel(%d).IsDangerous() = %v, want %v", tt.give, got, tt.want)
			}
		})
	}
}

func TestParseSafetyLevel(t *testing.T) {
	tests := []struct {
		give     string
		wantLvl  SafetyLevel
		wantOK   bool
	}{
		{give: "safe", wantLvl: SafetyLevelSafe, wantOK: true},
		{give: "moderate", wantLvl: SafetyLevelModerate, wantOK: true},
		{give: "dangerous", wantLvl: SafetyLevelDangerous, wantOK: true},
		{give: "", wantLvl: SafetyLevelDangerous, wantOK: false},
		{give: "unknown", wantLvl: SafetyLevelDangerous, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			lvl, ok := ParseSafetyLevel(tt.give)
			if lvl != tt.wantLvl || ok != tt.wantOK {
				t.Errorf("ParseSafetyLevel(%q) = (%d, %v), want (%d, %v)", tt.give, lvl, ok, tt.wantLvl, tt.wantOK)
			}
		})
	}
}
