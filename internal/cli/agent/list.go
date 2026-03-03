package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/langoai/lango/internal/agentregistry"
	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

type agentEntry struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Source      string `json:"source,omitempty"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	Status      string `json:"status,omitempty"`
}

func newListCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var (
		jsonOutput bool
		check      bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available sub-agents, user-defined agents, and remote agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			var entries []agentEntry

			// Load agents from registry (embedded + user-defined).
			reg := agentregistry.New()
			embeddedStore := agentregistry.NewEmbeddedStore()
			if loadErr := reg.LoadFromStore(embeddedStore); loadErr != nil {
				return fmt.Errorf("load embedded agents: %w", loadErr)
			}
			if cfg.Agent.AgentsDir != "" {
				userStore := agentregistry.NewFileStore(cfg.Agent.AgentsDir)
				_ = reg.LoadFromStore(userStore) // non-fatal
			}

			for _, def := range reg.Active() {
				entries = append(entries, agentEntry{
					Name:        def.Name,
					Type:        "local",
					Source:      agentSourceLabel(def.Source),
					Description: def.Description,
				})
			}

			// Add remote A2A agents.
			for _, ra := range cfg.A2A.RemoteAgents {
				e := agentEntry{
					Name:   ra.Name,
					Type:   "remote",
					Source: "a2a",
					URL:    ra.AgentCardURL,
				}
				if check {
					e.Status = checkConnectivity(ra.AgentCardURL)
				}
				entries = append(entries, e)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(entries)
			}

			if len(entries) == 0 {
				fmt.Println("No agents found.")
				return nil
			}

			// Print local agents (builtin + user).
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tSOURCE\tDESCRIPTION")
			for _, e := range entries {
				if e.Type != "local" {
					continue
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", e.Name, e.Source, truncate(e.Description, 60))
			}
			if err := w.Flush(); err != nil {
				return fmt.Errorf("flush table: %w", err)
			}

			// Print remote agents if any.
			hasRemote := false
			for _, e := range entries {
				if e.Type == "remote" {
					hasRemote = true
					break
				}
			}
			if hasRemote {
				fmt.Println()
				w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				if check {
					fmt.Fprintln(w, "NAME\tSOURCE\tURL\tSTATUS")
				} else {
					fmt.Fprintln(w, "NAME\tSOURCE\tURL")
				}
				for _, e := range entries {
					if e.Type != "remote" {
						continue
					}
					if check {
						fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.Name, e.Source, e.URL, e.Status)
					} else {
						fmt.Fprintf(w, "%s\t%s\t%s\n", e.Name, e.Source, e.URL)
					}
				}
				if err := w.Flush(); err != nil {
					return fmt.Errorf("flush table: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&check, "check", false, "Test connectivity to remote agents")

	return cmd
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func agentSourceLabel(source agentregistry.AgentSource) string {
	switch source {
	case agentregistry.SourceBuiltin:
		return "builtin"
	case agentregistry.SourceEmbedded:
		return "embedded"
	case agentregistry.SourceUser:
		return "user"
	case agentregistry.SourceRemote:
		return "remote"
	default:
		return "unknown"
	}
}

func checkConnectivity(url string) string {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "unreachable"
	}
	resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return "ok"
	}
	return fmt.Sprintf("http %d", resp.StatusCode)
}
