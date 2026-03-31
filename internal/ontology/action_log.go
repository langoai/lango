package ontology

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/actionlog"
)

// ActionLogStore provides Ent-backed CRUD for action execution records.
type ActionLogStore struct {
	client *ent.Client
}

// NewActionLogStore creates a new ActionLogStore.
func NewActionLogStore(client *ent.Client) *ActionLogStore {
	return &ActionLogStore{client: client}
}

// Create inserts a new action log with status "started".
func (s *ActionLogStore) Create(ctx context.Context, actionName, principal string, params map[string]string) (uuid.UUID, error) {
	rec, err := s.client.ActionLog.Create().
		SetActionName(actionName).
		SetPrincipal(principal).
		SetParams(params).
		SetStatus(actionlog.StatusStarted).
		Save(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create action log: %w", err)
	}
	return rec.ID, nil
}

// Complete updates an action log to "completed" with effects.
func (s *ActionLogStore) Complete(ctx context.Context, id uuid.UUID, effects *ActionEffects) error {
	now := time.Now()
	builder := s.client.ActionLog.UpdateOneID(id).
		SetStatus(actionlog.StatusCompleted).
		SetCompletedAt(now)
	if effects != nil {
		if m, err := effectsToMap(effects); err == nil {
			builder = builder.SetEffects(m)
		}
	}
	return builder.Exec(ctx)
}

// Fail updates an action log to "failed" with error message.
func (s *ActionLogStore) Fail(ctx context.Context, id uuid.UUID, errMsg string) error {
	now := time.Now()
	return s.client.ActionLog.UpdateOneID(id).
		SetStatus(actionlog.StatusFailed).
		SetErrorMessage(errMsg).
		SetCompletedAt(now).
		Exec(ctx)
}

// Compensated updates an action log to "compensated".
func (s *ActionLogStore) Compensated(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return s.client.ActionLog.UpdateOneID(id).
		SetStatus(actionlog.StatusCompensated).
		SetCompletedAt(now).
		Exec(ctx)
}

// Get retrieves a single action log by ID.
func (s *ActionLogStore) Get(ctx context.Context, id uuid.UUID) (*ActionLogEntry, error) {
	rec, err := s.client.ActionLog.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get action log: %w", err)
	}
	return toLogEntry(rec), nil
}

// List retrieves action logs filtered by action name, ordered by started_at desc.
func (s *ActionLogStore) List(ctx context.Context, actionName string, limit int) ([]ActionLogEntry, error) {
	query := s.client.ActionLog.Query().
		Order(ent.Desc(actionlog.FieldStartedAt))
	if actionName != "" {
		query = query.Where(actionlog.ActionNameEQ(actionName))
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	recs, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list action logs: %w", err)
	}
	entries := make([]ActionLogEntry, len(recs))
	for i, r := range recs {
		entries[i] = *toLogEntry(r)
	}
	return entries, nil
}

func toLogEntry(rec *ent.ActionLog) *ActionLogEntry {
	entry := &ActionLogEntry{
		ID:         rec.ID,
		ActionName: rec.ActionName,
		Principal:  rec.Principal,
		Params:     rec.Params,
		Status:     ActionStatus(rec.Status),
		StartedAt:  rec.StartedAt,
		CompletedAt: rec.CompletedAt,
	}
	if rec.ErrorMessage != nil {
		entry.Error = *rec.ErrorMessage
	}
	if rec.Effects != nil {
		entry.Effects = mapToEffects(rec.Effects)
	}
	return entry
}

func effectsToMap(e *ActionEffects) (map[string]any, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func mapToEffects(m map[string]any) *ActionEffects {
	data, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	var e ActionEffects
	if err := json.Unmarshal(data, &e); err != nil {
		return nil
	}
	return &e
}
