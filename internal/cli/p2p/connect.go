package p2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

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
			defer boot.DBClient.Close()

			deps, err := initP2PDeps(boot)
			if err != nil {
				return err
			}
			defer deps.cleanup()

			maddr, err := ma.NewMultiaddr(args[0])
			if err != nil {
				return fmt.Errorf("parse multiaddr: %w", err)
			}

			pi, err := peer.AddrInfoFromP2pAddr(maddr)
			if err != nil {
				return fmt.Errorf("parse peer info: %w", err)
			}

			if err := deps.node.Host().Connect(context.Background(), *pi); err != nil {
				return fmt.Errorf("connect to %s: %w", pi.ID, err)
			}

			fmt.Printf("Connected to peer %s\n", pi.ID)
			return nil
		},
	}

	return cmd
}
