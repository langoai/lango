package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func TestInitWorkspaceDisabledReturnsNilWithoutNode(t *testing.T) {
	t.Parallel()

	got := initWorkspace(&config.Config{}, nil, "did:lango:test-local", nil)

	assert.Nil(t, got)
}

func TestInitWorkspaceReturnsNilWhenDataDirCannotBeCreated(t *testing.T) {
	t.Parallel()

	blocker := filepath.Join(t.TempDir(), "not-a-directory")
	require.NoError(t, os.WriteFile(blocker, []byte("block workspace data dir"), 0o600))

	cfg := &config.Config{}
	cfg.P2P.Workspace.Enabled = true
	cfg.P2P.Workspace.DataDir = blocker

	got := initWorkspace(cfg, nil, "did:lango:test-local", nil)

	assert.Nil(t, got)
}

func TestInitWorkspaceReturnsNilWhenBoltDBCannotOpen(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dataDir, "workspaces.db"), 0o700))

	cfg := &config.Config{}
	cfg.P2P.Workspace.Enabled = true
	cfg.P2P.Workspace.DataDir = dataDir

	got := initWorkspace(cfg, nil, "did:lango:test-local", nil)

	assert.Nil(t, got)
}
