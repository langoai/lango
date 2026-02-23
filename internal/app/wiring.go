package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"database/sql"

	"github.com/consensys/gnark/frontend"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/langoai/lango/internal/a2a"
	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	cronpkg "github.com/langoai/lango/internal/cron"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/embedding"
	"github.com/langoai/lango/internal/gateway"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/learning"
	"github.com/langoai/lango/internal/librarian"
	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/orchestration"
	"github.com/langoai/lango/internal/p2p"
	"github.com/langoai/lango/internal/p2p/discovery"
	"github.com/langoai/lango/internal/p2p/firewall"
	"github.com/langoai/lango/internal/p2p/handshake"
	"github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/p2p/paygate"
	p2pproto "github.com/langoai/lango/internal/p2p/protocol"
	"github.com/langoai/lango/internal/p2p/reputation"
	"github.com/langoai/lango/internal/payment"
	"github.com/langoai/lango/internal/payment/contracts"
	"github.com/langoai/lango/internal/prompt"
	"github.com/langoai/lango/internal/provider"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/skill"
	"github.com/langoai/lango/internal/supervisor"
	"github.com/langoai/lango/internal/wallet"
	"github.com/langoai/lango/internal/workflow"
	"github.com/langoai/lango/internal/zkp"
	"github.com/langoai/lango/internal/zkp/circuits"
	x402pkg "github.com/langoai/lango/internal/x402"
	"github.com/langoai/lango/skills"
	libp2pproto "github.com/libp2p/go-libp2p/core/protocol"
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

// buildSubAgentPromptFunc creates a SubAgentPromptFunc that injects shared
// prompt sections (Safety, ConversationRules) into each sub-agent's system
// prompt, with optional per-agent overrides from <promptsDir>/agents/<name>/.
func buildSubAgentPromptFunc(cfg *config.AgentConfig) orchestration.SubAgentPromptFunc {
	// Build a shared base containing only safety + conversation rules.
	// Sub-agents should NOT inherit the global identity or tool usage sections.
	shared := prompt.NewBuilder()

	if cfg.PromptsDir != "" {
		// Load shared sections from user's prompts directory.
		full := prompt.LoadFromDir(cfg.PromptsDir, logger())
		if full.Has(prompt.SectionSafety) {
			// Re-load: LoadFromDir returns a full builder. We extract only
			// what we need by building a fresh shared base from the directory.
			shared = prompt.LoadFromDir(cfg.PromptsDir, logger())
		}
	} else {
		shared = prompt.DefaultBuilder()
	}
	// Remove sections that are agent-global, not sub-agent shared.
	shared.Remove(prompt.SectionIdentity)
	shared.Remove(prompt.SectionToolUsage)

	return func(agentName, defaultInstruction string) string {
		b := shared.Clone()

		// Insert the spec's default instruction as agent identity (priority 150).
		b.Add(prompt.NewStaticSection(
			prompt.SectionAgentIdentity, 150, "", defaultInstruction,
		))

		// Apply per-agent overrides if the directory exists.
		if cfg.PromptsDir != "" {
			agentDir := filepath.Join(cfg.PromptsDir, "agents", agentName)
			b = prompt.LoadAgentFromDir(b, agentDir, logger())
		}

		return b.Build()
	}
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
	observer learning.ToolResultObserver
}

// initKnowledge creates the self-learning components if enabled.
// When gc is provided, a GraphEngine is used as the observer instead of the base Engine.
func initKnowledge(cfg *config.Config, store session.Store, gc *graphComponents) *knowledgeComponents {
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

	kStore := knowledge.NewStore(client, kLogger)

	engine := learning.NewEngine(kStore, kLogger)

	// Select observer: GraphEngine when graph store is available, otherwise base Engine.
	var observer learning.ToolResultObserver = engine
	if gc != nil {
		graphEngine := learning.NewGraphEngine(kStore, gc.store, kLogger)
		graphEngine.SetGraphCallback(func(triples []graph.Triple) {
			gc.buffer.Enqueue(graph.GraphRequest{Triples: triples})
		})
		observer = graphEngine
		logger().Info("graph-enhanced learning engine initialized")
	}

	logger().Info("knowledge system initialized")
	return &knowledgeComponents{
		store:    kStore,
		engine:   engine,
		observer: observer,
	}
}

// initSkills creates the file-based skill registry.
func initSkills(cfg *config.Config, baseTools []*agent.Tool) *skill.Registry {
	if !cfg.Skill.Enabled {
		logger().Info("skill system disabled")
		return nil
	}

	dir := cfg.Skill.SkillsDir
	if dir == "" {
		dir = "~/.lango/skills"
	}
	// Expand ~ to home directory.
	if len(dir) > 1 && dir[:2] == "~/" {
		if home, err := os.UserHomeDir(); err == nil {
			dir = filepath.Join(home, dir[2:])
		}
	}

	sLogger := logger()
	store := skill.NewFileSkillStore(dir, sLogger)

	// Deploy embedded default skills.
	defaultFS, err := skills.DefaultFS()
	if err == nil {
		if err := store.EnsureDefaults(defaultFS); err != nil {
			sLogger.Warnw("deploy default skills error", "error", err)
		}
	}

	registry := skill.NewRegistry(store, baseTools, sLogger)
	ctx := context.Background()
	if err := registry.LoadSkills(ctx); err != nil {
		sLogger.Warnw("load skills error", "error", err)
	}

	sLogger.Infow("skill system initialized", "dir", dir)
	return registry
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

// initConversationAnalysis creates the conversation analysis pipeline if both
// knowledge and observational memory are enabled.
func initConversationAnalysis(cfg *config.Config, sv *supervisor.Supervisor, store session.Store, kc *knowledgeComponents, gc *graphComponents) *learning.AnalysisBuffer {
	if kc == nil {
		return nil
	}
	if !cfg.ObservationalMemory.Enabled {
		return nil
	}

	// Create LLM proxy reusing the observational memory provider/model.
	omProvider := cfg.ObservationalMemory.Provider
	if omProvider == "" {
		omProvider = cfg.Agent.Provider
	}
	omModel := cfg.ObservationalMemory.Model
	if omModel == "" {
		omModel = cfg.Agent.Model
	}

	proxy := supervisor.NewProviderProxy(sv, omProvider, omModel)
	generator := &providerTextGenerator{proxy: proxy}

	aLogger := logger()

	analyzer := learning.NewConversationAnalyzer(generator, kc.store, aLogger)
	learner := learning.NewSessionLearner(generator, kc.store, aLogger)

	// Wire graph callbacks if graph store is available.
	if gc != nil && gc.buffer != nil {
		graphCB := func(triples []graph.Triple) {
			gc.buffer.Enqueue(graph.GraphRequest{Triples: triples})
		}
		analyzer.SetGraphCallback(graphCB)
		learner.SetGraphCallback(graphCB)
	}

	// Message provider.
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

	turnThreshold := cfg.Knowledge.AnalysisTurnThreshold
	tokenThreshold := cfg.Knowledge.AnalysisTokenThreshold

	buf := learning.NewAnalysisBuffer(analyzer, learner, getMessages, turnThreshold, tokenThreshold, aLogger)

	logger().Infow("conversation analysis initialized",
		"turnThreshold", turnThreshold,
		"tokenThreshold", tokenThreshold,
	)

	return buf
}

// graphComponents holds optional graph store components.
type graphComponents struct {
	store      graph.Store
	buffer     *graph.GraphBuffer
	ragService *graph.GraphRAGService
}

// initGraphStore creates the graph store if enabled.
func initGraphStore(cfg *config.Config) *graphComponents {
	if !cfg.Graph.Enabled {
		logger().Info("graph store disabled")
		return nil
	}

	dbPath := cfg.Graph.DatabasePath
	if dbPath == "" {
		// Default: graph.db next to session database.
		if cfg.Session.DatabasePath != "" {
			dbPath = filepath.Join(filepath.Dir(cfg.Session.DatabasePath), "graph.db")
		} else {
			dbPath = "graph.db"
		}
	}

	store, err := graph.NewBoltStore(dbPath)
	if err != nil {
		logger().Warnw("graph store init error, skipping", "error", err)
		return nil
	}

	buffer := graph.NewGraphBuffer(store, logger())

	logger().Infow("graph store initialized", "backend", "bolt", "path", dbPath)
	return &graphComponents{
		store:  store,
		buffer: buffer,
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
func initAgent(ctx context.Context, sv *supervisor.Supervisor, cfg *config.Config, store session.Store, tools []*agent.Tool, kc *knowledgeComponents, mc *memoryComponents, ec *embeddingComponents, gc *graphComponents, scanner *agent.SecretScanner, sr *skill.Registry, lc *librarianComponents) (*adk.Agent, error) {
	// Adapt tools to ADK format with optional per-tool timeout.
	toolTimeout := cfg.Agent.ToolTimeout
	var adkTools []adk_tool.Tool
	for _, t := range tools {
		at, err := adk.AdaptToolWithTimeout(t, toolTimeout)
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

	// Add automation prompt section if any automation system is enabled.
	if cfg.Cron.Enabled || cfg.Background.Enabled || cfg.Workflow.Enabled {
		builder.Add(buildAutomationPromptSection(cfg))
	}

	systemPrompt := builder.Build()

	// If knowledge is enabled, wrap with context-aware adapter
	var llm model.LLM = modelAdapter
	if kc != nil {
		retriever := knowledge.NewContextRetriever(
			kc.store,
			cfg.Knowledge.MaxContextPerLayer,
			logger(),
		)

		// Wire skill provider from file-based registry.
		if sr != nil {
			retriever.WithSkillProvider(&skillProviderAdapter{registry: sr})
		}

		// Wire inquiry provider from proactive librarian.
		if lc != nil {
			retriever.WithInquiryProvider(&inquiryProviderAdapter{store: lc.inquiryStore})
		}

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
			ctxAdapter.WithMemory(mc.store)

			// Apply memory context limits from config.
			maxRef := cfg.ObservationalMemory.MaxReflectionsInContext
			maxObs := cfg.ObservationalMemory.MaxObservationsInContext
			if maxRef <= 0 {
				maxRef = 5
			}
			if maxObs <= 0 {
				maxObs = 20
			}
			ctxAdapter.WithMemoryLimits(maxRef, maxObs)
		}

		// Wire in RAG if available and enabled
		if ec != nil && cfg.Embedding.RAG.Enabled {
			ragOpts := embedding.RetrieveOptions{
				Limit:       cfg.Embedding.RAG.MaxResults,
				Collections: cfg.Embedding.RAG.Collections,
				MaxDistance: cfg.Embedding.RAG.MaxDistance,
			}
			if ragOpts.Limit <= 0 {
				ragOpts.Limit = 5
			}
			ctxAdapter.WithRAG(ec.ragService, ragOpts)

			// Wire in Graph RAG if graph store is available.
			if gc != nil && gc.ragService != nil {
				ctxAdapter.WithGraphRAG(gc.ragService)
			}
		}

		llm = ctxAdapter
	} else if mc != nil {
		// OM without knowledge system — create minimal context-aware adapter
		ctxAdapter := adk.NewContextAwareModelAdapter(modelAdapter, nil, builder, logger())
		ctxAdapter.WithMemory(mc.store)

		// Apply memory context limits from config.
		maxRef := cfg.ObservationalMemory.MaxReflectionsInContext
		maxObs := cfg.ObservationalMemory.MaxObservationsInContext
		if maxRef <= 0 {
			maxRef = 5
		}
		if maxObs <= 0 {
			maxObs = 20
		}
		ctxAdapter.WithMemoryLimits(maxRef, maxObs)

		// Wire in RAG if available and enabled
		if ec != nil && cfg.Embedding.RAG.Enabled {
			ragOpts := embedding.RetrieveOptions{
				Limit:       cfg.Embedding.RAG.MaxResults,
				Collections: cfg.Embedding.RAG.Collections,
				MaxDistance: cfg.Embedding.RAG.MaxDistance,
			}
			if ragOpts.Limit <= 0 {
				ragOpts.Limit = 5
			}
			ctxAdapter.WithRAG(ec.ragService, ragOpts)

			// Wire in Graph RAG if graph store is available.
			if gc != nil && gc.ragService != nil {
				ctxAdapter.WithGraphRAG(gc.ragService)
			}
		}

		llm = ctxAdapter
	}

	// If PII redaction is enabled, wrap with PII-redacting adapter
	if cfg.Security.Interceptor.Enabled && cfg.Security.Interceptor.RedactPII {
		redactor := agent.NewPIIRedactor(agent.PIIConfig{
			RedactEmail:       true,
			RedactPhone:       true,
			CustomRegex:       cfg.Security.Interceptor.PIIRegexPatterns,
			DisabledBuiltins:  cfg.Security.Interceptor.PIIDisabledPatterns,
			CustomPatterns:    cfg.Security.Interceptor.PIICustomPatterns,
			PresidioEnabled:   cfg.Security.Interceptor.Presidio.Enabled,
			PresidioURL:       cfg.Security.Interceptor.Presidio.URL,
			PresidioThreshold: cfg.Security.Interceptor.Presidio.ScoreThreshold,
			PresidioLanguage:  cfg.Security.Interceptor.Presidio.Language,
		})
		llm = adk.NewPIIRedactingModelAdapter(llm, redactor, scanner)
		logger().Info("PII redaction interceptor enabled")
	}

	// Multi-agent orchestration mode.
	if cfg.Agent.MultiAgent {
		logger().Info("initializing multi-agent orchestration...")

		// Build orchestrator-specific prompt: strip tool-related sections that
		// cause the LLM to hallucinate agent names like "browser" or "exec".
		orchBuilder := buildPromptBuilder(&cfg.Agent)
		orchBuilder.Remove(prompt.SectionToolUsage)
		orchBuilder.Add(prompt.NewStaticSection(
			prompt.SectionIdentity, 100, "",
			"You are Lango, a production-grade AI assistant built for developers and teams.\n"+
				"You coordinate specialized sub-agents to handle tasks. "+
				"You do not have direct access to tools — delegate to sub-agents instead.",
		))
		orchestratorPrompt := orchBuilder.Build()

		orchCfg := orchestration.Config{
			Tools:               tools,
			Model:               llm,
			SystemPrompt:        orchestratorPrompt,
			AdaptTool:           adk.AdaptTool,
			MaxDelegationRounds: 5,
			SubAgentPrompt:      buildSubAgentPromptFunc(&cfg.Agent),
		}

		// Load remote A2A agents BEFORE building the tree so they are included.
		if cfg.A2A.Enabled && len(cfg.A2A.RemoteAgents) > 0 {
			remoteAgents, err := a2a.LoadRemoteAgents(cfg.A2A.RemoteAgents, logger())
			if err != nil {
				logger().Warnw("load remote A2A agents", "error", err)
			}
			if len(remoteAgents) > 0 {
				orchCfg.RemoteAgents = remoteAgents
			}
		}

		agentTree, err := orchestration.BuildAgentTree(orchCfg)
		if err != nil {
			return nil, fmt.Errorf("build agent tree: %w", err)
		}

		adkAgent, err := adk.NewAgentFromADK(agentTree, store)
		if err != nil {
			return nil, fmt.Errorf("adk multi-agent: %w", err)
		}
		return adkAgent, nil
	}

	// Single-agent mode (default).
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
		RequestTimeout:   cfg.Agent.RequestTimeout,
	}, adkAgent, nil, store, auth)
}

// wireGraphCallbacks connects graph store callbacks to knowledge and memory stores.
// It also creates the Entity Extractor pipeline and Memory GraphHooks.
func wireGraphCallbacks(gc *graphComponents, kc *knowledgeComponents, mc *memoryComponents, sv *supervisor.Supervisor, cfg *config.Config) {
	if gc == nil || gc.buffer == nil {
		return
	}

	// Create Entity Extractor for async triple extraction from content.
	var extractor *graph.Extractor
	if sv != nil {
		provider := cfg.Agent.Provider
		mdl := cfg.Agent.Model
		proxy := supervisor.NewProviderProxy(sv, provider, mdl)
		generator := &providerTextGenerator{proxy: proxy}
		extractor = graph.NewExtractor(generator, logger())
		logger().Info("graph entity extractor initialized")
	}

	graphCB := func(id, collection, content string, metadata map[string]string) {
		// Basic containment triple.
		gc.buffer.Enqueue(graph.GraphRequest{
			Triples: []graph.Triple{
				{
					Subject:   collection + ":" + id,
					Predicate: graph.Contains,
					Object:    "collection:" + collection,
					Metadata:  metadata,
				},
			},
		})

		// Async entity extraction via LLM.
		if extractor != nil && content != "" {
			go func() {
				ctx := context.Background()
				triples, err := extractor.Extract(ctx, content, id)
				if err != nil {
					logger().Debugw("entity extraction error", "id", id, "error", err)
					return
				}
				if len(triples) > 0 {
					gc.buffer.Enqueue(graph.GraphRequest{Triples: triples})
				}
			}()
		}
	}

	if kc != nil {
		kc.store.SetGraphCallback(graphCB)
	}
	if mc != nil {
		mc.store.SetGraphCallback(graphCB)

		// Wire Memory GraphHooks for temporal/session triples.
		tripleCallback := func(triples []graph.Triple) {
			gc.buffer.Enqueue(graph.GraphRequest{Triples: triples})
		}
		hooks := memory.NewGraphHooks(tripleCallback, logger())
		mc.store.SetGraphHooks(hooks)
		logger().Info("memory graph hooks wired")
	}
}

// initGraphRAG creates the Graph RAG service if both graph store and vector RAG are available.
func initGraphRAG(cfg *config.Config, gc *graphComponents, ec *embeddingComponents) {
	if gc == nil || ec == nil || ec.ragService == nil {
		return
	}

	maxDepth := cfg.Graph.MaxTraversalDepth
	if maxDepth <= 0 {
		maxDepth = 2
	}
	maxExpand := cfg.Graph.MaxExpansionResults
	if maxExpand <= 0 {
		maxExpand = 10
	}

	// Create a VectorRetriever adapter from embedding.RAGService.
	adapter := &ragServiceAdapter{inner: ec.ragService}

	gc.ragService = graph.NewGraphRAGService(adapter, gc.store, maxDepth, maxExpand, logger())
	logger().Info("graph RAG hybrid retrieval initialized")
}

// ragServiceAdapter adapts embedding.RAGService to graph.VectorRetriever interface.
type ragServiceAdapter struct {
	inner *embedding.RAGService
}

func (a *ragServiceAdapter) Retrieve(ctx context.Context, query string, opts graph.VectorRetrieveOptions) ([]graph.VectorResult, error) {
	embOpts := embedding.RetrieveOptions{
		Collections: opts.Collections,
		Limit:       opts.Limit,
		SessionKey:  opts.SessionKey,
		MaxDistance: opts.MaxDistance,
	}

	results, err := a.inner.Retrieve(ctx, query, embOpts)
	if err != nil {
		return nil, err
	}

	graphResults := make([]graph.VectorResult, len(results))
	for i, r := range results {
		graphResults[i] = graph.VectorResult{
			Collection: r.Collection,
			SourceID:   r.SourceID,
			Content:    r.Content,
			Distance:   r.Distance,
		}
	}
	return graphResults, nil
}

// paymentComponents holds optional blockchain payment components.
type paymentComponents struct {
	wallet  wallet.WalletProvider
	service *payment.Service
	limiter wallet.SpendingLimiter
	secrets *security.SecretsStore
	chainID int64
}

// initPayment creates the payment components if enabled.
// Follows the same graceful degradation pattern as initGraphStore.
func initPayment(cfg *config.Config, store session.Store, secrets *security.SecretsStore) *paymentComponents {
	if !cfg.Payment.Enabled {
		logger().Info("payment system disabled")
		return nil
	}

	if secrets == nil {
		logger().Warn("payment system requires security.signer, skipping")
		return nil
	}

	entStore, ok := store.(*session.EntStore)
	if !ok {
		logger().Warn("payment system requires EntStore, skipping")
		return nil
	}

	client := entStore.Client()

	// Create RPC client for blockchain interaction
	rpcClient, err := ethclient.Dial(cfg.Payment.Network.RPCURL)
	if err != nil {
		logger().Warnw("payment RPC connection failed, skipping", "error", err, "rpcUrl", cfg.Payment.Network.RPCURL)
		return nil
	}

	// Create wallet provider based on configuration
	var wp wallet.WalletProvider
	switch cfg.Payment.WalletProvider {
	case "local":
		wp = wallet.NewLocalWallet(secrets, cfg.Payment.Network.RPCURL, cfg.Payment.Network.ChainID)
	case "rpc":
		wp = wallet.NewRPCWallet()
	case "composite":
		local := wallet.NewLocalWallet(secrets, cfg.Payment.Network.RPCURL, cfg.Payment.Network.ChainID)
		rpc := wallet.NewRPCWallet()
		wp = wallet.NewCompositeWallet(rpc, local, nil)
	default:
		logger().Warnw("unknown wallet provider, using local", "provider", cfg.Payment.WalletProvider)
		wp = wallet.NewLocalWallet(secrets, cfg.Payment.Network.RPCURL, cfg.Payment.Network.ChainID)
	}

	// Create spending limiter
	limiter, err := wallet.NewEntSpendingLimiter(client,
		cfg.Payment.Limits.MaxPerTx,
		cfg.Payment.Limits.MaxDaily,
	)
	if err != nil {
		logger().Warnw("spending limiter init failed, skipping", "error", err)
		return nil
	}

	// Create transaction builder
	builder := payment.NewTxBuilder(rpcClient,
		cfg.Payment.Network.ChainID,
		cfg.Payment.Network.USDCContract,
	)

	// Create payment service
	svc := payment.NewService(wp, limiter, builder, client, rpcClient, cfg.Payment.Network.ChainID)

	logger().Infow("payment system initialized",
		"walletProvider", cfg.Payment.WalletProvider,
		"chainId", cfg.Payment.Network.ChainID,
		"network", wallet.NetworkName(cfg.Payment.Network.ChainID),
		"maxPerTx", cfg.Payment.Limits.MaxPerTx,
		"maxDaily", cfg.Payment.Limits.MaxDaily,
	)

	return &paymentComponents{
		wallet:  wp,
		service: svc,
		limiter: limiter,
		secrets: secrets,
		chainID: cfg.Payment.Network.ChainID,
	}
}

// x402Components holds optional X402 interceptor components.
type x402Components struct {
	interceptor *x402pkg.Interceptor
}

// initX402 creates the X402 interceptor if payment is enabled.
func initX402(cfg *config.Config, secrets *security.SecretsStore, limiter wallet.SpendingLimiter) *x402Components {
	if !cfg.Payment.Enabled {
		return nil
	}
	if secrets == nil {
		return nil
	}

	signerProvider := x402pkg.NewLocalSignerProvider(secrets)

	maxAutoPayAmt := cfg.Payment.Limits.MaxPerTx
	if maxAutoPayAmt == "" {
		maxAutoPayAmt = "1.00"
	}

	x402Cfg := x402pkg.Config{
		Enabled:          true,
		ChainID:          cfg.Payment.Network.ChainID,
		MaxAutoPayAmount: maxAutoPayAmt,
	}

	interceptor := x402pkg.NewInterceptor(signerProvider, limiter, x402Cfg, logger())

	logger().Infow("X402 interceptor configured",
		"chainId", x402Cfg.ChainID,
		"maxAutoPayAmount", maxAutoPayAmt,
	)

	return &x402Components{
		interceptor: interceptor,
	}
}

// p2pComponents holds optional P2P networking components.
type p2pComponents struct {
	node       *p2p.Node
	sessions   *handshake.SessionStore
	handshaker *handshake.Handshaker
	fw         *firewall.Firewall
	gossip     *discovery.GossipService
	identity   *identity.WalletDIDProvider
	handler    *p2pproto.Handler
	payGate    *paygate.Gate
	reputation *reputation.Store
	pricingCfg config.P2PPricingConfig
}

// initP2P creates the P2P networking components if enabled.
func initP2P(cfg *config.Config, wp wallet.WalletProvider, pc *paymentComponents, dbClient *ent.Client) *p2pComponents {
	if !cfg.P2P.Enabled {
		logger().Info("P2P networking disabled")
		return nil
	}

	if wp == nil {
		logger().Warn("P2P networking requires wallet provider, skipping")
		return nil
	}

	pLogger := logger()

	// Create P2P node.
	node, err := p2p.NewNode(cfg.P2P, pLogger)
	if err != nil {
		pLogger.Warnw("P2P node creation failed, skipping", "error", err)
		return nil
	}

	// Create identity provider from wallet.
	idProvider := identity.NewProvider(wp, pLogger)

	// Create session store.
	sessionTTL := cfg.P2P.SessionTokenTTL
	if sessionTTL <= 0 {
		sessionTTL = 24 * time.Hour
	}
	sessions, err := handshake.NewSessionStore(sessionTTL)
	if err != nil {
		pLogger.Warnw("P2P session store creation failed, skipping", "error", err)
		return nil
	}

	// Initialize ZKP prover (optional).
	zkProver := initZKP(cfg)

	// Create handshaker.
	hsTimeout := cfg.P2P.HandshakeTimeout
	if hsTimeout <= 0 {
		hsTimeout = 30 * time.Second
	}
	hsCfg := handshake.Config{
		Wallet:           wp,
		Sessions:         sessions,
		ZKEnabled:        cfg.P2P.ZKHandshake,
		Timeout:          hsTimeout,
		AutoApproveKnown: cfg.P2P.AutoApproveKnownPeers,
		Logger:           pLogger,
	}

	// Wire ZK prover/verifier into handshake if available.
	if zkProver != nil && cfg.P2P.ZKHandshake {
		hsCfg.ZKProver = func(ctx context.Context, challenge []byte) ([]byte, error) {
			assignment := &circuits.WalletOwnershipCircuit{
				Challenge: challenge,
				Response:  challenge, // simplified: use challenge as witness in MVP
			}
			proof, err := zkProver.Prove(ctx, "wallet_ownership", assignment)
			if err != nil {
				return nil, err
			}
			return proof.Data, nil
		}
		hsCfg.ZKVerifier = func(ctx context.Context, proof, challenge, publicKey []byte) (bool, error) {
			p := &zkp.Proof{
				CircuitID: "wallet_ownership",
				Data:      proof,
				Scheme:    zkProver.Scheme(),
			}
			return zkProver.Verify(ctx, p, &circuits.WalletOwnershipCircuit{})
		}
		pLogger.Info("ZK handshake prover/verifier wired")
	}

	handshaker := handshake.NewHandshaker(hsCfg)

	// Create firewall.
	var aclRules []firewall.ACLRule
	for _, r := range cfg.P2P.FirewallRules {
		aclRules = append(aclRules, firewall.ACLRule{
			PeerDID:   r.PeerDID,
			Action:    r.Action,
			Tools:     r.Tools,
			RateLimit: r.RateLimit,
		})
	}
	fw := firewall.New(aclRules, pLogger)

	// Wire Owner Shield if configured.
	ownerCfg := cfg.P2P.OwnerProtection
	if ownerCfg.OwnerName != "" || ownerCfg.OwnerEmail != "" || ownerCfg.OwnerPhone != "" {
		blockConv := true
		if ownerCfg.BlockConversations != nil {
			blockConv = *ownerCfg.BlockConversations
		}
		shield := firewall.NewOwnerShield(firewall.OwnerProtectionConfig{
			OwnerName:          ownerCfg.OwnerName,
			OwnerEmail:         ownerCfg.OwnerEmail,
			OwnerPhone:         ownerCfg.OwnerPhone,
			ExtraTerms:         ownerCfg.ExtraTerms,
			BlockConversations: blockConv,
		}, pLogger)
		fw.SetOwnerShield(shield)
		pLogger.Info("P2P owner data shield enabled")
	}

	// Wire ZK attestation into firewall if available.
	if zkProver != nil && cfg.P2P.ZKAttestation {
		fw.SetZKAttestFunc(func(responseHash, agentDIDHash []byte) ([]byte, error) {
			assignment := &circuits.ResponseAttestationCircuit{
				ResponseHash: responseHash,
				AgentDIDHash: agentDIDHash,
				Timestamp:    time.Now().Unix(),
			}
			proof, err := zkProver.Prove(context.Background(), "response_attestation", assignment)
			if err != nil {
				return nil, err
			}
			return proof.Data, nil
		})
		pLogger.Info("ZK response attestation wired to firewall")
	}

	// Wire reputation system if DB client is available.
	var repStore *reputation.Store
	if dbClient != nil {
		repStore = reputation.NewStore(dbClient, pLogger)
		minScore := cfg.P2P.MinTrustScore
		if minScore <= 0 {
			minScore = 0.3
		}
		fw.SetReputationChecker(func(ctx context.Context, peerDID string) (float64, error) {
			return repStore.GetScore(ctx, peerDID)
		}, minScore)
		pLogger.Infow("P2P reputation system enabled", "minTrustScore", minScore)
	}

	// Register handshake protocol handler on the host.
	node.Host().SetStreamHandler(libp2pproto.ID(handshake.ProtocolID), handshaker.StreamHandler())

	// Get local DID for protocol handler.
	var localDID string
	ctx := context.Background()
	d, err := idProvider.DID(ctx)
	if err == nil && d != nil {
		localDID = d.ID
	}

	// Create A2A-over-P2P protocol handler.
	handler := p2pproto.NewHandler(p2pproto.HandlerConfig{
		Sessions: sessions,
		Firewall: fw,
		LocalDID: localDID,
		Logger:   pLogger,
	})
	node.Host().SetStreamHandler(libp2pproto.ID(p2pproto.ProtocolID), handler.StreamHandler())

	// Create gossip discovery service.
	var gossip *discovery.GossipService
	gossipInterval := cfg.P2P.GossipInterval
	if gossipInterval <= 0 {
		gossipInterval = 30 * time.Second
	}

	agentName := cfg.A2A.AgentName
	if agentName == "" {
		agentName = "lango"
	}
	// Wire payment gate if pricing is enabled.
	var pg *paygate.Gate
	if cfg.P2P.Pricing.Enabled && pc != nil {
		walletAddr := ""
		ctx2 := context.Background()
		if a, err := wp.Address(ctx2); err == nil {
			walletAddr = a
		}
		usdcAddr, _ := contracts.LookupUSDC(pc.chainID)

		pricingFn := func(toolName string) (string, bool) {
			if price, ok := cfg.P2P.Pricing.ToolPrices[toolName]; ok {
				return price, false
			}
			if cfg.P2P.Pricing.PerQuery != "" {
				return cfg.P2P.Pricing.PerQuery, false
			}
			return "", true // free by default
		}

		pg = paygate.New(paygate.Config{
			PricingFn: pricingFn,
			LocalAddr: walletAddr,
			ChainID:   pc.chainID,
			USDCAddr:  usdcAddr,
			Logger:    pLogger,
		})

		// Wire PayGate to handler via adapter.
		handler.SetPayGate(&payGateAdapter{gate: pg, chainID: pc.chainID, usdcAddr: usdcAddr})
		pLogger.Infow("P2P payment gate enabled",
			"perQuery", cfg.P2P.Pricing.PerQuery,
			"toolPrices", len(cfg.P2P.Pricing.ToolPrices),
		)
	}

	localCard := &discovery.GossipCard{
		Name:   agentName,
		DID:    localDID,
		PeerID: node.PeerID().String(),
	}
	for _, a := range node.Multiaddrs() {
		localCard.Multiaddrs = append(localCard.Multiaddrs, a.String())
	}

	// Set pricing info on gossip card if pricing is enabled.
	if cfg.P2P.Pricing.Enabled {
		localCard.Pricing = &discovery.PricingInfo{
			Currency:   "USDC",
			PerQuery:   cfg.P2P.Pricing.PerQuery,
			ToolPrices: cfg.P2P.Pricing.ToolPrices,
		}
	}

	gossip, err = discovery.NewGossipService(discovery.GossipConfig{
		Host:      node.Host(),
		LocalCard: localCard,
		Interval:  gossipInterval,
		Logger:    pLogger,
	})
	if err != nil {
		pLogger.Warnw("gossip service creation failed", "error", err)
	}

	pLogger.Infow("P2P networking initialized",
		"peerID", node.PeerID(),
		"did", localDID,
		"listenAddrs", cfg.P2P.ListenAddrs,
		"zkHandshake", cfg.P2P.ZKHandshake,
		"firewallRules", len(aclRules),
	)

	return &p2pComponents{
		node:       node,
		sessions:   sessions,
		handshaker: handshaker,
		fw:         fw,
		gossip:     gossip,
		identity:   idProvider,
		handler:    handler,
		payGate:    pg,
		reputation: repStore,
		pricingCfg: cfg.P2P.Pricing,
	}
}

// payGateAdapter adapts paygate.Gate to protocol.PayGateChecker.
type payGateAdapter struct {
	gate     *paygate.Gate
	chainID  int64
	usdcAddr common.Address
}

func (a *payGateAdapter) Check(peerDID, toolName string, payload map[string]interface{}) (p2pproto.PayGateResult, error) {
	result, err := a.gate.Check(peerDID, toolName, payload)
	if err != nil {
		return p2pproto.PayGateResult{}, err
	}
	pgr := p2pproto.PayGateResult{
		Status: string(result.Status),
	}
	if result.Auth != nil {
		pgr.Auth = result.Auth
	}
	if result.PriceQuote != nil {
		pgr.PriceQuote = map[string]interface{}{
			"toolName":     result.PriceQuote.ToolName,
			"price":        result.PriceQuote.Price,
			"currency":     result.PriceQuote.Currency,
			"usdcContract": result.PriceQuote.USDCContract,
			"chainId":      result.PriceQuote.ChainID,
			"sellerAddr":   result.PriceQuote.SellerAddr,
			"quoteExpiry":  result.PriceQuote.QuoteExpiry,
			"isFree":       false,
		}
	}
	return pgr, nil
}

// initZKP creates ZKP components if enabled.
func initZKP(cfg *config.Config) *zkp.ProverService {
	if !cfg.P2P.ZKHandshake && !cfg.P2P.ZKAttestation {
		return nil
	}

	prover, err := zkp.NewProverService(zkp.Config{
		CacheDir: cfg.P2P.ZKP.ProofCacheDir,
		Scheme:   cfg.P2P.ZKP.ProvingScheme,
		Logger:   logger(),
	})
	if err != nil {
		logger().Warnw("ZKP prover init error, skipping", "error", err)
		return nil
	}

	// Compile all 4 circuits.
	circuitDefs := map[string]interface {
		Define(frontend.API) error
	}{
		"wallet_ownership":     &circuits.WalletOwnershipCircuit{},
		"response_attestation": &circuits.ResponseAttestationCircuit{},
		"balance_range":        &circuits.BalanceRangeCircuit{},
		"agent_capability":     &circuits.AgentCapabilityCircuit{},
	}

	for id, circuit := range circuitDefs {
		if err := prover.Compile(id, circuit); err != nil {
			logger().Warnw("ZKP circuit compile error", "circuitID", id, "error", err)
		}
	}

	logger().Infow("ZKP prover initialized",
		"scheme", prover.Scheme(),
		"circuits", len(circuitDefs),
	)
	return prover
}

// agentRunnerAdapter adapts app.runAgent to cron.AgentRunner / background.AgentRunner / workflow.AgentRunner.
type agentRunnerAdapter struct {
	app *App
}

func (r *agentRunnerAdapter) Run(ctx context.Context, sessionKey, promptText string) (string, error) {
	return r.app.runAgent(ctx, sessionKey, promptText)
}

// initCron creates the cron scheduling system if enabled.
func initCron(cfg *config.Config, store session.Store, app *App) *cronpkg.Scheduler {
	if !cfg.Cron.Enabled {
		logger().Info("cron scheduling disabled")
		return nil
	}

	entStore, ok := store.(*session.EntStore)
	if !ok {
		logger().Warn("cron scheduling requires EntStore, skipping")
		return nil
	}

	client := entStore.Client()
	cronStore := cronpkg.NewEntStore(client)
	sender := newChannelSender(app)
	delivery := cronpkg.NewDelivery(sender, sender, logger())
	runner := &agentRunnerAdapter{app: app}
	executor := cronpkg.NewExecutor(runner, delivery, cronStore, logger())

	maxJobs := cfg.Cron.MaxConcurrentJobs
	if maxJobs <= 0 {
		maxJobs = 5
	}

	tz := cfg.Cron.Timezone
	if tz == "" {
		tz = "UTC"
	}

	scheduler := cronpkg.New(cronStore, executor, tz, maxJobs, logger())

	logger().Infow("cron scheduling initialized",
		"timezone", tz,
		"maxConcurrentJobs", maxJobs,
	)

	return scheduler
}

// initBackground creates the background task manager if enabled.
func initBackground(cfg *config.Config, app *App) *background.Manager {
	if !cfg.Background.Enabled {
		logger().Info("background tasks disabled")
		return nil
	}

	runner := &agentRunnerAdapter{app: app}
	sender := newChannelSender(app)
	notify := background.NewNotification(sender, sender, logger())

	maxTasks := cfg.Background.MaxConcurrentTasks
	if maxTasks <= 0 {
		maxTasks = 3
	}

	taskTimeout := cfg.Background.TaskTimeout
	if taskTimeout <= 0 {
		taskTimeout = 30 * time.Minute
	}

	mgr := background.NewManager(runner, notify, maxTasks, taskTimeout, logger())

	logger().Infow("background task manager initialized",
		"maxConcurrentTasks", maxTasks,
		"yieldMs", cfg.Background.YieldMs,
	)

	return mgr
}

// initWorkflow creates the workflow engine if enabled.
func initWorkflow(cfg *config.Config, store session.Store, app *App) *workflow.Engine {
	if !cfg.Workflow.Enabled {
		logger().Info("workflow engine disabled")
		return nil
	}

	entStore, ok := store.(*session.EntStore)
	if !ok {
		logger().Warn("workflow engine requires EntStore, skipping")
		return nil
	}

	client := entStore.Client()
	state := workflow.NewStateStore(client, logger())
	runner := &agentRunnerAdapter{app: app}
	sender := newChannelSender(app)

	maxConcurrent := cfg.Workflow.MaxConcurrentSteps
	if maxConcurrent <= 0 {
		maxConcurrent = 4
	}

	defaultTimeout := cfg.Workflow.DefaultTimeout
	if defaultTimeout <= 0 {
		defaultTimeout = 10 * time.Minute
	}

	engine := workflow.NewEngine(runner, state, sender, maxConcurrent, defaultTimeout, logger())

	logger().Infow("workflow engine initialized",
		"maxConcurrentSteps", maxConcurrent,
		"defaultTimeout", defaultTimeout,
	)

	return engine
}

// librarianComponents holds optional proactive librarian components.
type librarianComponents struct {
	inquiryStore    *librarian.InquiryStore
	proactiveBuffer *librarian.ProactiveBuffer
}

// initLibrarian creates the proactive librarian components if enabled.
// Requires: librarian.enabled && knowledge.enabled && observationalMemory.enabled.
func initLibrarian(
	cfg *config.Config,
	sv *supervisor.Supervisor,
	store session.Store,
	kc *knowledgeComponents,
	mc *memoryComponents,
	gc *graphComponents,
) *librarianComponents {
	if !cfg.Librarian.Enabled {
		logger().Info("proactive librarian disabled")
		return nil
	}
	if kc == nil {
		logger().Warn("proactive librarian requires knowledge system, skipping")
		return nil
	}
	if mc == nil {
		logger().Warn("proactive librarian requires observational memory, skipping")
		return nil
	}

	entStore, ok := store.(*session.EntStore)
	if !ok {
		logger().Warn("proactive librarian requires EntStore, skipping")
		return nil
	}

	client := entStore.Client()
	lLogger := logger()

	inquiryStore := librarian.NewInquiryStore(client, lLogger)

	// Create LLM proxy.
	provider := cfg.Librarian.Provider
	if provider == "" {
		provider = cfg.Agent.Provider
	}
	lModel := cfg.Librarian.Model
	if lModel == "" {
		lModel = cfg.Agent.Model
	}

	proxy := supervisor.NewProviderProxy(sv, provider, lModel)
	generator := &providerTextGenerator{proxy: proxy}

	analyzer := librarian.NewObservationAnalyzer(generator, lLogger)
	processor := librarian.NewInquiryProcessor(generator, inquiryStore, kc.store, lLogger)

	// Message provider.
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

	// Observation provider.
	getObservations := librarian.ObservationProvider(mc.store.ListObservations)

	bufCfg := librarian.ProactiveBufferConfig{
		ObservationThreshold: cfg.Librarian.ObservationThreshold,
		CooldownTurns:        cfg.Librarian.InquiryCooldownTurns,
		MaxPending:           cfg.Librarian.MaxPendingInquiries,
		AutoSaveConfidence:   cfg.Librarian.AutoSaveConfidence,
	}
	buffer := librarian.NewProactiveBuffer(
		analyzer, processor, inquiryStore, kc.store,
		getMessages, getObservations, bufCfg, lLogger,
	)

	// Wire graph callback if available.
	if gc != nil && gc.buffer != nil {
		buffer.SetGraphCallback(func(triples []librarian.Triple) {
			graphTriples := make([]graph.Triple, len(triples))
			for i, t := range triples {
				graphTriples[i] = graph.Triple{
					Subject:   t.Subject,
					Predicate: t.Predicate,
					Object:    t.Object,
					Metadata:  t.Metadata,
				}
			}
			gc.buffer.Enqueue(graph.GraphRequest{Triples: graphTriples})
		})
	}

	logger().Infow("proactive librarian initialized",
		"provider", provider,
		"model", lModel,
		"observationThreshold", bufCfg.ObservationThreshold,
		"cooldownTurns", bufCfg.CooldownTurns,
		"maxPending", bufCfg.MaxPending,
	)

	return &librarianComponents{
		inquiryStore:    inquiryStore,
		proactiveBuffer: buffer,
	}
}

// inquiryProviderAdapter bridges librarian.InquiryStore → knowledge.InquiryProvider.
type inquiryProviderAdapter struct {
	store *librarian.InquiryStore
}

func (a *inquiryProviderAdapter) PendingInquiryItems(ctx context.Context, sessionKey string, limit int) ([]knowledge.ContextItem, error) {
	inquiries, err := a.store.ListPendingInquiries(ctx, sessionKey, limit)
	if err != nil {
		return nil, err
	}

	items := make([]knowledge.ContextItem, 0, len(inquiries))
	for _, inq := range inquiries {
		items = append(items, knowledge.ContextItem{
			Layer:   knowledge.LayerPendingInquiries,
			Key:     inq.Topic,
			Content: inq.Question,
			Source:  inq.Context,
		})
	}
	return items, nil
}

// skillProviderAdapter adapts *skill.Registry to knowledge.SkillProvider.
type skillProviderAdapter struct {
	registry *skill.Registry
}

func (a *skillProviderAdapter) ListActiveSkillInfos(ctx context.Context) ([]knowledge.SkillInfo, error) {
	entries, err := a.registry.ListActiveSkills(ctx)
	if err != nil {
		return nil, err
	}
	infos := make([]knowledge.SkillInfo, len(entries))
	for i, e := range entries {
		infos[i] = knowledge.SkillInfo{
			Name:        e.Name,
			Description: e.Description,
			Type:        string(e.Type),
		}
	}
	return infos, nil
}

// buildAutomationPromptSection creates a dynamic prompt section describing
// available automation capabilities (cron, background, workflow).
func buildAutomationPromptSection(cfg *config.Config) *prompt.StaticSection {
	var parts []string

	parts = append(parts, "## Automation Capabilities\n")
	parts = append(parts, "You have access to automation tools that let you schedule recurring tasks, run background jobs, and execute multi-step workflows.\n")

	if cfg.Cron.Enabled {
		parts = append(parts, `### Cron Scheduling
- Use cron_add to create scheduled jobs (e.g., "매일 아침 9시에 뉴스 요약" → cron_add with schedule_type=cron, schedule="0 9 * * *")
- Schedule types: cron (crontab expression), every (interval like "1h"), at (one-time RFC3339 datetime)
- deliver_to is optional — if omitted, the current channel is auto-detected as channel:id (e.g. telegram:CHAT_ID)
- Use cron_list to show all jobs, cron_pause/cron_resume to toggle, cron_remove to delete
- Use cron_history to check execution history
`)
	}

	if cfg.Background.Enabled {
		parts = append(parts, `### Background Tasks
- Use bg_submit to run a prompt asynchronously (returns immediately with task_id)
- channel is optional — if omitted, the current channel is auto-detected as channel:id (e.g. telegram:CHAT_ID)
- Use bg_status/bg_result to check progress and retrieve results
- Use bg_list to see all tasks, bg_cancel to stop a task
`)
	}

	if cfg.Workflow.Enabled {
		parts = append(parts, `### Workflow Pipelines
- Use workflow_run to execute a multi-step workflow from YAML (file path or inline content)
- deliver_to in YAML is optional — if omitted, the current channel is auto-detected as channel:id (e.g. telegram:CHAT_ID)
- Use workflow_status to monitor progress, workflow_list for recent runs
- Use workflow_cancel to stop a running workflow
- Use workflow_save to persist a workflow YAML for future use
`)
	}

	parts = append(parts, `### Important
- ALWAYS use the built-in automation tools above. NEVER use exec to run "lango cron", "lango bg", or "lango workflow" commands — this will fail because spawning a new lango process requires passphrase authentication.
`)

	content := strings.Join(parts, "\n")
	return prompt.NewStaticSection(prompt.SectionAutomation, 450, "Automation", content)
}
