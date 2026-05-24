package dbmigrate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/sqlitedriver"
)

func createPlaintextDB(t *testing.T, dir string) string {
	t.Helper()
	dbPath := filepath.Join(dir, "test.db")
	db, err := sqlitedriver.Open(dbPath, false)
	require.NoError(t, err)
	require.NoError(t, sqlitedriver.ConfigureConnection(db, false))

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO test (name) VALUES ('hello'), ('world')")
	require.NoError(t, err)
	require.NoError(t, db.Close())
	return dbPath
}

func TestIsEncrypted_PlaintextDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := createPlaintextDB(t, dir)
	assert.False(t, IsEncrypted(dbPath))
}

func TestIsEncrypted_NonexistentFile(t *testing.T) {
	assert.False(t, IsEncrypted("/tmp/nonexistent_db_file_for_test.db"))
}

func TestIsEncrypted_SmallFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tiny.db")
	require.NoError(t, os.WriteFile(path, []byte("short"), 0600))
	assert.False(t, IsEncrypted(path))
}

func TestMigrateToEncrypted_EmptyPassphrase(t *testing.T) {
	dir := t.TempDir()
	dbPath := createPlaintextDB(t, dir)
	err := MigrateToEncrypted(dbPath, "", 4096)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "passphrase must not be empty")
}

func TestDecryptToPlaintext_EmptyPassphrase(t *testing.T) {
	dir := t.TempDir()
	dbPath := createPlaintextDB(t, dir)
	err := DecryptToPlaintext(dbPath, "", 4096)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "passphrase must not be empty")
}

func TestDecryptToPlaintext_NotEncrypted(t *testing.T) {
	dir := t.TempDir()
	dbPath := createPlaintextDB(t, dir)
	err := DecryptToPlaintext(dbPath, "test-pass", 4096)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSQLCipherUnsupported)
}

func TestIsSQLCipherAvailable(t *testing.T) {
	assert.False(t, IsSQLCipherAvailable())
}

func TestMigrateToEncrypted_Unsupported(t *testing.T) {
	dir := t.TempDir()
	dbPath := createPlaintextDB(t, dir)
	err := MigrateToEncrypted(dbPath, "test-passphrase", 4096)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSQLCipherUnsupported)
}

func TestDecryptToPlaintext_Unsupported(t *testing.T) {
	dir := t.TempDir()
	dbPath := createPlaintextDB(t, dir)
	err := DecryptToPlaintext(dbPath, "test-passphrase", 4096)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSQLCipherUnsupported)
}

func TestSecureDeleteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secret.txt")
	require.NoError(t, os.WriteFile(path, []byte("sensitive data here"), 0600))

	require.NoError(t, secureDeleteFile(path))

	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

func TestSecureDeleteFile_NonexistentFile(t *testing.T) {
	err := secureDeleteFile("/tmp/nonexistent_file_for_test_12345.txt")
	require.Error(t, err)
}
