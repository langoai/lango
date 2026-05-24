## Why

Production-readiness work cannot start from a red baseline. The repository had three avoidable failures: clock-sensitive tests that depended on historical calendar dates, Mission Control threshold assertions that did not pin the projector clock, and a parallel execution result that could report a zero duration for a completed failing handler.

## What Changes

- Harden clock-sensitive verification so proposal-lifecycle and Mission Control agenda tests use execution-relative or explicitly injected clocks.
- Guarantee that `ParallelReadOnlyExecutor` records a positive duration for every completed eligible invocation, including handler failures.
- Re-run the repository build and test baseline after the hardening change to restore confidence before larger UX and architecture work.

## Capabilities

### New Capabilities
- `quality-verification-baseline`: Covers deterministic clock-sensitive verification and positive-duration observability for completed parallel tool invocations.

### Modified Capabilities
- None.

## Impact

- Affected code: `internal/streamx/parallel_executor.go`
- Affected verification: `internal/app/modules_test.go`, `internal/cli/cockpit/missioncontrol_projector_test.go`, `internal/streamx/parallel_executor_test.go`
- No public CLI, TUI, API, or documentation surface changes
