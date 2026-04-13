package security

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"golang.org/x/crypto/pbkdf2"

	"github.com/langoai/lango/internal/ent"
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
	if err := client.Schema.Create(context.Background(), schema.WithForeignKeys(false)); err != nil {
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
