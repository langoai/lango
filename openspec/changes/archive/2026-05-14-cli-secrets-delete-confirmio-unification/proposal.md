## Why

`lango security secrets delete` still implements its own yes/no parser even though the repository now has a shared confirmation helper with command-stream seams. That leaves another security-sensitive delete flow inconsistent with the rest of the CLI.

## What Changes

- Replace the inline `secrets delete` confirmation parser with the shared `prompt.ConfirmIO(...)` helper
- Preserve the existing `--force` bypass and non-interactive refusal guidance
- Add regressions for abort, confirm, force, and non-interactive guidance
- Clarify the confirmation contract in spec/docs

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-secrets-management`: `secrets delete` confirmation uses the shared helper and Cobra command streams

## Impact

- Affected code: `internal/cli/security/secrets.go`, `internal/cli/security/security_test.go`
- Affected docs/specs: `docs/cli/security.md`, `openspec/specs/cli-secrets-management/spec.md`
- No feature expansion; this is a consistency and testability improvement
