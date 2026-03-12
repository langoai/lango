package gitbundle

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"
)

// TaskBranchPrefix is the ref prefix for per-task branches.
const TaskBranchPrefix = "task/"

// BranchInfo describes a branch in the workspace repository.
type BranchInfo struct {
	Name       string    `json:"name"`
	CommitHash string    `json:"commitHash"`
	IsHead     bool      `json:"isHead"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// MergeResult describes the outcome of a branch merge.
type MergeResult struct {
	Success       bool     `json:"success"`
	MergeCommit   string   `json:"mergeCommit,omitempty"`
	ConflictFiles []string `json:"conflictFiles,omitempty"`
	Message       string   `json:"message"`
}

// CreateTaskBranch creates a task/{taskID} branch in the workspace bare repo.
// If baseBranch is empty, it defaults to the current HEAD.
// The operation is idempotent — if the branch already exists, it returns nil.
func (s *Service) CreateTaskBranch(ctx context.Context, workspaceID, taskID, baseBranch string) error {
	if taskID == "" {
		return fmt.Errorf("empty task ID")
	}

	repoPath := s.store.RepoPath(workspaceID)
	branchName := TaskBranchPrefix + taskID

	// Check if branch already exists.
	checkCmd := exec.CommandContext(ctx, "git", "rev-parse", "--verify", "refs/heads/"+branchName)
	checkCmd.Dir = repoPath
	if err := checkCmd.Run(); err == nil {
		// Branch already exists — idempotent.
		return nil
	}

	args := []string{"branch", branchName}
	if baseBranch != "" {
		args = append(args, baseBranch)
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repoPath

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("create branch %s: %s: %w", branchName, stderr.String(), err)
	}

	s.logger.Info("created task branch",
		zap.String("workspace", workspaceID),
		zap.String("branch", branchName))
	return nil
}

// MergeTaskBranch merges task/{taskID} into targetBranch using git merge-tree.
// This works on bare repos without a working tree by using merge-tree --write-tree,
// commit-tree, and update-ref.
func (s *Service) MergeTaskBranch(ctx context.Context, workspaceID, taskID, targetBranch string) (*MergeResult, error) {
	if taskID == "" {
		return nil, fmt.Errorf("empty task ID")
	}
	if targetBranch == "" {
		targetBranch = "main"
	}

	repoPath := s.store.RepoPath(workspaceID)
	sourceBranch := TaskBranchPrefix + taskID

	// Resolve branch refs.
	sourceHash, err := s.resolveRef(ctx, repoPath, sourceBranch)
	if err != nil {
		return nil, fmt.Errorf("resolve source branch %s: %w", sourceBranch, err)
	}
	targetHash, err := s.resolveRef(ctx, repoPath, targetBranch)
	if err != nil {
		return nil, fmt.Errorf("resolve target branch %s: %w", targetBranch, err)
	}

	// git merge-tree --write-tree <target> <source>
	mergeCmd := exec.CommandContext(ctx, "git", "merge-tree", "--write-tree", targetHash, sourceHash)
	mergeCmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	mergeCmd.Stdout = &stdout
	mergeCmd.Stderr = &stderr

	if err := mergeCmd.Run(); err != nil {
		// Non-zero exit indicates conflicts.
		conflicts := parseConflictFiles(stdout.String())
		return &MergeResult{
			Success:       false,
			ConflictFiles: conflicts,
			Message:       fmt.Sprintf("merge conflicts between %s and %s", sourceBranch, targetBranch),
		}, nil
	}

	// First line of stdout is the tree hash.
	treeHash := strings.TrimSpace(strings.Split(stdout.String(), "\n")[0])
	if treeHash == "" {
		return nil, fmt.Errorf("merge-tree produced empty tree hash")
	}

	// Create merge commit using commit-tree.
	commitMsg := fmt.Sprintf("Merge %s into %s", sourceBranch, targetBranch)
	commitCmd := exec.CommandContext(ctx, "git", "commit-tree", treeHash,
		"-p", targetHash, "-p", sourceHash, "-m", commitMsg)
	commitCmd.Dir = repoPath

	var commitOut bytes.Buffer
	commitCmd.Stdout = &commitOut
	commitCmd.Stderr = &stderr

	if err := commitCmd.Run(); err != nil {
		return nil, fmt.Errorf("commit-tree: %s: %w", stderr.String(), err)
	}

	mergeCommitHash := strings.TrimSpace(commitOut.String())

	// Update target ref to point to merge commit.
	updateCmd := exec.CommandContext(ctx, "git", "update-ref", "refs/heads/"+targetBranch, mergeCommitHash)
	updateCmd.Dir = repoPath

	if err := updateCmd.Run(); err != nil {
		return nil, fmt.Errorf("update-ref %s: %w", targetBranch, err)
	}

	s.logger.Info("merged task branch",
		zap.String("workspace", workspaceID),
		zap.String("source", sourceBranch),
		zap.String("target", targetBranch),
		zap.String("mergeCommit", mergeCommitHash))

	return &MergeResult{
		Success:     true,
		MergeCommit: mergeCommitHash,
		Message:     commitMsg,
	}, nil
}

// ListBranches returns all branches in the workspace repository.
func (s *Service) ListBranches(ctx context.Context, workspaceID string) ([]BranchInfo, error) {
	repoPath := s.store.RepoPath(workspaceID)

	cmd := exec.CommandContext(ctx, "git", "for-each-ref",
		"--format=%(refname:short) %(objectname) %(HEAD) %(creatordate:iso8601)",
		"refs/heads/")
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("for-each-ref: %s: %w", stderr.String(), err)
	}

	var branches []BranchInfo
	for _, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		info := BranchInfo{
			Name:       parts[0],
			CommitHash: parts[1],
			IsHead:     parts[2] == "*",
		}

		// Parse date if available (remaining parts after first 3).
		if len(parts) > 3 {
			dateStr := strings.Join(parts[3:], " ")
			if t, err := time.Parse("2006-01-02 15:04:05 -0700", dateStr); err == nil {
				info.UpdatedAt = t
			}
		}

		branches = append(branches, info)
	}

	return branches, nil
}

// DeleteTaskBranch deletes the task/{taskID} branch from the workspace.
// The operation is idempotent — if the branch doesn't exist, it returns nil.
func (s *Service) DeleteTaskBranch(ctx context.Context, workspaceID, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("empty task ID")
	}

	repoPath := s.store.RepoPath(workspaceID)
	branchName := TaskBranchPrefix + taskID

	// Check if branch exists first for idempotency.
	checkCmd := exec.CommandContext(ctx, "git", "rev-parse", "--verify", "refs/heads/"+branchName)
	checkCmd.Dir = repoPath
	if err := checkCmd.Run(); err != nil {
		// Branch doesn't exist — idempotent.
		return nil
	}

	cmd := exec.CommandContext(ctx, "git", "branch", "-D", branchName)
	cmd.Dir = repoPath

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("delete branch %s: %s: %w", branchName, stderr.String(), err)
	}

	s.logger.Info("deleted task branch",
		zap.String("workspace", workspaceID),
		zap.String("branch", branchName))
	return nil
}

// resolveRef resolves a branch name to its commit hash.
func (s *Service) resolveRef(ctx context.Context, repoPath, branchName string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "refs/heads/"+branchName)
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("rev-parse %s: %s: %w", branchName, stderr.String(), err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// parseConflictFiles extracts conflicting file paths from git merge-tree output.
func parseConflictFiles(output string) []string {
	var files []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		// merge-tree conflict output contains lines with file paths after "CONFLICT" markers.
		if strings.HasPrefix(line, "CONFLICT") {
			// Extract filename from patterns like "CONFLICT (content): Merge conflict in <file>"
			if idx := strings.Index(line, "Merge conflict in "); idx >= 0 {
				file := strings.TrimSpace(line[idx+len("Merge conflict in "):])
				if file != "" {
					files = append(files, file)
				}
			}
		}
	}
	return files
}
