package provenance

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	provenancepkg "github.com/langoai/lango/internal/provenance"
)

func newSessionCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "View session tree and hierarchy",
	}

	cmd.AddCommand(newSessionTreeCmd(bootLoader))
	cmd.AddCommand(newSessionListCmd(bootLoader))

	return cmd
}

func newSessionTreeCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var maxDepth int

	cmd := &cobra.Command{
		Use:   "tree <session-key>",
		Short: "Show session hierarchy tree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return err
			}
			defer boot.DBClient.Close()

			if !boot.Config.Provenance.Enabled {
				cmd.Println("Provenance is disabled. Enable with: lango config set provenance.enabled true")
				return nil
			}

			store := provenancepkg.NewMemoryTreeStore()
			tree := provenancepkg.NewSessionTree(store)

			nodes, err := tree.GetTree(context.Background(), args[0], maxDepth)
			if err != nil {
				return fmt.Errorf("get session tree: %w", err)
			}

			for _, node := range nodes {
				indent := strings.Repeat("  ", node.Depth)
				cmd.Printf("%s%s [%s] %s (%s)\n",
					indent, node.SessionKey, node.AgentName, node.Goal, node.Status)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&maxDepth, "depth", 10, "Maximum tree depth to display")

	return cmd
}

func newSessionListCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all session nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return err
			}
			defer boot.DBClient.Close()

			if !boot.Config.Provenance.Enabled {
				cmd.Println("Provenance is disabled. Enable with: lango config set provenance.enabled true")
				return nil
			}

			store := provenancepkg.NewMemoryTreeStore()
			nodes, err := store.ListAll(context.Background(), 50)
			if err != nil {
				return fmt.Errorf("list sessions: %w", err)
			}

			if len(nodes) == 0 {
				cmd.Println("No session nodes found.")
				return nil
			}

			for _, node := range nodes {
				cmd.Printf("%s\t%s\t%s\t%s\tdepth=%d\n",
					node.SessionKey, node.AgentName, node.Status, node.Goal, node.Depth)
			}
			return nil
		},
	}
}
