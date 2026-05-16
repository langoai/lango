package p2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

var connectToPeer = func(boot *bootstrap.Result, target string) (string, func(), error) {
	deps, err := initP2PDeps(boot)
	if err != nil {
		return "", nil, err
	}

	maddr, err := ma.NewMultiaddr(target)
	if err != nil {
		deps.cleanup()
		return "", nil, fmt.Errorf("parse multiaddr: %w", err)
	}

	pi, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		deps.cleanup()
		return "", nil, fmt.Errorf("parse peer info: %w", err)
	}

	if err := deps.node.Host().Connect(context.Background(), *pi); err != nil {
		deps.cleanup()
		return "", nil, fmt.Errorf("connect to %s: %w", pi.ID, err)
	}

	return pi.ID.String(), deps.cleanup, nil
}

func newConnectCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect <multiaddr>",
		Short: "Connect to a peer by multiaddr",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			peerID, cleanup, err := connectToPeer(boot, args[0])
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Connected to peer %s\n", peerID)
			return nil
		},
	}

	return cmd
}
