## Context

Multi-agent control is scattered across 3 locations (orchestration prompt, agent.go budget, agent.go recovery). Since TUI became the default entry point (`847dcea`), TurnRunner changes affect TUI/Gateway/Channel entirely. Currently the app has only a single root `*adk.Agent`, injected as a single executor into TurnRunner.

Base commit: `847dcea` (dev), `app.New(boot, ...AppOption)` pattern, `AppModeLocalChat`, `lifecycle.SetMaxPriority()` introduced.

## Goals / Non-Goals

**Goals:**
- Place a policy/observation wrapper (`CoordinatingExecutor`) on top of `turnrunner.Executor` to separate delegation monitoring, budget mirroring, and recovery decisions into code
- Extend turntrace diagnostic infrastructure (typed events, delegation graph, metrics, retention)
- Provide CLI diagnostic surface (`lango agent trace/graph/trace metrics`)
- Strengthen doctor health checks (loop/timeout frequency, trace growth rate)
- Gateway WebSocket events (agent.delegation, agent.budget_warning)
- Externalize policy parameters via config

**Non-Goals:**
- agent.go refactoring (budget/recovery authoritative promotion is v2)
- Per-agent direct execution (only via root orchestrator)
- TaskQueue, Mailbox, Swarm, Pipeline patterns
- EventBus async/priority introduction
- Real-time delegation/budget display in TUI statusbar (only callback infrastructure provided)
- Prompt routing reduction (in v1 root orchestrator LLM retains routing ownership)

## Decisions

### D1: CoordinatingExecutor implements turnrunner.Executor

**Choice:** `CoordinatingExecutor` is a wrapper implementing the `turnrunner.Executor` interface (`RunStreamingDetailed`).
**Alternative:** Custom `Coordinate(sessionID, input) (string, error)` interface → TurnRunner loses chunk streaming, onEvent trace hook, idle timeout.
**Rationale:** TurnRunner is the sole turn boundary. To avoid breaking the existing streaming/tracing pipeline, it must implement the same interface.

### D2: DelegationGuard is a post-hoc observer (does not own routing)

**Choice:** `DelegationGuard` observes delegation via ADK event hooks and manages circuit breaker state. Routing decisions remain owned by the root orchestrator LLM.
**Alternative:** `StructuredRouter` selects agents proactively → requires per-agent direct execution (impossible since the app has only one root agent).
**Rationale:** Achievable scope in v1. Honest naming (`DelegationGuard`) constrains the role.

### D3: BudgetPolicy is observational (not authoritative)

**Choice:** The inner executor's (agent.go) hardcoded budget is authoritative. BudgetPolicy mirrors turn/delegation from event hooks and only issues threshold notifications.
**Alternative:** Modify agent.go to delegate budget logic to external policy → exceeds v1 scope, regression risk.
**Rationale:** v1 is the observation layer. agent.go modification is v2.
**Mirroring rule:** Same criteria as inner budget — only `hasFunctionCall(event)` && `!isDelegation(event)` counts as a turn (ref agent.go:350). `RecordDelegation(target string)` tracks uniqueAgents.

### D4: RecoveryPolicy has substantive control

**Choice:** On inner executor failure, `RecoveryPolicy.Decide()` determines retry/hint-retry/direct-answer/escalation. Since it decides whether to re-invoke the inner executor from outside, it has substantive control even in v1.
**Actions:** `RecoveryRetry` (retry with same input), `RecoveryRetryWithHint` (retry with "try a different specialist" hint added to root), `RecoveryDirectAnswer` (use partial result), `RecoveryEscalate` (return error).
**`RecoveryRetryWithHint` is not a reroute:** It retries with hint-augmented input to the root orchestrator to encourage a different selection.

### D5: turntrace Store extension includes doctor requirements

**Choice:** Add `RecentByOutcome(ctx, outcome, since, limit)` to support doctor's time-window + outcome-filter queries.
**Alternative:** Doctor uses Ent client directly → breaks Store interface abstraction.
**Rationale:** Consistent access through Store interface.

### D6: Event hook composition via opts, not onChunk wrapping

**Choice:** Compose policy hooks into opts using `adk.WithOnEvent()`. Delegation events are only visible in ADK event hooks (not in onChunk).
**Rationale:** TurnRunner's traceRecorder also works via `adk.WithOnEvent()` (ref runner.go:227). Same pattern.

### D7: CoordinatingExecutor is not a lifecycle component

**Choice:** Injected via executor wrapping (`initAgentRuntime` returns the executor, which is passed to TurnRunner). Not registered in lifecycle.Registry.
**Rationale:** Must be independent of lifecycle priority limits (LocalChat's `SetMaxPriority(PriorityBuffer)`). Only RetentionCleaner is registered as a lifecycle component.

## Risks / Trade-offs

- **[Risk] BudgetPolicy mirroring error** — Event hook timing and inner budget counting may not be precisely synchronized → **Mitigation:** Apply same criteria as inner budget (hasFunctionCall && !isDelegation). Notifications are advisory; enforcement is left to inner.
- **[Risk] RecoveryRetryWithHint infinite loop** — Root orchestrator keeps selecting the same specialist → **Mitigation:** maxRetries (default 2) limit. Include failed agent in excludeAgents hint.
- **[Risk] runner.go modified concurrently by Unit 1 (event constants) and Unit 5 (callback)** → **Mitigation:** Phase separation (Unit 1 Phase 1, Unit 5 Phase 4). Modification locations differ (constant replacement vs callback addition).
- **[Risk] OnDelegation/OnBudgetWarning not set in TUI results in missed events** → **Mitigation:** Callbacks are optional (nil means no-op). TUI display is explicitly stated as follow-up work.
- **[Trade-off] No budget enforcement in v1** — Only observation provided, actual control is in inner executor → Resolved in v2 with agent.go refactoring.
