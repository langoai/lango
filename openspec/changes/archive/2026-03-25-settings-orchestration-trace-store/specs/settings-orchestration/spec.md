## ADDED Requirements

### Requirement: Orchestration mode field in Multi-Agent form
The Multi-Agent settings form SHALL include an `orchestration_mode` field (InputSelect: "classic", "structured") that is visible only when `multi_agent` is enabled.

#### Scenario: Mode field hidden when multi-agent disabled
- **WHEN** `multi_agent` is unchecked (false)
- **THEN** `orchestration_mode` field SHALL NOT appear in `VisibleFields()`

#### Scenario: Mode field visible when multi-agent enabled
- **WHEN** `multi_agent` is checked (true)
- **THEN** `orchestration_mode` field SHALL appear in `VisibleFields()` with options ["classic", "structured"]

### Requirement: Circuit breaker fields in Multi-Agent form
The form SHALL include `orc_cb_failure_threshold` (InputInt) and `orc_cb_reset_timeout` (InputText/duration) fields visible only when `multi_agent=true` AND `orchestration_mode=structured`.

#### Scenario: CB fields hidden in classic mode
- **WHEN** `multi_agent=true` AND `orchestration_mode=classic`
- **THEN** `orc_cb_failure_threshold` and `orc_cb_reset_timeout` SHALL NOT appear in `VisibleFields()`

#### Scenario: CB fields visible in structured mode
- **WHEN** `multi_agent=true` AND `orchestration_mode=structured`
- **THEN** `orc_cb_failure_threshold` and `orc_cb_reset_timeout` SHALL appear in `VisibleFields()`

### Requirement: Budget policy fields in Multi-Agent form
The form SHALL include `orc_budget_tool_call_limit` (InputInt), `orc_budget_delegation_limit` (InputInt), and `orc_budget_alert_threshold` (InputText/float) fields visible only in structured mode.

#### Scenario: Budget fields with default values
- **WHEN** config has zero-value orchestration (fresh profile)
- **THEN** budget fields SHALL display defaults from `OrchestrationDefaults()`: tool_call_limit=50, delegation_limit=15, alert_threshold=0.80

#### Scenario: Alert threshold validation
- **WHEN** user enters a value outside 0.0-1.0 range for `orc_budget_alert_threshold`
- **THEN** validation SHALL reject the input with an error message

### Requirement: Recovery policy fields in Multi-Agent form
The form SHALL include `orc_recovery_max_retries` (InputInt) and `orc_recovery_cooldown` (InputText/duration) fields visible only in structured mode.

#### Scenario: Recovery fields with default values
- **WHEN** config has zero-value orchestration (fresh profile)
- **THEN** recovery fields SHALL display defaults: max_retries=2, cooldown=5m0s

### Requirement: Trace store fields in Observability form
The Observability form SHALL include `obs_trace_max_age` (InputText/duration), `obs_trace_max_traces` (InputInt), `obs_trace_failed_multiplier` (InputInt), and `obs_trace_cleanup_interval` (InputText/duration) fields visible when `obs_enabled=true`.

#### Scenario: Trace fields hidden when observability disabled
- **WHEN** `obs_enabled` is unchecked (false)
- **THEN** all `obs_trace_*` fields SHALL NOT appear in `VisibleFields()`

#### Scenario: Trace fields visible when observability enabled
- **WHEN** `obs_enabled` is checked (true)
- **THEN** all `obs_trace_*` fields SHALL appear in `VisibleFields()` with defaults from `TraceStoreDefaults()`

### Requirement: Zero-value default semantics
All orchestration integer and float fields SHALL accept 0 as valid input. Field descriptions SHALL document `(0 = use default: N)` where N is the runtime default value.

#### Scenario: Zero input preserved
- **WHEN** user enters 0 for `orc_cb_failure_threshold`
- **THEN** form SHALL accept the value and state binding SHALL store 0 in config (runtime applies default)

### Requirement: Values preserved when hidden
Field values SHALL persist in config when their `VisibleWhen` condition becomes false. Toggling `multi_agent` off and back on, or switching `orchestration_mode` between classic and structured, SHALL restore previously entered values.

#### Scenario: Round-trip value preservation
- **WHEN** user sets `orc_cb_failure_threshold=5`, then unchecks `multi_agent`, then re-checks it
- **THEN** `orc_cb_failure_threshold` SHALL still display 5
