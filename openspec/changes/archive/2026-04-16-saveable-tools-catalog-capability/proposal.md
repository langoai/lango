## Why

Wave 1 (Unit B) added `DefaultSaveableTools` as a hard-coded constant in `toolchain/hook_knowledge.go`. This works but drifts when tools are added/renamed/removed. The `ToolCapability` struct already has `ReadOnly bool` and `Activity ActivityKind` fields that semantically indicate which tools produce saveable read-type results. This change replaces the hard-coded list with a catalog query so `KnowledgeSaveHook.SaveableTools` is automatically derived from tool metadata.

## What Changes

- Add `KnowledgeSaveable() bool` helper method on `ToolCapability` (true when `ReadOnly || Activity ∈ {read, query}`)
- Add `SaveableToolNames()` method on `Catalog` that returns names of tools where `KnowledgeSaveable()` is true
- Replace `DefaultSaveableTools` constant usage in `app.go` builder with `catalog.SaveableToolNames()`
- Remove or deprecate the `DefaultSaveableTools` constant (keep as fallback if catalog is nil)
- Update `lango agent hooks` output to show "catalog-derived" source instead of "constant"

## Capabilities

### New Capabilities
_(none)_

### Modified Capabilities
- `cli-agent-tools-hooks`: SaveableTools source changes from constant to catalog-derived

## Impact

- `internal/agent/capability.go` — new `KnowledgeSaveable()` method
- `internal/toolcatalog/catalog.go` — new `SaveableToolNames()` method
- `internal/app/app.go` — wire catalog-derived list into KnowledgeSaveHook
- `internal/toolchain/hook_knowledge.go` — deprecate `DefaultSaveableTools` (retain as fallback)
- `internal/cli/agent/hooks.go` — minor: indicate source is catalog-derived
