package app

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/p2p/workspace"
)

func newWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceToolComponents(t *testing.T, maxWorkspaces int) *wsComponents {
	t.Helper()

	db, err := bolt.Open(filepath.Join(t.TempDir(), "workspaces.db"), 0o600, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	manager, err := workspace.NewManager(workspace.ManagerConfig{
		DB:            db,
		LocalDID:      "did:lango:runChatBuilderErrorRunsStartupCleanupWithoutTui9-local",
		MaxWorkspaces: maxWorkspaces,
		Logger:        zap.NewNop().Sugar(),
	})
	require.NoError(t, err)

	return &wsComponents{
		manager:  manager,
		tracker:  workspace.NewContributionTracker(),
		db:       db,
		localDID: "did:lango:runChatBuilderErrorRunsStartupCleanupWithoutTui9-local",
	}
}

func workspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceToolsByName(t *testing.T, wc *wsComponents) map[string]*agent.Tool {
	t.Helper()

	tools := make(map[string]*agent.Tool)
	for _, tool := range buildWorkspaceTools(wc) {
		tools[tool.Name] = tool
	}
	return tools
}

func requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceTool(t *testing.T, tools map[string]*agent.Tool, name string) *agent.Tool {
	t.Helper()

	tool := tools[name]
	require.NotNil(t, tool)
	return tool
}

func requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PMapPayload(t *testing.T, got interface{}) map[string]interface{} {
	t.Helper()

	payload, ok := got.(map[string]interface{})
	require.True(t, ok)
	return payload
}

func requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PStringField(t *testing.T, payload map[string]interface{}, key string) string {
	t.Helper()

	value, ok := payload[key].(string)
	require.True(t, ok)
	return value
}

func TestWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2P(t *testing.T) {
	t.Parallel()

	wc := newWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceToolComponents(t, 4)
	tools := workspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceToolsByName(t, wc)

	createTool := requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceTool(t, tools, "p2p_workspace_create")
	listTool := requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceTool(t, tools, "p2p_workspace_list")
	statusTool := requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceTool(t, tools, "p2p_workspace_status")
	leaveTool := requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceTool(t, tools, "p2p_workspace_leave")

	created, err := createTool.Handler(context.Background(), map[string]interface{}{
		"name": "runChatBuilderErrorRunsStartupCleanupWithoutTui9-workspace",
		"goal": "Exercise deterministic workspace tool branches",
	})
	require.NoError(t, err)
	createPayload := requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PMapPayload(t, created)
	workspaceID := requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PStringField(t, createPayload, "id")
	require.NotEmpty(t, workspaceID)
	assert.Equal(t, "runChatBuilderErrorRunsStartupCleanupWithoutTui9-workspace", createPayload["name"])
	assert.Equal(t, "Exercise deterministic workspace tool branches", createPayload["goal"])
	assert.Equal(t, string(workspace.StatusForming), createPayload["status"])
	_, err = time.Parse(time.RFC3339, requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PStringField(t, createPayload, "createdAt"))
	require.NoError(t, err)

	wc.tracker.RecordCommit(workspaceID, wc.localDID, 128)
	wc.tracker.RecordMessage(workspaceID, wc.localDID)

	listed, err := listTool.Handler(context.Background(), nil)
	require.NoError(t, err)
	listPayload := requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PMapPayload(t, listed)
	assert.Equal(t, 1, listPayload["count"])
	workspaces, ok := listPayload["workspaces"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, workspaces, 1)
	assert.Equal(t, workspaceID, workspaces[0]["id"])
	assert.Equal(t, "runChatBuilderErrorRunsStartupCleanupWithoutTui9-workspace", workspaces[0]["name"])
	assert.Equal(t, 1, workspaces[0]["members"])

	status, err := statusTool.Handler(context.Background(), map[string]interface{}{
		"workspaceId": workspaceID,
	})
	require.NoError(t, err)
	statusPayload := requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PMapPayload(t, status)
	assert.Equal(t, workspaceID, statusPayload["id"])
	assert.Equal(t, "runChatBuilderErrorRunsStartupCleanupWithoutTui9-workspace", statusPayload["name"])
	assert.Equal(t, string(workspace.StatusForming), statusPayload["status"])

	members, ok := statusPayload["members"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, members, 1)
	assert.Equal(t, wc.localDID, members[0]["did"])
	assert.Equal(t, workspace.RoleCreator, members[0]["role"])
	_, err = time.Parse(time.RFC3339, requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PStringField(t, members[0], "joinedAt"))
	require.NoError(t, err)

	contributions, ok := statusPayload["contributions"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, contributions, 1)
	assert.Equal(t, wc.localDID, contributions[0]["did"])
	assert.Equal(t, 1, contributions[0]["commits"])
	assert.Equal(t, int64(128), contributions[0]["codeBytes"])
	assert.Equal(t, 1, contributions[0]["messages"])
	_, err = time.Parse(time.RFC3339, requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PStringField(t, contributions[0], "lastActive"))
	require.NoError(t, err)

	left, err := leaveTool.Handler(context.Background(), map[string]interface{}{
		"workspaceId": workspaceID,
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"left": workspaceID}, left)

	afterLeave, err := statusTool.Handler(context.Background(), map[string]interface{}{
		"workspaceId": workspaceID,
	})
	require.NoError(t, err)
	afterLeavePayload := requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PMapPayload(t, afterLeave)
	afterLeaveMembers, ok := afterLeavePayload["members"].([]map[string]interface{})
	require.True(t, ok)
	assert.Empty(t, afterLeaveMembers)
}

func TestWorkspaceCreateWrapsManagerLimitError(t *testing.T) {
	t.Parallel()

	tools := workspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceToolsByName(t, newWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceToolComponents(t, 1))
	create := requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceTool(t, tools, "p2p_workspace_create")

	_, err := create.Handler(context.Background(), map[string]interface{}{"name": "first"})
	require.NoError(t, err)

	got, err := create.Handler(context.Background(), map[string]interface{}{"name": "second"})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "create workspace: max workspaces reached (1)")
}

func TestWorkspaceJoinAndErrorBranchesWithoutP2P(t *testing.T) {
	t.Parallel()

	wc := newWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceToolComponents(t, 3)
	tools := workspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceToolsByName(t, wc)
	createTool := requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceTool(t, tools, "p2p_workspace_create")
	joinTool := requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceTool(t, tools, "p2p_workspace_join")

	created, err := createTool.Handler(context.Background(), map[string]interface{}{"name": "joinable"})
	require.NoError(t, err)
	workspaceID := requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PStringField(t, requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PMapPayload(t, created), "id")

	joined, err := joinTool.Handler(context.Background(), map[string]interface{}{
		"workspaceId": workspaceID,
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"joined": workspaceID}, joined)

	for _, tc := range []struct {
		name string
		tool string
	}{
		{name: "join unknown workspace", tool: "p2p_workspace_join"},
		{name: "leave unknown workspace", tool: "p2p_workspace_leave"},
		{name: "status unknown workspace", tool: "p2p_workspace_status"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tool := requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceTool(t, tools, tc.tool)
			got, err := tool.Handler(context.Background(), map[string]interface{}{
				"workspaceId": "missing-workspace",
			})
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Contains(t, err.Error(), "workspace missing-workspace")
			assert.Contains(t, err.Error(), "not found")
		})
	}
}

func TestWorkspaceListReturnsEmptyPayloadWithoutNetwork(t *testing.T) {
	t.Parallel()

	tools := workspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceToolsByName(t, newWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceToolComponents(t, 2))
	listTool := requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceTool(t, tools, "p2p_workspace_list")

	got, err := listTool.Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)
	payload := requireWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PMapPayload(t, got)
	assert.Equal(t, 0, payload["count"])
	workspaces, ok := payload["workspaces"].([]map[string]interface{})
	require.True(t, ok)
	assert.Empty(t, workspaces)
}
