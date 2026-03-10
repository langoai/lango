package p2p

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

func newWorkspaceCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage P2P collaborative workspaces",
		Long: `Create, join, and manage collaborative workspaces where agents share code and messages.

Workspaces provide a shared context for P2P agent collaboration with
git-based code sharing and GossipSub messaging.`,
	}

	cmd.AddCommand(newWorkspaceCreateCmd(bootLoader))
	cmd.AddCommand(newWorkspaceListCmd(bootLoader))
	cmd.AddCommand(newWorkspaceStatusCmd(bootLoader))
	cmd.AddCommand(newWorkspaceJoinCmd(bootLoader))
	cmd.AddCommand(newWorkspaceLeaveCmd(bootLoader))

	return cmd
}

func newWorkspaceCreateCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		goal       string
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new workspace",
		Long:  "Create a new P2P collaborative workspace with a name and optional goal.",
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
			if !boot.Config.P2P.Workspace.Enabled {
				return fmt.Errorf("P2P workspace is not enabled (set p2p.workspace.enabled = true)")
			}

			name := args[0]
			result := map[string]interface{}{
				"name":   name,
				"goal":   goal,
				"status": "Use 'lango serve' and create workspaces via the agent API",
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			fmt.Printf("Workspace creation requires a running server.\n")
			fmt.Printf("Start the server with 'lango serve' and use the agent tools.\n")
			fmt.Printf("\nExample: p2p_workspace_create name=%q goal=%q\n", name, goal)
			return nil
		},
	}

	cmd.Flags().StringVar(&goal, "goal", "", "Workspace goal/description")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func newWorkspaceListCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workspaces",
		Long:  "List all P2P collaborative workspaces.",
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

			fmt.Println("No workspaces found.")
			fmt.Println()
			fmt.Println("Workspaces are runtime structures managed via the agent API.")
			fmt.Println("Start the server with 'lango serve' and use p2p_workspace_create.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func newWorkspaceStatusCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status <workspace-id>",
		Short: "Show workspace details",
		Long:  "Show detailed information about a P2P workspace including members and contributions.",
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

			_ = args[0] // workspaceID

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]interface{}{
					"error": "workspace not found (workspaces are runtime-only)",
				})
			}

			fmt.Println("Workspace not found.")
			fmt.Println()
			fmt.Println("Workspaces are runtime structures. Use the server API for inspection.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func newWorkspaceJoinCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join <workspace-id>",
		Short: "Join a workspace",
		Long:  "Join an existing P2P collaborative workspace.",
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

			_ = args[0] // workspaceID
			fmt.Println("Joining a workspace requires a running server.")
			fmt.Println("Use 'lango serve' and the p2p_workspace_join tool.")
			return nil
		},
	}

	return cmd
}

func newWorkspaceLeaveCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "leave <workspace-id>",
		Short: "Leave a workspace",
		Long:  "Leave a P2P collaborative workspace.",
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

			_ = args[0] // workspaceID
			fmt.Println("Leaving a workspace requires a running server.")
			fmt.Println("Use 'lango serve' and the p2p_workspace_leave tool.")
			return nil
		},
	}

	return cmd
}
