package p2p

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

func newTeamCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "team",
		Short: "Manage P2P agent teams",
		Long: `List, inspect, and disband dynamic P2P agent teams.

Teams are runtime-only structures that exist while the lango server is running.
Use the server API for live team management.`,
	}

	cmd.AddCommand(newTeamListCmd(bootLoader))
	cmd.AddCommand(newTeamStatusCmd(bootLoader))
	cmd.AddCommand(newTeamDisbandCmd(bootLoader))

	return cmd
}

func newTeamListCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active P2P teams",
		Long: `List all currently active agent teams on the P2P network.

Note: Teams are runtime-only and exist only while the server is running.
Connect to the running server API to inspect live teams.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.DBClient.Close()

			if !boot.Config.P2P.Enabled {
				return fmt.Errorf("P2P networking is not enabled (set p2p.enabled = true)")
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode([]interface{}{})
			}

			fmt.Println("No active teams.")
			fmt.Println()
			fmt.Println("Teams are runtime-only structures created during agent collaboration.")
			fmt.Println("Start the server with 'lango serve' and form teams via the API.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func newTeamStatusCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status <team-id>",
		Short: "Show team details",
		Long:  "Show detailed information about a specific P2P agent team including members, budget, and status.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.DBClient.Close()

			if !boot.Config.P2P.Enabled {
				return fmt.Errorf("P2P networking is not enabled (set p2p.enabled = true)")
			}

			_ = args[0] // teamID

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]interface{}{
					"error": "team not found (teams are runtime-only)",
				})
			}

			fmt.Println("Team not found.")
			fmt.Println()
			fmt.Println("Teams are runtime-only structures that exist only while the server is running.")
			fmt.Println("Use the server API (GET /api/p2p/teams/<id>) for live team inspection.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func newTeamDisbandCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disband <team-id>",
		Short: "Disband a team",
		Long:  "Disband a P2P agent team, releasing all members.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.DBClient.Close()

			if !boot.Config.P2P.Enabled {
				return fmt.Errorf("P2P networking is not enabled (set p2p.enabled = true)")
			}

			_ = args[0] // teamID

			fmt.Println("Team not found.")
			fmt.Println()
			fmt.Println("Teams are runtime-only structures. Use the server API")
			fmt.Println("(DELETE /api/p2p/teams/<id>) to disband a live team.")
			return nil
		},
	}

	return cmd
}
