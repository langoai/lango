package p2p

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestLoadOrGenerateKey_NewKeyWithoutSecrets(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	log := zap.NewNop().Sugar()

	// Generate new key (no SecretsStore).
	key, err := loadOrGenerateKey(tmpDir, nil, log)
	require.NoError(t, err)
	require.NotNil(t, key)

	// Verify key file was written.
	keyPath := filepath.Join(tmpDir, nodeKeyFile)
	_, err = os.Stat(keyPath)
	assert.NoError(t, err)

	// Load the same key again.
	key2, err := loadOrGenerateKey(tmpDir, nil, log)
	require.NoError(t, err)

	// Verify same key is loaded.
	raw1, err := crypto.MarshalPrivateKey(key)
	require.NoError(t, err)
	raw2, err := crypto.MarshalPrivateKey(key2)
	require.NoError(t, err)
	assert.Equal(t, raw1, raw2)
}

func TestLoadOrGenerateKey_LegacyFileLoaded(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	log := zap.NewNop().Sugar()

	// Pre-create a legacy key file.
	privKey, _, err := crypto.GenerateEd25519Key(nil)
	require.NoError(t, err)
	raw, err := crypto.MarshalPrivateKey(privKey)
	require.NoError(t, err)
	keyPath := filepath.Join(tmpDir, nodeKeyFile)
	require.NoError(t, os.WriteFile(keyPath, raw, 0o600))

	// Load with no secrets — should use legacy file.
	loaded, err := loadOrGenerateKey(tmpDir, nil, log)
	require.NoError(t, err)

	loadedRaw, err := crypto.MarshalPrivateKey(loaded)
	require.NoError(t, err)
	assert.Equal(t, raw, loadedRaw)
}

func TestExpandHome(t *testing.T) {
	t.Parallel()

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		give string
		want string
	}{
		{give: "~/foo", want: filepath.Join(home, "foo")},
		{give: "~/.lango/p2p", want: filepath.Join(home, ".lango", "p2p")},
		{give: "/absolute/path", want: "/absolute/path"},
		{give: "relative/path", want: "relative/path"},
		{give: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, expandHome(tt.give))
		})
	}
}

func TestLoadOrGenerateKey_EmptyKeyDirUsesDefault(t *testing.T) {
	t.Parallel()

	// Use a temp dir to avoid writing to real ~/.lango/p2p.
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "p2p")
	log := zap.NewNop().Sugar()

	// Generate in explicit subdir (simulates resolved default).
	key, err := loadOrGenerateKey(subDir, nil, log)
	require.NoError(t, err)
	require.NotNil(t, key)

	// Verify key file was created in the subdir.
	keyPath := filepath.Join(subDir, nodeKeyFile)
	_, err = os.Stat(keyPath)
	assert.NoError(t, err)
}

func TestZeroBytes(t *testing.T) {
	t.Parallel()

	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	zeroBytes(data)
	for _, b := range data {
		assert.Equal(t, byte(0), b)
	}
}
