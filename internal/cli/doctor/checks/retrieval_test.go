package checks

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/config"
)

func TestRetrievalCheck_Run(t *testing.T) {
	tests := []struct {
		give       string
		setup      func(*config.Config)
		wantStatus Status
	}{
		{
			give:       "disabled returns skip",
			setup:      func(cfg *config.Config) {},
			wantStatus: StatusSkip,
		},
		{
			give: "enabled with knowledge returns pass",
			setup: func(cfg *config.Config) {
				cfg.Retrieval.Enabled = true
				cfg.Knowledge.Enabled = true
			},
			wantStatus: StatusPass,
		},
		{
			give: "enabled without knowledge returns warn",
			setup: func(cfg *config.Config) {
				cfg.Retrieval.Enabled = true
				cfg.Knowledge.Enabled = false
			},
			wantStatus: StatusWarn,
		},
		{
			give: "autoAdjust invalid mode returns fail",
			setup: func(cfg *config.Config) {
				cfg.Retrieval.AutoAdjust.Enabled = true
				cfg.Retrieval.AutoAdjust.Mode = "invalid"
				cfg.Retrieval.AutoAdjust.MinScore = 0.1
				cfg.Retrieval.AutoAdjust.MaxScore = 5.0
			},
			wantStatus: StatusFail,
		},
		{
			give: "autoAdjust minScore >= maxScore returns fail",
			setup: func(cfg *config.Config) {
				cfg.Retrieval.AutoAdjust.Enabled = true
				cfg.Retrieval.AutoAdjust.Mode = "shadow"
				cfg.Retrieval.AutoAdjust.MinScore = 5.0
				cfg.Retrieval.AutoAdjust.MaxScore = 5.0
			},
			wantStatus: StatusFail,
		},
		{
			give: "autoAdjust active mode returns warn",
			setup: func(cfg *config.Config) {
				cfg.Knowledge.Enabled = true
				cfg.Retrieval.AutoAdjust.Enabled = true
				cfg.Retrieval.AutoAdjust.Mode = "active"
				cfg.Retrieval.AutoAdjust.BoostDelta = 0.05
				cfg.Retrieval.AutoAdjust.DecayDelta = 0.01
				cfg.Retrieval.AutoAdjust.MinScore = 0.1
				cfg.Retrieval.AutoAdjust.MaxScore = 5.0
			},
			wantStatus: StatusWarn,
		},
		{
			give: "autoAdjust shadow mode valid returns pass",
			setup: func(cfg *config.Config) {
				cfg.Knowledge.Enabled = true
				cfg.Retrieval.AutoAdjust.Enabled = true
				cfg.Retrieval.AutoAdjust.Mode = "shadow"
				cfg.Retrieval.AutoAdjust.BoostDelta = 0.05
				cfg.Retrieval.AutoAdjust.DecayDelta = 0.01
				cfg.Retrieval.AutoAdjust.MinScore = 0.1
				cfg.Retrieval.AutoAdjust.MaxScore = 5.0
				cfg.Retrieval.AutoAdjust.WarmupTurns = 50
			},
			wantStatus: StatusPass,
		},
		{
			give:       "nil config returns skip",
			setup:      nil,
			wantStatus: StatusSkip,
		},
	}

	check := &RetrievalCheck{}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			var cfg *config.Config
			if tt.setup != nil {
				cfg = config.DefaultConfig()
				tt.setup(cfg)
			}
			result := check.Run(context.Background(), cfg)
			if result.Status != tt.wantStatus {
				t.Errorf("want status %v, got %v: %s", tt.wantStatus, result.Status, result.Message)
			}
		})
	}
}
