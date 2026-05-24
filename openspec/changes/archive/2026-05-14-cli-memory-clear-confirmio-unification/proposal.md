## Why

`lango memory clear` still uses a local scanner-based confirmation parser even though the repository now has a shared confirmation helper with command-stream seams. That leaves another user-facing destructive action outside the shared prompt layer.

## What Changes

- Replace the inline `memory clear` confirmation parser with the shared `prompt.ConfirmIO(...)` helper
- Preserve the existing warning line, `Continue?` prompt text, and `--force` bypass
- Extend memory CLI regressions to keep the same command-stream contract
- Record the confirmation-helper usage in OpenSpec

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-agent-memory`: `lango memory clear` confirmation uses the shared helper and Cobra command streams

## Impact

- Affected code: `internal/cli/memory/clear.go`, `internal/cli/memory/memory_test.go`
- Affected docs/specs: `openspec/specs/cli-agent-memory/spec.md`
- No feature expansion; this is a prompt-layer consistency improvement
