package security

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internalsecurity "github.com/langoai/lango/internal/security"
)

func TestRecoveryRestoreCmd_NoBootstrapDependency(t *testing.T) {
	// Verify the restore command is constructed without bootLoader parameter.
	// newRecoveryRestoreCmd() takes no arguments — this is a compile-time guarantee.
	cmd := newRecoveryRestoreCmd()
	require.NotNil(t, cmd)
	assert.Equal(t, "restore", cmd.Use)
}

func TestRecoveryRestoreCmd_EnvelopeDirectLoad(t *testing.T) {
	// Create a temp dir with a valid envelope containing a mnemonic slot.
	tmpDir := t.TempDir()

	envelope := &internalsecurity.MasterKeyEnvelope{
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	// Add a passphrase slot so envelope is valid (needs at least one slot).
	mk, err := internalsecurity.GenerateMasterKey()
	require.NoError(t, err)
	defer internalsecurity.ZeroBytes(mk)

	err = envelope.AddSlot(internalsecurity.KEKSlotPassphrase, "primary", mk, "test-passphrase", internalsecurity.NewDefaultKDFParams())
	require.NoError(t, err)

	// Persist the envelope.
	data, err := json.Marshal(envelope)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "envelope.json"), data, 0600)
	require.NoError(t, err)

	// Verify LoadEnvelopeFile can load it directly (the path restore uses).
	loaded, err := internalsecurity.LoadEnvelopeFile(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, loaded, "envelope should be loaded directly without bootstrap")
	assert.Equal(t, 1, loaded.Version)
}

func TestRecoveryRestoreCmd_NoEnvelope(t *testing.T) {
	tmpDir := t.TempDir()

	// No envelope.json in the directory.
	loaded, err := internalsecurity.LoadEnvelopeFile(tmpDir)
	require.NoError(t, err)
	assert.Nil(t, loaded, "LoadEnvelopeFile should return nil for missing file")
}
