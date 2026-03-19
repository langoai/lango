package runledger

import (
	"fmt"
	"os/exec"
	"strings"
)

// WorkspaceManager handles git worktree isolation for coding steps.
// Coding-only, fail-closed: if worktree creation fails, the step is not executed.
type WorkspaceManager struct{}

// NewWorkspaceManager creates a new WorkspaceManager.
func NewWorkspaceManager() *WorkspaceManager {
	return &WorkspaceManager{}
}

// NeedsIsolation returns true if the step requires workspace isolation
// based on its validator type.
func NeedsIsolation(step *Step) bool {
	switch step.Validator.Type {
	case ValidatorFileChanged, ValidatorBuildPass, ValidatorTestPass:
		return true
	}
	return false
}

// CheckDirtyTree returns an error if the git working tree has uncommitted changes.
func (m *WorkspaceManager) CheckDirtyTree() error {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	if len(strings.TrimSpace(string(output))) > 0 {
		return fmt.Errorf("working tree has uncommitted changes — stash or commit before proceeding")
	}
	return nil
}

// CreateWorktree creates a git worktree at the given path for isolated execution.
// The branch name is derived from the run ID and step ID.
func (m *WorkspaceManager) CreateWorktree(path, runID, stepID string) error {
	branch := fmt.Sprintf("runledger/%s/%s", runID, stepID)
	cmd := exec.Command("git", "worktree", "add", "-b", branch, path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("create worktree: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

// RemoveWorktree removes a git worktree.
func (m *WorkspaceManager) RemoveWorktree(path string) error {
	cmd := exec.Command("git", "worktree", "remove", "--force", path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("remove worktree: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

// ExportPatch generates a patch file from a worktree using git format-patch.
// Auto-merge is intentionally forbidden — only git am is allowed.
func (m *WorkspaceManager) ExportPatch(worktreePath, outputPath string) error {
	cmd := exec.Command("git", "-C", worktreePath, "format-patch", "HEAD~1", "-o", outputPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("format-patch: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

// ApplyPatch applies a patch to the main tree using git am.
func (m *WorkspaceManager) ApplyPatch(patchPath string) error {
	cmd := exec.Command("git", "am", patchPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git am: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}
