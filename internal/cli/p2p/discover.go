package p2p

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/p2p/discovery"
)

var loadDiscoverCommandData = func(boot *bootstrap.Result, tag string) ([]*discovery.GossipCard, func(), error) {
	deps, err := initP2PDeps(boot)
	if err != nil {
		return nil, nil, err
	}

	gossip, err := discovery.NewGossipService(discovery.GossipConfig{
		Host:     deps.node.Host(),
		Interval: deps.config.GossipInterval,
	})
	if err != nil {
		deps.cleanup()
		return nil, nil, fmt.Errorf("init gossip service: %w", err)
	}

	var cards []*discovery.GossipCard
	if tag != "" {
		cards = gossip.FindByCapability(tag)
	} else {
		cards = gossip.KnownPeers()
	}

	return cards, func() {
		gossip.Stop()
		deps.cleanup()
	}, nil
}

func newDiscoverCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		tag    string
		output string
	)

	cmd := &cobra.Command{
		Use:           "discover",
		Short:         "Discover agents by capability",
		Long:          "Search for agents on the P2P network that advertise specific capabilities via GossipSub.",
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

			cards, cleanup, err := loadDiscoverCommandData(boot, tag)
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), cards)
			}

			if len(cards) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No agents discovered. Try connecting to bootstrap peers first.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tDID\tCAPABILITIES\tPEER ID")
			for _, c := range cards {
				caps := strings.Join(c.Capabilities, ", ")
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.Name, c.DID, caps, c.PeerID)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&tag, "tag", "", "Filter agents by capability tag")
	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}
