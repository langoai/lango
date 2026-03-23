package adk

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	adk_agent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/plugin"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"

	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/runledger"
	internal "github.com/langoai/lango/internal/session"
)

func logger() *zap.SugaredLogger { return logging.Agent() }

// ErrorFixProvider returns a known fix for a tool error if one exists.
// Implemented by learning.Engine.
type ErrorFixProvider interface {
	GetFixForError(ctx context.Context, toolName string, err error) (string, bool)
}

// defaultMaxTurns is the default maximum number of tool-calling iterations per agent run.
const defaultMaxTurns = 50

// maxConsecutiveSameToolCalls is the safety limit for detecting tool churn loops.
// When the same tool is called this many times in a row without any other tool
// call or text generation, the agent run is force-stopped.
const maxConsecutiveSameToolCalls = 5

// AgentOption configures optional Agent behavior at construction time.
type AgentOption func(*agentOptions)

type agentOptions struct {
	tokenBudget         int
	maxTurns            int
	errorFixProvider    ErrorFixProvider
	rootSessionObserver func(string)
	childLifecycleHook  func(internal.SessionLifecycleEvent)
	isolatedAgents      []string
	plugins             []*plugin.Plugin
}

// WithAgentTokenBudget sets the session history token budget.
// Use ModelTokenBudget(modelName) to derive an appropriate value.
func WithAgentTokenBudget(budget int) AgentOption {
	return func(o *agentOptions) { o.tokenBudget = budget }
}

// WithAgentMaxTurns sets the maximum number of tool-calling turns per run.
func WithAgentMaxTurns(n int) AgentOption {
	return func(o *agentOptions) { o.maxTurns = n }
}

// WithAgentErrorFixProvider sets a learning-based error correction provider.
func WithAgentErrorFixProvider(p ErrorFixProvider) AgentOption {
	return func(o *agentOptions) { o.errorFixProvider = p }
}

// WithAgentRootSessionObserver records root session creation events.
func WithAgentRootSessionObserver(fn func(string)) AgentOption {
	return func(o *agentOptions) { o.rootSessionObserver = fn }
}

// WithAgentChildLifecycleHook records synthetic child-session lifecycle events.
func WithAgentChildLifecycleHook(fn func(internal.SessionLifecycleEvent)) AgentOption {
	return func(o *agentOptions) { o.childLifecycleHook = fn }
}

// WithAgentIsolatedAgents marks agent names that should use child session history routing.
func WithAgentIsolatedAgents(names []string) AgentOption {
	return func(o *agentOptions) { o.isolatedAgents = append([]string(nil), names...) }
}

// WithPlugins adds ADK plugins to the runner configuration.
// Plugins provide agent-level callbacks (BeforeTool, AfterTool, OnEvent, etc.)
// that are executed by the ADK runner for every tool invocation.
// Zero plugins preserves current behavior.
func WithPlugins(plugins ...*plugin.Plugin) AgentOption {
	return func(o *agentOptions) { o.plugins = append(o.plugins, plugins...) }
}

// Agent wraps the ADK runner for integration with Lango.
type Agent struct {
	runner           *runner.Runner
	adkAgent         adk_agent.Agent
	maxTurns         int              // 0 = defaultMaxTurns
	errorFixProvider ErrorFixProvider // optional: for self-correction on errors
	sessionService   *SessionServiceAdapter
	isolatedAgents   map[string]struct{}
}

// NewAgent creates a new Agent instance.
func NewAgent(ctx context.Context, tools []tool.Tool, mod model.LLM, systemPrompt string, store internal.Store, opts ...AgentOption) (*Agent, error) {
	var o agentOptions
	for _, fn := range opts {
		fn(&o)
	}

	// Create LLM Agent
	cfg := llmagent.Config{
		Name:        "lango-agent",
		Description: "Lango Assistant",
		Model:       mod,
		Tools:       tools,
		Instruction: systemPrompt,
	}

	adkAgent, err := llmagent.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("create llm agent: %w", err)
	}

	// Create Session Service
	sessService := NewSessionServiceAdapter(store, "lango-agent")
	if o.tokenBudget > 0 {
		sessService.WithTokenBudget(o.tokenBudget)
	}
	if o.rootSessionObserver != nil {
		sessService.WithRootSessionObserver(o.rootSessionObserver)
	}
	if o.childLifecycleHook != nil {
		sessService.WithChildLifecycleHook(o.childLifecycleHook)
	}
	if len(o.isolatedAgents) > 0 {
		sessService.WithIsolatedAgents(o.isolatedAgents)
	}

	// Create Runner
	runnerCfg := runner.Config{
		AppName:        "lango",
		Agent:          adkAgent,
		SessionService: sessService,
	}
	if len(o.plugins) > 0 {
		runnerCfg.PluginConfig = runner.PluginConfig{
			Plugins: o.plugins,
		}
	}

	r, err := runner.New(runnerCfg)
	if err != nil {
		return nil, fmt.Errorf("create runner: %w", err)
	}

	return &Agent{
		runner:           r,
		adkAgent:         adkAgent,
		maxTurns:         o.maxTurns,
		errorFixProvider: o.errorFixProvider,
		sessionService:   sessService,
		isolatedAgents:   makeIsolatedAgentSet(o.isolatedAgents),
	}, nil
}

// NewAgentFromADK creates a Lango Agent wrapping a pre-built ADK agent.
// Used for multi-agent orchestration where the agent tree is built externally.
func NewAgentFromADK(adkAgent adk_agent.Agent, store internal.Store, opts ...AgentOption) (*Agent, error) {
	var o agentOptions
	for _, fn := range opts {
		fn(&o)
	}

	sessService := NewSessionServiceAdapter(store, adkAgent.Name())
	if o.tokenBudget > 0 {
		sessService.WithTokenBudget(o.tokenBudget)
	}
	if o.rootSessionObserver != nil {
		sessService.WithRootSessionObserver(o.rootSessionObserver)
	}
	if o.childLifecycleHook != nil {
		sessService.WithChildLifecycleHook(o.childLifecycleHook)
	}
	if len(o.isolatedAgents) > 0 {
		sessService.WithIsolatedAgents(o.isolatedAgents)
	}

	runnerCfg := runner.Config{
		AppName:        "lango",
		Agent:          adkAgent,
		SessionService: sessService,
	}
	if len(o.plugins) > 0 {
		runnerCfg.PluginConfig = runner.PluginConfig{
			Plugins: o.plugins,
		}
	}

	r, err := runner.New(runnerCfg)
	if err != nil {
		return nil, fmt.Errorf("create runner: %w", err)
	}

	return &Agent{
		runner:           r,
		adkAgent:         adkAgent,
		maxTurns:         o.maxTurns,
		errorFixProvider: o.errorFixProvider,
		sessionService:   sessService,
		isolatedAgents:   makeIsolatedAgentSet(o.isolatedAgents),
	}, nil
}

// WithMaxTurns sets the maximum number of tool-calling turns per run.
// Zero or negative values use the default (50).
func (a *Agent) WithMaxTurns(n int) *Agent {
	a.maxTurns = n
	return a
}

// WithErrorFixProvider sets an optional provider for learning-based error correction.
// When set, the agent will attempt to apply known fixes on errors before giving up.
func (a *Agent) WithErrorFixProvider(p ErrorFixProvider) *Agent {
	a.errorFixProvider = p
	return a
}

// ADKAgent returns the underlying ADK agent, or nil if not available.
func (a *Agent) ADKAgent() adk_agent.Agent {
	return a.adkAgent
}

// Run executes the agent for a given session and returns an event iterator.
// It enforces a maximum turn limit to prevent unbounded tool-calling loops.
func (a *Agent) Run(ctx context.Context, sessionID string, input string) iter.Seq2[*session.Event, error] {
	ctx = runledger.WithSnapshotCache(ctx)

	// Create user content
	userMsg := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{{Text: input}},
	}

	// Config for run
	runCfg := adk_agent.RunConfig{
		// Defaults
	}

	maxTurns := a.maxTurns
	if maxTurns <= 0 {
		maxTurns = defaultMaxTurns
	}

	// Execute via Runner with turn limit enforcement.
	inner := a.runner.Run(ctx, "user", sessionID, userMsg, runCfg)

	return func(yield func(*session.Event, error) bool) {
		turnCount := 0
		warnedAtThreshold := false

		// Delegation tracking (loop-level scope for budget expansion).
		budgetExpanded := false
		delegationCount := 0
		uniqueAgents := make(map[string]struct{})
		plannerInvolved := false

		// Wrap-up tracking (loop-level scope).
		wrapUpBudget := 1 // default: 1 wrap-up turn
		wrapUpRemaining := 0
		inWrapUp := false

		// Tool churn detection: force-stop when the same tool is called
		// consecutively without the model producing any other tool call or
		// text response (indicates the model is stuck in a loop).
		lastToolName := ""
		consecutiveSameToolCalls := 0

		for event, err := range inner {
			if err != nil {
				yield(nil, err)
				return
			}

			// Track delegations for dynamic budget expansion.
			if isDelegationEvent(event) {
				target := event.Actions.TransferToAgent
				delegationCount++
				if target != "" && target != "lango-orchestrator" {
					uniqueAgents[target] = struct{}{}
					if target == "planner" {
						plannerInvolved = true
					}
				}

				// Expand budget once when multi-agent complexity is detected.
				// Triggers: planner involvement OR 3+ delegations OR 2+ unique agents.
				if !budgetExpanded && (plannerInvolved || delegationCount >= 3 || len(uniqueAgents) >= 2) {
					budgetExpanded = true
					oldMax := maxTurns
					maxTurns = maxTurns * 3 / 2
					wrapUpBudget = 3
					logger().Infow("multi-agent task detected, expanding turn budget",
						"session", sessionID,
						"oldMaxTurns", oldMax,
						"newMaxTurns", maxTurns,
						"uniqueAgents", len(uniqueAgents),
						"delegationCount", delegationCount,
						"plannerInvolved", plannerInvolved)
				}
			}

			// Reset tool churn counter when the model generates a text-only
			// response (reasoning step, not a tool call), indicating it is
			// not stuck in a single-tool loop.
			if event.Content != nil && hasText(event) && !hasFunctionCalls(event) {
				lastToolName = ""
				consecutiveSameToolCalls = 0
			}

			// Count events containing function calls as agent turns.
			// Delegation transfers (agent-to-agent routing) are not counted
			// because they are routing overhead, not actual tool work.
			if event.Content != nil && hasFunctionCalls(event) && !isDelegationEvent(event) {
				// Tool churn detection: same tool called consecutively.
				if toolName := extractPrimaryToolName(event); toolName != "" {
					if toolName == lastToolName {
						consecutiveSameToolCalls++
					} else {
						lastToolName = toolName
						consecutiveSameToolCalls = 1
					}
					if consecutiveSameToolCalls >= maxConsecutiveSameToolCalls {
						logger().Warnw("tool churn detected, forcing stop",
							"session", sessionID,
							"tool", toolName,
							"consecutiveCalls", consecutiveSameToolCalls,
							"maxAllowed", maxConsecutiveSameToolCalls)
						yield(nil, fmt.Errorf("tool %q called %d times consecutively, forcing stop", toolName, consecutiveSameToolCalls))
						return
					}
				}

				turnCount++

				// Log a warning at 80% of the turn limit for observability.
				if !warnedAtThreshold && maxTurns > 0 && turnCount == maxTurns*4/5 {
					warnedAtThreshold = true
					logger().Warnw("agent nearing turn limit",
						"session", sessionID,
						"turns", turnCount,
						"maxTurns", maxTurns,
						"remaining", maxTurns-turnCount)
				}

				if turnCount > maxTurns {
					if !inWrapUp {
						inWrapUp = true
						wrapUpRemaining = wrapUpBudget
						logger().Warnw("agent turn limit reached, granting wrap-up",
							"session", sessionID,
							"turns", turnCount,
							"maxTurns", maxTurns,
							"wrapUpBudget", wrapUpBudget)
					}
					wrapUpRemaining--
					if wrapUpRemaining < 0 {
						logger().Warnw("agent max turns exceeded",
							"session", sessionID,
							"turns", turnCount,
							"maxTurns", maxTurns)
						yield(nil, fmt.Errorf("agent exceeded maximum turn limit (%d)", maxTurns))
						return
					}
					if !yield(event, nil) {
						return
					}
					continue
				}
			}
			if !yield(event, nil) {
				return
			}
		}
	}
}

// hasFunctionCalls reports whether the event contains any FunctionCall parts.
func hasFunctionCalls(e *session.Event) bool {
	if e.Content == nil {
		return false
	}
	for _, p := range e.Content.Parts {
		if p.FunctionCall != nil {
			return true
		}
	}
	return false
}

// isDelegationEvent reports whether the event is a pure agent-to-agent
// delegation transfer (routing overhead, not actual tool work).
func isDelegationEvent(e *session.Event) bool {
	return e.Actions.TransferToAgent != ""
}

// extractPrimaryToolName returns the name of the first FunctionCall in the event,
// or empty string if the event contains no function calls.
func extractPrimaryToolName(e *session.Event) string {
	if e.Content == nil {
		return ""
	}
	for _, p := range e.Content.Parts {
		if p.FunctionCall != nil {
			return p.FunctionCall.Name
		}
	}
	return ""
}

// RunAndCollect executes the agent and returns the full text response.
// If the agent encounters a "failed to find agent" error (hallucinated agent
// name), it sends a correction message and retries once.
func (a *Agent) RunAndCollect(ctx context.Context, sessionID, input string, opts ...RunOption) (string, error) {
	var ro runOptions
	for _, o := range opts {
		o(&ro)
	}
	start := time.Now()
	resp, err := a.runAndCollectOnce(ctx, sessionID, input, &ro)
	if err == nil {
		// Safety net: detect [REJECT] text from sub-agents that failed to
		// call transfer_to_agent and force re-routing through the orchestrator.
		if resp != "" && containsRejectPattern(resp) && len(a.adkAgent.SubAgents()) > 0 {
			if a.sessionService != nil {
				_ = a.sessionService.DiscardActiveChildWithReason(sessionID, "escalated without producing a result")
			}
			logger().Warnw("sub-agent REJECT detected in text, forcing re-route",
				"session", sessionID,
				"response_preview", truncate(resp, 100))
			correction := fmt.Sprintf(
				"[System: A sub-agent could not handle this request. "+
					"Re-evaluate and route to a different agent or answer directly. "+
					"Original user request: %s]", input)
			retryResp, retryErr := a.runAndCollectOnce(ctx, sessionID, correction, &ro)
			if retryErr == nil && retryResp != "" && !containsRejectPattern(retryResp) {
				if a.sessionService != nil {
					_ = a.sessionService.CloseActiveChild(sessionID)
				}
				return retryResp, nil
			}
			// Fall through with original response if retry also fails.
		}
		if a.sessionService != nil {
			_ = a.sessionService.CloseActiveChild(sessionID)
		}

		if resp == "" {
			logger().Warnw("agent returned empty response",
				"session", sessionID,
				"elapsed", time.Since(start).String())
		} else {
			logger().Debugw("agent run completed",
				"session", sessionID,
				"elapsed", time.Since(start).String(),
				"response_len", len(resp))
		}
		return resp, nil
	}

	badAgent := extractMissingAgent(err)
	if badAgent == "" || len(a.adkAgent.SubAgents()) == 0 {
		// Tool churn recovery: if a sub-agent was stopped due to repeated
		// same-tool calls, discard the stuck child session and let the
		// orchestrator respond using whatever information was gathered.
		var agentErr *AgentError
		if errors.As(err, &agentErr) && agentErr.Code == ErrToolChurn && len(a.adkAgent.SubAgents()) > 0 {
			if a.sessionService != nil {
				_ = a.sessionService.DiscardActiveChildWithReason(sessionID, "same tool loop detected")
			}
			recovery := "[System: The previous sub-agent was stopped because it called the same tool repeatedly without producing a response. " +
				"Do NOT delegate to the same sub-agent again for this request. " +
				"Respond to the user directly using whatever information has already been gathered in this conversation. " +
				"If no useful information was found, apologize and tell the user you were unable to complete the search.]"
			logger().Infow("tool churn recovery attempt",
				"session", sessionID,
				"elapsed", time.Since(start).String())
			retryResp, retryErr := a.runAndCollectOnce(ctx, sessionID, recovery, &ro)
			if retryErr == nil && retryResp != "" {
				if a.sessionService != nil {
					_ = a.sessionService.CloseActiveChild(sessionID)
				}
				logger().Infow("tool churn recovery succeeded",
					"session", sessionID,
					"elapsed", time.Since(start).String(),
					"response_len", len(retryResp))
				return retryResp, nil
			}
			logger().Warnw("tool churn recovery failed",
				"session", sessionID,
				"elapsed", time.Since(start).String(),
				"error", retryErr)
			if retryResp != "" && len(retryResp) > len(resp) {
				resp = retryResp
			}
		}

		// Try learning-based error correction before giving up.
		if a.errorFixProvider != nil {
			if fix, ok := a.errorFixProvider.GetFixForError(ctx, "", err); ok {
				correction := fmt.Sprintf(
					"[System: Previous action failed with: %s. Suggested fix: %s. Please retry.]",
					err.Error(), fix)
				logger().Infow("applying learned fix for error",
					"session", sessionID,
					"fix", fix,
					"elapsed", time.Since(start).String())
				retryResp, retryErr := a.runAndCollectOnce(ctx, sessionID, correction, &ro)
				if retryErr == nil {
					return retryResp, nil
				}
				logger().Warnw("learned fix retry failed",
					"session", sessionID,
					"error", retryErr)
				// Prefer whichever partial result is longer.
				if retryResp != "" && len(retryResp) > len(resp) {
					resp = retryResp
				}
			}
		}

		logger().Warnw("agent run failed",
			"session", sessionID,
			"elapsed", time.Since(start).String(),
			"error", err)
		// Return partial result from the best attempt if available.
		if a.sessionService != nil {
			_ = a.sessionService.DiscardActiveChildWithReason(sessionID, "agent error")
		}
		return resp, err
	}

	// Build correction message and retry once.
	names := subAgentNames(a.adkAgent)
	correction := fmt.Sprintf(
		"[System: Agent %q does not exist. Valid agents: %s. Please retry using one of the valid agent names listed above.]",
		badAgent, strings.Join(names, ", "))
	logger().Warnw("agent name hallucination detected, retrying",
		"hallucinated", badAgent,
		"valid_agents", names,
		"session", sessionID,
		"elapsed", time.Since(start).String())

	retryStart := time.Now()
	retryResp, retryErr := a.runAndCollectOnce(ctx, sessionID, correction, &ro)
	if retryErr != nil {
		logger().Errorw("agent hallucination retry failed",
			"session", sessionID,
			"retry_elapsed", time.Since(retryStart).String(),
			"total_elapsed", time.Since(start).String(),
			"error", retryErr)
		// Return best partial result from either attempt.
		if retryResp != "" && len(retryResp) > len(resp) {
			resp = retryResp
		}
		if a.sessionService != nil {
			_ = a.sessionService.DiscardActiveChildWithReason(sessionID, "agent error")
		}
		return resp, retryErr
	}
	resp = retryResp
	err = nil
	if a.sessionService != nil {
		_ = a.sessionService.CloseActiveChild(sessionID)
	}

	logger().Infow("agent hallucination retry succeeded",
		"session", sessionID,
		"retry_elapsed", time.Since(retryStart).String(),
		"total_elapsed", time.Since(start).String(),
		"response_len", len(resp))
	return resp, nil
}

// runAndCollectOnce executes a single agent run and collects text output.
// It tracks whether partial (streaming) events were seen to avoid
// double-counting text that appears in both partial chunks and the
// final non-partial response.
func (a *Agent) runAndCollectOnce(ctx context.Context, sessionID, input string, ro *runOptions) (string, error) {
	var b strings.Builder
	var sawVisiblePartial bool

	start := time.Now()

	for event, err := range a.Run(ctx, sessionID, input) {
		if err != nil {
			partial := b.String()
			if a.sessionService != nil {
				_ = a.sessionService.DiscardActiveChildWithReason(sessionID, discardReasonForError(err))
			}
			return partial, &AgentError{
				Code:    classifyError(err),
				Message: "agent error",
				Cause:   err,
				Partial: partial,
				Elapsed: time.Since(start),
			}
		}

		// Log agent event for multi-agent observability.
		if event.Author != "" {
			if event.Actions.TransferToAgent != "" {
				logger().Debugw("agent delegation",
					"from", event.Author,
					"to", event.Actions.TransferToAgent,
					"session", sessionID)
			} else if hasText(event) {
				logger().Debugw("agent response",
					"agent", event.Author,
					"session", sessionID)
			}
		}

		if event.Content == nil {
			continue
		}

		// Signal activity for deadline extension.
		if ro != nil && ro.onActivity != nil {
			if hasText(event) || hasFunctionCalls(event) {
				ro.onActivity()
			}
		}

		if event.Partial {
			// Streaming text chunk — collect incrementally.
			sawVisiblePartial = true
			for _, part := range event.Content.Parts {
				if part.Text != "" {
					b.WriteString(part.Text)
				}
			}
		} else if !sawVisiblePartial {
			// Non-streaming mode: no partial events were seen,
			// so collect from the final complete response.
			for _, part := range event.Content.Parts {
				if part.Text != "" {
					b.WriteString(part.Text)
				}
			}
		}
		// If sawVisiblePartial && !event.Partial: this is the final done event
		// in streaming mode. Its text duplicates partial chunks, so skip.
	}

	// ADK's streaming iterator silently terminates on context deadline
	// without yielding an error. Check context after iteration to detect
	// timeout that the iterator failed to propagate.
	if err := ctx.Err(); err != nil {
		partial := b.String()
		if a.sessionService != nil {
			_ = a.sessionService.DiscardActiveChildWithReason(sessionID, discardReasonForError(err))
		}
		return partial, &AgentError{
			Code:    ErrTimeout,
			Message: "agent error",
			Cause:   err,
			Partial: partial,
			Elapsed: time.Since(start),
		}
	}

	return b.String(), nil
}

// containsRejectPattern reports whether the text contains a [REJECT] marker.
// Uses strings.Contains for efficiency since the pattern is a literal string.
func containsRejectPattern(text string) bool {
	return strings.Contains(text, "[REJECT]")
}

// truncate returns the first n runes of s, appending "..." if truncated.
// Uses rune counting to avoid splitting multi-byte UTF-8 characters.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

// reAgentNotFound matches ADK's "failed to find agent: <name>" error.
var reAgentNotFound = regexp.MustCompile(`failed to find agent: (\S+)`)

// extractMissingAgent returns the hallucinated agent name from an error,
// or an empty string if the error does not match the pattern.
func extractMissingAgent(err error) string {
	m := reAgentNotFound.FindStringSubmatch(err.Error())
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

// subAgentNames returns the names of all immediate sub-agents.
func subAgentNames(a adk_agent.Agent) []string {
	subs := a.SubAgents()
	names := make([]string, len(subs))
	for i, s := range subs {
		names[i] = s.Name()
	}
	return names
}

// RunOption configures optional behavior for a single agent run.
type RunOption func(*runOptions)

type runOptions struct {
	onActivity func()
}

// WithOnActivity sets a callback that is invoked whenever the agent produces
// activity (text chunks, function calls). Useful for extending deadlines.
func WithOnActivity(fn func()) RunOption {
	return func(o *runOptions) { o.onActivity = fn }
}

// ChunkCallback is called for each streaming text chunk during agent execution.
type ChunkCallback func(chunk string)

// RunStreaming executes the agent and streams partial text chunks via the callback.
// It returns the full accumulated response text for backward compatibility.
func (a *Agent) RunStreaming(ctx context.Context, sessionID, input string, onChunk ChunkCallback, opts ...RunOption) (string, error) {
	var ro runOptions
	for _, o := range opts {
		o(&ro)
	}

	var b strings.Builder
	var sawVisiblePartial bool
	start := time.Now()

	for event, err := range a.Run(ctx, sessionID, input) {
		if err != nil {
			partial := b.String()
			return partial, &AgentError{
				Code:    classifyError(err),
				Message: "agent error",
				Cause:   err,
				Partial: partial,
				Elapsed: time.Since(start),
			}
		}

		if event.Content == nil {
			continue
		}

		// Signal activity for deadline extension.
		if ro.onActivity != nil {
			if hasText(event) || hasFunctionCalls(event) {
				ro.onActivity()
			}
		}

		if event.Partial {
			sawVisiblePartial = true
			for _, part := range event.Content.Parts {
				if part.Text != "" {
					b.WriteString(part.Text)
					if onChunk != nil {
						onChunk(part.Text)
					}
				}
			}
		} else if !sawVisiblePartial {
			// Non-streaming mode: collect from final response.
			for _, part := range event.Content.Parts {
				if part.Text != "" {
					b.WriteString(part.Text)
				}
			}
		}
	}

	// ADK's streaming iterator silently terminates on context deadline
	// without yielding an error. Check context after iteration to detect
	// timeout that the iterator failed to propagate.
	if err := ctx.Err(); err != nil {
		partial := b.String()
		if a.sessionService != nil {
			_ = a.sessionService.DiscardActiveChildWithReason(sessionID, discardReasonForError(err))
		}
		return partial, &AgentError{
			Code:    ErrTimeout,
			Message: "agent error",
			Cause:   err,
			Partial: partial,
			Elapsed: time.Since(start),
		}
	}
	if a.sessionService != nil {
		_ = a.sessionService.CloseActiveChild(sessionID)
	}
	return b.String(), nil
}

func discardReasonForError(err error) string {
	switch classifyError(err) {
	case ErrToolChurn:
		return "same tool loop detected"
	default:
		return "agent error"
	}
}

// hasText reports whether the event contains any non-empty text part.
func hasText(e *session.Event) bool {
	if e.Content == nil {
		return false
	}
	for _, p := range e.Content.Parts {
		if p.Text != "" {
			return true
		}
	}
	return false
}

func makeIsolatedAgentSet(names []string) map[string]struct{} {
	if len(names) == 0 {
		return nil
	}

	out := make(map[string]struct{}, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		out[name] = struct{}{}
	}
	return out
}

func (a *Agent) shouldCollectUserText(author string) bool {
	if len(a.isolatedAgents) == 0 || author == "" {
		return true
	}
	_, isolated := a.isolatedAgents[author]
	return !isolated
}
