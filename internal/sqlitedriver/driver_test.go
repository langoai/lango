package sqlitedriver

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadWriteDSNUsesSharedCache(t *testing.T) {
	t.Parallel()

	dsn := ReadWriteDSN("/tmp/lango-test.db")
	assert.Equal(t, "file:/tmp/lango-test.db?cache=shared", dsn)
}

func TestReadOnlyDSNUsesReadOnlyModeAndSharedCache(t *testing.T) {
	t.Parallel()

	dsn := ReadOnlyDSN("/tmp/lango-test.db")
	assert.Equal(t, "file:/tmp/lango-test.db?mode=ro&cache=shared", dsn)
}

func TestMemoryDSNDefaultsNameAndEnablesSharedMemoryDB(t *testing.T) {
	t.Parallel()

	dsn := MemoryDSN("")
	assert.Equal(t, "file:ent?mode=memory&cache=shared&_fk=1", dsn)
}

func TestCheckFileHeaderAllowsMissingAndEmptyFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.db")
	require.NoError(t, CheckFileHeader(missing))

	empty := filepath.Join(dir, "empty.db")
	require.NoError(t, osWriteFile(empty, nil))
	require.NoError(t, CheckFileHeader(empty))
}

func TestCheckFileHeaderRejectsLegacyOrUnreadableHeader(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "legacy.db")
	require.NoError(t, osWriteFile(path, bytes.Repeat([]byte{0x42}, 64)))

	err := CheckFileHeader(path)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrLegacyEncryptedOrUnreadableDB)
	assert.Contains(t, err.Error(), "downgrade/export required")
}

func TestCheckFileHeaderAcceptsRealSQLiteHeader(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "sqlite.db")
	payload := append([]byte("SQLite format 3\x00"), bytes.Repeat([]byte{0x00}, 64)...)
	require.NoError(t, osWriteFile(path, payload))
	require.NoError(t, CheckFileHeader(path))
}

func TestExpandPathLeavesNonHomePathUntouched(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "/tmp/lango.db", ExpandPath("/tmp/lango.db"))
}

func TestConfigureConnectionAllowsRealSQLiteHandle(t *testing.T) {
	t.Parallel()

	db, err := Open(filepath.Join(t.TempDir(), "cfg.db"), false)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	require.NoError(t, ConfigureConnection(db, false))

	var mode string
	require.NoError(t, db.QueryRow("PRAGMA journal_mode").Scan(&mode))
	assert.Equal(t, "wal", strings.ToLower(mode))
}

func osWriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o600)
}
