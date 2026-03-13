## Context

After v0.2.0, 13 tool handler files in `internal/app/` share identical parameter extraction patterns (`params["key"].(string)` + empty check + custom error message). Escrow hub client methods repeat the same contract call boilerplate. Sentinel detectors duplicate sliding-window counter logic. Config validation uses inline maps. These patterns increase maintenance cost and inconsistency.

## Goals / Non-Goals

**Goals:**
- Eliminate duplicated parameter extraction across all tool handlers via a shared `toolparam` package
- Establish domain-specific sentinel errors for `errors.Is`/`errors.As` matching
- Centralize hardcoded string constants (config validation, contract methods, transaction types)
- Reduce boilerplate in hub client and sentinel detectors through shared helpers

**Non-Goals:**
- Changing `ToolHandler` function signatures or `Tool.Parameters` JSON schema definitions
- Introducing generics or code generation for tool parameters
- Refactoring MCP or external-facing APIs
- Modifying test infrastructure beyond what's needed for changed interfaces

## Decisions

### 1. `toolparam` as a standalone package (not embedded in `app`)
**Rationale**: Keeps tool helpers reusable across any package that handles `map[string]interface{}` params. Avoids import cycle with `internal/app/`.
**Alternative**: Helper functions inside `internal/app/` — rejected because it would make the already-large `app` package even bigger.

### 2. `ErrMissingParam` custom error type instead of sentinel variables
**Rationale**: A single type with a `Name` field covers all parameter names dynamically, avoiding dozens of `ErrMissingWorkspaceID`, `ErrMissingTaskID`, etc. Supports `errors.As()` matching.
**Alternative**: One sentinel per parameter — rejected due to combinatorial explosion.

### 3. `writeMethod`/`readMethod` helpers on HubClient (not extracted to a base type)
**Rationale**: The helpers are receiver methods on `HubClient`, keeping the pattern simple. A base type would add unnecessary abstraction for two client structs.
**Alternative**: Shared `ContractClientBase` embedded type — rejected as premature abstraction.

### 4. `windowCounter` as an embedded struct (not an interface)
**Rationale**: Two detectors share identical sliding-window logic. A concrete struct with `record(key) int` method is simpler and testable. Interface would be overengineered for two consumers.

### 5. `AlertMetadata` as a typed struct (not keeping `map[string]interface{}`)
**Rationale**: All 5 detectors write to known fields. A struct provides compile-time safety and eliminates string key typos. JSON marshaling behavior is preserved via struct tags with `omitempty`.

## Risks / Trade-offs

- **[Breaking change in error messages]** → Error wrapping format changes slightly (e.g., "deposit deal %s" → "deposit"). Mitigated by updating test assertions. No external consumers of these error strings.
- **[AlertMetadata struct rigidity]** → Adding new metadata fields requires struct modification. Mitigated by `omitempty` tags allowing sparse usage. Acceptable trade-off for type safety.
- **[toolparam adoption scope]** → Only migrating `internal/app/tools_*.go` files. Other packages using similar patterns are out of scope to limit blast radius.
