## Why

The agent runtime executes all tool calls sequentially, even when models return multiple FunctionCalls in a single response and some tools are marked ReadOnly+ConcurrencySafe. Before implementing parallel execution, we need to understand where in the ADK tool execution ownership chain a parallel dispatch hook could be inserted. This is a read-only spike (E5) that produces an analysis report, not code.

## What Changes

- **Analysis document**: `internal/streamx/SPIKE_REPORT.md` mapping the complete tool execution flow from model response through ADK runner to tool handler
- Identifies 6 hook point options with feasibility/risk/effort ratings
- Documents ADK v0.6.0 constraints on concurrent tool dispatch
- Recommends best hook point for future parallel execution implementation

## Capabilities

### New Capabilities
- `adk-tool-dispatch-analysis`: Read-only spike documenting ADK tool execution ownership, hook points for parallel dispatch, and constraints

### Modified Capabilities
<!-- None — this is a read-only analysis that does not change any existing behavior or specs -->

## Impact

- No code changes — report document only
- Informs future E-series implementation (custom runner for parallel tool dispatch)
- Depends on: E3 (ParallelReadOnlyExecutor), A1 (ToolCapability flags)
