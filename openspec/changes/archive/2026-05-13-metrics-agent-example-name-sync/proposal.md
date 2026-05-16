## Why

The `lango metrics agents` docs example still used older built-in teammate names like `executor` and `researcher`, even though the current built-in registry uses `operator`, `librarian`, and `planner`. The metrics command prints whatever agent names the runtime emits, so the public example should reflect the current built-in set.

## What Changes

- Update the `lango metrics agents` example to use current built-in teammate names.
- Extend downstream docs sync coverage so metrics examples stay aligned with the current built-in registry.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `downstream-docs-sync`: metrics docs examples now follow the current built-in teammate names.

## Impact

- Affected docs: `docs/cli/metrics.md`
- Affected specs: `openspec/specs/downstream-docs-sync/spec.md`
