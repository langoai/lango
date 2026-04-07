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
	policy := DefaultToolPolicy(work)

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
}

func TestCompileBwrapArgs_StrictToolPolicy(t *testing.T) {
	work := t.TempDir()
	gitDir := filepath.Join(work, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0o755))

	policy := StrictToolPolicy(work)

	args, err := compileBwrapArgs(policy)
	require.NoError(t, err)

	assertContainsPair(t, args, "--ro-bind", "/", "/")
	assertContainsPair(t, args, "--bind", work, work)
	assertContainsSingle(t, args, "--tmpfs", gitDir)
}

func TestCompileBwrapArgs_MCPServerPolicy(t *testing.T) {
	policy := MCPServerPolicy()

	args, err := compileBwrapArgs(policy)
	require.NoError(t, err)

	assertContainsPair(t, args, "--ro-bind", "/", "/")
	assertContainsPair(t, args, "--bind", "/tmp", "/tmp")
	assert.NotContains(t, args, "--unshare-net", "MCPServerPolicy uses NetworkAllow — must not unshare net")
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
