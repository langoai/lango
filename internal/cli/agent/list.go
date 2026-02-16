package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/langowarny/lango/internal/config"
	"github.com/spf13/cobra"
)

type agentEntry struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	Status      string `json:"status,omitempty"`
}

var localAgents = []agentEntry{
	{
		Name:        "executor",
		Type:        "local",
		Description: "Executes tools including shell commands, file operations, browser automation",
	},
	{
		Name:        "researcher",
		Type:        "local",
		Description: "Searches knowledge bases, performs RAG retrieval, graph traversal",
	},
	{
		Name:        "planner",
		Type:        "local",
		Description: "Decomposes complex tasks into steps and designs execution plans",
	},
	{
		Name:        "memory-manager",
		Type:        "local",
		Description: "Manages conversational memory including observations, reflections",
	},
}

func newListCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var (
		jsonOutput bool
		check      bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available sub-agents and remote A2A agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			var entries []agentEntry

			// Add local sub-agents.
			entries = append(entries, localAgents...)

			// Add remote A2A agents.
			for _, ra := range cfg.A2A.RemoteAgents {
				e := agentEntry{
					Name: ra.Name,
					Type: "remote",
					URL:  ra.AgentCardURL,
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

			// Print local agents.
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tTYPE\tDESCRIPTION")
			for _, e := range entries {
				if e.Type != "local" {
					continue
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", e.Name, e.Type, e.Description)
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
					fmt.Fprintln(w, "NAME\tTYPE\tURL\tSTATUS")
				} else {
					fmt.Fprintln(w, "NAME\tTYPE\tURL")
				}
				for _, e := range entries {
					if e.Type != "remote" {
						continue
					}
					if check {
						fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.Name, e.Type, e.URL, e.Status)
					} else {
						fmt.Fprintf(w, "%s\t%s\t%s\n", e.Name, e.Type, e.URL)
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
