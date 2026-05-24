## Why

`fs_write` and `fs_edit` declared required wrapper inputs in their tool schemas, but the handlers still accepted missing values by falling back to empty content or zero line numbers. That weakens the operator-facing contract and makes invalid edit requests reach lower-level file mutation code unnecessarily.

## What Changes

- Tighten `fs_write` and `fs_edit` to enforce their declared required inputs at the wrapper layer.
- Add regression coverage for missing `content` and missing edit line-range inputs.
- Sync filesystem prompt/spec coverage for the wrapper guard contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `tool-filesystem`: `fs_write` and `fs_edit` now preserve actionable wrapper-level missing-parameter errors for their declared required inputs.
- `downstream-docs-sync`: filesystem prompt guidance mentions the stricter write/edit required-input contract.

## Impact

- Affected code: `internal/tools/filesystem/tools.go`
- Affected tests: `internal/tools/filesystem/filesystem_test.go`
- Affected prompts: `prompts/TOOL_USAGE.md`
- Affected specs: `openspec/specs/tool-filesystem/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
