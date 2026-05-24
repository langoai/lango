## Why

`internal/cli/extension` still keeps a local confirmation wrapper only to enforce a non-TTY guard and EOF-as-deny behavior. That logic belongs next to the rest of the shared prompt helpers so future commands do not keep re-implementing the same adapter.

## What Changes

- Add a shared prompt helper for TTY-guarded confirmation with EOF-as-deny behavior
- Replace the extension-local `promptConfirm(...)` wrapper with the shared helper
- Move the direct non-TTY helper coverage into the prompt package
- Record the guarded-confirmation helper contract in OpenSpec

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-prompt-helpers`: shared prompt helpers include a terminal-guarded confirmation wrapper with EOF-as-deny semantics
- `extension-pack-cli`: extension install/remove confirmation uses the shared terminal-guarded helper

## Impact

- Affected code: `internal/cli/prompt/*`, `internal/cli/extension/*`
- Affected specs: `openspec/specs/cli-prompt-helpers/spec.md`, `openspec/specs/extension-pack-cli/spec.md`
- No user-facing behavior change; this is a duplication and testability reduction
