package os

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompileBwrapArgs_DefaultToolPolicy(t *testing.T) {
	work := resolveSymlinks(t, t.TempDir())
	dataRoot := resolveSymlinks(t, t.TempDir())
	// .git baseline deny is now part of DefaultToolPolicy. compileBwrapArgs
	// requires deny paths to exist as directories.
	gitDir := filepath.Join(work, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0o755))

	policy := DefaultToolPolicy(work, dataRoot)

	args, err := compileBwrapArgs(policy)
	require.NoError(t, err)

	tmpResolved := resolveSymlinks(t, "/tmp")
	assertContainsPair(t, args, "--ro-bind", "/", "/")
	assertContainsPair(t, args, "--bind", work, work)
	assertContainsPair(t, args, "--bind", tmpResolved, tmpResolved)
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
	work := resolveSymlinks(t, t.TempDir())
	dataRoot := resolveSymlinks(t, t.TempDir())
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
	dataRoot := resolveSymlinks(t, t.TempDir())
	// Empty workspacePath — this test focuses on the dataRoot deny shape.
	// A dedicated test (TestMCPServerPolicy_DenyWorkspaceGit) covers walk-up.
	policy := MCPServerPolicy("", dataRoot)

	args, err := compileBwrapArgs(policy)
	require.NoError(t, err)

	tmpResolved := resolveSymlinks(t, "/tmp")
	assertContainsPair(t, args, "--ro-bind", "/", "/")
	assertContainsPair(t, args, "--bind", tmpResolved, tmpResolved)
	assert.NotContains(t, args, "--unshare-net", "MCPServerPolicy uses NetworkAllow — must not unshare net")
	assertContainsSingle(t, args, "--tmpfs", dataRoot)
}

// TestCompileBwrapArgs_DefaultToolPolicy_NoGitDir verifies that a non-repo
// workspace (no .git directory) does not cause compileBwrapArgs to fail.
// Regression: PR 4 added .git unconditionally to DenyPaths, which combined
// with bwrap's strict stat+IsDir check rejected every non-repo workspace.
func TestCompileBwrapArgs_DefaultToolPolicy_NoGitDir(t *testing.T) {
	workDir := resolveSymlinks(t, t.TempDir()) // no .git
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
	workDir := resolveSymlinks(t, t.TempDir())
	gitFile := filepath.Join(workDir, ".git")
	// Malformed pointer — no "gitdir:" prefix — so findGitRoot returns a
	// pointer-only gitRoot and DefaultToolPolicy denies only the file itself.
	// compileBwrapArgs's file-level deny (PR 5c Stage 1) handles the file.
	require.NoError(t, os.WriteFile(gitFile, []byte("not a pointer\n"), 0o600))

	policy := DefaultToolPolicy(workDir, "")

	// .git pointer file is denied at file level (--ro-bind /dev/null).
	assert.Contains(t, policy.Filesystem.DenyPaths, gitFile)

	args, err := compileBwrapArgs(policy)
	require.NoError(t, err, "compileBwrapArgs must succeed when .git is a malformed worktree pointer")
	assertContainsPair(t, args, "--bind", workDir, workDir)
	assertContainsPair(t, args, "--ro-bind", "/dev/null", gitFile)
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
	work := resolveSymlinks(t, t.TempDir())
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
	a := resolveSymlinks(t, t.TempDir())
	b := resolveSymlinks(t, t.TempDir())

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

// TestCompileBwrapArgs_DenyPathFileGetsRoBindDevNull verifies that a regular
// file in DenyPaths is translated to `--ro-bind /dev/null <file>` so the
// sandboxed child sees EOF on read and EACCES on write. This closes the
// file-level deny gap: PR 3/4 rejected files with "must be a directory";
// PR 5c supports them via the /dev/null bind trick.
func TestCompileBwrapArgs_DenyPathFileGetsRoBindDevNull(t *testing.T) {
	work := resolveSymlinks(t, t.TempDir())
	filePath := filepath.Join(work, "regular-file")
	require.NoError(t, os.WriteFile(filePath, []byte("hello"), 0o600))

	policy := Policy{
		Filesystem: FilesystemPolicy{
			ReadOnlyGlobal: true,
			DenyPaths:      []string{filePath},
		},
		Network: NetworkDeny,
	}

	args, err := compileBwrapArgs(policy)
	require.NoError(t, err, "regular file should no longer error; expected --ro-bind /dev/null")
	assertContainsPair(t, args, "--ro-bind", "/dev/null", filePath)
	// Must NOT be emitted as --tmpfs (that's the directory path).
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--tmpfs" && args[i+1] == filePath {
			t.Fatalf("regular file deny should not use --tmpfs: %v", args)
		}
	}
}

// TestCompileBwrapArgs_DenyPathDirectoryStillGetsTmpfs verifies directories
// still use --tmpfs (the pre-5c behavior). Regression guard against an
// accidental flip to ro-bind for directories.
func TestCompileBwrapArgs_DenyPathDirectoryStillGetsTmpfs(t *testing.T) {
	work := resolveSymlinks(t, t.TempDir())
	dirPath := filepath.Join(work, "denied-dir")
	require.NoError(t, os.Mkdir(dirPath, 0o755))

	policy := Policy{
		Filesystem: FilesystemPolicy{
			ReadOnlyGlobal: true,
			DenyPaths:      []string{dirPath},
		},
		Network: NetworkDeny,
	}

	args, err := compileBwrapArgs(policy)
	require.NoError(t, err)
	assertContainsSingle(t, args, "--tmpfs", dirPath)
}

// TestCompileBwrapArgs_DenyPathUnsupportedMode verifies that non-regular,
// non-directory deny paths (device, socket, fifo) are rejected with a clear
// error. /dev/null is a character device on all POSIX systems and is a
// portable test target — the exact test works on both Linux and macOS.
func TestCompileBwrapArgs_DenyPathUnsupportedMode(t *testing.T) {
	// /dev/null exists on Linux and macOS as a character device; it's the
	// most portable non-regular, non-directory file on supported platforms.
	const devNode = "/dev/null"
	fi, err := os.Stat(devNode)
	if err != nil || fi.Mode().IsDir() || fi.Mode().IsRegular() {
		t.Skipf("%s is not a special file on this platform (mode=%v)", devNode, fi.Mode())
	}

	policy := Policy{
		Filesystem: FilesystemPolicy{
			ReadOnlyGlobal: true,
			DenyPaths:      []string{devNode},
		},
		Network: NetworkDeny,
	}

	_, compileErr := compileBwrapArgs(policy)
	require.Error(t, compileErr)
	assert.Contains(t, compileErr.Error(), "unsupported file mode")
}

// TestCompileBwrapArgs_SymlinkedDenyPath verifies that a DenyPath pointing
// to a symlink is resolved to the target BEFORE the file-vs-dir classification.
// Without EvalSymlinks, bwrap would mount on the symlink itself, leaving the
// real target unprotected.
func TestCompileBwrapArgs_SymlinkedDenyPath(t *testing.T) {
	work := resolveSymlinks(t, t.TempDir())
	realDir := filepath.Join(work, "real-secrets")
	require.NoError(t, os.Mkdir(realDir, 0o755))
	link := filepath.Join(work, "secret-link")
	require.NoError(t, os.Symlink(realDir, link))

	policy := Policy{
		Filesystem: FilesystemPolicy{
			ReadOnlyGlobal: true,
			DenyPaths:      []string{link},
		},
		Network: NetworkDeny,
	}

	args, err := compileBwrapArgs(policy)
	require.NoError(t, err)
	// The deny should be emitted at the RESOLVED path (realDir), not the
	// symlink itself — symlink escape is closed.
	assertContainsSingle(t, args, "--tmpfs", realDir)
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--tmpfs" && args[i+1] == link {
			t.Fatalf("deny should use resolved target, not symlink: %v", args)
		}
	}
}

// TestNormalizePath_NonexistentFallback verifies the EvalSymlinks fallback:
// when the path doesn't exist, normalizePath returns the pre-resolve Abs path
// so downstream os.Stat catches the missing-path error with the existing
// error message format (preserved UX for "user typo'd a path").
func TestNormalizePath_NonexistentFallback(t *testing.T) {
	missing := "/this/path/does/not/exist/lango-normalize-test"
	results, err := normalizePath(missing)
	require.NoError(t, err, "nonexistent path must not error during normalizePath")
	require.Len(t, results, 1)
	assert.Equal(t, missing, results[0],
		"fallback must return the pre-resolve Abs path unchanged")
}

// TestNormalizePath_GlobExpansion verifies canonical pipeline step 3:
// wildcards expand via filepath.Glob, each match flows through the rest
// of the pipeline independently.
func TestNormalizePath_GlobExpansion(t *testing.T) {
	work := resolveSymlinks(t, t.TempDir())
	// Create three .db files and one .txt file — the pattern should match
	// exactly the three .db files.
	for _, name := range []string{"one.db", "two.db", "three.db", "ignored.txt"} {
		require.NoError(t, os.WriteFile(filepath.Join(work, name), []byte("x"), 0o600))
	}
	pattern := filepath.Join(work, "*.db")

	results, err := normalizePath(pattern)
	require.NoError(t, err)
	assert.Len(t, results, 3, "glob should expand to 3 .db files")
	for _, r := range results {
		assert.True(t, strings.HasSuffix(r, ".db"),
			"every expanded match should end with .db: %s", r)
	}
}

// TestNormalizePath_UnmatchedGlobSilentSkip verifies that a glob pattern
// with zero matches returns an empty slice + nil error (shell nullglob
// semantics, matching the best-effort baseline deny philosophy).
func TestNormalizePath_UnmatchedGlobSilentSkip(t *testing.T) {
	work := resolveSymlinks(t, t.TempDir())
	pattern := filepath.Join(work, "no-such-*.foo")

	results, err := normalizePath(pattern)
	require.NoError(t, err, "unmatched glob must not error")
	assert.Empty(t, results, "unmatched glob must return empty slice (silent skip)")
}

// TestNormalizePath_InvalidGlobErrors verifies that an invalid glob pattern
// (unclosed bracket) surfaces as an error — user config mistakes should
// fail loudly at startup, not be silently dropped.
func TestNormalizePath_InvalidGlobErrors(t *testing.T) {
	// filepath.Glob only returns ErrBadPattern for syntactically invalid
	// patterns like an unclosed bracket.
	_, err := normalizePath("/tmp/[unclosed")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid glob pattern")
}

// TestCompileBwrapArgs_DenyPathWithGlob verifies that a glob pattern in
// DenyPaths expands and each match becomes its own emission.
func TestCompileBwrapArgs_DenyPathWithGlob(t *testing.T) {
	work := resolveSymlinks(t, t.TempDir())
	for _, name := range []string{"a.log", "b.log", "c.log"} {
		require.NoError(t, os.WriteFile(filepath.Join(work, name), []byte("x"), 0o600))
	}
	pattern := filepath.Join(work, "*.log")

	policy := Policy{
		Filesystem: FilesystemPolicy{
			ReadOnlyGlobal: true,
			DenyPaths:      []string{pattern},
		},
		Network: NetworkDeny,
	}

	args, err := compileBwrapArgs(policy)
	require.NoError(t, err)
	// Each matched file should be emitted as its own --ro-bind /dev/null deny.
	for _, name := range []string{"a.log", "b.log", "c.log"} {
		assertContainsPair(t, args, "--ro-bind", "/dev/null", filepath.Join(work, name))
	}
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
