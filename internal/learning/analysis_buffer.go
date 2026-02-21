package learning

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/langowarny/lango/internal/session"
	"github.com/langowarny/lango/internal/types"
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
	lastAnalyzed map[string]int // session_key → last analyzed message index

	queue  chan AnalysisRequest
	stopCh chan struct{}
	done   chan struct{}
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
	return &AnalysisBuffer{
		analyzer:       analyzer,
		learner:        learner,
		getMessages:    getMessages,
		turnThreshold:  turnThreshold,
		tokenThreshold: tokenThreshold,
		lastAnalyzed:   make(map[string]int),
		queue:          make(chan AnalysisRequest, 32),
		stopCh:         make(chan struct{}),
		done:           make(chan struct{}),
		logger:         logger,
	}
}

// Start launches the background analysis goroutine.
func (b *AnalysisBuffer) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(b.done)
		b.run()
	}()
}

// Trigger checks if analysis is needed for a session and enqueues if thresholds are met.
func (b *AnalysisBuffer) Trigger(sessionKey string) {
	select {
	case b.queue <- AnalysisRequest{SessionKey: sessionKey}:
	default:
		b.logger.Warnw("analysis queue full, dropping trigger", "sessionKey", sessionKey)
	}
}

// TriggerSessionEnd enqueues a session-end analysis request.
func (b *AnalysisBuffer) TriggerSessionEnd(sessionKey string) {
	select {
	case b.queue <- AnalysisRequest{SessionKey: sessionKey, SessionEnd: true}:
	default:
		b.logger.Warnw("analysis queue full, dropping session-end trigger", "sessionKey", sessionKey)
	}
}

// Stop signals the background goroutine to drain and exit.
func (b *AnalysisBuffer) Stop() {
	close(b.stopCh)
	<-b.done
}

func (b *AnalysisBuffer) run() {
	for {
		select {
		case req := <-b.queue:
			b.process(req)
		case <-b.stopCh:
			// Drain remaining.
			for {
				select {
				case req := <-b.queue:
					b.process(req)
				default:
					return
				}
			}
		}
	}
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
