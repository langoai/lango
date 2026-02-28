## 1. Provider API Integration

- [x] 1.1 Replace hardcoded Gemini ListModels with `p.client.Models.All(ctx)`, strip "models/" prefix, map InputTokenLimit to ContextWindow
- [x] 1.2 Replace hardcoded Anthropic ListModels with `p.client.Models.ListAutoPaging(ctx, params)` with Limit 1000, graceful partial failure
- [x] 1.3 Update Anthropic test to skip when API key is not set (live API test)

## 2. InputSearchSelect Component

- [x] 2.1 Add `InputSearchSelect` constant to `InputType` enum in `field.go`
- [x] 2.2 Add `FilteredOptions`, `SelectCursor`, `SelectOpen` fields to `Field` struct
- [x] 2.3 Implement `applySearchFilter()` method with case-insensitive substring matching and cursor clamping
- [x] 2.4 Initialize TextInput and FilteredOptions in `AddField()` for InputSearchSelect type
- [x] 2.5 Add `HasOpenDropdown()` method to FormModel
- [x] 2.6 Implement dropdown open/close/navigate/select key handling in `Update()` (intercepts keys before form navigation when open)
- [x] 2.7 Implement dropdown rendering in `View()` with max 8 visible items, scroll, match count, "... N more"
- [x] 2.8 Add context-dependent help bar (dropdown keys vs form keys)

## 3. Esc Key Bug Fix

- [x] 3.1 Add `HasOpenDropdown()` check in `editor.go` StepForm Esc handler to pass Esc to form when dropdown is open

## 4. Forms and Embedding Filter

- [x] 4.1 Add `FetchEmbeddingModelOptions()` with "embed"/"embedding" pattern filtering and full-list fallback
- [x] 4.2 Convert agent model field to InputSearchSelect in `forms_impl.go`
- [x] 4.3 Convert fallback model field to InputSearchSelect
- [x] 4.4 Convert embedding model field to InputSearchSelect with FetchEmbeddingModelOptions
- [x] 4.5 Convert observational memory model field to InputSearchSelect
- [x] 4.6 Convert librarian model field to InputSearchSelect

## 5. Tests

- [x] 5.1 Add `form_test.go` with InputSearchSelect filter/select/Esc/Tab/cursor tests
- [x] 5.2 Add `model_fetcher_test.go` with embedding filter/fallback/current-model tests
