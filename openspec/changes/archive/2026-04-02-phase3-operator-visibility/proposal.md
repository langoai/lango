## Why

Phase 1-2 secured and verified the runtime. Phase 3 connects it to standard observability tooling so operators can monitor, trace, and receive alerts through production infrastructure (Prometheus, OpenTelemetry, webhooks). Additionally, 6 files still used `log.Printf` instead of the project's structured logging framework, creating inconsistent log output.

## What Changes

- **Unit 19**: Migrate 6 files from `log.Printf` to `logging.SubsystemSugar()` (payment, contract, configstore, runledger)
- **Unit 18**: Add external alert delivery via webhook channels with severity filtering and async dispatch
- **Unit 16**: Add Prometheus exposition endpoint at `/metrics/prometheus` with event-driven counters/gauges/histograms
- **Unit 17**: Add OpenTelemetry tool-level tracing with stdout exporter and outermost middleware placement
- **Codex review fixes**: Async webhook delivery, tracer lifecycle management, TUI delivery settings preservation, tracked-sessions gauge accuracy

## Capabilities

### New Capabilities

(none — these are infrastructure additions, not new user-facing features)

### Modified Capabilities

- `alerting`: External delivery channels (webhook) with severity filtering
- `observability`: Prometheus metrics endpoint + OpenTelemetry tracing
- `tool-execution-hooks`: Tracing middleware added as outermost layer

## Impact

- **New files**: `internal/alerting/delivery.go`, `internal/observability/prometheus.go`, `internal/observability/tracing.go`, `internal/toolchain/mw_tracing.go` + tests
- **Modified files**: `internal/config/types.go`, `internal/config/types_observability.go`, `internal/app/wiring_observability.go`, `internal/app/app.go`, `internal/app/routes_observability.go`, `internal/cli/settings/forms_alerting.go`, `internal/cli/tuicore/state_update.go`, `internal/toolchain/chain_order_test.go`, 6 logging migration files
- **Dependencies**: `prometheus/client_golang` + `go.opentelemetry.io/otel` promoted to direct
