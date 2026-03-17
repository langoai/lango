package gitbundle

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"errors"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.uber.org/zap"
)

var errLimitReached = errors.New("limit reached")

// ErrEmptyRepo indicates the workspace repository has no commits.
var ErrEmptyRepo = errors.New("empty repository")

// ErrMissingPrerequisite indicates the bundle requires commits not present in the repo.
var ErrMissingPrerequisite = errors.New("missing prerequisite commits")

// runGit executes a git command in the given repo directory and returns stdout.
func runGit(ctx context.Context, repoPath string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %s: %w", args[0], stderr.String(), err)
	}
	return stdout.String(), nil
}

// CommitInfo represents a summary of a git commit.
type CommitInfo struct {
	Hash      string    `json:"hash"`
	Message   string    `json:"message"`
	Author    string    `json:"author"`
	Timestamp time.Time `json:"timestamp"`
}

// Service provides git bundle operations for workspace repositories.
type Service struct {
	store  *BareRepoStore
	logger *zap.Logger
}

// NewService creates a new git bundle service.
func NewService(store *BareRepoStore, logger *zap.Logger) *Service {
	return &Service{
		store:  store,
		logger: logger,
	}
}

// Init initializes a bare repository for a workspace.
func (s *Service) Init(ctx context.Context, workspaceID string) error {
	return s.store.Init(workspaceID)
}

// CreateBundle creates a git bundle containing all refs in the workspace repo.
// Returns the bundle bytes and the HEAD commit hash.
func (s *Service) CreateBundle(ctx context.Context, workspaceID string) ([]byte, string, error) {
	repoPath := s.store.RepoPath(workspaceID)

	// Use git CLI for bundle creation since go-git doesn't support bundles natively.
	cmd := exec.CommandContext(ctx, "git", "bundle", "create", "-", "--all")
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Empty repo is not an error — just no bundle to create.
		// Check English error messages and also handle non-English locales by
		// detecting exit code 128 (git's fatal error for empty bundle).
		stderrStr := stderr.String()
		isEmptyBundle := strings.Contains(stderrStr, "empty bundle") ||
			strings.Contains(stderrStr, "Refusing to create empty bundle")
		if !isEmptyBundle {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && exitErr.ExitCode() == 128 {
				isEmptyBundle = true
			}
		}
		if isEmptyBundle {
			return nil, "", nil
		}
		return nil, "", fmt.Errorf("git bundle create: %s: %w", stderrStr, err)
	}

	// Get HEAD hash.
	repo, err := s.store.Repo(workspaceID)
	if err != nil {
		return stdout.Bytes(), "", nil
	}

	head, err := repo.Head()
	if err != nil {
		return stdout.Bytes(), "", nil
	}

	return stdout.Bytes(), head.Hash().String(), nil
}

// ApplyBundle applies a git bundle to the workspace repository.
func (s *Service) ApplyBundle(ctx context.Context, workspaceID string, bundle []byte) error {
	if err := s.store.Init(workspaceID); err != nil {
		return fmt.Errorf("init repo: %w", err)
	}

	repoPath := s.store.RepoPath(workspaceID)

	// Write bundle to a temp pipe via stdin.
	cmd := exec.CommandContext(ctx, "git", "bundle", "unbundle", "-")
	cmd.Dir = repoPath
	cmd.Stdin = bytes.NewReader(bundle)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git bundle unbundle: %s: %w", stderr.String(), err)
	}

	s.logger.Info("applied git bundle", zap.String("workspace", workspaceID))
	return nil
}

// Log returns the most recent commits from the workspace repository.
func (s *Service) Log(ctx context.Context, workspaceID string, limit int) ([]CommitInfo, error) {
	if limit <= 0 {
		limit = 20
	}

	repo, err := s.store.Repo(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}

	// Collect commits from all references (sprawling DAG, no single HEAD).
	seen := make(map[plumbing.Hash]bool)
	var commits []CommitInfo

	refs, err := repo.References()
	if err != nil {
		return nil, fmt.Errorf("list refs: %w", err)
	}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if len(commits) >= limit {
			return nil
		}

		hash := ref.Hash()
		if hash.IsZero() {
			return nil
		}

		iter, err := repo.Log(&git.LogOptions{
			From:  hash,
			Order: git.LogOrderCommitterTime,
		})
		if err != nil {
			return nil
		}

		return iter.ForEach(func(c *object.Commit) error {
			if len(commits) >= limit {
				return errLimitReached
			}
			if seen[c.Hash] {
				return nil
			}
			seen[c.Hash] = true

			commits = append(commits, CommitInfo{
				Hash:      c.Hash.String(),
				Message:   strings.TrimSpace(c.Message),
				Author:    c.Author.Name,
				Timestamp: c.Author.When,
			})
			return nil
		})
	})
	if err != nil && !errors.Is(err, errLimitReached) {
		return nil, fmt.Errorf("iterate commits: %w", err)
	}

	return commits, nil
}

// Diff returns the diff between two commits.
func (s *Service) Diff(ctx context.Context, workspaceID, from, to string) (string, error) {
	return runGit(ctx, s.store.RepoPath(workspaceID), "diff", from, to)
}

// Leaves returns the DAG leaf commits (commits with no children).
func (s *Service) Leaves(ctx context.Context, workspaceID string) ([]string, error) {
	repo, err := s.store.Repo(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}

	// A leaf is a commit that is not a parent of any other commit.
	// Collect all commits and track which are parents.
	allCommits := make(map[plumbing.Hash]bool)
	parentCommits := make(map[plumbing.Hash]bool)

	refs, err := repo.References()
	if err != nil {
		return nil, fmt.Errorf("list refs: %w", err)
	}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		hash := ref.Hash()
		if hash.IsZero() {
			return nil
		}

		iter, iterErr := repo.Log(&git.LogOptions{From: hash})
		if iterErr != nil {
			return nil
		}

		return iter.ForEach(func(c *object.Commit) error {
			allCommits[c.Hash] = true
			for _, parent := range c.ParentHashes {
				parentCommits[parent] = true
			}
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("scan commits: %w", err)
	}

	var leaves []string
	for h := range allCommits {
		if !parentCommits[h] {
			leaves = append(leaves, h.String())
		}
	}

	return leaves, nil
}

// validateCommitHash validates a 40-character lowercase hex commit hash.
func validateCommitHash(hash string) bool {
	if len(hash) != 40 {
		return false
	}
	for _, c := range hash {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// HasCommit checks if a commit exists in the workspace repository.
func (s *Service) HasCommit(ctx context.Context, workspaceID, commitHash string) (bool, error) {
	if !validateCommitHash(commitHash) {
		return false, fmt.Errorf("invalid commit hash: %s", commitHash)
	}
	repo, err := s.store.Repo(workspaceID)
	if err != nil {
		return false, fmt.Errorf("open repo: %w", err)
	}
	_, err = repo.CommitObject(plumbing.NewHash(commitHash))
	if err != nil {
		return false, nil
	}
	return true, nil
}

// CreateIncrementalBundle creates a bundle containing only commits after baseCommit.
// If baseCommit is not found, it falls back to a full bundle.
func (s *Service) CreateIncrementalBundle(ctx context.Context, workspaceID, baseCommit string) ([]byte, string, error) {
	if !validateCommitHash(baseCommit) {
		return nil, "", fmt.Errorf("invalid base commit: %s", baseCommit)
	}

	has, err := s.HasCommit(ctx, workspaceID, baseCommit)
	if err != nil {
		return nil, "", fmt.Errorf("check base commit: %w", err)
	}
	if !has {
		// Base commit not found — fallback to full bundle.
		s.logger.Info("base commit not found, falling back to full bundle",
			zap.String("workspace", workspaceID), zap.String("base", baseCommit))
		return s.CreateBundle(ctx, workspaceID)
	}

	repoPath := s.store.RepoPath(workspaceID)
	cmd := exec.CommandContext(ctx, "git", "bundle", "create", "-", baseCommit+"..HEAD")
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "empty bundle") || strings.Contains(stderrStr, "Refusing to create empty bundle") {
			return nil, "", nil
		}
		// Fallback to full bundle on any other error.
		s.logger.Info("incremental bundle failed, falling back to full bundle",
			zap.String("workspace", workspaceID), zap.String("error", stderrStr))
		return s.CreateBundle(ctx, workspaceID)
	}

	repo, err := s.store.Repo(workspaceID)
	if err != nil {
		return stdout.Bytes(), "", nil
	}
	head, err := repo.Head()
	if err != nil {
		return stdout.Bytes(), "", nil
	}
	return stdout.Bytes(), head.Hash().String(), nil
}

// VerifyBundle verifies that a bundle's prerequisites are present in the workspace repo.
func (s *Service) VerifyBundle(ctx context.Context, workspaceID string, bundleData []byte) error {
	repoPath := s.store.RepoPath(workspaceID)

	tmpFile, err := os.CreateTemp("", "lango-bundle-verify-*.bundle")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(bundleData); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write temp bundle: %w", err)
	}
	tmpFile.Close()

	cmd := exec.CommandContext(ctx, "git", "bundle", "verify", tmpFile.Name())
	cmd.Dir = repoPath

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "does not have") || strings.Contains(stderrStr, "prerequisite") {
			return ErrMissingPrerequisite
		}
		return fmt.Errorf("git bundle verify: %s: %w", stderrStr, err)
	}
	return nil
}

// snapshotRefs captures the current state of all refs in the workspace repo.
func (s *Service) snapshotRefs(ctx context.Context, workspaceID string) (map[string]string, error) {
	output, err := runGit(ctx, s.store.RepoPath(workspaceID), "for-each-ref", "--format=%(refname) %(objectname)")
	if err != nil {
		return nil, fmt.Errorf("for-each-ref: %w", err)
	}

	snapshot := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			snapshot[parts[0]] = parts[1]
		}
	}
	return snapshot, nil
}

// restoreRefs restores refs from a previously captured snapshot.
func (s *Service) restoreRefs(ctx context.Context, workspaceID string, snapshot map[string]string) error {
	repoPath := s.store.RepoPath(workspaceID)
	for ref, hash := range snapshot {
		cmd := exec.CommandContext(ctx, "git", "update-ref", ref, hash)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("restore ref %s: %w", ref, err)
		}
	}
	return nil
}

// SafeApplyBundle verifies, snapshots refs, applies, and rolls back on failure.
func (s *Service) SafeApplyBundle(ctx context.Context, workspaceID string, bundleData []byte) error {
	if err := s.VerifyBundle(ctx, workspaceID, bundleData); err != nil {
		return fmt.Errorf("verify bundle: %w", err)
	}

	snapshot, err := s.snapshotRefs(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("snapshot refs: %w", err)
	}

	if err := s.ApplyBundle(ctx, workspaceID, bundleData); err != nil {
		// Rollback on failure.
		s.logger.Warn("bundle apply failed, rolling back refs",
			zap.String("workspace", workspaceID), zap.Error(err))
		if rbErr := s.restoreRefs(ctx, workspaceID, snapshot); rbErr != nil {
			s.logger.Error("ref rollback failed",
				zap.String("workspace", workspaceID), zap.Error(rbErr))
		}
		return fmt.Errorf("apply bundle: %w", err)
	}

	s.logger.Info("safely applied bundle", zap.String("workspace", workspaceID))
	return nil
}
