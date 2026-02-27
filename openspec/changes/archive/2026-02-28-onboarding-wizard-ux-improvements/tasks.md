## 1. Settings Package — Export Functions

- [x] 1.1 Rename `fetchModelOptions` → `FetchModelOptions` in `internal/cli/settings/model_fetcher.go`
- [x] 1.2 Rename `newProviderFromConfig` → `NewProviderFromConfig` in `internal/cli/settings/model_fetcher.go`
- [x] 1.3 Update all 5 call sites in `internal/cli/settings/forms_impl.go` to use `FetchModelOptions`
- [x] 1.4 Add "github" to `NewProviderForm` options in `internal/cli/settings/forms_impl.go`

## 2. Onboard Steps — Descriptions

- [x] 2.1 Add Description to all 4 Provider Step fields (type, id, apikey, baseurl)
- [x] 2.2 Add Description to all 4 Agent Step fields (provider, model, maxtokens, temp)
- [x] 2.3 Add Description to all Channel form fields (telegram_token, discord_token, slack_token, slack_app_token)
- [x] 2.4 Add Description to all 3 Security Step fields (interceptor_enabled, interceptor_pii, interceptor_policy)

## 3. Onboard Steps — Model Auto-Fetch

- [x] 3.1 Import `settings` package in `internal/cli/onboard/steps.go`
- [x] 3.2 Add `settings.FetchModelOptions()` call after model field in `NewAgentStepForm`, converting to InputSelect on success

## 4. Onboard Steps — Validators

- [x] 4.1 Add Temperature validator: `strconv.ParseFloat` + 0.0–2.0 range check
- [x] 4.2 Strengthen Max Tokens validator: add `v <= 0` positive check

## 5. Onboard Steps — Conditional Visibility

- [x] 5.1 Capture `interceptorEnabled` field pointer in `NewSecurityStepForm`
- [x] 5.2 Add `VisibleWhen` closures to interceptor_pii and interceptor_policy referencing `interceptorEnabled.Checked`
- [x] 5.3 Indent sub-field labels: `"  Redact PII"`, `"  Approval Policy"`

## 6. Onboard Steps — GitHub Provider

- [x] 6.1 Add "github" to `NewProviderStepForm` type options
- [x] 6.2 Add "github" to `buildProviderOptions` fallback list
- [x] 6.3 Add `case "github": return "gpt-4o"` to `suggestModel`

## 7. Tests

- [x] 7.1 Add `TestAllFormsHaveDescriptions` — verify all form fields have non-empty Description
- [x] 7.2 Add `TestProviderOptionsIncludeGitHub` — verify "github" in provider type options and fallback list
- [x] 7.3 Add `TestTemperatureValidator` — table test (0.0, 1.5, 2.0, 2.1, -0.1, "abc")
- [x] 7.4 Add `TestMaxTokensValidator` — table test (4096, 1, 0, -1, "abc")
- [x] 7.5 Add `TestSecurityConditionalVisibility` — toggle interceptor and count visible fields
- [x] 7.6 Add `{give: "github", want: "gpt-4o"}` to `TestSuggestModel` table

## 8. Verification

- [x] 8.1 Run `go build ./...` — verify no build errors
- [x] 8.2 Run `go test ./internal/cli/onboard/...` — all tests pass
- [x] 8.3 Run `go test ./internal/cli/settings/...` — all tests pass
