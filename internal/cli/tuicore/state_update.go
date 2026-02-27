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
			s.Current.Embedding.ProviderID = "" // clear deprecated field
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

		// Security Keyring
		case "keyring_enabled":
			s.Current.Security.Keyring.Enabled = f.Checked

		// Security DB Encryption
		case "db_encryption_enabled":
			s.Current.Security.DBEncryption.Enabled = f.Checked
		case "db_cipher_page_size":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Security.DBEncryption.CipherPageSize = i
			}

		// Security KMS
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
