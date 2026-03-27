package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/langoai/lango/internal/provider"
	"github.com/langoai/lango/internal/types"
	"github.com/spf13/viper"
)

var envVarRegex = regexp.MustCompile(`\$\{([^}]+)\}`)

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		DataRoot: "~/.lango",
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
			OutputManager: OutputManagerConfig{
				TokenBudget: 2000,
				HeadRatio:   0.7,
				TailRatio:   0.3,
			},
		},
		Security: SecurityConfig{
			Interceptor: InterceptorConfig{
				Enabled:        true,
				ApprovalPolicy: ApprovalPolicyDangerous,
				Presidio: PresidioConfig{
					URL:            "http://localhost:5002",
					ScoreThreshold: 0.7,
					Language:       "en",
				},
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
			DefaultJobTimeout:  30 * time.Minute,
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
		ObservationalMemory: ObservationalMemoryConfig{
			Enabled:                          false,
			MessageTokenThreshold:            1000,
			ObservationTokenThreshold:        2000,
			MaxMessageTokenBudget:            8000,
			MaxReflectionsInContext:          5,
			MaxObservationsInContext:         20,
			MemoryTokenBudget:                4000,
			ReflectionConsolidationThreshold: 5,
		},
		Librarian: LibrarianConfig{
			Enabled:              false,
			ObservationThreshold: 2,
			InquiryCooldownTurns: 3,
			MaxPendingInquiries:  2,
			AutoSaveConfidence:   types.ConfidenceHigh,
		},
		Retrieval: RetrievalConfig{
			Enabled:  false,
			Shadow:   true,
			Feedback: false,
		},
		MCP: MCPConfig{
			Enabled:              false,
			DefaultTimeout:       30 * time.Second,
			MaxOutputTokens:      25000,
			HealthCheckInterval:  30 * time.Second,
			AutoReconnect:        true,
			MaxReconnectAttempts: 5,
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

// setDefaultsFromStruct recursively walks a struct using mapstructure tags
// and calls v.SetDefault for each non-zero leaf value. This ensures
// DefaultConfig() is the single source of truth for all default values.
func setDefaultsFromStruct(v *viper.Viper, prefix string, val reflect.Value) {
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		tag := field.Tag.Get("mapstructure")
		if tag == "" || tag == "-" {
			continue
		}

		key := tag
		if prefix != "" {
			key = prefix + "." + tag
		}

		// Dereference pointers.
		actual := fieldVal
		if actual.Kind() == reflect.Ptr {
			if actual.IsNil() {
				continue
			}
			actual = actual.Elem()
		}

		switch actual.Kind() {
		case reflect.Struct:
			// Recurse into nested config sections.
			setDefaultsFromStruct(v, key, actual)
		case reflect.Map:
			// Skip maps — they contain dynamic user content, not static defaults.
			continue
		case reflect.Slice:
			if actual.Len() > 0 {
				v.SetDefault(key, actual.Interface())
			}
		default:
			if actual.IsZero() {
				continue
			}
			// Convert string-based custom types (ApprovalPolicy, Confidence, etc.)
			// to plain string so viper stores them consistently.
			if actual.Kind() == reflect.String {
				v.SetDefault(key, actual.String())
			} else {
				v.SetDefault(key, actual.Interface())
			}
		}
	}
}

// Load reads configuration from file and environment
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults from DefaultConfig — single source of truth.
	defaults := DefaultConfig()
	setDefaultsFromStruct(v, "", reflect.ValueOf(defaults).Elem())

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

	// Post-load: migrate, substitute env vars, normalize paths, validate.
	if err := PostLoad(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// PostLoad applies post-load processing: legacy migration, env substitution,
// path normalization, path validation, and full config validation.
// All operations are idempotent — safe to call multiple times on the same config.
func PostLoad(cfg *Config) error {
	cfg.MigrateEmbeddingProvider()
	substituteEnvVars(cfg)
	NormalizePaths(cfg)
	if err := ValidateDataPaths(cfg); err != nil {
		return err
	}
	return Validate(cfg)
}

// substituteEnvVars replaces ${VAR} patterns with environment variable values
func substituteEnvVars(cfg *Config) {
	// Provider credentials
	for id, pCfg := range cfg.Providers {
		pCfg.APIKey = ExpandEnvVars(pCfg.APIKey)
		cfg.Providers[id] = pCfg
	}

	// Channel tokens
	cfg.Channels.Telegram.BotToken = ExpandEnvVars(cfg.Channels.Telegram.BotToken)
	cfg.Channels.Discord.BotToken = ExpandEnvVars(cfg.Channels.Discord.BotToken)
	cfg.Channels.Slack.BotToken = ExpandEnvVars(cfg.Channels.Slack.BotToken)
	cfg.Channels.Slack.AppToken = ExpandEnvVars(cfg.Channels.Slack.AppToken)
	cfg.Channels.Slack.SigningSecret = ExpandEnvVars(cfg.Channels.Slack.SigningSecret)

	// Auth OIDC provider credentials
	for id, aCfg := range cfg.Auth.Providers {
		aCfg.ClientID = ExpandEnvVars(aCfg.ClientID)
		aCfg.ClientSecret = ExpandEnvVars(aCfg.ClientSecret)
		cfg.Auth.Providers[id] = aCfg
	}

	// Payment
	cfg.Payment.Network.RPCURL = ExpandEnvVars(cfg.Payment.Network.RPCURL)

	// MCP server env/headers
	for name, srv := range cfg.MCP.Servers {
		for k, v := range srv.Env {
			srv.Env[k] = ExpandEnvVars(v)
		}
		for k, v := range srv.Headers {
			srv.Headers[k] = ExpandEnvVars(v)
		}
		cfg.MCP.Servers[name] = srv
	}

	// Paths
	cfg.Session.DatabasePath = ExpandEnvVars(cfg.Session.DatabasePath)
}

// ExpandEnvVars replaces ${VAR} patterns in s with environment variable values.
// Variables that are not set in the environment are left as-is.
func ExpandEnvVars(s string) string {
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

	// Validate agent.fallbackProvider references an existing key in providers map
	if cfg.Agent.FallbackProvider != "" && len(cfg.Providers) > 0 {
		if _, ok := cfg.Providers[cfg.Agent.FallbackProvider]; !ok {
			errs = append(errs, fmt.Sprintf("agent.fallbackProvider %q not found in providers map (available: %v)", cfg.Agent.FallbackProvider, providerKeys(cfg.Providers)))
		}
	}

	// Validate provider-model compatibility (primary)
	if cfg.Agent.Provider != "" && cfg.Agent.Model != "" {
		if pCfg, ok := cfg.Providers[cfg.Agent.Provider]; ok {
			if err := provider.ValidateModelProvider(string(pCfg.Type), cfg.Agent.Model); err != nil {
				errs = append(errs, fmt.Sprintf("agent.model %q incompatible with provider %q (type %s): %v", cfg.Agent.Model, cfg.Agent.Provider, pCfg.Type, err))
			}
		}
	}

	// Validate provider-model compatibility (fallback)
	if cfg.Agent.FallbackProvider != "" && cfg.Agent.FallbackModel != "" {
		if pCfg, ok := cfg.Providers[cfg.Agent.FallbackProvider]; ok {
			if err := provider.ValidateModelProvider(string(pCfg.Type), cfg.Agent.FallbackModel); err != nil {
				errs = append(errs, fmt.Sprintf("agent.fallbackModel %q incompatible with fallbackProvider %q (type %s): %v", cfg.Agent.FallbackModel, cfg.Agent.FallbackProvider, pCfg.Type, err))
			}
		}
	}

	// Validate logging config
	if !ValidLogLevels[cfg.Logging.Level] {
		errs = append(errs, fmt.Sprintf("invalid log level: %s (must be debug, info, warn, or error)", cfg.Logging.Level))
	}

	if !ValidLogFormats[cfg.Logging.Format] {
		errs = append(errs, fmt.Sprintf("invalid log format: %s (must be json or console)", cfg.Logging.Format))
	}

	// Validate security config
	if cfg.Security.Signer.Provider != "" {
		if !ValidSignerProviders[cfg.Security.Signer.Provider] {
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
		if !ValidWalletProviders[cfg.Payment.WalletProvider] {
			errs = append(errs, fmt.Sprintf("invalid payment.walletProvider: %q (must be local, rpc, or composite)", cfg.Payment.WalletProvider))
		}
	}

	// Validate P2P config
	if cfg.P2P.Enabled {
		if !cfg.Payment.Enabled {
			errs = append(errs, "p2p requires payment.enabled (wallet needed for identity)")
		}
		if cfg.P2P.ZKP.ProvingScheme != "" && !ValidZKPSchemes[cfg.P2P.ZKP.ProvingScheme] {
			errs = append(errs, fmt.Sprintf("invalid p2p.zkp.provingScheme: %q (must be plonk or groth16)", cfg.P2P.ZKP.ProvingScheme))
		}
	}

	// Validate container sandbox config
	if cfg.P2P.ToolIsolation.Container.Enabled {
		if !ValidContainerRuntimes[cfg.P2P.ToolIsolation.Container.Runtime] {
			errs = append(errs, fmt.Sprintf("invalid p2p.toolIsolation.container.runtime: %q (must be auto, docker, gvisor, or native)", cfg.P2P.ToolIsolation.Container.Runtime))
		}
	}

	// Validate MCP config
	if cfg.MCP.Enabled {
		for name, srv := range cfg.MCP.Servers {
			if !ValidMCPTransports[srv.Transport] {
				errs = append(errs, fmt.Sprintf("mcp.servers.%s.transport %q is not supported (must be stdio, http, or sse)", name, srv.Transport))
			}
			switch srv.Transport {
			case "", "stdio":
				if srv.Command == "" {
					errs = append(errs, fmt.Sprintf("mcp.servers.%s.command is required for stdio transport", name))
				}
			case "http", "sse":
				if srv.URL == "" {
					errs = append(errs, fmt.Sprintf("mcp.servers.%s.url is required for %s transport", name, srv.Transport))
				}
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}

// expandTilde replaces a leading ~ with the given home directory.
func expandTilde(path, home string) string {
	if home == "" || (!strings.HasPrefix(path, "~/") && path != "~") {
		return path
	}
	return filepath.Join(home, path[1:])
}

// NormalizePaths resolves relative data paths to be under DataRoot and
// expands ~ in all path fields. Call after Load/Unmarshal.
func NormalizePaths(cfg *Config) {
	home, _ := os.UserHomeDir()

	if cfg.DataRoot == "" {
		cfg.DataRoot = "~/.lango"
	}
	cfg.DataRoot = expandTilde(cfg.DataRoot, home)

	// Normalize each configurable data path.
	normalizePath(&cfg.Session.DatabasePath, cfg.DataRoot, home)
	normalizePath(&cfg.Graph.DatabasePath, cfg.DataRoot, home)
	normalizePath(&cfg.Skill.SkillsDir, cfg.DataRoot, home)
	normalizePath(&cfg.Workflow.StateDir, cfg.DataRoot, home)
	normalizePath(&cfg.P2P.KeyDir, cfg.DataRoot, home)
	normalizePath(&cfg.P2P.ZKP.ProofCacheDir, cfg.DataRoot, home)
	normalizePath(&cfg.P2P.Workspace.DataDir, cfg.DataRoot, home)
}

// normalizePath expands ~ and resolves relative paths under dataRoot.
func normalizePath(p *string, dataRoot, home string) {
	if p == nil || *p == "" {
		return
	}
	*p = expandTilde(*p, home)

	// If path is relative (not starting with /), resolve under dataRoot.
	if !filepath.IsAbs(*p) {
		*p = filepath.Join(dataRoot, *p)
	}
	*p = filepath.Clean(*p)
}

// ValidateDataPaths checks that all configurable data paths reside under DataRoot.
// Must be called after NormalizePaths (paths are already cleaned absolute paths).
func ValidateDataPaths(cfg *Config) error {
	if cfg.DataRoot == "" {
		return nil
	}

	root := filepath.Clean(cfg.DataRoot)
	rootPrefix := root + string(os.PathSeparator)

	type pathEntry struct {
		field string
		value string
	}

	entries := []pathEntry{
		{"session.databasePath", cfg.Session.DatabasePath},
		{"graph.databasePath", cfg.Graph.DatabasePath},
		{"skill.skillsDir", cfg.Skill.SkillsDir},
		{"workflow.stateDir", cfg.Workflow.StateDir},
		{"p2p.keyDir", cfg.P2P.KeyDir},
		{"p2p.zkp.proofCacheDir", cfg.P2P.ZKP.ProofCacheDir},
		{"p2p.workspace.dataDir", cfg.P2P.Workspace.DataDir},
	}

	var errs []string
	for _, e := range entries {
		if e.value == "" {
			continue
		}
		cleaned := filepath.Clean(e.value)
		// Path must be equal to or under the data root.
		if cleaned != root && !strings.HasPrefix(cleaned, rootPrefix) {
			errs = append(errs, fmt.Sprintf("%s (%q) must be under data root (%s)", e.field, e.value, root))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("data path validation failed:\n  - %s", strings.Join(errs, "\n  - "))
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
