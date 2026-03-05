package checks

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/config"
)

// ApprovalCheck validates the approval system configuration.
type ApprovalCheck struct{}

// Name returns the check name.
func (c *ApprovalCheck) Name() string {
	return "Approval System"
}

// Run checks approval system configuration.
func (c *ApprovalCheck) Run(_ context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	if !cfg.Security.Interceptor.Enabled {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "Security interceptor is not enabled (approval system inactive)",
		}
	}

	policy := string(cfg.Security.Interceptor.ApprovalPolicy)
	if policy == "" {
		policy = "dangerous"
	}

	validPolicies := map[string]bool{
		"dangerous":  true,
		"all":        true,
		"configured": true,
		"none":       true,
	}

	if !validPolicies[policy] {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: fmt.Sprintf("Unknown approval policy: %q", policy),
			Details: "Valid policies: dangerous, all, configured, none",
		}
	}

	if policy == "none" {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: "Approval policy is 'none' (all tools auto-approved)",
			Details: "Consider using 'dangerous' or 'all' for better security.",
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: fmt.Sprintf("Approval system active (policy=%s, pii_redaction=%v)", policy, cfg.Security.Interceptor.RedactPII),
	}
}

// Fix delegates to Run as automatic fixing is not supported.
func (c *ApprovalCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
