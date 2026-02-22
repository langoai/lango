package p2p

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

func newPeersCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "peers",
		Short: "List connected peers",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.DBClient.Close()

			deps, err := initP2PDeps(boot)
			if err != nil {
				return err
			}
			defer deps.cleanup()

			peers := deps.node.ConnectedPeers()
			host := deps.node.Host()

			type peerInfo struct {
				PeerID string   `json:"peerId"`
				Addrs  []string `json:"addrs"`
			}

			infos := make([]peerInfo, 0, len(peers))
			for _, pid := range peers {
				conns := host.Network().ConnsToPeer(pid)
				addrs := make([]string, 0)
				for _, c := range conns {
					addrs = append(addrs, c.RemoteMultiaddr().String())
				}
				infos = append(infos, peerInfo{
					PeerID: pid.String(),
					Addrs:  addrs,
				})
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(infos)
			}

			if len(infos) == 0 {
				fmt.Println("No connected peers.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
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

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
