package configcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"

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

func TestConfigGet_InvalidPathSuggestsNearbyKeyAndDiscoveryCommand(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewGetCmd(func() (*config.Config, error) { return cfg, nil })

	_, err := executeConfigCommand(t, cmd, "agent.providr")
	if err == nil {
		t.Fatal("expected invalid path error")
	}

	errText := err.Error()
	if !strings.Contains(errText, "agent.provider") {
		t.Fatalf("expected error to suggest agent.provider, got %q", errText)
	}
	if !strings.Contains(errText, "lango config keys agent") {
		t.Fatalf("expected error to include agent key discovery command, got %q", errText)
	}
}

func TestConfigGet_UnknownTopLevelPathIncludesDiscoveryCommand(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewGetCmd(func() (*config.Config, error) { return cfg, nil })

	_, err := executeConfigCommand(t, cmd, "made.up.path")
	if err == nil {
		t.Fatal("expected invalid path error")
	}
	if !strings.Contains(err.Error(), "lango config keys") {
		t.Fatalf("expected error to include generic key discovery command, got %q", err.Error())
	}
}

func TestConfigGet_LeafExtensionPathIncludesDiscoveryCommand(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewGetCmd(func() (*config.Config, error) { return cfg, nil })

	_, err := executeConfigCommand(t, cmd, "agent.provider.extra")
	if err == nil {
		t.Fatal("expected invalid path error")
	}

	errText := err.Error()
	if !strings.Contains(errText, "agent.provider") {
		t.Fatalf("expected error to reference agent.provider, got %q", errText)
	}
	if !strings.Contains(errText, "lango config keys agent") {
		t.Fatalf("expected error to include agent key discovery command, got %q", errText)
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

func TestSetConfigPath_CreatesProviderMapEntry(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers = nil

	err := setConfigPath(cfg, "providers.openai.type", "openai")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Providers == nil {
		t.Fatal("expected providers map to be created")
	}
	if got := cfg.Providers["openai"].Type; got != "openai" {
		t.Fatalf("expected providers.openai.type to be set, got %q", got)
	}
}

func TestSetConfigPath_UpdatesProviderMapEntryWithoutLosingFields(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"openai": {
			Type:   "openai",
			APIKey: "existing-key",
		},
	}

	err := setConfigPath(cfg, "providers.openai.baseUrl", "http://localhost:11434/v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := cfg.Providers["openai"]
	if got.BaseURL != "http://localhost:11434/v1" {
		t.Fatalf("expected baseUrl to be updated, got %q", got.BaseURL)
	}
	if got.Type != "openai" {
		t.Fatalf("expected provider type to be preserved, got %q", got.Type)
	}
	if got.APIKey != "existing-key" {
		t.Fatalf("expected apiKey to be preserved, got %q", got.APIKey)
	}
}

func TestSetConfigPath_CreatesNestedStringMapEntry(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.MCP.Servers = nil

	err := setConfigPath(cfg, "mcp.servers.docs.env.LOG_LEVEL", "debug")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MCP.Servers == nil {
		t.Fatal("expected mcp.servers map to be created")
	}
	if cfg.MCP.Servers["docs"].Env == nil {
		t.Fatal("expected mcp.servers.docs.env map to be created")
	}
	if got := cfg.MCP.Servers["docs"].Env["LOG_LEVEL"]; got != "debug" {
		t.Fatalf("expected mcp.servers.docs.env.LOG_LEVEL to be set, got %q", got)
	}
}

func TestSetConfigPath_CreatesAuthProviderMapEntry(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Auth.Providers = nil

	err := setConfigPath(cfg, "auth.providers.google.clientId", "${GOOGLE_CLIENT_ID}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Auth.Providers == nil {
		t.Fatal("expected auth.providers map to be created")
	}
	if got := cfg.Auth.Providers["google"].ClientID; got != "${GOOGLE_CLIENT_ID}" {
		t.Fatalf("expected auth.providers.google.clientId to be set, got %q", got)
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

func TestConfigSet_FromEnvRedactsProviderAPIKeyOutputAndSavesRawValue(t *testing.T) {
	cfg := config.DefaultConfig()
	rawSecret := "sk-env-secret"
	t.Setenv("LANGO_TEST_OPENAI_API_KEY", rawSecret)
	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			return cfg, nil, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			cfg = updated
			return nil
		},
	)

	out, err := executeConfigCommand(
		t,
		cmd,
		"providers.openai.apiKey",
		"--from-env",
		"LANGO_TEST_OPENAI_API_KEY",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Set providers.openai.apiKey = <redacted>") {
		t.Fatalf("expected redacted confirmation output, got %q", out)
	}
	if strings.Contains(out, rawSecret) {
		t.Fatalf("confirmation output must not contain raw secret, got %q", out)
	}
	if got := cfg.Providers["openai"].APIKey; got != rawSecret {
		t.Fatalf("expected raw provider API key to be saved, got %q", got)
	}
}

func TestConfigSet_FromEnvShowsNonSensitiveOutputAndSavesRawValue(t *testing.T) {
	cfg := config.DefaultConfig()
	t.Setenv("LANGO_TEST_AGENT_PROVIDER", "openai")
	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			return cfg, nil, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			cfg = updated
			return nil
		},
	)

	out, err := executeConfigCommand(
		t,
		cmd,
		"agent.provider",
		"--from-env",
		"LANGO_TEST_AGENT_PROVIDER",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Set agent.provider = openai") {
		t.Fatalf("expected non-sensitive confirmation output, got %q", out)
	}
	if cfg.Agent.Provider != "openai" {
		t.Fatalf("expected agent.provider to be saved, got %q", cfg.Agent.Provider)
	}
}

func TestConfigSet_FromEnvAllowsEmptyEnvValue(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers = make(map[string]config.ProviderConfig)
	cfg.Providers["openai"] = config.ProviderConfig{
		Type:   "openai",
		APIKey: "old-secret",
	}
	t.Setenv("LANGO_TEST_EMPTY_SECRET", "")
	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			return cfg, nil, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			cfg = updated
			return nil
		},
	)

	out, err := executeConfigCommand(
		t,
		cmd,
		"providers.openai.apiKey",
		"--from-env",
		"LANGO_TEST_EMPTY_SECRET",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Set providers.openai.apiKey = <redacted>") {
		t.Fatalf("expected redacted confirmation output, got %q", out)
	}
	if got := cfg.Providers["openai"].APIKey; got != "" {
		t.Fatalf("expected empty provider API key to be saved, got %q", got)
	}
}

func TestConfigSet_FromEnvMissingVariableRejectsBeforeLoadOrSave(t *testing.T) {
	envName := "LANGO_TEST_MISSING_OPENAI_API_KEY"
	if err := os.Unsetenv(envName); err != nil {
		t.Fatalf("unset env: %v", err)
	}
	loadCalled := false
	saveCalled := false
	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			loadCalled = true
			return config.DefaultConfig(), nil, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			saveCalled = true
			return nil
		},
	)

	out, err := executeConfigCommand(t, cmd, "providers.openai.apiKey", "--from-env", envName)
	if err == nil {
		t.Fatal("expected missing env error")
	}
	if !strings.Contains(err.Error(), `environment variable "LANGO_TEST_MISSING_OPENAI_API_KEY" is not set`) {
		t.Fatalf("expected missing env error, got %v", err)
	}
	if out != "" {
		t.Fatalf("expected empty command output, got %q", out)
	}
	if loadCalled {
		t.Fatal("cfgLoader must not run when --from-env variable is missing")
	}
	if saveCalled {
		t.Fatal("cfgSaver must not run when --from-env variable is missing")
	}
}

func TestConfigSet_FromEnvRejectsPositionalValueBeforeLoadOrSave(t *testing.T) {
	loadCalled := false
	saveCalled := false
	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			loadCalled = true
			return config.DefaultConfig(), nil, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			saveCalled = true
			return nil
		},
	)

	out, err := executeConfigCommand(
		t,
		cmd,
		"providers.openai.apiKey",
		"raw-secret",
		"--from-env",
		"LANGO_TEST_OPENAI_API_KEY",
	)
	if err == nil {
		t.Fatal("expected --from-env and positional value conflict")
	}
	if !strings.Contains(err.Error(), "--from-env cannot be combined with a value argument") {
		t.Fatalf("expected --from-env conflict error, got %v", err)
	}
	if out != "" {
		t.Fatalf("expected empty command output, got %q", out)
	}
	if loadCalled {
		t.Fatal("cfgLoader must not run when --from-env is combined with a value argument")
	}
	if saveCalled {
		t.Fatal("cfgSaver must not run when --from-env is combined with a value argument")
	}
}

func TestConfigSet_RedactsProviderAPIKeyOutputAndSavesRawValue(t *testing.T) {
	cfg := config.DefaultConfig()
	rawSecret := "sk-raw-secret"
	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			return cfg, nil, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			cfg = updated
			return nil
		},
	)

	out, err := executeConfigCommand(t, cmd, "providers.openai.apiKey", rawSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Set providers.openai.apiKey = <redacted>") {
		t.Fatalf("expected redacted confirmation output, got %q", out)
	}
	if strings.Contains(out, rawSecret) {
		t.Fatalf("confirmation output must not contain raw secret, got %q", out)
	}
	if got := cfg.Providers["openai"].APIKey; got != rawSecret {
		t.Fatalf("expected raw provider API key to be saved, got %q", got)
	}
}

func TestConfigSet_RedactsMCPEnvSecretOutputAndSavesRawValue(t *testing.T) {
	cfg := config.DefaultConfig()
	rawSecret := "docs-env-secret"
	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			return cfg, nil, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			cfg = updated
			return nil
		},
	)

	out, err := executeConfigCommand(t, cmd, "mcp.servers.docs.env.API_KEY", rawSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Set mcp.servers.docs.env.API_KEY = <redacted>") {
		t.Fatalf("expected redacted confirmation output, got %q", out)
	}
	if strings.Contains(out, rawSecret) {
		t.Fatalf("confirmation output must not contain raw secret, got %q", out)
	}
	if got := cfg.MCP.Servers["docs"].Env["API_KEY"]; got != rawSecret {
		t.Fatalf("expected raw MCP env value to be saved, got %q", got)
	}
}

func TestConfigSet_RedactsMCPEnvAPIKeyVariantOutputAndSavesRawValue(t *testing.T) {
	cfg := config.DefaultConfig()
	rawSecret := "docs-openai-secret"
	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			return cfg, nil, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			cfg = updated
			return nil
		},
	)

	out, err := executeConfigCommand(t, cmd, "mcp.servers.docs.env.OPENAI_API_KEY", rawSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Set mcp.servers.docs.env.OPENAI_API_KEY = <redacted>") {
		t.Fatalf("expected redacted confirmation output, got %q", out)
	}
	if strings.Contains(out, rawSecret) {
		t.Fatalf("confirmation output must not contain raw secret, got %q", out)
	}
	if got := cfg.MCP.Servers["docs"].Env["OPENAI_API_KEY"]; got != rawSecret {
		t.Fatalf("expected raw MCP env value to be saved, got %q", got)
	}
}

func TestConfigSet_RedactsMCPHeaderAuthorizationVariantOutputAndSavesRawValue(t *testing.T) {
	cfg := config.DefaultConfig()
	rawSecret := "Bearer docs-secret"
	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			return cfg, nil, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			cfg = updated
			return nil
		},
	)

	out, err := executeConfigCommand(t, cmd, "mcp.servers.docs.headers.Proxy-Authorization", rawSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Set mcp.servers.docs.headers.Proxy-Authorization = <redacted>") {
		t.Fatalf("expected redacted confirmation output, got %q", out)
	}
	if strings.Contains(out, rawSecret) {
		t.Fatalf("confirmation output must not contain raw secret, got %q", out)
	}
	if got := cfg.MCP.Servers["docs"].Headers["Proxy-Authorization"]; got != rawSecret {
		t.Fatalf("expected raw MCP header value to be saved, got %q", got)
	}
}

func TestConfigSet_RedactsPKCS11PinOutputAndSavesRawValue(t *testing.T) {
	cfg := config.DefaultConfig()
	rawSecret := "123456"
	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			return cfg, nil, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			cfg = updated
			return nil
		},
	)

	out, err := executeConfigCommand(t, cmd, "security.kms.pkcs11.pin", rawSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Set security.kms.pkcs11.pin = <redacted>") {
		t.Fatalf("expected redacted confirmation output, got %q", out)
	}
	if strings.Contains(out, rawSecret) {
		t.Fatalf("confirmation output must not contain raw secret, got %q", out)
	}
	if got := cfg.Security.KMS.PKCS11.Pin; got != rawSecret {
		t.Fatalf("expected raw PKCS#11 pin to be saved, got %q", got)
	}
}

func TestConfigSet_DoesNotRedactNonSecretKeyDirOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	keyDir := "/tmp/lango-node-keys"
	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			return cfg, nil, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			cfg = updated
			return nil
		},
	)

	out, err := executeConfigCommand(t, cmd, "p2p.keyDir", keyDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Set p2p.keyDir = "+keyDir) {
		t.Fatalf("expected keyDir confirmation to remain visible, got %q", out)
	}
	if strings.Contains(out, "<redacted>") {
		t.Fatalf("keyDir output must not be redacted, got %q", out)
	}
	if cfg.P2P.KeyDir != keyDir {
		t.Fatalf("expected raw keyDir to be saved, got %q", cfg.P2P.KeyDir)
	}
}

func TestConfigSet_DoesNotRedactNonSecretCredentialAgeOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	credentialAge := "48h"
	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			return cfg, nil, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			cfg = updated
			return nil
		},
	)

	out, err := executeConfigCommand(t, cmd, "p2p.zkp.maxCredentialAge", credentialAge)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Set p2p.zkp.maxCredentialAge = "+credentialAge) {
		t.Fatalf("expected maxCredentialAge confirmation to remain visible, got %q", out)
	}
	if strings.Contains(out, "<redacted>") {
		t.Fatalf("maxCredentialAge output must not be redacted, got %q", out)
	}
	if cfg.P2P.ZKP.MaxCredentialAge != credentialAge {
		t.Fatalf("expected maxCredentialAge to be saved, got %q", cfg.P2P.ZKP.MaxCredentialAge)
	}
}

func TestConfigSet_DoesNotRedactNonSecretTokenCountOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			return cfg, nil, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			cfg = updated
			return nil
		},
	)

	out, err := executeConfigCommand(t, cmd, "agent.maxTokens", "8192")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Set agent.maxTokens = 8192") {
		t.Fatalf("expected maxTokens confirmation to remain visible, got %q", out)
	}
	if strings.Contains(out, "<redacted>") {
		t.Fatalf("maxTokens output must not be redacted, got %q", out)
	}
	if cfg.Agent.MaxTokens != 8192 {
		t.Fatalf("expected maxTokens to be saved, got %d", cfg.Agent.MaxTokens)
	}
}

func TestConfigSet_DoesNotRedactNonSecretKeyIDOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	keyID := "signer-key-id"
	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			return cfg, nil, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			cfg = updated
			return nil
		},
	)

	out, err := executeConfigCommand(t, cmd, "security.signer.keyId", keyID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Set security.signer.keyId = "+keyID) {
		t.Fatalf("expected keyId confirmation to remain visible, got %q", out)
	}
	if strings.Contains(out, "<redacted>") {
		t.Fatalf("keyId output must not be redacted, got %q", out)
	}
	if cfg.Security.Signer.KeyID != keyID {
		t.Fatalf("expected keyId to be saved, got %q", cfg.Security.Signer.KeyID)
	}
}

func TestConfigSetPathIsSensitive(t *testing.T) {
	tests := []struct {
		givePath string
		want     bool
	}{
		{givePath: "providers.openai.apiKey", want: true},
		{givePath: "channels.slack.botToken", want: true},
		{givePath: "channels.slack.signingSecret", want: true},
		{givePath: "mcp.servers.docs.headers.Authorization", want: true},
		{givePath: "mcp.servers.docs.headers.Proxy-Authorization", want: true},
		{givePath: "mcp.servers.docs.headers.X-API-Key", want: true},
		{givePath: "mcp.servers.docs.env.API_KEY", want: true},
		{givePath: "mcp.servers.docs.env.OPENAI_API_KEY", want: true},
		{givePath: "security.kms.pkcs11.pin", want: true},
		{givePath: "auth.providers.okta.clientCredential", want: true},
		{givePath: "agent.maxTokens", want: false},
		{givePath: "p2p.keyDir", want: false},
		{givePath: "p2p.zkp.maxCredentialAge", want: false},
		{givePath: "security.signer.keyId", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.givePath, func(t *testing.T) {
			got := configSetPathIsSensitive(tt.givePath)
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
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

func TestConfigSet_InvalidPathSuggestsNearbyKeyAndDoesNotSave(t *testing.T) {
	cfg := config.DefaultConfig()
	saveCalled := false

	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			return cfg, nil, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			saveCalled = true
			return nil
		},
	)

	_, err := executeConfigCommand(t, cmd, "knowledge.enable", "false")
	if err == nil {
		t.Fatal("expected invalid path error")
	}

	errText := err.Error()
	if !strings.Contains(errText, "knowledge.enabled") {
		t.Fatalf("expected error to suggest knowledge.enabled, got %q", errText)
	}
	if !strings.Contains(errText, "lango config keys knowledge") {
		t.Fatalf("expected error to include knowledge key discovery command, got %q", errText)
	}
	if saveCalled {
		t.Fatal("save must not be called after invalid path")
	}
}

func TestConfigSet_InvalidMapBackedPathDoesNotSave(t *testing.T) {
	cfg := config.DefaultConfig()
	saveCalled := false

	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			return cfg, nil, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			saveCalled = true
			return nil
		},
	)

	_, err := executeConfigCommand(t, cmd, "providers.openai.notAField", "value")
	if err == nil {
		t.Fatal("expected invalid path error")
	}
	if saveCalled {
		t.Fatal("save must not be called after invalid map-backed path")
	}
}

func TestConfigSet_LeafExtensionPathIncludesDiscoveryCommandAndDoesNotSave(t *testing.T) {
	cfg := config.DefaultConfig()
	saveCalled := false

	cmd := NewSetCmd(
		func() (*config.Config, map[string]bool, func(), error) {
			return cfg, nil, func() {}, nil
		},
		func(updated *config.Config, explicitKeys map[string]bool) error {
			saveCalled = true
			return nil
		},
	)

	_, err := executeConfigCommand(t, cmd, "agent.provider.extra", "openai")
	if err == nil {
		t.Fatal("expected invalid path error")
	}

	errText := err.Error()
	if !strings.Contains(errText, "agent.provider") {
		t.Fatalf("expected error to reference agent.provider, got %q", errText)
	}
	if !strings.Contains(errText, "lango config keys agent") {
		t.Fatalf("expected error to include agent key discovery command, got %q", errText)
	}
	if saveCalled {
		t.Fatal("save must not be called after invalid path")
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
