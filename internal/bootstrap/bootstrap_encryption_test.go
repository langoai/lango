package bootstrap

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsDBEncrypted_PlaintextDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "plain.db")

	db, err := sql.Open("sqlite3", "file:"+dbPath)
	require.NoError(t, err)
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY)")
	require.NoError(t, err)
	require.NoError(t, db.Close())

	assert.False(t, IsDBEncrypted(dbPath))
}

func TestIsDBEncrypted_NonexistentFile(t *testing.T) {
	assert.False(t, IsDBEncrypted("/tmp/nonexistent_db_test_bootstrap.db"))
}

func TestIsDBEncrypted_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.db")
	require.NoError(t, os.WriteFile(path, []byte{}, 0600))
	assert.False(t, IsDBEncrypted(path))
}

func TestIsDBEncrypted_RandomBytes(t *testing.T) {
	// A file with random bytes (simulating encrypted) should return true
	// since it won't have the SQLite magic header.
	dir := t.TempDir()
	path := filepath.Join(dir, "random.db")
	data := make([]byte, 4096)
	for i := range data {
		data[i] = byte(i % 256)
	}
	require.NoError(t, os.WriteFile(path, data, 0600))
	assert.True(t, IsDBEncrypted(path))
}

func TestOpenDatabase_Plaintext(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	client, rawDB, err := openDatabase(dbPath, "", false, 0)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotNil(t, rawDB)

	// Verify DB is usable.
	require.NoError(t, rawDB.Ping())

	client.Close()
}

func TestOpenDatabase_WithEncryptionKey_NoSQLCipher(t *testing.T) {
	// When SQLCipher is not available, PRAGMA key is accepted silently
	// by standard SQLite (it's a no-op). The DB opens but is NOT encrypted.
	// This is expected behavior -- actual encryption requires SQLCipher support.
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	client, rawDB, err := openDatabase(dbPath, "test-key", false, 4096)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotNil(t, rawDB)

	// Verify DB is usable (PRAGMA key is silently ignored without SQLCipher).
	require.NoError(t, rawDB.Ping())

	client.Close()

	// Without SQLCipher, the file should still be plaintext.
	assert.False(t, IsDBEncrypted(dbPath))
}
