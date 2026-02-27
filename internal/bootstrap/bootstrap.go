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

	"github.com/langoai/lango/internal/cli/prompt"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/keyring"
	"github.com/langoai/lango/internal/security/passphrase"
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

// Run executes the full bootstrap sequence:
//  1. Ensure ~/.lango/ directory
//  2. Detect DB encryption status
//  3. Acquire passphrase
//  4. Open SQLite/SQLCipher DB + ent schema migration
//  5. Load security state (salt/checksum)
//  6. Initialize crypto provider
//  7. Load or create configuration profile
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

	// 2. Determine whether the DB file is encrypted (or encryption is configured).
	dbEncrypted := IsDBEncrypted(opts.DBPath)
	needsDBKey := dbEncrypted || opts.DBEncryption.Enabled

	// 3. Acquire passphrase.
	// If the DB is encrypted or encryption is enabled, we need the passphrase
	// BEFORE opening the database so we can derive the DB encryption key.
	//
	// DetectSecureProvider returns the highest-tier hardware-backed provider:
	//   TierBiometric (macOS Touch ID) > TierTPM (Linux) > TierNone.
	// When TierNone, secureProvider is nil and keyring auto-read is disabled,
	// preventing same-UID attacks on plain OS keyring storage.
	var secureProvider keyring.Provider
	var tier keyring.SecurityTier
	if !opts.SkipSecureDetection {
		secureProvider, tier = keyring.DetectSecureProvider()
	}

	// Determine if this is a first-run scenario: no DB file at all.
	_, statErr := os.Stat(opts.DBPath)
	dbExists := statErr == nil
	firstRunGuess := !dbExists

	pass, source, err := passphrase.Acquire(passphrase.Options{
		KeyfilePath:     opts.KeyfilePath,
		AllowCreation:   firstRunGuess,
		KeyringProvider: secureProvider, // nil on TierNone → keyring read skipped
	})
	if err != nil {
		return nil, fmt.Errorf("acquire passphrase: %w", err)
	}

	// 3b. Offer to store passphrase only when a secure hardware provider is available.
	if source == passphrase.SourceInteractive && secureProvider != nil {
		tierLabel := tier.String() // "biometric" or "tpm"
		msg := fmt.Sprintf("Secure storage available (%s). Store passphrase?", tierLabel)
		if ok, promptErr := prompt.Confirm(msg); promptErr == nil && ok {
			if storeErr := secureProvider.Set(keyring.Service, keyring.KeyMasterPassphrase, pass); storeErr != nil {
				fmt.Fprintf(os.Stderr, "warning: store passphrase: %v\n", storeErr)
			}
		}
	}

	// 4. Open database with encryption key if needed.
	var dbKey string
	if needsDBKey {
		dbKey = pass
	}
	client, rawDB, err := openDatabase(opts.DBPath, dbKey, opts.DBEncryption.CipherPageSize)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// 5. Check existing salt to determine first-run vs returning user.
	salt, checksum, firstRun, err := loadSecurityState(rawDB)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("load security state: %w", err)
	}

	// 6. Initialize crypto provider.
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

	// 6b. Shred keyfile after successful crypto initialization.
	if source == passphrase.SourceKeyfile && !opts.KeepKeyfile {
		if err := passphrase.ShredKeyfile(opts.KeyfilePath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: shred keyfile: %v\n", err)
		}
	}

	// 7. Load or create configuration profile.
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
			cfg, profileName, err = handleNoProfile(ctx, store)
			if err != nil {
				client.Close()
				return nil, err
			}
		}
	}

	return &Result{
		Config:      cfg,
		DBClient:    client,
		RawDB:       rawDB,
		Crypto:      provider,
		ConfigStore: store,
		ProfileName: profileName,
	}, nil
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
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
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
