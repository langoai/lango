## Capability: parallel-tool-execution

### Requirements

1. **REQ-PTE-001**: `IsEligible(tool)` returns true iff `tool.Capability.ReadOnly && tool.Capability.ConcurrencySafe`
2. **REQ-PTE-002**: `IsEligible(nil)` returns false
3. **REQ-PTE-003**: `NewParallelReadOnlyExecutor(n)` clamps `n` to minimum 1
4. **REQ-PTE-004**: `ExecuteParallel` returns results in same order as input invocations
5. **REQ-PTE-005**: Eligible tools execute concurrently up to `maxConcurrency` limit
6. **REQ-PTE-006**: Non-eligible tools receive error result without handler execution
7. **REQ-PTE-007**: Handler errors are captured in `ToolResult.Error`, not propagated to group
8. **REQ-PTE-008**: Context cancellation stops pending tool invocations
9. **REQ-PTE-009**: Each `ToolResult` includes `Duration` measuring handler execution time
10. **REQ-PTE-010**: Empty invocation slice returns empty result slice
11. **REQ-PTE-011**: Nil tool in invocation produces error result at corresponding index

### Types

```go
type ToolInvocation struct {
    Tool   *agent.Tool
    Params map[string]any
}

type ToolResult struct {
    ToolName string
    Result   any
    Error    error
    Duration time.Duration
}

type ParallelReadOnlyExecutor struct {
    maxConcurrency int
}
```

### Public API

- `NewParallelReadOnlyExecutor(maxConcurrency int) *ParallelReadOnlyExecutor`
- `IsEligible(t *agent.Tool) bool`
- `(*ParallelReadOnlyExecutor).ExecuteParallel(ctx context.Context, invocations []ToolInvocation) []ToolResult`
