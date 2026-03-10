package gitbundle

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

func newTestService(t *testing.T) *Service {
	t.Helper()
	store := newTestStore(t)
	logger := zap.NewNop()
	return NewService(store, logger)
}

func TestService_Init(t *testing.T) {
	svc := newTestService(t)

	err := svc.Init(context.Background(), "ws-1")
	require.NoError(t, err)

	// Verify the store has the repo.
	repo, err := svc.store.Repo("ws-1")
	require.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestService_Log_EmptyRepo(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	err := svc.Init(ctx, "ws-1")
	require.NoError(t, err)

	commits, err := svc.Log(ctx, "ws-1", 10)
	require.NoError(t, err)
	assert.Empty(t, commits)
}

func TestService_Leaves_EmptyRepo(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	err := svc.Init(ctx, "ws-1")
	require.NoError(t, err)

	leaves, err := svc.Leaves(ctx, "ws-1")
	require.NoError(t, err)
	assert.Empty(t, leaves)
}

func TestService_CreateBundle_EmptyRepo(t *testing.T) {
	skipIfNoGit(t)

	svc := newTestService(t)
	ctx := context.Background()

	err := svc.Init(ctx, "ws-1")
	require.NoError(t, err)

	bundle, hash, err := svc.CreateBundle(ctx, "ws-1")
	require.NoError(t, err)
	assert.Nil(t, bundle, "empty repo should produce nil bundle")
	assert.Empty(t, hash)
}
