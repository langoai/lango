# Tasks: Onboard/Settings Split

## Phase 1: Extract Shared Components
- [x] Create `internal/cli/tuicore/field.go` — Field, InputType types with exported TextInput
- [x] Create `internal/cli/tuicore/form.go` — FormModel (Init/Update/View)
- [x] Create `internal/cli/tuicore/state.go` — ConfigState + dirty tracking
- [x] Create `internal/cli/tuicore/state_update.go` — UpdateConfigFromForm/UpdateProviderFromForm/UpdateAuthProviderFromForm + splitCSV helper
- [x] Verify: `go build ./internal/cli/tuicore/...`

## Phase 2: Create Settings Package
- [x] Create `internal/cli/settings/settings.go` — Cobra command (lango settings --profile)
- [x] Create `internal/cli/settings/editor.go` — Editor model (renamed from Wizard)
- [x] Create `internal/cli/settings/menu.go` — Configuration menu (17 categories)
- [x] Create `internal/cli/settings/forms_impl.go` — All 15 form constructors using tuicore types
- [x] Create `internal/cli/settings/providers_list.go` — Provider management UI
- [x] Create `internal/cli/settings/auth_providers_list.go` — OIDC provider management UI
- [x] Create `internal/cli/settings/forms_impl_test.go` — Migrated tests
- [x] Register `settings.NewCommand()` in `cmd/lango/main.go`
- [x] Verify: `go build ./... && go test ./internal/cli/settings/...`

## Phase 3: Rewrite Onboard Wizard
- [x] Delete moved files from onboard (form.go, menu.go, forms_impl.go, state.go, state_update.go, providers_list.go, auth_providers_list.go, wizard.go)
- [x] Update `onboard/onboard.go` — New description, launches Wizard
- [x] Create `onboard/wizard.go` — 5-step stepper model (StepProvider → StepAgent → StepChannel → StepSecurity → StepTest)
- [x] Create `onboard/progress.go` — Progress bar + step list renderer
- [x] Create `onboard/steps.go` — Wizard form constructors (essential fields only)
- [x] Create `onboard/test_step.go` — Configuration validation (5 checks)
- [x] Verify: `go build ./internal/cli/onboard/...`

## Phase 4: Tests
- [x] Create `onboard/steps_test.go` — Table-driven tests for step forms
- [x] Create `onboard/test_step_test.go` — Config validation scenarios
- [x] Create `onboard/progress_test.go` — Progress bar rendering tests
- [x] Verify: `go test ./internal/cli/onboard/...`

## Phase 5: Final Verification
- [x] `go build ./...` — Full project compilation
- [x] `go test ./...` — All tests pass
