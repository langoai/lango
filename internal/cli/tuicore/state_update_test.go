package tuicore

import (
	"reflect"
	"testing"
	"time"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/types"
)

func TestUpdateConfigFromFormIgnoresNilForm(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "openai"
	state := NewConfigStateWith(cfg)

	state.UpdateConfigFromForm(nil)

	if state.Current.Agent.Provider != "openai" {
		t.Fatalf("Provider = %q, want unchanged %q", state.Current.Agent.Provider, "openai")
	}
}

func TestUpdateConfigFromFormAppliesPrimarySections(t *testing.T) {
	state := NewConfigStateWith(config.DefaultConfig())

	form := &FormModel{Fields: []*Field{
		textField("provider", "anthropic"),
		textField("model", "claude-3-5-sonnet"),
		textField("maxtokens", "12000"),
		textField("temp", "0.25"),
		textField("prompts_dir", "/tmp/prompts"),
		textField("fallback_provider", "openai"),
		textField("fallback_model", "gpt-4o"),
		textField("request_timeout", "45s"),
		textField("tool_timeout", "20s"),
		boolField("auto_extend_timeout", true),
		textField("max_request_timeout", "2m"),
		boolField("multi_agent", true),
		textField("max_delegation_rounds", "7"),
		textField("max_turns", "33"),
		boolField("error_correction_enabled", false),
		textField("agents_dir", "/tmp/agents"),
		textField("host", "127.0.0.1"),
		textField("port", "19000"),
		boolField("http", true),
		boolField("ws", true),
		boolField("telegram_enabled", true),
		textField("telegram_token", "telegram-token"),
		boolField("discord_enabled", true),
		textField("discord_token", "discord-token"),
		boolField("slack_enabled", true),
		textField("slack_token", "slack-token"),
		textField("slack_app_token", "slack-app-token"),
		textField("exec_timeout", "9s"),
		boolField("exec_bg", true),
		boolField("browser_enabled", true),
		boolField("browser_headless", true),
		textField("browser_session_timeout", "3m"),
		textField("fs_max_read", "65536"),
		textField("ttl", "24h"),
		textField("max_history_turns", "99"),
		boolField("interceptor_enabled", true),
		boolField("interceptor_pii", true),
		textField("interceptor_policy", "configured"),
		textField("interceptor_exempt_tools", "read, write, ,exec"),
		textField("interceptor_timeout", "42"),
		textField("interceptor_notify", "slack"),
		textField("interceptor_sensitive_tools", "exec, payment"),
		textField("interceptor_pii_disabled", "EMAIL_ADDRESS,PHONE_NUMBER"),
		textField("interceptor_pii_custom", "token:[A-Z]+, account:acct-[0-9]+, invalid"),
		boolField("presidio_enabled", true),
		textField("presidio_url", "http://presidio.test"),
		textField("presidio_language", "ko"),
		textField("signer_provider", "aws-kms"),
		textField("signer_rpc", "http://signer.test"),
		textField("signer_keyid", "key-1"),
		boolField("knowledge_enabled", true),
		textField("knowledge_max_context", "12"),
		boolField("skill_enabled", true),
		textField("skill_dir", "/tmp/skills"),
		boolField("skill_allow_import", true),
		textField("skill_max_bulk", "13"),
		textField("skill_import_concurrency", "4"),
		textField("skill_import_timeout", "55s"),
		boolField("om_enabled", true),
		textField("om_provider", "openai"),
		textField("om_model", "gpt-4o-mini"),
		textField("om_msg_threshold", "100"),
		textField("om_obs_threshold", "200"),
		textField("om_max_budget", "300"),
		textField("om_max_reflections", "5"),
		textField("om_max_observations", "8"),
		textField("emb_provider_id", "local"),
		textField("emb_model", "nomic-embed-text"),
		textField("emb_dimensions", "768"),
		textField("emb_local_baseurl", "http://ollama.test"),
		boolField("emb_rag_enabled", true),
		textField("emb_rag_max_results", "6"),
		textField("emb_rag_collections", "notes, docs"),
		boolField("graph_enabled", true),
		textField("graph_backend", "bolt"),
		textField("graph_db_path", "/tmp/graph.db"),
		textField("graph_max_depth", "4"),
		textField("graph_max_expand", "20"),
		textField("orchestration_mode", "structured"),
		textField("orc_cb_failure_threshold", "5"),
		textField("orc_cb_reset_timeout", "10s"),
		textField("orc_budget_tool_call_limit", "80"),
		textField("orc_budget_delegation_limit", "12"),
		textField("orc_budget_alert_threshold", "0.75"),
		textField("orc_recovery_max_retries", "3"),
		textField("orc_recovery_cooldown", "90s"),
		boolField("a2a_enabled", true),
		textField("a2a_base_url", "https://agent.test"),
		textField("a2a_agent_name", "local-agent"),
		textField("a2a_agent_desc", "local description"),
		boolField("cron_enabled", true),
		textField("cron_timezone", "Asia/Seoul"),
		textField("cron_max_jobs", "3"),
		textField("cron_session_mode", "main"),
		textField("cron_history_retention", "72h"),
		textField("cron_default_deliver", "slack, email"),
		boolField("bg_enabled", true),
		textField("bg_yield_ms", "250"),
		textField("bg_max_tasks", "6"),
		textField("bg_default_deliver", "discord,telegram"),
		boolField("wf_enabled", true),
		textField("wf_max_steps", "9"),
		textField("wf_timeout", "11m"),
		textField("wf_state_dir", "/tmp/workflows"),
		textField("wf_default_deliver", "slack"),
		boolField("runledger_enabled", true),
		boolField("runledger_shadow", true),
		boolField("runledger_write_through", true),
		boolField("runledger_authoritative_read", true),
		boolField("runledger_workspace_isolation", true),
		textField("runledger_stale_ttl", "2h"),
		textField("runledger_max_history", "44"),
		textField("runledger_validator_timeout", "30s"),
		textField("runledger_planner_retries", "4"),
		boolField("provenance_enabled", true),
		boolField("provenance_auto_on_step_complete", true),
		boolField("provenance_auto_on_policy", true),
		textField("provenance_max_per_session", "14"),
		textField("provenance_retention_days", "21"),
		boolField("mcp_enabled", true),
		textField("mcp_default_timeout", "17s"),
		textField("mcp_max_output_tokens", "1234"),
		textField("mcp_health_check_interval", "19s"),
		boolField("mcp_auto_reconnect", true),
		textField("mcp_max_reconnect_attempts", "8"),
		boolField("payment_enabled", true),
		textField("payment_wallet_provider", "local"),
		textField("payment_chain_id", "8453"),
		textField("payment_rpc_url", "https://rpc.test"),
		textField("payment_usdc_contract", "0xusdc"),
		textField("payment_max_per_tx", "1.00"),
		textField("payment_max_daily", "10.00"),
		textField("payment_auto_approve", "0.10"),
		boolField("payment_x402_auto", true),
		textField("payment_x402_max", "0.25"),
		boolField("p2p_enabled", true),
		textField("p2p_listen_addrs", "/ip4/0.0.0.0/tcp/9000, /ip4/127.0.0.1/tcp/9001"),
		textField("p2p_bootstrap_peers", "peer-a, peer-b"),
		boolField("p2p_enable_relay", true),
		boolField("p2p_enable_mdns", true),
		textField("p2p_max_peers", "50"),
		textField("p2p_handshake_timeout", "12s"),
		textField("p2p_session_token_ttl", "6h"),
		boolField("p2p_auto_approve", true),
		textField("p2p_gossip_interval", "15s"),
		boolField("p2p_zk_handshake", true),
		boolField("p2p_zk_attestation", true),
		boolField("p2p_require_signed_challenge", true),
		textField("p2p_min_trust_score", "0.42"),
		textField("zkp_proof_cache_dir", "/tmp/proofs"),
		textField("zkp_proving_scheme", "plonk"),
		textField("zkp_srs_mode", "file"),
		textField("zkp_srs_path", "/tmp/srs"),
		textField("zkp_max_credential_age", "48h"),
		boolField("pricing_enabled", true),
		textField("pricing_per_query", "0.50"),
		textField("pricing_tool_prices", "search:0.10,exec:1.50"),
		textField("owner_name", "Owner"),
		textField("owner_email", "owner@example.com"),
		textField("owner_phone", "+15551234567"),
		textField("owner_extra_terms", "secret, private"),
		boolField("owner_block_conversations", true),
		boolField("sandbox_enabled", true),
		textField("sandbox_timeout", "22s"),
		textField("sandbox_max_memory_mb", "512"),
		boolField("container_enabled", true),
		textField("container_runtime", "docker"),
		textField("container_image", "lango-sandbox:test"),
		textField("container_network_mode", "none"),
		boolField("container_readonly_rootfs", true),
		textField("container_cpu_quota", "50000"),
		textField("container_pool_size", "2"),
		textField("container_pool_idle_timeout", "7m"),
		boolField("os_sandbox_enabled", true),
		boolField("os_sandbox_fail_closed", true),
		textField("os_sandbox_backend", "seatbelt"),
		textField("os_sandbox_workspace_path", "/tmp/workspace"),
		textField("os_sandbox_network_mode", "deny"),
		textField("os_sandbox_allowed_ips", "127.0.0.1,10.0.0.1"),
		textField("os_sandbox_allowed_write_paths", "/tmp/a,/tmp/b"),
		textField("os_sandbox_excluded_commands", "git,docker"),
		textField("os_sandbox_timeout", "33s"),
		textField("os_sandbox_seccomp_profile", "strict"),
		textField("os_sandbox_seatbelt_profile", "/tmp/profile.sb"),
	}}

	state.UpdateConfigFromForm(form)
	cfg := state.Current

	if cfg.Agent.Provider != "anthropic" || cfg.Agent.Model != "claude-3-5-sonnet" {
		t.Fatalf("agent provider/model = %q/%q", cfg.Agent.Provider, cfg.Agent.Model)
	}
	if cfg.Agent.MaxTokens != 12000 || cfg.Agent.Temperature != 0.25 {
		t.Fatalf("agent numeric fields = %d/%v", cfg.Agent.MaxTokens, cfg.Agent.Temperature)
	}
	if cfg.Agent.PromptsDir != "/tmp/prompts" || cfg.Agent.FallbackProvider != "openai" || cfg.Agent.FallbackModel != "gpt-4o" {
		t.Fatalf("agent fallback/prompts not updated: %+v", cfg.Agent)
	}
	if cfg.Agent.RequestTimeout != 45*time.Second || cfg.Agent.ToolTimeout != 20*time.Second ||
		!cfg.Agent.AutoExtendTimeout || cfg.Agent.MaxRequestTimeout != 2*time.Minute {
		t.Fatalf("agent timeout fields not updated: %+v", cfg.Agent)
	}
	if !cfg.Agent.MultiAgent || cfg.Agent.MaxDelegationRounds != 7 || cfg.Agent.MaxTurns != 33 ||
		cfg.Agent.ErrorCorrectionEnabled == nil || *cfg.Agent.ErrorCorrectionEnabled || cfg.Agent.AgentsDir != "/tmp/agents" {
		t.Fatalf("agent orchestration fields not updated: %+v", cfg.Agent)
	}
	if cfg.Server.Host != "127.0.0.1" || cfg.Server.Port != 19000 || !cfg.Server.HTTPEnabled || !cfg.Server.WebSocketEnabled {
		t.Fatalf("server fields not updated: %+v", cfg.Server)
	}
	if !cfg.Channels.Telegram.Enabled || cfg.Channels.Telegram.BotToken != "telegram-token" ||
		!cfg.Channels.Discord.Enabled || cfg.Channels.Discord.BotToken != "discord-token" ||
		!cfg.Channels.Slack.Enabled || cfg.Channels.Slack.BotToken != "slack-token" || cfg.Channels.Slack.AppToken != "slack-app-token" {
		t.Fatalf("channel fields not updated: %+v", cfg.Channels)
	}
	if cfg.Tools.Exec.DefaultTimeout != 9*time.Second || !cfg.Tools.Exec.AllowBackground ||
		!cfg.Tools.Browser.Enabled || !cfg.Tools.Browser.Headless || cfg.Tools.Browser.SessionTimeout != 3*time.Minute ||
		cfg.Tools.Filesystem.MaxReadSize != 65536 {
		t.Fatalf("tool fields not updated: %+v", cfg.Tools)
	}
	if cfg.Session.TTL != 24*time.Hour || cfg.Session.MaxHistoryTurns != 99 {
		t.Fatalf("session fields not updated: %+v", cfg.Session)
	}
	if !cfg.Security.Interceptor.Enabled || !cfg.Security.Interceptor.RedactPII ||
		cfg.Security.Interceptor.ApprovalPolicy != config.ApprovalPolicyConfigured ||
		!reflect.DeepEqual(cfg.Security.Interceptor.ExemptTools, []string{"read", "write", "exec"}) ||
		cfg.Security.Interceptor.ApprovalTimeoutSec != 42 ||
		cfg.Security.Interceptor.NotifyChannel != "slack" ||
		!reflect.DeepEqual(cfg.Security.Interceptor.SensitiveTools, []string{"exec", "payment"}) ||
		!reflect.DeepEqual(cfg.Security.Interceptor.PIIDisabledPatterns, []string{"EMAIL_ADDRESS", "PHONE_NUMBER"}) ||
		!reflect.DeepEqual(cfg.Security.Interceptor.PIICustomPatterns, map[string]string{"token": "[A-Z]+", "account": "acct-[0-9]+"}) {
		t.Fatalf("interceptor fields not updated: %+v", cfg.Security.Interceptor)
	}
	if !cfg.Security.Interceptor.Presidio.Enabled || cfg.Security.Interceptor.Presidio.URL != "http://presidio.test" ||
		cfg.Security.Interceptor.Presidio.Language != "ko" ||
		cfg.Security.Signer.Provider != "aws-kms" || cfg.Security.Signer.RPCUrl != "http://signer.test" ||
		cfg.Security.Signer.KeyID != "key-1" {
		t.Fatalf("security fields not updated: %+v", cfg.Security)
	}
	if !cfg.Knowledge.Enabled || cfg.Knowledge.MaxContextPerLayer != 12 ||
		!cfg.Skill.Enabled || cfg.Skill.SkillsDir != "/tmp/skills" || !cfg.Skill.AllowImport ||
		cfg.Skill.MaxBulkImport != 13 || cfg.Skill.ImportConcurrency != 4 || cfg.Skill.ImportTimeout != 55*time.Second {
		t.Fatalf("knowledge/skill fields not updated: %+v / %+v", cfg.Knowledge, cfg.Skill)
	}
	if !cfg.ObservationalMemory.Enabled || cfg.ObservationalMemory.Provider != "openai" ||
		cfg.ObservationalMemory.Model != "gpt-4o-mini" || cfg.ObservationalMemory.MessageTokenThreshold != 100 ||
		cfg.ObservationalMemory.ObservationTokenThreshold != 200 || cfg.ObservationalMemory.MaxMessageTokenBudget != 300 ||
		cfg.ObservationalMemory.MaxReflectionsInContext != 5 || cfg.ObservationalMemory.MaxObservationsInContext != 8 {
		t.Fatalf("observational memory fields not updated: %+v", cfg.ObservationalMemory)
	}
	//nolint:staticcheck // intentional: regression test confirms legacy ProviderID is cleared after migration.
	if cfg.Embedding.Provider != "local" || cfg.Embedding.ProviderID != "" || cfg.Embedding.Model != "nomic-embed-text" ||
		cfg.Embedding.Dimensions != 768 || cfg.Embedding.Local.BaseURL != "http://ollama.test" ||
		!cfg.Embedding.RAG.Enabled || cfg.Embedding.RAG.MaxResults != 6 ||
		!reflect.DeepEqual(cfg.Embedding.RAG.Collections, []string{"notes", "docs"}) {
		t.Fatalf("embedding fields not updated: %+v", cfg.Embedding)
	}
	if !cfg.Graph.Enabled || cfg.Graph.Backend != "bolt" || cfg.Graph.DatabasePath != "/tmp/graph.db" ||
		cfg.Graph.MaxTraversalDepth != 4 || cfg.Graph.MaxExpansionResults != 20 {
		t.Fatalf("graph fields not updated: %+v", cfg.Graph)
	}
	if cfg.Agent.Orchestration.Mode != "structured" ||
		cfg.Agent.Orchestration.CircuitBreaker.FailureThreshold != 5 ||
		cfg.Agent.Orchestration.CircuitBreaker.ResetTimeout != 10*time.Second ||
		cfg.Agent.Orchestration.Budget.ToolCallLimit != 80 ||
		cfg.Agent.Orchestration.Budget.DelegationLimit != 12 ||
		cfg.Agent.Orchestration.Budget.AlertThreshold != 0.75 ||
		cfg.Agent.Orchestration.Recovery.MaxRetries != 3 ||
		cfg.Agent.Orchestration.Recovery.CircuitBreakerCooldown != 90*time.Second {
		t.Fatalf("orchestration fields not updated: %+v", cfg.Agent.Orchestration)
	}
	if !cfg.A2A.Enabled || cfg.A2A.BaseURL != "https://agent.test" || cfg.A2A.AgentName != "local-agent" ||
		cfg.A2A.AgentDescription != "local description" {
		t.Fatalf("a2a fields not updated: %+v", cfg.A2A)
	}
	if !cfg.Cron.Enabled || cfg.Cron.Timezone != "Asia/Seoul" || cfg.Cron.MaxConcurrentJobs != 3 ||
		cfg.Cron.DefaultSessionMode != "main" || cfg.Cron.HistoryRetention != "72h" ||
		!reflect.DeepEqual(cfg.Cron.DefaultDeliverTo, []string{"slack", "email"}) {
		t.Fatalf("cron fields not updated: %+v", cfg.Cron)
	}
	if !cfg.Background.Enabled || cfg.Background.YieldMs != 250 || cfg.Background.MaxConcurrentTasks != 6 ||
		!reflect.DeepEqual(cfg.Background.DefaultDeliverTo, []string{"discord", "telegram"}) {
		t.Fatalf("background fields not updated: %+v", cfg.Background)
	}
	if !cfg.Workflow.Enabled || cfg.Workflow.MaxConcurrentSteps != 9 || cfg.Workflow.DefaultTimeout != 11*time.Minute ||
		cfg.Workflow.StateDir != "/tmp/workflows" || !reflect.DeepEqual(cfg.Workflow.DefaultDeliverTo, []string{"slack"}) {
		t.Fatalf("workflow fields not updated: %+v", cfg.Workflow)
	}
	if !cfg.RunLedger.Enabled || !cfg.RunLedger.Shadow || !cfg.RunLedger.WriteThrough || !cfg.RunLedger.AuthoritativeRead ||
		!cfg.RunLedger.WorkspaceIsolation || cfg.RunLedger.StaleTTL != 2*time.Hour || cfg.RunLedger.MaxRunHistory != 44 ||
		cfg.RunLedger.ValidatorTimeout != 30*time.Second || cfg.RunLedger.PlannerMaxRetries != 4 {
		t.Fatalf("runledger fields not updated: %+v", cfg.RunLedger)
	}
	if !cfg.Provenance.Enabled || !cfg.Provenance.Checkpoints.AutoOnStepComplete ||
		!cfg.Provenance.Checkpoints.AutoOnPolicy || cfg.Provenance.Checkpoints.MaxPerSession != 14 ||
		cfg.Provenance.Checkpoints.RetentionDays != 21 {
		t.Fatalf("provenance fields not updated: %+v", cfg.Provenance)
	}
	if !cfg.MCP.Enabled || cfg.MCP.DefaultTimeout != 17*time.Second || cfg.MCP.MaxOutputTokens != 1234 ||
		cfg.MCP.HealthCheckInterval != 19*time.Second || !cfg.MCP.AutoReconnect || cfg.MCP.MaxReconnectAttempts != 8 {
		t.Fatalf("mcp fields not updated: %+v", cfg.MCP)
	}
	if !cfg.Payment.Enabled || cfg.Payment.WalletProvider != "local" || cfg.Payment.Network.ChainID != 8453 ||
		cfg.Payment.Network.RPCURL != "https://rpc.test" || cfg.Payment.Network.USDCContract != "0xusdc" ||
		cfg.Payment.Limits.MaxPerTx != "1.00" || cfg.Payment.Limits.MaxDaily != "10.00" ||
		cfg.Payment.Limits.AutoApproveBelow != "0.10" || !cfg.Payment.X402.AutoIntercept ||
		cfg.Payment.X402.MaxAutoPayAmount != "0.25" {
		t.Fatalf("payment fields not updated: %+v", cfg.Payment)
	}
	assertPrimaryP2PConfig(t, cfg.P2P)
	assertPrimarySandboxConfig(t, cfg.Sandbox)
}

func TestUpdateConfigFromFormAppliesAdvancedPolicySections(t *testing.T) {
	state := NewConfigStateWith(config.DefaultConfig())
	state.Current.Alerting.Delivery = []config.AlertDeliveryConfig{
		{Type: "email", WebhookURL: "mailto:ops@example.com", MinSeverity: "critical"},
	}

	form := &FormModel{Fields: []*Field{
		boolField("ontology_enabled", true),
		boolField("ontology_acl_enabled", true),
		textField("ontology_acl_roles", "alice=admin,bob=read"),
		textField("ontology_acl_p2p_permission", "write"),
		boolField("ontology_gov_enabled", true),
		textField("ontology_gov_max_new_per_day", "10"),
		textField("ontology_gov_quarantine_hrs", "24"),
		textField("ontology_gov_shadow_hrs", "48"),
		textField("ontology_gov_min_usage", "3"),
		textField("ontology_gov_explosion_budget", "100"),
		textField("ontology_gov_admission_mode", "observe"),
		textField("ontology_gov_learning_conf", "0.7"),
		textField("ontology_gov_librarian_conf", "0.6"),
		boolField("ontology_ex_enabled", true),
		textField("ontology_ex_min_trust_schema", "0.55"),
		textField("ontology_ex_min_trust_facts", "0.75"),
		textField("ontology_ex_auto_import_mode", "governed"),
		textField("ontology_ex_max_types", "15"),
		boolField("alerting_enabled", true),
		textField("alerting_policy_block_rate", "12"),
		textField("alerting_recovery_retries", "9"),
		textField("alerting_webhook_url", "https://hooks.test"),
		textField("kms_backend", "gcp-kms"),
		textField("kms_region", "us-central1"),
		textField("kms_key_id", "projects/p/keys/k"),
		textField("kms_endpoint", "https://kms.test"),
		boolField("kms_fallback_to_local", true),
		textField("kms_timeout", "6s"),
		textField("kms_max_retries", "5"),
		textField("kms_azure_vault_url", "https://vault.test"),
		textField("kms_azure_key_version", "v1"),
		textField("kms_pkcs11_module", "/usr/lib/pkcs11.so"),
		textField("kms_pkcs11_slot_id", "2"),
		textField("kms_pkcs11_pin", "1234"),
		textField("kms_pkcs11_key_label", "signing-key"),
		textField("log_level", "debug"),
		textField("log_format", "json"),
		textField("log_output_path", "/tmp/lango.log"),
		boolField("gk_enabled", true),
		boolField("gk_strip_thought_tags", true),
		boolField("gk_strip_internal_markers", true),
		boolField("gk_strip_raw_json", true),
		textField("gk_raw_json_threshold", "777"),
		textField("gk_custom_patterns", "secret,internal"),
		boolField("om_mgr_enabled", true),
		textField("om_mgr_token_budget", "5000"),
		textField("om_mgr_head_ratio", "0.6"),
		textField("om_mgr_tail_ratio", "0.4"),
		boolField("hooks_enabled", true),
		boolField("hooks_security_filter", true),
		boolField("hooks_access_control", true),
		boolField("hooks_event_publishing", true),
		boolField("hooks_knowledge_save", true),
		textField("hooks_blocked_commands", "rm,curl"),
		boolField("agent_memory_enabled", true),
		boolField("lib_enabled", true),
		textField("lib_obs_threshold", "2"),
		textField("lib_cooldown", "3"),
		textField("lib_max_inquiries", "4"),
		textField("lib_auto_save", "medium"),
		textField("lib_provider", "anthropic"),
		textField("lib_model", "claude"),
		boolField("economy_enabled", true),
		textField("economy_budget_default_max", "20.00"),
		boolField("economy_budget_hard_limit", true),
		textField("economy_budget_alert_thresholds", "0.5, 0.8, bad, 0.95"),
		textField("economy_risk_escrow_threshold", "5.00"),
		textField("economy_risk_high_trust", "0.9"),
		textField("economy_risk_medium_trust", "0.6"),
		boolField("economy_negotiate_enabled", true),
		textField("economy_negotiate_max_rounds", "6"),
		textField("economy_negotiate_timeout", "5m"),
		boolField("economy_negotiate_auto", true),
		textField("economy_negotiate_max_discount", "0.2"),
		boolField("economy_escrow_enabled", true),
		textField("economy_escrow_default_timeout", "24h"),
		textField("economy_escrow_max_milestones", "8"),
		boolField("economy_escrow_auto_release", true),
		textField("economy_escrow_dispute_window", "1h"),
		boolField("economy_escrow_onchain_enabled", true),
		textField("economy_escrow_onchain_mode", "hub"),
		textField("economy_escrow_onchain_hub_address", "0xhub"),
		textField("economy_escrow_onchain_vault_factory", "0xfactory"),
		textField("economy_escrow_onchain_vault_impl", "0ximpl"),
		textField("economy_escrow_onchain_arbitrator", "0xarb"),
		textField("economy_escrow_onchain_token", "0xtoken"),
		textField("economy_escrow_onchain_poll_interval", "15s"),
		textField("economy_escrow_onchain_confirmation_depth", "2"),
		textField("economy_escrow_settlement_receipt_timeout", "3m"),
		textField("economy_escrow_settlement_max_retries", "4"),
		boolField("economy_pricing_enabled", true),
		textField("economy_pricing_trust_discount", "0.15"),
		textField("economy_pricing_volume_discount", "0.05"),
		textField("economy_pricing_min_price", "0.01"),
		boolField("obs_enabled", true),
		boolField("obs_tokens_enabled", true),
		boolField("obs_tokens_persist", true),
		textField("obs_tokens_retention", "30"),
		boolField("obs_health_enabled", true),
		textField("obs_health_interval", "45s"),
		boolField("obs_audit_enabled", true),
		textField("obs_audit_retention", "90"),
		boolField("obs_metrics_enabled", true),
		textField("obs_metrics_format", "prometheus"),
		textField("obs_trace_max_age", "720h"),
		textField("obs_trace_max_traces", "10000"),
		textField("obs_trace_failed_multiplier", "3"),
		textField("obs_trace_cleanup_interval", "1h"),
		boolField("sa_enabled", true),
		textField("sa_factory_address", "0xfactory"),
		textField("sa_entrypoint_address", "0xentry"),
		textField("sa_singleton_address", "0xsafe"),
		textField("sa_safe7579_address", "0x7579"),
		textField("sa_fallback_handler", "0xfallback"),
		textField("sa_bundler_url", "https://bundler.test"),
		textField("sa_session_max_duration", "2h"),
		textField("sa_session_default_gas_limit", "100000"),
		textField("sa_session_max_active_keys", "5"),
		boolField("sa_paymaster_enabled", true),
		textField("sa_paymaster_provider", "circle"),
		textField("sa_paymaster_mode", "permit"),
		textField("sa_paymaster_rpc_url", "https://paymaster.test"),
		textField("sa_paymaster_token_address", "0xusdc"),
		textField("sa_paymaster_address", "0xpaymaster"),
		textField("sa_paymaster_policy_id", "policy-1"),
		textField("sa_paymaster_fallback_mode", "direct"),
		textField("sa_modules_session_validator", "0xvalidator"),
		textField("sa_modules_spending_hook", "0xhook"),
		textField("sa_modules_escrow_executor", "0xexecutor"),
		textField("ctx_profile", "balanced"),
		boolField("retrieval_enabled", true),
		boolField("retrieval_feedback", true),
		boolField("aa_enabled", true),
		textField("aa_mode", "active"),
		textField("aa_boost_delta", "0.2"),
		textField("aa_decay_delta", "0.03"),
		textField("aa_decay_interval", "25"),
		textField("aa_min_score", "0.2"),
		textField("aa_max_score", "3.5"),
		textField("aa_warmup_turns", "10"),
		textField("ctx_model_window", "128000"),
		textField("ctx_response_reserve", "4096"),
		textField("ctx_alloc_knowledge", "0.2"),
		textField("ctx_alloc_rag", "0.3"),
		textField("ctx_alloc_memory", "0.1"),
		textField("ctx_alloc_run_summary", "0.15"),
		textField("ctx_alloc_headroom", "0.25"),
	}}

	state.UpdateConfigFromForm(form)
	cfg := state.Current

	assertOntologyConfig(t, cfg.Ontology)
	assertAlertingConfig(t, cfg.Alerting)
	assertSecurityKMSConfig(t, cfg.Security)
	assertLoggingGatekeeperOutputConfig(t, cfg)
	assertHooksMemoryLibrarianConfig(t, cfg)
	assertEconomyConfig(t, cfg.Economy)
	assertObservabilityConfig(t, cfg.Observability)
	assertSmartAccountConfig(t, cfg.SmartAccount)
	if cfg.ContextProfile != config.ContextProfileName("balanced") {
		t.Fatalf("ContextProfile = %q", cfg.ContextProfile)
	}
	if !cfg.Retrieval.Enabled || !cfg.Retrieval.Feedback || !cfg.Retrieval.AutoAdjust.Enabled ||
		cfg.Retrieval.AutoAdjust.Mode != "active" || cfg.Retrieval.AutoAdjust.BoostDelta != 0.2 ||
		cfg.Retrieval.AutoAdjust.DecayDelta != 0.03 || cfg.Retrieval.AutoAdjust.DecayInterval != 25 ||
		cfg.Retrieval.AutoAdjust.MinScore != 0.2 || cfg.Retrieval.AutoAdjust.MaxScore != 3.5 ||
		cfg.Retrieval.AutoAdjust.WarmupTurns != 10 {
		t.Fatalf("retrieval fields not updated: %+v", cfg.Retrieval)
	}
	if cfg.Context.ModelWindow != 128000 || cfg.Context.ResponseReserve != 4096 ||
		cfg.Context.Allocation.Knowledge != 0.2 || cfg.Context.Allocation.RAG != 0.3 ||
		cfg.Context.Allocation.Memory != 0.1 || cfg.Context.Allocation.RunSummary != 0.15 ||
		cfg.Context.Allocation.Headroom != 0.25 {
		t.Fatalf("context fields not updated: %+v", cfg.Context)
	}
}

func TestUpdateConfigFromFormPreservesOntologyAbsentDefaultsUntilEdited(t *testing.T) {
	state := NewConfigStateWith(config.DefaultConfig())
	state.Current.Ontology.Governance.AdmissionMode = "off"
	state.Current.Ontology.Governance.AdmissionModePresent = false
	state.Current.Ontology.Governance.LearningDefaultConfidence = config.OntologyLearningDefaultConfidenceFallback
	state.Current.Ontology.Governance.LearningDefaultConfidencePresent = false

	form := &FormModel{Fields: []*Field{
		{
			Key:                       "ontology_gov_admission_mode",
			Value:                     "off",
			InitialValue:              "off",
			PreserveAbsentIfUntouched: true,
		},
		{
			Key:                       "ontology_gov_learning_conf",
			Value:                     "0.6",
			InitialValue:              "0.6",
			PreserveAbsentIfUntouched: true,
		},
	}}

	state.UpdateConfigFromForm(form)

	if state.Current.Ontology.Governance.AdmissionModePresent {
		t.Fatal("AdmissionModePresent should remain false when untouched default is preserved")
	}
	if state.Current.Ontology.Governance.LearningDefaultConfidencePresent {
		t.Fatal("LearningDefaultConfidencePresent should remain false when untouched default is preserved")
	}

	form.Fields[0].Value = "observe"
	form.Fields[1].Value = "0.8"
	state.UpdateConfigFromForm(form)

	if state.Current.Ontology.Governance.AdmissionMode != "observe" ||
		!state.Current.Ontology.Governance.AdmissionModePresent {
		t.Fatalf("admission mode was not marked present after edit: %+v", state.Current.Ontology.Governance)
	}
	if state.Current.Ontology.Governance.LearningDefaultConfidence != 0.8 ||
		!state.Current.Ontology.Governance.LearningDefaultConfidencePresent ||
		state.Current.Ontology.Governance.LearningDefaultConfidenceBackfillNeeded {
		t.Fatalf("learning confidence was not marked present after edit: %+v", state.Current.Ontology.Governance)
	}
}

func TestUpdateConfigFromFormRemovesWebhookWithoutDroppingOtherDeliveryChannels(t *testing.T) {
	state := NewConfigStateWith(config.DefaultConfig())
	state.Current.Alerting.Delivery = []config.AlertDeliveryConfig{
		{Type: "email", WebhookURL: "mailto:ops@example.com", MinSeverity: "critical"},
		{Type: "webhook", WebhookURL: "https://old.example.com", MinSeverity: "warning"},
	}

	state.UpdateConfigFromForm(&FormModel{Fields: []*Field{textField("alerting_webhook_url", "")}})

	if len(state.Current.Alerting.Delivery) != 1 || state.Current.Alerting.Delivery[0].Type != "email" {
		t.Fatalf("Delivery = %+v, want only non-webhook channel preserved", state.Current.Alerting.Delivery)
	}
}

func TestUpdateConfigFromFormKeepsExistingValuesOnInvalidNumbers(t *testing.T) {
	state := NewConfigStateWith(config.DefaultConfig())
	state.Current.Agent.MaxTokens = 123
	state.Current.Server.Port = 456
	state.Current.Agent.RequestTimeout = 7 * time.Second
	state.Current.Payment.Network.ChainID = 8453

	state.UpdateConfigFromForm(&FormModel{Fields: []*Field{
		textField("maxtokens", "not-int"),
		textField("port", "not-int"),
		textField("request_timeout", "not-duration"),
		textField("payment_chain_id", "not-int64"),
	}})

	if state.Current.Agent.MaxTokens != 123 || state.Current.Server.Port != 456 ||
		state.Current.Agent.RequestTimeout != 7*time.Second || state.Current.Payment.Network.ChainID != 8453 {
		t.Fatalf("invalid inputs changed config: %+v", state.Current)
	}
}

func TestUpdateAuthProviderFromFormUsesExplicitOrFormID(t *testing.T) {
	state := NewConfigStateWith(config.DefaultConfig())

	state.UpdateAuthProviderFromForm("", &FormModel{Fields: []*Field{
		textField("oidc_id", "corp"),
		textField("oidc_issuer", "https://issuer.test"),
		textField("oidc_client_id", "client"),
		textField("oidc_client_secret", "secret"),
		textField("oidc_redirect", "https://app.test/callback"),
		textField("oidc_scopes", "openid, profile, email"),
	}})

	provider := state.Current.Auth.Providers["corp"]
	if provider.IssuerURL != "https://issuer.test" || provider.ClientID != "client" ||
		provider.ClientSecret != "secret" || provider.RedirectURL != "https://app.test/callback" ||
		!reflect.DeepEqual(provider.Scopes, []string{"openid", "profile", "email"}) || !state.IsDirty("auth") {
		t.Fatalf("OIDC provider not updated: %+v dirty=%v", provider, state.Dirty)
	}

	state.UpdateAuthProviderFromForm("corp", &FormModel{Fields: []*Field{textField("oidc_client_id", "updated")}})
	if state.Current.Auth.Providers["corp"].ClientID != "updated" {
		t.Fatalf("explicit OIDC provider update was not applied: %+v", state.Current.Auth.Providers["corp"])
	}
}

func TestUpdateProviderFromFormUsesExplicitOrFormID(t *testing.T) {
	state := NewConfigStateWith(config.DefaultConfig())

	state.UpdateProviderFromForm("", &FormModel{Fields: []*Field{
		textField("id", "local-openai"),
		textField("type", "openai"),
		textField("apikey", "${OPENAI_API_KEY}"),
		textField("baseurl", "https://api.openai.com/v1"),
	}})

	provider := state.Current.Providers["local-openai"]
	if provider.Type != types.ProviderOpenAI || provider.APIKey != "${OPENAI_API_KEY}" ||
		provider.BaseURL != "https://api.openai.com/v1" || !state.IsDirty("providers") {
		t.Fatalf("provider not updated: %+v dirty=%v", provider, state.Dirty)
	}

	state.UpdateProviderFromForm("local-openai", &FormModel{Fields: []*Field{textField("type", "anthropic")}})
	if state.Current.Providers["local-openai"].Type != types.ProviderAnthropic {
		t.Fatalf("explicit provider update was not applied: %+v", state.Current.Providers["local-openai"])
	}
}

func TestUpdateMCPServerFromFormUsesExplicitOrFormName(t *testing.T) {
	state := NewConfigStateWith(config.DefaultConfig())

	state.UpdateMCPServerFromForm("", &FormModel{Fields: []*Field{
		textField("mcp_srv_name", "filesystem"),
		textField("mcp_srv_transport", "stdio"),
		textField("mcp_srv_command", "lango-mcp"),
		textField("mcp_srv_args", "--root, /tmp"),
		textField("mcp_srv_url", "http://mcp.test"),
		textField("mcp_srv_env", "A=1,B=two"),
		textField("mcp_srv_headers", "Authorization=Bearer test"),
		boolField("mcp_srv_enabled", false),
		textField("mcp_srv_timeout", "12s"),
		textField("mcp_srv_safety", "moderate"),
	}})

	server := state.Current.MCP.Servers["filesystem"]
	if server.Transport != "stdio" || server.Command != "lango-mcp" ||
		!reflect.DeepEqual(server.Args, []string{"--root", "/tmp"}) ||
		server.URL != "http://mcp.test" ||
		!reflect.DeepEqual(server.Env, map[string]string{"A": "1", "B": "two"}) ||
		!reflect.DeepEqual(server.Headers, map[string]string{"Authorization": "Bearer test"}) ||
		server.Enabled == nil || *server.Enabled || server.Timeout != 12*time.Second ||
		server.SafetyLevel != "moderate" || !state.IsDirty("mcp") {
		t.Fatalf("MCP server not updated: %+v dirty=%v", server, state.Dirty)
	}

	state.UpdateMCPServerFromForm("filesystem", &FormModel{Fields: []*Field{textField("mcp_srv_timeout", "")}})
	if state.Current.MCP.Servers["filesystem"].Timeout != 0 {
		t.Fatalf("empty timeout should clear override: %+v", state.Current.MCP.Servers["filesystem"])
	}
}

func textField(key, value string) *Field {
	return &Field{Key: key, Value: value, InitialValue: value}
}

func boolField(key string, checked bool) *Field {
	return &Field{Key: key, Type: InputBool, Checked: checked}
}

func assertBoolPtr(t *testing.T, name string, got *bool, want bool) {
	t.Helper()
	if got == nil || *got != want {
		t.Fatalf("%s = %v, want %v", name, got, want)
	}
}

func assertPrimaryP2PConfig(t *testing.T, cfg config.P2PConfig) {
	t.Helper()
	if !cfg.Enabled || !reflect.DeepEqual(cfg.ListenAddrs, []string{"/ip4/0.0.0.0/tcp/9000", "/ip4/127.0.0.1/tcp/9001"}) ||
		!reflect.DeepEqual(cfg.BootstrapPeers, []string{"peer-a", "peer-b"}) || !cfg.EnableRelay ||
		!cfg.EnableMDNS || cfg.MaxPeers != 50 || cfg.HandshakeTimeout != 12*time.Second ||
		cfg.SessionTokenTTL != 6*time.Hour || !cfg.AutoApproveKnownPeers || cfg.GossipInterval != 15*time.Second ||
		!cfg.ZKHandshake || !cfg.ZKAttestation || !cfg.RequireSignedChallenge || cfg.MinTrustScore != 0.42 {
		t.Fatalf("p2p network fields not updated: %+v", cfg)
	}
	if cfg.ZKP.ProofCacheDir != "/tmp/proofs" || cfg.ZKP.ProvingScheme != "plonk" ||
		cfg.ZKP.SRSMode != "file" || cfg.ZKP.SRSPath != "/tmp/srs" || cfg.ZKP.MaxCredentialAge != "48h" {
		t.Fatalf("p2p zkp fields not updated: %+v", cfg.ZKP)
	}
	if !cfg.Pricing.Enabled || cfg.Pricing.PerQuery != "0.50" ||
		!reflect.DeepEqual(cfg.Pricing.ToolPrices, map[string]string{"search": "0.10", "exec": "1.50"}) {
		t.Fatalf("p2p pricing fields not updated: %+v", cfg.Pricing)
	}
	if cfg.OwnerProtection.OwnerName != "Owner" || cfg.OwnerProtection.OwnerEmail != "owner@example.com" ||
		cfg.OwnerProtection.OwnerPhone != "+15551234567" ||
		!reflect.DeepEqual(cfg.OwnerProtection.ExtraTerms, []string{"secret", "private"}) {
		t.Fatalf("p2p owner fields not updated: %+v", cfg.OwnerProtection)
	}
	assertBoolPtr(t, "owner block conversations", cfg.OwnerProtection.BlockConversations, true)
	if !cfg.ToolIsolation.Enabled || cfg.ToolIsolation.TimeoutPerTool != 22*time.Second ||
		cfg.ToolIsolation.MaxMemoryMB != 512 || !cfg.ToolIsolation.Container.Enabled ||
		cfg.ToolIsolation.Container.Runtime != "docker" || cfg.ToolIsolation.Container.Image != "lango-sandbox:test" ||
		cfg.ToolIsolation.Container.NetworkMode != "none" || cfg.ToolIsolation.Container.CPUQuotaUS != 50000 ||
		cfg.ToolIsolation.Container.PoolSize != 2 || cfg.ToolIsolation.Container.PoolIdleTimeout != 7*time.Minute {
		t.Fatalf("p2p sandbox fields not updated: %+v", cfg.ToolIsolation)
	}
	assertBoolPtr(t, "container readonly rootfs", cfg.ToolIsolation.Container.ReadOnlyRootfs, true)
}

func assertPrimarySandboxConfig(t *testing.T, cfg config.SandboxConfig) {
	t.Helper()
	if !cfg.Enabled || !cfg.FailClosed || cfg.Backend != "seatbelt" || cfg.WorkspacePath != "/tmp/workspace" ||
		cfg.NetworkMode != "deny" || !reflect.DeepEqual(cfg.AllowedNetworkIPs, []string{"127.0.0.1", "10.0.0.1"}) ||
		!reflect.DeepEqual(cfg.AllowedWritePaths, []string{"/tmp/a", "/tmp/b"}) ||
		!reflect.DeepEqual(cfg.ExcludedCommands, []string{"git", "docker"}) || cfg.TimeoutPerTool != 33*time.Second ||
		cfg.OS.SeccompProfile != "strict" || cfg.OS.SeatbeltCustomProfile != "/tmp/profile.sb" {
		t.Fatalf("sandbox fields not updated: %+v", cfg)
	}
}

func assertOntologyConfig(t *testing.T, cfg config.OntologyConfig) {
	t.Helper()
	if !cfg.Enabled || !cfg.ACL.Enabled || !reflect.DeepEqual(cfg.ACL.Roles, map[string]string{"alice": "admin", "bob": "read"}) ||
		cfg.ACL.P2PPermission != "write" {
		t.Fatalf("ontology ACL fields not updated: %+v", cfg)
	}
	if !cfg.Governance.Enabled || cfg.Governance.MaxNewPerDay != 10 || cfg.Governance.QuarantinePeriodHrs != 24 ||
		cfg.Governance.ShadowModeDurationHrs != 48 || cfg.Governance.MinUsageForPromotion != 3 ||
		cfg.Governance.SchemaExplosionBudget != 100 || cfg.Governance.AdmissionMode != "observe" ||
		!cfg.Governance.AdmissionModePresent || cfg.Governance.LearningDefaultConfidence != 0.7 ||
		!cfg.Governance.LearningDefaultConfidencePresent || cfg.Governance.LearningDefaultConfidenceBackfillNeeded ||
		cfg.Governance.LibrarianDefaultConfidence != 0.6 || !cfg.Governance.LibrarianDefaultConfidencePresent ||
		cfg.Governance.LibrarianDefaultConfidenceBackfillNeeded {
		t.Fatalf("ontology governance fields not updated: %+v", cfg.Governance)
	}
	if !cfg.Exchange.Enabled || cfg.Exchange.MinTrustForSchema != 0.55 || cfg.Exchange.MinTrustForFacts != 0.75 ||
		cfg.Exchange.AutoImportMode != "governed" || cfg.Exchange.MaxTypesPerImport != 15 {
		t.Fatalf("ontology exchange fields not updated: %+v", cfg.Exchange)
	}
}

func assertAlertingConfig(t *testing.T, cfg config.AlertingConfig) {
	t.Helper()
	if !cfg.Enabled || cfg.PolicyBlockRate != 12 || cfg.RecoveryRetries != 9 || len(cfg.Delivery) != 2 {
		t.Fatalf("alerting fields not updated: %+v", cfg)
	}
	if cfg.Delivery[0].Type != "email" || cfg.Delivery[1].Type != "webhook" ||
		cfg.Delivery[1].WebhookURL != "https://hooks.test" || cfg.Delivery[1].MinSeverity != "warning" {
		t.Fatalf("alert delivery fields not updated: %+v", cfg.Delivery)
	}
}

func assertSecurityKMSConfig(t *testing.T, cfg config.SecurityConfig) {
	t.Helper()
	if cfg.Signer.Provider != "gcp-kms" || cfg.KMS.Region != "us-central1" ||
		cfg.KMS.KeyID != "projects/p/keys/k" || cfg.KMS.Endpoint != "https://kms.test" ||
		!cfg.KMS.FallbackToLocal || cfg.KMS.TimeoutPerOperation != 6*time.Second || cfg.KMS.MaxRetries != 5 ||
		cfg.KMS.Azure.VaultURL != "https://vault.test" || cfg.KMS.Azure.KeyVersion != "v1" ||
		cfg.KMS.PKCS11.ModulePath != "/usr/lib/pkcs11.so" || cfg.KMS.PKCS11.SlotID != 2 ||
		cfg.KMS.PKCS11.Pin != "1234" || cfg.KMS.PKCS11.KeyLabel != "signing-key" {
		t.Fatalf("KMS fields not updated: %+v", cfg)
	}
}

func assertLoggingGatekeeperOutputConfig(t *testing.T, cfg *config.Config) {
	t.Helper()
	if cfg.Logging.Level != "debug" || cfg.Logging.Format != "json" || cfg.Logging.OutputPath != "/tmp/lango.log" {
		t.Fatalf("logging fields not updated: %+v", cfg.Logging)
	}
	assertBoolPtr(t, "gatekeeper enabled", cfg.Gatekeeper.Enabled, true)
	assertBoolPtr(t, "gatekeeper strip thought tags", cfg.Gatekeeper.StripThoughtTags, true)
	assertBoolPtr(t, "gatekeeper strip internal markers", cfg.Gatekeeper.StripInternalMarkers, true)
	assertBoolPtr(t, "gatekeeper strip raw json", cfg.Gatekeeper.StripRawJSON, true)
	if cfg.Gatekeeper.RawJSONThreshold != 777 || !reflect.DeepEqual(cfg.Gatekeeper.CustomPatterns, []string{"secret", "internal"}) {
		t.Fatalf("gatekeeper fields not updated: %+v", cfg.Gatekeeper)
	}
	assertBoolPtr(t, "output manager enabled", cfg.Tools.OutputManager.Enabled, true)
	if cfg.Tools.OutputManager.TokenBudget != 5000 || cfg.Tools.OutputManager.HeadRatio != 0.6 ||
		cfg.Tools.OutputManager.TailRatio != 0.4 {
		t.Fatalf("output manager fields not updated: %+v", cfg.Tools.OutputManager)
	}
}

func assertHooksMemoryLibrarianConfig(t *testing.T, cfg *config.Config) {
	t.Helper()
	if !cfg.Hooks.Enabled || !cfg.Hooks.SecurityFilter || !cfg.Hooks.AccessControl ||
		!cfg.Hooks.EventPublishing || !cfg.Hooks.KnowledgeSave ||
		!reflect.DeepEqual(cfg.Hooks.BlockedCommands, []string{"rm", "curl"}) {
		t.Fatalf("hooks fields not updated: %+v", cfg.Hooks)
	}
	if !cfg.AgentMemory.Enabled {
		t.Fatal("agent memory should be enabled")
	}
	if !cfg.Librarian.Enabled || cfg.Librarian.ObservationThreshold != 2 ||
		cfg.Librarian.InquiryCooldownTurns != 3 || cfg.Librarian.MaxPendingInquiries != 4 ||
		cfg.Librarian.AutoSaveConfidence != types.ConfidenceMedium ||
		cfg.Librarian.Provider != "anthropic" || cfg.Librarian.Model != "claude" {
		t.Fatalf("librarian fields not updated: %+v", cfg.Librarian)
	}
}

func assertEconomyConfig(t *testing.T, cfg config.EconomyConfig) {
	t.Helper()
	if !cfg.Enabled || cfg.Budget.DefaultMax != "20.00" ||
		!reflect.DeepEqual(cfg.Budget.AlertThresholds, []float64{0.5, 0.8, 0.95}) {
		t.Fatalf("economy budget fields not updated: %+v", cfg)
	}
	assertBoolPtr(t, "budget hard limit", cfg.Budget.HardLimit, true)
	if cfg.Risk.EscrowThreshold != "5.00" || cfg.Risk.HighTrustScore != 0.9 || cfg.Risk.MediumTrustScore != 0.6 {
		t.Fatalf("economy risk fields not updated: %+v", cfg.Risk)
	}
	if !cfg.Negotiate.Enabled || cfg.Negotiate.MaxRounds != 6 || cfg.Negotiate.Timeout != 5*time.Minute ||
		!cfg.Negotiate.AutoNegotiate || cfg.Negotiate.MaxDiscount != 0.2 {
		t.Fatalf("economy negotiation fields not updated: %+v", cfg.Negotiate)
	}
	if !cfg.Escrow.Enabled || cfg.Escrow.DefaultTimeout != 24*time.Hour || cfg.Escrow.MaxMilestones != 8 ||
		!cfg.Escrow.AutoRelease || cfg.Escrow.DisputeWindow != time.Hour {
		t.Fatalf("economy escrow fields not updated: %+v", cfg.Escrow)
	}
	if !cfg.Escrow.OnChain.Enabled || cfg.Escrow.OnChain.Mode != "hub" ||
		cfg.Escrow.OnChain.HubAddress != "0xhub" || cfg.Escrow.OnChain.VaultFactoryAddress != "0xfactory" ||
		cfg.Escrow.OnChain.VaultImplementation != "0ximpl" || cfg.Escrow.OnChain.ArbitratorAddress != "0xarb" ||
		cfg.Escrow.OnChain.TokenAddress != "0xtoken" || cfg.Escrow.OnChain.PollInterval != 15*time.Second ||
		cfg.Escrow.OnChain.ConfirmationDepth != 2 {
		t.Fatalf("economy on-chain fields not updated: %+v", cfg.Escrow.OnChain)
	}
	if cfg.Escrow.Settlement.ReceiptTimeout != 3*time.Minute || cfg.Escrow.Settlement.MaxRetries != 4 ||
		!cfg.Pricing.Enabled || cfg.Pricing.TrustDiscount != 0.15 || cfg.Pricing.VolumeDiscount != 0.05 ||
		cfg.Pricing.MinPrice != "0.01" {
		t.Fatalf("economy settlement/pricing fields not updated: %+v", cfg)
	}
}

func assertObservabilityConfig(t *testing.T, cfg config.ObservabilityConfig) {
	t.Helper()
	if !cfg.Enabled || !cfg.Tokens.Enabled || !cfg.Tokens.PersistHistory || cfg.Tokens.RetentionDays != 30 ||
		!cfg.Health.Enabled || cfg.Health.Interval != 45*time.Second || !cfg.Audit.Enabled ||
		cfg.Audit.RetentionDays != 90 || !cfg.Metrics.Enabled || cfg.Metrics.Format != "prometheus" ||
		cfg.TraceStore.MaxAge != 720*time.Hour || cfg.TraceStore.MaxTraces != 10000 ||
		cfg.TraceStore.FailedTraceMultiplier != 3 || cfg.TraceStore.CleanupInterval != time.Hour {
		t.Fatalf("observability fields not updated: %+v", cfg)
	}
}

func assertSmartAccountConfig(t *testing.T, cfg config.SmartAccountConfig) {
	t.Helper()
	if !cfg.Enabled || cfg.FactoryAddress != "0xfactory" || cfg.EntryPointAddress != "0xentry" ||
		cfg.SafeSingletonAddress != "0xsafe" || cfg.Safe7579Address != "0x7579" ||
		cfg.FallbackHandler != "0xfallback" || cfg.BundlerURL != "https://bundler.test" {
		t.Fatalf("smart account fields not updated: %+v", cfg)
	}
	if cfg.Session.MaxDuration != 2*time.Hour || cfg.Session.DefaultGasLimit != 100000 || cfg.Session.MaxActiveKeys != 5 {
		t.Fatalf("smart account session fields not updated: %+v", cfg.Session)
	}
	if !cfg.Paymaster.Enabled || cfg.Paymaster.Provider != "circle" || cfg.Paymaster.Mode != "permit" ||
		cfg.Paymaster.RPCURL != "https://paymaster.test" || cfg.Paymaster.TokenAddress != "0xusdc" ||
		cfg.Paymaster.PaymasterAddress != "0xpaymaster" || cfg.Paymaster.PolicyID != "policy-1" ||
		cfg.Paymaster.FallbackMode != "direct" {
		t.Fatalf("smart account paymaster fields not updated: %+v", cfg.Paymaster)
	}
	if cfg.Modules.SessionValidatorAddress != "0xvalidator" || cfg.Modules.SpendingHookAddress != "0xhook" ||
		cfg.Modules.EscrowExecutorAddress != "0xexecutor" {
		t.Fatalf("smart account modules fields not updated: %+v", cfg.Modules)
	}
}
