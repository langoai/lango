package provenance

import (
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
	var depth int
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

			svcs := loadServices(boot)
			nodes, err := svcs.tree.GetTree(cmd.Context(), args[0], depth)
			if err != nil {
				return fmt.Errorf("get session tree: %w", err)
			}
			for _, line := range renderSessionTree(nodes) {
				cmd.Println(line)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&depth, "depth", 10, "Maximum descendant depth to display")
	return cmd
}

func newSessionListCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		limit  int
		status string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List persisted session nodes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			boot, err := bootLoader()
			if err != nil {
				return err
			}
			defer boot.DBClient.Close()

			svcs := loadServices(boot)
			nodes, err := svcs.treeStore.ListAll(cmd.Context(), limit)
			if err != nil {
				return fmt.Errorf("list session nodes: %w", err)
			}
			wantStatus := provenancepkg.SessionStatus(status)
			for _, node := range nodes {
				if status != "" && node.Status != wantStatus {
					continue
				}
				cmd.Printf("%s\tparent=%s\tagent=%s\tdepth=%d\tstatus=%s\n",
					node.SessionKey, emptyDash(node.ParentKey), node.AgentName, node.Depth, node.Status)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum entries to display")
	cmd.Flags().StringVar(&status, "status", "", "Optional status filter (active, merged, discarded, completed)")
	return cmd
}

func renderSessionTree(nodes []provenancepkg.SessionNode) []string {
	index := make(map[string][]provenancepkg.SessionNode)
	roots := make([]provenancepkg.SessionNode, 0, len(nodes))
	for _, node := range nodes {
		if node.ParentKey == "" {
			roots = append(roots, node)
			continue
		}
		index[node.ParentKey] = append(index[node.ParentKey], node)
	}

	var lines []string
	var walk func(node provenancepkg.SessionNode, depth int)
	walk = func(node provenancepkg.SessionNode, depth int) {
		lines = append(lines, fmt.Sprintf("%s%s [%s]", strings.Repeat("  ", depth), node.SessionKey, node.Status))
		for _, child := range index[node.SessionKey] {
			walk(child, depth+1)
		}
	}
	for _, root := range roots {
		walk(root, 0)
	}
	return lines
}

func emptyDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
