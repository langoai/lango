package os

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompileBwrapArgs_DefaultToolPolicy(t *testing.T) {
	work := t.TempDir()
	dataRoot := t.TempDir()
	// .git baseline deny is now part of DefaultToolPolicy. compileBwrapArgs
	// requires deny paths to exist as directories.
	gitDir := filepath.Join(work, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0o755))

	policy := DefaultToolPolicy(work, dataRoot)

	args, err := compileBwrapArgs(policy)
	require.NoError(t, err)

	assertContainsPair(t, args, "--ro-bind", "/", "/")
	assertContainsPair(t, args, "--bind", work, work)
	assertContainsPair(t, args, "--bind", "/tmp", "/tmp")
	assert.Contains(t, args, "--unshare-net")

	// Standard namespace flags
	assert.Contains(t, args, "--die-with-parent")
	assert.Contains(t, args, "--unshare-pid")
	assert.Contains(t, args, "--unshare-ipc")
	assert.Contains(t, args, "--unshare-uts")
	assert.Contains(t, args, "--unshare-cgroup-try")

	// .git baseline + dataRoot control-plane deny (both as --tmpfs masks).
	assertContainsSingle(t, args, "--tmpfs", gitDir)
	assertContainsSingle(t, args, "--tmpfs", dataRoot)
}

func TestCompileBwrapArgs_StrictToolPolicy(t *testing.T) {
	work := t.TempDir()
	dataRoot := t.TempDir()
	gitDir := filepath.Join(work, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0o755))

	policy := StrictToolPolicy(work, dataRoot)

	args, err := compileBwrapArgs(policy)
	require.NoError(t, err)

	assertContainsPair(t, args, "--ro-bind", "/", "/")
	assertContainsPair(t, args, "--bind", work, work)
	assertContainsSingle(t, args, "--tmpfs", gitDir)
	assertContainsSingle(t, args, "--tmpfs", dataRoot)
}

func TestCompileBwrapArgs_MCPServerPolicy(t *testing.T) {
	dataRoot := t.TempDir()
	// Empty workspacePath — this test focuses on the dataRoot deny shape.
	// A dedicated test (TestMCPServerPolicy_DenyWorkspaceGit) covers walk-up.
	policy := MCPServerPolicy("", dataRoot)

	args, err := compileBwrapArgs(policy)
	require.NoError(t, err)

	assertContainsPair(t, args, "--ro-bind", "/", "/")
	assertContainsPair(t, args, "--bind", "/tmp", "/tmp")
	assert.NotContains(t, args, "--unshare-net", "MCPServerPolicy uses NetworkAllow — must not unshare net")
	assertContainsSingle(t, args, "--tmpfs", dataRoot)
}

// TestCompileBwrapArgs_DefaultToolPolicy_NoGitDir verifies that a non-repo
// workspace (no .git directory) does not cause compileBwrapArgs to fail.
// Regression: PR 4 added .git unconditionally to DenyPaths, which combined
// with bwrap's strict stat+IsDir check rejected every non-repo workspace.
func TestCompileBwrapArgs_DefaultToolPolicy_NoGitDir(t *testing.T) {
	workDir := t.TempDir() // no .git
	policy := DefaultToolPolicy(workDir, "")

	// isDir guard skips missing .git; DenyPaths should be empty.
	assert.Empty(t, policy.Filesystem.DenyPaths)

	args, err := compileBwrapArgs(policy)
	require.NoError(t, err, "compileBwrapArgs must succeed for non-repo workspace")
	assertContainsPair(t, args, "--bind", workDir, workDir)
}

// TestCompileBwrapArgs_DefaultToolPolicy_GitFile verifies that a linked
// worktree (where .git is a file containing "gitdir: ...") does not cause
// compileBwrapArgs to fail. The isDir guard skips .git since bwrap --tmpfs
// requires a directory.
func TestCompileBwrapArgs_DefaultToolPolicy_GitFile(t *testing.T) {
	workDir := t.TempDir()
	gitFile := filepath.Join(workDir, ".git")
	require.NoError(t, os.WriteFile(gitFile, []byte("gitdir: /tmp/nowhere\n"), 0o600))

	policy := DefaultToolPolicy(workDir, "")

	// .git is a file, so isDir guard skips it.
	assert.NotContains(t, policy.Filesystem.DenyPaths, gitFile)

	args, err := compileBwrapArgs(policy)
	require.NoError(t, err, "compileBwrapArgs must succeed when .git is a file (worktree)")
	assertContainsPair(t, args, "--bind", workDir, workDir)
}

// TestCompileBwrapArgs_RootBindBeforeSpecialMounts verifies the load-bearing
// mount ordering: `--ro-bind / /` must appear BEFORE `--proc /proc`,
// `--dev /dev`, and `--tmpfs /run`. bubblewrap processes options
// left-to-right, so a later root bind would shadow the earlier specialised
// mounts and leak the host's /proc + /dev into the sandboxed child,
// weakening PID and device isolation. This test regression-guards the fix
// for the mount order bug Codex flagged after the first round of PR 4
// follow-up fixes.
func TestCompileBwrapArgs_RootBindBeforeSpecialMounts(t *testing.T) {
	policy := Policy{
		Filesystem: FilesystemPolicy{ReadOnlyGlobal: true},
		Network:    NetworkDeny,
		Process:    ProcessPolicy{AllowFork: true},
	}

	args, err := compileBwrapArgs(policy)
	require.NoError(t, err)

	rootBindIdx := -1
	procIdx := -1
	devIdx := -1
	runIdx := -1
	for i := 0; i < len(args)-2; i++ {
		if args[i] == "--ro-bind" && args[i+1] == "/" && args[i+2] == "/" {
			rootBindIdx = i
		}
		if args[i] == "--proc" && args[i+1] == "/proc" {
			procIdx = i
		}
		if args[i] == "--dev" && args[i+1] == "/dev" {
			devIdx = i
		}
		if args[i] == "--tmpfs" && args[i+1] == "/run" {
			runIdx = i
		}
	}

	require.NotEqual(t, -1, rootBindIdx, "missing --ro-bind / / in args: %v", args)
	require.NotEqual(t, -1, procIdx, "missing --proc /proc in args: %v", args)
	require.NotEqual(t, -1, devIdx, "missing --dev /dev in args: %v", args)
	require.NotEqual(t, -1, runIdx, "missing --tmpfs /run in args: %v", args)

	assert.Less(t, rootBindIdx, procIdx,
		"--ro-bind / / must come before --proc /proc so the root bind does not shadow procfs")
	assert.Less(t, rootBindIdx, devIdx,
		"--ro-bind / / must come before --dev /dev so the root bind does not shadow devfs")
	assert.Less(t, rootBindIdx, runIdx,
		"--ro-bind / / must come before --tmpfs /run so the root bind does not shadow /run")
}

// TestCompileBwrapArgs_DenyOverlapsWritePath verifies that when a deny path
// overlaps with a write path, the deny mount is emitted AFTER the write mount
// in the bwrap argv. bwrap applies mounts in order so the later --tmpfs wins.
func TestCompileBwrapArgs_DenyOverlapsWritePath(t *testing.T) {
	work := t.TempDir()
	// Create a sub-directory inside the writable workspace that will also be
	// listed as a deny path. This is the pattern used when a user's
	// AllowedWritePaths happens to include a child of dataRoot.
	denyChild := filepath.Join(work, "denied-child")
	require.NoError(t, os.Mkdir(denyChild, 0o755))

	policy := Policy{
		Filesystem: FilesystemPolicy{
			ReadOnlyGlobal: true,
			WritePaths:     []string{work},
			DenyPaths:      []string{denyChild},
		},
		Network: NetworkDeny,
	}

	args, err := compileBwrapArgs(policy)
	require.NoError(t, err)

	// Locate the --bind for the work dir and the --tmpfs for the deny child.
	bindIdx := -1
	denyIdx := -1
	for i := 0; i < len(args)-2; i++ {
		if args[i] == "--bind" && args[i+1] == work && args[i+2] == work {
			bindIdx = i
		}
		if args[i] == "--tmpfs" && args[i+1] == denyChild {
			denyIdx = i
		}
	}
	require.NotEqual(t, -1, bindIdx, "expected --bind for work dir, args=%v", args)
	require.NotEqual(t, -1, denyIdx, "expected --tmpfs for deny child, args=%v", args)
	assert.Greater(t, denyIdx, bindIdx,
		"deny mount must come after write mount so the later mount wins (deny precedence)")
}

func TestCompileBwrapArgs_NetworkUnixOnlyTreatedAsDeny(t *testing.T) {
	policy := Policy{
		Filesystem: FilesystemPolicy{ReadOnlyGlobal: true},
		Network:    NetworkUnixOnly,
	}

	args, err := compileBwrapArgs(policy)
	require.NoError(t, err)

	assert.Contains(t, args, "--unshare-net",
		"NetworkUnixOnly should currently be treated as deny under bwrap (no AF_UNIX-only filter)")
}

func TestCompileBwrapArgs_NetworkAllowOmitsUnshareNet(t *testing.T) {
	policy := Policy{
		Filesystem: FilesystemPolicy{ReadOnlyGlobal: true},
		Network:    NetworkAllow,
	}

	args, err := compileBwrapArgs(policy)
	require.NoError(t, err)

	assert.NotContains(t, args, "--unshare-net")
}

func TestCompileBwrapArgs_ReadPathsWhenNotGlobal(t *testing.T) {
	a := t.TempDir()
	b := t.TempDir()

	policy := Policy{
		Filesystem: FilesystemPolicy{
			ReadOnlyGlobal: false,
			ReadPaths:      []string{a, b},
		},
		Network: NetworkDeny,
	}

	args, err := compileBwrapArgs(policy)
	require.NoError(t, err)

	assert.NotContains(t, args, "/", "should not bind / when ReadOnlyGlobal=false")
	assertContainsPair(t, args, "--ro-bind", a, a)
	assertContainsPair(t, args, "--ro-bind", b, b)
}

func TestCompileBwrapArgs_RejectsInjectionInPaths(t *testing.T) {
	tests := []struct {
		give string
		path string
	}{
		{give: "semicolon", path: "/tmp/foo;bar"},
		{give: "newline", path: "/tmp/foo\nbar"},
		{give: "double quote", path: `/tmp/foo"bar`},
		{give: "open paren", path: "/tmp/foo(bar"},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			policy := Policy{
				Filesystem: FilesystemPolicy{
					ReadOnlyGlobal: true,
					WritePaths:     []string{tt.path},
				},
				Network: NetworkDeny,
			}
			_, err := compileBwrapArgs(policy)
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidPolicy)
		})
	}
}

func TestCompileBwrapArgs_DenyPathMustBeDirectory(t *testing.T) {
	work := t.TempDir()
	filePath := filepath.Join(work, "regular-file")
	require.NoError(t, os.WriteFile(filePath, []byte("hello"), 0o600))

	policy := Policy{
		Filesystem: FilesystemPolicy{
			ReadOnlyGlobal: true,
			DenyPaths:      []string{filePath},
		},
		Network: NetworkDeny,
	}

	_, err := compileBwrapArgs(policy)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be a directory")
}

func TestCompileBwrapArgs_DenyPathMissing(t *testing.T) {
	policy := Policy{
		Filesystem: FilesystemPolicy{
			ReadOnlyGlobal: true,
			DenyPaths:      []string{"/this/path/does/not/exist/lango-test"},
		},
		Network: NetworkDeny,
	}

	_, err := compileBwrapArgs(policy)
	require.Error(t, err)
}

func TestCompileBwrapArgs_EmptyPolicyKeepsStandardArgs(t *testing.T) {
	args, err := compileBwrapArgs(Policy{})
	require.NoError(t, err)

	assert.Contains(t, args, "--die-with-parent")
	assert.Contains(t, args, "--unshare-pid")
	assert.NotContains(t, args, "--ro-bind", "empty policy should not mount root")
	assert.NotContains(t, args, "--unshare-net", "empty policy has no Network — defaults to no flag")
}

// assertContainsPair checks that args contains [flag, a, b] as three consecutive
// elements (e.g. "--bind /workspace /workspace").
func assertContainsPair(t *testing.T, args []string, flag, a, b string) {
	t.Helper()
	for i := 0; i < len(args)-2; i++ {
		if args[i] == flag && args[i+1] == a && args[i+2] == b {
			return
		}
	}
	assert.Failf(t, "missing argument triple", "expected [%s %s %s] in args: %v", flag, a, b, args)
}

// assertContainsSingle checks that args contains [flag, a] as two consecutive
// elements (e.g. "--tmpfs /workspace/.git").
func assertContainsSingle(t *testing.T, args []string, flag, a string) {
	t.Helper()
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag && args[i+1] == a {
			return
		}
	}
	assert.Failf(t, "missing argument pair", "expected [%s %s] in args: %v", flag, a, args)
}
