// Package dbmigrate provides tools for converting between plaintext SQLite
// and SQLCipher-encrypted databases.
//
// SQLCipher support requires building with CGO and a SQLite library that includes
// SQLCipher (e.g., via system libsqlcipher). When built with the standard
// mattn/go-sqlite3 amalgamation, PRAGMA key is a no-op and encryption is unavailable.
package dbmigrate

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3" // SQLite driver (SQLCipher when linked with libsqlcipher)
)

// escapePassphrase escapes single quotes for SQLCipher PRAGMA values.
// SQLCipher PRAGMA key does not support parameterized queries, so
// single quotes must be doubled to prevent SQL injection.
func escapePassphrase(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// MigrateToEncrypted converts a plaintext SQLite DB to a SQLCipher-encrypted database.
// The original file is backed up and securely deleted after successful migration.
func MigrateToEncrypted(dbPath, passphrase string, cipherPageSize int) error {
	if passphrase == "" {
		return fmt.Errorf("passphrase must not be empty")
	}
	if cipherPageSize <= 0 {
		cipherPageSize = 4096
	}

	// Validate that the source DB is NOT already encrypted.
	if IsEncrypted(dbPath) {
		return fmt.Errorf("database is already encrypted")
	}

	tmpPath := dbPath + ".enc"
	defer os.Remove(tmpPath) // clean up temp file on error

	// Open the plaintext source database.
	srcDB, err := sql.Open("sqlite3", "file:"+dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return fmt.Errorf("open source db: %w", err)
	}
	defer srcDB.Close()

	// Verify SQLCipher is available by checking for the sqlcipher_export function.
	if err := verifySQLCipherAvailable(srcDB); err != nil {
		return err
	}

	// Verify we can actually read the source.
	if err := srcDB.Ping(); err != nil {
		return fmt.Errorf("ping source db: %w", err)
	}

	// Attach the encrypted target and export.
	attachSQL := fmt.Sprintf("ATTACH DATABASE '%s' AS target KEY '%s'", escapePassphrase(tmpPath), escapePassphrase(passphrase))
	if _, err := srcDB.Exec(attachSQL); err != nil {
		return fmt.Errorf("attach encrypted target: %w", err)
	}

	pragmaSQL := fmt.Sprintf("PRAGMA target.cipher_page_size = %d", cipherPageSize)
	if _, err := srcDB.Exec(pragmaSQL); err != nil {
		return fmt.Errorf("set cipher_page_size: %w", err)
	}

	if _, err := srcDB.Exec("SELECT sqlcipher_export('target')"); err != nil {
		return fmt.Errorf("sqlcipher_export: %w", err)
	}

	if _, err := srcDB.Exec("DETACH DATABASE target"); err != nil {
		return fmt.Errorf("detach target: %w", err)
	}

	srcDB.Close()

	// Verify the new encrypted DB can be opened.
	if err := verifyEncryptedDB(tmpPath, passphrase, cipherPageSize); err != nil {
		return fmt.Errorf("verify encrypted db: %w", err)
	}

	// Atomic swap: original -> .bak, encrypted -> original.
	bakPath := dbPath + ".bak"
	if err := os.Rename(dbPath, bakPath); err != nil {
		return fmt.Errorf("rename original to backup: %w", err)
	}
	if err := os.Rename(tmpPath, dbPath); err != nil {
		// Rollback: restore backup.
		_ = os.Rename(bakPath, dbPath)
		return fmt.Errorf("rename encrypted to original: %w", err)
	}

	// Secure-delete the backup.
	if err := secureDeleteFile(bakPath); err != nil {
		// Non-fatal: warn only.
		fmt.Fprintf(os.Stderr, "warning: secure delete of backup: %v\n", err)
	}

	return nil
}

// DecryptToPlaintext converts a SQLCipher-encrypted DB back to a plaintext SQLite database.
func DecryptToPlaintext(dbPath, passphrase string, cipherPageSize int) error {
	if passphrase == "" {
		return fmt.Errorf("passphrase must not be empty")
	}
	if cipherPageSize <= 0 {
		cipherPageSize = 4096
	}

	// Validate that the source DB IS encrypted.
	if !IsEncrypted(dbPath) {
		return fmt.Errorf("database is not encrypted")
	}

	tmpPath := dbPath + ".dec"
	defer os.Remove(tmpPath)

	// Open the encrypted source database with the key.
	srcDB, err := sql.Open("sqlite3", "file:"+dbPath+"?_busy_timeout=5000")
	if err != nil {
		return fmt.Errorf("open encrypted source: %w", err)
	}
	defer srcDB.Close()

	// Set the key PRAGMA to decrypt.
	if _, err := srcDB.Exec(fmt.Sprintf("PRAGMA key = '%s'", escapePassphrase(passphrase))); err != nil {
		return fmt.Errorf("set pragma key: %w", err)
	}
	if _, err := srcDB.Exec(fmt.Sprintf("PRAGMA cipher_page_size = %d", cipherPageSize)); err != nil {
		return fmt.Errorf("set cipher_page_size: %w", err)
	}

	if err := srcDB.Ping(); err != nil {
		return fmt.Errorf("open encrypted source (wrong passphrase?): %w", err)
	}

	// Attach plaintext target (empty key = no encryption).
	attachSQL := fmt.Sprintf("ATTACH DATABASE '%s' AS target KEY ''", escapePassphrase(tmpPath))
	if _, err := srcDB.Exec(attachSQL); err != nil {
		return fmt.Errorf("attach plaintext target: %w", err)
	}

	if _, err := srcDB.Exec("SELECT sqlcipher_export('target')"); err != nil {
		return fmt.Errorf("sqlcipher_export: %w", err)
	}

	if _, err := srcDB.Exec("DETACH DATABASE target"); err != nil {
		return fmt.Errorf("detach target: %w", err)
	}

	srcDB.Close()

	// Verify the new plaintext DB can be opened.
	if err := verifyPlaintextDB(tmpPath); err != nil {
		return fmt.Errorf("verify plaintext db: %w", err)
	}

	// Atomic swap.
	bakPath := dbPath + ".bak"
	if err := os.Rename(dbPath, bakPath); err != nil {
		return fmt.Errorf("rename original to backup: %w", err)
	}
	if err := os.Rename(tmpPath, dbPath); err != nil {
		_ = os.Rename(bakPath, dbPath)
		return fmt.Errorf("rename plaintext to original: %w", err)
	}

	if err := secureDeleteFile(bakPath); err != nil {
		fmt.Fprintf(os.Stderr, "warning: secure delete of backup: %v\n", err)
	}

	return nil
}

// IsEncrypted checks whether a database file is encrypted by inspecting the magic header.
// A standard SQLite file starts with "SQLite format 3\000"; an encrypted file does not.
func IsEncrypted(dbPath string) bool {
	f, err := os.Open(dbPath)
	if err != nil {
		return false
	}
	defer f.Close()
	header := make([]byte, 16)
	n, err := f.Read(header)
	if err != nil || n < 16 {
		return false
	}
	return string(header[:15]) != "SQLite format 3"
}

// IsSQLCipherAvailable checks whether the SQLite driver supports SQLCipher operations.
func IsSQLCipherAvailable() bool {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return false
	}
	defer db.Close()
	// sqlcipher_export is only available when SQLCipher is linked.
	var version string
	err = db.QueryRow("PRAGMA cipher_version").Scan(&version)
	return err == nil && version != ""
}

// verifySQLCipherAvailable returns an error if SQLCipher is not available.
func verifySQLCipherAvailable(db *sql.DB) error {
	var version string
	err := db.QueryRow("PRAGMA cipher_version").Scan(&version)
	if err != nil || version == "" {
		return fmt.Errorf("SQLCipher not available: binary must be built with SQLCipher support " +
			"(install libsqlcipher-dev and rebuild with CGO_LDFLAGS=-lsqlcipher)")
	}
	return nil
}

// verifyEncryptedDB opens the encrypted DB to verify it is readable.
func verifyEncryptedDB(path, passphrase string, cipherPageSize int) error {
	db, err := sql.Open("sqlite3", "file:"+path)
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := db.Exec(fmt.Sprintf("PRAGMA key = '%s'", escapePassphrase(passphrase))); err != nil {
		return err
	}
	if _, err := db.Exec(fmt.Sprintf("PRAGMA cipher_page_size = %d", cipherPageSize)); err != nil {
		return err
	}
	return db.Ping()
}

// verifyPlaintextDB opens the plaintext DB to verify it is readable.
func verifyPlaintextDB(path string) error {
	db, err := sql.Open("sqlite3", "file:"+path)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Ping()
}

// secureDeleteFile overwrites a file with zeros before removing it.
func secureDeleteFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}

	zeros := make([]byte, 4096)
	remaining := info.Size()
	for remaining > 0 {
		n := int64(len(zeros))
		if n > remaining {
			n = remaining
		}
		written, err := f.Write(zeros[:n])
		if err != nil {
			f.Close()
			os.Remove(path)
			return err
		}
		remaining -= int64(written)
	}

	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(path)
		return err
	}
	f.Close()

	return os.Remove(path)
}
