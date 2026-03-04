package checks

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/config"
)

// ToolHooksCheck validates tool hooks configuration.
type ToolHooksCheck struct{}

// Name returns the check name.
func (c *ToolHooksCheck) Name() string {
	return "Tool Hooks"
}

// Run checks tool hooks configuration.
func (c *ToolHooksCheck) Run(_ context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	var hooks []string

	// Learning hook depends on knowledge being enabled.
	if cfg.Knowledge.Enabled {
		hooks = append(hooks, "learning_observer", "knowledge_saver")
	}

	// Approval hook depends on interceptor.
	if cfg.Security.Interceptor.Enabled {
		hooks = append(hooks, "approval_gate", "security_filter")
	}

	if len(hooks) == 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "No tool hooks active (knowledge and interceptor disabled)",
			Details: "Enable knowledge.enabled or security.interceptor.enabled to activate tool hooks.",
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: fmt.Sprintf("%d tool hooks configured: %v", len(hooks), hooks),
	}
}

// Fix delegates to Run as automatic fixing is not supported.
func (c *ToolHooksCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
