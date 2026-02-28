## Context

The TUI settings editor uses `InputSelect` (left/right arrow cycling) for model selection. With hardcoded 3 models this was adequate, but live API calls return 50-200+ models, making arrow cycling impractical. The form's Esc handling also conflicts with dropdown state.

## Goals / Non-Goals

**Goals:**
- Live model lists from Gemini and Anthropic APIs
- Searchable dropdown component for large option sets
- Correct Esc key layering (dropdown close → form exit)
- Embedding model filtering for embedding-specific fields

**Non-Goals:**
- Caching API model lists across sessions
- Model metadata display (context window, pricing)
- Custom model input alongside search (text fallback remains when API fails)

## Decisions

1. **InputSearchSelect as new InputType**: Added to the existing `InputType` enum rather than creating a separate component. This keeps the form model unified and avoids a parallel rendering path.

2. **Two-state Esc handling**: When `SelectOpen == true`, Esc closes the dropdown only. The editor checks `HasOpenDropdown()` before consuming Esc for form exit. This creates a natural 2-step exit: Esc → close dropdown, Esc → exit form.

3. **Filter state on Field struct**: `FilteredOptions`, `SelectCursor`, `SelectOpen` live on `Field` rather than a separate model. This avoids cross-struct synchronization and keeps `FormModel.Update()` as the single control point.

4. **Embedding filtering by name pattern**: Uses substring matching ("embed", "embedding") rather than API metadata. Provider APIs don't consistently expose model capability tags, so name-based heuristics with full-list fallback is pragmatic.

5. **Graceful degradation**: If API calls fail or return empty, fields remain as `InputText` for manual entry. No user-visible errors for model fetch failures.

## Risks / Trade-offs

- **API latency on form open**: Model fetching adds up to 5s (timeout) when opening settings forms. Mitigated by the existing `modelFetchTimeout` constant.
- **Embedding pattern matching**: May miss unusually named embedding models or include non-embedding models with "embed" in the name. Fallback to full list prevents data loss.
- **Dropdown max 8 visible items**: Fixed limit may be too small for browsing without typing. Chosen for terminal height compatibility.
