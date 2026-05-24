package a2a

import (
	"fmt"
	"text/tabwriter"

	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

func newCardCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "card",
		Short:         "Show local A2A agent card configuration",
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

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), out)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "A2A Agent Card\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Enabled:      %v\n", out.Enabled)
			if out.Enabled {
				fmt.Fprintf(cmd.OutOrStdout(), "  Base URL:     %s\n", out.BaseURL)
				fmt.Fprintf(cmd.OutOrStdout(), "  Agent Name:   %s\n", out.AgentName)
				fmt.Fprintf(cmd.OutOrStdout(), "  Description:  %s\n", out.AgentDescription)
			}
			fmt.Fprintln(cmd.OutOrStdout())

			if len(out.RemoteAgents) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Remote Agents (%d)\n", len(out.RemoteAgents))
				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
				fmt.Fprintln(w, "  NAME\tAGENT CARD URL")
				for _, r := range out.RemoteAgents {
					fmt.Fprintf(w, "  %s\t%s\n", r.Name, r.AgentCardURL)
				}
				return w.Flush()
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "No remote agents configured.")
			return err
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}
