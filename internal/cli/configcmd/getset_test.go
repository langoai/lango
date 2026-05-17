package configcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"reflect"
	"strings"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
	"github.com/langoai/lango/internal/storage"
)

func executeConfigCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

type stubProfileStore struct {
	cfg *config.Config
}

func (s stubProfileStore) Save(context.Context, string, *config.Config, map[string]bool) error {
	return errors.New("not implemented")
}
func (s stubProfileStore) Load(context.Context, string) (*config.Config, map[string]bool, error) {
	return s.cfg, nil, nil
}
func (s stubProfileStore) LoadActive(context.Context) (string, *config.Config, map[string]bool, error) {
	return "default", s.cfg, nil, nil
}
func (s stubProfileStore) SetActive(context.Context, string) error {
	return errors.New("not implemented")
}
func (s stubProfileStore) List(context.Context) ([]configstore.ProfileInfo, error) {
	return nil, errors.New("not implemented")
}
func (s stubProfileStore) Delete(context.Context, string) error { return errors.New("not implemented") }
func (s stubProfileStore) Exists(context.Context, string) (bool, error) {
	return false, errors.New("not implemented")
}

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

func TestConfigGet_WritesToCommandOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewGetCmd(func() (*config.Config, error) { return cfg, nil })

	out, err := executeConfigCommand(t, cmd, "agent.provider")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "anthropic") {
		t.Fatalf("expected output to contain provider, got %q", out)
	}
}

func TestConfigGet_JSONOutputIsValid(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewGetCmd(func() (*config.Config, error) { return cfg, nil })

	out, err := executeConfigCommand(t, cmd, "agent", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if _, ok := decoded["provider"]; !ok {
		t.Fatalf("expected JSON output to include provider, got %v", decoded)
	}
}

func TestConfigGet_InvalidOutputRejectsBeforeLoad(t *testing.T) {
	called := false
	cmd := NewGetCmd(func() (*config.Config, error) {
		called = true
		return config.DefaultConfig(), nil
	})

	out, err := executeConfigCommand(t, cmd, "agent", "--output", "yaml")
	if err == nil {
		t.Fatal("expected invalid output format error")
	}
	if !strings.Contains(err.Error(), `unknown output format "yaml"`) {
		t.Fatalf("expected invalid output error, got %v", err)
	}
	if out != "" {
		t.Fatalf("expected empty command output, got %q", out)
	}
	if called {
		t.Fatal("cfgLoader must not run when output format validation fails")
	}
}

func TestConfigSet_WritesToCommandOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) { return cfg, nil, func() {}, nil },
		func(updated *config.Config, explicitKeys map[string]bool) error {
			cfg = updated
			if explicitKeys != nil {
				t.Fatalf("expected nil explicit keys for unrelated set without loaded explicit keys, got %v", explicitKeys)
			}
			return nil
		},
	)

	out, err := executeConfigCommand(t, cmd, "agent.provider", "openai")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Set agent.provider = openai") {
		t.Fatalf("expected confirmation output, got %q", out)
	}
}

func TestConfigSet_PreservesExistingExplicitKeys(t *testing.T) {
	cfg := config.DefaultConfig()
	loadedExplicit := map[string]bool{
		"knowledge.enabled": true,
	}
	var savedExplicit map[string]bool

	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			return cfg, loadedExplicit, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			cfg = updated
			savedExplicit = explicitKeys
			return nil
		},
	)

	_, err := executeConfigCommand(t, cmd, "agent.provider", "openai")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Agent.Provider != "openai" {
		t.Fatalf("expected agent.provider to be updated, got %q", cfg.Agent.Provider)
	}
	if !savedExplicit["knowledge.enabled"] {
		t.Fatalf("expected saved explicit keys to preserve knowledge.enabled, got %v", savedExplicit)
	}
}

func TestConfigSet_MarksContextRelatedPathExplicit(t *testing.T) {
	cfg := config.DefaultConfig()
	var savedExplicit map[string]bool

	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			return cfg, nil, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			cfg = updated
			savedExplicit = explicitKeys
			return nil
		},
	)

	_, err := executeConfigCommand(t, cmd, "knowledge.enabled", "false")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Knowledge.Enabled {
		t.Fatal("expected knowledge.enabled to be false")
	}
	if !savedExplicit["knowledge.enabled"] {
		t.Fatalf("expected knowledge.enabled to be marked explicit, got %v", savedExplicit)
	}
}

func TestConfigSet_InvalidPathDoesNotMutateExplicitKeysOrSave(t *testing.T) {
	cfg := config.DefaultConfig()
	loadedExplicit := map[string]bool{
		"knowledge.enabled": true,
	}
	saveCalled := false

	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			return cfg, loadedExplicit, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			saveCalled = true
			return nil
		},
	)

	_, err := executeConfigCommand(t, cmd, "invalid.key", "value")
	if err == nil {
		t.Fatal("expected invalid path error")
	}
	if saveCalled {
		t.Fatal("save must not be called after setConfigPath fails")
	}
	if _, ok := loadedExplicit["invalid.key"]; ok {
		t.Fatalf("invalid path must not mutate loaded explicit keys: %v", loadedExplicit)
	}
}

func TestConfigKeys_WritesToCommandOutput(t *testing.T) {
	cmd := NewKeysCmd()

	out, err := executeConfigCommand(t, cmd, "agent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "agent.provider") {
		t.Fatalf("expected output to contain agent.provider, got %q", out)
	}
}

func TestConfigExport_WritesToCommandOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	bootLoader := func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			ProfileName: "default",
			Storage:     storage.NewFacade(stubProfileStore{cfg: cfg}, nil),
		}, nil
	}
	cmd := newExportCmd(bootLoader)

	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"default"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), `"agent"`) {
		t.Fatalf("expected JSON config output, got %q", out.String())
	}
	if !strings.Contains(errBuf.String(), "WARNING: exported configuration contains sensitive values in plaintext.") {
		t.Fatalf("expected warning on stderr, got %q", errBuf.String())
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("expected valid JSON export, got error: %v", err)
	}
	if _, ok := decoded["agent"]; !ok {
		t.Fatalf("expected JSON export to include agent config, got %v", decoded)
	}
}

func TestConfigValidate_WritesToCommandOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	bootLoader := func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Config:      cfg,
			ProfileName: "default",
		}, nil
	}
	cmd := newValidateCmd(bootLoader)

	out, err := executeConfigCommand(t, cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, `Profile "default" configuration is valid.`) {
		t.Fatalf("expected validation confirmation, got %q", out)
	}
}
