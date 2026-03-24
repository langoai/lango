package turnrunner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	adksession "google.golang.org/adk/session"
	"google.golang.org/genai"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/deadline"
	"github.com/langoai/lango/internal/logging"
	langosession "github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/tools/browser"
	"github.com/langoai/lango/internal/turntrace"
)

func logger() *zap.SugaredLogger { return logging.App().Named("turnrunner") }

const traceWriteTimeout = 5 * time.Second

// EmptyResponseFallback is used only for truly empty successful turns.
const EmptyResponseFallback = "I processed your message but couldn't formulate a visible response. Could you try rephrasing your question?"

// TurnOutcome is the shared classification for a completed turn.
type TurnOutcome = turntrace.Outcome

// Executor is the subset of adk.Agent used by the runner.
type Executor interface {
	RunStreamingDetailed(
		ctx context.Context,
		sessionID, input string,
		onChunk adk.ChunkCallback,
		opts ...adk.RunOption,
	) (adk.RunReport, error)
}

// Sanitizer is the subset of gatekeeper.Sanitizer needed by the runner.
type Sanitizer interface {
	Enabled() bool
	Sanitize(string) string
}

// TurnCallback fires after a turn finishes.
type TurnCallback func(sessionKey string)

// Config controls runner timeouts and durable trace storage.
type Config struct {
	IdleTimeout time.Duration
	HardCeiling time.Duration
	TraceStore  turntrace.Store

	// DelegationBudgetMax is the delegation count threshold for budget warnings.
	// Zero means use default (15).
	DelegationBudgetMax int
}

// Request is a single turn execution request.
type Request struct {
	SessionKey string
	Input      string
	Entrypoint string
	OnChunk    func(string)
	OnWarning  func(elapsed, hardCeiling time.Duration)

	// OnDelegation is called when a delegation event is observed in the trace.
	// from/to are agent names, reason is optional context.
	OnDelegation func(from, to, reason string)

	// OnBudgetWarning is called when delegation count approaches the threshold.
	OnBudgetWarning func(used, max int)
}

// Result is the structured result of a completed turn.
type Result struct {
	ResponseText string
	Outcome      TurnOutcome
	TraceID      string
	UserMessage  string
	Elapsed      time.Duration
	ErrorCode    string
	CauseClass   string
	CauseDetail  string
	OperatorSummary string
	Summary      string
}

// Runner owns timeout handling, durable tracing, and outcome classification.
type Runner struct {
	executor            Executor
	sessionStore        langosession.Store
	sanitizer           Sanitizer
	traceStore          turntrace.Store
	idleTimeout         time.Duration
	hardCeiling         time.Duration
	delegationBudgetMax int

	mu        sync.RWMutex
	callbacks []TurnCallback
}

// New creates a new turn runner.
func New(
	cfg Config,
	executor Executor,
	sessionStore langosession.Store,
	sanitizer Sanitizer,
) *Runner {
	hardCeiling := cfg.HardCeiling
	if hardCeiling <= 0 {
		hardCeiling = 5 * time.Minute
	}

	delegMax := cfg.DelegationBudgetMax
	if delegMax <= 0 {
		delegMax = 15
	}

	return &Runner{
		executor:            executor,
		sessionStore:        sessionStore,
		sanitizer:           sanitizer,
		traceStore:          cfg.TraceStore,
		idleTimeout:         cfg.IdleTimeout,
		hardCeiling:         hardCeiling,
		delegationBudgetMax: delegMax,
	}
}

// OnTurnComplete registers a callback fired after each completed turn.
func (r *Runner) OnTurnComplete(cb TurnCallback) {
	if cb == nil {
		return
	}
	r.mu.Lock()
	r.callbacks = append(r.callbacks, cb)
	r.mu.Unlock()
}

// Run executes a single turn through the shared runtime.
func (r *Runner) Run(parent context.Context, req Request) (Result, error) {
	if r.executor == nil {
		return Result{}, fmt.Errorf("turn runner: executor is nil")
	}
	if req.SessionKey == "" {
		return Result{}, fmt.Errorf("turn runner: session key is required")
	}

	start := time.Now()
	traceID := uuid.NewString()
	entrypoint := strings.TrimSpace(req.Entrypoint)
	if entrypoint == "" {
		entrypoint = "direct"
	}

	traceCtx, traceCancel := context.WithTimeout(
		context.WithoutCancel(parent),
		traceWriteTimeout,
	)
	defer traceCancel()
	recorder := newTraceRecorder(traceCtx, r.traceStore, traceID, r.delegationBudgetMax)
	recorder.onDelegation = req.OnDelegation
	recorder.onBudgetWarning = req.OnBudgetWarning
	recorder.start(req.SessionKey, entrypoint, start)

	ctx, cancel, _, runOpts := r.prepareContext(parent, req, recorder, start)
	defer cancel()

	ctx = langosession.WithSessionKey(ctx, req.SessionKey)
	ctx = approval.WithTurnApprovalState(ctx, approval.NewTurnApprovalState())
	ctx = browser.WithRequestState(ctx, browser.NewRequestState())

	report, runErr := r.executor.RunStreamingDetailed(
		ctx,
		req.SessionKey,
		req.Input,
		r.wrapChunkCallback(req.OnChunk),
		runOpts...,
	)

	elapsed := time.Since(start)
	result := r.classifyResult(report, runErr, elapsed)
	if result.TraceID == "" {
		result.TraceID = traceID
	}
	if result.ResponseText != "" && r.sanitizer != nil && r.sanitizer.Enabled() {
		result.ResponseText = r.sanitizer.Sanitize(result.ResponseText)
	}
	if result.UserMessage != "" && r.sanitizer != nil && r.sanitizer.Enabled() {
		result.UserMessage = r.sanitizer.Sanitize(result.UserMessage)
	}
	if result.Summary == "" {
		result.Summary = result.ResponseText
		if result.Summary == "" {
			result.Summary = result.UserMessage
		}
	}

	if result.Outcome == turntrace.OutcomeTimeout && r.sessionStore != nil {
		_ = r.sessionStore.AnnotateTimeout(req.SessionKey, "")
	}

	if result.Outcome != turntrace.OutcomeSuccess {
		recorder.recordTerminalError(result)
	}
	recorder.finish(
		result.Outcome,
		result.Summary,
		result.ErrorCode,
		result.CauseClass,
		result.CauseDetail,
		time.Now(),
	)
	r.fireCallbacks(req.SessionKey)
	return result, nil
}

func (r *Runner) prepareContext(
	parent context.Context,
	req Request,
	recorder *traceRecorder,
	start time.Time,
) (context.Context, context.CancelFunc, *deadline.ExtendableDeadline, []adk.RunOption) {
	var (
		ctx         context.Context
		cancel      context.CancelFunc
		extDeadline *deadline.ExtendableDeadline
		runOpts     []adk.RunOption
	)

	if r.idleTimeout > 0 {
		ctx, extDeadline = deadline.New(parent, r.idleTimeout, r.hardCeiling)
		cancel = extDeadline.Stop
		runOpts = append(runOpts, adk.WithOnActivity(extDeadline.Extend))
	} else {
		ctx, cancel = context.WithTimeout(parent, r.hardCeiling)
	}

	runOpts = append(runOpts, adk.WithOnEvent(func(event *adksession.Event) {
		recorder.recordEvent(event)
	}))

	if req.OnWarning != nil {
		timer := time.AfterFunc(time.Duration(float64(r.hardCeiling)*0.8), func() {
			req.OnWarning(time.Since(start), r.hardCeiling)
		})
		runOpts = append(runOpts, adk.WithOnFinish(func() {
			timer.Stop()
		}))
	}

	return ctx, cancel, extDeadline, runOpts
}

func (r *Runner) wrapChunkCallback(onChunk func(string)) adk.ChunkCallback {
	if onChunk == nil {
		return nil
	}
	return func(chunk string) {
		if chunk == "" {
			return
		}
		if r.sanitizer != nil && r.sanitizer.Enabled() {
			chunk = r.sanitizer.Sanitize(chunk)
		}
		if chunk == "" {
			return
		}
		onChunk(chunk)
	}
}

func (r *Runner) classifyResult(report adk.RunReport, runErr error, elapsed time.Duration) Result {
	result := Result{
		Outcome: turntrace.OutcomeSuccess,
		TraceID: report.TraceID,
		Elapsed: elapsed,
	}

	if runErr == nil {
		switch {
		case strings.TrimSpace(report.Response) != "":
			result.ResponseText = report.Response
			result.Summary = truncateText(report.Response, 240)
			return result
		case report.Diagnostics.ToolResultCount > 0:
			agentErr := &adk.AgentError{
				Code:    adk.ErrEmptyAfterToolUse,
				Message: "agent produced no visible completion after tool output",
				Elapsed: elapsed,
			}
			result.Outcome = turntrace.OutcomeEmptyAfterToolUse
			result.ErrorCode = string(agentErr.Code)
			result.CauseClass = adk.CauseEmptyAfterToolUse
			result.CauseDetail = "tool results were present but no visible assistant completion was produced"
			result.OperatorSummary = "specialist turn ended after tool output without visible completion"
			result.UserMessage = agentErr.UserMessage()
			result.ResponseText = agentErr.UserMessage()
			result.Summary = fmt.Sprintf(
				"empty_after_tool_use after %d tool result(s)",
				report.Diagnostics.ToolResultCount,
			)
			return result
		default:
			result.ResponseText = EmptyResponseFallback
			result.Summary = "generic empty-response fallback"
			return result
		}
	}

		var agentErr *adk.AgentError
	if errors.As(runErr, &agentErr) {
		result.ErrorCode = string(agentErr.Code)
		result.CauseClass = agentErr.CauseClass
		result.CauseDetail = truncateText(agentErr.CauseDetail, 512)
		result.OperatorSummary = agentErr.DiagnosticSummary()
		result.UserMessage = agentErr.UserMessage()
		result.ResponseText = agentErr.UserMessage()
		result.Outcome = outcomeFromAgentError(agentErr)
		result.Summary = agentErr.DiagnosticSummary()
		return result
	}

	result.Outcome = turntrace.OutcomeInternalError
	result.UserMessage = fmt.Sprintf("An error occurred: %s", runErr.Error())
	result.ResponseText = result.UserMessage
	result.CauseClass = adk.CauseInternalRuntimeError
	result.CauseDetail = truncateText(runErr.Error(), 512)
	result.OperatorSummary = fmt.Sprintf("[%s] %s", adk.ErrInternal, adk.CauseInternalRuntimeError)
	result.Summary = truncateText(runErr.Error(), 240)
	return result
}

func outcomeFromAgentError(err *adk.AgentError) TurnOutcome {
	switch err.Code {
	case adk.ErrTimeout, adk.ErrIdleTimeout:
		return turntrace.OutcomeTimeout
	case adk.ErrToolChurn:
		return turntrace.OutcomeLoopDetected
	case adk.ErrEmptyAfterToolUse:
		return turntrace.OutcomeEmptyAfterToolUse
	case adk.ErrToolError:
		return turntrace.OutcomeToolError
	case adk.ErrModelError:
		return turntrace.OutcomeModelError
	default:
		return turntrace.OutcomeInternalError
	}
}

func (r *Runner) fireCallbacks(sessionKey string) {
	r.mu.RLock()
	callbacks := append([]TurnCallback(nil), r.callbacks...)
	r.mu.RUnlock()

	for _, cb := range callbacks {
		cb(sessionKey)
	}
}

type traceRecorder struct {
	ctx             context.Context
	store           turntrace.Store
	traceID         string
	seq             int64
	onDelegation    func(from, to, reason string)
	onBudgetWarning func(used, max int)
	delegationCount int
	delegationMax   int
}

func newTraceRecorder(ctx context.Context, store turntrace.Store, traceID string, delegationMax int) *traceRecorder {
	if delegationMax <= 0 {
		delegationMax = 15
	}
	return &traceRecorder{
		ctx:           ctx,
		store:         store,
		traceID:       traceID,
		delegationMax: delegationMax,
	}
}

func (r *traceRecorder) start(sessionKey, entrypoint string, startedAt time.Time) {
	if r.store == nil {
		return
	}
	if err := r.store.CreateTrace(r.ctx, turntrace.Trace{
		TraceID:    r.traceID,
		SessionKey: sessionKey,
		Entrypoint: entrypoint,
		Outcome:    turntrace.OutcomeRunning,
		StartedAt:  startedAt,
	}); err != nil {
		logger().Warnw("create turn trace", "trace_id", r.traceID, "error", err)
	}
}

func (r *traceRecorder) finish(
	outcome turntrace.Outcome,
	summary string,
	errorCode string,
	causeClass string,
	causeDetail string,
	endedAt time.Time,
) {
	if r.store == nil {
		return
	}
	if err := r.store.FinishTrace(
		r.ctx,
		r.traceID,
		outcome,
		summary,
		errorCode,
		causeClass,
		causeDetail,
		endedAt,
	); err != nil {
		logger().Warnw("finish turn trace", "trace_id", r.traceID, "error", err)
	}
}

func (r *traceRecorder) recordTerminalError(result Result) {
	if r.store == nil {
		return
	}
	r.append(turntrace.Event{
		TraceID:   r.traceID,
		EventType: turntrace.EventTerminalError,
		PayloadJSON: marshalTracePayload(map[string]any{
			"error_code":       result.ErrorCode,
			"cause_class":      result.CauseClass,
			"cause_detail":     result.CauseDetail,
			"operator_summary": result.OperatorSummary,
			"elapsed_ms":       result.Elapsed.Milliseconds(),
		}),
	})
}

func (r *traceRecorder) recordEvent(event *adksession.Event) {
	if r.store == nil || event == nil {
		return
	}
	if event.Actions.TransferToAgent != "" {
		target := event.Actions.TransferToAgent
		eventType := turntrace.EventDelegation
		if target == "lango-orchestrator" {
			eventType = turntrace.EventDelegationReturn
		}
		r.append(turntrace.Event{
			TraceID:     r.traceID,
			EventType:   eventType,
			AgentName:   event.Author,
			PayloadJSON: marshalTracePayload(map[string]any{"to": target}),
		})

		// Fire delegation callback.
		if r.onDelegation != nil {
			r.onDelegation(event.Author, target, "")
		}

		// Track delegation count for budget warning.
		if target != "lango-orchestrator" {
			r.delegationCount++
			if r.onBudgetWarning != nil && r.delegationCount == r.delegationMax*80/100 {
				r.onBudgetWarning(r.delegationCount, r.delegationMax)
			}
		}
	}
	if event.Content == nil {
		return
	}

	for _, part := range event.Content.Parts {
		if part.Text != "" {
			r.append(turntrace.Event{
				TraceID:     r.traceID,
				EventType:   turntrace.EventText,
				AgentName:   event.Author,
				PayloadJSON: marshalTracePayload(map[string]any{"text": truncateText(part.Text, 512)}),
			})
		}
		if part.FunctionCall != nil {
			r.append(turntrace.Event{
				TraceID:       r.traceID,
				EventType:     turntrace.EventToolCall,
				AgentName:     event.Author,
				ToolName:      part.FunctionCall.Name,
				CallSignature: callSignature(event.Author, part.FunctionCall),
				PayloadJSON: marshalTracePayload(map[string]any{
					"id":   part.FunctionCall.ID,
					"args": part.FunctionCall.Args,
				}),
			})
		}
		if part.FunctionResponse != nil {
			r.append(turntrace.Event{
				TraceID:   r.traceID,
				EventType: turntrace.EventToolResult,
				AgentName: event.Author,
				ToolName:  part.FunctionResponse.Name,
				PayloadJSON: marshalTracePayload(map[string]any{
					"id":       part.FunctionResponse.ID,
					"response": part.FunctionResponse.Response,
				}),
			})
		}
	}
}

func (r *traceRecorder) append(event turntrace.Event) {
	r.seq++
	event.Seq = r.seq
	event.CreatedAt = time.Now()
	payload, truncated := truncatePayload(event.PayloadJSON, 1024)
	event.PayloadJSON = payload
	event.PayloadTruncated = truncated
	if err := r.store.AppendEvent(r.ctx, event); err != nil {
		logger().Warnw("append turn trace event",
			"trace_id", r.traceID,
			"seq", event.Seq,
			"event_type", event.EventType,
			"payload_truncated", event.PayloadTruncated,
			"error", err)
	}
}

func callSignature(author string, fc *genai.FunctionCall) string {
	if fc == nil {
		return ""
	}
	args := "{}"
	if len(fc.Args) > 0 {
		if data, err := json.Marshal(fc.Args); err == nil {
			args = string(data)
		}
	}
	return fmt.Sprintf("%s|%s|%s", author, fc.Name, args)
}

func marshalTracePayload(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(data)
}

func truncatePayload(s string, max int) (string, bool) {
	if max <= 0 || len(s) <= max {
		return s, false
	}
	if max <= 3 {
		return s[:max], true
	}
	return s[:max-3] + "...", true
}

func truncateText(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
