package security

import (
	"bytes"
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

func TestConfirmWord_RejectsMismatchedWord(t *testing.T) {
	var out bytes.Buffer

	err := confirmWord(bytes.NewBufferString("wrong\n"), &out, []string{"alpha", "beta"}, 2)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "confirmation word 2 did not match")
	assert.Contains(t, out.String(), "Enter word 2 to confirm: ")
}

func TestConfirmWord_RejectsEmptyEOF(t *testing.T) {
	var out bytes.Buffer

	err := confirmWord(bytes.NewBufferString(""), &out, []string{"alpha", "beta"}, 1)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "read confirmation word")
	assert.Contains(t, err.Error(), "EOF")
	assert.Contains(t, out.String(), "Enter word 1 to confirm: ")
}

func TestPickConfirmationIndexes_BoundaryAndDistinctness(t *testing.T) {
	first, second := pickConfirmationIndexes(1)
	assert.Equal(t, 1, first)
	assert.Equal(t, 1, second)

	for range 100 {
		first, second = pickConfirmationIndexes(2)
		assert.GreaterOrEqual(t, first, 1)
		assert.LessOrEqual(t, first, 2)
		assert.GreaterOrEqual(t, second, 1)
		assert.LessOrEqual(t, second, 2)
		assert.NotEqual(t, first, second)
	}
}
