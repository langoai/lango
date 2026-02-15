package onboard

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

	// Iterate over fields and update config
	// This is a manual mapping based on keys defined in forms_impl.go
	for _, f := range form.Fields {
		val := f.Value
		// For boolean fields, value might be empty string, check Checked
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
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				s.Current.Agent.Temperature = f
			}
		case "system_prompt_path":
			s.Current.Agent.SystemPromptPath = val
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

		// Security / Session
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
		case "interceptor_approval":
			s.Current.Security.Interceptor.ApprovalRequired = f.Checked
		case "interceptor_timeout":
			if i, err := strconv.Atoi(val); err == nil {
				s.Current.Security.Interceptor.ApprovalTimeoutSec = i
			}
		case "interceptor_notify":
			s.Current.Security.Interceptor.NotifyChannel = val
		case "interceptor_sensitive_tools":
			if val != "" {
				parts := strings.Split(val, ",")
				tools := make([]string, 0, len(parts))
				for _, p := range parts {
					if t := strings.TrimSpace(p); t != "" {
						tools = append(tools, t)
					}
				}
				s.Current.Security.Interceptor.SensitiveTools = tools
			} else {
				s.Current.Security.Interceptor.SensitiveTools = nil
			}

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
		case "emb_provider":
			s.Current.Embedding.Provider = val
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
			if val != "" {
				parts := strings.Split(val, ",")
				cols := make([]string, 0, len(parts))
				for _, p := range parts {
					if c := strings.TrimSpace(p); c != "" {
						cols = append(cols, c)
					}
				}
				s.Current.Embedding.RAG.Collections = cols
			} else {
				s.Current.Embedding.RAG.Collections = nil
			}
		}
	}
}

// UpdateProviderFromForm updates a specific provider config from the form.
func (s *ConfigState) UpdateProviderFromForm(id string, form *FormModel) {
	if form == nil {
		return
	}

	if s.Current.Providers == nil {
		s.Current.Providers = make(map[string]config.ProviderConfig)
	}

	// If id is empty, look for "id" field in form
	if id == "" {
		for _, f := range form.Fields {
			if f.Key == "id" {
				id = f.Value
				break
			}
		}
	}

	if id == "" {
		return // Should not happen if validation works
	}

	// Get or create provider config
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
