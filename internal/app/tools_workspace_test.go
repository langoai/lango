package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/p2p/workspace"
)

func newWorkspaceToolComponents(t *testing.T) *wsComponents {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "workspaces.db")
	db, err := bolt.Open(dbPath, 0o600, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	manager, err := workspace.NewManager(workspace.ManagerConfig{
		DB:            db,
		LocalDID:      "did:lango:test-local",
		MaxWorkspaces: 10,
		Logger:        zap.NewNop().Sugar(),
	})
	require.NoError(t, err)

	return &wsComponents{manager: manager}
}

func findWorkspaceTool(tools []*agent.Tool, name string) *agent.Tool {
	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}
	return nil
}

func TestWorkspaceCreateSchema_KeepsGoalOptional(t *testing.T) {
	t.Parallel()

	tool := findWorkspaceTool(buildWorkspaceTools(newWorkspaceToolComponents(t)), "p2p_workspace_create")
	require.NotNil(t, tool)

	required, _ := tool.Parameters["required"].([]string)
	assert.Equal(t, []string{"name"}, required)
	assert.NotContains(t, required, "goal")
}

func TestWorkspaceCreate_AllowsMissingGoal(t *testing.T) {
	t.Parallel()

	tool := findWorkspaceTool(buildWorkspaceTools(newWorkspaceToolComponents(t)), "p2p_workspace_create")
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"name": "research-project",
	})
	require.NoError(t, err)

	payload, ok := got.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "research-project", payload["name"])
	assert.Equal(t, "", payload["goal"])
}

func TestWorkspaceTools_RequireCanonicalInputs(t *testing.T) {
	t.Parallel()

	tools := buildWorkspaceTools(newWorkspaceToolComponents(t))

	testCases := []struct {
		name    string
		tool    string
		params  map[string]interface{}
		wantErr string
	}{
		{name: "join requires workspace id", tool: "p2p_workspace_join", params: map[string]interface{}{}, wantErr: "missing workspaceId parameter"},
		{name: "leave requires workspace id", tool: "p2p_workspace_leave", params: map[string]interface{}{}, wantErr: "missing workspaceId parameter"},
		{name: "status requires workspace id", tool: "p2p_workspace_status", params: map[string]interface{}{}, wantErr: "missing workspaceId parameter"},
		{name: "post requires workspace id", tool: "p2p_workspace_post", params: map[string]interface{}{"content": "hello"}, wantErr: "missing workspaceId parameter"},
		{name: "post requires content", tool: "p2p_workspace_post", params: map[string]interface{}{"workspaceId": "ws-1"}, wantErr: "missing content parameter"},
		{name: "read requires workspace id", tool: "p2p_workspace_read", params: map[string]interface{}{}, wantErr: "missing workspaceId parameter"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tool := findWorkspaceTool(tools, tc.tool)
			require.NotNil(t, tool)

			got, err := tool.Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}

func TestWorkspaceGitTools_RequireCanonicalInputs(t *testing.T) {
	t.Parallel()

	tools := buildGitTools(newWorkspaceToolComponents(t))

	testCases := []struct {
		name    string
		tool    string
		params  map[string]interface{}
		wantErr string
	}{
		{name: "git init requires workspace id", tool: "p2p_git_init", params: map[string]interface{}{}, wantErr: "missing workspaceId parameter"},
		{name: "git push requires workspace id", tool: "p2p_git_push", params: map[string]interface{}{}, wantErr: "missing workspaceId parameter"},
		{name: "git log requires workspace id", tool: "p2p_git_log", params: map[string]interface{}{}, wantErr: "missing workspaceId parameter"},
		{name: "git diff requires workspace id", tool: "p2p_git_diff", params: map[string]interface{}{"from": "abc", "to": "def"}, wantErr: "missing workspaceId parameter"},
		{name: "git diff requires from", tool: "p2p_git_diff", params: map[string]interface{}{"workspaceId": "ws-1", "to": "def"}, wantErr: "missing from parameter"},
		{name: "git diff requires to", tool: "p2p_git_diff", params: map[string]interface{}{"workspaceId": "ws-1", "from": "abc"}, wantErr: "missing to parameter"},
		{name: "git leaves requires workspace id", tool: "p2p_git_leaves", params: map[string]interface{}{}, wantErr: "missing workspaceId parameter"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tool := findWorkspaceTool(tools, tc.tool)
			require.NotNil(t, tool)

			got, err := tool.Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}

func TestWorkspaceReadOnlyTools_CapabilityMetadataMatchesInspection(t *testing.T) {
	t.Parallel()

	tools := buildWorkspaceTools(newWorkspaceToolComponents(t))

	listTool := findWorkspaceTool(tools, "p2p_workspace_list")
	require.NotNil(t, listTool)
	assert.Equal(t, agent.ActivityQuery, listTool.Capability.Activity)
	assert.True(t, listTool.Capability.ReadOnly)

	statusTool := findWorkspaceTool(tools, "p2p_workspace_status")
	require.NotNil(t, statusTool)
	assert.Equal(t, agent.ActivityQuery, statusTool.Capability.Activity)
	assert.True(t, statusTool.Capability.ReadOnly)

	readTool := findWorkspaceTool(tools, "p2p_workspace_read")
	require.NotNil(t, readTool)
	assert.Equal(t, agent.ActivityRead, readTool.Capability.Activity)
	assert.True(t, readTool.Capability.ReadOnly)
}

func TestWorkspaceGitReadTools_CapabilityMetadataMatchesInspection(t *testing.T) {
	t.Parallel()

	tools := buildGitTools(newWorkspaceToolComponents(t))

	logTool := findWorkspaceTool(tools, "p2p_git_log")
	require.NotNil(t, logTool)
	assert.Equal(t, agent.ActivityRead, logTool.Capability.Activity)
	assert.True(t, logTool.Capability.ReadOnly)

	diffTool := findWorkspaceTool(tools, "p2p_git_diff")
	require.NotNil(t, diffTool)
	assert.Equal(t, agent.ActivityRead, diffTool.Capability.Activity)
	assert.True(t, diffTool.Capability.ReadOnly)

	leavesTool := findWorkspaceTool(tools, "p2p_git_leaves")
	require.NotNil(t, leavesTool)
	assert.Equal(t, agent.ActivityRead, leavesTool.Capability.Activity)
	assert.True(t, leavesTool.Capability.ReadOnly)
}
