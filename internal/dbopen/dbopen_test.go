package dbopen

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/sqlitedriver"
)

func TestOpenManagedCreatesUsableDatabaseAndParentDir(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "nested", "managed.db")
	client, rawDB, err := OpenManaged(dbPath, "", false, 0)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotNil(t, rawDB)
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	require.FileExists(t, dbPath)
	require.NoError(t, rawDB.Ping())
}

func TestOpenReadOnlyOpensExistingManagedDatabase(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "readonly.db")
	client, rawDB, err := OpenManaged(dbPath, "", false, 0)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotNil(t, rawDB)
	require.NoError(t, client.Close())
	require.NoError(t, rawDB.Close())

	roClient, roDB, err := OpenReadOnly(dbPath, "", false, 0)
	require.NoError(t, err)
	require.NotNil(t, roClient)
	require.NotNil(t, roDB)
	t.Cleanup(func() { require.NoError(t, roClient.Close()) })

	require.NoError(t, roDB.Ping())
}

func TestOpenReadOnlyRejectsMissingDatabase(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "missing.db")
	client, rawDB, err := OpenReadOnly(dbPath, "", false, 0)
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Nil(t, rawDB)
	assert.Contains(t, err.Error(), "read-only db open: stat")
}

func TestOpenManagedIgnoresDeprecatedEncryptionArgsForPlaintextDB(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "deprecated-args.db")
	client, rawDB, err := OpenManaged(dbPath, "legacy-passphrase", false, 4096)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotNil(t, rawDB)
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	require.NoError(t, rawDB.Ping())
}

func TestOpenReadOnlyIgnoresDeprecatedEncryptionArgsForPlaintextDB(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "deprecated-readonly.db")
	client, rawDB, err := OpenManaged(dbPath, "", false, 0)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotNil(t, rawDB)
	require.NoError(t, client.Close())
	require.NoError(t, rawDB.Close())

	roClient, roDB, err := OpenReadOnly(dbPath, "legacy-passphrase", false, 4096)
	require.NoError(t, err)
	require.NotNil(t, roClient)
	require.NotNil(t, roDB)
	t.Cleanup(func() { require.NoError(t, roClient.Close()) })

	require.NoError(t, roDB.Ping())
}

func TestOpenManagedRejectsLegacyHeaderBeforeCreatingClient(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "legacy.db")
	require.NoError(t, osWriteFile(dbPath, bytes.Repeat([]byte{0x99}, 64)))

	client, rawDB, err := OpenManaged(dbPath, "", false, 0)
	require.Error(t, err)
	assert.ErrorIs(t, err, sqlitedriver.ErrLegacyEncryptedOrUnreadableDB)
	assert.Nil(t, client)
	assert.Nil(t, rawDB)
}

func osWriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o600)
}

func TestOpenManagedIsSafeForConcurrentInvocations(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	paths := []string{
		filepath.Join(dir, "one.db"),
		filepath.Join(dir, "two.db"),
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(paths))
	panicCh := make(chan interface{}, len(paths))

	for _, dbPath := range paths {
		dbPath := dbPath
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panicCh <- r
				}
			}()

			client, rawDB, err := OpenManaged(dbPath, "", false, 0)
			if err != nil {
				errCh <- err
				return
			}
			defer client.Close()
			defer rawDB.Close()

			if err := rawDB.Ping(); err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)
	close(panicCh)

	for p := range panicCh {
		t.Fatalf("concurrent OpenManaged panicked: %v", p)
	}
	for err := range errCh {
		require.NoError(t, err)
	}
}
