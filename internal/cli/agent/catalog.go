package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
)

// toolCategoryInfo describes a tool category and its config-derived enabled state.
type toolCategoryInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ConfigKey   string `json:"config_key,omitempty"`
	Enabled     bool   `json:"enabled"`
}

func newToolsCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var (
		jsonOutput bool
		category   string
	)

	cmd := &cobra.Command{
		Use:   "tools",
		Short: "List tool categories and their availability based on config",
		Long:  "Show which tool categories are available given the current configuration. Individual tools are registered at runtime when the server starts.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			categories := buildToolCategories(cfg)

			if category != "" {
				var filtered []toolCategoryInfo
				for _, c := range categories {
					if c.Name == category {
						filtered = append(filtered, c)
					}
				}
				categories = filtered
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(categories)
			}

			if len(categories) == 0 {
				if category != "" {
					fmt.Printf("No tool category %q found.\n", category)
				} else {
					fmt.Println("No tool categories configured.")
				}
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "CATEGORY\tENABLED\tDESCRIPTION")
			for _, c := range categories {
				enabled := "yes"
				if !c.Enabled {
					enabled = "no"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", c.Name, enabled, c.Description)
			}
			return w.Flush()
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&category, "category", "", "Filter by category name")

	return cmd
}

// buildToolCategories returns the known tool categories with enabled status derived from config.
func buildToolCategories(cfg *config.Config) []toolCategoryInfo {
	return []toolCategoryInfo{
		{Name: "exec", Description: "Shell command execution", Enabled: true},
		{Name: "filesystem", Description: "File system operations", Enabled: true},
		{Name: "browser", Description: "Web browsing", ConfigKey: "tools.browser.enabled", Enabled: cfg.Tools.Browser.Enabled},
		{Name: "crypto", Description: "Cryptographic operations", ConfigKey: "security.signer.provider", Enabled: cfg.Security.Signer.Provider != ""},
		{Name: "meta", Description: "Knowledge, learning, and skill management", ConfigKey: "knowledge.enabled", Enabled: cfg.Knowledge.Enabled},
		{Name: "graph", Description: "Knowledge graph traversal", ConfigKey: "graph.enabled", Enabled: cfg.Graph.Enabled},
		{Name: "rag", Description: "Retrieval-augmented generation", ConfigKey: "embedding.rag.enabled", Enabled: cfg.Embedding.RAG.Enabled},
		{Name: "memory", Description: "Observational memory", ConfigKey: "observationalMemory.enabled", Enabled: cfg.ObservationalMemory.Enabled},
		{Name: "agent_memory", Description: "Per-agent persistent memory", ConfigKey: "agentMemory.enabled", Enabled: cfg.AgentMemory.Enabled},
		{Name: "payment", Description: "Blockchain payments (USDC on Base)", ConfigKey: "payment.enabled", Enabled: cfg.Payment.Enabled},
		{Name: "p2p", Description: "Peer-to-peer networking", ConfigKey: "p2p.enabled", Enabled: cfg.P2P.Enabled},
		{Name: "workspace", Description: "P2P workspace collaboration and git bundles", ConfigKey: "p2p.workspace.enabled", Enabled: cfg.P2P.Workspace.Enabled},
		{Name: "librarian", Description: "Knowledge inquiries and gap detection", ConfigKey: "librarian.enabled", Enabled: cfg.Librarian.Enabled},
		{Name: "cron", Description: "Cron job scheduling", ConfigKey: "cron.enabled", Enabled: cfg.Cron.Enabled},
		{Name: "background", Description: "Background task execution", ConfigKey: "background.enabled", Enabled: cfg.Background.Enabled},
		{Name: "workflow", Description: "Workflow pipeline execution", ConfigKey: "workflow.enabled", Enabled: cfg.Workflow.Enabled},
	}
}
