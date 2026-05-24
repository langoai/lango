package agentrt

import (
	"context"
	"encoding/json"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/runledger"
)

const (
	runLedgerMirrorTargetAgentRunProjection = "agent_run_projection"
	runLedgerMirrorPhaseAppendJournal       = "append_journal"
	runLedgerMirrorPhaseRefreshSnapshot     = "refresh_snapshot"
)

// RunLedgerMirrorStore decorates an AgentRunStore with best-effort RunLedger mirroring.
// Transition derivation is intentionally best effort: concurrent projection writes
// for the same run may emit duplicate approval-block events, but the durable
// snapshot still converges on the latest blocked state.
type RunLedgerMirrorStore struct {
	base   AgentRunStore
	ledger runledger.RunLedgerStore
	bus    *eventbus.Bus
}

var _ AgentRunStore = (*RunLedgerMirrorStore)(nil)

func NewRunLedgerMirrorStore(base AgentRunStore, ledger runledger.RunLedgerStore, bus *eventbus.Bus) *RunLedgerMirrorStore {
	return &RunLedgerMirrorStore{
		base:   base,
		ledger: ledger,
		bus:    bus,
	}
}

func (s *RunLedgerMirrorStore) Create(run *AgentRun) error {
	return s.base.Create(run)
}

func (s *RunLedgerMirrorStore) Get(id string) (*AgentRun, error) {
	return s.base.Get(id)
}

func (s *RunLedgerMirrorStore) List() []*AgentRun {
	return s.base.List()
}

func (s *RunLedgerMirrorStore) UpdateStatus(id string, status AgentRunStatus, result, errMsg string) error {
	return s.base.UpdateStatus(id, status, result, errMsg)
}

func (s *RunLedgerMirrorStore) Cancel(id string) error {
	return s.base.Cancel(id)
}

func (s *RunLedgerMirrorStore) UpdateProjection(id string, patch RunProjectionPatch) error {
	beforeRun, err := s.base.Get(id)
	if err != nil {
		return err
	}
	// Keep a private snapshot even when the underlying store changes its Get
	// semantics later. The current in-memory store already returns a copy.
	before := copyRun(beforeRun)
	if err := s.base.UpdateProjection(id, patch); err != nil {
		return err
	}
	afterRun, err := s.base.Get(id)
	if err != nil {
		return err
	}
	// Mirror the same defensive snapshotting for the post-write state.
	after := copyRun(afterRun)

	beforeBlocked := before.RuntimeCondition == AgentRunConditionBlockedWaitingApproval
	afterBlocked := after.RuntimeCondition == AgentRunConditionBlockedWaitingApproval
	changedBlockedState := before.BlockedReason != after.BlockedReason ||
		before.GrantRequestID != after.GrantRequestID ||
		before.GrantAttempt != after.GrantAttempt ||
		before.GrantState != after.GrantState

	switch {
	case afterBlocked && (!beforeBlocked || changedBlockedState):
		s.appendMirrorEvent(id, runledger.JournalEvent{
			RunID: id,
			Type:  runledger.EventTeammateApprovalBlocked,
			Payload: marshalRunLedgerPayload(runledger.TeammateApprovalBlockedPayload{
				RuntimeCondition: string(after.RuntimeCondition),
				BlockedReason:    after.BlockedReason,
				GrantRequestID:   after.GrantRequestID,
				GrantAttempt:     after.GrantAttempt,
				GrantState:       after.GrantState,
			}),
		})
	case beforeBlocked && !afterBlocked:
		s.appendMirrorEvent(id, runledger.JournalEvent{
			RunID:   id,
			Type:    runledger.EventTeammateApprovalUnblocked,
			Payload: marshalRunLedgerPayload(runledger.TeammateApprovalUnblockedPayload{}),
		})
	}

	return nil
}

func (s *RunLedgerMirrorStore) appendMirrorEvent(runID string, ev runledger.JournalEvent) {
	ctx := context.Background()
	if err := s.ledger.AppendJournalEvent(ctx, ev); err != nil {
		s.observeMirrorFailure(runID, runLedgerMirrorPhaseAppendJournal, err)
		return
	}
	if _, err := s.ledger.GetRunSnapshot(ctx, runID); err != nil {
		s.observeMirrorFailure(runID, runLedgerMirrorPhaseRefreshSnapshot, err)
	}
}

func (s *RunLedgerMirrorStore) observeMirrorFailure(runID, phase string, err error) {
	logger().Warnw("runledger mirror failed", "target", runLedgerMirrorTargetAgentRunProjection, "phase", phase, "runID", runID, "error", err)
	if s.bus != nil {
		s.bus.Publish(eventbus.RunLedgerMirrorFailureEvent{
			Target: runLedgerMirrorTargetAgentRunProjection,
			Phase:  phase,
			RunID:  runID,
			Error:  err.Error(),
		})
	}
}

func marshalRunLedgerPayload(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		logger().Warnw("runledger mirror payload marshal failed", "error", err)
		return json.RawMessage(`{}`)
	}
	return data
}
