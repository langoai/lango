package bootstrap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/security/passphrase"
)

func TestRun_ShredsKeyfileAfterCryptoInit(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	keyfilePath := filepath.Join(dir, "keyfile")
	pass := "test-passphrase-for-shred"

	require.NoError(t, passphrase.WriteKeyfile(keyfilePath, pass))

	result, err := Run(Options{
		DBPath:              dbPath,
		KeyfilePath:         keyfilePath,
		SkipSecureDetection: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		result.DBClient.Close()
	})

	_, statErr := os.Stat(keyfilePath)
	assert.True(t, os.IsNotExist(statErr), "keyfile should be deleted after bootstrap")
}

func TestRun_KeepsKeyfileWhenOptedOut(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	keyfilePath := filepath.Join(dir, "keyfile")
	pass := "test-passphrase-for-keep"

	require.NoError(t, passphrase.WriteKeyfile(keyfilePath, pass))

	result, err := Run(Options{
		DBPath:              dbPath,
		KeyfilePath:         keyfilePath,
		KeepKeyfile:         true,
		SkipSecureDetection: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		result.DBClient.Close()
	})

	_, statErr := os.Stat(keyfilePath)
	assert.NoError(t, statErr, "keyfile should still exist when KeepKeyfile is true")
}
