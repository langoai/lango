package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var envVarRegex = regexp.MustCompile(`\$\{([^}]+)\}`)

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:             "localhost",
			Port:             18789,
			HTTPEnabled:      true,
			WebSocketEnabled: true,
		},
		Agent: AgentConfig{
			Provider:    "anthropic",
			Model:       "claude-sonnet-4-20250514",
			MaxTokens:   4096,
			Temperature: 0.7,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "console",
		},
		Session: SessionConfig{
			DatabasePath:    "~/.lango/data.db",
			TTL:             24 * time.Hour,
			MaxHistoryTurns: 50,
		},
		Tools: ToolsConfig{
			Exec: ExecToolConfig{
				DefaultTimeout:  30 * time.Second,
				AllowBackground: true,
			},
			Filesystem: FilesystemToolConfig{
				MaxReadSize: 10 * 1024 * 1024, // 10MB
			},
			Browser: BrowserToolConfig{
				Enabled:        false,
				Headless:       true,
				SessionTimeout: 5 * time.Minute,
			},
		},
		Security: SecurityConfig{
			Interceptor: InterceptorConfig{
				Enabled:        true,
				ApprovalPolicy: ApprovalPolicyDangerous,
			},
		},
		Knowledge: KnowledgeConfig{
			Enabled:            false,
			MaxLearnings:       10,
			MaxKnowledge:       20,
			MaxContextPerLayer: 5,
			MaxSkillsPerDay:    5,
		},
	}
}

// Load reads configuration from file and environment
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults from DefaultConfig
	defaults := DefaultConfig()
	v.SetDefault("server.host", defaults.Server.Host)
	v.SetDefault("server.port", defaults.Server.Port)
	v.SetDefault("server.httpEnabled", defaults.Server.HTTPEnabled)
	v.SetDefault("server.wsEnabled", defaults.Server.WebSocketEnabled)
	v.SetDefault("agent.provider", defaults.Agent.Provider)
	v.SetDefault("agent.model", defaults.Agent.Model)
	v.SetDefault("agent.maxTokens", defaults.Agent.MaxTokens)
	v.SetDefault("agent.temperature", defaults.Agent.Temperature)
	v.SetDefault("logging.level", defaults.Logging.Level)
	v.SetDefault("logging.format", defaults.Logging.Format)
	v.SetDefault("session.databasePath", defaults.Session.DatabasePath)
	v.SetDefault("session.ttl", defaults.Session.TTL)
	v.SetDefault("session.maxHistoryTurns", defaults.Session.MaxHistoryTurns)
	v.SetDefault("tools.exec.defaultTimeout", defaults.Tools.Exec.DefaultTimeout)
	v.SetDefault("tools.exec.allowBackground", defaults.Tools.Exec.AllowBackground)
	v.SetDefault("tools.filesystem.maxReadSize", defaults.Tools.Filesystem.MaxReadSize)
	v.SetDefault("tools.browser.enabled", defaults.Tools.Browser.Enabled)
	v.SetDefault("tools.browser.headless", defaults.Tools.Browser.Headless)
	v.SetDefault("tools.browser.sessionTimeout", defaults.Tools.Browser.SessionTimeout)
	v.SetDefault("security.interceptor.enabled", defaults.Security.Interceptor.Enabled)
	v.SetDefault("security.interceptor.approvalPolicy", string(defaults.Security.Interceptor.ApprovalPolicy))

	// Configure viper
	v.SetConfigType("json")
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.lango")
	v.AddConfigPath("/etc/lango")

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("lango")
	}

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
		// Config file not found, use defaults
	}

	// Unmarshal into struct
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Migrate legacy approval settings to ApprovalPolicy
	migrateApprovalPolicy(cfg)

	// Apply environment variable substitution
	substituteEnvVars(cfg)

	// Validate configuration
	if err := Validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// migrateApprovalPolicy migrates legacy approvalRequired + sensitiveTools
// configuration to the new ApprovalPolicy field.
func migrateApprovalPolicy(cfg *Config) {
	ic := &cfg.Security.Interceptor

	// If ApprovalPolicy is already explicitly set, keep it.
	if ic.ApprovalPolicy != "" {
		return
	}

	// Migrate from legacy fields.
	if ic.ApprovalRequired {
		if len(ic.SensitiveTools) > 0 {
			ic.ApprovalPolicy = ApprovalPolicyConfigured
		} else {
			ic.ApprovalPolicy = ApprovalPolicyDangerous
		}
		return
	}

	// ApprovalRequired was explicitly set to false (or defaulted to false
	// in an existing config). Respect the explicit opt-out for existing users.
	// New installations get "dangerous" via DefaultConfig / SetDefault.
}

// substituteEnvVars replaces ${VAR} patterns with environment variable values
func substituteEnvVars(cfg *Config) {
	// Provider credentials
	for id, pCfg := range cfg.Providers {
		pCfg.APIKey = expandEnvVars(pCfg.APIKey)
		cfg.Providers[id] = pCfg
	}

	// Channel tokens
	cfg.Channels.Telegram.BotToken = expandEnvVars(cfg.Channels.Telegram.BotToken)
	cfg.Channels.Discord.BotToken = expandEnvVars(cfg.Channels.Discord.BotToken)
	cfg.Channels.Slack.BotToken = expandEnvVars(cfg.Channels.Slack.BotToken)
	cfg.Channels.Slack.AppToken = expandEnvVars(cfg.Channels.Slack.AppToken)
	cfg.Channels.Slack.SigningSecret = expandEnvVars(cfg.Channels.Slack.SigningSecret)

	// Auth OIDC provider credentials
	for id, aCfg := range cfg.Auth.Providers {
		aCfg.ClientID = expandEnvVars(aCfg.ClientID)
		aCfg.ClientSecret = expandEnvVars(aCfg.ClientSecret)
		cfg.Auth.Providers[id] = aCfg
	}

	// Paths
	cfg.Session.DatabasePath = expandEnvVars(cfg.Session.DatabasePath)
}

func expandEnvVars(s string) string {
	return envVarRegex.ReplaceAllStringFunc(s, func(match string) string {
		varName := strings.TrimSuffix(strings.TrimPrefix(match, "${"), "}")
		if val := os.Getenv(varName); val != "" {
			return val
		}
		return match // Keep original if not found
	})
}

// Validate checks if the configuration is valid
func Validate(cfg *Config) error {
	var errs []string

	// Validate server config
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		errs = append(errs, fmt.Sprintf("invalid port: %d (must be 1-65535)", cfg.Server.Port))
	}

	// Validate agent.provider references an existing key in providers map
	if cfg.Agent.Provider != "" && len(cfg.Providers) > 0 {
		if _, ok := cfg.Providers[cfg.Agent.Provider]; !ok {
			errs = append(errs, fmt.Sprintf("agent.provider %q not found in providers map (available: %v)", cfg.Agent.Provider, providerKeys(cfg.Providers)))
		}
	}

	// Validate logging config
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[cfg.Logging.Level] {
		errs = append(errs, fmt.Sprintf("invalid log level: %s (must be debug, info, warn, or error)", cfg.Logging.Level))
	}

	validFormats := map[string]bool{"json": true, "console": true}
	if !validFormats[cfg.Logging.Format] {
		errs = append(errs, fmt.Sprintf("invalid log format: %s (must be json or console)", cfg.Logging.Format))
	}

	// Validate security config
	if cfg.Security.Signer.Provider != "" {
		validProviders := map[string]bool{"local": true, "rpc": true, "enclave": true}
		if !validProviders[cfg.Security.Signer.Provider] {
			errs = append(errs, fmt.Sprintf("invalid security.signer.provider: %q (must be local, rpc, or enclave)", cfg.Security.Signer.Provider))
		}
		if cfg.Security.Signer.Provider == "rpc" && cfg.Security.Signer.RPCUrl == "" {
			errs = append(errs, "security.signer.rpcUrl is required when provider is 'rpc'")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}

func providerKeys(providers map[string]ProviderConfig) []string {
	keys := make([]string, 0, len(providers))
	for k := range providers {
		keys = append(keys, k)
	}
	return keys
}

