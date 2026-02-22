package a2a

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/config"
)

// fakeAgent implements agent.Agent for testing.
type fakeAgent struct {
	name        string
	description string
	subAgents   []fakeAgent
}

func (a fakeAgent) Name() string           { return a.name }
func (a fakeAgent) Description() string    { return a.description }
func (a fakeAgent) SubAgents() []fakeAgent { return a.subAgents }

func TestAgentCard(t *testing.T) {
	cfg := config.A2AConfig{
		Enabled:          true,
		BaseURL:          "http://localhost:8080",
		AgentName:        "test-agent",
		AgentDescription: "Test agent description",
	}

	card := &AgentCard{
		Name:        cfg.AgentName,
		Description: cfg.AgentDescription,
		URL:         cfg.BaseURL,
		Skills: []AgentSkill{
			{ID: "skill-1", Name: "skill-1", Description: "Test skill"},
		},
	}

	s := &Server{
		cfg:    cfg,
		card:   card,
		logger: zap.NewNop().Sugar(),
	}

	t.Run("agent card served", func(t *testing.T) {
		router := chi.NewRouter()
		s.RegisterRoutes(router)

		req := httptest.NewRequest(http.MethodGet, "/.well-known/agent.json", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", rec.Code)
		}

		ct := rec.Header().Get("Content-Type")
		if ct != "application/json" {
			t.Fatalf("want application/json, got %s", ct)
		}

		var got AgentCard
		if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
			t.Fatalf("decode agent card: %v", err)
		}

		if got.Name != "test-agent" {
			t.Errorf("want name 'test-agent', got %q", got.Name)
		}
		if got.Description != "Test agent description" {
			t.Errorf("want description 'Test agent description', got %q", got.Description)
		}
		if got.URL != "http://localhost:8080" {
			t.Errorf("want URL 'http://localhost:8080', got %q", got.URL)
		}
		if len(got.Skills) != 1 {
			t.Fatalf("want 1 skill, got %d", len(got.Skills))
		}
		if got.Skills[0].ID != "skill-1" {
			t.Errorf("want skill ID 'skill-1', got %q", got.Skills[0].ID)
		}
	})
}

func TestAgentCardEmpty(t *testing.T) {
	cfg := config.A2AConfig{
		Enabled: true,
		BaseURL: "http://localhost:9090",
	}

	card := &AgentCard{
		Name:        "empty-agent",
		Description: "",
		URL:         cfg.BaseURL,
	}

	s := &Server{
		cfg:    cfg,
		card:   card,
		logger: zap.NewNop().Sugar(),
	}

	router := chi.NewRouter()
	s.RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/agent.json", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}

	var got AgentCard
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Name != "empty-agent" {
		t.Errorf("want 'empty-agent', got %q", got.Name)
	}
	if len(got.Skills) != 0 {
		t.Errorf("want 0 skills, got %d", len(got.Skills))
	}
}
