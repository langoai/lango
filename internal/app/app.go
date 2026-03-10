package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/a2a"
	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/agentmemory"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/lifecycle"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/observability/audit"
	"github.com/langoai/lango/internal/sandbox"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/p2p/gitbundle"
	"github.com/langoai/lango/internal/p2p/workspace"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/toolcatalog"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/langoai/lango/internal/tools/browser"
	"github.com/langoai/lango/internal/tools/filesystem"
	"github.com/langoai/lango/internal/wallet"
	x402pkg "github.com/langoai/lango/internal/x402"
)

func logger() *zap.SugaredLogger { return logging.App() }

// New creates a new application instance from a bootstrap result.
func New(boot *bootstrap.Result) (*App, error) {
	cfg := boot.Config
	bus := eventbus.New()
	app := &App{
		Config:   cfg,
		EventBus: bus,
		registry: lifecycle.NewRegistry(),
	}

	// 1. Supervisor (holds provider secrets, exec tool)
	sv, err := initSupervisor(cfg)
	if err != nil {
		return nil, fmt.Errorf("create supervisor: %w", err)
	}

	// 2. Session Store — reuse the DB client opened during bootstrap.
	store, err := initSessionStore(cfg, boot)
	if err != nil {
		return nil, fmt.Errorf("create session store: %w", err)
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

	automationAvailable := map[string]bool{
		"cron":       cfg.Cron.Enabled,
		"background": cfg.Background.Enabled,
		"workflow":   cfg.Workflow.Enabled,
	}
	tools := buildTools(sv, fsConfig, browserSM, automationAvailable)

	// Tool Catalog — register every built-in tool for dynamic discovery/dispatch.
	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{Name: "exec", Description: "Shell command execution", Enabled: true})
	catalog.RegisterCategory(toolcatalog.Category{Name: "filesystem", Description: "File system operations", Enabled: true})
	if cfg.Tools.Browser.Enabled {
		catalog.RegisterCategory(toolcatalog.Category{Name: "browser", Description: "Web browsing", ConfigKey: "tools.browser.enabled", Enabled: true})
	}
	// Register base tools (exec, fs, browser) all at once.
	for _, t := range tools {
		switch {
		case strings.HasPrefix(t.Name, "exec"):
			catalog.Register("exec", []*agent.Tool{t})
		case strings.HasPrefix(t.Name, "fs_"):
			catalog.Register("filesystem", []*agent.Tool{t})
		case strings.HasPrefix(t.Name, "browser_"):
			catalog.Register("browser", []*agent.Tool{t})
		}
	}

	// 4b. Crypto/Secrets tools (if security is enabled)
	// RefStore holds opaque references; plaintext never reaches agent context.
	// SecretScanner detects leaked secrets in model output.
	refs := security.NewRefStore()
	scanner := agent.NewSecretScanner()

	// Register config secrets to prevent leakage in model output.
	registerConfigSecrets(scanner, cfg)

	if app.Crypto != nil && app.Keys != nil {
		ct := buildCryptoTools(app.Crypto, app.Keys, refs, scanner)
		tools = append(tools, ct...)
		catalog.RegisterCategory(toolcatalog.Category{Name: "crypto", Description: "Cryptographic operations", ConfigKey: "security.signer.provider", Enabled: true})
		catalog.Register("crypto", ct)
		logger().Info("crypto tools registered")
	}
	if app.Secrets != nil {
		st := buildSecretsTools(app.Secrets, refs, scanner)
		tools = append(tools, st...)
		catalog.RegisterCategory(toolcatalog.Category{Name: "secrets", Description: "Secret management", ConfigKey: "security.secrets.enabled", Enabled: true})
		catalog.Register("secrets", st)
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
		tools = toolchain.ChainAll(tools, toolchain.WithLearning(kc.observer))

		// Add meta-tools
		metaTools := buildMetaTools(kc.store, kc.engine, registry, cfg.Skill)
		tools = append(tools, metaTools...)
		catalog.RegisterCategory(toolcatalog.Category{Name: "meta", Description: "Knowledge, learning, and skill management", ConfigKey: "knowledge.enabled", Enabled: true})
		catalog.Register("meta", metaTools)
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
		gt := buildGraphTools(gc.store)
		tools = append(tools, gt...)
		catalog.RegisterCategory(toolcatalog.Category{Name: "graph", Description: "Knowledge graph traversal", ConfigKey: "graph.enabled", Enabled: true})
		catalog.Register("graph", gt)
	}

	// 5f. RAG tools (optional)
	if ec != nil && ec.ragService != nil {
		rt := buildRAGTools(ec.ragService)
		tools = append(tools, rt...)
		catalog.RegisterCategory(toolcatalog.Category{Name: "rag", Description: "Retrieval-augmented generation", ConfigKey: "embedding.rag.enabled", Enabled: true})
		catalog.Register("rag", rt)
	}

	// 5g. Memory agent tools (optional)
	if mc != nil {
		mt := buildMemoryAgentTools(mc.store)
		tools = append(tools, mt...)
		catalog.RegisterCategory(toolcatalog.Category{Name: "memory", Description: "Observational memory", ConfigKey: "observationalMemory.enabled", Enabled: true})
		catalog.Register("memory", mt)
	}

	// 5g'. Agent Memory tools (optional, per-agent persistent memory)
	if cfg.AgentMemory.Enabled {
		amStore := agentmemory.NewInMemoryStore()
		app.AgentMemoryStore = amStore
		amTools := buildAgentMemoryTools(amStore)
		tools = append(tools, amTools...)
		catalog.RegisterCategory(toolcatalog.Category{Name: "agent_memory", Description: "Per-agent persistent memory", ConfigKey: "agentMemory.enabled", Enabled: true})
		catalog.Register("agent_memory", amTools)
		logger().Info("agent memory tools enabled")
	}

	// 5h. Payment tools (optional)
	pc := initPayment(cfg, store, app.Secrets)
	var p2pc *p2pComponents
	var x402Interceptor *x402pkg.Interceptor
	if pc != nil {
		app.WalletProvider = pc.wallet
		app.PaymentService = pc.service

		// 5h'. X402 interceptor (optional, requires payment)
		xc := initX402(cfg, app.Secrets, pc.limiter)
		if xc != nil {
			x402Interceptor = xc.interceptor
			app.X402Interceptor = xc.interceptor
		}

		pt := buildPaymentTools(pc, x402Interceptor)
		tools = append(tools, pt...)
		catalog.RegisterCategory(toolcatalog.Category{Name: "payment", Description: "Blockchain payments (USDC on Base)", ConfigKey: "payment.enabled", Enabled: true})
		catalog.Register("payment", pt)

		// 5h''. P2P networking (optional, requires wallet)
		// Use the single global bus so settlement and other P2P subscribers
		// receive tool execution events published by EventBusHook.
		p2pc = initP2P(cfg, pc.wallet, pc, boot.DBClient, app.Secrets, bus)
		if p2pc != nil {
			app.P2PNode = p2pc.node
			app.P2PAgentPool = p2pc.agentPool
			app.P2PTeamCoordinator = p2pc.coordinator
			app.P2PAgentProvider = p2pc.provider

			// Register NonceCache lifecycle so it is stopped on shutdown.
			if p2pc.nonceCache != nil {
				nc := p2pc.nonceCache
				app.registry.Register(lifecycle.NewFuncComponent("p2p-nonce-cache",
					func(_ context.Context, _ *sync.WaitGroup) error { return nil },
					func(_ context.Context) error {
						nc.Stop()
						return nil
					},
				), lifecycle.PriorityNetwork)
			}

			// Wire P2P payment tool.
			p2pTools := buildP2PTools(p2pc)
			p2pTools = append(p2pTools, buildP2PPaymentTool(p2pc, pc)...)
			p2pTools = append(p2pTools, buildP2PPaidInvokeTool(p2pc, pc)...)
			tools = append(tools, p2pTools...)
			catalog.RegisterCategory(toolcatalog.Category{Name: "p2p", Description: "Peer-to-peer networking", ConfigKey: "p2p.enabled", Enabled: true})
			catalog.Register("p2p", p2pTools)

			// 5h'''. P2P Workspace + Git (optional, requires P2P node)
			var sessionValidator gitbundle.SessionValidator
			if p2pc.sessions != nil {
				sess := p2pc.sessions
				sessionValidator = func(token string) (string, bool) {
					for _, s := range sess.ActiveSessions() {
						if s.Token == token {
							return s.PeerDID, true
						}
					}
					return "", false
				}
			}

			var localDID string
			if p2pc.identity != nil {
				d, idErr := p2pc.identity.DID(context.Background())
				if idErr == nil && d != nil {
					localDID = d.ID
				}
			}

			wsc := initWorkspace(cfg, p2pc.node, localDID, sessionValidator)
			if wsc != nil {
				// Wire chronicler triple adder to graph store if available.
				if wsc.chronicler != nil && app.GraphStore != nil {
					gs := app.GraphStore
					wsc.chronicler = workspace.NewChronicler(func(ctx context.Context, triples []workspace.Triple) error {
						// Convert workspace triples to graph triples.
						type graphTriple struct {
							Subject   string
							Predicate string
							Object    string
						}
						gTriples := make([]graphTriple, len(triples))
						for i, t := range triples {
							gTriples[i] = graphTriple{Subject: t.Subject, Predicate: t.Predicate, Object: t.Object}
						}
						_ = gs // graph store wiring deferred to avoid import cycle
						return nil
					}, logger())
				}

				// Build and register workspace tools.
				wsTools := buildWorkspaceTools(&workspaceComponents{
					manager:    wsc.manager,
					gitService: wsc.gitService,
					gossip:     wsc.gossip,
					tracker:    wsc.tracker,
				})
				tools = append(tools, wsTools...)
				catalog.RegisterCategory(toolcatalog.Category{
					Name:        "workspace",
					Description: "P2P collaborative workspaces and git sharing",
					ConfigKey:   "p2p.workspace.enabled",
					Enabled:     true,
				})
				catalog.Register("workspace", wsTools)

				// Register workspace DB lifecycle for graceful shutdown.
				wsDB := wsc.db
				app.registry.Register(lifecycle.NewFuncComponent("p2p-workspace-db",
					func(_ context.Context, _ *sync.WaitGroup) error { return nil },
					func(_ context.Context) error {
						if wsDB != nil {
							return wsDB.Close()
						}
						return nil
					},
				), lifecycle.PriorityNetwork)

				// Register workspace gossip lifecycle.
				if wsc.gossip != nil {
					wsGossip := wsc.gossip
					app.registry.Register(lifecycle.NewFuncComponent("p2p-workspace-gossip",
						func(_ context.Context, _ *sync.WaitGroup) error { return nil },
						func(_ context.Context) error {
							wsGossip.Stop()
							return nil
						},
					), lifecycle.PriorityNetwork)
				}

				logger().Info("P2P workspace tools registered")
			}
		}
	}

	// 5i. Librarian tools (optional)
	if lc != nil {
		lt := buildLibrarianTools(lc.inquiryStore)
		tools = append(tools, lt...)
		catalog.RegisterCategory(toolcatalog.Category{Name: "librarian", Description: "Knowledge inquiries and gap detection", ConfigKey: "librarian.enabled", Enabled: true})
		catalog.Register("librarian", lt)
	}

	// 5j. Cron Scheduling (optional) — initialized before agent so tools get approval-wrapped.
	app.CronScheduler = initCron(cfg, store, app)
	if app.CronScheduler != nil {
		cronTools := buildCronTools(app.CronScheduler, cfg.Cron.DefaultDeliverTo)
		tools = append(tools, cronTools...)
		catalog.RegisterCategory(toolcatalog.Category{Name: "cron", Description: "Cron job scheduling", ConfigKey: "cron.enabled", Enabled: true})
		catalog.Register("cron", cronTools)
		logger().Info("cron tools registered")
	}

	// 5k. Background Tasks (optional)
	app.BackgroundManager = initBackground(cfg, app)
	if app.BackgroundManager != nil {
		bgTools := buildBackgroundTools(app.BackgroundManager, cfg.Background.DefaultDeliverTo)
		tools = append(tools, bgTools...)
		catalog.RegisterCategory(toolcatalog.Category{Name: "background", Description: "Background task execution", ConfigKey: "background.enabled", Enabled: true})
		catalog.Register("background", bgTools)
		logger().Info("background tools registered")
	}

	// 5l. Workflow Engine (optional)
	app.WorkflowEngine = initWorkflow(cfg, store, app)
	if app.WorkflowEngine != nil {
		wfTools := buildWorkflowTools(app.WorkflowEngine, cfg.Workflow.StateDir, cfg.Workflow.DefaultDeliverTo)
		tools = append(tools, wfTools...)
		catalog.RegisterCategory(toolcatalog.Category{Name: "workflow", Description: "Workflow pipeline execution", ConfigKey: "workflow.enabled", Enabled: true})
		catalog.Register("workflow", wfTools)
		logger().Info("workflow tools registered")
	}

	// 5m. Dispatcher tools — dynamic access to all registered built-in tools.
	dispatcherTools := toolcatalog.BuildDispatcher(catalog)
	tools = append(tools, dispatcherTools...)
	app.ToolCatalog = catalog

	// 5n. MCP Plugins (optional — external MCP server tools)
	mcpc := initMCP(cfg)
	if mcpc != nil {
		app.MCPManager = mcpc.manager
		tools = append(tools, mcpc.tools...)
		catalog.RegisterCategory(toolcatalog.Category{
			Name:        "mcp",
			Description: "MCP plugin tools (external servers)",
			ConfigKey:   "mcp.enabled",
			Enabled:     true,
		})
		catalog.Register("mcp", mcpc.tools)
		// Register management meta-tools
		mgmtTools := buildMCPManagementTools(mcpc.manager)
		tools = append(tools, mgmtTools...)
		catalog.Register("mcp", mgmtTools)
	}

	// 5o. Economy Layer (optional — budget, risk, pricing, negotiation, escrow)
	econc := initEconomy(cfg, p2pc, pc, bus)
	if econc != nil {
		app.EconomyBudget = econc.budgetEngine
		app.EconomyRisk = econc.riskEngine
		app.EconomyPricing = econc.pricingEngine
		app.EconomyNegotiation = econc.negotiationEngine
		app.EconomyEscrow = econc.escrowEngine

		econTools := buildEconomyTools(econc)
		tools = append(tools, econTools...)
		catalog.RegisterCategory(toolcatalog.Category{
			Name:        "economy",
			Description: "P2P economy (budget, risk, pricing, negotiation, escrow)",
			ConfigKey:   "economy.enabled",
			Enabled:     true,
		})
		catalog.Register("economy", econTools)
		logger().Info("economy tools registered")

		// 5o'. On-chain escrow tools (if escrow engine is available)
		if econc.escrowEngine != nil && econc.escrowSettler != nil {
			escrowTools := buildOnChainEscrowTools(econc.escrowEngine, econc.escrowSettler)
			tools = append(tools, escrowTools...)
			catalog.RegisterCategory(toolcatalog.Category{
				Name:        "escrow",
				Description: "On-chain escrow management (hub/vault/custodian)",
				ConfigKey:   "economy.escrow.enabled",
				Enabled:     true,
			})
			catalog.Register("escrow", escrowTools)
			logger().Info("on-chain escrow tools registered")
		}

		// 5o''. Sentinel tools (if sentinel engine is available)
		if econc.sentinelEngine != nil {
			sentTools := buildSentinelTools(econc.sentinelEngine)
			tools = append(tools, sentTools...)
			catalog.RegisterCategory(toolcatalog.Category{
				Name:        "sentinel",
				Description: "Security Sentinel anomaly detection",
				ConfigKey:   "economy.escrow.enabled",
				Enabled:     true,
			})
			catalog.Register("sentinel", sentTools)
			logger().Info("sentinel tools registered")
		}
	}

	// 5p. Contract interaction (optional, requires payment)
	cc := initContract(pc)
	if cc != nil {
		ctTools := buildContractTools(cc.caller)
		tools = append(tools, ctTools...)
		catalog.RegisterCategory(toolcatalog.Category{
			Name:        "contract",
			Description: "Smart contract interaction",
			ConfigKey:   "payment.enabled",
			Enabled:     true,
		})
		catalog.Register("contract", ctTools)
		logger().Info("contract interaction tools registered")
	}

	// 5p'. Smart Account (optional, requires payment + contract)
	sacc := initSmartAccount(cfg, pc, econc, bus)
	if sacc != nil {
		app.SmartAccountManager = sacc.manager
		app.SmartAccountComponents = sacc
		saTools := buildSmartAccountTools(sacc)
		tools = append(tools, saTools...)
		catalog.RegisterCategory(toolcatalog.Category{
			Name:        "smartaccount",
			Description: "ERC-7579 smart account management",
			ConfigKey:   "smartAccount.enabled",
			Enabled:     true,
		})
		catalog.Register("smartaccount", saTools)
		logger().Info("smart account tools registered")
	}

	// 5q. Observability (optional — metrics, health, token tracking)
	obsc := initObservability(cfg, boot.DBClient, bus)
	if obsc != nil {
		app.MetricsCollector = obsc.collector
		app.HealthRegistry = obsc.healthRegistry
		app.TokenStore = obsc.tokenStore
	}

	// 6. Auth
	auth := initAuth(cfg, store)

	// 7. Gateway (created before agent so we can wire approval)
	app.Gateway = initGateway(cfg, nil, app.Store, auth)

	// 7b. Tool Execution Hooks
	if cfg.Hooks.Enabled || cfg.Agent.MultiAgent {
		hookRegistry := toolchain.NewHookRegistry()

		// Register built-in hooks based on configuration.
		if cfg.Hooks.SecurityFilter {
			hookRegistry.RegisterPre(toolchain.NewSecurityFilterHook(cfg.Hooks.BlockedCommands))
		}
		if cfg.Hooks.AccessControl {
			hookRegistry.RegisterPre(toolchain.NewAgentAccessControlHook(nil))
		}
		if cfg.Hooks.EventPublishing && bus != nil {
			ebHook := toolchain.NewEventBusHook(bus)
			hookRegistry.RegisterPre(ebHook)
			hookRegistry.RegisterPost(ebHook)
		}

		tools = toolchain.ChainAll(tools, toolchain.WithHooks(hookRegistry))
		logger().Infow("tool hooks enabled",
			"preHooks", len(hookRegistry.PreHooks()),
			"postHooks", len(hookRegistry.PostHooks()),
		)
	}

	// 8. Build composite approval provider and tool approval wrapper
	composite := approval.NewCompositeProvider()
	composite.Register(approval.NewGatewayProvider(app.Gateway))
	if cfg.Security.Interceptor.HeadlessAutoApprove {
		composite.SetTTYFallback(&approval.HeadlessProvider{})
		logger().Warn("headless auto-approve enabled — all tool executions will be auto-approved")
	} else {
		composite.SetTTYFallback(&approval.TTYProvider{})
	}
	// P2P sessions use a dedicated fallback to prevent HeadlessProvider
	// from auto-approving remote peer requests.
	if cfg.P2P.Enabled {
		composite.SetP2PFallback(&approval.TTYProvider{})
		logger().Info("P2P approval routed to TTY (HeadlessProvider blocked for remote peers)")
	}
	app.ApprovalProvider = composite

	grantStore := approval.NewGrantStore()
	// P2P grants expire after 1 hour to limit the window of implicit trust.
	if cfg.P2P.Enabled {
		grantStore.SetTTL(time.Hour)
	}
	app.GrantStore = grantStore

	policy := cfg.Security.Interceptor.ApprovalPolicy
	if policy == "" {
		policy = config.ApprovalPolicyDangerous
	}
	if policy != config.ApprovalPolicyNone {
		var limiter wallet.SpendingLimiter
		if pc != nil {
			limiter = pc.limiter
		}
		tools = toolchain.ChainAll(tools,
			toolchain.WithApproval(cfg.Security.Interceptor, composite, grantStore, limiter))
		logger().Infow("tool approval enabled", "policy", string(policy))
	}

	// 9. ADK Agent (scanner is passed for output-side secret scanning)
	adkAgent, err := initAgent(context.Background(), &agentDeps{
		sv:       sv,
		cfg:      cfg,
		store:    store,
		tools:    tools,
		kc:       kc,
		mc:       mc,
		ec:       ec,
		gc:       gc,
		scanner:  scanner,
		sr:       registry,
		lc:       lc,
		catalog:  catalog,
		p2pc:     p2pc,
		eventBus: bus,
	})
	if err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}
	app.Agent = adkAgent

	// Update gateway with the created agent
	app.Gateway.SetAgent(adkAgent)

	// 9b. A2A Server (if multi-agent and A2A enabled)
	if cfg.A2A.Enabled && cfg.Agent.MultiAgent && adkAgent.ADKAgent() != nil {
		a2aServer := a2a.NewServer(cfg.A2A, adkAgent.ADKAgent(), logger())
		a2aServer.RegisterRoutes(app.Gateway.Router())
	}

	// 9c. P2P executor + REST API routes (if P2P enabled)
	if p2pc != nil {
		// Wire executor callback so remote peers can invoke local tools.
		// Capture the tools slice in a closure for direct tool dispatch.
		if p2pc.handler != nil {
			toolIndex := make(map[string]*agent.Tool, len(tools))
			for _, t := range tools {
				toolIndex[t.Name] = t
			}
			p2pc.handler.SetExecutor(func(ctx context.Context, toolName string, params map[string]interface{}) (map[string]interface{}, error) {
				t, ok := toolIndex[toolName]
				if !ok {
					return nil, fmt.Errorf("tool %q not found", toolName)
				}
				result, err := t.Handler(ctx, params)
				if err != nil {
					return nil, err
				}
				// Coerce the result to map[string]interface{}.
				switch v := result.(type) {
				case map[string]interface{}:
					return v, nil
				default:
					return map[string]interface{}{"result": v}, nil
				}
			})

			// Wire sandbox executor for P2P tool isolation if enabled.
			if cfg.P2P.ToolIsolation.Enabled {
				sbxCfg := sandbox.Config{
					Enabled:        true,
					TimeoutPerTool: cfg.P2P.ToolIsolation.TimeoutPerTool,
					MaxMemoryMB:    cfg.P2P.ToolIsolation.MaxMemoryMB,
				}
				var sbxExec sandbox.Executor
				if cfg.P2P.ToolIsolation.Container.Enabled {
					containerExec, err := sandbox.NewContainerExecutor(sbxCfg, cfg.P2P.ToolIsolation.Container)
					if err != nil {
						logger().Warnf("Container sandbox unavailable, falling back to subprocess: %v", err)
						sbxExec = sandbox.NewSubprocessExecutor(sbxCfg)
					} else {
						sbxExec = containerExec
						logger().Infof("P2P tool isolation enabled (container mode: %s)", containerExec.RuntimeName())
					}
				} else {
					sbxExec = sandbox.NewSubprocessExecutor(sbxCfg)
					logger().Info("P2P tool isolation enabled (subprocess mode)")
				}
				p2pc.handler.SetSandboxExecutor(func(ctx context.Context, toolName string, params map[string]interface{}) (map[string]interface{}, error) {
					return sbxExec.Execute(ctx, toolName, params)
				})
			}

			// Wire owner approval callback for inbound remote tool invocations.
			if pc != nil {
				p2pc.handler.SetApprovalFunc(func(ctx context.Context, peerDID, toolName string, params map[string]interface{}) (bool, error) {
					// Never auto-approve dangerous tools via P2P.
					// Unknown tools (not in index) are also treated as dangerous.
					t, known := toolIndex[toolName]
					if !known || t.SafetyLevel.IsDangerous() {
						goto requireApproval
					}

					// For non-dangerous paid tools, check if the amount is auto-approvable.
					if p2pc.pricingFn != nil {
						if priceStr, isFree := p2pc.pricingFn(toolName); !isFree {
							amt, err := wallet.ParseUSDC(priceStr)
							if err == nil {
								if autoOK, checkErr := pc.limiter.IsAutoApprovable(ctx, amt); checkErr == nil && autoOK {
									if grantStore != nil {
										grantStore.Grant("p2p:"+peerDID, toolName)
									}
									return true, nil
								}
							}
						}
					}

				requireApproval:
					// Fall back to composite approval provider.
					req := approval.ApprovalRequest{
						ID:         fmt.Sprintf("p2p-%d", time.Now().UnixNano()),
						ToolName:   toolName,
						SessionKey: "p2p:" + peerDID,
						Params:     params,
						Summary:    fmt.Sprintf("Remote peer %s wants to invoke tool '%s'", truncate(peerDID, 16), toolName),
						CreatedAt:  time.Now(),
					}
					resp, err := composite.RequestApproval(ctx, req)
					if err != nil {
						return false, nil // fail-closed
					}
					// Record grant to avoid double-approval (handler approvalFn + tool's wrapWithApproval).
					if resp.Approved && grantStore != nil {
						grantStore.Grant("p2p:"+peerDID, toolName)
					}
					return resp.Approved, nil
				})
			}
		}
		registerP2PRoutes(app.Gateway.Router(), p2pc)
		logger().Info("P2P REST API routes registered")
	}

	// 9d. Observability API routes
	if obsc != nil {
		registerObservabilityRoutes(app.Gateway.Router(), obsc.collector, obsc.healthRegistry, obsc.tokenStore)
		logger().Info("observability API routes registered")
	}

	// 9e. Audit recorder (optional)
	if cfg.Observability.Audit.Enabled && boot.DBClient != nil {
		auditRec := audit.NewRecorder(boot.DBClient)
		auditRec.Subscribe(bus)
		logger().Info("audit recorder wired to event bus")
	}

	// 10. Channels
	if err := app.initChannels(); err != nil {
		logger().Errorw("initialize channels", "error", err)
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

	// 16. Observability lifecycle (token store cleanup on shutdown).
	registerObservabilityLifecycle(app.registry, obsc, cfg)

	// 17. Register lifecycle components for ordered startup/shutdown.
	app.registerLifecycleComponents()

	return app, nil
}

// registerLifecycleComponents registers all startable/stoppable components
// with the lifecycle registry using appropriate adapters and priorities.
func (a *App) registerLifecycleComponents() {
	reg := a.registry

	// Gateway — runs blocking in a goroutine, shutdown via context.
	reg.Register(lifecycle.NewFuncComponent("gateway",
		func(_ context.Context, wg *sync.WaitGroup) error {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := a.Gateway.Start(); err != nil {
					logger().Errorw("gateway server error", "error", err)
				}
			}()
			return nil
		},
		func(ctx context.Context) error {
			return a.Gateway.Shutdown(ctx)
		},
	), lifecycle.PriorityNetwork)

	// Buffers — all implement Startable (Start(*sync.WaitGroup) / Stop()).
	if a.MemoryBuffer != nil {
		reg.Register(lifecycle.NewSimpleComponent("memory-buffer", a.MemoryBuffer), lifecycle.PriorityBuffer)
	}
	if a.EmbeddingBuffer != nil {
		reg.Register(lifecycle.NewSimpleComponent("embedding-buffer", a.EmbeddingBuffer), lifecycle.PriorityBuffer)
	}
	if a.GraphBuffer != nil {
		reg.Register(lifecycle.NewSimpleComponent("graph-buffer", a.GraphBuffer), lifecycle.PriorityBuffer)
	}
	if a.AnalysisBuffer != nil {
		reg.Register(lifecycle.NewSimpleComponent("analysis-buffer", a.AnalysisBuffer), lifecycle.PriorityBuffer)
	}
	if a.LibrarianProactiveBuffer != nil {
		reg.Register(lifecycle.NewSimpleComponent("librarian-proactive-buffer", a.LibrarianProactiveBuffer), lifecycle.PriorityBuffer)
	}

	// P2P Node — Start(*sync.WaitGroup) error / Stop() error.
	if a.P2PNode != nil {
		reg.Register(lifecycle.NewFuncComponent("p2p-node",
			func(_ context.Context, wg *sync.WaitGroup) error {
				return a.P2PNode.Start(wg)
			},
			func(_ context.Context) error {
				return a.P2PNode.Stop()
			},
		), lifecycle.PriorityNetwork)
	}

	// Cron Scheduler — Start(ctx) error / Stop().
	if a.CronScheduler != nil {
		reg.Register(lifecycle.NewFuncComponent("cron-scheduler",
			func(ctx context.Context, _ *sync.WaitGroup) error {
				return a.CronScheduler.Start(ctx)
			},
			func(_ context.Context) error {
				a.CronScheduler.Stop()
				return nil
			},
		), lifecycle.PriorityAutomation)
	}

	// Background Manager — no Start, only Shutdown().
	if a.BackgroundManager != nil {
		reg.Register(lifecycle.NewFuncComponent("background-manager",
			func(_ context.Context, _ *sync.WaitGroup) error { return nil },
			func(_ context.Context) error {
				a.BackgroundManager.Shutdown()
				return nil
			},
		), lifecycle.PriorityAutomation)
	}

	// Workflow Engine — no Start, only Shutdown().
	if a.WorkflowEngine != nil {
		reg.Register(lifecycle.NewFuncComponent("workflow-engine",
			func(_ context.Context, _ *sync.WaitGroup) error { return nil },
			func(_ context.Context) error {
				a.WorkflowEngine.Shutdown()
				return nil
			},
		), lifecycle.PriorityAutomation)
	}

	// MCP Manager — disconnect all servers on shutdown.
	if a.MCPManager != nil {
		reg.Register(lifecycle.NewFuncComponent("mcp-manager",
			func(_ context.Context, _ *sync.WaitGroup) error { return nil },
			func(ctx context.Context) error { return a.MCPManager.DisconnectAll(ctx) },
		), lifecycle.PriorityNetwork)
	}

	// Channels — each runs blocking in a goroutine, Stop() to signal.
	for i, ch := range a.Channels {
		ch := ch // capture for closure
		name := fmt.Sprintf("channel-%d", i)
		reg.Register(lifecycle.NewFuncComponent(name,
			func(ctx context.Context, wg *sync.WaitGroup) error {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := ch.Start(ctx); err != nil {
						logger().Errorw("channel start error", "error", err)
					}
				}()
				return nil
			},
			func(_ context.Context) error {
				ch.Stop()
				return nil
			},
		), lifecycle.PriorityNetwork)
	}
}

// Start starts the application services using the lifecycle registry.
func (a *App) Start(ctx context.Context) error {
	logger().Info("starting application")
	return a.registry.StartAll(ctx, &a.wg)
}

// Stop stops the application services and waits for all goroutines to exit.
func (a *App) Stop(ctx context.Context) error {
	logger().Info("stopping application")

	// Stop all lifecycle-managed components in reverse startup order.
	_ = a.registry.StopAll(ctx)

	// Wait for all background goroutines to finish.
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

	// Close non-lifecycle resources (browser, stores) after all components stop.
	if a.Browser != nil {
		if err := a.Browser.Close(); err != nil {
			logger().Warnw("browser close error", "error", err)
		}
	}

	if a.Store != nil {
		if err := a.Store.Close(); err != nil {
			logger().Warnw("session store close error", "error", err)
		}
	}

	if a.GraphStore != nil {
		if err := a.GraphStore.Close(); err != nil {
			logger().Warnw("graph store close error", "error", err)
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

	// MCP server secrets (headers and env vars)
	for name, srv := range cfg.MCP.Servers {
		for hk, hv := range srv.Headers {
			register("mcp."+name+".header."+hk, hv)
		}
		for ek, ev := range srv.Env {
			register("mcp."+name+".env."+ek, ev)
		}
	}
}
