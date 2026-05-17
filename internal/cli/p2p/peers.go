package p2p

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

type peersCommandInfo struct {
	PeerID string   `json:"peerId"`
	Addrs  []string `json:"addrs"`
}

var loadPeersCommandData = func(ctx context.Context, boot *bootstrap.Result) ([]peersCommandInfo, func(), error) {
	deps, err := initP2PDeps(ctx, boot)
	if err != nil {
		return nil, nil, err
	}

	peers := deps.node.ConnectedPeers()
	host := deps.node.Host()
	infos := make([]peersCommandInfo, 0, len(peers))
	for _, pid := range peers {
		conns := host.Network().ConnsToPeer(pid)
		addrs := make([]string, 0)
		for _, c := range conns {
			addrs = append(addrs, c.RemoteMultiaddr().String())
		}
		infos = append(infos, peersCommandInfo{
			PeerID: pid.String(),
			Addrs:  addrs,
		})
	}

	return infos, deps.cleanup, nil
}

func newPeersCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "peers",
		Short:         "List connected peers",
		Long:          "List connected peers (creates an ephemeral node). For the running server's peers, use GET /api/p2p/peers.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			infos, cleanup, err := loadPeersCommandData(cmd.Context(), boot)
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), infos)
			}

			if len(infos) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No connected peers.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "PEER ID\tADDRESS")
			for _, p := range infos {
				addr := ""
				if len(p.Addrs) > 0 {
					addr = p.Addrs[0]
				}
				fmt.Fprintf(w, "%s\t%s\n", p.PeerID, addr)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}
