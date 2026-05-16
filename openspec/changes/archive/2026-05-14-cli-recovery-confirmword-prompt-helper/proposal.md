## Why

`lango security recovery setup` still keeps a local line-reader implementation for confirmation-word entry even though the prompt package already centralizes other interactive CLI input paths. Keeping that last local prompt parser weakens consistency in a sensitive recovery flow.

## What Changes

- Add a small shared prompt helper for visible line-entry prompts
- Replace the local recovery confirmation-word parser with the shared helper
- Add regression coverage for the shared line prompt and the recovery confirmation-word path
- Clarify in spec/docs that the confirmation-word prompt uses the shared command-stream prompt layer

## Capabilities

### New Capabilities
- `cli-prompt-helpers`: shared visible line-entry prompts for command-stream-driven CLI interactions

### Modified Capabilities
- `recovery-mnemonic`: recovery confirmation-word input uses the shared prompt helper

## Impact

- Affected code: `internal/cli/prompt/*`, `internal/cli/security/recovery.go`, `internal/cli/security/security_test.go`
- Affected docs/specs: `docs/cli/security.md`, `openspec/specs/cli-prompt-helpers/spec.md`, `openspec/specs/recovery-mnemonic/spec.md`
- No feature expansion; this is a prompt-layer consistency and testability improvement
