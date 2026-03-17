## 1. Provider Token Capture

- [x] 1.1 Add `Usage` struct and `Usage *Usage` field to `StreamEvent` in `internal/provider/provider.go`
- [x] 1.2 Capture OpenAI token usage with `StreamOptions.IncludeUsage` in `internal/provider/openai/openai.go`
- [x] 1.3 Capture Anthropic token usage from `stream.Message.Usage` in `internal/provider/anthropic/anthropic.go`
- [x] 1.4 Capture Gemini token usage from `resp.UsageMetadata` in `internal/provider/gemini/gemini.go`

## 2. Observability Core

- [x] 2.1 Create `internal/observability/types.go` with TokenUsage, ToolMetric, AgentMetric, SessionMetric, SystemSnapshot types
- [x] 2.2 Create `internal/observability/collector.go` with thread-safe MetricsCollector
- [x] 2.3 Write table-driven tests for MetricsCollector in `collector_test.go`

## 3. Health Check System

- [x] 3.1 Create `internal/observability/health/types.go` with Checker interface and Status types
- [x] 3.2 Create `internal/observability/health/registry.go` with HealthRegistry
- [x] 3.3 Create `internal/observability/health/checks.go` with DatabaseCheck, MemoryCheck, ProviderCheck
- [x] 3.4 Write tests for health registry in `registry_test.go`

## 4. Token Cost Calculator

- [x] 4.1 ~~Create `internal/observability/token/cost.go` with model pricing table and Calculate function~~ (intentionally removed — see archive `2026-03-07-remove-cost-calculator`)
- [x] 4.2 ~~Write tests for cost calculator with prefix matching in `cost_test.go`~~ (removed with 4.1)

## 5. Event Bus Wiring

- [x] 5.1 Create `TokenUsageEvent` in `internal/eventbus/observability_events.go`
- [x] 5.2 Create `TokenTracker` in `internal/observability/token/tracker.go`
- [x] 5.3 Write tests for TokenTracker in `tracker_test.go`

## 6. ModelAdapter Token Forwarding

- [x] 6.1 Add `OnTokenUsage` callback to `ModelAdapter` in `internal/adk/model.go`
- [x] 6.2 Forward `evt.Usage` to callback on `StreamEventDone` in both streaming and non-streaming paths

## 7. Persistent Token Storage

- [x] 7.1 Create Ent schema `internal/ent/schema/token_usage.go` and run `go generate`
- [x] 7.2 Create `EntTokenStore` in `internal/observability/token/store.go` with Save, Query, Aggregate, Cleanup

## 8. ToolExecutedEvent Duration Fix

- [x] 8.1 Add PreToolHook to `EventBusHook` with `sync.Map` timing in `internal/toolchain/hook_eventbus.go`
- [x] 8.2 Update wiring to register EventBusHook as both pre and post hook

## 9. Config and Wiring

- [x] 9.1 Create `ObservabilityConfig` in `internal/config/types_observability.go`
- [x] 9.2 Add `Observability` field to `Config` in `internal/config/types.go`
- [x] 9.3 Create `initObservability()` and `wireModelAdapterTokenUsage()` in `internal/app/wiring_observability.go`
- [x] 9.4 Wire observability into `app.New()` and pass event bus to `initAgent()`
- [x] 9.5 Add observability fields to `App` struct in `types.go`
- [x] 9.6 Register lifecycle component for token store cleanup

## 10. CLI and API Exposure

- [x] 10.1 Create `lango metrics` CLI commands in `internal/cli/metrics/`
- [x] 10.2 Create gateway API routes (`/metrics`, `/metrics/*`, `/health/detailed`) in `internal/app/routes_observability.go`
- [x] 10.3 Register metrics CLI command in `cmd/lango/main.go`

## 11. Audit Recorder

- [x] 11.1 Create `AuditRecorder` in `internal/observability/audit/recorder.go`
- [x] 11.2 Wire audit recorder to event bus in `app.New()` when `audit.enabled`

## 12. Verification

- [x] 12.1 Run `go build ./...` — all packages compile
- [x] 12.2 Run `go test ./...` — all tests pass
