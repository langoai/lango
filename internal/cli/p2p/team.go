package p2p

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

func newTeamCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "team",
		Short: "Manage P2P agent teams",
		Long: `Inspect the truth-aligned team operator surface for the running P2P runtime.

Teams are real runtime-only structures that exist while the lango server is running.
The current CLI primarily explains how to use the server-backed runtime plus
the concrete team tool surface (team_form, team_form_with_budget,
team_status, team_list, team_disband) rather than providing full live team
control.`,
	}

	cmd.AddCommand(newTeamListCmd(bootLoader))
	cmd.AddCommand(newTeamStatusCmd(bootLoader))
	cmd.AddCommand(newTeamDisbandCmd(bootLoader))

	return cmd
}

func newTeamListCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List active P2P teams",
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `Describe how to inspect active agent teams in the running P2P runtime.

Note: Teams are runtime-only and exist only while the server is running.
		Use lango serve plus the server-backed runtime and team_list, team_form,
and team_form_with_budget for live teams.`,
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

			if !boot.Config.P2P.Enabled {
				return fmt.Errorf("P2P networking is not enabled (set p2p.enabled = true)")
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), []interface{}{})
			}

			fmt.Fprintln(cmd.OutOrStdout(), "No active teams.")
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout(), "Teams are runtime-only structures created during agent collaboration.")
			fmt.Fprintln(cmd.OutOrStdout(), "Start the server with 'lango serve' and inspect/form teams via team_list, team_form, and team_form_with_budget.")
			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}

func newTeamStatusCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "status <team-id>",
		Short:         "Show team details",
		Long:          "Explain how to inspect a specific runtime-backed P2P agent team, including members, budget, and status.",
		Args:          cobra.ExactArgs(1),
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

			if !boot.Config.P2P.Enabled {
				return fmt.Errorf("P2P networking is not enabled (set p2p.enabled = true)")
			}

			_ = args[0] // teamID

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), map[string]interface{}{
					"error": "team not found (teams are runtime-only)",
				})
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Team not found.")
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout(), "Teams are runtime-only structures that exist only while the server is running.")
			fmt.Fprintln(cmd.OutOrStdout(), "Use the running server plus the team_status tool for live inspection.")
			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}

func newTeamDisbandCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disband <team-id>",
		Short: "Disband a team",
		Long:  "Explain how to disband a runtime-backed P2P agent team and release its members.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			if !boot.Config.P2P.Enabled {
				return fmt.Errorf("P2P networking is not enabled (set p2p.enabled = true)")
			}

			_ = args[0] // teamID

			fmt.Fprintln(cmd.OutOrStdout(), "Team not found.")
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout(), "Teams are runtime-only structures.")
			fmt.Fprintln(cmd.OutOrStdout(), "Use the running server plus the team_disband tool to disband a live team.")
			return nil
		},
	}

	return cmd
}
