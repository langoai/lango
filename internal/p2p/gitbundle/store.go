package gitbundle

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-git/go-git/v5"
	"go.uber.org/zap"
)

// BareRepoStore manages bare git repositories per workspace.
type BareRepoStore struct {
	baseDir string
	logger  *zap.Logger
	mu      sync.RWMutex
	repos   map[string]*git.Repository
}

// NewBareRepoStore creates a BareRepoStore rooted at baseDir.
func NewBareRepoStore(baseDir string, logger *zap.Logger) *BareRepoStore {
	return &BareRepoStore{
		baseDir: baseDir,
		logger:  logger,
		repos:   make(map[string]*git.Repository),
	}
}

// Init initializes a bare git repository for the given workspace.
func (s *BareRepoStore) Init(workspaceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.repos[workspaceID]; exists {
		return nil
	}

	repoPath := s.repoPath(workspaceID)
	if err := os.MkdirAll(repoPath, 0o700); err != nil {
		return fmt.Errorf("create repo dir %s: %w", repoPath, err)
	}

	repo, err := git.PlainInit(repoPath, true)
	if err != nil {
		// Try opening existing repo
		repo, err = git.PlainOpen(repoPath)
		if err != nil {
			return fmt.Errorf("init bare repo %s: %w", repoPath, err)
		}
	}

	// Ensure default config
	cfg, err := repo.Config()
	if err != nil {
		return fmt.Errorf("read repo config: %w", err)
	}
	cfg.Core.IsBare = true
	if err := repo.SetConfig(cfg); err != nil {
		return fmt.Errorf("set repo config: %w", err)
	}

	s.repos[workspaceID] = repo
	s.logger.Info("initialized bare repo", zap.String("workspace", workspaceID), zap.String("path", repoPath))
	return nil
}

// Repo returns the git.Repository for a workspace, opening it if needed.
func (s *BareRepoStore) Repo(workspaceID string) (*git.Repository, error) {
	s.mu.RLock()
	if repo, ok := s.repos[workspaceID]; ok {
		s.mu.RUnlock()
		return repo, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if repo, ok := s.repos[workspaceID]; ok {
		return repo, nil
	}

	repoPath := s.repoPath(workspaceID)
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("open repo %s: %w", workspaceID, err)
	}

	s.repos[workspaceID] = repo
	return repo, nil
}

// RepoPath returns the filesystem path for a workspace's bare repo.
func (s *BareRepoStore) RepoPath(workspaceID string) string {
	return s.repoPath(workspaceID)
}

// List returns all workspace IDs that have initialized repos.
func (s *BareRepoStore) List() ([]string, error) {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read dir %s: %w", s.baseDir, err)
	}

	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			ids = append(ids, e.Name())
		}
	}
	return ids, nil
}

// Remove deletes the bare repo for a workspace.
func (s *BareRepoStore) Remove(workspaceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.repos, workspaceID)
	repoPath := s.repoPath(workspaceID)
	if err := os.RemoveAll(repoPath); err != nil {
		return fmt.Errorf("remove repo %s: %w", repoPath, err)
	}
	s.logger.Info("removed bare repo", zap.String("workspace", workspaceID))
	return nil
}

func (s *BareRepoStore) repoPath(workspaceID string) string {
	return filepath.Join(s.baseDir, workspaceID, "repo.git")
}
