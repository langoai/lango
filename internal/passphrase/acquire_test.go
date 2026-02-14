package passphrase

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcquire_Keyfile(t *testing.T) {
	dir := t.TempDir()
	keyfilePath := filepath.Join(dir, "keyfile")
	wantPass := "keyfile-passphrase"

	require.NoError(t, WriteKeyfile(keyfilePath, wantPass))

	got, source, err := Acquire(Options{KeyfilePath: keyfilePath})
	require.NoError(t, err)
	assert.Equal(t, wantPass, got)
	assert.Equal(t, SourceKeyfile, source)
}

func TestAcquire_KeyfilePriority(t *testing.T) {
	// When keyfile exists with valid permissions, it should be used
	// even if stdin is a pipe with data
	dir := t.TempDir()
	keyfilePath := filepath.Join(dir, "keyfile")
	wantPass := "keyfile-wins"

	require.NoError(t, WriteKeyfile(keyfilePath, wantPass))

	// Set up a pipe on stdin (simulating piped input)
	r, w, err := os.Pipe()
	require.NoError(t, err)

	origStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = origStdin })

	_, err = w.WriteString("stdin-passphrase\n")
	require.NoError(t, err)
	require.NoError(t, w.Close())

	got, source, err := Acquire(Options{KeyfilePath: keyfilePath})
	require.NoError(t, err)
	assert.Equal(t, wantPass, got)
	assert.Equal(t, SourceKeyfile, source)
}

func TestAcquire_StdinPipe(t *testing.T) {
	// No keyfile, stdin is a pipe — should read from stdin
	dir := t.TempDir()
	keyfilePath := filepath.Join(dir, "nonexistent-keyfile")

	r, w, err := os.Pipe()
	require.NoError(t, err)

	origStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = origStdin })

	wantPass := "stdin-passphrase"
	_, err = w.WriteString(wantPass + "\n")
	require.NoError(t, err)
	require.NoError(t, w.Close())

	got, source, err := Acquire(Options{KeyfilePath: keyfilePath})
	require.NoError(t, err)
	assert.Equal(t, wantPass, got)
	assert.Equal(t, SourceStdin, source)
}

func TestAcquire_InvalidKeyfilePermissions(t *testing.T) {
	// Keyfile exists but has wrong permissions — should fall through
	dir := t.TempDir()
	keyfilePath := filepath.Join(dir, "keyfile")

	require.NoError(t, os.WriteFile(keyfilePath, []byte("bad-perms\n"), 0644))

	r, w, err := os.Pipe()
	require.NoError(t, err)

	origStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = origStdin })

	wantPass := "fallback-stdin"
	_, err = w.WriteString(wantPass + "\n")
	require.NoError(t, err)
	require.NoError(t, w.Close())

	got, source, err := Acquire(Options{KeyfilePath: keyfilePath})
	require.NoError(t, err)
	assert.Equal(t, wantPass, got)
	assert.Equal(t, SourceStdin, source)
}

func TestAcquire_NoSourceAvailable(t *testing.T) {
	// No keyfile, stdin is a pipe but empty — should return error
	dir := t.TempDir()
	keyfilePath := filepath.Join(dir, "nonexistent-keyfile")

	r, w, err := os.Pipe()
	require.NoError(t, err)

	origStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = origStdin })

	require.NoError(t, w.Close())

	_, _, err = Acquire(Options{KeyfilePath: keyfilePath})
	assert.Error(t, err)
}

func TestDefaultKeyfilePath(t *testing.T) {
	got, err := defaultKeyfilePath()
	require.NoError(t, err)

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	want := filepath.Join(home, ".lango", "keyfile")
	assert.Equal(t, want, got)
}
