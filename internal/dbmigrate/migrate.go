package dbmigrate

import (
	"errors"
	"fmt"
	"os"

	"github.com/langoai/lango/internal/sqlitedriver"
)

// ErrSQLCipherUnsupported indicates that legacy SQLCipher workflows are no
// longer supported by the current runtime.
var ErrSQLCipherUnsupported = errors.New("SQLCipher database workflows are no longer supported")

// MigrateToEncrypted is retained as a tombstone entrypoint for older CLI
// surfaces. New runtimes do not support SQLCipher migration.
func MigrateToEncrypted(dbPath, passphrase string, cipherPageSize int) error {
	_ = dbPath
	_ = cipherPageSize
	if passphrase == "" {
		return fmt.Errorf("passphrase must not be empty")
	}
	return fmt.Errorf("%w: use an older build to export legacy databases before upgrading", ErrSQLCipherUnsupported)
}

// DecryptToPlaintext is retained as a tombstone entrypoint for older CLI
// surfaces. New runtimes do not support SQLCipher decryption workflows.
func DecryptToPlaintext(dbPath, passphrase string, cipherPageSize int) error {
	_ = dbPath
	_ = cipherPageSize
	if passphrase == "" {
		return fmt.Errorf("passphrase must not be empty")
	}
	return fmt.Errorf("%w: use an older build to export legacy databases before upgrading", ErrSQLCipherUnsupported)
}

// IsEncrypted checks for a non-SQLite file header, which now means "legacy
// encrypted or unreadable" rather than a supported encrypted runtime format.
func IsEncrypted(dbPath string) bool {
	return errors.Is(sqlitedriver.CheckFileHeader(dbPath), sqlitedriver.ErrLegacyEncryptedOrUnreadableDB)
}

// IsSQLCipherAvailable always reports false in the brokered payload-protection runtime.
func IsSQLCipherAvailable() bool {
	return false
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
			_ = os.Remove(path)
			return err
		}
		remaining -= int64(written)
	}

	if err := f.Sync(); err != nil {
		f.Close()
		_ = os.Remove(path)
		return err
	}
	_ = f.Close()

	return os.Remove(path)
}
