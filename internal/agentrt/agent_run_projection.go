package agentrt

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/langoai/lango/internal/background"
)

// Compile-time interface satisfaction check.
var _ background.Projection = (*AgentRunProjection)(nil)

type pendingAgentRunContextKey struct{}

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

type backgroundRunLedgerProjection interface {
	PrepareTask(context.Context, string, background.Origin) (string, error)
	PrepareTaskWithID(context.Context, string, background.Origin, string) error
	SyncTask(context.Context, background.TaskSnapshot) error
}

type BackgroundProjection struct {
	agentRuns   *AgentRunProjection
	runLedger   backgroundRunLedgerProjection
	mu          sync.Mutex
	agentRunIDs map[string]bool
}

// NewAgentRunProjection creates a new AgentRunProjection backed by the given store.
func NewAgentRunProjection(store AgentRunStore) *AgentRunProjection {
	return &AgentRunProjection{
		store:   store,
		pending: make(map[string]string),
	}
}

func NewBackgroundProjection(
	agentRuns *AgentRunProjection,
	runLedgerProjection backgroundRunLedgerProjection,
) *BackgroundProjection {
	return &BackgroundProjection{
		agentRuns:   agentRuns,
		runLedger:   runLedgerProjection,
		agentRunIDs: make(map[string]bool),
	}
}

func withPendingAgentRunID(ctx context.Context, runID string) context.Context {
	return context.WithValue(ctx, pendingAgentRunContextKey{}, runID)
}

func pendingAgentRunIDFromContext(ctx context.Context) string {
	runID, _ := ctx.Value(pendingAgentRunContextKey{}).(string)
	return runID
}

// RegisterPending pre-registers an AgentRun ID so that the next PrepareTask
// call returns it instead of generating a new one. Called by the spawn path
// (D2) before bgManager.Submit.
func (p *AgentRunProjection) RegisterPending(agentRunID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pending[agentRunID] = agentRunID
}

func (p *AgentRunProjection) consumePending() (string, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for id := range p.pending {
		delete(p.pending, id)
		return id, true
	}

	return "", false
}

// PrepareTask implements background.Projection. It returns a pre-assigned
// AgentRun ID if one was registered via RegisterPending. If no pending ID
// exists, it returns an error — callers must always register before submit.
//
// PrepareTask does NOT change AgentRun status; the run stays in its current
// state (typically Spawned) until SyncTask is called by the manager.
func (p *AgentRunProjection) PrepareTask(ctx context.Context, _ string, _ background.Origin) (string, error) {
	if id := pendingAgentRunIDFromContext(ctx); id != "" {
		return id, nil
	}
	if id, ok := p.consumePending(); ok {
		return id, nil
	}
	return "", fmt.Errorf("prepare task: no pending agent run ID registered")
}

func (p *BackgroundProjection) PrepareTask(
	ctx context.Context,
	prompt string,
	origin background.Origin,
) (string, error) {
	if id := pendingAgentRunIDFromContext(ctx); id != "" {
		if p.runLedger != nil {
			if err := p.runLedger.PrepareTaskWithID(ctx, prompt, origin, id); err != nil {
				return "", err
			}
		}
		p.mu.Lock()
		p.agentRunIDs[id] = true
		p.mu.Unlock()
		return id, nil
	}

	if p.agentRuns != nil {
		if id, ok := p.agentRuns.consumePending(); ok {
			if p.runLedger != nil {
				if err := p.runLedger.PrepareTaskWithID(ctx, prompt, origin, id); err != nil {
					return "", err
				}
			}
			p.mu.Lock()
			p.agentRunIDs[id] = true
			p.mu.Unlock()
			return id, nil
		}
	}

	if p.runLedger != nil {
		return p.runLedger.PrepareTask(ctx, prompt, origin)
	}

	return uuid.NewString(), nil
}

func (p *BackgroundProjection) SyncTask(ctx context.Context, snap background.TaskSnapshot) error {
	var errs []error

	if p.isAgentRun(snap.ID) && p.agentRuns != nil {
		if err := p.agentRuns.SyncTask(ctx, snap); err != nil {
			errs = append(errs, err)
		}
	}
	if p.runLedger != nil {
		if err := p.runLedger.SyncTask(ctx, snap); err != nil {
			errs = append(errs, err)
		}
	}
	if snap.Status == background.Done || snap.Status == background.Failed || snap.Status == background.Cancelled {
		p.clearAgentRun(snap.ID)
	}

	return errors.Join(errs...)
}

func (p *BackgroundProjection) isAgentRun(id string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.agentRunIDs[id]
}

func (p *BackgroundProjection) clearAgentRun(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.agentRunIDs, id)
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
