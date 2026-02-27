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
		Value:       cfg.Agent.Provider,
		Options:     providerOpts,
		Description: "LLM provider to use for agent inference",
	})

	form.AddField(&tuicore.Field{
		Key: "model", Label: "Model ID", Type: tuicore.InputText,
		Value:       cfg.Agent.Model,
		Placeholder: "e.g. claude-3-5-sonnet-20240620",
		Description: "Model identifier from the selected provider",
	})

	// Try to fetch models dynamically from the selected provider
	if modelOpts := FetchModelOptions(cfg.Agent.Provider, cfg, cfg.Agent.Model); len(modelOpts) > 0 {
		f := form.Fields[len(form.Fields)-1]
		f.Type = tuicore.InputSelect
		f.Options = modelOpts
		f.Placeholder = ""
		f.Description = fmt.Sprintf("Fetched %d models from provider; use ←→ to browse", len(modelOpts))
	}

	form.AddField(&tuicore.Field{
		Key: "maxtokens", Label: "Max Tokens", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Agent.MaxTokens),
		Description: "Maximum number of tokens the model can generate per response",
		Validate: func(s string) error {
			if _, err := strconv.Atoi(s); err != nil {
				return fmt.Errorf("must be integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "temp", Label: "Temperature", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%.1f", cfg.Agent.Temperature),
		Placeholder: "0.0 to 2.0",
		Description: "Controls randomness: 0.0 = deterministic, 2.0 = maximum creativity",
		Validate: func(s string) error {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("must be a number")
			}
			if f < 0 || f > 2.0 {
				return fmt.Errorf("must be between 0.0 and 2.0")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "prompts_dir", Label: "Prompts Directory", Type: tuicore.InputText,
		Value:       cfg.Agent.PromptsDir,
		Placeholder: "~/.lango/prompts (supports agents/<name>/ for per-agent overrides)",
		Description: "Directory containing system prompt templates; supports per-agent overrides",
	})

	fallbackOpts := append([]string{""}, providerOpts...)
	form.AddField(&tuicore.Field{
		Key: "fallback_provider", Label: "Fallback Provider", Type: tuicore.InputSelect,
		Value:       cfg.Agent.FallbackProvider,
		Options:     fallbackOpts,
		Description: "Alternative provider used when primary provider fails or is unavailable",
	})

	form.AddField(&tuicore.Field{
		Key: "fallback_model", Label: "Fallback Model", Type: tuicore.InputText,
		Value:       cfg.Agent.FallbackModel,
		Placeholder: "e.g. gpt-4o",
		Description: "Model to use with the fallback provider",
	})

	if cfg.Agent.FallbackProvider != "" {
		if fbModelOpts := FetchModelOptions(cfg.Agent.FallbackProvider, cfg, cfg.Agent.FallbackModel); len(fbModelOpts) > 0 {
			fbModelOpts = append([]string{""}, fbModelOpts...)
			form.Fields[len(form.Fields)-1].Type = tuicore.InputSelect
			form.Fields[len(form.Fields)-1].Options = fbModelOpts
			form.Fields[len(form.Fields)-1].Placeholder = ""
		}
	}

	form.AddField(&tuicore.Field{
		Key: "request_timeout", Label: "Request Timeout", Type: tuicore.InputText,
		Value:       cfg.Agent.RequestTimeout.String(),
		Placeholder: "5m (e.g. 30s, 2m, 5m)",
		Description: "Maximum time to wait for a single LLM API request",
	})

	form.AddField(&tuicore.Field{
		Key: "tool_timeout", Label: "Tool Timeout", Type: tuicore.InputText,
		Value:       cfg.Agent.ToolTimeout.String(),
		Placeholder: "2m (e.g. 30s, 1m, 2m)",
		Description: "Maximum execution time allowed for a single tool invocation",
	})

	return &form
}

// NewServerForm creates the Server configuration form.
func NewServerForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Server Configuration")

	form.AddField(&tuicore.Field{
		Key: "host", Label: "Host", Type: tuicore.InputText,
		Value:       cfg.Server.Host,
		Description: "Network interface to bind to; use 0.0.0.0 to listen on all interfaces",
	})

	form.AddField(&tuicore.Field{
		Key: "port", Label: "Port", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Server.Port),
		Validate:    validatePort,
		Description: "TCP port for the HTTP/WebSocket server (1-65535)",
	})

	form.AddField(&tuicore.Field{
		Key: "http", Label: "Generic HTTP", Type: tuicore.InputBool,
		Checked:     cfg.Server.HTTPEnabled,
		Description: "Enable REST API endpoint for HTTP-based integrations",
	})

	form.AddField(&tuicore.Field{
		Key: "ws", Label: "WebSockets", Type: tuicore.InputBool,
		Checked:     cfg.Server.WebSocketEnabled,
		Description: "Enable WebSocket endpoint for real-time bidirectional communication",
	})

	return &form
}

// NewChannelsForm creates the Channels configuration form.
func NewChannelsForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Channels Configuration")

	telegramEnabled := &tuicore.Field{
		Key: "telegram_enabled", Label: "Telegram", Type: tuicore.InputBool,
		Checked:     cfg.Channels.Telegram.Enabled,
		Description: "Enable Telegram bot channel for receiving and sending messages",
	}
	form.AddField(telegramEnabled)
	form.AddField(&tuicore.Field{
		Key: "telegram_token", Label: "  Bot Token", Type: tuicore.InputPassword,
		Value:       cfg.Channels.Telegram.BotToken,
		Placeholder: "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
		Description: "Bot token from @BotFather; use ${ENV_VAR} to reference environment variables",
		VisibleWhen: func() bool { return telegramEnabled.Checked },
	})

	discordEnabled := &tuicore.Field{
		Key: "discord_enabled", Label: "Discord", Type: tuicore.InputBool,
		Checked:     cfg.Channels.Discord.Enabled,
		Description: "Enable Discord bot channel for receiving and sending messages",
	}
	form.AddField(discordEnabled)
	form.AddField(&tuicore.Field{
		Key: "discord_token", Label: "  Bot Token", Type: tuicore.InputPassword,
		Value:       cfg.Channels.Discord.BotToken,
		Description: "Bot token from Discord Developer Portal; use ${ENV_VAR} for security",
		VisibleWhen: func() bool { return discordEnabled.Checked },
	})

	slackEnabled := &tuicore.Field{
		Key: "slack_enabled", Label: "Slack", Type: tuicore.InputBool,
		Checked:     cfg.Channels.Slack.Enabled,
		Description: "Enable Slack bot channel using Socket Mode",
	}
	form.AddField(slackEnabled)
	form.AddField(&tuicore.Field{
		Key: "slack_token", Label: "  Bot Token", Type: tuicore.InputPassword,
		Value:       cfg.Channels.Slack.BotToken,
		Placeholder: "xoxb-...",
		Description: "Slack Bot User OAuth Token (starts with xoxb-)",
		VisibleWhen: func() bool { return slackEnabled.Checked },
	})
	form.AddField(&tuicore.Field{
		Key: "slack_app_token", Label: "  App Token", Type: tuicore.InputPassword,
		Value:       cfg.Channels.Slack.AppToken,
		Placeholder: "xapp-...",
		Description: "Slack App-Level Token for Socket Mode (starts with xapp-)",
		VisibleWhen: func() bool { return slackEnabled.Checked },
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
		Description: "Default timeout for shell command execution",
	})
	form.AddField(&tuicore.Field{
		Key: "exec_bg", Label: "Allow Background", Type: tuicore.InputBool,
		Checked:     cfg.Tools.Exec.AllowBackground,
		Description: "Allow the agent to run shell commands in the background",
	})

	form.AddField(&tuicore.Field{
		Key: "browser_enabled", Label: "Browser Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Tools.Browser.Enabled,
		Description: "Enable headless browser tool for web scraping and interaction",
	})
	form.AddField(&tuicore.Field{
		Key: "browser_headless", Label: "Browser Headless", Type: tuicore.InputBool,
		Checked:     cfg.Tools.Browser.Headless,
		Description: "Run browser without visible UI window; disable for debugging",
	})
	form.AddField(&tuicore.Field{
		Key: "browser_session_timeout", Label: "Browser Session Timeout", Type: tuicore.InputText,
		Value:       cfg.Tools.Browser.SessionTimeout.String(),
		Placeholder: "5m",
		Description: "Maximum duration for a single browser session before auto-close",
	})

	form.AddField(&tuicore.Field{
		Key: "fs_max_read", Label: "Max Read Size", Type: tuicore.InputInt,
		Value:       strconv.FormatInt(cfg.Tools.Filesystem.MaxReadSize, 10),
		Description: "Maximum file size in bytes that the filesystem tool can read",
		Validate: func(s string) error {
			if i, err := strconv.ParseInt(s, 10, 64); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	return &form
}

// NewSessionForm creates the Session configuration form.
func NewSessionForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Session Configuration")

	form.AddField(&tuicore.Field{
		Key: "ttl", Label: "Session TTL", Type: tuicore.InputText,
		Value:       cfg.Session.TTL.String(),
		Description: "Time-to-live before an idle session expires (e.g. 24h, 7d)",
	})

	form.AddField(&tuicore.Field{
		Key: "max_history_turns", Label: "Max History Turns", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Session.MaxHistoryTurns),
		Description: "Maximum number of conversation turns kept in session history",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	return &form
}

// NewSecurityForm creates the Security configuration form.
func NewSecurityForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Security Configuration")

	interceptorEnabled := &tuicore.Field{
		Key: "interceptor_enabled", Label: "Privacy Interceptor", Type: tuicore.InputBool,
		Checked:     cfg.Security.Interceptor.Enabled,
		Description: "Enable the privacy interceptor to filter outgoing data",
	}
	form.AddField(interceptorEnabled)
	isInterceptorOn := func() bool { return interceptorEnabled.Checked }

	form.AddField(&tuicore.Field{
		Key: "interceptor_pii", Label: "  Redact PII", Type: tuicore.InputBool,
		Checked:     cfg.Security.Interceptor.RedactPII,
		Description: "Automatically redact personally identifiable information from messages",
		VisibleWhen: isInterceptorOn,
	})
	policyVal := string(cfg.Security.Interceptor.ApprovalPolicy)
	if policyVal == "" {
		policyVal = "dangerous"
	}
	form.AddField(&tuicore.Field{
		Key: "interceptor_policy", Label: "  Approval Policy", Type: tuicore.InputSelect,
		Value:       policyVal,
		Options:     []string{"dangerous", "all", "configured", "none"},
		Description: "When to require user approval: dangerous=risky tools, all=every tool, none=skip",
		VisibleWhen: isInterceptorOn,
	})

	form.AddField(&tuicore.Field{
		Key: "interceptor_timeout", Label: "  Approval Timeout (s)", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Security.Interceptor.ApprovalTimeoutSec),
		Description: "Seconds to wait for user approval before auto-denying; 0 = wait forever",
		VisibleWhen: isInterceptorOn,
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "interceptor_notify", Label: "  Notify Channel", Type: tuicore.InputSelect,
		Value:       cfg.Security.Interceptor.NotifyChannel,
		Options:     []string{"", string(types.ChannelTelegram), string(types.ChannelDiscord), string(types.ChannelSlack)},
		Description: "Channel to send approval notifications to; empty = no notification",
		VisibleWhen: isInterceptorOn,
	})

	form.AddField(&tuicore.Field{
		Key: "interceptor_sensitive_tools", Label: "  Sensitive Tools", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Security.Interceptor.SensitiveTools, ","),
		Placeholder: "exec,browser (comma-separated)",
		Description: "Tools that always require approval regardless of approval policy",
		VisibleWhen: isInterceptorOn,
	})

	form.AddField(&tuicore.Field{
		Key: "interceptor_exempt_tools", Label: "  Exempt Tools", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Security.Interceptor.ExemptTools, ","),
		Placeholder: "filesystem (comma-separated)",
		Description: "Tools that never require approval, even with 'all' policy",
		VisibleWhen: isInterceptorOn,
	})

	// PII Pattern Management
	form.AddField(&tuicore.Field{
		Key: "interceptor_pii_disabled", Label: "  Disabled PII Patterns", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Security.Interceptor.PIIDisabledPatterns, ","),
		Placeholder: "kr_bank_account,passport,ipv4 (comma-separated)",
		Description: "Built-in PII pattern names to disable (e.g. ipv4, passport)",
		VisibleWhen: isInterceptorOn,
	})
	form.AddField(&tuicore.Field{
		Key: "interceptor_pii_custom", Label: "  Custom PII Patterns", Type: tuicore.InputText,
		Value:       formatCustomPatterns(cfg.Security.Interceptor.PIICustomPatterns),
		Placeholder: `my_id:\bID-\d{6}\b (name:regex, comma-sep)`,
		Description: "Custom regex patterns for PII detection in name:regex format",
		VisibleWhen: isInterceptorOn,
	})

	// Presidio Integration
	presidioEnabled := &tuicore.Field{
		Key: "presidio_enabled", Label: "  Presidio (Docker)", Type: tuicore.InputBool,
		Checked:     cfg.Security.Interceptor.Presidio.Enabled,
		Description: "Use Microsoft Presidio (Docker) for advanced NLP-based PII detection",
		VisibleWhen: isInterceptorOn,
	}
	form.AddField(presidioEnabled)
	isPresidioOn := func() bool { return isInterceptorOn() && presidioEnabled.Checked }
	form.AddField(&tuicore.Field{
		Key: "presidio_url", Label: "    Presidio URL", Type: tuicore.InputText,
		Value:       cfg.Security.Interceptor.Presidio.URL,
		Placeholder: "http://localhost:5002",
		Description: "URL of the Presidio analyzer service endpoint",
		VisibleWhen: isPresidioOn,
	})
	presidioLang := cfg.Security.Interceptor.Presidio.Language
	if presidioLang == "" {
		presidioLang = "en"
	}
	form.AddField(&tuicore.Field{
		Key: "presidio_language", Label: "    Presidio Language", Type: tuicore.InputSelect,
		Value:       presidioLang,
		Options:     []string{"en", "ko", "ja", "zh", "de", "fr", "es", "it", "pt", "nl", "ru"},
		Description: "Primary language for Presidio NLP analysis",
		VisibleWhen: isPresidioOn,
	})

	// Signer Configuration
	signerField := &tuicore.Field{
		Key: "signer_provider", Label: "Signer Provider", Type: tuicore.InputSelect,
		Value:       cfg.Security.Signer.Provider,
		Options:     []string{"local", "rpc", "enclave", "aws-kms", "gcp-kms", "azure-kv", "pkcs11"},
		Description: "Cryptographic signer backend for message signing and verification",
	}
	form.AddField(signerField)
	form.AddField(&tuicore.Field{
		Key: "signer_rpc", Label: "  RPC URL", Type: tuicore.InputText,
		Value:       cfg.Security.Signer.RPCUrl,
		Placeholder: "http://localhost:8080",
		Description: "URL of the remote signing service",
		VisibleWhen: func() bool { return signerField.Value == "rpc" },
	})
	form.AddField(&tuicore.Field{
		Key: "signer_keyid", Label: "  Key ID", Type: tuicore.InputText,
		Value:       cfg.Security.Signer.KeyID,
		Placeholder: "key-123",
		Description: "Key identifier for the signer (ARN for AWS, key name for GCP/Azure)",
		VisibleWhen: func() bool {
			v := signerField.Value
			return v == "rpc" || v == "aws-kms" || v == "gcp-kms" || v == "azure-kv" || v == "pkcs11"
		},
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
		Checked:     cfg.Knowledge.Enabled,
		Description: "Enable the knowledge layer for persistent learning across sessions",
	})

	form.AddField(&tuicore.Field{
		Key: "knowledge_max_context", Label: "Max Context/Layer", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Knowledge.MaxContextPerLayer),
		Description: "Maximum tokens of context injected per knowledge layer per turn",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	return &form
}

// NewSkillForm creates the Skill configuration form.
func NewSkillForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Skill Configuration")

	form.AddField(&tuicore.Field{
		Key: "skill_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Skill.Enabled,
		Description: "Enable file-based skill system for reusable agent capabilities",
	})

	form.AddField(&tuicore.Field{
		Key: "skill_dir", Label: "Skills Directory", Type: tuicore.InputText,
		Value:       cfg.Skill.SkillsDir,
		Placeholder: "~/.lango/skills",
		Description: "Directory where skill YAML files are stored and loaded from",
	})

	form.AddField(&tuicore.Field{
		Key: "skill_allow_import", Label: "Allow Import", Type: tuicore.InputBool,
		Checked:     cfg.Skill.AllowImport,
		Description: "Allow importing skills from external sources (URLs, P2P peers)",
	})

	form.AddField(&tuicore.Field{
		Key: "skill_max_bulk", Label: "Max Bulk Import", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Skill.MaxBulkImport),
		Placeholder: "50",
		Description: "Maximum number of skills to import in a single bulk operation",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "skill_import_concurrency", Label: "Import Concurrency", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Skill.ImportConcurrency),
		Placeholder: "5",
		Description: "Number of skills to import in parallel during bulk operations",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "skill_import_timeout", Label: "Import Timeout", Type: tuicore.InputText,
		Value:       cfg.Skill.ImportTimeout.String(),
		Placeholder: "2m (e.g. 30s, 1m, 5m)",
		Description: "Maximum time allowed for a single skill import operation",
	})

	return &form
}

// NewObservationalMemoryForm creates the Observational Memory configuration form.
func NewObservationalMemoryForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Observational Memory")

	form.AddField(&tuicore.Field{
		Key: "om_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.ObservationalMemory.Enabled,
		Description: "Enable observational memory for automatic user behavior learning",
	})

	omProviderOpts := append([]string{""}, buildProviderOptions(cfg)...)
	form.AddField(&tuicore.Field{
		Key: "om_provider", Label: "Provider", Type: tuicore.InputSelect,
		Value:       cfg.ObservationalMemory.Provider,
		Options:     omProviderOpts,
		Placeholder: "(inherits from Agent)",
		Description: fmt.Sprintf("LLM provider for memory processing; empty = inherit from Agent (%s)", cfg.Agent.Provider),
	})

	form.AddField(&tuicore.Field{
		Key: "om_model", Label: "Model", Type: tuicore.InputText,
		Value:       cfg.ObservationalMemory.Model,
		Placeholder: "(inherits from Agent)",
		Description: fmt.Sprintf("Model for observation/reflection generation; empty = inherit from Agent (%s)", cfg.Agent.Model),
	})

	omFetchProvider := cfg.ObservationalMemory.Provider
	if omFetchProvider == "" {
		omFetchProvider = cfg.Agent.Provider
	}
	if omModelOpts := FetchModelOptions(omFetchProvider, cfg, cfg.ObservationalMemory.Model); len(omModelOpts) > 0 {
		omModelOpts = append([]string{""}, omModelOpts...)
		form.Fields[len(form.Fields)-1].Type = tuicore.InputSelect
		form.Fields[len(form.Fields)-1].Options = omModelOpts
		form.Fields[len(form.Fields)-1].Placeholder = ""
	}

	form.AddField(&tuicore.Field{
		Key: "om_msg_threshold", Label: "Message Token Threshold",
		Type:        tuicore.InputInt,
		Value:       strconv.Itoa(cfg.ObservationalMemory.MessageTokenThreshold),
		Description: "Minimum tokens in a message before it triggers observation",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "om_obs_threshold", Label: "Observation Token Threshold",
		Type:        tuicore.InputInt,
		Value:       strconv.Itoa(cfg.ObservationalMemory.ObservationTokenThreshold),
		Description: "Token threshold to trigger consolidation into reflections",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "om_max_budget", Label: "Max Message Token Budget",
		Type:        tuicore.InputInt,
		Value:       strconv.Itoa(cfg.ObservationalMemory.MaxMessageTokenBudget),
		Description: "Maximum tokens allocated for memory context in each turn",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "om_max_reflections", Label: "Max Reflections in Context",
		Type:        tuicore.InputInt,
		Value:       strconv.Itoa(cfg.ObservationalMemory.MaxReflectionsInContext),
		Description: "Max reflections injected per turn; 0 = unlimited",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer (0 = unlimited)")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "om_max_observations", Label: "Max Observations in Context",
		Type:        tuicore.InputInt,
		Value:       strconv.Itoa(cfg.ObservationalMemory.MaxObservationsInContext),
		Description: "Max raw observations injected per turn; 0 = unlimited",
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

	form.AddField(&tuicore.Field{
		Key: "emb_provider_id", Label: "Provider", Type: tuicore.InputSelect,
		Value:       cfg.Embedding.Provider,
		Options:     providerOpts,
		Description: "Embedding provider; 'local' uses a local model via Ollama/compatible API",
	})

	form.AddField(&tuicore.Field{
		Key: "emb_model", Label: "Model", Type: tuicore.InputText,
		Value:       cfg.Embedding.Model,
		Placeholder: "e.g. text-embedding-3-small",
		Description: "Embedding model name; must be supported by the selected provider",
	})

	if cfg.Embedding.Provider != "" {
		if embModelOpts := FetchModelOptions(cfg.Embedding.Provider, cfg, cfg.Embedding.Model); len(embModelOpts) > 0 {
			embModelOpts = append([]string{""}, embModelOpts...)
			form.Fields[len(form.Fields)-1].Type = tuicore.InputSelect
			form.Fields[len(form.Fields)-1].Options = embModelOpts
			form.Fields[len(form.Fields)-1].Placeholder = ""
		}
	}

	form.AddField(&tuicore.Field{
		Key: "emb_dimensions", Label: "Dimensions", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Embedding.Dimensions),
		Description: "Vector dimensions for embeddings; 0 = use model default",
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
		Description: "API base URL for the local embedding server (Ollama, vLLM, etc.)",
	})

	form.AddField(&tuicore.Field{
		Key: "emb_rag_enabled", Label: "RAG Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Embedding.RAG.Enabled,
		Description: "Enable Retrieval-Augmented Generation for knowledge-enhanced responses",
	})

	form.AddField(&tuicore.Field{
		Key: "emb_rag_max_results", Label: "RAG Max Results", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Embedding.RAG.MaxResults),
		Description: "Maximum number of retrieved chunks injected into context per query",
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
		Description: "Vector store collections to search during RAG retrieval",
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
			Description: "Unique identifier for this OIDC provider configuration",
		})
	}

	form.AddField(&tuicore.Field{
		Key: "oidc_issuer", Label: "Issuer URL", Type: tuicore.InputText,
		Value:       cfg.IssuerURL,
		Placeholder: "https://accounts.google.com",
		Description: "OIDC issuer URL used for auto-discovery of endpoints",
	})

	form.AddField(&tuicore.Field{
		Key: "oidc_client_id", Label: "Client ID", Type: tuicore.InputPassword,
		Value:       cfg.ClientID,
		Placeholder: "${ENV_VAR} or value",
		Description: "OAuth2 Client ID from the identity provider; supports ${ENV_VAR} syntax",
	})

	form.AddField(&tuicore.Field{
		Key: "oidc_client_secret", Label: "Client Secret", Type: tuicore.InputPassword,
		Value:       cfg.ClientSecret,
		Placeholder: "${ENV_VAR} or value",
		Description: "OAuth2 Client Secret; strongly recommend using ${ENV_VAR} for security",
	})

	form.AddField(&tuicore.Field{
		Key: "oidc_redirect", Label: "Redirect URL", Type: tuicore.InputText,
		Value:       cfg.RedirectURL,
		Placeholder: "http://localhost:18789/auth/callback/<name>",
		Description: "Callback URL registered with the identity provider",
	})

	form.AddField(&tuicore.Field{
		Key: "oidc_scopes", Label: "Scopes", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Scopes, ","),
		Placeholder: "openid,email,profile",
		Description: "OAuth2 scopes to request; openid is required for OIDC",
	})

	return &form
}

// NewGraphForm creates the Graph Store configuration form.
func NewGraphForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Graph Store Configuration")

	form.AddField(&tuicore.Field{
		Key: "graph_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Graph.Enabled,
		Description: "Enable knowledge graph for structured entity and relationship storage",
	})

	form.AddField(&tuicore.Field{
		Key: "graph_backend", Label: "Backend", Type: tuicore.InputSelect,
		Value:       cfg.Graph.Backend,
		Options:     []string{"bolt"},
		Description: "Graph database backend; 'bolt' uses embedded BoltDB",
	})

	form.AddField(&tuicore.Field{
		Key: "graph_db_path", Label: "Database Path", Type: tuicore.InputText,
		Value:       cfg.Graph.DatabasePath,
		Placeholder: "~/.lango/graph.db",
		Description: "File path for the graph database storage",
	})

	form.AddField(&tuicore.Field{
		Key: "graph_max_depth", Label: "Max Traversal Depth", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Graph.MaxTraversalDepth),
		Description: "Maximum depth for graph traversal queries (hop count)",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "graph_max_expand", Label: "Max Expansion Results", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Graph.MaxExpansionResults),
		Description: "Maximum nodes returned per expansion query",
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
		Checked:     cfg.Agent.MultiAgent,
		Description: "Allow the agent to spawn and coordinate sub-agents for complex tasks",
	})

	return &form
}

// NewA2AForm creates the A2A Protocol configuration form.
func NewA2AForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("A2A Protocol Configuration")

	form.AddField(&tuicore.Field{
		Key: "a2a_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.A2A.Enabled,
		Description: "Enable Google A2A (Agent-to-Agent) protocol for inter-agent communication",
	})

	form.AddField(&tuicore.Field{
		Key: "a2a_base_url", Label: "Base URL", Type: tuicore.InputText,
		Value:       cfg.A2A.BaseURL,
		Placeholder: "https://your-agent.example.com",
		Description: "Public URL where this agent's A2A endpoint is accessible",
	})

	form.AddField(&tuicore.Field{
		Key: "a2a_agent_name", Label: "Agent Name", Type: tuicore.InputText,
		Value:       cfg.A2A.AgentName,
		Placeholder: "my-lango-agent",
		Description: "Human-readable name advertised in the A2A agent card",
	})

	form.AddField(&tuicore.Field{
		Key: "a2a_agent_desc", Label: "Agent Description", Type: tuicore.InputText,
		Value:       cfg.A2A.AgentDescription,
		Placeholder: "A helpful AI assistant",
		Description: "Description of this agent's capabilities for A2A discovery",
	})

	return &form
}

// NewPaymentForm creates the Payment configuration form.
func NewPaymentForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Payment Configuration")

	form.AddField(&tuicore.Field{
		Key: "payment_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Payment.Enabled,
		Description: "Enable blockchain-based USDC payment capabilities",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_wallet_provider", Label: "Wallet Provider", Type: tuicore.InputSelect,
		Value:       cfg.Payment.WalletProvider,
		Options:     []string{"local", "rpc", "composite"},
		Description: "Wallet backend: local=embedded key, rpc=remote signer, composite=multi-wallet",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_chain_id", Label: "Chain ID", Type: tuicore.InputInt,
		Value:       strconv.FormatInt(cfg.Payment.Network.ChainID, 10),
		Description: "EVM chain ID (e.g. 84532 for Base Sepolia, 8453 for Base Mainnet)",
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
		Description: "Ethereum JSON-RPC endpoint URL for blockchain interactions",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_usdc_contract", Label: "USDC Contract", Type: tuicore.InputText,
		Value:       cfg.Payment.Network.USDCContract,
		Placeholder: "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		Description: "USDC token contract address on the selected chain",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_max_per_tx", Label: "Max Per Transaction (USDC)", Type: tuicore.InputText,
		Value:       cfg.Payment.Limits.MaxPerTx,
		Placeholder: "1.00",
		Description: "Maximum USDC amount allowed per single transaction",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_max_daily", Label: "Max Daily (USDC)", Type: tuicore.InputText,
		Value:       cfg.Payment.Limits.MaxDaily,
		Placeholder: "10.00",
		Description: "Maximum total USDC spending allowed per 24-hour period",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_auto_approve", Label: "Auto-Approve Below (USDC)", Type: tuicore.InputText,
		Value:       cfg.Payment.Limits.AutoApproveBelow,
		Placeholder: "0.10",
		Description: "Transactions below this amount are auto-approved without user confirmation",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_x402_auto", Label: "X402 Auto-Intercept", Type: tuicore.InputBool,
		Checked:     cfg.Payment.X402.AutoIntercept,
		Description: "Automatically handle HTTP 402 Payment Required responses with USDC",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_x402_max", Label: "X402 Max Auto-Pay (USDC)", Type: tuicore.InputText,
		Value:       cfg.Payment.X402.MaxAutoPayAmount,
		Placeholder: "0.50",
		Description: "Maximum USDC to auto-pay for a single X402 response",
	})

	return &form
}

// NewCronForm creates the Cron Scheduler configuration form.
func NewCronForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Cron Scheduler Configuration")

	form.AddField(&tuicore.Field{
		Key: "cron_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Cron.Enabled,
		Description: "Enable the cron scheduler for recurring automated tasks",
	})

	form.AddField(&tuicore.Field{
		Key: "cron_timezone", Label: "Timezone", Type: tuicore.InputText,
		Value:       cfg.Cron.Timezone,
		Placeholder: "UTC or Asia/Seoul",
		Description: "IANA timezone for cron schedule evaluation (e.g. America/New_York)",
	})

	form.AddField(&tuicore.Field{
		Key: "cron_max_jobs", Label: "Max Concurrent Jobs", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Cron.MaxConcurrentJobs),
		Description: "Maximum number of cron jobs that can run simultaneously",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	sessionMode := cfg.Cron.DefaultSessionMode
	if sessionMode == "" {
		sessionMode = "isolated"
	}
	form.AddField(&tuicore.Field{
		Key: "cron_session_mode", Label: "Session Mode", Type: tuicore.InputSelect,
		Value:       sessionMode,
		Options:     []string{"isolated", "main"},
		Description: "isolated=separate session per job, main=shared with main conversation",
	})

	form.AddField(&tuicore.Field{
		Key: "cron_history_retention", Label: "History Retention", Type: tuicore.InputText,
		Value:       cfg.Cron.HistoryRetention,
		Placeholder: "30d or 720h",
		Description: "How long to keep cron job execution history",
	})

	form.AddField(&tuicore.Field{
		Key: "cron_default_deliver", Label: "Default Deliver To", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Cron.DefaultDeliverTo, ","),
		Placeholder: "telegram,discord,slack (comma-separated)",
		Description: "Default channels to deliver cron job results to",
	})

	return &form
}

// NewBackgroundForm creates the Background Tasks configuration form.
func NewBackgroundForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Background Tasks Configuration")

	form.AddField(&tuicore.Field{
		Key: "bg_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Background.Enabled,
		Description: "Enable asynchronous background task execution",
	})

	form.AddField(&tuicore.Field{
		Key: "bg_yield_ms", Label: "Yield Time (ms)", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Background.YieldMs),
		Description: "Milliseconds to yield between task steps to avoid CPU monopolization",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "bg_max_tasks", Label: "Max Concurrent Tasks", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Background.MaxConcurrentTasks),
		Description: "Maximum number of background tasks running at the same time",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "bg_default_deliver", Label: "Default Deliver To", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Background.DefaultDeliverTo, ","),
		Placeholder: "telegram,discord,slack (comma-separated)",
		Description: "Default channels to deliver background task results to",
	})

	return &form
}

// NewWorkflowForm creates the Workflow Engine configuration form.
func NewWorkflowForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Workflow Engine Configuration")

	form.AddField(&tuicore.Field{
		Key: "wf_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Workflow.Enabled,
		Description: "Enable the DAG-based workflow engine for multi-step pipelines",
	})

	form.AddField(&tuicore.Field{
		Key: "wf_max_steps", Label: "Max Concurrent Steps", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Workflow.MaxConcurrentSteps),
		Description: "Maximum workflow steps executed in parallel",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "wf_timeout", Label: "Default Timeout", Type: tuicore.InputText,
		Value:       cfg.Workflow.DefaultTimeout.String(),
		Placeholder: "10m",
		Description: "Default timeout for an entire workflow execution",
	})

	form.AddField(&tuicore.Field{
		Key: "wf_state_dir", Label: "State Directory", Type: tuicore.InputText,
		Value:       cfg.Workflow.StateDir,
		Placeholder: "~/.lango/workflows",
		Description: "Directory to persist workflow state and checkpoints",
	})

	form.AddField(&tuicore.Field{
		Key: "wf_default_deliver", Label: "Default Deliver To", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Workflow.DefaultDeliverTo, ","),
		Placeholder: "telegram,discord,slack (comma-separated)",
		Description: "Default channels to deliver workflow completion results to",
	})

	return &form
}

// NewLibrarianForm creates the Librarian configuration form.
func NewLibrarianForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Librarian Configuration")

	form.AddField(&tuicore.Field{
		Key: "lib_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Librarian.Enabled,
		Description: "Enable proactive knowledge extraction from conversations",
	})

	form.AddField(&tuicore.Field{
		Key: "lib_obs_threshold", Label: "Observation Threshold", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Librarian.ObservationThreshold),
		Description: "Minimum observations before the librarian triggers knowledge extraction",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "lib_cooldown", Label: "Inquiry Cooldown Turns", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Librarian.InquiryCooldownTurns),
		Description: "Minimum turns between librarian inquiries to avoid being intrusive",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "lib_max_inquiries", Label: "Max Pending Inquiries", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Librarian.MaxPendingInquiries),
		Description: "Maximum unanswered inquiries before pausing new ones",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "lib_auto_save", Label: "Auto-Save Confidence", Type: tuicore.InputSelect,
		Value:       string(cfg.Librarian.AutoSaveConfidence),
		Options:     []string{"high", "medium", "low"},
		Description: "Confidence threshold for auto-saving extracted knowledge without confirmation",
	})

	libProviderOpts := append([]string{""}, buildProviderOptions(cfg)...)
	form.AddField(&tuicore.Field{
		Key: "lib_provider", Label: "Provider", Type: tuicore.InputSelect,
		Value:       cfg.Librarian.Provider,
		Options:     libProviderOpts,
		Placeholder: "(inherits from Agent)",
		Description: fmt.Sprintf("LLM provider for librarian processing; empty = inherit from Agent (%s)", cfg.Agent.Provider),
	})

	form.AddField(&tuicore.Field{
		Key: "lib_model", Label: "Model", Type: tuicore.InputText,
		Value:       cfg.Librarian.Model,
		Placeholder: "(inherits from Agent)",
		Description: fmt.Sprintf("Model for knowledge extraction; empty = inherit from Agent (%s)", cfg.Agent.Model),
	})

	libFetchProvider := cfg.Librarian.Provider
	if libFetchProvider == "" {
		libFetchProvider = cfg.Agent.Provider
	}
	if libModelOpts := FetchModelOptions(libFetchProvider, cfg, cfg.Librarian.Model); len(libModelOpts) > 0 {
		libModelOpts = append([]string{""}, libModelOpts...)
		form.Fields[len(form.Fields)-1].Type = tuicore.InputSelect
		form.Fields[len(form.Fields)-1].Options = libModelOpts
		form.Fields[len(form.Fields)-1].Placeholder = ""
	}

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
		Checked:     cfg.P2P.Enabled,
		Description: "Enable libp2p-based peer-to-peer networking for agent discovery",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_listen_addrs", Label: "Listen Addresses", Type: tuicore.InputText,
		Value:       strings.Join(cfg.P2P.ListenAddrs, ","),
		Placeholder: "/ip4/0.0.0.0/tcp/9000 (comma-separated)",
		Description: "Multiaddr listen addresses for incoming P2P connections",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_bootstrap_peers", Label: "Bootstrap Peers", Type: tuicore.InputText,
		Value:       strings.Join(cfg.P2P.BootstrapPeers, ","),
		Placeholder: "/ip4/host/tcp/port/p2p/peerID (comma-separated)",
		Description: "Initial peers to connect to for network discovery",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_enable_relay", Label: "Enable Relay", Type: tuicore.InputBool,
		Checked:     cfg.P2P.EnableRelay,
		Description: "Allow relaying connections for peers behind NAT",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_enable_mdns", Label: "Enable mDNS", Type: tuicore.InputBool,
		Checked:     cfg.P2P.EnableMDNS,
		Description: "Use multicast DNS for local network peer discovery",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_max_peers", Label: "Max Peers", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.P2P.MaxPeers),
		Description: "Maximum number of simultaneous peer connections",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_handshake_timeout", Label: "Handshake Timeout", Type: tuicore.InputText,
		Value:       cfg.P2P.HandshakeTimeout.String(),
		Placeholder: "30s",
		Description: "Maximum time to wait for peer handshake completion",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_session_token_ttl", Label: "Session Token TTL", Type: tuicore.InputText,
		Value:       cfg.P2P.SessionTokenTTL.String(),
		Placeholder: "24h",
		Description: "Lifetime of P2P session tokens before re-authentication is required",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_auto_approve", Label: "Auto-Approve Known Peers", Type: tuicore.InputBool,
		Checked:     cfg.P2P.AutoApproveKnownPeers,
		Description: "Skip approval for previously authenticated and trusted peers",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_gossip_interval", Label: "Gossip Interval", Type: tuicore.InputText,
		Value:       cfg.P2P.GossipInterval.String(),
		Placeholder: "30s",
		Description: "Interval between gossip protocol broadcasts for peer discovery",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_zk_handshake", Label: "ZK Handshake", Type: tuicore.InputBool,
		Checked:     cfg.P2P.ZKHandshake,
		Description: "Use zero-knowledge proofs during peer handshake for privacy",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_zk_attestation", Label: "ZK Attestation", Type: tuicore.InputBool,
		Checked:     cfg.P2P.ZKAttestation,
		Description: "Require ZK attestation proofs for tool execution results",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_require_signed_challenge", Label: "Require Signed Challenge", Type: tuicore.InputBool,
		Checked:     cfg.P2P.RequireSignedChallenge,
		Description: "Require cryptographic challenge-response during peer authentication",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_min_trust_score", Label: "Min Trust Score", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%.1f", cfg.P2P.MinTrustScore),
		Placeholder: "0.3 (0.0 to 1.0)",
		Description: "Minimum trust score (0.0-1.0) required to interact with a peer",
		Validate: func(s string) error {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("must be a number")
			}
			if f < 0 || f > 1.0 {
				return fmt.Errorf("must be between 0.0 and 1.0")
			}
			return nil
		},
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
		Description: "Directory to cache generated zero-knowledge proofs",
	})

	provingScheme := cfg.P2P.ZKP.ProvingScheme
	if provingScheme == "" {
		provingScheme = "plonk"
	}
	form.AddField(&tuicore.Field{
		Key: "zkp_proving_scheme", Label: "Proving Scheme", Type: tuicore.InputSelect,
		Value:       provingScheme,
		Options:     []string{"plonk", "groth16"},
		Description: "ZKP proving system: plonk=universal setup, groth16=faster but circuit-specific",
	})

	srsMode := cfg.P2P.ZKP.SRSMode
	if srsMode == "" {
		srsMode = "unsafe"
	}
	form.AddField(&tuicore.Field{
		Key: "zkp_srs_mode", Label: "SRS Mode", Type: tuicore.InputSelect,
		Value:       srsMode,
		Options:     []string{"unsafe", "file"},
		Description: "Structured Reference String mode: unsafe=dev-only random, file=from trusted setup",
	})

	form.AddField(&tuicore.Field{
		Key: "zkp_srs_path", Label: "SRS File Path", Type: tuicore.InputText,
		Value:       cfg.P2P.ZKP.SRSPath,
		Placeholder: "/path/to/srs.bin (when SRS mode = file)",
		Description: "Path to the SRS file from a trusted ceremony (required when mode=file)",
	})

	form.AddField(&tuicore.Field{
		Key: "zkp_max_credential_age", Label: "Max Credential Age", Type: tuicore.InputText,
		Value:       cfg.P2P.ZKP.MaxCredentialAge,
		Placeholder: "24h",
		Description: "Maximum age of a ZKP credential before it must be refreshed",
	})

	return &form
}

// NewP2PPricingForm creates the P2P Pricing configuration form.
func NewP2PPricingForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("P2P Pricing Configuration")

	form.AddField(&tuicore.Field{
		Key: "pricing_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.P2P.Pricing.Enabled,
		Description: "Enable paid tool invocations from P2P peers",
	})

	form.AddField(&tuicore.Field{
		Key: "pricing_per_query", Label: "Price Per Query (USDC)", Type: tuicore.InputText,
		Value:       cfg.P2P.Pricing.PerQuery,
		Placeholder: "0.50",
		Description: "USDC price charged per incoming P2P query",
	})

	form.AddField(&tuicore.Field{
		Key: "pricing_tool_prices", Label: "Tool Prices", Type: tuicore.InputText,
		Value:       formatKeyValueMap(cfg.P2P.Pricing.ToolPrices),
		Placeholder: "exec:0.10,browser:0.50 (name:price, comma-sep)",
		Description: "Per-tool USDC pricing overrides in tool_name:price format",
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
		Description: "Owner's real name to prevent leaking via P2P responses",
	})

	form.AddField(&tuicore.Field{
		Key: "owner_email", Label: "Owner Email", Type: tuicore.InputText,
		Value:       cfg.P2P.OwnerProtection.OwnerEmail,
		Placeholder: "your@email.com",
		Description: "Owner's email address to redact from P2P responses",
	})

	form.AddField(&tuicore.Field{
		Key: "owner_phone", Label: "Owner Phone", Type: tuicore.InputText,
		Value:       cfg.P2P.OwnerProtection.OwnerPhone,
		Placeholder: "+82-10-1234-5678",
		Description: "Owner's phone number to redact from P2P responses",
	})

	form.AddField(&tuicore.Field{
		Key: "owner_extra_terms", Label: "Extra Terms", Type: tuicore.InputText,
		Value:       strings.Join(cfg.P2P.OwnerProtection.ExtraTerms, ","),
		Placeholder: "company-name,project-name (comma-sep)",
		Description: "Additional terms to block from P2P responses (company names, etc.)",
	})

	form.AddField(&tuicore.Field{
		Key: "owner_block_conversations", Label: "Block Conversations", Type: tuicore.InputBool,
		Checked:     derefBool(cfg.P2P.OwnerProtection.BlockConversations, true),
		Description: "Block P2P peers from accessing owner's conversation history",
	})

	return &form
}

// NewP2PSandboxForm creates the P2P Sandbox configuration form.
func NewP2PSandboxForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("P2P Sandbox Configuration")

	form.AddField(&tuicore.Field{
		Key: "sandbox_enabled", Label: "Tool Isolation Enabled", Type: tuicore.InputBool,
		Checked:     cfg.P2P.ToolIsolation.Enabled,
		Description: "Isolate P2P tool executions in sandboxed environments",
	})

	form.AddField(&tuicore.Field{
		Key: "sandbox_timeout", Label: "Timeout Per Tool", Type: tuicore.InputText,
		Value:       cfg.P2P.ToolIsolation.TimeoutPerTool.String(),
		Placeholder: "30s",
		Description: "Maximum execution time for a single sandboxed tool invocation",
	})

	form.AddField(&tuicore.Field{
		Key: "sandbox_max_memory_mb", Label: "Max Memory (MB)", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.P2P.ToolIsolation.MaxMemoryMB),
		Placeholder: "256",
		Description: "Memory limit in MB for each sandboxed tool execution",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	containerEnabled := &tuicore.Field{
		Key: "container_enabled", Label: "Container Sandbox", Type: tuicore.InputBool,
		Checked:     cfg.P2P.ToolIsolation.Container.Enabled,
		Description: "Use container-based isolation (Docker/gVisor) for stronger security",
	}
	form.AddField(containerEnabled)
	isContainerOn := func() bool { return containerEnabled.Checked }

	runtime := cfg.P2P.ToolIsolation.Container.Runtime
	if runtime == "" {
		runtime = "auto"
	}
	form.AddField(&tuicore.Field{
		Key: "container_runtime", Label: "  Runtime", Type: tuicore.InputSelect,
		Value:       runtime,
		Options:     []string{"auto", "docker", "gvisor", "native"},
		Description: "Container runtime: auto=detect best, gvisor=strongest isolation",
		VisibleWhen: isContainerOn,
	})

	form.AddField(&tuicore.Field{
		Key: "container_image", Label: "  Image", Type: tuicore.InputText,
		Value:       cfg.P2P.ToolIsolation.Container.Image,
		Placeholder: "lango-sandbox:latest",
		Description: "Docker image to use for sandboxed tool execution",
		VisibleWhen: isContainerOn,
	})

	networkMode := cfg.P2P.ToolIsolation.Container.NetworkMode
	if networkMode == "" {
		networkMode = "none"
	}
	form.AddField(&tuicore.Field{
		Key: "container_network_mode", Label: "  Network Mode", Type: tuicore.InputSelect,
		Value:       networkMode,
		Options:     []string{"none", "host", "bridge"},
		Description: "Container network: none=no network, host=full access, bridge=isolated",
		VisibleWhen: isContainerOn,
	})

	form.AddField(&tuicore.Field{
		Key: "container_readonly_rootfs", Label: "  Read-Only Rootfs", Type: tuicore.InputBool,
		Checked:     derefBool(cfg.P2P.ToolIsolation.Container.ReadOnlyRootfs, true),
		Description: "Mount container root filesystem as read-only for security",
		VisibleWhen: isContainerOn,
	})

	form.AddField(&tuicore.Field{
		Key: "container_cpu_quota", Label: "  CPU Quota (us)", Type: tuicore.InputInt,
		Value:       strconv.FormatInt(cfg.P2P.ToolIsolation.Container.CPUQuotaUS, 10),
		Placeholder: "0 (0 = unlimited)",
		Description: "CPU quota in microseconds per 100ms period; 0 = unlimited",
		VisibleWhen: isContainerOn,
		Validate: func(s string) error {
			if i, err := strconv.ParseInt(s, 10, 64); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "container_pool_size", Label: "  Pool Size", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.P2P.ToolIsolation.Container.PoolSize),
		Placeholder: "0 (0 = disabled)",
		Description: "Number of pre-warmed containers in the pool; 0 = create on demand",
		VisibleWhen: isContainerOn,
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "container_pool_idle_timeout", Label: "  Pool Idle Timeout", Type: tuicore.InputText,
		Value:       cfg.P2P.ToolIsolation.Container.PoolIdleTimeout.String(),
		Placeholder: "5m",
		Description: "Time before idle pooled containers are destroyed",
		VisibleWhen: isContainerOn,
	})

	return &form
}

// NewDBEncryptionForm creates the Security DB Encryption configuration form.
func NewDBEncryptionForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Security DB Encryption Configuration")

	form.AddField(&tuicore.Field{
		Key: "db_encryption_enabled", Label: "SQLCipher Encryption", Type: tuicore.InputBool,
		Checked:     cfg.Security.DBEncryption.Enabled,
		Description: "Encrypt the SQLite database at rest using SQLCipher",
	})

	form.AddField(&tuicore.Field{
		Key: "db_cipher_page_size", Label: "Cipher Page Size", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Security.DBEncryption.CipherPageSize),
		Placeholder: "4096",
		Description: "SQLCipher page size; must match database creation settings (default: 4096)",
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

	// Backend selector mirrors signer provider to drive field visibility.
	signerProv := cfg.Security.Signer.Provider
	if signerProv == "" {
		signerProv = "local"
	}
	backendField := &tuicore.Field{
		Key: "kms_backend", Label: "KMS Backend", Type: tuicore.InputSelect,
		Value:       signerProv,
		Options:     []string{"local", "aws-kms", "gcp-kms", "azure-kv", "pkcs11"},
		Description: "Cloud KMS or HSM backend; must match Signer Provider in Security settings",
	}
	form.AddField(backendField)

	isCloudKMS := func() bool {
		v := backendField.Value
		return v == "aws-kms" || v == "gcp-kms" || v == "azure-kv"
	}
	isAnyKMS := func() bool {
		return backendField.Value != "local"
	}

	form.AddField(&tuicore.Field{
		Key: "kms_region", Label: "Region", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.Region,
		Placeholder: "us-east-1 or us-central1",
		Description: "Cloud region for KMS API calls (AWS region or GCP location)",
		VisibleWhen: isCloudKMS,
	})

	form.AddField(&tuicore.Field{
		Key: "kms_key_id", Label: "Key ID", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.KeyID,
		Placeholder: "arn:aws:kms:... or alias/my-key",
		Description: "KMS key identifier (AWS ARN, GCP resource name, or Azure key name)",
		VisibleWhen: isAnyKMS,
	})

	form.AddField(&tuicore.Field{
		Key: "kms_endpoint", Label: "Endpoint", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.Endpoint,
		Placeholder: "http://localhost:8080 (optional)",
		Description: "Custom KMS API endpoint; leave empty for default cloud endpoints",
		VisibleWhen: isCloudKMS,
	})

	form.AddField(&tuicore.Field{
		Key: "kms_fallback_to_local", Label: "Fallback to Local", Type: tuicore.InputBool,
		Checked:     cfg.Security.KMS.FallbackToLocal,
		Description: "Fall back to local key signing if cloud KMS is unavailable",
		VisibleWhen: isAnyKMS,
	})

	form.AddField(&tuicore.Field{
		Key: "kms_timeout", Label: "Timeout Per Operation", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.TimeoutPerOperation.String(),
		Placeholder: "5s",
		Description: "Timeout for each individual KMS API call",
		VisibleWhen: isAnyKMS,
	})

	form.AddField(&tuicore.Field{
		Key: "kms_max_retries", Label: "Max Retries", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Security.KMS.MaxRetries),
		Placeholder: "3",
		Description: "Number of retry attempts for failed KMS operations",
		VisibleWhen: isAnyKMS,
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	isAzure := func() bool { return backendField.Value == "azure-kv" }
	form.AddField(&tuicore.Field{
		Key: "kms_azure_vault_url", Label: "Azure Vault URL", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.Azure.VaultURL,
		Placeholder: "https://myvault.vault.azure.net",
		Description: "Azure Key Vault URL (required for Azure backend)",
		VisibleWhen: isAzure,
	})

	form.AddField(&tuicore.Field{
		Key: "kms_azure_key_version", Label: "Azure Key Version", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.Azure.KeyVersion,
		Placeholder: "empty = latest",
		Description: "Specific key version to use; empty = always use latest version",
		VisibleWhen: isAzure,
	})

	isPKCS11 := func() bool { return backendField.Value == "pkcs11" }
	form.AddField(&tuicore.Field{
		Key: "kms_pkcs11_module", Label: "PKCS#11 Module Path", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.PKCS11.ModulePath,
		Placeholder: "/usr/lib/pkcs11/opensc-pkcs11.so",
		Description: "Path to the PKCS#11 shared library for HSM access",
		VisibleWhen: isPKCS11,
	})

	form.AddField(&tuicore.Field{
		Key: "kms_pkcs11_slot_id", Label: "PKCS#11 Slot ID", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Security.KMS.PKCS11.SlotID),
		Placeholder: "0",
		Description: "HSM slot index to use for key operations",
		VisibleWhen: isPKCS11,
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "kms_pkcs11_pin", Label: "PKCS#11 PIN", Type: tuicore.InputPassword,
		Value:       cfg.Security.KMS.PKCS11.Pin,
		Placeholder: "prefer LANGO_PKCS11_PIN env var",
		Description: "HSM PIN/password; strongly prefer LANGO_PKCS11_PIN env var for security",
		VisibleWhen: isPKCS11,
	})

	form.AddField(&tuicore.Field{
		Key: "kms_pkcs11_key_label", Label: "PKCS#11 Key Label", Type: tuicore.InputText,
		Value:       cfg.Security.KMS.PKCS11.KeyLabel,
		Placeholder: "my-signing-key",
		Description: "Label of the signing key stored in the HSM",
		VisibleWhen: isPKCS11,
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
		Value:       string(cfg.Type),
		Options:     []string{"openai", "anthropic", "gemini", "ollama", "github"},
		Description: "LLM provider type; determines API format and authentication method",
	})

	if id == "" {
		form.AddField(&tuicore.Field{
			Key: "id", Label: "Provider Name", Type: tuicore.InputText,
			Placeholder: "e.g. my-openai, production-claude",
			Description: "Unique identifier to reference this provider in other settings",
		})
	}

	form.AddField(&tuicore.Field{
		Key: "apikey", Label: "API Key", Type: tuicore.InputPassword,
		Value:       cfg.APIKey,
		Placeholder: "${ENV_VAR} or key",
		Description: "API key for authentication; use ${ENV_VAR} to reference environment variables",
	})

	form.AddField(&tuicore.Field{
		Key: "baseurl", Label: "Base URL", Type: tuicore.InputText,
		Value:       cfg.BaseURL,
		Placeholder: "https://api.example.com/v1",
		Description: "Custom API base URL; leave empty for provider's default endpoint",
	})

	return &form
}
