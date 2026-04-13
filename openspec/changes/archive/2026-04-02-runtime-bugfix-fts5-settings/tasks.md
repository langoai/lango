## 1. FTS5 Query Safety

- [x] 1.1 Add missing punctuation chars to `escapeFTS5Token` replacer in `internal/search/fts5.go:265`
- [x] 1.2 Add empty-token skip in `sanitizeFTS5Query` for normal, prefix, and unclosed-quote paths
- [x] 1.3 Add tests in `internal/search/fts5_test.go`: `?`, `.`, `!`, prefix `?*`, unclosed `"?`, Korean text

## 2. FTS5 Bulk Index

- [x] 2.1 Fix table name `knowledge` → `knowledges` in `internal/app/wiring_knowledge.go:131`
- [x] 2.2 Add `COALESCE(error_pattern, '')` and `COALESCE(fix, '')` in learning bulk index query at line 164

## 3. Settings Forms

- [x] 3.1 Create `internal/cli/settings/forms_ontology.go` with `NewOntologyForm` (15 fields, VisibleWhen conditionals)
- [x] 3.2 Create `internal/cli/settings/forms_alerting.go` with `NewAlertingForm` (3 fields)

## 4. Settings Wiring

- [x] 4.1 Add ontology + alerting cases to `internal/cli/tuicore/state_update.go` `UpdateConfigFromForm` switch
- [x] 4.2 Add ontology + alerting menu entries to `internal/cli/settings/menu.go`
- [x] 4.3 Add `case "ontology"` and `case "alerting"` to `createFormForCategory` in `setup_flow.go`
- [x] 4.4 Add `case "ontology"` and `case "alerting"` to `categoryIsEnabled` in `editor.go`
- [x] 4.5 Add alerting dependency on observability in `dependencies.go`

## 5. Settings Tests

- [x] 5.1 Add `createFormForCategory` non-nil tests for ontology and alerting
- [x] 5.2 Add `UpdateConfigFromForm` round-trip tests for ontology/alerting keys in `forms_impl_test.go`
- [x] 5.3 Add alerting dependency evaluation test in `dependencies_test.go`

## 6. Downstream Sync

- [x] 6.1 Re-sync `internal/cli/settings/settings.go` Long description with `menu.go` categories
- [x] 6.2 Re-sync `docs/cli/core.md` settings section with `menu.go` categories
- [x] 6.3 Re-sync `openspec/specs/cli-settings/spec.md` section SHALL list with `menu.go` categories
- [x] 6.4 Audit `docs/configuration.md` settings labels (e.g., "Sandbox" → "OS Sandbox")
- [x] 6.5 Grep-based audit of `docs/` for `lango settings` references and fix mismatches

## 7. Verification

- [x] 7.1 `go build -tags "fts5,vec" ./...` + `go build ./...`
- [x] 7.2 `go test ./...` + `go test -tags "fts5,vec" ./...`
