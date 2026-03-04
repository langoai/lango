package a2a

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

func newCardCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "card",
		Short: "Show local A2A agent card configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			type remoteEntry struct {
				Name         string `json:"name"`
				AgentCardURL string `json:"agent_card_url"`
			}

			type cardOutput struct {
				Enabled          bool          `json:"enabled"`
				BaseURL          string        `json:"base_url,omitempty"`
				AgentName        string        `json:"agent_name,omitempty"`
				AgentDescription string        `json:"agent_description,omitempty"`
				RemoteAgents     []remoteEntry `json:"remote_agents,omitempty"`
			}

			remotes := make([]remoteEntry, 0, len(cfg.A2A.RemoteAgents))
			for _, r := range cfg.A2A.RemoteAgents {
				remotes = append(remotes, remoteEntry{
					Name:         r.Name,
					AgentCardURL: r.AgentCardURL,
				})
			}

			out := cardOutput{
				Enabled:          cfg.A2A.Enabled,
				BaseURL:          cfg.A2A.BaseURL,
				AgentName:        cfg.A2A.AgentName,
				AgentDescription: cfg.A2A.AgentDescription,
				RemoteAgents:     remotes,
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			fmt.Printf("A2A Agent Card\n")
			fmt.Printf("  Enabled:      %v\n", out.Enabled)
			if out.Enabled {
				fmt.Printf("  Base URL:     %s\n", out.BaseURL)
				fmt.Printf("  Agent Name:   %s\n", out.AgentName)
				fmt.Printf("  Description:  %s\n", out.AgentDescription)
			}
			fmt.Println()

			if len(out.RemoteAgents) > 0 {
				fmt.Printf("Remote Agents (%d)\n", len(out.RemoteAgents))
				w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
				fmt.Fprintln(w, "  NAME\tAGENT CARD URL")
				for _, r := range out.RemoteAgents {
					fmt.Fprintf(w, "  %s\t%s\n", r.Name, r.AgentCardURL)
				}
				return w.Flush()
			}

			fmt.Println("No remote agents configured.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
