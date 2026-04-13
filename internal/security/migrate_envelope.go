package security

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/pbkdf2"

	"github.com/langoai/lango/internal/ent"
)

// migrationBackupSuffix is the filename suffix for the WAL-safe pre-migration DB copy.
const migrationBackupSuffix = ".pre-migration"

// MigrateToEnvelope performs a one-time legacy → envelope migration.
//
// Flow (SQLCipher):
//  1. Derive legacy key via PBKDF2(passphrase, oldSalt, …) and verify checksum.
//  2. Generate a fresh MK and build an envelope with PendingMigration=true
//     (and PendingRekey=true if the DB is SQLCipher-encrypted). Persist the
//     envelope file BEFORE touching any data so a crash leaves a discoverable
//     marker.
//  3. PRAGMA wal_checkpoint(TRUNCATE) + VACUUM INTO <dbpath>.pre-migration for
//     a WAL-safe backup.
//  4. Re-encrypt every secret and config profile row inside a single ent
//     transaction. Count-verify before/after.
//  5. Clear PendingMigration and persist the envelope.
//  6. PRAGMA rekey to MK-derived raw key (SQLCipher only).
//  7. Clear PendingRekey and persist the envelope.
//
// Returns (envelope, mk, error). The caller takes ownership of the raw MK and
// must ZeroBytes it when done.
func MigrateToEnvelope(
	ctx context.Context,
	db *sql.DB,
	client *ent.Client,
	langoDir string,
	passphrase string,
	oldSalt, oldChecksum []byte,
	dbEncrypted bool,
) (*MasterKeyEnvelope, []byte, error) {
	// 1. Verify legacy key material.
	if len(oldSalt) != SaltSize {
		return nil, nil, fmt.Errorf("migrate: invalid legacy salt size %d", len(oldSalt))
	}
	oldKey := pbkdf2.Key([]byte(passphrase), oldSalt, Iterations, KeySize, sha256.New)
	defer ZeroBytes(oldKey)

	if oldChecksum != nil {
		mac := hmac.New(sha256.New, oldSalt)
		mac.Write([]byte(passphrase))
		computed := mac.Sum(nil)
		if !hmac.Equal(oldChecksum, computed) {
			return nil, nil, fmt.Errorf("migrate: legacy passphrase checksum mismatch")
		}
	}

	// 2. Generate MK and build envelope. Persist BEFORE data changes.
	mk, err := GenerateMasterKey()
	if err != nil {
		return nil, nil, fmt.Errorf("migrate: generate mk: %w", err)
	}
	now := time.Now().UTC()
	envelope := &MasterKeyEnvelope{
		Version:          EnvelopeVersion,
		PendingMigration: true,
		PendingRekey:     dbEncrypted,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := envelope.AddSlot(KEKSlotPassphrase, "", mk, passphrase, NewDefaultKDFParams()); err != nil {
		ZeroBytes(mk)
		return nil, nil, fmt.Errorf("migrate: build envelope: %w", err)
	}
	if err := StoreEnvelopeFile(langoDir, envelope); err != nil {
		ZeroBytes(mk)
		return nil, nil, fmt.Errorf("migrate: persist envelope: %w", err)
	}

	// 3. WAL-safe backup. Failure here is fatal — we need the backup before
	//    we touch data.
	if err := backupDatabase(ctx, db); err != nil {
		ZeroBytes(mk)
		return nil, nil, fmt.Errorf("migrate: backup db: %w", err)
	}

	// 4. Re-encrypt data in a single transaction with count verification.
	if err := reencryptAll(ctx, client, oldKey, mk); err != nil {
		ZeroBytes(mk)
		return nil, nil, fmt.Errorf("migrate: re-encrypt: %w", err)
	}

	// 5. Clear PendingMigration and persist.
	envelope.PendingMigration = false
	envelope.UpdatedAt = time.Now().UTC()
	if err := StoreEnvelopeFile(langoDir, envelope); err != nil {
		ZeroBytes(mk)
		return nil, nil, fmt.Errorf("migrate: persist envelope after re-encrypt: %w", err)
	}

	// 6. PRAGMA rekey (SQLCipher only).
	if dbEncrypted {
		if err := rekeyDatabase(db, mk); err != nil {
			ZeroBytes(mk)
			return nil, nil, fmt.Errorf("migrate: rekey: %w", err)
		}
		envelope.PendingRekey = false
		envelope.UpdatedAt = time.Now().UTC()
		if err := StoreEnvelopeFile(langoDir, envelope); err != nil {
			ZeroBytes(mk)
			return nil, nil, fmt.Errorf("migrate: persist envelope after rekey: %w", err)
		}
	}

	return envelope, mk, nil
}

// RetryMigration re-runs the data re-encryption phase using an already-unwrapped MK.
// Used by bootstrap when a previous migration crashed after envelope write but
// before data re-encryption completed. The caller is responsible for clearing
// PendingMigration and persisting the envelope afterwards.
func RetryMigration(ctx context.Context, client *ent.Client, mk []byte, passphrase string, oldSalt []byte) error {
	if len(oldSalt) != SaltSize {
		return fmt.Errorf("retry migration: invalid legacy salt size")
	}
	oldKey := pbkdf2.Key([]byte(passphrase), oldSalt, Iterations, KeySize, sha256.New)
	defer ZeroBytes(oldKey)
	return reencryptAll(ctx, client, oldKey, mk)
}

// RetryRekey retries `PRAGMA rekey` using the MK-derived raw key.
// Used when the previous migration crashed between data re-encryption and rekey.
func RetryRekey(db *sql.DB, mk []byte) error {
	return rekeyDatabase(db, mk)
}

// reencryptAll decrypts every secret + config_profile row with oldKey and
// re-encrypts with the provided Master Key. Count-verified before/after.
func reencryptAll(ctx context.Context, client *ent.Client, oldKey, mk []byte) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	// --- secrets ---
	beforeSecrets, err := tx.Secret.Query().Count(ctx)
	if err != nil {
		return fmt.Errorf("count secrets: %w", err)
	}
	secrets, err := tx.Secret.Query().All(ctx)
	if err != nil {
		return fmt.Errorf("list secrets: %w", err)
	}
	for _, row := range secrets {
		plain, derr := aesGCMDecrypt(row.EncryptedValue, oldKey)
		if derr != nil {
			return fmt.Errorf("decrypt secret %q: %w", row.Name, derr)
		}
		newCT, eerr := aesGCMEncrypt(plain, mk)
		ZeroBytes(plain)
		if eerr != nil {
			return fmt.Errorf("re-encrypt secret %q: %w", row.Name, eerr)
		}
		if _, uerr := tx.Secret.UpdateOneID(row.ID).SetEncryptedValue(newCT).Save(ctx); uerr != nil {
			return fmt.Errorf("update secret %q: %w", row.Name, uerr)
		}
	}
	afterSecrets, err := tx.Secret.Query().Count(ctx)
	if err != nil {
		return fmt.Errorf("count secrets after: %w", err)
	}
	if beforeSecrets != afterSecrets {
		return fmt.Errorf("secret count changed during migration: %d → %d", beforeSecrets, afterSecrets)
	}

	// --- config profiles ---
	beforeProfiles, err := tx.ConfigProfile.Query().Count(ctx)
	if err != nil {
		return fmt.Errorf("count profiles: %w", err)
	}
	profiles, err := tx.ConfigProfile.Query().All(ctx)
	if err != nil {
		return fmt.Errorf("list profiles: %w", err)
	}
	for _, row := range profiles {
		plain, derr := aesGCMDecrypt(row.EncryptedData, oldKey)
		if derr != nil {
			return fmt.Errorf("decrypt profile %q: %w", row.Name, derr)
		}
		newCT, eerr := aesGCMEncrypt(plain, mk)
		ZeroBytes(plain)
		if eerr != nil {
			return fmt.Errorf("re-encrypt profile %q: %w", row.Name, eerr)
		}
		if _, uerr := tx.ConfigProfile.UpdateOneID(row.ID).SetEncryptedData(newCT).Save(ctx); uerr != nil {
			return fmt.Errorf("update profile %q: %w", row.Name, uerr)
		}
	}
	afterProfiles, err := tx.ConfigProfile.Query().Count(ctx)
	if err != nil {
		return fmt.Errorf("count profiles after: %w", err)
	}
	if beforeProfiles != afterProfiles {
		return fmt.Errorf("config profile count changed during migration: %d → %d", beforeProfiles, afterProfiles)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	committed = true
	return nil
}

// backupDatabase creates a WAL-safe backup via PRAGMA wal_checkpoint(TRUNCATE)
// followed by VACUUM INTO. Requires the DB to be open with write access.
func backupDatabase(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `PRAGMA wal_checkpoint(TRUNCATE)`); err != nil {
		return fmt.Errorf("wal_checkpoint: %w", err)
	}
	// Determine the DB path from the connection. SQLite exposes it via
	// `PRAGMA database_list`, and we copy the "main" entry.
	rows, err := db.QueryContext(ctx, `PRAGMA database_list`)
	if err != nil {
		return fmt.Errorf("database_list: %w", err)
	}
	var dbPath string
	for rows.Next() {
		var seq int
		var name, file string
		if err := rows.Scan(&seq, &name, &file); err != nil {
			rows.Close()
			return fmt.Errorf("scan database_list: %w", err)
		}
		if name == "main" {
			dbPath = file
		}
	}
	rows.Close()
	if dbPath == "" {
		return fmt.Errorf("backup: main db path not found")
	}
	backupPath := dbPath + migrationBackupSuffix
	// Remove stale backup from a prior crashed migration.
	_ = os.Remove(backupPath)
	// VACUUM INTO requires a literal quoted path — escape single quotes.
	escaped := backupPath
	// Quote-escape by doubling single quotes (standard SQL).
	var sb []byte
	for _, c := range []byte(escaped) {
		if c == '\'' {
			sb = append(sb, '\'', '\'')
		} else {
			sb = append(sb, c)
		}
	}
	if _, err := db.ExecContext(ctx, fmt.Sprintf(`VACUUM INTO '%s'`, sb)); err != nil {
		return fmt.Errorf("vacuum into: %w", err)
	}
	// Best-effort: enforce 0600 permissions on the backup.
	_ = os.Chmod(backupPath, 0o600)
	// Ensure the parent dir at least exists (it always should).
	_ = os.MkdirAll(filepath.Dir(backupPath), 0o700)
	return nil
}

// rekeyDatabase issues `PRAGMA rekey` with a raw MK-derived key.
// The caller is responsible for reopening the connection with the new key
// before issuing further writes.
func rekeyDatabase(db *sql.DB, mk []byte) error {
	hexKey := DeriveDBKeyHex(mk)
	if _, err := db.Exec(fmt.Sprintf(`PRAGMA rekey = "x'%s'"`, hexKey)); err != nil {
		return fmt.Errorf("pragma rekey: %w", err)
	}
	return nil
}

// aesGCMEncrypt encrypts plaintext with AES-256-GCM. The returned bytes are
// `nonce || ciphertext` to match the format used by LocalCryptoProvider.Encrypt.
func aesGCMEncrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// aesGCMDecrypt is the inverse of aesGCMEncrypt — it reads the 12-byte nonce
// prefix and verifies the GCM authentication tag.
func aesGCMDecrypt(ciphertext, key []byte) ([]byte, error) {
	if len(ciphertext) < NonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}
	nonce, ct := ciphertext[:NonceSize], ciphertext[NonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}
	return plaintext, nil
}
