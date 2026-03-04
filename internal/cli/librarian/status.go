package librarian

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
		Short: "Show librarian configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			fmt.Printf("Librarian Status\n")
			fmt.Printf("  Enabled:               %v\n", out.Enabled)
			fmt.Printf("  Observation Threshold: %d\n", out.ObservationThreshold)
			fmt.Printf("  Inquiry Cooldown:      %d turns\n", out.InquiryCooldownTurns)
			fmt.Printf("  Max Pending Inquiries: %d\n", out.MaxPendingInquiries)
			fmt.Printf("  Auto-Save Confidence:  %s\n", out.AutoSaveConfidence)
			if out.Provider != "" {
				fmt.Printf("  LLM Provider:          %s\n", out.Provider)
			}
			if out.Model != "" {
				fmt.Printf("  LLM Model:             %s\n", out.Model)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
