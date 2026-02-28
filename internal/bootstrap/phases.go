package bootstrap

import (
	"context"
	"crypto/hmac"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/langoai/lango/internal/cli/prompt"
	"github.com/langoai/lango/internal/configstore"
	"github.com/langoai/lango/internal/keyring"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/security/passphrase"
)

// DefaultPhases returns the standard bootstrap phase sequence.
func DefaultPhases() []Phase {
	return []Phase{
		phaseEnsureDataDir(),
		phaseDetectEncryption(),
		phaseAcquirePassphrase(),
		phaseOpenDatabase(),
		phaseLoadSecurityState(),
		phaseInitCrypto(),
		phaseLoadProfile(),
	}
}

// phaseEnsureDataDir creates ~/.lango/ directory and resolves default paths.
func phaseEnsureDataDir() Phase {
	return Phase{
		Name: "ensure data directory",
		Run: func(_ context.Context, s *State) error {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("resolve home directory: %w", err)
			}
			s.Home = home
			s.LangoDir = filepath.Join(home, ".lango")

			if s.Options.DBPath == "" {
				s.Options.DBPath = filepath.Join(s.LangoDir, "lango.db")
			}
			if s.Options.KeyfilePath == "" {
				s.Options.KeyfilePath = filepath.Join(s.LangoDir, "keyfile")
			}

			if err := os.MkdirAll(s.LangoDir, 0700); err != nil {
				return fmt.Errorf("create data directory: %w", err)
			}
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

// phaseAcquirePassphrase acquires the passphrase from keyring, keyfile, or interactive prompt.
// Also offers to store the passphrase when secure hardware is available.
func phaseAcquirePassphrase() Phase {
	return Phase{
		Name: "acquire passphrase",
		Run: func(_ context.Context, s *State) error {
			// Detect secure provider (biometric/TPM).
			if !s.Options.SkipSecureDetection {
				s.SecureProvider, s.SecurityTier = keyring.DetectSecureProvider()
			}

			// Determine if this is a first-run scenario: no DB file.
			_, statErr := os.Stat(s.Options.DBPath)
			s.FirstRunGuess = statErr != nil

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

// phaseOpenDatabase opens SQLite/SQLCipher DB and runs ent schema migration.
func phaseOpenDatabase() Phase {
	return Phase{
		Name: "open database",
		Run: func(_ context.Context, s *State) error {
			if s.NeedsDBKey {
				s.DBKey = s.Passphrase
			}
			client, rawDB, err := openDatabase(s.Options.DBPath, s.DBKey, s.Options.DBEncryption.CipherPageSize)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			s.Client = client
			s.RawDB = rawDB
			// Populate Result early so later phases can build on it.
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

// phaseLoadSecurityState reads salt and checksum from the database.
func phaseLoadSecurityState() Phase {
	return Phase{
		Name: "load security state",
		Run: func(_ context.Context, s *State) error {
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

// phaseInitCrypto initializes the crypto provider and shreds keyfile if needed.
func phaseInitCrypto() Phase {
	return Phase{
		Name: "initialize crypto",
		Run: func(_ context.Context, s *State) error {
			provider := security.NewLocalCryptoProvider()

			if s.FirstRun {
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
			} else {
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

// phaseLoadProfile loads or creates the configuration profile.
func phaseLoadProfile() Phase {
	return Phase{
		Name: "load profile",
		Run: func(ctx context.Context, s *State) error {
			store := configstore.NewStore(s.Client, s.Crypto)
			s.Result.ConfigStore = store

			profileName := s.Options.ForceProfile

			if profileName != "" {
				cfg, err := store.Load(ctx, profileName)
				if err != nil {
					return fmt.Errorf("load profile %q: %w", profileName, err)
				}
				s.Result.Config = cfg
				s.Result.ProfileName = profileName
				return nil
			}

			name, cfg, err := store.LoadActive(ctx)
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
				return nil
			}

			s.Result.Config = cfg
			s.Result.ProfileName = name
			return nil
		},
	}
}
