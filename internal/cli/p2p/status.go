package p2p

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

func newStatusCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show P2P node status",
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

			peerID := deps.node.PeerID().String()
			addrs := deps.node.Multiaddrs()
			connectedPeers := deps.node.ConnectedPeers()

			listenAddrs := make([]string, len(addrs))
			for i, a := range addrs {
				listenAddrs[i] = a.String()
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]interface{}{
					"peerId":         peerID,
					"listenAddrs":    listenAddrs,
					"connectedPeers": len(connectedPeers),
					"maxPeers":       deps.config.MaxPeers,
					"mdns":           deps.config.EnableMDNS,
					"relay":          deps.config.EnableRelay,
					"zkHandshake":    deps.config.ZKHandshake,
				})
			}

			fmt.Println("P2P Node Status")
			fmt.Printf("  Peer ID:          %s\n", peerID)
			fmt.Printf("  Listen Addrs:     %v\n", listenAddrs)
			fmt.Printf("  Connected Peers:  %d / %d\n", len(connectedPeers), deps.config.MaxPeers)
			fmt.Printf("  mDNS:             %v\n", deps.config.EnableMDNS)
			fmt.Printf("  Relay:            %v\n", deps.config.EnableRelay)
			fmt.Printf("  ZK Handshake:     %v\n", deps.config.ZKHandshake)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
