package passphrase

import (
	"bytes"
	"errors"
	"io"
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

	got, source, err := acquireWithIO(Options{KeyfilePath: keyfilePath}, bytes.NewBufferString("stdin-passphrase\n"), io.Discard, false)
	require.NoError(t, err)
	assert.Equal(t, wantPass, got)
	assert.Equal(t, SourceKeyfile, source)
}

func TestAcquire_StdinPipe(t *testing.T) {
	// No keyfile, stdin is a pipe — should read from stdin
	dir := t.TempDir()
	keyfilePath := filepath.Join(dir, "nonexistent-keyfile")

	wantPass := "stdin-passphrase"
	got, source, err := acquireWithIO(Options{KeyfilePath: keyfilePath}, bytes.NewBufferString(wantPass+"\n"), io.Discard, false)
	require.NoError(t, err)
	assert.Equal(t, wantPass, got)
	assert.Equal(t, SourceStdin, source)
}

func TestAcquire_PublicWrapperUsesStdinSeam(t *testing.T) {
	restorePassphraseStdioSeams(t)

	dir := t.TempDir()
	keyfilePath := filepath.Join(dir, "nonexistent-keyfile")

	passphraseStdin = bytes.NewBufferString("wrapped-stdin-passphrase\n")
	passphraseStderr = io.Discard
	passphraseIsTerminal = func() bool { return false }

	got, source, err := Acquire(Options{KeyfilePath: keyfilePath})
	require.NoError(t, err)
	assert.Equal(t, "wrapped-stdin-passphrase", got)
	assert.Equal(t, SourceStdin, source)
}

func TestAcquire_InteractivePromptUsesInjectedWriter(t *testing.T) {
	dir := t.TempDir()
	keyfilePath := filepath.Join(dir, "nonexistent-keyfile")
	var errBuf bytes.Buffer

	oldPrompt := passphrasePrompt
	passphrasePrompt = func(out io.Writer, promptText string) (string, error) {
		_, err := out.Write([]byte(promptText))
		return "interactive-passphrase", err
	}
	t.Cleanup(func() { passphrasePrompt = oldPrompt })

	got, source, err := acquireWithIO(Options{KeyfilePath: keyfilePath}, bytes.NewBuffer(nil), &errBuf, true)
	require.NoError(t, err)
	assert.Equal(t, "interactive-passphrase", got)
	assert.Equal(t, SourceInteractive, source)
	assert.Equal(t, "Enter passphrase: ", errBuf.String())
}

func TestAcquire_InteractiveCreationPromptsUseInjectedWriter(t *testing.T) {
	dir := t.TempDir()
	keyfilePath := filepath.Join(dir, "nonexistent-keyfile")
	var errBuf bytes.Buffer

	oldConfirm := passphraseConfirmPrompt
	passphraseConfirmPrompt = func(out io.Writer, promptText, confirmPromptText string) (string, error) {
		if _, err := out.Write([]byte(promptText)); err != nil {
			return "", err
		}
		if _, err := out.Write([]byte(confirmPromptText)); err != nil {
			return "", err
		}
		return "created-passphrase", nil
	}
	t.Cleanup(func() { passphraseConfirmPrompt = oldConfirm })

	got, source, err := acquireWithIO(Options{KeyfilePath: keyfilePath, AllowCreation: true}, bytes.NewBuffer(nil), &errBuf, true)
	require.NoError(t, err)
	assert.Equal(t, "created-passphrase", got)
	assert.Equal(t, SourceInteractive, source)
	assert.Equal(t, "Enter new passphrase: Confirm passphrase: ", errBuf.String())
}

func TestAcquire_InvalidKeyfilePermissions(t *testing.T) {
	// Keyfile exists but has wrong permissions — should fall through
	dir := t.TempDir()
	keyfilePath := filepath.Join(dir, "keyfile")

	require.NoError(t, os.WriteFile(keyfilePath, []byte("bad-perms\n"), 0644))

	wantPass := "fallback-stdin"
	got, source, err := acquireWithIO(Options{KeyfilePath: keyfilePath}, bytes.NewBufferString(wantPass+"\n"), io.Discard, false)
	require.NoError(t, err)
	assert.Equal(t, wantPass, got)
	assert.Equal(t, SourceStdin, source)
}

func TestAcquire_NoSourceAvailable(t *testing.T) {
	// No keyfile, stdin is a pipe but empty — should return error
	dir := t.TempDir()
	keyfilePath := filepath.Join(dir, "nonexistent-keyfile")

	_, _, err := acquireWithIO(Options{KeyfilePath: keyfilePath}, bytes.NewBuffer(nil), io.Discard, false)
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

type stubAcquireKeyringProvider struct {
	pass string
	err  error
}

func (s stubAcquireKeyringProvider) Get(service, key string) (string, error) { return s.pass, s.err }
func (s stubAcquireKeyringProvider) Set(service, key, value string) error    { return nil }
func (s stubAcquireKeyringProvider) Delete(service, key string) error        { return nil }

func TestAcquire_KeyringErrorWarnsAndFallsThrough(t *testing.T) {
	dir := t.TempDir()
	keyfilePath := filepath.Join(dir, "nonexistent-keyfile")
	var errBuf bytes.Buffer

	got, source, err := acquireWithIO(Options{
		KeyfilePath:     keyfilePath,
		KeyringProvider: stubAcquireKeyringProvider{err: errors.New("boom")},
	}, bytes.NewBufferString("stdin-passphrase\n"), &errBuf, false)
	require.NoError(t, err)
	assert.Equal(t, "stdin-passphrase", got)
	assert.Equal(t, SourceStdin, source)
	assert.Contains(t, errBuf.String(), "warning: keyring read failed: boom")
}

func TestAcquire_PublicWrapperUsesStderrSeam(t *testing.T) {
	restorePassphraseStdioSeams(t)

	dir := t.TempDir()
	keyfilePath := filepath.Join(dir, "keyfile")
	require.NoError(t, WriteKeyfile(keyfilePath, "keyfile-passphrase"))

	var errBuf bytes.Buffer
	passphraseStdin = bytes.NewBuffer(nil)
	passphraseStderr = &errBuf
	passphraseIsTerminal = func() bool { return false }

	got, source, err := Acquire(Options{
		KeyfilePath:     keyfilePath,
		KeyringProvider: stubAcquireKeyringProvider{err: errors.New("boom")},
	})
	require.NoError(t, err)
	assert.Equal(t, "keyfile-passphrase", got)
	assert.Equal(t, SourceKeyfile, source)
	assert.Equal(t, "warning: keyring read failed: boom\n", errBuf.String())
}

func restorePassphraseStdioSeams(t *testing.T) {
	t.Helper()

	oldStdin := passphraseStdin
	oldStderr := passphraseStderr
	oldIsTerminal := passphraseIsTerminal

	t.Cleanup(func() {
		passphraseStdin = oldStdin
		passphraseStderr = oldStderr
		passphraseIsTerminal = oldIsTerminal
	})
}
