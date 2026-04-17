package sqlitedriver

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

const (
	driverName         = "sqlite"
	defaultBusyTimeout = 5000
)

var ErrLegacyEncryptedOrUnreadableDB = errors.New("legacy encrypted or unreadable DB")

func DriverName() string {
	return driverName
}

func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func ReadWriteDSN(path string) string {
	return buildDSN(path, false)
}

func ReadOnlyDSN(path string) string {
	return buildDSN(path, true)
}

func Open(path string, readonly bool) (*sql.DB, error) {
	dsn := ReadWriteDSN(path)
	if readonly {
		dsn = ReadOnlyDSN(path)
	}
	db, err := sql.Open(DriverName(), dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func MemoryDSN(name string) string {
	if strings.TrimSpace(name) == "" {
		name = "ent"
	}
	return fmt.Sprintf("file:%s?mode=memory&cache=shared&_fk=1", name)
}

func CheckFileHeader(path string) error {
	path = ExpandPath(path)
	if strings.HasPrefix(path, "file:") {
		return nil
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat sqlite db %q: %w", path, err)
	}
	if info.Size() < 16 {
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open sqlite db header %q: %w", path, err)
	}
	defer f.Close()

	header := make([]byte, 16)
	n, err := f.Read(header)
	if err != nil || n < 16 {
		return nil
	}
	if string(header[:15]) == "SQLite format 3" {
		return nil
	}
	return fmt.Errorf("%w: downgrade/export required", ErrLegacyEncryptedOrUnreadableDB)
}

func buildDSN(path string, readonly bool) string {
	path = ExpandPath(path)
	base := path
	if !strings.HasPrefix(base, "file:") {
		base = "file:" + base
	}

	params := []string{}
	if readonly {
		params = append(params, "mode=ro")
	}
	params = append(params, "cache=shared")

	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	return base + sep + strings.Join(params, "&")
}

func ConfigureConnection(db *sql.DB, readonly bool) error {
	if db == nil {
		return nil
	}
	if _, err := db.Exec(fmt.Sprintf("PRAGMA busy_timeout = %d", defaultBusyTimeout)); err != nil {
		return fmt.Errorf("set busy_timeout: %w", err)
	}
	if !readonly {
		if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
			return fmt.Errorf("set journal_mode: %w", err)
		}
	}
	return nil
}
