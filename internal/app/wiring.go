package app

import (
	"context"
	"fmt"
	"strings"

	"database/sql"

	"github.com/langowarny/lango/internal/adk"
	"github.com/langowarny/lango/internal/agent"
	"github.com/langowarny/lango/internal/bootstrap"
	"github.com/langowarny/lango/internal/config"
	"github.com/langowarny/lango/internal/embedding"
	"github.com/langowarny/lango/internal/gateway"
	"github.com/langowarny/lango/internal/knowledge"
	"github.com/langowarny/lango/internal/learning"
	"github.com/langowarny/lango/internal/memory"
	"github.com/langowarny/lango/internal/prompt"
	"github.com/langowarny/lango/internal/provider"
	"github.com/langowarny/lango/internal/security"
	"github.com/langowarny/lango/internal/session"
	"github.com/langowarny/lango/internal/skill"
	"github.com/langowarny/lango/internal/supervisor"
	"google.golang.org/adk/model"
	adk_tool "google.golang.org/adk/tool"
)

// buildPromptBuilder returns a prompt.Builder configured from the agent settings.
// Priority: PromptsDir (directory of .md files) > defaults.
func buildPromptBuilder(cfg *config.AgentConfig) *prompt.Builder {
	if cfg.PromptsDir != "" {
		return prompt.LoadFromDir(cfg.PromptsDir, logger())
	}
	return prompt.DefaultBuilder()
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
// When a bootstrap result is available, the shared DB client is reused to avoid
// opening a second database connection.
func initSessionStore(cfg *config.Config, boot *bootstrap.Result) (session.Store, error) {
	var storeOpts []session.StoreOption
	if cfg.Session.MaxHistoryTurns > 0 {
		storeOpts = append(storeOpts, session.WithMaxHistoryTurns(cfg.Session.MaxHistoryTurns))
	}
	if cfg.Session.TTL > 0 {
		storeOpts = append(storeOpts, session.WithTTL(cfg.Session.TTL))
	}

	logger().Info("initializing session store...")

	// Reuse the ent client opened during bootstrap.
	if boot != nil && boot.DBClient != nil {
		return session.NewEntStoreWithClient(boot.DBClient, storeOpts...), nil
	}

	// Fallback: open a fresh connection (e.g., for tests).
	store, err := session.NewEntStore(cfg.Session.DatabasePath, storeOpts...)
	if err != nil {
		return nil, fmt.Errorf("session store: %w", err)
	}
	return store, nil
}

// initSecurity creates and initializes the security stack.
// When a bootstrap result provides an already-initialized CryptoProvider, it is
// reused for the "local" provider case to avoid re-acquiring the passphrase.
func initSecurity(cfg *config.Config, store session.Store, boot *bootstrap.Result) (security.CryptoProvider, *security.KeyRegistry, *security.SecretsStore, error) {
	if cfg.Security.Signer.Provider == "" {
		return nil, nil, nil, nil
	}

	switch cfg.Security.Signer.Provider {
	case "local":
		// Reuse the crypto provider initialized during bootstrap.
		if boot != nil && boot.Crypto != nil && boot.DBClient != nil {
			keys := security.NewKeyRegistry(boot.DBClient)
			secrets := security.NewSecretsStore(boot.DBClient, keys, boot.Crypto)

			ctx := context.Background()
			if _, err := keys.RegisterKey(ctx, "default", "local", security.KeyTypeEncryption); err != nil {
				return nil, nil, nil, fmt.Errorf("register default key: %w", err)
			}

			logger().Info("security initialized (local provider, from bootstrap)")
			return boot.Crypto, keys, secrets, nil
		}

		// Fallback: standalone initialization (should not happen in normal flow).
		return nil, nil, nil, fmt.Errorf("local security provider requires bootstrap")

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

// embeddingComponents holds optional embedding/RAG components.
type embeddingComponents struct {
	buffer     *embedding.EmbeddingBuffer
	ragService *embedding.RAGService
}

// initEmbedding creates the embedding pipeline and RAG service if configured.
func initEmbedding(cfg *config.Config, rawDB *sql.DB, kc *knowledgeComponents, mc *memoryComponents) *embeddingComponents {
	emb := cfg.Embedding
	if emb.Provider == "" && emb.ProviderID == "" {
		logger().Info("embedding system disabled (no provider configured)")
		return nil
	}

	backendType, apiKey := cfg.ResolveEmbeddingProvider()
	if backendType == "" {
		logger().Warnw("embedding provider type could not be resolved",
			"providerID", emb.ProviderID, "provider", emb.Provider)
		return nil
	}

	providerCfg := embedding.ProviderConfig{
		Provider:   backendType,
		Model:      emb.Model,
		Dimensions: emb.Dimensions,
		APIKey:     apiKey,
		BaseURL:    emb.Local.BaseURL,
	}
	if backendType == "local" && emb.Local.Model != "" {
		providerCfg.Model = emb.Local.Model
	}

	registry, err := embedding.NewRegistry(providerCfg, nil, logger())
	if err != nil {
		logger().Warnw("embedding provider init failed, skipping", "error", err)
		return nil
	}

	provider := registry.Provider()
	dimensions := provider.Dimensions()

	// Create vector store using the shared database.
	if rawDB == nil {
		logger().Warn("embedding requires raw DB handle, skipping")
		return nil
	}
	vecStore, err := embedding.NewSQLiteVecStore(rawDB, dimensions)
	if err != nil {
		logger().Warnw("sqlite-vec store init failed, skipping", "error", err)
		return nil
	}

	embLogger := logger()

	// Create buffer.
	buffer := embedding.NewEmbeddingBuffer(provider, vecStore, embLogger)

	// Create resolver and RAG service.
	var ks *knowledge.Store
	var ms *memory.Store
	if kc != nil {
		ks = kc.store
	}
	if mc != nil {
		ms = mc.store
	}
	resolver := embedding.NewStoreResolver(ks, ms)
	ragService := embedding.NewRAGService(provider, vecStore, resolver, embLogger)

	// Wire embed callbacks into stores so saves trigger async embedding.
	embedCB := func(id, collection, content string, metadata map[string]string) {
		buffer.Enqueue(embedding.EmbedRequest{
			ID:         id,
			Collection: collection,
			Content:    content,
			Metadata:   metadata,
		})
	}
	if kc != nil {
		kc.store.SetEmbedCallback(embedCB)
	}
	if mc != nil {
		mc.store.SetEmbedCallback(embedCB)
	}

	logger().Infow("embedding system initialized",
		"provider", backendType,
		"providerID", emb.ProviderID,
		"dimensions", dimensions,
		"ragEnabled", emb.RAG.Enabled,
	)

	return &embeddingComponents{
		buffer:     buffer,
		ragService: ragService,
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
func initAgent(ctx context.Context, sv *supervisor.Supervisor, cfg *config.Config, store session.Store, tools []*agent.Tool, kc *knowledgeComponents, mc *memoryComponents, ec *embeddingComponents, scanner *agent.SecretScanner) (*adk.Agent, error) {
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
	modelAdapter := adk.NewModelAdapter(proxy, cfg.Agent.Model)

	// Build structured system prompt
	builder := buildPromptBuilder(&cfg.Agent)
	systemPrompt := builder.Build()

	// If knowledge is enabled, wrap with context-aware adapter
	var llm model.LLM = modelAdapter
	if kc != nil {
		retriever := knowledge.NewContextRetriever(
			kc.store,
			cfg.Knowledge.MaxContextPerLayer,
			logger(),
		)

		// Wire tool registry and runtime context providers
		toolAdapter := adk.NewToolRegistryAdapter(tools)
		retriever.WithToolRegistry(toolAdapter)

		runtimeAdapter := adk.NewRuntimeContextAdapter(
			len(tools),
			cfg.Security.Signer.Provider != "",
			cfg.Knowledge.Enabled,
			cfg.ObservationalMemory.Enabled,
		)
		retriever.WithRuntimeContext(runtimeAdapter)

		ctxAdapter := adk.NewContextAwareModelAdapter(modelAdapter, retriever, builder, logger())
		ctxAdapter.WithRuntimeAdapter(runtimeAdapter)

		// Wire in observational memory if available
		if mc != nil {
			ctxAdapter.WithMemory(mc.store, "")
		}

		// Wire in RAG if available and enabled
		if ec != nil && cfg.Embedding.RAG.Enabled {
			ragOpts := embedding.RetrieveOptions{
				Limit:       cfg.Embedding.RAG.MaxResults,
				Collections: cfg.Embedding.RAG.Collections,
			}
			if ragOpts.Limit <= 0 {
				ragOpts.Limit = 5
			}
			ctxAdapter.WithRAG(ec.ragService, ragOpts)
		}

		llm = ctxAdapter
	} else if mc != nil {
		// OM without knowledge system — create minimal context-aware adapter
		ctxAdapter := adk.NewContextAwareModelAdapter(modelAdapter, nil, builder, logger())
		ctxAdapter.WithMemory(mc.store, "")

		// Wire in RAG if available and enabled
		if ec != nil && cfg.Embedding.RAG.Enabled {
			ragOpts := embedding.RetrieveOptions{
				Limit:       cfg.Embedding.RAG.MaxResults,
				Collections: cfg.Embedding.RAG.Collections,
			}
			if ragOpts.Limit <= 0 {
				ragOpts.Limit = 5
			}
			ctxAdapter.WithRAG(ec.ragService, ragOpts)
		}

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
		AllowedOrigins:   cfg.Server.AllowedOrigins,
	}, adkAgent, nil, store, auth)
}
