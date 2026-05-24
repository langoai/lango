package memory

import (
	"fmt"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/langoai/lango/internal/agentmemory"
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/session"
)

func newAgentsCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "agents",
		Short:         "Show agent memory configuration and status",
		Long:          "Display agent memory system configuration. Agent memories are persistent, retained across restarts.",
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

			if !cfg.AgentMemory.Enabled {
				return fmt.Errorf("agent memory is not enabled (set agentMemory.enabled = true)")
			}

			store, cleanup, err := initAgentMemoryStore(cfg)
			if err != nil {
				return err
			}
			defer cleanup()

			names, err := store.ListAgentNames()
			if err != nil {
				return fmt.Errorf("list agent memories: %w", err)
			}
			sort.Strings(names)

			type agentSummary struct {
				AgentName   string `json:"agent_name"`
				EntryCount  int    `json:"entry_count"`
				LastUpdated string `json:"last_updated"`
			}
			summaries := make([]agentSummary, 0, len(names))
			for _, name := range names {
				entries, err := store.ListAll(name)
				if err != nil {
					return fmt.Errorf("list memories for %s: %w", name, err)
				}
				lastUpdated := ""
				if len(entries) > 0 {
					lastUpdated = entries[0].UpdatedAt.Format(time.DateTime)
				}
				summaries = append(summaries, agentSummary{
					AgentName:   name,
					EntryCount:  len(entries),
					LastUpdated: lastUpdated,
				})
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), summaries)
			}

			if len(summaries) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No agent memories found.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "AGENT\tENTRIES\tLAST UPDATED")
			for _, summary := range summaries {
				lastUpdated := summary.LastUpdated
				if lastUpdated == "" {
					lastUpdated = "-"
				}
				fmt.Fprintf(w, "%s\t%d\t%s\n", summary.AgentName, summary.EntryCount, lastUpdated)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")

	return cmd
}

func newAgentCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var (
		output string
		limit  int
	)

	cmd := &cobra.Command{
		Use:           "agent <name>",
		Short:         "Show memories for a specific agent",
		Long:          "Display stored memories for a named agent. Agent memories are persistent, retained across restarts.",
		Args:          cobra.ExactArgs(1),
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

			if !cfg.AgentMemory.Enabled {
				return fmt.Errorf("agent memory is not enabled (set agentMemory.enabled = true)")
			}

			agentName := args[0]
			store, cleanup, err := initAgentMemoryStore(cfg)
			if err != nil {
				return err
			}
			defer cleanup()

			entries, err := store.ListAll(agentName)
			if err != nil {
				return fmt.Errorf("list agent memories: %w", err)
			}
			if limit > 0 && len(entries) > limit {
				entries = entries[:limit]
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), entries)
			}

			if len(entries) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No memories found for agent %q.\n", agentName)
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tCONTENT\tUSES\tCREATED")
			for _, entry := range entries {
				content := entry.Content
				if len(content) > 60 {
					content = content[:57] + "..."
				}
				id := entry.ID
				if len(id) > 8 {
					id = id[:8]
				}
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
					id,
					content,
					entry.UseCount,
					entry.CreatedAt.Format("2006-01-02 15:04"),
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum entries to show")

	return cmd
}

func initAgentMemoryStore(cfg *config.Config) (agentmemory.Store, func(), error) {
	store, err := session.NewEntStore(cfg.Session.DatabasePath)
	if err != nil {
		return nil, nil, fmt.Errorf("open session store: %w", err)
	}
	cleanup := func() {
		_ = store.Close()
	}
	return agentmemory.NewEntStore(store.Client()), cleanup, nil
}
