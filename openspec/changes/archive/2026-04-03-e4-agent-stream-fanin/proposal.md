## Why

Child agent orchestration requires merging output streams from multiple concurrent child agents into a single observable stream, with lifecycle tracking for monitoring and debugging. The stream-combinators (E1) and ProgressBus (E2) primitives are complete; this unit composes them into the agent-specific fan-in pattern needed by the multi-agent runtime.

## What Changes

- Add `AgentStreamFanIn` struct that wraps the `Merge` combinator for child agent `Stream[string]` outputs
- Emit `ProgressStarted`, `ProgressCompleted`, and `ProgressFailed` events per child via `ProgressBus`
- Support nil bus (optional dependency) for contexts where progress tracking is not needed
- Wrap each child stream to detect completion/error and emit corresponding lifecycle events

## Capabilities

### New Capabilities
- `agent-stream-fanin`: Child-agent stream fan-in using Merge combinator with ProgressBus lifecycle emission

### Modified Capabilities
- `stream-combinators`: No spec-level requirement changes (implementation only consumes existing API)

## Impact

- New files: `internal/streamx/agent_fanin.go`, `internal/streamx/agent_fanin_test.go`
- Depends on: `internal/streamx/` (Stream, Tag, Merge, ProgressBus)
- No breaking changes to existing APIs
- Downstream: multi-agent orchestration layer will consume AgentStreamFanIn
