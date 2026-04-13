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
	eventbus.SubscribeTyped[eventbus.PolicyDecisionEvent](bus, r.handlePolicyDecision)
	eventbus.SubscribeTyped[eventbus.AlertEvent](bus, r.handleAlert)
	eventbus.SubscribeTyped[eventbus.SandboxDecisionEvent](bus, r.handleSandboxDecision)
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

func (r *Recorder) handlePolicyDecision(evt eventbus.PolicyDecisionEvent) {
	details := map[string]interface{}{
		"verdict":   evt.Verdict,
		"reason":    evt.Reason,
		"unwrapped": evt.Unwrapped,
	}
	if evt.Message != "" {
		details["message"] = evt.Message
	}

	actor := evt.AgentName
	if actor == "" {
		actor = "system"
	}

	_, _ = r.client.AuditLog.Create().
		SetSessionKey(evt.SessionKey).
		SetAction(auditlog.ActionPolicyDecision).
		SetActor(actor).
		SetTarget(evt.Command).
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

func (r *Recorder) handleSandboxDecision(evt eventbus.SandboxDecisionEvent) {
	details := map[string]interface{}{
		"decision": evt.Decision,
		"source":   evt.Source,
		"backend":  evt.Backend,
	}
	if evt.Reason != "" {
		details["reason"] = evt.Reason
	}
	if evt.Pattern != "" {
		details["pattern"] = evt.Pattern
	}

	create := r.client.AuditLog.Create().
		SetAction(auditlog.ActionSandboxDecision).
		SetActor(evt.Source).
		SetTarget(evt.Command).
		SetDetails(details)
	if evt.SessionKey != "" {
		create = create.SetSessionKey(evt.SessionKey)
	}
	_, _ = create.Save(context.Background())
}

func (r *Recorder) handleAlert(evt eventbus.AlertEvent) {
	details := map[string]interface{}{
		"severity":  evt.Severity,
		"message":   evt.Message,
		"timestamp": evt.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
	}
	for k, v := range evt.Details {
		details[k] = v
	}

	_, _ = r.client.AuditLog.Create().
		SetSessionKey(evt.SessionKey).
		SetAction(auditlog.Action("alert")).
		SetActor("system").
		SetTarget(evt.Type).
		SetDetails(details).
		Save(context.Background())
}
