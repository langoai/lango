package app

import (
	"context"
	"crypto/hmac"
	"fmt"
	"os"
	"strings"

	"github.com/langowarny/lango/internal/adk"
	"github.com/langowarny/lango/internal/agent"
	"github.com/langowarny/lango/internal/config"
	"github.com/langowarny/lango/internal/gateway"
	"github.com/langowarny/lango/internal/knowledge"
	"github.com/langowarny/lango/internal/learning"
	"github.com/langowarny/lango/internal/memory"
	"github.com/langowarny/lango/internal/provider"
	"github.com/langowarny/lango/internal/security"
	"github.com/langowarny/lango/internal/session"
	"github.com/langowarny/lango/internal/skill"
	"github.com/langowarny/lango/internal/supervisor"
	"google.golang.org/adk/model"
	adk_tool "google.golang.org/adk/tool"
)

const _defaultSystemPrompt = "You are Lango, a powerful AI assistant. You have access to tools for shell command execution and file system operations. Use them when appropriate to help the user."

// loadSystemPrompt returns the system prompt from file or the default.
func loadSystemPrompt(path string) string {
	if path == "" {
		return _defaultSystemPrompt
	}
	data, err := os.ReadFile(path)
	if err != nil {
		logger().Warnw("system prompt file not found, using default", "path", path, "error", err)
		return _defaultSystemPrompt
	}
	prompt := string(data)
	if prompt == "" {
		return _defaultSystemPrompt
	}
	logger().Infow("loaded custom system prompt", "path", path)
	return prompt
}

// initSupervisor creates and initializes the Supervisor.
func initSupervisor(cfg *config.Config) (*supervisor.Supervisor, error) {
	logger().Info("initializing supervisor...")
	sv, err := supervisor.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("supervisor: %w", err)
	}
	return sv, nil
}

// initSessionStore creates and initializes the session store.
func initSessionStore(cfg *config.Config) (session.Store, error) {
	passphrase := os.Getenv("LANGO_PASSPHRASE")
	if passphrase == "" {
		passphrase = cfg.Security.Passphrase
	}

	var storeOpts []session.StoreOption
	if passphrase != "" {
		storeOpts = append(storeOpts, session.WithPassphrase(passphrase))
	}
	if cfg.Session.MaxHistoryTurns > 0 {
		storeOpts = append(storeOpts, session.WithMaxHistoryTurns(cfg.Session.MaxHistoryTurns))
	}
	if cfg.Session.TTL > 0 {
		storeOpts = append(storeOpts, session.WithTTL(cfg.Session.TTL))
	}

	logger().Info("initializing session store...")
	store, err := session.NewEntStore(cfg.Session.DatabasePath, storeOpts...)
	if err != nil {
		return nil, fmt.Errorf("session store: %w", err)
	}
	return store, nil
}

// initSecurity creates and initializes the security stack.
func initSecurity(cfg *config.Config, store session.Store) (security.CryptoProvider, *security.KeyRegistry, *security.SecretsStore, error) {
	if cfg.Security.Signer.Provider == "" {
		return nil, nil, nil, nil
	}

	switch cfg.Security.Signer.Provider {
	case "local":
		passphrase := os.Getenv("LANGO_PASSPHRASE")
		if passphrase == "" {
			passphrase = cfg.Security.Passphrase
		}
		if passphrase == "" {
			return nil, nil, nil, fmt.Errorf("local security provider requires a passphrase")
		}

		provider := security.NewLocalCryptoProvider()

		entStore, ok := store.(*session.EntStore)
		if !ok {
			return nil, nil, nil, fmt.Errorf("local security provider requires EntStore")
		}

		salt, err := entStore.GetSalt("default")
		if err != nil {
			// First-time setup: initialize with new salt
			if err := provider.Initialize(passphrase); err != nil {
				return nil, nil, nil, fmt.Errorf("initialize crypto provider: %w", err)
			}
			if err := entStore.SetSalt("default", provider.Salt()); err != nil {
				return nil, nil, nil, fmt.Errorf("store salt: %w", err)
			}
			checksum := provider.CalculateChecksum(passphrase, provider.Salt())
			if err := entStore.SetChecksum("default", checksum); err != nil {
				return nil, nil, nil, fmt.Errorf("store checksum: %w", err)
			}
		} else {
			// Existing salt found: initialize with it
			if err := provider.InitializeWithSalt(passphrase, salt); err != nil {
				return nil, nil, nil, fmt.Errorf("initialize crypto provider with salt: %w", err)
			}
			// Verify checksum
			storedChecksum, err := entStore.GetChecksum("default")
			if err == nil {
				computed := provider.CalculateChecksum(passphrase, salt)
				if !hmac.Equal(storedChecksum, computed) {
					return nil, nil, nil, fmt.Errorf("passphrase checksum mismatch: incorrect passphrase")
				}
			}
		}

		client := entStore.Client()
		keys := security.NewKeyRegistry(client)
		secrets := security.NewSecretsStore(client, keys, provider)

		// Register default encryption key
		ctx := context.Background()
		if _, err := keys.RegisterKey(ctx, "default", "local", security.KeyTypeEncryption); err != nil {
			return nil, nil, nil, fmt.Errorf("register default key: %w", err)
		}

		logger().Info("security initialized (local provider)")
		return provider, keys, secrets, nil

	case "rpc":
		provider := security.NewRPCProvider()

		entStore, ok := store.(*session.EntStore)
		if !ok {
			return nil, nil, nil, fmt.Errorf("rpc security provider requires EntStore")
		}

		client := entStore.Client()
		keys := security.NewKeyRegistry(client)
		secrets := security.NewSecretsStore(client, keys, provider)

		logger().Info("security initialized (rpc provider)")
		return provider, keys, secrets, nil

	case "enclave":
		return nil, nil, nil, fmt.Errorf("enclave provider not yet implemented")

	default:
		return nil, nil, nil, fmt.Errorf("unknown security provider: %s", cfg.Security.Signer.Provider)
	}
}

// knowledgeComponents holds optional self-learning components.
type knowledgeComponents struct {
	store    *knowledge.Store
	engine   *learning.Engine
	registry *skill.Registry
}

// initKnowledge creates the self-learning components if enabled.
func initKnowledge(cfg *config.Config, store session.Store, baseTools []*agent.Tool) *knowledgeComponents {
	if !cfg.Knowledge.Enabled {
		logger().Info("knowledge system disabled")
		return nil
	}

	entStore, ok := store.(*session.EntStore)
	if !ok {
		logger().Warn("knowledge system requires EntStore, skipping")
		return nil
	}

	client := entStore.Client()
	kLogger := logger()

	kStore := knowledge.NewStore(
		client, kLogger,
		cfg.Knowledge.MaxKnowledge,
		cfg.Knowledge.MaxLearnings,
		cfg.Knowledge.MaxSkillsPerDay,
	)

	engine := learning.NewEngine(kStore, kLogger)
	registry, err := skill.NewRegistry(kStore, baseTools, kLogger)
	if err != nil {
		logger().Warnw("skill registry init error, skipping knowledge system", "error", err)
		return nil
	}

	ctx := context.Background()
	if err := registry.LoadSkills(ctx); err != nil {
		logger().Warnw("load skills error", "error", err)
	}

	logger().Info("knowledge system initialized")
	return &knowledgeComponents{
		store:    kStore,
		engine:   engine,
		registry: registry,
	}
}

// memoryComponents holds optional observational memory components.
type memoryComponents struct {
	store     *memory.Store
	observer  *memory.Observer
	reflector *memory.Reflector
	buffer    *memory.Buffer
}

// providerTextGenerator adapts a supervisor.ProviderProxy to the memory.TextGenerator interface.
type providerTextGenerator struct {
	proxy *supervisor.ProviderProxy
}

func (g *providerTextGenerator) GenerateText(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	params := provider.GenerateParams{
		Messages: []provider.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	stream, err := g.proxy.Generate(ctx, params)
	if err != nil {
		return "", fmt.Errorf("generate text: %w", err)
	}

	var result strings.Builder
	for evt, err := range stream {
		if err != nil {
			return "", fmt.Errorf("stream text: %w", err)
		}
		if evt.Type == provider.StreamEventPlainText {
			result.WriteString(evt.Text)
		}
		if evt.Type == provider.StreamEventError && evt.Error != nil {
			return "", evt.Error
		}
	}
	return result.String(), nil
}

// initMemory creates the observational memory components if enabled.
func initMemory(cfg *config.Config, store session.Store, sv *supervisor.Supervisor) *memoryComponents {
	if !cfg.ObservationalMemory.Enabled {
		logger().Info("observational memory disabled")
		return nil
	}

	entStore, ok := store.(*session.EntStore)
	if !ok {
		logger().Warn("observational memory requires EntStore, skipping")
		return nil
	}

	client := entStore.Client()
	mLogger := logger()
	mStore := memory.NewStore(client, mLogger)

	// Create provider proxy for observer/reflector LLM calls
	provider := cfg.ObservationalMemory.Provider
	if provider == "" {
		provider = cfg.Agent.Provider
	}
	omModel := cfg.ObservationalMemory.Model
	if omModel == "" {
		omModel = cfg.Agent.Model
	}

	proxy := supervisor.NewProviderProxy(sv, provider, omModel)
	generator := &providerTextGenerator{proxy: proxy}

	observer := memory.NewObserver(generator, mStore, mLogger)
	reflector := memory.NewReflector(generator, mStore, mLogger)

	// Apply defaults for thresholds
	msgThreshold := cfg.ObservationalMemory.MessageTokenThreshold
	if msgThreshold <= 0 {
		msgThreshold = 1000
	}
	obsThreshold := cfg.ObservationalMemory.ObservationTokenThreshold
	if obsThreshold <= 0 {
		obsThreshold = 2000
	}

	// Message provider retrieves messages for a session key
	getMessages := func(sessionKey string) ([]session.Message, error) {
		sess, err := store.Get(sessionKey)
		if err != nil {
			return nil, err
		}
		if sess == nil {
			return nil, nil
		}
		return sess.History, nil
	}

	buffer := memory.NewBuffer(observer, reflector, mStore, msgThreshold, obsThreshold, getMessages, mLogger)

	logger().Infow("observational memory initialized",
		"provider", provider,
		"model", omModel,
		"messageTokenThreshold", msgThreshold,
		"observationTokenThreshold", obsThreshold,
	)

	return &memoryComponents{
		store:     mStore,
		observer:  observer,
		reflector: reflector,
		buffer:    buffer,
	}
}

// initAuth creates the auth manager if OIDC providers are configured.
func initAuth(cfg *config.Config, store session.Store) *gateway.AuthManager {
	if len(cfg.Auth.Providers) == 0 {
		return nil
	}

	auth, err := gateway.NewAuthManager(cfg.Auth, store)
	if err != nil {
		logger().Warnw("auth manager init error, skipping", "error", err)
		return nil
	}

	logger().Info("auth manager initialized")
	return auth
}

// initAgent creates the ADK agent with the given tools and provider proxy.
func initAgent(ctx context.Context, sv *supervisor.Supervisor, cfg *config.Config, store session.Store, tools []*agent.Tool, kc *knowledgeComponents, mc *memoryComponents, scanner *agent.SecretScanner) (*adk.Agent, error) {
	// Adapt tools to ADK format
	var adkTools []adk_tool.Tool
	for _, t := range tools {
		at, err := adk.AdaptTool(t)
		if err != nil {
			logger().Warnw("adapt tool error", "name", t.Name, "error", err)
			continue
		}
		adkTools = append(adkTools, at)
	}

	// Create provider proxy with temperature, maxTokens, and fallback options
	var proxyOpts []supervisor.ProxyOption
	if cfg.Agent.Temperature != 0 {
		proxyOpts = append(proxyOpts, supervisor.WithTemperature(cfg.Agent.Temperature))
	}
	if cfg.Agent.MaxTokens != 0 {
		proxyOpts = append(proxyOpts, supervisor.WithMaxTokens(cfg.Agent.MaxTokens))
	}
	if cfg.Agent.FallbackProvider != "" {
		proxyOpts = append(proxyOpts, supervisor.WithFallback(cfg.Agent.FallbackProvider, cfg.Agent.FallbackModel))
	}

	proxy := supervisor.NewProviderProxy(sv, cfg.Agent.Provider, cfg.Agent.Model, proxyOpts...)
	modelAdapter := adk.NewModelAdapter(proxy)

	// Load system prompt (from file or default)
	systemPrompt := loadSystemPrompt(cfg.Agent.SystemPromptPath)

	// If knowledge is enabled, wrap with context-aware adapter
	var llm model.LLM = modelAdapter
	if kc != nil {
		retriever := knowledge.NewContextRetriever(
			kc.store,
			cfg.Knowledge.MaxContextPerLayer,
			logger(),
		)
		ctxAdapter := adk.NewContextAwareModelAdapter(modelAdapter, retriever, systemPrompt, logger())

		// Wire in observational memory if available
		if mc != nil {
			ctxAdapter.WithMemory(mc.store, "")
		}

		llm = ctxAdapter
	} else if mc != nil {
		// OM without knowledge system — create minimal context-aware adapter
		ctxAdapter := adk.NewContextAwareModelAdapter(modelAdapter, nil, systemPrompt, logger())
		ctxAdapter.WithMemory(mc.store, "")
		llm = ctxAdapter
	}

	// If PII redaction is enabled, wrap with PII-redacting adapter
	if cfg.Security.Interceptor.Enabled && cfg.Security.Interceptor.RedactPII {
		redactor := agent.NewPIIRedactor(agent.PIIConfig{
			RedactEmail: true,
			RedactPhone: true,
			CustomRegex: cfg.Security.Interceptor.PIIRegexPatterns,
		})
		llm = adk.NewPIIRedactingModelAdapter(llm, redactor, scanner)
		logger().Info("PII redaction interceptor enabled")
	}

	logger().Info("initializing agent runtime (ADK)...")
	adkAgent, err := adk.NewAgent(ctx, adkTools, llm, systemPrompt, store)
	if err != nil {
		return nil, fmt.Errorf("adk agent: %w", err)
	}
	return adkAgent, nil
}

// initGateway creates the gateway server.
func initGateway(cfg *config.Config, adkAgent *adk.Agent, store session.Store, auth *gateway.AuthManager) *gateway.Server {
	return gateway.New(gateway.Config{
		Host:             cfg.Server.Host,
		Port:             cfg.Server.Port,
		HTTPEnabled:      cfg.Server.HTTPEnabled,
		WebSocketEnabled: cfg.Server.WebSocketEnabled,
	}, adkAgent, nil, store, auth)
}
