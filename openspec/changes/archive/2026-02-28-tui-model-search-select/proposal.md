## Why

The TUI settings model selection has several usability issues: Gemini and Anthropic providers show only 3 hardcoded models (missing latest releases), arrow-key rapid navigation causes premature menu exit, and browsing hundreds of models via left/right arrows alone is impractical. Embedding model selection also lacks filtering for embedding-specific models.

## What Changes

- Replace hardcoded model lists in Gemini and Anthropic providers with live API calls (`Models.All()` and `Models.ListAutoPaging()`)
- Add `InputSearchSelect` TUI component: searchable dropdown with type-to-filter, up/down navigation, and Enter/Esc handling
- Fix Esc key bug where pressing Esc while a dropdown is open exits the entire form instead of just closing the dropdown
- Add `FetchEmbeddingModelOptions()` that filters model lists for embedding-capable models
- Convert all model selection fields (agent, fallback, embedding, observational memory, librarian) from `InputSelect` to `InputSearchSelect`

## Capabilities

### New Capabilities
- `input-search-select`: Searchable dropdown select component for TUI forms with real-time filtering, keyboard navigation, and multi-state Esc handling

### Modified Capabilities
- `provider-anthropic`: ListModels now calls the live API instead of returning hardcoded values
- `cli-tuicore`: New InputSearchSelect field type with FilteredOptions, SelectCursor, SelectOpen state management
- `cli-settings`: Model fields use InputSearchSelect; embedding model selection uses filtered model list

## Impact

- `internal/provider/gemini/gemini.go` — ListModels uses live API
- `internal/provider/anthropic/anthropic.go` — ListModels uses live API with pagination
- `internal/cli/tuicore/field.go` — New InputSearchSelect type and filter state
- `internal/cli/tuicore/form.go` — Dropdown open/close/navigate/filter logic and rendering
- `internal/cli/settings/editor.go` — Esc key passthrough for open dropdowns
- `internal/cli/settings/model_fetcher.go` — FetchEmbeddingModelOptions with pattern filtering
- `internal/cli/settings/forms_impl.go` — All model fields converted to InputSearchSelect
