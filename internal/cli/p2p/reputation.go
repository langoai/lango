package p2p

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	p2preputation "github.com/langoai/lango/internal/p2p/reputation"
)

var loadReputationDetails = func(boot *bootstrap.Result, peerDID string) (*p2preputation.PeerDetails, error) {
	if boot.Storage == nil {
		return nil, fmt.Errorf("p2p reputation storage unavailable")
	}
	return boot.Storage.ReputationDetails(context.Background(), peerDID)
}

func newReputationCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		peerDID string
		output  string
	)

	cmd := &cobra.Command{
		Use:           "reputation",
		Short:         "Show peer reputation and trust score",
		Long:          "Query the reputation system for a peer's trust score, exchange history, and interaction timeline.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}

			if peerDID == "" {
				return fmt.Errorf("--peer-did is required")
			}

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			details, err := loadReputationDetails(boot, peerDID)
			if err != nil {
				return fmt.Errorf("get reputation: %w", err)
			}
			if details == nil {
				if output == "json" {
					return printJSON(cmd.OutOrStdout(), map[string]interface{}{
						"peerDid":    peerDID,
						"trustScore": 0.0,
						"message":    "no reputation record found",
					})
				}
				fmt.Fprintf(cmd.OutOrStdout(), "No reputation record found for %s\n", peerDID)
				return nil
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), details)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Peer Reputation")
			fmt.Fprintf(cmd.OutOrStdout(), "  Peer DID:          %s\n", details.PeerDID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Trust Score:       %.4f\n", details.TrustScore)
			fmt.Fprintf(cmd.OutOrStdout(), "  Successes:         %d\n", details.SuccessfulExchanges)
			fmt.Fprintf(cmd.OutOrStdout(), "  Failures:          %d\n", details.FailedExchanges)
			fmt.Fprintf(cmd.OutOrStdout(), "  Timeouts:          %d\n", details.TimeoutCount)
			fmt.Fprintf(cmd.OutOrStdout(), "  First Seen:        %s\n", details.FirstSeen.Format("2006-01-02 15:04:05"))
			fmt.Fprintf(cmd.OutOrStdout(), "  Last Interaction:  %s\n", details.LastInteraction.Format("2006-01-02 15:04:05"))

			return nil
		},
	}

	cmd.Flags().StringVar(&peerDID, "peer-did", "", "The DID of the peer to query")
	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}
