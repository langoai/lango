package os

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultToolPolicy(t *testing.T) {
	policy := DefaultToolPolicy("/home/user/project")

	assert.True(t, policy.Filesystem.ReadOnlyGlobal)
	assert.Contains(t, policy.Filesystem.WritePaths, "/home/user/project")
	assert.Contains(t, policy.Filesystem.WritePaths, "/tmp")
	assert.Empty(t, policy.Filesystem.DenyPaths)
	assert.Equal(t, NetworkDeny, policy.Network)
	assert.True(t, policy.Process.AllowFork)
	assert.False(t, policy.Process.AllowSignals)
}

func TestStrictToolPolicy(t *testing.T) {
	policy := StrictToolPolicy("/home/user/project")

	assert.True(t, policy.Filesystem.ReadOnlyGlobal)
	assert.Contains(t, policy.Filesystem.WritePaths, "/home/user/project")
	assert.Contains(t, policy.Filesystem.DenyPaths, "/home/user/project/.git")
	assert.Equal(t, NetworkDeny, policy.Network)
}

func TestMCPServerPolicy(t *testing.T) {
	policy := MCPServerPolicy()

	assert.True(t, policy.Filesystem.ReadOnlyGlobal)
	assert.Contains(t, policy.Filesystem.WritePaths, "/tmp")
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
			give:       "default policy allows global read and denies network",
			givePolicy: DefaultToolPolicy("/tmp/work"),
			wantContains: []string{
				"(allow file-read*)",
				`(allow file-write* (subpath "/tmp/work"))`,
				`(allow file-write* (subpath "/tmp"))`,
				"(deny network*)",
				"(deny default)",
			},
		},
		{
			give:       "strict policy denies .git writes",
			givePolicy: StrictToolPolicy("/tmp/work"),
			wantContains: []string{
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
