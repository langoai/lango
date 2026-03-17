package workspace

import (
	"time"
)

// MessageType identifies the kind of workspace message.
type MessageType string

const (
	MessageTypeTaskProposal   MessageType = "TASK_PROPOSAL"
	MessageTypeLogStream      MessageType = "LOG_STREAM"
	MessageTypeCommitSignal   MessageType = "COMMIT_SIGNAL"
	MessageTypeKnowledgeShare MessageType = "KNOWLEDGE_SHARE"
	MessageTypeMemberJoined   MessageType = "MEMBER_JOINED"
	MessageTypeMemberLeft     MessageType = "MEMBER_LEFT"

	// Conflict and branch collaboration message types.
	MessageTypeConflictReport MessageType = "CONFLICT_REPORT"
	MessageTypeBranchCreated  MessageType = "BRANCH_CREATED"
	MessageTypeBranchMerged   MessageType = "BRANCH_MERGED"
	MessageTypeSyncRequest    MessageType = "SYNC_REQUEST"
)

// Message represents a message posted to a workspace.
type Message struct {
	ID          string            `json:"id"`
	Type        MessageType       `json:"type"`
	WorkspaceID string            `json:"workspaceId"`
	SenderDID   string            `json:"senderDid"`
	Content     string            `json:"content"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	ParentID    string            `json:"parentId,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
}
