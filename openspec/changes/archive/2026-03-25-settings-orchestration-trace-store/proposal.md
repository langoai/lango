## Why

The structured multi-agent control plane (`agentrt` package) and turn trace retention (`turntrace` package) are fully implemented with config structs (`OrchestrationConfig`, `TraceStoreConfig`) and runtime defaults, but users cannot configure these settings through the TUI settings editor. This makes the structured orchestration mode operationally inaccessible without manual YAML editing.

## What Changes

- Extend `NewMultiAgentForm` with 8 orchestration fields: `orchestration_mode` (classic/structured), circuit breaker threshold/timeout, budget limits/alert threshold, recovery retries/cooldown
- Extend `NewObservabilityForm` with 4 trace store fields: max age, max traces, failed trace multiplier, cleanup interval
- Add 12 state binding cases in `UpdateConfigFromForm` for config persistence
- Implement 2-level `VisibleWhen` chain: orchestration fields appear only when `multi_agent=true && mode=structured`
- Trace store fields gated by `obs_enabled` toggle
- Zero-value semantics: `0` means "use runtime default" — documented in field descriptions

## Capabilities

### New Capabilities
- `settings-orchestration`: TUI form fields for structured orchestration control plane (circuit breaker, budget, recovery policies) within the existing Multi-Agent settings category

### Modified Capabilities
- `cli-settings`: New field keys added to `UpdateConfigFromForm` state binding (8 orchestration + 4 trace store cases)

## Impact

- `internal/cli/settings/forms_agent.go` — 8 new fields in `NewMultiAgentForm`
- `internal/cli/settings/forms_observability.go` — 4 new fields in `NewObservabilityForm`
- `internal/cli/tuicore/state_update.go` — 12 new case statements in `UpdateConfigFromForm`
- `internal/cli/settings/forms_impl_test.go` — 8 new test functions
- No menu, dispatcher, or config struct changes required
