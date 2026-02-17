package onboard

import (
	"testing"

	"github.com/langowarny/lango/internal/cli/tuicore"
	"github.com/langowarny/lango/internal/config"
)

func fieldByKey(form *tuicore.FormModel, key string) *tuicore.Field {
	for _, f := range form.Fields {
		if f.Key == key {
			return f
		}
	}
	return nil
}

func TestNewProviderStepForm(t *testing.T) {
	cfg := config.DefaultConfig()
	form := NewProviderStepForm(cfg)

	wantKeys := []string{"type", "id", "apikey", "baseurl"}

	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing field %q", key)
		}
	}

	if f := fieldByKey(form, "type"); f.Type != tuicore.InputSelect {
		t.Errorf("type: want InputSelect, got %d", f.Type)
	}
	if f := fieldByKey(form, "apikey"); f.Type != tuicore.InputPassword {
		t.Errorf("apikey: want InputPassword, got %d", f.Type)
	}
}

func TestNewAgentStepForm(t *testing.T) {
	cfg := config.DefaultConfig()
	form := NewAgentStepForm(cfg)

	wantKeys := []string{"provider", "model", "maxtokens", "temp"}

	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing field %q", key)
		}
	}

	if f := fieldByKey(form, "provider"); f.Type != tuicore.InputSelect {
		t.Errorf("provider: want InputSelect, got %d", f.Type)
	}
}

func TestNewChannelStepForm_Telegram(t *testing.T) {
	cfg := config.DefaultConfig()
	form := NewChannelStepForm("telegram", cfg)
	if form == nil {
		t.Fatal("form should not be nil for telegram")
	}
	if len(form.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(form.Fields))
	}
	if f := fieldByKey(form, "telegram_token"); f == nil {
		t.Error("missing telegram_token field")
	}
}

func TestNewChannelStepForm_Discord(t *testing.T) {
	cfg := config.DefaultConfig()
	form := NewChannelStepForm("discord", cfg)
	if form == nil {
		t.Fatal("form should not be nil for discord")
	}
	if len(form.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(form.Fields))
	}
	if f := fieldByKey(form, "discord_token"); f == nil {
		t.Error("missing discord_token field")
	}
}

func TestNewChannelStepForm_Slack(t *testing.T) {
	cfg := config.DefaultConfig()
	form := NewChannelStepForm("slack", cfg)
	if form == nil {
		t.Fatal("form should not be nil for slack")
	}
	if len(form.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(form.Fields))
	}
	if f := fieldByKey(form, "slack_token"); f == nil {
		t.Error("missing slack_token field")
	}
	if f := fieldByKey(form, "slack_app_token"); f == nil {
		t.Error("missing slack_app_token field")
	}
}

func TestNewChannelStepForm_Skip(t *testing.T) {
	cfg := config.DefaultConfig()
	form := NewChannelStepForm("skip", cfg)
	if form != nil {
		t.Error("form should be nil for skip")
	}
}

func TestNewSecurityStepForm(t *testing.T) {
	cfg := config.DefaultConfig()
	form := NewSecurityStepForm(cfg)

	wantKeys := []string{"interceptor_enabled", "interceptor_pii", "interceptor_policy"}

	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing field %q", key)
		}
	}
}

func TestSuggestModel(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{give: "anthropic", want: "claude-sonnet-4-5-20250929"},
		{give: "openai", want: "gpt-4o"},
		{give: "gemini", want: "gemini-2.0-flash"},
		{give: "ollama", want: "llama3.1"},
		{give: "unknown", want: "claude-sonnet-4-5-20250929"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := suggestModel(tt.give)
			if got != tt.want {
				t.Errorf("suggestModel(%q): want %q, got %q", tt.give, tt.want, got)
			}
		})
	}
}
