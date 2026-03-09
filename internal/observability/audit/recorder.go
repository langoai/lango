// Package audit records events to the existing AuditLog Ent schema.
package audit

import (
	"context"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/auditlog"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/toolchain"
)

// Recorder writes audit log entries to the database.
type Recorder struct {
	client *ent.Client
}

// NewRecorder creates a new audit Recorder.
func NewRecorder(client *ent.Client) *Recorder {
	return &Recorder{client: client}
}

// Subscribe registers the recorder on the event bus.
func (r *Recorder) Subscribe(bus *eventbus.Bus) {
	eventbus.SubscribeTyped[toolchain.ToolExecutedEvent](bus, r.handleToolExecuted)
	eventbus.SubscribeTyped[eventbus.TokenUsageEvent](bus, r.handleTokenUsage)
}

func (r *Recorder) handleToolExecuted(evt toolchain.ToolExecutedEvent) {
	details := map[string]interface{}{
		"duration": evt.Duration.String(),
		"success":  evt.Success,
	}
	if evt.Error != "" {
		details["error"] = evt.Error
	}

	_, _ = r.client.AuditLog.Create().
		SetSessionKey(evt.SessionKey).
		SetAction(auditlog.ActionToolCall).
		SetActor(evt.AgentName).
		SetTarget(evt.ToolName).
		SetDetails(details).
		Save(context.Background())
}

func (r *Recorder) handleTokenUsage(evt eventbus.TokenUsageEvent) {
	details := map[string]interface{}{
		"provider":     evt.Provider,
		"model":        evt.Model,
		"inputTokens":  evt.InputTokens,
		"outputTokens": evt.OutputTokens,
		"totalTokens":  evt.TotalTokens,
	}
	if evt.CacheTokens > 0 {
		details["cacheTokens"] = evt.CacheTokens
	}

	actor := evt.AgentName
	if actor == "" {
		actor = "system"
	}

	_, _ = r.client.AuditLog.Create().
		SetSessionKey(evt.SessionKey).
		SetAction(auditlog.ActionToolCall).
		SetActor(actor).
		SetTarget(evt.Model).
		SetDetails(details).
		Save(context.Background())
}
