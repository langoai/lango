## Why

`p2p_workspace_create` already behaves as if `goal` is optional, and the public CLI/prompt docs describe it that way. But the tool schema still marked `goal` as required, which makes the declared contract drift from both the implementation and the operator guidance.

## What Changes

- Update the `p2p_workspace_create` tool schema so only `name` is required.
- Add regressions that pin the optional-goal schema and behavior.
- Sync the workspace spec coverage for the optional `goal` contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `p2p-workspace`: workspace creation now has an explicit optional-goal tool contract that matches the existing runtime behavior and docs.

## Impact

- Affected code: `internal/app/tools_workspace.go`
- Affected tests: `internal/app/tools_workspace_test.go`
- Affected specs: `openspec/specs/p2p-workspace/spec.md`
