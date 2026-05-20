package gitbundle

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
)

func TestServiceCreateBundle_RunsBundleCreatedHook(t *testing.T) {
	skipIfNoGit(t)

	ctx := context.Background()
	svc := newTestService(t)
	require.NoError(t, svc.Init(ctx, "ws-1"))

	workDir := initServiceCreateBundleRunsBundleCreatedHookWorktree(t, svc.store.RepoPath("ws-1"))
	first := commitServiceCreateBundleRunsBundleCreatedHookFile(t, workDir, "README.md", "# Test\n", "initial commit")
	pushServiceCreateBundleRunsBundleCreatedHookBranch(t, workDir, "main")
	setServiceCreateBundleRunsBundleCreatedHookBareHead(t, svc.store.RepoPath("ws-1"), "main")
	require.NotEmpty(t, first)

	var called int
	var gotWorkspace, gotHead string
	var gotSize int
	svc.SetBundleCreatedHook(func(_ context.Context, workspaceID, headCommit string, bundleSize int) {
		called++
		gotWorkspace = workspaceID
		gotHead = headCommit
		gotSize = bundleSize
	})

	bundle, head, err := svc.CreateBundle(ctx, "ws-1")
	require.NoError(t, err)
	require.NotEmpty(t, bundle)
	assert.Equal(t, first, head)
	assert.Equal(t, 1, called)
	assert.Equal(t, "ws-1", gotWorkspace)
	assert.Equal(t, first, gotHead)
	assert.Equal(t, len(bundle), gotSize)
}

func TestServiceDiffAndNumStat_ReturnContentAndParsedStats(t *testing.T) {
	skipIfNoGit(t)

	ctx := context.Background()
	svc := newTestService(t)
	require.NoError(t, svc.Init(ctx, "ws-1"))

	workDir := initServiceCreateBundleRunsBundleCreatedHookWorktree(t, svc.store.RepoPath("ws-1"))
	from := commitServiceCreateBundleRunsBundleCreatedHookFile(t, workDir, "README.md", "old line\n", "initial commit")
	to := commitServiceCreateBundleRunsBundleCreatedHookFile(t, workDir, "README.md", "new line\nextra line\n", "update readme")
	pushServiceCreateBundleRunsBundleCreatedHookBranch(t, workDir, "main")
	setServiceCreateBundleRunsBundleCreatedHookBareHead(t, svc.store.RepoPath("ws-1"), "main")

	diff, err := svc.Diff(ctx, "ws-1", from, to)
	require.NoError(t, err)
	assert.Contains(t, diff, "-old line")
	assert.Contains(t, diff, "+new line")
	assert.Contains(t, diff, "+extra line")

	stats, err := svc.DiffNumStat(ctx, "ws-1", from, to)
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.Equal(t, FileStat{
		FilePath:     "README.md",
		LinesAdded:   2,
		LinesRemoved: 1,
	}, stats[0])
}

func TestServiceDiffNumStat_ParsesBinaryFileStatsAsZeroCounts(t *testing.T) {
	skipIfNoGit(t)

	ctx := context.Background()
	svc := newTestService(t)
	require.NoError(t, svc.Init(ctx, "ws-1"))

	workDir := initServiceCreateBundleRunsBundleCreatedHookWorktree(t, svc.store.RepoPath("ws-1"))
	from := commitServiceCreateBundleRunsBundleCreatedHookFileBytes(t, workDir, "image.bin", []byte{0x00, 0x01}, "add binary")
	to := commitServiceCreateBundleRunsBundleCreatedHookFileBytes(t, workDir, "image.bin", []byte{0x00, 0x02, 0x03}, "update binary")
	pushServiceCreateBundleRunsBundleCreatedHookBranch(t, workDir, "main")
	setServiceCreateBundleRunsBundleCreatedHookBareHead(t, svc.store.RepoPath("ws-1"), "main")

	stats, err := svc.DiffNumStat(ctx, "ws-1", from, to)
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.Equal(t, FileStat{FilePath: "image.bin"}, stats[0])
}

func TestServiceCreateIncrementalBundle_ExistingBaseCreatesVerifiableBundle(t *testing.T) {
	skipIfNoGit(t)

	ctx := context.Background()
	svc := newTestService(t)
	require.NoError(t, svc.Init(ctx, "ws-1"))

	workDir := initServiceCreateBundleRunsBundleCreatedHookWorktree(t, svc.store.RepoPath("ws-1"))
	base := commitServiceCreateBundleRunsBundleCreatedHookFile(t, workDir, "README.md", "# Test\n", "initial commit")
	head := commitServiceCreateBundleRunsBundleCreatedHookFile(t, workDir, "feature.txt", "feature\n", "add feature")
	pushServiceCreateBundleRunsBundleCreatedHookBranch(t, workDir, "main")
	setServiceCreateBundleRunsBundleCreatedHookBareHead(t, svc.store.RepoPath("ws-1"), "main")

	bundle, gotHead, err := svc.CreateIncrementalBundle(ctx, "ws-1", base)
	require.NoError(t, err)
	require.NotEmpty(t, bundle)
	assert.Equal(t, head, gotHead)
	require.NoError(t, svc.VerifyBundle(ctx, "ws-1", bundle))
}

func TestServiceApplyBundle_UnbundlesObjectsIntoInitializedRepo(t *testing.T) {
	skipIfNoGit(t)

	ctx := context.Background()
	source := newTestService(t)
	require.NoError(t, source.Init(ctx, "source"))
	workDir := initServiceCreateBundleRunsBundleCreatedHookWorktree(t, source.store.RepoPath("source"))
	head := commitServiceCreateBundleRunsBundleCreatedHookFile(t, workDir, "README.md", "# Source\n", "initial commit")
	pushServiceCreateBundleRunsBundleCreatedHookBranch(t, workDir, "main")
	setServiceCreateBundleRunsBundleCreatedHookBareHead(t, source.store.RepoPath("source"), "main")

	bundle, _, err := source.CreateBundle(ctx, "source")
	require.NoError(t, err)
	require.NotEmpty(t, bundle)

	target := newTestService(t)
	require.NoError(t, target.ApplyBundle(ctx, "target", bundle))

	got := runServiceCreateBundleRunsBundleCreatedHookGit(t, target.store.RepoPath("target"), "cat-file", "-t", head)
	assert.Equal(t, "commit", got)
}

func TestServiceSafeApplyBundle_VerifiesSnapshotsAndApplies(t *testing.T) {
	skipIfNoGit(t)

	ctx := context.Background()
	source := newTestService(t)
	require.NoError(t, source.Init(ctx, "source"))
	sourceWorkDir := initServiceCreateBundleRunsBundleCreatedHookWorktree(t, source.store.RepoPath("source"))
	base := commitServiceCreateBundleRunsBundleCreatedHookFile(t, sourceWorkDir, "README.md", "# Source\n", "initial commit")
	head := commitServiceCreateBundleRunsBundleCreatedHookFile(t, sourceWorkDir, "feature.txt", "feature\n", "add feature")
	pushServiceCreateBundleRunsBundleCreatedHookBranch(t, sourceWorkDir, "main")
	setServiceCreateBundleRunsBundleCreatedHookBareHead(t, source.store.RepoPath("source"), "main")

	bundle, _, err := source.CreateIncrementalBundle(ctx, "source", base)
	require.NoError(t, err)
	require.NotEmpty(t, bundle)

	target := newTestService(t)
	require.NoError(t, target.Init(ctx, "target"))
	targetWorkDir := initServiceCreateBundleRunsBundleCreatedHookWorktree(t, target.store.RepoPath("target"))
	targetBase := commitServiceCreateBundleRunsBundleCreatedHookFile(t, targetWorkDir, "README.md", "# Source\n", "initial commit")
	pushServiceCreateBundleRunsBundleCreatedHookBranch(t, targetWorkDir, "main")
	setServiceCreateBundleRunsBundleCreatedHookBareHead(t, target.store.RepoPath("target"), "main")
	require.Equal(t, base, targetBase, "test setup must create matching prerequisite commits")

	before, err := target.snapshotRefs(ctx, "target")
	require.NoError(t, err)
	require.NotEmpty(t, before)

	require.NoError(t, target.SafeApplyBundle(ctx, "target", bundle))

	got := runServiceCreateBundleRunsBundleCreatedHookGit(t, target.store.RepoPath("target"), "cat-file", "-t", head)
	assert.Equal(t, "commit", got)
	after, err := target.snapshotRefs(ctx, "target")
	require.NoError(t, err)
	assert.Equal(t, before, after, "unbundle imports objects without moving refs")
}

func TestServiceRestoreRefs_ReturnsUpdateRefErrors(t *testing.T) {
	skipIfNoGit(t)

	ctx := context.Background()
	svc := newTestService(t)
	require.NoError(t, svc.Init(ctx, "ws-1"))

	err := svc.restoreRefs(ctx, "ws-1", map[string]string{
		"refs/heads/main": "not-a-commit",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "restore ref refs/heads/main")
}

func TestServiceMergeTaskBranch_RunsMergeHookWithFileStats(t *testing.T) {
	skipIfNoGit(t)
	skipIfNoGitMergeTreeWriteTree(t)

	ctx := context.Background()
	svc := newTestService(t)
	require.NoError(t, svc.Init(ctx, "ws-1"))

	workDir := initServiceCreateBundleRunsBundleCreatedHookWorktree(t, svc.store.RepoPath("ws-1"))
	initial := commitServiceCreateBundleRunsBundleCreatedHookFile(t, workDir, "README.md", "# Test\n", "initial commit")
	pushServiceCreateBundleRunsBundleCreatedHookBranch(t, workDir, "main")
	setServiceCreateBundleRunsBundleCreatedHookBareHead(t, svc.store.RepoPath("ws-1"), "main")

	runServiceCreateBundleRunsBundleCreatedHookGit(t, workDir, "checkout", "-B", "task/feat-1", "main")
	source := commitServiceCreateBundleRunsBundleCreatedHookFile(t, workDir, "feature.txt", "feature\n", "add feature")
	pushServiceCreateBundleRunsBundleCreatedHookBranch(t, workDir, "task/feat-1")

	var events []MergeHookEvent
	svc.SetMergeHook(func(_ context.Context, event MergeHookEvent) {
		events = append(events, event)
	})

	result, err := svc.MergeTaskBranch(ctx, "ws-1", "feat-1", "main")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)
	require.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, "ws-1", event.WorkspaceID)
	assert.Equal(t, "feat-1", event.TaskID)
	assert.Equal(t, "main", event.TargetBranch)
	assert.Equal(t, result.MergeCommit, event.MergeCommit)
	assert.Equal(t, source, event.SourceCommit)
	assert.Equal(t, initial, event.PreviousTarget)
	require.Len(t, event.Files, 1)
	assert.Equal(t, FileStat{
		FilePath:     "feature.txt",
		LinesAdded:   1,
		LinesRemoved: 0,
	}, event.Files[0])
}

func initServiceCreateBundleRunsBundleCreatedHookWorktree(t *testing.T, remotePath string) string {
	t.Helper()

	workDir := filepath.Join(t.TempDir(), "work")
	require.NoError(t, os.MkdirAll(workDir, 0o700))
	runServiceCreateBundleRunsBundleCreatedHookGit(t, workDir, "init")
	runServiceCreateBundleRunsBundleCreatedHookGit(t, workDir, "config", "user.email", "channelMetadataApprovalAndStop8@example.test")
	runServiceCreateBundleRunsBundleCreatedHookGit(t, workDir, "config", "user.name", "Bundle Author")
	runServiceCreateBundleRunsBundleCreatedHookGit(t, workDir, "config", "commit.gpgsign", "false")
	runServiceCreateBundleRunsBundleCreatedHookGit(t, workDir, "checkout", "-B", "main")
	runServiceCreateBundleRunsBundleCreatedHookGit(t, workDir, "remote", "add", "origin", remotePath)
	return workDir
}

func commitServiceCreateBundleRunsBundleCreatedHookFile(t *testing.T, workDir, name, content, message string) string {
	t.Helper()
	return commitServiceCreateBundleRunsBundleCreatedHookFileBytes(t, workDir, name, []byte(content), message)
}

func commitServiceCreateBundleRunsBundleCreatedHookFileBytes(t *testing.T, workDir, name string, content []byte, message string) string {
	t.Helper()

	path := filepath.Join(workDir, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, content, 0o600))
	runServiceCreateBundleRunsBundleCreatedHookGit(t, workDir, "add", name)
	runServiceCreateBundleRunsBundleCreatedHookGit(t, workDir, "commit", "-m", message)
	return runServiceCreateBundleRunsBundleCreatedHookGit(t, workDir, "rev-parse", "HEAD")
}

func pushServiceCreateBundleRunsBundleCreatedHookBranch(t *testing.T, workDir, branch string) {
	t.Helper()
	runServiceCreateBundleRunsBundleCreatedHookGit(t, workDir, "push", "-f", "origin", "HEAD:refs/heads/"+branch)
}

func setServiceCreateBundleRunsBundleCreatedHookBareHead(t *testing.T, repoPath, branch string) {
	t.Helper()
	runServiceCreateBundleRunsBundleCreatedHookGit(t, repoPath, "symbolic-ref", "HEAD", "refs/heads/"+branch)
}

func runServiceCreateBundleRunsBundleCreatedHookGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_DATE=2000-01-01T00:00:00Z",
		"GIT_COMMITTER_DATE=2000-01-01T00:00:00Z",
	)
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("git %v in %s timed out after 5s: %s", args, dir, stderr.String())
	}
	require.NoError(t, err, "git %v in %s failed: %s", args, dir, stderr.String())
	return stringsTrimSpace(stdout.String())
}

func skipIfNoGitMergeTreeWriteTree(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "merge-tree", "--write-tree", "--stdin")
	cmd.Stdin = bytes.NewBufferString("")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Skip("git merge-tree --write-tree capability probe timed out")
	}
	if err == nil {
		return
	}
	if bytes.Contains(stderr.Bytes(), []byte("unknown option")) ||
		bytes.Contains(stderr.Bytes(), []byte("usage:")) {
		t.Skip("git merge-tree --write-tree not supported by installed git")
	}
}

func stringsTrimSpace(s string) string {
	return string(bytes.TrimSpace([]byte(s)))
}
