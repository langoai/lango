package p2p

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/p2p/workspace"
)

type workspaceCLIManager interface {
	Create(context.Context, workspace.CreateRequest) (*workspace.Workspace, error)
	List(context.Context) ([]*workspace.Workspace, error)
	Get(context.Context, string) (*workspace.Workspace, error)
	Join(context.Context, string) error
	Leave(context.Context, string) error
	Close() error
}

type localWorkspaceCLIManager struct {
	manager *workspace.Manager
	db      *bbolt.DB
}

func (m *localWorkspaceCLIManager) Create(ctx context.Context, req workspace.CreateRequest) (*workspace.Workspace, error) {
	return m.manager.Create(ctx, req)
}

func (m *localWorkspaceCLIManager) List(ctx context.Context) ([]*workspace.Workspace, error) {
	return m.manager.List(ctx)
}

func (m *localWorkspaceCLIManager) Get(ctx context.Context, workspaceID string) (*workspace.Workspace, error) {
	return m.manager.Get(ctx, workspaceID)
}

func (m *localWorkspaceCLIManager) Join(ctx context.Context, workspaceID string) error {
	return m.manager.Join(ctx, workspaceID)
}

func (m *localWorkspaceCLIManager) Leave(ctx context.Context, workspaceID string) error {
	return m.manager.Leave(ctx, workspaceID)
}

func (m *localWorkspaceCLIManager) Close() error {
	if m == nil || m.db == nil {
		return nil
	}
	return m.db.Close()
}

var (
	openWorkspaceCLIManager = openLocalWorkspaceCLIManager
	workspaceDBOpenTimeout  = 2 * time.Second
)

func openLocalWorkspaceCLIManager(boot *bootstrap.Result) (workspaceCLIManager, error) {
	if boot == nil || boot.Config == nil {
		return nil, fmt.Errorf("load config: missing bootstrap config")
	}
	if !boot.Config.P2P.Enabled {
		return nil, errP2PDisabled
	}
	if !boot.Config.P2P.Workspace.Enabled {
		return nil, fmt.Errorf("P2P workspace is not enabled (set p2p.workspace.enabled = true)")
	}

	dataDir := boot.Config.P2P.Workspace.DataDir
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolve workspace data dir: %w", err)
		}
		dataDir = filepath.Join(home, ".lango", "workspaces")
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, fmt.Errorf("create workspace data dir: %w", err)
	}

	db, err := bbolt.Open(filepath.Join(dataDir, "workspaces.db"), 0o600, &bbolt.Options{
		Timeout: workspaceDBOpenTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("open workspace db: %w", err)
	}

	log := logging.Sugar()
	if log == nil {
		log = zap.NewNop().Sugar()
	}
	localDID := resolveIdentityDID(boot)
	if localDID == "" {
		localDID = "did:lango:local-cli"
	}

	manager, err := workspace.NewManager(workspace.ManagerConfig{
		DB:            db,
		LocalDID:      localDID,
		MaxWorkspaces: boot.Config.P2P.Workspace.MaxWorkspaces,
		Logger:        log,
	})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open workspace manager: %w", err)
	}

	return &localWorkspaceCLIManager{manager: manager, db: db}, nil
}

func workspaceView(ws *workspace.Workspace) map[string]interface{} {
	members := make([]map[string]interface{}, 0, len(ws.Members))
	for _, member := range ws.Members {
		if member == nil {
			continue
		}
		members = append(members, map[string]interface{}{
			"did":      member.DID,
			"name":     member.Name,
			"role":     member.Role,
			"joinedAt": member.JoinedAt,
		})
	}

	return map[string]interface{}{
		"id":          ws.ID,
		"name":        ws.Name,
		"goal":        ws.Goal,
		"status":      ws.Status,
		"memberCount": len(members),
		"members":     members,
		"createdAt":   ws.CreatedAt,
		"updatedAt":   ws.UpdatedAt,
	}
}

func newWorkspaceCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage P2P collaborative workspaces",
		Long: `Manage local P2P collaborative workspace records.

These commands persist local workspace lifecycle state through the same
workspace manager used by the runtime. Distributed workspace messaging and peer
	exchange still require the running server tools:
	p2p_workspace_create, p2p_workspace_join, p2p_workspace_leave,
	p2p_workspace_list, p2p_workspace_status, and p2p_workspace_read.`,
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
		Long:          "Create a local P2P workspace record with a name and optional goal.",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedOutput, err := resolveOutput(cmd)
			if err != nil {
				return err
			}

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			manager, err := openWorkspaceCLIManager(boot)
			if err != nil {
				return err
			}
			defer manager.Close()

			name := args[0]
			ws, err := manager.Create(cmd.Context(), workspace.CreateRequest{Name: name, Goal: goal})
			if err != nil {
				return err
			}
			result := workspaceView(ws)

			if resolvedOutput == "json" {
				return printJSON(cmd.OutOrStdout(), result)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Workspace created")
			fmt.Fprintf(cmd.OutOrStdout(), "  ID:      %s\n", ws.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Name:    %s\n", ws.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "  Goal:    %s\n", ws.Goal)
			fmt.Fprintf(cmd.OutOrStdout(), "  Status:  %s\n", ws.Status)
			fmt.Fprintf(cmd.OutOrStdout(), "  Members: %d\n", len(ws.Members))
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
		Long:          "List locally persisted P2P collaborative workspaces.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedOutput, err := resolveOutput(cmd)
			if err != nil {
				return err
			}

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			manager, err := openWorkspaceCLIManager(boot)
			if err != nil {
				return err
			}
			defer manager.Close()

			workspaces, err := manager.List(cmd.Context())
			if err != nil {
				return err
			}
			items := make([]map[string]interface{}, 0, len(workspaces))
			for _, ws := range workspaces {
				items = append(items, workspaceView(ws))
			}
			result := map[string]interface{}{"workspaces": items, "count": len(items)}

			if resolvedOutput == "json" {
				return printJSON(cmd.OutOrStdout(), result)
			}

			if len(workspaces) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No local workspaces found.")
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Local Workspaces")
			for _, ws := range workspaces {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s  %s  %s  members=%d\n", ws.ID, ws.Name, ws.Status, len(ws.Members))
			}
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
		Long:          "Show one locally persisted P2P collaborative workspace including members.",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedOutput, err := resolveOutput(cmd)
			if err != nil {
				return err
			}

			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			manager, err := openWorkspaceCLIManager(boot)
			if err != nil {
				return err
			}
			defer manager.Close()

			ws, err := manager.Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			result := workspaceView(ws)

			if resolvedOutput == "json" {
				return printJSON(cmd.OutOrStdout(), result)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Workspace")
			fmt.Fprintf(cmd.OutOrStdout(), "  ID:      %s\n", ws.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Name:    %s\n", ws.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "  Goal:    %s\n", ws.Goal)
			fmt.Fprintf(cmd.OutOrStdout(), "  Status:  %s\n", ws.Status)
			fmt.Fprintf(cmd.OutOrStdout(), "  Members: %d\n", len(ws.Members))
			for _, member := range ws.Members {
				if member == nil {
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "    %s  %s\n", member.DID, member.Role)
			}
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
		Long:  "Join an existing local P2P collaborative workspace.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			manager, err := openWorkspaceCLIManager(boot)
			if err != nil {
				return err
			}
			defer manager.Close()

			workspaceID := args[0]
			if err := manager.Join(cmd.Context(), workspaceID); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Joined workspace %s\n", workspaceID)
			return nil
		},
	}

	return cmd
}

func newWorkspaceLeaveCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "leave <workspace-id>",
		Short: "Leave a workspace",
		Long:  "Leave a local P2P collaborative workspace.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			manager, err := openWorkspaceCLIManager(boot)
			if err != nil {
				return err
			}
			defer manager.Close()

			workspaceID := args[0]
			if err := manager.Leave(cmd.Context(), workspaceID); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Left workspace %s\n", workspaceID)
			return nil
		},
	}

	return cmd
}
