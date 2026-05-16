## Why

`lango security keyring clear` still owns a bespoke yes/no parser even though the repository now has a shared confirmation helper with command-stream seams. That duplication makes one of the core security commands drift away from the rest of the CLI interaction contract.

## What Changes

- Replace the inline `keyring clear` confirmation parser with the shared `prompt.ConfirmIO(...)` helper
- Preserve the existing non-interactive `--force` safeguard and command-stream routing
- Add regressions for abort, confirm, force, and non-interactive guidance paths
- Clarify the command-stream confirmation contract in spec/docs

## Capabilities

### New Capabilities

### Modified Capabilities
- `passphrase-management`: `security keyring clear` confirmation uses the shared helper and Cobra command streams

## Impact

- Affected code: `internal/cli/security/keyring.go`, `internal/cli/security/security_test.go`
- Affected docs/specs: `docs/cli/security.md`, `openspec/specs/passphrase-management/spec.md`
- No feature expansion; this is a consistency and testability improvement
