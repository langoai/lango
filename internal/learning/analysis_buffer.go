package learning

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/asyncbuf"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
)

// perMessageOverhead is the token overhead per message for role/formatting.
const perMessageOverhead = 4

// MessageProvider retrieves messages for a session key.
type MessageProvider func(sessionKey string) ([]session.Message, error)

// AnalysisRequest represents a request to analyze a session's conversation.
type AnalysisRequest struct {
	SessionKey string
	SessionEnd bool // true for session-end analysis
}

// AnalysisBuffer manages async conversation analysis processing.
type AnalysisBuffer struct {
	analyzer       *ConversationAnalyzer
	learner        *SessionLearner
	getMessages    MessageProvider
	turnThreshold  int
	tokenThreshold int

	mu           sync.Mutex
	lastAnalyzed map[string]int // session_key -> last analyzed message index

	inner  *asyncbuf.TriggerBuffer[AnalysisRequest]
	logger *zap.SugaredLogger
}

// NewAnalysisBuffer creates a new async analysis buffer.
func NewAnalysisBuffer(
	analyzer *ConversationAnalyzer,
	learner *SessionLearner,
	getMessages MessageProvider,
	turnThreshold, tokenThreshold int,
	logger *zap.SugaredLogger,
) *AnalysisBuffer {
	if turnThreshold <= 0 {
		turnThreshold = 10
	}
	if tokenThreshold <= 0 {
		tokenThreshold = 2000
	}
	b := &AnalysisBuffer{
		analyzer:       analyzer,
		learner:        learner,
		getMessages:    getMessages,
		turnThreshold:  turnThreshold,
		tokenThreshold: tokenThreshold,
		lastAnalyzed:   make(map[string]int),
		logger:         logger,
	}
	b.inner = asyncbuf.NewTriggerBuffer[AnalysisRequest](asyncbuf.TriggerConfig{
		QueueSize: 32,
	}, b.process, logger)
	return b
}

// Start launches the background analysis goroutine.
func (b *AnalysisBuffer) Start(wg *sync.WaitGroup) {
	b.inner.Start(wg)
}

// Trigger checks if analysis is needed for a session and enqueues if thresholds are met.
func (b *AnalysisBuffer) Trigger(sessionKey string) {
	b.inner.Enqueue(AnalysisRequest{SessionKey: sessionKey})
}

// TriggerSessionEnd enqueues a session-end analysis request.
func (b *AnalysisBuffer) TriggerSessionEnd(sessionKey string) {
	b.inner.Enqueue(AnalysisRequest{SessionKey: sessionKey, SessionEnd: true})
}

// Stop signals the background goroutine to drain and exit.
func (b *AnalysisBuffer) Stop() {
	b.inner.Stop()
}

func (b *AnalysisBuffer) process(req AnalysisRequest) {
	ctx := context.Background()

	messages, err := b.getMessages(req.SessionKey)
	if err != nil {
		b.logger.Errorw("get messages for analysis", "sessionKey", req.SessionKey, "error", err)
		return
	}

	if req.SessionEnd {
		if err := b.learner.LearnFromSession(ctx, req.SessionKey, messages); err != nil {
			b.logger.Errorw("session learning failed", "sessionKey", req.SessionKey, "error", err)
		}
		return
	}

	// Check thresholds for conversation analysis.
	lastIdx := b.getLastAnalyzed(req.SessionKey)
	if lastIdx+1 >= len(messages) {
		return
	}

	unanalyzed := messages[lastIdx+1:]
	turnCount := len(unanalyzed)

	var tokenCount int
	for _, msg := range unanalyzed {
		tokenCount += perMessageOverhead + types.EstimateTokens(msg.Content)
	}

	if turnCount < b.turnThreshold && tokenCount < b.tokenThreshold {
		return
	}

	if err := b.analyzer.Analyze(ctx, req.SessionKey, unanalyzed); err != nil {
		b.logger.Errorw("conversation analysis failed", "sessionKey", req.SessionKey, "error", err)
		return
	}

	b.setLastAnalyzed(req.SessionKey, len(messages)-1)
}

func (b *AnalysisBuffer) getLastAnalyzed(sessionKey string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	if idx, ok := b.lastAnalyzed[sessionKey]; ok {
		return idx
	}
	b.lastAnalyzed[sessionKey] = -1
	return -1
}

func (b *AnalysisBuffer) setLastAnalyzed(sessionKey string, idx int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lastAnalyzed[sessionKey] = idx
}
