package gitbundle

import (
	"context"
	"os/exec"
	"strings"
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

func TestValidateCommitHash(t *testing.T) {
	tests := []struct {
		give string
		want bool
	}{
		{give: "a" + strings.Repeat("0", 39), want: true},
		{give: strings.Repeat("f", 40), want: true},
		{give: strings.Repeat("0", 40), want: true},
		{give: strings.Repeat("0", 39), want: false},
		{give: strings.Repeat("0", 41), want: false},
		{give: strings.Repeat("g", 40), want: false},
		{give: strings.Repeat("A", 40), want: false},
		{give: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, validateCommitHash(tt.give))
		})
	}
}

func TestService_HasCommit_InvalidHash(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	require.NoError(t, svc.Init(ctx, "ws-1"))

	_, err := svc.HasCommit(ctx, "ws-1", "not-a-hash")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid commit hash")
}

func TestService_HasCommit_EmptyRepo(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	require.NoError(t, svc.Init(ctx, "ws-1"))

	exists, err := svc.HasCommit(ctx, "ws-1", strings.Repeat("a", 40))
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestService_CreateIncrementalBundle_InvalidBase(t *testing.T) {
	skipIfNoGit(t)

	svc := newTestService(t)
	ctx := context.Background()

	require.NoError(t, svc.Init(ctx, "ws-1"))

	_, _, err := svc.CreateIncrementalBundle(ctx, "ws-1", "bad-hash")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid base commit")
}

func TestService_CreateIncrementalBundle_MissingBase_FallbackToFull(t *testing.T) {
	skipIfNoGit(t)

	svc := newTestService(t)
	ctx := context.Background()

	require.NoError(t, svc.Init(ctx, "ws-1"))

	// Empty repo — fallback should still produce nil bundle gracefully.
	bundle, _, err := svc.CreateIncrementalBundle(ctx, "ws-1", strings.Repeat("a", 40))
	require.NoError(t, err)
	assert.Nil(t, bundle)
}

func TestService_VerifyBundle_EmptyRepo(t *testing.T) {
	skipIfNoGit(t)

	svc := newTestService(t)
	ctx := context.Background()

	require.NoError(t, svc.Init(ctx, "ws-1"))

	// Create a bundle from another empty workspace — should fail verification.
	err := svc.VerifyBundle(ctx, "ws-1", []byte("not-a-valid-bundle"))
	require.Error(t, err)
}

func TestService_SnapshotAndRestoreRefs_EmptyRepo(t *testing.T) {
	skipIfNoGit(t)

	svc := newTestService(t)
	ctx := context.Background()

	require.NoError(t, svc.Init(ctx, "ws-1"))

	snapshot, err := svc.snapshotRefs(ctx, "ws-1")
	require.NoError(t, err)
	assert.Empty(t, snapshot, "empty repo should have no refs")

	// Restore with empty snapshot should be a no-op.
	err = svc.restoreRefs(ctx, "ws-1", snapshot)
	require.NoError(t, err)
}

func TestService_SafeApplyBundle_InvalidBundle(t *testing.T) {
	skipIfNoGit(t)

	svc := newTestService(t)
	ctx := context.Background()

	require.NoError(t, svc.Init(ctx, "ws-1"))

	err := svc.SafeApplyBundle(ctx, "ws-1", []byte("garbage"))
	require.Error(t, err)
}
