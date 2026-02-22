package p2p

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

func newIdentityCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Show local DID and peer identity",
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

			peerID := deps.node.PeerID().String()
			addrs := deps.node.Multiaddrs()

			listenAddrs := make([]string, len(addrs))
			for i, a := range addrs {
				listenAddrs[i] = a.String()
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]interface{}{
					"peerId":      peerID,
					"listenAddrs": listenAddrs,
					"keyDir":      deps.config.KeyDir,
				})
			}

			fmt.Println("P2P Identity")
			fmt.Printf("  Peer ID:      %s\n", peerID)
			fmt.Printf("  Key Dir:      %s\n", deps.config.KeyDir)
			fmt.Printf("  Listen Addrs:\n")
			for _, a := range listenAddrs {
				fmt.Printf("    %s\n", a)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
