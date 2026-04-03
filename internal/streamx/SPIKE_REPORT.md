# E5 Spike Report: ADK Tool Execution Ownership Analysis

**Date**: 2026-04-03
**Unit**: E5 — Parallel Safe-Tool Execution Hook Point Analysis
**Status**: READ-ONLY spike — no code changes
**Branch**: dev

---

## 1. Current Flow: Model Response to Tool Dispatch

### 1.1 End-to-End Sequence

```
Provider (OpenAI/Gemini/Anthropic)
  │
  ▼
ModelAdapter.GenerateContent()          [internal/adk/model.go:167]
  │ Streaming: yields partial LLMResponse with FunctionCall parts
  │ toolCallAccumulator batches deltas → done() emits []*genai.Part
  │
  ▼
ADK runner.Runner.Run()                 [google.golang.org/adk/runner — opaque]
  │ Receives LLMResponse with FunctionCall parts
  │ Dispatches each FunctionCall to the matching tool.Tool.Run()
  │ ── This is the BLACK BOX — ADK runner internals ──
  │
  ▼
functiontool handler                    [internal/adk/tools.go:168]
  │ adaptToolWithOptions wraps agent.Tool.Handler into ADK's handler signature
  │ Injects agentName via ctxkeys, enforces timeout
  │
  ▼
Middleware chain (pre-ADK wrapping)     [internal/toolchain/middleware.go]
  │ Chain()/ChainAll() wraps agent.Tool.Handler BEFORE AdaptTool
  │ Order: Tracing → ExecPolicy → RunLedgerGuard → Approval → Principal
  │         → Hooks(Security+ACL+EventBus) → OutputManager → Learning
  │
  ▼
Actual tool implementation              [internal/tools/*]
```

### 1.2 Key Observation: Tool Dispatch is SEQUENTIAL

The ADK runner (`google.golang.org/adk@v0.6.0`) is an opaque dependency. Based on
observable behavior from Lango's integration:

1. **`runner.Run()` returns `iter.Seq2[*session.Event, error]`** (agent.go:274).
   Lango iterates this single stream sequentially in the `for event, err := range inner`
   loop (agent.go:297).

2. **Each event with FunctionCall(s) is processed one at a time.** The ADK runner emits
   events sequentially — there is no concurrent dispatch visible at the Lango layer.

3. **When the model returns multiple FunctionCalls in one response** (e.g., 2-3 parallel
   tool_use blocks from Claude), the `toolCallAccumulator` (model.go:42) correctly
   assembles them into multiple `genai.Part` entries in a single `LLMResponse.Content`.
   However, the ADK runner receives these as a batch and dispatches them **one by one**,
   yielding separate FunctionResponse events per tool.

4. **Evidence**: The `convertMessages` function (model.go:373) splits multiple
   FunctionResponse parts into separate provider messages (model.go:383-386),
   confirming responses arrive individually, not as a batch.

### 1.3 Middleware Chain Application Point

Middleware wrapping happens at the `agent.Tool.Handler` level, BEFORE `AdaptTool`
converts to ADK's `tool.Tool`:

```
app.go:149-239 (ChainAll calls)
  → toolchain.Chain(tool, middlewares...)    [middleware.go:12]
     → Wraps tool.Handler with each middleware in order
  → adk.AdaptToolForAgentWithTimeout(tool)  [tools.go:160]
     → functiontool.New(cfg, handler)       [tools.go:202]
```

This means each individual tool handler already includes the full middleware stack
when the ADK runner calls it. There is no turn-level interception point in the
current architecture that sees all FunctionCalls before they are individually dispatched.

---

## 2. Hook Point Options for Parallel Execution

### Option A: ADK Plugin BeforeToolCallback Batch Interceptor

**Where**: `plugin.Config.BeforeToolCallback` (plugin.go:236)

**Mechanism**: The ADK plugin system fires `BeforeToolCallback` for each tool
invocation. To enable parallelism, a custom plugin would need to:
1. Detect that multiple FunctionCalls arrived in the same model response
2. Buffer tool calls and execute them concurrently
3. Return results back to the ADK runner

**Feasibility**: LOW
- BeforeToolCallback fires per-tool, not per-batch. It has signature:
  `func(tool.Context, tool.Tool, map[string]any) (map[string]any, error)`
- Returning a non-nil map blocks execution (skips the tool), but there is no way
  to signal "I'll handle this tool asynchronously and return the result later"
- No access to the pending FunctionCall queue — only the current invocation
- Would require maintaining cross-callback state (which calls are from the same
  model response) with no official API to correlate them

**Risk**: HIGH — fighting the ADK's execution model
**Effort**: HIGH

### Option B: Lango-Level Pre-Runner Interceptor

**Where**: Between `runner.Run()` output and Lango's event iteration loop (agent.go:274-297)

**Mechanism**: Instead of modifying tool dispatch, intercept the event stream from
`runner.Run()` and detect batched FunctionCall events. When a batch is detected:
1. Pause the event stream
2. Use `ParallelReadOnlyExecutor` (already exists in streamx) to execute eligible
   tools concurrently
3. Inject FunctionResponse events back into the stream

**Feasibility**: LOW
- `runner.Run()` returns an opaque iterator. The ADK runner internally dispatches
  tool calls and expects to receive results via the session service. Lango cannot
  intercept mid-turn tool execution because the runner owns the tool-call/response
  lifecycle.
- The event stream is a read-only output; Lango cannot inject FunctionResponse
  events that the runner would recognize.

**Risk**: VERY HIGH — requires forking or patching the ADK runner
**Effort**: HIGH

### Option C: Custom Model Adapter with Tool Call Batching

**Where**: `ModelAdapter.GenerateContent()` (model.go:167) — modify the model
response to signal batch intent, and use a custom tool execution layer

**Mechanism**: Instead of the ADK runner dispatching tools, the ModelAdapter would:
1. Detect multiple FunctionCalls in the model response
2. NOT yield them as FunctionCall parts to the ADK runner
3. Execute them directly using `ParallelReadOnlyExecutor`
4. Yield pre-computed FunctionResponse parts instead

**Feasibility**: LOW
- The ADK runner expects to manage the tool-call/response cycle. Bypassing it
  would break session history, plugin callbacks, and agent delegation.
- The model adapter's job is to translate provider responses, not execute tools.
  This violates separation of concerns.

**Risk**: VERY HIGH — breaks ADK session state machine
**Effort**: HIGH

### Option D: Replace ADK Tool Adapter with Parallel-Aware Wrapper

**Where**: `adaptToolWithOptions()` (tools.go:160) — wrap the functiontool handler
to coordinate with sibling tool calls

**Mechanism**: Each adapted tool handler would:
1. Register itself in a shared per-turn batch coordinator
2. Wait for a brief collection window (or until all expected calls arrive)
3. Eligible tools execute concurrently via `ParallelReadOnlyExecutor`
4. Return results to the ADK runner as each completes

**Feasibility**: MEDIUM
- Requires a shared coordinator (per session/turn) that tool handlers can access
  via context. This is architecturally clean using `ctxkeys`.
- The ADK runner calls tools sequentially, so the first tool's handler would need
  to wait for all sibling calls to be dispatched before executing the batch.
  Problem: the ADK runner waits for each tool to return before dispatching the next.
  **This creates a deadlock**: tool 1 waits for tools 2-3 to arrive, but the runner
  won't dispatch tool 2 until tool 1 returns.

**Risk**: HIGH — deadlock unless the ADK runner dispatches concurrently
**Effort**: MEDIUM (if no deadlock), HIGH (if workaround needed)

### Option E: Bypass ADK Runner for Tool Execution (Recommended Investigation)

**Where**: Replace `runner.Runner` with a custom Lango runner that wraps ADK's agent
but handles tool dispatch directly

**Mechanism**:
1. Use `llmagent` to build the agent (already done in agent.go:138)
2. Instead of `runner.New()`, implement a custom run loop that:
   a. Calls the LLM via `ModelAdapter`
   b. Parses FunctionCall parts from the response
   c. Classifies tools by eligibility (`IsEligible` from parallel_executor.go:42)
   d. Dispatches eligible tools via `ParallelReadOnlyExecutor`
   e. Dispatches non-eligible tools sequentially
   f. Constructs FunctionResponse content and feeds back to the LLM
   g. Manages session state via `SessionServiceAdapter`
3. Plugin callbacks (BeforeTool, AfterTool) would be called explicitly in the custom
   runner loop.

**Feasibility**: MEDIUM-HIGH
- Lango already has all the building blocks:
  - `ParallelReadOnlyExecutor` (streamx/parallel_executor.go)
  - `ModelAdapter` (adk/model.go)
  - `SessionServiceAdapter` (adk/session_service.go)
  - Tool eligibility checks (agent.ToolCapability)
  - Middleware chain (toolchain/*)
- The main challenge is reimplementing what `runner.Runner.Run()` does:
  - LLM call → tool dispatch → response assembly → session persistence → loop
  - Agent delegation (transfer_to_agent)
  - Plugin callback orchestration
- This is the **only option that gives full control** over the tool dispatch loop

**Risk**: MEDIUM — significant implementation effort but architecturally sound
**Effort**: HIGH (estimated 3-5 days for MVP)

### Option F: Contribute Parallel Tool Execution to ADK Upstream

**Where**: `google.golang.org/adk/runner`

**Mechanism**: Submit a PR to the ADK project adding a configuration option for
concurrent tool dispatch, respecting tool-level capability metadata.

**Feasibility**: UNKNOWN — depends on ADK project maintainers' roadmap
**Risk**: LOW (no risk to Lango codebase), but UNCERTAIN timeline
**Effort**: MEDIUM for the PR, UNKNOWN for acceptance timeline

---

## 3. ADK Constraints Analysis

### 3.1 What We Know (from Lango's integration surface)

| Aspect | Finding |
|---|---|
| Runner API | `runner.Run()` returns `iter.Seq2[*session.Event, error]` — sequential event stream |
| Tool dispatch | One tool at a time; runner waits for tool return before proceeding |
| Plugin model | Per-tool callbacks (BeforeTool/AfterTool), NOT per-batch |
| Plugin scope | Agent-level (fires for ALL tools), not selective |
| Tool interface | `tool.Tool` has `Run()` but no concurrency metadata |
| Session model | Events are appended sequentially; no parallel event support |

### 3.2 What We Cannot Verify (ADK runner source inaccessible)

- Whether the runner has internal support for concurrent tool dispatch that is not
  exposed in the public API
- Whether `RunConfig` or future versions will add a `ConcurrentTools` option
- How the runner handles multiple FunctionCalls internally (sequential loop vs. batch)
- Whether there is an internal `ToolExecutor` interface that could be replaced

### 3.3 ADK v0.6.0 Plugin Callback Signatures

```go
// Per-tool, sequential — no batch awareness
BeforeToolCallback:  func(tool.Context, tool.Tool, args map[string]any) (map[string]any, error)
AfterToolCallback:   func(tool.Context, tool.Tool, args, result map[string]any, err error) (map[string]any, error)
OnToolErrorCallback: func(tool.Context, tool.Tool, args map[string]any, err error) (map[string]any, error)

// Agent-level, but per-event — not per-batch
OnEventCallback:     func(agent.InvocationContext, *session.Event) (*session.Event, error)
```

None of these callbacks provide batch-level access to multiple simultaneous tool calls.

---

## 4. Recommendation

### Primary: Option E — Custom Runner (Long-term)

**Rationale**: This is the only approach that gives Lango full control over tool
dispatch without fighting the ADK's sequential execution model. All building blocks
already exist in the codebase.

**Prerequisites before implementation**:
1. Audit the ADK runner source (need read access to Go module cache) to understand
   the exact run loop, session state machine, and delegation handling
2. Identify which ADK runner behaviors must be replicated (agent delegation,
   error recovery, session persistence, max turns)
3. Decide whether to keep the ADK runner as a fallback for non-parallel paths

### Secondary: Option F — Upstream Contribution (Opportunistic)

**Rationale**: If the ADK project is receptive, this is the cleanest solution.
Monitor the ADK issue tracker for concurrent tool execution discussions.

### Avoid: Options A-D

Options A-D all try to inject parallelism into the ADK's sequential execution model
without controlling the dispatch loop. They range from impractical (A, B, C) to
deadlock-prone (D).

---

## 5. Existing Assets for Future Implementation

| Asset | Location | Relevance |
|---|---|---|
| `ParallelReadOnlyExecutor` | `internal/streamx/parallel_executor.go` | Core executor — ready to use |
| `IsEligible()` | `internal/streamx/parallel_executor.go:42` | Eligibility check — ready |
| `ToolCapability` | `internal/agent/capability.go:52` | ReadOnly + ConcurrencySafe flags |
| `ToolInvocation` / `ToolResult` | `internal/streamx/parallel_executor.go:13-24` | Input/output types — ready |
| `toolCallAccumulator` | `internal/adk/model.go:42` | Batches streaming tool calls — shows how FunctionCalls arrive as a group |
| `Middleware` / `Chain` / `ChainAll` | `internal/toolchain/middleware.go` | Middleware would need to run inside parallel executor |
| `HookRegistry` | `internal/toolchain/hook_registry.go` | Pre/post hooks must fire per-tool even in parallel |
| `SessionServiceAdapter` | `internal/adk/session_service.go` | Session persistence for custom runner |
| `AgentStreamFanIn` | `internal/streamx/agent_fanin.go` | Pattern for merging concurrent agent streams |
| `errgroup` usage | `internal/streamx/parallel_executor.go:56` | Bounded concurrency pattern already established |

---

## 6. Effort Estimates

| Option | Feasibility | Risk | Effort | Verdict |
|---|---|---|---|---|
| A: Plugin BeforeToolCallback | Low | High | High | REJECT |
| B: Pre-Runner Interceptor | Low | Very High | High | REJECT |
| C: Custom Model Adapter | Low | Very High | High | REJECT |
| D: Parallel-Aware Tool Wrapper | Medium | High (deadlock) | Medium-High | REJECT |
| E: Custom Runner | Medium-High | Medium | High (3-5 days) | RECOMMENDED |
| F: Upstream Contribution | Unknown | Low | Medium + unknown | OPPORTUNISTIC |

---

## 7. Next Steps

1. **Immediate**: Gain read access to ADK runner source at
   `$GOMODCACHE/google.golang.org/adk@v0.6.0/runner/` to audit the exact run loop
2. **Design**: Draft a custom runner spec that preserves all current ADK behaviors
   (delegation, session persistence, error recovery, plugin callbacks) while adding
   parallel tool dispatch for eligible tools
3. **Prototype**: Build a minimal custom runner that handles the single-agent case
   (no delegation) with parallel tool execution, then extend to multi-agent
4. **Validate**: Benchmark against the current sequential runner to quantify latency
   improvement with 2-5 concurrent read-only tools
