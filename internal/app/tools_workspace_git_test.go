package app

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/p2p/gitbundle"
)

func TestWorkspaceGitToolsRoundTripWithLocalBareRepo(t *testing.T) {
	t.Parallel()
	workspaceGitSkipIfNoGit(t)

	ctx := context.Background()
	wc, repoRoot := newWorkspaceGitToolComponents(t)
	tools := workspaceGitToolsByName(t, wc)

	initTool := workspaceGitRequireTool(t, tools, "p2p_git_init")
	pushTool := workspaceGitRequireTool(t, tools, "p2p_git_push")
	logTool := workspaceGitRequireTool(t, tools, "p2p_git_log")
	diffTool := workspaceGitRequireTool(t, tools, "p2p_git_diff")
	leavesTool := workspaceGitRequireTool(t, tools, "p2p_git_leaves")

	initialized, err := initTool.Handler(ctx, map[string]interface{}{"workspaceId": "ws-1"})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"initialized": true, "workspaceId": "ws-1"}, initialized)

	emptyPush, err := pushTool.Handler(ctx, map[string]interface{}{"workspaceId": "ws-1"})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"pushed": false, "reason": "empty repository"}, emptyPush)

	workDir := workspaceGitInitWorktree(t, repoRoot, "ws-1")
	from := workspaceGitCommitFile(t, workDir, "README.md", "old line\n", "initial commit")
	to := workspaceGitCommitFile(t, workDir, "README.md", "new line\nextra line\n", "update readme")
	workspaceGitPushBranch(t, workDir, "main")
	workspaceGitSetBareHead(t, repoRoot, "ws-1", "main")

	logged, err := logTool.Handler(ctx, map[string]interface{}{"workspaceId": "ws-1", "limit": 10})
	require.NoError(t, err)
	logPayload := workspaceGitPayload(t, logged)
	assert.Equal(t, 2, logPayload["count"])
	commits, ok := logPayload["commits"].([]map[string]interface{})
	require.True(t, ok)
	gotMessages := make(map[string]bool, len(commits))
	for _, commit := range commits {
		gotMessages[commit["message"].(string)] = true
		assert.Equal(t, "Workspace Git Author", commit["author"])
		assert.NotEmpty(t, commit["hash"])
		_, err := time.Parse(time.RFC3339, commit["timestamp"].(string))
		require.NoError(t, err)
	}
	assert.True(t, gotMessages["initial commit"])
	assert.True(t, gotMessages["update readme"])

	diffed, err := diffTool.Handler(ctx, map[string]interface{}{"workspaceId": "ws-1", "from": from, "to": to})
	require.NoError(t, err)
	diffPayload := workspaceGitPayload(t, diffed)
	diff, ok := diffPayload["diff"].(string)
	require.True(t, ok)
	assert.Contains(t, diff, "-old line")
	assert.Contains(t, diff, "+new line")
	assert.Contains(t, diff, "+extra line")

	leaves, err := leavesTool.Handler(ctx, map[string]interface{}{"workspaceId": "ws-1"})
	require.NoError(t, err)
	leavesPayload := workspaceGitPayload(t, leaves)
	assert.Equal(t, 1, leavesPayload["count"])
	assert.Equal(t, []string{to}, leavesPayload["leaves"])

	pushed, err := pushTool.Handler(ctx, map[string]interface{}{
		"workspaceId": "ws-1",
		"message":     "share update",
	})
	require.NoError(t, err)
	pushPayload := workspaceGitPayload(t, pushed)
	assert.Equal(t, true, pushPayload["pushed"])
	assert.Equal(t, to, pushPayload["headCommit"])
	assert.Equal(t, "share update", pushPayload["message"])
	assert.Greater(t, pushPayload["bundleSize"].(int), 0)
}

func TestWorkspaceToolsRegisterGitToolsOnlyWhenServiceExists(t *testing.T) {
	t.Parallel()

	withoutGit := workspaceToolsNames(buildWorkspaceTools(&wsComponents{
		manager: newWorkspaceToolComponents(t).manager,
	}))
	assert.NotContains(t, withoutGit, "p2p_git_init")
	assert.NotContains(t, withoutGit, "p2p_git_push")

	wc, _ := newWorkspaceGitToolComponents(t)
	wc.manager = newWorkspaceToolComponents(t).manager
	withGit := workspaceToolsNames(buildWorkspaceTools(wc))
	for _, name := range []string{
		"p2p_git_init",
		"p2p_git_push",
		"p2p_git_log",
		"p2p_git_diff",
		"p2p_git_leaves",
	} {
		assert.Contains(t, withGit, name)
	}
}

func TestWorkspaceGitToolsPropagateServiceErrors(t *testing.T) {
	t.Parallel()
	workspaceGitSkipIfNoGit(t)

	ctx := context.Background()
	wc, _ := newWorkspaceGitToolComponents(t)
	tools := workspaceGitToolsByName(t, wc)

	for _, tc := range []struct {
		name   string
		tool   string
		params map[string]interface{}
	}{
		{name: "push missing repo", tool: "p2p_git_push", params: map[string]interface{}{"workspaceId": "missing"}},
		{name: "log missing repo", tool: "p2p_git_log", params: map[string]interface{}{"workspaceId": "missing"}},
		{name: "diff missing repo", tool: "p2p_git_diff", params: map[string]interface{}{"workspaceId": "missing", "from": "a", "to": "b"}},
		{name: "leaves missing repo", tool: "p2p_git_leaves", params: map[string]interface{}{"workspaceId": "missing"}},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := workspaceGitRequireTool(t, tools, tc.tool).Handler(ctx, tc.params)
			require.Error(t, err)
			assert.Nil(t, got)
		})
	}
}

func TestWorkspaceGitInitPropagatesRepoStoreError(t *testing.T) {
	t.Parallel()
	workspaceGitSkipIfNoGit(t)

	blocker := filepath.Join(t.TempDir(), "repo-root")
	require.NoError(t, os.WriteFile(blocker, []byte("not a directory"), 0o600))
	wc := &wsComponents{
		gitService: gitbundle.NewService(gitbundle.NewBareRepoStore(blocker, zap.NewNop()), zap.NewNop()),
	}
	tools := workspaceGitToolsByName(t, wc)

	got, err := workspaceGitRequireTool(t, tools, "p2p_git_init").Handler(context.Background(), map[string]interface{}{
		"workspaceId": "ws-1",
	})
	require.Error(t, err)
	assert.Nil(t, got)
}

func newWorkspaceGitToolComponents(t *testing.T) (*wsComponents, string) {
	t.Helper()

	repoRoot := t.TempDir()
	store := gitbundle.NewBareRepoStore(repoRoot, zap.NewNop())
	return &wsComponents{
		gitService: gitbundle.NewService(store, zap.NewNop()),
		localDID:   "did:lango:workspace-git-test",
	}, repoRoot
}

func workspaceGitToolsByName(t *testing.T, wc *wsComponents) map[string]*agent.Tool {
	t.Helper()

	tools := make(map[string]*agent.Tool)
	for _, tool := range buildGitTools(wc) {
		tools[tool.Name] = tool
	}
	return tools
}

func workspaceGitRequireTool(t *testing.T, tools map[string]*agent.Tool, name string) *agent.Tool {
	t.Helper()

	tool := tools[name]
	require.NotNil(t, tool)
	return tool
}

func workspaceGitPayload(t *testing.T, got interface{}) map[string]interface{} {
	t.Helper()

	payload, ok := got.(map[string]interface{})
	require.True(t, ok)
	return payload
}

func workspaceToolsNames(tools []*agent.Tool) map[string]bool {
	names := make(map[string]bool, len(tools))
	for _, tool := range tools {
		names[tool.Name] = true
	}
	return names
}

func workspaceGitSkipIfNoGit(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git executable not available")
	}
}

func workspaceGitInitWorktree(t *testing.T, repoRoot, workspaceID string) string {
	t.Helper()

	workDir := filepath.Join(t.TempDir(), "work")
	require.NoError(t, os.MkdirAll(workDir, 0o700))
	workspaceGitRun(t, workDir, "init")
	workspaceGitRun(t, workDir, "config", "user.email", "workspace-git@example.test")
	workspaceGitRun(t, workDir, "config", "user.name", "Workspace Git Author")
	workspaceGitRun(t, workDir, "config", "commit.gpgsign", "false")
	workspaceGitRun(t, workDir, "checkout", "-B", "main")
	workspaceGitRun(t, workDir, "remote", "add", "origin", workspaceGitRepoPath(repoRoot, workspaceID))
	return workDir
}

func workspaceGitCommitFile(t *testing.T, workDir, name, content, message string) string {
	t.Helper()

	path := filepath.Join(workDir, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	workspaceGitRun(t, workDir, "add", name)
	workspaceGitRun(t, workDir, "commit", "-m", message)
	return workspaceGitRun(t, workDir, "rev-parse", "HEAD")
}

func workspaceGitPushBranch(t *testing.T, workDir, branch string) {
	t.Helper()
	workspaceGitRun(t, workDir, "push", "-f", "origin", "HEAD:refs/heads/"+branch)
}

func workspaceGitSetBareHead(t *testing.T, repoRoot, workspaceID, branch string) {
	t.Helper()
	workspaceGitRun(t, workspaceGitRepoPath(repoRoot, workspaceID), "symbolic-ref", "HEAD", "refs/heads/"+branch)
}

func workspaceGitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_DATE=2000-01-01T00:00:00Z",
		"GIT_COMMITTER_DATE=2000-01-01T00:00:00Z",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("git %v in %s timed out: %s", args, dir, stderr.String())
	}
	require.NoError(t, err, "git %v in %s failed: %s", args, dir, stderr.String())
	return string(bytes.TrimSpace(stdout.Bytes()))
}

func workspaceGitRepoPath(repoRoot, workspaceID string) string {
	return filepath.Join(repoRoot, workspaceID, "repo.git")
}
