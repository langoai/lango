package learning

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	entlearning "github.com/langowarny/lango/internal/ent/learning"
	"github.com/langowarny/lango/internal/graph"
	"github.com/langowarny/lango/internal/knowledge"
	"github.com/langowarny/lango/internal/session"
	"github.com/langowarny/lango/internal/types"
)

const sessionLearnerPrompt = `You are a session analysis assistant. Analyze this complete conversation session and extract high-confidence learnings.

Focus only on insights that are clearly established and would be useful across future sessions:
- Confirmed user preferences (tools, styles, approaches)
- Verified domain knowledge (facts, rules, constraints)
- Established workflows (successful multi-step patterns)
- Important corrections (where the user explicitly corrected behavior)

For each learning, output a JSON object with:
- "type": one of "fact", "pattern", "correction", "preference"
- "category": brief category label
- "content": clear, reusable statement
- "confidence": MUST be "high" (only extract high-confidence learnings)
- "subject": (optional) graph triple subject
- "predicate": (optional) graph triple predicate
- "object": (optional) graph triple object

Output a JSON array. If nothing high-confidence is found, output [].`

// SessionLearner extracts high-confidence knowledge at session end.
type SessionLearner struct {
	generator     TextGenerator
	store         *knowledge.Store
	graphCallback GraphCallback
	logger        *zap.SugaredLogger
}

// NewSessionLearner creates a new session learner.
func NewSessionLearner(
	generator TextGenerator,
	store *knowledge.Store,
	logger *zap.SugaredLogger,
) *SessionLearner {
	return &SessionLearner{
		generator: generator,
		store:     store,
		logger:    logger,
	}
}

// SetGraphCallback sets the optional graph update hook.
func (l *SessionLearner) SetGraphCallback(cb GraphCallback) {
	l.graphCallback = cb
}

// LearnFromSession analyzes a complete session and stores high-confidence results.
func (l *SessionLearner) LearnFromSession(ctx context.Context, sessionKey string, messages []session.Message) error {
	if len(messages) < 4 {
		l.logger.Debugw("session too short for learning", "sessionKey", sessionKey, "turns", len(messages))
		return nil
	}

	sampled := sampleMessages(messages)
	userPrompt := formatMessagesForAnalysis(sampled)

	response, err := l.generator.GenerateText(ctx, sessionLearnerPrompt, userPrompt)
	if err != nil {
		return fmt.Errorf("session learning: %w", err)
	}

	results, err := parseAnalysisResponse(response)
	if err != nil {
		l.logger.Debugw("parse session learning response", "error", err)
		return nil
	}

	stored := 0
	for _, r := range results {
		if r.Content == "" || r.Confidence != types.ConfidenceHigh {
			continue
		}
		l.saveSessionResult(ctx, sessionKey, r)
		stored++
	}

	l.logger.Infow("session learning complete",
		"sessionKey", sessionKey,
		"extracted", len(results),
		"stored", stored,
	)
	return nil
}

func (l *SessionLearner) saveSessionResult(ctx context.Context, sessionKey string, r analysisResult) {
	switch r.Type {
	case "fact", "preference":
		key := fmt.Sprintf("session:%s:%s", sessionKey, sanitizeForNode(r.Content[:min(len(r.Content), 32)]))
		entry := knowledge.KnowledgeEntry{
			Key:      key,
			Category: mapKnowledgeCategory(r.Type),
			Content:  r.Content,
			Source:   "session_learning",
		}
		if err := l.store.SaveKnowledge(ctx, sessionKey, entry); err != nil {
			l.logger.Debugw("save session knowledge", "error", err)
		}

	case "pattern", "correction":
		entry := knowledge.LearningEntry{
			Trigger:   fmt.Sprintf("session:%s", r.Category),
			Diagnosis: r.Content,
			Category:  mapLearningCategory(r.Type),
		}
		if r.Type == "correction" {
			entry.Fix = r.Content
			entry.Category = entlearning.CategoryUserCorrection
		}
		if err := l.store.SaveLearning(ctx, sessionKey, entry); err != nil {
			l.logger.Debugw("save session learning", "error", err)
		}
	}

	// Cross-reference graph triple.
	if l.graphCallback != nil && r.Subject != "" && r.Predicate != "" && r.Object != "" {
		l.graphCallback([]graph.Triple{{
			Subject:   r.Subject,
			Predicate: r.Predicate,
			Object:    r.Object,
		}})
	}
}

// sampleMessages samples messages from long sessions for efficient LLM processing.
// For sessions <= 20 messages, returns all. Otherwise: first 3 + every 5th + last 5.
func sampleMessages(msgs []session.Message) []session.Message {
	if len(msgs) <= 20 {
		return msgs
	}

	sampled := make([]session.Message, 0, 15)

	// First 3
	sampled = append(sampled, msgs[:3]...)

	// Every 5th from index 3 to len-5
	for i := 3; i < len(msgs)-5; i += 5 {
		sampled = append(sampled, msgs[i])
	}

	// Last 5
	sampled = append(sampled, msgs[len(msgs)-5:]...)

	return sampled
}
