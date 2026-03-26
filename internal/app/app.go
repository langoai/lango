package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/a2a"
	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	cronpkg "github.com/langoai/lango/internal/cron"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/gateway"
	"github.com/langoai/lango/internal/learning"
	"github.com/langoai/lango/internal/lifecycle"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/observability/audit"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/sandbox"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/skill"
	"github.com/langoai/lango/internal/toolcatalog"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/langoai/lango/internal/tooloutput"
	"github.com/langoai/lango/internal/turnrunner"
	"github.com/langoai/lango/internal/turntrace"
	"github.com/langoai/lango/internal/wallet"
	"github.com/langoai/lango/internal/workflow"
)

func logger() *zap.SugaredLogger { return logging.App() }

// cleanupEntry pairs a name with its rollback function.
type cleanupEntry struct {
	name string
	fn   func()
}

// cleanupStack accumulates rollback functions during Phase B wiring.
// On failure, rollback() executes all cleanups in reverse order.
// On success, clear() discards the stack (lifecycle registry takes ownership).
type cleanupStack struct {
	entries []cleanupEntry
}

// push adds a named cleanup function to the stack.
func (s *cleanupStack) push(name string, fn func()) {
	s.entries = append(s.entries, cleanupEntry{name: name, fn: fn})
}

// rollback executes all cleanups in reverse order (last-in, first-out).
func (s *cleanupStack) rollback() {
	for i := len(s.entries) - 1; i >= 0; i-- {
		e := s.entries[i]
		logger().Infow("rolling back Phase B step", "step", e.name)
		e.fn()
	}
	s.entries = nil
}

// clear discards the cleanup stack without executing any cleanups.
func (s *cleanupStack) clear() {
	s.entries = nil
}

// New creates a new application instance from a bootstrap result.
func New(boot *bootstrap.Result, opts ...AppOption) (*App, error) {
	var options appOptions
	for _, o := range opts {
		o(&options)
	}

	cfg := boot.Config
	bus := eventbus.New()
	ctx, cancel := context.WithCancel(context.Background())
	app := &App{
		Config:   cfg,
		EventBus: bus,
		registry: lifecycle.NewRegistry(),
		ctx:      ctx,
		cancel:   cancel,
	}

	// LocalChat mode: skip Network and Automation lifecycle components.
	if options.mode == AppModeLocalChat {
		app.registry.SetMaxPriority(lifecycle.PriorityBuffer)
	}

	// ── Phase A: Module Build ──

	builder := appinit.NewBuilder()
	builder.AddModule(&foundationModule{cfg: cfg, boot: boot})
	builder.AddModule(&intelligenceModule{cfg: cfg, boot: boot, rawDB: boot.RawDB, bus: bus})
	builder.AddModule(&automationModule{cfg: cfg, app: app})
	builder.AddModule(&networkModule{cfg: cfg, boot: boot, bus: bus, app: app})
	builder.AddModule(&extensionModule{cfg: cfg, boot: boot, bus: bus})
	builder.AddModule(&runLedgerModule{cfg: cfg, boot: boot})
	builder.AddModule(&provenanceModule{cfg: cfg, boot: boot})

	buildResult, err := builder.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("module build: %w", err)
	}

	resolver := buildResult.Resolver
	tools := buildResult.Tools

	// ── Phase B: Post-Build Wiring ──
	// Cleanup stack accumulates rollback functions during Phase B.
	// On failure, cleanups run in reverse order (bootstrap pipeline pattern).
	// On success, the stack is discarded — ownership transfers to the lifecycle registry.
	var cleanups cleanupStack

	logger().Info("Phase B: starting post-build wiring")

	// B1. Populate app fields from resolver.
	populateAppFields(app, resolver)

	// B1b. Provenance runtime capture + transport wiring.
	wireProvenanceRuntime(app, resolver)

	// B2. Build catalog from module CatalogEntries.
	catalog := buildCatalogFromEntries(buildResult.CatalogEntries)

	// B3. Dispatcher tools — dynamic access to all registered built-in tools.
	dispatcherTools := toolcatalog.BuildDispatcher(catalog)
	tools = append(tools, dispatcherTools...)
	app.ToolCatalog = catalog

	// B4. Cross-cutting middleware (order matters).
	// B4a. WithLearning — wrap all tools with learning observer.
	iv, _ := resolver.Resolve(appinit.ProvidesKnowledge).(*intelligenceValues)
	if iv != nil && iv.Observer != nil {
		if obs, ok := iv.Observer.(learning.ToolResultObserver); ok {
			tools = toolchain.ChainAll(tools, toolchain.WithLearning(obs))
		}
	}

	// B4b. Tool output management — token-based tiered compression.
	outputStore := tooloutput.NewOutputStore(10 * time.Minute)
	app.registry.Register(outputStore, lifecycle.PriorityCore)
	app.OutputStore = outputStore
	cleanups.push("output-store", func() {
		_ = outputStore.Stop(context.Background())
	})
	outputTools := tooloutput.BuildTools(outputStore)
	tools = append(tools, outputTools...)
	catalog.RegisterCategory(toolcatalog.Category{Name: "output", Description: "Tool output retrieval", Enabled: true})
	catalog.Register("output", outputTools)
	tools = toolchain.ChainAll(tools, toolchain.WithOutputManager(cfg.Tools.OutputManager, outputStore))

	// B4c. Tool Execution Hooks.
	hookRegistry := buildHookRegistry(cfg, bus)
	tools = toolchain.ChainAll(tools, toolchain.WithHooks(hookRegistry))
	app.HookRegistry = hookRegistry
	logger().Infow("tool hooks enabled",
		"preHooks", len(hookRegistry.PreHooks()),
		"postHooks", len(hookRegistry.PostHooks()),
	)

	// B5. Auth + Gateway.
	fv := resolver.Resolve(appinit.ProvidesSupervisor).(*foundationValues)
	auth := initAuth(cfg, fv.Store)
	app.Gateway = initGateway(cfg, nil, app.Store, auth)
	if app.Sanitizer != nil {
		app.Gateway.SetSanitizer(app.Sanitizer)
	}
	if app.RunLedgerStore != nil {
		app.Gateway.SetRunLedgerStore(app.RunLedgerStore)
	}
	cleanups.push("gateway", func() {
		_ = app.Gateway.Shutdown(context.Background())
	})

	// B4d. Build composite approval provider and tool approval wrapper.
	composite, grantStore := buildApprovalProvider(cfg, app.Gateway)
	app.ApprovalProvider = composite
	app.GrantStore = grantStore

	policy := cfg.Security.Interceptor.ApprovalPolicy
	if policy == "" {
		policy = config.ApprovalPolicyDangerous
	}
	if policy == config.ApprovalPolicyNone {
		logger().Warnw("tool approval policy is set to 'none' -- all tool calls will execute without user confirmation; not recommended for production")
	}
	if policy != config.ApprovalPolicyNone {
		var limiter wallet.SpendingLimiter
		nv, _ := resolver.Resolve(appinit.ProvidesPayment).(*paymentComponents)
		if nv != nil {
			limiter = nv.limiter
		}
		tools = toolchain.ChainAll(tools,
			toolchain.WithApproval(cfg.Security.Interceptor, composite, grantStore, limiter))
		logger().Infow("tool approval enabled", "policy", string(policy))
	}

	if app.RunLedgerStore != nil && cfg.RunLedger.WorkspaceIsolation {
		tools = toolchain.ChainAll(tools, runledger.ToolProfileGuard(app.RunLedgerStore))
	}

	// Log tool registration summary for diagnostics.
	logToolRegistrationSummary(catalog)

	// B6. Agent creation.
	scanner := fv.Scanner
	p2pc, _ := resolver.Resolve(appinit.ProvidesP2P).(*p2pComponents)
	adkAgent, err := initAgent(context.Background(), &agentDeps{
		sv:       fv.Supervisor,
		cfg:      cfg,
		store:    fv.Store,
		tools:    tools,
		kc:       resolveKC(iv),
		mc:       resolveMC(iv),
		ec:       resolveEC(iv),
		gc:       resolveGC(iv),
		scanner:  scanner,
		sr:       resolveSR(iv),
		lc:       resolveLC(iv),
		catalog:  catalog,
		p2pc:     p2pc,
		eventBus: bus,
		rls:      app.RunLedgerStore,
		prov: func() *provenanceValues {
			pv, _ := resolver.Resolve(appinit.ProvidesProvenance).(*provenanceValues)
			return pv
		}(),
	})
	if err != nil {
		cleanups.rollback()
		return nil, fmt.Errorf("create agent: %w", err)
	}
	app.Agent = adkAgent
	app.Gateway.SetAgent(adkAgent)
	app.TurnTraceStore = initTurnTraceStore(app.Store)
	idleTimeout, hardCeiling := app.resolveTimeouts()
	var errorFixProvider adk.ErrorFixProvider
	if iv != nil && iv.KC != nil && iv.KC.engine != nil {
		errorFixProvider = iv.KC.engine
	}
	executor := initAgentRuntime(cfg, adkAgent, bus, errorFixProvider)
	app.TurnRunner = turnrunner.New(turnrunner.Config{
		IdleTimeout:         idleTimeout,
		HardCeiling:         hardCeiling,
		TraceStore:          app.TurnTraceStore,
		DelegationBudgetMax: cfg.Agent.Orchestration.Budget.DelegationLimit,
	}, executor, app.Store, app.Sanitizer)
	app.Gateway.SetTurnRunner(app.TurnRunner)

	// B7. Post-agent wiring.
	wirePostAgent(app, resolver, tools, bus, composite, grantStore, boot, auth)

	// B8. Channels (skip in local-chat mode).
	if options.mode != AppModeLocalChat {
		if err := app.initChannels(); err != nil {
			logger().Errorw("initialize channels", "error", err)
		}
	}

	// B9. Memory compaction + turn callbacks.
	wireMemoryAndTurnCallbacks(app, iv, fv)

	// B10. Lifecycle registration (module components + gateway + channels).
	for _, entry := range buildResult.Components {
		app.registry.Register(entry.Component, entry.Priority)
	}
	if options.mode != AppModeLocalChat {
		registerPostBuildLifecycle(app)
	}

	// B11. Trace retention cleaner (if configured).
	if tsCfg := cfg.Observability.TraceStore; tsCfg.MaxAge > 0 || tsCfg.MaxTraces > 0 {
		cleaner := turntrace.NewRetentionCleaner(app.TurnTraceStore, turntrace.RetentionConfig{
			MaxAge:                tsCfg.MaxAge,
			MaxTraces:             tsCfg.MaxTraces,
			FailedTraceMultiplier: tsCfg.FailedTraceMultiplier,
			CleanupInterval:       tsCfg.CleanupInterval,
		})
		app.registry.Register(cleaner, lifecycle.PriorityCore)
	}

	// Phase B succeeded — discard rollback cleanups; lifecycle registry owns everything now.
	cleanups.clear()

	return app, nil
}

// populateAppFields maps resolver values to app struct fields.
func populateAppFields(app *App, r appinit.Resolver) {
	// Foundation.
	if fv, ok := r.Resolve(appinit.ProvidesSupervisor).(*foundationValues); ok {
		app.Store = fv.Store
		app.Crypto = fv.Crypto
		app.Keys = fv.Keys
		app.Secrets = fv.Secrets
		app.Sanitizer = fv.Sanitizer
		if fv.BrowserSM != nil {
			app.Browser = fv.BrowserSM
		}
	}

	// Intelligence.
	if iv, ok := r.Resolve(appinit.ProvidesKnowledge).(*intelligenceValues); ok {
		if iv.KC != nil {
			app.KnowledgeStore = iv.KC.store
			app.LearningEngine = iv.KC.engine
		}
		if iv.MC != nil {
			app.MemoryStore = iv.MC.store
			app.MemoryBuffer = iv.MC.buffer
		}
		if iv.EC != nil {
			app.EmbeddingBuffer = iv.EC.buffer
			app.RAGService = iv.EC.ragService
		}
		if iv.GC != nil {
			app.GraphStore = iv.GC.store
			app.GraphBuffer = iv.GC.buffer
		}
		if iv.LC != nil {
			app.LibrarianInquiryStore = iv.LC.inquiryStore
			app.LibrarianProactiveBuffer = iv.LC.proactiveBuffer
		}
		if iv.AB != nil {
			if ab, ok := iv.AB.(*learning.AnalysisBuffer); ok {
				app.AnalysisBuffer = ab
			}
		}
		if sr, ok := iv.SkillRegistry.(*skill.Registry); ok {
			app.SkillRegistry = sr
		}
		app.AgentMemoryStore = iv.AgentMemoryStore
		app.FeatureStatuses = iv.FeatureStatuses
		if app.Gateway != nil && iv.FeatureStatuses != nil {
			app.Gateway.SetFeatureStatuses(iv.FeatureStatuses.All())
		}
	}

	// Automation.
	if av, ok := r.Resolve(appinit.ProvidesAutomation).(*automationValues); ok {
		if cs, ok := av.CronScheduler.(*cronpkg.Scheduler); ok {
			app.CronScheduler = cs
		}
		if bm, ok := av.BackgroundManager.(*background.Manager); ok {
			app.BackgroundManager = bm
		}
		if we, ok := av.WorkflowEngine.(*workflow.Engine); ok {
			app.WorkflowEngine = we
		}
	}

	// Network.
	if pc, ok := r.Resolve(appinit.ProvidesPayment).(*paymentComponents); ok && pc != nil {
		app.WalletProvider = pc.wallet
		app.PaymentService = pc.service
	}
	if p2pc, ok := r.Resolve(appinit.ProvidesP2P).(*p2pComponents); ok && p2pc != nil {
		app.P2PNode = p2pc.node
		app.P2PAgentPool = p2pc.agentPool
		app.P2PTeamCoordinator = p2pc.coordinator
		app.P2PAgentProvider = p2pc.provider
	}
	if econc, ok := r.Resolve(appinit.ProvidesEconomy).(*economyComponents); ok && econc != nil {
		app.EconomyBudget = econc.budgetEngine
		app.EconomyRisk = econc.riskEngine
		app.EconomyPricing = econc.pricingEngine
		app.EconomyNegotiation = econc.negotiationEngine
		app.EconomyEscrow = econc.escrowEngine
	}
	if sacc, ok := r.Resolve(appinit.ProvidesSmartAccount).(*smartAccountComponents); ok && sacc != nil {
		app.SmartAccountManager = sacc.manager
		app.SmartAccountComponents = sacc
	}

	// Extension.
	if mcpc, ok := r.Resolve(appinit.ProvidesMCP).(*mcpComponents); ok && mcpc != nil {
		app.MCPManager = mcpc.manager
	}
	if obsc, ok := r.Resolve(appinit.ProvidesObservability).(*observabilityComponents); ok && obsc != nil {
		app.MetricsCollector = obsc.collector
		app.HealthRegistry = obsc.healthRegistry
		app.TokenStore = obsc.tokenStore
	}

	// RunLedger.
	if rlv, ok := r.Resolve(appinit.ProvidesRunLedger).(*runLedgerValues); ok && rlv != nil {
		app.RunLedgerStore = rlv.store
		app.RunLedgerPEV = rlv.pev
	}

	// Provenance.
	if pv, ok := r.Resolve(appinit.ProvidesProvenance).(*provenanceValues); ok && pv != nil {
		app.ProvenanceCheckpoints = pv.checkpointService
		app.ProvenanceSessionTree = pv.sessionTree
		app.ProvenanceAttribution = pv.attribution
		app.ProvenanceBundle = pv.bundle
	}
}

// buildCatalogFromEntries converts module CatalogEntries into a toolcatalog.Catalog.
func buildCatalogFromEntries(entries []appinit.CatalogEntry) *toolcatalog.Catalog {
	catalog := toolcatalog.New()
	for _, e := range entries {
		catalog.RegisterCategory(toolcatalog.Category{
			Name:        e.Category,
			Description: e.Description,
			ConfigKey:   e.ConfigKey,
			Enabled:     e.Enabled,
		})
		if len(e.Tools) > 0 {
			catalog.Register(e.Category, e.Tools)
		}
	}
	return catalog
}

// buildHookRegistry constructs the tool execution hook registry.
func buildHookRegistry(cfg *config.Config, bus *eventbus.Bus) *toolchain.HookRegistry {
	hookRegistry := toolchain.NewHookRegistry()
	hookRegistry.RegisterPre(toolchain.NewSecurityFilterHook(cfg.Hooks.BlockedCommands))
	if cfg.Hooks.AccessControl {
		hookRegistry.RegisterPre(toolchain.NewAgentAccessControlHook(nil))
	}
	if (cfg.Hooks.Enabled || cfg.Agent.MultiAgent) && cfg.Hooks.EventPublishing && bus != nil {
		ebHook := toolchain.NewEventBusHook(bus)
		hookRegistry.RegisterPre(ebHook)
		hookRegistry.RegisterPost(ebHook)
	}
	return hookRegistry
}

// buildApprovalProvider constructs the composite approval provider and grant store.
func buildApprovalProvider(cfg *config.Config, gw *gateway.Server) (*approval.CompositeProvider, *approval.GrantStore) {
	composite := approval.NewCompositeProvider()
	composite.Register(approval.NewGatewayProvider(gw))
	if cfg.Security.Interceptor.HeadlessAutoApprove {
		composite.SetTTYFallback(&approval.HeadlessProvider{})
		logger().Warn("headless auto-approve enabled — all tool executions will be auto-approved")
	} else {
		composite.SetTTYFallback(&approval.TTYProvider{})
	}
	if cfg.P2P.Enabled {
		composite.SetP2PFallback(&approval.TTYProvider{})
		logger().Info("P2P approval routed to TTY (HeadlessProvider blocked for remote peers)")
	}

	grantStore := approval.NewGrantStore()
	if cfg.P2P.Enabled {
		grantStore.SetTTL(time.Hour)
	}

	return composite, grantStore
}

// wirePostAgent handles A2A, P2P executor, routes, and audit after agent creation.
func wirePostAgent(app *App, r appinit.Resolver, tools []*agent.Tool, bus *eventbus.Bus, composite *approval.CompositeProvider, grantStore *approval.GrantStore, boot *bootstrap.Result, auth *gateway.AuthManager) {
	cfg := app.Config
	adkAgent := app.Agent

	// A2A Server.
	if cfg.A2A.Enabled && cfg.Agent.MultiAgent && adkAgent.ADKAgent() != nil {
		a2aServer := a2a.NewServer(cfg.A2A, adkAgent.ADKAgent(), logger())
		a2aServer.RegisterRoutes(app.Gateway.Router())
	}

	// P2P executor + REST API routes.
	p2pc, _ := r.Resolve(appinit.ProvidesP2P).(*p2pComponents)
	pc, _ := r.Resolve(appinit.ProvidesPayment).(*paymentComponents)
	if p2pc != nil {
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
				switch v := result.(type) {
				case map[string]interface{}:
					return v, nil
				default:
					return map[string]interface{}{"result": v}, nil
				}
			})

			// Sandbox executor for P2P tool isolation.
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

			// Owner approval callback for inbound remote tool invocations.
			if pc != nil {
				p2pc.handler.SetApprovalFunc(func(ctx context.Context, peerDID, toolName string, params map[string]interface{}) (bool, error) {
					t, known := toolIndex[toolName]
					if !known || t.SafetyLevel.IsDangerous() {
						goto requireApproval
					}
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
						return false, nil
					}
					if resp.Approved && grantStore != nil {
						grantStore.Grant("p2p:"+peerDID, toolName)
					}
					return resp.Approved, nil
				})
			}
		}
		registerP2PRoutes(app.Gateway.Router(), app, p2pc, auth)
		logger().Info("P2P REST API routes registered")
	}

	// Observability API routes.
	obsc, _ := r.Resolve(appinit.ProvidesObservability).(*observabilityComponents)
	if obsc != nil {
		registerObservabilityRoutes(app.Gateway.Router(), obsc.collector, obsc.healthRegistry, obsc.tokenStore)
		logger().Info("observability API routes registered")
	}

	// Audit recorder.
	if cfg.Observability.Audit.Enabled && boot.DBClient != nil {
		auditRec := audit.NewRecorder(boot.DBClient)
		auditRec.Subscribe(bus)
		logger().Info("audit recorder wired to event bus")
	}
}

// wireMemoryAndTurnCallbacks wires memory compaction and gateway turn callbacks.
func wireMemoryAndTurnCallbacks(app *App, iv *intelligenceValues, fv *foundationValues) {
	// Memory compaction.
	if iv != nil && iv.MC != nil && iv.MC.buffer != nil {
		if entStore, ok := fv.Store.(*session.EntStore); ok {
			iv.MC.buffer.SetCompactor(entStore.CompactMessages)
			logger().Info("observational memory compaction wired")
		}
	}

	// Gateway turn callbacks for buffer triggers.
	if app.MemoryBuffer != nil {
		app.TurnRunner.OnTurnComplete(func(sessionKey string) {
			app.MemoryBuffer.Trigger(sessionKey)
		})
	}
	if iv != nil && iv.AB != nil {
		if ab, ok := iv.AB.(interface{ Trigger(string) }); ok {
			app.TurnRunner.OnTurnComplete(func(sessionKey string) {
				ab.Trigger(sessionKey)
			})
		}
	}
	if app.LibrarianProactiveBuffer != nil {
		app.TurnRunner.OnTurnComplete(func(sessionKey string) {
			app.LibrarianProactiveBuffer.Trigger(sessionKey)
		})
	}
}

func initTurnTraceStore(store session.Store) turntrace.Store {
	entStore, ok := store.(*session.EntStore)
	if !ok {
		return nil
	}
	return turntrace.NewEntStore(entStore.Client())
}

// registerPostBuildLifecycle registers gateway and channel lifecycle components.
// Module components are already registered from buildResult.Components.
func registerPostBuildLifecycle(app *App) {
	reg := app.registry

	// Gateway.
	reg.Register(lifecycle.NewFuncComponent("gateway",
		func(_ context.Context, wg *sync.WaitGroup) error {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := app.Gateway.Start(); err != nil {
					logger().Errorw("gateway server error", "error", err)
				}
			}()
			return nil
		},
		func(ctx context.Context) error {
			return app.Gateway.Shutdown(ctx)
		},
	), lifecycle.PriorityNetwork)

	// Channels.
	for i, ch := range app.Channels {
		ch := ch
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
			func(ctx context.Context) error {
				return ch.Stop(ctx)
			},
		), lifecycle.PriorityNetwork)
	}
}

// Resolver helper functions for safe type assertions.
func resolveKC(iv *intelligenceValues) *knowledgeComponents {
	if iv == nil {
		return nil
	}
	return iv.KC
}
func resolveMC(iv *intelligenceValues) *memoryComponents {
	if iv == nil {
		return nil
	}
	return iv.MC
}
func resolveEC(iv *intelligenceValues) *embeddingComponents {
	if iv == nil {
		return nil
	}
	return iv.EC
}
func resolveGC(iv *intelligenceValues) *graphComponents {
	if iv == nil {
		return nil
	}
	return iv.GC
}
func resolveLC(iv *intelligenceValues) *librarianComponents {
	if iv == nil {
		return nil
	}
	return iv.LC
}
func resolveSR(iv *intelligenceValues) *skill.Registry {
	if iv == nil || iv.SkillRegistry == nil {
		return nil
	}
	sr, _ := iv.SkillRegistry.(*skill.Registry)
	return sr
}

// Start starts the application services using the lifecycle registry.
func (a *App) Start(ctx context.Context) error {
	logger().Info("starting application")
	return a.registry.StartAll(ctx, &a.wg)
}

// Stop stops the application services and waits for all goroutines to exit.
func (a *App) Stop(ctx context.Context) error {
	logger().Info("stopping application")

	// Cancel app-level context to signal fire-and-forget goroutines.
	if a.cancel != nil {
		a.cancel()
	}

	// Stop all lifecycle-managed components in reverse startup order.
	stopErr := a.registry.StopAll(ctx)

	// Wait for all background goroutines to finish.
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	var waitErr error
	select {
	case <-done:
		logger().Info("all services stopped")
	case <-ctx.Done():
		waitErr = ctx.Err()
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

	return errors.Join(stopErr, waitErr)
}

// logToolRegistrationSummary logs a diagnostic summary of registered tool categories.
func logToolRegistrationSummary(catalog *toolcatalog.Catalog) {
	categories := catalog.ListCategories()
	var enabledNames []string
	var disabledNames []string
	for _, cat := range categories {
		if cat.Enabled {
			enabledNames = append(enabledNames, fmt.Sprintf("%s(%d)", cat.Name, len(catalog.ToolNamesForCategory(cat.Name))))
		} else {
			disabledNames = append(disabledNames, cat.Name)
		}
	}
	logger().Infow("tool registration complete",
		"total", catalog.ToolCount(),
		"enabled", strings.Join(enabledNames, ", "),
		"disabled", strings.Join(disabledNames, ", "),
	)
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
