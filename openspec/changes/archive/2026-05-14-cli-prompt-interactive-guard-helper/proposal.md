## Why

Several security CLI commands still repeat the same `if !prompt.IsInteractive()` check with command-specific error strings. That duplication keeps the interactive-terminal policy outside the shared prompt layer and makes future consistency fixes harder.

## What Changes

- Add a shared prompt helper for interactive-terminal requirements
- Replace repeated `prompt.IsInteractive()` guards in security CLI commands with the shared helper
- Add prompt helper regression coverage for the new guard
- Record the helper contract in OpenSpec

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-prompt-helpers`: shared prompt helpers include an interactive-terminal guard helper

## Impact

- Affected code: `internal/cli/prompt/*`, `internal/cli/security/change_passphrase.go`, `internal/cli/security/migrate.go`, `internal/cli/security/recovery.go`, `internal/cli/security/keyring.go`, `internal/cli/security/secrets.go`
- Affected specs: `openspec/specs/cli-prompt-helpers/spec.md`
- No user-facing behavior change; this is a prompt-layer consolidation
