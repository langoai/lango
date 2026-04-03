## Why

The agent runtime needs to batch concurrent execution of read-only, concurrency-safe tools to reduce turn latency. Currently all tool invocations are sequential, even when multiple tools are safe to run in parallel. Unit E3 adds a `ParallelReadOnlyExecutor` to `internal/streamx/` that leverages `agent.Tool.Capability` flags (A1) and the existing stream combinator package (E1).

## What Changes

- Add `ToolInvocation` and `ToolResult` types for representing tool call requests and outcomes
- Add `ParallelReadOnlyExecutor` that uses `errgroup.SetLimit` to run eligible tools concurrently
- Add `IsEligible` function to check `ReadOnly && ConcurrencySafe` capability flags
- Non-eligible tools are rejected with descriptive errors without execution
- Results preserve invocation order via indexed slice (not channels)
- Context cancellation propagates to pending goroutines

## Capabilities

### New Capabilities
- `parallel-tool-execution`: Concurrent execution of read-only, concurrency-safe tool invocations with configurable concurrency limits

### Modified Capabilities
- `stream-combinators`: Extended with parallel executor types (ToolInvocation, ToolResult) and ParallelReadOnlyExecutor

## Impact

- **Code**: `internal/streamx/parallel_executor.go` (new file)
- **Dependencies**: Uses `golang.org/x/sync/errgroup` (already in go.mod), `internal/agent.Tool` and `agent.ToolCapability`
- **APIs**: New public types `ToolInvocation`, `ToolResult`, `ParallelReadOnlyExecutor`, and function `IsEligible`
- **Downstream**: Future agent runtime can use this executor to batch parallel tool calls
