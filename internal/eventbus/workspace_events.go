package eventbus

import "time"

// WorkspaceCreatedEvent is published when a new workspace is created.
type WorkspaceCreatedEvent struct {
	WorkspaceID string
	Name        string
	Goal        string
	CreatorDID  string
	CreatedAt   time.Time
}

// EventName implements Event.
func (e WorkspaceCreatedEvent) EventName() string { return "workspace.created" }

// WorkspaceMemberJoinedEvent is published when a member joins a workspace.
type WorkspaceMemberJoinedEvent struct {
	WorkspaceID string
	MemberDID   string
	JoinedAt    time.Time
}

// EventName implements Event.
func (e WorkspaceMemberJoinedEvent) EventName() string { return "workspace.member.joined" }

// WorkspaceMemberLeftEvent is published when a member leaves a workspace.
type WorkspaceMemberLeftEvent struct {
	WorkspaceID string
	MemberDID   string
	LeftAt      time.Time
}

// EventName implements Event.
func (e WorkspaceMemberLeftEvent) EventName() string { return "workspace.member.left" }

// WorkspaceCommitReceivedEvent is published when a git commit is received in a workspace.
type WorkspaceCommitReceivedEvent struct {
	WorkspaceID string
	CommitHash  string
	SenderDID   string
	Message     string
	ReceivedAt  time.Time
}

// EventName implements Event.
func (e WorkspaceCommitReceivedEvent) EventName() string { return "workspace.commit.received" }

// WorkspaceMessagePostedEvent is published when a message is posted to a workspace.
type WorkspaceMessagePostedEvent struct {
	WorkspaceID string
	MessageID   string
	MessageType string
	SenderDID   string
	PostedAt    time.Time
}

// EventName implements Event.
func (e WorkspaceMessagePostedEvent) EventName() string { return "workspace.message.posted" }

// WorkspaceArchivedEvent is published when a workspace is archived.
type WorkspaceArchivedEvent struct {
	WorkspaceID string
	ArchivedAt  time.Time
}

// EventName implements Event.
func (e WorkspaceArchivedEvent) EventName() string { return "workspace.archived" }

// WorkspaceGitDivergenceEvent is published when team members have divergent git HEADs.
type WorkspaceGitDivergenceEvent struct {
	WorkspaceID  string
	MajorityHead string
	Divergent    []GitDivergenceInfo
	DetectedAt   time.Time
}

// GitDivergenceInfo describes a member whose HEAD diverges from the majority.
type GitDivergenceInfo struct {
	MemberDID string
	HeadHash  string
}

// EventName implements Event.
func (e WorkspaceGitDivergenceEvent) EventName() string { return "workspace.git.divergence" }
