package checks

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/langoai/lango/internal/config"
)

// WorkspaceCheck validates P2P workspace configuration and dependencies.
type WorkspaceCheck struct{}

// Name returns the check name.
func (c *WorkspaceCheck) Name() string {
	return "P2P Workspaces"
}

// Run checks workspace configuration validity.
func (c *WorkspaceCheck) Run(_ context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	if !cfg.P2P.Workspace.Enabled {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "P2P workspaces are not enabled",
		}
	}

	var issues []string
	status := StatusPass

	// Check git binary availability.
	if _, err := exec.LookPath("git"); err != nil {
		issues = append(issues, "git binary not found in PATH (required for git bundle operations)")
		if status < StatusWarn {
			status = StatusWarn
		}
	}

	// Resolve data directory.
	dataDir := cfg.P2P.Workspace.DataDir
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".lango", "workspaces")
	}

	// Check data directory exists.
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		issues = append(issues, fmt.Sprintf("data directory %s does not exist", dataDir))
		if status < StatusWarn {
			status = StatusWarn
		}
		return Result{
			Name:      c.Name(),
			Status:    status,
			Message:   fmt.Sprintf("Workspace data dir missing: %s", dataDir),
			Details:   joinIssues(issues),
			Fixable:   true,
			FixAction: fmt.Sprintf("Create directory %s", dataDir),
		}
	}

	if len(issues) == 0 {
		msg := fmt.Sprintf("Workspaces configured (dir=%s, max=%d, chronicler=%v, contributions=%v)",
			dataDir, cfg.P2P.Workspace.MaxWorkspaces,
			cfg.P2P.Workspace.ChroniclerEnabled, cfg.P2P.Workspace.ContributionTracking)
		return Result{Name: c.Name(), Status: StatusPass, Message: msg}
	}

	return Result{
		Name:    c.Name(),
		Status:  status,
		Message: "Workspace issues:\n" + joinIssues(issues),
	}
}

// Fix attempts to create the workspace data directory.
func (c *WorkspaceCheck) Fix(_ context.Context, cfg *config.Config) Result {
	if cfg == nil || !cfg.P2P.Workspace.Enabled {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Not applicable"}
	}

	dataDir := cfg.P2P.Workspace.DataDir
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".lango", "workspaces")
	}

	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: fmt.Sprintf("create data dir: %v", err),
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: fmt.Sprintf("Created workspace data directory: %s", dataDir),
	}
}

func joinIssues(issues []string) string {
	result := ""
	for _, issue := range issues {
		result += fmt.Sprintf("- %s\n", issue)
	}
	return result
}
