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
)

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
		phaseMigrateEnvelope(),
		phaseLoadSecurityState(),
		phaseInitCrypto(),
		phaseDeriveIdentityKey(),
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

// phaseAcquireCredential acquires the passphrase, or offers mnemonic recovery
// if the envelope contains a mnemonic slot and the terminal is interactive.
func phaseAcquireCredential() Phase {
	return Phase{
		Name: "acquire credential",
		Run: func(_ context.Context, s *State) error {
			// Detect secure provider (biometric/TPM).
			if !s.Options.SkipSecureDetection {
				s.SecureProvider, s.SecurityTier = keyring.DetectSecureProvider()
			}

			// Determine if this is a first-run scenario: no DB file AND no envelope.
			_, statErr := os.Stat(s.Options.DBPath)
			s.FirstRunGuess = statErr != nil && s.Envelope == nil

			// Recovery path: offer mnemonic choice when the envelope has a
			// mnemonic slot and we're on an interactive terminal. The interactive
			// check relies on passphrase.Acquire's TTY detection; the choice
			// prompt itself uses prompt.Confirm which already gates on TTY.
			if s.Envelope != nil && s.Envelope.HasSlotType(security.KEKSlotMnemonic) {
				if ok, promptErr := prompt.Confirm("Recovery mnemonic slot detected. Recover with mnemonic?"); promptErr == nil && ok {
					mnemonic, mErr := prompt.Passphrase("Enter 24-word recovery mnemonic: ")
					if mErr != nil {
						return fmt.Errorf("read mnemonic: %w", mErr)
					}
					if err := security.ValidateMnemonic(mnemonic); err != nil {
						return fmt.Errorf("invalid mnemonic: %w", err)
					}
					mk, _, unwrapErr := s.Envelope.UnwrapFromMnemonic(mnemonic)
					if unwrapErr != nil {
						return fmt.Errorf("mnemonic does not match any envelope slot: %w", unwrapErr)
					}
					s.MasterKey = mk
					s.RecoveryMode = true
					// We still need a DBKey for SQLCipher below, but that will
					// be derived from the MK in phaseOpenDatabase.
					return nil
				}
			}

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
//  1. MasterKey already set (by mnemonic recovery) — no-op.
//  2. Envelope exists — derive KEK from passphrase and unwrap the MK.
//  3. No envelope and no legacy DB — treat as first run: create a new MK +
//     envelope and persist the envelope file immediately.
//  4. No envelope but legacy DB exists — mark LegacyMode; migration will handle it.
func phaseUnwrapOrCreateMK() Phase {
	return Phase{
		Name: "unwrap or create master key",
		Run: func(_ context.Context, s *State) error {
			if s.MasterKey != nil {
				// Already unwrapped via mnemonic recovery.
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
			client, rawDB, err := openDatabase(s.Options.DBPath, dbKey, rawKey, s.Options.DBEncryption.CipherPageSize)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			s.Client = client
			s.RawDB = rawDB
			s.Result.DBClient = client
			s.Result.RawDB = rawDB
			return nil
		},
		Cleanup: func(s *State) {
			if s.Client != nil {
				s.Client.Close()
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
				// Envelope path already resolved; legacy salt/checksum are no longer
				// authoritative. Skip to keep FirstRun semantics consistent.
				s.FirstRun = false
				return nil
			}
			salt, checksum, firstRun, err := loadSecurityState(s.RawDB)
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
				if err := storeSalt(s.RawDB, provider.Salt()); err != nil {
					return fmt.Errorf("store salt: %w", err)
				}
				cs := provider.CalculateChecksum(s.Passphrase, provider.Salt())
				if err := storeChecksum(s.RawDB, cs); err != nil {
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
			return nil
		},
	}
}
