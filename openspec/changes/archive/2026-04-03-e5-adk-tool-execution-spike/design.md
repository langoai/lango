## Context

The ADK runner (`google.golang.org/adk@v0.6.0`) owns the tool dispatch loop. Lango wraps tools via `adaptToolWithOptions()` and applies middleware via `toolchain.ChainAll()` before handing them to the ADK. When the model returns multiple FunctionCalls in one response, the ADK runner dispatches them sequentially. This spike analyzes where parallel execution could be inserted.

Key files analyzed:
- `internal/adk/agent.go` — Runner integration, event iteration loop
- `internal/adk/model.go` — `toolCallAccumulator`, `ModelAdapter.GenerateContent()`
- `internal/adk/tools.go` — `adaptToolWithOptions()`, `functiontool.New()`
- `internal/adk/plugin.go` — ADK plugin callbacks (BeforeTool/AfterTool)
- `internal/toolchain/middleware.go` — `Chain()`, `ChainAll()`
- `internal/toolchain/hooks.go` — `HookRegistry`, `PreToolHook`, `PostToolHook`
- `internal/streamx/parallel_executor.go` — Existing `ParallelReadOnlyExecutor`

## Goals / Non-Goals

**Goals:**
- Map the complete tool execution ownership chain from model response to tool handler
- Identify all possible hook points for inserting parallel tool dispatch
- Rate each hook point by feasibility, risk, and effort
- Recommend the best path forward

**Non-Goals:**
- No code changes (read-only spike)
- No ADK runner source audit (module cache inaccessible during this spike)
- No implementation prototype

## Decisions

| Decision | Rationale |
|---|---|
| Recommend custom runner (Option E) over plugin-based approaches | ADK plugin callbacks are per-tool, not per-batch. No way to intercept the full FunctionCall batch before individual dispatch. |
| Reject tool-wrapper approach (Option D) | ADK runner dispatches tools sequentially and waits for return. A wrapper that buffers sibling calls would deadlock. |
| Document 6 options exhaustively | Future implementers need full context to avoid rediscovering rejected approaches. |

## Risks / Trade-offs

- **[Risk]** ADK runner source was not audited (Go module cache inaccessible) → Mitigation: Recommendations are based on observable API surface and behavior. First next step is to audit the runner source.
- **[Risk]** Custom runner (Option E) may need to reimplement complex ADK behaviors (delegation, session state machine) → Mitigation: Start with single-agent case, extend incrementally.
- **[Risk]** ADK future versions may add native parallel support, making custom runner unnecessary → Mitigation: Monitor ADK releases; Option F (upstream contribution) kept as opportunistic path.
