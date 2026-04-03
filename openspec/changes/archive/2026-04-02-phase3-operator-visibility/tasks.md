## 1. Structured Logging Migration (Unit 19)

- [x] 1.1 `internal/payment/service.go` — log.Printf → logging.SubsystemSugar("payment")
- [x] 1.2 `internal/payment/tx_builder.go` — log.Printf → logging.SubsystemSugar("payment")
- [x] 1.3 `internal/contract/caller.go` — log.Printf → logging.SubsystemSugar("contract")
- [x] 1.4 `internal/configstore/migrate.go` — log.Printf → logging.SubsystemSugar("configstore")
- [x] 1.5 `internal/runledger/writethrough.go` — log.Printf → logging.SubsystemSugar("runledger")
- [x] 1.6 `internal/runledger/types.go` — log.Printf → logging.SubsystemSugar("runledger")

## 2. Alert External Delivery (Unit 18)

- [x] 2.1 Create `internal/alerting/delivery.go` with DeliveryChannel interface, WebhookDelivery, DeliveryRouter
- [x] 2.2 Add `AlertDeliveryConfig` and `Delivery` field to `AlertingConfig` in `internal/config/types.go`
- [x] 2.3 Wire DeliveryRouter in `internal/app/wiring_observability.go`
- [x] 2.4 Add webhook URL field to `internal/cli/settings/forms_alerting.go`
- [x] 2.5 Add webhook URL state update in `internal/cli/tuicore/state_update.go` (preserving existing channels)
- [x] 2.6 Create `internal/alerting/delivery_test.go` with webhook, severity filter, fan-out, unknown type tests
- [x] 2.7 Make webhook dispatch async (goroutine) to avoid blocking EventBus

## 3. Prometheus Metrics (Unit 16)

- [x] 3.1 Promote `prometheus/client_golang` to direct dependency
- [x] 3.2 Create `internal/observability/prometheus.go` with PrometheusExporter (5 instruments)
- [x] 3.3 Add `/metrics/prometheus` route in `internal/app/routes_observability.go` (preserving `/metrics` JSON)
- [x] 3.4 Wire exporter in `internal/app/wiring_observability.go` when format=="prometheus"
- [x] 3.5 Link collector to exporter for tracked_sessions gauge updates
- [x] 3.6 Create `internal/observability/prometheus_test.go`

## 4. OpenTelemetry Tracing (Unit 17)

- [x] 4.1 Promote `go.opentelemetry.io/otel` + sdk/trace + stdout exporter to direct deps
- [x] 4.2 Create `internal/observability/tracing.go` with InitTracer (stdout/none)
- [x] 4.3 Add TracingConfig to `internal/config/types_observability.go`
- [x] 4.4 Create `internal/toolchain/mw_tracing.go` with WithTracing middleware
- [x] 4.5 Add WithTracing as outermost middleware (B4f) in `internal/app/app.go`
- [x] 4.6 Wire tracer initialization in `internal/app/wiring_observability.go`
- [x] 4.7 Register TracerShutdown in App.Stop() for span flush
- [x] 4.8 Update `internal/toolchain/chain_order_test.go` with tracing as outermost
- [x] 4.9 Create `internal/observability/tracing_test.go`

## 5. Verification

- [x] 5.1 `go build ./...` passes
- [x] 5.2 `grep log.Printf internal/` returns 0 non-test/non-ent results
- [x] 5.3 All affected package tests pass
