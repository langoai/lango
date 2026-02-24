package security

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	sec "github.com/langoai/lango/internal/security"
)

func newKMSCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kms",
		Short: "Manage Cloud KMS / HSM integration",
	}

	cmd.AddCommand(newKMSStatusCmd(bootLoader))
	cmd.AddCommand(newKMSTestCmd(bootLoader))
	cmd.AddCommand(newKMSKeysCmd(bootLoader))

	return cmd
}

func newKMSStatusCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show KMS provider status",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.DBClient.Close()

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
				kmsProvider, provErr := sec.NewKMSProvider(provider, cfg.Security.KMS)
				if provErr != nil {
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

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(s)
			}

			fmt.Println("KMS Status")
			fmt.Printf("  Provider:      %s\n", s.Provider)
			fmt.Printf("  Key ID:        %s\n", s.KeyID)
			if s.Region != "" {
				fmt.Printf("  Region:        %s\n", s.Region)
			}
			fmt.Printf("  Fallback:      %s\n", s.Fallback)
			fmt.Printf("  Status:        %s\n", s.Status)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
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
			defer boot.DBClient.Close()

			cfg := boot.Config
			provider := cfg.Security.Signer.Provider
			if !isKMSProvider(provider) {
				return fmt.Errorf("current provider %q is not a KMS provider", provider)
			}

			kmsProvider, err := sec.NewKMSProvider(provider, cfg.Security.KMS)
			if err != nil {
				return fmt.Errorf("create KMS provider: %w", err)
			}

			ctx := context.Background()
			keyID := cfg.Security.KMS.KeyID

			// Generate random test data.
			testData := make([]byte, 32)
			if _, err := rand.Read(testData); err != nil {
				return fmt.Errorf("generate test data: %w", err)
			}

			fmt.Printf("Testing KMS roundtrip with key %q...\n", keyID)

			// Encrypt.
			ciphertext, err := kmsProvider.Encrypt(ctx, keyID, testData)
			if err != nil {
				return fmt.Errorf("encrypt: %w", err)
			}
			fmt.Printf("  Encrypt: OK (%d bytes → %d bytes)\n", len(testData), len(ciphertext))

			// Decrypt.
			plaintext, err := kmsProvider.Decrypt(ctx, keyID, ciphertext)
			if err != nil {
				return fmt.Errorf("decrypt: %w", err)
			}
			fmt.Printf("  Decrypt: OK (%d bytes)\n", len(plaintext))

			// Verify roundtrip.
			if len(plaintext) != len(testData) {
				return fmt.Errorf("roundtrip mismatch: got %d bytes, want %d", len(plaintext), len(testData))
			}
			for i := range testData {
				if plaintext[i] != testData[i] {
					return fmt.Errorf("roundtrip mismatch at byte %d", i)
				}
			}

			fmt.Println("  Roundtrip: PASS")
			return nil
		},
	}
}

func newKMSKeysCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "keys",
		Short: "List KMS keys registered in KeyRegistry",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.DBClient.Close()

			ctx := context.Background()
			registry := sec.NewKeyRegistry(boot.DBClient)
			keys, err := registry.ListKeys(ctx)
			if err != nil {
				return fmt.Errorf("list keys: %w", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(keys)
			}

			if len(keys) == 0 {
				fmt.Println("No keys registered.")
				return nil
			}

			fmt.Printf("%-36s  %-20s  %-12s  %-40s\n", "ID", "NAME", "TYPE", "REMOTE KEY ID")
			for _, k := range keys {
				fmt.Printf("%-36s  %-20s  %-12s  %-40s\n",
					k.ID.String(), k.Name, k.Type, k.RemoteKeyID)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func isKMSProvider(provider string) bool {
	switch provider {
	case "aws-kms", "gcp-kms", "azure-kv", "pkcs11":
		return true
	}
	return false
}
