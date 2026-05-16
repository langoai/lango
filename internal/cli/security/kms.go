package security

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	sec "github.com/langoai/lango/internal/security"
)

var newKMSProvider = sec.NewKMSProvider

func newKMSCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kms",
		Short: "Manage Cloud KMS / HSM integration",
	}

	cmd.AddCommand(newKMSStatusCmd(bootLoader))
	cmd.AddCommand(newKMSTestCmd(bootLoader))
	cmd.AddCommand(newKMSKeysCmd(bootLoader))
	cmd.AddCommand(newKMSWrapCmd(bootLoader))
	cmd.AddCommand(newKMSDetachCmd(bootLoader))

	return cmd
}

func newKMSStatusCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "status",
		Short:         "Show KMS provider status",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			cfg := boot.Config

			type kmsStatus struct {
				Provider string `json:"provider"`
				KeyID    string `json:"key_id"`
				Region   string `json:"region,omitempty"`
				Fallback string `json:"fallback"`
				Status   string `json:"status"`
			}

			provider := cfg.Security.Signer.Provider
			isKMS := isKMSProvider(provider)

			s := kmsStatus{
				Provider: provider,
				KeyID:    cfg.Security.KMS.KeyID,
				Region:   cfg.Security.KMS.Region,
				Fallback: boolToStatus(cfg.Security.KMS.FallbackToLocal),
				Status:   "not configured",
			}

			if isKMS {
				// Try to create the provider to check connectivity.
				kmsProvider, provErr := newKMSProvider(sec.KMSProviderName(provider), cfg.Security.KMS)
				if provErr != nil { //nolint:staticcheck // stubs always error; real impls use kms_* build tags
					s.Status = fmt.Sprintf("error: %v", provErr)
				} else {
					checker := sec.NewKMSHealthChecker(kmsProvider, cfg.Security.KMS.KeyID, 0)
					if checker.IsConnected() {
						s.Status = "connected"
					} else {
						s.Status = "unreachable"
					}
				}
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), s)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "KMS Status")
			fmt.Fprintf(cmd.OutOrStdout(), "  Provider:      %s\n", s.Provider)
			fmt.Fprintf(cmd.OutOrStdout(), "  Key ID:        %s\n", s.KeyID)
			if s.Region != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  Region:        %s\n", s.Region)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  Fallback:      %s\n", s.Fallback)
			fmt.Fprintf(cmd.OutOrStdout(), "  Status:        %s\n", s.Status)

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}

func newKMSTestCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test KMS encrypt/decrypt roundtrip",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			cfg := boot.Config
			provider := cfg.Security.Signer.Provider
			if !isKMSProvider(provider) {
				return fmt.Errorf("current provider %q is not a KMS provider", provider)
			}

			kmsProvider, err := newKMSProvider(sec.KMSProviderName(provider), cfg.Security.KMS)
			if err != nil { //nolint:staticcheck // stubs always error; real impls use kms_* build tags
				return fmt.Errorf("create KMS provider: %w", err)
			}

			ctx := context.Background()
			keyID := cfg.Security.KMS.KeyID

			// Generate random test data.
			testData := make([]byte, 32)
			if _, err := rand.Read(testData); err != nil {
				return fmt.Errorf("generate test data: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Testing KMS roundtrip with key %q...\n", keyID)

			// Encrypt.
			ciphertext, err := kmsProvider.Encrypt(ctx, keyID, testData)
			if err != nil {
				return fmt.Errorf("encrypt: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  Encrypt: OK (%d bytes → %d bytes)\n", len(testData), len(ciphertext))

			// Decrypt.
			plaintext, err := kmsProvider.Decrypt(ctx, keyID, ciphertext)
			if err != nil {
				return fmt.Errorf("decrypt: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  Decrypt: OK (%d bytes)\n", len(plaintext))

			// Verify roundtrip.
			if len(plaintext) != len(testData) {
				return fmt.Errorf("roundtrip mismatch: got %d bytes, want %d", len(plaintext), len(testData))
			}
			for i := range testData {
				if plaintext[i] != testData[i] {
					return fmt.Errorf("roundtrip mismatch at byte %d", i)
				}
			}

			fmt.Fprintln(cmd.OutOrStdout(), "  Roundtrip: PASS")
			return nil
		},
	}
}

func newKMSKeysCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "keys",
		Short:         "List KMS keys registered in KeyRegistry",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			ctx := context.Background()
			var registry *sec.KeyRegistry
			if boot.Storage != nil {
				registry = boot.Storage.KeyRegistry()
			}
			if registry == nil {
				return fmt.Errorf("key registry unavailable")
			}
			keys, err := registry.ListKeys(ctx)
			if err != nil {
				return fmt.Errorf("list keys: %w", err)
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), keys)
			}

			if len(keys) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No keys registered.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-36s  %-20s  %-12s  %-40s\n", "ID", "NAME", "TYPE", "REMOTE KEY ID")
			for _, k := range keys {
				fmt.Fprintf(cmd.OutOrStdout(), "%-36s  %-20s  %-12s  %-40s\n",
					k.ID.String(), k.Name, k.Type, k.RemoteKeyID)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}

func isKMSProvider(provider string) bool {
	return sec.KMSProviderName(provider).Valid()
}

func newKMSWrapCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		provider string
		keyID    string
	)

	cmd := &cobra.Command{
		Use:   "wrap",
		Short: "Add a KMS KEK slot to protect the Master Key",
		Long: `Wraps the Master Key with a KMS key and adds a KMS slot to the envelope.
This enables passphraseless bootstrap when KMS credentials are available.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if provider == "" || keyID == "" {
				return fmt.Errorf("--provider and --key-id are required")
			}
			if !sec.KMSProviderName(provider).Valid() {
				return fmt.Errorf("unknown KMS provider %q (supported: aws-kms, gcp-kms, azure-kv, pkcs11)", provider)
			}

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			crypto, ok := boot.Crypto.(*sec.LocalCryptoProvider)
			if !ok || crypto == nil {
				return fmt.Errorf("local crypto provider not available")
			}
			envelope := crypto.Envelope()
			if envelope == nil {
				return fmt.Errorf("envelope not available (legacy mode)")
			}

			// Unwrap MK from existing passphrase slot to wrap with KMS.
			// The MK is already in memory via bootstrap.
			mk := crypto.MasterKey()
			if mk == nil {
				return fmt.Errorf("master key not available")
			}

			// Create the KMS provider.
			kmsProvider, err := newKMSProvider(sec.KMSProviderName(provider), boot.Config.Security.KMS)
			if err != nil { //nolint:staticcheck // stubs always error; real impls use kms_* build tags
				return fmt.Errorf("create KMS provider: %w", err)
			}

			ctx := context.Background()
			if err := envelope.AddKMSSlot(ctx, "kms", mk, kmsProvider, provider, keyID); err != nil {
				return fmt.Errorf("add KMS slot: %w", err)
			}

			if err := sec.StoreEnvelopeFile(boot.LangoDir, envelope); err != nil {
				return fmt.Errorf("persist envelope: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "KMS slot added (provider=%s, keyID=%s)\n", provider, keyID)
			fmt.Fprintln(cmd.OutOrStdout(), "Next bootstrap can use KMS for passphraseless unlock.")
			return nil
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "", "KMS provider name (aws-kms, gcp-kms, azure-kv, pkcs11)")
	cmd.Flags().StringVar(&keyID, "key-id", "", "KMS key identifier")

	return cmd
}

func newKMSDetachCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var slotID string

	cmd := &cobra.Command{
		Use:   "detach",
		Short: "Remove a KMS KEK slot from the envelope",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.Close()

			crypto, ok := boot.Crypto.(*sec.LocalCryptoProvider)
			if !ok || crypto == nil {
				return fmt.Errorf("local crypto provider not available")
			}
			envelope := crypto.Envelope()
			if envelope == nil {
				return fmt.Errorf("envelope not available (legacy mode)")
			}

			// Find KMS slots.
			var kmsSlots []sec.KEKSlot
			for _, s := range envelope.Slots {
				if s.Type == sec.KEKSlotHardware {
					kmsSlots = append(kmsSlots, s)
				}
			}

			if len(kmsSlots) == 0 {
				return fmt.Errorf("no KMS slots found in envelope")
			}

			// Determine which slot to remove.
			var targetID string
			if len(kmsSlots) == 1 {
				targetID = kmsSlots[0].ID
			} else if slotID != "" {
				targetID = slotID
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "Multiple KMS slots found. Specify --slot-id:")
				for _, s := range kmsSlots {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s  provider=%s  keyID=%s  label=%s\n",
						s.ID, s.KMSProvider, s.KMSKeyID, s.Label)
				}
				return fmt.Errorf("--slot-id required when multiple KMS slots exist")
			}

			// Ensure at least one non-KMS slot remains.
			nonKMSCount := 0
			for _, s := range envelope.Slots {
				if s.Type != sec.KEKSlotHardware {
					nonKMSCount++
				}
			}
			if nonKMSCount == 0 {
				return fmt.Errorf("cannot remove KMS slot: at least one passphrase or mnemonic slot must remain")
			}

			if err := envelope.RemoveSlot(targetID); err != nil {
				return fmt.Errorf("remove slot: %w", err)
			}

			if err := sec.StoreEnvelopeFile(boot.LangoDir, envelope); err != nil {
				return fmt.Errorf("persist envelope: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "KMS slot %s removed.\n", targetID)
			return nil
		},
	}

	cmd.Flags().StringVar(&slotID, "slot-id", "", "UUID of the KMS slot to remove (required when multiple exist)")

	return cmd
}
