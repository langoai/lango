package p2p

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

func newDisconnectCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disconnect <peer-id>",
		Short: "Disconnect from a peer",
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

			peerID, err := peer.Decode(args[0])
			if err != nil {
				return fmt.Errorf("parse peer ID: %w", err)
			}

			if err := deps.node.Host().Network().ClosePeer(peerID); err != nil {
				return fmt.Errorf("disconnect from %s: %w", peerID, err)
			}

			fmt.Printf("Disconnected from peer %s\n", peerID)
			return nil
		},
	}

	return cmd
}
