package team

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamToolsFormReturnsTeamSummary(t *testing.T) {
	coord, _ := setupCoordinator(t)
	tool := findToolByName(BuildTools(coord), "team_form")
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"name":        "research",
		"goal":        "collect references",
		"capability":  "search",
		"memberCount": float64(2),
		"leaderDid":   "did:leader",
	})

	require.NoError(t, err)
	result := got.(map[string]interface{})
	assert.NotEmpty(t, result["teamId"])
	assert.Equal(t, "research", result["name"])
	assert.Equal(t, "collect references", result["goal"])
	assert.Equal(t, string(StatusActive), result["status"])
	assert.NotEmpty(t, result["createdAt"])

	members := result["members"].([]map[string]interface{})
	require.GreaterOrEqual(t, len(members), 2)
	var foundLeader bool
	for _, member := range members {
		if member["did"] == "did:leader" {
			foundLeader = true
			assert.Equal(t, string(RoleLeader), member["role"])
		}
	}
	assert.True(t, foundLeader)
}

func TestTeamToolsDelegateIncludesConflictErrorAndIndividualResults(t *testing.T) {
	coord, _ := setupCoordinator(t)
	coord.resolver = func(_ []TaskResult) (map[string]interface{}, error) {
		return nil, fmt.Errorf("conflicting answers")
	}

	teamTool := findToolByName(BuildTools(coord), "team_form")
	formed, err := teamTool.Handler(context.Background(), map[string]interface{}{
		"name":        "research",
		"goal":        "collect references",
		"capability":  "search",
		"memberCount": float64(2),
		"leaderDid":   "did:leader",
	})
	require.NoError(t, err)
	teamID := formed.(map[string]interface{})["teamId"].(string)

	delegateTool := findToolByName(BuildTools(coord), "team_delegate")
	got, err := delegateTool.Handler(context.Background(), map[string]interface{}{
		"teamId":   teamID,
		"toolName": "web_search",
		"params": map[string]interface{}{
			"q": "lango",
		},
	})

	require.NoError(t, err)
	result := got.(map[string]interface{})
	assert.Equal(t, teamID, result["teamId"])
	assert.Equal(t, "web_search", result["toolName"])
	assert.Equal(t, "conflicting answers", result["conflictError"])
	assert.NotContains(t, result, "resolvedResult")

	individual := result["individualResults"].([]map[string]interface{})
	require.Len(t, individual, 2)
	assert.Contains(t, individual[0], "result")
	assert.Contains(t, individual[0], "duration")
}

func TestTeamToolsStatusListAndDisbandValidateTeamLifecycle(t *testing.T) {
	coord, _ := setupCoordinator(t)
	tools := BuildTools(coord)
	formTool := findToolByName(tools, "team_form")
	formed, err := formTool.Handler(context.Background(), map[string]interface{}{
		"name":        "research",
		"goal":        "collect references",
		"capability":  "search",
		"memberCount": float64(1),
		"leaderDid":   "did:leader",
	})
	require.NoError(t, err)
	teamID := formed.(map[string]interface{})["teamId"].(string)

	status, err := findToolByName(tools, "team_status").Handler(context.Background(), map[string]interface{}{
		"teamId": teamID,
	})
	require.NoError(t, err)
	statusMap := status.(map[string]interface{})
	assert.Equal(t, teamID, statusMap["teamId"])
	assert.Equal(t, "did:leader", statusMap["leaderDid"])

	list, err := findToolByName(tools, "team_list").Handler(context.Background(), nil)
	require.NoError(t, err)
	listMap := list.(map[string]interface{})
	assert.Equal(t, 1, listMap["count"])

	disbanded, err := findToolByName(tools, "team_disband").Handler(context.Background(), map[string]interface{}{
		"teamId": teamID,
	})
	require.NoError(t, err)
	assert.Equal(t, teamID, disbanded.(map[string]interface{})["disbanded"])

	_, err = findToolByName(tools, "team_status").Handler(context.Background(), map[string]interface{}{
		"teamId": teamID,
	})
	require.ErrorIs(t, err, ErrTeamNotFound)
}
