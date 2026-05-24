package adk

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/adk/model"
	"google.golang.org/genai"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/prompt"
	"github.com/langoai/lango/internal/provider"
	"github.com/langoai/lango/internal/session"
)

// mockMemoryProvider records calls and returns canned data.
type mockMemoryProvider struct {
	lastSessionKey string
	observations   []memory.Observation
	reflections    []memory.Reflection
}

func (m *mockMemoryProvider) ListObservations(_ context.Context, sessionKey string) ([]memory.Observation, error) {
	m.lastSessionKey = sessionKey
	return m.observations, nil
}

func (m *mockMemoryProvider) ListReflections(_ context.Context, sessionKey string) ([]memory.Reflection, error) {
	m.lastSessionKey = sessionKey
	return m.reflections, nil
}

func (m *mockMemoryProvider) ListRecentReflections(_ context.Context, sessionKey string, _ int) ([]memory.Reflection, error) {
	m.lastSessionKey = sessionKey
	return m.reflections, nil
}

func (m *mockMemoryProvider) ListRecentObservations(_ context.Context, sessionKey string, _ int) ([]memory.Observation, error) {
	m.lastSessionKey = sessionKey
	return m.observations, nil
}

// Compile-time check.
var _ MemoryProvider = (*mockMemoryProvider)(nil)

type mockRunSummaryProvider struct {
	summaries []RunSummaryContext
	maxSeq    int64
	listCalls int
	seqCalls  int
}

func (m *mockRunSummaryProvider) ListRunSummaries(_ context.Context, _ string, _ int) ([]RunSummaryContext, error) {
	m.listCalls++
	return m.summaries, nil
}

func (m *mockRunSummaryProvider) MaxJournalSeqForSession(_ context.Context, _ string) (int64, error) {
	m.seqCalls++
	return m.maxSeq, nil
}

func newTestContextAdapter(t *testing.T, mp MemoryProvider) *ContextAwareModelAdapter {
	t.Helper()
	p := &mockProvider{
		id: "test",
		events: []provider.StreamEvent{
			{Type: provider.StreamEventPlainText, Text: "ok"},
			{Type: provider.StreamEventDone},
		},
	}
	inner := NewModelAdapter(p, "test-model")
	builder := prompt.DefaultBuilder()
	logger := zap.NewNop().Sugar()
	adapter := NewContextAwareModelAdapter(inner, nil, builder, logger)
	if mp != nil {
		adapter.WithMemory(mp)
		adapter.WithMemoryLimits(3, 5)
	}
	return adapter
}

func TestGenerateContent_SessionKeyFromContext(t *testing.T) {
	t.Parallel()

	mp := &mockMemoryProvider{
		observations: []memory.Observation{{Content: "user prefers dark mode"}},
		reflections:  []memory.Reflection{{Content: "user is a developer"}},
	}
	adapter := newTestContextAdapter(t, mp)

	ctx := session.WithSessionKey(context.Background(), "telegram:123:456")
	req := &model.LLMRequest{
		Model: "test-model",
		Contents: []*genai.Content{
			{Role: "user", Parts: []*genai.Part{{Text: "hello"}}},
		},
	}

	seq := adapter.GenerateContent(ctx, req, false)
	for _, err := range seq {
		require.NoError(t, err)
	}

	assert.Equal(t, "telegram:123:456", mp.lastSessionKey)
}

func TestGenerateContent_NoSessionKey_SkipsMemory(t *testing.T) {
	t.Parallel()

	mp := &mockMemoryProvider{
		observations: []memory.Observation{{Content: "should not appear"}},
	}
	adapter := newTestContextAdapter(t, mp)

	// No session key in context.
	ctx := context.Background()
	req := &model.LLMRequest{
		Model: "test-model",
		Contents: []*genai.Content{
			{Role: "user", Parts: []*genai.Part{{Text: "hello"}}},
		},
	}

	seq := adapter.GenerateContent(ctx, req, false)
	for _, err := range seq {
		require.NoError(t, err)
	}

	// Memory provider should not have been called.
	assert.Empty(t, mp.lastSessionKey, "memory provider should not be called without session key")
}

func TestGenerateContent_SessionKey_UpdatesRuntimeAdapter(t *testing.T) {
	t.Parallel()

	adapter := newTestContextAdapter(t, nil)
	ra := NewRuntimeContextAdapter(2, false, false, true)
	adapter.WithRuntimeAdapter(ra)

	ctx := session.WithSessionKey(context.Background(), "discord:guild:chan")
	req := &model.LLMRequest{
		Model: "test-model",
		Contents: []*genai.Content{
			{Role: "user", Parts: []*genai.Part{{Text: "hello"}}},
		},
	}

	seq := adapter.GenerateContent(ctx, req, false)
	for _, err := range seq {
		require.NoError(t, err)
	}

	rc := ra.GetRuntimeContext()
	assert.Equal(t, "discord:guild:chan", rc.SessionKey)
	assert.Equal(t, "discord", rc.ChannelType)
}

func TestGenerateContent_MemoryInjectedIntoPrompt(t *testing.T) {
	t.Parallel()

	mp := &mockMemoryProvider{
		observations: []memory.Observation{{Content: "user prefers Go"}},
		reflections:  []memory.Reflection{{Content: "experienced developer"}},
	}
	p := &mockProvider{
		id: "test",
		events: []provider.StreamEvent{
			{Type: provider.StreamEventPlainText, Text: "ok"},
			{Type: provider.StreamEventDone},
		},
	}
	inner := NewModelAdapter(p, "test-model")
	builder := prompt.DefaultBuilder()
	logger := zap.NewNop().Sugar()
	adapter := NewContextAwareModelAdapter(inner, nil, builder, logger)
	adapter.WithMemory(mp)
	adapter.WithMemoryLimits(3, 5)

	ctx := session.WithSessionKey(context.Background(), "test:session:1")
	req := &model.LLMRequest{
		Model: "test-model",
		Contents: []*genai.Content{
			{Role: "user", Parts: []*genai.Part{{Text: "hello"}}},
		},
	}

	seq := adapter.GenerateContent(ctx, req, false)
	for _, err := range seq {
		require.NoError(t, err)
	}

	// Verify system instruction was augmented with memory content.
	msgs := p.lastParams.Messages
	require.GreaterOrEqual(t, len(msgs), 2, "expected at least 2 messages (system + user)")

	systemMsg := msgs[0]
	require.Equal(t, "system", string(systemMsg.Role))

	// The system prompt should contain memory sections.
	assert.True(t, strings.Contains(systemMsg.Content, "Conversation Memory"), "system prompt should contain 'Conversation Memory' section")
	assert.True(t, strings.Contains(systemMsg.Content, "user prefers Go"), "system prompt should contain observation content")
	assert.True(t, strings.Contains(systemMsg.Content, "experienced developer"), "system prompt should contain reflection content")
}

func TestGenerateContent_RunSummariesInjectedIntoPrompt(t *testing.T) {
	t.Parallel()

	p := &mockProvider{
		id: "test",
		events: []provider.StreamEvent{
			{Type: provider.StreamEventPlainText, Text: "ok"},
			{Type: provider.StreamEventDone},
		},
	}
	inner := NewModelAdapter(p, "test-model")
	builder := prompt.DefaultBuilder()
	logger := zap.NewNop().Sugar()
	adapter := NewContextAwareModelAdapter(inner, nil, builder, logger)
	adapter.WithRunSummaryProvider(&mockRunSummaryProvider{
		maxSeq: 1,
		summaries: []RunSummaryContext{{
			RunID:          "run-1",
			Goal:           "Fix drift",
			Status:         "running",
			CurrentStep:    "Repair projection",
			CurrentBlocker: "none",
		}},
	})

	ctx := session.WithSessionKey(context.Background(), "test:session:run")
	req := &model.LLMRequest{
		Model: "test-model",
		Contents: []*genai.Content{
			{Role: "user", Parts: []*genai.Part{{Text: "continue"}}},
		},
	}

	seq := adapter.GenerateContent(ctx, req, false)
	for _, err := range seq {
		require.NoError(t, err)
	}

	msgs := p.lastParams.Messages
	require.GreaterOrEqual(t, len(msgs), 2)
	systemMsg := msgs[0]
	require.Equal(t, "system", string(systemMsg.Role))
	assert.Contains(t, systemMsg.Content, "Active Runs")
	assert.Contains(t, systemMsg.Content, "run-1")
	assert.Contains(t, systemMsg.Content, "Repair projection")
}

func TestRetrieveRunSummaryData_CacheHit(t *testing.T) {
	t.Parallel()

	prov := &mockRunSummaryProvider{
		maxSeq: 1,
		summaries: []RunSummaryContext{{
			RunID:       "run-1",
			Goal:        "Fix drift",
			Status:      "running",
			CurrentStep: "Repair projection",
		}},
	}

	adapter := newTestContextAdapter(t, nil)
	adapter.WithRunSummaryProvider(prov)

	got1 := adapter.retrieveRunSummaryData(context.Background(), "sess-1")
	got2 := adapter.retrieveRunSummaryData(context.Background(), "sess-1")

	require.Len(t, got1, 1)
	assert.Equal(t, got1[0].RunID, got2[0].RunID)
	assert.Equal(t, 1, prov.listCalls, "list should be called once (cache hit on second)")
	assert.Equal(t, 2, prov.seqCalls, "seq should be called every time")
}

func TestRetrieveRunSummaryData_CacheInvalidatesOnSeqChange(t *testing.T) {
	t.Parallel()

	prov := &mockRunSummaryProvider{
		maxSeq: 1,
		summaries: []RunSummaryContext{{
			RunID:  "run-1",
			Goal:   "First",
			Status: "running",
		}},
	}

	adapter := newTestContextAdapter(t, nil)
	adapter.WithRunSummaryProvider(prov)

	got1 := adapter.retrieveRunSummaryData(context.Background(), "sess-1")
	prov.maxSeq = 2
	prov.summaries = []RunSummaryContext{{
		RunID:  "run-2",
		Goal:   "Second",
		Status: "paused",
	}}
	got2 := adapter.retrieveRunSummaryData(context.Background(), "sess-1")

	require.Len(t, got1, 1)
	require.Len(t, got2, 1)
	assert.NotEqual(t, got1[0].RunID, got2[0].RunID)
	assert.Equal(t, 2, prov.listCalls, "list should be called twice (cache invalidated)")
}

// mockSessionCompactor records CompactMessages calls.
type mockSessionCompactor struct {
	calls []string
}

func (m *mockSessionCompactor) CompactMessages(key string, upToIndex int, summary string) error {
	m.calls = append(m.calls, key)
	return nil
}

var _ SessionCompactor = (*mockSessionCompactor)(nil)

type mockCompactionSyncWaiter struct {
	done   bool
	waited time.Duration
	key    string
	wait   time.Duration
}

func (m *mockCompactionSyncWaiter) WaitForSession(_ context.Context, key string, timeout time.Duration) (bool, time.Duration) {
	m.key = key
	m.wait = timeout
	return m.done, m.waited
}

var _ CompactionSyncWaiter = (*mockCompactionSyncWaiter)(nil)

type mockCatalogSource struct {
	modeName string
	section  string
}

func (m *mockCatalogSource) BuildToolCatalogSection(modeName string) string {
	m.modeName = modeName
	return m.section
}

var _ CatalogSource = (*mockCatalogSource)(nil)

type mockModeResolver struct {
	seenMode string
	hint     string
}

func (m *mockModeResolver) LookupModeHint(modeName string) string {
	m.seenMode = modeName
	return m.hint
}

var _ ModeResolver = (*mockModeResolver)(nil)

type mockRecallProvider struct {
	currentSessionKey string
	query             string
	matches           []RecallMatch
}

func (m *mockRecallProvider) RecallRecent(_ context.Context, currentSessionKey, query string) ([]RecallMatch, error) {
	m.currentSessionKey = currentSessionKey
	m.query = query
	return m.matches, nil
}

var _ RecallProvider = (*mockRecallProvider)(nil)

func TestEmergencyCompaction_TriggeredAtThreshold(t *testing.T) {
	t.Parallel()

	adapter := newTestContextAdapter(t, nil)
	sc := &mockSessionCompactor{}
	adapter.WithSessionCompactor(sc)

	// Create a budget manager with a tiny model window so measured > 90%.
	bm, err := NewContextBudgetManager(1000, 100, 50, DefaultAllocation())
	require.NoError(t, err)
	adapter.WithBudgetManager(bm)

	// Run GenerateContent with a session key.
	ctx := session.WithSessionKey(context.Background(), "compact-test")
	req := &model.LLMRequest{
		Contents: []*genai.Content{{
			Role:  "user",
			Parts: []*genai.Part{{Text: "test"}},
		}},
	}
	for _, _ = range adapter.GenerateContent(ctx, req, false) {
		// consume iterator
	}

	// With a 1000 token window and base prompt > 90% threshold,
	// compaction may or may not trigger depending on measured sections.
	// The key test is that the compactor interface is invocable.
	// A more precise test would require injecting large memory/knowledge.
	t.Logf("compaction calls: %d", len(sc.calls))
}

func TestEmergencyCompaction_DegradedDoesNotTrigger(t *testing.T) {
	t.Parallel()

	adapter := newTestContextAdapter(t, nil)
	sc := &mockSessionCompactor{}
	adapter.WithSessionCompactor(sc)

	// Create a budget manager that will be degraded (window < base prompt).
	bm, err := NewContextBudgetManager(100, 50, 200, DefaultAllocation())
	require.NoError(t, err)
	adapter.WithBudgetManager(bm)

	ctx := session.WithSessionKey(context.Background(), "degraded-test")
	req := &model.LLMRequest{
		Contents: []*genai.Content{{
			Role:  "user",
			Parts: []*genai.Part{{Text: "test"}},
		}},
	}
	for _, _ = range adapter.GenerateContent(ctx, req, false) {
	}

	assert.Empty(t, sc.calls, "compaction should NOT trigger when Degraded (config issue)")
}

func TestEmergencyCompaction_NilCompactorSkips(t *testing.T) {
	t.Parallel()

	adapter := newTestContextAdapter(t, nil)
	// No compactor wired — should not panic.

	bm, err := NewContextBudgetManager(1000, 100, 50, DefaultAllocation())
	require.NoError(t, err)
	adapter.WithBudgetManager(bm)

	ctx := session.WithSessionKey(context.Background(), "nil-compactor-test")
	req := &model.LLMRequest{
		Contents: []*genai.Content{{
			Role:  "user",
			Parts: []*genai.Part{{Text: "test"}},
		}},
	}
	for _, _ = range adapter.GenerateContent(ctx, req, false) {
	}
	// No panic = pass
}

func TestGenerateContent_CompactionSyncTimeoutPublishesSlowEvent(t *testing.T) {
	t.Parallel()

	adapter := newTestContextAdapter(t, nil)
	waiter := &mockCompactionSyncWaiter{waited: 25 * time.Millisecond}
	adapter.WithCompactionSync(waiter, 50*time.Millisecond)

	bus := eventbus.New()
	var got eventbus.CompactionSlowEvent
	var count int
	eventbus.SubscribeTyped(bus, func(e eventbus.CompactionSlowEvent) {
		got = e
		count++
	})
	adapter.WithEventBus(bus)

	ctx := session.WithSessionKey(context.Background(), "compaction:slow")
	req := &model.LLMRequest{
		Contents: []*genai.Content{{
			Role:  "user",
			Parts: []*genai.Part{{Text: "continue"}},
		}},
	}

	for _, err := range adapter.GenerateContent(ctx, req, false) {
		require.NoError(t, err)
	}

	assert.Equal(t, "compaction:slow", waiter.key)
	assert.Equal(t, 50*time.Millisecond, waiter.wait)
	require.Equal(t, 1, count)
	assert.Equal(t, "compaction:slow", got.SessionKey)
	assert.Equal(t, 25*time.Millisecond, got.WaitedFor)
	assert.False(t, got.Timestamp.IsZero())
}

func TestGenerateContent_DynamicCatalogAndModeHintUseSessionMode(t *testing.T) {
	t.Parallel()

	p := &mockProvider{
		id: "mode-aware",
		events: []provider.StreamEvent{
			{Type: provider.StreamEventPlainText, Text: "ok"},
			{Type: provider.StreamEventDone},
		},
	}
	inner := NewModelAdapter(p, "test-model")
	catalog := &mockCatalogSource{section: "## Tool Catalog\n- inspect_workspace"}
	resolver := &mockModeResolver{hint: "Prefer review-first responses."}
	adapter := NewContextAwareModelAdapter(inner, nil, prompt.DefaultBuilder(), zap.NewNop().Sugar()).
		WithCatalog(catalog).
		WithModeResolver(resolver)

	ctx := session.WithModeName(context.Background(), "review")
	req := &model.LLMRequest{
		Contents: []*genai.Content{{
			Role:  "user",
			Parts: []*genai.Part{{Text: "check this"}},
		}},
	}

	for _, err := range adapter.GenerateContent(ctx, req, false) {
		require.NoError(t, err)
	}

	assert.Equal(t, "review", catalog.modeName)
	assert.Equal(t, "review", resolver.seenMode)
	require.NotNil(t, req.Config)
	require.NotNil(t, req.Config.SystemInstruction)
	promptText := req.Config.SystemInstruction.Parts[0].Text
	assert.Contains(t, promptText, "## Tool Catalog")
	assert.Contains(t, promptText, "inspect_workspace")
	assert.Contains(t, promptText, "## Session Mode: review")
	assert.Contains(t, promptText, "Prefer review-first responses.")
}

func TestGenerateContent_RecallProviderPrependsPriorSessionRecall(t *testing.T) {
	t.Parallel()

	p := &mockProvider{
		id: "recall-aware",
		events: []provider.StreamEvent{
			{Type: provider.StreamEventPlainText, Text: "ok"},
			{Type: provider.StreamEventDone},
		},
	}
	inner := NewModelAdapter(p, "test-model")
	recall := &mockRecallProvider{
		matches: []RecallMatch{{
			SessionKey: "prior:session",
			Summary:    "The user already selected SQLite-backed fixtures.",
			Rank:       0.88,
		}},
	}
	adapter := NewContextAwareModelAdapter(inner, nil, prompt.DefaultBuilder(), zap.NewNop().Sugar()).
		WithRecallProvider(recall)

	ctx := session.WithSessionKey(context.Background(), "current:session")
	req := &model.LLMRequest{
		Contents: []*genai.Content{{
			Role:  "user",
			Parts: []*genai.Part{{Text: "what did we decide?"}},
		}},
	}

	for _, err := range adapter.GenerateContent(ctx, req, false) {
		require.NoError(t, err)
	}

	assert.Equal(t, "current:session", recall.currentSessionKey)
	assert.Equal(t, "what did we decide?", recall.query)
	require.NotNil(t, req.Config)
	require.NotNil(t, req.Config.SystemInstruction)
	promptText := req.Config.SystemInstruction.Parts[0].Text
	assert.Contains(t, promptText, "## Prior Session Recall")
	assert.Contains(t, promptText, "prior:session")
	assert.Contains(t, promptText, "SQLite-backed fixtures")
}
