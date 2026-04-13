## Context

lango uses `google.golang.org/adk v0.6.0` to drive the ADK-based agent runtime. 34 Go files import 7 ADK sub-packages, with core integration points concentrated in the `internal/adk/` package (ModelAdapter, SessionServiceAdapter, Agent wrapper, tool adapters).

v1.0.0 has been released as GA, and after diffing the actual Go source from the module cache, all interfaces used by lango are identical at the source level or have only additive changes (variadic parameters, new optional struct fields).

## Goals / Non-Goals

**Goals:**
- Upgrade ADK dependency from v0.6.0 to v1.0.0 GA
- Maintain passing build, vet, and test suite
- Fix type references in MCP spike test

**Non-Goals:**
- Adopting v1.0.0 new features (AutoCreateSession, HITL, workflow agents, RunOption, etc.) — separate change
- ADK adapter refactoring or architecture changes
- Production code changes

## Decisions

### D1: Single go.mod bump strategy

Directly change the ADK version in go.mod and resolve transitive dependencies with `go mod tidy`.

**Rationale**: All public interfaces are source-compatible, so incremental migration is unnecessary. A single change produces the smallest diff and minimizes review burden.

**Alternative**: Gradual upgrade through intermediate versions (v0.7.0, etc.) — impossible since no intermediate versions exist.

### D2: MCP spike test type reference fix

Change `mcptoolset.ConfirmationProvider` → `tool.ConfirmationProvider`.

**Rationale**: In v1.0.0, the `ConfirmationProvider` type was moved from the `tool/mcptoolset` package to the `tool` package. Only used in spike tests with no production impact.

### D3: No production code changes

No changes to production code (`internal/adk/*.go`, `internal/orchestration/`, `internal/a2a/`, etc.).

**Rationale**: Actual diff verification results:
- `session.Service` — identical
- `model.LLM` — identical
- `runner.Runner.Run()` — variadic `opts ...RunOption` added (existing calls work as-is)
- `runner.Config` — `AutoCreateSession` field added (zero value = false)
- `agent.Agent` — identical
- `tool.Tool` / `functiontool.Config` — identical
- `plugin.Config` — identical

## Risks / Trade-offs

- **Transitive dependency conflicts** → `go mod tidy` resolves automatically. Only minor bumps occur: `grpc v1.78.0→v1.79.3`, `a2a-go v0.3.3→v0.3.10`, etc.
- **ADK internal behavior changes** → Covered by golden tests, session tests, model adapter tests. Verified with full test suite passing.
- **MCP spike test semantic changes** → Spike tests are unrelated to production. Only type reference changed, logic identical.
