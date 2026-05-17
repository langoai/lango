package p2p

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
)

const defaultConnectTimeout = 30 * time.Second

type connectDeps struct {
	config  *config.P2PConfig
	connect func(context.Context, peer.AddrInfo) error
	cleanup func()
}

var loadConnectDeps = func(boot *bootstrap.Result) (connectDeps, error) {
	deps, err := initP2PDeps(boot)
	if err != nil {
		return connectDeps{}, err
	}

	return connectDeps{
		config:  deps.config,
		connect: connectHost(deps.node.Host()),
		cleanup: deps.cleanup,
	}, nil
}

var connectHost = func(host host.Host) func(context.Context, peer.AddrInfo) error {
	return func(ctx context.Context, pi peer.AddrInfo) error {
		return host.Connect(ctx, pi)
	}
}

var connectToPeer = func(ctx context.Context, boot *bootstrap.Result, target string) (string, func(), error) {
	deps, err := loadConnectDeps(boot)
	if err != nil {
		return "", nil, err
	}

	maddr, err := ma.NewMultiaddr(target)
	if err != nil {
		if deps.cleanup != nil {
			deps.cleanup()
		}
		return "", nil, fmt.Errorf("parse multiaddr: %w", err)
	}

	pi, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		if deps.cleanup != nil {
			deps.cleanup()
		}
		return "", nil, fmt.Errorf("parse peer info: %w", err)
	}

	timeout := connectTimeout(deps.config)
	commandDeadlineApplies := commandDeadlineBeforeTimeout(ctx, timeout, time.Now())
	connectCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := deps.connect(connectCtx, *pi); err != nil {
		if deps.cleanup != nil {
			deps.cleanup()
		}
		return "", nil, connectPeerError(ctx, pi.ID, timeout, commandDeadlineApplies, err)
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

			peerID, cleanup, err := connectToPeer(cmd.Context(), boot, args[0])
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

func connectTimeout(cfg *config.P2PConfig) time.Duration {
	if cfg != nil && cfg.HandshakeTimeout > 0 {
		return cfg.HandshakeTimeout
	}
	return defaultConnectTimeout
}

func commandDeadlineBeforeTimeout(parent context.Context, timeout time.Duration, startedAt time.Time) bool {
	parentDeadline, ok := parent.Deadline()
	if !ok {
		return false
	}
	configDeadline := startedAt.Add(timeout)
	return !parentDeadline.After(configDeadline)
}

func connectPeerError(parent context.Context, peerID peer.ID, timeout time.Duration, commandDeadlineApplies bool, err error) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded) && commandDeadlineApplies:
		return fmt.Errorf("connect to %s timed out by command context deadline: %w", peerID, err)
	case errors.Is(err, context.DeadlineExceeded):
		return fmt.Errorf("connect to %s timed out after %s: %w", peerID, timeout, err)
	case errors.Is(err, context.Canceled) && errors.Is(parent.Err(), context.Canceled):
		return fmt.Errorf("connect to %s canceled by command context: %w", peerID, err)
	case errors.Is(err, context.Canceled):
		return fmt.Errorf("connect to %s canceled: %w", peerID, err)
	default:
		return fmt.Errorf("connect to %s: %w", peerID, err)
	}
}
