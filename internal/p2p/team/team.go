// Package team defines the types and coordination primitives for P2P agent teams.
// A team is a dynamic, task-scoped group of agents that collaborate on a goal.
package team

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Sentinel errors for team operations.
var (
	ErrTeamFull       = errors.New("team is at maximum capacity")
	ErrBudgetExceeded = errors.New("team budget exceeded")
	ErrAlreadyMember  = errors.New("agent is already a team member")
	ErrNotMember      = errors.New("agent is not a team member")
	ErrTeamDisbanded  = errors.New("team has been disbanded")
	ErrConflict       = errors.New("conflicting results from team members")
)

// MemberStatus represents the operational state of a team member.
type MemberStatus string

const (
	MemberIdle   MemberStatus = "idle"
	MemberBusy   MemberStatus = "busy"
	MemberFailed MemberStatus = "failed"
	MemberLeft   MemberStatus = "left"
)

// Role describes a member's function within a team.
type Role string

const (
	RoleLeader   Role = "leader"
	RoleWorker   Role = "worker"
	RoleReviewer Role = "reviewer"
	RoleObserver Role = "observer"
)

// TeamStatus represents the lifecycle state of a team.
type TeamStatus string

const (
	StatusForming   TeamStatus = "forming"
	StatusActive    TeamStatus = "active"
	StatusCompleted TeamStatus = "completed"
	StatusDisbanded TeamStatus = "disbanded"
)

// Member represents an agent participating in a team.
type Member struct {
	DID          string            `json:"did"`
	Name         string            `json:"name"`
	PeerID       string            `json:"peerId"`
	Role         Role              `json:"role"`
	Status       MemberStatus      `json:"status"`
	Capabilities []string          `json:"capabilities"`
	TrustScore   float64           `json:"trustScore"`
	JoinedAt     time.Time         `json:"joinedAt"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Team is a task-scoped group of P2P agents coordinating on a shared goal.
type Team struct {
	mu sync.RWMutex

	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Goal        string     `json:"goal"`
	LeaderDID   string     `json:"leaderDid"`
	Status      TeamStatus `json:"status"`
	MaxMembers  int        `json:"maxMembers"`
	Budget      float64    `json:"budget"`
	Spent       float64    `json:"spent"`
	CreatedAt   time.Time  `json:"createdAt"`
	DisbandedAt time.Time  `json:"disbandedAt,omitempty"`

	members map[string]*Member // keyed by DID
}

// NewTeam creates a team in the forming state.
func NewTeam(id, name, goal, leaderDID string, maxMembers int) *Team {
	return &Team{
		ID:         id,
		Name:       name,
		Goal:       goal,
		LeaderDID:  leaderDID,
		Status:     StatusForming,
		MaxMembers: maxMembers,
		CreatedAt:  time.Now(),
		members:    make(map[string]*Member),
	}
}

// AddMember adds an agent to the team.
func (t *Team) AddMember(m *Member) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Status == StatusDisbanded {
		return ErrTeamDisbanded
	}
	if _, ok := t.members[m.DID]; ok {
		return ErrAlreadyMember
	}
	if t.MaxMembers > 0 && len(t.members) >= t.MaxMembers {
		return ErrTeamFull
	}

	m.JoinedAt = time.Now()
	t.members[m.DID] = m
	return nil
}

// RemoveMember removes an agent from the team.
func (t *Team) RemoveMember(did string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, ok := t.members[did]; !ok {
		return ErrNotMember
	}
	delete(t.members, did)
	return nil
}

// GetMember returns a member by DID.
func (t *Team) GetMember(did string) *Member {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.members[did]
}

// Clone returns a deep copy of the Member.
func (m *Member) Clone() *Member {
	c := *m
	if len(m.Capabilities) > 0 {
		c.Capabilities = make([]string, len(m.Capabilities))
		copy(c.Capabilities, m.Capabilities)
	}
	if len(m.Metadata) > 0 {
		c.Metadata = make(map[string]string, len(m.Metadata))
		for k, v := range m.Metadata {
			c.Metadata[k] = v
		}
	}
	return &c
}

// Members returns copies of all current members (safe for concurrent use).
func (t *Team) Members() []*Member {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]*Member, 0, len(t.members))
	for _, m := range t.members {
		result = append(result, m.Clone())
	}
	return result
}

// MemberCount returns the number of members.
func (t *Team) MemberCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return len(t.members)
}

// ActiveMembers returns members that are not in MemberLeft or MemberFailed state.
func (t *Team) ActiveMembers() []*Member {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var result []*Member
	for _, m := range t.members {
		if m.Status != MemberLeft && m.Status != MemberFailed {
			result = append(result, m)
		}
	}
	return result
}

// AddSpend adds to the team's spent total. Returns ErrBudgetExceeded if budget is exceeded.
func (t *Team) AddSpend(amount float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Budget > 0 && t.Spent+amount > t.Budget {
		return ErrBudgetExceeded
	}
	t.Spent += amount
	return nil
}

// Activate transitions the team to active status.
func (t *Team) Activate() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Status = StatusActive
}

// Disband marks the team as disbanded.
func (t *Team) Disband() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Status = StatusDisbanded
	t.DisbandedAt = time.Now()
}

// ScopedContext wraps a context with team-specific metadata so that downstream
// handlers can identify which team and member is executing a request.
type ScopedContext struct {
	TeamID    string
	MemberDID string
	Role      Role
}

type scopedContextKey struct{}

// WithScopedContext returns a context carrying the team scope.
func WithScopedContext(ctx context.Context, sc ScopedContext) context.Context {
	return context.WithValue(ctx, scopedContextKey{}, sc)
}

// ScopedContextFromContext extracts the team scope from ctx.
func ScopedContextFromContext(ctx context.Context) (ScopedContext, bool) {
	sc, ok := ctx.Value(scopedContextKey{}).(ScopedContext)
	return sc, ok
}

// ContextFilter determines which context data is shared with a team member.
type ContextFilter struct {
	// AllowedKeys restricts shared metadata to these keys. Empty means allow all.
	AllowedKeys []string
	// ExcludeKeys removes these keys from shared metadata.
	ExcludeKeys []string
}

// Filter applies the filter to a metadata map and returns a new filtered copy.
func (f *ContextFilter) Filter(metadata map[string]string) map[string]string {
	if metadata == nil {
		return nil
	}

	excluded := make(map[string]struct{}, len(f.ExcludeKeys))
	for _, k := range f.ExcludeKeys {
		excluded[k] = struct{}{}
	}

	allowed := make(map[string]struct{}, len(f.AllowedKeys))
	for _, k := range f.AllowedKeys {
		allowed[k] = struct{}{}
	}

	result := make(map[string]string)
	for k, v := range metadata {
		if _, ok := excluded[k]; ok {
			continue
		}
		if len(allowed) > 0 {
			if _, ok := allowed[k]; !ok {
				continue
			}
		}
		result[k] = v
	}
	return result
}

// TaskResultSummary holds the summarized result of a delegated task.
type TaskResultSummary struct {
	TaskID     string  `json:"taskId"`
	AgentDID   string  `json:"agentDid"`
	AgentName  string  `json:"agentName"`
	Success    bool    `json:"success"`
	Result     string  `json:"result"`
	Error      string  `json:"error,omitempty"`
	DurationMs int64   `json:"durationMs"`
	Cost       float64 `json:"cost"`
}
