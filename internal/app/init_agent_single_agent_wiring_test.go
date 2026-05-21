package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/supervisor"
	"github.com/langoai/lango/internal/testutil"
)

func TestInitAgentSingleAgentRunsWithLocalProviderFixture(t *testing.T) {
	t.Parallel()

	fixture := newInitAgentOpenAITestServer(t, "single-agent-ok")

	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "openai"
	cfg.Agent.Model = "gpt-test"
	cfg.Providers = map[string]config.ProviderConfig{
		"openai": {
			Type:    "openai",
			APIKey:  "test-key",
			BaseURL: fixture.server.URL,
		},
	}
	cfg.Agent.MaxTurns = 7
	cfg.Agent.ToolTimeout = 25 * time.Millisecond
	cfg.Agent.Temperature = 0.1
	cfg.Agent.MaxTokens = 64
	cfg.Agent.FallbackProvider = "fallback-openai"
	cfg.Agent.FallbackModel = "fallback-model"
	cfg.Cron.Enabled = true
	cfg.Background.Enabled = true
	cfg.Workflow.Enabled = true

	sv, err := supervisor.New(cfg)
	require.NoError(t, err)

	tool := &agent.Tool{
		Name:        "init_agent_single_noop",
		Description: "No-op tool for initAgent wiring tests",
		Parameters: map[string]interface{}{
			"type":                 "object",
			"properties":           map[string]interface{}{},
			"additionalProperties": false,
		},
		Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"ok": true}, nil
		},
		SafetyLevel: agent.SafetyLevelSafe,
	}

	got, err := initAgent(context.Background(), &agentDeps{
		sv:       sv,
		cfg:      cfg,
		store:    &stubSessionStore{},
		tools:    []*agent.Tool{tool},
		scanner:  agent.NewSecretScanner(),
		eventBus: eventbus.New(),
	})

	require.NoError(t, err)
	require.NotNil(t, got)

	response, err := got.RunAndCollect(context.Background(), "init-agent-single", "hello single")
	require.NoError(t, err)
	require.Equal(t, "single-agent-ok", response)

	requests := fixture.requests()
	require.Len(t, requests, 1)
	require.Equal(t, "gpt-test", requests[0]["model"])
	require.Equal(t, true, requests[0]["stream"])
	require.Contains(t, fixture.body(0), "hello single")
	require.Contains(t, fixture.body(0), "init_agent_single_noop")
}

func TestInitAgentMemoryOnlyRunsContextAdapterWithLocalProviderFixture(t *testing.T) {
	t.Parallel()

	fixture := newInitAgentOpenAITestServer(t, "memory-only-ok")

	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "openai"
	cfg.Agent.Model = "gpt-test"
	cfg.Providers = map[string]config.ProviderConfig{
		"openai": {
			Type:    "openai",
			APIKey:  "test-key",
			BaseURL: fixture.server.URL,
		},
	}
	cfg.ObservationalMemory.Enabled = true
	cfg.ObservationalMemory.MaxReflectionsInContext = 0
	cfg.ObservationalMemory.MaxObservationsInContext = 0
	cfg.ObservationalMemory.MemoryTokenBudget = 512
	cfg.RunLedger.Enabled = true
	cfg.RunLedger.AuthoritativeRead = true

	sv, err := supervisor.New(cfg)
	require.NoError(t, err)

	entStore := session.NewEntStoreWithClient(testutil.TestEntClient(t))
	memoryStore := memory.NewStore(entStore.Client(), zap.NewNop().Sugar())
	const sessionKey = "init-agent-memory-only"
	ctx := session.WithSessionKey(context.Background(), sessionKey)
	require.NoError(t, memoryStore.SaveObservation(ctx, memory.Observation{
		SessionKey: sessionKey,
		Content:    "memory-only observation fixture",
		TokenCount: 4,
	}))
	require.NoError(t, memoryStore.SaveReflection(ctx, memory.Reflection{
		SessionKey: sessionKey,
		Content:    "memory-only reflection fixture",
		TokenCount: 4,
	}))

	got, err := initAgent(ctx, &agentDeps{
		sv:       sv,
		cfg:      cfg,
		store:    entStore,
		mc:       &memoryComponents{store: memoryStore},
		scanner:  agent.NewSecretScanner(),
		eventBus: eventbus.New(),
	})

	require.NoError(t, err)
	require.NotNil(t, got)

	response, err := got.RunAndCollect(ctx, sessionKey, "hello memory")
	require.NoError(t, err)
	require.Equal(t, "memory-only-ok", response)

	requests := fixture.requests()
	require.Len(t, requests, 1)
	require.Equal(t, "gpt-test", requests[0]["model"])
	require.Equal(t, true, requests[0]["stream"])
	require.Contains(t, fixture.body(0), "hello memory")
	require.Contains(t, fixture.body(0), "Conversation Memory")
	require.Contains(t, fixture.body(0), "memory-only observation fixture")
	require.Contains(t, fixture.body(0), "memory-only reflection fixture")
}

type initAgentOpenAITestServer struct {
	server *httptest.Server
	mu     sync.Mutex
	bodies []string
	reqs   []map[string]interface{}
}

func newInitAgentOpenAITestServer(t *testing.T, response string) *initAgentOpenAITestServer {
	t.Helper()

	fixture := &initAgentOpenAITestServer{}
	fixture.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/chat/completions", r.URL.Path)
		require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		var payload map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		raw, err := json.Marshal(payload)
		require.NoError(t, err)

		fixture.mu.Lock()
		fixture.reqs = append(fixture.reqs, payload)
		fixture.bodies = append(fixture.bodies, string(raw))
		fixture.mu.Unlock()

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"id\":\"chatcmpl-test\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"" + response + "\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	t.Cleanup(fixture.server.Close)
	return fixture
}

func (s *initAgentOpenAITestServer) requests() []map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]map[string]interface{}, len(s.reqs))
	copy(out, s.reqs)
	return out
}

func (s *initAgentOpenAITestServer) body(index int) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.bodies[index]
}
