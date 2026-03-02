## Why

TUI Settings and Onboard forms fail to display model lists when API keys use `${ENV_VAR}` references because `NewProviderFromConfig()` does not expand environment variables. Additionally, changing the provider field does not refresh the model list — users must exit and re-enter the form to see updated models.

## What Changes

- Export `config.ExpandEnvVars()` and apply it in `NewProviderFromConfig()` so API keys and base URLs with `${VAR}` references resolve correctly when fetching models
- Add `OnChange` callback and `Loading`/`LoadError` fields to `tuicore.Field` for reactive field dependencies
- Add `FieldOptionsLoadedMsg` message type and async handling in `FormModel.Update()` to refresh model options when a provider field changes
- Add `FetchModelOptionsCmd()` and `FetchEmbeddingModelOptionsCmd()` Bubble Tea Cmd wrappers for async model fetching
- Wire `OnChange` callbacks in all provider→model field pairs across Settings (Agent, Fallback, OM, Embedding, Librarian) and Onboard (Agent step)
- Forward non-key messages in `onboard/wizard.go` to the active form so `FieldOptionsLoadedMsg` reaches forms
- Upgrade Onboard model field from `InputSelect` to `InputSearchSelect` with error visibility
- Add debug logging to OpenAI provider's `ListModels()`

## Capabilities

### New Capabilities
- `tui-reactive-fields`: Reactive field dependency system for TUI forms — `OnChange` callbacks, async loading state, and `FieldOptionsLoadedMsg` pattern

### Modified Capabilities
- `cli-tuicore`: Add `OnChange`, `Loading`, `LoadError` to Field; handle `FieldOptionsLoadedMsg` in FormModel
- `cli-settings`: Wire reactive provider→model dependencies in Agent, Knowledge, Embedding, and Librarian forms
- `cli-onboard`: Wire reactive provider→model in Agent step; forward async messages in Wizard; improve error visibility
- `config-system`: Export `ExpandEnvVars` for use outside the config loader
- `provider-openai-compatible`: Add debug logging to `ListModels()`

## Impact

- `internal/config/loader.go` — `expandEnvVars` renamed to `ExpandEnvVars` (exported)
- `internal/config/loader_test.go` — Updated test references
- `internal/cli/tuicore/field.go` — New fields on `Field` struct
- `internal/cli/tuicore/messages.go` — New file
- `internal/cli/tuicore/form.go` — `FieldOptionsLoadedMsg` handler + `OnChange` invocation + Loading view
- `internal/cli/tuicore/form_test.go` — 5 new test cases
- `internal/cli/settings/model_fetcher.go` — Env var expansion + Cmd wrappers
- `internal/cli/settings/forms_impl.go` — Reactive wiring for Agent + Fallback
- `internal/cli/settings/forms_knowledge.go` — Reactive wiring for OM, Embedding, Librarian
- `internal/cli/onboard/steps.go` — Reactive Agent step + error visibility
- `internal/cli/onboard/wizard.go` — Default message forwarding
- `internal/provider/openai/openai.go` — Debug logging
