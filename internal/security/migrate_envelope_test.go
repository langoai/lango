package security

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"golang.org/x/crypto/pbkdf2"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/testutil/schemautil"
	_ "github.com/mattn/go-sqlite3"
)

// setupLegacyDB builds an ent-backed SQLite DB with a legacy salt/checksum and
// a seeded secret + config profile encrypted with the passphrase-derived key.
// It mirrors what a pre-envelope install would look like on disk.
func setupLegacyDB(t *testing.T, dbPath, passphrase string) (*ent.Client, *sql.DB) {
	t.Helper()
	db, err := sql.Open("sqlite3", "file:"+dbPath+"?cache=shared&_journal_mode=WAL&_busy_timeout=5000&_fk=1")
	if err != nil {
		t.Fatalf("sql open: %v", err)
	}
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))
	if err := schemautil.CreateSchema(context.Background(), client); err != nil {
		t.Fatalf("schema create: %v", err)
	}

	// Legacy salt + checksum in security_config.
	store := NewSecurityConfigStore(db)
	if err := store.EnsureTable(); err != nil {
		t.Fatalf("ensure security_config: %v", err)
	}
	salt := []byte("legacy-salt16bts")
	if err := store.StoreSalt(salt); err != nil {
		t.Fatalf("store salt: %v", err)
	}
	mac := hmac.New(sha256.New, salt)
	mac.Write([]byte(passphrase))
	if err := store.StoreChecksum(mac.Sum(nil)); err != nil {
		t.Fatalf("store checksum: %v", err)
	}

	// Seed an encrypted secret that we can verify survives migration.
	legacyKey := pbkdf2.Key([]byte(passphrase), salt, Iterations, KeySize, sha256.New)
	ct, err := aesGCMEncrypt([]byte("hunter2"), legacyKey)
	if err != nil {
		t.Fatalf("encrypt legacy secret: %v", err)
	}

	// Register a key row so the Secret edge is valid (ent enforces the edge).
	keyEntity, err := client.Key.Create().
		SetName("default").
		SetRemoteKeyID("local").
		SetType("encryption").
		Save(context.Background())
	if err != nil {
		t.Fatalf("create key row: %v", err)
	}
	if _, err := client.Secret.Create().
		SetName("api-key").
		SetEncryptedValue(ct).
		SetKey(keyEntity).
		Save(context.Background()); err != nil {
		t.Fatalf("create secret row: %v", err)
	}

	// Seed a config profile row.
	profileCT, err := aesGCMEncrypt([]byte(`{"version":1}`), legacyKey)
	if err != nil {
		t.Fatalf("encrypt profile: %v", err)
	}
	if _, err := client.ConfigProfile.Create().
		SetName("default").
		SetEncryptedData(profileCT).
		SetActive(true).
		SetVersion(1).
		Save(context.Background()); err != nil {
		t.Fatalf("create profile row: %v", err)
	}

	return client, db
}

func TestMigrateToEnvelope_Plaintext_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	passphrase := "migration-test-pass"

	client, db := setupLegacyDB(t, dbPath, passphrase)

	// Load legacy state.
	store := NewSecurityConfigStore(db)
	salt, err := store.LoadSalt()
	if err != nil {
		t.Fatalf("load salt: %v", err)
	}
	checksum, err := store.LoadChecksum()
	if err != nil {
		t.Fatalf("load checksum: %v", err)
	}

	// Run migration (dbEncrypted=false → no PRAGMA rekey).
	env, mk, err := MigrateToEnvelope(
		context.Background(), db, client, dir,
		passphrase, salt, checksum, false,
	)
	if err != nil {
		t.Fatalf("MigrateToEnvelope: %v", err)
	}
	defer ZeroBytes(mk)

	if env == nil {
		t.Fatal("expected envelope, got nil")
	}
	if env.PendingMigration || env.PendingRekey {
		t.Fatalf("pending flags should be clear after successful migration: %+v", env)
	}
	if !HasEnvelopeFile(dir) {
		t.Fatal("envelope file should exist after migration")
	}

	// Verify seed data survived: read the secret and decrypt with MK.
	secretRow, err := client.Secret.Query().First(context.Background())
	if err != nil {
		t.Fatalf("re-read secret: %v", err)
	}
	plain, err := aesGCMDecrypt(secretRow.EncryptedValue, mk)
	if err != nil {
		t.Fatalf("decrypt with MK: %v", err)
	}
	if string(plain) != "hunter2" {
		t.Fatalf("secret plaintext mismatch: %q", plain)
	}

	// Verify envelope file contains the expected metadata.
	loaded, err := LoadEnvelopeFile(dir)
	if err != nil {
		t.Fatalf("LoadEnvelopeFile: %v", err)
	}
	if loaded.SlotCount() != 1 || !loaded.HasSlotType(KEKSlotPassphrase) {
		t.Fatalf("unexpected envelope slots: %+v", loaded.Slots)
	}
	unwrapped, _, err := loaded.UnwrapFromPassphrase(passphrase)
	if err != nil {
		t.Fatalf("unwrap from passphrase: %v", err)
	}
	defer ZeroBytes(unwrapped)

	client.Close()
	os.Remove(EnvelopeFilePath(dir))
}

func TestMigrateToEnvelope_AllowsMissingChecksum(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	passphrase := "migration-test-pass"

	client, db := setupLegacyDB(t, dbPath, passphrase)
	defer client.Close()

	store := NewSecurityConfigStore(db)
	salt, err := store.LoadSalt()
	if err != nil {
		t.Fatalf("load salt: %v", err)
	}

	env, mk, err := MigrateToEnvelope(
		context.Background(), db, client, dir,
		passphrase, salt, nil, false,
	)
	if err != nil {
		t.Fatalf("MigrateToEnvelope with missing checksum: %v", err)
	}
	defer ZeroBytes(mk)

	if env.PendingMigration || env.PendingRekey {
		t.Fatalf("pending flags should be clear after successful migration: %+v", env)
	}
	loaded, err := LoadEnvelopeFile(dir)
	if err != nil {
		t.Fatalf("LoadEnvelopeFile: %v", err)
	}
	if loaded.PendingMigration || loaded.PendingRekey {
		t.Fatalf("persisted envelope pending flags should be clear: %+v", loaded)
	}
}

func TestMigrateToEnvelope_WrongPassphrase(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	passphrase := "correct-pass-1234"

	client, db := setupLegacyDB(t, dbPath, passphrase)
	defer client.Close()

	store := NewSecurityConfigStore(db)
	salt, _ := store.LoadSalt()
	checksum, _ := store.LoadChecksum()

	_, _, err := MigrateToEnvelope(
		context.Background(), db, client, dir,
		"wrong-passphrase-999", salt, checksum, false,
	)
	if err == nil {
		t.Fatal("expected error on wrong passphrase")
	}

	// Envelope file should still exist (we wrote it before the TX) BUT
	// since the checksum verification happens BEFORE envelope write, no file
	// should exist. Confirm that.
	if HasEnvelopeFile(dir) {
		t.Fatal("envelope file should not exist when migration is rejected upfront")
	}
}

func TestMigrateToEnvelope_RejectsInvalidLegacySaltBeforeWritingEnvelope(t *testing.T) {
	dir := t.TempDir()

	env, mk, err := MigrateToEnvelope(
		context.Background(), nil, nil, dir,
		"passphrase", []byte("short"), nil, false,
	)
	if err == nil {
		t.Fatal("expected invalid legacy salt error")
	}
	if env != nil {
		t.Fatalf("expected nil envelope, got %+v", env)
	}
	if mk != nil {
		t.Fatalf("expected nil master key, got %d bytes", len(mk))
	}
	if !strings.Contains(err.Error(), "invalid legacy salt size") {
		t.Fatalf("expected invalid salt error, got %v", err)
	}
	if HasEnvelopeFile(dir) {
		t.Fatal("envelope file should not exist when salt validation fails")
	}
}

func TestRetryMigration_LegacyData(t *testing.T) {
	// RetryMigration is designed to be called by bootstrap when a previous
	// migration crashed between envelope creation and TX commit. In that
	// state, the envelope file says "PendingMigration=true" but the data is
	// still legacy-encrypted. RetryMigration re-runs the re-encryption step.
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	passphrase := "retry-legacy-pass"

	client, db := setupLegacyDB(t, dbPath, passphrase)
	defer client.Close()

	store := NewSecurityConfigStore(db)
	salt, _ := store.LoadSalt()

	// Simulate the crash: an envelope exists on disk with PendingMigration=true
	// but no data has been re-encrypted yet. Generate the MK that the envelope
	// would have used.
	mk, err := GenerateMasterKey()
	if err != nil {
		t.Fatal(err)
	}
	defer ZeroBytes(mk)

	// Call RetryMigration directly — this should succeed and migrate the legacy
	// secret to MK-based encryption.
	if err := RetryMigration(context.Background(), client, mk, passphrase, salt); err != nil {
		t.Fatalf("RetryMigration: %v", err)
	}

	// Verify: the secret is now decryptable with MK.
	secretRow, err := client.Secret.Query().First(context.Background())
	if err != nil {
		t.Fatalf("re-read secret: %v", err)
	}
	plain, err := aesGCMDecrypt(secretRow.EncryptedValue, mk)
	if err != nil {
		t.Fatalf("decrypt with MK after retry: %v", err)
	}
	if string(plain) != "hunter2" {
		t.Fatalf("unexpected plaintext after retry: %q", plain)
	}
}

func TestRetryMigration_RejectsInvalidLegacySaltBeforeOpeningTransaction(t *testing.T) {
	err := RetryMigration(context.Background(), nil, make([]byte, KeySize), "passphrase", []byte("short"))
	if err == nil {
		t.Fatal("expected invalid legacy salt error")
	}
	if !strings.Contains(err.Error(), "invalid legacy salt size") {
		t.Fatalf("expected invalid salt error, got %v", err)
	}
}

func TestRetryRekey_ReturnsErrorWhenDatabaseIsClosed(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql open: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}

	err = RetryRekey(db, make([]byte, KeySize))
	if err == nil {
		t.Fatal("expected rekey error for closed database")
	}
	if !strings.Contains(err.Error(), "pragma rekey") {
		t.Fatalf("expected pragma rekey error, got %v", err)
	}
}

func TestAESGCMHelpers_RoundTrip(t *testing.T) {
	key := make([]byte, KeySize)
	for i := range key {
		key[i] = byte(i)
	}
	plaintext := []byte("round trip secret")

	ct, err := aesGCMEncrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	pt, err := aesGCMDecrypt(ct, key)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(pt) != string(plaintext) {
		t.Fatalf("mismatch: %q vs %q", pt, plaintext)
	}
}

func TestAESGCMDecrypt_WrongKey(t *testing.T) {
	key1 := make([]byte, KeySize)
	key2 := make([]byte, KeySize)
	key2[0] = 1
	ct, err := aesGCMEncrypt([]byte("data"), key1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = aesGCMDecrypt(ct, key2)
	if err == nil {
		t.Fatal("expected error with wrong key")
	}
	if !errors.Is(err, ErrDecryptionFailed) {
		t.Fatalf("expected ErrDecryptionFailed, got %v", err)
	}
}

func TestAESGCMHelpers_RejectInvalidKeySizes(t *testing.T) {
	invalidKey := make([]byte, KeySize-1)

	if _, err := aesGCMEncrypt([]byte("data"), invalidKey); err == nil {
		t.Fatal("expected encrypt error for invalid key size")
	} else if !strings.Contains(err.Error(), "new cipher") {
		t.Fatalf("expected new cipher error, got %v", err)
	}

	if _, err := aesGCMDecrypt(make([]byte, NonceSize), invalidKey); err == nil {
		t.Fatal("expected decrypt error for invalid key size")
	} else if !strings.Contains(err.Error(), "new cipher") {
		t.Fatalf("expected new cipher error, got %v", err)
	}
}

func TestAESGCMDecrypt_RejectsCiphertextShorterThanNonce(t *testing.T) {
	_, err := aesGCMDecrypt(make([]byte, NonceSize-1), make([]byte, KeySize))
	if err == nil {
		t.Fatal("expected ciphertext too short error")
	}
	if errors.Is(err, ErrDecryptionFailed) {
		t.Fatalf("ciphertext length validation should not wrap ErrDecryptionFailed: %v", err)
	}
	if !strings.Contains(err.Error(), "ciphertext too short") {
		t.Fatalf("expected ciphertext too short error, got %v", err)
	}
}

func TestAESGCMDecrypt_PreservesDecryptionFailureForTamperedCiphertext(t *testing.T) {
	key := make([]byte, KeySize)
	ct, err := aesGCMEncrypt([]byte("data"), key)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	ct[len(ct)-1] ^= 0xff

	_, err = aesGCMDecrypt(ct, key)
	if err == nil {
		t.Fatal("expected error for tampered ciphertext")
	}
	if !errors.Is(err, ErrDecryptionFailed) {
		t.Fatalf("expected ErrDecryptionFailed, got %v", err)
	}
}

func TestBackupDatabase_ReturnsErrorWhenDatabaseIsClosed(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql open: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}

	err = backupDatabase(context.Background(), db)
	if err == nil {
		t.Fatal("expected backup error for closed database")
	}
	if !strings.Contains(err.Error(), "wal_checkpoint") {
		t.Fatalf("expected wal_checkpoint error, got %v", err)
	}
}

func TestBackupDatabase_ReturnsErrorWhenMainDatabaseHasNoPath(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql open: %v", err)
	}
	defer db.Close()

	err = backupDatabase(context.Background(), db)
	if err == nil {
		t.Fatal("expected backup error for in-memory database")
	}
	if !strings.Contains(err.Error(), "main db path not found") {
		t.Fatalf("expected missing main db path error, got %v", err)
	}
}

func TestBackupDatabase_CreatesBackupForQuotedPath(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "quoted'name.db")
	db, err := sql.Open("sqlite3", "file:"+dbPath+"?_journal_mode=WAL")
	if err != nil {
		t.Fatalf("sql open: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE t (id INTEGER PRIMARY KEY, value TEXT)`); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO t (value) VALUES ('ok')`); err != nil {
		t.Fatalf("insert row: %v", err)
	}

	if err := backupDatabase(context.Background(), db); err != nil {
		t.Fatalf("backupDatabase: %v", err)
	}
	backupPath := dbPath + migrationBackupSuffix
	info, err := os.Stat(backupPath)
	if err != nil {
		t.Fatalf("stat backup: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected backup mode 0600, got %v", info.Mode().Perm())
	}
}

func TestRekeyDatabase_ReturnsErrorWhenDatabaseIsClosed(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql open: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}

	err = rekeyDatabase(db, make([]byte, KeySize))
	if err == nil {
		t.Fatal("expected rekey error for closed database")
	}
	if !strings.Contains(err.Error(), "pragma rekey") {
		t.Fatalf("expected pragma rekey error, got %v", err)
	}
}
