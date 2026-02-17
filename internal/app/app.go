package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/langowarny/lango/internal/a2a"
	"github.com/langowarny/lango/internal/agent"
	"github.com/langowarny/lango/internal/approval"
	"github.com/langowarny/lango/internal/bootstrap"
	"github.com/langowarny/lango/internal/config"
	"github.com/langowarny/lango/internal/logging"
	"github.com/langowarny/lango/internal/security"
	"github.com/langowarny/lango/internal/session"
	"github.com/langowarny/lango/internal/tools/browser"
	"github.com/langowarny/lango/internal/tools/filesystem"
)

func logger() *zap.SugaredLogger { return logging.App() }

// New creates a new application instance from a bootstrap result.
func New(boot *bootstrap.Result) (*App, error) {
	cfg := boot.Config
	app := &App{
		Config: cfg,
	}

	// 1. Supervisor (holds provider secrets, exec tool)
	sv, err := initSupervisor(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create supervisor: %w", err)
	}

	// 2. Session Store — reuse the DB client opened during bootstrap.
	store, err := initSessionStore(cfg, boot)
	if err != nil {
		return nil, fmt.Errorf("failed to create session store: %w", err)
	}
	app.Store = store

	// 3. Security — reuse the crypto provider initialized during bootstrap.
	crypto, keys, secrets, err := initSecurity(cfg, store, boot)
	if err != nil {
		return nil, fmt.Errorf("security init: %w", err)
	}
	app.Crypto = crypto
	app.Keys = keys
	app.Secrets = secrets

	// 4. Base tools (exec + filesystem + optional browser)
	// Block agent access to the ~/.lango/ directory.
	var blockedPaths []string
	if home, err := os.UserHomeDir(); err == nil {
		blockedPaths = append(blockedPaths,
			filepath.Join(home, ".lango")+string(os.PathSeparator))
	}
	fsConfig := filesystem.Config{
		MaxReadSize:  cfg.Tools.Filesystem.MaxReadSize,
		AllowedPaths: cfg.Tools.Filesystem.AllowedPaths,
		BlockedPaths: blockedPaths,
	}

	var browserSM *browser.SessionManager
	if cfg.Tools.Browser.Enabled {
		bt, err := browser.New(browser.Config{
			Headless:       cfg.Tools.Browser.Headless,
			BrowserBin:     cfg.Tools.Browser.BrowserBin,
			SessionTimeout: cfg.Tools.Browser.SessionTimeout,
		})
		if err != nil {
			return nil, fmt.Errorf("create browser tool: %w", err)
		}
		browserSM = browser.NewSessionManager(bt)
		app.Browser = browserSM
		logger().Info("browser tools enabled")
	}

	tools := buildTools(sv, fsConfig, browserSM)

	// 4b. Crypto/Secrets tools (if security is enabled)
	// RefStore holds opaque references; plaintext never reaches agent context.
	// SecretScanner detects leaked secrets in model output.
	refs := security.NewRefStore()
	scanner := agent.NewSecretScanner()

	// Register config secrets to prevent leakage in model output.
	registerConfigSecrets(scanner, cfg)

	if app.Crypto != nil && app.Keys != nil {
		tools = append(tools, buildCryptoTools(app.Crypto, app.Keys, refs, scanner)...)
		logger().Info("crypto tools registered")
	}
	if app.Secrets != nil {
		tools = append(tools, buildSecretsTools(app.Secrets, refs, scanner)...)
		logger().Info("secrets tools registered")
	}

	// 5d. Graph Store (optional) — initialized before knowledge so GraphEngine can be wired.
	gc := initGraphStore(cfg)
	if gc != nil {
		app.GraphStore = gc.store
		app.GraphBuffer = gc.buffer
	}

	// 5. Skills (file-based, independent of knowledge)
	registry := initSkills(cfg, tools)
	if registry != nil {
		app.SkillRegistry = registry
		tools = append(tools, registry.LoadedSkills()...)
	}

	// 5a. Knowledge system (optional, non-blocking)
	kc := initKnowledge(cfg, store, gc)
	if kc != nil {
		app.KnowledgeStore = kc.store
		app.LearningEngine = kc.engine

		// Wrap base tools with learning observer (Engine or GraphEngine)
		wrapped := make([]*agent.Tool, len(tools))
		for i, t := range tools {
			wrapped[i] = wrapWithLearning(t, kc.observer)
		}
		tools = wrapped

		// Add meta-tools
		metaTools := buildMetaTools(kc.store, kc.engine, registry)
		tools = append(tools, metaTools...)
	}

	// 5b. Observational Memory (optional)
	mc := initMemory(cfg, store, sv)
	if mc != nil {
		app.MemoryStore = mc.store
		app.MemoryBuffer = mc.buffer
	}

	// 5c. Embedding / RAG (optional)
	ec := initEmbedding(cfg, boot.RawDB, kc, mc)
	if ec != nil {
		app.EmbeddingBuffer = ec.buffer
		app.RAGService = ec.ragService
	}

	// 5d'. Wire graph callbacks into knowledge and memory stores.
	if gc != nil {
		wireGraphCallbacks(gc, kc, mc, sv, cfg)
		// Initialize Graph RAG hybrid retrieval.
		initGraphRAG(cfg, gc, ec)
	}

	// 5d''. Conversation Analysis (optional)
	ab := initConversationAnalysis(cfg, sv, store, kc, gc)
	if ab != nil {
		app.AnalysisBuffer = ab
	}

	// 5d'''. Proactive Librarian (optional)
	lc := initLibrarian(cfg, sv, store, kc, mc, gc)
	if lc != nil {
		app.LibrarianInquiryStore = lc.inquiryStore
		app.LibrarianProactiveBuffer = lc.proactiveBuffer
	}

	// 5e. Graph tools (optional)
	if gc != nil {
		tools = append(tools, buildGraphTools(gc.store)...)
	}

	// 5f. RAG tools (optional)
	if ec != nil && ec.ragService != nil {
		tools = append(tools, buildRAGTools(ec.ragService)...)
	}

	// 5g. Memory agent tools (optional)
	if mc != nil {
		tools = append(tools, buildMemoryAgentTools(mc.store)...)
	}

	// 5h. Payment tools (optional)
	pc := initPayment(cfg, store, app.Secrets)
	if pc != nil {
		app.WalletProvider = pc.wallet
		app.PaymentService = pc.service
		tools = append(tools, buildPaymentTools(pc.service, pc.limiter)...)
	}

	// 5i. Librarian tools (optional)
	if lc != nil {
		tools = append(tools, buildLibrarianTools(lc.inquiryStore)...)
	}

	// 5j. Cron Scheduling (optional) — initialized before agent so tools get approval-wrapped.
	app.CronScheduler = initCron(cfg, store, app)
	if app.CronScheduler != nil {
		tools = append(tools, buildCronTools(app.CronScheduler)...)
		logger().Info("cron tools registered")
	}

	// 5k. Background Tasks (optional)
	app.BackgroundManager = initBackground(cfg, app)
	if app.BackgroundManager != nil {
		tools = append(tools, buildBackgroundTools(app.BackgroundManager)...)
		logger().Info("background tools registered")
	}

	// 5l. Workflow Engine (optional)
	app.WorkflowEngine = initWorkflow(cfg, store, app)
	if app.WorkflowEngine != nil {
		tools = append(tools, buildWorkflowTools(app.WorkflowEngine, cfg.Workflow.StateDir)...)
		logger().Info("workflow tools registered")
	}

	// 6. Auth
	auth := initAuth(cfg, store)

	// 7. Gateway (created before agent so we can wire approval)
	app.Gateway = initGateway(cfg, nil, app.Store, auth)

	// 8. Build composite approval provider and tool approval wrapper
	composite := approval.NewCompositeProvider()
	composite.Register(approval.NewGatewayProvider(app.Gateway))
	if cfg.Security.Interceptor.HeadlessAutoApprove {
		composite.SetTTYFallback(&approval.HeadlessProvider{})
		logger().Warn("headless auto-approve enabled — all tool executions will be auto-approved")
	} else {
		composite.SetTTYFallback(&approval.TTYProvider{})
	}
	app.ApprovalProvider = composite

	policy := cfg.Security.Interceptor.ApprovalPolicy
	if policy == "" {
		policy = config.ApprovalPolicyDangerous
	}
	if policy != config.ApprovalPolicyNone {
		for i, t := range tools {
			tools[i] = wrapWithApproval(t, cfg.Security.Interceptor, composite)
		}
		logger().Infow("tool approval enabled", "policy", string(policy))
	}

	// 9. ADK Agent (scanner is passed for output-side secret scanning)
	adkAgent, err := initAgent(context.Background(), sv, cfg, store, tools, kc, mc, ec, gc, scanner, registry, lc)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}
	app.Agent = adkAgent

	// Update gateway with the created agent
	app.Gateway.SetAgent(adkAgent)

	// 9b. A2A Server (if multi-agent and A2A enabled)
	if cfg.A2A.Enabled && cfg.Agent.MultiAgent && adkAgent.ADKAgent() != nil {
		a2aServer := a2a.NewServer(cfg.A2A, adkAgent.ADKAgent(), logger())
		a2aServer.RegisterRoutes(app.Gateway.Router())
	}

	// 10. Channels
	if err := app.initChannels(); err != nil {
		logger().Errorw("failed to initialize channels", "error", err)
	}

	// 11. Wire memory compaction (optional)
	if mc != nil && mc.buffer != nil {
		if entStore, ok := store.(*session.EntStore); ok {
			mc.buffer.SetCompactor(entStore.CompactMessages)
			logger().Info("observational memory compaction wired")
		}
	}

	// 15. Wire gateway turn callbacks for buffer triggers
	if app.MemoryBuffer != nil {
		app.Gateway.OnTurnComplete(func(sessionKey string) {
			app.MemoryBuffer.Trigger(sessionKey)
		})
	}
	if app.AnalysisBuffer != nil {
		app.Gateway.OnTurnComplete(func(sessionKey string) {
			app.AnalysisBuffer.Trigger(sessionKey)
		})
	}
	if app.LibrarianProactiveBuffer != nil {
		app.Gateway.OnTurnComplete(func(sessionKey string) {
			app.LibrarianProactiveBuffer.Trigger(sessionKey)
		})
	}

	return app, nil
}

// Start starts the application services
func (a *App) Start(ctx context.Context) error {
	logger().Info("starting application")

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.Gateway.Start(); err != nil {
			logger().Errorw("gateway server error", "error", err)
		}
	}()

	// Start observational memory buffer if enabled
	if a.MemoryBuffer != nil {
		a.MemoryBuffer.Start(&a.wg)
		logger().Info("observational memory buffer started")
	}

	// Start embedding buffer if enabled
	if a.EmbeddingBuffer != nil {
		a.EmbeddingBuffer.Start(&a.wg)
		logger().Info("embedding buffer started")
	}

	// Start graph buffer if enabled
	if a.GraphBuffer != nil {
		a.GraphBuffer.Start(&a.wg)
		logger().Info("graph buffer started")
	}

	// Start analysis buffer if enabled
	if a.AnalysisBuffer != nil {
		a.AnalysisBuffer.Start(&a.wg)
		logger().Info("conversation analysis buffer started")
	}

	// Start proactive librarian buffer if enabled
	if a.LibrarianProactiveBuffer != nil {
		a.LibrarianProactiveBuffer.Start(&a.wg)
		logger().Info("proactive librarian buffer started")
	}

	// Start cron scheduler if enabled
	if a.CronScheduler != nil {
		if err := a.CronScheduler.Start(ctx); err != nil {
			logger().Errorw("cron scheduler start error", "error", err)
		} else {
			logger().Info("cron scheduler started")
		}
	}

	logger().Info("starting channels...")
	for _, ch := range a.Channels {
		a.wg.Add(1)
		go func(c Channel) {
			defer a.wg.Done()
			if err := c.Start(ctx); err != nil {
				logger().Errorw("channel start error", "error", err)
			}
		}(ch)
	}

	return nil
}

// Stop stops the application services and waits for all goroutines to exit.
func (a *App) Stop(ctx context.Context) error {
	logger().Info("stopping application")

	// Stop cron scheduler
	if a.CronScheduler != nil {
		a.CronScheduler.Stop()
		logger().Info("cron scheduler stopped")
	}

	// Stop background manager
	if a.BackgroundManager != nil {
		a.BackgroundManager.Shutdown()
		logger().Info("background manager stopped")
	}

	// Stop workflow engine
	if a.WorkflowEngine != nil {
		a.WorkflowEngine.Shutdown()
		logger().Info("workflow engine stopped")
	}

	// Signal gateway and channels to stop
	if err := a.Gateway.Shutdown(ctx); err != nil {
		logger().Errorw("gateway shutdown error", "error", err)
	}

	for _, ch := range a.Channels {
		ch.Stop()
	}

	// Stop observational memory buffer
	if a.MemoryBuffer != nil {
		a.MemoryBuffer.Stop()
		logger().Info("observational memory buffer stopped")
	}

	// Stop embedding buffer
	if a.EmbeddingBuffer != nil {
		a.EmbeddingBuffer.Stop()
		logger().Info("embedding buffer stopped")
	}

	// Stop analysis buffer
	if a.AnalysisBuffer != nil {
		a.AnalysisBuffer.Stop()
		logger().Info("conversation analysis buffer stopped")
	}

	// Stop proactive librarian buffer
	if a.LibrarianProactiveBuffer != nil {
		a.LibrarianProactiveBuffer.Stop()
		logger().Info("proactive librarian buffer stopped")
	}

	// Stop graph buffer
	if a.GraphBuffer != nil {
		a.GraphBuffer.Stop()
		logger().Info("graph buffer stopped")
	}

	// Wait for all background goroutines to finish
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger().Info("all services stopped")
	case <-ctx.Done():
		logger().Warnw("shutdown timed out waiting for services", "error", ctx.Err())
	}

	if a.Browser != nil {
		if err := a.Browser.Close(); err != nil {
			logger().Errorw("browser close error", "error", err)
		}
	}

	if a.Store != nil {
		if err := a.Store.Close(); err != nil {
			logger().Errorw("session store close error", "error", err)
		}
	}

	if a.GraphStore != nil {
		if err := a.GraphStore.Close(); err != nil {
			logger().Errorw("graph store close error", "error", err)
		}
	}

	return nil
}

// registerConfigSecrets extracts sensitive values from config and registers
// them with the secret scanner so they are redacted from model output.
func registerConfigSecrets(scanner *agent.SecretScanner, cfg *config.Config) {
	register := func(name, value string) {
		if value != "" {
			scanner.Register(name, []byte(value))
		}
	}

	// Provider credentials
	for id, p := range cfg.Providers {
		register("provider."+id+".apiKey", p.APIKey)
	}

	// Channel tokens
	register("telegram.botToken", cfg.Channels.Telegram.BotToken)
	register("discord.botToken", cfg.Channels.Discord.BotToken)
	register("slack.botToken", cfg.Channels.Slack.BotToken)
	register("slack.appToken", cfg.Channels.Slack.AppToken)
	register("slack.signingSecret", cfg.Channels.Slack.SigningSecret)

	// Auth provider secrets
	for id, a := range cfg.Auth.Providers {
		register("auth."+id+".clientSecret", a.ClientSecret)
	}
}
