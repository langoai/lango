package p2p

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

type statusCommandData struct {
	peerID         string
	listenAddrs    []string
	connectedPeers int
	maxPeers       int
	mdns           bool
	relay          bool
	zkHandshake    bool
}

var loadStatusCommandData = func(boot *bootstrap.Result) (statusCommandData, func(), error) {
	deps, err := initP2PDeps(boot)
	if err != nil {
		return statusCommandData{}, nil, err
	}

	peerID := deps.node.PeerID().String()
	addrs := deps.node.Multiaddrs()
	connectedPeers := deps.node.ConnectedPeers()

	listenAddrs := make([]string, len(addrs))
	for i, a := range addrs {
		listenAddrs[i] = a.String()
	}

	return statusCommandData{
		peerID:         peerID,
		listenAddrs:    listenAddrs,
		connectedPeers: len(connectedPeers),
		maxPeers:       deps.config.MaxPeers,
		mdns:           deps.config.EnableMDNS,
		relay:          deps.config.EnableRelay,
		zkHandshake:    deps.config.ZKHandshake,
	}, deps.cleanup, nil
}

func newStatusCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "status",
		Short:         "Show P2P node status",
		Long:          "Show P2P node status (creates an ephemeral node). For the running server's node, use GET /api/p2p/status.",
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

			data, cleanup, err := loadStatusCommandData(boot)
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), map[string]interface{}{
					"peerId":         data.peerID,
					"listenAddrs":    data.listenAddrs,
					"connectedPeers": data.connectedPeers,
					"maxPeers":       data.maxPeers,
					"mdns":           data.mdns,
					"relay":          data.relay,
					"zkHandshake":    data.zkHandshake,
				})
			}

			fmt.Fprintln(cmd.OutOrStdout(), "P2P Node Status")
			fmt.Fprintf(cmd.OutOrStdout(), "  Peer ID:          %s\n", data.peerID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Listen Addrs:     %v\n", data.listenAddrs)
			fmt.Fprintf(cmd.OutOrStdout(), "  Connected Peers:  %d / %d\n", data.connectedPeers, data.maxPeers)
			fmt.Fprintf(cmd.OutOrStdout(), "  mDNS:             %v\n", data.mdns)
			fmt.Fprintf(cmd.OutOrStdout(), "  Relay:            %v\n", data.relay)
			fmt.Fprintf(cmd.OutOrStdout(), "  ZK Handshake:     %v\n", data.zkHandshake)

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}
