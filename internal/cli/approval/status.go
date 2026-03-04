package approval

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

func newStatusCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show approval providers and policy configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			ic := cfg.Security.Interceptor

			type statusOutput struct {
				InterceptorEnabled  bool     `json:"interceptor_enabled"`
				ApprovalPolicy      string   `json:"approval_policy"`
				HeadlessAutoApprove bool     `json:"headless_auto_approve"`
				ApprovalTimeoutSec  int      `json:"approval_timeout_sec"`
				NotifyChannel       string   `json:"notify_channel,omitempty"`
				SensitiveTools      []string `json:"sensitive_tools,omitempty"`
				ExemptTools         []string `json:"exempt_tools,omitempty"`
				RedactPII           bool     `json:"redact_pii"`
			}

			out := statusOutput{
				InterceptorEnabled:  ic.Enabled,
				ApprovalPolicy:      string(ic.ApprovalPolicy),
				HeadlessAutoApprove: ic.HeadlessAutoApprove,
				ApprovalTimeoutSec:  ic.ApprovalTimeoutSec,
				NotifyChannel:       ic.NotifyChannel,
				SensitiveTools:      ic.SensitiveTools,
				ExemptTools:         ic.ExemptTools,
				RedactPII:           ic.RedactPII,
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			fmt.Printf("Approval Status\n")
			fmt.Printf("  Interceptor Enabled:   %v\n", out.InterceptorEnabled)
			fmt.Printf("  Approval Policy:       %s\n", out.ApprovalPolicy)
			fmt.Printf("  Headless Auto-Approve: %v\n", out.HeadlessAutoApprove)
			fmt.Printf("  Approval Timeout:      %d sec\n", out.ApprovalTimeoutSec)
			fmt.Printf("  Redact PII:            %v\n", out.RedactPII)
			if out.NotifyChannel != "" {
				fmt.Printf("  Notify Channel:        %s\n", out.NotifyChannel)
			}
			fmt.Println()

			if len(out.SensitiveTools) > 0 {
				fmt.Printf("Sensitive Tools (%d)\n", len(out.SensitiveTools))
				for _, t := range out.SensitiveTools {
					fmt.Printf("  - %s\n", t)
				}
				fmt.Println()
			}

			if len(out.ExemptTools) > 0 {
				fmt.Printf("Exempt Tools (%d)\n", len(out.ExemptTools))
				for _, t := range out.ExemptTools {
					fmt.Printf("  - %s\n", t)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
