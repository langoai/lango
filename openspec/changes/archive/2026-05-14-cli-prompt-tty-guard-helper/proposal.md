## Why

Several CLI commands still repeat the same non-terminal stdin guard logic before prompting the user. That duplication makes the prompt layer harder to evolve consistently and leaves tiny policy differences easy to reintroduce.

## What Changes

- Add a shared prompt helper for enforcing terminal-only input with caller-supplied guidance errors
- Reuse that helper in `payment send`, `security keyring clear`, and `security secrets delete`
- Add prompt helper coverage for the shared guard path
- Record the helper contract in OpenSpec

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-prompt-helpers`: shared prompt helpers include a reusable TTY-input guard

## Impact

- Affected code: `internal/cli/prompt/*`, `internal/cli/payment/send.go`, `internal/cli/security/keyring.go`, `internal/cli/security/secrets.go`
- Affected specs: `openspec/specs/cli-prompt-helpers/spec.md`
- No user-facing behavior change; this is a duplication and consistency reduction
