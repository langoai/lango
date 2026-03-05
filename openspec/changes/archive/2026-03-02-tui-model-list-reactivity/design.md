## Context

The TUI forms in Settings and Onboard fetch model lists from AI providers at form creation time. Two bugs exist: (1) `${ENV_VAR}` references in API keys are not expanded when creating provider instances for model fetching, causing authentication failures for OpenAI and other providers; (2) changing the provider field has no effect on the model list until the form is recreated.

The existing `FormModel` has no concept of field interdependency or asynchronous updates — fields are static after creation.

## Goals / Non-Goals

**Goals:**
- Fix model list fetching when API keys use `${ENV_VAR}` syntax
- Enable real-time model list refresh when provider selection changes
- Show loading state and error feedback during async model fetches
- Guard against race conditions from rapid provider switching
- Maintain backward compatibility — forms without `OnChange` work identically

**Non-Goals:**
- Full reactive form framework (only provider→model dependency needed)
- Server-side model caching or rate limiting
- Refactoring the entire FormModel architecture

## Decisions

### 1. Export `ExpandEnvVars` from config package
**Rationale**: The function already exists as private `expandEnvVars`. Making it public allows `model_fetcher.go` to expand env vars at the point of provider creation, without duplicating regex logic. Alternative was to add a separate utility function — rejected because it would duplicate the regex and behavior.

### 2. Bubble Tea Cmd pattern for async model fetching
**Rationale**: Bubble Tea's architecture requires side effects to be expressed as `tea.Cmd` functions that return `tea.Msg`. Using `FetchModelOptionsCmd()` that returns `FieldOptionsLoadedMsg` integrates naturally with the existing update loop. Alternative was goroutine+channel — rejected because it bypasses Bubble Tea's message queue and creates concurrency issues.

### 3. `OnChange` callback on Field struct
**Rationale**: A simple callback `func(string) tea.Cmd` on the Field struct provides a minimal reactive mechanism without requiring a full event system. The form's Update method invokes it when InputSelect value changes. Alternative was a form-level event bus — rejected as over-engineered for the current use case.

### 4. ProviderID in FieldOptionsLoadedMsg for race-condition defense
**Rationale**: If a user rapidly switches providers (A→B→C), the fetch for A may complete after C's fetch starts. Including the provider ID at request time allows the handler to detect and discard stale results. Current implementation ignores stale results silently — no error shown since the user already moved on.

### 5. Fallback to InputText on fetch error
**Rationale**: When model fetching fails, the field type changes from InputSearchSelect to InputText so users can still manually type a model ID. This preserves functionality while showing the error in the description.

## Risks / Trade-offs

- **[Stale config reference]** OnChange closures capture a `*config.Config` pointer. If config is mutated elsewhere during the form session, fetches use the updated state. → Acceptable since config is not mutated during TUI sessions.
- **[Multiple concurrent fetches]** Rapid provider switching spawns multiple goroutines. → Mitigated by ProviderID guard in the handler; old results are discarded. Goroutine count is bounded by user interaction speed.
- **[No retry on fetch failure]** A transient network error shows an error message but does not retry. → User can switch provider away and back to trigger a new fetch. Adding auto-retry would complicate the UX with no clear benefit.
