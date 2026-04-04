package approval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassifyTier(t *testing.T) {
	tests := []struct {
		give        string
		safetyLevel string
		category    string
		activity    string
		want        DisplayTier
	}{
		{
			give:        "dangerous + filesystem → fullscreen",
			safetyLevel: "dangerous",
			category:    "filesystem",
			activity:    "write",
			want:        TierFullscreen,
		},
		{
			give:        "dangerous + automation → fullscreen",
			safetyLevel: "dangerous",
			category:    "automation",
			activity:    "execute",
			want:        TierFullscreen,
		},
		{
			give:        "dangerous + execute activity → fullscreen",
			safetyLevel: "dangerous",
			category:    "browser",
			activity:    "execute",
			want:        TierFullscreen,
		},
		{
			give:        "dangerous + write activity → fullscreen",
			safetyLevel: "dangerous",
			category:    "crypto",
			activity:    "write",
			want:        TierFullscreen,
		},
		{
			give:        "dangerous + read activity → inline",
			safetyLevel: "dangerous",
			category:    "browser",
			activity:    "read",
			want:        TierInline,
		},
		{
			give:        "moderate + filesystem → inline",
			safetyLevel: "moderate",
			category:    "filesystem",
			activity:    "write",
			want:        TierInline,
		},
		{
			give:        "safe + any → inline",
			safetyLevel: "safe",
			category:    "filesystem",
			activity:    "execute",
			want:        TierInline,
		},
		{
			give:        "empty safety level → inline",
			safetyLevel: "",
			category:    "filesystem",
			activity:    "write",
			want:        TierInline,
		},
		{
			give:        "dangerous + unknown category + query → inline",
			safetyLevel: "dangerous",
			category:    "knowledge",
			activity:    "query",
			want:        TierInline,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := ClassifyTier(tt.safetyLevel, tt.category, tt.activity)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestComputeRisk(t *testing.T) {
	tests := []struct {
		give        string
		safetyLevel string
		category    string
		wantLevel   string
	}{
		{
			give:        "dangerous + filesystem → critical",
			safetyLevel: "dangerous",
			category:    "filesystem",
			wantLevel:   "critical",
		},
		{
			give:        "dangerous + automation → critical",
			safetyLevel: "dangerous",
			category:    "automation",
			wantLevel:   "critical",
		},
		{
			give:        "dangerous + other → high",
			safetyLevel: "dangerous",
			category:    "browser",
			wantLevel:   "high",
		},
		{
			give:        "moderate → moderate",
			safetyLevel: "moderate",
			category:    "filesystem",
			wantLevel:   "moderate",
		},
		{
			give:        "safe → low",
			safetyLevel: "safe",
			category:    "",
			wantLevel:   "low",
		},
		{
			give:        "empty → low",
			safetyLevel: "",
			category:    "",
			wantLevel:   "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := ComputeRisk(tt.safetyLevel, tt.category)
			assert.Equal(t, tt.wantLevel, got.Level)
			assert.NotEmpty(t, got.Label)
		})
	}
}

func TestNewViewModel(t *testing.T) {
	req := ApprovalRequest{
		ID:          "req-123",
		ToolName:    "exec",
		SafetyLevel: "dangerous",
		Category:    "automation",
		Activity:    "execute",
	}

	vm := NewViewModel(req)

	assert.Equal(t, TierFullscreen, vm.Tier)
	assert.Equal(t, "critical", vm.Risk.Level)
	assert.Equal(t, req, vm.Request)
	assert.Empty(t, vm.DiffContent)
}

func TestNewViewModel_InlineTier(t *testing.T) {
	req := ApprovalRequest{
		ID:          "req-456",
		ToolName:    "browser_search",
		SafetyLevel: "moderate",
		Category:    "browser",
		Activity:    "read",
	}

	vm := NewViewModel(req)

	assert.Equal(t, TierInline, vm.Tier)
	assert.Equal(t, "moderate", vm.Risk.Level)
}
