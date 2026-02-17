package tuicore

import (
	"strconv"
	"strings"
	"time"

	"github.com/langowarny/lango/internal/config"
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
		case "db_path":
			s.Current.Session.DatabasePath = val
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
		case "knowledge_max_learnings":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Knowledge.MaxLearnings = i
			}
		case "knowledge_max_knowledge":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Knowledge.MaxKnowledge = i
			}
		case "knowledge_max_context":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Knowledge.MaxContextPerLayer = i
			}
		case "knowledge_auto_approve":
			s.Current.Knowledge.AutoApproveSkills = f.Checked
		case "knowledge_max_skills_day":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Knowledge.MaxSkillsPerDay = i
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

		// Embedding & RAG
		case "emb_provider_id":
			if val == "local" {
				s.Current.Embedding.ProviderID = ""
				s.Current.Embedding.Provider = "local"
			} else {
				s.Current.Embedding.ProviderID = val
				s.Current.Embedding.Provider = ""
			}
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
			p.Type = val
		case "apikey":
			p.APIKey = val
		case "baseurl":
			p.BaseURL = val
		}
	}

	s.Current.Providers[id] = p
	s.MarkDirty("providers")
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
