package runledger

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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
// The error includes a count of changed files and a suggested remediation command.
func (m *WorkspaceManager) CheckDirtyTree() error {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	trimmed := strings.TrimSpace(string(output))
	if len(trimmed) == 0 {
		return nil
	}
	lines := strings.Split(trimmed, "\n")
	return fmt.Errorf(
		"working tree has %d uncommitted change(s) — workspace isolation requires a clean tree\n"+
			"  suggestion: git stash push -m \"lango-workspace-isolation\"",
		len(lines),
	)
}

// CreateWorktree creates a git worktree at the given path for isolated execution.
func (m *WorkspaceManager) CreateWorktree(path, branch string) error {
	cmd := exec.Command("git", "worktree", "add", "-b", branch, path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("create worktree: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

// DeleteBranch deletes a worktree branch after cleanup.
func (m *WorkspaceManager) DeleteBranch(branch string) error {
	cmd := exec.Command("git", "branch", "-D", branch)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("delete branch: %s: %w", strings.TrimSpace(string(output)), err)
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

// PrepareStepWorkspace handles the full workspace lifecycle for a coding step:
// 1. Check if isolation needed (based on validator type)
// 2. If not needed, return (WorkDir stays empty = current dir)
// 3. Check dirty tree -> fail if dirty
// 4. Create worktree -> set step.Validator.WorkDir
// Returns a cleanup function that must be deferred.
func (m *WorkspaceManager) PrepareStepWorkspace(step *Step, runID string) (cleanup func(), err error) {
	if !NeedsIsolation(step) {
		return func() {}, nil
	}

	if err := m.CheckDirtyTree(); err != nil {
		return nil, fmt.Errorf("workspace isolation: %w", err)
	}

	suffix := fmt.Sprintf("%s-%d", step.StepID, time.Now().UnixNano())
	path := filepath.Join(os.TempDir(), "runledger", runID, suffix)
	branch := fmt.Sprintf("runledger/%s/%s", runID, suffix)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create workspace parent: %w", err)
	}
	if err := m.CreateWorktree(path, branch); err != nil {
		return nil, fmt.Errorf("create worktree: %w", err)
	}

	step.Validator.WorkDir = path

	return func() {
		_ = m.RemoveWorktree(path)
		_ = m.DeleteBranch(branch)
		step.Validator.WorkDir = ""
	}, nil
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
// On failure, the error includes rollback instructions.
func (m *WorkspaceManager) ApplyPatch(patchPath string) error {
	cmd := exec.Command("git", "am", patchPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"git am: %s: %w\n"+
				"  to abort and return to the previous state: git am --abort",
			strings.TrimSpace(string(output)), err,
		)
	}
	return nil
}
