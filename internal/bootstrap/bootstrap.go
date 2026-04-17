package bootstrap

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/mattn/go-sqlite3"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/security"
)

// PhaseTimingEntry records the duration of a bootstrap phase.
type PhaseTimingEntry struct {
	Phase    string        `json:"phase"`
	Duration time.Duration `json:"duration"`
}

// Result holds everything produced by the bootstrap process.
type Result struct {
	// Config is the decrypted, active configuration.
	Config *config.Config
	// ExplicitKeys tracks which context-related config keys the user explicitly set.
	// nil for legacy profiles or when explicit key tracking is unavailable.
	ExplicitKeys map[string]bool
	// AutoEnabled records which context subsystems were auto-enabled during config resolution.
	AutoEnabled config.AutoEnabledSet
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
	// LangoDir is the resolved data directory (Options.LangoDir or ~/.lango).
	// Downstream components (CLI, status, change-passphrase) use this to locate
	// the envelope file without reaching into bootstrap internals.
	LangoDir string
	// IdentityKey is the Ed25519 identity key derived from the Master Key (Phase 3).
	// nil when MK is unavailable (legacy mode).
	IdentityKey ed25519.PrivateKey
	// PQSigningKeySeed is the 32-byte HKDF seed for ML-DSA-65 PQ signing (Phase 5).
	// nil when MK is unavailable. Downstream calls mldsa65.NewKeyFromSeed to derive the key.
	PQSigningKeySeed []byte
	// KMSUnwrap indicates the MK was unwrapped via a KMS KEK slot (not passphrase/mnemonic).
	KMSUnwrap bool
	// PhaseTiming records the duration of each bootstrap phase.
	PhaseTiming []PhaseTimingEntry `json:"phaseTiming,omitempty"`
}

// Options configures the bootstrap process.
type Options struct {
	// LangoDir is the lango data directory (default: ~/.lango/).
	// When set, envelope.json, keyfile (if default), and skills/ are placed under this path.
	// Primarily used by tests to isolate state in t.TempDir().
	LangoDir string
	// DBPath is the SQLite database path (default: <LangoDir>/lango.db).
	DBPath string
	// KeyfilePath is the path to the passphrase keyfile (default: <LangoDir>/keyfile).
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
	// KMSConfig provides KMS settings for KMS KEK slot unwrapping during bootstrap.
	// When set and the envelope has a hardware slot, bootstrap attempts KMS-based
	// MK unwrap before falling back to passphrase.
	KMSConfig *config.KMSConfig
	// KMSProviderName identifies which KMS backend to use for KEK unwrap (e.g., "aws-kms").
	KMSProviderName string
	// Version is the application version string, recorded in bootstrap timing diagnostics.
	Version string
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
	// Explicit Options > env vars: only read env when Options are empty.
	if opts.KMSConfig == nil && opts.KMSProviderName == "" {
		opts.KMSConfig, opts.KMSProviderName = KMSConfigFromEnv()
	}
	pipeline := NewPipeline(DefaultPhases()...)
	return pipeline.Execute(context.Background(), opts)
}

// openDatabase opens the SQLite/SQLCipher database and runs ent schema migration.
// When encryptionKey is non-empty, PRAGMA key is executed after sql.Open.
//
// rawKey distinguishes two SQLCipher key modes:
//   - rawKey=true: encryptionKey is a hex-encoded 32-byte raw key. Uses
//     `PRAGMA key = "x'<hex>'"`. SQLCipher skips its internal PBKDF2.
//     This is the envelope-mode path (DB key = HKDF(MK)).
//   - rawKey=false: encryptionKey is a passphrase. Uses `PRAGMA key = '<passphrase>'`.
//     SQLCipher runs its internal PBKDF2. This is the legacy path.
//
// When encryptionKey is empty, no PRAGMA key is issued and the DB opens in plaintext mode.
func openDatabase(dbPath, encryptionKey string, rawKey bool, cipherPageSize int) (*ent.Client, *sql.DB, error) {
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
		var pragmaKey string
		if rawKey {
			pragmaKey = fmt.Sprintf(`PRAGMA key = "x'%s'"`, encryptionKey)
		} else {
			pragmaKey = fmt.Sprintf("PRAGMA key = '%s'", encryptionKey)
		}
		if _, err := db.Exec(pragmaKey); err != nil {
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

// OpenDatabaseReadOnly opens the SQLite/SQLCipher database in read-only mode
// without invoking ent schema migration.
//
// Contract:
//   - Read-only: the connection uses `file:path?mode=ro`, so writes fail at the
//     SQLite layer.
//   - No schema migration: `client.Schema.Create` is not called, so this
//     function can be used by commands that must not modify the DB.
//   - No prompt: the caller is responsible for obtaining the encryption key
//     non-interactively. Failure to open returns an error; callers must
//     gracefully degrade (zero counts, "unavailable" fields).
//
// rawKey semantics match openDatabase: rawKey=true uses
// `PRAGMA key = "x'<hex>'"`, rawKey=false uses `PRAGMA key = '<passphrase>'`,
// and an empty encryptionKey skips the PRAGMA entirely (plaintext).
//
// The returned *ent.Client shares the underlying *sql.DB so callers should
// Close() the client (which closes the DB) when done.
func OpenDatabaseReadOnly(dbPath, encryptionKey string, rawKey bool, cipherPageSize int) (*ent.Client, *sql.DB, error) {
	// Expand tilde.
	if strings.HasPrefix(dbPath, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			dbPath = filepath.Join(home, dbPath[2:])
		}
	}

	// If the file doesn't exist, fail fast — read-only open on a nonexistent
	// DB would still succeed but produce an empty database, which is misleading.
	if _, err := os.Stat(dbPath); err != nil {
		return nil, nil, fmt.Errorf("read-only db open: stat %q: %w", dbPath, err)
	}

	connStr := "file:" + dbPath + "?mode=ro&cache=shared&_journal_mode=WAL&_busy_timeout=5000"
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, nil, fmt.Errorf("read-only sql open: %w", err)
	}

	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(2)

	if encryptionKey != "" {
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
			db.Close()
			return nil, nil, fmt.Errorf("read-only PRAGMA key: %w", err)
		}
		if _, err := db.Exec(fmt.Sprintf("PRAGMA cipher_page_size = %d", cipherPageSize)); err != nil {
			db.Close()
			return nil, nil, fmt.Errorf("read-only cipher_page_size: %w", err)
		}
	}

	// Sanity check: run a trivial query to confirm the key (if any) is correct.
	// This catches "wrong PRAGMA key → DB looks OK but queries fail later".
	if _, err := db.Exec("PRAGMA schema_version"); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("read-only db verify: %w", err)
	}

	// NOTE: intentionally no Schema.Create — this is the whole point of the
	// read-only path. Callers that need a migrated schema should use bootstrap.Run.
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))
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

// loadSecurityState reads existing salt and checksum from the database via
// the shared SecurityConfigStore. Returns (salt, checksum, firstRun, error).
func loadSecurityState(db *sql.DB) ([]byte, []byte, bool, error) {
	store := security.NewSecurityConfigStore(db)
	if err := store.EnsureTable(); err != nil {
		return nil, nil, false, err
	}
	salt, err := store.LoadSalt()
	if err != nil {
		return nil, nil, false, err
	}
	if salt == nil {
		return nil, nil, true, nil // first run
	}
	checksum, err := store.LoadChecksum()
	if err != nil {
		return salt, nil, false, err
	}
	return salt, checksum, false, nil
}

// storeSalt writes the encryption salt into the security_config table.
func storeSalt(db *sql.DB, salt []byte) error {
	return security.NewSecurityConfigStore(db).StoreSalt(salt)
}

// storeChecksum writes the passphrase checksum into the security_config table.
func storeChecksum(db *sql.DB, checksum []byte) error {
	return security.NewSecurityConfigStore(db).StoreChecksum(checksum)
}

// handleNoProfile handles the case where no active profile exists.
// It creates a default profile with sensible defaults.
func handleNoProfile(
	ctx context.Context,
	store *configstore.Store,
) (*config.Config, string, error) {
	cfg := config.DefaultConfig()
	if err := store.Save(ctx, "default", cfg, nil); err != nil {
		return nil, "", fmt.Errorf("save default profile: %w", err)
	}
	if err := store.SetActive(ctx, "default"); err != nil {
		return nil, "", fmt.Errorf("activate default profile: %w", err)
	}

	return cfg, "default", nil
}
