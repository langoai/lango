package p2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

var disconnectFromPeer = func(ctx context.Context, boot *bootstrap.Result, target string) (string, func(), error) {
	deps, err := initP2PDeps(ctx, boot)
	if err != nil {
		return "", nil, err
	}

	peerID, err := peer.Decode(target)
	if err != nil {
		deps.cleanup()
		return "", nil, fmt.Errorf("parse peer ID: %w", err)
	}

	if err := deps.node.Host().Network().ClosePeer(peerID); err != nil {
		deps.cleanup()
		return "", nil, fmt.Errorf("disconnect from %s: %w", peerID, err)
	}

	return peerID.String(), deps.cleanup, nil
}

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
			defer boot.Close()

			peerID, cleanup, err := disconnectFromPeer(cmd.Context(), boot, args[0])
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Disconnected from peer %s\n", peerID)
			return nil
		},
	}

	return cmd
}
