package p2p

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/p2p/handshake"
)

func newSessionCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage P2P sessions",
		Long:  "List, revoke, or revoke-all authenticated peer sessions.",
	}

	cmd.AddCommand(newSessionListCmd(bootLoader))
	cmd.AddCommand(newSessionRevokeCmd(bootLoader))
	cmd.AddCommand(newSessionRevokeAllCmd(bootLoader))

	return cmd
}

func newSessionListCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active P2P sessions",
		Long:  "List all active (non-expired, non-invalidated) peer sessions.",
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

			sessions := deps.sessions.ActiveSessions()

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(sessions)
			}

			if len(sessions) == 0 {
				fmt.Println("No active sessions.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "PEER DID\tCREATED\tEXPIRES\tZK VERIFIED")
			for _, s := range sessions {
				fmt.Fprintf(w, "%s\t%s\t%s\t%v\n",
					s.PeerDID,
					s.CreatedAt.Format(time.RFC3339),
					s.ExpiresAt.Format(time.RFC3339),
					s.ZKVerified,
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func newSessionRevokeCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var peerDID string

	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke a peer's session",
		Long:  "Explicitly invalidate and revoke the session for a specific peer DID.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if peerDID == "" {
				return fmt.Errorf("--peer-did is required")
			}

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

			deps.sessions.Invalidate(peerDID, handshake.ReasonManualRevoke)
			fmt.Printf("Session for %s revoked.\n", peerDID)
			return nil
		},
	}

	cmd.Flags().StringVar(&peerDID, "peer-did", "", "The DID of the peer to revoke")
	return cmd
}

func newSessionRevokeAllCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke-all",
		Short: "Revoke all active sessions",
		Long:  "Invalidate and revoke all active peer sessions.",
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

			deps.sessions.InvalidateAll(handshake.ReasonManualRevoke)
			fmt.Println("All sessions revoked.")
			return nil
		},
	}

	return cmd
}
