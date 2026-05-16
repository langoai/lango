package p2p

import (
	"fmt"

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
server plus the concrete workspace tool surface (p2p_workspace_create,
p2p_workspace_join, p2p_workspace_leave, p2p_workspace_list,
p2p_workspace_status, p2p_workspace_read) instead of providing full live
control.`,
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
		goal   string
		output string
	)

	cmd := &cobra.Command{
		Use:           "create <name>",
		Short:         "Create a new workspace",
		Long:          "Describe how to create a runtime-backed P2P workspace with a name and optional goal.",
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
				return errP2PDisabled
			}
			if !boot.Config.P2P.Workspace.Enabled {
				return fmt.Errorf("P2P workspace is not enabled (set p2p.workspace.enabled = true)")
			}

			name := args[0]
			result := map[string]interface{}{
				"name":   name,
				"goal":   goal,
				"status": "Use 'lango serve' and create workspaces via p2p_workspace_create.",
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), result)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Workspace creation requires a running server.")
			fmt.Fprintln(cmd.OutOrStdout(), "Start the server with 'lango serve' and use p2p_workspace_create.")
			fmt.Fprintf(cmd.OutOrStdout(), "\nExample: p2p_workspace_create name=%q goal=%q\n", name, goal)
			return nil
		},
	}

	cmd.Flags().StringVar(&goal, "goal", "", "Workspace goal/description")
	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}

func newWorkspaceListCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List workspaces",
		Long:          "Describe how to inspect runtime-backed P2P collaborative workspaces.",
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
				return errP2PDisabled
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), []interface{}{})
			}

			fmt.Fprintln(cmd.OutOrStdout(), "No workspaces found.")
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout(), "Workspaces are runtime structures managed by the running server.")
			fmt.Fprintln(cmd.OutOrStdout(), "Start the server with 'lango serve' and use p2p_workspace_list, p2p_workspace_create, or p2p_workspace_join.")
			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	return cmd
}

func newWorkspaceStatusCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "status <workspace-id>",
		Short:         "Show workspace details",
		Long:          "Explain how to inspect a runtime-backed P2P workspace including members and contributions.",
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
				return errP2PDisabled
			}

			_ = args[0] // workspaceID

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), map[string]interface{}{
					"error": "workspace not found (workspaces are runtime-only)",
				})
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Workspace not found.")
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout(), "Workspaces are runtime structures.")
			fmt.Fprintln(cmd.OutOrStdout(), "Use the running server plus the p2p_workspace_status or p2p_workspace_read tools for inspection.")
			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
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
			fmt.Fprintln(cmd.OutOrStdout(), "Joining a workspace requires a running server.")
			fmt.Fprintln(cmd.OutOrStdout(), "Use 'lango serve' and the server-backed runtime or p2p_workspace_join tool.")
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
			fmt.Fprintln(cmd.OutOrStdout(), "Leaving a workspace requires a running server.")
			fmt.Fprintln(cmd.OutOrStdout(), "Use 'lango serve' and the server-backed runtime or p2p_workspace_leave tool.")
			return nil
		},
	}

	return cmd
}
