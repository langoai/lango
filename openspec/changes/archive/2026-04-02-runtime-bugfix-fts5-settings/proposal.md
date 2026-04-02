## Why

FTS5 search fails with `syntax error near "?"` on special characters (`?`, `.`, `!`) causing fallback to LIKE (performance degradation). FTS5 bulk index fails due to wrong table name (`knowledge` vs `knowledges`) and nullable columns scanned into plain strings. Additionally, `lango settings` does not expose ontology or alerting configuration, preventing users from enabling these features through the TUI.

## What Changes

- Fix `escapeFTS5Token()` to strip all FTS5-problematic characters (`.?!@#$%&=|~<>;,[]\/`)
- Add empty-token filtering in `sanitizeFTS5Query()` across all three code paths (normal, prefix, unclosed quote)
- Fix knowledge bulk index table name from `knowledge` to `knowledges` (Ent table name)
- COALESCE nullable columns (`error_pattern`, `fix`) in learning bulk index query
- Add ontology settings form (15 fields covering ACL, governance, exchange)
- Add alerting settings form (3 fields)
- Wire state update, menu, factory, enabled checker, and dependency for both categories
- Re-synchronize downstream docs (`settings.go`, `core.md`, `cli-settings/spec.md`, `configuration.md`) with `menu.go` as source of truth

## Capabilities

### New Capabilities

- `fts5-query-safety`: Robust FTS5 query sanitization that handles all special characters and edge cases without syntax errors
- `settings-ontology-alerting`: Ontology and alerting configuration forms in `lango settings` TUI with full save-path wiring

### Modified Capabilities

- `cli-settings`: Adding ontology and alerting categories to the settings menu, and re-synchronizing category lists across docs

## Impact

- `internal/search/fts5.go` — sanitizer and query builder
- `internal/app/wiring_knowledge.go` — bulk index SQL queries
- `internal/cli/settings/` — new form files, menu, factory, editor, dependencies
- `internal/cli/tuicore/state_update.go` — config write-back keys
- `docs/cli/core.md`, `openspec/specs/cli-settings/spec.md`, `docs/configuration.md` — category list sync
