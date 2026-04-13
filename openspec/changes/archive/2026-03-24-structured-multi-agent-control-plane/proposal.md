## Why

Lango's multi-agent control is scattered across 3 locations, making debugging, policy changes, and observation difficult:
- Routing/judgment protocol is buried in orchestrator prompt prose (`orchestration/tools.go:606`)
- Budget extension is hardcoded as ad-hoc logic inside agent.go (`adk/agent.go:270-328`)
- Recovery strategy is scattered inline within RunAndCollect (`adk/agent.go:473-593`)

Since TUI became the default entry point (`847dcea`), TurnRunner changes immediately affect TUI/Gateway/Channel entirely, increasing the urgency to separate policies into code.

## What Changes

- Introduce `internal/agentrt/` new package: `CoordinatingExecutor` (turnrunner.Executor wrapper), `DelegationGuard` (circuit breaker), `BudgetPolicy` (observational budget mirroring), `RecoveryPolicy` (retry/reroute/direct-answer/escalation decisions)
- Extend `internal/turntrace/`: typed event constants, Store interface extension (EventsForTrace, TracesForSession, RecentByOutcome, PurgeTraces, TraceCount, OldTraces), delegation graph computation, agent metrics computation, retention cleaner
- Extend `internal/turnrunner/`: OnDelegation/OnBudgetWarning callbacks added to TurnRunner Request
- Extend `internal/config/`: OrchestrationConfig, BudgetCfg, RecoveryCfg, CircuitBreakerCfg, TraceStoreConfig
- Extend `internal/cli/agent/`: `lango agent trace list/trace <id>/graph/trace metrics` CLI commands
- Extend `internal/cli/doctor/checks/multi_agent.go`: loop frequency, timeout frequency, trace growth rate, average turn time checks
- Extend `internal/gateway/server.go`: agent.delegation, agent.budget_warning WebSocket events
- New `internal/app/wiring_agentrt.go`: structured mode wiring, RetentionCleaner lifecycle registration

What is **not done** in v1: agent.go refactoring (budget/recovery authoritative promotion is v2), per-agent direct execution, TaskQueue, Mailbox, Swarm, EventBus async

## Capabilities

### New Capabilities
- `agent-control-plane`: CoordinatingExecutor (turnrunner.Executor wrapper) + DelegationGuard + BudgetPolicy + RecoveryPolicy — multi-agent policy/observation control plane
- `turntrace-diagnostics`: typed event constants, delegation graph, agent metrics, retention cleaner, Store extension — turn diagnostic infrastructure
- `agent-cli-diagnostics`: `lango agent trace/graph/trace metrics` CLI commands — operator diagnostic surface

### Modified Capabilities
- `agent-turn-tracing`: EventsForTrace, TracesForSession, RecentByOutcome, PurgeTraces, TraceCount, OldTraces added to Store interface
- `agent-error-handling`: RecoveryPolicy captures existing inline recovery patterns as code policy (applied in wrapper without modifying agent.go)
- `multi-agent-orchestration`: DelegationGuard observes orchestrator delegation post-hoc, manages circuit breaker state
- `agent-runtime`: CoordinatingExecutor injected as turnrunner.Executor, wrapping existing execution path

## Impact

- **Code**: `internal/agentrt/` (new), `internal/turntrace/`, `internal/turnrunner/`, `internal/config/`, `internal/cli/agent/`, `internal/cli/doctor/checks/`, `internal/gateway/`, `internal/app/`
- **APIs**: OnDelegation/OnBudgetWarning callbacks added to TurnRunner.Request (backward compatible — optional fields)
- **Config**: `agent.orchestration.mode` ("classic"|"structured"), `agent.orchestration.budget.*`, `agent.orchestration.recovery.*`, `agent.orchestration.circuitBreaker.*`, `observability.traceStore.*`
- **Dependencies**: No new external dependencies (stdlib + existing internal packages only)
- **TUI**: Since TUI also uses TurnRunner after `847dcea`, policies are automatically applied in structured mode. TUI statusbar display is follow-up work
