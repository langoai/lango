package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/langoai/lango/internal/types"
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
			Provider:       "anthropic",
			Model:          "",
			MaxTokens:      4096,
			Temperature:    0.7,
			RequestTimeout: 5 * time.Minute,
			ToolTimeout:    2 * time.Minute,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "console",
		},
		Session: SessionConfig{
			DatabasePath:    "~/.lango/lango.db",
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
			Keyring: KeyringConfig{
				Enabled: true,
			},
			DBEncryption: DBEncryptionConfig{
				Enabled:        false,
				CipherPageSize: 4096,
			},
			KMS: KMSConfig{
				FallbackToLocal:     true,
				TimeoutPerOperation: 5 * time.Second,
				MaxRetries:          3,
			},
		},
		Knowledge: KnowledgeConfig{
			Enabled:            false,
			MaxContextPerLayer: 5,
		},
		Skill: SkillConfig{
			Enabled:           true,
			SkillsDir:         "~/.lango/skills",
			AllowImport:       true,
			MaxBulkImport:     50,
			ImportConcurrency: 5,
			ImportTimeout:     2 * time.Minute,
		},
		Graph: GraphConfig{
			Enabled:             false,
			Backend:             "bolt",
			MaxTraversalDepth:   2,
			MaxExpansionResults: 10,
		},
		A2A: A2AConfig{
			Enabled: false,
		},
		Payment: PaymentConfig{
			Enabled:        false,
			WalletProvider: "local",
			Network: PaymentNetworkConfig{
				ChainID:      84532, // Base Sepolia
				USDCContract: "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
			},
			Limits: SpendingLimitsConfig{
				MaxPerTx:         "1.00",
				MaxDaily:         "10.00",
				AutoApproveBelow: "0.10",
			},
			X402: X402Config{
				AutoIntercept:    false,
				MaxAutoPayAmount: "0.50",
			},
		},
		Cron: CronConfig{
			Enabled:            false,
			Timezone:           "UTC",
			MaxConcurrentJobs:  5,
			DefaultSessionMode: "isolated",
			HistoryRetention:   "720h",
		},
		Background: BackgroundConfig{
			Enabled:            false,
			YieldMs:            30000,
			MaxConcurrentTasks: 3,
		},
		Workflow: WorkflowConfig{
			Enabled:            false,
			MaxConcurrentSteps: 4,
			DefaultTimeout:     10 * time.Minute,
			StateDir:           "~/.lango/workflows/",
		},
		Librarian: LibrarianConfig{
			Enabled:              false,
			ObservationThreshold: 2,
			InquiryCooldownTurns: 3,
			MaxPendingInquiries:  2,
			AutoSaveConfidence:   types.ConfidenceHigh,
		},
		P2P: P2PConfig{
			Enabled: false,
			ListenAddrs: []string{
				"/ip4/0.0.0.0/tcp/9000",
				"/ip4/0.0.0.0/udp/9000/quic-v1",
			},
			KeyDir:           "~/.lango/p2p",
			EnableRelay:      true,
			EnableMDNS:       true,
			MaxPeers:         50,
			HandshakeTimeout: 30 * time.Second,
			SessionTokenTTL:  24 * time.Hour,
			GossipInterval:   30 * time.Second,
			ZKHandshake:      true,
			ZKAttestation:    true,
			ZKP: ZKPConfig{
				ProofCacheDir:    "~/.lango/p2p/zkp-cache",
				ProvingScheme:    "plonk",
				SRSMode:          "unsafe",
				MaxCredentialAge: "24h",
			},
			ToolIsolation: ToolIsolationConfig{
				Enabled:        false,
				TimeoutPerTool: 30 * time.Second,
				MaxMemoryMB:    256,
				Container: ContainerSandboxConfig{
					Enabled:         false,
					Runtime:         "auto",
					Image:           "lango-sandbox:latest",
					NetworkMode:     "none",
					ReadOnlyRootfs:  boolPtr(true),
					PoolSize:        0,
					PoolIdleTimeout: 5 * time.Minute,
				},
			},
		},
	}
}

// boolPtr returns a pointer to a bool value.
func boolPtr(b bool) *bool {
	return &b
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
	v.SetDefault("agent.requestTimeout", defaults.Agent.RequestTimeout)
	v.SetDefault("agent.toolTimeout", defaults.Agent.ToolTimeout)
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
	v.SetDefault("security.keyring.enabled", defaults.Security.Keyring.Enabled)
	v.SetDefault("security.dbEncryption.enabled", defaults.Security.DBEncryption.Enabled)
	v.SetDefault("security.dbEncryption.cipherPageSize", defaults.Security.DBEncryption.CipherPageSize)
	v.SetDefault("security.kms.fallbackToLocal", defaults.Security.KMS.FallbackToLocal)
	v.SetDefault("security.kms.timeoutPerOperation", defaults.Security.KMS.TimeoutPerOperation)
	v.SetDefault("security.kms.maxRetries", defaults.Security.KMS.MaxRetries)
	v.SetDefault("graph.enabled", defaults.Graph.Enabled)
	v.SetDefault("graph.backend", defaults.Graph.Backend)
	v.SetDefault("graph.maxTraversalDepth", defaults.Graph.MaxTraversalDepth)
	v.SetDefault("graph.maxExpansionResults", defaults.Graph.MaxExpansionResults)
	v.SetDefault("a2a.enabled", defaults.A2A.Enabled)
	v.SetDefault("payment.enabled", defaults.Payment.Enabled)
	v.SetDefault("payment.walletProvider", defaults.Payment.WalletProvider)
	v.SetDefault("payment.network.chainId", defaults.Payment.Network.ChainID)
	v.SetDefault("payment.network.usdcContract", defaults.Payment.Network.USDCContract)
	v.SetDefault("payment.limits.maxPerTx", defaults.Payment.Limits.MaxPerTx)
	v.SetDefault("payment.limits.maxDaily", defaults.Payment.Limits.MaxDaily)
	v.SetDefault("payment.limits.autoApproveBelow", defaults.Payment.Limits.AutoApproveBelow)
	v.SetDefault("payment.x402.autoIntercept", defaults.Payment.X402.AutoIntercept)
	v.SetDefault("payment.x402.maxAutoPayAmount", defaults.Payment.X402.MaxAutoPayAmount)
	v.SetDefault("cron.enabled", defaults.Cron.Enabled)
	v.SetDefault("cron.timezone", defaults.Cron.Timezone)
	v.SetDefault("cron.maxConcurrentJobs", defaults.Cron.MaxConcurrentJobs)
	v.SetDefault("cron.defaultSessionMode", defaults.Cron.DefaultSessionMode)
	v.SetDefault("cron.historyRetention", defaults.Cron.HistoryRetention)
	v.SetDefault("cron.defaultDeliverTo", defaults.Cron.DefaultDeliverTo)
	v.SetDefault("background.enabled", defaults.Background.Enabled)
	v.SetDefault("background.yieldMs", defaults.Background.YieldMs)
	v.SetDefault("background.maxConcurrentTasks", defaults.Background.MaxConcurrentTasks)
	v.SetDefault("background.defaultDeliverTo", defaults.Background.DefaultDeliverTo)
	v.SetDefault("workflow.enabled", defaults.Workflow.Enabled)
	v.SetDefault("workflow.maxConcurrentSteps", defaults.Workflow.MaxConcurrentSteps)
	v.SetDefault("workflow.defaultTimeout", defaults.Workflow.DefaultTimeout)
	v.SetDefault("workflow.stateDir", defaults.Workflow.StateDir)
	v.SetDefault("workflow.defaultDeliverTo", defaults.Workflow.DefaultDeliverTo)
	v.SetDefault("librarian.enabled", defaults.Librarian.Enabled)
	v.SetDefault("librarian.observationThreshold", defaults.Librarian.ObservationThreshold)
	v.SetDefault("librarian.inquiryCooldownTurns", defaults.Librarian.InquiryCooldownTurns)
	v.SetDefault("librarian.maxPendingInquiries", defaults.Librarian.MaxPendingInquiries)
	v.SetDefault("librarian.autoSaveConfidence", defaults.Librarian.AutoSaveConfidence)
	v.SetDefault("security.interceptor.presidio.url", "http://localhost:5002")
	v.SetDefault("security.interceptor.presidio.scoreThreshold", 0.7)
	v.SetDefault("security.interceptor.presidio.language", "en")
	v.SetDefault("skill.enabled", defaults.Skill.Enabled)
	v.SetDefault("skill.skillsDir", defaults.Skill.SkillsDir)
	v.SetDefault("skill.allowImport", defaults.Skill.AllowImport)
	v.SetDefault("skill.maxBulkImport", defaults.Skill.MaxBulkImport)
	v.SetDefault("skill.importConcurrency", defaults.Skill.ImportConcurrency)
	v.SetDefault("skill.importTimeout", defaults.Skill.ImportTimeout)
	v.SetDefault("p2p.enabled", defaults.P2P.Enabled)
	v.SetDefault("p2p.listenAddrs", defaults.P2P.ListenAddrs)
	v.SetDefault("p2p.keyDir", defaults.P2P.KeyDir)
	v.SetDefault("p2p.nodeKeyName", "p2p.node.privatekey")
	v.SetDefault("p2p.enableRelay", defaults.P2P.EnableRelay)
	v.SetDefault("p2p.enableMdns", defaults.P2P.EnableMDNS)
	v.SetDefault("p2p.maxPeers", defaults.P2P.MaxPeers)
	v.SetDefault("p2p.handshakeTimeout", defaults.P2P.HandshakeTimeout)
	v.SetDefault("p2p.sessionTokenTtl", defaults.P2P.SessionTokenTTL)
	v.SetDefault("p2p.gossipInterval", defaults.P2P.GossipInterval)
	v.SetDefault("p2p.zkHandshake", defaults.P2P.ZKHandshake)
	v.SetDefault("p2p.zkAttestation", defaults.P2P.ZKAttestation)
	v.SetDefault("p2p.zkp.proofCacheDir", defaults.P2P.ZKP.ProofCacheDir)
	v.SetDefault("p2p.zkp.provingScheme", defaults.P2P.ZKP.ProvingScheme)
	v.SetDefault("p2p.toolIsolation.container.enabled", defaults.P2P.ToolIsolation.Container.Enabled)
	v.SetDefault("p2p.toolIsolation.container.runtime", defaults.P2P.ToolIsolation.Container.Runtime)
	v.SetDefault("p2p.toolIsolation.container.image", defaults.P2P.ToolIsolation.Container.Image)
	v.SetDefault("p2p.toolIsolation.container.networkMode", defaults.P2P.ToolIsolation.Container.NetworkMode)
	v.SetDefault("p2p.toolIsolation.container.poolSize", defaults.P2P.ToolIsolation.Container.PoolSize)
	v.SetDefault("p2p.toolIsolation.container.poolIdleTimeout", defaults.P2P.ToolIsolation.Container.PoolIdleTimeout)

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

	// Migrate legacy fields.
	cfg.MigrateEmbeddingProvider()

	// Apply environment variable substitution
	substituteEnvVars(cfg)

	// Validate configuration
	if err := Validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
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

	// Payment
	cfg.Payment.Network.RPCURL = expandEnvVars(cfg.Payment.Network.RPCURL)

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
		validProviders := map[string]bool{
			"local": true, "rpc": true, "enclave": true,
			"aws-kms": true, "gcp-kms": true, "azure-kv": true, "pkcs11": true,
		}
		if !validProviders[cfg.Security.Signer.Provider] {
			errs = append(errs, fmt.Sprintf("invalid security.signer.provider: %q (must be local, rpc, enclave, aws-kms, gcp-kms, azure-kv, or pkcs11)", cfg.Security.Signer.Provider))
		}
		if cfg.Security.Signer.Provider == "rpc" && cfg.Security.Signer.RPCUrl == "" {
			errs = append(errs, "security.signer.rpcUrl is required when provider is 'rpc'")
		}
		// Validate KMS-specific config.
		switch cfg.Security.Signer.Provider {
		case "aws-kms", "gcp-kms":
			if cfg.Security.KMS.KeyID == "" {
				errs = append(errs, fmt.Sprintf("security.kms.keyId is required when provider is %q", cfg.Security.Signer.Provider))
			}
		case "azure-kv":
			if cfg.Security.KMS.Azure.VaultURL == "" {
				errs = append(errs, "security.kms.azure.vaultUrl is required when provider is 'azure-kv'")
			}
			if cfg.Security.KMS.KeyID == "" {
				errs = append(errs, "security.kms.keyId is required when provider is 'azure-kv'")
			}
		case "pkcs11":
			if cfg.Security.KMS.PKCS11.ModulePath == "" {
				errs = append(errs, "security.kms.pkcs11.modulePath is required when provider is 'pkcs11'")
			}
		}
	}

	// Validate graph config
	if cfg.Graph.Enabled && cfg.Graph.Backend != "bolt" {
		errs = append(errs, fmt.Sprintf("graph.backend %q is not supported (must be \"bolt\")", cfg.Graph.Backend))
	}

	// Validate A2A config
	if cfg.A2A.Enabled {
		if cfg.A2A.BaseURL == "" {
			errs = append(errs, "a2a.baseUrl is required when A2A is enabled")
		}
		if cfg.A2A.AgentName == "" {
			errs = append(errs, "a2a.agentName is required when A2A is enabled")
		}
	}

	// Validate payment config
	if cfg.Payment.Enabled {
		if cfg.Payment.Network.RPCURL == "" {
			errs = append(errs, "payment.network.rpcUrl is required when payment is enabled")
		}
		validWalletProviders := map[string]bool{"local": true, "rpc": true, "composite": true}
		if !validWalletProviders[cfg.Payment.WalletProvider] {
			errs = append(errs, fmt.Sprintf("invalid payment.walletProvider: %q (must be local, rpc, or composite)", cfg.Payment.WalletProvider))
		}
	}

	// Validate P2P config
	if cfg.P2P.Enabled {
		if !cfg.Payment.Enabled {
			errs = append(errs, "p2p requires payment.enabled (wallet needed for identity)")
		}
		validSchemes := map[string]bool{"plonk": true, "groth16": true}
		if cfg.P2P.ZKP.ProvingScheme != "" && !validSchemes[cfg.P2P.ZKP.ProvingScheme] {
			errs = append(errs, fmt.Sprintf("invalid p2p.zkp.provingScheme: %q (must be plonk or groth16)", cfg.P2P.ZKP.ProvingScheme))
		}
	}

	// Validate container sandbox config
	if cfg.P2P.ToolIsolation.Container.Enabled {
		validRuntimes := map[string]bool{"auto": true, "docker": true, "gvisor": true, "native": true}
		if !validRuntimes[cfg.P2P.ToolIsolation.Container.Runtime] {
			errs = append(errs, fmt.Sprintf("invalid p2p.toolIsolation.container.runtime: %q (must be auto, docker, gvisor, or native)", cfg.P2P.ToolIsolation.Container.Runtime))
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
