package app

import (
	"context"
	"io"
	"sync"

	"github.com/langowarny/lango/internal/adk"
	"github.com/langowarny/lango/internal/approval"
	"github.com/langowarny/lango/internal/background"
	"github.com/langowarny/lango/internal/config"
	cronpkg "github.com/langowarny/lango/internal/cron"
	"github.com/langowarny/lango/internal/embedding"
	"github.com/langowarny/lango/internal/gateway"
	"github.com/langowarny/lango/internal/graph"
	"github.com/langowarny/lango/internal/knowledge"
	"github.com/langowarny/lango/internal/learning"
	"github.com/langowarny/lango/internal/librarian"
	"github.com/langowarny/lango/internal/memory"
	"github.com/langowarny/lango/internal/payment"
	"github.com/langowarny/lango/internal/security"
	"github.com/langowarny/lango/internal/session"
	"github.com/langowarny/lango/internal/skill"
	"github.com/langowarny/lango/internal/wallet"
	"github.com/langowarny/lango/internal/workflow"
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
	KnowledgeStore  *knowledge.Store
	LearningEngine  *learning.Engine
	SkillRegistry   *skill.Registry

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
	WalletProvider wallet.WalletProvider
	PaymentService *payment.Service

	// Cron Scheduling Components (optional)
	CronScheduler *cronpkg.Scheduler

	// Background Task Components (optional)
	BackgroundManager *background.Manager

	// Workflow Engine Components (optional)
	WorkflowEngine *workflow.Engine

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
