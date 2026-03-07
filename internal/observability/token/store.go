package token

import (
	"context"
	"time"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/tokenusage"
	"github.com/langoai/lango/internal/observability"
)

// EntTokenStore persists token usage data via Ent.
type EntTokenStore struct {
	client *ent.Client
}

// NewEntTokenStore creates a new EntTokenStore.
func NewEntTokenStore(client *ent.Client) *EntTokenStore {
	return &EntTokenStore{client: client}
}

// Save persists a token usage record.
func (s *EntTokenStore) Save(usage observability.TokenUsage) error {
	_, err := s.client.TokenUsage.Create().
		SetSessionKey(usage.SessionKey).
		SetProvider(usage.Provider).
		SetModel(usage.Model).
		SetAgentName(usage.AgentName).
		SetInputTokens(usage.InputTokens).
		SetOutputTokens(usage.OutputTokens).
		SetTotalTokens(usage.TotalTokens).
		SetCacheTokens(usage.CacheTokens).
		SetTimestamp(usage.Timestamp).
		Save(context.Background())
	return err
}

// QueryBySession returns all token usage records for a session.
func (s *EntTokenStore) QueryBySession(ctx context.Context, sessionKey string) ([]observability.TokenUsage, error) {
	rows, err := s.client.TokenUsage.Query().
		Where(tokenusage.SessionKeyEQ(sessionKey)).
		Order(ent.Desc(tokenusage.FieldTimestamp)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toTokenUsages(rows), nil
}

// QueryByAgent returns all token usage records for an agent.
func (s *EntTokenStore) QueryByAgent(ctx context.Context, agentName string) ([]observability.TokenUsage, error) {
	rows, err := s.client.TokenUsage.Query().
		Where(tokenusage.AgentNameEQ(agentName)).
		Order(ent.Desc(tokenusage.FieldTimestamp)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toTokenUsages(rows), nil
}

// QueryByTimeRange returns all token usage records within a time range.
func (s *EntTokenStore) QueryByTimeRange(ctx context.Context, from, to time.Time) ([]observability.TokenUsage, error) {
	rows, err := s.client.TokenUsage.Query().
		Where(
			tokenusage.TimestampGTE(from),
			tokenusage.TimestampLTE(to),
		).
		Order(ent.Desc(tokenusage.FieldTimestamp)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toTokenUsages(rows), nil
}

// AggregateResult holds aggregated token usage data.
type AggregateResult struct {
	TotalInput  int64
	TotalOutput int64
	TotalTokens int64
	RecordCount int
}

// Aggregate returns aggregated stats for all records.
func (s *EntTokenStore) Aggregate(ctx context.Context) (*AggregateResult, error) {
	rows, err := s.client.TokenUsage.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	result := &AggregateResult{RecordCount: len(rows)}
	for _, r := range rows {
		result.TotalInput += r.InputTokens
		result.TotalOutput += r.OutputTokens
		result.TotalTokens += r.TotalTokens
	}
	return result, nil
}

// Cleanup deletes records older than retentionDays.
func (s *EntTokenStore) Cleanup(ctx context.Context, retentionDays int) (int, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	count, err := s.client.TokenUsage.Delete().
		Where(tokenusage.TimestampLT(cutoff)).
		Exec(ctx)
	return count, err
}

func toTokenUsages(rows []*ent.TokenUsage) []observability.TokenUsage {
	out := make([]observability.TokenUsage, len(rows))
	for i, r := range rows {
		out[i] = observability.TokenUsage{
			Provider:     r.Provider,
			Model:        r.Model,
			SessionKey:   r.SessionKey,
			AgentName:    r.AgentName,
			InputTokens:  r.InputTokens,
			OutputTokens: r.OutputTokens,
			TotalTokens:  r.TotalTokens,
			CacheTokens:  r.CacheTokens,
			Timestamp:    r.Timestamp,
		}
	}
	return out
}
