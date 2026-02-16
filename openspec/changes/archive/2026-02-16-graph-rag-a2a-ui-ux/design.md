## Context

Graph Store, Multi-Agent, A2A, Graph RAG, and Self-Learning Graph features exist in `internal/graph/`, `internal/orchestration/`, `internal/a2a/`, `internal/learning/` but have no user-facing surface. All configuration is stored in an encrypted SQLite database (AES-256-GCM) — no plaintext config files exist during normal operation. New UI must work through this encrypted profile system.

## Goals / Non-Goals

**Goals:**
- Expose graph, agent, and A2A features via CLI commands users can run
- Add diagnostic checks so `lango doctor` validates new subsystems
- Add onboard wizard screens for interactive configuration
- Provide sensible defaults and validation for new config fields
- Document everything in README

**Non-Goals:**
- Implementing new core graph/A2A/orchestration logic (already exists)
- Adding remote agent management forms (too complex for simple TUI forms; documented as `config import/export` workflow)
- GraphQL or REST API for graph queries

## Decisions

**1. Graph Store API Extension — Add to existing interface vs new interface**
- Decision: Extend existing `Store` interface with `Count`, `PredicateStats`, `ClearAll`
- Rationale: These are fundamental store operations. A separate interface would fragment the API. All implementors (currently just BoltStore) can implement them.

**2. CLI Pattern — Follow memory CLI pattern exactly**
- Decision: Use `cfgLoader func() (*config.Config, error)` injection, `--json` flag, tabwriter output
- Rationale: Consistency with existing `lango memory` and `lango security` commands. Users learn one pattern.

**3. Doctor Checks — Warn vs fail on remote agent connectivity**
- Decision: Remote agent unreachability produces a warning, not a failure
- Rationale: Remote agents may be intentionally offline. Hard failure would block `lango doctor --fix` flows.

**4. Onboard — Simple forms, no remote agent list editor**
- Decision: Graph/Multi-Agent/A2A forms use simple fields. Remote agents require `config export` → edit → `config import`
- Rationale: A list-of-structs editor in bubbletea is significantly more complex. The existing providers list model uses a different pattern (CRUD list). Remote agents are a power-user feature.

**5. Sub-agent list — Hardcoded vs dynamic**
- Decision: Hardcode 4 local sub-agents in `lango agent list`
- Rationale: Sub-agents are defined at compile time in `orchestrator.go`. No runtime registration mechanism exists.

## Risks / Trade-offs

- **Interface extension breaks test fakes** → Mitigated by updating `fakeGraphStore` in `learning/graph_engine_test.go`. Any external consumers would also need updates.
- **BoltDB Stats().KeyN accuracy** → BoltDB bucket stats may lag behind recent writes within the same transaction. Acceptable for a diagnostic command.
- **No tests for new CLI commands** → CLI commands are thin wrappers over tested core logic. Integration tests would require bootstrap setup. Risk is low for read-only commands.
