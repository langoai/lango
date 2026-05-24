package gitbundle

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceLogReturnsUniqueCommitsUpToLimit(t *testing.T) {
	skipIfNoGit(t)

	ctx := context.Background()
	svc := newTestService(t)
	require.NoError(t, svc.Init(ctx, "ws-1"))

	workDir := initServiceCreateBundleRunsBundleCreatedHookWorktree(t, svc.store.RepoPath("ws-1"))
	first := commitServiceCreateBundleRunsBundleCreatedHookFile(t, workDir, "README.md", "# Test\n", "initial commit")
	second := commitServiceCreateBundleRunsBundleCreatedHookFile(t, workDir, "feature.txt", "feature\n", "add feature")
	third := commitServiceCreateBundleRunsBundleCreatedHookFile(t, workDir, "notes.txt", "notes\n", "add notes")
	pushServiceCreateBundleRunsBundleCreatedHookBranch(t, workDir, "main")
	setServiceCreateBundleRunsBundleCreatedHookBareHead(t, svc.store.RepoPath("ws-1"), "main")

	commits, err := svc.Log(ctx, "ws-1", 2)
	require.NoError(t, err)
	require.Len(t, commits, 2)

	gotHashes := map[string]bool{
		commits[0].Hash: true,
		commits[1].Hash: true,
	}
	assert.True(t, gotHashes[third], "most recent commit should be included")
	assert.True(t, gotHashes[second], "second most recent commit should be included")
	assert.False(t, gotHashes[first], "limit should stop before older commits")
	for _, commit := range commits {
		assert.NotEmpty(t, commit.Message)
		assert.Equal(t, "Bundle Author", commit.Author)
		assert.False(t, commit.Timestamp.IsZero())
	}
}

func TestServiceLeavesReturnsBranchTipsOnly(t *testing.T) {
	skipIfNoGit(t)

	ctx := context.Background()
	svc := newTestService(t)
	require.NoError(t, svc.Init(ctx, "ws-1"))

	workDir := initServiceCreateBundleRunsBundleCreatedHookWorktree(t, svc.store.RepoPath("ws-1"))
	base := commitServiceCreateBundleRunsBundleCreatedHookFile(t, workDir, "README.md", "# Test\n", "initial commit")

	runServiceCreateBundleRunsBundleCreatedHookGit(t, workDir, "checkout", "-B", "task/one", base)
	leafOne := commitServiceCreateBundleRunsBundleCreatedHookFile(t, workDir, "one.txt", "one\n", "add one")
	pushServiceCreateBundleRunsBundleCreatedHookBranch(t, workDir, "task/one")

	runServiceCreateBundleRunsBundleCreatedHookGit(t, workDir, "checkout", "-B", "task/two", base)
	leafTwo := commitServiceCreateBundleRunsBundleCreatedHookFile(t, workDir, "two.txt", "two\n", "add two")
	pushServiceCreateBundleRunsBundleCreatedHookBranch(t, workDir, "task/two")
	setServiceCreateBundleRunsBundleCreatedHookBareHead(t, svc.store.RepoPath("ws-1"), "task/one")

	leaves, err := svc.Leaves(ctx, "ws-1")
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{leafOne, leafTwo}, leaves)
	assert.NotContains(t, leaves, base)
}

func TestServiceCreateIncrementalBundleAtHeadReturnsEmptyBundle(t *testing.T) {
	skipIfNoGit(t)

	ctx := context.Background()
	svc := newTestService(t)
	require.NoError(t, svc.Init(ctx, "ws-1"))

	workDir := initServiceCreateBundleRunsBundleCreatedHookWorktree(t, svc.store.RepoPath("ws-1"))
	head := commitServiceCreateBundleRunsBundleCreatedHookFile(t, workDir, "README.md", "# Test\n", "initial commit")
	pushServiceCreateBundleRunsBundleCreatedHookBranch(t, workDir, "main")
	setServiceCreateBundleRunsBundleCreatedHookBareHead(t, svc.store.RepoPath("ws-1"), "main")

	bundle, gotHead, err := svc.CreateIncrementalBundle(ctx, "ws-1", head)
	require.NoError(t, err)
	assert.Nil(t, bundle)
	assert.Empty(t, gotHead)
}

func TestServiceVerifyBundleReportsMissingPrerequisite(t *testing.T) {
	skipIfNoGit(t)

	ctx := context.Background()
	source := newTestService(t)
	require.NoError(t, source.Init(ctx, "source"))
	workDir := initServiceCreateBundleRunsBundleCreatedHookWorktree(t, source.store.RepoPath("source"))
	base := commitServiceCreateBundleRunsBundleCreatedHookFile(t, workDir, "README.md", "# Source\n", "initial commit")
	commitServiceCreateBundleRunsBundleCreatedHookFile(t, workDir, "feature.txt", "feature\n", "add feature")
	pushServiceCreateBundleRunsBundleCreatedHookBranch(t, workDir, "main")
	setServiceCreateBundleRunsBundleCreatedHookBareHead(t, source.store.RepoPath("source"), "main")

	bundle, _, err := source.CreateIncrementalBundle(ctx, "source", base)
	require.NoError(t, err)
	require.NotEmpty(t, bundle)

	target := newTestService(t)
	require.NoError(t, target.Init(ctx, "target"))

	err = target.VerifyBundle(ctx, "target", bundle)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrMissingPrerequisite), "got %v", err)
}
