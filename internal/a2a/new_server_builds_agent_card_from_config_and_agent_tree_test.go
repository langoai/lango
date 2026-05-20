package a2a

import (
	"encoding/json"
	"iter"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/session"

	"github.com/langoai/lango/internal/config"
)

func TestNewServerBuildsAgentCardFromConfigAndAgentTree(t *testing.T) {
	t.Parallel()

	operator := newNewServerBuildsAgentCardFromConfigAndAgentTreeA2AAgent(t, "operator", "runs shell commands", nil)
	reviewer := newNewServerBuildsAgentCardFromConfigAndAgentTreeA2AAgent(t, "reviewer", "reviews proposed changes", nil)
	root := newNewServerBuildsAgentCardFromConfigAndAgentTreeA2AAgent(t, "root-agent", "coordinates sub-agents", []adkagent.Agent{
		operator,
		reviewer,
	})

	server := NewServer(config.A2AConfig{
		BaseURL:          "https://agents.example/a2a",
		AgentName:        "public-root",
		AgentDescription: "public description",
	}, root, zap.NewNop().Sugar())

	card := server.Card()
	require.NotNil(t, card)
	assert.Equal(t, "public-root", card.Name)
	assert.Equal(t, "public description", card.Description)
	assert.Equal(t, "https://agents.example/a2a", card.URL)
	require.Len(t, card.Skills, 3)
	assert.Equal(t, AgentSkill{
		ID:          "root-agent",
		Name:        "root-agent",
		Description: "coordinates sub-agents",
		Tags:        []string{SkillTagOrchestration},
	}, card.Skills[0])
	assert.Equal(t, AgentSkill{
		ID:          "operator",
		Name:        "operator",
		Description: "runs shell commands",
		Tags:        []string{SkillTagSubAgentPrefix + "root-agent"},
	}, card.Skills[1])
	assert.Equal(t, AgentSkill{
		ID:          "reviewer",
		Name:        "reviewer",
		Description: "reviews proposed changes",
		Tags:        []string{SkillTagSubAgentPrefix + "root-agent"},
	}, card.Skills[2])
}

func TestNewServerFallsBackToAgentMetadata(t *testing.T) {
	t.Parallel()

	root := newNewServerBuildsAgentCardFromConfigAndAgentTreeA2AAgent(t, "metadata-agent", "agent supplied description", nil)

	server := NewServer(config.A2AConfig{
		BaseURL: "https://metadata.example/a2a",
	}, root, zap.NewNop().Sugar())

	card := server.Card()
	require.NotNil(t, card)
	assert.Equal(t, "metadata-agent", card.Name)
	assert.Equal(t, "agent supplied description", card.Description)
	assert.Equal(t, "https://metadata.example/a2a", card.URL)
	require.Len(t, card.Skills, 1)
	assert.Equal(t, "metadata-agent", card.Skills[0].ID)
	assert.Equal(t, []string{SkillTagOrchestration}, card.Skills[0].Tags)
}

func TestSetP2PInfoAndPricingMutateRouteOutput(t *testing.T) {
	t.Parallel()

	root := newNewServerBuildsAgentCardFromConfigAndAgentTreeA2AAgent(t, "priced-agent", "agent with p2p metadata", nil)
	server := NewServer(config.A2AConfig{
		BaseURL: "https://priced.example/a2a",
	}, root, zap.NewNop().Sugar())

	multiaddrs := []string{"/ip4/127.0.0.1/tcp/4001", "/dns4/node.example/tcp/443/wss"}
	capabilities := []string{"chat", "code-review"}
	server.SetP2PInfo("did:example:agent", multiaddrs, capabilities)
	pricing := &PricingInfo{
		Currency:  "USD",
		PerQuery:  "0.05",
		PerMinute: "0.20",
		ToolPrices: map[string]string{
			"review": "0.10",
		},
	}
	server.SetPricing(pricing)

	card := server.Card()
	assert.Equal(t, "did:example:agent", card.DID)
	assert.Equal(t, multiaddrs, card.Multiaddrs)
	assert.Equal(t, capabilities, card.Capabilities)
	assert.Equal(t, pricing, card.Pricing)

	req := httptest.NewRequest(http.MethodGet, AgentCardRoute, nil)
	rec := httptest.NewRecorder()
	server.handleAgentCard(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, ContentTypeJSON, rec.Header().Get("Content-Type"))

	var got AgentCard
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, "did:example:agent", got.DID)
	assert.Equal(t, multiaddrs, got.Multiaddrs)
	assert.Equal(t, capabilities, got.Capabilities)
	require.NotNil(t, got.Pricing)
	assert.Equal(t, "USD", got.Pricing.Currency)
	assert.Equal(t, "0.05", got.Pricing.PerQuery)
	assert.Equal(t, "0.20", got.Pricing.PerMinute)
	assert.Equal(t, map[string]string{"review": "0.10"}, got.Pricing.ToolPrices)
}

func TestRegisterRoutesMountsAgentCardHandlerWithoutListener(t *testing.T) {
	t.Parallel()

	root := newNewServerBuildsAgentCardFromConfigAndAgentTreeA2AAgent(t, "route-agent", "route description", nil)
	server := NewServer(config.A2AConfig{
		BaseURL: "https://route.example/a2a",
	}, root, zap.NewNop().Sugar())
	mux := &newServerBuildsAgentCardFromConfigAndAgentTreeRouteRecorder{routes: make(map[string]http.HandlerFunc)}

	server.RegisterRoutes(mux)

	handler, ok := mux.routes[AgentCardRoute]
	require.True(t, ok, "agent card route was not registered")

	req := httptest.NewRequest(http.MethodGet, AgentCardRoute, nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var got AgentCard
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, "route-agent", got.Name)
	assert.Equal(t, "route description", got.Description)
}

func TestBuildSkillsExtractsRootOnlyAndSubAgentBranches(t *testing.T) {
	t.Parallel()

	rootOnly := newNewServerBuildsAgentCardFromConfigAndAgentTreeA2AAgent(t, "solo", "solo description", nil)
	assert.Equal(t, []AgentSkill{
		{
			ID:          "solo",
			Name:        "solo",
			Description: "solo description",
			Tags:        []string{SkillTagOrchestration},
		},
	}, buildSkills(rootOnly))

	child := newNewServerBuildsAgentCardFromConfigAndAgentTreeA2AAgent(t, "child", "child description", nil)
	root := newNewServerBuildsAgentCardFromConfigAndAgentTreeA2AAgent(t, "parent", "parent description", []adkagent.Agent{child})
	assert.Equal(t, []AgentSkill{
		{
			ID:          "parent",
			Name:        "parent",
			Description: "parent description",
			Tags:        []string{SkillTagOrchestration},
		},
		{
			ID:          "child",
			Name:        "child",
			Description: "child description",
			Tags:        []string{SkillTagSubAgentPrefix + "parent"},
		},
	}, buildSkills(root))
}

func TestAgentCardJSONOmitsEmptyP2PExtensions(t *testing.T) {
	t.Parallel()

	card := AgentCard{
		Name:        "minimal",
		Description: "minimal card",
		URL:         "https://minimal.example/a2a",
		Skills: []AgentSkill{
			{ID: "minimal", Name: "minimal", Description: "minimal card"},
		},
	}

	payload, err := json.Marshal(card)
	require.NoError(t, err)

	var got map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(payload, &got))
	for _, absent := range []string{"did", "multiaddrs", "capabilities", "pricing", "zkCredentials"} {
		assert.NotContains(t, got, absent)
	}
}

func TestZKCredentialRoundTripsOnAgentCard(t *testing.T) {
	t.Parallel()

	issuedAt := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	expiresAt := issuedAt.Add(time.Hour)
	card := AgentCard{
		Name:        "credentialed",
		Description: "credentialed card",
		URL:         "https://credentialed.example/a2a",
		ZKCredentials: []ZKCredential{
			{
				CapabilityID: "review",
				Proof:        []byte{1, 2, 3},
				IssuedAt:     issuedAt,
				ExpiresAt:    expiresAt,
			},
		},
	}

	payload, err := json.Marshal(card)
	require.NoError(t, err)

	var got AgentCard
	require.NoError(t, json.Unmarshal(payload, &got))
	require.Len(t, got.ZKCredentials, 1)
	assert.Equal(t, "review", got.ZKCredentials[0].CapabilityID)
	assert.Equal(t, []byte{1, 2, 3}, got.ZKCredentials[0].Proof)
	assert.True(t, got.ZKCredentials[0].IssuedAt.Equal(issuedAt))
	assert.True(t, got.ZKCredentials[0].ExpiresAt.Equal(expiresAt))
}

func newNewServerBuildsAgentCardFromConfigAndAgentTreeA2AAgent(
	t *testing.T,
	name string,
	description string,
	subAgents []adkagent.Agent,
) adkagent.Agent {
	t.Helper()

	created, err := adkagent.New(adkagent.Config{
		Name:        name,
		Description: description,
		SubAgents:   subAgents,
		Run: func(adkagent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(func(*session.Event, error) bool) {}
		},
	})
	require.NoError(t, err)
	return created
}

type newServerBuildsAgentCardFromConfigAndAgentTreeRouteRecorder struct {
	routes map[string]http.HandlerFunc
}

func (r *newServerBuildsAgentCardFromConfigAndAgentTreeRouteRecorder) Get(path string, handler http.HandlerFunc) {
	r.routes[path] = handler
}
