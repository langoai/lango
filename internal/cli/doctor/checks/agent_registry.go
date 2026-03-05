package checks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/langoai/lango/internal/config"
)

// AgentRegistryCheck validates agent registry configuration.
type AgentRegistryCheck struct{}

// Name returns the check name.
func (c *AgentRegistryCheck) Name() string {
	return "Agent Registry"
}

// Run checks agent registry configuration and agents directory.
func (c *AgentRegistryCheck) Run(_ context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	if !cfg.Agent.MultiAgent {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "Multi-agent mode is not enabled",
		}
	}

	agentsDir := cfg.Agent.AgentsDir
	if agentsDir == "" {
		return Result{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "Multi-agent enabled (no custom agents directory configured)",
			Details: "Set agent.agentsDir to load custom AGENT.md definitions.",
		}
	}

	// Expand ~ in path.
	if len(agentsDir) > 1 && agentsDir[:2] == "~/" {
		if home, err := os.UserHomeDir(); err == nil {
			agentsDir = filepath.Join(home, agentsDir[2:])
		}
	}

	info, err := os.Stat(agentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return Result{
				Name:    c.Name(),
				Status:  StatusWarn,
				Message: fmt.Sprintf("Agents directory does not exist: %s", agentsDir),
				Details: "Create the directory and add AGENT.md files to define custom agents.",
			}
		}
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: fmt.Sprintf("access agents directory: %v", err),
		}
	}

	if !info.IsDir() {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: fmt.Sprintf("agent.agentsDir is not a directory: %s", agentsDir),
		}
	}

	// Count subdirectories (potential agent definitions).
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: fmt.Sprintf("read agents directory: %v", err),
		}
	}

	agentCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			agentMD := filepath.Join(agentsDir, entry.Name(), "AGENT.md")
			if _, statErr := os.Stat(agentMD); statErr == nil {
				agentCount++
			}
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: fmt.Sprintf("Agent registry healthy (%d user-defined agents in %s)", agentCount, cfg.Agent.AgentsDir),
	}
}

// Fix delegates to Run as automatic fixing is not supported.
func (c *AgentRegistryCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
