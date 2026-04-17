package dbopen

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/sqlitedriver"
)

const dataDirPerm = 0o700

// OpenManaged opens the application database in read-write mode and applies
// schema migration.
func OpenManaged(dbPath, encryptionKey string, rawKey bool, cipherPageSize int) (*ent.Client, *sql.DB, error) {
	dbPath = sqlitedriver.ExpandPath(dbPath)
	if err := sqlitedriver.CheckFileHeader(dbPath); err != nil {
		return nil, nil, err
	}

	// Ensure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(dbPath), dataDirPerm); err != nil {
		return nil, nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sqlitedriver.Open(dbPath, false)
	if err != nil {
		return nil, nil, fmt.Errorf("sql open: %w", err)
	}

	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)

	if err := sqlitedriver.ConfigureConnection(db, false); err != nil {
		db.Close()
		return nil, nil, err
	}

	if err := applyEncryptionPragmas(db, encryptionKey, rawKey, cipherPageSize, false); err != nil {
		db.Close()
		return nil, nil, err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))

	if err := client.Schema.Create(
		context.Background(),
		schema.WithForeignKeys(false),
	); err != nil {
		client.Close()
		return nil, nil, fmt.Errorf("schema migration: %w", err)
	}

	return client, db, nil
}

// OpenReadOnly opens the application database in read-only mode without
// invoking ent schema migration.
func OpenReadOnly(dbPath, encryptionKey string, rawKey bool, cipherPageSize int) (*ent.Client, *sql.DB, error) {
	dbPath = sqlitedriver.ExpandPath(dbPath)
	if err := sqlitedriver.CheckFileHeader(dbPath); err != nil {
		return nil, nil, err
	}

	if _, err := os.Stat(dbPath); err != nil {
		return nil, nil, fmt.Errorf("read-only db open: stat %q: %w", dbPath, err)
	}

	db, err := sqlitedriver.Open(dbPath, true)
	if err != nil {
		return nil, nil, fmt.Errorf("read-only sql open: %w", err)
	}

	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(2)

	if err := sqlitedriver.ConfigureConnection(db, true); err != nil {
		db.Close()
		return nil, nil, err
	}

	if err := applyEncryptionPragmas(db, encryptionKey, rawKey, cipherPageSize, true); err != nil {
		db.Close()
		return nil, nil, err
	}

	if _, err := db.Exec("PRAGMA schema_version"); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("read-only db verify: %w", err)
	}

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))
	return client, db, nil
}

func applyEncryptionPragmas(db *sql.DB, encryptionKey string, rawKey bool, cipherPageSize int, readonly bool) error {
	if encryptionKey == "" {
		return nil
	}
	if cipherPageSize <= 0 {
		cipherPageSize = 4096
	}
	var pragmaKey string
	if rawKey {
		pragmaKey = fmt.Sprintf(`PRAGMA key = "x'%s'"`, encryptionKey)
	} else {
		pragmaKey = fmt.Sprintf("PRAGMA key = '%s'", encryptionKey)
	}
	if _, err := db.Exec(pragmaKey); err != nil {
		if readonly {
			return fmt.Errorf("read-only PRAGMA key: %w", err)
		}
		return fmt.Errorf("set PRAGMA key: %w", err)
	}
	if _, err := db.Exec(fmt.Sprintf("PRAGMA cipher_page_size = %d", cipherPageSize)); err != nil {
		if readonly {
			return fmt.Errorf("read-only cipher_page_size: %w", err)
		}
		return fmt.Errorf("set cipher_page_size: %w", err)
	}
	return nil
}
