package workspace

import (
	"sync"
	"time"
)

// Contribution records a member's contribution to a workspace.
type Contribution struct {
	DID        string    `json:"did"`
	Commits    int       `json:"commits"`
	CodeBytes  int64     `json:"codeBytes"`
	Messages   int       `json:"messages"`
	LastActive time.Time `json:"lastActive"`
}

// ContributionTracker tracks contributions per member per workspace.
type ContributionTracker struct {
	mu   sync.RWMutex
	data map[string]map[string]*Contribution // workspaceID -> DID -> Contribution
}

// NewContributionTracker creates a new contribution tracker.
func NewContributionTracker() *ContributionTracker {
	return &ContributionTracker{
		data: make(map[string]map[string]*Contribution),
	}
}

// RecordCommit records a commit contribution.
func (t *ContributionTracker) RecordCommit(workspaceID, did string, codeBytes int64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	c := t.getOrCreate(workspaceID, did)
	c.Commits++
	c.CodeBytes += codeBytes
	c.LastActive = time.Now()
}

// RecordMessage records a message contribution.
func (t *ContributionTracker) RecordMessage(workspaceID, did string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	c := t.getOrCreate(workspaceID, did)
	c.Messages++
	c.LastActive = time.Now()
}

// Get returns the contribution for a member in a workspace.
func (t *ContributionTracker) Get(workspaceID, did string) *Contribution {
	t.mu.RLock()
	defer t.mu.RUnlock()

	ws, ok := t.data[workspaceID]
	if !ok {
		return nil
	}
	return ws[did]
}

// List returns all contributions for a workspace.
func (t *ContributionTracker) List(workspaceID string) []*Contribution {
	t.mu.RLock()
	defer t.mu.RUnlock()

	ws, ok := t.data[workspaceID]
	if !ok {
		return nil
	}

	result := make([]*Contribution, 0, len(ws))
	for _, c := range ws {
		result = append(result, c)
	}
	return result
}

func (t *ContributionTracker) getOrCreate(workspaceID, did string) *Contribution {
	ws, ok := t.data[workspaceID]
	if !ok {
		ws = make(map[string]*Contribution)
		t.data[workspaceID] = ws
	}

	c, ok := ws[did]
	if !ok {
		c = &Contribution{DID: did}
		ws[did] = c
	}
	return c
}
