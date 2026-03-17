## 1. Dependency Registry

- [x] 1.1 Create `internal/cli/settings/dependencies.go` with DepStatus, Dependency, DepResult, DependencyIndex types
- [x] 1.2 Implement Evaluate, UnmetRequired, AllTransitiveUnmet (with cycle guard), Dependents, HasDependencies methods
- [x] 1.3 Define defaultDependencies() with all 20+ dependency relationships (Smart Account, P2P, Economy, Librarian, etc.)
- [x] 1.4 Create shared check functions (checkSmartAccountEnabled, checkP2PEnabled, checkEconomyEnabled)

## 2. Menu Warning Badges + @ready Filter

- [x] 2.1 Add BadgeDependencyStyle to `internal/cli/tui/styles.go`
- [x] 2.2 Add DependencyChecker callback field to MenuModel in `internal/cli/settings/menu.go`
- [x] 2.3 Add @ready case in applyFilter() and update filter hint text
- [x] 2.4 Render dependency warning badge in renderItem() after ADV badge
- [x] 2.5 Wire DependencyChecker callback in wireMenuCheckers() in `internal/cli/settings/editor.go`

## 3. Prerequisite Panel Component

- [x] 3.1 Create `internal/cli/settings/dependency_panel.go` with DependencyPanel type
- [x] 3.2 Implement NewDependencyPanel (returns nil if all met), View, cursor navigation (MoveUp/MoveDown)
- [x] 3.3 Implement SelectedCategoryID, SelectedIsUnmet, UnmetCount, StatusSummary helpers

## 4. Editor Panel Integration + Quick-Jump

- [x] 4.1 Add depIndex, depPanel, panelFocus, navStack fields to Editor struct
- [x] 4.2 Initialize depIndex in NewEditor and NewEditorWithConfig constructors
- [x] 4.3 Implement attachDependencyPanel helper
- [x] 4.4 Add panel key handling in StepForm Update (up/down/enter/tab/esc/s)
- [x] 4.5 Implement jumpToDependency (push nav stack, open dep's form)
- [x] 4.6 Implement popNavStack (return to original category)
- [x] 4.7 Render dependency panel above form in StepForm View
- [x] 4.8 Add breadcrumb navigation chain for jumped forms

## 5. Guided Setup Flow

- [x] 5.1 Create `internal/cli/settings/setup_flow.go` with SetupFlow, SetupStep, SetupFlowState types
- [x] 5.2 Implement NewSetupFlow with deduplication of transitive deps
- [x] 5.3 Implement NextStep, SkipStep, Cancel methods
- [x] 5.4 Implement View with progress bar and step list rendering
- [x] 5.5 Create shared createFormForCategory() factory function

## 6. Editor Setup Flow Integration

- [x] 6.1 Add StepSetupFlow constant and setupFlow field to Editor
- [x] 6.2 Implement startSetupFlow (from panel 's' key)
- [x] 6.3 Add StepSetupFlow Update handling (Ctrl+N next, Ctrl+S skip, Esc cancel)
- [x] 6.4 Implement completeSetupFlow (open target form after completion)
- [x] 6.5 Add StepSetupFlow View rendering and breadcrumb

## 7. Refactoring + Code Quality

- [x] 7.1 Refactor handleMenuSelection to use createFormForCategory for all form-opening cases
- [x] 7.2 Update welcome screen tips to include @ready

## 8. Tests

- [x] 8.1 TestDependencyIndex_SmartAccountUnmet (table-driven, 4 cases)
- [x] 8.2 TestDependencyIndex_Evaluate (3 results check)
- [x] 8.3 TestDependencyIndex_TransitiveResolution (depth-first order)
- [x] 8.4 TestDependencyIndex_CycleGuard
- [x] 8.5 TestDependencyIndex_Dependents (reverse lookup)
- [x] 8.6 TestDependencyIndex_NoDepsCategory
- [x] 8.7 TestDependencyPanel_NilWhenAllMet
- [x] 8.8 TestDependencyPanel_CreatedWhenUnmet
- [x] 8.9 TestDependencyPanel_Navigation
- [x] 8.10 TestDependencyPanel_View
- [x] 8.11 TestSetupFlow_Creation (table-driven)
- [x] 8.12 TestSetupFlow_StepProgression
- [x] 8.13 TestSetupFlow_Cancel
- [x] 8.14 TestSetupFlow_View
- [x] 8.15 TestSetupFlow_DeduplicatesDeps
- [x] 8.16 TestMenuModel_ReadyFilter

## 9. Verification

- [x] 9.1 go build ./... passes
- [x] 9.2 go test ./internal/cli/settings/... -v passes (all 16 tests)
- [x] 9.3 go vet ./internal/cli/settings/... passes
