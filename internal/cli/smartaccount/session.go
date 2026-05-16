package smartaccount

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	sa "github.com/langoai/lango/internal/smartaccount"
)

type sessionCreateResult struct {
	ID        string   `json:"id"`
	Address   string   `json:"address"`
	Targets   []string `json:"allowedTargets"`
	Functions []string `json:"allowedFunctions"`
	Limit     string   `json:"spendLimit"`
	ExpiresAt string   `json:"expiresAt"`
	CreatedAt string   `json:"createdAt"`
}

type sessionListEntry struct {
	ID        string `json:"id"`
	Address   string `json:"address"`
	ParentID  string `json:"parentId,omitempty"`
	ExpiresAt string `json:"expiresAt"`
	Limit     string `json:"spendLimit"`
	Status    string `json:"status"`
}

var executeSessionCreate = func(bootLoader BootLoader, targets, functions []string, limit, duration string) (sessionCreateResult, func(), error) {
	boot, err := bootLoader()
	if err != nil {
		return sessionCreateResult{}, nil, fmt.Errorf("bootstrap: %w", err)
	}

	deps, err := initSmartAccountDeps(boot)
	if err != nil {
		boot.Close()
		return sessionCreateResult{}, nil, err
	}

	dur, err := time.ParseDuration(duration)
	if err != nil {
		deps.cleanup()
		boot.Close()
		return sessionCreateResult{}, nil, fmt.Errorf("parse duration %q: %w", duration, err)
	}

	spendLimit := new(big.Int)
	if limit != "" && limit != "0" {
		if _, ok := spendLimit.SetString(limit, 10); !ok {
			deps.cleanup()
			boot.Close()
			return sessionCreateResult{}, nil, fmt.Errorf("parse spend limit %q: provide a wei amount (integer)", limit)
		}
	}

	allowedTargets := make([]common.Address, 0, len(targets))
	for _, t := range targets {
		if !common.IsHexAddress(t) {
			deps.cleanup()
			boot.Close()
			return sessionCreateResult{}, nil, fmt.Errorf("invalid target address: %s", t)
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
		deps.cleanup()
		boot.Close()
		return sessionCreateResult{}, nil, fmt.Errorf("create session: %w", err)
	}

	targetStrs := make([]string, 0, len(sk.Policy.AllowedTargets))
	for _, a := range sk.Policy.AllowedTargets {
		targetStrs = append(targetStrs, a.Hex())
	}

	result := sessionCreateResult{
		ID:        sk.ID,
		Address:   sk.Address.Hex(),
		Targets:   targetStrs,
		Functions: sk.Policy.AllowedFunctions,
		Limit:     sk.Policy.SpendLimit.String(),
		ExpiresAt: sk.ExpiresAt.Format(time.RFC3339),
		CreatedAt: sk.CreatedAt.Format(time.RFC3339),
	}

	return result, func() {
		deps.cleanup()
		boot.Close()
	}, nil
}

var loadSessionList = func(bootLoader BootLoader) ([]sessionListEntry, func(), error) {
	boot, err := bootLoader()
	if err != nil {
		return nil, nil, fmt.Errorf("bootstrap: %w", err)
	}

	deps, err := initSmartAccountDeps(boot)
	if err != nil {
		boot.Close()
		return nil, nil, err
	}

	ctx := context.Background()
	sessions, err := deps.sessionManager.List(ctx)
	if err != nil {
		deps.cleanup()
		boot.Close()
		return nil, nil, fmt.Errorf("list sessions: %w", err)
	}

	entries := make([]sessionListEntry, 0, len(sessions))
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
		entries = append(entries, sessionListEntry{
			ID:        sk.ID,
			Address:   sk.Address.Hex(),
			ParentID:  sk.ParentID,
			ExpiresAt: sk.ExpiresAt.Format(time.RFC3339),
			Limit:     limitStr,
			Status:    status,
		})
	}

	return entries, func() {
		deps.cleanup()
		boot.Close()
	}, nil
}

var executeSessionRevoke = func(bootLoader BootLoader, all bool, sessionID string) (string, func(), error) {
	boot, err := bootLoader()
	if err != nil {
		return "", nil, fmt.Errorf("bootstrap: %w", err)
	}

	deps, err := initSmartAccountDeps(boot)
	if err != nil {
		boot.Close()
		return "", nil, err
	}

	ctx := context.Background()
	if all {
		if revokeErr := deps.sessionManager.RevokeAll(ctx); revokeErr != nil {
			deps.cleanup()
			boot.Close()
			return "", nil, fmt.Errorf("revoke all sessions: %w", revokeErr)
		}
		return "All active session keys revoked.", func() {
			deps.cleanup()
			boot.Close()
		}, nil
	}

	if revokeErr := deps.sessionManager.Revoke(ctx, sessionID); revokeErr != nil {
		deps.cleanup()
		boot.Close()
		return "", nil, fmt.Errorf("revoke session %s: %w", sessionID, revokeErr)
	}

	return fmt.Sprintf("Session key %s revoked.", sessionID), func() {
		deps.cleanup()
		boot.Close()
	}, nil
}

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
		Use:           "create",
		Short:         "Create a new session key",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveTableOrJSONOutput(cmd)
			if err != nil {
				return err
			}
			result, cleanup, err := executeSessionCreate(bootLoader, targets, functions, limit, duration)
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), result)
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
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
	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List active session keys",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveTableOrJSONOutput(cmd)
			if err != nil {
				return err
			}
			entries, cleanup, err := loadSessionList(bootLoader)
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), entries)
			}

			if len(entries) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No session keys found.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
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

	cmd.Flags().String("output", "table", "output format (table|json)")
	return cmd
}

func sessionRevokeCmd(bootLoader BootLoader) *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "revoke [session-id]",
		Short: "Revoke a session key or all session keys",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !all && len(args) == 0 {
				return fmt.Errorf("provide a session ID or use --all to revoke all sessions")
			}

			sessionID := ""
			if !all {
				sessionID = args[0]
			}

			message, cleanup, err := executeSessionRevoke(bootLoader, all, sessionID)
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}
			fmt.Fprintln(cmd.OutOrStdout(), message)
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "revoke all active session keys")
	return cmd
}
