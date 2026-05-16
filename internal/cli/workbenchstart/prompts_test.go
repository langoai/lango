package workbenchstart

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildStarterPrompts_DefaultWhenNoWorkDir(t *testing.T) {
	t.Parallel()

	assert.Equal(t, DefaultPrompts(), BuildPrompts(""))
}

func TestBuildStarterPrompts_ContextAwareForNestedGoRepo(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Lango Test")
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/demo\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n"), 0o644))
	runGit(t, root, "add", "go.mod", "main.go")
	runGit(t, root, "commit", "-m", "init")
	nested := filepath.Join(root, "internal", "feature")
	require.NoError(t, os.MkdirAll(nested, 0o755))
	branch := runGitOutput(t, root, "rev-parse", "--abbrev-ref", "HEAD")

	got := BuildPrompts(nested)

	assert.Equal(t, []string{
		"Summarize the " + filepath.Base(root) + " repository and its current purpose",
		"Explain the Go package layout in " + filepath.Base(root) + " and where to start editing",
		"Review the current state of branch " + branch + " in " + filepath.Base(root) + " and suggest the best next change",
	}, got)
}

func TestBuildStarterPrompts_FallbackWhenNoRepoMarkers(t *testing.T) {
	t.Parallel()

	got := BuildPrompts(t.TempDir())
	assert.Equal(t, DefaultPrompts(), got)
}

func TestBuildStarterPrompts_DirtyBranchMentionsUncommittedChanges(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Lango Test")
	require.NoError(t, os.WriteFile(filepath.Join(root, "README.md"), []byte("hello\n"), 0o644))
	runGit(t, root, "add", "README.md")
	runGit(t, root, "commit", "-m", "init")
	require.NoError(t, os.WriteFile(filepath.Join(root, "README.md"), []byte("changed\n"), 0o644))
	branch := runGitOutput(t, root, "rev-parse", "--abbrev-ref", "HEAD")

	got := BuildPrompts(root)
	assert.Contains(t, got[2], "Review the uncommitted changes on branch "+branch)
	assert.Contains(t, got[2], "`README.md`")
}

func TestBuildStarterPrompts_DirtyBranchSummarizesMultipleTargets(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Lango Test")
	require.NoError(t, os.MkdirAll(filepath.Join(root, "internal"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "docs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "README.md"), []byte("hello\n"), 0o644))
	runGit(t, root, "add", "README.md")
	runGit(t, root, "commit", "-m", "init")
	require.NoError(t, os.WriteFile(filepath.Join(root, "README.md"), []byte("changed\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "internal", "app.go"), []byte("package internal\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "docs", "guide.md"), []byte("# docs\n"), 0o644))

	got := BuildPrompts(root)
	assert.Contains(t, got[2], "`README.md`")
	assert.Contains(t, got[2], "`internal`")
	assert.Contains(t, got[2], "`docs`")
}

func TestDefaultPrompt_UsesChangedReviewForDirtyRepo(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Lango Test")
	require.NoError(t, os.WriteFile(filepath.Join(root, "README.md"), []byte("hello\n"), 0o644))
	runGit(t, root, "add", "README.md")
	runGit(t, root, "commit", "-m", "init")
	require.NoError(t, os.WriteFile(filepath.Join(root, "README.md"), []byte("changed\n"), 0o644))
	branch := runGitOutput(t, root, "rev-parse", "--abbrev-ref", "HEAD")

	got := DefaultPrompt(root)
	assert.Contains(t, got, "Review the uncommitted changes on branch "+branch)
	assert.Contains(t, got, "`README.md`")
}

func TestDefaultPrompt_FallsBackToSummaryWhenClean(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Lango Test")
	require.NoError(t, os.WriteFile(filepath.Join(root, "README.md"), []byte("hello\n"), 0o644))
	runGit(t, root, "add", "README.md")
	runGit(t, root, "commit", "-m", "init")

	got := DefaultPrompt(root)
	assert.Contains(t, got, "Summarize the "+filepath.Base(root)+" repository and its current purpose")
}

func TestPostTurnDefaultPrompt_UsesStructureStarterWhenNoWorkspaceContext(t *testing.T) {
	t.Parallel()

	got := PostTurnDefaultPrompt("")
	assert.Equal(t, "Explain the current project structure", got)
}

func TestPostTurnDefaultPrompt_UsesNextChangeStarterForDetectedWorkspace(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Lango Test")
	require.NoError(t, os.WriteFile(filepath.Join(root, "README.md"), []byte("hello\n"), 0o644))
	runGit(t, root, "add", "README.md")
	runGit(t, root, "commit", "-m", "init")
	branch := runGitOutput(t, root, "rev-parse", "--abbrev-ref", "HEAD")

	got := PostTurnDefaultPrompt(root)
	assert.Contains(t, got, "Review the current state of branch "+branch)
}

func TestRecoveryDefaultPrompt_UsesRecentChangeStarterWhenNoWorkspaceContext(t *testing.T) {
	t.Parallel()

	got := RecoveryDefaultPrompt("")
	assert.Equal(t, "Review recent changes", got)
}

func TestRecoveryDefaultPrompt_UsesNextChangeStarterForDetectedWorkspace(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Lango Test")
	require.NoError(t, os.WriteFile(filepath.Join(root, "README.md"), []byte("hello\n"), 0o644))
	runGit(t, root, "add", "README.md")
	runGit(t, root, "commit", "-m", "init")
	branch := runGitOutput(t, root, "rev-parse", "--abbrev-ref", "HEAD")

	got := RecoveryDefaultPrompt(root)
	assert.Contains(t, got, "Review the current state of branch "+branch)
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed: %s", args, string(out))
}

func runGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed: %s", args, string(out))
	return string(bytes.TrimSpace(out))
}
