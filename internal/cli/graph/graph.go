package graph

import (
	"fmt"

	"github.com/langowarny/lango/internal/config"
	graphstore "github.com/langowarny/lango/internal/graph"
	"github.com/spf13/cobra"
)

// NewGraphCmd creates the graph command with lazy config loading.
func NewGraphCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "graph",
		Short: "Manage the knowledge graph store",
	}

	cmd.AddCommand(newStatusCmd(cfgLoader))
	cmd.AddCommand(newQueryCmd(cfgLoader))
	cmd.AddCommand(newStatsCmd(cfgLoader))
	cmd.AddCommand(newClearCmd(cfgLoader))

	return cmd
}

func initGraphStore(cfg *config.Config) (graphstore.Store, error) {
	if !cfg.Graph.Enabled {
		return nil, fmt.Errorf("graph store is not enabled (set graph.enabled via lango onboard)")
	}
	if cfg.Graph.DatabasePath == "" {
		return nil, fmt.Errorf("graph database path is not configured")
	}
	store, err := graphstore.NewBoltStore(cfg.Graph.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("open graph store: %w", err)
	}
	return store, nil
}
