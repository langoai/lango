## Why

Several read-only tool surfaces promise that omitting `session_key` falls back to the current request session. That behavior already exists in the implementation for observational-memory and librarian inquiry tools, but there was no direct tool-entrypoint regression proving it, and the main specs did not capture the fallback explicitly.

## What Changes

- Add direct tool-entrypoint regressions for `memory_list_observations`, `memory_list_reflections`, and `librarian_pending_inquiries` using the current session when `session_key` is omitted.
- Update observational-memory and proactive-librarian specs to describe the current-session fallback contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `observational-memory`: tool-entrypoint session fallback is now directly covered.
- `proactive-librarian`: pending inquiry tool session fallback is now directly covered.

## Impact

- Affected tests: `internal/memory/tools_test.go`, `internal/librarian/tools_test.go`
- Affected specs: `openspec/specs/observational-memory/spec.md`, `openspec/specs/proactive-librarian/spec.md`
