package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
)

func newHooksCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "hooks",
		Short: "Show active hook configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			h := cfg.Hooks

			type hooksOutput struct {
				Enabled         bool     `json:"enabled"`
				SecurityFilter  bool     `json:"security_filter"`
				AccessControl   bool     `json:"access_control"`
				EventPublishing bool     `json:"event_publishing"`
				KnowledgeSave   bool     `json:"knowledge_save"`
				BlockedCommands []string `json:"blocked_commands,omitempty"`
			}

			out := hooksOutput{
				Enabled:         h.Enabled,
				SecurityFilter:  h.SecurityFilter,
				AccessControl:   h.AccessControl,
				EventPublishing: h.EventPublishing,
				KnowledgeSave:   h.KnowledgeSave,
				BlockedCommands: h.BlockedCommands,
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			fmt.Println("Hook Configuration")
			fmt.Printf("  Enabled:          %v\n", out.Enabled)
			fmt.Printf("  Security Filter:  %v\n", out.SecurityFilter)
			fmt.Printf("  Access Control:   %v\n", out.AccessControl)
			fmt.Printf("  Event Publishing: %v\n", out.EventPublishing)
			fmt.Printf("  Knowledge Save:   %v\n", out.KnowledgeSave)
			if len(out.BlockedCommands) > 0 {
				fmt.Printf("  Blocked Commands: %s\n", strings.Join(out.BlockedCommands, ", "))
			} else {
				fmt.Printf("  Blocked Commands: (none)\n")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
