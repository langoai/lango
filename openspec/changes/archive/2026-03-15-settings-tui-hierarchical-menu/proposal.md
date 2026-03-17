## Why

The Settings TUI currently displays all 49 categories in a flat list grouped by 7 section headers. On smaller terminals, the list gets clipped at top/bottom, making some categories inaccessible. A two-level hierarchical menu solves this by showing only 7 sections at Level 1 (always fits), then drilling into categories at Level 2.

## What Changes

- Replace flat grouped list with two-level hierarchical navigation in the Settings menu
- Level 1 shows 7 sections + Save/Cancel (max 9 items, fits any terminal)
- Level 2 shows categories within a selected section with Basic/Advanced tab filtering
- Esc at Level 2 returns to Level 1 (not Welcome screen)
- Tab key only active at Level 2 (no-op at Level 1)
- Breadcrumb shows section name at Level 2 (e.g., "Settings > Core")
- Search (/) works globally from both levels
- Welcome screen tip updated to reflect new navigation flow

## Capabilities

### New Capabilities

- `settings-hierarchical-menu`: Two-level navigation for the Settings TUI menu with section list (Level 1) and category detail (Level 2) views

### Modified Capabilities

- `cli-settings`: Menu navigation flow changes from flat list to hierarchical, affecting Esc/Tab/Enter key behavior and View rendering

## Impact

- `internal/cli/settings/menu.go` — Primary change: new navigation state, level-aware rendering, Enter/Esc/Tab dispatch
- `internal/cli/settings/editor.go` — Minor: Esc guard for Level 2, breadcrumb at Level 2, welcome tip text
- `internal/cli/settings/editor_test.go` — New tests for hierarchical navigation
- No API, dependency, or config changes
