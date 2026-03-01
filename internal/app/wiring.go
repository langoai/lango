package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/langoai/lango/internal/a2a"
	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/embedding"
	"github.com/langoai/lango/internal/gateway"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/orchestration"
	"github.com/langoai/lango/internal/prompt"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/skill"
	"github.com/langoai/lango/internal/supervisor"
	"github.com/langoai/lango/internal/toolcatalog"
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

	case "aws-kms", "gcp-kms", "azure-kv", "pkcs11":
		kmsProvider, err := security.NewKMSProvider(security.KMSProviderName(cfg.Security.Signer.Provider), cfg.Security.KMS)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("KMS provider %q: %w", cfg.Security.Signer.Provider, err)
		}

		if boot == nil || boot.DBClient == nil {
			return nil, nil, nil, fmt.Errorf("KMS security provider requires bootstrap")
		}

		keys := security.NewKeyRegistry(boot.DBClient)
		ctx := context.Background()
		if _, err := keys.RegisterKey(ctx, "kms-default", cfg.Security.KMS.KeyID, security.KeyTypeEncryption); err != nil {
			return nil, nil, nil, fmt.Errorf("register KMS key: %w", err)
		}

		var finalProvider = kmsProvider

		// Wrap with CompositeCryptoProvider for fallback when configured.
		if cfg.Security.KMS.FallbackToLocal && boot.Crypto != nil {
			checker := security.NewKMSHealthChecker(kmsProvider, cfg.Security.KMS.KeyID, 0)
			finalProvider = security.NewCompositeCryptoProvider(kmsProvider, boot.Crypto, checker)
			logger().Infow("security initialized (KMS provider with local fallback)",
				"provider", cfg.Security.Signer.Provider,
				"keyID", cfg.Security.KMS.KeyID)
		} else {
			logger().Infow("security initialized (KMS provider)",
				"provider", cfg.Security.Signer.Provider,
				"keyID", cfg.Security.KMS.KeyID)
		}

		secrets := security.NewSecretsStore(boot.DBClient, keys, finalProvider)
		return finalProvider, keys, secrets, nil

	default:
		return nil, nil, nil, fmt.Errorf("unknown security provider: %s", cfg.Security.Signer.Provider)
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
func initAgent(ctx context.Context, sv *supervisor.Supervisor, cfg *config.Config, store session.Store, tools []*agent.Tool, kc *knowledgeComponents, mc *memoryComponents, ec *embeddingComponents, gc *graphComponents, scanner *agent.SecretScanner, sr *skill.Registry, lc *librarianComponents, catalog *toolcatalog.Catalog) (*adk.Agent, error) {
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
			if cfg.ObservationalMemory.MemoryTokenBudget > 0 {
				ctxAdapter.WithMemoryTokenBudget(cfg.ObservationalMemory.MemoryTokenBudget)
			}
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
		if cfg.ObservationalMemory.MemoryTokenBudget > 0 {
			ctxAdapter.WithMemoryTokenBudget(cfg.ObservationalMemory.MemoryTokenBudget)
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
		orchIdentity := "You are Lango, a production-grade AI assistant built for developers and teams.\n" +
			"You coordinate specialized sub-agents to handle tasks."
		if catalog != nil && catalog.ToolCount() > 0 {
			orchIdentity += " You also have builtin_list and builtin_invoke tools for direct access to any registered built-in tool."
		} else {
			orchIdentity += " You do not have direct access to tools — delegate to sub-agents instead."
		}
		orchBuilder.Add(prompt.NewStaticSection(
			prompt.SectionIdentity, 100, "",
			orchIdentity,
		))
		orchestratorPrompt := orchBuilder.Build()

		var universalTools []*agent.Tool
		if catalog != nil {
			universalTools = toolcatalog.BuildDispatcher(catalog)
		}

		orchCfg := orchestration.Config{
			Tools:               tools,
			Model:               llm,
			SystemPrompt:        orchestratorPrompt,
			AdaptTool:           adk.AdaptTool,
			MaxDelegationRounds: cfg.Agent.MaxDelegationRounds,
			SubAgentPrompt:      buildSubAgentPromptFunc(&cfg.Agent),
			UniversalTools:      universalTools,
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

		// Build agent options for multi-agent mode.
		agentOpts := buildAgentOptions(cfg, kc)
		adkAgent, err := adk.NewAgentFromADK(agentTree, store, agentOpts...)
		if err != nil {
			return nil, fmt.Errorf("adk multi-agent: %w", err)
		}
		return adkAgent, nil
	}

	// Single-agent mode (default).
	logger().Info("initializing agent runtime (ADK)...")
	agentOpts := buildAgentOptions(cfg, kc)
	adkAgent, err := adk.NewAgent(ctx, adkTools, llm, systemPrompt, store, agentOpts...)
	if err != nil {
		return nil, fmt.Errorf("adk agent: %w", err)
	}
	return adkAgent, nil
}

// buildAgentOptions constructs AgentOption slice from config and knowledge components.
func buildAgentOptions(cfg *config.Config, kc *knowledgeComponents) []adk.AgentOption {
	var opts []adk.AgentOption

	// Token budget derived from the configured model.
	opts = append(opts, adk.WithAgentTokenBudget(adk.ModelTokenBudget(cfg.Agent.Model)))

	// Max turns: use explicit config if set, otherwise raise default for multi-agent mode
	// where delegation overhead consumes more turns.
	if cfg.Agent.MaxTurns > 0 {
		opts = append(opts, adk.WithAgentMaxTurns(cfg.Agent.MaxTurns))
	} else if cfg.Agent.MultiAgent {
		opts = append(opts, adk.WithAgentMaxTurns(50))
	}

	// Error correction: enabled by default when knowledge system is available.
	errorCorrectionEnabled := true
	if cfg.Agent.ErrorCorrectionEnabled != nil {
		errorCorrectionEnabled = *cfg.Agent.ErrorCorrectionEnabled
	}
	if errorCorrectionEnabled && kc != nil && kc.engine != nil {
		opts = append(opts, adk.WithAgentErrorFixProvider(kc.engine))
	}

	return opts
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
- ALWAYS use the built-in tools. NEVER use exec to run ANY "lango" CLI command — this includes "lango cron", "lango bg", "lango workflow", "lango graph", "lango memory", "lango p2p", "lango security", "lango payment", "lango config", "lango doctor", or any other subcommand. Every lango CLI invocation requires passphrase authentication during bootstrap and will fail when spawned as a non-interactive subprocess.
- If you need functionality without a built-in tool equivalent (e.g., config management, diagnostics), ask the user to run the command in their terminal.
`)

	content := strings.Join(parts, "\n")
	return prompt.NewStaticSection(prompt.SectionAutomation, 450, "Automation", content)
}
