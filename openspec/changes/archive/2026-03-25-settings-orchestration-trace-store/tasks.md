## 1. Multi-Agent Form Extension

- [x] 1.1 Refactor `multi_agent` field to named variable in `NewMultiAgentForm` (`forms_agent.go`)
- [x] 1.2 Add default fallback logic using `OrchestrationDefaults()` for zero-value config fields
- [x] 1.3 Add `orchestration_mode` InputSelect field with `VisibleWhen: isMultiAgentOn`
- [x] 1.4 Add 7 structured policy fields (`orc_cb_*`, `orc_budget_*`, `orc_recovery_*`) with `VisibleWhen: isStructured`
- [x] 1.5 Add `time` import to `forms_agent.go`

## 2. Observability Form Extension

- [x] 2.1 Refactor `obs_enabled` field to named variable in `NewObservabilityForm` (`forms_observability.go`)
- [x] 2.2 Add default fallback logic using `TraceStoreDefaults()` for zero-value trace store config
- [x] 2.3 Add 4 trace store fields (`obs_trace_*`) with `VisibleWhen: isObsEnabled`
- [x] 2.4 Add `time` import to `forms_observability.go`

## 3. State Binding

- [x] 3.1 Add 8 orchestration case statements in `UpdateConfigFromForm` after `agents_dir` (`state_update.go`)
- [x] 3.2 Add 4 trace store case statements in `UpdateConfigFromForm` after `obs_metrics_format` (`state_update.go`)

## 4. Tests

- [x] 4.1 Add `visibleKeys` helper and orchestration field existence test (`forms_impl_test.go`)
- [x] 4.2 Add orchestration defaults test verifying fallback values
- [x] 4.3 Add VisibleWhen 3-level chain test (multi_agent off → classic → structured)
- [x] 4.4 Add trace store field existence and visibility tests
- [x] 4.5 Add state binding tests for orchestration and trace store fields
- [x] 4.6 Add invalid value test for parse error resilience

## 5. Verification

- [x] 5.1 Run `go build ./...` — no compile errors
- [x] 5.2 Run `go test ./internal/cli/settings/... ./internal/cli/tuicore/... ./internal/config/...` — all pass
