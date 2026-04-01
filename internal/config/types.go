package config

import (
	"encoding/json"
	"time"

	"github.com/langoai/lango/internal/types"
)

// Clone returns a deep copy of the Config via JSON roundtrip.
func (c *Config) Clone() *Config {
	if c == nil {
		return nil
	}
	data, _ := json.Marshal(c)
	var clone Config
	_ = json.Unmarshal(data, &clone)
	return &clone
}

// Config is the root configuration structure for lango
type Config struct {
	// DataRoot is the root directory for all lango data files (default: ~/.lango/).
	// All configurable data paths (database, graph, skills, etc.) must reside under this directory.
	// This can be overridden (e.g., for Docker: /data/lango/) but all sub-paths must stay under it.
	DataRoot string `mapstructure:"dataRoot" json:"dataRoot,omitempty"`

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

	// Hooks configuration (tool execution hooks)
	Hooks HooksConfig `mapstructure:"hooks" json:"hooks"`

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

	// Librarian configuration (proactive knowledge agent)
	Librarian LibrarianConfig `mapstructure:"librarian" json:"librarian"`

	// P2P network configuration
	P2P P2PConfig `mapstructure:"p2p" json:"p2p"`

	// Agent Memory configuration (per-agent persistent memory)
	AgentMemory AgentMemoryConfig `mapstructure:"agentMemory" json:"agentMemory"`

	// MCP server integration configuration
	MCP MCPConfig `mapstructure:"mcp" json:"mcp"`

	// Economy layer configuration (budget, risk, escrow, pricing, negotiation)
	Economy EconomyConfig `mapstructure:"economy" json:"economy"`

	// Smart Account configuration (ERC-7579 modular accounts)
	SmartAccount SmartAccountConfig `mapstructure:"smartAccount" json:"smartAccount"`

	// Retrieval coordinator configuration (agentic retrieval)
	Retrieval RetrievalConfig `mapstructure:"retrieval" json:"retrieval"`

	// Gatekeeper configuration (response sanitization)
	Gatekeeper GatekeeperConfig `mapstructure:"gatekeeper" json:"gatekeeper"`

	// Observability configuration (token tracking, health, audit, metrics)
	Observability ObservabilityConfig `mapstructure:"observability" json:"observability"`

	// Alerting configuration (operational alert thresholds and delivery)
	Alerting AlertingConfig `mapstructure:"alerting" json:"alerting"`

	// RunLedger configuration (Task OS — durable execution engine)
	RunLedger RunLedgerConfig `mapstructure:"runLedger" json:"runLedger"`

	// Provenance configuration (session checkpoints, attribution, session tree)
	Provenance ProvenanceConfig `mapstructure:"provenance" json:"provenance"`

	// Sandbox configuration (OS-level tool execution isolation)
	Sandbox SandboxConfig `mapstructure:"sandbox" json:"sandbox"`

	// Ontology subsystem configuration (typed objects and predicates)
	Ontology OntologyConfig `mapstructure:"ontology" json:"ontology,omitempty"`

	// ContextProfile selects a named preset that bundles context subsystem settings.
	// Valid values: "off", "lite", "balanced", "full", or empty (no profile).
	ContextProfile ContextProfileName `mapstructure:"contextProfile" json:"contextProfile,omitempty"`

	// Context budget configuration (advanced). contextProfile stays top-level;
	// these settings control token budget allocation across prompt sections.
	Context ContextConfig `mapstructure:"context" json:"context"`

	// Providers configuration
	Providers map[string]ProviderConfig `mapstructure:"providers" json:"providers"`
}

// ContextConfig controls token budget allocation across prompt sections.
// These are advanced settings; most users should use contextProfile instead.
type ContextConfig struct {
	// ModelWindow overrides the auto-detected model context window size (tokens).
	// 0 = auto-detect from model registry.
	ModelWindow int `mapstructure:"modelWindow" json:"modelWindow"`

	// ResponseReserve overrides the response token reserve.
	// 0 = use agent.maxTokens. Clamped to [1024, 25% of modelWindow].
	ResponseReserve int `mapstructure:"responseReserve" json:"responseReserve"`

	// Allocation controls the ratio of available tokens allocated to each section.
	// All values must sum to 1.0.
	Allocation ContextAllocationConfig `mapstructure:"allocation" json:"allocation"`
}

// ContextAllocationConfig defines per-section token allocation ratios.
type ContextAllocationConfig struct {
	Knowledge  float64 `mapstructure:"knowledge" json:"knowledge"`
	RAG        float64 `mapstructure:"rag" json:"rag"`
	Memory     float64 `mapstructure:"memory" json:"memory"`
	RunSummary float64 `mapstructure:"runSummary" json:"runSummary"`
	Headroom   float64 `mapstructure:"headroom" json:"headroom"`
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

	// RequestTimeout is the maximum duration for a single agent request (default: 5m).
	RequestTimeout time.Duration `mapstructure:"requestTimeout" json:"requestTimeout"`

	// ToolTimeout is the maximum duration for a single tool call execution (default: 2m).
	ToolTimeout time.Duration `mapstructure:"toolTimeout" json:"toolTimeout"`

	// MaxTurns limits the number of tool-calling iterations per agent run (default: 25).
	// Zero means use the default.
	MaxTurns int `mapstructure:"maxTurns" json:"maxTurns"`

	// ErrorCorrectionEnabled enables learning-based error correction (default: true).
	// When nil, defaults to true if the knowledge system is enabled.
	ErrorCorrectionEnabled *bool `mapstructure:"errorCorrectionEnabled" json:"errorCorrectionEnabled"`

	// MaxDelegationRounds limits orchestrator→sub-agent delegation rounds per turn (default: 10).
	// Zero means use the default.
	MaxDelegationRounds int `mapstructure:"maxDelegationRounds" json:"maxDelegationRounds"`

	// AutoExtendTimeout enables automatic deadline extension when agent activity is detected.
	// When true, the timeout is extended on each agent event (tool call, text chunk) up to MaxRequestTimeout.
	AutoExtendTimeout bool `mapstructure:"autoExtendTimeout" json:"autoExtendTimeout"`

	// MaxRequestTimeout is the absolute maximum duration for a request when auto-extend is enabled.
	// Defaults to 3x RequestTimeout (e.g. 15m if RequestTimeout is 5m).
	MaxRequestTimeout time.Duration `mapstructure:"maxRequestTimeout" json:"maxRequestTimeout"`

	// Orchestration configures the structured multi-agent control plane.
	// When Mode is "structured", a CoordinatingExecutor wraps the agent executor
	// to apply DelegationGuard, BudgetPolicy, and RecoveryPolicy.
	Orchestration OrchestrationConfig `mapstructure:"orchestration" json:"orchestration"`

	// IdleTimeout is the duration of inactivity (no streaming chunks, tool calls, or model responses)
	// before a request is timed out. When set, RequestTimeout becomes the hard ceiling.
	// Set to -1 to disable idle timeout and use fixed RequestTimeout instead.
	// Default: 0 (disabled for backward compatibility; new installs should set 2m).
	IdleTimeout time.Duration `mapstructure:"idleTimeout" json:"idleTimeout"`

	// AgentsDir is the directory containing user-defined AGENT.md files.
	// Structure: <dir>/<name>/AGENT.md
	// If empty, only built-in agents are used.
	AgentsDir string `mapstructure:"agentsDir" json:"agentsDir"`
}

// ProviderConfig defines AI provider settings
type ProviderConfig struct {
	// Provider type (openai, anthropic, gemini)
	Type types.ProviderType `mapstructure:"type" json:"type"`

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

// SessionConfig defines session storage settings.
// The primary database is always ~/.lango/lango.db (opened during bootstrap).
// DatabasePath is used as a fallback for standalone CLI commands.
type SessionConfig struct {
	// Database path for standalone CLI access (defaults to ~/.lango/lango.db).
	// In normal operation the bootstrap Ent client is reused instead.
	DatabasePath string `mapstructure:"databasePath" json:"databasePath"`

	// Session TTL before expiration
	TTL time.Duration `mapstructure:"ttl" json:"ttl"`

	// Maximum history turns per session
	MaxHistoryTurns int `mapstructure:"maxHistoryTurns" json:"maxHistoryTurns"`
}

// ToolsConfig defines tool-specific settings
type ToolsConfig struct {
	Exec           ExecToolConfig       `mapstructure:"exec" json:"exec"`
	Filesystem     FilesystemToolConfig `mapstructure:"filesystem" json:"filesystem"`
	Browser        BrowserToolConfig    `mapstructure:"browser" json:"browser"`
	OutputManager  OutputManagerConfig  `mapstructure:"outputManager" json:"outputManager"`
	MaxOutputChars int                  `mapstructure:"maxOutputChars" json:"maxOutputChars"`
}

// HooksConfig defines tool execution hook settings.
type HooksConfig struct {
	// Enabled activates the hook system (default: true when multi-agent is enabled).
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// SecurityFilter enables the security filter hook that blocks dangerous commands.
	SecurityFilter bool `mapstructure:"securityFilter" json:"securityFilter"`

	// AccessControl enables per-agent tool access control.
	AccessControl bool `mapstructure:"accessControl" json:"accessControl"`

	// EventPublishing enables publishing tool execution events to the event bus.
	EventPublishing bool `mapstructure:"eventPublishing" json:"eventPublishing"`

	// KnowledgeSave enables automatic knowledge saving from tool results.
	KnowledgeSave bool `mapstructure:"knowledgeSave" json:"knowledgeSave"`

	// BlockedCommands is a list of command patterns to block (security filter).
	BlockedCommands []string `mapstructure:"blockedCommands" json:"blockedCommands"`
}

// ExecToolConfig defines shell execution settings
type ExecToolConfig struct {
	// Default timeout for commands
	DefaultTimeout time.Duration `mapstructure:"defaultTimeout" json:"defaultTimeout"`

	// Allow background processes
	AllowBackground bool `mapstructure:"allowBackground" json:"allowBackground"`

	// Working directory (empty = current)
	WorkDir string `mapstructure:"workDir" json:"workDir"`

	// AdditionalProtectedPaths specifies extra paths that the exec tool
	// should block access to (in addition to the DataRoot).
	AdditionalProtectedPaths []string `mapstructure:"additionalProtectedPaths" json:"additionalProtectedPaths,omitempty"`
}

// FilesystemToolConfig defines file access settings
type FilesystemToolConfig struct {
	// Maximum file size to read
	MaxReadSize int64 `mapstructure:"maxReadSize" json:"maxReadSize"`

	// Allowed paths (empty = allow all)
	AllowedPaths []string `mapstructure:"allowedPaths" json:"allowedPaths"`
}

// AgentMemoryConfig defines agent-scoped persistent memory settings.
type AgentMemoryConfig struct {
	// Enable agent memory system
	Enabled bool `mapstructure:"enabled" json:"enabled"`
}

// RetrievalConfig controls the agentic retrieval coordinator.
type RetrievalConfig struct {
	Enabled    bool             `mapstructure:"enabled" json:"enabled"`       // Enable agentic retrieval coordinator
	Feedback   bool             `mapstructure:"feedback" json:"feedback"`     // Context injection observability
	AutoAdjust AutoAdjustConfig `mapstructure:"autoAdjust" json:"autoAdjust"` // Relevance score auto-adjustment
}

// AutoAdjustConfig controls relevance score auto-adjustment.
// Primarily affects LIKE fallback search path and coordinator merge priority.
type AutoAdjustConfig struct {
	Enabled       bool    `mapstructure:"enabled" json:"enabled"`             // Master switch (default: false)
	Mode          string  `mapstructure:"mode" json:"mode"`                   // "shadow" or "active" (default: "shadow")
	BoostDelta    float64 `mapstructure:"boostDelta" json:"boostDelta"`       // Per-injection boost (default: 0.05)
	DecayDelta    float64 `mapstructure:"decayDelta" json:"decayDelta"`       // Per-interval decay (default: 0.01)
	DecayInterval int     `mapstructure:"decayInterval" json:"decayInterval"` // Turns between global decay (default: 100)
	MinScore      float64 `mapstructure:"minScore" json:"minScore"`           // Floor (default: 0.1)
	MaxScore      float64 `mapstructure:"maxScore" json:"maxScore"`           // Cap (default: 5.0)
	WarmupTurns   int     `mapstructure:"warmupTurns" json:"warmupTurns"`     // Turns before activation (default: 50)
}

// GatekeeperConfig defines response sanitization (output gatekeeper) settings.
type GatekeeperConfig struct {
	// Master switch (nil defaults to true — enabled by default)
	Enabled *bool `mapstructure:"enabled" json:"enabled"`

	// Strip <thought>/<thinking> tags from responses
	StripThoughtTags *bool `mapstructure:"stripThoughtTags" json:"stripThoughtTags"`

	// Strip lines starting with [INTERNAL], [DEBUG], [SYSTEM], [OBSERVATION]
	StripInternalMarkers *bool `mapstructure:"stripInternalMarkers" json:"stripInternalMarkers"`

	// Replace large raw JSON code blocks with a placeholder
	StripRawJSON *bool `mapstructure:"stripRawJSON" json:"stripRawJSON"`

	// Character threshold for raw JSON replacement (default: 500)
	RawJSONThreshold int `mapstructure:"rawJsonThreshold" json:"rawJsonThreshold"`

	// Additional regex patterns to strip from responses
	CustomPatterns []string `mapstructure:"customPatterns" json:"customPatterns"`
}

// OutputManagerConfig defines token-based output management settings.
type OutputManagerConfig struct {
	// Master switch (nil defaults to true — enabled by default)
	Enabled *bool `mapstructure:"enabled" json:"enabled"`

	// Maximum token budget for tool output (default: 2000)
	TokenBudget int `mapstructure:"tokenBudget" json:"tokenBudget"`

	// Ratio of head content to preserve during compression (default: 0.7)
	HeadRatio float64 `mapstructure:"headRatio" json:"headRatio"`

	// Ratio of tail content to preserve during compression (default: 0.3)
	TailRatio float64 `mapstructure:"tailRatio" json:"tailRatio"`
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

// AlertingConfig defines operational alerting thresholds and delivery settings.
type AlertingConfig struct {
	// Master switch (default: false)
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// AdminChannel routes alerts to a configured channel (e.g., "slack", "discord")
	AdminChannel string `mapstructure:"adminChannel" json:"adminChannel"`

	// PolicyBlockRate is the threshold for policy block events per 5min window (default: 10)
	PolicyBlockRate int `mapstructure:"policyBlockRateThreshold" json:"policyBlockRateThreshold"`

	// RecoveryRetries is the threshold for recovery retry events per session (default: 5)
	RecoveryRetries int `mapstructure:"recoveryRetryThreshold" json:"recoveryRetryThreshold"`
}
