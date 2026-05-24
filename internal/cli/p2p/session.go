package p2p

import (
	"context"
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/p2p/handshake"
)

var loadSessionListData = func(ctx context.Context, boot *bootstrap.Result) ([]*handshake.Session, func(), error) {
	deps, err := initP2PDeps(ctx, boot)
	if err != nil {
		return nil, nil, err
	}
	return deps.sessions.ActiveSessions(), deps.cleanup, nil
}

var revokeSessionForPeer = func(ctx context.Context, boot *bootstrap.Result, peerDID string) (func(), error) {
	deps, err := initP2PDeps(ctx, boot)
	if err != nil {
		return nil, err
	}
	deps.sessions.Invalidate(peerDID, handshake.ReasonManualRevoke)
	return deps.cleanup, nil
}

var revokeAllSessions = func(ctx context.Context, boot *bootstrap.Result) (func(), error) {
	deps, err := initP2PDeps(ctx, boot)
	if err != nil {
		return nil, err
	}
	deps.sessions.InvalidateAll(handshake.ReasonManualRevoke)
	return deps.cleanup, nil
}

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
	var output string

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List active P2P sessions",
		Long:          "List all active (non-expired, non-invalidated) peer sessions.",
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

			sessions, cleanup, err := loadSessionListData(cmd.Context(), boot)
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), sessions)
			}

			if len(sessions) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No active sessions.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
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

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
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
			defer boot.Close()

			cleanup, err := revokeSessionForPeer(cmd.Context(), boot, peerDID)
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Session for %s revoked.\n", peerDID)
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
			defer boot.Close()

			cleanup, err := revokeAllSessions(cmd.Context(), boot)
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}
			fmt.Fprintln(cmd.OutOrStdout(), "All sessions revoked.")
			return nil
		},
	}

	return cmd
}
