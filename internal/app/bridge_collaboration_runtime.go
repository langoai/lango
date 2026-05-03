package app

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/langoai/lango/internal/agentrt"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/mission"
	"github.com/langoai/lango/internal/turntrace"
)

const collaborationRuntimeBufferLimit = 16

type collaborationMissionLinkReader struct {
	store mission.Store
}

func (r *collaborationMissionLinkReader) ListMissionExecutionLinks(ctx context.Context, missionID string) ([]CollaborationMissionExecutionLink, error) {
	if r == nil || r.store == nil {
		return nil, nil
	}
	links, err := r.store.ListExecutionLinks(ctx, missionID)
	if err != nil {
		return nil, err
	}
	out := make([]CollaborationMissionExecutionLink, 0, len(links))
	for _, link := range links {
		if link == nil {
			continue
		}
		out = append(out, CollaborationMissionExecutionLink{
			ExecutionKind: string(link.ExecutionKind),
			ExecutionRef:  link.ExecutionRef,
		})
	}
	return out, nil
}

type collaborationAgentRunReader struct {
	store agentrt.AgentRunStore
}

func (r *collaborationAgentRunReader) ListAgentRuns() []CollaborationAgentRunView {
	if r == nil || r.store == nil {
		return nil
	}
	runs := r.store.List()
	out := make([]CollaborationAgentRunView, 0, len(runs))
	for _, run := range runs {
		if run == nil {
			continue
		}
		out = append(out, CollaborationAgentRunView{
			ID:               run.ID,
			RequestedAgent:   run.RequestedAgent,
			RuntimeCondition: string(run.RuntimeCondition),
			BlockedReason:    run.BlockedReason,
			WaitingOnRunID:   run.WaitingOnRunID,
			RecoveryState:    run.RecoveryState,
		})
	}
	return out
}

type collaborationDelegationReader struct {
	store turntrace.Store
}

func (r *collaborationDelegationReader) ListDelegationsForSession(ctx context.Context, sessionKey string) ([]CollaborationDelegationRecord, error) {
	if r == nil || r.store == nil {
		return nil, nil
	}
	traces, err := r.store.TracesForSession(ctx, sessionKey)
	if err != nil {
		return nil, err
	}
	out := make([]CollaborationDelegationRecord, 0)
	for _, trace := range traces {
		events, err := r.store.EventsForTrace(ctx, trace.TraceID)
		if err != nil {
			return nil, err
		}
		for _, ev := range events {
			if ev.EventType != turntrace.EventDelegation && ev.EventType != turntrace.EventDelegationReturn {
				continue
			}
			to := extractDelegationTarget(ev.PayloadJSON)
			if strings.TrimSpace(to) == "" {
				continue
			}
			out = append(out, CollaborationDelegationRecord{
				From:      ev.AgentName,
				To:        to,
				Timestamp: ev.CreatedAt,
			})
		}
	}
	return out, nil
}

func extractDelegationTarget(payload string) string {
	if strings.TrimSpace(payload) == "" {
		return ""
	}
	var decoded struct {
		To string `json:"to"`
	}
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		return ""
	}
	return strings.TrimSpace(decoded.To)
}

type collaborationRuntimeBridge struct {
	mu               sync.RWMutex
	sessionMission   map[string]string
	budgetByMission  map[string][]CollaborationBudgetRecord
	recoverByMission map[string][]CollaborationRecoveryRecord
}

func newCollaborationRuntimeBridge(bus *eventbus.Bus) *collaborationRuntimeBridge {
	b := &collaborationRuntimeBridge{
		sessionMission:   make(map[string]string),
		budgetByMission:  make(map[string][]CollaborationBudgetRecord),
		recoverByMission: make(map[string][]CollaborationRecoveryRecord),
	}
	if bus == nil {
		return b
	}
	eventbus.SubscribeTyped(bus, func(e agentrt.BudgetAlertEvent) {
		b.recordBudget(e)
	})
	eventbus.SubscribeTyped(bus, func(e agentrt.RecoveryDecisionEvent) {
		b.recordRecovery(e)
	})
	return b
}

func (b *collaborationRuntimeBridge) BindSessionMission(sessionKey, missionID string) {
	sessionKey = strings.TrimSpace(sessionKey)
	missionID = strings.TrimSpace(missionID)
	if sessionKey == "" || missionID == "" {
		return
	}
	b.mu.Lock()
	b.sessionMission[sessionKey] = missionID
	b.mu.Unlock()
}

func (b *collaborationRuntimeBridge) ListBudgetSignals(missionID string) []CollaborationBudgetRecord {
	b.mu.RLock()
	defer b.mu.RUnlock()
	records := b.budgetByMission[strings.TrimSpace(missionID)]
	return append([]CollaborationBudgetRecord(nil), records...)
}

func (b *collaborationRuntimeBridge) ListRecoverySignals(missionID string) []CollaborationRecoveryRecord {
	b.mu.RLock()
	defer b.mu.RUnlock()
	records := b.recoverByMission[strings.TrimSpace(missionID)]
	return append([]CollaborationRecoveryRecord(nil), records...)
}

func (b *collaborationRuntimeBridge) recordBudget(evt agentrt.BudgetAlertEvent) {
	sessionKey := strings.TrimSpace(evt.SessionID)
	if sessionKey == "" {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	missionID := strings.TrimSpace(b.sessionMission[sessionKey])
	if missionID == "" {
		return
	}
	records := append(b.budgetByMission[missionID], CollaborationBudgetRecord{
		MissionID: missionID,
		Used:      evt.Used,
		Max:       evt.Limit,
		Timestamp: eventTime(),
	})
	if len(records) > collaborationRuntimeBufferLimit {
		records = append([]CollaborationBudgetRecord(nil), records[len(records)-collaborationRuntimeBufferLimit:]...)
	}
	b.budgetByMission[missionID] = records
}

func (b *collaborationRuntimeBridge) recordRecovery(evt agentrt.RecoveryDecisionEvent) {
	sessionKey := strings.TrimSpace(evt.SessionKey)
	if sessionKey == "" {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	missionID := strings.TrimSpace(b.sessionMission[sessionKey])
	if missionID == "" {
		return
	}
	records := append(b.recoverByMission[missionID], CollaborationRecoveryRecord{
		MissionID:  missionID,
		Action:     evt.Action,
		CauseClass: evt.CauseClass,
		Timestamp:  eventTime(),
	})
	if len(records) > collaborationRuntimeBufferLimit {
		records = append([]CollaborationRecoveryRecord(nil), records[len(records)-collaborationRuntimeBufferLimit:]...)
	}
	b.recoverByMission[missionID] = records
}

var eventTime = func() time.Time {
	return time.Now()
}
