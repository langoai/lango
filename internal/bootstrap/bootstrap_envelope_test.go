package bootstrap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/security/passphrase"
)

// TestRun_FreshInstall_CreatesEnvelope verifies that a clean bootstrap on an
// empty LangoDir generates and persists a MasterKeyEnvelope file.
func TestRun_FreshInstall_CreatesEnvelope(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	keyfilePath := filepath.Join(dir, "keyfile")

	require.NoError(t, passphrase.WriteKeyfile(keyfilePath, "fresh-install-pass"))

	result, err := Run(Options{
		LangoDir:            dir,
		DBPath:              dbPath,
		KeyfilePath:         keyfilePath,
		KeepKeyfile:         true,
		SkipSecureDetection: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { result.DBClient.Close() })

	// Envelope file must exist.
	assert.True(t, security.HasEnvelopeFile(dir), "envelope file should be created on fresh install")

	env, err := security.LoadEnvelopeFile(dir)
	require.NoError(t, err)
	require.NotNil(t, env)
	assert.Equal(t, security.EnvelopeVersion, env.Version)
	assert.Equal(t, 1, env.SlotCount(), "fresh envelope should have exactly one passphrase slot")
	assert.True(t, env.HasSlotType(security.KEKSlotPassphrase))
	assert.False(t, env.PendingMigration)
	assert.False(t, env.PendingRekey)

	// Envelope file must be 0600.
	info, err := os.Stat(security.EnvelopeFilePath(dir))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

// TestRun_ReturningUser_UnwrapsEnvelope verifies that a second bootstrap on
// the same LangoDir successfully unwraps the MK from the existing envelope.
func TestRun_ReturningUser_UnwrapsEnvelope(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	keyfilePath := filepath.Join(dir, "keyfile")
	pass := "returning-user-pass"

	require.NoError(t, passphrase.WriteKeyfile(keyfilePath, pass))

	// First run creates the envelope.
	result1, err := Run(Options{
		LangoDir:            dir,
		DBPath:              dbPath,
		KeyfilePath:         keyfilePath,
		KeepKeyfile:         true,
		SkipSecureDetection: true,
	})
	require.NoError(t, err)
	result1.DBClient.Close()

	// Second run reopens using the same passphrase.
	result2, err := Run(Options{
		LangoDir:            dir,
		DBPath:              dbPath,
		KeyfilePath:         keyfilePath,
		KeepKeyfile:         true,
		SkipSecureDetection: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { result2.DBClient.Close() })
	require.NotNil(t, result2.Crypto)
}

// TestRun_WrongPassphrase_EnvelopeMode verifies that a wrong passphrase on an
// envelope-based installation is rejected with ErrUnwrapFailed.
func TestRun_WrongPassphrase_EnvelopeMode(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	keyfilePath := filepath.Join(dir, "keyfile")

	require.NoError(t, passphrase.WriteKeyfile(keyfilePath, "correct-passphrase"))
	result, err := Run(Options{
		LangoDir:            dir,
		DBPath:              dbPath,
		KeyfilePath:         keyfilePath,
		KeepKeyfile:         true,
		SkipSecureDetection: true,
	})
	require.NoError(t, err)
	result.DBClient.Close()

	// Overwrite keyfile with a wrong passphrase.
	require.NoError(t, passphrase.WriteKeyfile(keyfilePath, "wrong-passphrase"))
	_, err = Run(Options{
		LangoDir:            dir,
		DBPath:              dbPath,
		KeyfilePath:         keyfilePath,
		KeepKeyfile:         true,
		SkipSecureDetection: true,
	})
	assert.Error(t, err)
}
