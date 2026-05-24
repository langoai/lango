## Why

`internal/cli/extension` still carries its own yes/no parsing helper even though the repository now has a shared confirmation helper with stream seams and coverage. Keeping a second confirmation implementation invites drift in prompt behavior and weakens the effort to make CLI interaction rules consistent.

## What Changes

- Replace the extension-specific confirmation parser with the shared `prompt.ConfirmIO(...)` flow
- Keep the existing non-TTY safeguard for scripted installs/removals without `--yes`
- Add extension CLI regressions that verify install/remove confirmation through Cobra command streams
- Document the shared confirmation contract on the extension CLI surface

## Capabilities

### New Capabilities

### Modified Capabilities
- `extension-pack-cli`: install/remove confirmation flows use the shared confirmation helper and Cobra command streams

## Impact

- Affected code: `internal/cli/extension/*`, `internal/cli/prompt/*`
- Affected docs: `README.md`
- Affected specs: `openspec/specs/extension-pack-cli/spec.md`
- No user-facing feature expansion; this is a consistency and testability improvement
