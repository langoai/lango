package supervisor

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/langowarny/lango/internal/config"
	"github.com/langowarny/lango/internal/provider"
)

func defaultTestConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = ""
	cfg.Providers = nil
	return cfg
}

func TestNew_NoProviders(t *testing.T) {
	cfg := defaultTestConfig()

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	if sv == nil {
		t.Fatal("expected supervisor to be non-nil")
	}
	if sv.Config != cfg {
		t.Error("expected supervisor.Config to be the provided config")
	}
	if sv.registry == nil {
		t.Error("expected registry to be initialized")
	}
	if sv.execTool == nil {
		t.Error("expected execTool to be initialized")
	}
}

func TestNew_OpenAIProvider(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"openai": {
			Type:   "openai",
			APIKey: "test-key",
		},
	}

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	p, ok := sv.registry.Get("openai")
	if !ok {
		t.Fatal("expected openai provider to be registered")
	}
	if p.ID() != "openai" {
		t.Errorf("expected provider ID 'openai', got %q", p.ID())
	}
}

func TestNew_AnthropicProvider(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {
			Type:   "anthropic",
			APIKey: "test-key",
		},
	}

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	p, ok := sv.registry.Get("anthropic")
	if !ok {
		t.Fatal("expected anthropic provider to be registered")
	}
	if p.ID() != "anthropic" {
		t.Errorf("expected provider ID 'anthropic', got %q", p.ID())
	}
}

func TestNew_OllamaProvider(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"ollama": {
			Type: "ollama",
		},
	}

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	p, ok := sv.registry.Get("ollama")
	if !ok {
		t.Fatal("expected ollama provider to be registered")
	}
	if p.ID() != "ollama" {
		t.Errorf("expected provider ID 'ollama', got %q", p.ID())
	}
}

func TestNew_GitHubProvider(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"github": {
			Type:   "github",
			APIKey: "test-key",
		},
	}

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	p, ok := sv.registry.Get("github")
	if !ok {
		t.Fatal("expected github provider to be registered")
	}
	if p.ID() != "github" {
		t.Errorf("expected provider ID 'github', got %q", p.ID())
	}
}

func TestNew_OllamaProvider_DefaultBaseURL(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"ollama": {
			Type: "ollama",
			// No BaseURL — should use default localhost:11434
		},
	}

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	if _, ok := sv.registry.Get("ollama"); !ok {
		t.Fatal("expected ollama provider to be registered with default base URL")
	}
}

func TestNew_GitHubProvider_DefaultBaseURL(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"github": {
			Type:   "github",
			APIKey: "test-key",
			// No BaseURL — should use default Azure endpoint
		},
	}

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	if _, ok := sv.registry.Get("github"); !ok {
		t.Fatal("expected github provider to be registered with default base URL")
	}
}

func TestNew_UnknownProviderType_Skipped(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"unknown": {
			Type:   "not-a-real-provider",
			APIKey: "test-key",
		},
	}

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if _, ok := sv.registry.Get("unknown"); ok {
		t.Error("unknown provider type should not be registered")
	}
}

func TestNew_DefaultProviderNotFound_Error(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Agent.Provider = "nonexistent"
	cfg.Providers = map[string]config.ProviderConfig{
		"openai": {
			Type:   "openai",
			APIKey: "test-key",
		},
	}

	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected error when default provider is not in providers list")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("expected error to mention 'nonexistent', got: %v", err)
	}
}

func TestNew_AutoSelectSingleProvider(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Agent.Provider = "" // No default set
	cfg.Providers = map[string]config.ProviderConfig{
		"openai": {
			Type:   "openai",
			APIKey: "test-key",
		},
	}

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if sv.Config.Agent.Provider != "openai" {
		t.Errorf("expected auto-selected provider 'openai', got %q", sv.Config.Agent.Provider)
	}
}

func TestNew_MultipleProviders(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Agent.Provider = "openai"
	cfg.Providers = map[string]config.ProviderConfig{
		"openai": {
			Type:   "openai",
			APIKey: "test-key-1",
		},
		"anthropic": {
			Type:   "anthropic",
			APIKey: "test-key-2",
		},
	}

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if _, ok := sv.registry.Get("openai"); !ok {
		t.Error("expected openai to be registered")
	}
	if _, ok := sv.registry.Get("anthropic"); !ok {
		t.Error("expected anthropic to be registered")
	}
}

func TestNew_ExecToolConfig(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Tools.Exec.DefaultTimeout = 10 * time.Second
	cfg.Tools.Exec.AllowBackground = true
	cfg.Tools.Exec.WorkDir = "/tmp"

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	if sv.execTool == nil {
		t.Fatal("expected execTool to be initialized")
	}
}

// --- Generate tests ---

func TestGenerate_ProviderNotFound(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"openai": {
			Type:   "openai",
			APIKey: "test-key",
		},
	}
	cfg.Agent.Provider = "openai"

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	_, err = sv.Generate(context.Background(), "nonexistent", "model", provider.GenerateParams{})
	if err == nil {
		t.Fatal("expected error for nonexistent provider")
	}
	if !strings.Contains(err.Error(), "provider not found") {
		t.Errorf("expected 'provider not found' error, got: %v", err)
	}
}

func TestGenerate_FallsBackToDefaultProvider(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"openai": {
			Type:   "openai",
			APIKey: "test-key",
		},
	}
	cfg.Agent.Provider = "openai"

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	// Pass empty providerID — should fall back to cfg.Agent.Provider ("openai")
	// This will try to actually call the provider, which will fail due to invalid key,
	// but the point is it doesn't fail with "provider not found"
	_, err = sv.Generate(context.Background(), "", "gpt-4", provider.GenerateParams{
		Messages: []provider.Message{{Role: "user", Content: "test"}},
	})
	// The error should be from the HTTP call (invalid key), not "provider not found"
	if err != nil && strings.Contains(err.Error(), "provider not found") {
		t.Errorf("expected fallback to default provider, got: %v", err)
	}
}

func TestGenerate_SetsModelFromParam(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"openai": {
			Type:   "openai",
			APIKey: "test-key",
		},
	}
	cfg.Agent.Provider = "openai"

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	// Verify that when params.Model is empty, the model argument is used
	params := provider.GenerateParams{
		Messages: []provider.Message{{Role: "user", Content: "test"}},
	}
	// We can't easily assert the model without mocking, but verify no "provider not found" error
	_, err = sv.Generate(context.Background(), "openai", "gpt-4", params)
	if err != nil && strings.Contains(err.Error(), "provider not found") {
		t.Errorf("unexpected provider not found error: %v", err)
	}
}

// --- ExecuteTool tests ---

func TestExecuteTool_SimpleCommand(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Tools.Exec.DefaultTimeout = 5 * time.Second

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	output, err := sv.ExecuteTool(context.Background(), "echo hello")
	if err != nil {
		t.Fatalf("ExecuteTool() returned error: %v", err)
	}
	if !strings.Contains(output, "hello") {
		t.Errorf("expected output to contain 'hello', got %q", output)
	}
}

func TestExecuteTool_NonZeroExitCode(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Tools.Exec.DefaultTimeout = 5 * time.Second

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	output, err := sv.ExecuteTool(context.Background(), "exit 1")
	if err != nil {
		t.Fatalf("ExecuteTool() returned error: %v (expected nil with exit code in output)", err)
	}
	if !strings.Contains(output, "exit code 1") {
		t.Errorf("expected output to contain 'exit code 1', got %q", output)
	}
}

func TestExecuteTool_StderrIncluded(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Tools.Exec.DefaultTimeout = 5 * time.Second

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	output, err := sv.ExecuteTool(context.Background(), "echo error >&2")
	if err != nil {
		t.Fatalf("ExecuteTool() returned error: %v", err)
	}
	if !strings.Contains(output, "error") {
		t.Errorf("expected output to contain stderr 'error', got %q", output)
	}
}

// --- Background process tests ---

func TestStartBackground_AndGetStatus(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Tools.Exec.AllowBackground = true
	cfg.Tools.Exec.DefaultTimeout = 5 * time.Second

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	id, err := sv.StartBackground("sleep 10")
	if err != nil {
		t.Fatalf("StartBackground() returned error: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty process ID")
	}

	status, err := sv.GetBackgroundStatus(id)
	if err != nil {
		t.Fatalf("GetBackgroundStatus() returned error: %v", err)
	}

	if status["id"] != id {
		t.Errorf("expected status id %q, got %q", id, status["id"])
	}
	if status["command"] != "sleep 10" {
		t.Errorf("expected command 'sleep 10', got %q", status["command"])
	}
	if done, ok := status["done"].(bool); !ok || done {
		t.Error("expected process to still be running")
	}

	// Cleanup
	if err := sv.StopBackground(id); err != nil {
		t.Errorf("StopBackground() returned error: %v", err)
	}
}

func TestStopBackground_Success(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Tools.Exec.AllowBackground = true

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	id, err := sv.StartBackground("sleep 30")
	if err != nil {
		t.Fatalf("StartBackground() returned error: %v", err)
	}

	if err := sv.StopBackground(id); err != nil {
		t.Fatalf("StopBackground() returned error: %v", err)
	}

	// After stop, getting status should fail (process removed)
	_, err = sv.GetBackgroundStatus(id)
	if err == nil {
		t.Error("expected error getting status of stopped process")
	}
}

func TestGetBackgroundStatus_NotFound(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Tools.Exec.AllowBackground = true

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	_, err = sv.GetBackgroundStatus("nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent process ID")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestStopBackground_NotFound(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Tools.Exec.AllowBackground = true

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	err = sv.StopBackground("nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent process ID")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

// --- ProviderProxy tests ---

func TestNewProviderProxy(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"openai": {
			Type:   "openai",
			APIKey: "test-key",
		},
	}
	cfg.Agent.Provider = "openai"

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	proxy := NewProviderProxy(sv, "openai", "gpt-4")
	if proxy.ID() != "openai" {
		t.Errorf("expected proxy ID 'openai', got %q", proxy.ID())
	}
}

func TestProviderProxy_ListModels(t *testing.T) {
	cfg := defaultTestConfig()
	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	proxy := NewProviderProxy(sv, "openai", "gpt-4")
	models, err := proxy.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels() returned error: %v", err)
	}
	// Returns empty list until proxy-based model listing is implemented.
	if len(models) != 0 {
		t.Errorf("expected empty models list, got %d", len(models))
	}
}

func TestProviderProxy_Generate_ProviderNotFound(t *testing.T) {
	cfg := defaultTestConfig()
	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	proxy := NewProviderProxy(sv, "nonexistent", "model")
	_, err = proxy.Generate(context.Background(), provider.GenerateParams{})
	if err == nil {
		t.Fatal("expected error for nonexistent provider")
	}
	if !strings.Contains(err.Error(), "provider not found") {
		t.Errorf("expected 'provider not found' error, got: %v", err)
	}
}

// --- Environment variable filtering via Supervisor ---

func TestExecuteTool_EnvWhitelist(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Tools.Exec.DefaultTimeout = 5 * time.Second

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	// The supervisor configures EnvWhitelist with PATH, HOME, USER, etc.
	// Verify that sensitive env vars are NOT visible to child processes.
	t.Setenv("OPENAI_API_KEY", "super-secret-key")

	output, err := sv.ExecuteTool(context.Background(), "env")
	if err != nil {
		t.Fatalf("ExecuteTool() returned error: %v", err)
	}

	if strings.Contains(output, "super-secret-key") {
		t.Error("OPENAI_API_KEY should be filtered out by env whitelist")
	}
	if strings.Contains(output, "OPENAI_API_KEY") {
		t.Error("OPENAI_API_KEY variable name should not appear in filtered env")
	}
}

func TestExecuteTool_WhitelistedVarsPresent(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Tools.Exec.DefaultTimeout = 5 * time.Second

	sv, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	// PATH should be whitelisted and visible
	output, err := sv.ExecuteTool(context.Background(), "env")
	if err != nil {
		t.Fatalf("ExecuteTool() returned error: %v", err)
	}

	if !strings.Contains(output, "PATH=") {
		t.Error("expected PATH to be present in whitelisted environment")
	}
}
