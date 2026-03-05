## 1. Environment Variable Expansion (Bug 1 Core Fix)

- [x] 1.1 Rename `expandEnvVars` to `ExpandEnvVars` in `internal/config/loader.go` and update all internal call sites
- [x] 1.2 Update `internal/config/loader_test.go` to reference `ExpandEnvVars`
- [x] 1.3 Apply `config.ExpandEnvVars()` to `APIKey` and `BaseURL` in `NewProviderFromConfig()` in `internal/cli/settings/model_fetcher.go`

## 2. Reactive Field Infrastructure (Bug 2 Core Fix)

- [x] 2.1 Add `OnChange`, `Loading`, `LoadError` fields to `tuicore.Field` struct in `internal/cli/tuicore/field.go`
- [x] 2.2 Create `internal/cli/tuicore/messages.go` with `FieldOptionsLoadedMsg` type
- [x] 2.3 Add `FieldOptionsLoadedMsg` handler in `FormModel.Update()` in `internal/cli/tuicore/form.go`
- [x] 2.4 Add `OnChange` invocation after InputSelect value change in `FormModel.Update()`
- [x] 2.5 Add "Loading models..." display in `FormModel.View()` when `field.Loading == true`

## 3. Async Cmd Wrappers

- [x] 3.1 Add `FetchModelOptionsCmd()` function in `internal/cli/settings/model_fetcher.go`
- [x] 3.2 Add `FetchEmbeddingModelOptionsCmd()` function in `internal/cli/settings/model_fetcher.go`

## 4. Settings Forms Reactive Wiring

- [x] 4.1 Wire `OnChange` on provider field in `NewAgentForm()` to fetch models for "model" field
- [x] 4.2 Wire `OnChange` on fallback_provider field in `NewAgentForm()` to fetch models for "fallback_model" field
- [x] 4.3 Wire `OnChange` on om_provider field in `NewObservationalMemoryForm()` to fetch models for "om_model" field
- [x] 4.4 Wire `OnChange` on emb_provider_id field in `NewEmbeddingForm()` to fetch embedding models for "emb_model" field
- [x] 4.5 Wire `OnChange` on lib_provider field in `NewLibrarianForm()` to fetch models for "lib_model" field

## 5. Onboard Reactive Wiring

- [x] 5.1 Wire `OnChange` on provider field in `NewAgentStepForm()` to fetch models and update placeholder
- [x] 5.2 Upgrade model field from `InputSelect` to `InputSearchSelect` with error visibility in `NewAgentStepForm()`
- [x] 5.3 Add default message forwarding case in `Wizard.Update()` in `internal/cli/onboard/wizard.go`

## 6. Debug Logging

- [x] 6.1 Add debug logging to `ListModels()` in `internal/provider/openai/openai.go`

## 7. Tests

- [x] 7.1 Add `TestInputSelect_OnChangeCallback` test
- [x] 7.2 Add `TestInputSelect_OnChangeNotCalledWhenNoChange` test
- [x] 7.3 Add `TestFieldOptionsLoadedMsg_UpdatesField` test
- [x] 7.4 Add `TestFieldOptionsLoadedMsg_Error` test
- [x] 7.5 Add `TestFieldOptionsLoadedMsg_WrongFieldKey` test
- [x] 7.6 Verify `go build ./...` succeeds
- [x] 7.7 Verify `go test ./internal/cli/tuicore/... ./internal/cli/settings/... ./internal/cli/onboard/... ./internal/config/...` passes
