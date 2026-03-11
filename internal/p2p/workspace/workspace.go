package workspace

import (
	"time"
)

// Status represents the lifecycle state of a workspace.
type Status string

const (
	StatusForming  Status = "forming"
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
)

// Workspace represents a collaborative workspace where agents share code and messages.
type Workspace struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Goal      string            `json:"goal"`
	Status    Status            `json:"status"`
	Members   []*Member         `json:"members"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Role represents a member's role in a workspace.
type Role string

const (
	RoleCreator Role = "creator"
	RoleMember  Role = "member"
)

// Member represents a participant in a workspace.
type Member struct {
	DID      string    `json:"did"`
	Name     string    `json:"name,omitempty"`
	Role     Role      `json:"role"`
	JoinedAt time.Time `json:"joinedAt"`
}

// CreateRequest holds parameters for creating a new workspace.
type CreateRequest struct {
	Name     string            `json:"name"`
	Goal     string            `json:"goal"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ReadOptions controls workspace message listing.
type ReadOptions struct {
	Limit     int       `json:"limit,omitempty"`
	Before    time.Time `json:"before,omitempty"`
	After     time.Time `json:"after,omitempty"`
	SenderDID string    `json:"senderDID,omitempty"`
	Types     []string  `json:"types,omitempty"`
	ParentID  string    `json:"parentID,omitempty"`
}
