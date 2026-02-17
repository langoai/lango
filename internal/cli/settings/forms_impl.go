package settings

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/langowarny/lango/internal/cli/tuicore"
	"github.com/langowarny/lango/internal/config"
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
		Placeholder: "~/.lango/prompts (directory of .md files)",
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
		Key: "db_path", Label: "Database Path", Type: tuicore.InputText,
		Value: cfg.Session.DatabasePath,
	})

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
		Options: []string{"local", "rpc", "enclave"},
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
		Options: []string{"", "telegram", "discord", "slack"},
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

// NewKnowledgeForm creates the Knowledge configuration form.
func NewKnowledgeForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Knowledge Configuration")

	form.AddField(&tuicore.Field{
		Key: "knowledge_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked: cfg.Knowledge.Enabled,
	})

	form.AddField(&tuicore.Field{
		Key: "knowledge_max_learnings", Label: "Max Learnings", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Knowledge.MaxLearnings),
	})

	form.AddField(&tuicore.Field{
		Key: "knowledge_max_knowledge", Label: "Max Knowledge", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Knowledge.MaxKnowledge),
	})

	form.AddField(&tuicore.Field{
		Key: "knowledge_max_context", Label: "Max Context/Layer", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Knowledge.MaxContextPerLayer),
	})

	form.AddField(&tuicore.Field{
		Key: "knowledge_auto_approve", Label: "Auto Approve Skills", Type: tuicore.InputBool,
		Checked: cfg.Knowledge.AutoApproveSkills,
	})

	form.AddField(&tuicore.Field{
		Key: "knowledge_max_skills_day", Label: "Max Skills/Day", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Knowledge.MaxSkillsPerDay),
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

// NewProviderForm creates a Provider configuration form.
func NewProviderForm(id string, cfg config.ProviderConfig) *tuicore.FormModel {
	title := "Edit Provider: " + id
	if id == "" {
		title = "Add New Provider"
	}
	form := tuicore.NewFormModel(title)

	form.AddField(&tuicore.Field{
		Key: "type", Label: "Type", Type: tuicore.InputSelect,
		Value:   cfg.Type,
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
