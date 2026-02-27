## Context

The Settings editor and Onboarding wizard both use `tuicore.FormModel`/`tuicore.Field` to render configuration forms. The Settings editor was enhanced with inline descriptions, model auto-fetch from provider APIs, field validators, and conditional visibility (`VisibleWhen`). The Onboarding wizard still uses bare fields without these features, leading to inconsistent UX for identical settings.

## Goals / Non-Goals

**Goals:**
- Consistent field descriptions across Settings editor and Onboarding wizard for the same fields
- Model auto-fetch in Onboarding's Agent Step via reuse of the settings package's fetcher
- Input validation for Temperature (0.0–2.0) and Max Tokens (positive integer) in Onboarding
- Conditional visibility for Security Step sub-fields (PII redaction, approval policy hidden when interceptor disabled)
- "github" provider support in all provider option lists

**Non-Goals:**
- Refactoring the form rendering engine or `tuicore` package
- Adding new onboarding steps or changing the wizard flow
- Changing Settings editor behavior beyond function exports and github option

## Decisions

1. **Export fetcher functions from settings package** — `FetchModelOptions` and `NewProviderFromConfig` are exported so `onboard` can import and call them directly. Alternative: duplicate the logic in onboard. Rejected because it violates DRY and would drift over time.

2. **Graceful fallback for model auto-fetch** — If the provider API key is missing or fetch fails, the model field remains a text input with a placeholder suggestion. This matches the Settings editor pattern exactly.

3. **Pointer capture for VisibleWhen** — The `interceptorEnabled` field pointer is captured in a closure for `VisibleWhen` on the PII and Policy fields. This is the same pattern used in the Settings editor's conditional fields.

4. **Indented labels for conditional sub-fields** — PII and Policy labels are prefixed with two spaces (`"  Redact PII"`, `"  Approval Policy"`) to visually indicate hierarchy, matching Settings editor conventions.

## Risks / Trade-offs

- [Cross-package coupling] The onboard package now imports settings for `FetchModelOptions` → Acceptable because both packages are in the same CLI layer and share the same config types.
- [Network call during onboarding] Model auto-fetch makes a network request with 5s timeout → Mitigated by fallback to text input if it fails, same as Settings editor behavior.
