package app

import (
	"context"
	"io"
	"sync"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/config"
	cronpkg "github.com/langoai/lango/internal/cron"
	"github.com/langoai/lango/internal/embedding"
	"github.com/langoai/lango/internal/gateway"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/learning"
	"github.com/langoai/lango/internal/librarian"
	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/p2p"
	"github.com/langoai/lango/internal/payment"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/skill"
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

	// P2P Components (optional)
	P2PNode *p2p.Node

	// Channels
	Channels []Channel

	// wg tracks background goroutines for graceful shutdown
	wg sync.WaitGroup
}

// Channel represents a communication channel (Telegram, Discord, Slack)
type Channel interface {
	Start(ctx context.Context) error
	Stop()
}
