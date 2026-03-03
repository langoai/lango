package protocol

import "time"

// Team-specific request types for P2P team coordination.
const (
	// RequestTeamInvite invites a remote agent to join a team.
	RequestTeamInvite RequestType = "team_invite"

	// RequestTeamAccept acknowledges acceptance of a team invitation.
	RequestTeamAccept RequestType = "team_accept"

	// RequestTeamTask delegates a task to a team member.
	RequestTeamTask RequestType = "team_task"

	// RequestTeamResult reports the result of a delegated task back to the leader.
	RequestTeamResult RequestType = "team_result"

	// RequestTeamDisband notifies team members that the team is disbanding.
	RequestTeamDisband RequestType = "team_disband"
)

// TeamInvitePayload is the payload for a team invitation.
type TeamInvitePayload struct {
	TeamID      string   `json:"teamId"`
	TeamName    string   `json:"teamName"`
	Goal        string   `json:"goal"`
	LeaderDID   string   `json:"leaderDid"`
	Role        string   `json:"role"`
	Capabilities []string `json:"capabilities"`
}

// TeamAcceptPayload is the payload for accepting a team invitation.
type TeamAcceptPayload struct {
	TeamID    string `json:"teamId"`
	MemberDID string `json:"memberDid"`
	Accepted  bool   `json:"accepted"`
	Reason    string `json:"reason,omitempty"`
}

// TeamTaskPayload is the payload for delegating a task to a team member.
type TeamTaskPayload struct {
	TeamID   string                 `json:"teamId"`
	TaskID   string                 `json:"taskId"`
	ToolName string                 `json:"toolName"`
	Params   map[string]interface{} `json:"params"`
	Deadline time.Time              `json:"deadline,omitempty"`
}

// TeamResultPayload is the payload for reporting a task result.
type TeamResultPayload struct {
	TeamID    string                 `json:"teamId"`
	TaskID    string                 `json:"taskId"`
	MemberDID string                 `json:"memberDid"`
	Result    map[string]interface{} `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Duration  int64                  `json:"durationMs"`
}

// TeamDisbandPayload is the payload for disbanding a team.
type TeamDisbandPayload struct {
	TeamID string `json:"teamId"`
	Reason string `json:"reason"`
}
