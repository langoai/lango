package smartaccount

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
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

			cfg := boot.Config
			if !cfg.SmartAccount.Enabled {
				return fmt.Errorf("smart account not enabled in config")
			}

			type createInfo struct {
				Targets   []string `json:"targets"`
				Functions []string `json:"functions"`
				Limit     string   `json:"spendLimit"`
				Duration  string   `json:"duration"`
				Status    string   `json:"status"`
			}

			info := createInfo{
				Targets:   targets,
				Functions: functions,
				Limit:     limit,
				Duration:  duration,
				Status:    "pending (requires running server)",
			}

			if output == "json" {
				data, marshalErr := json.MarshalIndent(info, "", "  ")
				if marshalErr != nil {
					return fmt.Errorf("marshal json: %w", marshalErr)
				}
				fmt.Println(string(data))
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "Session Key Creation Request")
			fmt.Fprintln(w, "----------------------------")
			fmt.Fprintf(w, "Targets:\t%s\n", strings.Join(targets, ", "))
			fmt.Fprintf(w, "Functions:\t%s\n", strings.Join(functions, ", "))
			fmt.Fprintf(w, "Spend Limit:\t%s\n", limit)
			fmt.Fprintf(w, "Duration:\t%s\n", duration)
			if flushErr := w.Flush(); flushErr != nil {
				return fmt.Errorf("flush output: %w", flushErr)
			}

			fmt.Println()
			fmt.Println("Note: Session key creation requires a running server (lango serve).")
			fmt.Println("Use the 'smart_account_session_create' agent tool for actual creation.")

			return nil
		},
	}

	cmd.Flags().StringSliceVar(&targets, "targets", nil, "allowed target addresses (comma-separated)")
	cmd.Flags().StringSliceVar(&functions, "functions", nil, "allowed function selectors (comma-separated)")
	cmd.Flags().StringVar(&limit, "limit", "0", "spend limit in ETH")
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

			cfg := boot.Config
			if !cfg.SmartAccount.Enabled {
				return fmt.Errorf("smart account not enabled in config")
			}

			type sessionSummary struct {
				Status     string `json:"status"`
				MaxKeys    int    `json:"maxActiveKeys"`
				MaxDur     string `json:"maxDuration"`
				DefaultGas uint64 `json:"defaultGasLimit"`
			}

			info := sessionSummary{
				Status:     "configured (requires running server for live data)",
				MaxKeys:    cfg.SmartAccount.Session.MaxActiveKeys,
				MaxDur:     cfg.SmartAccount.Session.MaxDuration.String(),
				DefaultGas: cfg.SmartAccount.Session.DefaultGasLimit,
			}

			if output == "json" {
				data, marshalErr := json.MarshalIndent(info, "", "  ")
				if marshalErr != nil {
					return fmt.Errorf("marshal json: %w", marshalErr)
				}
				fmt.Println(string(data))
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tADDRESS\tPARENT\tEXPIRES\tSPEND_LIMIT\tSTATUS")
			fmt.Fprintln(w, "(no live data available)")
			if flushErr := w.Flush(); flushErr != nil {
				return fmt.Errorf("flush output: %w", flushErr)
			}

			fmt.Println()
			fmt.Fprintf(cmd.ErrOrStderr(), "Session config: max_keys=%d, max_duration=%s, default_gas=%d\n",
				info.MaxKeys, info.MaxDur, info.DefaultGas)
			fmt.Fprintln(cmd.ErrOrStderr(), "Note: Live session listing requires a running server (lango serve).")

			return nil
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

			cfg := boot.Config
			if !cfg.SmartAccount.Enabled {
				return fmt.Errorf("smart account not enabled in config")
			}

			if !all && len(args) == 0 {
				return fmt.Errorf("provide a session ID or use --all to revoke all sessions")
			}

			if all {
				fmt.Println("Revoking all session keys...")
			} else {
				fmt.Printf("Revoking session key: %s\n", args[0])
			}

			fmt.Println()
			fmt.Println("Note: Session revocation requires a running server (lango serve).")
			fmt.Println("Use the agent tool 'smart_account_session_revoke' for actual revocation.")

			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "revoke all active session keys")
	return cmd
}
