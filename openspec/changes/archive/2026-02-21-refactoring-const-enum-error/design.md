## Context

The Lango codebase grew rapidly through feature additions (multi-agent, automation, graph store, etc.) without enforcing consistent patterns for constants, enums, errors, and shared types. Current state:

- Magic strings for channel types ("telegram", "discord", "slack") in 6+ files
- Provider types as raw strings across supervisor and config
- Message roles with inline normalization logic
- "failed to" error prefixes in 100+ locations (violating project go-errors.md rule)
- String matching for error handling (`strings.Contains(err.Error(), "session not found")`)
- Duplicate `EmbedCallback` definitions in `knowledge/store.go` and `memory/store.go`
- Duplicate `EstimateTokens()` in `memory/token.go` and `learning/token.go`
- Package named `cli/common` violating Go naming guidelines
- Existing typed enums (`ApprovalPolicy`, `PIICategory`, etc.) lacking `Valid()`/`Values()` methods

## Goals / Non-Goals

**Goals:**
- Establish `internal/types/` as the leaf package for shared cross-cutting types
- Define `Enum[T]` interface pattern and apply it to all typed enums
- Replace all magic strings with typed enum constants
- Replace string-based error matching with sentinel errors and `errors.Is()`
- Remove "failed to" prefix from all error messages per project rules
- Consolidate duplicate types (callbacks, token estimation, sender functions)
- Rename `cli/common` to `cli/clitypes` per Go naming guidelines

**Non-Goals:**
- Changing external API behavior or user-facing functionality
- Adding new features or capabilities
- Modifying the Ent ORM generated code
- Restructuring package hierarchy beyond what's needed for type consolidation
- Adding comprehensive error handling where none exists (only fixing format)

## Decisions

### D1: `internal/types/` as leaf package
**Decision**: Create `internal/types/` importing only stdlib (+ `context`).
**Rationale**: Avoids import cycles. Callback types exist specifically to break cycles — moving them to a shared leaf package maintains this property while eliminating duplication.
**Alternative**: Keep types in their respective packages → rejected because it doesn't solve duplication.

### D2: `Enum[T]` interface pattern
**Decision**: Define a generic `Enum[T any]` interface with `Valid() bool` and `Values() []T`. Each enum type implements this as value receiver methods.
**Rationale**: Provides a consistent contract for enum validation across the codebase. Generic parameter allows type-safe usage.
**Alternative**: Code generation with `go generate` → rejected as overkill for the current scale.

### D3: Mirror types for graph.Triple
**Decision**: Keep `types.Triple` and `graph.Triple` as mirror types (existing pattern).
**Rationale**: `graph` package has BoltDB-specific concerns. Mirror types avoid coupling `types/` to storage internals. Conversion happens at the boundary.

### D4: Sentinel errors over string matching
**Decision**: Define package-level `var ErrXxx = errors.New("...")` and use `errors.Is()`.
**Rationale**: Type-safe, refactor-friendly, and follows Go best practices. The current `strings.Contains` pattern is fragile and breaks if error messages change.

### D5: Error message format — remove "failed to"
**Decision**: Batch-replace `"failed to X: %w"` → `"X: %w"` across all files.
**Rationale**: Project rule in `go-errors.md` explicitly states: "Keep context succinct — avoid 'failed to' prefix."

### D6: Package rename cli/common → cli/clitypes
**Decision**: Rename with importer updates.
**Rationale**: Go naming guidelines prohibit "common", "util", "shared", "lib" as package names. `clitypes` accurately describes the package content.

## Risks / Trade-offs

- **[Risk] Large number of file changes** → Mitigated by phased PR strategy (6 independent PRs). Each PR is small and focused.
- **[Risk] Sentinel error behavior change** → `strings.Contains` → `errors.Is()` could miss wrapped errors. Mitigated by verifying `%w` wrapping at error creation sites.
- **[Risk] mapstructure deserialization with typed strings** → Go `type X string` works with mapstructure by default. Verified: no custom decoder needed.
- **[Trade-off] Mirror types create conversion overhead** → Acceptable; happens at package boundaries, not hot paths. Maintains clean dependency graph.
- **[Trade-off] "failed to" removal is mechanical** → Low risk but tedious. Can be validated with grep post-change.
