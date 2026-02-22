package p2p

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/p2p/discovery"
)

func newDiscoverCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		tag        string
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover agents by capability",
		Long:  "Search for agents on the P2P network that advertise specific capabilities via GossipSub.",
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

			gossip, err := discovery.NewGossipService(discovery.GossipConfig{
				Host:     deps.node.Host(),
				Interval: deps.config.GossipInterval,
			})
			if err != nil {
				return fmt.Errorf("init gossip service: %w", err)
			}
			defer gossip.Stop()

			var cards []*discovery.GossipCard
			if tag != "" {
				cards = gossip.FindByCapability(tag)
			} else {
				cards = gossip.KnownPeers()
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(cards)
			}

			if len(cards) == 0 {
				fmt.Println("No agents discovered. Try connecting to bootstrap peers first.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tDID\tCAPABILITIES\tPEER ID")
			for _, c := range cards {
				caps := strings.Join(c.Capabilities, ", ")
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.Name, c.DID, caps, c.PeerID)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&tag, "tag", "", "Filter agents by capability tag")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
