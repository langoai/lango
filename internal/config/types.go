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

	// Embedding / RAG configuration
	Embedding EmbeddingConfig `mapstructure:"embedding" json:"embedding"`

	// Graph store configuration
	Graph GraphConfig `mapstructure:"graph" json:"graph"`

	// A2A protocol configuration
	A2A A2AConfig `mapstructure:"a2a" json:"a2a"`

	// Payment configuration (blockchain micropayments)
	Payment PaymentConfig `mapstructure:"payment" json:"payment"`

	// Cron scheduling configuration
	Cron CronConfig `mapstructure:"cron" json:"cron"`

	// Background task execution configuration
	Background BackgroundConfig `mapstructure:"background" json:"background"`

	// Workflow engine configuration
	Workflow WorkflowConfig `mapstructure:"workflow" json:"workflow"`

	// Skill configuration (file-based skills)
	Skill SkillConfig `mapstructure:"skill" json:"skill"`

	// Providers configuration
	Providers map[string]ProviderConfig `mapstructure:"providers" json:"providers"`
}

// CronConfig defines cron scheduling settings.
type CronConfig struct {
	// Enable the cron scheduling system.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Default timezone for cron schedules (e.g. "Asia/Seoul").
	Timezone string `mapstructure:"timezone" json:"timezone"`

	// Maximum number of concurrently executing jobs.
	MaxConcurrentJobs int `mapstructure:"maxConcurrentJobs" json:"maxConcurrentJobs"`

	// Default session mode for jobs: "isolated" or "main".
	DefaultSessionMode string `mapstructure:"defaultSessionMode" json:"defaultSessionMode"`

	// How long to retain job execution history (e.g. "30d", "720h").
	HistoryRetention string `mapstructure:"historyRetention" json:"historyRetention"`
}

// BackgroundConfig defines background task execution settings.
type BackgroundConfig struct {
	// Enable the background task system.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Time in milliseconds before an agent turn is auto-yielded to background.
	YieldMs int `mapstructure:"yieldMs" json:"yieldMs"`

	// Maximum number of concurrently running background tasks.
	MaxConcurrentTasks int `mapstructure:"maxConcurrentTasks" json:"maxConcurrentTasks"`
}

// WorkflowConfig defines workflow engine settings.
type WorkflowConfig struct {
	// Enable the workflow engine.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Maximum number of concurrently executing workflow steps.
	MaxConcurrentSteps int `mapstructure:"maxConcurrentSteps" json:"maxConcurrentSteps"`

	// Default timeout for a single workflow step (e.g. "10m").
	DefaultTimeout time.Duration `mapstructure:"defaultTimeout" json:"defaultTimeout"`

	// Directory to store workflow state for resume capability.
	StateDir string `mapstructure:"stateDir" json:"stateDir"`
}

// SkillConfig defines file-based skill settings.
type SkillConfig struct {
	// Enable the skill system.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// SkillsDir is the directory containing skill files (default: ~/.lango/skills).
	SkillsDir string `mapstructure:"skillsDir" json:"skillsDir"`
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

	// AnalysisTurnThreshold is the number of new turns before triggering conversation analysis (default: 10).
	AnalysisTurnThreshold int `mapstructure:"analysisTurnThreshold" json:"analysisTurnThreshold"`

	// AnalysisTokenThreshold is the token count before triggering conversation analysis (default: 2000).
	AnalysisTokenThreshold int `mapstructure:"analysisTokenThreshold" json:"analysisTokenThreshold"`
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

// EmbeddingConfig defines embedding and RAG settings.
type EmbeddingConfig struct {
	// ProviderID references a key in the providers map (e.g., "gemini-1", "my-openai").
	// The embedding backend type and API key are resolved from this provider.
	// For local (Ollama) embeddings, leave ProviderID empty and set Provider to "local".
	ProviderID string `mapstructure:"providerID" json:"providerID"`

	// Provider is used only for local (Ollama) embeddings where no entry in the
	// providers map is needed. Set to "local" to enable local embeddings.
	Provider string `mapstructure:"provider" json:"provider"`

	// Model is the embedding model identifier.
	Model string `mapstructure:"model" json:"model"`

	// Dimensions is the embedding vector dimensionality.
	Dimensions int `mapstructure:"dimensions" json:"dimensions"`

	// Local holds settings for the local (Ollama) provider.
	Local LocalEmbeddingConfig `mapstructure:"local" json:"local"`

	// RAG holds retrieval-augmented generation settings.
	RAG RAGConfig `mapstructure:"rag" json:"rag"`
}

// LocalEmbeddingConfig defines settings for a local embedding provider.
type LocalEmbeddingConfig struct {
	// BaseURL is the Ollama endpoint (default: http://localhost:11434/v1).
	BaseURL string `mapstructure:"baseUrl" json:"baseUrl"`
	// Model overrides the embedding model for local provider.
	Model string `mapstructure:"model" json:"model"`
}

// RAGConfig defines retrieval-augmented generation settings.
type RAGConfig struct {
	// Enabled activates RAG context injection.
	Enabled bool `mapstructure:"enabled" json:"enabled"`
	// MaxResults is the maximum number of results to inject.
	MaxResults int `mapstructure:"maxResults" json:"maxResults"`
	// Collections to search (empty means all).
	Collections []string `mapstructure:"collections" json:"collections"`
	// MaxDistance is the maximum cosine distance for RAG results (0.0 = disabled).
	MaxDistance float32 `mapstructure:"maxDistance" json:"maxDistance"`
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

// ApprovalPolicy determines which tools require approval before execution.
type ApprovalPolicy string

const (
	// ApprovalPolicyDangerous requires approval for Dangerous-level tools (default).
	ApprovalPolicyDangerous ApprovalPolicy = "dangerous"
	// ApprovalPolicyAll requires approval for all tools.
	ApprovalPolicyAll ApprovalPolicy = "all"
	// ApprovalPolicyConfigured requires approval only for explicitly listed SensitiveTools.
	ApprovalPolicyConfigured ApprovalPolicy = "configured"
	// ApprovalPolicyNone disables approval entirely.
	ApprovalPolicyNone ApprovalPolicy = "none"
)

// InterceptorConfig defines AI Privacy Interceptor settings
type InterceptorConfig struct {
	Enabled             bool           `mapstructure:"enabled" json:"enabled"`
	RedactPII           bool           `mapstructure:"redactPii" json:"redactPii"`
	ApprovalPolicy      ApprovalPolicy `mapstructure:"approvalPolicy" json:"approvalPolicy"`             // default: "dangerous"
	HeadlessAutoApprove bool           `mapstructure:"headlessAutoApprove" json:"headlessAutoApprove"`
	NotifyChannel       string         `mapstructure:"notifyChannel" json:"notifyChannel"`               // e.g. "discord", "telegram"
	SensitiveTools      []string       `mapstructure:"sensitiveTools" json:"sensitiveTools"`
	ExemptTools         []string       `mapstructure:"exemptTools" json:"exemptTools"`                   // Tools exempt from approval regardless of policy
	PIIRegexPatterns    []string       `mapstructure:"piiRegexPatterns" json:"piiRegexPatterns"`
	ApprovalTimeoutSec  int            `mapstructure:"approvalTimeoutSec" json:"approvalTimeoutSec"`     // default 30
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

	// PromptsDir is the directory containing section .md files (AGENTS.md, SAFETY.md, etc.)
	// If empty, built-in default sections are used.
	PromptsDir string `mapstructure:"promptsDir" json:"promptsDir"`

	// Fallback provider ID
	FallbackProvider string `mapstructure:"fallbackProvider" json:"fallbackProvider"`

	// Fallback model ID
	FallbackModel string `mapstructure:"fallbackModel" json:"fallbackModel"`

	// MultiAgent enables hierarchical sub-agent orchestration.
	// When false (default), a single monolithic agent handles all tasks.
	MultiAgent bool `mapstructure:"multiAgent" json:"multiAgent"`
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

	// Path to browser binary (empty = auto-detect via launcher.LookPath)
	BrowserBin string `mapstructure:"browserBin" json:"browserBin"`

	// Session timeout
	SessionTimeout time.Duration `mapstructure:"sessionTimeout" json:"sessionTimeout"`
}

// ProviderTypeToEmbeddingType maps a provider config type to the corresponding
// embedding backend type.
var ProviderTypeToEmbeddingType = map[string]string{
	"openai":    "openai",
	"gemini":    "google",
	"google":    "google",
	"anthropic": "",
	"ollama":    "local",
}

// GraphConfig defines graph store settings for relationship-aware retrieval.
type GraphConfig struct {
	// Enable the graph store.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Backend type: "bolt" (default, embedded BoltDB) or "rocksdb".
	Backend string `mapstructure:"backend" json:"backend"`

	// DatabasePath is the file path for the graph database.
	// Defaults to a "graph.db" file next to the session database.
	DatabasePath string `mapstructure:"databasePath" json:"databasePath"`

	// MaxTraversalDepth limits graph expansion depth (default: 2).
	MaxTraversalDepth int `mapstructure:"maxTraversalDepth" json:"maxTraversalDepth"`

	// MaxExpansionResults limits how many graph-expanded results to return (default: 10).
	MaxExpansionResults int `mapstructure:"maxExpansionResults" json:"maxExpansionResults"`
}

// A2AConfig defines Agent-to-Agent protocol settings.
type A2AConfig struct {
	// Enable A2A protocol support.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// BaseURL is the external URL where this agent is reachable.
	BaseURL string `mapstructure:"baseUrl" json:"baseUrl"`

	// AgentName is the name advertised in the Agent Card.
	AgentName string `mapstructure:"agentName" json:"agentName"`

	// AgentDescription is the description in the Agent Card.
	AgentDescription string `mapstructure:"agentDescription" json:"agentDescription"`

	// RemoteAgents is a list of external A2A agents to integrate as sub-agents.
	RemoteAgents []RemoteAgentConfig `mapstructure:"remoteAgents" json:"remoteAgents"`
}

// RemoteAgentConfig defines an external A2A agent to connect to.
type RemoteAgentConfig struct {
	// Name is the local name for this remote agent.
	Name string `mapstructure:"name" json:"name"`

	// AgentCardURL is the URL to fetch the agent card from.
	// Typically: https://host/.well-known/agent.json
	AgentCardURL string `mapstructure:"agentCardUrl" json:"agentCardUrl"`
}

// PaymentConfig defines blockchain payment settings.
type PaymentConfig struct {
	// Enable blockchain payment features.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// WalletProvider selects the wallet backend: "local", "rpc", or "composite".
	WalletProvider string `mapstructure:"walletProvider" json:"walletProvider"`

	// Network defines blockchain network parameters.
	Network PaymentNetworkConfig `mapstructure:"network" json:"network"`

	// Limits defines spending restrictions.
	Limits SpendingLimitsConfig `mapstructure:"limits" json:"limits"`

	// X402 defines X402 protocol interception settings.
	X402 X402Config `mapstructure:"x402" json:"x402"`
}

// PaymentNetworkConfig defines blockchain network parameters.
type PaymentNetworkConfig struct {
	// ChainID is the EVM chain ID (default: 84532 = Base Sepolia).
	ChainID int64 `mapstructure:"chainId" json:"chainId"`

	// RPCURL is the JSON-RPC endpoint for the blockchain network.
	RPCURL string `mapstructure:"rpcUrl" json:"rpcUrl"`

	// USDCContract is the USDC token contract address on the target chain.
	USDCContract string `mapstructure:"usdcContract" json:"usdcContract"`
}

// SpendingLimitsConfig defines spending restrictions for payment transactions.
type SpendingLimitsConfig struct {
	// MaxPerTx is the maximum amount per transaction in USDC (e.g. "1.00").
	MaxPerTx string `mapstructure:"maxPerTx" json:"maxPerTx"`

	// MaxDaily is the maximum daily spending in USDC (e.g. "10.00").
	MaxDaily string `mapstructure:"maxDaily" json:"maxDaily"`

	// AutoApproveBelow is the amount below which transactions are auto-approved.
	AutoApproveBelow string `mapstructure:"autoApproveBelow" json:"autoApproveBelow"`
}

// X402Config defines X402 protocol interception settings.
type X402Config struct {
	// AutoIntercept enables automatic interception of HTTP 402 responses.
	AutoIntercept bool `mapstructure:"autoIntercept" json:"autoIntercept"`

	// MaxAutoPayAmount is the maximum amount to auto-pay for X402 challenges.
	MaxAutoPayAmount string `mapstructure:"maxAutoPayAmount" json:"maxAutoPayAmount"`
}

// ResolveEmbeddingProvider returns the embedding backend type and API key
// for the configured embedding provider.
// Priority: ProviderID (from providers map) > Provider "local" (Ollama).
func (c *Config) ResolveEmbeddingProvider() (backendType, apiKey string) {
	emb := c.Embedding

	// Explicit provider ID — resolve type and key from providers map.
	if emb.ProviderID != "" {
		p, ok := c.Providers[emb.ProviderID]
		if !ok {
			return "", ""
		}
		bt := ProviderTypeToEmbeddingType[p.Type]
		if bt == "" {
			return "", ""
		}
		return bt, p.APIKey
	}

	// Local (Ollama) provider — no API key needed.
	if emb.Provider == "local" {
		return "local", ""
	}

	return "", ""
}
