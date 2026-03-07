package protocol

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamRequestTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give RequestType
		want string
	}{
		{give: RequestTeamInvite, want: "team_invite"},
		{give: RequestTeamAccept, want: "team_accept"},
		{give: RequestTeamTask, want: "team_task"},
		{give: RequestTeamResult, want: "team_result"},
		{give: RequestTeamDisband, want: "team_disband"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, string(tt.give))
		})
	}
}

func TestTeamInvitePayload_JSON(t *testing.T) {
	t.Parallel()

	p := TeamInvitePayload{
		TeamID:       "t1",
		TeamName:     "search-team",
		Goal:         "find information",
		LeaderDID:    "did:leader:123",
		Role:         "worker",
		Capabilities: []string{"search", "summarize"},
	}

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var decoded TeamInvitePayload
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, p.TeamID, decoded.TeamID)
	assert.Len(t, decoded.Capabilities, 2)
}

func TestTeamTaskPayload_JSON(t *testing.T) {
	t.Parallel()

	p := TeamTaskPayload{
		TeamID:   "t1",
		TaskID:   "task-42",
		ToolName: "web_search",
		Params:   map[string]interface{}{"query": "hello"},
		Deadline: time.Now().Add(5 * time.Minute),
	}

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var decoded TeamTaskPayload
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "web_search", decoded.ToolName)
	assert.Equal(t, "hello", decoded.Params["query"])
}

func TestTeamResultPayload_JSON(t *testing.T) {
	t.Parallel()

	p := TeamResultPayload{
		TeamID:    "t1",
		TaskID:    "task-42",
		MemberDID: "did:worker:1",
		Result:    map[string]interface{}{"answer": "42"},
		Duration:  1500,
	}

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var decoded TeamResultPayload
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, int64(1500), decoded.Duration)
}

func TestTeamDisbandPayload_JSON(t *testing.T) {
	t.Parallel()

	p := TeamDisbandPayload{
		TeamID: "t1",
		Reason: "task complete",
	}

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var decoded TeamDisbandPayload
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "task complete", decoded.Reason)
}
