package bootstrap

import (
	"context"
	"crypto/hmac"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/langoai/lango/internal/cli/prompt"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
	"github.com/langoai/lango/internal/keyring"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/security/passphrase"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/storagebroker"
)

var startStorageBroker = func(ctx context.Context) (storagebroker.API, error) {
	return storagebroker.Start(ctx)
}

func loadSecurityStateForState(ctx context.Context, s *State) ([]byte, []byte, bool, error) {
	if s != nil && s.Broker != nil {
		return loadSecurityStateViaBroker(ctx, s.Broker)
	}
	return loadSecurityState(s.RawDB)
}

func storeSaltForState(ctx context.Context, s *State, salt []byte) error {
	if s != nil && s.Broker != nil {
		return storeSaltViaBroker(ctx, s.Broker, salt)
	}
	return storeSalt(s.RawDB, salt)
}

func storeChecksumForState(ctx context.Context, s *State, checksum []byte) error {
	if s != nil && s.Broker != nil {
		return storeChecksumViaBroker(ctx, s.Broker, checksum)
	}
	return storeChecksum(s.RawDB, checksum)
}

// dataDirPerm is the permission mode for all ~/.lango/ directories.
// 0700 restricts access to the owner only (appropriate for data containing
// encrypted secrets, database files, and keyfiles).
const dataDirPerm = 0700

// DefaultPhases returns the standard bootstrap phase sequence (11 phases).
//
// Envelope-aware order: envelope loads BEFORE DB open so recovery credentials
// (mnemonic) and MK-derived DB keys are available when we actually open
// SQLCipher. Legacy installations follow the same pipeline but land in
// MigrateEnvelope which performs a one-time upgrade.
func DefaultPhases() []Phase {
	return []Phase{
		phaseEnsureDataDir(),
		phaseDetectEncryption(),
		phaseLoadEnvelopeFile(),
		phaseAcquireCredential(),
		phaseUnwrapOrCreateMK(),
		phaseOpenDatabase(),
		phaseLoadSecurityState(),
		phaseMigrateEnvelope(),
		phaseInitCrypto(),
		phaseDeriveIdentityKey(),
		phaseDerivePQKey(),
		phaseLoadProfile(),
	}
}

// phaseEnsureDataDir resolves the lango data directory (honoring Options.LangoDir),
// creates it with 0700 permissions, and fills in default paths for DBPath and KeyfilePath.
func phaseEnsureDataDir() Phase {
	return Phase{
		Name: "ensure data directory",
		Run: func(_ context.Context, s *State) error {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("resolve home directory: %w", err)
			}
			s.Home = home
			if s.Options.LangoDir != "" {
				s.LangoDir = s.Options.LangoDir
			} else {
				s.LangoDir = filepath.Join(home, ".lango")
			}

			if s.Options.DBPath == "" {
				s.Options.DBPath = filepath.Join(s.LangoDir, "lango.db")
			}
			if s.Options.KeyfilePath == "" {
				s.Options.KeyfilePath = filepath.Join(s.LangoDir, "keyfile")
			}

			if err := os.MkdirAll(s.LangoDir, dataDirPerm); err != nil {
				return fmt.Errorf("create data directory: %w", err)
			}

			// Verify the directory is actually writable by the current user.
			// Docker volumes may have stale ownership from a previous build.
			testPath := filepath.Join(s.LangoDir, ".write-test")
			if err := os.WriteFile(testPath, []byte{}, 0600); err != nil {
				return fmt.Errorf("data directory not writable (uid %d): %w", os.Getuid(), err)
			}
			os.Remove(testPath)

			// Pre-create the skills directory so FileSkillStore can write immediately.
			skillsDir := filepath.Join(s.LangoDir, "skills")
			if err := os.MkdirAll(skillsDir, dataDirPerm); err != nil {
				return fmt.Errorf("create skills directory: %w", err)
			}

			// Expose the resolved lango dir to downstream CLI/tools.
			s.Result.LangoDir = s.LangoDir
			return nil
		},
	}
}

// phaseDetectEncryption checks if DB is encrypted or encryption is configured.
func phaseDetectEncryption() Phase {
	return Phase{
		Name: "detect encryption",
		Run: func(_ context.Context, s *State) error {
			s.DBEncrypted = IsDBEncrypted(s.Options.DBPath)
			s.NeedsDBKey = s.DBEncrypted || s.Options.DBEncryption.Enabled
			return nil
		},
	}
}

// phaseLoadEnvelopeFile attempts to load <LangoDir>/envelope.json.
// On success populates s.Envelope; on "file does not exist" sets it to nil.
// A corrupt file (invalid JSON, wrong version) fails the phase.
func phaseLoadEnvelopeFile() Phase {
	return Phase{
		Name: "load envelope file",
		Run: func(_ context.Context, s *State) error {
			env, err := security.LoadEnvelopeFile(s.LangoDir)
			if err != nil {
				return fmt.Errorf("load envelope: %w", err)
			}
			s.Envelope = env
			return nil
		},
	}
}

// phaseAcquireCredential attempts KMS unwrap first (if configured),
// then falls back to the passphrase acquisition chain (keyring, keyfile, interactive, stdin).
func phaseAcquireCredential() Phase {
	return Phase{
		Name: "acquire credential",
		Run: func(ctx context.Context, s *State) error {
			// KMS path: if envelope has a hardware/KMS slot and KMS config
			// is provided, attempt to create a bare KMS provider and unwrap
			// the MK directly. On success, skip passphrase entirely.
			// On failure, fall through to passphrase acquisition.
			if s.Envelope != nil && s.Envelope.HasSlotType(security.KEKSlotHardware) &&
				s.Options.KMSConfig != nil && s.Options.KMSProviderName != "" {

				kmsProvider, kmsErr := security.NewKMSProvider(
					security.KMSProviderName(s.Options.KMSProviderName),
					*s.Options.KMSConfig,
				)
				if kmsErr == nil { //nolint:staticcheck // stubs always error; real impls use kms_* build tags
					mk, _, unwrapErr := s.Envelope.UnwrapFromKMS(
						ctx, kmsProvider, s.Options.KMSProviderName, s.Options.KMSConfig.KeyID,
					)
					if unwrapErr == nil {
						s.MasterKey = mk
						s.KMSUnwrap = true
						s.KMSProvider = kmsProvider
						s.Result.KMSUnwrap = true
						return nil // Skip passphrase entirely.
					}
					fmt.Fprintf(os.Stderr, "warning: KMS unwrap failed: %v (falling back to passphrase)\n", unwrapErr)
				} else {
					fmt.Fprintf(os.Stderr, "warning: KMS provider init failed: %v (falling back to passphrase)\n", kmsErr)
				}
			}

			// Detect secure provider (biometric/TPM).
			if !s.Options.SkipSecureDetection {
				s.SecureProvider, s.SecurityTier = keyring.DetectSecureProvider()
			}

			// Determine if this is a first-run scenario: no DB file AND no envelope.
			_, statErr := os.Stat(s.Options.DBPath)
			s.FirstRunGuess = statErr != nil && s.Envelope == nil

			pass, source, err := passphrase.Acquire(passphrase.Options{
				KeyfilePath:     s.Options.KeyfilePath,
				AllowCreation:   s.FirstRunGuess,
				KeyringProvider: s.SecureProvider,
			})
			if err != nil {
				return fmt.Errorf("acquire passphrase: %w", err)
			}
			s.Passphrase = pass
			s.PassSource = source

			// Offer to store passphrase when secure hardware is available.
			if source == passphrase.SourceInteractive && s.SecureProvider != nil {
				tierLabel := s.SecurityTier.String()
				msg := fmt.Sprintf("Secure storage available (%s). Store passphrase?", tierLabel)
				if ok, promptErr := prompt.Confirm(msg); promptErr == nil && ok {
					if storeErr := s.SecureProvider.Set(keyring.Service, keyring.KeyMasterPassphrase, pass); storeErr != nil {
						if errors.Is(storeErr, keyring.ErrEntitlement) {
							fmt.Fprintf(os.Stderr, "warning: biometric storage unavailable (binary not codesigned)\n")
							fmt.Fprintf(os.Stderr, "  Tip: codesign the binary for Touch ID support: make codesign\n")
							fmt.Fprintf(os.Stderr, "  Note: also ensure device passcode is set (required for biometric Keychain)\n")
						} else {
							fmt.Fprintf(os.Stderr, "warning: store passphrase failed: %v\n", storeErr)
						}
					} else {
						fmt.Fprintf(os.Stderr, "Passphrase saved. Next launch will load it automatically.\n")
					}
				}
			}

			return nil
		},
	}
}

// phaseUnwrapOrCreateMK covers three cases:
//  1. MasterKey already set (by KMS unwrap) — no-op.
//  2. Envelope exists — derive KEK from passphrase and unwrap the MK.
//  3. No envelope and no legacy DB — first run: create MK + envelope.
//  4. No envelope but legacy DB — mark LegacyMode; migration handles it.
func phaseUnwrapOrCreateMK() Phase {
	return Phase{
		Name: "unwrap or create master key",
		Run: func(_ context.Context, s *State) error {
			if s.MasterKey != nil {
				// Already unwrapped (e.g. KMS).
				return nil
			}
			if s.Envelope != nil {
				mk, _, err := s.Envelope.UnwrapFromPassphrase(s.Passphrase)
				if err != nil {
					return fmt.Errorf("unwrap master key: %w", err)
				}
				s.MasterKey = mk
				return nil
			}
			// No envelope. Decide between first-run and legacy upgrade.
			if s.FirstRunGuess {
				env, mk, err := security.NewEnvelope(s.Passphrase)
				if err != nil {
					return fmt.Errorf("create envelope: %w", err)
				}
				if err := security.StoreEnvelopeFile(s.LangoDir, env); err != nil {
					security.ZeroBytes(mk)
					return fmt.Errorf("store envelope: %w", err)
				}
				s.Envelope = env
				s.MasterKey = mk
				return nil
			}
			// Legacy DB exists; migration phase will handle it after DB open.
			s.LegacyMode = true
			return nil
		},
	}
}

// phaseOpenDatabase opens SQLite/SQLCipher DB and runs ent schema migration.
//
// Key selection matrix:
//   - MK available + no pending flags  → MK-derived raw key (rawKey=true)
//   - MK available + pending flags     → legacy passphrase (fallback)
//   - No MK (legacy mode)              → legacy passphrase
//   - Encryption disabled              → no key
func phaseOpenDatabase() Phase {
	return Phase{
		Name: "open database",
		Run: func(_ context.Context, s *State) error {
			var (
				dbKey  string
				rawKey bool
			)
			if s.NeedsDBKey {
				switch {
				case s.MasterKey != nil && s.Envelope != nil &&
					!s.Envelope.PendingMigration && !s.Envelope.PendingRekey:
					dbKey = security.DeriveDBKeyHex(s.MasterKey)
					rawKey = true
				case s.Passphrase != "":
					dbKey = s.Passphrase
					rawKey = false
				default:
					return fmt.Errorf("open database: no credential available for encrypted db")
				}
				s.DBKey = dbKey
			}
			if s.Options.StartStorageBroker {
				brokerClient, err := startStorageBroker(context.Background())
				if err != nil {
					return fmt.Errorf("start storage broker: %w", err)
				}
				if _, err := brokerClient.OpenDB(context.Background(), storagebroker.OpenDBRequest{
					DBPath:         s.Options.DBPath,
					EncryptionKey:  dbKey,
					RawKey:         rawKey,
					CipherPageSize: s.Options.DBEncryption.CipherPageSize,
				}); err != nil {
					_ = brokerClient.Close(context.Background())
					return fmt.Errorf("storage broker open_db: %w", err)
				}
				s.Broker = brokerClient
				s.Result.Broker = brokerClient
			}
			client, rawDB, err := openDatabase(s.Options.DBPath, dbKey, rawKey, s.Options.DBEncryption.CipherPageSize)
			if err != nil {
				if s.Broker != nil {
					_ = s.Broker.Close(context.Background())
					s.Broker = nil
					s.Result.Broker = nil
				}
				return fmt.Errorf("open database: %w", err)
			}
			s.Client = client
			s.RawDB = rawDB
			return nil
		},
		Cleanup: func(s *State) {
			if s.Broker != nil {
				_ = s.Broker.Close(context.Background())
				s.Broker = nil
			}
			if s.Client != nil {
				_ = s.Client.Close()
			}
		},
	}
}

// phaseMigrateEnvelope performs three conditional actions:
//
//   - LegacyMode: run the full legacy → envelope migration.
//   - PendingMigration=true: retry data re-encryption using the already-unwrapped MK.
//   - PendingRekey=true: retry `PRAGMA rekey` on SQLCipher DB.
//
// On failure the envelope retains its pending flags so the next bootstrap can retry.
func phaseMigrateEnvelope() Phase {
	return Phase{
		Name: "migrate envelope",
		Run: func(ctx context.Context, s *State) error {
			if s.LegacyMode {
				fmt.Fprintln(os.Stderr, "Upgrading encryption format (one-time migration)...")
				env, mk, err := security.MigrateToEnvelope(
					ctx, s.RawDB, s.Client, s.LangoDir,
					s.Passphrase, s.Salt, s.Checksum,
					s.NeedsDBKey,
				)
				if err != nil {
					return fmt.Errorf("legacy migration: %w", err)
				}
				s.Envelope = env
				s.MasterKey = mk
				return nil
			}

			if s.Envelope != nil && s.Envelope.PendingMigration {
				if s.MasterKey == nil {
					return fmt.Errorf("pending migration retry requires unwrapped master key")
				}
				if err := security.RetryMigration(ctx, s.Client, s.MasterKey, s.Passphrase, s.Salt); err != nil {
					return fmt.Errorf("retry migration: %w", err)
				}
				s.Envelope.PendingMigration = false
				if err := security.StoreEnvelopeFile(s.LangoDir, s.Envelope); err != nil {
					return fmt.Errorf("persist envelope after retry migration: %w", err)
				}
			}

			if s.Envelope != nil && s.Envelope.PendingRekey {
				if s.MasterKey == nil {
					return fmt.Errorf("pending rekey retry requires unwrapped master key")
				}
				if err := security.RetryRekey(s.RawDB, s.MasterKey); err != nil {
					return fmt.Errorf("retry rekey: %w", err)
				}
				s.Envelope.PendingRekey = false
				if err := security.StoreEnvelopeFile(s.LangoDir, s.Envelope); err != nil {
					return fmt.Errorf("persist envelope after retry rekey: %w", err)
				}
			}

			return nil
		},
	}
}

// phaseLoadSecurityState reads legacy salt and checksum from the database.
// These are only used when an envelope has NOT been installed yet (i.e. on
// installations that predate this change and will be migrated in Phase 7).
func phaseLoadSecurityState() Phase {
	return Phase{
		Name: "load security state",
		Run: func(_ context.Context, s *State) error {
			if s.Envelope != nil {
				// When pending flags are set, legacy salt/checksum are still
				// needed for crash-recovery retry in phaseMigrateEnvelope.
				if s.Envelope.PendingMigration || s.Envelope.PendingRekey {
					salt, checksum, _, err := loadSecurityStateForState(context.Background(), s)
					if err != nil {
						return fmt.Errorf("load security state for pending migration: %w", err)
					}
					s.Salt = salt
					s.Checksum = checksum
				}
				s.FirstRun = false
				return nil
			}
			salt, checksum, firstRun, err := loadSecurityStateForState(context.Background(), s)
			if err != nil {
				return fmt.Errorf("load security state: %w", err)
			}
			s.Salt = salt
			s.Checksum = checksum
			s.FirstRun = firstRun
			return nil
		},
	}
}

// phaseInitCrypto installs the Master Key (or legacy passphrase-derived key)
// into the LocalCryptoProvider and shreds the keyfile on success.
func phaseInitCrypto() Phase {
	return Phase{
		Name: "initialize crypto",
		Run: func(_ context.Context, s *State) error {
			provider := security.NewLocalCryptoProvider()

			switch {
			case s.MasterKey != nil && s.Envelope != nil:
				// Envelope path: install the already-unwrapped MK.
				if err := provider.InitializeWithEnvelope(s.MasterKey, s.Envelope); err != nil {
					return fmt.Errorf("initialize crypto with envelope: %w", err)
				}
			case s.FirstRun:
				// Legacy first run (should be rare now — UnwrapOrCreateMK handles
				// first runs via NewEnvelope). Kept as a safety net.
				if err := provider.Initialize(s.Passphrase); err != nil {
					return fmt.Errorf("initialize crypto: %w", err)
				}
				if err := storeSaltForState(context.Background(), s, provider.Salt()); err != nil {
					return fmt.Errorf("store salt: %w", err)
				}
				cs := provider.CalculateChecksum(s.Passphrase, provider.Salt())
				if err := storeChecksumForState(context.Background(), s, cs); err != nil {
					return fmt.Errorf("store checksum: %w", err)
				}
			default:
				// Legacy returning-user path. Used for DBs that still have
				// salt/checksum and no envelope (shouldn't reach here after
				// successful migration).
				if err := provider.InitializeWithSalt(s.Passphrase, s.Salt); err != nil {
					return fmt.Errorf("initialize crypto with salt: %w", err)
				}
				if s.Checksum != nil {
					computed := provider.CalculateChecksum(s.Passphrase, s.Salt)
					if !hmac.Equal(s.Checksum, computed) {
						return fmt.Errorf("passphrase checksum mismatch: incorrect passphrase")
					}
				}
			}

			// Shred keyfile after successful crypto initialization.
			if s.PassSource == passphrase.SourceKeyfile && !s.Options.KeepKeyfile {
				if err := passphrase.ShredKeyfile(s.Options.KeyfilePath); err != nil {
					fmt.Fprintf(os.Stderr, "warning: shred keyfile: %v\n", err)
				}
			}

			s.Crypto = provider
			s.Result.Crypto = provider
			return nil
		},
	}
}

// phaseDeriveIdentityKey derives the Ed25519 identity key from the Master Key
// via HKDF. This key is used by BundleProvider for DID v2 identity.
// No-op when MK is unavailable (legacy mode).
func phaseDeriveIdentityKey() Phase {
	return Phase{
		Name: "derive identity key",
		Run: func(_ context.Context, s *State) error {
			if s.MasterKey == nil {
				return nil // No MK = no identity key (legacy mode)
			}
			s.IdentityKey = security.DeriveIdentityKey(s.MasterKey, 0)
			s.Result.IdentityKey = s.IdentityKey
			return nil
		},
	}
}

// phaseDerivePQKey derives the ML-DSA-65 PQ signing key seed from the Master Key
// via HKDF. The 32-byte seed is stored in Result; downstream code derives the
// full ML-DSA-65 keypair via mldsa65.NewKeyFromSeed.
// No-op when MK is unavailable (legacy mode).
func phaseDerivePQKey() Phase {
	return Phase{
		Name: "derive PQ signing key",
		Run: func(_ context.Context, s *State) error {
			if s.MasterKey == nil {
				return nil
			}
			s.PQSigningKeySeed = security.DerivePQSigningSeed(s.MasterKey, 0)
			s.Result.PQSigningKeySeed = s.PQSigningKeySeed
			return nil
		},
	}
}

// phaseLoadProfile loads or creates the configuration profile.
func phaseLoadProfile() Phase {
	return Phase{
		Name: "load profile",
		Run: func(ctx context.Context, s *State) error {
			store := configstore.NewStore(s.Client, s.Crypto)
			s.Result.ConfigStore = store
			profileName := s.Options.ForceProfile

			var explicitKeys map[string]bool

			if profileName != "" {
				cfg, keys, err := store.Load(ctx, profileName)
				if err != nil {
					return fmt.Errorf("load profile %q: %w", profileName, err)
				}
				s.Result.Config = cfg
				s.Result.ProfileName = profileName
				explicitKeys = keys
			} else {
				name, cfg, keys, err := store.LoadActive(ctx)
				if err != nil && !errors.Is(err, configstore.ErrNoActiveProfile) {
					return fmt.Errorf("load active profile: %w", err)
				}
				if errors.Is(err, configstore.ErrNoActiveProfile) {
					resultCfg, resultName, handleErr := handleNoProfile(ctx, store)
					if handleErr != nil {
						return handleErr
					}
					s.Result.Config = resultCfg
					s.Result.ProfileName = resultName
					// New profile: no explicit keys.
				} else {
					s.Result.Config = cfg
					s.Result.ProfileName = name
					explicitKeys = keys
				}
			}

			// Apply context profile and auto-enable resolution.
			config.ApplyContextProfile(s.Result.Config, explicitKeys)
			autoEnabled := config.ResolveContextAutoEnable(s.Result.Config, explicitKeys)
			s.Result.ExplicitKeys = explicitKeys
			s.Result.AutoEnabled = autoEnabled

			// Single post-load: normalize + validate for all branches.
			if err := config.PostLoad(s.Result.Config); err != nil {
				return fmt.Errorf("post-load config: %w", err)
			}

			s.Result.Storage = storage.NewFacade(
				store,
				security.NewSecurityConfigStore(s.RawDB),
				storage.WithEntClient(s.Client),
				storage.WithRawDB(s.RawDB),
				storage.WithSessionClient(s.Client),
				storage.WithSessionDBPath(s.Result.Config.Session.DatabasePath),
			)
			return nil
		},
	}
}
