## Context

4 runtime bugs were found during Change-1~3 code review. All in areas not covered by tests — shared pointer mutation, semantic mismatch, lifecycle timing.

## Goals / Non-Goals

**Goals:**
- Restore explicitKeys semantics (embedded save and standalone save behave identically)
- Live config isolation (embedded editor does not contaminate runtime config)
- Correct width rendering on first context panel toggle
- Synchronize sidebar cursor with active page

**Non-Goals:**
- Narrow terminal min-width clamp (separate change)
- Editor width-aware rendering (already deferred in Change-3)

## Decisions

### D1: explicitKeys → context-related dotted paths
Instead of passing `state.Dirty` (category keys) from embedded save, pass `config.ContextRelatedKeys()` as explicit keys, same as standalone. This ensures auto-enable respects the user's explicit settings.

### D2: Config.Clone() via JSON roundtrip
Deep copy using `encoding/json` Marshal/Unmarshal. Correctness over performance (settings save is not a hot path). Includes nil guard.

### D3: Clone location is NewEditorForEmbedding
Clone at the Editor constructor, not SettingsPage. Reason: Editor is the config ownership boundary. All embedded use cases are automatically protected.

### D4: Propagate width to panel in toggleContext()
`propagateResize()` only updates child+pages. contextPanel is not a page, so it requires separate handling. On toggle ON, call `contextPanel.Update(WindowSizeMsg{Width: cpw})`.

### D5: Synchronize cursor in SetActive()
In SetActive(id), iterate items to find the matching index and assign it to cursor. Minimal change for maximum safety.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| JSON roundtrip may alter some zero-value fields | Config's json tags use omitempty only for slices/maps. Primitive types are safe. Verified with Clone tests. |
| explicitKeys change may subtly differ from standalone save | Guaranteed since both use the same `config.ContextRelatedKeys()` source |
