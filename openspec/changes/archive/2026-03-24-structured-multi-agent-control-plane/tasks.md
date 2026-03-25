## 1. turntrace Foundation (Phase 1)

- [x] 1.1 Create `internal/turntrace/events.go` — `type EventType = string` alias + typed constants (EventToolCall, EventToolResult, EventDelegation, EventDelegationReturn, EventText, EventTerminalError, EventBudgetWarning, EventRecoveryAttempt)
- [x] 1.2 Expand `internal/turntrace/store.go` Store interface with EventsForTrace, TracesForSession, PurgeTraces, TraceCount, OldTraces, RecentByOutcome
- [x] 1.3 Implement all new Store methods in EntStore (nil-safe pattern)
- [x] 1.4 Update `internal/turnrunner/runner.go` — replace string literals with EventType constants, add EventDelegationReturn emission
- [x] 1.5 Create `internal/turntrace/delegation.go` — DelegationEdge, DelegationGraph, AgentNode types + BuildDelegationGraph pure function
- [x] 1.6 Create `internal/turntrace/metrics.go` — AgentMetrics, AgentMetricsSummary types + ComputeAgentMetrics pure function with percentile calculation
- [x] 1.7 Create `internal/turntrace/retention.go` — RetentionCleaner lifecycle.Component with configurable maxAge/maxTraces/cleanupInterval
- [x] 1.8 Write tests: turntrace/{events,delegation,metrics,retention,store_expansion}_test.go

## 2. CLI Diagnostics (Phase 1, with turntrace)

- [x] 2.1 Update `internal/cli/agent/agent.go` — change NewAgentCmd signature to accept bootLoader for DB access
- [x] 2.2 Create `internal/cli/agent/trace.go` — `lango agent trace list` (--session, --limit, --outcome, --json) + `lango agent trace <trace-id>` (event timeline)
- [x] 2.3 Create `internal/cli/agent/graph.go` — `lango agent graph <session-key>` (delegation graph view, --json)
- [x] 2.4 Create `internal/cli/agent/tracemetrics.go` — `lango agent trace metrics` (per-agent performance table, --json, --agent)
- [x] 2.5 Update `cmd/lango/main.go` — pass bootLoader to NewAgentCmd
- [x] 2.6 Write CLI command tests

## 3. Config Schema (Phase 1)

- [x] 3.1 Add OrchestrationConfig (Mode, CircuitBreakerCfg, BudgetCfg, RecoveryCfg) to AgentConfig
- [x] 3.2 Add TraceStoreConfig (MaxAge, MaxTraces, FailedTraceMultiplier, CleanupInterval) to ObservabilityConfig
- [x] 3.3 Add default values in presets/defaults
- [x] 3.4 Write config tests (defaults, validation)

## 4. Doctor MultiAgentCheck Extension (Phase 2, after turntrace)

- [x] 4.1 Extend `internal/cli/doctor/checks/multi_agent.go` — add loop frequency check (RecentByOutcome, warn >3/24h)
- [x] 4.2 Add timeout frequency check (warn >5/24h)
- [x] 4.3 Add trace store growth check (TraceCount vs maxTraces, warn >80%)
- [x] 4.4 Add average turn duration check (warn >2min)
- [x] 4.5 Write/update multi_agent_test.go for new checks

## 5. agentrt Control Plane (Phase 3, after turntrace)

- [x] 5.1 Create `internal/agentrt/coordinating_executor.go` — CoordinatingExecutor implementing turnrunner.Executor, event hook composition via opts (not onChunk), recovery action dispatch
- [x] 5.2 Create `internal/agentrt/delegation_guard.go` — DelegationGuard with per-agent CircuitBreaker (closed/open/half-open state machine), Observe/IsOpen/RecordOutcome methods
- [x] 5.3 Create `internal/agentrt/budget.go` — BudgetPolicy (observational), RecordTurn/RecordDelegation(target)/AlertIfNeeded, inner budget mirroring semantics (hasFunctionCall && !isDelegation)
- [x] 5.4 Create `internal/agentrt/recovery.go` — RecoveryPolicy with RecoveryAction enum (None/Retry/RetryWithHint/DirectAnswer/Escalate), Decide method, addRerouteHint helper
- [x] 5.5 Create `internal/agentrt/events.go` — DelegationObservedEvent, BudgetAlertEvent, RecoveryEvent, CircuitBreakerTrippedEvent (EventBus events)
- [x] 5.6 Write tests: agentrt/{coordinating_executor,delegation_guard,budget,recovery}_test.go (table-driven)

## 6. Gateway Events + TurnRunner Callbacks (Phase 4, after turntrace)

- [x] 6.1 Extend `turnrunner.Request` with OnDelegation func(from, to, reason string) and OnBudgetWarning func(used, max int)
- [x] 6.2 Update runner.go traceRecorder to invoke OnDelegation callback on delegation events, track delegation count for OnBudgetWarning
- [x] 6.3 Update `internal/gateway/server.go` handleChatMessage to set OnDelegation/OnBudgetWarning callbacks that broadcast WebSocket events
- [x] 6.4 Write integration tests (server_test.go pattern)

## 7. App Wiring + Docs (Phase 5, after agentrt + config)

- [x] 7.1 Create `internal/app/wiring_agentrt.go` — initAgentRuntime() that wraps executor with CoordinatingExecutor in structured mode
- [x] 7.2 Update `internal/app/app.go` — call initAgentRuntime() before TurnRunner creation, register RetentionCleaner lifecycle
- [x] 7.3 Update README.md — structured orchestration mode architecture section
- [x] 7.4 Update docs/cli/core.md — add agent trace/graph/trace metrics commands (below existing TUI section)
- [x] 7.5 Update docs/cli/index.md — add agent diagnostic command rows

## 8. Verification

- [x] 8.1 `go build ./...` passes
- [x] 8.2 `go vet ./...` passes
- [x] 8.3 `go test ./... -short -timeout 120s` passes (all existing + new tests)
- [x] 8.4 Classic mode (default) — verify no behavioral change
- [x] 8.5 Structured mode — verify CoordinatingExecutor wrapping works end-to-end
