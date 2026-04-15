package app

import (
	"context"
	"io"
	"sync"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/agentmemory"
	"github.com/langoai/lango/internal/agentregistry"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/config"
	cronpkg "github.com/langoai/lango/internal/cron"
	"github.com/langoai/lango/internal/economy/budget"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/economy/negotiation"
	"github.com/langoai/lango/internal/economy/pricing"
	"github.com/langoai/lango/internal/economy/risk"
	"github.com/langoai/lango/internal/embedding"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/extension"
	"github.com/langoai/lango/internal/gatekeeper"
	"github.com/langoai/lango/internal/gateway"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/learning"
	"github.com/langoai/lango/internal/librarian"
	"github.com/langoai/lango/internal/lifecycle"
	"github.com/langoai/lango/internal/mcp"
	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/observability/health"
	"github.com/langoai/lango/internal/observability/token"
	"github.com/langoai/lango/internal/p2p"
	"github.com/langoai/lango/internal/p2p/agentpool"
	"github.com/langoai/lango/internal/p2p/team"
	"github.com/langoai/lango/internal/payment"
	"github.com/langoai/lango/internal/provenance"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/skill"
	sa "github.com/langoai/lango/internal/smartaccount"
	"github.com/langoai/lango/internal/toolcatalog"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/langoai/lango/internal/tooloutput"
	"github.com/langoai/lango/internal/turnrunner"
	"github.com/langoai/lango/internal/turntrace"
	"github.com/langoai/lango/internal/wallet"
	"github.com/langoai/lango/internal/workflow"
	x402pkg "github.com/langoai/lango/internal/x402"
)

// App is the root application structure
type App struct {
	Config *config.Config

	// Core Components
	Agent   *adk.Agent
	Gateway *gateway.Server
	Store   session.Store

	// Browser (optional, io.Closer)
	Browser io.Closer

	// Security Components (optional)
	Crypto  security.CryptoProvider
	Keys    *security.KeyRegistry
	Secrets *security.SecretsStore

	// Approval Provider (composite, routes to channel-specific providers)
	ApprovalProvider approval.Provider
	GrantStore       *approval.GrantStore
	ApprovalHistory  *approval.HistoryStore

	// Self-Learning Components
	KnowledgeStore            *knowledge.Store
	LearningEngine            *learning.Engine
	LearningSuggestionEmitter *learning.SuggestionEmitter
	SkillRegistry             *skill.Registry

	// Agent Memory Components (optional, per-agent persistent memory)
	AgentMemoryStore agentmemory.Store

	// Observational Memory Components (optional)
	MemoryStore  *memory.Store
	MemoryBuffer *memory.Buffer

	// Embedding / RAG Components (optional)
	EmbeddingBuffer *embedding.EmbeddingBuffer
	RAGService      *embedding.RAGService

	// Conversation Analysis Components (optional)
	AnalysisBuffer *learning.AnalysisBuffer

	// Compaction Components (optional, Phase 3 background hygiene)
	CompactionBuffer *session.CompactionBuffer

	// ExtensionRegistry holds the set of installed extension packs
	// discovered at startup. nil when the subsystem is disabled or the
	// extensions directory does not exist.
	ExtensionRegistry *extension.Registry

	// recallIndex is the FTS5-backed session recall index. nil when disabled
	// or when FTS5 is unavailable in the build.
	recallIndex *session.RecallIndex

	// compactionSync is the indirect sync-point handle installed on the
	// context adapter at build time. wireCompactionBuffer plugs the real
	// buffer into it once it exists.
	compactionSync *compactionSyncHolder

	// Proactive Librarian Components (optional)
	LibrarianInquiryStore    *librarian.InquiryStore
	LibrarianProactiveBuffer *librarian.ProactiveBuffer

	// Graph Components (optional)
	GraphStore  graph.Store
	GraphBuffer *graph.GraphBuffer

	// Payment Components (optional)
	WalletProvider  wallet.WalletProvider
	PaymentService  *payment.Service
	X402Interceptor *x402pkg.Interceptor

	// Cron Scheduling Components (optional)
	CronScheduler *cronpkg.Scheduler

	// Background Task Components (optional)
	BackgroundManager *background.Manager

	// Workflow Engine Components (optional)
	WorkflowEngine *workflow.Engine

	// Economy Components (optional, P2P economy layer)
	EconomyBudget      *budget.Engine
	EconomyRisk        *risk.Engine
	EconomyPricing     *pricing.Engine
	EconomyNegotiation *negotiation.Engine
	EconomyEscrow      *escrow.Engine

	// Smart Account Components (optional, ERC-7579 modular accounts)
	SmartAccountManager    sa.AccountManager
	SmartAccountComponents *smartAccountComponents // full components for CLI access

	// Output Store (compressed tool output retrieval)
	OutputStore *tooloutput.OutputStore

	// Gatekeeper (response sanitizer)
	Sanitizer *gatekeeper.Sanitizer

	// Turn Runtime (shared execution + durable traces)
	TurnRunner    *turnrunner.Runner
	TurnTraceStore turntrace.Store

	// RunLedger Components (optional, Task OS durable execution)
	RunLedgerStore runledger.RunLedgerStore
	RunLedgerPEV   *runledger.PEVEngine

	// Provenance Components (optional)
	ProvenanceCheckpoints *provenance.CheckpointService
	ProvenanceSessionTree *provenance.SessionTree
	ProvenanceAttribution *provenance.AttributionService
	ProvenanceBundle      *provenance.BundleService

	// MCP Components (optional, external MCP server integration)
	MCPManager *mcp.ServerManager

	// Observability Components (optional)
	MetricsCollector *observability.MetricsCollector
	HealthRegistry   *health.Registry
	TokenStore       *token.EntTokenStore
	TracerShutdown   func(context.Context) error

	// Tool Catalog (built-in tool discovery + dynamic dispatch)
	ToolCatalog *toolcatalog.Catalog

	// P2P Components (optional)
	P2PNode            *p2p.Node
	P2PAgentPool       *agentpool.Pool
	P2PTeamCoordinator *team.Coordinator
	P2PAgentProvider   agentpool.DynamicAgentProvider

	// Event Bus (app-level, for hooks and cross-component communication)
	EventBus *eventbus.Bus

	// Agent Registry (dynamic agent definitions)
	AgentRegistry *agentregistry.Registry

	// Hook Registry (tool execution hooks)
	HookRegistry *toolchain.HookRegistry

	// FeatureStatuses holds aggregated init diagnostics for context subsystems.
	FeatureStatuses *StatusCollector

	// Channels
	Channels []Channel

	// Lifecycle registry manages component startup/shutdown ordering.
	registry *lifecycle.Registry

	// ctx/cancel for signalling shutdown to fire-and-forget goroutines.
	ctx    context.Context
	cancel context.CancelFunc

	// wg tracks background goroutines for graceful shutdown
	wg sync.WaitGroup
}

// Channel represents a communication channel (Telegram, Discord, Slack)
type Channel interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// AppMode determines the application operating mode.
type AppMode int

const (
	// AppModeServer is the default mode — starts all components including
	// gateway, channels, automation, and network.
	AppModeServer AppMode = iota

	// AppModeLocalChat starts only core components (Infra, Core, Buffer)
	// and skips network/automation/gateway/channel lifecycle. Used for
	// interactive TUI chat.
	AppModeLocalChat

	// AppModeCockpit starts core + buffer components and optionally initializes
	// channels (if configured), but skips the HTTP gateway and automation lifecycle.
	// Channel Start/Stop is managed externally by the caller (e.g., runCockpit).
	AppModeCockpit
)

// AppOption configures optional behavior for App construction.
type AppOption func(*appOptions)

type appOptions struct {
	mode AppMode
}

// WithLocalChat creates an App in local-chat mode. Network, automation,
// gateway, and channel lifecycle components are not started.
func WithLocalChat() AppOption {
	return func(o *appOptions) { o.mode = AppModeLocalChat }
}

// WithCockpit creates an App in cockpit mode. Core components start normally,
// channels are initialized if configured, but gateway and automation are skipped.
func WithCockpit() AppOption {
	return func(o *appOptions) { o.mode = AppModeCockpit }
}
