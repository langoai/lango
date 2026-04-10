## Why

3 P1 and 1 P2 issues were found during Change-1~3 code review. Tests pass but these are real bugs that break user experience at runtime: embedded settings save disables auto-enable, live config contamination means runtime state changes even on cancel, context panel renders with width=0 on first toggle, sidebar cursor desynchronized from active page.

## What Changes

- Pass `config.ContextRelatedKeys()` (dotted path) instead of `state.Dirty` (category keys) on embedded save to restore explicitKeys semantics
- Add `Config.Clone()` method, deep copy in `NewEditorForEmbedding` to isolate live config
- Propagate correct width to contextPanel on visibility toggle in `toggleContext()`
- Synchronize cursor in `SetActive(id)` to match visual active with keyboard cursor

## Capabilities

### New Capabilities
- `config-clone`: Config.Clone() deep copy method (json roundtrip)

### Modified Capabilities
- `cockpit-settings-page`: embedded save's explicitKeys changed to dotted context-related paths, editor works with config deep copy
- `cockpit-shell`: Added width propagation to context panel in toggleContext()
- `cockpit-sidebar`: Added cursor synchronization in SetActive()

## Impact

- **Modified**: `internal/config/types.go` — Clone() added
- **Modified**: `internal/cli/settings/editor.go` — explicitKeys fix + cfg.Clone() call
- **Modified**: `internal/cli/cockpit/cockpit.go` — toggleContext() width propagation
- **Modified**: `internal/cli/cockpit/sidebar/sidebar.go` — SetActive() cursor synchronization
- **Tests**: config/types_test.go, editor_embed_test.go, cockpit_test.go, sidebar_test.go
