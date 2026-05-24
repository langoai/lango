## Why

`lango graph clear` still uses a local scanner-based confirmation parser even though the repository now routes destructive confirmations through the shared prompt helper. That leaves the graph management surface behind the rest of the CLI consistency work.

## What Changes

- Replace the inline `graph clear` confirmation parser with the shared `prompt.ConfirmIO(...)` helper
- Preserve the existing warning line, `Continue?` prompt text, and `--force` bypass
- Keep graph clear regressions on the command-stream contract
- Record the shared helper usage in OpenSpec

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-graph-management`: `lango graph clear` confirmation uses the shared helper and Cobra command streams

## Impact

- Affected code: `internal/cli/graph/clear.go`, `internal/cli/graph/graph_test.go`
- Affected specs/docs: `openspec/specs/cli-graph-management/spec.md`, `docs/cli/agent-memory.md`
- No feature expansion; this is a prompt-layer consistency improvement
