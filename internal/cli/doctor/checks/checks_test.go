package checks

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/config"
)

func TestConfigCheck_Run_ValidConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	check := &ConfigCheck{}
	result := check.Run(context.Background(), cfg)

	// Accept both pass and warn (validation warnings are acceptable)
	if result.Status != StatusPass && result.Status != StatusWarn {
		t.Errorf("expected StatusPass or StatusWarn, got %v: %s", result.Status, result.Message)
	}
}

func TestConfigCheck_Run_NilConfig(t *testing.T) {
	check := &ConfigCheck{}
	result := check.Run(context.Background(), nil)

	if result.Status != StatusFail {
		t.Errorf("expected StatusFail, got %v", result.Status)
	}
}

func TestConfigCheck_Fix_GuidesToOnboard(t *testing.T) {
	check := &ConfigCheck{}
	result := check.Fix(context.Background(), nil)

	if result.Status != StatusFail {
		t.Errorf("expected StatusFail (guidance), got %v", result.Status)
	}
	if result.Details == "" {
		t.Error("expected details with onboard guidance")
	}
}

func TestProvidersCheck_Run_NoConfig(t *testing.T) {
	check := &ProvidersCheck{}
	result := check.Run(context.Background(), nil)

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip, got %v", result.Status)
	}
}

func TestProvidersCheck_Run_ExplicitProvider(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"openai": {Type: "openai", APIKey: "sk-test"},
		},
	}

	check := &ProvidersCheck{}
	result := check.Run(context.Background(), cfg)

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %v: %s", result.Status, result.Message)
	}
}

func TestProvidersCheck_Run_ProviderInMap(t *testing.T) {
	cfg := &config.Config{
		Agent: config.AgentConfig{
			Provider: "gemini",
		},
		Providers: map[string]config.ProviderConfig{
			"gemini": {Type: "gemini", APIKey: "test-key"},
		},
	}

	check := &ProvidersCheck{}
	result := check.Run(context.Background(), cfg)

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %v: %s", result.Status, result.Message)
	}
}

func TestProvidersCheck_Run_MissingKey(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"anthropic": {Type: "anthropic", APIKey: ""},
		},
	}

	check := &ProvidersCheck{}
	result := check.Run(context.Background(), cfg)

	if result.Status != StatusFail {
		t.Errorf("expected StatusFail, got %v", result.Status)
	}
}

func TestChannelCheck_Run_NoChannelsEnabled(t *testing.T) {
	cfg := &config.Config{}

	check := &ChannelCheck{}
	result := check.Run(context.Background(), cfg)

	if result.Status != StatusWarn {
		t.Errorf("expected StatusWarn for no channels, got %v", result.Status)
	}
}

func TestChannelCheck_Run_TelegramEnabledNoToken(t *testing.T) {
	cfg := &config.Config{}
	cfg.Channels.Telegram.Enabled = true

	check := &ChannelCheck{}
	result := check.Run(context.Background(), cfg)

	if result.Status != StatusFail {
		t.Errorf("expected StatusFail for missing token, got %v", result.Status)
	}
}

func TestNetworkCheck_Run_PortAvailable(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "127.0.0.1",
			Port: 19999, // Use a high port unlikely to be in use
		},
	}

	check := &NetworkCheck{}
	result := check.Run(context.Background(), cfg)

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %v: %s", result.Status, result.Message)
	}
}

func TestDatabaseCheck_Run_DirectoryNotExist(t *testing.T) {
	cfg := &config.Config{
		Session: config.SessionConfig{
			DatabasePath: "/nonexistent/path/lango.db",
		},
	}

	check := &DatabaseCheck{}
	result := check.Run(context.Background(), cfg)

	if result.Status != StatusFail {
		t.Errorf("expected StatusFail, got %v", result.Status)
	}
	if !result.Fixable {
		t.Error("expected Fixable to be true")
	}
}

func TestSecurityCheck_Run_EnclaveProvider(t *testing.T) {
	cfg := &config.Config{
		Session: config.SessionConfig{
			DatabasePath: "", // skip DB checks
		},
	}
	cfg.Security.Signer.Provider = "enclave"

	check := &SecurityCheck{}
	result := check.Run(context.Background(), cfg)

	if result.Status == StatusFail {
		t.Errorf("enclave provider should not return Fail, got: %s", result.Message)
	}
}

func TestSecurityCheck_Run_UnknownProvider(t *testing.T) {
	cfg := &config.Config{}
	cfg.Security.Signer.Provider = "some-unknown-provider"

	check := &SecurityCheck{}
	result := check.Run(context.Background(), cfg)

	if result.Status != StatusFail {
		t.Errorf("expected StatusFail for unknown provider, got %v: %s", result.Status, result.Message)
	}
}

func TestNewSummary(t *testing.T) {
	results := []Result{
		{Name: "Test1", Status: StatusPass},
		{Name: "Test2", Status: StatusPass},
		{Name: "Test3", Status: StatusWarn},
		{Name: "Test4", Status: StatusFail},
		{Name: "Test5", Status: StatusSkip},
	}

	summary := NewSummary(results)

	if summary.Passed != 2 {
		t.Errorf("expected 2 passed, got %d", summary.Passed)
	}
	if summary.Warnings != 1 {
		t.Errorf("expected 1 warning, got %d", summary.Warnings)
	}
	if summary.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", summary.Failed)
	}
	if summary.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", summary.Skipped)
	}
	if !summary.HasErrors() {
		t.Error("expected HasErrors to be true")
	}
	if !summary.HasWarnings() {
		t.Error("expected HasWarnings to be true")
	}
}
