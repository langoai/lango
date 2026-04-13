## 1. Schema and Constants

- [x] 1.1 Add `"policy_decision"` to action enum Values in `internal/ent/schema/audit_log.go`
- [x] 1.2 Add `EventPolicyDecision EventType = "policy_decision"` constant to `internal/turntrace/events.go`

## 2. Metrics Collector

- [x] 2.1 Add `PolicyMetrics` struct to `internal/observability/types.go` with Blocks, Observes, ByReason fields
- [x] 2.2 Add `PolicyMetrics` field to `SystemSnapshot` in `internal/observability/types.go`
- [x] 2.3 Add policy counters (`policyBlocks`, `policyObserves`, `policyByReason`) to MetricsCollector in `internal/observability/collector.go`
- [x] 2.4 Add `RecordPolicyDecision(verdict, reason string)` method to MetricsCollector
- [x] 2.5 Update `Snapshot()` to include PolicyMetrics in the returned snapshot
- [x] 2.6 Update `Reset()` to clear policy counters
- [x] 2.7 Add `NewCollector()` initialization for `policyByReason` map

## 3. Audit Recorder

- [x] 3.1 Add `handlePolicyDecision` handler method to `internal/observability/audit/recorder.go`
- [x] 3.2 Add `SubscribeTyped[eventbus.PolicyDecisionEvent]` call in `Subscribe()` method

## 4. HTTP and CLI

- [x] 4.1 Add `/metrics/policy` GET endpoint to `internal/app/routes_observability.go`
- [x] 4.2 Create `internal/cli/metrics/policy.go` with `newPolicyCmd()` subcommand
- [x] 4.3 Register policy subcommand in `internal/cli/metrics/metrics.go`

## 5. Wiring

- [x] 5.1 Subscribe to `PolicyDecisionEvent` in `initObservability` in `internal/app/wiring_observability.go`

## 6. Tests

- [x] 6.1 Add test for `RecordPolicyDecision` in `internal/observability/collector_test.go`
- [x] 6.2 Add test for `EventPolicyDecision` constant in `internal/turntrace/events_test.go`
