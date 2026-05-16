package librarian

import (
	"fmt"

	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

func newStatusCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "status",
		Short:         "Show librarian configuration",
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

			type statusOutput struct {
				Enabled              bool   `json:"enabled"`
				ObservationThreshold int    `json:"observation_threshold"`
				InquiryCooldownTurns int    `json:"inquiry_cooldown_turns"`
				MaxPendingInquiries  int    `json:"max_pending_inquiries"`
				AutoSaveConfidence   string `json:"auto_save_confidence"`
				Provider             string `json:"provider,omitempty"`
				Model                string `json:"model,omitempty"`
			}

			out := statusOutput{
				Enabled:              cfg.Librarian.Enabled,
				ObservationThreshold: cfg.Librarian.ObservationThreshold,
				InquiryCooldownTurns: cfg.Librarian.InquiryCooldownTurns,
				MaxPendingInquiries:  cfg.Librarian.MaxPendingInquiries,
				AutoSaveConfidence:   string(cfg.Librarian.AutoSaveConfidence),
				Provider:             cfg.Librarian.Provider,
				Model:                cfg.Librarian.Model,
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), out)
			}

			writer := cmd.OutOrStdout()
			fmt.Fprintf(writer, "Librarian Status\n")
			fmt.Fprintf(writer, "  Enabled:               %v\n", out.Enabled)
			fmt.Fprintf(writer, "  Observation Threshold: %d\n", out.ObservationThreshold)
			fmt.Fprintf(writer, "  Inquiry Cooldown:      %d turns\n", out.InquiryCooldownTurns)
			fmt.Fprintf(writer, "  Max Pending Inquiries: %d\n", out.MaxPendingInquiries)
			fmt.Fprintf(writer, "  Auto-Save Confidence:  %s\n", out.AutoSaveConfidence)
			if out.Provider != "" {
				fmt.Fprintf(writer, "  LLM Provider:          %s\n", out.Provider)
			}
			if out.Model != "" {
				fmt.Fprintf(writer, "  LLM Model:             %s\n", out.Model)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}
