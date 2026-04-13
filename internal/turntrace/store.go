package turntrace

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/langoai/lango/internal/ent"
	entmessage "github.com/langoai/lango/internal/ent/message"
	entturntrace "github.com/langoai/lango/internal/ent/turntrace"
	entturntraceevent "github.com/langoai/lango/internal/ent/turntraceevent"
)

// Outcome classifies the terminal state of a turn.
type Outcome string

const (
	OutcomeRunning           Outcome = "running"
	OutcomeSuccess           Outcome = "success"
	OutcomeTimeout           Outcome = "timeout"
	OutcomeLoopDetected      Outcome = "loop_detected"
	OutcomeEmptyAfterToolUse Outcome = "empty_after_tool_use"
	OutcomeToolError         Outcome = "tool_error"
	OutcomeModelError        Outcome = "model_error"
	OutcomeInternalError     Outcome = "internal_error"
)

// Trace is the durable summary row for a single turn.
type Trace struct {
	TraceID    string
	SessionKey string
	Entrypoint string
	Outcome    Outcome
	ErrorCode  string
	CauseClass string
	CauseDetail string
	Summary    string
	StartedAt  time.Time
	EndedAt    *time.Time
}

// Event is a single append-only trace event.
type Event struct {
	TraceID       string
	Seq           int64
	EventType     string
	AgentName     string
	ToolName      string
	CallSignature string
	PayloadJSON   string
	PayloadTruncated bool
	CreatedAt     time.Time
}

// Store persists durable turn traces.
type Store interface {
	CreateTrace(ctx context.Context, trace Trace) error
	AppendEvent(ctx context.Context, event Event) error
	FinishTrace(
		ctx context.Context,
		traceID string,
		outcome Outcome,
		summary string,
		errorCode string,
		causeClass string,
		causeDetail string,
		endedAt time.Time,
	) error
	RecentFailures(ctx context.Context, limit int) ([]Trace, error)
	IsolationLeakCount(ctx context.Context, isolatedAgents []string) (int, error)

	// EventsForTrace returns all events for a trace, ordered by seq.
	EventsForTrace(ctx context.Context, traceID string) ([]Event, error)
	// TracesForSession returns all traces for a session key, ordered by started_at.
	TracesForSession(ctx context.Context, sessionKey string) ([]Trace, error)
	// PurgeTraces deletes traces and their associated events by trace IDs.
	PurgeTraces(ctx context.Context, traceIDs []string) error
	// TraceCount returns the total number of traces.
	TraceCount(ctx context.Context) (int, error)
	// OldTraces returns trace IDs older than the given cutoff.
	OldTraces(ctx context.Context, cutoff time.Time, onlySuccess bool, limit int) ([]string, error)
	// RecentByOutcome returns traces matching outcome within a time window.
	RecentByOutcome(ctx context.Context, outcome Outcome, since time.Time, limit int) ([]Trace, error)
}

// EntStore implements Store using the shared ent client.
type EntStore struct {
	client *ent.Client
}

// NewEntStore creates a trace store backed by ent.
func NewEntStore(client *ent.Client) *EntStore {
	return &EntStore{client: client}
}

// CreateTrace inserts the initial trace row.
func (s *EntStore) CreateTrace(ctx context.Context, trace Trace) error {
	if s == nil || s.client == nil {
		return nil
	}
	startedAt := trace.StartedAt
	if startedAt.IsZero() {
		startedAt = time.Now()
	}

	builder := s.client.TurnTrace.Create().
		SetTraceID(trace.TraceID).
		SetSessionKey(trace.SessionKey).
		SetEntrypoint(trace.Entrypoint).
		SetOutcome(string(trace.Outcome)).
		SetStartedAt(startedAt)
	if trace.ErrorCode != "" {
		builder.SetErrorCode(trace.ErrorCode)
	}
	if trace.CauseClass != "" {
		builder.SetCauseClass(trace.CauseClass)
	}
	if strings.TrimSpace(trace.CauseDetail) != "" {
		builder.SetCauseDetail(trace.CauseDetail)
	}
	if strings.TrimSpace(trace.Summary) != "" {
		builder.SetSummary(trace.Summary)
	}

	if _, err := builder.Save(ctx); err != nil {
		return fmt.Errorf("create trace %q: %w", trace.TraceID, err)
	}
	return nil
}

// AppendEvent persists an append-only trace event.
func (s *EntStore) AppendEvent(ctx context.Context, event Event) error {
	if s == nil || s.client == nil {
		return nil
	}
	createdAt := event.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	builder := s.client.TurnTraceEvent.Create().
		SetTraceID(event.TraceID).
		SetSeq(event.Seq).
		SetEventType(event.EventType).
		SetCreatedAt(createdAt)
	if event.AgentName != "" {
		builder.SetAgentName(event.AgentName)
	}
	if event.ToolName != "" {
		builder.SetToolName(event.ToolName)
	}
	if event.CallSignature != "" {
		builder.SetCallSignature(event.CallSignature)
	}
		if event.PayloadJSON != "" {
			builder.SetPayloadJSON(event.PayloadJSON)
		}
	if event.PayloadTruncated {
		builder.SetPayloadTruncated(true)
	}

	if _, err := builder.Save(ctx); err != nil {
		return fmt.Errorf("append trace event %q/%d: %w", event.TraceID, event.Seq, err)
	}
	return nil
}

// FinishTrace finalizes a trace with its terminal outcome.
func (s *EntStore) FinishTrace(
	ctx context.Context,
	traceID string,
	outcome Outcome,
	summary string,
	errorCode string,
	causeClass string,
	causeDetail string,
	endedAt time.Time,
) error {
	if s == nil || s.client == nil {
		return nil
	}
	builder := s.client.TurnTrace.Update().
		Where(entturntrace.TraceID(traceID)).
		SetOutcome(string(outcome)).
		SetEndedAt(endedAt)
	if strings.TrimSpace(summary) != "" {
		builder.SetSummary(summary)
	}
	if errorCode != "" {
		builder.SetErrorCode(errorCode)
	}
	if causeClass != "" {
		builder.SetCauseClass(causeClass)
	}
	if strings.TrimSpace(causeDetail) != "" {
		builder.SetCauseDetail(causeDetail)
	}

	if _, err := builder.Save(ctx); err != nil {
		return fmt.Errorf("finish trace %q: %w", traceID, err)
	}
	return nil
}

// RecentFailures returns the most recent non-success traces.
func (s *EntStore) RecentFailures(ctx context.Context, limit int) ([]Trace, error) {
	if s == nil || s.client == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 5
	}
	rows, err := s.client.TurnTrace.Query().
		Where(
			entturntrace.OutcomeNEQ(string(OutcomeSuccess)),
			entturntrace.OutcomeNEQ(string(OutcomeRunning)),
		).
		Order(ent.Desc(entturntrace.FieldStartedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query recent failures: %w", err)
	}

	out := make([]Trace, 0, len(rows))
	for _, row := range rows {
		out = append(out, entToTrace(row))
	}
	return out, nil
}

// IsolationLeakCount counts persisted raw turns authored by isolated agents.
func (s *EntStore) IsolationLeakCount(ctx context.Context, isolatedAgents []string) (int, error) {
	if s == nil || s.client == nil || len(isolatedAgents) == 0 {
		return 0, nil
	}

	count, err := s.client.Message.Query().
		Where(entmessage.AuthorIn(isolatedAgents...)).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count isolation leaks: %w", err)
	}
	return count, nil
}

// EventsForTrace returns all events for a trace, ordered by seq.
func (s *EntStore) EventsForTrace(ctx context.Context, traceID string) ([]Event, error) {
	if s == nil || s.client == nil {
		return nil, nil
	}
	rows, err := s.client.TurnTraceEvent.Query().
		Where(entturntraceevent.TraceID(traceID)).
		Order(ent.Asc(entturntraceevent.FieldSeq)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query events for trace %q: %w", traceID, err)
	}
	out := make([]Event, 0, len(rows))
	for _, row := range rows {
		out = append(out, entToEvent(row))
	}
	return out, nil
}

// TracesForSession returns all traces for a session key, ordered by started_at.
func (s *EntStore) TracesForSession(ctx context.Context, sessionKey string) ([]Trace, error) {
	if s == nil || s.client == nil {
		return nil, nil
	}
	rows, err := s.client.TurnTrace.Query().
		Where(entturntrace.SessionKey(sessionKey)).
		Order(ent.Asc(entturntrace.FieldStartedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query traces for session %q: %w", sessionKey, err)
	}
	out := make([]Trace, 0, len(rows))
	for _, row := range rows {
		out = append(out, entToTrace(row))
	}
	return out, nil
}

// PurgeTraces deletes traces and their associated events by trace IDs.
func (s *EntStore) PurgeTraces(ctx context.Context, traceIDs []string) error {
	if s == nil || s.client == nil || len(traceIDs) == 0 {
		return nil
	}
	// Delete events first (no FK cascade in ent).
	if _, err := s.client.TurnTraceEvent.Delete().
		Where(entturntraceevent.TraceIDIn(traceIDs...)).
		Exec(ctx); err != nil {
		return fmt.Errorf("purge trace events: %w", err)
	}
	if _, err := s.client.TurnTrace.Delete().
		Where(entturntrace.TraceIDIn(traceIDs...)).
		Exec(ctx); err != nil {
		return fmt.Errorf("purge traces: %w", err)
	}
	return nil
}

// TraceCount returns the total number of traces.
func (s *EntStore) TraceCount(ctx context.Context) (int, error) {
	if s == nil || s.client == nil {
		return 0, nil
	}
	return s.client.TurnTrace.Query().Count(ctx)
}

// OldTraces returns trace IDs older than the given cutoff.
func (s *EntStore) OldTraces(ctx context.Context, cutoff time.Time, onlySuccess bool, limit int) ([]string, error) {
	if s == nil || s.client == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 100
	}
	q := s.client.TurnTrace.Query().
		Where(entturntrace.StartedAtLT(cutoff))
	if onlySuccess {
		q = q.Where(entturntrace.Outcome(string(OutcomeSuccess)))
	}
	rows, err := q.
		Order(ent.Asc(entturntrace.FieldStartedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query old traces: %w", err)
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.TraceID)
	}
	return ids, nil
}

// RecentByOutcome returns traces matching outcome within a time window.
func (s *EntStore) RecentByOutcome(ctx context.Context, outcome Outcome, since time.Time, limit int) ([]Trace, error) {
	if s == nil || s.client == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.client.TurnTrace.Query().
		Where(
			entturntrace.Outcome(string(outcome)),
			entturntrace.StartedAtGTE(since),
		).
		Order(ent.Desc(entturntrace.FieldStartedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query recent by outcome %q: %w", outcome, err)
	}
	out := make([]Trace, 0, len(rows))
	for _, row := range rows {
		out = append(out, entToTrace(row))
	}
	return out, nil
}

func entToEvent(row *ent.TurnTraceEvent) Event {
	return Event{
		TraceID:          row.TraceID,
		Seq:              row.Seq,
		EventType:        row.EventType,
		AgentName:        row.AgentName,
		ToolName:         row.ToolName,
		CallSignature:    row.CallSignature,
		PayloadJSON:      row.PayloadJSON,
		PayloadTruncated: row.PayloadTruncated,
		CreatedAt:        row.CreatedAt,
	}
}

func entToTrace(row *ent.TurnTrace) Trace {
	trace := Trace{
		TraceID:    row.TraceID,
		SessionKey: row.SessionKey,
		Entrypoint: row.Entrypoint,
		Outcome:    Outcome(row.Outcome),
		ErrorCode:  row.ErrorCode,
		CauseClass: row.CauseClass,
		CauseDetail: row.CauseDetail,
		Summary:    row.Summary,
		StartedAt:  row.StartedAt,
	}
	if row.EndedAt != nil {
		trace.EndedAt = row.EndedAt
	}
	return trace
}
