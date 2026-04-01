## Context

PolicyDecisionEvent is already defined in `internal/eventbus/events.go` and published by the `policyBusAdapter` in `internal/app/app.go`. However, no subsystem subscribes to this event. The audit recorder, metrics collector, and turntrace modules are unaware of policy decisions. This change wires subscribers to make these decisions observable.

The existing observability infrastructure follows a consistent pattern: event bus subscriptions in `wiring_observability.go`, metrics aggregation in `collector.go`, HTTP endpoints in `routes_observability.go`, and CLI subcommands in `internal/cli/metrics/`.

## Goals / Non-Goals

**Goals:**
- Make policy decisions (observe/block) visible in audit logs, metrics, and turn traces
- Follow existing subscription and aggregation patterns exactly
- Provide CLI access to policy metrics via `lango metrics policy`

**Non-Goals:**
- Alerting or notification on policy decisions
- Historical persistence of policy metrics (in-memory only, like existing tool/token metrics)
- Policy decision replay or forensic analysis
- Changes to the PolicyEvaluator itself

## Decisions

1. **Audit action enum value**: Add `"policy_decision"` to the AuditLog schema enum. This follows the existing pattern where each event type gets its own action value. `go generate` will be deferred to a coordinator step.

2. **Metrics aggregation approach**: Use simple counters (`policyBlocks`, `policyObserves`) plus a `policyByReason` map for reason-code breakdown. This mirrors the existing `toolExecs` + `tools map` pattern in `collector.go`. In-memory only — no persistent history needed for v1.

3. **Turntrace constant only**: Add `EventPolicyDecision` constant to `turntrace/events.go` for future trace timeline integration. The actual trace recording wiring is out of scope (matches the existing pattern where constants are defined ahead of full integration).

4. **HTTP endpoint placement**: Add `/metrics/policy` alongside existing `/metrics/sessions`, `/metrics/tools`, `/metrics/agents` in `routes_observability.go`.

5. **CLI subcommand**: New `internal/cli/metrics/policy.go` file following the exact same pattern as tools/sessions/agents subcommands.

## Risks / Trade-offs

- [Risk] Adding enum value without `go generate` causes compile errors → Mitigation: The ent-generated code is handled by the coordinator after all parallel changes land. The schema change is additive-only.
- [Risk] High-frequency policy events could grow the `policyByReason` map unboundedly → Mitigation: Reason codes are a fixed set defined by the PolicyEvaluator (ReasonLangoCLI, ReasonSkillImport, etc.), so the map is naturally bounded.
