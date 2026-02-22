package learning

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/knowledge"
)

// ToolResultObserver observes tool execution results for learning.
// Both Engine and GraphEngine implement this interface.
type ToolResultObserver interface {
	OnToolResult(ctx context.Context, sessionKey, toolName string, params map[string]interface{}, result interface{}, err error)
}

// Compile-time interface check.
var _ ToolResultObserver = (*Engine)(nil)

// Engine observes tool execution results and learns from errors.
type Engine struct {
	store  *knowledge.Store
	logger *zap.SugaredLogger
}

// NewEngine creates a new learning engine.
func NewEngine(store *knowledge.Store, logger *zap.SugaredLogger) *Engine {
	return &Engine{store: store, logger: logger}
}

// OnToolResult observes a tool execution result and records learnings.
func (e *Engine) OnToolResult(ctx context.Context, sessionKey, toolName string, params map[string]interface{}, result interface{}, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}

	details := make(map[string]interface{}, 3)
	details["status"] = status
	if summarized := summarizeParams(params); summarized != nil {
		details["params"] = summarized
	}
	if err != nil {
		details["error"] = err.Error()
	}

	auditErr := e.store.SaveAuditLog(ctx, knowledge.AuditEntry{
		SessionKey: sessionKey,
		Action:     "tool_call",
		Actor:      "agent",
		Target:     toolName,
		Details:    details,
	})
	if auditErr != nil {
		e.logger.Warnw("save audit log:", "error", auditErr)
	}

	if err != nil {
		e.handleError(ctx, sessionKey, toolName, err)
		return
	}

	e.handleSuccess(ctx, toolName)
}

// GetFixForError returns a known fix for a given tool error if one exists with sufficient confidence.
func (e *Engine) GetFixForError(ctx context.Context, toolName string, err error) (string, bool) {
	pattern := extractErrorPattern(err)

	entities, searchErr := e.store.SearchLearningEntities(ctx, pattern, 5)
	if searchErr != nil {
		e.logger.Warnw("search learnings:", "error", searchErr)
		return "", false
	}

	for _, entity := range entities {
		if entity.Confidence > 0.5 && entity.Fix != "" {
			return entity.Fix, true
		}
	}
	return "", false
}

// RecordUserCorrection saves a user-provided correction as a high-confidence learning.
func (e *Engine) RecordUserCorrection(ctx context.Context, sessionKey, trigger, diagnosis, fix string) error {
	return e.store.SaveLearning(ctx, sessionKey, knowledge.LearningEntry{
		Trigger:   trigger,
		Diagnosis: diagnosis,
		Fix:       fix,
		Category:  entlearning.CategoryUserCorrection,
	})
}

func (e *Engine) handleError(ctx context.Context, sessionKey, toolName string, err error) {
	pattern := extractErrorPattern(err)

	entities, searchErr := e.store.SearchLearningEntities(ctx, pattern, 5)
	if searchErr != nil {
		e.logger.Warnw("search learnings:", "error", searchErr)
		return
	}

	for _, entity := range entities {
		if entity.Confidence > 0.5 {
			e.logger.Infow("known fix exists for error",
				"tool", toolName,
				"pattern", pattern,
				"fix", entity.Fix,
			)
			return
		}
	}

	category := categorizeError(toolName, err)
	saveErr := e.store.SaveLearning(ctx, sessionKey, knowledge.LearningEntry{
		Trigger:      fmt.Sprintf("tool:%s", toolName),
		ErrorPattern: pattern,
		Diagnosis:    err.Error(),
		Category:     category,
	})
	if saveErr != nil {
		e.logger.Warnw("save learning",
			"session", sessionKey,
			"tool", toolName,
			"error", saveErr)
	}
}

func (e *Engine) handleSuccess(ctx context.Context, toolName string) {
	entities, searchErr := e.store.SearchLearningEntities(ctx, toolName, 5)
	if searchErr != nil {
		e.logger.Warnw("search learnings:", "error", searchErr)
		return
	}

	for _, entity := range entities {
		if boostErr := e.store.BoostLearningConfidence(ctx, entity.ID, 1, 0.0); boostErr != nil {
			e.logger.Warnw("boost learning confidence:", "error", boostErr)
		}
	}
}
