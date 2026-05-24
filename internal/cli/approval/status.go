package approval

import (
	"fmt"

	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

func newStatusCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "status",
		Short:         "Show approval providers and policy configuration",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}
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

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), out)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Approval Status\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Interceptor Enabled:   %v\n", out.InterceptorEnabled)
			fmt.Fprintf(cmd.OutOrStdout(), "  Approval Policy:       %s\n", out.ApprovalPolicy)
			fmt.Fprintf(cmd.OutOrStdout(), "  Headless Auto-Approve: %v\n", out.HeadlessAutoApprove)
			fmt.Fprintf(cmd.OutOrStdout(), "  Approval Timeout:      %d sec\n", out.ApprovalTimeoutSec)
			fmt.Fprintf(cmd.OutOrStdout(), "  Redact PII:            %v\n", out.RedactPII)
			if out.NotifyChannel != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  Notify Channel:        %s\n", out.NotifyChannel)
			}
			fmt.Fprintln(cmd.OutOrStdout())

			if len(out.SensitiveTools) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Sensitive Tools (%d)\n", len(out.SensitiveTools))
				for _, t := range out.SensitiveTools {
					fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", t)
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}

			if len(out.ExemptTools) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Exempt Tools (%d)\n", len(out.ExemptTools))
				for _, t := range out.ExemptTools {
					fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", t)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}
