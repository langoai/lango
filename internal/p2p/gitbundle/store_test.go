package gitbundle

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestStore(t *testing.T) *BareRepoStore {
	t.Helper()
	baseDir := t.TempDir()
	logger := zap.NewNop()
	return NewBareRepoStore(baseDir, logger)
}

func TestBareRepoStore_Init(t *testing.T) {
	store := newTestStore(t)

	err := store.Init("ws-1")
	require.NoError(t, err)

	// Verify the bare repo directory was created.
	repoPath := store.RepoPath("ws-1")
	assert.DirExists(t, repoPath)
}

func TestBareRepoStore_Init_Idempotent(t *testing.T) {
	store := newTestStore(t)

	err := store.Init("ws-1")
	require.NoError(t, err)

	// Init again should not error.
	err = store.Init("ws-1")
	require.NoError(t, err)
}

func TestBareRepoStore_Repo(t *testing.T) {
	store := newTestStore(t)

	err := store.Init("ws-1")
	require.NoError(t, err)

	repo, err := store.Repo("ws-1")
	require.NoError(t, err)
	assert.NotNil(t, repo)

	// Verify it's a bare repo.
	cfg, err := repo.Config()
	require.NoError(t, err)
	assert.True(t, cfg.Core.IsBare)
}

func TestBareRepoStore_Repo_NotFound(t *testing.T) {
	store := newTestStore(t)

	_, err := store.Repo("ws-nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "open repo")
}

func TestBareRepoStore_RepoPath(t *testing.T) {
	store := newTestStore(t)

	path := store.RepoPath("ws-1")
	expected := filepath.Join(store.baseDir, "ws-1", "repo.git")
	assert.Equal(t, expected, path)
}

func TestBareRepoStore_List(t *testing.T) {
	store := newTestStore(t)

	// Empty initially.
	ids, err := store.List()
	require.NoError(t, err)
	assert.Empty(t, ids)

	// Init multiple workspaces.
	for _, id := range []string{"ws-a", "ws-b", "ws-c"} {
		err := store.Init(id)
		require.NoError(t, err)
	}

	ids, err = store.List()
	require.NoError(t, err)
	assert.Len(t, ids, 3)

	// Verify all IDs are present (order may vary).
	idSet := make(map[string]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}
	assert.True(t, idSet["ws-a"])
	assert.True(t, idSet["ws-b"])
	assert.True(t, idSet["ws-c"])
}

func TestBareRepoStore_Remove(t *testing.T) {
	store := newTestStore(t)

	err := store.Init("ws-1")
	require.NoError(t, err)

	// Verify it exists.
	repoPath := store.RepoPath("ws-1")
	assert.DirExists(t, repoPath)

	// Remove it.
	err = store.Remove("ws-1")
	require.NoError(t, err)

	// Verify the repo directory is gone.
	assert.NoDirExists(t, repoPath)

	// Repo should fail after removal.
	_, err = store.Repo("ws-1")
	require.Error(t, err)
}
