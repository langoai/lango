package app

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/p2p/workspace"
)

func TestWorkspacePostAndReadMessagesWithoutP2P(t *testing.T) {
	t.Parallel()

	wc := newWorkspaceCreateListStatusAndLeaveRoundTripWithoutP2PWorkspaceToolComponents(t, 2)
	tools := workspacePostAndReadMessagesWithoutP2PToolsByName(t, wc)
	createTool := workspacePostAndReadMessagesWithoutP2PRequireTool(t, tools, "p2p_workspace_create")
	postTool := workspacePostAndReadMessagesWithoutP2PRequireTool(t, tools, "p2p_workspace_post")
	readTool := workspacePostAndReadMessagesWithoutP2PRequireTool(t, tools, "p2p_workspace_read")

	created, err := createTool.Handler(context.Background(), map[string]interface{}{
		"name": "message-workspace",
	})
	require.NoError(t, err)
	workspaceID := workspacePostAndReadMessagesWithoutP2PStringField(t, workspacePostAndReadMessagesWithoutP2PMapPayload(t, created), "id")

	first, err := postTool.Handler(context.Background(), map[string]interface{}{
		"workspaceId": workspaceID,
		"content":     "first knowledge share",
	})
	require.NoError(t, err)
	firstPayload := workspacePostAndReadMessagesWithoutP2PMapPayload(t, first)
	firstID := workspacePostAndReadMessagesWithoutP2PStringField(t, firstPayload, "messageId")
	require.NotEmpty(t, firstID)
	assert.Equal(t, true, firstPayload["posted"])
	assert.Equal(t, workspaceID, firstPayload["workspaceId"])

	second, err := postTool.Handler(context.Background(), map[string]interface{}{
		"workspaceId": workspaceID,
		"type":        string(workspace.MessageTypeTaskProposal),
		"content":     "proposal response",
		"parentId":    firstID,
	})
	require.NoError(t, err)
	secondID := workspacePostAndReadMessagesWithoutP2PStringField(t, workspacePostAndReadMessagesWithoutP2PMapPayload(t, second), "messageId")
	require.NotEmpty(t, secondID)

	readAll, err := readTool.Handler(context.Background(), map[string]interface{}{
		"workspaceId": workspaceID,
		"limit":       10,
	})
	require.NoError(t, err)
	allPayload := workspacePostAndReadMessagesWithoutP2PMapPayload(t, readAll)
	assert.Equal(t, 2, allPayload["count"])
	allMessages := workspacePostAndReadMessagesWithoutP2PMessages(t, allPayload)
	require.Len(t, allMessages, 2)
	allByID := workspacePostAndReadMessagesWithoutP2PMessagesByID(t, allMessages)
	firstMessage := allByID[firstID]
	require.NotNil(t, firstMessage)
	assert.Equal(t, string(workspace.MessageTypeKnowledgeShare), firstMessage["type"])
	assert.Equal(t, "first knowledge share", firstMessage["content"])
	assert.Empty(t, firstMessage["parentId"])
	_, err = time.Parse(time.RFC3339, workspacePostAndReadMessagesWithoutP2PStringField(t, firstMessage, "timestamp"))
	require.NoError(t, err)

	secondMessage := allByID[secondID]
	require.NotNil(t, secondMessage)
	assert.Equal(t, string(workspace.MessageTypeTaskProposal), secondMessage["type"])
	assert.Equal(t, firstID, secondMessage["parentId"])

	readFiltered, err := readTool.Handler(context.Background(), map[string]interface{}{
		"workspaceId": workspaceID,
		"type":        string(workspace.MessageTypeTaskProposal),
	})
	require.NoError(t, err)
	filteredPayload := workspacePostAndReadMessagesWithoutP2PMapPayload(t, readFiltered)
	assert.Equal(t, 1, filteredPayload["count"])
	filteredMessages := workspacePostAndReadMessagesWithoutP2PMessages(t, filteredPayload)
	require.Len(t, filteredMessages, 1)
	assert.Equal(t, secondID, filteredMessages[0]["id"])
	assert.Equal(t, "proposal response", filteredMessages[0]["content"])
	assert.Equal(t, firstID, filteredMessages[0]["parentId"])
}

func workspacePostAndReadMessagesWithoutP2PToolsByName(t *testing.T, wc *wsComponents) map[string]*agent.Tool {
	t.Helper()

	tools := make(map[string]*agent.Tool)
	for _, tool := range buildWorkspaceTools(wc) {
		tools[tool.Name] = tool
	}
	return tools
}

func workspacePostAndReadMessagesWithoutP2PRequireTool(t *testing.T, tools map[string]*agent.Tool, name string) *agent.Tool {
	t.Helper()

	tool := tools[name]
	require.NotNil(t, tool)
	return tool
}

func workspacePostAndReadMessagesWithoutP2PMapPayload(t *testing.T, got interface{}) map[string]interface{} {
	t.Helper()

	payload, ok := got.(map[string]interface{})
	require.True(t, ok)
	return payload
}

func workspacePostAndReadMessagesWithoutP2PMessages(t *testing.T, payload map[string]interface{}) []map[string]interface{} {
	t.Helper()

	messages, ok := payload["messages"].([]map[string]interface{})
	require.True(t, ok)
	return messages
}

func workspacePostAndReadMessagesWithoutP2PMessagesByID(t *testing.T, messages []map[string]interface{}) map[string]map[string]interface{} {
	t.Helper()

	result := make(map[string]map[string]interface{}, len(messages))
	for _, message := range messages {
		id := workspacePostAndReadMessagesWithoutP2PStringField(t, message, "id")
		result[id] = message
	}
	return result
}

func workspacePostAndReadMessagesWithoutP2PStringField(t *testing.T, payload map[string]interface{}, key string) string {
	t.Helper()

	value, ok := payload[key].(string)
	require.True(t, ok)
	return value
}
