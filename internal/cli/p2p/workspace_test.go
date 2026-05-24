package p2p

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.etcd.io/bbolt"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	workspacepkg "github.com/langoai/lango/internal/p2p/workspace"
)

func TestWorkspaceCommands_ManageLocalPersistentWorkspace(t *testing.T) {
	cfg, bootLoader := workspaceCommandTestBoot(t)

	createOut, err := executeP2PCmd(t, newWorkspaceCreateCmd(bootLoader), "research-project", "--goal", "shared context", "--output", "json")
	require.NoError(t, err)

	var created map[string]any
	require.NoError(t, json.Unmarshal([]byte(createOut), &created))
	workspaceID, ok := created["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, workspaceID)
	assert.Equal(t, "research-project", created["name"])
	assert.Equal(t, "shared context", created["goal"])
	assert.Equal(t, "forming", created["status"])
	assert.Equal(t, float64(1), created["memberCount"])

	listOut, err := executeP2PCmd(t, newWorkspaceListCmd(bootLoader), "--output", "json")
	require.NoError(t, err)

	var listed map[string]any
	require.NoError(t, json.Unmarshal([]byte(listOut), &listed))
	require.Equal(t, float64(1), listed["count"])
	workspaces, ok := listed["workspaces"].([]any)
	require.True(t, ok)
	require.Len(t, workspaces, 1)

	statusOut, err := executeP2PCmd(t, newWorkspaceStatusCmd(bootLoader), workspaceID, "--output", "json")
	require.NoError(t, err)

	var status map[string]any
	require.NoError(t, json.Unmarshal([]byte(statusOut), &status))
	assert.Equal(t, workspaceID, status["id"])
	assert.Equal(t, "research-project", status["name"])
	assert.Equal(t, float64(1), status["memberCount"])

	// A second command invocation opens the same local workspace DB, proving
	// the CLI path uses persisted state instead of in-memory command state.
	tableOut, err := executeP2PCmd(t, newWorkspaceListCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg, LangoDir: filepath.Dir(cfg.P2P.Workspace.DataDir)}, nil
	}))
	require.NoError(t, err)
	assert.Contains(t, tableOut, workspaceID)
	assert.Contains(t, tableOut, "research-project")
}

func TestWorkspaceCommands_MutateLocalMembership(t *testing.T) {
	_, bootLoader := workspaceCommandTestBoot(t)

	createOut, err := executeP2PCmd(t, newWorkspaceCreateCmd(bootLoader), "membership-project", "--output", "json")
	require.NoError(t, err)

	var created map[string]any
	require.NoError(t, json.Unmarshal([]byte(createOut), &created))
	workspaceID, ok := created["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, workspaceID)

	leaveOut, err := executeP2PCmd(t, newWorkspaceLeaveCmd(bootLoader), workspaceID)
	require.NoError(t, err)
	assert.Contains(t, leaveOut, "Left workspace")

	statusAfterLeave, err := executeP2PCmd(t, newWorkspaceStatusCmd(bootLoader), workspaceID, "--output", "json")
	require.NoError(t, err)
	var leftStatus map[string]any
	require.NoError(t, json.Unmarshal([]byte(statusAfterLeave), &leftStatus))
	assert.Equal(t, float64(0), leftStatus["memberCount"])

	joinOut, err := executeP2PCmd(t, newWorkspaceJoinCmd(bootLoader), workspaceID)
	require.NoError(t, err)
	assert.Contains(t, joinOut, "Joined workspace")

	statusAfterJoin, err := executeP2PCmd(t, newWorkspaceStatusCmd(bootLoader), workspaceID, "--output", "json")
	require.NoError(t, err)
	var joinedStatus map[string]any
	require.NoError(t, json.Unmarshal([]byte(statusAfterJoin), &joinedStatus))
	assert.Equal(t, float64(1), joinedStatus["memberCount"])
}

func TestWorkspaceListCmd_TableOutputShowsEmptyState(t *testing.T) {
	_, bootLoader := workspaceCommandTestBoot(t)

	out, err := executeP2PCmd(t, newWorkspaceListCmd(bootLoader))
	require.NoError(t, err)

	assert.Contains(t, out, "No local workspaces found.")
}

func TestWorkspaceStatusCmd_TableOutputShowsWorkspaceAndLocalMember(t *testing.T) {
	_, bootLoader := workspaceCommandTestBoot(t)

	createOut, err := executeP2PCmd(t, newWorkspaceCreateCmd(bootLoader), "table-project", "--output", "json")
	require.NoError(t, err)

	var created map[string]any
	require.NoError(t, json.Unmarshal([]byte(createOut), &created))
	workspaceID, ok := created["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, workspaceID)

	statusOut, err := executeP2PCmd(t, newWorkspaceStatusCmd(bootLoader), workspaceID)
	require.NoError(t, err)

	assert.Contains(t, statusOut, "Workspace")
	assert.Contains(t, statusOut, "ID:")
	assert.Contains(t, statusOut, "Members:")
	assert.Contains(t, statusOut, "did:lango:local-cli  creator")
}

func TestWorkspaceOpenLocalManagerRequiresBootstrapConfig(t *testing.T) {
	manager, err := openLocalWorkspaceCLIManager(nil)
	require.Error(t, err)
	assert.Nil(t, manager)
	assert.EqualError(t, err, "load config: missing bootstrap config")

	manager, err = openLocalWorkspaceCLIManager(&bootstrap.Result{})
	require.Error(t, err)
	assert.Nil(t, manager)
	assert.EqualError(t, err, "load config: missing bootstrap config")
}

func TestWorkspaceViewSkipsNilMembers(t *testing.T) {
	joinedAt := time.Unix(100, 0).UTC()
	ws := &workspacepkg.Workspace{
		ID:     "workspace-nil-members",
		Name:   "nil-members",
		Status: workspacepkg.StatusForming,
		Members: []*workspacepkg.Member{
			nil,
			{
				DID:      "did:lango:member",
				Name:     "member",
				Role:     workspacepkg.RoleMember,
				JoinedAt: joinedAt,
			},
		},
		CreatedAt: joinedAt,
		UpdatedAt: joinedAt,
	}

	view := workspaceView(ws)

	assert.Equal(t, 1, view["memberCount"])
	members, ok := view["members"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, members, 1)
	assert.Equal(t, "did:lango:member", members[0]["did"])
	assert.Equal(t, workspacepkg.RoleMember, members[0]["role"])
}

func TestWorkspaceCommands_RequireFeatureGatesBeforeOpeningStore(t *testing.T) {
	cfg, bootLoader := workspaceCommandTestBoot(t)
	dataDir := cfg.P2P.Workspace.DataDir

	cfg.P2P.Enabled = false
	out, err := executeP2PCmd(t, newWorkspaceCreateCmd(bootLoader), "blocked-project")
	require.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, err.Error(), "P2P networking is not enabled")
	assert.NoDirExists(t, dataDir)

	cfg.P2P.Enabled = true
	cfg.P2P.Workspace.Enabled = false
	out, err = executeP2PCmd(t, newWorkspaceListCmd(bootLoader))
	require.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, err.Error(), "P2P workspace is not enabled")
	assert.NoDirExists(t, dataDir)
}

func TestWorkspaceCommands_ReturnErrorWhenWorkspaceDBIsLocked(t *testing.T) {
	cfg, bootLoader := workspaceCommandTestBoot(t)
	require.NoError(t, os.MkdirAll(cfg.P2P.Workspace.DataDir, 0o700))

	dbPath := filepath.Join(cfg.P2P.Workspace.DataDir, "workspaces.db")
	lockedDB, err := bbolt.Open(dbPath, 0o600, &bbolt.Options{Timeout: time.Second})
	require.NoError(t, err)
	defer lockedDB.Close()

	originalTimeout := workspaceDBOpenTimeout
	workspaceDBOpenTimeout = 10 * time.Millisecond
	t.Cleanup(func() { workspaceDBOpenTimeout = originalTimeout })

	started := time.Now()
	out, err := executeP2PCmd(t, newWorkspaceListCmd(bootLoader))
	require.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, err.Error(), "open workspace db")
	assert.Less(t, time.Since(started), 500*time.Millisecond)
}

func workspaceCommandTestBoot(t *testing.T) (*config.Config, func() (*bootstrap.Result, error)) {
	t.Helper()

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.Workspace.Enabled = true
	langoDir := t.TempDir()
	cfg.P2P.Workspace.DataDir = filepath.Join(langoDir, "workspaces")

	return cfg, func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg, LangoDir: langoDir}, nil
	}
}

func TestWorkspaceCmd_HelpUsesConcreteToolSurfaceWording(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.Workspace.Enabled = true

	cmd := newWorkspaceCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg}, nil
	})
	out, err := executeP2PCmd(t, cmd, "--help")
	require.NoError(t, err)

	assert.Contains(t, out, "p2p_workspace_create")
	assert.Contains(t, out, "p2p_workspace_join")
	assert.Contains(t, out, "p2p_workspace_leave")
	assert.Contains(t, out, "p2p_workspace_list")
	assert.Contains(t, out, "p2p_workspace_status")
	assert.Contains(t, out, "p2p_workspace_read")
	assert.NotContains(t, out, "agent/tool-backed flows")
}

func TestWorkspaceCreateCmd_WritesJSONToCommandWriter(t *testing.T) {
	_, bootLoader := workspaceCommandTestBoot(t)

	cmd := newWorkspaceCreateCmd(bootLoader)
	out, err := executeP2PCmd(t, cmd, "research-project", "--goal", "shared context", "--output", "json")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "research-project", decoded["name"])
	assert.Equal(t, "shared context", decoded["goal"])
	assert.NotEmpty(t, decoded["id"])
}

func TestWorkspaceListCmd_WritesJSONToCommandWriter(t *testing.T) {
	_, bootLoader := workspaceCommandTestBoot(t)

	cmd := newWorkspaceListCmd(bootLoader)
	out, err := executeP2PCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(0), decoded["count"])
	assert.Empty(t, decoded["workspaces"])
}

func TestWorkspaceStatusCmd_WritesJSONToCommandWriter(t *testing.T) {
	_, bootLoader := workspaceCommandTestBoot(t)
	cmd := newWorkspaceStatusCmd(bootLoader)
	out, err := executeP2PCmd(t, cmd, "workspace-123", "--output", "json")
	require.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, err.Error(), "workspace-123")
}
