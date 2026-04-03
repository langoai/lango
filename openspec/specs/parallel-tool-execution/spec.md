## Purpose

Parallel Tool Execution provides a concurrency-safe mechanism for executing multiple read-only agent tools simultaneously. It validates tool eligibility based on capability metadata, enforces a configurable concurrency limit, and preserves result ordering to match input invocation order.

## Requirements

### Requirement: IsEligible validates tool concurrency safety

`IsEligible(tool)` SHALL return `true` if and only if the tool's capability declares both `ReadOnly` and `ConcurrencySafe` as true. A nil tool SHALL never be eligible.

#### Scenario: Eligible tool with both capabilities
- **WHEN** `IsEligible` is called with a tool that has `ReadOnly: true` and `ConcurrencySafe: true`
- **THEN** it SHALL return `true`

#### Scenario: Ineligible tool missing ReadOnly
- **WHEN** `IsEligible` is called with a tool that has `ReadOnly: false`
- **THEN** it SHALL return `false`

#### Scenario: Nil tool returns false
- **WHEN** `IsEligible(nil)` is called
- **THEN** it SHALL return `false`

### Requirement: Constructor clamps concurrency to minimum 1

`NewParallelReadOnlyExecutor(n)` SHALL create an executor with `maxConcurrency` clamped to a minimum of 1, regardless of the input value.

#### Scenario: Zero concurrency clamped to 1
- **WHEN** `NewParallelReadOnlyExecutor(0)` is called
- **THEN** the executor's `maxConcurrency` SHALL be 1

#### Scenario: Negative concurrency clamped to 1
- **WHEN** `NewParallelReadOnlyExecutor(-5)` is called
- **THEN** the executor's `maxConcurrency` SHALL be 1

### Requirement: ExecuteParallel preserves result ordering

`ExecuteParallel` SHALL return results in the same positional order as the input invocations slice, regardless of completion order.

#### Scenario: Results match input order
- **WHEN** `ExecuteParallel` is called with invocations [A, B, C]
- **THEN** the result slice SHALL contain results for [A, B, C] in that exact order

### Requirement: Eligible tools execute concurrently up to maxConcurrency

Eligible tools SHALL execute concurrently, bounded by the executor's `maxConcurrency` limit. No more than `maxConcurrency` tool handlers SHALL run simultaneously.

#### Scenario: Concurrent execution within limit
- **WHEN** `ExecuteParallel` is called with 5 eligible invocations and `maxConcurrency` is 3
- **THEN** at most 3 tool handlers SHALL execute simultaneously

### Requirement: Non-eligible tools receive error without handler execution

When a non-eligible tool is included in the invocations, its corresponding `ToolResult` SHALL contain an error and the tool's handler SHALL NOT be executed.

#### Scenario: Non-eligible tool error result
- **WHEN** `ExecuteParallel` is called with a non-eligible tool
- **THEN** the corresponding `ToolResult.Error` SHALL be non-nil
- **AND** the tool's handler SHALL not be invoked

### Requirement: Handler errors are captured per-result

Handler errors SHALL be captured in the individual `ToolResult.Error` field and SHALL NOT propagate to the group or cancel other tool executions.

#### Scenario: One handler fails others succeed
- **WHEN** one tool handler returns an error during `ExecuteParallel`
- **THEN** that tool's `ToolResult.Error` SHALL contain the error
- **AND** other tools' results SHALL be unaffected

### Requirement: Context cancellation stops pending invocations

When the context is cancelled during `ExecuteParallel`, pending (not yet started) tool invocations SHALL be stopped. Already-running handlers follow standard context cancellation behavior.

#### Scenario: Context cancelled mid-execution
- **WHEN** the context is cancelled while tools are executing
- **THEN** pending invocations SHALL not start
- **AND** the function SHALL return without blocking indefinitely

### Requirement: Each ToolResult includes Duration

Each `ToolResult` SHALL include a `Duration` field measuring the wall-clock time of the tool handler execution.

#### Scenario: Duration is recorded
- **WHEN** a tool handler executes for approximately 100ms
- **THEN** the `ToolResult.Duration` SHALL be approximately 100ms

### Requirement: Empty invocation slice returns empty result

When `ExecuteParallel` is called with an empty invocation slice, it SHALL return an empty result slice without error.

#### Scenario: No invocations
- **WHEN** `ExecuteParallel` is called with an empty slice
- **THEN** it SHALL return an empty `[]ToolResult`

### Requirement: Nil tool in invocation produces error at corresponding index

When an invocation contains a nil tool, the result at that index SHALL contain an error. Other invocations SHALL proceed normally.

#### Scenario: Nil tool in invocation list
- **WHEN** an invocation at index 2 has a nil tool
- **THEN** `results[2].Error` SHALL be non-nil
- **AND** results at other indices SHALL be computed normally

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
