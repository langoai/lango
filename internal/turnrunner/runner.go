package turnrunner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	syncatomic "sync/atomic"
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

	// StaleTimeout is how long to wait for the next streaming chunk before
	// considering the stream stale and cancelling the attempt. Zero means
	// use default (30s). Negative disables stale detection.
	StaleTimeout time.Duration

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

	// OnToolCall is called when a tool invocation begins.
	OnToolCall func(callID, toolName string, params map[string]any)

	// OnToolResult is called when a tool invocation completes.
	OnToolResult func(callID, toolName string, success bool, duration time.Duration, preview string)

	// OnThinking is called when thinking/reasoning is detected via genai.Part.Thought.
	// started=true means thinking began (summary is the thought text);
	// started=false means thinking ended.
	OnThinking func(agentName string, started bool, summary string)
}

// Result is the structured result of a completed turn.
type Result struct {
	ResponseText    string
	Outcome         TurnOutcome
	TraceID         string
	UserMessage     string
	Elapsed         time.Duration
	ErrorCode       string
	CauseClass      string
	CauseDetail     string
	OperatorSummary string
	Summary         string
}

// Runner owns timeout handling, durable tracing, and outcome classification.
type Runner struct {
	executor            Executor
	sessionStore        langosession.Store
	sanitizer           Sanitizer
	traceStore          turntrace.Store
	idleTimeout         time.Duration
	hardCeiling         time.Duration
	staleTimeout        time.Duration
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

	staleTimeout := cfg.StaleTimeout
	if staleTimeout == 0 {
		staleTimeout = 30 * time.Second
	}

	return &Runner{
		executor:            executor,
		sessionStore:        sessionStore,
		sanitizer:           sanitizer,
		traceStore:          cfg.TraceStore,
		idleTimeout:         cfg.IdleTimeout,
		hardCeiling:         hardCeiling,
		staleTimeout:        staleTimeout,
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

	recorder := newTraceRecorder(parent, r.traceStore, traceID, r.delegationBudgetMax)
	recorder.onDelegation = req.OnDelegation
	recorder.onBudgetWarning = req.OnBudgetWarning
	recorder.onToolCall = req.OnToolCall
	recorder.onToolResult = req.OnToolResult
	recorder.onThinking = req.OnThinking
	recorder.toolStartedAt = make(map[string]time.Time)
	recorder.start(req.SessionKey, entrypoint, start)

	const maxAttempts = 3

	var result Result
	for attempt := 0; attempt < maxAttempts; attempt++ {
		ctx, cancel, _, runOpts := r.prepareContext(parent, req, recorder, start)

		ctx = langosession.WithSessionKey(ctx, req.SessionKey)
		ctx = langosession.WithTurnID(ctx, traceID)
		ctx = approval.WithTurnApprovalState(ctx, approval.NewTurnApprovalState())
		ctx = browser.WithRequestState(ctx, browser.NewRequestState())

		// Propagate the session's active mode (if any) so the context model
		// adapter and middleware-level enforcement see the same value.
		if r.sessionStore != nil {
			if s, getErr := r.sessionStore.Get(req.SessionKey); getErr == nil && s != nil {
				if name := s.Mode(); name != "" {
					ctx = langosession.WithModeName(ctx, name)
				}
			}
		}

		// Guard: track whether visible chunks were emitted in this attempt.
		// If a provider fails mid-stream after the user saw partial output,
		// retrying would append duplicate content to the TUI.
		var chunksEmitted syncatomic.Bool
		guardedOnChunk := func(chunk string) {
			if chunk != "" {
				chunksEmitted.Store(true)
			}
			if req.OnChunk != nil {
				req.OnChunk(chunk)
			}
		}

		chunkCb, stopStale, staleFlag := r.wrapChunkCallbackWithStale(guardedOnChunk, cancel)
		report, runErr := r.executor.RunStreamingDetailed(
			ctx,
			req.SessionKey,
			req.Input,
			chunkCb,
			runOpts...,
		)
		stopStale()
		cancel()

		elapsed := time.Since(start)
		result = r.classifyResult(report, runErr, elapsed)

		action := adk.RecoveryActionFor(adk.FailureClassification{
			Code:       adk.ErrorCode(result.ErrorCode),
			CauseClass: result.CauseClass,
		})
		if action != adk.RecoveryRetry || attempt >= maxAttempts-1 {
			break
		}
		// Don't retry when partial content was already streamed to the
		// user — UNLESS the failure was caused by a stale stream timeout
		// (stream stalled after some output, retry is appropriate).
		if chunksEmitted.Load() && !staleFlag.Load() {
			break
		}

		// Record recovery event and backoff before next attempt.
		recorder.recordRecovery(adk.RecoveryInfo{
			Action:     "retry",
			CauseClass: result.CauseClass,
			Attempt:    attempt + 1,
			Backoff:    retryBackoff(attempt),
		})
		logger().Infow("retrying after transient failure",
			"attempt", attempt+1,
			"causeClass", result.CauseClass,
			"backoff", retryBackoff(attempt),
		)
		timer := time.NewTimer(retryBackoff(attempt))
		select {
		case <-timer.C:
		case <-parent.Done():
			timer.Stop()
		}
	}
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
	runOpts = append(runOpts, adk.WithOnRecovery(func(info adk.RecoveryInfo) {
		recorder.recordRecovery(info)
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

// wrapChunkCallbackWithStale wraps the chunk callback with a stale stream
// watchdog. The watchdog starts on the first chunk and resets on each
// subsequent chunk. If no chunk arrives for staleTimeout, attemptCancel is
// called. Returns the wrapped callback and a stop function that MUST be called
// when the attempt finishes (to prevent the timer from firing after cancel).
func (r *Runner) wrapChunkCallbackWithStale(onChunk func(string), attemptCancel context.CancelFunc) (adk.ChunkCallback, func(), *syncatomic.Bool) {
	staleFlag := &syncatomic.Bool{}
	inner := r.wrapChunkCallback(onChunk)
	if inner == nil || r.staleTimeout < 0 {
		return inner, func() {}, staleFlag
	}

	var staleTimer *time.Timer
	var mu sync.Mutex
	gotFirstChunk := false

	stop := func() {
		mu.Lock()
		defer mu.Unlock()
		if staleTimer != nil {
			staleTimer.Stop()
		}
	}

	return func(chunk string) {
		mu.Lock()
		if !gotFirstChunk && chunk != "" {
			gotFirstChunk = true
			staleTimer = time.AfterFunc(r.staleTimeout, func() {
				logger().Warnw("stale stream detected, cancelling attempt",
					"staleTimeout", r.staleTimeout)
				staleFlag.Store(true)
				attemptCancel()
			})
		} else if gotFirstChunk && staleTimer != nil {
			staleTimer.Reset(r.staleTimeout)
		}
		mu.Unlock()
		inner(chunk)
	}, stop, staleFlag
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
	parentCtx       context.Context
	store           turntrace.Store
	traceID         string
	seq             int64
	onDelegation    func(from, to, reason string)
	onBudgetWarning func(used, max int)
	onToolCall      func(callID, toolName string, params map[string]any)
	onToolResult    func(callID, toolName string, success bool, duration time.Duration, preview string)
	onThinking      func(agentName string, started bool, summary string)
	delegationCount int
	delegationMax   int
	toolStartedAt   map[string]time.Time    // callID → start time for duration calculation
	inThinking      bool                    // tracks thinking state for boundary detection
	thinkingStart   time.Time               // when thinking phase started
	thinkingText    strings.Builder         // accumulates thought text across chunks
}

func newTraceRecorder(parentCtx context.Context, store turntrace.Store, traceID string, delegationMax int) *traceRecorder {
	if delegationMax <= 0 {
		delegationMax = 15
	}
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	return &traceRecorder{
		parentCtx:     parentCtx,
		store:         store,
		traceID:       traceID,
		delegationMax: delegationMax,
	}
}

func (r *traceRecorder) writeContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(r.parentCtx), traceWriteTimeout)
}

func (r *traceRecorder) start(sessionKey, entrypoint string, startedAt time.Time) {
	if r.store == nil {
		return
	}
	ctx, cancel := r.writeContext()
	defer cancel()
	if err := r.store.CreateTrace(ctx, turntrace.Trace{
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
	ctx, cancel := r.writeContext()
	defer cancel()
	if err := r.store.FinishTrace(
		ctx,
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
		// Thinking detection: fire OnThinking only on boundary transitions.
		if part.Thought && part.Text != "" {
			if !r.inThinking {
				r.inThinking = true
				r.thinkingStart = time.Now()
				r.thinkingText.Reset()
				if r.onThinking != nil {
					r.onThinking(event.Author, true, part.Text)
				}
			}
			r.thinkingText.WriteString(part.Text)
		} else if r.inThinking && !part.Thought {
			summary := r.thinkingText.String()
			duration := time.Since(r.thinkingStart)
			r.inThinking = false
			r.thinkingText.Reset()
			if r.onThinking != nil {
				r.onThinking(event.Author, false, summary)
			}
			_ = duration // available for future use
		}

		if part.Text != "" && !part.Thought {
			r.append(turntrace.Event{
				TraceID:     r.traceID,
				EventType:   turntrace.EventText,
				AgentName:   event.Author,
				PayloadJSON: marshalTracePayload(map[string]any{"text": truncateText(part.Text, 512)}),
			})
		}
		if part.FunctionCall != nil {
			callID := part.FunctionCall.ID
			toolName := part.FunctionCall.Name
			r.append(turntrace.Event{
				TraceID:       r.traceID,
				EventType:     turntrace.EventToolCall,
				AgentName:     event.Author,
				ToolName:      toolName,
				CallSignature: callSignature(event.Author, part.FunctionCall),
				PayloadJSON: marshalTracePayload(map[string]any{
					"id":   callID,
					"args": part.FunctionCall.Args,
				}),
			})

			// Fire OnToolCall callback and record start time.
			if r.toolStartedAt != nil {
				r.toolStartedAt[callID] = time.Now()
			}
			if r.onToolCall != nil {
				r.onToolCall(callID, toolName, part.FunctionCall.Args)
			}
		}
		if part.FunctionResponse != nil {
			callID := part.FunctionResponse.ID
			toolName := part.FunctionResponse.Name
			r.append(turntrace.Event{
				TraceID:   r.traceID,
				EventType: turntrace.EventToolResult,
				AgentName: event.Author,
				ToolName:  toolName,
				PayloadJSON: marshalTracePayload(map[string]any{
					"id":       callID,
					"response": part.FunctionResponse.Response,
				}),
			})

			// Fire OnToolResult callback with computed duration.
			if r.onToolResult != nil {
				var dur time.Duration
				if started, ok := r.toolStartedAt[callID]; ok {
					dur = time.Since(started)
					delete(r.toolStartedAt, callID)
				}
				success := !isErrorResponse(part.FunctionResponse.Response)
				preview := truncateText(responsePreview(part.FunctionResponse.Response), 200)
				r.onToolResult(callID, toolName, success, dur, preview)
			}
		}
	}
}

func (r *traceRecorder) recordRecovery(info adk.RecoveryInfo) {
	if r.store == nil {
		return
	}
	r.append(turntrace.Event{
		TraceID:   r.traceID,
		EventType: turntrace.EventRecoveryAttempt,
		AgentName: info.AgentName,
		PayloadJSON: marshalTracePayload(map[string]any{
			"action": info.Action,
			"agent":  info.AgentName,
			"error":  info.Error,
		}),
	})
}

func (r *traceRecorder) append(event turntrace.Event) {
	r.seq++
	event.Seq = r.seq
	event.CreatedAt = time.Now()
	payload, truncated := truncatePayload(event.PayloadJSON, 1024)
	event.PayloadJSON = payload
	event.PayloadTruncated = truncated
	ctx, cancel := r.writeContext()
	defer cancel()
	if err := r.store.AppendEvent(ctx, event); err != nil {
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

// isErrorResponse checks if a tool response indicates an error.
func isErrorResponse(resp any) bool {
	if resp == nil {
		return false
	}
	switch v := resp.(type) {
	case map[string]any:
		if _, ok := v["error"]; ok {
			return true
		}
	case string:
		return strings.HasPrefix(v, "Error:") || strings.HasPrefix(v, "error:")
	}
	return false
}

// responsePreview returns a short text preview of a tool response.
func responsePreview(resp any) string {
	if resp == nil {
		return ""
	}
	switch v := resp.(type) {
	case string:
		return v
	case map[string]any:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	default:
		return fmt.Sprintf("%v", v)
	}
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

// retryBackoff returns a jittered exponential backoff duration for the given
// zero-based attempt index: base * 2^attempt + random jitter up to 500ms.
func retryBackoff(attempt int) time.Duration {
	base := time.Duration(1<<uint(attempt)) * time.Second // 1s, 2s, 4s
	jitter := time.Duration(rand.Int63n(int64(500 * time.Millisecond)))
	return base + jitter
}
