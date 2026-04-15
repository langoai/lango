package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveExtensionsDefaults(t *testing.T) {
	t.Parallel()

	got := ExtensionsConfig{}.ResolveExtensions()
	assert.NotNil(t, got.Enabled)
	assert.True(t, *got.Enabled)
	assert.Equal(t, DefaultExtensionsDir, got.Dir)
	assert.False(t, got.EnforceIntegrity)
}

func TestResolveExtensionsPreservesUser(t *testing.T) {
	t.Parallel()

	f := false
	in := ExtensionsConfig{Enabled: &f, Dir: "/data/packs", EnforceIntegrity: true}
	got := in.ResolveExtensions()

	assert.False(t, *got.Enabled)
	assert.Equal(t, "/data/packs", got.Dir)
	assert.True(t, got.EnforceIntegrity)
}

func TestResolveExtensionsDoesNotMutateReceiver(t *testing.T) {
	t.Parallel()

	in := ExtensionsConfig{}
	_ = in.ResolveExtensions()
	assert.Nil(t, in.Enabled, "receiver must not be mutated")
	assert.Empty(t, in.Dir)
}

func TestResolvedDirExpandsTilde(t *testing.T) {
	t.Parallel()

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home dir: %v", err)
	}

	c := ExtensionsConfig{Dir: "~/foo/bar"}
	assert.Equal(t, filepath.Join(home, "foo/bar"), c.ResolvedDir())
}

func TestResolvedDirPreservesAbsolute(t *testing.T) {
	t.Parallel()

	c := ExtensionsConfig{Dir: "/abs/path"}
	assert.Equal(t, "/abs/path", c.ResolvedDir())
}

func TestResolvedDirDefaultsWhenEmpty(t *testing.T) {
	t.Parallel()

	home, _ := os.UserHomeDir()
	got := ExtensionsConfig{}.ResolvedDir()
	if home != "" {
		assert.Equal(t, filepath.Join(home, ".lango/extensions"), got)
	} else {
		assert.Equal(t, DefaultExtensionsDir, got)
	}
}

func TestIsEnabledDefaultTrue(t *testing.T) {
	t.Parallel()

	assert.True(t, ExtensionsConfig{}.IsEnabled())

	f := false
	assert.False(t, ExtensionsConfig{Enabled: &f}.IsEnabled())

	tr := true
	assert.True(t, ExtensionsConfig{Enabled: &tr}.IsEnabled())
}
