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

func TestNewKeyringForm_AllFields(t *testing.T) {
	cfg := defaultTestConfig()
	form := NewKeyringForm(cfg)

	if len(form.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(form.Fields))
	}

	if f := fieldByKey(form, "keyring_enabled"); f == nil {
		t.Error("missing field keyring_enabled")
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
		"security_keyring", "security_db", "security_kms",
	}

	for _, id := range wantIDs {
		found := false
		for _, cat := range menu.Categories {
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

func TestDerefBool(t *testing.T) {
	tests := []struct {
		give    *bool
		def     bool
		want    bool
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
