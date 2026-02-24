package security

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	sec "github.com/langoai/lango/internal/security"
)

func newStatusCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show security configuration status",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.DBClient.Close()

			cfg := boot.Config

			type statusOutput struct {
				SignerProvider string `json:"signer_provider"`
				EncryptionKeys int    `json:"encryption_keys"`
				StoredSecrets  int    `json:"stored_secrets"`
				Interceptor    string `json:"interceptor"`
				PIIRedaction   string `json:"pii_redaction"`
				ApprovalPolicy string `json:"approval_policy"`
				DBEncryption   string `json:"db_encryption"`
				KMSProvider    string `json:"kms_provider,omitempty"`
				KMSKeyID       string `json:"kms_key_id,omitempty"`
				KMSFallback    string `json:"kms_fallback,omitempty"`
			}

			policy := string(cfg.Security.Interceptor.ApprovalPolicy)
			if policy == "" {
				policy = "dangerous"
			}

			// Determine DB encryption status.
			dbEncStatus := "disabled (plaintext)"
			dbPath := cfg.Session.DatabasePath
			if strings.HasPrefix(dbPath, "~/") {
				if h, err := os.UserHomeDir(); err == nil {
					dbPath = filepath.Join(h, dbPath[2:])
				}
			}
			if bootstrap.IsDBEncrypted(dbPath) {
				dbEncStatus = "encrypted (active)"
			} else if cfg.Security.DBEncryption.Enabled {
				dbEncStatus = "enabled (pending migration)"
			}

			s := statusOutput{
				SignerProvider: cfg.Security.Signer.Provider,
				Interceptor:    boolToStatus(cfg.Security.Interceptor.Enabled),
				PIIRedaction:   boolToStatus(cfg.Security.Interceptor.RedactPII),
				ApprovalPolicy: policy,
				DBEncryption:   dbEncStatus,
			}

			// Populate KMS fields when a KMS provider is configured.
			if isKMSProvider(cfg.Security.Signer.Provider) {
				s.KMSProvider = cfg.Security.Signer.Provider
				s.KMSKeyID = cfg.Security.KMS.KeyID
				s.KMSFallback = boolToStatus(cfg.Security.KMS.FallbackToLocal)
			}

			ctx := context.Background()
			registry := sec.NewKeyRegistry(boot.DBClient)
			keys, err := registry.ListKeys(ctx)
			if err == nil {
				s.EncryptionKeys = len(keys)
			}

			secrets, err := boot.DBClient.Secret.Query().Count(ctx)
			if err == nil {
				s.StoredSecrets = secrets
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(s)
			}

			fmt.Println("Security Status")
			fmt.Printf("  Signer Provider:    %s\n", s.SignerProvider)
			fmt.Printf("  Encryption Keys:    %d\n", s.EncryptionKeys)
			fmt.Printf("  Stored Secrets:     %d\n", s.StoredSecrets)
			fmt.Printf("  Interceptor:        %s\n", s.Interceptor)
			fmt.Printf("  PII Redaction:      %s\n", s.PIIRedaction)
			fmt.Printf("  Approval Policy:    %s\n", s.ApprovalPolicy)
			fmt.Printf("  DB Encryption:      %s\n", s.DBEncryption)
			if s.KMSProvider != "" {
				fmt.Printf("  KMS Provider:       %s\n", s.KMSProvider)
				fmt.Printf("  KMS Key ID:         %s\n", s.KMSKeyID)
				fmt.Printf("  KMS Fallback:       %s\n", s.KMSFallback)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func boolToStatus(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}
