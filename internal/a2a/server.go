package a2a

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/server/adka2a"

	"github.com/langoai/lango/internal/config"
)

// Server exposes a Lango agent as an A2A-compatible server.
type Server struct {
	cfg      config.A2AConfig
	agent    agent.Agent
	executor *adka2a.Executor
	card     *AgentCard
	logger   *zap.SugaredLogger
}

// AgentCard is a simplified representation of the A2A Agent Card
// served at /.well-known/agent.json.
type AgentCard struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	URL         string       `json:"url"`
	Skills      []AgentSkill `json:"skills"`

	// P2P extensions
	DID           string         `json:"did,omitempty"`
	Multiaddrs    []string       `json:"multiaddrs,omitempty"`
	Capabilities  []string       `json:"capabilities,omitempty"`
	Pricing       *PricingInfo   `json:"pricing,omitempty"`
	ZKCredentials []ZKCredential `json:"zkCredentials,omitempty"`
}

// PricingInfo describes the pricing for an agent's services.
type PricingInfo struct {
	Currency   string            `json:"currency"`
	PerQuery   string            `json:"perQuery,omitempty"`
	PerMinute  string            `json:"perMinute,omitempty"`
	ToolPrices map[string]string `json:"toolPrices,omitempty"`
}

// ZKCredential is a zero-knowledge proof of agent capability.
type ZKCredential struct {
	CapabilityID string    `json:"capabilityId"`
	Proof        []byte    `json:"proof"`
	IssuedAt     time.Time `json:"issuedAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

// AgentSkill describes a capability of the agent.
type AgentSkill struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags,omitempty"`
}

// NewServer creates a new A2A server for the given agent.
func NewServer(cfg config.A2AConfig, adkAgent agent.Agent, logger *zap.SugaredLogger) *Server {
	// Build agent card from agent tree.
	skills := buildSkills(adkAgent)

	name := cfg.AgentName
	if name == "" {
		name = adkAgent.Name()
	}
	desc := cfg.AgentDescription
	if desc == "" {
		desc = adkAgent.Description()
	}

	card := &AgentCard{
		Name:        name,
		Description: desc,
		URL:         cfg.BaseURL,
		Skills:      skills,
	}

	return &Server{
		cfg:    cfg,
		agent:  adkAgent,
		card:   card,
		logger: logger,
	}
}

// Card returns the agent card (used by P2P gossip and protocol).
func (s *Server) Card() *AgentCard { return s.card }

// SetP2PInfo adds P2P networking information to the agent card.
func (s *Server) SetP2PInfo(did string, multiaddrs, capabilities []string) {
	s.card.DID = did
	s.card.Multiaddrs = multiaddrs
	s.card.Capabilities = capabilities
}

// SetPricing sets the pricing information on the agent card.
func (s *Server) SetPricing(pricing *PricingInfo) {
	s.card.Pricing = pricing
}

// RegisterRoutes mounts the A2A routes on the given HTTP mux.
// - GET /.well-known/agent.json — serves the Agent Card
func (s *Server) RegisterRoutes(mux interface {
	Get(string, http.HandlerFunc)
}) {
	mux.Get("/.well-known/agent.json", s.handleAgentCard)
	s.logger.Infow("a2a routes registered",
		"agentCard", "/.well-known/agent.json",
	)
}

// handleAgentCard serves the Agent Card JSON.
func (s *Server) handleAgentCard(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.card); err != nil {
		s.logger.Warnw("encode agent card: %w", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

// buildSkills extracts agent skills from the agent tree.
func buildSkills(adkAgent agent.Agent) []AgentSkill {
	var skills []AgentSkill

	// Root agent skill.
	skills = append(skills, AgentSkill{
		ID:          adkAgent.Name(),
		Name:        adkAgent.Name(),
		Description: adkAgent.Description(),
		Tags:        []string{"orchestration"},
	})

	// Sub-agent skills.
	for _, sub := range adkAgent.SubAgents() {
		skills = append(skills, AgentSkill{
			ID:          sub.Name(),
			Name:        sub.Name(),
			Description: sub.Description(),
			Tags:        []string{"sub_agent:" + adkAgent.Name()},
		})
	}

	return skills
}
