package security

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/keyring"
	"github.com/langoai/lango/internal/p2p/identity"
	sec "github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/security/passphrase"
	"github.com/langoai/lango/internal/storagebroker"
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
	SignerProvider       string                `json:"signer_provider"`
	EncryptionKeys       int                   `json:"encryption_keys"`
	StoredSecrets        int                   `json:"stored_secrets"`
	Interceptor          string                `json:"interceptor"`
	PIIRedaction         string                `json:"pii_redaction"`
	ApprovalPolicy       string                `json:"approval_policy"`
	ExportabilityEnabled bool                  `json:"exportability_enabled"`
	DBEncryption         string                `json:"db_encryption"`
	Envelope             envelopeSection       `json:"envelope"`
	IdentityBundle       identityBundleSection `json:"identity_bundle"`
	DBAvailable          bool                  `json:"db_available"`
	KMSProvider          string                `json:"kms_provider,omitempty"`
	KMSKeyID             string                `json:"kms_key_id,omitempty"`
	KMSFallback          string                `json:"kms_fallback,omitempty"`
	PQHandshakeEnabled   bool                  `json:"pq_handshake_enabled"`
	PQHandshakeAlgo      string                `json:"pq_handshake_algorithm,omitempty"`
}

var (
	acquireNonInteractivePassphrase           = passphrase.AcquireNonInteractive
	statusErrWriter                 io.Writer = os.Stderr
)

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
//  3. Start a transient storage broker client and request a read-only DB status
//     summary. The CLI process does not open the database directly.
//  4. Read encryption key count and stored secret count from the broker result.
//  5. Close the broker client.
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
		dbKey       string
		rawKey      bool
		masterKey   []byte // non-nil when envelope unwrap succeeded
		usedKeyring bool   // true when passphrase came from keyring (stale fallback possible)
	)
	if needsKey {
		keyringProvider, _ := keyring.DetectSecureProvider()
		pass, source, err := acquireNonInteractivePassphrase(passphrase.Options{
			KeyfilePath:     filepath.Join(langoDir, "keyfile"),
			KeyringProvider: keyringProvider,
		})
		if err != nil {
			if !errors.Is(err, passphrase.ErrNoNonInteractiveSource) {
				fmt.Fprintf(statusErrWriter, "warning: status non-interactive passphrase: %v\n", err)
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
			kfPass, _, kfErr := acquireNonInteractivePassphrase(passphrase.Options{
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

	brokerClient, err := storagebroker.Start(context.Background())
	if err != nil {
		return result
	}
	defer func() { _ = brokerClient.Close(context.Background()) }()

	summary, err := brokerClient.DBStatusSummary(context.Background(), storagebroker.DBStatusSummaryRequest{
		DBPath:         dbPath,
		EncryptionKey:  dbKey,
		RawKey:         rawKey,
		CipherPageSize: 0,
	})
	if err != nil {
		// For legacy mode with stale keyring, retry with keyfile-only.
		if needsKey && !rawKey && usedKeyring {
			kfPass, _, kfErr := acquireNonInteractivePassphrase(passphrase.Options{
				KeyfilePath: filepath.Join(langoDir, "keyfile"),
			})
			if kfErr == nil {
				summary, err = brokerClient.DBStatusSummary(context.Background(), storagebroker.DBStatusSummaryRequest{
					DBPath:         dbPath,
					EncryptionKey:  kfPass,
					RawKey:         false,
					CipherPageSize: 0,
				})
			}
		}
		if err != nil {
			return result
		}
	}

	_ = masterKey
	_ = envelope
	result.available = summary.Available
	result.encryptionKeys = summary.EncryptionKeys
	result.storedSecrets = summary.StoredSecrets
	if cfg, ok := loadActiveStatusConfig(brokerClient); ok {
		result.config = cfg
	}
	return result
}

type activeConfigLoader interface {
	ConfigLoadActive(context.Context) (storagebroker.ConfigLoadActiveResult, error)
}

func loadActiveStatusConfig(brokerClient activeConfigLoader) (*config.Config, bool) {
	if brokerClient == nil {
		return nil, false
	}
	cfgResult, err := brokerClient.ConfigLoadActive(context.Background())
	if err != nil {
		return nil, false
	}
	var cfg config.Config
	if err := json.Unmarshal(cfgResult.Config, &cfg); err != nil {
		return nil, false
	}
	if err := config.PostLoad(&cfg); err != nil {
		return nil, false
	}
	return &cfg, true
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
	var output string
	var fullBootstrap bool

	cmd := &cobra.Command{
		Use:           "status",
		Short:         "Show security configuration status",
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `Show security configuration status.

By default, the command runs in passphrase-free mode: it reads envelope.json
directly, attempts a non-interactive DB read via keyring/keyfile, and
gracefully degrades DB-dependent fields when no credential is available.

Use --full to force a full bootstrap (which may prompt for a passphrase in
		interactive terminals).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}
			if fullBootstrap {
				return runStatusFullBootstrap(cmd.OutOrStdout(), bootLoader, output)
			}
			return runStatusNonInteractive(cmd.OutOrStdout(), output)
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	cmd.Flags().BoolVar(&fullBootstrap, "full", false, "Run full bootstrap (may prompt for passphrase)")
	return cmd
}

// runStatusNonInteractive is the default status path.
// It NEVER triggers an interactive passphrase prompt.
func runStatusNonInteractive(writer io.Writer, output string) error {
	langoDir := defaultLangoDir()
	dbPath := filepath.Join(langoDir, "lango.db")

	envelope := readEnvelopeStatus(langoDir)

	var envPtr *sec.MasterKeyEnvelope
	if envelope.Present {
		envPtr, _ = sec.LoadEnvelopeFile(langoDir)
	}

	needsKey := false
	dbStatus := readDBStatusNonInteractive(langoDir, dbPath, envPtr, needsKey)

	// Use the active config if DB read succeeded; otherwise fall back to defaults.
	cfg := dbStatus.config
	if cfg == nil {
		cfg = resolveStatusConfig()
	}

	dbEncStatus := "disabled (plaintext)"
	if bootstrap.IsDBEncrypted(dbPath) {
		dbEncStatus = "legacy encrypted or unreadable DB (unsupported)"
	} else if cfg.Security.DBEncryption.Enabled {
		dbEncStatus = "deprecated config (ignored)"
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
		SignerProvider:       signer,
		EncryptionKeys:       dbStatus.encryptionKeys,
		StoredSecrets:        dbStatus.storedSecrets,
		Interceptor:          boolToStatus(cfg.Security.Interceptor.Enabled),
		PIIRedaction:         boolToStatus(cfg.Security.Interceptor.RedactPII),
		ApprovalPolicy:       policy,
		ExportabilityEnabled: cfg.Security.Exportability.Enabled,
		DBEncryption:         dbEncStatus,
		Envelope:             envelope,
		IdentityBundle:       readIdentityBundleStatus(langoDir),
		DBAvailable:          dbStatus.available,
		PQHandshakeEnabled:   cfg.P2P.EnablePQHandshake,
		PQHandshakeAlgo:      pqAlgorithmLabel(cfg.P2P.EnablePQHandshake),
	}
	return renderStatus(writer, s, output)
}

// runStatusFullBootstrap is the --full path. It runs a full bootstrap (may
// prompt), reads decrypted config values, and surfaces KMS provider details.
func runStatusFullBootstrap(writer io.Writer, bootLoader func() (*bootstrap.Result, error), output string) error {
	boot, err := bootLoader()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	defer boot.Close()

	cfg := boot.Config

	policy := string(cfg.Security.Interceptor.ApprovalPolicy)
	if policy == "" {
		policy = "dangerous"
	}

	dbEncStatus := "disabled (plaintext)"
	dbPath := expandPath(cfg.Session.DatabasePath)
	if bootstrap.IsDBEncrypted(dbPath) {
		dbEncStatus = "legacy encrypted or unreadable DB (unsupported)"
	} else if cfg.Security.DBEncryption.Enabled {
		dbEncStatus = "deprecated config (ignored)"
	}

	langoDir := boot.LangoDir
	if langoDir == "" {
		langoDir = defaultLangoDir()
	}

	s := statusOutput{
		SignerProvider:       cfg.Security.Signer.Provider,
		Interceptor:          boolToStatus(cfg.Security.Interceptor.Enabled),
		PIIRedaction:         boolToStatus(cfg.Security.Interceptor.RedactPII),
		ApprovalPolicy:       policy,
		ExportabilityEnabled: cfg.Security.Exportability.Enabled,
		DBEncryption:         dbEncStatus,
		Envelope:             readEnvelopeStatus(langoDir),
		IdentityBundle:       readIdentityBundleStatus(langoDir),
		DBAvailable:          true,
		PQHandshakeEnabled:   cfg.P2P.EnablePQHandshake,
		PQHandshakeAlgo:      pqAlgorithmLabel(cfg.P2P.EnablePQHandshake),
	}

	if isKMSProvider(cfg.Security.Signer.Provider) {
		s.KMSProvider = cfg.Security.Signer.Provider
		s.KMSKeyID = cfg.Security.KMS.KeyID
		s.KMSFallback = boolToStatus(cfg.Security.KMS.FallbackToLocal)
	}

	ctx := context.Background()
	if boot.Storage != nil {
		if summary, err := boot.Storage.SecuritySummary(ctx); err == nil {
			s.EncryptionKeys = summary.EncryptionKeys
			s.StoredSecrets = summary.StoredSecrets
		}
	}

	return renderStatus(writer, s, output)
}

func renderStatus(writer io.Writer, s statusOutput, output string) error {
	if output == "json" {
		return printJSON(writer, s)
	}

	signer := s.SignerProvider
	if signer == "" {
		signer = "unavailable"
	}

	fmt.Fprintln(writer, "Security Status")
	fmt.Fprintf(writer, "  Signer Provider:    %s\n", signer)
	fmt.Fprintf(writer, "  Encryption Keys:    %d\n", s.EncryptionKeys)
	fmt.Fprintf(writer, "  Stored Secrets:     %d\n", s.StoredSecrets)
	fmt.Fprintf(writer, "  Interceptor:        %s\n", s.Interceptor)
	fmt.Fprintf(writer, "  PII Redaction:      %s\n", s.PIIRedaction)
	fmt.Fprintf(writer, "  Approval Policy:    %s\n", s.ApprovalPolicy)
	fmt.Fprintf(writer, "  Exportability:      %s\n", boolToStatus(s.ExportabilityEnabled))
	fmt.Fprintf(writer, "  DB Encryption:      %s\n", s.DBEncryption)
	if !s.DBAvailable {
		fmt.Fprintln(writer, "  DB Access:          unavailable (no non-interactive credential)")
	}
	fmt.Fprintln(writer, "  Master Key Envelope:")
	if s.Envelope.Present {
		fmt.Fprintf(writer, "    Version:          %d\n", s.Envelope.Version)
		fmt.Fprintf(writer, "    KEK Slots:        %d (%s)\n", s.Envelope.SlotCount, strings.Join(s.Envelope.SlotTypes, ", "))
		fmt.Fprintf(writer, "    Recovery Setup:   %s\n", boolToStatus(s.Envelope.RecoverySetup))
		if s.Envelope.KMSProtected {
			fmt.Fprintf(writer, "    KMS Protection:   enabled (%s)\n", s.Envelope.KMSProvider)
		} else {
			fmt.Fprintln(writer, "    KMS Protection:   disabled")
		}
		if s.Envelope.PendingMigration {
			fmt.Fprintln(writer, "    PendingMigration: TRUE (migration incomplete)")
		}
		if s.Envelope.PendingRekey {
			fmt.Fprintln(writer, "    PendingRekey:     TRUE (PRAGMA rekey incomplete)")
		}
	} else {
		fmt.Fprintln(writer, "    absent (legacy format)")
	}
	// Identity bundle section.
	fmt.Fprintln(writer, "  Identity Bundle:")
	if s.IdentityBundle.Present {
		fmt.Fprintf(writer, "    DID v2:           %s\n", s.IdentityBundle.DIDv2)
		fmt.Fprintf(writer, "    Signing Key:      %s\n", s.IdentityBundle.SigningAlgorithm)
		fmt.Fprintf(writer, "    Settlement Key:   %s\n", boolToStatus(s.IdentityBundle.HasSettlement))
		fmt.Fprintf(writer, "    Legacy DID:       %s\n", s.IdentityBundle.LegacyDID)
		if s.IdentityBundle.PQSigningKeyAvailable {
			fmt.Fprintf(writer, "    PQ Signing Key:   available (%s)\n", s.IdentityBundle.PQSigningAlgorithm)
		} else {
			fmt.Fprintln(writer, "    PQ Signing Key:   not available")
		}
	} else {
		fmt.Fprintln(writer, "    absent (v1 identity only)")
	}
	// PQ handshake section.
	if s.PQHandshakeEnabled {
		fmt.Fprintf(writer, "  PQ Handshake:       enabled (%s)\n", s.PQHandshakeAlgo)
	} else {
		fmt.Fprintln(writer, "  PQ Handshake:       disabled")
	}
	if s.KMSProvider != "" {
		fmt.Fprintf(writer, "  KMS Provider:       %s\n", s.KMSProvider)
		fmt.Fprintf(writer, "  KMS Key ID:         %s\n", s.KMSKeyID)
		fmt.Fprintf(writer, "  KMS Fallback:       %s\n", s.KMSFallback)
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
