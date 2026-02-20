package settings

import (
	"strconv"
	"testing"
	"time"

	"github.com/langowarny/lango/internal/cli/tuicore"
	"github.com/langowarny/lango/internal/config"
)

func defaultTestConfig() *config.Config {
	return config.DefaultConfig()
}

func fieldByKey(form *tuicore.FormModel, key string) *tuicore.Field {
	for _, f := range form.Fields {
		if f.Key == key {
			return f
		}
	}
	return nil
}

func TestNewAgentForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewAgentForm(cfg)

	wantKeys := []string{
		"provider", "model", "maxtokens", "temp",
		"prompts_dir", "fallback_provider", "fallback_model",
		"request_timeout", "tool_timeout",
	}

	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing field %q", key)
		}
	}

	if f := fieldByKey(form, "provider"); f.Value != "anthropic" {
		t.Errorf("provider: want %q, got %q", "anthropic", f.Value)
	}
	if f := fieldByKey(form, "fallback_provider"); f.Type != tuicore.InputSelect {
		t.Errorf("fallback_provider: want InputSelect, got %d", f.Type)
	}
}

func TestNewToolsForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewToolsForm(cfg)

	wantKeys := []string{
		"exec_timeout", "exec_bg",
		"browser_enabled", "browser_headless", "browser_session_timeout",
		"fs_max_read",
	}

	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing field %q", key)
		}
	}

	if f := fieldByKey(form, "browser_enabled"); f.Checked != false {
		t.Error("browser_enabled: want false by default")
	}
	if f := fieldByKey(form, "browser_headless"); f.Checked != true {
		t.Error("browser_headless: want true by default")
	}
}

func TestNewSessionForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewSessionForm(cfg)

	wantKeys := []string{
		"ttl", "max_history_turns",
	}

	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing field %q", key)
		}
	}

	if f := fieldByKey(form, "max_history_turns"); f.Value != "50" {
		t.Errorf("max_history_turns: want %q, got %q", "50", f.Value)
	}
}

func TestNewSecurityForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewSecurityForm(cfg)

	wantKeys := []string{
		"interceptor_enabled", "interceptor_pii", "interceptor_policy",
		"interceptor_timeout", "interceptor_notify", "interceptor_sensitive_tools",
		"interceptor_exempt_tools",
		"signer_provider", "signer_rpc", "signer_keyid",
	}

	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing field %q", key)
		}
	}
}

func TestNewKnowledgeForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewKnowledgeForm(cfg)

	wantKeys := []string{
		"knowledge_enabled", "knowledge_max_context",
	}

	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing field %q", key)
		}
	}

	if f := fieldByKey(form, "knowledge_enabled"); f.Checked != false {
		t.Error("knowledge_enabled: want false by default")
	}
	if f := fieldByKey(form, "knowledge_max_context"); f.Value != "5" {
		t.Errorf("knowledge_max_context: want %q, got %q", "5", f.Value)
	}
}

func TestUpdateConfigFromForm_AgentAdvancedFields(t *testing.T) {
	state := tuicore.NewConfigState()
	form := tuicore.NewFormModel("test")
	form.AddField(&tuicore.Field{Key: "prompts_dir", Type: tuicore.InputText, Value: "~/.lango/prompts"})
	form.AddField(&tuicore.Field{Key: "fallback_provider", Type: tuicore.InputSelect, Value: "openai"})
	form.AddField(&tuicore.Field{Key: "fallback_model", Type: tuicore.InputText, Value: "gpt-4o"})

	state.UpdateConfigFromForm(&form)

	if state.Current.Agent.PromptsDir != "~/.lango/prompts" {
		t.Errorf("PromptsDir: want %q, got %q", "~/.lango/prompts", state.Current.Agent.PromptsDir)
	}
	if state.Current.Agent.FallbackProvider != "openai" {
		t.Errorf("FallbackProvider: want %q, got %q", "openai", state.Current.Agent.FallbackProvider)
	}
	if state.Current.Agent.FallbackModel != "gpt-4o" {
		t.Errorf("FallbackModel: want %q, got %q", "gpt-4o", state.Current.Agent.FallbackModel)
	}
}

func TestUpdateConfigFromForm_BrowserFields(t *testing.T) {
	state := tuicore.NewConfigState()
	form := tuicore.NewFormModel("test")
	form.AddField(&tuicore.Field{Key: "browser_enabled", Type: tuicore.InputBool, Checked: true})
	form.AddField(&tuicore.Field{Key: "browser_session_timeout", Type: tuicore.InputText, Value: "10m"})

	state.UpdateConfigFromForm(&form)

	if !state.Current.Tools.Browser.Enabled {
		t.Error("Browser.Enabled: want true")
	}
	if state.Current.Tools.Browser.SessionTimeout != 10*time.Minute {
		t.Errorf("Browser.SessionTimeout: want 10m, got %v", state.Current.Tools.Browser.SessionTimeout)
	}
}

func TestUpdateConfigFromForm_MaxHistoryTurns(t *testing.T) {
	state := tuicore.NewConfigState()
	form := tuicore.NewFormModel("test")
	form.AddField(&tuicore.Field{Key: "max_history_turns", Type: tuicore.InputInt, Value: "100"})

	state.UpdateConfigFromForm(&form)

	if state.Current.Session.MaxHistoryTurns != 100 {
		t.Errorf("MaxHistoryTurns: want 100, got %d", state.Current.Session.MaxHistoryTurns)
	}
}

func TestUpdateConfigFromForm_KnowledgeFields(t *testing.T) {
	state := tuicore.NewConfigState()
	form := tuicore.NewFormModel("test")
	form.AddField(&tuicore.Field{Key: "knowledge_enabled", Type: tuicore.InputBool, Checked: true})
	form.AddField(&tuicore.Field{Key: "knowledge_max_context", Type: tuicore.InputInt, Value: "8"})
	state.UpdateConfigFromForm(&form)

	k := state.Current.Knowledge
	if !k.Enabled {
		t.Error("Knowledge.Enabled: want true")
	}
	if k.MaxContextPerLayer != 8 {
		t.Errorf("MaxContextPerLayer: want 8, got %d", k.MaxContextPerLayer)
	}
}

func TestNewObservationalMemoryForm_ProviderIsSelect(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewObservationalMemoryForm(cfg)

	wantKeys := []string{
		"om_enabled", "om_provider", "om_model",
		"om_msg_threshold", "om_obs_threshold", "om_max_budget",
		"om_max_reflections", "om_max_observations",
	}

	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing field %q", key)
		}
	}

	f := fieldByKey(form, "om_provider")
	if f.Type != tuicore.InputSelect {
		t.Errorf("om_provider: want InputSelect, got %d", f.Type)
	}
	if len(f.Options) == 0 {
		t.Fatal("om_provider: options must not be empty")
	}
	if f.Options[0] != "" {
		t.Errorf("om_provider: first option should be empty string, got %q", f.Options[0])
	}

	if mf := fieldByKey(form, "om_model"); mf.Type != tuicore.InputText {
		t.Errorf("om_model: want InputText, got %d", mf.Type)
	}
}

func TestNewEmbeddingForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewEmbeddingForm(cfg)

	wantKeys := []string{
		"emb_provider_id", "emb_model", "emb_dimensions",
		"emb_local_baseurl",
		"emb_rag_enabled", "emb_rag_max_results", "emb_rag_collections",
	}

	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing field %q", key)
		}
	}

	if f := fieldByKey(form, "emb_provider_id"); f.Type != tuicore.InputSelect {
		t.Errorf("emb_provider_id: want InputSelect, got %d", f.Type)
	}
	if f := fieldByKey(form, "emb_rag_enabled"); f.Type != tuicore.InputBool {
		t.Errorf("emb_rag_enabled: want InputBool, got %d", f.Type)
	}
}

func TestNewEmbeddingForm_ProviderOptionsFromProviders(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"gemini-1":  {Type: "gemini", APIKey: "test-key"},
		"my-openai": {Type: "openai", APIKey: "sk-test"},
	}
	cfg.Embedding.ProviderID = "gemini-1"

	form := NewEmbeddingForm(cfg)
	f := fieldByKey(form, "emb_provider_id")
	if f == nil {
		t.Fatal("missing emb_provider_id field")
	}

	if len(f.Options) < 3 {
		t.Errorf("expected at least 3 options, got %d: %v", len(f.Options), f.Options)
	}

	if f.Value != "gemini-1" {
		t.Errorf("value: want %q, got %q", "gemini-1", f.Value)
	}
}

func TestUpdateConfigFromForm_EmbeddingFields(t *testing.T) {
	state := tuicore.NewConfigState()
	state.Current.Providers = map[string]config.ProviderConfig{
		"my-openai": {Type: "openai", APIKey: "sk-test"},
	}

	form := tuicore.NewFormModel("test")
	form.AddField(&tuicore.Field{Key: "emb_provider_id", Type: tuicore.InputSelect, Value: "my-openai"})
	form.AddField(&tuicore.Field{Key: "emb_model", Type: tuicore.InputText, Value: "text-embedding-3-small"})
	form.AddField(&tuicore.Field{Key: "emb_dimensions", Type: tuicore.InputInt, Value: "1536"})
	form.AddField(&tuicore.Field{Key: "emb_local_baseurl", Type: tuicore.InputText, Value: "http://localhost:11434/v1"})
	form.AddField(&tuicore.Field{Key: "emb_rag_enabled", Type: tuicore.InputBool, Checked: true})
	form.AddField(&tuicore.Field{Key: "emb_rag_max_results", Type: tuicore.InputInt, Value: "5"})
	form.AddField(&tuicore.Field{Key: "emb_rag_collections", Type: tuicore.InputText, Value: "docs,wiki"})

	state.UpdateConfigFromForm(&form)

	e := state.Current.Embedding
	if e.ProviderID != "my-openai" {
		t.Errorf("ProviderID: want %q, got %q", "my-openai", e.ProviderID)
	}
	if e.Provider != "" {
		t.Errorf("Provider: want empty (non-local), got %q", e.Provider)
	}
	if e.Model != "text-embedding-3-small" {
		t.Errorf("Model: want %q, got %q", "text-embedding-3-small", e.Model)
	}
	if e.Dimensions != 1536 {
		t.Errorf("Dimensions: want 1536, got %d", e.Dimensions)
	}
	if e.Local.BaseURL != "http://localhost:11434/v1" {
		t.Errorf("Local.BaseURL: want %q, got %q", "http://localhost:11434/v1", e.Local.BaseURL)
	}
	if !e.RAG.Enabled {
		t.Error("RAG.Enabled: want true")
	}
	if e.RAG.MaxResults != 5 {
		t.Errorf("RAG.MaxResults: want 5, got %d", e.RAG.MaxResults)
	}
	if len(e.RAG.Collections) != 2 || e.RAG.Collections[0] != "docs" || e.RAG.Collections[1] != "wiki" {
		t.Errorf("RAG.Collections: want [docs wiki], got %v", e.RAG.Collections)
	}
}

func TestUpdateConfigFromForm_EmbeddingProviderIDLocal(t *testing.T) {
	state := tuicore.NewConfigState()
	form := tuicore.NewFormModel("test")
	form.AddField(&tuicore.Field{Key: "emb_provider_id", Type: tuicore.InputSelect, Value: "local"})

	state.UpdateConfigFromForm(&form)

	e := state.Current.Embedding
	if e.ProviderID != "" {
		t.Errorf("ProviderID: want empty, got %q", e.ProviderID)
	}
	if e.Provider != "local" {
		t.Errorf("Provider: want %q, got %q", "local", e.Provider)
	}
}

func TestUpdateConfigFromForm_SecurityInterceptorFields(t *testing.T) {
	state := tuicore.NewConfigState()
	form := tuicore.NewFormModel("test")
	form.AddField(&tuicore.Field{Key: "interceptor_timeout", Type: tuicore.InputInt, Value: "60"})
	form.AddField(&tuicore.Field{Key: "interceptor_notify", Type: tuicore.InputSelect, Value: "telegram"})
	form.AddField(&tuicore.Field{Key: "interceptor_sensitive_tools", Type: tuicore.InputText, Value: "exec, browser"})

	state.UpdateConfigFromForm(&form)

	ic := state.Current.Security.Interceptor
	if ic.ApprovalTimeoutSec != 60 {
		t.Errorf("ApprovalTimeoutSec: want 60, got %d", ic.ApprovalTimeoutSec)
	}
	if ic.NotifyChannel != "telegram" {
		t.Errorf("NotifyChannel: want %q, got %q", "telegram", ic.NotifyChannel)
	}
	if len(ic.SensitiveTools) != 2 || ic.SensitiveTools[0] != "exec" || ic.SensitiveTools[1] != "browser" {
		t.Errorf("SensitiveTools: want [exec browser], got %v", ic.SensitiveTools)
	}
}

func TestNewMenuModel_HasEmbeddingCategory(t *testing.T) {
	menu := NewMenuModel()

	found := false
	for _, cat := range menu.Categories {
		if cat.ID == "embedding" {
			found = true
			break
		}
	}
	if !found {
		t.Error("menu missing 'embedding' category")
	}
}

func TestNewMenuModel_HasKnowledgeCategory(t *testing.T) {
	menu := NewMenuModel()

	found := false
	for _, cat := range menu.Categories {
		if cat.ID == "knowledge" {
			found = true
			break
		}
	}
	if !found {
		t.Error("menu missing 'knowledge' category")
	}
}

func TestNewSkillForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewSkillForm(cfg)

	wantKeys := []string{
		"skill_enabled", "skill_dir",
		"skill_allow_import", "skill_max_bulk",
		"skill_import_concurrency", "skill_import_timeout",
	}

	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing field %q", key)
		}
	}

	if f := fieldByKey(form, "skill_enabled"); !f.Checked {
		t.Error("skill_enabled: want true by default")
	}
	if f := fieldByKey(form, "skill_max_bulk"); f.Value != "50" {
		t.Errorf("skill_max_bulk: want %q, got %q", "50", f.Value)
	}
	if f := fieldByKey(form, "skill_import_concurrency"); f.Value != "5" {
		t.Errorf("skill_import_concurrency: want %q, got %q", "5", f.Value)
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		give    string
		wantErr bool
	}{
		{give: "8080", wantErr: false},
		{give: "0", wantErr: true},
		{give: "65536", wantErr: true},
		{give: "abc", wantErr: true},
		{give: strconv.Itoa(18789), wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			err := validatePort(tt.give)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePort(%q): wantErr=%v, got %v", tt.give, tt.wantErr, err)
			}
		})
	}
}
