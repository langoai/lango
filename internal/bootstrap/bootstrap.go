package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/mattn/go-sqlite3"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/security"
)

// Result holds everything produced by the bootstrap process.
type Result struct {
	// Config is the decrypted, active configuration.
	Config *config.Config
	// DBClient is the shared ent.Client for the application database.
	DBClient *ent.Client
	// RawDB is the underlying *sql.DB for direct SQL operations (e.g., sqlite-vec).
	RawDB *sql.DB
	// Crypto is the initialized CryptoProvider for the session.
	Crypto security.CryptoProvider
	// ConfigStore provides encrypted profile CRUD operations.
	ConfigStore *configstore.Store
	// ProfileName is the name of the loaded profile.
	ProfileName string
}

// Options configures the bootstrap process.
type Options struct {
	// DBPath is the SQLite database path (default: ~/.lango/lango.db).
	DBPath string
	// KeyfilePath is the path to the passphrase keyfile (default: ~/.lango/keyfile).
	KeyfilePath string
	// ForceProfile overrides the active profile selection.
	ForceProfile string
	// KeepKeyfile prevents the keyfile from being shredded after crypto initialization.
	// Default (false) shreds the keyfile for security.
	KeepKeyfile bool
	// DBEncryption configures SQLCipher transparent database encryption.
	DBEncryption config.DBEncryptionConfig
	// SkipSecureDetection disables secure hardware provider detection (biometric/TPM).
	// When true, the bootstrap falls back to keyfile or interactive prompt only.
	// Useful for testing and headless environments.
	SkipSecureDetection bool
}

// Run executes the full bootstrap sequence using the phase pipeline:
//  1. Ensure ~/.lango/ directory
//  2. Detect DB encryption status
//  3. Acquire passphrase
//  4. Open SQLite/SQLCipher DB + ent schema migration
//  5. Load security state (salt/checksum)
//  6. Initialize crypto provider
//  7. Load or create configuration profile
func Run(opts Options) (*Result, error) {
	pipeline := NewPipeline(DefaultPhases()...)
	return pipeline.Execute(context.Background(), opts)
}

// openDatabase opens the SQLite/SQLCipher database and runs ent schema migration.
// When encryptionKey is non-empty, PRAGMA key is executed after opening the connection
// to enable SQLCipher transparent encryption.
func openDatabase(dbPath, encryptionKey string, cipherPageSize int) (*ent.Client, *sql.DB, error) {
	// Expand tilde.
	if strings.HasPrefix(dbPath, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			dbPath = filepath.Join(home, dbPath[2:])
		}
	}

	// Ensure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(dbPath), dataDirPerm); err != nil {
		return nil, nil, fmt.Errorf("create db directory: %w", err)
	}

	connStr := "file:" + dbPath + "?cache=shared&_journal_mode=WAL&_busy_timeout=5000"
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, nil, fmt.Errorf("sql open: %w", err)
	}

	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)

	// When encryption key is provided, set SQLCipher PRAGMAs.
	// This requires the binary to be built with SQLCipher support.
	if encryptionKey != "" {
		if cipherPageSize <= 0 {
			cipherPageSize = 4096
		}
		if _, err := db.Exec(fmt.Sprintf("PRAGMA key = '%s'", encryptionKey)); err != nil {
			db.Close()
			return nil, nil, fmt.Errorf("set PRAGMA key: %w", err)
		}
		if _, err := db.Exec(fmt.Sprintf("PRAGMA cipher_page_size = %d", cipherPageSize)); err != nil {
			db.Close()
			return nil, nil, fmt.Errorf("set cipher_page_size: %w", err)
		}
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

// IsDBEncrypted checks whether a SQLite database file is encrypted.
// An encrypted DB will not have the standard "SQLite format 3" magic header.
func IsDBEncrypted(dbPath string) bool {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return false
	}
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

// ensureSecurityTable creates the security_config table if it does not exist.
func ensureSecurityTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS security_config (
			name TEXT PRIMARY KEY,
			value BLOB NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create security_config table: %w", err)
	}

	var count int
	err = db.QueryRow(
		`SELECT count(*) FROM pragma_table_info('security_config') WHERE name='checksum'`,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("check checksum column: %w", err)
	}

	if count == 0 {
		_, err = db.Exec(`ALTER TABLE security_config ADD COLUMN checksum BLOB`)
		if err != nil {
			return fmt.Errorf("add checksum column: %w", err)
		}
	}

	return nil
}

// loadSecurityState reads existing salt and checksum from the database.
// Returns (salt, checksum, firstRun, error).
func loadSecurityState(db *sql.DB) ([]byte, []byte, bool, error) {
	if err := ensureSecurityTable(db); err != nil {
		return nil, nil, false, err
	}

	var salt []byte
	err := db.QueryRow(
		`SELECT value FROM security_config WHERE name = ?`, "default",
	).Scan(&salt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, true, nil // first run
		}
		return nil, nil, false, fmt.Errorf("query salt: %w", err)
	}

	var rawChecksum interface{}
	err = db.QueryRow(
		`SELECT checksum FROM security_config WHERE name = ?`, "default",
	).Scan(&rawChecksum)
	if err != nil && err != sql.ErrNoRows {
		return salt, nil, false, fmt.Errorf("query checksum: %w", err)
	}

	var checksum []byte
	if rawChecksum != nil {
		switch v := rawChecksum.(type) {
		case []byte:
			checksum = v
		case string:
			checksum = []byte(v)
		}
	}

	return salt, checksum, false, nil
}

// storeSalt writes the encryption salt into the security_config table.
func storeSalt(db *sql.DB, salt []byte) error {
	if err := ensureSecurityTable(db); err != nil {
		return err
	}
	_, err := db.Exec(
		`INSERT INTO security_config (name, value) VALUES (?, ?)
		 ON CONFLICT(name) DO UPDATE SET value=excluded.value`,
		"default", salt,
	)
	return err
}

// storeChecksum writes the passphrase checksum into the security_config table.
func storeChecksum(db *sql.DB, checksum []byte) error {
	_, err := db.Exec(
		`UPDATE security_config SET checksum = ? WHERE name = ?`,
		checksum, "default",
	)
	return err
}

// handleNoProfile handles the case where no active profile exists.
// It creates a default profile with sensible defaults.
func handleNoProfile(
	ctx context.Context,
	store *configstore.Store,
) (*config.Config, string, error) {
	cfg := config.DefaultConfig()
	if err := store.Save(ctx, "default", cfg); err != nil {
		return nil, "", fmt.Errorf("save default profile: %w", err)
	}
	if err := store.SetActive(ctx, "default"); err != nil {
		return nil, "", fmt.Errorf("activate default profile: %w", err)
	}

	return cfg, "default", nil
}
