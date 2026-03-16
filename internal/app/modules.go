package app

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/agentmemory"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/gatekeeper"
	"github.com/langoai/lango/internal/lifecycle"
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
	Supervisor  *supervisor.Supervisor
	Store       session.Store
	Crypto      security.CryptoProvider
	Keys        *security.KeyRegistry
	Secrets     *security.SecretsStore
	BrowserSM   *browser.SessionManager
	Refs        *security.RefStore
	Scanner     *agent.SecretScanner
	Sanitizer   *gatekeeper.Sanitizer
	CmdGuard    *execpkg.CommandGuard
	FsConfig    filesystem.Config
	AutoAvail   map[string]bool
}

// intelligenceValues holds the outputs of the intelligence module.
type intelligenceValues struct {
	KC *knowledgeComponents
	MC *memoryComponents
	EC *embeddingComponents
	GC *graphComponents
	LC *librarianComponents
	AB interface{} // *learning.AnalysisBuffer
	SkillRegistry interface{}
	AgentMemoryStore agentmemory.Store
}

// automationValues holds the outputs of the automation module.
type automationValues struct {
	CronScheduler     interface{}
	BackgroundManager interface{}
	WorkflowEngine    interface{}
}

// networkValues holds the outputs of the network module.
type networkValues struct {
	PC   *paymentComponents
	P2PC *p2pComponents
	EconC *economyComponents
	CC   *contractComponents
	SAC  *smartAccountComponents
	WSC  *wsComponents
	X402 interface{}
}

// ─── Foundation Module ───

type foundationModule struct {
	cfg  *config.Config
	boot *bootstrap.Result
}

func (m *foundationModule) Name() string         { return "foundation" }
func (m *foundationModule) Provides() []appinit.Provides {
	return []appinit.Provides{appinit.ProvidesSupervisor, appinit.ProvidesSessionStore, appinit.ProvidesSecurity}
}
func (m *foundationModule) DependsOn() []appinit.Provides { return nil }
func (m *foundationModule) Enabled() bool                  { return true }

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
			appinit.ProvidesSupervisor:   &foundationValues{
				Supervisor:  sv,
				Store:       store,
				Crypto:      crypto,
				Keys:        keys,
				Secrets:     secrets,
				BrowserSM:   browserSM,
				Refs:        refs,
				Scanner:     scanner,
				Sanitizer:   san,
				CmdGuard:    cmdGuard,
				FsConfig:    fsConfig,
				AutoAvail:   automationAvailable,
			},
			appinit.ProvidesSessionStore: store,
			appinit.ProvidesSecurity:     crypto,
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

func (m *intelligenceModule) Name() string         { return "intelligence" }
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

	// Graph Store (before knowledge).
	gc := initGraphStore(cfg)

	// Skills.
	baseTools := r.Resolve(appinit.ProvidesSessionStore) // used indirectly via foundation tools
	_ = baseTools
	skillReg := initSkills(cfg, nil) // skills don't depend on base tools for init
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

	return &appinit.ModuleResult{
		Tools:          tools,
		CatalogEntries: entries,
		Values: map[appinit.Provides]interface{}{
			appinit.ProvidesKnowledge: &intelligenceValues{
				KC: kc, MC: mc, EC: ec, GC: gc, LC: lc, AB: ab,
				SkillRegistry: skillReg, AgentMemoryStore: amStore,
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
	cfg  *config.Config
	app  *App // needed for AgentRunner interface at runtime
}

func (m *automationModule) Name() string         { return "automation" }
func (m *automationModule) Provides() []appinit.Provides {
	return []appinit.Provides{appinit.ProvidesAutomation}
}
func (m *automationModule) DependsOn() []appinit.Provides {
	return []appinit.Provides{appinit.ProvidesSessionStore}
}
func (m *automationModule) Enabled() bool {
	return m.cfg.Cron.Enabled || m.cfg.Background.Enabled || m.cfg.Workflow.Enabled
}

func (m *automationModule) Init(ctx context.Context, r appinit.Resolver) (*appinit.ModuleResult, error) {
	cfg := m.cfg
	fv := r.Resolve(appinit.ProvidesSupervisor).(*foundationValues)
	store := fv.Store

	var tools []*agent.Tool
	var entries []appinit.CatalogEntry
	var components []lifecycle.ComponentEntry

	cron := initCron(cfg, store, m.app)
	if cron != nil {
		cronTools := buildCronTools(cron, cfg.Cron.DefaultDeliverTo)
		tools = append(tools, cronTools...)
		entries = append(entries, appinit.CatalogEntry{Category: "cron", Description: "Cron job scheduling", ConfigKey: "cron.enabled", Enabled: true, Tools: cronTools})
		logger().Info("cron tools registered")
	}

	bg := initBackground(cfg, m.app)
	if bg != nil {
		bgTools := buildBackgroundTools(bg, cfg.Background.DefaultDeliverTo)
		tools = append(tools, bgTools...)
		entries = append(entries, appinit.CatalogEntry{Category: "background", Description: "Background task execution", ConfigKey: "background.enabled", Enabled: true, Tools: bgTools})
		logger().Info("background tools registered")
	}

	wf := initWorkflow(cfg, store, m.app)
	if wf != nil {
		wfTools := buildWorkflowTools(wf, cfg.Workflow.StateDir, cfg.Workflow.DefaultDeliverTo)
		tools = append(tools, wfTools...)
		entries = append(entries, appinit.CatalogEntry{Category: "workflow", Description: "Workflow pipeline execution", ConfigKey: "workflow.enabled", Enabled: true, Tools: wfTools})
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
	cfg      *config.Config
	boot     *bootstrap.Result
	bus      *eventbus.Bus
	app      *App
	registry *lifecycle.Registry
}

func (m *networkModule) Name() string         { return "network" }
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

	pc := initPayment(cfg, fv.Store, fv.Secrets)
	var p2pc *p2pComponents
	var econc *economyComponents
	var cc *contractComponents
	var sacc *smartAccountComponents

	if pc != nil {
		xc := initX402(cfg, fv.Secrets, pc.limiter)
		var x402Interceptor *x402pkg.Interceptor
		if xc != nil {
			x402Interceptor = xc.interceptor
		}

		pt := buildPaymentTools(pc, x402Interceptor)
		tools = append(tools, pt...)
		entries = append(entries, appinit.CatalogEntry{Category: "payment", Description: "Blockchain payments (USDC on Base)", ConfigKey: "payment.enabled", Enabled: true, Tools: pt})

		// P2P.
		p2pc = initP2P(cfg, pc.wallet, pc, m.boot.DBClient, fv.Secrets, m.bus)
		if p2pc != nil {
			p2pTools := buildP2PTools(p2pc)
			p2pTools = append(p2pTools, buildP2PPaymentTool(p2pc, pc)...)
			p2pTools = append(p2pTools, buildP2PPaidInvokeTool(p2pc, pc)...)
			tools = append(tools, p2pTools...)
			entries = append(entries, appinit.CatalogEntry{Category: "p2p", Description: "Peer-to-peer networking", ConfigKey: "p2p.enabled", Enabled: true, Tools: p2pTools})

			if p2pc.coordinator != nil {
				teamTools := buildTeamTools(p2pc.coordinator)
				tools = append(tools, teamTools...)
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
			registerEconomyLifecycle(m.registry, econc)
		} else {
			entries = append(entries, appinit.CatalogEntry{Category: "economy", Description: "P2P economy (disabled)", ConfigKey: "economy.enabled", Enabled: false})
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
		sacc = initSmartAccount(cfg, pc, econc, m.bus, m.registry)
		if sacc != nil {
			saTools := buildSmartAccountTools(sacc)
			tools = append(tools, saTools...)
			entries = append(entries, appinit.CatalogEntry{Category: "smartaccount", Description: "ERC-7579 smart account management", ConfigKey: "smartAccount.enabled", Enabled: true, Tools: saTools})
		} else {
			entries = append(entries, appinit.CatalogEntry{Category: "smartaccount", Description: "ERC-7579 smart account management (disabled)", ConfigKey: "smartAccount.enabled", Enabled: false})
		}

		_ = x402Interceptor
	} else {
		entries = append(entries, appinit.CatalogEntry{Category: "payment", Description: "Blockchain payments (disabled)", ConfigKey: "payment.enabled", Enabled: false})
		entries = append(entries, appinit.CatalogEntry{Category: "contract", Description: "Smart contract interaction (disabled)", ConfigKey: "payment.enabled", Enabled: false})
		if cfg.P2P.Enabled {
			entries = append(entries, appinit.CatalogEntry{Category: "p2p", Description: "P2P networking (disabled — payment required)", ConfigKey: "p2p.enabled", Enabled: false})
		} else {
			entries = append(entries, appinit.CatalogEntry{Category: "p2p", Description: "P2P networking (disabled)", ConfigKey: "p2p.enabled", Enabled: false})
		}
	}

	return &appinit.ModuleResult{
		Tools:          tools,
		CatalogEntries: entries,
		Values: map[appinit.Provides]interface{}{
			appinit.ProvidesPayment: pc,
			appinit.ProvidesP2P:     p2pc,
			appinit.ProvidesEconomy: econc,
			appinit.ProvidesContract: cc,
			appinit.ProvidesSmartAccount: sacc,
		},
	}, nil
}

// ─── Extension Module ───

type extensionModule struct {
	cfg      *config.Config
	boot     *bootstrap.Result
	bus      *eventbus.Bus
}

func (m *extensionModule) Name() string         { return "extension" }
func (m *extensionModule) Provides() []appinit.Provides {
	return []appinit.Provides{ProvidesMCPKey, ProvidesObsKey}
}
func (m *extensionModule) DependsOn() []appinit.Provides { return nil }
func (m *extensionModule) Enabled() bool                  { return true }

// Internal keys to avoid import cycle with appinit when needed.
const (
	ProvidesMCPKey appinit.Provides = "mcp"
	ProvidesObsKey appinit.Provides = "observability"
)

func (m *extensionModule) Init(ctx context.Context, r appinit.Resolver) (*appinit.ModuleResult, error) {
	cfg := m.cfg

	var tools []*agent.Tool
	var entries []appinit.CatalogEntry

	// MCP.
	mcpc := initMCP(cfg)
	if mcpc != nil {
		tools = append(tools, mcpc.tools...)
		entries = append(entries, appinit.CatalogEntry{Category: "mcp", Description: "MCP plugin tools (external servers)", ConfigKey: "mcp.enabled", Enabled: true, Tools: mcpc.tools})
		mgmtTools := buildMCPManagementTools(mcpc.manager)
		tools = append(tools, mgmtTools...)
	} else {
		entries = append(entries, appinit.CatalogEntry{Category: "mcp", Description: "MCP plugins (disabled)", ConfigKey: "mcp.enabled", Enabled: false})
	}

	// Observability.
	obsc := initObservability(cfg, m.boot.DBClient, m.bus)

	return &appinit.ModuleResult{
		Tools:          tools,
		CatalogEntries: entries,
		Values: map[appinit.Provides]interface{}{
			ProvidesMCPKey: mcpc,
			ProvidesObsKey: obsc,
		},
	}, nil
}

