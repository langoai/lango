package librarian

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/asyncbuf"
	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
)

// MessageProvider retrieves messages for a session key.
type MessageProvider func(sessionKey string) ([]session.Message, error)

// ObservationProvider retrieves recent observations for a session key.
type ObservationProvider func(ctx context.Context, sessionKey string) ([]memory.Observation, error)

// ProactiveBuffer manages the async proactive librarian pipeline.
type ProactiveBuffer struct {
	analyzer             *ObservationAnalyzer
	processor            *InquiryProcessor
	inquiryStore         *InquiryStore
	knowledgeStore       *knowledge.Store
	getMessages          MessageProvider
	getObservations      ObservationProvider
	observationThreshold int
	cooldownTurns        int
	maxPending           int
	autoSaveConfidence   types.Confidence
	bus                  *eventbus.Bus // Optional event bus for publishing triple events.

	mu          sync.Mutex
	turnCounter map[string]int // session_key -> turns since last inquiry

	inner  *asyncbuf.TriggerBuffer[string]
	logger *zap.SugaredLogger
}

// ProactiveBufferConfig holds configuration for the proactive buffer.
type ProactiveBufferConfig struct {
	ObservationThreshold int
	CooldownTurns        int
	MaxPending           int
	AutoSaveConfidence   types.Confidence
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
		cfg.AutoSaveConfidence = types.ConfidenceHigh
	}

	b := &ProactiveBuffer{
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
		logger:               logger,
	}
	b.inner = asyncbuf.NewTriggerBuffer[string](asyncbuf.TriggerConfig{
		QueueSize: 32,
	}, b.process, logger)
	return b
}

// SetEventBus sets the optional event bus for publishing triple events.
func (b *ProactiveBuffer) SetEventBus(bus *eventbus.Bus) {
	b.bus = bus
}

// Start launches the background processing goroutine.
func (b *ProactiveBuffer) Start(wg *sync.WaitGroup) {
	b.inner.Start(wg)
}

// Trigger enqueues a session for proactive analysis.
func (b *ProactiveBuffer) Trigger(sessionKey string) {
	b.inner.Enqueue(sessionKey)
}

// Stop signals the background goroutine to drain and exit.
func (b *ProactiveBuffer) Stop() {
	b.inner.Stop()
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
			cat, err := mapCategory(ext.Type)
			if err != nil {
				b.logger.Warnw("skip extraction: unknown type", "key", ext.Key, "type", ext.Type, "error", err)
				continue
			}
			entry := knowledge.KnowledgeEntry{
				Key:      ext.Key,
				Category: cat,
				Content:  ext.Content,
				Source:   "proactive_librarian",
			}
			if ext.Temporal != "" {
				entry.Tags = append(entry.Tags, "temporal:"+ext.Temporal)
			}
			if err := b.knowledgeStore.SaveKnowledge(ctx, sessionKey, entry); err != nil {
				b.logger.Warnw("auto-save knowledge", "key", ext.Key, "error", err)
			} else {
				b.logger.Infow("knowledge auto-saved", "key", ext.Key, "confidence", ext.Confidence)
			}

			// Dual-save: pattern/correction also go to learning store.
			dualSaveToLearning(ctx, b.knowledgeStore, sessionKey,
				ext.Type, ext.Key, ext.Content, "proactive:", b.logger)

			// Publish graph triples via event bus.
			if b.bus != nil && ext.Subject != "" && ext.Predicate != "" && ext.Object != "" {
				b.bus.Publish(eventbus.TriplesExtractedEvent{
					Triples: []eventbus.Triple{{
						Subject:   ext.Subject,
						Predicate: ext.Predicate,
						Object:    ext.Object,
					}},
					Source: "proactive_librarian",
				})
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
func (b *ProactiveBuffer) shouldAutoSave(confidence types.Confidence) bool {
	switch b.autoSaveConfidence {
	case types.ConfidenceLow:
		return true
	case types.ConfidenceMedium:
		return confidence == types.ConfidenceMedium || confidence == types.ConfidenceHigh
	default: // "high"
		return confidence == types.ConfidenceHigh
	}
}

// mapCategory maps LLM analysis type to a valid knowledge category.
func mapCategory(analysisType string) (entknowledge.Category, error) {
	switch analysisType {
	case "preference":
		return entknowledge.CategoryPreference, nil
	case "fact":
		return entknowledge.CategoryFact, nil
	case "rule":
		return entknowledge.CategoryRule, nil
	case "definition":
		return entknowledge.CategoryDefinition, nil
	case "pattern":
		return entknowledge.CategoryPattern, nil
	case "correction":
		return entknowledge.CategoryCorrection, nil
	default:
		return "", fmt.Errorf("unrecognized knowledge type: %q", analysisType)
	}
}

// dualSaveToLearning saves a pattern or correction knowledge entry to the learning store.
// It is a no-op for other categories.
func dualSaveToLearning(
	ctx context.Context,
	store *knowledge.Store,
	sessionKey string,
	category string,
	key string,
	content string,
	triggerPrefix string,
	logger *zap.SugaredLogger,
) {
	if category != "pattern" && category != "correction" {
		return
	}
	lEntry := knowledge.LearningEntry{
		Trigger:   triggerPrefix + key,
		Diagnosis: content,
		Category:  entlearning.CategoryGeneral,
	}
	if category == "correction" {
		lEntry.Fix = content
		lEntry.Category = entlearning.CategoryUserCorrection
	}
	if err := store.SaveLearning(ctx, sessionKey, lEntry); err != nil {
		logger.Warnw("dual-save learning", "key", key, "error", err)
	}
}
