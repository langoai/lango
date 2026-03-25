package settings

import (
	"strconv"
	"testing"
	"time"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
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

func visibleKeys(form *tuicore.FormModel) map[string]bool {
	keys := make(map[string]bool)
	for _, f := range form.VisibleFields() {
		keys[f.Key] = true
	}
	return keys
}

func TestNewAgentForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewAgentForm(cfg)

	wantKeys := []string{
		"provider", "model", "maxtokens", "temp",
		"prompts_dir", "fallback_provider", "fallback_model",
		"request_timeout", "tool_timeout",
		"auto_extend_timeout", "max_request_timeout",
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
		"interceptor_pii_disabled", "interceptor_pii_custom",
		"presidio_enabled", "presidio_url", "presidio_language",
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

func TestParseCustomPatterns(t *testing.T) {
	tests := []struct {
		give    string
		wantLen int
		wantKey string
		wantVal string
	}{
		{give: "", wantLen: 0},
		{give: "my_id:\\bID-\\d{6}\\b", wantLen: 1, wantKey: "my_id", wantVal: "\\bID-\\d{6}\\b"},
		{give: "a:\\d+,b:\\w+", wantLen: 2},
		{give: "invalid", wantLen: 0}, // no colon
		{give: ":noname", wantLen: 0}, // empty name
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result := ParseCustomPatterns(tt.give)
			if len(result) != tt.wantLen {
				t.Fatalf("want len %d, got %d: %v", tt.wantLen, len(result), result)
			}
			if tt.wantKey != "" {
				if val, ok := result[tt.wantKey]; !ok || val != tt.wantVal {
					t.Errorf("want %q=%q, got %q", tt.wantKey, tt.wantVal, val)
				}
			}
		})
	}
}

func TestFormatCustomPatterns(t *testing.T) {
	tests := []struct {
		give    map[string]string
		wantLen int // just check non-empty
	}{
		{give: nil, wantLen: 0},
		{give: map[string]string{"a": "\\d+"}, wantLen: 1},
	}

	for _, tt := range tests {
		result := formatCustomPatterns(tt.give)
		if tt.wantLen == 0 && result != "" {
			t.Errorf("want empty, got %q", result)
		}
		if tt.wantLen > 0 && result == "" {
			t.Error("want non-empty, got empty")
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

func TestNewRunLedgerForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewRunLedgerForm(cfg)

	wantKeys := []string{
		"runledger_enabled",
		"runledger_shadow",
		"runledger_write_through",
		"runledger_authoritative_read",
		"runledger_workspace_isolation",
		"runledger_stale_ttl",
		"runledger_max_history",
		"runledger_validator_timeout",
		"runledger_planner_retries",
	}

	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing field %q", key)
		}
	}

	if f := fieldByKey(form, "runledger_stale_ttl"); f.Value != "1h0m0s" {
		t.Errorf("runledger_stale_ttl: want %q, got %q", "1h0m0s", f.Value)
	}
	if f := fieldByKey(form, "runledger_validator_timeout"); f.Value != "2m0s" {
		t.Errorf("runledger_validator_timeout: want %q, got %q", "2m0s", f.Value)
	}
}

func TestUpdateConfigFromForm_RunLedgerFields(t *testing.T) {
	state := tuicore.NewConfigState()
	form := tuicore.NewFormModel("runledger")
	form.AddField(&tuicore.Field{Key: "runledger_enabled", Type: tuicore.InputBool, Checked: true})
	form.AddField(&tuicore.Field{Key: "runledger_shadow", Type: tuicore.InputBool, Checked: true})
	form.AddField(&tuicore.Field{Key: "runledger_write_through", Type: tuicore.InputBool, Checked: true})
	form.AddField(&tuicore.Field{Key: "runledger_authoritative_read", Type: tuicore.InputBool, Checked: true})
	form.AddField(&tuicore.Field{Key: "runledger_workspace_isolation", Type: tuicore.InputBool, Checked: true})
	form.AddField(&tuicore.Field{Key: "runledger_stale_ttl", Type: tuicore.InputText, Value: "90m"})
	form.AddField(&tuicore.Field{Key: "runledger_max_history", Type: tuicore.InputInt, Value: "250"})
	form.AddField(&tuicore.Field{Key: "runledger_validator_timeout", Type: tuicore.InputText, Value: "3m"})
	form.AddField(&tuicore.Field{Key: "runledger_planner_retries", Type: tuicore.InputInt, Value: "5"})

	state.UpdateConfigFromForm(&form)

	rl := state.Current.RunLedger
	if !rl.Enabled || !rl.Shadow || !rl.WriteThrough || !rl.AuthoritativeRead || !rl.WorkspaceIsolation {
		t.Fatal("expected all runledger booleans to be true")
	}
	if rl.StaleTTL != 90*time.Minute {
		t.Errorf("StaleTTL: want %v, got %v", 90*time.Minute, rl.StaleTTL)
	}
	if rl.MaxRunHistory != 250 {
		t.Errorf("MaxRunHistory: want 250, got %d", rl.MaxRunHistory)
	}
	if rl.ValidatorTimeout != 3*time.Minute {
		t.Errorf("ValidatorTimeout: want %v, got %v", 3*time.Minute, rl.ValidatorTimeout)
	}
	if rl.PlannerMaxRetries != 5 {
		t.Errorf("PlannerMaxRetries: want 5, got %d", rl.PlannerMaxRetries)
	}
}

func TestNewProvenanceForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewProvenanceForm(cfg)

	wantKeys := []string{
		"provenance_enabled",
		"provenance_auto_on_step_complete",
		"provenance_auto_on_policy",
		"provenance_max_per_session",
		"provenance_retention_days",
	}

	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing field %q", key)
		}
	}

	if f := fieldByKey(form, "provenance_max_per_session"); f.Value != "100" {
		t.Errorf("provenance_max_per_session: want %q, got %q", "100", f.Value)
	}
	if f := fieldByKey(form, "provenance_retention_days"); f.Value != "30" {
		t.Errorf("provenance_retention_days: want %q, got %q", "30", f.Value)
	}
}

func TestUpdateConfigFromForm_ProvenanceFields(t *testing.T) {
	state := tuicore.NewConfigState()
	form := tuicore.NewFormModel("provenance")
	form.AddField(&tuicore.Field{Key: "provenance_enabled", Type: tuicore.InputBool, Checked: true})
	form.AddField(&tuicore.Field{Key: "provenance_auto_on_step_complete", Type: tuicore.InputBool, Checked: false})
	form.AddField(&tuicore.Field{Key: "provenance_auto_on_policy", Type: tuicore.InputBool, Checked: false})
	form.AddField(&tuicore.Field{Key: "provenance_max_per_session", Type: tuicore.InputInt, Value: "250"})
	form.AddField(&tuicore.Field{Key: "provenance_retention_days", Type: tuicore.InputInt, Value: "90"})

	state.UpdateConfigFromForm(&form)

	p := state.Current.Provenance
	if !p.Enabled {
		t.Fatal("expected provenance enabled to be true")
	}
	if p.Checkpoints.AutoOnStepComplete {
		t.Fatal("expected auto_on_step_complete to be false")
	}
	if p.Checkpoints.AutoOnPolicy {
		t.Fatal("expected auto_on_policy to be false")
	}
	if p.Checkpoints.MaxPerSession != 250 {
		t.Errorf("MaxPerSession: want 250, got %d", p.Checkpoints.MaxPerSession)
	}
	if p.Checkpoints.RetentionDays != 90 {
		t.Errorf("RetentionDays: want 90, got %d", p.Checkpoints.RetentionDays)
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
	cfg.Embedding.Provider = "gemini-1"

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
	if e.Provider != "my-openai" {
		t.Errorf("Provider: want %q, got %q", "my-openai", e.Provider)
	}
	if e.ProviderID != "" { //nolint:staticcheck // intentional: verify deprecated field is cleared
		t.Errorf("ProviderID: want empty (deprecated), got %q", e.ProviderID) //nolint:staticcheck // intentional
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

func TestUpdateConfigFromForm_EmbeddingProviderLocal(t *testing.T) {
	state := tuicore.NewConfigState()
	form := tuicore.NewFormModel("test")
	form.AddField(&tuicore.Field{Key: "emb_provider_id", Type: tuicore.InputSelect, Value: "local"})

	state.UpdateConfigFromForm(&form)

	e := state.Current.Embedding
	if e.Provider != "local" {
		t.Errorf("Provider: want %q, got %q", "local", e.Provider)
	}
	if e.ProviderID != "" { //nolint:staticcheck // intentional: verify deprecated field is cleared
		t.Errorf("ProviderID: want empty (deprecated), got %q", e.ProviderID) //nolint:staticcheck // intentional
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
	for _, cat := range menu.AllCategories() {
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
	for _, cat := range menu.AllCategories() {
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

func TestNewP2PForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewP2PForm(cfg)

	wantKeys := []string{
		"p2p_enabled", "p2p_listen_addrs", "p2p_bootstrap_peers",
		"p2p_enable_relay", "p2p_enable_mdns", "p2p_max_peers",
		"p2p_handshake_timeout", "p2p_session_token_ttl",
		"p2p_auto_approve", "p2p_gossip_interval",
		"p2p_zk_handshake", "p2p_zk_attestation",
		"p2p_require_signed_challenge", "p2p_min_trust_score",
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

func TestNewP2PZKPForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewP2PZKPForm(cfg)

	wantKeys := []string{
		"zkp_proof_cache_dir", "zkp_proving_scheme",
		"zkp_srs_mode", "zkp_srs_path", "zkp_max_credential_age",
	}

	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing field %q", key)
		}
	}

	if f := fieldByKey(form, "zkp_proving_scheme"); f.Type != tuicore.InputSelect {
		t.Errorf("zkp_proving_scheme: want InputSelect, got %d", f.Type)
	}
}

func TestNewP2PPricingForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewP2PPricingForm(cfg)

	wantKeys := []string{
		"pricing_enabled", "pricing_per_query", "pricing_tool_prices",
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

func TestNewP2POwnerProtectionForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewP2POwnerProtectionForm(cfg)

	wantKeys := []string{
		"owner_name", "owner_email", "owner_phone",
		"owner_extra_terms", "owner_block_conversations",
	}

	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing field %q", key)
		}
	}

	if f := fieldByKey(form, "owner_block_conversations"); !f.Checked {
		t.Error("owner_block_conversations: want true by default (nil *bool)")
	}
}

func TestNewP2PSandboxForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewP2PSandboxForm(cfg)

	wantKeys := []string{
		"sandbox_enabled", "sandbox_timeout", "sandbox_max_memory_mb",
		"container_enabled", "container_runtime", "container_image",
		"container_network_mode", "container_readonly_rootfs",
		"container_cpu_quota", "container_pool_size", "container_pool_idle_timeout",
	}

	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing field %q", key)
		}
	}

	if f := fieldByKey(form, "container_runtime"); f.Type != tuicore.InputSelect {
		t.Errorf("container_runtime: want InputSelect, got %d", f.Type)
	}
}

func TestNewDBEncryptionForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewDBEncryptionForm(cfg)

	wantKeys := []string{
		"db_encryption_enabled", "db_cipher_page_size",
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

func TestNewKMSForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewKMSForm(cfg)

	wantKeys := []string{
		"kms_backend",
		"kms_region", "kms_key_id", "kms_endpoint",
		"kms_fallback_to_local", "kms_timeout", "kms_max_retries",
		"kms_azure_vault_url", "kms_azure_key_version",
		"kms_pkcs11_module", "kms_pkcs11_slot_id",
		"kms_pkcs11_pin", "kms_pkcs11_key_label",
	}

	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing field %q", key)
		}
	}

	if f := fieldByKey(form, "kms_pkcs11_pin"); f.Type != tuicore.InputPassword {
		t.Errorf("kms_pkcs11_pin: want InputPassword, got %d", f.Type)
	}
}

func TestNewMenuModel_HasP2PCategories(t *testing.T) {
	menu := NewMenuModel()

	wantIDs := []string{
		"p2p", "p2p_zkp", "p2p_pricing", "p2p_owner", "p2p_sandbox",
		"security_db", "security_kms",
	}

	for _, id := range wantIDs {
		found := false
		for _, cat := range menu.AllCategories() {
			if cat.ID == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("menu missing %q category", id)
		}
	}
}

func TestUpdateConfigFromForm_P2PFields(t *testing.T) {
	state := tuicore.NewConfigState()
	form := tuicore.NewFormModel("test")
	form.AddField(&tuicore.Field{Key: "p2p_enabled", Type: tuicore.InputBool, Checked: true})
	form.AddField(&tuicore.Field{Key: "p2p_listen_addrs", Type: tuicore.InputText, Value: "/ip4/0.0.0.0/tcp/9000,/ip4/0.0.0.0/udp/9000"})
	form.AddField(&tuicore.Field{Key: "p2p_max_peers", Type: tuicore.InputInt, Value: "50"})
	form.AddField(&tuicore.Field{Key: "p2p_handshake_timeout", Type: tuicore.InputText, Value: "45s"})
	form.AddField(&tuicore.Field{Key: "p2p_min_trust_score", Type: tuicore.InputText, Value: "0.5"})
	form.AddField(&tuicore.Field{Key: "p2p_zk_handshake", Type: tuicore.InputBool, Checked: true})

	state.UpdateConfigFromForm(&form)

	p := state.Current.P2P
	if !p.Enabled {
		t.Error("P2P.Enabled: want true")
	}
	if len(p.ListenAddrs) != 2 {
		t.Errorf("ListenAddrs: want 2, got %d", len(p.ListenAddrs))
	}
	if p.MaxPeers != 50 {
		t.Errorf("MaxPeers: want 50, got %d", p.MaxPeers)
	}
	if p.HandshakeTimeout != 45*time.Second {
		t.Errorf("HandshakeTimeout: want 45s, got %v", p.HandshakeTimeout)
	}
	if p.MinTrustScore != 0.5 {
		t.Errorf("MinTrustScore: want 0.5, got %f", p.MinTrustScore)
	}
	if !p.ZKHandshake {
		t.Error("ZKHandshake: want true")
	}
}

func TestUpdateConfigFromForm_P2PSandboxBoolPtr(t *testing.T) {
	state := tuicore.NewConfigState()
	form := tuicore.NewFormModel("test")
	form.AddField(&tuicore.Field{Key: "sandbox_enabled", Type: tuicore.InputBool, Checked: true})
	form.AddField(&tuicore.Field{Key: "container_readonly_rootfs", Type: tuicore.InputBool, Checked: false})
	form.AddField(&tuicore.Field{Key: "owner_block_conversations", Type: tuicore.InputBool, Checked: false})

	state.UpdateConfigFromForm(&form)

	if !state.Current.P2P.ToolIsolation.Enabled {
		t.Error("ToolIsolation.Enabled: want true")
	}
	ro := state.Current.P2P.ToolIsolation.Container.ReadOnlyRootfs
	if ro == nil {
		t.Fatal("ReadOnlyRootfs: want non-nil")
	}
	if *ro {
		t.Error("ReadOnlyRootfs: want false")
	}
	bc := state.Current.P2P.OwnerProtection.BlockConversations
	if bc == nil {
		t.Fatal("BlockConversations: want non-nil")
	}
	if *bc {
		t.Error("BlockConversations: want false")
	}
}

func TestUpdateConfigFromForm_KMSFields(t *testing.T) {
	state := tuicore.NewConfigState()
	form := tuicore.NewFormModel("test")
	form.AddField(&tuicore.Field{Key: "kms_region", Type: tuicore.InputText, Value: "us-east-1"})
	form.AddField(&tuicore.Field{Key: "kms_key_id", Type: tuicore.InputText, Value: "arn:aws:kms:us-east-1:123:key/abc"})
	form.AddField(&tuicore.Field{Key: "kms_fallback_to_local", Type: tuicore.InputBool, Checked: true})
	form.AddField(&tuicore.Field{Key: "kms_timeout", Type: tuicore.InputText, Value: "10s"})
	form.AddField(&tuicore.Field{Key: "kms_max_retries", Type: tuicore.InputInt, Value: "5"})
	form.AddField(&tuicore.Field{Key: "kms_azure_vault_url", Type: tuicore.InputText, Value: "https://myvault.vault.azure.net"})
	form.AddField(&tuicore.Field{Key: "kms_pkcs11_slot_id", Type: tuicore.InputInt, Value: "2"})
	form.AddField(&tuicore.Field{Key: "kms_pkcs11_pin", Type: tuicore.InputPassword, Value: "1234"})

	state.UpdateConfigFromForm(&form)

	k := state.Current.Security.KMS
	if k.Region != "us-east-1" {
		t.Errorf("Region: want %q, got %q", "us-east-1", k.Region)
	}
	if k.KeyID != "arn:aws:kms:us-east-1:123:key/abc" {
		t.Errorf("KeyID: want arn..., got %q", k.KeyID)
	}
	if !k.FallbackToLocal {
		t.Error("FallbackToLocal: want true")
	}
	if k.TimeoutPerOperation != 10*time.Second {
		t.Errorf("TimeoutPerOperation: want 10s, got %v", k.TimeoutPerOperation)
	}
	if k.MaxRetries != 5 {
		t.Errorf("MaxRetries: want 5, got %d", k.MaxRetries)
	}
	if k.Azure.VaultURL != "https://myvault.vault.azure.net" {
		t.Errorf("Azure.VaultURL: want vault url, got %q", k.Azure.VaultURL)
	}
	if k.PKCS11.SlotID != 2 {
		t.Errorf("PKCS11.SlotID: want 2, got %d", k.PKCS11.SlotID)
	}
	if k.PKCS11.Pin != "1234" {
		t.Errorf("PKCS11.Pin: want 1234, got %q", k.PKCS11.Pin)
	}
}

func TestNewMenuModel_ShowAdvancedByDefault(t *testing.T) {
	menu := NewMenuModel()
	if !menu.ShowAdvanced() {
		t.Error("showAdvanced should be true by default")
	}
}

func TestNewMenuModel_HasNewSystemCategories(t *testing.T) {
	menu := NewMenuModel()

	wantIDs := []string{"logging", "gatekeeper", "output_manager"}

	for _, id := range wantIDs {
		found := false
		for _, cat := range menu.AllCategories() {
			if cat.ID == id {
				found = true
				if cat.Tier != TierAdvanced {
					t.Errorf("category %q should be TierAdvanced", id)
				}
				break
			}
		}
		if !found {
			t.Errorf("menu missing %q category", id)
		}
	}
}

func TestNewLoggingForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewLoggingForm(cfg)

	wantKeys := []string{"log_level", "log_format", "log_output_path"}

	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing field %q", key)
		}
	}

	if f := fieldByKey(form, "log_level"); f.Value != "info" {
		t.Errorf("log_level: want %q, got %q", "info", f.Value)
	}
	if f := fieldByKey(form, "log_format"); f.Type != tuicore.InputSelect {
		t.Errorf("log_format: want InputSelect, got %d", f.Type)
	}
}

func TestNewGatekeeperForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewGatekeeperForm(cfg)

	wantKeys := []string{
		"gk_enabled", "gk_strip_thought_tags", "gk_strip_internal_markers",
		"gk_strip_raw_json", "gk_raw_json_threshold", "gk_custom_patterns",
	}

	if len(form.Fields) != len(wantKeys) {
		t.Fatalf("expected %d fields, got %d", len(wantKeys), len(form.Fields))
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing field %q", key)
		}
	}

	// Default: enabled = true (derefBool nil => true)
	if f := fieldByKey(form, "gk_enabled"); !f.Checked {
		t.Error("gk_enabled: want true by default")
	}
}

func TestNewOutputManagerForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewOutputManagerForm(cfg)

	wantKeys := []string{
		"om_mgr_enabled", "om_mgr_token_budget",
		"om_mgr_head_ratio", "om_mgr_tail_ratio",
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

func TestUpdateConfigFromForm_LoggingFields(t *testing.T) {
	state := tuicore.NewConfigState()
	form := tuicore.NewFormModel("test")
	form.AddField(&tuicore.Field{Key: "log_level", Type: tuicore.InputSelect, Value: "debug"})
	form.AddField(&tuicore.Field{Key: "log_format", Type: tuicore.InputSelect, Value: "json"})
	form.AddField(&tuicore.Field{Key: "log_output_path", Type: tuicore.InputText, Value: "/var/log/lango.log"})

	state.UpdateConfigFromForm(&form)

	if state.Current.Logging.Level != "debug" {
		t.Errorf("Logging.Level: want %q, got %q", "debug", state.Current.Logging.Level)
	}
	if state.Current.Logging.Format != "json" {
		t.Errorf("Logging.Format: want %q, got %q", "json", state.Current.Logging.Format)
	}
	if state.Current.Logging.OutputPath != "/var/log/lango.log" {
		t.Errorf("Logging.OutputPath: want %q, got %q", "/var/log/lango.log", state.Current.Logging.OutputPath)
	}
}

func TestUpdateConfigFromForm_GatekeeperFields(t *testing.T) {
	state := tuicore.NewConfigState()
	form := tuicore.NewFormModel("test")
	form.AddField(&tuicore.Field{Key: "gk_enabled", Type: tuicore.InputBool, Checked: false})
	form.AddField(&tuicore.Field{Key: "gk_strip_thought_tags", Type: tuicore.InputBool, Checked: true})
	form.AddField(&tuicore.Field{Key: "gk_raw_json_threshold", Type: tuicore.InputInt, Value: "1000"})
	form.AddField(&tuicore.Field{Key: "gk_custom_patterns", Type: tuicore.InputText, Value: "\\[SECRET\\],\\bkey=\\w+"})

	state.UpdateConfigFromForm(&form)

	gk := state.Current.Gatekeeper
	if gk.Enabled == nil || *gk.Enabled {
		t.Error("Gatekeeper.Enabled: want false")
	}
	if gk.StripThoughtTags == nil || !*gk.StripThoughtTags {
		t.Error("Gatekeeper.StripThoughtTags: want true")
	}
	if gk.RawJSONThreshold != 1000 {
		t.Errorf("Gatekeeper.RawJSONThreshold: want 1000, got %d", gk.RawJSONThreshold)
	}
	if len(gk.CustomPatterns) != 2 {
		t.Errorf("Gatekeeper.CustomPatterns: want 2 patterns, got %d", len(gk.CustomPatterns))
	}
}

func TestUpdateConfigFromForm_OutputManagerFields(t *testing.T) {
	state := tuicore.NewConfigState()
	form := tuicore.NewFormModel("test")
	form.AddField(&tuicore.Field{Key: "om_mgr_enabled", Type: tuicore.InputBool, Checked: true})
	form.AddField(&tuicore.Field{Key: "om_mgr_token_budget", Type: tuicore.InputInt, Value: "3000"})
	form.AddField(&tuicore.Field{Key: "om_mgr_head_ratio", Type: tuicore.InputText, Value: "0.6"})
	form.AddField(&tuicore.Field{Key: "om_mgr_tail_ratio", Type: tuicore.InputText, Value: "0.4"})

	state.UpdateConfigFromForm(&form)

	om := state.Current.Tools.OutputManager
	if om.Enabled == nil || !*om.Enabled {
		t.Error("OutputManager.Enabled: want true")
	}
	if om.TokenBudget != 3000 {
		t.Errorf("OutputManager.TokenBudget: want 3000, got %d", om.TokenBudget)
	}
	if om.HeadRatio != 0.6 {
		t.Errorf("OutputManager.HeadRatio: want 0.6, got %f", om.HeadRatio)
	}
	if om.TailRatio != 0.4 {
		t.Errorf("OutputManager.TailRatio: want 0.4, got %f", om.TailRatio)
	}
}

func TestUpdateConfigFromForm_OnChainEscrowFields(t *testing.T) {
	state := tuicore.NewConfigState()
	form := tuicore.NewFormModel("test")
	form.AddField(&tuicore.Field{Key: "economy_escrow_onchain_enabled", Type: tuicore.InputBool, Checked: true})
	form.AddField(&tuicore.Field{Key: "economy_escrow_onchain_mode", Type: tuicore.InputText, Value: "hub"})
	form.AddField(&tuicore.Field{Key: "economy_escrow_onchain_hub_address", Type: tuicore.InputText, Value: "0x1234"})
	form.AddField(&tuicore.Field{Key: "economy_escrow_onchain_vault_factory", Type: tuicore.InputText, Value: "0x5678"})
	form.AddField(&tuicore.Field{Key: "economy_escrow_onchain_vault_impl", Type: tuicore.InputText, Value: "0xabcd"})
	form.AddField(&tuicore.Field{Key: "economy_escrow_onchain_arbitrator", Type: tuicore.InputText, Value: "0xdead"})
	form.AddField(&tuicore.Field{Key: "economy_escrow_onchain_token", Type: tuicore.InputText, Value: "0xbeef"})
	form.AddField(&tuicore.Field{Key: "economy_escrow_onchain_poll_interval", Type: tuicore.InputText, Value: "30s"})
	form.AddField(&tuicore.Field{Key: "economy_escrow_onchain_confirmation_depth", Type: tuicore.InputInt, Value: "5"})
	form.AddField(&tuicore.Field{Key: "economy_escrow_settlement_receipt_timeout", Type: tuicore.InputText, Value: "3m"})
	form.AddField(&tuicore.Field{Key: "economy_escrow_settlement_max_retries", Type: tuicore.InputInt, Value: "5"})

	state.UpdateConfigFromForm(&form)

	oc := state.Current.Economy.Escrow.OnChain
	if !oc.Enabled {
		t.Error("OnChain.Enabled: want true")
	}
	if oc.Mode != "hub" {
		t.Errorf("OnChain.Mode: want %q, got %q", "hub", oc.Mode)
	}
	if oc.HubAddress != "0x1234" {
		t.Errorf("OnChain.HubAddress: want %q, got %q", "0x1234", oc.HubAddress)
	}
	if oc.VaultFactoryAddress != "0x5678" {
		t.Errorf("OnChain.VaultFactoryAddress: want %q, got %q", "0x5678", oc.VaultFactoryAddress)
	}
	if oc.VaultImplementation != "0xabcd" {
		t.Errorf("OnChain.VaultImplementation: want %q, got %q", "0xabcd", oc.VaultImplementation)
	}
	if oc.ArbitratorAddress != "0xdead" {
		t.Errorf("OnChain.ArbitratorAddress: want %q, got %q", "0xdead", oc.ArbitratorAddress)
	}
	if oc.TokenAddress != "0xbeef" {
		t.Errorf("OnChain.TokenAddress: want %q, got %q", "0xbeef", oc.TokenAddress)
	}
	if oc.PollInterval != 30*time.Second {
		t.Errorf("OnChain.PollInterval: want 30s, got %v", oc.PollInterval)
	}
	if oc.ConfirmationDepth != 5 {
		t.Errorf("OnChain.ConfirmationDepth: want 5, got %d", oc.ConfirmationDepth)
	}

	st := state.Current.Economy.Escrow.Settlement
	if st.ReceiptTimeout != 3*time.Minute {
		t.Errorf("Settlement.ReceiptTimeout: want 3m, got %v", st.ReceiptTimeout)
	}
	if st.MaxRetries != 5 {
		t.Errorf("Settlement.MaxRetries: want 5, got %d", st.MaxRetries)
	}
}

func TestMenuModel_SmartFilter_Basic(t *testing.T) {
	menu := NewMenuModel()
	menu.searching = true
	menu.searchInput.SetValue("@basic")
	menu.applyFilter()

	if menu.filtered == nil {
		t.Fatal("filtered should not be nil after @basic filter")
	}

	for _, cat := range menu.filtered {
		if cat.Tier != TierBasic {
			t.Errorf("@basic filter returned advanced category: %q", cat.ID)
		}
	}
}

func TestMenuModel_SmartFilter_Advanced(t *testing.T) {
	menu := NewMenuModel()
	menu.searching = true
	menu.searchInput.SetValue("@advanced")
	menu.applyFilter()

	if menu.filtered == nil {
		t.Fatal("filtered should not be nil after @advanced filter")
	}

	for _, cat := range menu.filtered {
		if cat.Tier != TierAdvanced {
			t.Errorf("@advanced filter returned basic category: %q", cat.ID)
		}
	}
}

func TestMenuModel_SmartFilter_Enabled(t *testing.T) {
	menu := NewMenuModel()
	menu.EnabledChecker = func(id string) bool {
		return id == "agent" || id == "knowledge"
	}
	menu.searching = true
	menu.searchInput.SetValue("@enabled")
	menu.applyFilter()

	if menu.filtered == nil {
		t.Fatal("filtered should not be nil after @enabled filter")
	}
	if len(menu.filtered) != 2 {
		t.Errorf("expected 2 enabled categories, got %d", len(menu.filtered))
	}
}

func TestMenuModel_SmartFilter_Modified(t *testing.T) {
	menu := NewMenuModel()
	menu.DirtyChecker = func(id string) bool {
		return id == "agent"
	}
	menu.searching = true
	menu.searchInput.SetValue("@modified")
	menu.applyFilter()

	if menu.filtered == nil {
		t.Fatal("filtered should not be nil after @modified filter")
	}
	if len(menu.filtered) != 1 || menu.filtered[0].ID != "agent" {
		t.Errorf("expected [agent], got %v", menu.filtered)
	}
}

func TestDerefBool(t *testing.T) {
	tests := []struct {
		give *bool
		def  bool
		want bool
	}{
		{give: nil, def: true, want: true},
		{give: nil, def: false, want: false},
		{give: boolP(true), def: false, want: true},
		{give: boolP(false), def: true, want: false},
	}

	for _, tt := range tests {
		got := derefBool(tt.give, tt.def)
		if got != tt.want {
			t.Errorf("derefBool(%v, %v): want %v, got %v", tt.give, tt.def, tt.want, got)
		}
	}
}

func boolP(b bool) *bool { return &b }

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

// --- Orchestration Settings Tests ---

func TestNewMultiAgentForm_OrchestrationFieldsExist(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewMultiAgentForm(cfg)

	wantKeys := []string{
		"orchestration_mode",
		"orc_cb_failure_threshold", "orc_cb_reset_timeout",
		"orc_budget_tool_call_limit", "orc_budget_delegation_limit", "orc_budget_alert_threshold",
		"orc_recovery_max_retries", "orc_recovery_cooldown",
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing orchestration field %q", key)
		}
	}
}

func TestNewMultiAgentForm_OrchestrationDefaults(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewMultiAgentForm(cfg)

	tests := []struct {
		key  string
		want string
	}{
		{"orchestration_mode", "classic"},
		{"orc_cb_failure_threshold", "3"},
		{"orc_cb_reset_timeout", "30s"},
		{"orc_budget_tool_call_limit", "50"},
		{"orc_budget_delegation_limit", "15"},
		{"orc_budget_alert_threshold", "0.80"},
		{"orc_recovery_max_retries", "2"},
		{"orc_recovery_cooldown", "5m0s"},
	}

	for _, tt := range tests {
		f := fieldByKey(form, tt.key)
		if f == nil {
			t.Errorf("field %q not found", tt.key)
			continue
		}
		if f.Value != tt.want {
			t.Errorf("field %q: want %q, got %q", tt.key, tt.want, f.Value)
		}
	}
}

func TestNewMultiAgentForm_VisibleWhen(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewMultiAgentForm(cfg)

	orcKeys := []string{
		"orc_cb_failure_threshold", "orc_cb_reset_timeout",
		"orc_budget_tool_call_limit", "orc_budget_delegation_limit", "orc_budget_alert_threshold",
		"orc_recovery_max_retries", "orc_recovery_cooldown",
	}

	// multi_agent=false -> orchestration_mode and orc_* hidden
	vis := visibleKeys(form)
	if vis["orchestration_mode"] {
		t.Error("orchestration_mode should be hidden when multi_agent=false")
	}
	for _, key := range orcKeys {
		if vis[key] {
			t.Errorf("%q should be hidden when multi_agent=false", key)
		}
	}

	// multi_agent=true, mode=classic -> orchestration_mode visible, orc_* hidden
	fieldByKey(form, "multi_agent").Checked = true
	vis = visibleKeys(form)
	if !vis["orchestration_mode"] {
		t.Error("orchestration_mode should be visible when multi_agent=true")
	}
	for _, key := range orcKeys {
		if vis[key] {
			t.Errorf("%q should be hidden when mode=classic", key)
		}
	}

	// multi_agent=true, mode=structured -> all orc_* visible
	fieldByKey(form, "orchestration_mode").Value = "structured"
	vis = visibleKeys(form)
	for _, key := range orcKeys {
		if !vis[key] {
			t.Errorf("%q should be visible when mode=structured", key)
		}
	}

	// multi_agent=false again -> everything hidden
	fieldByKey(form, "multi_agent").Checked = false
	vis = visibleKeys(form)
	if vis["orchestration_mode"] {
		t.Error("orchestration_mode should be hidden after disabling multi_agent")
	}
	for _, key := range orcKeys {
		if vis[key] {
			t.Errorf("%q should be hidden after disabling multi_agent", key)
		}
	}
}

// --- Trace Store Settings Tests ---

func TestNewObservabilityForm_TraceStoreFieldsExist(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewObservabilityForm(cfg)

	wantKeys := []string{
		"obs_trace_max_age", "obs_trace_max_traces",
		"obs_trace_failed_multiplier", "obs_trace_cleanup_interval",
	}

	for _, key := range wantKeys {
		if f := fieldByKey(form, key); f == nil {
			t.Errorf("missing trace store field %q", key)
		}
	}
}

func TestNewObservabilityForm_TraceStoreVisibility(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewObservabilityForm(cfg)

	traceKeys := []string{
		"obs_trace_max_age", "obs_trace_max_traces",
		"obs_trace_failed_multiplier", "obs_trace_cleanup_interval",
	}

	// obs_enabled=false -> trace fields hidden
	vis := visibleKeys(form)
	for _, key := range traceKeys {
		if vis[key] {
			t.Errorf("%q should be hidden when obs_enabled=false", key)
		}
	}

	// obs_enabled=true -> trace fields visible
	fieldByKey(form, "obs_enabled").Checked = true
	vis = visibleKeys(form)
	for _, key := range traceKeys {
		if !vis[key] {
			t.Errorf("%q should be visible when obs_enabled=true", key)
		}
	}
}

// --- State Binding Tests ---

func TestUpdateConfigFromForm_OrchestrationFields(t *testing.T) {
	state := tuicore.NewConfigState()
	form := tuicore.NewFormModel("test")

	form.AddField(&tuicore.Field{Key: "orchestration_mode", Type: tuicore.InputText, Value: "structured"})
	form.AddField(&tuicore.Field{Key: "orc_cb_failure_threshold", Type: tuicore.InputInt, Value: "5"})
	form.AddField(&tuicore.Field{Key: "orc_cb_reset_timeout", Type: tuicore.InputText, Value: "1m"})
	form.AddField(&tuicore.Field{Key: "orc_budget_tool_call_limit", Type: tuicore.InputInt, Value: "100"})
	form.AddField(&tuicore.Field{Key: "orc_budget_delegation_limit", Type: tuicore.InputInt, Value: "20"})
	form.AddField(&tuicore.Field{Key: "orc_budget_alert_threshold", Type: tuicore.InputText, Value: "0.75"})
	form.AddField(&tuicore.Field{Key: "orc_recovery_max_retries", Type: tuicore.InputInt, Value: "3"})
	form.AddField(&tuicore.Field{Key: "orc_recovery_cooldown", Type: tuicore.InputText, Value: "10m"})

	state.UpdateConfigFromForm(&form)

	orc := state.Current.Agent.Orchestration
	if orc.Mode != "structured" {
		t.Errorf("Mode: want %q, got %q", "structured", orc.Mode)
	}
	if orc.CircuitBreaker.FailureThreshold != 5 {
		t.Errorf("FailureThreshold: want 5, got %d", orc.CircuitBreaker.FailureThreshold)
	}
	if orc.CircuitBreaker.ResetTimeout != time.Minute {
		t.Errorf("ResetTimeout: want 1m, got %v", orc.CircuitBreaker.ResetTimeout)
	}
	if orc.Budget.ToolCallLimit != 100 {
		t.Errorf("ToolCallLimit: want 100, got %d", orc.Budget.ToolCallLimit)
	}
	if orc.Budget.DelegationLimit != 20 {
		t.Errorf("DelegationLimit: want 20, got %d", orc.Budget.DelegationLimit)
	}
	if orc.Budget.AlertThreshold != 0.75 {
		t.Errorf("AlertThreshold: want 0.75, got %f", orc.Budget.AlertThreshold)
	}
	if orc.Recovery.MaxRetries != 3 {
		t.Errorf("MaxRetries: want 3, got %d", orc.Recovery.MaxRetries)
	}
	if orc.Recovery.CircuitBreakerCooldown != 10*time.Minute {
		t.Errorf("CircuitBreakerCooldown: want 10m, got %v", orc.Recovery.CircuitBreakerCooldown)
	}
}

func TestUpdateConfigFromForm_TraceStoreFields(t *testing.T) {
	state := tuicore.NewConfigState()
	form := tuicore.NewFormModel("test")

	form.AddField(&tuicore.Field{Key: "obs_trace_max_age", Type: tuicore.InputText, Value: "168h"})
	form.AddField(&tuicore.Field{Key: "obs_trace_max_traces", Type: tuicore.InputInt, Value: "5000"})
	form.AddField(&tuicore.Field{Key: "obs_trace_failed_multiplier", Type: tuicore.InputInt, Value: "3"})
	form.AddField(&tuicore.Field{Key: "obs_trace_cleanup_interval", Type: tuicore.InputText, Value: "30m"})

	state.UpdateConfigFromForm(&form)

	ts := state.Current.Observability.TraceStore
	if ts.MaxAge != 168*time.Hour {
		t.Errorf("MaxAge: want 168h, got %v", ts.MaxAge)
	}
	if ts.MaxTraces != 5000 {
		t.Errorf("MaxTraces: want 5000, got %d", ts.MaxTraces)
	}
	if ts.FailedTraceMultiplier != 3 {
		t.Errorf("FailedTraceMultiplier: want 3, got %d", ts.FailedTraceMultiplier)
	}
	if ts.CleanupInterval != 30*time.Minute {
		t.Errorf("CleanupInterval: want 30m, got %v", ts.CleanupInterval)
	}
}

func TestUpdateConfigFromForm_InvalidValues(t *testing.T) {
	state := tuicore.NewConfigState()
	state.Current.Agent.Orchestration.CircuitBreaker.FailureThreshold = 3
	state.Current.Agent.Orchestration.CircuitBreaker.ResetTimeout = 30 * time.Second
	state.Current.Agent.Orchestration.Budget.AlertThreshold = 0.8

	form := tuicore.NewFormModel("test")
	form.AddField(&tuicore.Field{Key: "orc_cb_failure_threshold", Type: tuicore.InputInt, Value: "not-a-number"})
	form.AddField(&tuicore.Field{Key: "orc_cb_reset_timeout", Type: tuicore.InputText, Value: "invalid-duration"})
	form.AddField(&tuicore.Field{Key: "orc_budget_alert_threshold", Type: tuicore.InputText, Value: "not-a-float"})

	state.UpdateConfigFromForm(&form)

	orc := state.Current.Agent.Orchestration
	if orc.CircuitBreaker.FailureThreshold != 3 {
		t.Errorf("FailureThreshold should remain 3, got %d", orc.CircuitBreaker.FailureThreshold)
	}
	if orc.CircuitBreaker.ResetTimeout != 30*time.Second {
		t.Errorf("ResetTimeout should remain 30s, got %v", orc.CircuitBreaker.ResetTimeout)
	}
	if orc.Budget.AlertThreshold != 0.8 {
		t.Errorf("AlertThreshold should remain 0.8, got %f", orc.Budget.AlertThreshold)
	}
}
