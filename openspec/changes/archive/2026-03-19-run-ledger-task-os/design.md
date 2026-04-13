## Context

Lango is a Go-based multi-agent AI platform. Long-running tasks routinely fail because the agent self-certifies completion with no system-level verification. The existing workflow/background systems track execution state but have no append-only journal, no typed validators, and no policy-driven recovery. The orchestrator acts as a simple router rather than a policy supervisor.

This change introduces a RunLedger — a durable execution engine that records all state transitions in an immutable journal and validates every step result through a Propose-Evidence-Verify (PEV) protocol.

## Goals / Non-Goals

**Goals:**
- Establish an append-only journal as the single source of truth for run state
- Enforce that execution agents can only propose results, never certify completion
- Provide 6 typed validators with no custom/auto-pass escape hatch
- Enable the orchestrator to act as a policy supervisor with topology-change authority
- Support opt-in resume for paused runs (no automatic resurrection)
- Provide git worktree isolation for coding steps (fail-closed)
- Roll out progressively: shadow → write-through → authoritative read → projection retired

**Non-Goals:**
- Ent-backed persistent store (Phase 1 uses in-memory; Ent store is Phase 2)
- Orchestrator prompt changes (Phase 2)
- Command Context injection into LLM context (Phase 2)
- CLI commands for run management (Phase 4)
- Projection drift detection automation (Phase 4)

## Decisions

### 1. Append-Only Journal as Source of Truth

**Decision**: All state changes are recorded as immutable `JournalEvent` records. `RunSnapshot` is a cached projection derived by replaying the journal.

**Alternatives considered**:
- Direct mutable state in database: Simpler writes but no audit trail, no replay capability, harder to debug
- Event sourcing with separate read model: Overkill for current scale, added infrastructure complexity

**Rationale**: Append-only journal provides complete audit trail, deterministic replay for debugging, and natural cache invalidation via sequence numbers. Tail-replay optimization (only replay events after last cached seq) avoids full replay on every read.

### 2. PEV Protocol Instead of Self-Certification

**Decision**: Execution agents call `run_propose_step_result` (status → `verify_pending`). The PEV engine runs the typed validator. Only a passing validator transitions to `completed`.

**Alternatives considered**:
- Trust agent output: Current behavior — root cause of failures
- Post-hoc audit: Allows completion first, checks later — defeats the purpose
- Human-in-the-loop for every step: Too slow for most tasks

**Rationale**: PEV ensures "the system proves completion" without blocking on human approval for automatable checks (build_pass, test_pass). The `orchestrator_approval` type requires explicit human/orchestrator sign-off for non-automatable steps.

### 3. Typed Validators Only (No Custom)

**Decision**: 6 built-in validator types. No `custom` type. `orchestrator_approval` always returns failed — requires explicit `run_approve_step`.

**Alternatives considered**:
- Allow custom validators: Risk of auto-pass `func() { return true }` defeating the system
- LLM-as-judge: Non-deterministic, expensive, gameable

**Rationale**: The constraint "no auto-pass" is a core invariant. Every validator must be deterministic and verifiable. If a new validation strategy is needed, it should be added as a named built-in type with reviewed implementation.

### 4. Kahn's Algorithm for Dependency Validation

**Decision**: Step dependencies form a DAG. Validation uses Kahn's algorithm (topological sort) to detect cycles at plan creation time.

**Alternatives considered**:
- DFS-based cycle detection: Equivalent correctness but Kahn's also produces execution order
- No validation, runtime detection: Late failure is worse than early rejection

**Rationale**: Kahn's algorithm is O(V+E), provides both cycle detection and topological ordering, and matches the existing `internal/appinit/topo_sort.go` pattern in the codebase.

### 5. In-Memory Store for Phase 1

**Decision**: Phase 1 uses `MemoryStore`. Ent schemas are defined but the Ent-backed implementation is deferred to Phase 2.

**Alternatives considered**:
- Ent store immediately: More work upfront, blocks Phase 1 completion
- File-based store: Middle ground but adds serialization complexity

**Rationale**: MemoryStore enables full integration testing and shadow-mode deployment. The `RunLedgerStore` interface ensures the Ent implementation is a drop-in replacement.

### 6. 4-Stage Progressive Rollout

**Decision**: Shadow → Write-Through → Authoritative Read → Projection Retired.

**Alternatives considered**:
- Big-bang cutover: High risk, no fallback
- Feature flag per-run: Complex routing logic

**Rationale**: Progressive rollout minimizes blast radius. Shadow mode validates journaling without affecting existing systems. Each stage is independently reversible by changing config flags.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| MemoryStore data loss on restart | Phase 1 is shadow-only; real persistence comes in Phase 2 with Ent store |
| Validator execution timeout | `validatorTimeout` config (default: 2m); validators run with `context.WithTimeout` |
| Malformed planner JSON | 2-retry policy with error feedback; fallback to orchestrator-generated abbreviated plan |
| Git worktree creation failure | Fail-closed: step is aborted, not run on base tree; PolicyRequest escalates to orchestrator |
| Journal replay performance at scale | Snapshot caching with tail-replay avoids full replay; Ent store will add indexed queries |
| Access control bypass | Role check is first operation in tool handler; orchestrator name matching is explicit |

## Open Questions

1. **Ent store migration**: How to migrate from MemoryStore to EntStore without losing in-flight runs? → Proposal: drain all runs to terminal state before switching.
2. **Multi-session runs**: Should a run span multiple sessions? → Current design: one session per run, resume creates continuation in same session.
