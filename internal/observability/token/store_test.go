package token

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/ent/tokenusage"
	"github.com/langoai/lango/internal/observability"

	_ "github.com/mattn/go-sqlite3"
)

func TestEntTokenStore_SavePersistsTokenUsage(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	store := NewEntTokenStore(client)
	ts := time.Date(2026, 5, 20, 10, 11, 12, 0, time.UTC)

	usage := observability.TokenUsage{
		Provider:     "openai",
		Model:        "gpt-4o-mini",
		SessionKey:   "session-save",
		AgentName:    "writer",
		InputTokens:  11,
		OutputTokens: 22,
		TotalTokens:  33,
		CacheTokens:  4,
		Timestamp:    ts,
	}

	require.NoError(t, store.Save(usage))

	row, err := client.TokenUsage.Query().Only(ctx)
	require.NoError(t, err)
	require.Equal(t, usage.Provider, row.Provider)
	require.Equal(t, usage.Model, row.Model)
	require.Equal(t, usage.SessionKey, row.SessionKey)
	require.Equal(t, usage.AgentName, row.AgentName)
	require.Equal(t, usage.InputTokens, row.InputTokens)
	require.Equal(t, usage.OutputTokens, row.OutputTokens)
	require.Equal(t, usage.TotalTokens, row.TotalTokens)
	require.Equal(t, usage.CacheTokens, row.CacheTokens)
	require.True(t, usage.Timestamp.Equal(row.Timestamp))
}

func TestEntTokenStore_QueryFiltersAndOrdersByTimestamp(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	store := NewEntTokenStore(client)
	base := time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC)

	insertTokenUsages(t, store,
		tokenUsageFixture("session-a", "agent-a", base.Add(1*time.Hour), 10),
		tokenUsageFixture("session-a", "agent-b", base.Add(3*time.Hour), 30),
		tokenUsageFixture("session-b", "agent-a", base.Add(2*time.Hour), 20),
		tokenUsageFixture("session-a", "agent-a", base.Add(4*time.Hour), 40),
	)

	sessionRows, err := store.QueryBySession(ctx, "session-a")
	require.NoError(t, err)
	require.Equal(t, []int64{40, 30, 10}, totalTokens(sessionRows))

	agentRows, err := store.QueryByAgent(ctx, "agent-a")
	require.NoError(t, err)
	require.Equal(t, []int64{40, 20, 10}, totalTokens(agentRows))

	rangeRows, err := store.QueryByTimeRange(ctx, base.Add(2*time.Hour), base.Add(3*time.Hour))
	require.NoError(t, err)
	require.Equal(t, []int64{30, 20}, totalTokens(rangeRows))
}

func TestEntTokenStore_AggregateSumsAllRecords(t *testing.T) {
	ctx := context.Background()
	store := NewEntTokenStore(testEntClient(t))
	base := time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC)

	insertTokenUsages(t, store,
		observability.TokenUsage{
			Provider:     "openai",
			Model:        "gpt-4o-mini",
			SessionKey:   "session-a",
			AgentName:    "agent-a",
			InputTokens:  7,
			OutputTokens: 11,
			TotalTokens:  18,
			CacheTokens:  2,
			Timestamp:    base,
		},
		observability.TokenUsage{
			Provider:     "anthropic",
			Model:        "claude-3-5-haiku",
			SessionKey:   "session-b",
			AgentName:    "agent-b",
			InputTokens:  13,
			OutputTokens: 17,
			TotalTokens:  30,
			CacheTokens:  3,
			Timestamp:    base.Add(time.Minute),
		},
	)

	got, err := store.Aggregate(ctx)
	require.NoError(t, err)
	require.Equal(t, &AggregateResult{
		TotalInput:  20,
		TotalOutput: 28,
		TotalTokens: 48,
		RecordCount: 2,
	}, got)
}

func TestEntTokenStore_CleanupDeletesOnlyRowsOlderThanRetention(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	store := NewEntTokenStore(client)
	now := time.Now().UTC()

	insertTokenUsages(t, store,
		tokenUsageFixture("old-session", "agent-a", now.AddDate(0, 0, -10), 10),
		tokenUsageFixture("boundary-session", "agent-a", now.AddDate(0, 0, -2), 20),
		tokenUsageFixture("recent-session", "agent-a", now.AddDate(0, 0, -1), 30),
	)

	deleted, err := store.Cleanup(ctx, 3)
	require.NoError(t, err)
	require.Equal(t, 1, deleted)

	rows, err := client.TokenUsage.Query().
		Order(ent.Asc(tokenusage.FieldTotalTokens)).
		All(ctx)
	require.NoError(t, err)
	require.Equal(t, []int64{20, 30}, entTotalTokens(rows))
}

func TestToTokenUsagesCopiesAllFields(t *testing.T) {
	ts := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	rows := []*ent.TokenUsage{
		{
			Provider:     "openai",
			Model:        "gpt-4o",
			SessionKey:   "session-a",
			AgentName:    "agent-a",
			InputTokens:  1,
			OutputTokens: 2,
			TotalTokens:  3,
			CacheTokens:  4,
			Timestamp:    ts,
		},
	}

	got := toTokenUsages(rows)

	require.Equal(t, []observability.TokenUsage{
		{
			Provider:     "openai",
			Model:        "gpt-4o",
			SessionKey:   "session-a",
			AgentName:    "agent-a",
			InputTokens:  1,
			OutputTokens: 2,
			TotalTokens:  3,
			CacheTokens:  4,
			Timestamp:    ts,
		},
	}, got)
	require.Empty(t, toTokenUsages(nil))
}

func insertTokenUsages(t *testing.T, store *EntTokenStore, usages ...observability.TokenUsage) {
	t.Helper()
	for _, usage := range usages {
		require.NoError(t, store.Save(usage))
	}
}

func testEntClient(t *testing.T) *ent.Client {
	t.Helper()
	name := strings.NewReplacer("/", "-", " ", "-", ":", "-").Replace(t.Name())
	client := enttest.Open(t, "sqlite3", "file:"+name+"?mode=memory&_fk=1")
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})
	return client
}

func tokenUsageFixture(
	sessionKey string,
	agentName string,
	timestamp time.Time,
	totalTokens int64,
) observability.TokenUsage {
	return observability.TokenUsage{
		Provider:     "openai",
		Model:        "gpt-4o-mini",
		SessionKey:   sessionKey,
		AgentName:    agentName,
		InputTokens:  totalTokens,
		OutputTokens: totalTokens * 2,
		TotalTokens:  totalTokens,
		CacheTokens:  totalTokens / 10,
		Timestamp:    timestamp,
	}
}

func totalTokens(usages []observability.TokenUsage) []int64 {
	out := make([]int64, len(usages))
	for i, usage := range usages {
		out[i] = usage.TotalTokens
	}
	return out
}

func entTotalTokens(rows []*ent.TokenUsage) []int64 {
	out := make([]int64, len(rows))
	for i, row := range rows {
		out[i] = row.TotalTokens
	}
	return out
}
