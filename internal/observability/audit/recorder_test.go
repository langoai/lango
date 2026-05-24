package audit

import (
	"context"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/auditlog"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/toolchain"
	toolpayment "github.com/langoai/lango/internal/tools/payment"
	_ "github.com/mattn/go-sqlite3"
)

func TestNewRecorderKeepsClient(t *testing.T) {
	client := newAuditTestClient(t)

	recorder := NewRecorder(client)

	if recorder == nil {
		t.Fatal("NewRecorder returned nil")
	}
	if recorder.client != client {
		t.Fatal("NewRecorder did not keep the provided client")
	}
}

func TestSubscribeRecordsToolExecutedEvent(t *testing.T) {
	client := newAuditTestClient(t)
	bus := eventbus.New()
	NewRecorder(client).Subscribe(bus)

	bus.Publish(toolchain.ToolExecutedEvent{
		ToolName:   "filesystem.read",
		AgentName:  "navigator",
		SessionKey: "session-tool",
		Duration:   250 * time.Millisecond,
		Success:    false,
		Error:      "permission denied",
	})

	row := onlyAuditLog(t, client)
	assertAuditLog(t, row, auditlog.ActionToolCall, "session-tool", "navigator", "filesystem.read")
	assertDetail(t, row.Details, "duration", "250ms")
	assertDetail(t, row.Details, "success", false)
	assertDetail(t, row.Details, "error", "permission denied")
}

func TestSubscribeRecordsPolicyDecisionEvent(t *testing.T) {
	client := newAuditTestClient(t)
	bus := eventbus.New()
	NewRecorder(client).Subscribe(bus)

	bus.Publish(eventbus.PolicyDecisionEvent{
		Command:    "sqlite3 ~/.lango/lango.db",
		Unwrapped:  "sqlite3 /home/user/.lango/lango.db",
		Verdict:    "block",
		Reason:     "protected_path",
		Message:    "blocked protected database access",
		SessionKey: "session-policy",
	})

	row := onlyAuditLog(t, client)
	assertAuditLog(t, row, auditlog.ActionPolicyDecision, "session-policy", "system", "sqlite3 ~/.lango/lango.db")
	assertDetail(t, row.Details, "verdict", "block")
	assertDetail(t, row.Details, "reason", "protected_path")
	assertDetail(t, row.Details, "unwrapped", "sqlite3 /home/user/.lango/lango.db")
	assertDetail(t, row.Details, "message", "blocked protected database access")
}

func TestSubscribeRecordsTokenUsageEvent(t *testing.T) {
	client := newAuditTestClient(t)
	bus := eventbus.New()
	NewRecorder(client).Subscribe(bus)

	bus.Publish(eventbus.TokenUsageEvent{
		Provider:     "openai",
		Model:        "gpt-test",
		SessionKey:   "session-token",
		AgentName:    "planner",
		InputTokens:  11,
		OutputTokens: 13,
		TotalTokens:  24,
		CacheTokens:  5,
	})

	row := onlyAuditLog(t, client)
	assertAuditLog(t, row, auditlog.ActionToolCall, "session-token", "planner", "gpt-test")
	assertDetail(t, row.Details, "provider", "openai")
	assertDetail(t, row.Details, "model", "gpt-test")
	assertNumericDetail(t, row.Details, "inputTokens", 11)
	assertNumericDetail(t, row.Details, "outputTokens", 13)
	assertNumericDetail(t, row.Details, "totalTokens", 24)
	assertNumericDetail(t, row.Details, "cacheTokens", 5)
}

func TestSubscribeRecordsSandboxDecisionEvent(t *testing.T) {
	client := newAuditTestClient(t)
	bus := eventbus.New()
	NewRecorder(client).Subscribe(bus)

	bus.Publish(eventbus.SandboxDecisionEvent{
		SessionKey: "session-sandbox",
		Source:     "exec",
		Command:    "go test ./...",
		Decision:   "excluded",
		Backend:    "seatbelt",
		Reason:     "trusted command",
		Pattern:    "go test *",
	})

	row := onlyAuditLog(t, client)
	assertAuditLog(t, row, auditlog.ActionSandboxDecision, "session-sandbox", "exec", "go test ./...")
	assertDetail(t, row.Details, "decision", "excluded")
	assertDetail(t, row.Details, "source", "exec")
	assertDetail(t, row.Details, "backend", "seatbelt")
	assertDetail(t, row.Details, "reason", "trusted command")
	assertDetail(t, row.Details, "pattern", "go test *")
}

func TestSubscribeRecordsAlertEvent(t *testing.T) {
	client := newAuditTestClient(t)
	bus := eventbus.New()
	NewRecorder(client).Subscribe(bus)
	timestamp := time.Date(2026, 5, 20, 12, 34, 56, 0, time.UTC)

	bus.Publish(eventbus.AlertEvent{
		Type:       "policy_block_rate",
		Severity:   "critical",
		Message:    "block rate exceeded",
		SessionKey: "session-alert",
		Timestamp:  timestamp,
		Details: map[string]interface{}{
			"blocked": 3,
			"window":  "1m",
		},
	})

	row := onlyAuditLog(t, client)
	assertAuditLog(t, row, auditlog.ActionAlert, "session-alert", "system", "policy_block_rate")
	assertDetail(t, row.Details, "severity", "critical")
	assertDetail(t, row.Details, "message", "block rate exceeded")
	assertDetail(t, row.Details, "timestamp", "2026-05-20T12:34:56Z")
	assertNumericDetail(t, row.Details, "blocked", 3)
	assertDetail(t, row.Details, "window", "1m")
}

func TestRecordPaymentExecutionNilRecorderIsNoop(t *testing.T) {
	var recorder *Recorder

	err := recorder.RecordPaymentExecution(context.Background(), toolpayment.PaymentExecutionAuditEntry{
		ToolName:   "payment.execute",
		SessionKey: "session-payment",
		Outcome:    "denied",
	})

	if err != nil {
		t.Fatalf("RecordPaymentExecution on nil recorder returned error: %v", err)
	}
}

func TestRecordPaymentExecutionRecordsPaymentAuditLog(t *testing.T) {
	client := newAuditTestClient(t)
	recorder := NewRecorder(client)

	err := recorder.RecordPaymentExecution(context.Background(), toolpayment.PaymentExecutionAuditEntry{
		ToolName:             "payment.execute",
		SessionKey:           "session-payment",
		TransactionReceiptID: "tx-123",
		SubmissionReceiptID:  "submit-456",
		Outcome:              "denied",
		Reason:               "missing receipt",
	})
	if err != nil {
		t.Fatalf("RecordPaymentExecution returned error: %v", err)
	}

	row := onlyAuditLog(t, client)
	assertAuditLog(t, row, auditlog.ActionPaymentExecution, "session-payment", "agent", "payment.execute")
	assertDetail(t, row.Details, "toolName", "payment.execute")
	assertDetail(t, row.Details, "transactionReceiptId", "tx-123")
	assertDetail(t, row.Details, "submissionReceiptId", "submit-456")
	assertDetail(t, row.Details, "outcome", "denied")
	assertDetail(t, row.Details, "reason", "missing receipt")
}

func newAuditTestClient(t *testing.T) *ent.Client {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	client := enttest.Open(t, "sqlite3", "file:"+name+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Fatalf("close ent client: %v", err)
		}
	})
	return client
}

func onlyAuditLog(t *testing.T, client *ent.Client) *ent.AuditLog {
	t.Helper()

	rows, err := client.AuditLog.Query().All(context.Background())
	if err != nil {
		t.Fatalf("query audit logs: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("audit log count = %d, want 1", len(rows))
	}
	return rows[0]
}

func assertAuditLog(t *testing.T, row *ent.AuditLog, action auditlog.Action, sessionKey, actor, target string) {
	t.Helper()

	if row.Action != action {
		t.Fatalf("action = %q, want %q", row.Action, action)
	}
	if row.SessionKey != sessionKey {
		t.Fatalf("session key = %q, want %q", row.SessionKey, sessionKey)
	}
	if row.Actor != actor {
		t.Fatalf("actor = %q, want %q", row.Actor, actor)
	}
	if row.Target != target {
		t.Fatalf("target = %q, want %q", row.Target, target)
	}
}

func assertDetail(t *testing.T, details map[string]interface{}, key string, want interface{}) {
	t.Helper()

	got, ok := details[key]
	if !ok {
		t.Fatalf("details[%q] missing in %#v", key, details)
	}
	if got != want {
		t.Fatalf("details[%q] = %#v, want %#v", key, got, want)
	}
}

func assertNumericDetail(t *testing.T, details map[string]interface{}, key string, want float64) {
	t.Helper()

	got, ok := details[key]
	if !ok {
		t.Fatalf("details[%q] missing in %#v", key, details)
	}

	var gotFloat float64
	switch v := got.(type) {
	case int:
		gotFloat = float64(v)
	case int64:
		gotFloat = float64(v)
	case float64:
		gotFloat = v
	default:
		t.Fatalf("details[%q] = %T(%#v), want numeric %v", key, got, got, want)
	}
	if math.Abs(gotFloat-want) > 0.000001 {
		t.Fatalf("details[%q] = %v, want %v", key, gotFloat, want)
	}
}

func TestRecorderSubscribeDoesNotRecordUnsubscribedBus(t *testing.T) {
	client := newAuditTestClient(t)
	bus := eventbus.New()
	_ = NewRecorder(client)

	bus.Publish(toolchain.ToolExecutedEvent{
		ToolName:   "filesystem.read",
		AgentName:  "navigator",
		SessionKey: "session-tool",
		Success:    true,
	})

	count, err := client.AuditLog.Query().Count(context.Background())
	if err != nil {
		t.Fatalf("count audit logs: %v", err)
	}
	if count != 0 {
		t.Fatalf("audit log count = %d, want 0", count)
	}
}

func ExampleRecorder_RecordPaymentExecution_nilReceiver() {
	var recorder *Recorder
	err := recorder.RecordPaymentExecution(context.Background(), toolpayment.PaymentExecutionAuditEntry{})
	fmt.Println(err == nil)
	// Output: true
}
