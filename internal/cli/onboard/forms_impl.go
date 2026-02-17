package onboard

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/langowarny/lango/internal/config"
)

// buildProviderOptions builds provider options from registered providers
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

// Helper to create the Agent configuration form
func NewAgentForm(cfg *config.Config) *FormModel {
	form := NewFormModel("🤖 Agent Configuration")

	providerOpts := buildProviderOptions(cfg)
	form.AddField(&Field{
		Key: "provider", Label: "Provider", Type: InputSelect,
		Value:   cfg.Agent.Provider,
		Options: providerOpts,
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
		Key: "prompts_dir", Label: "Prompts Directory", Type: InputText,
		Value:       cfg.Agent.PromptsDir,
		Placeholder: "~/.lango/prompts (directory of .md files)",
	})

	fallbackOpts := append([]string{""}, providerOpts...)
	form.AddField(&Field{
		Key: "fallback_provider", Label: "Fallback Provider", Type: InputSelect,
		Value:   cfg.Agent.FallbackProvider,
		Options: fallbackOpts,
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

// Helper to create Session configuration form
func NewSessionForm(cfg *config.Config) *FormModel {
	form := NewFormModel("📂 Session Configuration")

	form.AddField(&Field{
		Key: "db_path", Label: "Database Path", Type: InputText,
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

	return &form
}

// Helper to create Security configuration form
func NewSecurityForm(cfg *config.Config) *FormModel {
	form := NewFormModel("🔒 Security Configuration")

	// Interceptor
	form.AddField(&Field{
		Key: "interceptor_enabled", Label: "Privacy Interceptor", Type: InputBool,
		Checked: cfg.Security.Interceptor.Enabled,
	})
	form.AddField(&Field{
		Key: "interceptor_pii", Label: "  Redact PII", Type: InputBool,
		Checked: cfg.Security.Interceptor.RedactPII,
	})
	policyVal := string(cfg.Security.Interceptor.ApprovalPolicy)
	if policyVal == "" {
		policyVal = "dangerous"
	}
	form.AddField(&Field{
		Key: "interceptor_policy", Label: "  Approval Policy", Type: InputSelect,
		Value:   policyVal,
		Options: []string{"dangerous", "all", "configured", "none"},
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

	// Approval Timeout
	form.AddField(&Field{
		Key: "interceptor_timeout", Label: "  Approval Timeout (s)", Type: InputInt,
		Value: strconv.Itoa(cfg.Security.Interceptor.ApprovalTimeoutSec),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	// Notify Channel
	form.AddField(&Field{
		Key: "interceptor_notify", Label: "  Notify Channel", Type: InputSelect,
		Value:   cfg.Security.Interceptor.NotifyChannel,
		Options: []string{"", "telegram", "discord", "slack"},
	})

	// Sensitive Tools
	form.AddField(&Field{
		Key: "interceptor_sensitive_tools", Label: "  Sensitive Tools", Type: InputText,
		Value:       strings.Join(cfg.Security.Interceptor.SensitiveTools, ","),
		Placeholder: "exec,browser (comma-separated)",
	})

	// Exempt Tools
	form.AddField(&Field{
		Key: "interceptor_exempt_tools", Label: "  Exempt Tools", Type: InputText,
		Value:       strings.Join(cfg.Security.Interceptor.ExemptTools, ","),
		Placeholder: "filesystem (comma-separated)",
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

	omProviderOpts := append([]string{""}, buildProviderOptions(cfg)...)
	form.AddField(&Field{
		Key: "om_provider", Label: "Provider", Type: InputSelect,
		Value:   cfg.ObservationalMemory.Provider,
		Options: omProviderOpts,
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

// Helper to create Embedding & RAG configuration form
func NewEmbeddingForm(cfg *config.Config) *FormModel {
	form := NewFormModel("🔗 Embedding & RAG Configuration")

	// Build provider options from user's registered providers + "local".
	providerOpts := []string{"local"}
	for id := range cfg.Providers {
		providerOpts = append(providerOpts, id)
	}
	sort.Strings(providerOpts)

	// Determine current selection: prefer ProviderID, fallback to "local" if Provider is "local".
	currentVal := cfg.Embedding.ProviderID
	if currentVal == "" && cfg.Embedding.Provider == "local" {
		currentVal = "local"
	}

	form.AddField(&Field{
		Key: "emb_provider_id", Label: "Provider", Type: InputSelect,
		Value:   currentVal,
		Options: providerOpts,
	})

	form.AddField(&Field{
		Key: "emb_model", Label: "Model", Type: InputText,
		Value:       cfg.Embedding.Model,
		Placeholder: "e.g. text-embedding-3-small",
	})

	form.AddField(&Field{
		Key: "emb_dimensions", Label: "Dimensions", Type: InputInt,
		Value: strconv.Itoa(cfg.Embedding.Dimensions),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&Field{
		Key: "emb_local_baseurl", Label: "Local Base URL", Type: InputText,
		Value:       cfg.Embedding.Local.BaseURL,
		Placeholder: "http://localhost:11434/v1",
	})

	form.AddField(&Field{
		Key: "emb_rag_enabled", Label: "RAG Enabled", Type: InputBool,
		Checked: cfg.Embedding.RAG.Enabled,
	})

	form.AddField(&Field{
		Key: "emb_rag_max_results", Label: "RAG Max Results", Type: InputInt,
		Value: strconv.Itoa(cfg.Embedding.RAG.MaxResults),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&Field{
		Key: "emb_rag_collections", Label: "RAG Collections", Type: InputText,
		Value:       strings.Join(cfg.Embedding.RAG.Collections, ","),
		Placeholder: "collection1,collection2 (comma-separated)",
	})

	return &form
}

// Helper to create OIDC Provider configuration form
func NewOIDCProviderForm(id string, cfg config.OIDCProviderConfig) *FormModel {
	title := "Edit OIDC Provider: " + id
	if id == "" {
		title = "Add New OIDC Provider"
	}
	form := NewFormModel(title)

	if id == "" {
		form.AddField(&Field{
			Key: "oidc_id", Label: "Provider Name", Type: InputText,
			Placeholder: "e.g. google, github",
		})
	}

	form.AddField(&Field{
		Key: "oidc_issuer", Label: "Issuer URL", Type: InputText,
		Value:       cfg.IssuerURL,
		Placeholder: "https://accounts.google.com",
	})

	form.AddField(&Field{
		Key: "oidc_client_id", Label: "Client ID", Type: InputPassword,
		Value:       cfg.ClientID,
		Placeholder: "${ENV_VAR} or value",
	})

	form.AddField(&Field{
		Key: "oidc_client_secret", Label: "Client Secret", Type: InputPassword,
		Value:       cfg.ClientSecret,
		Placeholder: "${ENV_VAR} or value",
	})

	form.AddField(&Field{
		Key: "oidc_redirect", Label: "Redirect URL", Type: InputText,
		Value:       cfg.RedirectURL,
		Placeholder: "http://localhost:18789/auth/callback/<name>",
	})

	form.AddField(&Field{
		Key: "oidc_scopes", Label: "Scopes", Type: InputText,
		Value:       strings.Join(cfg.Scopes, ","),
		Placeholder: "openid,email,profile",
	})

	return &form
}

// NewGraphForm creates the Graph Store configuration form.
func NewGraphForm(cfg *config.Config) *FormModel {
	form := NewFormModel("📊 Graph Store Configuration")

	form.AddField(&Field{
		Key: "graph_enabled", Label: "Enabled", Type: InputBool,
		Checked: cfg.Graph.Enabled,
	})

	form.AddField(&Field{
		Key: "graph_backend", Label: "Backend", Type: InputSelect,
		Value:   cfg.Graph.Backend,
		Options: []string{"bolt"},
	})

	form.AddField(&Field{
		Key: "graph_db_path", Label: "Database Path", Type: InputText,
		Value:       cfg.Graph.DatabasePath,
		Placeholder: "~/.lango/graph.db",
	})

	form.AddField(&Field{
		Key: "graph_max_depth", Label: "Max Traversal Depth", Type: InputInt,
		Value: strconv.Itoa(cfg.Graph.MaxTraversalDepth),
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&Field{
		Key: "graph_max_expand", Label: "Max Expansion Results", Type: InputInt,
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
func NewMultiAgentForm(cfg *config.Config) *FormModel {
	form := NewFormModel("🔀 Multi-Agent Configuration")

	form.AddField(&Field{
		Key: "multi_agent", Label: "Enable Multi-Agent Orchestration", Type: InputBool,
		Checked: cfg.Agent.MultiAgent,
	})

	return &form
}

// NewA2AForm creates the A2A Protocol configuration form.
func NewA2AForm(cfg *config.Config) *FormModel {
	form := NewFormModel("🌍 A2A Protocol Configuration")

	form.AddField(&Field{
		Key: "a2a_enabled", Label: "Enabled", Type: InputBool,
		Checked: cfg.A2A.Enabled,
	})

	form.AddField(&Field{
		Key: "a2a_base_url", Label: "Base URL", Type: InputText,
		Value:       cfg.A2A.BaseURL,
		Placeholder: "https://your-agent.example.com",
	})

	form.AddField(&Field{
		Key: "a2a_agent_name", Label: "Agent Name", Type: InputText,
		Value:       cfg.A2A.AgentName,
		Placeholder: "my-lango-agent",
	})

	form.AddField(&Field{
		Key: "a2a_agent_desc", Label: "Agent Description", Type: InputText,
		Value:       cfg.A2A.AgentDescription,
		Placeholder: "A helpful AI assistant",
	})

	return &form
}

// NewPaymentForm creates the Payment configuration form.
func NewPaymentForm(cfg *config.Config) *FormModel {
	form := NewFormModel("💸 Payment Configuration")

	form.AddField(&Field{
		Key: "payment_enabled", Label: "Enabled", Type: InputBool,
		Checked: cfg.Payment.Enabled,
	})

	form.AddField(&Field{
		Key: "payment_wallet_provider", Label: "Wallet Provider", Type: InputSelect,
		Value:   cfg.Payment.WalletProvider,
		Options: []string{"local", "rpc", "composite"},
	})

	form.AddField(&Field{
		Key: "payment_chain_id", Label: "Chain ID", Type: InputInt,
		Value: strconv.FormatInt(cfg.Payment.Network.ChainID, 10),
		Validate: func(s string) error {
			if _, err := strconv.ParseInt(s, 10, 64); err != nil {
				return fmt.Errorf("must be an integer")
			}
			return nil
		},
	})

	form.AddField(&Field{
		Key: "payment_rpc_url", Label: "RPC URL", Type: InputText,
		Value:       cfg.Payment.Network.RPCURL,
		Placeholder: "https://sepolia.base.org",
	})

	form.AddField(&Field{
		Key: "payment_usdc_contract", Label: "USDC Contract", Type: InputText,
		Value:       cfg.Payment.Network.USDCContract,
		Placeholder: "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
	})

	form.AddField(&Field{
		Key: "payment_max_per_tx", Label: "Max Per Transaction (USDC)", Type: InputText,
		Value:       cfg.Payment.Limits.MaxPerTx,
		Placeholder: "1.00",
	})

	form.AddField(&Field{
		Key: "payment_max_daily", Label: "Max Daily (USDC)", Type: InputText,
		Value:       cfg.Payment.Limits.MaxDaily,
		Placeholder: "10.00",
	})

	form.AddField(&Field{
		Key: "payment_auto_approve", Label: "Auto-Approve Below (USDC)", Type: InputText,
		Value:       cfg.Payment.Limits.AutoApproveBelow,
		Placeholder: "0.10",
	})

	form.AddField(&Field{
		Key: "payment_x402_auto", Label: "X402 Auto-Intercept", Type: InputBool,
		Checked: cfg.Payment.X402.AutoIntercept,
	})

	form.AddField(&Field{
		Key: "payment_x402_max", Label: "X402 Max Auto-Pay (USDC)", Type: InputText,
		Value:       cfg.Payment.X402.MaxAutoPayAmount,
		Placeholder: "0.50",
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

	form.AddField(&Field{
		Key: "type", Label: "Type", Type: InputSelect,
		Value:   cfg.Type,
		Options: []string{"openai", "anthropic", "gemini", "ollama"},
	})

	if id == "" {
		form.AddField(&Field{
			Key: "id", Label: "Provider Name", Type: InputText,
			Placeholder: "e.g. my-openai, production-claude",
		})
	}

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
