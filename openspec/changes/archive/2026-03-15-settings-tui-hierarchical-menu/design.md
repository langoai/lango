## Context

The Settings TUI menu renders all 49 categories as a flat list with 7 section headers. On terminals with fewer than ~60 rows, bottom categories are clipped and unreachable. The menu needs to fit within any reasonable terminal height while preserving search and filtering.

## Goals / Non-Goals

**Goals:**
- Level 1 shows max 9 items (7 sections + Save + Cancel) — fits any terminal
- Level 2 shows only the categories within one section (max 12 items)
- Preserve all existing functionality: search, smart filters, save/cancel
- Minimal changes to editor.go (menu owns its navigation internally)

**Non-Goals:**
- Changing form rendering or category definitions
- Adding new categories or sections
- Persisting navigation state across sessions
- Collapsible/expandable sections within a single view

## Decisions

### Two-level state machine internal to MenuModel

Navigation levels (`levelSections`, `levelCategories`) are tracked inside `MenuModel`. The `Selected` field remains the sole output to `editor.go`. This keeps the hierarchical logic encapsulated — editor.go only needs two new guards: `InCategoryLevel()` for Esc routing and `ActiveSectionTitle()` for breadcrumbs.

**Alternative considered**: Separate models for each level. Rejected because cursor state, search, and filters need to be shared.

### Synthetic section items via `__section_` ID prefix

Level 1 items reuse the existing `Category` type with synthetic IDs (`__section_0`, `__section_1`, etc.). On Enter, the prefix is detected and parsed to transition to Level 2 instead of setting `Selected`. Save/Cancel items have real IDs and flow through normally.

**Alternative considered**: A separate `sectionItem` type. Rejected to avoid duplicating cursor/rendering logic.

### Tab restricted to Level 2

Tab toggles Basic/Advanced at Level 2 only. At Level 1, section counts always show total categories regardless of filter. This prevents confusion where toggling at Level 1 would have no visible effect.

### Cursor restoration on Esc

A `sectionCursor` field stores the Level 1 cursor position when entering Level 2. On Esc back to Level 1, the cursor returns to the section the user came from, maintaining spatial context.

## Risks / Trade-offs

- [Synthetic IDs] `__section_` prefix is a convention, not enforced by types → mitigated by encapsulation within MenuModel, no external code uses these IDs
- [Search from Level 2] After selecting a search result from a different section, Esc from the form returns to the original Level 2 section → acceptable UX, consistent with spatial navigation
