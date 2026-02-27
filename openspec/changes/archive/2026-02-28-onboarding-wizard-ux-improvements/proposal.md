## Why

The Settings editor (`internal/cli/settings/`) received major UX improvements — inline descriptions, model auto-fetch, field validation, and conditional visibility — but the Onboarding wizard (`internal/cli/onboard/`) uses the same `tuicore.FormModel`/`tuicore.Field` and was not updated. Users encounter inconsistent UX between the two entry points for the same fields.

## What Changes

- Export `FetchModelOptions` and `NewProviderFromConfig` from the settings package so the onboard package can reuse model auto-fetch logic
- Add `Description` to all onboard form fields (Provider, Agent, Channel, Security steps) matching Settings wording
- Add model auto-fetch to Agent Step's model field using `settings.FetchModelOptions()`
- Add Temperature validator (0.0–2.0 range) and strengthen Max Tokens validator (positive integer)
- Add `VisibleWhen` conditional visibility to Security Step sub-fields (PII, Policy hidden when interceptor disabled)
- Add "github" provider to all provider option lists (Provider Step, `buildProviderOptions` fallback, `suggestModel`)
- Add "github" to Settings `NewProviderForm` options for consistency
- Update call sites in `forms_impl.go` to use exported function names

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-onboard`: Add field descriptions, model auto-fetch, validators, conditional visibility, and github provider support
- `cli-settings`: Export model fetcher functions for cross-package reuse; add github to provider form options

## Impact

- `internal/cli/settings/model_fetcher.go` — function exports (API change, non-breaking)
- `internal/cli/settings/forms_impl.go` — call site updates + github option
- `internal/cli/onboard/steps.go` — descriptions, auto-fetch, validators, conditional visibility, github
- `internal/cli/onboard/steps_test.go` — new tests for all improvements
