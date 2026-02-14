package bootstrap

import (
	"context"
	"crypto/hmac"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/mattn/go-sqlite3"

	"github.com/langowarny/lango/internal/config"
	"github.com/langowarny/lango/internal/configstore"
	"github.com/langowarny/lango/internal/ent"
	"github.com/langowarny/lango/internal/passphrase"
	"github.com/langowarny/lango/internal/security"
)

// Result holds everything produced by the bootstrap process.
type Result struct {
	// Config is the decrypted, active configuration.
	Config *config.Config
	// DBClient is the shared ent.Client for the application database.
	DBClient *ent.Client
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
	// MigrationPath is the path to an existing lango.json for migration.
	MigrationPath string
	// ForceProfile overrides the active profile selection.
	ForceProfile string
}

// Run executes the full bootstrap sequence:
//  1. Ensure ~/.lango/ directory
//  2. Open SQLite DB + ent schema migration
//  3. Acquire passphrase
//  4. Initialize crypto provider (salt/checksum)
//  5. Load or create configuration profile
func Run(opts Options) (*Result, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory: %w", err)
	}

	langoDir := filepath.Join(home, ".lango")
	if opts.DBPath == "" {
		opts.DBPath = filepath.Join(langoDir, "lango.db")
	}
	if opts.KeyfilePath == "" {
		opts.KeyfilePath = filepath.Join(langoDir, "keyfile")
	}

	// 1. Ensure data directory exists.
	if err := os.MkdirAll(langoDir, 0700); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	// 2. Open database and run schema migration.
	client, rawDB, err := openDatabase(opts.DBPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// 3. Check existing salt to determine first-run vs returning user.
	salt, checksum, firstRun, err := loadSecurityState(rawDB)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("load security state: %w", err)
	}

	// 4. Acquire passphrase.
	pass, _, err := passphrase.Acquire(passphrase.Options{
		KeyfilePath:   opts.KeyfilePath,
		AllowCreation: firstRun,
	})
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("acquire passphrase: %w", err)
	}

	// 5. Initialize crypto provider.
	provider := security.NewLocalCryptoProvider()
	if firstRun {
		if err := provider.Initialize(pass); err != nil {
			client.Close()
			return nil, fmt.Errorf("initialize crypto: %w", err)
		}
		if err := storeSalt(rawDB, provider.Salt()); err != nil {
			client.Close()
			return nil, fmt.Errorf("store salt: %w", err)
		}
		cs := provider.CalculateChecksum(pass, provider.Salt())
		if err := storeChecksum(rawDB, cs); err != nil {
			client.Close()
			return nil, fmt.Errorf("store checksum: %w", err)
		}
	} else {
		if err := provider.InitializeWithSalt(pass, salt); err != nil {
			client.Close()
			return nil, fmt.Errorf("initialize crypto with salt: %w", err)
		}
		if checksum != nil {
			computed := provider.CalculateChecksum(pass, salt)
			if !hmac.Equal(checksum, computed) {
				client.Close()
				return nil, fmt.Errorf("passphrase checksum mismatch: incorrect passphrase")
			}
		}
	}

	// 6. Load or create configuration profile.
	store := configstore.NewStore(client, provider)
	ctx := context.Background()

	profileName := opts.ForceProfile
	var cfg *config.Config

	if profileName != "" {
		cfg, err = store.Load(ctx, profileName)
		if err != nil {
			client.Close()
			return nil, fmt.Errorf("load profile %q: %w", profileName, err)
		}
	} else {
		profileName, cfg, err = store.LoadActive(ctx)
		if err != nil && !errors.Is(err, configstore.ErrNoActiveProfile) {
			client.Close()
			return nil, fmt.Errorf("load active profile: %w", err)
		}

		if errors.Is(err, configstore.ErrNoActiveProfile) {
			cfg, profileName, err = handleNoProfile(ctx, store, opts.MigrationPath)
			if err != nil {
				client.Close()
				return nil, err
			}
		}
	}

	return &Result{
		Config:      cfg,
		DBClient:    client,
		Crypto:      provider,
		ConfigStore: store,
		ProfileName: profileName,
	}, nil
}

// openDatabase opens the SQLite database and runs ent schema migration.
func openDatabase(dbPath string) (*ent.Client, *sql.DB, error) {
	// Expand tilde.
	if strings.HasPrefix(dbPath, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			dbPath = filepath.Join(home, dbPath[2:])
		}
	}

	// Ensure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		return nil, nil, fmt.Errorf("create db directory: %w", err)
	}

	connStr := "file:" + dbPath + "?cache=shared"
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, nil, fmt.Errorf("sql open: %w", err)
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
// It tries to migrate from an existing lango.json or creates a default profile.
func handleNoProfile(
	ctx context.Context,
	store *configstore.Store,
	migrationPath string,
) (*config.Config, string, error) {
	// Try migration from lango.json.
	jsonPaths := candidateJSONPaths(migrationPath)
	for _, p := range jsonPaths {
		if _, err := os.Stat(p); err == nil {
			if err := configstore.MigrateFromJSON(ctx, store, p, "default"); err != nil {
				return nil, "", fmt.Errorf("migrate from %s: %w", p, err)
			}
			cfg, err := store.Load(ctx, "default")
			if err != nil {
				return nil, "", fmt.Errorf("load migrated profile: %w", err)
			}
			return cfg, "default", nil
		}
	}

	// No existing config — create default profile.
	cfg := config.DefaultConfig()
	if err := store.Save(ctx, "default", cfg); err != nil {
		return nil, "", fmt.Errorf("save default profile: %w", err)
	}
	if err := store.SetActive(ctx, "default"); err != nil {
		return nil, "", fmt.Errorf("activate default profile: %w", err)
	}

	return cfg, "default", nil
}

// candidateJSONPaths returns paths to search for an existing lango.json.
func candidateJSONPaths(explicit string) []string {
	var paths []string
	if explicit != "" {
		paths = append(paths, explicit)
	}
	paths = append(paths, "lango.json")
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".lango", "lango.json"))
	}
	return paths
}
