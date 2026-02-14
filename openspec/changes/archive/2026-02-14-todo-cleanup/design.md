## Context

The codebase has 4 TODO comments remaining from earlier development phases. Code review confirmed none require new features — they are either dead code, misleading comments, or minor constant extractions.

## Goals / Non-Goals

**Goals:**
- Remove dead code (`UpdateField`) that has no callers
- Replace misleading TODO comments with accurate descriptions of current behavior
- Extract magic number (30s timeout) into a named constant for clarity

**Non-Goals:**
- Making the RPC timeout configurable at runtime (current hardcoded value is appropriate)
- Implementing ListModels proxying (not needed yet)
- Adding new save logic to the wizard (already handled by caller)

## Decisions

1. **Delete `UpdateField` rather than implement it**: The method body only calls `MarkDirty` and the real update paths (`UpdateConfigFromForm`, `UpdateProviderFromForm`) are already wired. No callers exist. Implementing reflection-based updates adds complexity with no benefit.

2. **Extract `rpcTimeout` as a package-level constant**: A named constant at package scope makes the value discoverable and documents intent, without adding the complexity of runtime configuration that isn't currently needed.

3. **Replace TODOs with descriptive comments**: Each TODO is replaced with a comment that accurately describes the current behavior and, where applicable, what would trigger future implementation.

## Risks / Trade-offs

- **[Minimal]** Removing `UpdateField` is a public API deletion → No external consumers exist; the method was only in the internal `onboard` package with zero callers.
