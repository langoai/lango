## Why

The filesystem tool implementation already defaults `fs_list` to the current working directory when `path` is omitted, but the tool schema still marked `path` as required. That makes the declared contract drift from the actual behavior and from the natural `ls`-style UX the tool already provides.

## What Changes

- Update `fs_list` so its schema treats `path` as optional.
- Add regression coverage for the optional-path schema and default-to-current-directory behavior.
- Sync prompt/spec coverage for the optional `path` contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `tool-filesystem`: `fs_list` now has an explicit optional-path contract that matches its existing default behavior.
- `downstream-docs-sync`: prompt guidance now mentions the current-directory default.

## Impact

- Affected code: `internal/tools/filesystem/tools.go`
- Affected tests: `internal/tools/filesystem/filesystem_test.go`
- Affected prompts: `prompts/TOOL_USAGE.md`
- Affected specs: `openspec/specs/tool-filesystem/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
