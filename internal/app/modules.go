package app

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/agentmemory"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/gatekeeper"
	"github.com/langoai/lango/internal/lifecycle"
	"github.com/langoai/lango/internal/p2p/gitbundle"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/supervisor"
	"github.com/langoai/lango/internal/tools/browser"
	execpkg "github.com/langoai/lango/internal/tools/exec"
	"github.com/langoai/lango/internal/tools/filesystem"
	x402pkg "github.com/langoai/lango/internal/x402"
)

// ─── Value holder types for Resolver ───

// foundationValues holds the outputs of the foundation module.
type foundationValues struct {
	Supervisor *supervisor.Supervisor
	Store      session.Store
	Crypto     security.CryptoProvider
	Keys       *security.KeyRegistry
	Secrets    *security.SecretsStore
	BrowserSM  *browser.SessionManager
	Refs       *security.RefStore
	Scanner    *agent.SecretScanner
	Sanitizer  *gatekeeper.Sanitizer
	CmdGuard   *execpkg.CommandGuard
	FsConfig   filesystem.Config
	AutoAvail  map[string]bool
}

// intelligenceValues holds the outputs of the intelligence module.
type intelligenceValues struct {
	KC               *knowledgeComponents
	MC               *memoryComponents
	EC               *embeddingComponents
	GC               *graphComponents
	LC               *librarianComponents
	AB               interface{} // *learning.AnalysisBuffer
	Observer         interface{} // learning.Observer — for WithLearning middleware
	SkillRegistry    interface{}
	AgentMemoryStore agentmemory.Store
}

// automationValues holds the outputs of the automation module.
type automationValues struct {
	CronScheduler     interface{}
	BackgroundManager interface{}
	WorkflowEngine    interface{}
}

// ─── Foundation Module ───

type foundationModule struct {
	cfg  *config.Config
	boot *bootstrap.Result
}

func (m *foundationModule) Name() string { return "foundation" }
func (m *foundationModule) Provides() []appinit.Provides {
	return []appinit.Provides{appinit.ProvidesSupervisor, appinit.ProvidesSessionStore, appinit.ProvidesSecurity}
}
func (m *foundationModule) DependsOn() []appinit.Provides { return nil }
func (m *foundationModule) Enabled() bool                 { return true }

func (m *foundationModule) Init(ctx context.Context, r appinit.Resolver) (*appinit.ModuleResult, error) {
	cfg := m.cfg

	sv, err := initSupervisor(cfg)
	if err != nil {
		return nil, fmt.Errorf("create supervisor: %w", err)
	}

	var san *gatekeeper.Sanitizer
	if s, initErr := gatekeeper.NewSanitizer(cfg.Gatekeeper); initErr != nil {
		logger().Warnw("gatekeeper sanitizer init error, disabled", "error", initErr)
	} else {
		san = s
	}

	store, err := initSessionStore(cfg, m.boot)
	if err != nil {
		return nil, fmt.Errorf("create session store: %w", err)
	}

	crypto, keys, secrets, err := initSecurity(cfg, store, m.boot)
	if err != nil {
		return nil, fmt.Errorf("security init: %w", err)
	}

	// Base tools: exec, filesystem, browser.
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
		logger().Info("browser tools enabled")
	}

	automationAvailable := map[string]bool{
		"cron":       cfg.Cron.Enabled,
		"background": cfg.Background.Enabled,
		"workflow":   cfg.Workflow.Enabled,
	}
	protectedPaths := []string{cfg.DataRoot}
	protectedPaths = append(protectedPaths, cfg.Tools.Exec.AdditionalProtectedPaths...)
	cmdGuard := execpkg.NewCommandGuard(protectedPaths)

	baseTools := buildTools(sv, fsConfig, browserSM, automationAvailable, cmdGuard)

	refs := security.NewRefStore()
	scanner := agent.NewSecretScanner()
	registerConfigSecrets(scanner, cfg)

	// Crypto tools.
	var cryptoTools []*agent.Tool
	if crypto != nil && keys != nil {
		cryptoTools = buildCryptoTools(crypto, keys, refs, scanner)
	}

	// Secrets tools.
	var secretsTools []*agent.Tool
	if secrets != nil {
		secretsTools = buildSecretsTools(secrets, refs, scanner)
	}

	allTools := append(baseTools, cryptoTools...)
	allTools = append(allTools, secretsTools...)

	// Catalog entries for base tools.
	entries := buildFoundationCatalogEntries(cfg, baseTools, cryptoTools, secretsTools)

	return &appinit.ModuleResult{
		Tools:          allTools,
		CatalogEntries: entries,
		Values: map[appinit.Provides]interface{}{
			appinit.ProvidesSupervisor: &foundationValues{
				Supervisor: sv,
				Store:      store,
				Crypto:     crypto,
				Keys:       keys,
				Secrets:    secrets,
				BrowserSM:  browserSM,
				Refs:       refs,
				Scanner:    scanner,
				Sanitizer:  san,
				CmdGuard:   cmdGuard,
				FsConfig:   fsConfig,
				AutoAvail:  automationAvailable,
			},
			appinit.ProvidesSessionStore: store,
			appinit.ProvidesSecurity:     crypto,
			appinit.ProvidesBaseTools:    allTools,
		},
	}, nil
}

func buildFoundationCatalogEntries(cfg *config.Config, base, crypto, secrets []*agent.Tool) []appinit.CatalogEntry {
	var entries []appinit.CatalogEntry

	// Split base tools by prefix.
	var execTools, fsTools, browserTools []*agent.Tool
	for _, t := range base {
		switch {
		case len(t.Name) >= 4 && t.Name[:4] == "exec":
			execTools = append(execTools, t)
		case len(t.Name) >= 3 && t.Name[:3] == "fs_":
			fsTools = append(fsTools, t)
		case len(t.Name) >= 8 && t.Name[:8] == "browser_":
			browserTools = append(browserTools, t)
		}
	}

	entries = append(entries, appinit.CatalogEntry{Category: "exec", Description: "Shell command execution", Enabled: true, Tools: execTools})
	entries = append(entries, appinit.CatalogEntry{Category: "filesystem", Description: "File system operations", Enabled: true, Tools: fsTools})

	if cfg.Tools.Browser.Enabled {
		entries = append(entries, appinit.CatalogEntry{Category: "browser", Description: "Web browsing", ConfigKey: "tools.browser.enabled", Enabled: true, Tools: browserTools})
	} else {
		entries = append(entries, appinit.CatalogEntry{Category: "browser", Description: "Web browsing (disabled)", ConfigKey: "tools.browser.enabled", Enabled: false})
	}

	if len(crypto) > 0 {
		entries = append(entries, appinit.CatalogEntry{Category: "crypto", Description: "Cryptographic operations", ConfigKey: "security.signer.provider", Enabled: true, Tools: crypto})
	} else {
		entries = append(entries, appinit.CatalogEntry{Category: "crypto", Description: "Cryptographic operations (disabled)", ConfigKey: "security.signer.provider", Enabled: false})
	}

	if len(secrets) > 0 {
		entries = append(entries, appinit.CatalogEntry{Category: "secrets", Description: "Secret management", ConfigKey: "security.secrets.enabled", Enabled: true, Tools: secrets})
	} else {
		entries = append(entries, appinit.CatalogEntry{Category: "secrets", Description: "Secret management (disabled)", ConfigKey: "security.secrets.enabled", Enabled: false})
	}

	return entries
}

// ─── Intelligence Module ───

type intelligenceModule struct {
	cfg   *config.Config
	boot  *bootstrap.Result
	rawDB *sql.DB
}

func (m *intelligenceModule) Name() string { return "intelligence" }
func (m *intelligenceModule) Provides() []appinit.Provides {
	return []appinit.Provides{appinit.ProvidesKnowledge, appinit.ProvidesMemory, appinit.ProvidesEmbedding, appinit.ProvidesGraph, appinit.ProvidesLibrarian, appinit.ProvidesSkills}
}
func (m *intelligenceModule) DependsOn() []appinit.Provides {
	return []appinit.Provides{appinit.ProvidesSessionStore, appinit.ProvidesSupervisor}
}
func (m *intelligenceModule) Enabled() bool { return true } // always enabled — subsystems check their own config

func (m *intelligenceModule) Init(ctx context.Context, r appinit.Resolver) (*appinit.ModuleResult, error) {
	cfg := m.cfg
	fv := r.Resolve(appinit.ProvidesSupervisor).(*foundationValues)
	store := fv.Store
	sv := fv.Supervisor

	var tools []*agent.Tool
	var entries []appinit.CatalogEntry
	var components []lifecycle.ComponentEntry

	// Graph Store (before knowledge).
	gc := initGraphStore(cfg)

	// Skills — resolve base tools from foundation for skill init.
	var baseToolSlice []*agent.Tool
	if bt := r.Resolve(appinit.ProvidesBaseTools); bt != nil {
		baseToolSlice, _ = bt.([]*agent.Tool)
	}
	skillReg := initSkills(cfg, baseToolSlice)
	if skillReg != nil {
		tools = append(tools, skillReg.LoadedSkills()...)
	}

	// Knowledge.
	kc := initKnowledge(cfg, store, gc)
	if kc != nil {
		metaTools := buildMetaTools(kc.store, kc.engine, skillReg, cfg.Skill)
		tools = append(tools, metaTools...)
		entries = append(entries, appinit.CatalogEntry{Category: "meta", Description: "Knowledge, learning, and skill management", ConfigKey: "knowledge.enabled", Enabled: true, Tools: metaTools})
	} else {
		entries = append(entries, appinit.CatalogEntry{Category: "meta", Description: "Knowledge & learning (disabled)", ConfigKey: "knowledge.enabled", Enabled: false})
	}

	// Observational Memory.
	mc := initMemory(cfg, store, sv)

	// Embedding / RAG.
	ec := initEmbedding(cfg, m.rawDB, kc, mc)

	// Graph callbacks.
	if gc != nil {
		wireGraphCallbacks(gc, kc, mc, sv, cfg)
		initGraphRAG(cfg, gc, ec)
	}

	// Conversation Analysis.
	ab := initConversationAnalysis(cfg, sv, store, kc, gc)

	// Librarian.
	lc := initLibrarian(cfg, sv, store, kc, mc, gc)

	// Graph tools.
	if gc != nil {
		gt := buildGraphTools(gc.store)
		tools = append(tools, gt...)
		entries = append(entries, appinit.CatalogEntry{Category: "graph", Description: "Knowledge graph traversal", ConfigKey: "graph.enabled", Enabled: true, Tools: gt})
	} else {
		entries = append(entries, appinit.CatalogEntry{Category: "graph", Description: "Knowledge graph (disabled)", ConfigKey: "graph.enabled", Enabled: false})
	}

	// RAG tools.
	if ec != nil && ec.ragService != nil {
		rt := buildRAGTools(ec.ragService)
		tools = append(tools, rt...)
		entries = append(entries, appinit.CatalogEntry{Category: "rag", Description: "Retrieval-augmented generation", ConfigKey: "embedding.rag.enabled", Enabled: true, Tools: rt})
	} else {
		entries = append(entries, appinit.CatalogEntry{Category: "rag", Description: "RAG retrieval (disabled)", ConfigKey: "embedding.provider", Enabled: false})
	}

	// Memory tools.
	if mc != nil {
		mt := buildMemoryAgentTools(mc.store)
		tools = append(tools, mt...)
		entries = append(entries, appinit.CatalogEntry{Category: "memory", Description: "Observational memory", ConfigKey: "observationalMemory.enabled", Enabled: true, Tools: mt})
	} else {
		entries = append(entries, appinit.CatalogEntry{Category: "memory", Description: "Observational memory (disabled)", ConfigKey: "observationalMemory.enabled", Enabled: false})
	}

	// Agent Memory.
	var amStore agentmemory.Store
	if cfg.AgentMemory.Enabled {
		amStore = agentmemory.NewInMemoryStore()
		amTools := buildAgentMemoryTools(amStore)
		tools = append(tools, amTools...)
		entries = append(entries, appinit.CatalogEntry{Category: "agent_memory", Description: "Per-agent persistent memory", ConfigKey: "agentMemory.enabled", Enabled: true, Tools: amTools})
		logger().Info("agent memory tools enabled")
	} else {
		entries = append(entries, appinit.CatalogEntry{Category: "agent_memory", Description: "Per-agent memory (disabled)", ConfigKey: "agentMemory.enabled", Enabled: false})
	}

	// Librarian tools.
	if lc != nil {
		lt := buildLibrarianTools(lc.inquiryStore)
		tools = append(tools, lt...)
		entries = append(entries, appinit.CatalogEntry{Category: "librarian", Description: "Knowledge inquiries and gap detection", ConfigKey: "librarian.enabled", Enabled: true, Tools: lt})
	} else {
		entries = append(entries, appinit.CatalogEntry{Category: "librarian", Description: "Knowledge inquiries (disabled)", ConfigKey: "librarian.enabled", Enabled: false})
	}

	// Lifecycle components for buffers.
	if mc != nil && mc.buffer != nil {
		components = append(components, lifecycle.ComponentEntry{
			Component: lifecycle.NewSimpleComponent("memory-buffer", mc.buffer),
			Priority:  lifecycle.PriorityBuffer,
		})
	}
	if ec != nil && ec.buffer != nil {
		components = append(components, lifecycle.ComponentEntry{
			Component: lifecycle.NewSimpleComponent("embedding-buffer", ec.buffer),
			Priority:  lifecycle.PriorityBuffer,
		})
	}
	if gc != nil && gc.buffer != nil {
		components = append(components, lifecycle.ComponentEntry{
			Component: lifecycle.NewSimpleComponent("graph-buffer", gc.buffer),
			Priority:  lifecycle.PriorityBuffer,
		})
	}
	if ab != nil {
		components = append(components, lifecycle.ComponentEntry{
			Component: lifecycle.NewSimpleComponent("analysis-buffer", ab),
			Priority:  lifecycle.PriorityBuffer,
		})
	}
	if lc != nil && lc.proactiveBuffer != nil {
		components = append(components, lifecycle.ComponentEntry{
			Component: lifecycle.NewSimpleComponent("librarian-proactive-buffer", lc.proactiveBuffer),
			Priority:  lifecycle.PriorityBuffer,
		})
	}

	// Observer for WithLearning middleware.
	var observer interface{}
	if kc != nil {
		observer = kc.observer
	}

	return &appinit.ModuleResult{
		Tools:          tools,
		Components:     components,
		CatalogEntries: entries,
		Values: map[appinit.Provides]interface{}{
			appinit.ProvidesKnowledge: &intelligenceValues{
				KC: kc, MC: mc, EC: ec, GC: gc, LC: lc, AB: ab,
				Observer:         observer,
				SkillRegistry:    skillReg,
				AgentMemoryStore: amStore,
			},
			appinit.ProvidesGraph:     gc,
			appinit.ProvidesMemory:    mc,
			appinit.ProvidesEmbedding: ec,
			appinit.ProvidesLibrarian: lc,
			appinit.ProvidesSkills:    skillReg,
		},
	}, nil
}

// ─── Automation Module ───

type automationModule struct {
	cfg *config.Config
	app *App // needed for AgentRunner interface at runtime
}

func (m *automationModule) Name() string { return "automation" }
func (m *automationModule) Provides() []appinit.Provides {
	return []appinit.Provides{appinit.ProvidesAutomation}
}
func (m *automationModule) DependsOn() []appinit.Provides {
	return []appinit.Provides{appinit.ProvidesSessionStore, appinit.ProvidesRunLedger}
}
func (m *automationModule) Enabled() bool {
	return m.cfg.Cron.Enabled || m.cfg.Background.Enabled || m.cfg.Workflow.Enabled
}

func (m *automationModule) Init(ctx context.Context, r appinit.Resolver) (*appinit.ModuleResult, error) {
	cfg := m.cfg
	fv := r.Resolve(appinit.ProvidesSupervisor).(*foundationValues)
	store := fv.Store
	rlv, _ := r.Resolve(appinit.ProvidesRunLedger).(*runLedgerValues)

	var tools []*agent.Tool
	var entries []appinit.CatalogEntry
	var components []lifecycle.ComponentEntry

	cron := initCron(cfg, store, m.app)
	if cron != nil {
		cronTools := buildCronTools(cron, cfg.Cron.DefaultDeliverTo)
		tools = append(tools, cronTools...)
		entries = append(entries, appinit.CatalogEntry{Category: "cron", Description: "Cron job scheduling", ConfigKey: "cron.enabled", Enabled: true, Tools: cronTools})
		cs := cron // capture for closure
		components = append(components, lifecycle.ComponentEntry{
			Component: lifecycle.NewFuncComponent("cron-scheduler",
				func(ctx context.Context, _ *sync.WaitGroup) error { return cs.Start(ctx) },
				func(_ context.Context) error { cs.Stop(); return nil },
			),
			Priority: lifecycle.PriorityAutomation,
		})
		logger().Info("cron tools registered")
	}

	bg := initBackground(cfg, m.app)
	if bg != nil {
		bgTools := buildBackgroundTools(bg, cfg.Background.DefaultDeliverTo)
		tools = append(tools, bgTools...)
		entries = append(entries, appinit.CatalogEntry{Category: "background", Description: "Background task execution", ConfigKey: "background.enabled", Enabled: true, Tools: bgTools})
		bm := bg // capture for closure
		components = append(components, lifecycle.ComponentEntry{
			Component: lifecycle.NewFuncComponent("background-manager",
				func(_ context.Context, _ *sync.WaitGroup) error { return nil },
				func(_ context.Context) error { bm.Shutdown(); return nil },
			),
			Priority: lifecycle.PriorityAutomation,
		})
		logger().Info("background tools registered")
	}

	wf := initWorkflow(cfg, store, m.app, rlv)
	if wf != nil {
		wfTools := buildWorkflowTools(wf, cfg.Workflow.StateDir, cfg.Workflow.DefaultDeliverTo)
		tools = append(tools, wfTools...)
		entries = append(entries, appinit.CatalogEntry{Category: "workflow", Description: "Workflow pipeline execution", ConfigKey: "workflow.enabled", Enabled: true, Tools: wfTools})
		we := wf // capture for closure
		components = append(components, lifecycle.ComponentEntry{
			Component: lifecycle.NewFuncComponent("workflow-engine",
				func(_ context.Context, _ *sync.WaitGroup) error { return nil },
				func(_ context.Context) error { we.Shutdown(); return nil },
			),
			Priority: lifecycle.PriorityAutomation,
		})
		logger().Info("workflow tools registered")
	}

	// Disabled category entries.
	if !cfg.Cron.Enabled {
		entries = append(entries, appinit.CatalogEntry{Category: "cron", Description: "Cron job scheduling (disabled)", ConfigKey: "cron.enabled", Enabled: false})
	}
	if !cfg.Background.Enabled {
		entries = append(entries, appinit.CatalogEntry{Category: "background", Description: "Background task execution (disabled)", ConfigKey: "background.enabled", Enabled: false})
	}
	if !cfg.Workflow.Enabled {
		entries = append(entries, appinit.CatalogEntry{Category: "workflow", Description: "Workflow pipeline execution (disabled)", ConfigKey: "workflow.enabled", Enabled: false})
	}

	return &appinit.ModuleResult{
		Tools:          tools,
		Components:     components,
		CatalogEntries: entries,
		Values: map[appinit.Provides]interface{}{
			appinit.ProvidesAutomation: &automationValues{
				CronScheduler:     cron,
				BackgroundManager: bg,
				WorkflowEngine:    wf,
			},
		},
	}, nil
}

// ─── Network Module ───

type networkModule struct {
	cfg  *config.Config
	boot *bootstrap.Result
	bus  *eventbus.Bus
	app  *App
}

func (m *networkModule) Name() string { return "network" }
func (m *networkModule) Provides() []appinit.Provides {
	return []appinit.Provides{appinit.ProvidesPayment, appinit.ProvidesP2P, appinit.ProvidesEconomy, appinit.ProvidesContract, appinit.ProvidesSmartAccount, appinit.ProvidesWorkspace}
}
func (m *networkModule) DependsOn() []appinit.Provides {
	return []appinit.Provides{appinit.ProvidesSecurity, appinit.ProvidesSessionStore}
}
func (m *networkModule) Enabled() bool {
	return m.cfg.Payment.Enabled || m.cfg.P2P.Enabled || m.cfg.Economy.Enabled
}

func (m *networkModule) Init(ctx context.Context, r appinit.Resolver) (*appinit.ModuleResult, error) {
	cfg := m.cfg
	fv := r.Resolve(appinit.ProvidesSupervisor).(*foundationValues)

	var tools []*agent.Tool
	var entries []appinit.CatalogEntry
	var components []lifecycle.ComponentEntry

	pc := initPayment(cfg, fv.Store, fv.Secrets)
	var p2pc *p2pComponents
	var econc *economyComponents
	var cc *contractComponents
	var sacc *smartAccountComponents
	var wsc *wsComponents
	var x402Interceptor *x402pkg.Interceptor

	if pc != nil {
		xc := initX402(cfg, fv.Secrets, pc.limiter)
		if xc != nil {
			x402Interceptor = xc.interceptor
		}

		pt := buildPaymentTools(pc, x402Interceptor)
		tools = append(tools, pt...)
		entries = append(entries, appinit.CatalogEntry{Category: "payment", Description: "Blockchain payments (USDC on Base)", ConfigKey: "payment.enabled", Enabled: true, Tools: pt})

		// P2P.
		p2pc = initP2P(cfg, pc.wallet, pc, m.boot.DBClient, fv.Secrets, m.bus)
		if p2pc != nil {
			// P2P Node lifecycle.
			if p2pc.node != nil {
				node := p2pc.node
				components = append(components, lifecycle.ComponentEntry{
					Component: lifecycle.NewFuncComponent("p2p-node",
						func(_ context.Context, wg *sync.WaitGroup) error { return node.Start(wg) },
						func(_ context.Context) error { return node.Stop() },
					),
					Priority: lifecycle.PriorityNetwork,
				})
			}

			// NonceCache lifecycle.
			if p2pc.nonceCache != nil {
				nc := p2pc.nonceCache
				components = append(components, lifecycle.ComponentEntry{
					Component: lifecycle.NewFuncComponent("p2p-nonce-cache",
						func(_ context.Context, _ *sync.WaitGroup) error { return nil },
						func(_ context.Context) error { nc.Stop(); return nil },
					),
					Priority: lifecycle.PriorityNetwork,
				})
			}

			p2pTools := buildP2PTools(p2pc)
			p2pTools = append(p2pTools, buildP2PPaymentTool(p2pc, pc)...)
			p2pTools = append(p2pTools, buildP2PPaidInvokeTool(p2pc, pc)...)
			tools = append(tools, p2pTools...)
			entries = append(entries, appinit.CatalogEntry{Category: "p2p", Description: "Peer-to-peer networking", ConfigKey: "p2p.enabled", Enabled: true, Tools: p2pTools})

			// Team coordination tools.
			if p2pc.coordinator != nil {
				teamTools := buildTeamTools(p2pc.coordinator)
				tools = append(tools, teamTools...)
			}

			// Workspace (optional, requires P2P node).
			var sessionValidator gitbundle.SessionValidator
			if p2pc.sessions != nil {
				sess := p2pc.sessions
				sessionValidator = func(token string) (string, bool) {
					return sess.GetByToken(token)
				}
			}

			var localDID string
			if p2pc.identity != nil {
				didCtx, didCancel := context.WithTimeout(context.Background(), 5*time.Second)
				d, idErr := p2pc.identity.DID(didCtx)
				didCancel()
				if idErr == nil && d != nil {
					localDID = d.ID
				}
			}

			wsc = initWorkspace(cfg, p2pc.node, localDID, sessionValidator)
			if wsc != nil {
				wsTools := buildWorkspaceTools(wsc)
				tools = append(tools, wsTools...)
				entries = append(entries, appinit.CatalogEntry{Category: "workspace", Description: "P2P collaborative workspaces and git sharing", ConfigKey: "p2p.workspace.enabled", Enabled: true, Tools: wsTools})

				// Workspace DB lifecycle.
				wsDB := wsc.db
				components = append(components, lifecycle.ComponentEntry{
					Component: lifecycle.NewFuncComponent("p2p-workspace-db",
						func(_ context.Context, _ *sync.WaitGroup) error { return nil },
						func(_ context.Context) error {
							if wsDB != nil {
								return wsDB.Close()
							}
							return nil
						},
					),
					Priority: lifecycle.PriorityNetwork,
				})

				// Workspace gossip lifecycle.
				if wsc.gossip != nil {
					wsGossip := wsc.gossip
					components = append(components, lifecycle.ComponentEntry{
						Component: lifecycle.NewFuncComponent("p2p-workspace-gossip",
							func(_ context.Context, _ *sync.WaitGroup) error { return nil },
							func(_ context.Context) error { wsGossip.Stop(); return nil },
						),
						Priority: lifecycle.PriorityNetwork,
					})
				}

				// Wire workspace-team bridge.
				if p2pc.coordinator != nil && wsc.manager != nil {
					wireWorkspaceTeamBridge(m.bus, wsc.manager, wsc.tracker, wsc.gossip, logger())
				}

				logger().Info("P2P workspace tools registered")
			} else if cfg.P2P.Workspace.Enabled {
				entries = append(entries, appinit.CatalogEntry{Category: "workspace", Description: "P2P workspaces (disabled)", ConfigKey: "p2p.workspace.enabled", Enabled: false})
			}

			// HealthMonitor lifecycle.
			if p2pc.healthMonitor != nil {
				components = append(components, lifecycle.ComponentEntry{
					Component: p2pc.healthMonitor,
					Priority:  lifecycle.PriorityAutomation,
				})
			}
		} else {
			entries = append(entries, appinit.CatalogEntry{Category: "p2p", Description: "P2P networking (disabled — payment required)", ConfigKey: "p2p.enabled", Enabled: false})
		}

		// Economy.
		econc = initEconomy(cfg, p2pc, pc, m.bus)
		if econc != nil {
			econTools := buildEconomyTools(econc)
			tools = append(tools, econTools...)
			entries = append(entries, appinit.CatalogEntry{Category: "economy", Description: "P2P economy (budget, risk, pricing, negotiation, escrow)", ConfigKey: "economy.enabled", Enabled: true, Tools: econTools})

			if econc.escrowEngine != nil && econc.escrowSettler != nil {
				escrowTools := buildOnChainEscrowTools(econc.escrowEngine, econc.escrowSettler)
				tools = append(tools, escrowTools...)
				entries = append(entries, appinit.CatalogEntry{Category: "escrow", Description: "On-chain escrow management", ConfigKey: "economy.escrow.enabled", Enabled: true, Tools: escrowTools})
			}
			if econc.sentinelEngine != nil {
				sentTools := buildSentinelTools(econc.sentinelEngine)
				tools = append(tools, sentTools...)
				entries = append(entries, appinit.CatalogEntry{Category: "sentinel", Description: "Security Sentinel anomaly detection", ConfigKey: "economy.escrow.enabled", Enabled: true, Tools: sentTools})
			}

			// Economy lifecycle components (EventMonitor, DanglingDetector).
			if econc.eventMonitor != nil {
				components = append(components, lifecycle.ComponentEntry{
					Component: econc.eventMonitor,
					Priority:  lifecycle.PriorityNetwork,
				})
			}
			if econc.danglingDetector != nil {
				components = append(components, lifecycle.ComponentEntry{
					Component: econc.danglingDetector,
					Priority:  lifecycle.PriorityAutomation,
				})
			}
		} else {
			entries = append(entries, appinit.CatalogEntry{Category: "economy", Description: "P2P economy (disabled)", ConfigKey: "economy.enabled", Enabled: false})
		}

		// Team-Economy Bridges (event-driven).
		if p2pc != nil && p2pc.coordinator != nil {
			if econc != nil && econc.escrowEngine != nil {
				wireTeamEscrowBridge(m.bus, econc.escrowEngine, p2pc.coordinator, logger())
			}
			if econc != nil && econc.budgetEngine != nil {
				wireTeamBudgetBridge(m.app.ctx, m.bus, econc.budgetEngine, p2pc.coordinator, logger())
			}
			if p2pc.reputation != nil {
				minRepScore := cfg.P2P.Team.MinReputationScore
				if minRepScore <= 0 {
					minRepScore = cfg.P2P.MinTrustScore
				}
				if minRepScore <= 0 {
					minRepScore = 0.3
				}
				initTeamReputationBridge(m.bus, p2pc.coordinator, p2pc.reputation, minRepScore, logger())
			}
			if econc != nil && econc.budgetEngine != nil {
				initTeamShutdownBridge(m.bus, p2pc.coordinator, logger())
			}
			// Team-Escrow convenience tools.
			if econc != nil && econc.escrowEngine != nil {
				teTools := buildTeamEscrowTools(p2pc.coordinator, econc.escrowEngine, econc.budgetEngine)
				tools = append(tools, teTools...)
			}
		}

		// Contract.
		cc = initContract(pc)
		if cc != nil {
			ctTools := buildContractTools(cc.caller)
			tools = append(tools, ctTools...)
			entries = append(entries, appinit.CatalogEntry{Category: "contract", Description: "Smart contract interaction", ConfigKey: "payment.enabled", Enabled: true, Tools: ctTools})
		} else {
			entries = append(entries, appinit.CatalogEntry{Category: "contract", Description: "Smart contract interaction (disabled)", ConfigKey: "payment.enabled", Enabled: false})
		}

		// Smart Account.
		saResult := initSmartAccount(cfg, pc, econc, m.bus)
		if saResult != nil {
			sacc = saResult.components
			components = append(components, saResult.lifecycle...)
			saTools := buildSmartAccountTools(sacc)
			tools = append(tools, saTools...)
			entries = append(entries, appinit.CatalogEntry{Category: "smartaccount", Description: "ERC-7579 smart account management", ConfigKey: "smartAccount.enabled", Enabled: true, Tools: saTools})
		} else {
			entries = append(entries, appinit.CatalogEntry{Category: "smartaccount", Description: "ERC-7579 smart account management (disabled)", ConfigKey: "smartAccount.enabled", Enabled: false})
		}
	} else {
		entries = append(entries, appinit.CatalogEntry{Category: "payment", Description: "Blockchain payments (disabled)", ConfigKey: "payment.enabled", Enabled: false})
		entries = append(entries, appinit.CatalogEntry{Category: "contract", Description: "Smart contract interaction (disabled)", ConfigKey: "payment.enabled", Enabled: false})
		if cfg.P2P.Enabled {
			entries = append(entries, appinit.CatalogEntry{Category: "p2p", Description: "P2P networking (disabled — payment required)", ConfigKey: "p2p.enabled", Enabled: false})
		} else {
			entries = append(entries, appinit.CatalogEntry{Category: "p2p", Description: "P2P networking (disabled)", ConfigKey: "p2p.enabled", Enabled: false})
		}
		if cfg.P2P.Workspace.Enabled {
			entries = append(entries, appinit.CatalogEntry{Category: "workspace", Description: "P2P workspaces (disabled)", ConfigKey: "p2p.workspace.enabled", Enabled: false})
		}
		entries = append(entries, appinit.CatalogEntry{Category: "smartaccount", Description: "ERC-7579 smart account management (disabled)", ConfigKey: "smartAccount.enabled", Enabled: false})
	}

	return &appinit.ModuleResult{
		Tools:          tools,
		Components:     components,
		CatalogEntries: entries,
		Values: map[appinit.Provides]interface{}{
			appinit.ProvidesPayment:      pc,
			appinit.ProvidesP2P:          p2pc,
			appinit.ProvidesEconomy:      econc,
			appinit.ProvidesContract:     cc,
			appinit.ProvidesSmartAccount: sacc,
			appinit.ProvidesWorkspace:    wsc,
		},
	}, nil
}

// ─── Extension Module ───

type extensionModule struct {
	cfg  *config.Config
	boot *bootstrap.Result
	bus  *eventbus.Bus
}

func (m *extensionModule) Name() string { return "extension" }
func (m *extensionModule) Provides() []appinit.Provides {
	return []appinit.Provides{appinit.ProvidesMCP, appinit.ProvidesObservability}
}
func (m *extensionModule) DependsOn() []appinit.Provides { return nil }
func (m *extensionModule) Enabled() bool                 { return true }

func (m *extensionModule) Init(ctx context.Context, r appinit.Resolver) (*appinit.ModuleResult, error) {
	cfg := m.cfg

	var tools []*agent.Tool
	var entries []appinit.CatalogEntry
	var components []lifecycle.ComponentEntry

	// MCP.
	mcpc := initMCP(cfg)
	if mcpc != nil {
		tools = append(tools, mcpc.tools...)
		entries = append(entries, appinit.CatalogEntry{Category: "mcp", Description: "MCP plugin tools (external servers)", ConfigKey: "mcp.enabled", Enabled: true, Tools: mcpc.tools})
		mgmtTools := buildMCPManagementTools(mcpc.manager)
		tools = append(tools, mgmtTools...)
		entries = append(entries, appinit.CatalogEntry{Category: "mcp", Description: "MCP management tools", ConfigKey: "mcp.enabled", Enabled: true, Tools: mgmtTools})
		// MCP Manager lifecycle.
		mgr := mcpc.manager
		components = append(components, lifecycle.ComponentEntry{
			Component: lifecycle.NewFuncComponent("mcp-manager",
				func(_ context.Context, _ *sync.WaitGroup) error { return nil },
				func(ctx context.Context) error { return mgr.DisconnectAll(ctx) },
			),
			Priority: lifecycle.PriorityNetwork,
		})
	} else {
		entries = append(entries, appinit.CatalogEntry{Category: "mcp", Description: "MCP plugins (disabled)", ConfigKey: "mcp.enabled", Enabled: false})
	}

	// Observability.
	obsc := initObservability(cfg, m.boot.DBClient, m.bus)
	if obsc == nil {
		entries = append(entries, appinit.CatalogEntry{Category: "observability", Description: "Metrics & health (disabled)", ConfigKey: "observability.enabled", Enabled: false})
	}

	// Observability token cleanup lifecycle.
	if obsc != nil && obsc.tokenStore != nil && cfg.Observability.Tokens.RetentionDays > 0 {
		retDays := cfg.Observability.Tokens.RetentionDays
		store := obsc.tokenStore
		components = append(components, lifecycle.ComponentEntry{
			Component: lifecycle.NewFuncComponent("observability-token-cleanup",
				func(_ context.Context, _ *sync.WaitGroup) error { return nil },
				func(ctx context.Context) error {
					count, err := store.Cleanup(ctx, retDays)
					if err != nil {
						logger().Warnw("token usage cleanup", "error", err)
						return nil
					}
					if count > 0 {
						logger().Infow("token usage cleanup", "deleted", count, "retentionDays", retDays)
					}
					return nil
				},
			),
			Priority: lifecycle.PriorityCore,
		})
	}

	return &appinit.ModuleResult{
		Tools:          tools,
		Components:     components,
		CatalogEntries: entries,
		Values: map[appinit.Provides]interface{}{
			appinit.ProvidesMCP:           mcpc,
			appinit.ProvidesObservability: obsc,
		},
	}, nil
}
