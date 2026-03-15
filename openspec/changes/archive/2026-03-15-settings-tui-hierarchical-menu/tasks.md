## 1. MenuModel State & Types

- [x] 1.1 Add `menuLevel` type with `levelSections` and `levelCategories` constants
- [x] 1.2 Add `level`, `activeSectionIdx`, `sectionCursor` fields to `MenuModel`
- [x] 1.3 Initialize new fields in `NewMenuModel()` (`level: levelSections`, `activeSectionIdx: -1`)
- [x] 1.4 Add `InCategoryLevel()` and `ActiveSectionTitle()` public methods

## 2. Navigation Logic

- [x] 2.1 Add `level1Items()` method — builds synthetic section items + save/cancel
- [x] 2.2 Add `activeSectionCategories()` method — returns tier-filtered categories for active section
- [x] 2.3 Modify `selectableItems()` to dispatch by level (Level 1 → `level1Items`, Level 2 → `activeSectionCategories`)
- [x] 2.4 Add `esc` case in normal mode Update — Level 2 → Level 1 with cursor restoration
- [x] 2.5 Modify `tab` case — no-op at Level 1, toggle at Level 2
- [x] 2.6 Modify `enter` case — detect `__section_` prefix for Level 1 → Level 2 transition

## 3. Rendering

- [x] 3.1 Add `renderSectionListView()` — Level 1 items with separator before save/cancel
- [x] 3.2 Add `renderCategoryDetailView()` — Level 2 items with "No basic settings" empty state
- [x] 3.3 Add `renderTabIndicator()` — `[Basic]` / `[All]` styled labels
- [x] 3.4 Modify `View()` — section header + tab indicator at Level 2, dispatch body by level
- [x] 3.5 Update help footer — Level 1 omits Tab hint, Level 2 includes Tab hint

## 4. Editor Integration

- [x] 4.1 Add `InCategoryLevel()` guard to editor.go Esc handler at StepMenu
- [x] 4.2 Update breadcrumb to show section title at Level 2
- [x] 4.3 Update welcome screen tip text

## 5. Tests

- [x] 5.1 Add `TestEditor_EscAtMenuLevel2_StaysAtMenu`
- [x] 5.2 Add `TestMenu_EnterSection_TransitionsToLevel2`
- [x] 5.3 Add `TestMenu_EscAtLevel2_ReturnsToLevel1`
- [x] 5.4 Add `TestMenu_TabOnlyAtLevel2`
- [x] 5.5 Add `TestMenu_SearchAtBothLevels`
- [x] 5.6 Add `TestMenu_SaveCancelFromLevel1`

## 6. Verification

- [x] 6.1 `go build ./...` passes
- [x] 6.2 `go test ./internal/cli/settings/...` passes
- [x] 6.3 `go test ./...` full test suite passes
