package security

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
	"github.com/langoai/lango/internal/keyring"
	"github.com/langoai/lango/internal/p2p/identity"
	sec "github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/security/passphrase"
)

// envelopeSection captures passphrase-free envelope state for status output.
// Populated by reading envelope.json directly; never requires a passphrase.
type envelopeSection struct {
	Present          bool     `json:"present"`
	Version          int      `json:"version,omitempty"`
	SlotCount        int      `json:"slot_count,omitempty"`
	SlotTypes        []string `json:"slot_types,omitempty"`
	RecoverySetup    bool     `json:"recovery_setup"`
	PendingMigration bool     `json:"pending_migration,omitempty"`
	PendingRekey     bool     `json:"pending_rekey,omitempty"`
	KMSProtected     bool     `json:"kms_protected"`
	KMSProvider      string   `json:"kms_provider,omitempty"`
}

// dbStatusResult holds the DB-dependent fields populated by the non-interactive
// mini-bootstrap. Zero values indicate "unavailable" — the caller should not
// treat missing data as an error.
type dbStatusResult struct {
	available      bool
	encryptionKeys int
	storedSecrets  int
	config         *config.Config // non-nil when DB was opened and config loaded
}

// statusOutput is the full status payload (envelope + DB + config fields).
// identityBundleSection captures DID v2 identity bundle state.
type identityBundleSection struct {
	Present               bool   `json:"present"`
	DIDv2                 string `json:"did_v2,omitempty"`
	SigningAlgorithm      string `json:"signing_algorithm,omitempty"`
	HasSettlement         bool   `json:"has_settlement"`
	LegacyDID             string `json:"legacy_did,omitempty"`
	PQSigningKeyAvailable bool   `json:"pq_signing_key_available"`
	PQSigningAlgorithm    string `json:"pq_signing_algorithm,omitempty"`
}

type statusOutput struct {
	SignerProvider string                `json:"signer_provider"`
	EncryptionKeys int                   `json:"encryption_keys"`
	StoredSecrets  int                   `json:"stored_secrets"`
	Interceptor    string                `json:"interceptor"`
	PIIRedaction   string                `json:"pii_redaction"`
	ApprovalPolicy string                `json:"approval_policy"`
	DBEncryption   string                `json:"db_encryption"`
	Envelope       envelopeSection       `json:"envelope"`
	IdentityBundle identityBundleSection `json:"identity_bundle"`
	DBAvailable        bool                  `json:"db_available"`
	KMSProvider        string                `json:"kms_provider,omitempty"`
	KMSKeyID           string                `json:"kms_key_id,omitempty"`
	KMSFallback        string                `json:"kms_fallback,omitempty"`
	PQHandshakeEnabled bool                  `json:"pq_handshake_enabled"`
	PQHandshakeAlgo    string                `json:"pq_handshake_algorithm,omitempty"`
}

// readIdentityBundleStatus reads the identity bundle file from langoDir.
func readIdentityBundleStatus(langoDir string) identityBundleSection {
	if langoDir == "" {
		return identityBundleSection{}
	}
	bundle, err := identity.LoadBundleFile(langoDir)
	if err != nil || bundle == nil {
		return identityBundleSection{}
	}
	didV2, _ := identity.ComputeDIDv2(bundle)
	hasPQ := bundle.PQSigningKey != nil && len(bundle.PQSigningKey.PublicKey) > 0
	var pqAlgo string
	if hasPQ {
		pqAlgo = bundle.PQSigningKey.Algorithm
	}
	return identityBundleSection{
		Present:               true,
		DIDv2:                 didV2,
		SigningAlgorithm:      bundle.SigningKey.Algorithm,
		HasSettlement:         len(bundle.SettlementKey.PublicKey) > 0,
		LegacyDID:             bundle.LegacyDID,
		PQSigningKeyAvailable: hasPQ,
		PQSigningAlgorithm:    pqAlgo,
	}
}

// readEnvelopeStatus loads the envelope file from langoDir without requiring
// a passphrase. Returns a zero envelopeSection if the file is missing or corrupt —
// status output must never fail because of envelope state.
func readEnvelopeStatus(langoDir string) envelopeSection {
	if langoDir == "" {
		return envelopeSection{}
	}
	env, err := sec.LoadEnvelopeFile(langoDir)
	if err != nil || env == nil {
		return envelopeSection{}
	}
	types := make([]string, 0, env.SlotCount())
	seen := make(map[sec.KEKSlotType]bool)
	for _, slot := range env.Slots {
		if !seen[slot.Type] {
			types = append(types, string(slot.Type))
			seen[slot.Type] = true
		}
	}
	// Check for KMS KEK slot.
	kmsProtected := env.HasSlotType(sec.KEKSlotHardware)
	var kmsProvider string
	if kmsProtected {
		for _, slot := range env.Slots {
			if slot.Type == sec.KEKSlotHardware && slot.KMSProvider != "" {
				kmsProvider = slot.KMSProvider
				break
			}
		}
	}

	return envelopeSection{
		Present:          true,
		Version:          env.Version,
		SlotCount:        env.SlotCount(),
		SlotTypes:        types,
		RecoverySetup:    env.HasSlotType(sec.KEKSlotMnemonic),
		PendingMigration: env.PendingMigration,
		PendingRekey:     env.PendingRekey,
		KMSProtected:     kmsProtected,
		KMSProvider:      kmsProvider,
	}
}

// defaultLangoDir returns the default data directory (~/.lango) for the current user.
func defaultLangoDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".lango")
}

// expandPath expands a leading "~/" to the user's home directory.
func expandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		if h, err := os.UserHomeDir(); err == nil {
			return filepath.Join(h, p[2:])
		}
	}
	return p
}

// readDBStatusNonInteractive runs a minimal bootstrap-free DB read.
//
// Steps:
//  1. Acquire passphrase non-interactively (keyring → keyfile). If neither is
//     available, return a zero result — the caller renders "unavailable".
//  2. If an envelope is present, unwrap the MK and derive the raw DB key via
//     HKDF. Otherwise, fall back to the passphrase as the DB key (legacy path).
//  3. Open the DB via bootstrap.OpenDatabaseReadOnly — no schema migration,
//     no writes.
//  4. Read encryption key count and stored secret count.
//  5. Close the DB.
//
// This helper NEVER triggers an interactive prompt. Any failure (wrong
// passphrase, corrupt DB, schema drift) results in a zero result instead of
// an error, matching the spec's "graceful degrade" requirement.
func readDBStatusNonInteractive(
	langoDir, dbPath string,
	envelope *sec.MasterKeyEnvelope,
	needsKey bool,
) dbStatusResult {
	result := dbStatusResult{}
	if _, err := os.Stat(dbPath); err != nil {
		return result
	}

	var (
		dbKey      string
		rawKey     bool
		masterKey  []byte // non-nil when envelope unwrap succeeded
		usedKeyring bool  // true when passphrase came from keyring (stale fallback possible)
	)
	if needsKey {
		keyringProvider, _ := keyring.DetectSecureProvider()
		pass, source, err := passphrase.AcquireNonInteractive(passphrase.Options{
			KeyfilePath:     filepath.Join(langoDir, "keyfile"),
			KeyringProvider: keyringProvider,
		})
		if err != nil {
			if !errors.Is(err, passphrase.ErrNoNonInteractiveSource) {
				fmt.Fprintf(os.Stderr, "warning: status non-interactive passphrase: %v\n", err)
			}
			return result
		}

		usedKeyring = source == passphrase.SourceKeyring

		// retryWithKeyfile attempts keyfile-only acquisition when the first
		// passphrase (possibly from a stale keyring) fails to work.
		retryWithKeyfile := func() (string, bool) {
			if source != passphrase.SourceKeyring {
				return "", false // first attempt was already keyfile
			}
			kfPass, _, kfErr := passphrase.AcquireNonInteractive(passphrase.Options{
				KeyfilePath: filepath.Join(langoDir, "keyfile"),
			})
			if kfErr != nil {
				return "", false
			}
			return kfPass, true
		}

		if envelope != nil && !envelope.PendingMigration && !envelope.PendingRekey {
			mk, _, uerr := envelope.UnwrapFromPassphrase(pass)
			if uerr != nil {
				if fallback, ok := retryWithKeyfile(); ok {
					mk, _, uerr = envelope.UnwrapFromPassphrase(fallback)
				}
			}
			if uerr != nil {
				return result
			}
			masterKey = mk
			defer sec.ZeroBytes(masterKey)
			dbKey = sec.DeriveDBKeyHex(mk)
			rawKey = true
		} else {
			// Legacy mode OR migration in progress — use passphrase as DB key.
			dbKey = pass
			rawKey = false
		}
	}

	client, rawDB, err := bootstrap.OpenDatabaseReadOnly(dbPath, dbKey, rawKey, 0)
	if err != nil {
		// For legacy mode with stale keyring, retry with keyfile-only.
		if needsKey && !rawKey && usedKeyring {
			kfPass, _, kfErr := passphrase.AcquireNonInteractive(passphrase.Options{
				KeyfilePath: filepath.Join(langoDir, "keyfile"),
			})
			if kfErr == nil {
				client, rawDB, err = bootstrap.OpenDatabaseReadOnly(dbPath, kfPass, false, 0)
			}
		}
		if err != nil {
			return result
		}
	}
	defer client.Close()
	defer rawDB.Close()

	ctx := context.Background()
	registry := sec.NewKeyRegistry(client)
	if keys, err := registry.ListKeys(ctx); err == nil {
		result.encryptionKeys = len(keys)
	}
	if n, err := client.Secret.Query().Count(ctx); err == nil {
		result.storedSecrets = n
	}

	// Try to load the active config profile when MK is available.
	if masterKey != nil && envelope != nil {
		crypto := sec.NewLocalCryptoProvider()
		if initErr := crypto.InitializeWithEnvelope(masterKey, envelope); initErr == nil {
			store := configstore.NewStore(client, crypto)
			if _, cfg, _, loadErr := store.LoadActive(ctx); loadErr == nil {
				result.config = cfg
			}
			crypto.Close()
		}
	}

	result.available = true
	return result
}

// resolveStatusConfig loads the config without opening the encrypted DB.
// Returns a default config if loading fails, so the status command can still
// render configuration-derived fields (signer provider, interceptor, approval).
func resolveStatusConfig() *config.Config {
	// Config currently lives inside the encrypted profile store, which does
	// require bootstrap to read. For the status default path we keep things
	// simple: return DefaultConfig so the command never depends on decrypting.
	// Future work: surface a plaintext config snapshot, if one is useful.
	return config.DefaultConfig()
}

func newStatusCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool
	var fullBootstrap bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show security configuration status",
		Long: `Show security configuration status.

By default, the command runs in passphrase-free mode: it reads envelope.json
directly, attempts a non-interactive DB read via keyring/keyfile, and
gracefully degrades DB-dependent fields when no credential is available.

Use --full to force a full bootstrap (which may prompt for a passphrase in
interactive terminals).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if fullBootstrap {
				return runStatusFullBootstrap(bootLoader, jsonOutput)
			}
			return runStatusNonInteractive(jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&fullBootstrap, "full", false, "Run full bootstrap (may prompt for passphrase)")
	return cmd
}

// runStatusNonInteractive is the default status path.
// It NEVER triggers an interactive passphrase prompt.
func runStatusNonInteractive(jsonOutput bool) error {
	langoDir := defaultLangoDir()
	dbPath := filepath.Join(langoDir, "lango.db")

	envelope := readEnvelopeStatus(langoDir)

	var envPtr *sec.MasterKeyEnvelope
	if envelope.Present {
		envPtr, _ = sec.LoadEnvelopeFile(langoDir)
	}

	needsKey := bootstrap.IsDBEncrypted(dbPath)
	dbStatus := readDBStatusNonInteractive(langoDir, dbPath, envPtr, needsKey)

	// Use the active config if DB read succeeded; otherwise fall back to defaults.
	cfg := dbStatus.config
	if cfg == nil {
		cfg = resolveStatusConfig()
	}

	dbEncStatus := "disabled (plaintext)"
	if bootstrap.IsDBEncrypted(dbPath) {
		dbEncStatus = "encrypted (active)"
	} else if cfg.Security.DBEncryption.Enabled {
		dbEncStatus = "enabled (pending migration)"
	}

	policy := string(cfg.Security.Interceptor.ApprovalPolicy)
	if policy == "" {
		policy = "dangerous"
	}

	signer := cfg.Security.Signer.Provider
	if !dbStatus.available {
		signer = "unavailable"
	}

	s := statusOutput{
		SignerProvider: signer,
		EncryptionKeys: dbStatus.encryptionKeys,
		StoredSecrets:  dbStatus.storedSecrets,
		Interceptor:    boolToStatus(cfg.Security.Interceptor.Enabled),
		PIIRedaction:   boolToStatus(cfg.Security.Interceptor.RedactPII),
		ApprovalPolicy: policy,
		DBEncryption:   dbEncStatus,
		Envelope:           envelope,
		IdentityBundle:     readIdentityBundleStatus(langoDir),
		DBAvailable:        dbStatus.available,
		PQHandshakeEnabled: cfg.P2P.EnablePQHandshake,
		PQHandshakeAlgo:    pqAlgorithmLabel(cfg.P2P.EnablePQHandshake),
	}
	return renderStatus(s, jsonOutput)
}

// runStatusFullBootstrap is the --full path. It runs a full bootstrap (may
// prompt), reads decrypted config values, and surfaces KMS provider details.
func runStatusFullBootstrap(bootLoader func() (*bootstrap.Result, error), jsonOutput bool) error {
	boot, err := bootLoader()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	defer boot.DBClient.Close()

	cfg := boot.Config

	policy := string(cfg.Security.Interceptor.ApprovalPolicy)
	if policy == "" {
		policy = "dangerous"
	}

	dbEncStatus := "disabled (plaintext)"
	dbPath := expandPath(cfg.Session.DatabasePath)
	if bootstrap.IsDBEncrypted(dbPath) {
		dbEncStatus = "encrypted (active)"
	} else if cfg.Security.DBEncryption.Enabled {
		dbEncStatus = "enabled (pending migration)"
	}

	langoDir := boot.LangoDir
	if langoDir == "" {
		langoDir = defaultLangoDir()
	}

	s := statusOutput{
		SignerProvider: cfg.Security.Signer.Provider,
		Interceptor:    boolToStatus(cfg.Security.Interceptor.Enabled),
		PIIRedaction:   boolToStatus(cfg.Security.Interceptor.RedactPII),
		ApprovalPolicy: policy,
		DBEncryption:   dbEncStatus,
		Envelope:           readEnvelopeStatus(langoDir),
		IdentityBundle:     readIdentityBundleStatus(langoDir),
		DBAvailable:        true,
		PQHandshakeEnabled: cfg.P2P.EnablePQHandshake,
		PQHandshakeAlgo:    pqAlgorithmLabel(cfg.P2P.EnablePQHandshake),
	}

	if isKMSProvider(cfg.Security.Signer.Provider) {
		s.KMSProvider = cfg.Security.Signer.Provider
		s.KMSKeyID = cfg.Security.KMS.KeyID
		s.KMSFallback = boolToStatus(cfg.Security.KMS.FallbackToLocal)
	}

	ctx := context.Background()
	registry := sec.NewKeyRegistry(boot.DBClient)
	if keys, err := registry.ListKeys(ctx); err == nil {
		s.EncryptionKeys = len(keys)
	}
	if secrets, err := boot.DBClient.Secret.Query().Count(ctx); err == nil {
		s.StoredSecrets = secrets
	}

	return renderStatus(s, jsonOutput)
}

func renderStatus(s statusOutput, jsonOutput bool) error {
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(s)
	}

	signer := s.SignerProvider
	if signer == "" {
		signer = "unavailable"
	}

	fmt.Println("Security Status")
	fmt.Printf("  Signer Provider:    %s\n", signer)
	fmt.Printf("  Encryption Keys:    %d\n", s.EncryptionKeys)
	fmt.Printf("  Stored Secrets:     %d\n", s.StoredSecrets)
	fmt.Printf("  Interceptor:        %s\n", s.Interceptor)
	fmt.Printf("  PII Redaction:      %s\n", s.PIIRedaction)
	fmt.Printf("  Approval Policy:    %s\n", s.ApprovalPolicy)
	fmt.Printf("  DB Encryption:      %s\n", s.DBEncryption)
	if !s.DBAvailable {
		fmt.Println("  DB Access:          unavailable (no non-interactive credential)")
	}
	fmt.Println("  Master Key Envelope:")
	if s.Envelope.Present {
		fmt.Printf("    Version:          %d\n", s.Envelope.Version)
		fmt.Printf("    KEK Slots:        %d (%s)\n", s.Envelope.SlotCount, strings.Join(s.Envelope.SlotTypes, ", "))
		fmt.Printf("    Recovery Setup:   %s\n", boolToStatus(s.Envelope.RecoverySetup))
		if s.Envelope.KMSProtected {
			fmt.Printf("    KMS Protection:   enabled (%s)\n", s.Envelope.KMSProvider)
		} else {
			fmt.Println("    KMS Protection:   disabled")
		}
		if s.Envelope.PendingMigration {
			fmt.Println("    PendingMigration: TRUE (migration incomplete)")
		}
		if s.Envelope.PendingRekey {
			fmt.Println("    PendingRekey:     TRUE (PRAGMA rekey incomplete)")
		}
	} else {
		fmt.Println("    absent (legacy format)")
	}
	// Identity bundle section.
	fmt.Println("  Identity Bundle:")
	if s.IdentityBundle.Present {
		fmt.Printf("    DID v2:           %s\n", s.IdentityBundle.DIDv2)
		fmt.Printf("    Signing Key:      %s\n", s.IdentityBundle.SigningAlgorithm)
		fmt.Printf("    Settlement Key:   %s\n", boolToStatus(s.IdentityBundle.HasSettlement))
		fmt.Printf("    Legacy DID:       %s\n", s.IdentityBundle.LegacyDID)
		if s.IdentityBundle.PQSigningKeyAvailable {
			fmt.Printf("    PQ Signing Key:   available (%s)\n", s.IdentityBundle.PQSigningAlgorithm)
		} else {
			fmt.Println("    PQ Signing Key:   not available")
		}
	} else {
		fmt.Println("    absent (v1 identity only)")
	}
	// PQ handshake section.
	if s.PQHandshakeEnabled {
		fmt.Printf("  PQ Handshake:       enabled (%s)\n", s.PQHandshakeAlgo)
	} else {
		fmt.Println("  PQ Handshake:       disabled")
	}
	if s.KMSProvider != "" {
		fmt.Printf("  KMS Provider:       %s\n", s.KMSProvider)
		fmt.Printf("  KMS Key ID:         %s\n", s.KMSKeyID)
		fmt.Printf("  KMS Fallback:       %s\n", s.KMSFallback)
	}
	return nil
}

// pqAlgorithmLabel returns the algorithm label for PQ handshake status display.
func pqAlgorithmLabel(enabled bool) string {
	if enabled {
		return "X25519-MLKEM768"
	}
	return ""
}

func boolToStatus(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}
