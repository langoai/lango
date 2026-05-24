## Why

The recovery setup confirmation-word path uses the shared visible line prompt helper, which intentionally preserves partial-line-on-EOF behavior from the lower-level raw line reader. However, `confirmWord(...)` still turns that specific case into a hard read failure even when the operator already supplied the correct word.

## What Changes

- Accept a matching confirmation word when EOF follows the final line without a trailing newline
- Add a regression for the partial-line EOF success case
- Document the behavior in recovery CLI docs and recovery spec

## Capabilities

### New Capabilities

### Modified Capabilities
- `recovery-mnemonic`: confirmation-word prompts accept matching final lines without trailing newline

## Impact

- Affected code: `internal/cli/security/recovery.go`, `internal/cli/security/security_test.go`
- Affected docs: `docs/cli/security.md`
- Affected specs: `openspec/specs/recovery-mnemonic/spec.md`
- No feature expansion; this is input-path hardening for wrapper-driven recovery setup
