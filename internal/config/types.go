package config

import "time"

// Config is the root configuration structure for lango
type Config struct {
	// Server configuration
	Server ServerConfig `mapstructure:"server" json:"server"`

	// Agent configuration
	Agent AgentConfig `mapstructure:"agent" json:"agent"`

	// Channel configurations
	Channels ChannelsConfig `mapstructure:"channels" json:"channels"`

	// Logging configuration
	Logging LoggingConfig `mapstructure:"logging" json:"logging"`

	// Session configuration
	Session SessionConfig `mapstructure:"session" json:"session"`

	// Tools configuration
	Tools ToolsConfig `mapstructure:"tools" json:"tools"`

	// Auth configuration
	Auth AuthConfig `mapstructure:"auth" json:"auth"`

	// Security configuration
	Security SecurityConfig `mapstructure:"security" json:"security"`

	// Knowledge configuration
	Knowledge KnowledgeConfig `mapstructure:"knowledge" json:"knowledge"`

	// Observational Memory configuration
	ObservationalMemory ObservationalMemoryConfig `mapstructure:"observationalMemory" json:"observationalMemory"`

	// Providers configuration
	Providers map[string]ProviderConfig `mapstructure:"providers" json:"providers"`
}

// KnowledgeConfig defines self-learning knowledge system settings
type KnowledgeConfig struct {
	// Enable the knowledge/learning system
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Maximum learning entries per session
	MaxLearnings int `mapstructure:"maxLearnings" json:"maxLearnings"`

	// Maximum knowledge entries per session
	MaxKnowledge int `mapstructure:"maxKnowledge" json:"maxKnowledge"`

	// Maximum context items per layer in retrieval
	MaxContextPerLayer int `mapstructure:"maxContextPerLayer" json:"maxContextPerLayer"`

	// Auto-approve new skills without human review
	AutoApproveSkills bool `mapstructure:"autoApproveSkills" json:"autoApproveSkills"`

	// Maximum new skills per day
	MaxSkillsPerDay int `mapstructure:"maxSkillsPerDay" json:"maxSkillsPerDay"`
}

// ObservationalMemoryConfig defines Observational Memory settings
type ObservationalMemoryConfig struct {
	// Enable the observational memory system
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// LLM provider for observer/reflector (empty = use agent default)
	Provider string `mapstructure:"provider" json:"provider"`

	// Model ID for observer/reflector (empty = use agent default)
	Model string `mapstructure:"model" json:"model"`

	// Token threshold to trigger observation (default: 1000)
	MessageTokenThreshold int `mapstructure:"messageTokenThreshold" json:"messageTokenThreshold"`

	// Token threshold to trigger reflection (default: 2000)
	ObservationTokenThreshold int `mapstructure:"observationTokenThreshold" json:"observationTokenThreshold"`

	// Max token budget for recent messages in context (default: 8000)
	MaxMessageTokenBudget int `mapstructure:"maxMessageTokenBudget" json:"maxMessageTokenBudget"`
}

// AuthConfig defines authentication settings
type AuthConfig struct {
	// OIDC Providers
	Providers map[string]OIDCProviderConfig `mapstructure:"providers" json:"providers"`
}

// OIDCProviderConfig defines a single OIDC provider
type OIDCProviderConfig struct {
	IssuerURL    string   `mapstructure:"issuerUrl" json:"issuerUrl"`
	ClientID     string   `mapstructure:"clientId" json:"clientId"`
	ClientSecret string   `mapstructure:"clientSecret" json:"clientSecret"`
	RedirectURL  string   `mapstructure:"redirectUrl" json:"redirectUrl"`
	Scopes       []string `mapstructure:"scopes" json:"scopes"`
}

// SecurityConfig defines security settings
type SecurityConfig struct {
	// Interceptor configuration
	Interceptor InterceptorConfig `mapstructure:"interceptor" json:"interceptor"`
	// Signer configuration
	Signer SignerConfig `mapstructure:"signer" json:"signer"`
}

// InterceptorConfig defines AI Privacy Interceptor settings
type InterceptorConfig struct {
	Enabled            bool     `mapstructure:"enabled" json:"enabled"`
	RedactPII          bool     `mapstructure:"redactPii" json:"redactPii"`
	ApprovalRequired   bool     `mapstructure:"approvalRequired" json:"approvalRequired"`
	NotifyChannel      string   `mapstructure:"notifyChannel" json:"notifyChannel"` // e.g. "discord", "telegram"
	SensitiveTools     []string `mapstructure:"sensitiveTools" json:"sensitiveTools"`
	PIIRegexPatterns   []string `mapstructure:"piiRegexPatterns" json:"piiRegexPatterns"`
	ApprovalTimeoutSec int      `mapstructure:"approvalTimeoutSec" json:"approvalTimeoutSec"` // default 30
}

// SignerConfig defines Secure Signer settings
type SignerConfig struct {
	Provider string `mapstructure:"provider" json:"provider"` // "local", "rpc", "enclave"
	RPCUrl   string `mapstructure:"rpcUrl" json:"rpcUrl"`     // for RPC provider
	KeyID    string `mapstructure:"keyId" json:"keyId"`       // Key identifier
}

// ServerConfig defines gateway server settings
type ServerConfig struct {
	// Host to bind to (default: "localhost")
	Host string `mapstructure:"host" json:"host"`

	// Port to listen on (default: 18789)
	Port int `mapstructure:"port" json:"port"`

	// Enable HTTP API endpoints
	HTTPEnabled bool `mapstructure:"httpEnabled" json:"httpEnabled"`

	// Enable WebSocket server
	WebSocketEnabled bool `mapstructure:"wsEnabled" json:"wsEnabled"`

	// Allowed origins for WebSocket CORS (empty = same-origin, ["*"] = allow all)
	AllowedOrigins []string `mapstructure:"allowedOrigins" json:"allowedOrigins"`
}

// AgentConfig defines LLM agent settings
type AgentConfig struct {
	// Default model provider (anthropic, openai, google)
	Provider string `mapstructure:"provider" json:"provider"`

	// Model ID to use
	Model string `mapstructure:"model" json:"model"`

	// Maximum tokens for context window
	MaxTokens int `mapstructure:"maxTokens" json:"maxTokens"`

	// Temperature for generation
	Temperature float64 `mapstructure:"temperature" json:"temperature"`

	// System prompt template path
	SystemPromptPath string `mapstructure:"systemPromptPath" json:"systemPromptPath"`

	// Fallback provider ID
	FallbackProvider string `mapstructure:"fallbackProvider" json:"fallbackProvider"`

	// Fallback model ID
	FallbackModel string `mapstructure:"fallbackModel" json:"fallbackModel"`
}

// ProviderConfig defines AI provider settings
type ProviderConfig struct {
	// Provider type (openai, anthropic, gemini)
	Type string `mapstructure:"type" json:"type"`

	// API key for the provider (supports ${ENV_VAR} substitution)
	APIKey string `mapstructure:"apiKey" json:"apiKey"`

	// Base URL for OpenAI-compatible providers
	BaseURL string `mapstructure:"baseUrl" json:"baseUrl"`

}

// ChannelsConfig holds all channel configurations
type ChannelsConfig struct {
	Telegram TelegramConfig `mapstructure:"telegram" json:"telegram"`
	Discord  DiscordConfig  `mapstructure:"discord" json:"discord"`
	Slack    SlackConfig    `mapstructure:"slack" json:"slack"`
}

// TelegramConfig defines Telegram bot settings
type TelegramConfig struct {
	// Enable Telegram channel
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Bot token from BotFather
	BotToken string `mapstructure:"botToken" json:"botToken"`

	// Allowed user/group IDs (empty = allow all)
	Allowlist []int64 `mapstructure:"allowlist" json:"allowlist"`

}

// DiscordConfig defines Discord bot settings
type DiscordConfig struct {
	// Enable Discord channel
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Bot token from Discord Developer Portal
	BotToken string `mapstructure:"botToken" json:"botToken"`

	// Application ID for slash commands
	ApplicationID string `mapstructure:"applicationId" json:"applicationId"`

	// Allowed guild IDs (empty = allow all)
	AllowedGuilds []string `mapstructure:"allowedGuilds" json:"allowedGuilds"`
}

// SlackConfig defines Slack app settings
type SlackConfig struct {
	// Enable Slack channel
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Bot OAuth token
	BotToken string `mapstructure:"botToken" json:"botToken"`

	// App-level token for Socket Mode
	AppToken string `mapstructure:"appToken" json:"appToken"`

	// Signing secret for request verification
	SigningSecret string `mapstructure:"signingSecret" json:"signingSecret"`
}

// LoggingConfig defines logging settings
type LoggingConfig struct {
	// Log level (debug, info, warn, error)
	Level string `mapstructure:"level" json:"level"`

	// Output format (json, console)
	Format string `mapstructure:"format" json:"format"`

	// Output file path (empty = stdout)
	OutputPath string `mapstructure:"outputPath" json:"outputPath"`
}

// SessionConfig defines session storage settings
type SessionConfig struct {
	// Database path (SQLite)
	DatabasePath string `mapstructure:"databasePath" json:"databasePath"`

	// Session TTL before expiration
	TTL time.Duration `mapstructure:"ttl" json:"ttl"`

	// Maximum history turns per session
	MaxHistoryTurns int `mapstructure:"maxHistoryTurns" json:"maxHistoryTurns"`
}

// ToolsConfig defines tool-specific settings
type ToolsConfig struct {
	Exec       ExecToolConfig       `mapstructure:"exec" json:"exec"`
	Filesystem FilesystemToolConfig `mapstructure:"filesystem" json:"filesystem"`
	Browser    BrowserToolConfig    `mapstructure:"browser" json:"browser"`
}

// ExecToolConfig defines shell execution settings
type ExecToolConfig struct {
	// Default timeout for commands
	DefaultTimeout time.Duration `mapstructure:"defaultTimeout" json:"defaultTimeout"`

	// Allow background processes
	AllowBackground bool `mapstructure:"allowBackground" json:"allowBackground"`

	// Working directory (empty = current)
	WorkDir string `mapstructure:"workDir" json:"workDir"`
}

// FilesystemToolConfig defines file access settings
type FilesystemToolConfig struct {
	// Maximum file size to read
	MaxReadSize int64 `mapstructure:"maxReadSize" json:"maxReadSize"`

	// Allowed paths (empty = allow all)
	AllowedPaths []string `mapstructure:"allowedPaths" json:"allowedPaths"`
}

// BrowserToolConfig defines browser automation settings
type BrowserToolConfig struct {
	// Enable browser tools (requires Chromium)
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Run headless
	Headless bool `mapstructure:"headless" json:"headless"`

	// Session timeout
	SessionTimeout time.Duration `mapstructure:"sessionTimeout" json:"sessionTimeout"`
}
