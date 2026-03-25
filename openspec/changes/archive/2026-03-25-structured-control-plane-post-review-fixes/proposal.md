## Why

Structured control plane review found three correctness gaps that affect real runtime behavior. Structured mode can drop existing trace hooks, circuit breaker outcomes can be attributed to the wrong agent, and trace metrics can attribute non-delegated turns using current config mode instead of trace evidence.

## What Changes

- Preserve existing `RunOption` event hooks when `CoordinatingExecutor` adds its policy observer.
- Isolate `CoordinatingExecutor` mutable state per run/attempt so concurrent turns do not share delegation or budget tracking state.
- Record circuit breaker outcomes against the delegated specialist for the current attempt only, not the return-to-root transfer.
- Attribute non-delegated trace metrics from trace/event evidence instead of transport entrypoints or current config mode.
- Update tests to cover hook chaining, per-run state isolation, breaker attribution, and evidence-based metrics attribution.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `agent-control-plane`: Preserve caller-installed event hooks, isolate per-run observational state, and record breaker outcomes for the correct delegated specialist.
- `turntrace-diagnostics`: Attribute non-delegated turns from trace/event agent evidence instead of transport or current config.

## Impact

- Affected code: `internal/adk/agent.go`, `internal/agentrt/coordinating_executor.go`, `internal/agentrt/budget.go`, `internal/turntrace/metrics.go`, `internal/cli/agent/tracemetrics.go`
- Affected tests: `internal/agentrt/*_test.go`, `internal/turntrace/metrics_test.go`
- No new external dependencies
- No new config keys or CLI flags
