package configcmd

import (
	"reflect"
	"testing"

	"github.com/langoai/lango/internal/config"
)

func TestResolveConfigPath_AgentProvider(t *testing.T) {
	cfg := config.DefaultConfig()
	val, err := resolveConfigPath(cfg, "agent.provider")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "anthropic" {
		t.Errorf("want %q, got %q", "anthropic", val)
	}
}

func TestResolveConfigPath_NestedField(t *testing.T) {
	cfg := config.DefaultConfig()
	val, err := resolveConfigPath(cfg, "logging.level")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "info" {
		t.Errorf("want %q, got %q", "info", val)
	}
}

func TestResolveConfigPath_BoolField(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	val, err := resolveConfigPath(cfg, "p2p.enabled")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != true {
		t.Errorf("want true, got %v", val)
	}
}

func TestResolveConfigPath_NotFound(t *testing.T) {
	cfg := config.DefaultConfig()
	_, err := resolveConfigPath(cfg, "agent.nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestResolveConfigPath_DeepNested(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Budget.DefaultMax = "50.00"
	val, err := resolveConfigPath(cfg, "economy.budget.defaultMax")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "50.00" {
		t.Errorf("want %q, got %q", "50.00", val)
	}
}

func TestSetConfigPath_StringField(t *testing.T) {
	cfg := config.DefaultConfig()
	err := setConfigPath(cfg, "agent.provider", "openai")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Agent.Provider != "openai" {
		t.Errorf("want %q, got %q", "openai", cfg.Agent.Provider)
	}
}

func TestSetConfigPath_BoolField(t *testing.T) {
	cfg := config.DefaultConfig()
	err := setConfigPath(cfg, "p2p.enabled", "true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.P2P.Enabled {
		t.Error("want true, got false")
	}
}

func TestSetConfigPath_IntField(t *testing.T) {
	cfg := config.DefaultConfig()
	err := setConfigPath(cfg, "server.port", "9999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 9999 {
		t.Errorf("want 9999, got %d", cfg.Server.Port)
	}
}

func TestSetConfigPath_FloatField(t *testing.T) {
	cfg := config.DefaultConfig()
	err := setConfigPath(cfg, "agent.temperature", "0.5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Agent.Temperature != 0.5 {
		t.Errorf("want 0.5, got %f", cfg.Agent.Temperature)
	}
}

func TestSetConfigPath_NotFound(t *testing.T) {
	cfg := config.DefaultConfig()
	err := setConfigPath(cfg, "agent.nonexistent", "val")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestCollectKeys_ContainsExpectedKeys(t *testing.T) {
	keys := collectKeys(reflect.TypeOf(config.Config{}), "")

	wantKeys := []string{
		"agent.provider",
		"logging.level",
		"server.port",
		"p2p.enabled",
	}

	keySet := make(map[string]bool, len(keys))
	for _, k := range keys {
		keySet[k] = true
	}

	for _, wk := range wantKeys {
		if !keySet[wk] {
			t.Errorf("expected key %q in output", wk)
		}
	}
}

func TestFormatPlain(t *testing.T) {
	tests := []struct {
		give interface{}
		want string
	}{
		{give: "hello", want: "hello"},
		{give: true, want: "true"},
		{give: 42, want: "42"},
		{give: nil, want: "<nil>"},
		{give: []string{"a", "b"}, want: "a,b"},
	}

	for _, tt := range tests {
		got := formatPlain(tt.give)
		if got != tt.want {
			t.Errorf("formatPlain(%v): want %q, got %q", tt.give, tt.want, got)
		}
	}
}
