package settings

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/types"
)

// buildProviderOptions builds provider options from registered providers.
func buildProviderOptions(cfg *config.Config) []string {
	opts := make([]string, 0, len(cfg.Providers))
	for id := range cfg.Providers {
		opts = append(opts, id)
	}
	sort.Strings(opts)
	if len(opts) == 0 {
		opts = []string{"anthropic", "openai", "gemini", "ollama"}
	}
	return opts
}

// NewAgentForm creates the Agent configuration form.
func NewAgentForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Agent Configuration")

	providerOpts := buildProviderOptions(cfg)
	form.AddField(&tuicore.Field{
		Key: "provider", Label: "Provider", Type: tuicore.InputSelect,
		Value:   cfg.Agent.Provider,
		Options: providerOpts,
	})

	form.AddField(&tuicore.Field{
		Key: "model", Label: "Model ID", Type: tuicore.InputText,
		Value:       cfg.Agent.Model,
		Placeholder: "e.g. claude-3-5-sonnet-20240620",
	})

	form.AddField(&tuicore.Field{
		Key: "maxtokens", Label: "Max Tokens", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Agent.MaxTokens),
		Validate: func(s string) error {
			if _, err := strconv.Atoi(s); err != nil {
				return fmt.Errorf("must be integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "temp", Label: "Temperature", Type: tuicore.InputText,
		Value: fmt.Sprintf("%.1f", cfg.Agent.Temperature),
	})

	form.AddField(&tuicore.Field{
		Key: "prompts_dir", Label: "Prompts Directory", Type: tuicore.InputText,
		Value:       cfg.Agent.PromptsDir,
		Placeholder: "~/.lango/prompts (supports agents/<name>/ for per-agent overrides)",
	})

	fallbackOpts := append([]string{""}, providerOpts...)
	form.AddField(&tuicore.Field{
		Key: "fallback_provider", Label: "Fallback Provider", Type: tuicore.InputSelect,
		Value:   cfg.Agent.FallbackProvider,
		Options: fallbackOpts,
	})

	form.AddField(&tuicore.Field{
		Key: "fallback_model", Label: "Fallback Model", Type: tuicore.InputText,
		Value:       cfg.Agent.FallbackModel,
		Placeholder: "e.g. gpt-4o",
	})

	form.AddField(&tuicore.Field{
		Key: "request_timeout", Label: "Request Timeout", Type: tuicore.InputText,
		Value:       cfg.Agent.RequestTimeout.String(),
		Placeholder: "5m (e.g. 30s, 2m, 5m)",
	})

	form.AddField(&tuicore.Field{
		Key: "tool_timeout", Label: "Tool Timeout", Type: tuicore.InputText,
		Value:       cfg.Agent.ToolTimeout.String(),
		Placeholder: "2m (e.g. 30s, 1m, 2m)",
	})

	return &form
}

// NewServerForm creates the Server configuration form.
func NewServerForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Server Configuration")

	form.AddField(&tuicore.Field{
		Key: "host", Label: "Host", Type: tuicore.InputText,
		Value: cfg.Server.Host,
	})

	form.AddField(&tuicore.Field{
		Key: "port", Label: "Port", Type: tuicore.InputInt,
		Value:    strconv.Itoa(cfg.Server.Port),
		Validate: validatePort,
	})

	form.AddField(&tuicore.Field{
		Key: "http", Label: "Generic HTTP", Type: tuicore.InputBool,
		Checked: cfg.Server.HTTPEnabled,
	})

	form.AddField(&tuicore.Field{
		Key: "ws", Label: "WebSockets", Type: tuicore.InputBool,
		Checked: cfg.Server.WebSocketEnabled,
	})

	return &form
}

// NewChannelsForm creates the Channels configuration form.
func NewChannelsForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Channels Configuration")

	form.AddField(&tuicore.Field{
		Key: "telegram_enabled", Label: "Telegram", Type: tuicore.InputBool,
		Checked: cfg.Channels.Telegram.Enabled,
	})
	form.AddField(&tuicore.Field{
		Key: "telegram_token", Label: "  Bot Token", Type: tuicore.InputPassword,
		Value:       cfg.Channels.Telegram.BotToken,
		Placeholder: "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
	})

	form.AddField(&tuicore.Field{
		Key: "discord_enabled", Label: "Discord", Type: tuicore.InputBool,
		Checked: cfg.Channels.Discord.Enabled,
	})
	form.AddField(&tuicore.Field{
		Key: "discord_token", Label: "  Bot Token", Type: tuicore.InputPassword,
		Value: cfg.Channels.Discord.BotToken,
	})

	form.AddField(&tuicore.Field{
		Key: "slack_enabled", Label: "Slack", Type: tuicore.InputBool,
		Checked: cfg.Channels.Slack.Enabled,
	})
	form.AddField(&tuicore.Field{
		Key: "slack_token", Label: "  Bot Token", Type: tuicore.InputPassword,
		Value:       cfg.Channels.Slack.BotToken,
		Placeholder: "xoxb-...",
	})
	form.AddField(&tuicore.Field{
		Key: "slack_app_token", Label: "  App Token", Type: tuicore.InputPassword,
		Value:       cfg.Channels.Slack.AppToken,
		Placeholder: "xapp-...",
	})

	return &form
}

// NewToolsForm creates the Tools configuration form.
func NewToolsForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Tools Configuration")

	form.AddField(&tuicore.Field{
		Key: "exec_timeout", Label: "Exec Timeout", Type: tuicore.InputText,
		Value:       cfg.Tools.Exec.DefaultTimeout.String(),
		Placeholder: "30s",
	})
	form.AddField(&tuicore.Field{
		Key: "exec_bg", Label: "Allow Background", Type: tuicore.InputBool,
		Checked: cfg.Tools.Exec.AllowBackground,
	})

	form.AddField(&tuicore.Field{
		Key: "browser_enabled", Label: "Browser Enabled", Type: tuicore.InputBool,
		Checked: cfg.Tools.Browser.Enabled,
	})
	form.AddField(&tuicore.Field{
		Key: "browser_headless", Label: "Browser Headless", Type: tuicore.InputBool,
		Checked: cfg.Tools.Browser.Headless,
	})
	form.AddField(&tuicore.Field{
		Key: "browser_session_timeout", Label: "Browser Session Timeout", Type: tuicore.InputText,
		Value:       cfg.Tools.Browser.SessionTimeout.String(),
		Placeholder: "5m",
	})

	form.AddField(&tuicore.Field{
		Key: "fs_max_read", Label: "Max Read Size", Type: tuicore.InputInt,
		Value: strconv.FormatInt(cfg.Tools.Filesystem.MaxReadSize, 10),
	})

	return &form
}

// NewSessionForm creates the Session configuration form.
func NewSessionForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Session Configuration")

	form.AddField(&tuicore.Field{
		Key: "ttl", Label: "Session TTL", Type: tuicore.InputText,
		Value: cfg.Session.TTL.String(),
	})

	form.AddField(&tuicore.Field{
		Key: "max_history_turns", Label: "Max History Turns", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Session.MaxHistoryTurns),
	})

	return &form
}

// NewSecurityForm creates the Security configuration form.
func NewSecurityForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Security Configuration")

	form.AddField(&tuicore.Field{
		Key: "interceptor_enabled", Label: "Privacy Interceptor", Type: tuicore.InputBool,
		Checked: cfg.Security.Interceptor.Enabled,
	})
	form.AddField(&tuicore.Field{
		Key: "interceptor_pii", Label: "  Redact PII", Type: tuicore.InputBool,
		Checked: cfg.Security.Interceptor.RedactPII,
	})
	policyVal := string(cfg.Security.Interceptor.ApprovalPolicy)
	if policyVal == "" {
		policyVal = "dangerous"
	}
	form.AddField(&tuicore.Field{
		Key: "interceptor_policy", Label: "  Approval Policy", Type: tuicore.InputSelect,
		Value:   policyVal,
		Options: []string{"dangerous", "all", "configured", "none"},
	})

	form.AddField(&tuicore.Field{
		Key: "signer_provider", Label: "Signer Provider", Type: tuicore.InputSelect,
		Value:   cfg.Security.Signer.Provider,
		Options: []string{"local", "rpc", "enclave", "aws-kms", "gcp-kms", "azure-kv", "pkcs11"},
	})
	form.AddField(&tuicore.Field{
		Key: "signer_rpc", Label: "  RPC URL", Type: tuicore.InputText,
		Value:       cfg.Security.Signer.RPCUrl,
		Placeholder: "http://localhost:8080",
	})
	form.AddField(&tuicore.Field{
		Key: "signer_keyid", Label: "  Key ID", Type: tuicore.InputText,
		Value:       cfg.Security.Signer.KeyID,
		Placeholder: "key-123",
	})

	form.AddField(&tuicore.Field{
		Key: "interceptor_timeout", Label: "  Approval Timeout (s)", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Security.Interceptor.ApprovalTimeoutSec),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "interceptor_notify", Label: "  Notify Channel", Type: tuicore.InputSelect,
		Value:   cfg.Security.Interceptor.NotifyChannel,
		Options: []string{"", string(types.ChannelTelegram), string(types.ChannelDiscord), string(types.ChannelSlack)},
	})

	form.AddField(&tuicore.Field{
		Key: "interceptor_sensitive_tools", Label: "  Sensitive Tools", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Security.Interceptor.SensitiveTools, ","),
		Placeholder: "exec,browser (comma-separated)",
	})

	form.AddField(&tuicore.Field{
		Key: "interceptor_exempt_tools", Label: "  Exempt Tools", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Security.Interceptor.ExemptTools, ","),
		Placeholder: "filesystem (comma-separated)",
	})

	// PII Pattern Management
	form.AddField(&tuicore.Field{
		Key: "interceptor_pii_disabled", Label: "  Disabled PII Patterns", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Security.Interceptor.PIIDisabledPatterns, ","),
		Placeholder: "kr_bank_account,passport,ipv4 (comma-separated)",
	})
	form.AddField(&tuicore.Field{
		Key: "interceptor_pii_custom", Label: "  Custom PII Patterns", Type: tuicore.InputText,
		Value:       formatCustomPatterns(cfg.Security.Interceptor.PIICustomPatterns),
		Placeholder: `my_id:\bID-\d{6}\b (name:regex, comma-sep)`,
	})

	// Presidio Integration
	form.AddField(&tuicore.Field{
		Key: "presidio_enabled", Label: "  Presidio (Docker)", Type: tuicore.InputBool,
		Checked: cfg.Security.Interceptor.Presidio.Enabled,
	})
	form.AddField(&tuicore.Field{
		Key: "presidio_url", Label: "  Presidio URL", Type: tuicore.InputText,
		Value:       cfg.Security.Interceptor.Presidio.URL,
		Placeholder: "http://localhost:5002",
	})
	form.AddField(&tuicore.Field{
		Key: "presidio_language", Label: "  Presidio Language", Type: tuicore.InputText,
		Value:       cfg.Security.Interceptor.Presidio.Language,
		Placeholder: "en",
	})

	return &form
}

// formatCustomPatterns formats a map of custom patterns into a comma-separated
// "name:regex" string for display in the TUI.
func formatCustomPatterns(patterns map[string]string) string {
	if len(patterns) == 0 {
		return ""
	}
	parts := make([]string, 0, len(patterns))
	for name, regex := range patterns {
		parts = append(parts, name+":"+regex)
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

// ParseCustomPatterns parses a comma-separated "name:regex" string into a map.
func ParseCustomPatterns(val string) map[string]string {
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

func validatePort(s string) error {
	p, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("invalid number")
	}
	if p < 1 || p > 65535 {
		return fmt.Errorf("port out of range")
	}
	return nil
}

// NewKnowledgeForm creates the Knowledge configuration form.
func NewKnowledgeForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Knowledge Configuration")

	form.AddField(&tuicore.Field{
		Key: "knowledge_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked: cfg.Knowledge.Enabled,
	})

	form.AddField(&tuicore.Field{
		Key: "knowledge_max_context", Label: "Max Context/Layer", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Knowledge.MaxContextPerLayer),
	})

	return &form
}

// NewSkillForm creates the Skill configuration form.
func NewSkillForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Skill Configuration")

	form.AddField(&tuicore.Field{
		Key: "skill_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked: cfg.Skill.Enabled,
	})

	form.AddField(&tuicore.Field{
		Key: "skill_dir", Label: "Skills Directory", Type: tuicore.InputText,
		Value:       cfg.Skill.SkillsDir,
		Placeholder: "~/.lango/skills",
	})

	form.AddField(&tuicore.Field{
		Key: "skill_allow_import", Label: "Allow Import", Type: tuicore.InputBool,
		Checked: cfg.Skill.AllowImport,
	})

	form.AddField(&tuicore.Field{
		Key: "skill_max_bulk", Label: "Max Bulk Import", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Skill.MaxBulkImport),
		Placeholder: "50",
	})

	form.AddField(&tuicore.Field{
		Key: "skill_import_concurrency", Label: "Import Concurrency", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Skill.ImportConcurrency),
		Placeholder: "5",
	})

	form.AddField(&tuicore.Field{
		Key: "skill_import_timeout", Label: "Import Timeout", Type: tuicore.InputText,
		Value:       cfg.Skill.ImportTimeout.String(),
		Placeholder: "2m (e.g. 30s, 1m, 5m)",
	})

	return &form
}

// NewObservationalMemoryForm creates the Observational Memory configuration form.
func NewObservationalMemoryForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Observational Memory")

	form.AddField(&tuicore.Field{
		Key: "om_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked: cfg.ObservationalMemory.Enabled,
	})

	omProviderOpts := append([]string{""}, buildProviderOptions(cfg)...)
	form.AddField(&tuicore.Field{
		Key: "om_provider", Label: "Provider", Type: tuicore.InputSelect,
		Value:   cfg.ObservationalMemory.Provider,
		Options: omProviderOpts,
	})

	form.AddField(&tuicore.Field{
		Key: "om_model", Label: "Model", Type: tuicore.InputText,
		Value:       cfg.ObservationalMemory.Model,
		Placeholder: "leave empty for agent default",
	})

	form.AddField(&tuicore.Field{
		Key: "om_msg_threshold", Label: "Message Token Threshold",
		Type:  tuicore.InputInt,
		Value: strconv.Itoa(cfg.ObservationalMemory.MessageTokenThreshold),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "om_obs_threshold", Label: "Observation Token Threshold",
		Type:  tuicore.InputInt,
		Value: strconv.Itoa(cfg.ObservationalMemory.ObservationTokenThreshold),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "om_max_budget", Label: "Max Message Token Budget",
		Type:  tuicore.InputInt,
		Value: strconv.Itoa(cfg.ObservationalMemory.MaxMessageTokenBudget),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "om_max_reflections", Label: "Max Reflections in Context",
		Type:  tuicore.InputInt,
		Value: strconv.Itoa(cfg.ObservationalMemory.MaxReflectionsInContext),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer (0 = unlimited)")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "om_max_observations", Label: "Max Observations in Context",
		Type:  tuicore.InputInt,
		Value: strconv.Itoa(cfg.ObservationalMemory.MaxObservationsInContext),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer (0 = unlimited)")
			}
			return nil
		},
	})

	return &form
}

// NewEmbeddingForm creates the Embedding & RAG configuration form.
func NewEmbeddingForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Embedding & RAG Configuration")

	providerOpts := []string{"local"}
	for id := range cfg.Providers {
		providerOpts = append(providerOpts, id)
	}
	sort.Strings(providerOpts)

	currentVal := cfg.Embedding.ProviderID
	if currentVal == "" && cfg.Embedding.Provider == "local" {
		currentVal = "local"
	}

	form.AddField(&tuicore.Field{
		Key: "emb_provider_id", Label: "Provider", Type: tuicore.InputSelect,
		Value:   currentVal,
		Options: providerOpts,
	})

	form.AddField(&tuicore.Field{
		Key: "emb_model", Label: "Model", Type: tuicore.InputText,
		Value:       cfg.Embedding.Model,
		Placeholder: "e.g. text-embedding-3-small",
	})

	form.AddField(&tuicore.Field{
		Key: "emb_dimensions", Label: "Dimensions", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Embedding.Dimensions),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "emb_local_baseurl", Label: "Local Base URL", Type: tuicore.InputText,
		Value:       cfg.Embedding.Local.BaseURL,
		Placeholder: "http://localhost:11434/v1",
	})

	form.AddField(&tuicore.Field{
		Key: "emb_rag_enabled", Label: "RAG Enabled", Type: tuicore.InputBool,
		Checked: cfg.Embedding.RAG.Enabled,
	})

	form.AddField(&tuicore.Field{
		Key: "emb_rag_max_results", Label: "RAG Max Results", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Embedding.RAG.MaxResults),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "emb_rag_collections", Label: "RAG Collections", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Embedding.RAG.Collections, ","),
		Placeholder: "collection1,collection2 (comma-separated)",
	})

	return &form
}

// NewOIDCProviderForm creates the OIDC Provider configuration form.
func NewOIDCProviderForm(id string, cfg config.OIDCProviderConfig) *tuicore.FormModel {
	title := "Edit OIDC Provider: " + id
	if id == "" {
		title = "Add New OIDC Provider"
	}
	form := tuicore.NewFormModel(title)

	if id == "" {
		form.AddField(&tuicore.Field{
			Key: "oidc_id", Label: "Provider Name", Type: tuicore.InputText,
			Placeholder: "e.g. google, github",
		})
	}

	form.AddField(&tuicore.Field{
		Key: "oidc_issuer", Label: "Issuer URL", Type: tuicore.InputText,
		Value:       cfg.IssuerURL,
		Placeholder: "https://accounts.google.com",
	})

	form.AddField(&tuicore.Field{
		Key: "oidc_client_id", Label: "Client ID", Type: tuicore.InputPassword,
		Value:       cfg.ClientID,
		Placeholder: "${ENV_VAR} or value",
	})

	form.AddField(&tuicore.Field{
		Key: "oidc_client_secret", Label: "Client Secret", Type: tuicore.InputPassword,
		Value:       cfg.ClientSecret,
		Placeholder: "${ENV_VAR} or value",
	})

	form.AddField(&tuicore.Field{
		Key: "oidc_redirect", Label: "Redirect URL", Type: tuicore.InputText,
		Value:       cfg.RedirectURL,
		Placeholder: "http://localhost:18789/auth/callback/<name>",
	})

	form.AddField(&tuicore.Field{
		Key: "oidc_scopes", Label: "Scopes", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Scopes, ","),
		Placeholder: "openid,email,profile",
	})

	return &form
}

// NewGraphForm creates the Graph Store configuration form.
func NewGraphForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Graph Store Configuration")

	form.AddField(&tuicore.Field{
		Key: "graph_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked: cfg.Graph.Enabled,
	})

	form.AddField(&tuicore.Field{
		Key: "graph_backend", Label: "Backend", Type: tuicore.InputSelect,
		Value:   cfg.Graph.Backend,
		Options: []string{"bolt"},
	})

	form.AddField(&tuicore.Field{
		Key: "graph_db_path", Label: "Database Path", Type: tuicore.InputText,
		Value:       cfg.Graph.DatabasePath,
		Placeholder: "~/.lango/graph.db",
	})

	form.AddField(&tuicore.Field{
		Key: "graph_max_depth", Label: "Max Traversal Depth", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Graph.MaxTraversalDepth),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "graph_max_expand", Label: "Max Expansion Results", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Graph.MaxExpansionResults),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	return &form
}

// NewMultiAgentForm creates the Multi-Agent configuration form.
func NewMultiAgentForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Multi-Agent Configuration")

	form.AddField(&tuicore.Field{
		Key: "multi_agent", Label: "Enable Multi-Agent Orchestration", Type: tuicore.InputBool,
		Checked: cfg.Agent.MultiAgent,
	})

	return &form
}

// NewA2AForm creates the A2A Protocol configuration form.
func NewA2AForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("A2A Protocol Configuration")

	form.AddField(&tuicore.Field{
		Key: "a2a_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked: cfg.A2A.Enabled,
	})

	form.AddField(&tuicore.Field{
		Key: "a2a_base_url", Label: "Base URL", Type: tuicore.InputText,
		Value:       cfg.A2A.BaseURL,
		Placeholder: "https://your-agent.example.com",
	})

	form.AddField(&tuicore.Field{
		Key: "a2a_agent_name", Label: "Agent Name", Type: tuicore.InputText,
		Value:       cfg.A2A.AgentName,
		Placeholder: "my-lango-agent",
	})

	form.AddField(&tuicore.Field{
		Key: "a2a_agent_desc", Label: "Agent Description", Type: tuicore.InputText,
		Value:       cfg.A2A.AgentDescription,
		Placeholder: "A helpful AI assistant",
	})

	return &form
}

// NewPaymentForm creates the Payment configuration form.
func NewPaymentForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Payment Configuration")

	form.AddField(&tuicore.Field{
		Key: "payment_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked: cfg.Payment.Enabled,
	})

	form.AddField(&tuicore.Field{
		Key: "payment_wallet_provider", Label: "Wallet Provider", Type: tuicore.InputSelect,
		Value:   cfg.Payment.WalletProvider,
		Options: []string{"local", "rpc", "composite"},
	})

	form.AddField(&tuicore.Field{
		Key: "payment_chain_id", Label: "Chain ID", Type: tuicore.InputInt,
		Value: strconv.FormatInt(cfg.Payment.Network.ChainID, 10),
		Validate: func(s string) error {
			if _, err := strconv.ParseInt(s, 10, 64); err != nil {
				return fmt.Errorf("must be an integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "payment_rpc_url", Label: "RPC URL", Type: tuicore.InputText,
		Value:       cfg.Payment.Network.RPCURL,
		Placeholder: "https://sepolia.base.org",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_usdc_contract", Label: "USDC Contract", Type: tuicore.InputText,
		Value:       cfg.Payment.Network.USDCContract,
		Placeholder: "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_max_per_tx", Label: "Max Per Transaction (USDC)", Type: tuicore.InputText,
		Value:       cfg.Payment.Limits.MaxPerTx,
		Placeholder: "1.00",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_max_daily", Label: "Max Daily (USDC)", Type: tuicore.InputText,
		Value:       cfg.Payment.Limits.MaxDaily,
		Placeholder: "10.00",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_auto_approve", Label: "Auto-Approve Below (USDC)", Type: tuicore.InputText,
		Value:       cfg.Payment.Limits.AutoApproveBelow,
		Placeholder: "0.10",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_x402_auto", Label: "X402 Auto-Intercept", Type: tuicore.InputBool,
		Checked: cfg.Payment.X402.AutoIntercept,
	})

	form.AddField(&tuicore.Field{
		Key: "payment_x402_max", Label: "X402 Max Auto-Pay (USDC)", Type: tuicore.InputText,
		Value:       cfg.Payment.X402.MaxAutoPayAmount,
		Placeholder: "0.50",
	})

	return &form
}

// NewCronForm creates the Cron Scheduler configuration form.
func NewCronForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Cron Scheduler Configuration")

	form.AddField(&tuicore.Field{
		Key: "cron_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked: cfg.Cron.Enabled,
	})

	form.AddField(&tuicore.Field{
		Key: "cron_timezone", Label: "Timezone", Type: tuicore.InputText,
		Value:       cfg.Cron.Timezone,
		Placeholder: "UTC or Asia/Seoul",
	})

	form.AddField(&tuicore.Field{
		Key: "cron_max_jobs", Label: "Max Concurrent Jobs", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Cron.MaxConcurrentJobs),
	})

	sessionMode := cfg.Cron.DefaultSessionMode
	if sessionMode == "" {
		sessionMode = "isolated"
	}
	form.AddField(&tuicore.Field{
		Key: "cron_session_mode", Label: "Session Mode", Type: tuicore.InputSelect,
		Value:   sessionMode,
		Options: []string{"isolated", "main"},
	})

	form.AddField(&tuicore.Field{
		Key: "cron_history_retention", Label: "History Retention", Type: tuicore.InputText,
		Value:       cfg.Cron.HistoryRetention,
		Placeholder: "30d or 720h",
	})

	form.AddField(&tuicore.Field{
		Key: "cron_default_deliver", Label: "Default Deliver To", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Cron.DefaultDeliverTo, ","),
		Placeholder: "telegram,discord,slack (comma-separated)",
	})

	return &form
}

// NewBackgroundForm creates the Background Tasks configuration form.
func NewBackgroundForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Background Tasks Configuration")

	form.AddField(&tuicore.Field{
		Key: "bg_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked: cfg.Background.Enabled,
	})

	form.AddField(&tuicore.Field{
		Key: "bg_yield_ms", Label: "Yield Time (ms)", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Background.YieldMs),
	})

	form.AddField(&tuicore.Field{
		Key: "bg_max_tasks", Label: "Max Concurrent Tasks", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Background.MaxConcurrentTasks),
	})

	form.AddField(&tuicore.Field{
		Key: "bg_default_deliver", Label: "Default Deliver To", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Background.DefaultDeliverTo, ","),
		Placeholder: "telegram,discord,slack (comma-separated)",
	})

	return &form
}

// NewWorkflowForm creates the Workflow Engine configuration form.
func NewWorkflowForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Workflow Engine Configuration")

	form.AddField(&tuicore.Field{
		Key: "wf_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked: cfg.Workflow.Enabled,
	})

	form.AddField(&tuicore.Field{
		Key: "wf_max_steps", Label: "Max Concurrent Steps", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Workflow.MaxConcurrentSteps),
	})

	form.AddField(&tuicore.Field{
		Key: "wf_timeout", Label: "Default Timeout", Type: tuicore.InputText,
		Value:       cfg.Workflow.DefaultTimeout.String(),
		Placeholder: "10m",
	})

	form.AddField(&tuicore.Field{
		Key: "wf_state_dir", Label: "State Directory", Type: tuicore.InputText,
		Value:       cfg.Workflow.StateDir,
		Placeholder: "~/.lango/workflows",
	})

	form.AddField(&tuicore.Field{
		Key: "wf_default_deliver", Label: "Default Deliver To", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Workflow.DefaultDeliverTo, ","),
		Placeholder: "telegram,discord,slack (comma-separated)",
	})

	return &form
}

// NewLibrarianForm creates the Librarian configuration form.
func NewLibrarianForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Librarian Configuration")

	form.AddField(&tuicore.Field{
		Key: "lib_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked: cfg.Librarian.Enabled,
	})

	form.AddField(&tuicore.Field{
		Key: "lib_obs_threshold", Label: "Observation Threshold", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Librarian.ObservationThreshold),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "lib_cooldown", Label: "Inquiry Cooldown Turns", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Librarian.InquiryCooldownTurns),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "lib_max_inquiries", Label: "Max Pending Inquiries", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Librarian.MaxPendingInquiries),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "lib_auto_save", Label: "Auto-Save Confidence", Type: tuicore.InputSelect,
		Value:   string(cfg.Librarian.AutoSaveConfidence),
		Options: []string{"high", "medium", "low"},
	})

	libProviderOpts := append([]string{""}, buildProviderOptions(cfg)...)
	form.AddField(&tuicore.Field{
		Key: "lib_provider", Label: "Provider", Type: tuicore.InputSelect,
		Value:   cfg.Librarian.Provider,
		Options: libProviderOpts,
	})

	form.AddField(&tuicore.Field{
		Key: "lib_model", Label: "Model", Type: tuicore.InputText,
		Value:       cfg.Librarian.Model,
		Placeholder: "leave empty for agent default",
	})

	return &form
}

// derefBool safely dereferences a *bool with a default value.
func derefBool(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

// formatKeyValueMap formats a map[string]string as "key:value" comma-separated.
func formatKeyValueMap(m map[string]string) string {
	return formatCustomPatterns(m)
}

// NewP2PForm creates the P2P Network configuration form.
func NewP2PForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("P2P Network Configuration")

	form.AddField(&tuicore.Field{
		Key: "p2p_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked: cfg.P2P.Enabled,
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_listen_addrs", Label: "Listen Addresses", Type: tuicore.InputText,
		Value:       strings.Join(cfg.P2P.ListenAddrs, ","),
		Placeholder: "/ip4/0.0.0.0/tcp/9000 (comma-separated)",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_bootstrap_peers", Label: "Bootstrap Peers", Type: tuicore.InputText,
		Value:       strings.Join(cfg.P2P.BootstrapPeers, ","),
		Placeholder: "/ip4/host/tcp/port/p2p/peerID (comma-separated)",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_enable_relay", Label: "Enable Relay", Type: tuicore.InputBool,
		Checked: cfg.P2P.EnableRelay,
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_enable_mdns", Label: "Enable mDNS", Type: tuicore.InputBool,
		Checked: cfg.P2P.EnableMDNS,
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_max_peers", Label: "Max Peers", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.P2P.MaxPeers),
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_handshake_timeout", Label: "Handshake Timeout", Type: tuicore.InputText,
		Value:       cfg.P2P.HandshakeTimeout.String(),
		Placeholder: "30s",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_session_token_ttl", Label: "Session Token TTL", Type: tuicore.InputText,
		Value:       cfg.P2P.SessionTokenTTL.String(),
		Placeholder: "24h",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_auto_approve", Label: "Auto-Approve Known Peers", Type: tuicore.InputBool,
		Checked: cfg.P2P.AutoApproveKnownPeers,
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_gossip_interval", Label: "Gossip Interval", Type: tuicore.InputText,
		Value:       cfg.P2P.GossipInterval.String(),
		Placeholder: "30s",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_zk_handshake", Label: "ZK Handshake", Type: tuicore.InputBool,
		Checked: cfg.P2P.ZKHandshake,
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_zk_attestation", Label: "ZK Attestation", Type: tuicore.InputBool,
		Checked: cfg.P2P.ZKAttestation,
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_require_signed_challenge", Label: "Require Signed Challenge", Type: tuicore.InputBool,
		Checked: cfg.P2P.RequireSignedChallenge,
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_min_trust_score", Label: "Min Trust Score", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%.1f", cfg.P2P.MinTrustScore),
		Placeholder: "0.3 (0.0 to 1.0)",
	})

	return &form
}

// NewP2PZKPForm creates the P2P ZKP configuration form.
func NewP2PZKPForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("P2P ZKP Configuration")

	form.AddField(&tuicore.Field{
		Key: "zkp_proof_cache_dir", Label: "Proof Cache Directory", Type: tuicore.InputText,
		Value:       cfg.P2P.ZKP.ProofCacheDir,
		Placeholder: "~/.lango/p2p/zkp-cache",
	})

	provingScheme := cfg.P2P.ZKP.ProvingScheme
	if provingScheme == "" {
		provingScheme = "plonk"
	}
	form.AddField(&tuicore.Field{
		Key: "zkp_proving_scheme", Label: "Proving Scheme", Type: tuicore.InputSelect,
		Value:   provingScheme,
		Options: []string{"plonk", "groth16"},
	})

	srsMode := cfg.P2P.ZKP.SRSMode
	if srsMode == "" {
		srsMode = "unsafe"
	}
	form.AddField(&tuicore.Field{
		Key: "zkp_srs_mode", Label: "SRS Mode", Type: tuicore.InputSelect,
		Value:   srsMode,
		Options: []string{"unsafe", "file"},
	})

	form.AddField(&tuicore.Field{
		Key: "zkp_srs_path", Label: "SRS File Path", Type: tuicore.InputText,
		Value:       cfg.P2P.ZKP.SRSPath,
		Placeholder: "/path/to/srs.bin (when SRS mode = file)",
	})

	form.AddField(&tuicore.Field{
		Key: "zkp_max_credential_age", Label: "Max Credential Age", Type: tuicore.InputText,
		Value:       cfg.P2P.ZKP.MaxCredentialAge,
		Placeholder: "24h",
	})

	return &form
}

// NewP2PPricingForm creates the P2P Pricing configuration form.
func NewP2PPricingForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("P2P Pricing Configuration")

	form.AddField(&tuicore.Field{
		Key: "pricing_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked: cfg.P2P.Pricing.Enabled,
	})

	form.AddField(&tuicore.Field{
		Key: "pricing_per_query", Label: "Price Per Query (USDC)", Type: tuicore.InputText,
		Value:       cfg.P2P.Pricing.PerQuery,
		Placeholder: "0.50",
	})

	form.AddField(&tuicore.Field{
		Key: "pricing_tool_prices", Label: "Tool Prices", Type: tuicore.InputText,
		Value:       formatKeyValueMap(cfg.P2P.Pricing.ToolPrices),
		Placeholder: "exec:0.10,browser:0.50 (name:price, comma-sep)",
	})

	return &form
}

// NewP2POwnerProtectionForm creates the P2P Owner Protection configuration form.
func NewP2POwnerProtectionForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("P2P Owner Protection")

	form.AddField(&tuicore.Field{
		Key: "owner_name", Label: "Owner Name", Type: tuicore.InputText,
		Value:       cfg.P2P.OwnerProtection.OwnerName,
		Placeholder: "Your name to block from P2P responses",
	})

	form.AddField(&tuicore.Field{
		Key: "owner_email", Label: "Owner Email", Type: tuicore.InputText,
		Value:       cfg.P2P.OwnerProtection.OwnerEmail,
		Placeholder: "your@email.com",
	})

	form.AddField(&tuicore.Field{
		Key: "owner_phone", Label: "Owner Phone", Type: tuicore.InputText,
		Value:       cfg.P2P.OwnerProtection.OwnerPhone,
		Placeholder: "+82-10-1234-5678",
	})

	form.AddField(&tuicore.Field{
		Key: "owner_extra_terms", Label: "Extra Terms", Type: tuicore.InputText,
		Value:       strings.Join(cfg.P2P.OwnerProtection.ExtraTerms, ","),
		Placeholder: "company-name,project-name (comma-sep)",
	})

	form.AddField(&tuicore.Field{
		Key: "owner_block_conversations", Label: "Block Conversations", Type: tuicore.InputBool,
		Checked: derefBool(cfg.P2P.OwnerProtection.BlockConversations, true),
	})

	return &form
}

// NewP2PSandboxForm creates the P2P Sandbox configuration form.
func NewP2PSandboxForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("P2P Sandbox Configuration")

	form.AddField(&tuicore.Field{
		Key: "sandbox_enabled", Label: "Tool Isolation Enabled", Type: tuicore.InputBool,
		Checked: cfg.P2P.ToolIsolation.Enabled,
	})

	form.AddField(&tuicore.Field{
		Key: "sandbox_timeout", Label: "Timeout Per Tool", Type: tuicore.InputText,
		Value:       cfg.P2P.ToolIsolation.TimeoutPerTool.String(),
		Placeholder: "30s",
	})

	form.AddField(&tuicore.Field{
		Key: "sandbox_max_memory_mb", Label: "Max Memory (MB)", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.P2P.ToolIsolation.MaxMemoryMB),
		Placeholder: "256",
	})

	form.AddField(&tuicore.Field{
		Key: "container_enabled", Label: "Container Sandbox", Type: tuicore.InputBool,
		Checked: cfg.P2P.ToolIsolation.Container.Enabled,
	})

	runtime := cfg.P2P.ToolIsolation.Container.Runtime
	if runtime == "" {
		runtime = "auto"
	}
	form.AddField(&tuicore.Field{
		Key: "container_runtime", Label: "  Runtime", Type: tuicore.InputSelect,
		Value:   runtime,
		Options: []string{"auto", "docker", "gvisor", "native"},
	})

	form.AddField(&tuicore.Field{
		Key: "container_image", Label: "  Image", Type: tuicore.InputText,
		Value:       cfg.P2P.ToolIsolation.Container.Image,
		Placeholder: "lango-sandbox:latest",
	})

	networkMode := cfg.P2P.ToolIsolation.Container.NetworkMode
	if networkMode == "" {
		networkMode = "none"
	}
	form.AddField(&tuicore.Field{
		Key: "container_network_mode", Label: "  Network Mode", Type: tuicore.InputSelect,
		Value:   networkMode,
		Options: []string{"none", "host", "bridge"},
	})

	form.AddField(&tuicore.Field{
		Key: "container_readonly_rootfs", Label: "  Read-Only Rootfs", Type: tuicore.InputBool,
		Checked: derefBool(cfg.P2P.ToolIsolation.Container.ReadOnlyRootfs, true),
	})

	form.AddField(&tuicore.Field{
		Key: "container_cpu_quota", Label: "  CPU Quota (us)", Type: tuicore.InputInt,
		Value:       strconv.FormatInt(cfg.P2P.ToolIsolation.Container.CPUQuotaUS, 10),
		Placeholder: "0 (0 = unlimited)",
	})

	form.AddField(&tuicore.Field{
		Key: "container_pool_size", Label: "  Pool Size", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.P2P.ToolIsolation.Container.PoolSize),
		Placeholder: "0 (0 = disabled)",
	})

	form.AddField(&tuicore.Field{
		Key: "container_pool_idle_timeout", Label: "  Pool Idle Timeout", Type: tuicore.InputText,
		Value:       cfg.P2P.ToolIsolation.Container.PoolIdleTimeout.String(),
		Placeholder: "5m",
	})

	return &form
}

// NewKeyringForm creates the Security Keyring configuration form.
func NewKeyringForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Security Keyring Configuration")

	form.AddField(&tuicore.Field{
		Key: "keyring_enabled", Label: "OS Keyring Enabled", Type: tuicore.InputBool,
		Checked: cfg.Security.Keyring.Enabled,
	})

	return &form
}

// NewDBEncryptionForm creates the Security DB Encryption configuration form.
func NewDBEncryptionForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Security DB Encryption Configuration")

	form.AddField(&tuicore.Field{
		Key: "db_encryption_enabled", Label: "SQLCipher Encryption", Type: tuicore.InputBool,
		Checked: cfg.Security.DBEncryption.Enabled,
	})

	form.AddField(&tuicore.Field{
		Key: "db_cipher_page_size", Label: "Cipher Page Size", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Security.DBEncryption.CipherPageSize),
		Placeholder: "4096",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	return &form
}

// NewKMSForm creates the Security KMS configuration form.
func NewKMSForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Security KMS Configuration")

	form.AddField(&tuicore.Field{
		Key: "kms_region", Label: "Region", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.Region,
		Placeholder: "us-east-1 or us-central1",
	})

	form.AddField(&tuicore.Field{
		Key: "kms_key_id", Label: "Key ID", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.KeyID,
		Placeholder: "arn:aws:kms:... or alias/my-key",
	})

	form.AddField(&tuicore.Field{
		Key: "kms_endpoint", Label: "Endpoint", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.Endpoint,
		Placeholder: "http://localhost:8080 (optional)",
	})

	form.AddField(&tuicore.Field{
		Key: "kms_fallback_to_local", Label: "Fallback to Local", Type: tuicore.InputBool,
		Checked: cfg.Security.KMS.FallbackToLocal,
	})

	form.AddField(&tuicore.Field{
		Key: "kms_timeout", Label: "Timeout Per Operation", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.TimeoutPerOperation.String(),
		Placeholder: "5s",
	})

	form.AddField(&tuicore.Field{
		Key: "kms_max_retries", Label: "Max Retries", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Security.KMS.MaxRetries),
		Placeholder: "3",
	})

	form.AddField(&tuicore.Field{
		Key: "kms_azure_vault_url", Label: "Azure Vault URL", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.Azure.VaultURL,
		Placeholder: "https://myvault.vault.azure.net",
	})

	form.AddField(&tuicore.Field{
		Key: "kms_azure_key_version", Label: "Azure Key Version", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.Azure.KeyVersion,
		Placeholder: "empty = latest",
	})

	form.AddField(&tuicore.Field{
		Key: "kms_pkcs11_module", Label: "PKCS#11 Module Path", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.PKCS11.ModulePath,
		Placeholder: "/usr/lib/pkcs11/opensc-pkcs11.so",
	})

	form.AddField(&tuicore.Field{
		Key: "kms_pkcs11_slot_id", Label: "PKCS#11 Slot ID", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Security.KMS.PKCS11.SlotID),
		Placeholder: "0",
	})

	form.AddField(&tuicore.Field{
		Key: "kms_pkcs11_pin", Label: "PKCS#11 PIN", Type: tuicore.InputPassword,
		Value:       cfg.Security.KMS.PKCS11.Pin,
		Placeholder: "prefer LANGO_PKCS11_PIN env var",
	})

	form.AddField(&tuicore.Field{
		Key: "kms_pkcs11_key_label", Label: "PKCS#11 Key Label", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.PKCS11.KeyLabel,
		Placeholder: "my-signing-key",
	})

	return &form
}

// NewProviderForm creates a Provider configuration form.
func NewProviderForm(id string, cfg config.ProviderConfig) *tuicore.FormModel {
	title := "Edit Provider: " + id
	if id == "" {
		title = "Add New Provider"
	}
	form := tuicore.NewFormModel(title)

	form.AddField(&tuicore.Field{
		Key: "type", Label: "Type", Type: tuicore.InputSelect,
		Value:   string(cfg.Type),
		Options: []string{"openai", "anthropic", "gemini", "ollama"},
	})

	if id == "" {
		form.AddField(&tuicore.Field{
			Key: "id", Label: "Provider Name", Type: tuicore.InputText,
			Placeholder: "e.g. my-openai, production-claude",
		})
	}

	form.AddField(&tuicore.Field{
		Key: "apikey", Label: "API Key", Type: tuicore.InputPassword,
		Value:       cfg.APIKey,
		Placeholder: "${ENV_VAR} or key",
	})

	form.AddField(&tuicore.Field{
		Key: "baseurl", Label: "Base URL", Type: tuicore.InputText,
		Value:       cfg.BaseURL,
		Placeholder: "https://api.example.com/v1",
	})

	return &form
}
