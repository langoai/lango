package p2p

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

func newGitCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git",
		Short: "Manage P2P git bundles",
		Long: `Manage git bundle exchange for P2P workspace code sharing.

Git bundles allow agents to share code changes without a central git server.
Each workspace has a bare git repository for storing shared commits.`,
	}

	cmd.AddCommand(newGitInitCmd(bootLoader))
	cmd.AddCommand(newGitLogCmd(bootLoader))
	cmd.AddCommand(newGitDiffCmd(bootLoader))
	cmd.AddCommand(newGitPushCmd(bootLoader))
	cmd.AddCommand(newGitFetchCmd(bootLoader))

	return cmd
}

func newGitInitCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <workspace-id>",
		Short: "Initialize git repo for a workspace",
		Long:  "Initialize a bare git repository for code sharing in a P2P workspace.",
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
			fmt.Println("Git init requires a running server.")
			fmt.Println("Use 'lango serve' and the p2p_git_init tool.")
			return nil
		},
	}
	return cmd
}

func newGitLogCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		limit      int
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "log <workspace-id>",
		Short: "Show commit log",
		Long:  "Show the commit log for a workspace's git repository.",
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
			_ = limit

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode([]interface{}{})
			}

			fmt.Println("No commits found.")
			fmt.Println("Git operations require a running server with workspace enabled.")
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum commits to show")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func newGitDiffCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <workspace-id> <from> <to>",
		Short: "Show diff between commits",
		Long:  "Show the diff between two commits in a workspace repository.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.DBClient.Close()

			if !boot.Config.P2P.Enabled {
				return fmt.Errorf("P2P networking is not enabled (set p2p.enabled = true)")
			}

			fmt.Println("Diff requires a running server. Use 'lango serve' and the p2p_git_diff tool.")
			return nil
		},
	}
	return cmd
}

func newGitPushCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push <workspace-id>",
		Short: "Push git bundle to peers",
		Long:  "Create and push a git bundle to workspace peers.",
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

			fmt.Println("Push requires a running server. Use 'lango serve' and the p2p_git_push tool.")
			return nil
		},
	}
	return cmd
}

func newGitFetchCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch <workspace-id>",
		Short: "Fetch git bundle from peers",
		Long:  "Fetch the latest git bundle from workspace peers.",
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

			fmt.Println("Fetch requires a running server. Use 'lango serve' and the agent tools.")
			return nil
		},
	}
	return cmd
}
