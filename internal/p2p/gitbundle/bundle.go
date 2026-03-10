package gitbundle

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.uber.org/zap"
)

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
		if strings.Contains(stderr.String(), "empty bundle") ||
			strings.Contains(stderr.String(), "Refusing to create empty bundle") {
			return nil, "", nil
		}
		return nil, "", fmt.Errorf("git bundle create: %s: %w", stderr.String(), err)
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

	// Fetch the unbundled refs.
	cmd2 := exec.CommandContext(ctx, "git", "fetch", "-", "--all")
	cmd2.Dir = repoPath
	cmd2.Stdin = bytes.NewReader(bundle)

	var stderr2 bytes.Buffer
	cmd2.Stderr = &stderr2

	// fetch from bundle may fail if refs already exist; that's OK.
	_ = cmd2.Run()

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
				return fmt.Errorf("limit reached")
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
	if err != nil && err.Error() != "limit reached" {
		return nil, fmt.Errorf("iterate commits: %w", err)
	}

	return commits, nil
}

// Diff returns the diff between two commits.
func (s *Service) Diff(ctx context.Context, workspaceID, from, to string) (string, error) {
	repoPath := s.store.RepoPath(workspaceID)

	cmd := exec.CommandContext(ctx, "git", "diff", from, to)
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git diff: %s: %w", stderr.String(), err)
	}

	return stdout.String(), nil
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
