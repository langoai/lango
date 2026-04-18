package bootstrap

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
	"github.com/langoai/lango/internal/dbopen"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/storagebroker"
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
	// Broker is the optional storage broker client used during transitional
	// broker-boundary rollout and diagnostics.
	Broker storagebroker.API
	// Crypto is the initialized CryptoProvider for the session.
	Crypto security.CryptoProvider
	// ConfigStore is retained as a transitional compatibility handle. New code
	// should prefer Storage.ConfigProfiles().
	ConfigStore *configstore.Store
	// Storage is the sole bootstrap-owned access point for database-backed
	// capabilities. Callers should not depend on raw DB handles.
	Storage *storage.Facade
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
	// StartStorageBroker enables the transitional bootstrap path that starts a
	// long-lived broker subprocess and completes an open-db handshake.
	StartStorageBroker bool
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

// Close releases resources held by the bootstrap result.
func (r *Result) Close() error {
	if r == nil {
		return nil
	}
	if r.Broker != nil {
		_ = r.Broker.Close(context.Background())
		r.Broker = nil
	}
	if r.Storage != nil {
		return r.Storage.Close()
	}
	return nil
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
	return dbopen.OpenManaged(dbPath, encryptionKey, rawKey, cipherPageSize)
}

// OpenDatabaseManaged opens the application database in read-write mode and
// applies schema migration. It is intended for infrastructure owners such as
// the storage broker.
func OpenDatabaseManaged(dbPath, encryptionKey string, rawKey bool, cipherPageSize int) (*ent.Client, *sql.DB, error) {
	return dbopen.OpenManaged(dbPath, encryptionKey, rawKey, cipherPageSize)
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
	return dbopen.OpenReadOnly(dbPath, encryptionKey, rawKey, cipherPageSize)
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

func loadSecurityStateViaBroker(ctx context.Context, broker storagebroker.API) ([]byte, []byte, bool, error) {
	if broker == nil {
		return nil, nil, false, fmt.Errorf("storage broker not available")
	}
	result, err := broker.LoadSecurityState(ctx)
	if err != nil {
		return nil, nil, false, err
	}
	return result.Salt, result.Checksum, result.FirstRun, nil
}

// storeSalt writes the encryption salt into the security_config table.
func storeSalt(db *sql.DB, salt []byte) error {
	return security.NewSecurityConfigStore(db).StoreSalt(salt)
}

func storeSaltViaBroker(ctx context.Context, broker storagebroker.API, salt []byte) error {
	if broker == nil {
		return fmt.Errorf("storage broker not available")
	}
	return broker.StoreSalt(ctx, salt)
}

// storeChecksum writes the passphrase checksum into the security_config table.
func storeChecksum(db *sql.DB, checksum []byte) error {
	return security.NewSecurityConfigStore(db).StoreChecksum(checksum)
}

func storeChecksumViaBroker(ctx context.Context, broker storagebroker.API, checksum []byte) error {
	if broker == nil {
		return fmt.Errorf("storage broker not available")
	}
	return broker.StoreChecksum(ctx, checksum)
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
