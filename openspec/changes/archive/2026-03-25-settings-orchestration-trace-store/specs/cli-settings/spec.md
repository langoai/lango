## ADDED Requirements

### Requirement: Orchestration state update mapping
The `UpdateConfigFromForm()` function SHALL handle orchestration form field keys, mapping them to `Config.Agent.Orchestration` sub-fields: `orchestration_mode` → `Mode`, `orc_cb_failure_threshold` → `CircuitBreaker.FailureThreshold`, `orc_cb_reset_timeout` → `CircuitBreaker.ResetTimeout`, `orc_budget_tool_call_limit` → `Budget.ToolCallLimit`, `orc_budget_delegation_limit` → `Budget.DelegationLimit`, `orc_budget_alert_threshold` → `Budget.AlertThreshold`, `orc_recovery_max_retries` → `Recovery.MaxRetries`, `orc_recovery_cooldown` → `Recovery.CircuitBreakerCooldown`.

#### Scenario: Orchestration fields saved to config
- **WHEN** user sets `orchestration_mode=structured`, `orc_cb_failure_threshold=5`, `orc_budget_alert_threshold=0.75`
- **THEN** `Config.Agent.Orchestration.Mode` SHALL be "structured", `CircuitBreaker.FailureThreshold` SHALL be 5, `Budget.AlertThreshold` SHALL be 0.75

#### Scenario: Invalid orchestration values ignored
- **WHEN** user enters "not-a-number" for `orc_cb_failure_threshold`
- **THEN** the config value SHALL remain unchanged (parse error silently skipped)

### Requirement: Trace store state update mapping
The `UpdateConfigFromForm()` function SHALL handle trace store form field keys, mapping them to `Config.Observability.TraceStore` sub-fields: `obs_trace_max_age` → `MaxAge`, `obs_trace_max_traces` → `MaxTraces`, `obs_trace_failed_multiplier` → `FailedTraceMultiplier`, `obs_trace_cleanup_interval` → `CleanupInterval`.

#### Scenario: Trace store fields saved to config
- **WHEN** user sets `obs_trace_max_age=168h`, `obs_trace_max_traces=5000`
- **THEN** `Config.Observability.TraceStore.MaxAge` SHALL be 168h, `MaxTraces` SHALL be 5000

#### Scenario: Duration parse for trace store
- **WHEN** user enters "30m" for `obs_trace_cleanup_interval`
- **THEN** `Config.Observability.TraceStore.CleanupInterval` SHALL be 30 minutes
