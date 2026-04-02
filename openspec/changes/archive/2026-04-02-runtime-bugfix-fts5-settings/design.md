## Context

FTS5 full-text search has three runtime bugs: the sanitizer misses common punctuation characters causing syntax errors, the knowledge bulk-index query uses the wrong table name, and the learning bulk-index scans nullable columns into plain strings. Separately, the `lango settings` TUI lacks ontology and alerting categories — both config structs exist but have no forms, state-update mappings, or menu entries.

## Goals / Non-Goals

**Goals:**
- FTS5 queries never produce syntax errors regardless of input characters
- FTS5 bulk index succeeds on existing databases with nullable learning fields and correct table names
- Ontology and alerting settings are fully editable and persistable via `lango settings`
- Downstream docs and specs are synchronized with the actual `menu.go` category list

**Non-Goals:**
- Changing FTS5 tokenizer configuration or adding new search features
- Modifying ontology or alerting runtime behavior (wiring_ontology.go, wiring_observability.go)
- Adding complex UI for ontology ACL roles (CSV key=value editing is sufficient for v1)

## Decisions

1. **FTS5 sanitizer: strip-all approach** — Add all known problematic characters to `escapeFTS5Token` replacer rather than switching to a whitelist approach. Rationale: minimal diff, preserves existing behavior for valid tokens (alphanumeric, CJK).

2. **Bulk index COALESCE** — Use `COALESCE(col, '')` in SQL rather than `sql.NullString` in Go. Rationale: simpler, no scan type changes, consistent with existing raw SQL patterns in the codebase.

3. **Ontology ACL roles: CSV editing** — Use `parseKeyValuePairs` (existing helper at state_update.go:1201) for `map[string]string` role editing as `"operator=write,librarian=read"`. Rationale: reuses existing code, avoids complex list-edit UI.

4. **Alerting depends on Observability** — Register alerting with required dependency on observability in `dependencies.go`, and gate `categoryIsEnabled` with `cfg.Observability.Enabled && cfg.Alerting.Enabled`. Rationale: matches actual wiring — `initObservability()` returns nil when disabled, skipping alerting setup.

5. **Ontology has no graph dependency** — Do not register graph as required dependency for ontology. Rationale: `wiring_ontology.go:23` initializes registry/property/ACL/governance/exchange without graph; graph is only used conditionally for truth maintenance.

6. **Downstream sync: full re-sync** — Re-synchronize all three category-list files (`settings.go`, `core.md`, `spec.md`) from `menu.go` source of truth, not just adding two new entries. Also audit `configuration.md` settings labels.

## Risks / Trade-offs

- [FTS5 strip-all may remove valid search characters in non-Latin scripts] → Mitigated by testing with Korean and CJK input; only ASCII punctuation is stripped.
- [CSV role editing is limited for complex ACL setups] → Acceptable for v1; advanced users can use `lango config import` with JSON.
- [Full re-sync of category lists may cause unrelated diff noise] → Acceptable; the alternative is leaving persistent drift.
