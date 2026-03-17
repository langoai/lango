package gitbundle

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_CreateTaskBranch_EmptyTaskID(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	require.NoError(t, svc.Init(ctx, "ws-1"))

	err := svc.CreateTaskBranch(ctx, "ws-1", "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty task ID")
}

func TestService_CreateTaskBranch_Idempotent(t *testing.T) {
	skipIfNoGit(t)

	svc := newTestService(t)
	ctx := context.Background()

	require.NoError(t, svc.Init(ctx, "ws-1"))
	seedCommit(t, svc.store.RepoPath("ws-1"))

	err := svc.CreateTaskBranch(ctx, "ws-1", "task-1", "main")
	require.NoError(t, err)

	// Second call should be idempotent.
	err = svc.CreateTaskBranch(ctx, "ws-1", "task-1", "main")
	require.NoError(t, err)
}

func TestService_DeleteTaskBranch_Idempotent(t *testing.T) {
	skipIfNoGit(t)

	svc := newTestService(t)
	ctx := context.Background()

	require.NoError(t, svc.Init(ctx, "ws-1"))

	// Delete non-existent branch should be a no-op.
	err := svc.DeleteTaskBranch(ctx, "ws-1", "nonexistent")
	require.NoError(t, err)
}

func TestService_DeleteTaskBranch_EmptyTaskID(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	require.NoError(t, svc.Init(ctx, "ws-1"))

	err := svc.DeleteTaskBranch(ctx, "ws-1", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty task ID")
}

func TestService_ListBranches_EmptyRepo(t *testing.T) {
	skipIfNoGit(t)

	svc := newTestService(t)
	ctx := context.Background()

	require.NoError(t, svc.Init(ctx, "ws-1"))

	branches, err := svc.ListBranches(ctx, "ws-1")
	require.NoError(t, err)
	assert.Empty(t, branches)
}

func TestService_ListBranches_WithBranches(t *testing.T) {
	skipIfNoGit(t)

	svc := newTestService(t)
	ctx := context.Background()

	require.NoError(t, svc.Init(ctx, "ws-1"))
	seedCommit(t, svc.store.RepoPath("ws-1"))

	// Create a task branch.
	err := svc.CreateTaskBranch(ctx, "ws-1", "feat-1", "main")
	require.NoError(t, err)

	branches, err := svc.ListBranches(ctx, "ws-1")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(branches), 1)

	// Find the task branch.
	found := false
	for _, b := range branches {
		if b.Name == "task/feat-1" {
			found = true
			assert.NotEmpty(t, b.CommitHash)
		}
	}
	assert.True(t, found, "task/feat-1 branch should exist")
}

func TestParseConflictFiles(t *testing.T) {
	tests := []struct {
		give string
		want []string
	}{
		{
			give: "CONFLICT (content): Merge conflict in README.md\nCONFLICT (content): Merge conflict in main.go",
			want: []string{"README.md", "main.go"},
		},
		{
			give: "some other output\nno conflicts here",
			want: nil,
		},
		{
			give: "",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := parseConflictFiles(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

// seedCommit creates an initial commit in a bare repo so branches can be created.
func seedCommit(t *testing.T, repoPath string) {
	t.Helper()

	tmpDir := filepath.Join(t.TempDir(), "work")

	// Clone the bare repo (may fail on empty repo).
	cmd := exec.Command("git", "clone", repoPath, tmpDir)
	if err := cmd.Run(); err != nil {
		// Empty repo — init a new repo and push.
		cmd = exec.Command("git", "init", tmpDir)
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "-C", tmpDir, "remote", "add", "origin", repoPath)
		require.NoError(t, cmd.Run())
	}

	// Configure git user for commits (disable GPG signing for test environments).
	for _, kv := range [][2]string{
		{"user.email", "test@test.com"},
		{"user.name", "Test"},
		{"commit.gpgsign", "false"},
	} {
		cmd = exec.Command("git", "-C", tmpDir, "config", kv[0], kv[1])
		require.NoError(t, cmd.Run())
	}

	// Create a file and commit.
	testFile := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(testFile, []byte("# Test\n"), 0o644))

	cmd = exec.Command("git", "-C", tmpDir, "add", ".")
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", tmpDir, "commit", "-m", "initial commit")
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", tmpDir, "push", "origin", "HEAD:refs/heads/main")
	require.NoError(t, cmd.Run())
}
