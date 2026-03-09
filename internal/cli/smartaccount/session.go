package smartaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	sa "github.com/langoai/lango/internal/smartaccount"
)

func sessionCmd(bootLoader BootLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage session keys",
		Long: `Manage ERC-7579 session keys for delegated transaction signing.

Examples:
  lango account session list
  lango account session create --targets 0x... --duration 24h --limit "10.00"
  lango account session revoke <session-id>
  lango account session revoke --all`,
	}

	cmd.AddCommand(sessionCreateCmd(bootLoader))
	cmd.AddCommand(sessionListCmd(bootLoader))
	cmd.AddCommand(sessionRevokeCmd(bootLoader))

	return cmd
}

func sessionCreateCmd(bootLoader BootLoader) *cobra.Command {
	var (
		targets   []string
		functions []string
		limit     string
		duration  string
		output    string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new session key",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			deps, err := initSmartAccountDeps(boot)
			if err != nil {
				return err
			}
			defer deps.cleanup()

			// Parse duration.
			dur, err := time.ParseDuration(duration)
			if err != nil {
				return fmt.Errorf("parse duration %q: %w", duration, err)
			}

			// Parse spend limit (in wei string).
			spendLimit := new(big.Int)
			if limit != "" && limit != "0" {
				if _, ok := spendLimit.SetString(limit, 10); !ok {
					// Try parsing as float ETH value and convert to wei.
					return fmt.Errorf("parse spend limit %q: provide a wei amount (integer)", limit)
				}
			}

			// Parse target addresses.
			allowedTargets := make([]common.Address, 0, len(targets))
			for _, t := range targets {
				if !common.IsHexAddress(t) {
					return fmt.Errorf("invalid target address: %s", t)
				}
				allowedTargets = append(allowedTargets, common.HexToAddress(t))
			}

			now := time.Now()
			p := sa.SessionPolicy{
				AllowedTargets:   allowedTargets,
				AllowedFunctions: functions,
				SpendLimit:       spendLimit,
				ValidAfter:       now,
				ValidUntil:       now.Add(dur),
				Active:           true,
			}

			ctx := context.Background()
			sk, err := deps.sessionManager.Create(ctx, p, "")
			if err != nil {
				return fmt.Errorf("create session: %w", err)
			}

			type sessionResult struct {
				ID        string   `json:"id"`
				Address   string   `json:"address"`
				Targets   []string `json:"allowedTargets"`
				Functions []string `json:"allowedFunctions"`
				Limit     string   `json:"spendLimit"`
				ExpiresAt string   `json:"expiresAt"`
				CreatedAt string   `json:"createdAt"`
			}

			targetStrs := make([]string, 0, len(sk.Policy.AllowedTargets))
			for _, a := range sk.Policy.AllowedTargets {
				targetStrs = append(targetStrs, a.Hex())
			}

			result := sessionResult{
				ID:        sk.ID,
				Address:   sk.Address.Hex(),
				Targets:   targetStrs,
				Functions: sk.Policy.AllowedFunctions,
				Limit:     sk.Policy.SpendLimit.String(),
				ExpiresAt: sk.ExpiresAt.Format(time.RFC3339),
				CreatedAt: sk.CreatedAt.Format(time.RFC3339),
			}

			if output == "json" {
				data, marshalErr := json.MarshalIndent(result, "", "  ")
				if marshalErr != nil {
					return fmt.Errorf("marshal json: %w", marshalErr)
				}
				fmt.Println(string(data))
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "Session Key Created")
			fmt.Fprintln(w, "-------------------")
			fmt.Fprintf(w, "ID:\t%s\n", result.ID)
			fmt.Fprintf(w, "Address:\t%s\n", result.Address)
			fmt.Fprintf(w, "Targets:\t%s\n", strings.Join(result.Targets, ", "))
			fmt.Fprintf(w, "Functions:\t%s\n", strings.Join(result.Functions, ", "))
			fmt.Fprintf(w, "Spend Limit:\t%s wei\n", result.Limit)
			fmt.Fprintf(w, "Expires:\t%s\n", result.ExpiresAt)
			fmt.Fprintf(w, "Created:\t%s\n", result.CreatedAt)
			return w.Flush()
		},
	}

	cmd.Flags().StringSliceVar(&targets, "targets", nil, "allowed target addresses (comma-separated)")
	cmd.Flags().StringSliceVar(&functions, "functions", nil, "allowed function selectors (comma-separated)")
	cmd.Flags().StringVar(&limit, "limit", "0", "spend limit in wei")
	cmd.Flags().StringVar(&duration, "duration", "24h", "session duration (e.g., 1h, 24h)")
	cmd.Flags().StringVar(&output, "output", "table", "output format (table|json)")

	return cmd
}

func sessionListCmd(bootLoader BootLoader) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active session keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			deps, err := initSmartAccountDeps(boot)
			if err != nil {
				return err
			}
			defer deps.cleanup()

			ctx := context.Background()
			sessions, err := deps.sessionManager.List(ctx)
			if err != nil {
				return fmt.Errorf("list sessions: %w", err)
			}

			type sessionEntry struct {
				ID        string `json:"id"`
				Address   string `json:"address"`
				ParentID  string `json:"parentId,omitempty"`
				ExpiresAt string `json:"expiresAt"`
				Limit     string `json:"spendLimit"`
				Status    string `json:"status"`
			}

			entries := make([]sessionEntry, 0, len(sessions))
			for _, sk := range sessions {
				status := "active"
				if sk.Revoked {
					status = "revoked"
				} else if sk.IsExpired() {
					status = "expired"
				}
				limitStr := "unlimited"
				if sk.Policy.SpendLimit != nil && sk.Policy.SpendLimit.Sign() > 0 {
					limitStr = sk.Policy.SpendLimit.String()
				}
				entries = append(entries, sessionEntry{
					ID:        sk.ID,
					Address:   sk.Address.Hex(),
					ParentID:  sk.ParentID,
					ExpiresAt: sk.ExpiresAt.Format(time.RFC3339),
					Limit:     limitStr,
					Status:    status,
				})
			}

			if output == "json" {
				data, marshalErr := json.MarshalIndent(entries, "", "  ")
				if marshalErr != nil {
					return fmt.Errorf("marshal json: %w", marshalErr)
				}
				fmt.Println(string(data))
				return nil
			}

			if len(entries) == 0 {
				fmt.Println("No session keys found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tADDRESS\tPARENT\tEXPIRES\tSPEND_LIMIT\tSTATUS")
			for _, e := range entries {
				parent := "-"
				if e.ParentID != "" {
					parent = e.ParentID[:8] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
					e.ID[:8]+"...", e.Address[:10]+"...", parent,
					e.ExpiresAt, e.Limit, e.Status)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "output format (table|json)")
	return cmd
}

func sessionRevokeCmd(bootLoader BootLoader) *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "revoke [session-id]",
		Short: "Revoke a session key or all session keys",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			deps, err := initSmartAccountDeps(boot)
			if err != nil {
				return err
			}
			defer deps.cleanup()

			if !all && len(args) == 0 {
				return fmt.Errorf("provide a session ID or use --all to revoke all sessions")
			}

			ctx := context.Background()

			if all {
				if revokeErr := deps.sessionManager.RevokeAll(ctx); revokeErr != nil {
					return fmt.Errorf("revoke all sessions: %w", revokeErr)
				}
				fmt.Println("All active session keys revoked.")
				return nil
			}

			sessionID := args[0]
			if revokeErr := deps.sessionManager.Revoke(ctx, sessionID); revokeErr != nil {
				return fmt.Errorf("revoke session %s: %w", sessionID, revokeErr)
			}
			fmt.Printf("Session key %s revoked.\n", sessionID)
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "revoke all active session keys")
	return cmd
}
