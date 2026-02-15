package security

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langowarny/lango/internal/bootstrap"
	sec "github.com/langowarny/lango/internal/security"
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
				SignerProvider  string `json:"signer_provider"`
				EncryptionKeys int    `json:"encryption_keys"`
				StoredSecrets  int    `json:"stored_secrets"`
				Interceptor    string `json:"interceptor"`
				PIIRedaction   string `json:"pii_redaction"`
				ApprovalReq    string `json:"approval_required"`
			}

			s := statusOutput{
				SignerProvider: cfg.Security.Signer.Provider,
				Interceptor:   boolToStatus(cfg.Security.Interceptor.Enabled),
				PIIRedaction:  boolToStatus(cfg.Security.Interceptor.RedactPII),
				ApprovalReq:   boolToStatus(cfg.Security.Interceptor.ApprovalRequired),
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
			fmt.Printf("  Approval Required:  %s\n", s.ApprovalReq)

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
