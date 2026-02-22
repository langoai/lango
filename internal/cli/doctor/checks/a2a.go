package checks

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/langoai/lango/internal/config"
)

// A2ACheck validates A2A protocol configuration.
type A2ACheck struct{}

// Name returns the check name.
func (c *A2ACheck) Name() string {
	return "A2A Protocol"
}

// Run checks A2A configuration validity.
func (c *A2ACheck) Run(_ context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	if !cfg.A2A.Enabled {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "A2A protocol is not enabled",
		}
	}

	var issues []string
	status := StatusPass

	if cfg.A2A.BaseURL == "" {
		issues = append(issues, "a2a.baseUrl is required when A2A is enabled")
		status = StatusFail
	}

	if cfg.A2A.AgentName == "" {
		issues = append(issues, "a2a.agentName is required when A2A is enabled")
		status = StatusFail
	}

	// Check remote agent connectivity (warn only, don't fail).
	for _, ra := range cfg.A2A.RemoteAgents {
		if ra.AgentCardURL == "" {
			issues = append(issues, fmt.Sprintf("remote agent %q has no agent card URL", ra.Name))
			if status < StatusWarn {
				status = StatusWarn
			}
			continue
		}

		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get(ra.AgentCardURL)
		if err != nil {
			issues = append(issues, fmt.Sprintf("remote agent %q unreachable: %v", ra.Name, err))
			if status < StatusWarn {
				status = StatusWarn
			}
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			issues = append(issues, fmt.Sprintf("remote agent %q returned HTTP %d", ra.Name, resp.StatusCode))
			if status < StatusWarn {
				status = StatusWarn
			}
		}
	}

	if len(issues) == 0 {
		msg := fmt.Sprintf("A2A configured (name=%s, url=%s, remotes=%d)",
			cfg.A2A.AgentName, cfg.A2A.BaseURL, len(cfg.A2A.RemoteAgents))
		return Result{Name: c.Name(), Status: StatusPass, Message: msg}
	}

	message := "A2A issues:\n"
	for _, issue := range issues {
		message += fmt.Sprintf("- %s\n", issue)
	}
	return Result{Name: c.Name(), Status: status, Message: message}
}

// Fix delegates to Run as automatic fixing is not supported.
func (c *A2ACheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
