package protocol

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTeamRequestTypes(t *testing.T) {
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
			if string(tt.give) != tt.want {
				t.Errorf("RequestType = %q, want %q", tt.give, tt.want)
			}
		})
	}
}

func TestTeamInvitePayload_JSON(t *testing.T) {
	p := TeamInvitePayload{
		TeamID:       "t1",
		TeamName:     "search-team",
		Goal:         "find information",
		LeaderDID:    "did:leader:123",
		Role:         "worker",
		Capabilities: []string{"search", "summarize"},
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded TeamInvitePayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.TeamID != p.TeamID {
		t.Errorf("TeamID = %q, want %q", decoded.TeamID, p.TeamID)
	}
	if len(decoded.Capabilities) != 2 {
		t.Errorf("Capabilities count = %d, want 2", len(decoded.Capabilities))
	}
}

func TestTeamTaskPayload_JSON(t *testing.T) {
	p := TeamTaskPayload{
		TeamID:   "t1",
		TaskID:   "task-42",
		ToolName: "web_search",
		Params:   map[string]interface{}{"query": "hello"},
		Deadline: time.Now().Add(5 * time.Minute),
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded TeamTaskPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.ToolName != "web_search" {
		t.Errorf("ToolName = %q, want %q", decoded.ToolName, "web_search")
	}
	if decoded.Params["query"] != "hello" {
		t.Errorf("Params[query] = %v, want %q", decoded.Params["query"], "hello")
	}
}

func TestTeamResultPayload_JSON(t *testing.T) {
	p := TeamResultPayload{
		TeamID:    "t1",
		TaskID:    "task-42",
		MemberDID: "did:worker:1",
		Result:    map[string]interface{}{"answer": "42"},
		Duration:  1500,
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded TeamResultPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.Duration != 1500 {
		t.Errorf("Duration = %d, want 1500", decoded.Duration)
	}
}

func TestTeamDisbandPayload_JSON(t *testing.T) {
	p := TeamDisbandPayload{
		TeamID: "t1",
		Reason: "task complete",
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded TeamDisbandPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.Reason != "task complete" {
		t.Errorf("Reason = %q, want %q", decoded.Reason, "task complete")
	}
}
