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
	"github.com/langoai/lango/internal/embedding"
	"github.com/langoai/lango/internal/eventbus"
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
	"github.com/langoai/lango/internal/toolcatalog"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/langoai/lango/internal/tooloutput"
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

	// Self-Learning Components
	KnowledgeStore *knowledge.Store
	LearningEngine *learning.Engine
	SkillRegistry  *skill.Registry

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
	EconomyBudget      interface{} // *budget.Engine
	EconomyRisk        interface{} // *risk.Engine
	EconomyPricing     interface{} // *pricing.Engine
	EconomyNegotiation interface{} // *negotiation.Engine
	EconomyEscrow      interface{} // *escrow.Engine

	// Smart Account Components (optional, ERC-7579 modular accounts)
	SmartAccountManager    interface{}             // *smartaccount.Manager
	SmartAccountComponents *smartAccountComponents // full components for CLI access

	// Output Store (compressed tool output retrieval)
	OutputStore *tooloutput.OutputStore

	// Gatekeeper (response sanitizer)
	Sanitizer *gatekeeper.Sanitizer

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
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
