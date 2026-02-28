package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/a2a"
	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/lifecycle"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/sandbox"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/langoai/lango/internal/wallet"
	"github.com/langoai/lango/internal/tools/browser"
	"github.com/langoai/lango/internal/tools/filesystem"
	x402pkg "github.com/langoai/lango/internal/x402"
)

func logger() *zap.SugaredLogger { return logging.App() }

// New creates a new application instance from a bootstrap result.
func New(boot *bootstrap.Result) (*App, error) {
	cfg := boot.Config
	app := &App{
		Config:   cfg,
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
		tools = toolchain.ChainAll(tools, toolchain.WithLearning(kc.observer))

		// Add meta-tools
		metaTools := buildMetaTools(kc.store, kc.engine, registry, cfg.Skill)
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

		tools = append(tools, buildPaymentTools(pc, x402Interceptor)...)

		// 5h''. P2P networking (optional, requires wallet)
		p2pc = initP2P(cfg, pc.wallet, pc, boot.DBClient, app.Secrets)
		if p2pc != nil {
			app.P2PNode = p2pc.node
			// Wire P2P payment tool.
			tools = append(tools, buildP2PTools(p2pc)...)
			tools = append(tools, buildP2PPaymentTool(p2pc, pc)...)
		}
	}

	// 5i. Librarian tools (optional)
	if lc != nil {
		tools = append(tools, buildLibrarianTools(lc.inquiryStore)...)
	}

	// 5j. Cron Scheduling (optional) — initialized before agent so tools get approval-wrapped.
	app.CronScheduler = initCron(cfg, store, app)
	if app.CronScheduler != nil {
		tools = append(tools, buildCronTools(app.CronScheduler, cfg.Cron.DefaultDeliverTo)...)
		logger().Info("cron tools registered")
	}

	// 5k. Background Tasks (optional)
	app.BackgroundManager = initBackground(cfg, app)
	if app.BackgroundManager != nil {
		tools = append(tools, buildBackgroundTools(app.BackgroundManager, cfg.Background.DefaultDeliverTo)...)
		logger().Info("background tools registered")
	}

	// 5l. Workflow Engine (optional)
	app.WorkflowEngine = initWorkflow(cfg, store, app)
	if app.WorkflowEngine != nil {
		tools = append(tools, buildWorkflowTools(app.WorkflowEngine, cfg.Workflow.StateDir, cfg.Workflow.DefaultDeliverTo)...)
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
	adkAgent, err := initAgent(context.Background(), sv, cfg, store, tools, kc, mc, ec, gc, scanner, registry, lc)
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

	// 16. Register lifecycle components for ordered startup/shutdown.
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
	a.registry.StopAll(ctx)

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
}
