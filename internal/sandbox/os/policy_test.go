package os

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultToolPolicy(t *testing.T) {
	workDir := t.TempDir()
	dataRoot := t.TempDir()
	// .git baseline deny requires the directory to exist; otherwise the
	// isDir guard silently skips it.
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
	workDir := t.TempDir()
	gitDir := filepath.Join(workDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0o755))

	// Empty dataRoot is allowed (used by isolated unit tests). The .git
	// baseline deny is still present because the directory exists.
	policy := DefaultToolPolicy(workDir, "")

	assert.Contains(t, policy.Filesystem.DenyPaths, gitDir)
	assert.Len(t, policy.Filesystem.DenyPaths, 1)
}

func TestDefaultToolPolicy_MissingGitNotDenied(t *testing.T) {
	// Non-repo workspace: workDir exists but has no .git. The isDir guard
	// must skip .git so that compileBwrapArgs does not fail on a missing
	// deny path.
	workDir := t.TempDir()

	policy := DefaultToolPolicy(workDir, "")

	assert.Empty(t, policy.Filesystem.DenyPaths,
		"non-repo workspace must produce an empty DenyPaths list")
}

func TestDefaultToolPolicy_GitFileNotDenied(t *testing.T) {
	// Linked worktree: .git is a regular file containing "gitdir: <path>".
	// The isDir guard must skip it since bwrap --tmpfs requires a directory.
	workDir := t.TempDir()
	gitFile := filepath.Join(workDir, ".git")
	require.NoError(t, os.WriteFile(gitFile, []byte("gitdir: /tmp/nowhere\n"), 0o600))

	policy := DefaultToolPolicy(workDir, "")

	assert.NotContains(t, policy.Filesystem.DenyPaths, gitFile,
		".git file (worktree) must not be added to DenyPaths")
	assert.Empty(t, policy.Filesystem.DenyPaths)
}

func TestDefaultToolPolicy_MissingDataRootNotDenied(t *testing.T) {
	// Missing dataRoot is silently skipped even when non-empty, so that
	// minimal environments (e.g. during initial setup) can still build a
	// policy without the control-plane mask rather than failing outright.
	workDir := t.TempDir()
	missingDataRoot := filepath.Join(t.TempDir(), "does-not-exist")

	policy := DefaultToolPolicy(workDir, missingDataRoot)

	assert.NotContains(t, policy.Filesystem.DenyPaths, missingDataRoot)
}

func TestStrictToolPolicy(t *testing.T) {
	// StrictToolPolicy is currently identical to DefaultToolPolicy — kept as a
	// separate symbol so future strict-only options can branch without another
	// signature migration.
	workDir := t.TempDir()
	dataRoot := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(workDir, ".git"), 0o755))

	policy := StrictToolPolicy(workDir, dataRoot)
	defaultPolicy := DefaultToolPolicy(workDir, dataRoot)
	assert.Equal(t, defaultPolicy, policy)
}

func TestMCPServerPolicy(t *testing.T) {
	dataRoot := t.TempDir()

	policy := MCPServerPolicy(dataRoot)

	assert.True(t, policy.Filesystem.ReadOnlyGlobal)
	assert.Contains(t, policy.Filesystem.WritePaths, "/tmp")
	// MCP server children are also blocked from reading the lango control-plane.
	assert.Contains(t, policy.Filesystem.DenyPaths, dataRoot)
	assert.Equal(t, NetworkAllow, policy.Network)
}

func TestMCPServerPolicy_EmptyDataRoot(t *testing.T) {
	policy := MCPServerPolicy("")

	assert.True(t, policy.Filesystem.ReadOnlyGlobal)
	assert.Empty(t, policy.Filesystem.DenyPaths)
	assert.Equal(t, NetworkAllow, policy.Network)
}

func TestMCPServerPolicy_MissingDataRoot(t *testing.T) {
	// Non-empty but missing dataRoot: silently skipped by the isDir guard.
	missing := filepath.Join(t.TempDir(), "does-not-exist")

	policy := MCPServerPolicy(missing)

	assert.Empty(t, policy.Filesystem.DenyPaths)
	assert.Equal(t, NetworkAllow, policy.Network)
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
					WritePaths:     []string{"/tmp/work", "/tmp"},
				},
				Network: NetworkDeny,
				Process: ProcessPolicy{AllowFork: true},
			},
			wantContains: []string{
				"(allow file-read*)",
				`(allow file-write* (subpath "/tmp/work"))`,
				`(allow file-write* (subpath "/tmp"))`,
				"(deny network*)",
				"(deny default)",
			},
		},
		{
			give: "default-shape policy denies .git writes",
			givePolicy: Policy{
				Filesystem: FilesystemPolicy{
					ReadOnlyGlobal: true,
					WritePaths:     []string{"/tmp/work", "/tmp"},
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
					WritePaths:     []string{"/tmp/work", "/tmp"},
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
