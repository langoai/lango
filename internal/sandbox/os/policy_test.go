package os

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resolveSymlinks returns p with all symlinks resolved via filepath.EvalSymlinks,
// falling back to p if the call fails. Tests use this to express expectations
// in the canonical form that the normalizePath pipeline produces — on macOS,
// t.TempDir() returns /var/folders/... which resolves to /private/var/folders/...,
// and /tmp resolves to /private/tmp.
func resolveSymlinks(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		return p
	}
	return r
}

func TestDefaultToolPolicy(t *testing.T) {
	workDir := resolveSymlinks(t, t.TempDir())
	dataRoot := resolveSymlinks(t, t.TempDir())
	gitDir := filepath.Join(workDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0o755))

	policy := DefaultToolPolicy(workDir, dataRoot)

	assert.True(t, policy.Filesystem.ReadOnlyGlobal)
	assert.Contains(t, policy.Filesystem.WritePaths, workDir)
	assert.Contains(t, policy.Filesystem.WritePaths, "/tmp")
	// .git is denied as a baseline (was strict-only before).
	assert.Contains(t, policy.Filesystem.DenyPaths, gitDir)
	// Control-plane masking: dataRoot is denied so sandboxed children cannot
	// read or write the agent's own state.
	assert.Contains(t, policy.Filesystem.DenyPaths, dataRoot)
	assert.Equal(t, NetworkDeny, policy.Network)
	assert.True(t, policy.Process.AllowFork)
	assert.False(t, policy.Process.AllowSignals)
}

func TestDefaultToolPolicy_EmptyDataRoot(t *testing.T) {
	workDir := resolveSymlinks(t, t.TempDir())
	gitDir := filepath.Join(workDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0o755))

	// Empty dataRoot is allowed (used by isolated unit tests). The .git
	// baseline deny is still present because the directory exists.
	policy := DefaultToolPolicy(workDir, "")

	assert.Contains(t, policy.Filesystem.DenyPaths, gitDir)
	assert.Len(t, policy.Filesystem.DenyPaths, 1)
}

func TestDefaultToolPolicy_MissingGitNotDenied(t *testing.T) {
	// Non-repo workspace: workDir exists but has no .git. findGitRoot walks
	// up without finding any ancestor .git, so DenyPaths stays empty.
	workDir := resolveSymlinks(t, t.TempDir())

	policy := DefaultToolPolicy(workDir, "")

	assert.Empty(t, policy.Filesystem.DenyPaths,
		"non-repo workspace must produce an empty DenyPaths list")
}

func TestDefaultToolPolicy_GitFileNotDenied(t *testing.T) {
	// Linked worktree: .git is a regular file pointing to a nonexistent gitdir.
	// After PR 5c's two-deny strategy, the pointer file IS denied (via
	// file-level deny from Stage 1). The malformed gitdir target "/tmp/nowhere"
	// is still parsed and added — the resolve step succeeds to the absolute
	// path even though the target doesn't actually exist. For this test we use
	// a MALFORMED pointer (no "gitdir:" prefix) so only the pointer file is
	// added, making the DenyPaths length predictable.
	workDir := resolveSymlinks(t, t.TempDir())
	gitFile := filepath.Join(workDir, ".git")
	require.NoError(t, os.WriteFile(gitFile, []byte("this is not a gitdir pointer\n"), 0o600))

	policy := DefaultToolPolicy(workDir, "")

	// After PR 5c: the pointer file IS in DenyPaths (file-level deny).
	// No gitdir target was parsed (malformed), so only the pointer is denied.
	assert.Contains(t, policy.Filesystem.DenyPaths, gitFile,
		"malformed .git pointer file must be denied at the pointer path")
	assert.Len(t, policy.Filesystem.DenyPaths, 1,
		"malformed pointer degrades to pointer-only deny")
}

func TestFindGitRoot(t *testing.T) {
	t.Run("direct parent with .git directory", func(t *testing.T) {
		workDir := resolveSymlinks(t, t.TempDir())
		gitDir := filepath.Join(workDir, ".git")
		require.NoError(t, os.Mkdir(gitDir, 0o755))

		got := findGitRoot(workDir)
		assert.Equal(t, gitDir, got.pointerPath)
		assert.Equal(t, gitDir, got.gitdirPath,
			"standard repo should have pointerPath == gitdirPath")
		assert.True(t, got.found())
	})

	t.Run("nested subdirectory walks up to ancestor .git", func(t *testing.T) {
		// Simulate supervisor cwd = /repo/cmd/lango while .git lives at /repo/.git.
		root := resolveSymlinks(t, t.TempDir())
		gitDir := filepath.Join(root, ".git")
		require.NoError(t, os.Mkdir(gitDir, 0o755))
		nested := filepath.Join(root, "cmd", "lango")
		require.NoError(t, os.MkdirAll(nested, 0o755))

		got := findGitRoot(nested)
		assert.Equal(t, gitDir, got.pointerPath,
			"walk-up from a subdirectory must return the ancestor .git path")
		assert.Equal(t, gitDir, got.gitdirPath)
	})

	t.Run("deeply nested subdirectory walks up multiple levels", func(t *testing.T) {
		root := resolveSymlinks(t, t.TempDir())
		gitDir := filepath.Join(root, ".git")
		require.NoError(t, os.Mkdir(gitDir, 0o755))
		deep := filepath.Join(root, "a", "b", "c", "d", "e")
		require.NoError(t, os.MkdirAll(deep, 0o755))

		got := findGitRoot(deep)
		assert.Equal(t, gitDir, got.pointerPath)
	})

	t.Run("worktree pointer with absolute gitdir target", func(t *testing.T) {
		// .git file points to an absolute gitdir target. findGitRoot must
		// return both pointerPath (the file) and gitdirPath (the resolved
		// target). Canonical two-deny strategy per PR 5c.
		root := resolveSymlinks(t, t.TempDir())
		gitdirTarget := resolveSymlinks(t, t.TempDir())
		gitFile := filepath.Join(root, ".git")
		content := "gitdir: " + gitdirTarget + "\n"
		require.NoError(t, os.WriteFile(gitFile, []byte(content), 0o600))

		got := findGitRoot(root)
		assert.Equal(t, gitFile, got.pointerPath)
		assert.Equal(t, gitdirTarget, got.gitdirPath,
			"worktree gitdir target must be resolved to absolute path")
		assert.NotEqual(t, got.pointerPath, got.gitdirPath,
			"pointer and gitdir should differ for worktrees")
	})

	t.Run("worktree pointer with relative gitdir target", func(t *testing.T) {
		// Relative gitdir paths are resolved against the pointer file's parent.
		root := resolveSymlinks(t, t.TempDir())
		gitFile := filepath.Join(root, ".git")
		// Create a real target inside the tempdir that the relative path
		// resolves to.
		targetDir := filepath.Join(root, "real-gitdir")
		require.NoError(t, os.Mkdir(targetDir, 0o755))
		require.NoError(t, os.WriteFile(gitFile, []byte("gitdir: real-gitdir\n"), 0o600))

		got := findGitRoot(root)
		assert.Equal(t, gitFile, got.pointerPath)
		assert.Equal(t, targetDir, got.gitdirPath,
			"relative gitdir target must resolve against pointer file parent")
	})

	t.Run("worktree pointer malformed degrades to pointer-only", func(t *testing.T) {
		// Pointer file lacks "gitdir:" prefix. findGitRoot returns pointer path
		// but empty gitdirPath — caller still gets file-level deny of the
		// pointer file itself (partial protection).
		root := resolveSymlinks(t, t.TempDir())
		gitFile := filepath.Join(root, ".git")
		require.NoError(t, os.WriteFile(gitFile, []byte("not a pointer\n"), 0o600))

		got := findGitRoot(root)
		assert.Equal(t, gitFile, got.pointerPath)
		assert.Empty(t, got.gitdirPath,
			"malformed pointer must have empty gitdirPath (degraded)")
		assert.True(t, got.found())
	})

	t.Run("empty workDir returns zero gitRoot", func(t *testing.T) {
		got := findGitRoot("")
		assert.False(t, got.found())
		assert.Empty(t, got.pointerPath)
		assert.Empty(t, got.gitdirPath)
	})

	t.Run("filesystem root terminates without panic", func(t *testing.T) {
		// The walk MUST terminate when filepath.Dir(cur)==cur. A regression
		// in that check would cause an infinite loop; the test would hang
		// rather than fail, but assert.NotPanics documents the expectation.
		assert.NotPanics(t, func() {
			_ = findGitRoot("/")
		})
	})

	t.Run("symlinked workDir resolves to canonical path", func(t *testing.T) {
		// Create a real repo and a symlink to its cmd subdirectory. findGitRoot
		// should resolve the symlink and find the .git at the real ancestor.
		realRoot := resolveSymlinks(t, t.TempDir())
		gitDir := filepath.Join(realRoot, ".git")
		require.NoError(t, os.Mkdir(gitDir, 0o755))
		realCmd := filepath.Join(realRoot, "cmd")
		require.NoError(t, os.Mkdir(realCmd, 0o755))

		// Create a symlink elsewhere that points to realCmd.
		symlinkParent := resolveSymlinks(t, t.TempDir())
		symLink := filepath.Join(symlinkParent, "link-to-cmd")
		require.NoError(t, os.Symlink(realCmd, symLink))

		got := findGitRoot(symLink)
		assert.Equal(t, gitDir, got.pointerPath,
			"symlinked workDir must resolve to canonical path before walk-up")
	})
}

func TestDefaultToolPolicy_WalksUpToGitRoot(t *testing.T) {
	// Regression guard: supervisor/skill executor may pass a subdirectory as
	// workDir while .git lives at the repository root. DefaultToolPolicy must
	// deny the ancestor .git, not silently skip because `workDir/.git` does
	// not exist.
	root := resolveSymlinks(t, t.TempDir())
	gitDir := filepath.Join(root, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0o755))
	nested := filepath.Join(root, "cmd", "lango")
	require.NoError(t, os.MkdirAll(nested, 0o755))

	policy := DefaultToolPolicy(nested, "")

	assert.Contains(t, policy.Filesystem.DenyPaths, gitDir,
		"walk-up must discover the ancestor .git and add it to DenyPaths")
	assert.NotContains(t, policy.Filesystem.DenyPaths, filepath.Join(nested, ".git"),
		"DenyPaths must not contain a fictional nested .git path")
}

func TestDefaultToolPolicy_WorktreeDenyBothPointerAndGitdir(t *testing.T) {
	// Linked worktree regression guard: when .git is a regular file pointing
	// to an existing gitdir target, DefaultToolPolicy must deny BOTH the
	// pointer file AND the resolved gitdir target (two-deny strategy).
	root := resolveSymlinks(t, t.TempDir())
	gitdirTarget := resolveSymlinks(t, t.TempDir())
	gitFile := filepath.Join(root, ".git")
	content := "gitdir: " + gitdirTarget + "\n"
	require.NoError(t, os.WriteFile(gitFile, []byte(content), 0o600))

	policy := DefaultToolPolicy(root, "")

	assert.Contains(t, policy.Filesystem.DenyPaths, gitFile,
		"worktree pointer file must be in DenyPaths (file-level deny)")
	assert.Contains(t, policy.Filesystem.DenyPaths, gitdirTarget,
		"worktree gitdir target must also be in DenyPaths (two-deny strategy)")
	assert.Len(t, policy.Filesystem.DenyPaths, 2,
		"worktree should produce exactly two DenyPaths entries")
}

func TestDefaultToolPolicy_MissingDataRootNotDenied(t *testing.T) {
	// Missing dataRoot is silently skipped even when non-empty, so that
	// minimal environments (e.g. during initial setup) can still build a
	// policy without the control-plane mask rather than failing outright.
	workDir := resolveSymlinks(t, t.TempDir())
	missingDataRoot := filepath.Join(t.TempDir(), "does-not-exist")

	policy := DefaultToolPolicy(workDir, missingDataRoot)

	assert.NotContains(t, policy.Filesystem.DenyPaths, missingDataRoot)
}

func TestStrictToolPolicy(t *testing.T) {
	// StrictToolPolicy is currently identical to DefaultToolPolicy — kept as a
	// separate symbol so future strict-only options can branch without another
	// signature migration.
	workDir := resolveSymlinks(t, t.TempDir())
	dataRoot := resolveSymlinks(t, t.TempDir())
	require.NoError(t, os.Mkdir(filepath.Join(workDir, ".git"), 0o755))

	policy := StrictToolPolicy(workDir, dataRoot)
	defaultPolicy := DefaultToolPolicy(workDir, dataRoot)
	assert.Equal(t, defaultPolicy, policy)
}

func TestMCPServerPolicy(t *testing.T) {
	dataRoot := resolveSymlinks(t, t.TempDir())

	// Empty workspacePath — this test checks the dataRoot deny shape only.
	policy := MCPServerPolicy("", dataRoot)

	assert.True(t, policy.Filesystem.ReadOnlyGlobal)
	assert.Contains(t, policy.Filesystem.WritePaths, "/tmp")
	// MCP server children are also blocked from reading the lango control-plane.
	assert.Contains(t, policy.Filesystem.DenyPaths, dataRoot)
	assert.Equal(t, NetworkAllow, policy.Network)
}

func TestDefaultToolPolicyWithProtectedPaths(t *testing.T) {
	workDir := t.TempDir()
	protected := filepath.Join(t.TempDir(), "graph.db")

	policy := DefaultToolPolicyWithProtectedPaths(workDir, "", []string{protected})
	assert.Contains(t, policy.Filesystem.DenyPaths, protected)
}

func TestMCPServerPolicy_EmptyDataRoot(t *testing.T) {
	policy := MCPServerPolicy("", "")

	assert.True(t, policy.Filesystem.ReadOnlyGlobal)
	assert.Empty(t, policy.Filesystem.DenyPaths)
	assert.Equal(t, NetworkAllow, policy.Network)
}

func TestMCPServerPolicy_MissingDataRoot(t *testing.T) {
	// Non-empty but missing dataRoot: silently skipped by the isDir guard.
	missing := filepath.Join(t.TempDir(), "does-not-exist")

	policy := MCPServerPolicy("", missing)

	assert.Empty(t, policy.Filesystem.DenyPaths)
	assert.Equal(t, NetworkAllow, policy.Network)
}

func TestMCPServerPolicy_DenyWorkspaceGit(t *testing.T) {
	// Regression guard: MCPServerPolicy must now apply the same .git walk-up
	// deny as DefaultToolPolicy so MCP children cannot read git metadata.
	root := resolveSymlinks(t, t.TempDir())
	gitDir := filepath.Join(root, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0o755))
	nested := filepath.Join(root, "cmd", "lango")
	require.NoError(t, os.MkdirAll(nested, 0o755))

	policy := MCPServerPolicy(nested, "")

	assert.Contains(t, policy.Filesystem.DenyPaths, gitDir,
		"MCPServerPolicy must walk up to find ancestor .git")
	assert.Equal(t, NetworkAllow, policy.Network)
}

func TestMCPServerPolicy_WorkspaceGitPlusDataRoot(t *testing.T) {
	// Both baseline denies present simultaneously: walk-up .git AND dataRoot.
	root := resolveSymlinks(t, t.TempDir())
	gitDir := filepath.Join(root, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0o755))
	dataRoot := resolveSymlinks(t, t.TempDir())

	policy := MCPServerPolicy(root, dataRoot)

	assert.Contains(t, policy.Filesystem.DenyPaths, gitDir)
	assert.Contains(t, policy.Filesystem.DenyPaths, dataRoot)
	assert.Len(t, policy.Filesystem.DenyPaths, 2)
}

func TestGenerateSeatbeltProfile(t *testing.T) {
	tests := []struct {
		give            string
		givePolicy      Policy
		wantContains    []string
		wantNotContains []string
		wantErr         bool
	}{
		{
			give: "default-shape policy allows global read and denies network",
			givePolicy: Policy{
				Filesystem: FilesystemPolicy{
					ReadOnlyGlobal: true,
					WritePaths:     []string{"/tmp/work", "/tmp/sandbox-extra"},
				},
				Network: NetworkDeny,
				Process: ProcessPolicy{AllowFork: true},
			},
			wantContains: []string{
				"(allow file-read*)",
				`(allow file-write* (subpath "/tmp/work"))`,
				`(allow file-write* (subpath "/tmp/sandbox-extra"))`,
				"(deny network*)",
				"(deny default)",
			},
		},
		{
			give: "default-shape policy denies .git writes",
			givePolicy: Policy{
				Filesystem: FilesystemPolicy{
					ReadOnlyGlobal: true,
					WritePaths:     []string{"/tmp/work", "/tmp/sandbox-extra"},
					DenyPaths:      []string{"/tmp/work/.git"},
				},
				Network: NetworkDeny,
				Process: ProcessPolicy{AllowFork: true},
			},
			wantContains: []string{
				`(deny file-write* (subpath "/tmp/work/.git"))`,
			},
		},
		{
			give: "default-shape policy denies dataRoot when provided",
			givePolicy: Policy{
				Filesystem: FilesystemPolicy{
					ReadOnlyGlobal: true,
					WritePaths:     []string{"/tmp/work", "/tmp/sandbox-extra"},
					DenyPaths:      []string{"/tmp/work/.git", "/home/user/.lango"},
				},
				Network: NetworkDeny,
				Process: ProcessPolicy{AllowFork: true},
			},
			wantContains: []string{
				`(deny file-write* (subpath "/home/user/.lango"))`,
			},
		},
		{
			give: "DenyPaths entries deny both read and write",
			givePolicy: Policy{
				Filesystem: FilesystemPolicy{
					ReadOnlyGlobal: true,
					DenyPaths:      []string{"/home/user/.lango"},
				},
				Network: NetworkDeny,
				Process: ProcessPolicy{AllowFork: true},
			},
			wantContains: []string{
				`(deny file-read* (subpath "/home/user/.lango"))`,
				`(deny file-write* (subpath "/home/user/.lango"))`,
			},
		},
		{
			give: "multiple DenyPaths each get both read and write deny",
			givePolicy: Policy{
				Filesystem: FilesystemPolicy{
					ReadOnlyGlobal: true,
					DenyPaths:      []string{"/home/user/.lango", "/tmp/work/.git"},
				},
				Network: NetworkDeny,
				Process: ProcessPolicy{AllowFork: true},
			},
			wantContains: []string{
				`(deny file-read* (subpath "/home/user/.lango"))`,
				`(deny file-write* (subpath "/home/user/.lango"))`,
				`(deny file-read* (subpath "/tmp/work/.git"))`,
				`(deny file-write* (subpath "/tmp/work/.git"))`,
			},
		},
		{
			give: "allow network mode",
			givePolicy: Policy{
				Filesystem: FilesystemPolicy{ReadOnlyGlobal: true},
				Network:    NetworkAllow,
				Process:    ProcessPolicy{AllowFork: true},
			},
			wantContains: []string{
				"(allow network*)",
			},
			wantNotContains: []string{
				"(deny network*)",
			},
		},
		{
			give: "unix-only network mode",
			givePolicy: Policy{
				Filesystem: FilesystemPolicy{ReadOnlyGlobal: true},
				Network:    NetworkUnixOnly,
				Process:    ProcessPolicy{AllowFork: true},
			},
			wantContains: []string{
				"(deny network*)",
				"(allow network* (local unix))",
			},
		},
		{
			give: "path with injection characters fails",
			givePolicy: Policy{
				Filesystem: FilesystemPolicy{
					WritePaths: []string{`/tmp/bad"path`},
				},
			},
			wantErr: true,
		},
		{
			give: "allowed IPs included in profile",
			givePolicy: Policy{
				Filesystem:        FilesystemPolicy{ReadOnlyGlobal: true},
				Network:           NetworkDeny,
				AllowedNetworkIPs: []string{"192.168.1.1", "10.0.0.1"},
				Process:           ProcessPolicy{AllowFork: true},
			},
			wantContains: []string{
				`(allow network-outbound (remote ip "192.168.1.1:*"))`,
				`(allow network-outbound (remote ip "10.0.0.1:*"))`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			profile, err := GenerateSeatbeltProfile(tt.givePolicy)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			for _, want := range tt.wantContains {
				assert.Contains(t, profile, want)
			}
			for _, notWant := range tt.wantNotContains {
				assert.NotContains(t, profile, notWant)
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		give    string
		wantErr bool
	}{
		{give: "/tmp/valid", wantErr: false},
		{give: "/tmp/also-valid_123", wantErr: false},
		{give: `/tmp/has"quote`, wantErr: true},
		{give: "/tmp/has(paren)", wantErr: true},
		{give: "/tmp/has;semi", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			_, err := sanitizePath(tt.give)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateIP(t *testing.T) {
	tests := []struct {
		give    string
		wantErr bool
	}{
		{give: "192.168.1.1", wantErr: false},
		{give: "10.0.0.1", wantErr: false},
		{give: "::1", wantErr: false},
		{give: "fe80::1", wantErr: false},
		{give: "", wantErr: true},
		{give: "evil;cmd", wantErr: true},
		{give: "192.168.1.1/24", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			err := validateIP(tt.give)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProbe(t *testing.T) {
	caps := Probe()
	assert.NotEmpty(t, caps.Platform)
	// Platform-specific: at least one of the capabilities should be detected on supported OS.
	assert.NotEmpty(t, caps.Summary())
}

func TestPlatformCapabilities_Summary(t *testing.T) {
	tests := []struct {
		give PlatformCapabilities
		want string
	}{
		{
			give: PlatformCapabilities{HasSeatbelt: true},
			want: "seatbelt (macOS)",
		},
		{
			give: PlatformCapabilities{HasLandlock: true, HasSeccomp: true},
			want: "landlock+seccomp (Linux)",
		},
		{
			give: PlatformCapabilities{HasLandlock: true},
			want: "landlock (Linux, no seccomp)",
		},
		{
			give: PlatformCapabilities{HasSeccomp: true},
			want: "seccomp (Linux, no landlock)",
		},
		{
			give: PlatformCapabilities{},
			want: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.give.Summary())
		})
	}
}
