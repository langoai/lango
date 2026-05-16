## Why

The exec tool cluster already declares `command` or `id` as required wrapper inputs, but there was no direct regression proving that those missing-input failures happen before policy evaluation or supervisor interaction. The agent prompt and operator-facing docs also did not state that contract explicitly.

## What Changes

- Add wrapper-level regression coverage for missing `command`/`id` across the exec tool cluster.
- Sync exec prompt and operator-facing docs with the required-input contract.
- Sync exec, downstream-docs, and production-readiness specs with the wrapper guard coverage.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `tool-exec`: wrapper-level required-input guards are now directly covered.
- `downstream-docs-sync`: exec tool docs now mention the `command`/`id` wrapper contract.
- `production-readiness`: wrapper guard coverage now includes the exec tool cluster.

## Impact

- Affected tests: `internal/tools/exec/tools_test.go`
- Affected prompts: `prompts/TOOL_USAGE.md`
- Affected docs: `README.md`, `docs/cli/agent-memory.md`, `docs/features/exec-safety.md`
- Affected specs: `openspec/specs/tool-exec/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`, `openspec/specs/production-readiness/spec.md`
