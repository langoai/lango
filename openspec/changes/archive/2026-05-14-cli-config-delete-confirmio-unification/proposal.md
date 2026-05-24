## Why

`lango config delete` still implements its own prompt rendering and input parsing even though the repository now has a shared confirmation helper with command-stream seams. Keeping that bespoke path makes the config CLI inconsistent with the rest of the stream-hardening work.

## What Changes

- Replace the inline delete confirmation parser in `configcmd` with the shared `prompt.ConfirmIO(...)` helper
- Add config profile command regressions for approve, deny, and `--force` paths
- Clarify in specs/docs that delete confirmation is driven through Cobra command streams

## Capabilities

### New Capabilities

### Modified Capabilities
- `config-cli-commands`: config delete confirmation uses the shared helper and Cobra command streams

## Impact

- Affected code: `internal/cli/configcmd/profile.go`, `internal/cli/configcmd/profile_test.go`
- Affected docs/specs: `docs/cli/config.md`, `openspec/specs/config-cli-commands/spec.md`
- No feature expansion; this is a consistency and testability improvement
