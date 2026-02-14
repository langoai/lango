package security

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/langowarny/lango/internal/config"
	sec "github.com/langowarny/lango/internal/security"
	"github.com/langowarny/lango/internal/session"
	"github.com/spf13/cobra"
)

func newStatusCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show security configuration status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

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

			// Try to read key/secret counts from DB
			if cfg.Session.DatabasePath != "" {
				store, err := session.NewEntStore(cfg.Session.DatabasePath)
				if err == nil {
					defer store.Close()

					ctx := context.Background()
					registry := sec.NewKeyRegistry(store.Client())
					keys, err := registry.ListKeys(ctx)
					if err == nil {
						s.EncryptionKeys = len(keys)
					}

					// Count secrets (just query count, no crypto needed)
					secrets, err := store.Client().Secret.Query().Count(ctx)
					if err == nil {
						s.StoredSecrets = secrets
					}
				}
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
