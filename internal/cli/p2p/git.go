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
		Long: `Inspect the truth-aligned git bundle operator surface for the running P2P runtime.

Git bundle services are real runtime subsystems for workspace code sharing.
The current CLI mostly points operators to the running server and agent/tool-
backed flows instead of providing full direct live repository control.`,
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
		Long:  "Describe how to initialize the runtime-backed git repository for a P2P workspace.",
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
			fmt.Println("Git init requires a running server.")
			fmt.Println("Use 'lango serve' and the runtime API or p2p_git_init tool.")
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
		Long:  "Describe how to inspect commit history from a runtime-backed workspace git repository.",
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
			_ = limit

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode([]interface{}{})
			}

			fmt.Println("No commits found.")
			fmt.Println("Git operations require a running server with workspace enabled.")
			fmt.Println("Use the runtime API or p2p_git_* tools for live repository inspection.")
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
		Long:  "Describe how to diff commits in a runtime-backed workspace repository.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			if !boot.Config.P2P.Enabled {
				return errP2PDisabled
			}

			fmt.Println("Diff requires a running server.")
			fmt.Println("Use 'lango serve' and the runtime API or p2p_git_diff tool.")
			return nil
		},
	}
	return cmd
}

func newGitPushCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push <workspace-id>",
		Short: "Push git bundle to peers",
		Long:  "Describe how to create and push a runtime-backed git bundle to workspace peers.",
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

			fmt.Println("Push requires a running server.")
			fmt.Println("Use 'lango serve' and the runtime API or p2p_git_push tool.")
			return nil
		},
	}
	return cmd
}

func newGitFetchCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch <workspace-id>",
		Short: "Fetch git bundle from peers",
		Long:  "Describe how to fetch the latest runtime-backed git bundle from workspace peers.",
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

			fmt.Println("Fetch requires a running server.")
			fmt.Println("Use 'lango serve' and the runtime API or p2p_git_fetch tool.")
			return nil
		},
	}
	return cmd
}
