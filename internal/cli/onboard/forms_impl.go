package onboard

import (
	"fmt"
	"strconv"

	"github.com/langowarny/lango/internal/config"
)

// Helper to create the Agent configuration form
func NewAgentForm(cfg *config.Config) *FormModel {
	form := NewFormModel("🤖 Agent Configuration")

	form.AddField(&Field{
		Key: "provider", Label: "Provider", Type: InputSelect,
		Value:   cfg.Agent.Provider,
		Options: []string{"anthropic", "openai", "gemini", "ollama"},
	})

	form.AddField(&Field{
		Key: "model", Label: "Model ID", Type: InputText,
		Value:       cfg.Agent.Model,
		Placeholder: "e.g. claude-3-5-sonnet-20240620",
	})

	form.AddField(&Field{
		Key: "maxtokens", Label: "Max Tokens", Type: InputInt,
		Value: strconv.Itoa(cfg.Agent.MaxTokens),
		Validate: func(s string) error {
			if _, err := strconv.Atoi(s); err != nil {
				return fmt.Errorf("must be integer")
			}
			return nil
		},
	})

	form.AddField(&Field{
		Key: "temp", Label: "Temperature", Type: InputText, // Float input as text for now
		Value: fmt.Sprintf("%.1f", cfg.Agent.Temperature),
	})

	form.AddField(&Field{
		Key: "system_prompt_path", Label: "System Prompt Path", Type: InputText,
		Value:       cfg.Agent.SystemPromptPath,
		Placeholder: "path/to/system_prompt.txt",
	})

	form.AddField(&Field{
		Key: "fallback_provider", Label: "Fallback Provider", Type: InputSelect,
		Value:   cfg.Agent.FallbackProvider,
		Options: []string{"", "anthropic", "openai", "gemini", "ollama"},
	})

	form.AddField(&Field{
		Key: "fallback_model", Label: "Fallback Model", Type: InputText,
		Value:       cfg.Agent.FallbackModel,
		Placeholder: "e.g. gpt-4o",
	})

	return &form
}

// Helper to create the Server configuration form
func NewServerForm(cfg *config.Config) *FormModel {
	form := NewFormModel("🌐 Server Configuration")

	form.AddField(&Field{
		Key: "host", Label: "Host", Type: InputText,
		Value: cfg.Server.Host,
	})

	form.AddField(&Field{
		Key: "port", Label: "Port", Type: InputInt,
		Value:    strconv.Itoa(cfg.Server.Port),
		Validate: validatePort,
	})

	form.AddField(&Field{
		Key: "http", Label: "Generic HTTP", Type: InputBool,
		Checked: cfg.Server.HTTPEnabled,
	})

	form.AddField(&Field{
		Key: "ws", Label: "WebSockets", Type: InputBool,
		Checked: cfg.Server.WebSocketEnabled,
	})

	return &form
}

// Helper to create Channels configuration form
func NewChannelsForm(cfg *config.Config) *FormModel {
	form := NewFormModel("📡 Channels Configuration")

	form.AddField(&Field{
		Key: "telegram_enabled", Label: "Telegram", Type: InputBool,
		Checked: cfg.Channels.Telegram.Enabled,
	})
	form.AddField(&Field{
		Key: "telegram_token", Label: "  Bot Token", Type: InputPassword,
		Value:       cfg.Channels.Telegram.BotToken,
		Placeholder: "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
	})

	form.AddField(&Field{
		Key: "discord_enabled", Label: "Discord", Type: InputBool,
		Checked: cfg.Channels.Discord.Enabled,
	})
	form.AddField(&Field{
		Key: "discord_token", Label: "  Bot Token", Type: InputPassword,
		Value: cfg.Channels.Discord.BotToken,
	})

	form.AddField(&Field{
		Key: "slack_enabled", Label: "Slack", Type: InputBool,
		Checked: cfg.Channels.Slack.Enabled,
	})
	form.AddField(&Field{
		Key: "slack_token", Label: "  Bot Token", Type: InputPassword,
		Value:       cfg.Channels.Slack.BotToken,
		Placeholder: "xoxb-...",
	})
	form.AddField(&Field{
		Key: "slack_app_token", Label: "  App Token", Type: InputPassword,
		Value:       cfg.Channels.Slack.AppToken,
		Placeholder: "xapp-...",
	})

	return &form
}

// Helper to create Tools configuration form
func NewToolsForm(cfg *config.Config) *FormModel {
	form := NewFormModel("🛠️ Tools Configuration")

	form.AddField(&Field{
		Key: "exec_timeout", Label: "Exec Timeout", Type: InputText,
		Value:       cfg.Tools.Exec.DefaultTimeout.String(),
		Placeholder: "30s",
	})
	form.AddField(&Field{
		Key: "exec_bg", Label: "Allow Background", Type: InputBool,
		Checked: cfg.Tools.Exec.AllowBackground,
	})

	form.AddField(&Field{
		Key: "browser_enabled", Label: "Browser Enabled", Type: InputBool,
		Checked: cfg.Tools.Browser.Enabled,
	})
	form.AddField(&Field{
		Key: "browser_headless", Label: "Browser Headless", Type: InputBool,
		Checked: cfg.Tools.Browser.Headless,
	})
	form.AddField(&Field{
		Key: "browser_session_timeout", Label: "Browser Session Timeout", Type: InputText,
		Value:       cfg.Tools.Browser.SessionTimeout.String(),
		Placeholder: "5m",
	})

	form.AddField(&Field{
		Key: "fs_max_read", Label: "Max Read Size", Type: InputInt,
		Value: strconv.FormatInt(cfg.Tools.Filesystem.MaxReadSize, 10),
	})

	return &form
}

// Helper to create Security configuration form
func NewSecurityForm(cfg *config.Config) *FormModel {
	form := NewFormModel("🔒 Security Configuration")

	form.AddField(&Field{
		Key: "db_path", Label: "Session DB Path", Type: InputText,
		Value: cfg.Session.DatabasePath,
	})

	form.AddField(&Field{
		Key: "ttl", Label: "Session TTL", Type: InputText,
		Value: cfg.Session.TTL.String(),
	})

	form.AddField(&Field{
		Key: "max_history_turns", Label: "Max History Turns", Type: InputInt,
		Value: strconv.Itoa(cfg.Session.MaxHistoryTurns),
	})

	// Interceptor
	form.AddField(&Field{
		Key: "interceptor_enabled", Label: "Privacy Interceptor", Type: InputBool,
		Checked: cfg.Security.Interceptor.Enabled,
	})
	form.AddField(&Field{
		Key: "interceptor_pii", Label: "  Redact PII", Type: InputBool,
		Checked: cfg.Security.Interceptor.RedactPII,
	})
	form.AddField(&Field{
		Key: "interceptor_approval", Label: "  Approval Req.", Type: InputBool,
		Checked: cfg.Security.Interceptor.ApprovalRequired,
	})

	// Signer
	form.AddField(&Field{
		Key: "signer_provider", Label: "Signer Provider", Type: InputSelect,
		Value:   cfg.Security.Signer.Provider,
		Options: []string{"local", "rpc", "enclave"},
	})
	form.AddField(&Field{
		Key: "signer_rpc", Label: "  RPC URL", Type: InputText,
		Value:       cfg.Security.Signer.RPCUrl,
		Placeholder: "http://localhost:8080",
	})
	form.AddField(&Field{
		Key: "signer_keyid", Label: "  Key ID", Type: InputText,
		Value:       cfg.Security.Signer.KeyID,
		Placeholder: "key-123",
	})

	// Passphrase
	form.AddField(&Field{
		Key: "passphrase", Label: "Local Passphrase", Type: InputPassword,
		Value:       cfg.Security.Passphrase,
		Placeholder: "${ENV_VAR} or plaintext",
	})

	return &form
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

// Helper to create Knowledge configuration form
func NewKnowledgeForm(cfg *config.Config) *FormModel {
	form := NewFormModel("🧠 Knowledge Configuration")

	form.AddField(&Field{
		Key: "knowledge_enabled", Label: "Enabled", Type: InputBool,
		Checked: cfg.Knowledge.Enabled,
	})

	form.AddField(&Field{
		Key: "knowledge_max_learnings", Label: "Max Learnings", Type: InputInt,
		Value: strconv.Itoa(cfg.Knowledge.MaxLearnings),
	})

	form.AddField(&Field{
		Key: "knowledge_max_knowledge", Label: "Max Knowledge", Type: InputInt,
		Value: strconv.Itoa(cfg.Knowledge.MaxKnowledge),
	})

	form.AddField(&Field{
		Key: "knowledge_max_context", Label: "Max Context/Layer", Type: InputInt,
		Value: strconv.Itoa(cfg.Knowledge.MaxContextPerLayer),
	})

	form.AddField(&Field{
		Key: "knowledge_auto_approve", Label: "Auto Approve Skills", Type: InputBool,
		Checked: cfg.Knowledge.AutoApproveSkills,
	})

	form.AddField(&Field{
		Key: "knowledge_max_skills_day", Label: "Max Skills/Day", Type: InputInt,
		Value: strconv.Itoa(cfg.Knowledge.MaxSkillsPerDay),
	})

	return &form
}

// Helper to create Observational Memory configuration form
func NewObservationalMemoryForm(cfg *config.Config) *FormModel {
	form := NewFormModel("🔬 Observational Memory")

	form.AddField(&Field{
		Key: "om_enabled", Label: "Enabled", Type: InputBool,
		Checked: cfg.ObservationalMemory.Enabled,
	})

	form.AddField(&Field{
		Key: "om_provider", Label: "Provider", Type: InputText,
		Value:       cfg.ObservationalMemory.Provider,
		Placeholder: "leave empty for agent default",
	})

	form.AddField(&Field{
		Key: "om_model", Label: "Model", Type: InputText,
		Value:       cfg.ObservationalMemory.Model,
		Placeholder: "leave empty for agent default",
	})

	form.AddField(&Field{
		Key: "om_msg_threshold", Label: "Message Token Threshold",
		Type:  InputInt,
		Value: strconv.Itoa(cfg.ObservationalMemory.MessageTokenThreshold),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&Field{
		Key: "om_obs_threshold", Label: "Observation Token Threshold",
		Type:  InputInt,
		Value: strconv.Itoa(cfg.ObservationalMemory.ObservationTokenThreshold),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&Field{
		Key: "om_max_budget", Label: "Max Message Token Budget",
		Type:  InputInt,
		Value: strconv.Itoa(cfg.ObservationalMemory.MaxMessageTokenBudget),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	return &form
}

// Helper to create Provider configuration form
func NewProviderForm(id string, cfg config.ProviderConfig) *FormModel {
	title := "Edit Provider: " + id
	if id == "" {
		title = "Add New Provider"
	}
	form := NewFormModel(title)

	if id == "" {
		form.AddField(&Field{
			Key: "id", Label: "ID", Type: InputText,
			Placeholder: "unique-provider-id",
		})
	}

	form.AddField(&Field{
		Key: "type", Label: "Type", Type: InputSelect,
		Value:   cfg.Type,
		Options: []string{"openai", "anthropic", "gemini", "ollama"},
	})

	form.AddField(&Field{
		Key: "apikey", Label: "API Key", Type: InputPassword,
		Value:       cfg.APIKey,
		Placeholder: "${ENV_VAR} or key",
	})

	form.AddField(&Field{
		Key: "baseurl", Label: "Base URL", Type: InputText,
		Value:       cfg.BaseURL,
		Placeholder: "https://api.example.com/v1",
	})

	return &form
}
