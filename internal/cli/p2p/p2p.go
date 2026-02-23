// Package p2p provides CLI commands for P2P network management.
package p2p

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/logging"
	p2pnet "github.com/langoai/lango/internal/p2p"
)

// p2pDeps holds lazily-initialized P2P dependencies.
type p2pDeps struct {
	config  *config.P2PConfig
	node    *p2pnet.Node
	cleanup func()
}

// NewP2PCmd creates the p2p command with lazy bootstrap loading.
func NewP2PCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "p2p",
		Short: "Manage P2P network",
		Long:  "Connect to peers, manage firewall rules, discover agents, and inspect P2P node identity on the Sovereign Agent Network.",
	}

	cmd.AddCommand(newStatusCmd(bootLoader))
	cmd.AddCommand(newPeersCmd(bootLoader))
	cmd.AddCommand(newConnectCmd(bootLoader))
	cmd.AddCommand(newDisconnectCmd(bootLoader))
	cmd.AddCommand(newFirewallCmd(bootLoader))
	cmd.AddCommand(newDiscoverCmd(bootLoader))
	cmd.AddCommand(newIdentityCmd(bootLoader))
	cmd.AddCommand(newReputationCmd(bootLoader))
	cmd.AddCommand(newPricingCmd(bootLoader))

	return cmd
}

// initP2PDeps creates P2P components from a bootstrap result.
func initP2PDeps(boot *bootstrap.Result) (*p2pDeps, error) {
	cfg := boot.Config
	if !cfg.P2P.Enabled {
		return nil, fmt.Errorf("P2P networking is not enabled (set p2p.enabled = true)")
	}

	logger := logging.Sugar()
	if logger == nil {
		l, _ := zap.NewProduction()
		logger = l.Sugar()
	}

	node, err := p2pnet.NewNode(cfg.P2P, logger)
	if err != nil {
		return nil, fmt.Errorf("create P2P node: %w", err)
	}

	var wg sync.WaitGroup
	if err := node.Start(&wg); err != nil {
		node.Stop()
		return nil, fmt.Errorf("start P2P node: %w", err)
	}

	return &p2pDeps{
		config: &cfg.P2P,
		node:   node,
		cleanup: func() {
			node.Stop()
		},
	}, nil
}
