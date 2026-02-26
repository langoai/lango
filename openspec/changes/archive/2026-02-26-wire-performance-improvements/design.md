## Context

Branch `claude/improve-lango-performance-JzdXK` added core performance primitives (turn limits, error correction, token budgets) to `internal/adk/`, `internal/memory/`, and `internal/orchestration/`. These features have builder methods and tests but are not connected to the config system or application wiring layer. The orchestrator's `MaxDelegationRounds` is hardcoded to `5` in `wiring.go` despite the default changing to `10` in `orchestrator.go`.

## Goals / Non-Goals

**Goals:**
- Wire all 6 performance features from config → application layer
- Expose new config fields with zero-value-means-default semantics
- Show performance settings in `lango agent status`
- Update documentation to reflect new config options

**Non-Goals:**
- Changing the core implementation of any performance feature
- Adding new performance features beyond what already exists
- Adding TUI settings pages for these fields
- Adding validation or integration tests (existing unit tests cover the core)

## Decisions

### D1: Functional Options for Agent Construction
**Decision**: Introduce `AgentOption` functional options for `NewAgent`/`NewAgentFromADK` instead of chaining builder methods after construction.

**Rationale**: Builder methods (`WithMaxTurns`, `WithErrorFixProvider`) require the caller to set fields after construction, which means `SessionServiceAdapter.tokenBudget` cannot be set before `runner.New()`. Functional options allow all configuration to happen atomically during construction.

**Alternative considered**: Keep builder methods only — rejected because token budget must be set on `SessionServiceAdapter` before the runner is created.

### D2: Zero-Value Defaults
**Decision**: Use `0` to mean "use code default" for all integer config fields. Use `*bool` (nil = default true) for `errorCorrectionEnabled`.

**Rationale**: Follows existing patterns in the codebase (`MaxReflectionsInContext`, `MaxObservationsInContext`). The `*bool` pattern matches `ReadOnlyRootfs *bool` in the same config package.

### D3: Centralized `buildAgentOptions` Helper
**Decision**: Extract a `buildAgentOptions(cfg, kc)` function in `wiring.go` that both single-agent and multi-agent paths share.

**Rationale**: Avoids duplicating option construction logic across the two code paths. Single source of truth for how config maps to agent options.

## Risks / Trade-offs

- **[Signature change]** `NewAgent`/`NewAgentFromADK` gain `opts ...AgentOption` variadic parameter. This is backward compatible (existing callers pass no options). → No migration needed.
- **[Config field proliferation]** 5 new fields added to config. → Mitigated by zero-value defaults; existing configs work unchanged.
- **[Error correction default]** Error correction defaults to `true` when knowledge system is available. Users who don't want it must explicitly set `errorCorrectionEnabled: false`. → Acceptable because the feature is beneficial by default and has no cost when no fix is found.
