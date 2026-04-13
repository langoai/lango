package security

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/security"
)

// TestReadEnvelopeStatus_Present verifies the passphrase-free envelope reader
// extracts the expected fields from an on-disk envelope file.
func TestReadEnvelopeStatus_Present(t *testing.T) {
	dir := t.TempDir()
	env, mk, err := security.NewEnvelope("smoke-test-pass")
	require.NoError(t, err)
	defer security.ZeroBytes(mk)

	// Add a mnemonic slot to exercise the RecoverySetup branch.
	mn := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	require.NoError(t, env.AddSlot(security.KEKSlotMnemonic, "recovery", mk, mn, security.NewDefaultKDFParams()))
	require.NoError(t, security.StoreEnvelopeFile(dir, env))

	got := readEnvelopeStatus(dir)
	assert.True(t, got.Present)
	assert.Equal(t, security.EnvelopeVersion, got.Version)
	assert.Equal(t, 2, got.SlotCount)
	assert.True(t, got.RecoverySetup, "mnemonic slot should flip RecoverySetup to true")
	assert.Contains(t, got.SlotTypes, string(security.KEKSlotPassphrase))
	assert.Contains(t, got.SlotTypes, string(security.KEKSlotMnemonic))
	assert.False(t, got.PendingMigration)
	assert.False(t, got.PendingRekey)
}

// TestReadEnvelopeStatus_Missing verifies the reader degrades to an empty
// struct when no envelope file exists.
func TestReadEnvelopeStatus_Missing(t *testing.T) {
	dir := t.TempDir()
	got := readEnvelopeStatus(dir)
	assert.False(t, got.Present)
	assert.Equal(t, 0, got.SlotCount)
	assert.False(t, got.RecoverySetup)
}

// TestReadEnvelopeStatus_CorruptReturnsEmpty verifies the reader never panics
// or propagates errors when the envelope file is malformed.
func TestReadEnvelopeStatus_CorruptReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	path := security.EnvelopeFilePath(dir)
	require.NoError(t, os.WriteFile(path, []byte("{not-json"), 0o600))

	got := readEnvelopeStatus(dir)
	assert.False(t, got.Present, "corrupt envelope must degrade to absent")
}

// TestReadDBStatusNonInteractive_NoKeyfileNoKeyring verifies the non-interactive
// mini-bootstrap degrades gracefully when no credential is available.
func TestReadDBStatusNonInteractive_NoKeyfileNoKeyring(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "nonexistent.db")

	got := readDBStatusNonInteractive(dir, dbPath, nil, false)
	assert.False(t, got.available)
	assert.Equal(t, 0, got.encryptionKeys)
	assert.Equal(t, 0, got.storedSecrets)
}
