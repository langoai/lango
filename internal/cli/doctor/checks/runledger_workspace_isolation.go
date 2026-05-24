package checks

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/langoai/lango/internal/config"
)

// RunLedgerWorkspaceIsolationCheck validates RunLedger workspace isolation
// configuration and environment health.
// Distinct from P2P WorkspaceCheck — this covers RunLedger coding-step isolation.
type RunLedgerWorkspaceIsolationCheck struct{}

func (c *RunLedgerWorkspaceIsolationCheck) Name() string {
	return "RunLedger Workspace Isolation"
}

func (c *RunLedgerWorkspaceIsolationCheck) Run(_ context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	if !cfg.RunLedger.WorkspaceIsolation {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "Workspace isolation is not enabled (runLedger.workspaceIsolation = false)",
		}
	}

	if _, err := exec.LookPath("git"); err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: "git not found in PATH — required for workspace isolation",
		}
	}

	worktrees, stale, wtErr := listRunLedgerWorktrees()
	if wtErr != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: "Could not inspect git worktrees",
			Details: wtErr.Error(),
		}
	}

	var details []string
	details = append(details, fmt.Sprintf("Active runledger worktrees: %d", len(worktrees)))
	if len(stale) > 0 {
		details = append(details, "Stale worktrees (path missing on disk):")
		for _, s := range stale {
			details = append(details, fmt.Sprintf("  - %s", s))
		}
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: fmt.Sprintf("Workspace isolation enabled — %d stale worktree(s) detected", len(stale)),
			Details: strings.Join(details, "\n"),
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: fmt.Sprintf("Workspace isolation enabled (%d active worktrees, no stale)", len(worktrees)),
		Details: strings.Join(details, "\n"),
	}
}

func (c *RunLedgerWorkspaceIsolationCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}

func listRunLedgerWorktrees() (active []string, stale []string, err error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, cmdErr := cmd.Output()
	if cmdErr != nil {
		return nil, nil, fmt.Errorf("git worktree list: %w", cmdErr)
	}

	classifyPath := func(p string) {
		if !strings.Contains(p, "runledger") {
			return
		}
		if _, err := os.Stat(p); os.IsNotExist(err) {
			stale = append(stale, p)
		} else {
			active = append(active, p)
		}
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	var currentPath string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "worktree ") {
			if currentPath != "" {
				classifyPath(currentPath)
			}
			currentPath = strings.TrimPrefix(line, "worktree ")
			continue
		}
		if line == "" && currentPath != "" {
			classifyPath(currentPath)
			currentPath = ""
		}
	}
	if currentPath != "" {
		classifyPath(currentPath)
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return active, stale, fmt.Errorf("parse worktree list: %w", scanErr)
	}
	return active, stale, nil
}
