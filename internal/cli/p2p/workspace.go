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
		Long: `Inspect the truth-aligned workspace operator surface for the running P2P runtime.

Workspaces are real runtime structures for shared context, code exchange, and
GossipSub messaging. The current CLI mostly points operators to the running
server and agent/tool-backed flows instead of providing full live control.`,
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
		Long:  "Describe how to create a runtime-backed P2P workspace with a name and optional goal.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			if !boot.Config.P2P.Enabled {
				return errP2PDisabled
			}
			if !boot.Config.P2P.Workspace.Enabled {
				return fmt.Errorf("P2P workspace is not enabled (set p2p.workspace.enabled = true)")
			}

			name := args[0]
			result := map[string]interface{}{
				"name":   name,
				"goal":   goal,
				"status": "Use 'lango serve' and create workspaces via the runtime API or agent tools",
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			fmt.Printf("Workspace creation requires a running server.\n")
			fmt.Printf("Start the server with 'lango serve' and use the runtime API or agent tools.\n")
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
		Long:  "Describe how to inspect runtime-backed P2P collaborative workspaces.",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			if !boot.Config.P2P.Enabled {
				return errP2PDisabled
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode([]interface{}{})
			}

			fmt.Println("No workspaces found.")
			fmt.Println()
			fmt.Println("Workspaces are runtime structures managed by the running server.")
			fmt.Println("Start the server with 'lango serve' and use the runtime API or p2p_workspace_* tools.")
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
		Long:  "Explain how to inspect a runtime-backed P2P workspace including members and contributions.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			if !boot.Config.P2P.Enabled {
				return errP2PDisabled
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
			fmt.Println("Workspaces are runtime structures.")
			fmt.Println("Use the running server plus workspace runtime integrations or agent tools for inspection.")
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
		Long:  "Describe how to join an existing runtime-backed P2P collaborative workspace.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			if !boot.Config.P2P.Enabled {
				return errP2PDisabled
			}

			_ = args[0] // workspaceID
			fmt.Println("Joining a workspace requires a running server.")
			fmt.Println("Use 'lango serve' and the runtime API or p2p_workspace_join tool.")
			return nil
		},
	}

	return cmd
}

func newWorkspaceLeaveCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "leave <workspace-id>",
		Short: "Leave a workspace",
		Long:  "Describe how to leave a runtime-backed P2P collaborative workspace.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			if !boot.Config.P2P.Enabled {
				return errP2PDisabled
			}

			_ = args[0] // workspaceID
			fmt.Println("Leaving a workspace requires a running server.")
			fmt.Println("Use 'lango serve' and the runtime API or p2p_workspace_leave tool.")
			return nil
		},
	}

	return cmd
}
