package agentrt

import (
	"context"
	"fmt"
	"sync"

	"github.com/langoai/lango/internal/background"
)

// Compile-time interface satisfaction check.
var _ background.Projection = (*AgentRunProjection)(nil)

// AgentRunProjection implements background.Projection to synchronize
// background task lifecycle events to AgentRunStore.
//
// ID unification: RegisterPending pre-assigns the AgentRun.ID so that
// PrepareTask returns it to the background manager, ensuring both layers
// share the same canonical ID.
type AgentRunProjection struct {
	store   AgentRunStore
	mu      sync.Mutex
	pending map[string]string // agentRunID → agentRunID (identity map for PrepareTask)
}

// NewAgentRunProjection creates a new AgentRunProjection backed by the given store.
func NewAgentRunProjection(store AgentRunStore) *AgentRunProjection {
	return &AgentRunProjection{
		store:   store,
		pending: make(map[string]string),
	}
}

// RegisterPending pre-registers an AgentRun ID so that the next PrepareTask
// call returns it instead of generating a new one. Called by the spawn path
// (D2) before bgManager.Submit.
func (p *AgentRunProjection) RegisterPending(agentRunID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pending[agentRunID] = agentRunID
}

// PrepareTask implements background.Projection. It returns a pre-assigned
// AgentRun ID if one was registered via RegisterPending. If no pending ID
// exists, it returns an error — callers must always register before submit.
//
// PrepareTask does NOT change AgentRun status; the run stays in its current
// state (typically Spawned) until SyncTask is called by the manager.
func (p *AgentRunProjection) PrepareTask(_ context.Context, _ string, _ background.Origin) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for id := range p.pending {
		delete(p.pending, id)
		return id, nil
	}
	return "", fmt.Errorf("prepare task: no pending agent run ID registered")
}

// SyncTask implements background.Projection. It maps background task status
// transitions to AgentRun status and persists them via the store.
//
// Status mapping:
//   - Pending   → AgentRunSpawned
//   - Running   → AgentRunRunning
//   - Done      → AgentRunCompleted (with result)
//   - Failed    → AgentRunFailed (with error)
//   - Cancelled → AgentRunCancelled
func (p *AgentRunProjection) SyncTask(_ context.Context, snap background.TaskSnapshot) error {
	status, err := mapBgStatus(snap.Status)
	if err != nil {
		return err
	}
	return p.store.UpdateStatus(snap.ID, status, snap.Result, snap.Error)
}

// mapBgStatus converts a background.Status to the corresponding AgentRunStatus.
func mapBgStatus(s background.Status) (AgentRunStatus, error) {
	switch s {
	case background.Pending:
		return AgentRunSpawned, nil
	case background.Running:
		return AgentRunRunning, nil
	case background.Done:
		return AgentRunCompleted, nil
	case background.Failed:
		return AgentRunFailed, nil
	case background.Cancelled:
		return AgentRunCancelled, nil
	default:
		return "", fmt.Errorf("map background status: unknown status %d", s)
	}
}
