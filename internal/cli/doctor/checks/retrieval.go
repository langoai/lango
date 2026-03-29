package checks

import (
	"context"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/config"
)

// RetrievalCheck validates retrieval coordinator configuration.
type RetrievalCheck struct{}

// Name returns the check name.
func (c *RetrievalCheck) Name() string {
	return "Retrieval Coordinator"
}

// Run checks retrieval coordinator configuration validity.
func (c *RetrievalCheck) Run(_ context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	// If retrieval is disabled AND autoAdjust is disabled, skip.
	if !cfg.Retrieval.Enabled && !cfg.Retrieval.AutoAdjust.Enabled {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "Retrieval coordinator disabled",
		}
	}

	var warnings []string

	// retrieval.enabled + !knowledge.enabled → dependency missing.
	if cfg.Retrieval.Enabled && !cfg.Knowledge.Enabled {
		warnings = append(warnings, "retrieval.enabled requires knowledge.enabled")
	}

	aa := cfg.Retrieval.AutoAdjust
	if aa.Enabled {
		// Mode validation.
		if aa.Mode != "shadow" && aa.Mode != "active" {
			return Result{
				Name:    c.Name(),
				Status:  StatusFail,
				Message: fmt.Sprintf("retrieval.autoAdjust.mode must be \"shadow\" or \"active\", got %q", aa.Mode),
			}
		}

		// minScore >= maxScore.
		if aa.MinScore >= aa.MaxScore {
			return Result{
				Name:    c.Name(),
				Status:  StatusFail,
				Message: fmt.Sprintf("retrieval.autoAdjust.minScore (%.2f) must be less than maxScore (%.2f)", aa.MinScore, aa.MaxScore),
			}
		}

		if aa.BoostDelta <= 0 || aa.BoostDelta > 1.0 {
			warnings = append(warnings,
				fmt.Sprintf("retrieval.autoAdjust.boostDelta (%.2f) should be in (0.0, 1.0]", aa.BoostDelta))
		}
		if aa.DecayDelta <= 0 || aa.DecayDelta > 1.0 {
			warnings = append(warnings,
				fmt.Sprintf("retrieval.autoAdjust.decayDelta (%.2f) should be in (0.0, 1.0]", aa.DecayDelta))
		}

		if aa.WarmupTurns < 0 {
			warnings = append(warnings,
				fmt.Sprintf("retrieval.autoAdjust.warmupTurns (%d) should not be negative", aa.WarmupTurns))
		}

		if aa.Mode == "active" {
			warnings = append(warnings,
				"retrieval.autoAdjust.mode is \"active\" — relevance scores will be mutated in production")
		}
	}

	if len(warnings) > 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: fmt.Sprintf("%d warning(s): %s", len(warnings), warnings[0]),
			Details: strings.Join(warnings, "; "),
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "Retrieval configuration valid",
	}
}

// Fix delegates to Run as automatic fixing is not supported.
func (c *RetrievalCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
