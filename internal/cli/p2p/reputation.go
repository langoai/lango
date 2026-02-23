package p2p

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/p2p/reputation"
)

func newReputationCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		peerDID    string
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "reputation",
		Short: "Show peer reputation and trust score",
		Long:  "Query the reputation system for a peer's trust score, exchange history, and interaction timeline.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if peerDID == "" {
				return fmt.Errorf("--peer-did is required")
			}

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.DBClient.Close()

			logger := logging.Sugar()
			if logger == nil {
				l, _ := zap.NewProduction()
				logger = l.Sugar()
			}

			store := reputation.NewStore(boot.DBClient, logger)
			details, err := store.GetDetails(cmd.Context(), peerDID)
			if err != nil {
				return fmt.Errorf("get reputation: %w", err)
			}

			if details == nil {
				if jsonOutput {
					enc := json.NewEncoder(os.Stdout)
					enc.SetIndent("", "  ")
					return enc.Encode(map[string]interface{}{
						"peerDid":    peerDID,
						"trustScore": 0.0,
						"message":    "no reputation record found",
					})
				}
				fmt.Printf("No reputation record found for %s\n", peerDID)
				return nil
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(details)
			}

			fmt.Println("Peer Reputation")
			fmt.Printf("  Peer DID:          %s\n", details.PeerDID)
			fmt.Printf("  Trust Score:       %.4f\n", details.TrustScore)
			fmt.Printf("  Successes:         %d\n", details.SuccessfulExchanges)
			fmt.Printf("  Failures:          %d\n", details.FailedExchanges)
			fmt.Printf("  Timeouts:          %d\n", details.TimeoutCount)
			fmt.Printf("  First Seen:        %s\n", details.FirstSeen.Format("2006-01-02 15:04:05"))
			fmt.Printf("  Last Interaction:  %s\n", details.LastInteraction.Format("2006-01-02 15:04:05"))

			return nil
		},
	}

	cmd.Flags().StringVar(&peerDID, "peer-did", "", "The DID of the peer to query")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
