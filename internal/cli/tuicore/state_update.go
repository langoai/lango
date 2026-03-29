package tuicore

import (
	"strconv"
	"strings"
	"time"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/types"
)

// UpdateConfigFromForm updates the config based on the form fields.
func (s *ConfigState) UpdateConfigFromForm(form *FormModel) {
	if form == nil {
		return
	}

	for _, f := range form.Fields {
		val := f.Value
		if f.Type == InputBool {
			val = strconv.FormatBool(f.Checked)
		}

		switch f.Key {
		// Agent
		case "provider":
			s.Current.Agent.Provider = val
		case "model":
			s.Current.Agent.Model = val
		case "maxtokens":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Agent.MaxTokens = i
			}
		case "temp":
			if fv, err := strconv.ParseFloat(val, 64); err == nil {
				s.Current.Agent.Temperature = fv
			}
		case "prompts_dir":
			s.Current.Agent.PromptsDir = val
		case "fallback_provider":
			s.Current.Agent.FallbackProvider = val
		case "fallback_model":
			s.Current.Agent.FallbackModel = val
		case "request_timeout":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.Agent.RequestTimeout = d
			}
		case "tool_timeout":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.Agent.ToolTimeout = d
			}
		case "auto_extend_timeout":
			s.Current.Agent.AutoExtendTimeout = (val == "true")
		case "max_request_timeout":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.Agent.MaxRequestTimeout = d
			}

		// Server
		case "host":
			s.Current.Server.Host = val
		case "port":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Server.Port = i
			}
		case "http":
			s.Current.Server.HTTPEnabled = f.Checked
		case "ws":
			s.Current.Server.WebSocketEnabled = f.Checked

		// Channels - Telegram
		case "telegram_enabled":
			s.Current.Channels.Telegram.Enabled = f.Checked
		case "telegram_token":
			s.Current.Channels.Telegram.BotToken = val

		// Channels - Discord
		case "discord_enabled":
			s.Current.Channels.Discord.Enabled = f.Checked
		case "discord_token":
			s.Current.Channels.Discord.BotToken = val

		// Channels - Slack
		case "slack_enabled":
			s.Current.Channels.Slack.Enabled = f.Checked
		case "slack_token":
			s.Current.Channels.Slack.BotToken = val
		case "slack_app_token":
			s.Current.Channels.Slack.AppToken = val

		// Tools
		case "exec_timeout":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.Tools.Exec.DefaultTimeout = d
			}
		case "exec_bg":
			s.Current.Tools.Exec.AllowBackground = f.Checked
		case "browser_enabled":
			s.Current.Tools.Browser.Enabled = f.Checked
		case "browser_headless":
			s.Current.Tools.Browser.Headless = f.Checked
		case "browser_session_timeout":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.Tools.Browser.SessionTimeout = d
			}
		case "fs_max_read":
			if i, err := strconv.ParseInt(val, 10, 64); err == nil {
				s.Current.Tools.Filesystem.MaxReadSize = i
			}

		// Session
		case "ttl":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.Session.TTL = d
			}
		case "max_history_turns":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Session.MaxHistoryTurns = i
			}

		// Security - Interceptor
		case "interceptor_enabled":
			s.Current.Security.Interceptor.Enabled = f.Checked
		case "interceptor_pii":
			s.Current.Security.Interceptor.RedactPII = f.Checked
		case "interceptor_policy":
			s.Current.Security.Interceptor.ApprovalPolicy = config.ApprovalPolicy(val)
		case "interceptor_exempt_tools":
			s.Current.Security.Interceptor.ExemptTools = splitCSV(val)
		case "interceptor_timeout":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Security.Interceptor.ApprovalTimeoutSec = i
			}
		case "interceptor_notify":
			s.Current.Security.Interceptor.NotifyChannel = val
		case "interceptor_sensitive_tools":
			s.Current.Security.Interceptor.SensitiveTools = splitCSV(val)
		case "interceptor_pii_disabled":
			s.Current.Security.Interceptor.PIIDisabledPatterns = splitCSV(val)
		case "interceptor_pii_custom":
			s.Current.Security.Interceptor.PIICustomPatterns = parseCustomPatterns(val)
		case "presidio_enabled":
			s.Current.Security.Interceptor.Presidio.Enabled = f.Checked
		case "presidio_url":
			s.Current.Security.Interceptor.Presidio.URL = val
		case "presidio_language":
			s.Current.Security.Interceptor.Presidio.Language = val

		// Security - Signer
		case "signer_provider":
			s.Current.Security.Signer.Provider = val
		case "signer_rpc":
			s.Current.Security.Signer.RPCUrl = val
		case "signer_keyid":
			s.Current.Security.Signer.KeyID = val

		// Knowledge
		case "knowledge_enabled":
			s.Current.Knowledge.Enabled = f.Checked
		case "knowledge_max_context":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Knowledge.MaxContextPerLayer = i
			}

		// Skill
		case "skill_enabled":
			s.Current.Skill.Enabled = f.Checked
		case "skill_dir":
			s.Current.Skill.SkillsDir = val
		case "skill_allow_import":
			s.Current.Skill.AllowImport = f.Checked
		case "skill_max_bulk":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Skill.MaxBulkImport = i
			}
		case "skill_import_concurrency":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Skill.ImportConcurrency = i
			}
		case "skill_import_timeout":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.Skill.ImportTimeout = d
			}

			// Observational Memory
		case "om_enabled":
			s.Current.ObservationalMemory.Enabled = f.Checked
		case "om_provider":
			s.Current.ObservationalMemory.Provider = val
		case "om_model":
			s.Current.ObservationalMemory.Model = val
		case "om_msg_threshold":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.ObservationalMemory.MessageTokenThreshold = i
			}
		case "om_obs_threshold":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.ObservationalMemory.ObservationTokenThreshold = i
			}
		case "om_max_budget":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.ObservationalMemory.MaxMessageTokenBudget = i
			}
		case "om_max_reflections":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.ObservationalMemory.MaxReflectionsInContext = i
			}
		case "om_max_observations":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.ObservationalMemory.MaxObservationsInContext = i
			}

		// Embedding & RAG
		case "emb_provider_id":
			s.Current.Embedding.Provider = val
			s.Current.Embedding.ProviderID = "" //nolint:staticcheck // intentional: clear deprecated field
		case "emb_model":
			s.Current.Embedding.Model = val
		case "emb_dimensions":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Embedding.Dimensions = i
			}
		case "emb_local_baseurl":
			s.Current.Embedding.Local.BaseURL = val
		case "emb_rag_enabled":
			s.Current.Embedding.RAG.Enabled = f.Checked
		case "emb_rag_max_results":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Embedding.RAG.MaxResults = i
			}
		case "emb_rag_collections":
			s.Current.Embedding.RAG.Collections = splitCSV(val)

		// Graph Store
		case "graph_enabled":
			s.Current.Graph.Enabled = f.Checked
		case "graph_backend":
			s.Current.Graph.Backend = val
		case "graph_db_path":
			s.Current.Graph.DatabasePath = val
		case "graph_max_depth":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Graph.MaxTraversalDepth = i
			}
		case "graph_max_expand":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Graph.MaxExpansionResults = i
			}

		// Multi-Agent
		case "multi_agent":
			s.Current.Agent.MultiAgent = f.Checked
		case "max_delegation_rounds":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Agent.MaxDelegationRounds = i
			}
		case "max_turns":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Agent.MaxTurns = i
			}
		case "error_correction_enabled":
			s.Current.Agent.ErrorCorrectionEnabled = boolPtr(f.Checked)
		case "agents_dir":
			s.Current.Agent.AgentsDir = val

		// A2A Protocol
		case "a2a_enabled":
			s.Current.A2A.Enabled = f.Checked
		case "a2a_base_url":
			s.Current.A2A.BaseURL = val
		case "a2a_agent_name":
			s.Current.A2A.AgentName = val
		case "a2a_agent_desc":
			s.Current.A2A.AgentDescription = val

		// Cron
		case "cron_enabled":
			s.Current.Cron.Enabled = f.Checked
		case "cron_timezone":
			s.Current.Cron.Timezone = val
		case "cron_max_jobs":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Cron.MaxConcurrentJobs = i
			}
		case "cron_session_mode":
			s.Current.Cron.DefaultSessionMode = val
		case "cron_history_retention":
			s.Current.Cron.HistoryRetention = val
		case "cron_default_deliver":
			s.Current.Cron.DefaultDeliverTo = splitCSV(val)

		// Background
		case "bg_enabled":
			s.Current.Background.Enabled = f.Checked
		case "bg_yield_ms":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Background.YieldMs = i
			}
		case "bg_max_tasks":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Background.MaxConcurrentTasks = i
			}
		case "bg_default_deliver":
			s.Current.Background.DefaultDeliverTo = splitCSV(val)

		// Workflow
		case "wf_enabled":
			s.Current.Workflow.Enabled = f.Checked
		case "wf_max_steps":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Workflow.MaxConcurrentSteps = i
			}
		case "wf_timeout":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.Workflow.DefaultTimeout = d
			}
		case "wf_state_dir":
			s.Current.Workflow.StateDir = val
		case "wf_default_deliver":
			s.Current.Workflow.DefaultDeliverTo = splitCSV(val)

		// MCP
		case "mcp_enabled":
			s.Current.MCP.Enabled = f.Checked
		case "mcp_default_timeout":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.MCP.DefaultTimeout = d
			}
		case "mcp_max_output_tokens":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.MCP.MaxOutputTokens = i
			}
		case "mcp_health_check_interval":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.MCP.HealthCheckInterval = d
			}
		case "mcp_auto_reconnect":
			s.Current.MCP.AutoReconnect = f.Checked
		case "mcp_max_reconnect_attempts":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.MCP.MaxReconnectAttempts = i
			}

		// Payment
		case "payment_enabled":
			s.Current.Payment.Enabled = f.Checked
		case "payment_wallet_provider":
			s.Current.Payment.WalletProvider = val
		case "payment_chain_id":
			if i, err := strconv.ParseInt(val, 10, 64); err == nil {
				s.Current.Payment.Network.ChainID = i
			}
		case "payment_rpc_url":
			s.Current.Payment.Network.RPCURL = val
		case "payment_usdc_contract":
			s.Current.Payment.Network.USDCContract = val
		case "payment_max_per_tx":
			s.Current.Payment.Limits.MaxPerTx = val
		case "payment_max_daily":
			s.Current.Payment.Limits.MaxDaily = val
		case "payment_auto_approve":
			s.Current.Payment.Limits.AutoApproveBelow = val
		case "payment_x402_auto":
			s.Current.Payment.X402.AutoIntercept = f.Checked
		case "payment_x402_max":
			s.Current.Payment.X402.MaxAutoPayAmount = val

		// P2P Network
		case "p2p_enabled":
			s.Current.P2P.Enabled = f.Checked
		case "p2p_listen_addrs":
			s.Current.P2P.ListenAddrs = splitCSV(val)
		case "p2p_bootstrap_peers":
			s.Current.P2P.BootstrapPeers = splitCSV(val)
		case "p2p_enable_relay":
			s.Current.P2P.EnableRelay = f.Checked
		case "p2p_enable_mdns":
			s.Current.P2P.EnableMDNS = f.Checked
		case "p2p_max_peers":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.P2P.MaxPeers = i
			}
		case "p2p_handshake_timeout":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.P2P.HandshakeTimeout = d
			}
		case "p2p_session_token_ttl":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.P2P.SessionTokenTTL = d
			}
		case "p2p_auto_approve":
			s.Current.P2P.AutoApproveKnownPeers = f.Checked
		case "p2p_gossip_interval":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.P2P.GossipInterval = d
			}
		case "p2p_zk_handshake":
			s.Current.P2P.ZKHandshake = f.Checked
		case "p2p_zk_attestation":
			s.Current.P2P.ZKAttestation = f.Checked
		case "p2p_require_signed_challenge":
			s.Current.P2P.RequireSignedChallenge = f.Checked
		case "p2p_min_trust_score":
			if fv, err := strconv.ParseFloat(val, 64); err == nil {
				s.Current.P2P.MinTrustScore = fv
			}

		// P2P ZKP
		case "zkp_proof_cache_dir":
			s.Current.P2P.ZKP.ProofCacheDir = val
		case "zkp_proving_scheme":
			s.Current.P2P.ZKP.ProvingScheme = val
		case "zkp_srs_mode":
			s.Current.P2P.ZKP.SRSMode = val
		case "zkp_srs_path":
			s.Current.P2P.ZKP.SRSPath = val
		case "zkp_max_credential_age":
			s.Current.P2P.ZKP.MaxCredentialAge = val

		// P2P Pricing
		case "pricing_enabled":
			s.Current.P2P.Pricing.Enabled = f.Checked
		case "pricing_per_query":
			s.Current.P2P.Pricing.PerQuery = val
		case "pricing_tool_prices":
			s.Current.P2P.Pricing.ToolPrices = parseCustomPatterns(val)

		// P2P Owner Protection
		case "owner_name":
			s.Current.P2P.OwnerProtection.OwnerName = val
		case "owner_email":
			s.Current.P2P.OwnerProtection.OwnerEmail = val
		case "owner_phone":
			s.Current.P2P.OwnerProtection.OwnerPhone = val
		case "owner_extra_terms":
			s.Current.P2P.OwnerProtection.ExtraTerms = splitCSV(val)
		case "owner_block_conversations":
			s.Current.P2P.OwnerProtection.BlockConversations = boolPtr(f.Checked)

		// P2P Sandbox
		case "sandbox_enabled":
			s.Current.P2P.ToolIsolation.Enabled = f.Checked
		case "sandbox_timeout":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.P2P.ToolIsolation.TimeoutPerTool = d
			}
		case "sandbox_max_memory_mb":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.P2P.ToolIsolation.MaxMemoryMB = i
			}
		case "container_enabled":
			s.Current.P2P.ToolIsolation.Container.Enabled = f.Checked
		case "container_runtime":
			s.Current.P2P.ToolIsolation.Container.Runtime = val
		case "container_image":
			s.Current.P2P.ToolIsolation.Container.Image = val
		case "container_network_mode":
			s.Current.P2P.ToolIsolation.Container.NetworkMode = val
		case "container_readonly_rootfs":
			s.Current.P2P.ToolIsolation.Container.ReadOnlyRootfs = boolPtr(f.Checked)
		case "container_cpu_quota":
			if i, err := strconv.ParseInt(val, 10, 64); err == nil {
				s.Current.P2P.ToolIsolation.Container.CPUQuotaUS = i
			}
		case "container_pool_size":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.P2P.ToolIsolation.Container.PoolSize = i
			}
		case "container_pool_idle_timeout":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.P2P.ToolIsolation.Container.PoolIdleTimeout = d
			}

		// Security DB Encryption
		case "db_encryption_enabled":
			s.Current.Security.DBEncryption.Enabled = f.Checked
		case "db_cipher_page_size":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Security.DBEncryption.CipherPageSize = i
			}

		// Security KMS
		case "kms_backend":
			// Syncs the KMS backend selector with signer provider.
			s.Current.Security.Signer.Provider = val
		case "kms_region":
			s.Current.Security.KMS.Region = val
		case "kms_key_id":
			s.Current.Security.KMS.KeyID = val
		case "kms_endpoint":
			s.Current.Security.KMS.Endpoint = val
		case "kms_fallback_to_local":
			s.Current.Security.KMS.FallbackToLocal = f.Checked
		case "kms_timeout":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.Security.KMS.TimeoutPerOperation = d
			}
		case "kms_max_retries":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Security.KMS.MaxRetries = i
			}
		case "kms_azure_vault_url":
			s.Current.Security.KMS.Azure.VaultURL = val
		case "kms_azure_key_version":
			s.Current.Security.KMS.Azure.KeyVersion = val
		case "kms_pkcs11_module":
			s.Current.Security.KMS.PKCS11.ModulePath = val
		case "kms_pkcs11_slot_id":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Security.KMS.PKCS11.SlotID = i
			}
		case "kms_pkcs11_pin":
			s.Current.Security.KMS.PKCS11.Pin = val
		case "kms_pkcs11_key_label":
			s.Current.Security.KMS.PKCS11.KeyLabel = val

		// Logging
		case "log_level":
			s.Current.Logging.Level = val
		case "log_format":
			s.Current.Logging.Format = val
		case "log_output_path":
			s.Current.Logging.OutputPath = val

		// Gatekeeper
		case "gk_enabled":
			s.Current.Gatekeeper.Enabled = boolPtr(f.Checked)
		case "gk_strip_thought_tags":
			s.Current.Gatekeeper.StripThoughtTags = boolPtr(f.Checked)
		case "gk_strip_internal_markers":
			s.Current.Gatekeeper.StripInternalMarkers = boolPtr(f.Checked)
		case "gk_strip_raw_json":
			s.Current.Gatekeeper.StripRawJSON = boolPtr(f.Checked)
		case "gk_raw_json_threshold":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Gatekeeper.RawJSONThreshold = i
			}
		case "gk_custom_patterns":
			s.Current.Gatekeeper.CustomPatterns = splitCSV(val)

		// Output Manager
		case "om_mgr_enabled":
			s.Current.Tools.OutputManager.Enabled = boolPtr(f.Checked)
		case "om_mgr_token_budget":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Tools.OutputManager.TokenBudget = i
			}
		case "om_mgr_head_ratio":
			if fv, err := strconv.ParseFloat(val, 64); err == nil {
				s.Current.Tools.OutputManager.HeadRatio = fv
			}
		case "om_mgr_tail_ratio":
			if fv, err := strconv.ParseFloat(val, 64); err == nil {
				s.Current.Tools.OutputManager.TailRatio = fv
			}

		// Hooks
		case "hooks_enabled":
			s.Current.Hooks.Enabled = f.Checked
		case "hooks_security_filter":
			s.Current.Hooks.SecurityFilter = f.Checked
		case "hooks_access_control":
			s.Current.Hooks.AccessControl = f.Checked
		case "hooks_event_publishing":
			s.Current.Hooks.EventPublishing = f.Checked
		case "hooks_knowledge_save":
			s.Current.Hooks.KnowledgeSave = f.Checked
		case "hooks_blocked_commands":
			s.Current.Hooks.BlockedCommands = splitCSV(val)

		// Agent Memory
		case "agent_memory_enabled":
			s.Current.AgentMemory.Enabled = f.Checked

		// Librarian
		case "lib_enabled":
			s.Current.Librarian.Enabled = f.Checked
		case "lib_obs_threshold":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Librarian.ObservationThreshold = i
			}
		case "lib_cooldown":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Librarian.InquiryCooldownTurns = i
			}
		case "lib_max_inquiries":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Librarian.MaxPendingInquiries = i
			}
		case "lib_auto_save":
			s.Current.Librarian.AutoSaveConfidence = types.Confidence(val)
		case "lib_provider":
			s.Current.Librarian.Provider = val
		case "lib_model":
			s.Current.Librarian.Model = val

		// Economy
		case "economy_enabled":
			s.Current.Economy.Enabled = f.Checked
		case "economy_budget_default_max":
			s.Current.Economy.Budget.DefaultMax = val
		case "economy_budget_hard_limit":
			s.Current.Economy.Budget.HardLimit = boolPtr(f.Checked)
		case "economy_budget_alert_thresholds":
			s.Current.Economy.Budget.AlertThresholds = parseFloatSlice(val)

		// Economy Risk
		case "economy_risk_escrow_threshold":
			s.Current.Economy.Risk.EscrowThreshold = val
		case "economy_risk_high_trust":
			if fv, err := strconv.ParseFloat(val, 64); err == nil {
				s.Current.Economy.Risk.HighTrustScore = fv
			}
		case "economy_risk_medium_trust":
			if fv, err := strconv.ParseFloat(val, 64); err == nil {
				s.Current.Economy.Risk.MediumTrustScore = fv
			}

		// Economy Negotiation
		case "economy_negotiate_enabled":
			s.Current.Economy.Negotiate.Enabled = f.Checked
		case "economy_negotiate_max_rounds":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Economy.Negotiate.MaxRounds = i
			}
		case "economy_negotiate_timeout":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.Economy.Negotiate.Timeout = d
			}
		case "economy_negotiate_auto":
			s.Current.Economy.Negotiate.AutoNegotiate = f.Checked
		case "economy_negotiate_max_discount":
			if fv, err := strconv.ParseFloat(val, 64); err == nil {
				s.Current.Economy.Negotiate.MaxDiscount = fv
			}

		// Economy Escrow
		case "economy_escrow_enabled":
			s.Current.Economy.Escrow.Enabled = f.Checked
		case "economy_escrow_default_timeout":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.Economy.Escrow.DefaultTimeout = d
			}
		case "economy_escrow_max_milestones":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Economy.Escrow.MaxMilestones = i
			}
		case "economy_escrow_auto_release":
			s.Current.Economy.Escrow.AutoRelease = f.Checked
		case "economy_escrow_dispute_window":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.Economy.Escrow.DisputeWindow = d
			}

		// Economy Escrow On-Chain
		case "economy_escrow_onchain_enabled":
			s.Current.Economy.Escrow.OnChain.Enabled = f.Checked
		case "economy_escrow_onchain_mode":
			s.Current.Economy.Escrow.OnChain.Mode = val
		case "economy_escrow_onchain_hub_address":
			s.Current.Economy.Escrow.OnChain.HubAddress = val
		case "economy_escrow_onchain_vault_factory":
			s.Current.Economy.Escrow.OnChain.VaultFactoryAddress = val
		case "economy_escrow_onchain_vault_impl":
			s.Current.Economy.Escrow.OnChain.VaultImplementation = val
		case "economy_escrow_onchain_arbitrator":
			s.Current.Economy.Escrow.OnChain.ArbitratorAddress = val
		case "economy_escrow_onchain_token":
			s.Current.Economy.Escrow.OnChain.TokenAddress = val
		case "economy_escrow_onchain_poll_interval":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.Economy.Escrow.OnChain.PollInterval = d
			}
		case "economy_escrow_onchain_confirmation_depth":
			if u, err := strconv.ParseUint(val, 10, 64); err == nil {
				s.Current.Economy.Escrow.OnChain.ConfirmationDepth = u
			}

		// Economy Escrow Settlement
		case "economy_escrow_settlement_receipt_timeout":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.Economy.Escrow.Settlement.ReceiptTimeout = d
			}
		case "economy_escrow_settlement_max_retries":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Economy.Escrow.Settlement.MaxRetries = i
			}

		// Economy Pricing
		case "economy_pricing_enabled":
			s.Current.Economy.Pricing.Enabled = f.Checked
		case "economy_pricing_trust_discount":
			if fv, err := strconv.ParseFloat(val, 64); err == nil {
				s.Current.Economy.Pricing.TrustDiscount = fv
			}
		case "economy_pricing_volume_discount":
			if fv, err := strconv.ParseFloat(val, 64); err == nil {
				s.Current.Economy.Pricing.VolumeDiscount = fv
			}
		case "economy_pricing_min_price":
			s.Current.Economy.Pricing.MinPrice = val

		// Observability
		case "obs_enabled":
			s.Current.Observability.Enabled = f.Checked
		case "obs_tokens_enabled":
			s.Current.Observability.Tokens.Enabled = f.Checked
		case "obs_tokens_persist":
			s.Current.Observability.Tokens.PersistHistory = f.Checked
		case "obs_tokens_retention":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Observability.Tokens.RetentionDays = i
			}
		case "obs_health_enabled":
			s.Current.Observability.Health.Enabled = f.Checked
		case "obs_health_interval":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.Observability.Health.Interval = d
			}
		case "obs_audit_enabled":
			s.Current.Observability.Audit.Enabled = f.Checked
		case "obs_audit_retention":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Observability.Audit.RetentionDays = i
			}
		case "obs_metrics_enabled":
			s.Current.Observability.Metrics.Enabled = f.Checked
		case "obs_metrics_format":
			s.Current.Observability.Metrics.Format = val

		// Smart Account
		case "sa_enabled":
			s.Current.SmartAccount.Enabled = f.Checked
		case "sa_factory_address":
			s.Current.SmartAccount.FactoryAddress = val
		case "sa_entrypoint_address":
			s.Current.SmartAccount.EntryPointAddress = val
		case "sa_singleton_address":
			s.Current.SmartAccount.SafeSingletonAddress = val
		case "sa_safe7579_address":
			s.Current.SmartAccount.Safe7579Address = val
		case "sa_fallback_handler":
			s.Current.SmartAccount.FallbackHandler = val
		case "sa_bundler_url":
			s.Current.SmartAccount.BundlerURL = val

		// Smart Account Session
		case "sa_session_max_duration":
			if d, err := time.ParseDuration(val); err == nil {
				s.Current.SmartAccount.Session.MaxDuration = d
			}
		case "sa_session_default_gas_limit":
			if i, err := strconv.ParseUint(val, 10, 64); err == nil {
				s.Current.SmartAccount.Session.DefaultGasLimit = i
			}
		case "sa_session_max_active_keys":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.SmartAccount.Session.MaxActiveKeys = i
			}

		// Smart Account Paymaster
		case "sa_paymaster_enabled":
			s.Current.SmartAccount.Paymaster.Enabled = f.Checked
		case "sa_paymaster_provider":
			s.Current.SmartAccount.Paymaster.Provider = val
		case "sa_paymaster_mode":
			s.Current.SmartAccount.Paymaster.Mode = val
		case "sa_paymaster_rpc_url":
			s.Current.SmartAccount.Paymaster.RPCURL = val
		case "sa_paymaster_token_address":
			s.Current.SmartAccount.Paymaster.TokenAddress = val
		case "sa_paymaster_address":
			s.Current.SmartAccount.Paymaster.PaymasterAddress = val
		case "sa_paymaster_policy_id":
			s.Current.SmartAccount.Paymaster.PolicyID = val
		case "sa_paymaster_fallback_mode":
			s.Current.SmartAccount.Paymaster.FallbackMode = val

		// Smart Account Modules
		case "sa_modules_session_validator":
			s.Current.SmartAccount.Modules.SessionValidatorAddress = val
		case "sa_modules_spending_hook":
			s.Current.SmartAccount.Modules.SpendingHookAddress = val
		case "sa_modules_escrow_executor":
			s.Current.SmartAccount.Modules.EscrowExecutorAddress = val

		// Context Profile
		case "ctx_profile":
			s.Current.ContextProfile = config.ContextProfileName(val)

		// Retrieval
		case "retrieval_enabled":
			s.Current.Retrieval.Enabled = f.Checked
		case "retrieval_feedback":
			s.Current.Retrieval.Feedback = f.Checked

		// Auto-Adjust
		case "aa_enabled":
			s.Current.Retrieval.AutoAdjust.Enabled = f.Checked
		case "aa_mode":
			s.Current.Retrieval.AutoAdjust.Mode = val
		case "aa_boost_delta":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				s.Current.Retrieval.AutoAdjust.BoostDelta = v
			}
		case "aa_decay_delta":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				s.Current.Retrieval.AutoAdjust.DecayDelta = v
			}
		case "aa_decay_interval":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Retrieval.AutoAdjust.DecayInterval = i
			}
		case "aa_min_score":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				s.Current.Retrieval.AutoAdjust.MinScore = v
			}
		case "aa_max_score":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				s.Current.Retrieval.AutoAdjust.MaxScore = v
			}
		case "aa_warmup_turns":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Retrieval.AutoAdjust.WarmupTurns = i
			}

		// Context Budget
		case "ctx_model_window":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Context.ModelWindow = i
			}
		case "ctx_response_reserve":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Context.ResponseReserve = i
			}
		case "ctx_alloc_knowledge":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				s.Current.Context.Allocation.Knowledge = v
			}
		case "ctx_alloc_rag":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				s.Current.Context.Allocation.RAG = v
			}
		case "ctx_alloc_memory":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				s.Current.Context.Allocation.Memory = v
			}
		case "ctx_alloc_run_summary":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				s.Current.Context.Allocation.RunSummary = v
			}
		case "ctx_alloc_headroom":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				s.Current.Context.Allocation.Headroom = v
			}
		}
	}
}

// UpdateAuthProviderFromForm updates a specific OIDC provider config from the form.
func (s *ConfigState) UpdateAuthProviderFromForm(id string, form *FormModel) {
	if form == nil {
		return
	}

	if s.Current.Auth.Providers == nil {
		s.Current.Auth.Providers = make(map[string]config.OIDCProviderConfig)
	}

	if id == "" {
		for _, f := range form.Fields {
			if f.Key == "oidc_id" {
				id = f.Value
				break
			}
		}
	}

	if id == "" {
		return
	}

	p, ok := s.Current.Auth.Providers[id]
	if !ok {
		p = config.OIDCProviderConfig{}
	}

	for _, f := range form.Fields {
		val := f.Value
		switch f.Key {
		case "oidc_issuer":
			p.IssuerURL = val
		case "oidc_client_id":
			p.ClientID = val
		case "oidc_client_secret":
			p.ClientSecret = val
		case "oidc_redirect":
			p.RedirectURL = val
		case "oidc_scopes":
			p.Scopes = splitCSV(val)
		}
	}

	s.Current.Auth.Providers[id] = p
	s.MarkDirty("auth")
}

// UpdateProviderFromForm updates a specific provider config from the form.
func (s *ConfigState) UpdateProviderFromForm(id string, form *FormModel) {
	if form == nil {
		return
	}

	if s.Current.Providers == nil {
		s.Current.Providers = make(map[string]config.ProviderConfig)
	}

	if id == "" {
		for _, f := range form.Fields {
			if f.Key == "id" {
				id = f.Value
				break
			}
		}
	}

	if id == "" {
		return
	}

	p, ok := s.Current.Providers[id]
	if !ok {
		p = config.ProviderConfig{}
	}

	for _, f := range form.Fields {
		val := f.Value
		switch f.Key {
		case "type":
			p.Type = types.ProviderType(val)
		case "apikey":
			p.APIKey = val
		case "baseurl":
			p.BaseURL = val
		}
	}

	s.Current.Providers[id] = p
	s.MarkDirty("providers")
}

// UpdateMCPServerFromForm updates a specific MCP server config from the form.
func (s *ConfigState) UpdateMCPServerFromForm(name string, form *FormModel) {
	if form == nil {
		return
	}

	if s.Current.MCP.Servers == nil {
		s.Current.MCP.Servers = make(map[string]config.MCPServerConfig)
	}

	if name == "" {
		for _, f := range form.Fields {
			if f.Key == "mcp_srv_name" {
				name = f.Value
				break
			}
		}
	}

	if name == "" {
		return
	}

	srv, ok := s.Current.MCP.Servers[name]
	if !ok {
		srv = config.MCPServerConfig{}
	}

	for _, f := range form.Fields {
		val := f.Value
		switch f.Key {
		case "mcp_srv_transport":
			srv.Transport = val
		case "mcp_srv_command":
			srv.Command = val
		case "mcp_srv_args":
			srv.Args = splitCSV(val)
		case "mcp_srv_url":
			srv.URL = val
		case "mcp_srv_env":
			srv.Env = parseKeyValuePairs(val)
		case "mcp_srv_headers":
			srv.Headers = parseKeyValuePairs(val)
		case "mcp_srv_enabled":
			srv.Enabled = boolPtr(f.Checked)
		case "mcp_srv_timeout":
			if val != "" {
				if d, err := time.ParseDuration(val); err == nil {
					srv.Timeout = d
				}
			} else {
				srv.Timeout = 0
			}
		case "mcp_srv_safety":
			srv.SafetyLevel = val
		}
	}

	s.Current.MCP.Servers[name] = srv
	s.MarkDirty("mcp")
}

// boolPtr returns a pointer to the given bool value.
func boolPtr(b bool) *bool { return &b }

// parseCustomPatterns parses a comma-separated "name:regex" string into a map.
func parseCustomPatterns(val string) map[string]string {
	if val == "" {
		return nil
	}
	result := make(map[string]string)
	parts := strings.Split(val, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		idx := strings.Index(p, ":")
		if idx <= 0 || idx >= len(p)-1 {
			continue
		}
		name := strings.TrimSpace(p[:idx])
		regex := strings.TrimSpace(p[idx+1:])
		if name != "" && regex != "" {
			result[name] = regex
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// splitCSV splits a comma-separated string, trims whitespace, and drops empty parts.
func splitCSV(val string) []string {
	if val == "" {
		return nil
	}
	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// parseFloatSlice parses a comma-separated string of floats into a float64 slice.
func parseFloatSlice(val string) []float64 {
	if val == "" {
		return nil
	}
	parts := strings.Split(val, ",")
	out := make([]float64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if f, err := strconv.ParseFloat(p, 64); err == nil {
			out = append(out, f)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// parseKeyValuePairs parses a comma-separated "KEY=VAL,KEY=VAL" string into a map.
func parseKeyValuePairs(val string) map[string]string {
	if val == "" {
		return nil
	}
	result := make(map[string]string)
	parts := strings.Split(val, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		idx := strings.Index(p, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(p[:idx])
		value := strings.TrimSpace(p[idx+1:])
		if key != "" {
			result[key] = value
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
