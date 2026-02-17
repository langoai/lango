package librarian

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/langowarny/lango/internal/graph"
	"github.com/langowarny/lango/internal/knowledge"
	"github.com/langowarny/lango/internal/memory"
	"github.com/langowarny/lango/internal/session"
)

// MessageProvider retrieves messages for a session key.
type MessageProvider func(sessionKey string) ([]session.Message, error)

// ObservationProvider retrieves recent observations for a session key.
type ObservationProvider func(ctx context.Context, sessionKey string) ([]memory.Observation, error)

// ProactiveBuffer manages the async proactive librarian pipeline.
type ProactiveBuffer struct {
	analyzer           *ObservationAnalyzer
	processor          *InquiryProcessor
	inquiryStore       *InquiryStore
	knowledgeStore     *knowledge.Store
	getMessages        MessageProvider
	getObservations    ObservationProvider
	observationThreshold int
	cooldownTurns      int
	maxPending         int
	autoSaveConfidence string
	graphCallback      GraphCallback

	mu          sync.Mutex
	turnCounter map[string]int // session_key → turns since last inquiry

	queue  chan string
	stopCh chan struct{}
	done   chan struct{}
	logger *zap.SugaredLogger
}

// ProactiveBufferConfig holds configuration for the proactive buffer.
type ProactiveBufferConfig struct {
	ObservationThreshold int
	CooldownTurns        int
	MaxPending           int
	AutoSaveConfidence   string
}

// NewProactiveBuffer creates a new proactive librarian buffer.
func NewProactiveBuffer(
	analyzer *ObservationAnalyzer,
	processor *InquiryProcessor,
	inquiryStore *InquiryStore,
	knowledgeStore *knowledge.Store,
	getMessages MessageProvider,
	getObservations ObservationProvider,
	cfg ProactiveBufferConfig,
	logger *zap.SugaredLogger,
) *ProactiveBuffer {
	if cfg.ObservationThreshold <= 0 {
		cfg.ObservationThreshold = 2
	}
	if cfg.CooldownTurns <= 0 {
		cfg.CooldownTurns = 3
	}
	if cfg.MaxPending <= 0 {
		cfg.MaxPending = 2
	}
	if cfg.AutoSaveConfidence == "" {
		cfg.AutoSaveConfidence = "high"
	}

	return &ProactiveBuffer{
		analyzer:             analyzer,
		processor:            processor,
		inquiryStore:         inquiryStore,
		knowledgeStore:       knowledgeStore,
		getMessages:          getMessages,
		getObservations:      getObservations,
		observationThreshold: cfg.ObservationThreshold,
		cooldownTurns:        cfg.CooldownTurns,
		maxPending:           cfg.MaxPending,
		autoSaveConfidence:   cfg.AutoSaveConfidence,
		turnCounter:          make(map[string]int),
		queue:                make(chan string, 32),
		stopCh:               make(chan struct{}),
		done:                 make(chan struct{}),
		logger:               logger,
	}
}

// SetGraphCallback sets the optional graph triple callback.
func (b *ProactiveBuffer) SetGraphCallback(cb GraphCallback) {
	b.graphCallback = cb
}

// Start launches the background processing goroutine.
func (b *ProactiveBuffer) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(b.done)
		b.run()
	}()
}

// Trigger enqueues a session for proactive analysis.
func (b *ProactiveBuffer) Trigger(sessionKey string) {
	select {
	case b.queue <- sessionKey:
	default:
		b.logger.Warnw("proactive buffer queue full, dropping trigger", "sessionKey", sessionKey)
	}
}

// Stop signals the background goroutine to drain and exit.
func (b *ProactiveBuffer) Stop() {
	close(b.stopCh)
	<-b.done
}

func (b *ProactiveBuffer) run() {
	for {
		select {
		case sessionKey := <-b.queue:
			b.process(sessionKey)
		case <-b.stopCh:
			for {
				select {
				case sessionKey := <-b.queue:
					b.process(sessionKey)
				default:
					return
				}
			}
		}
	}
}

func (b *ProactiveBuffer) process(sessionKey string) {
	ctx := context.Background()

	// Phase 1: Process pending inquiry answers.
	messages, err := b.getMessages(sessionKey)
	if err != nil {
		b.logger.Errorw("get messages for librarian", "sessionKey", sessionKey, "error", err)
		return
	}

	if len(messages) > 0 {
		if err := b.processor.ProcessAnswers(ctx, sessionKey, messages); err != nil {
			b.logger.Errorw("process inquiry answers", "sessionKey", sessionKey, "error", err)
		}
	}

	// Phase 2: Analyze new observations.
	observations, err := b.getObservations(ctx, sessionKey)
	if err != nil {
		b.logger.Errorw("get observations for librarian", "sessionKey", sessionKey, "error", err)
		return
	}

	if len(observations) < b.observationThreshold {
		return
	}

	output, err := b.analyzer.Analyze(ctx, observations)
	if err != nil {
		b.logger.Errorw("analyze observations", "sessionKey", sessionKey, "error", err)
		return
	}

	// Process extractions: auto-save high confidence, create inquiries for medium.
	for _, ext := range output.Extractions {
		if b.shouldAutoSave(ext.Confidence) {
			entry := knowledge.KnowledgeEntry{
				Key:      ext.Key,
				Category: mapCategory(ext.Type),
				Content:  ext.Content,
				Source:   "proactive_librarian",
			}
			if err := b.knowledgeStore.SaveKnowledge(ctx, sessionKey, entry); err != nil {
				b.logger.Warnw("auto-save knowledge", "key", ext.Key, "error", err)
			} else {
				b.logger.Infow("knowledge auto-saved", "key", ext.Key, "confidence", ext.Confidence)
			}

			// Graph callback for extracted triples.
			if b.graphCallback != nil && ext.Subject != "" && ext.Predicate != "" && ext.Object != "" {
				b.graphCallback([]Triple{{
					Subject:   ext.Subject,
					Predicate: ext.Predicate,
					Object:    ext.Object,
					Metadata:  map[string]string{"source": "proactive_librarian", "key": ext.Key},
				}})
			}
		}
	}

	// Process gaps: create inquiries with cooldown/limit checks.
	b.mu.Lock()
	b.turnCounter[sessionKey]++
	turnsSinceLastInquiry := b.turnCounter[sessionKey]
	b.mu.Unlock()

	if turnsSinceLastInquiry < b.cooldownTurns {
		return
	}

	pendingCount, err := b.inquiryStore.CountPendingBySession(ctx, sessionKey)
	if err != nil {
		b.logger.Warnw("count pending inquiries", "sessionKey", sessionKey, "error", err)
		return
	}

	for _, gap := range output.Gaps {
		if pendingCount >= b.maxPending {
			break
		}

		inq := Inquiry{
			SessionKey: sessionKey,
			Topic:      gap.Topic,
			Question:   gap.Question,
			Context:    gap.Context,
			Priority:   gap.Priority,
		}
		if err := b.inquiryStore.SaveInquiry(ctx, inq); err != nil {
			b.logger.Warnw("save inquiry", "topic", gap.Topic, "error", err)
			continue
		}

		pendingCount++
		b.logger.Infow("inquiry created", "topic", gap.Topic, "priority", gap.Priority)
	}

	// Reset cooldown counter after creating inquiries.
	if pendingCount > 0 {
		b.mu.Lock()
		b.turnCounter[sessionKey] = 0
		b.mu.Unlock()
	}
}

// shouldAutoSave checks if the extraction confidence meets the auto-save threshold.
func (b *ProactiveBuffer) shouldAutoSave(confidence string) bool {
	switch b.autoSaveConfidence {
	case "low":
		return true
	case "medium":
		return confidence == "medium" || confidence == "high"
	default: // "high"
		return confidence == "high"
	}
}

// mapCategory maps LLM analysis type to a valid knowledge category.
func mapCategory(analysisType string) string {
	switch analysisType {
	case "preference":
		return "preference"
	case "fact":
		return "fact"
	case "rule":
		return "rule"
	case "definition":
		return "definition"
	default:
		return "fact"
	}
}

// toGraphTriples converts librarian triples to graph.Triple for callback.
func toGraphTriples(triples []Triple) []graph.Triple {
	result := make([]graph.Triple, len(triples))
	for i, t := range triples {
		result[i] = graph.Triple{
			Subject:   t.Subject,
			Predicate: t.Predicate,
			Object:    t.Object,
			Metadata:  t.Metadata,
		}
	}
	return result
}
