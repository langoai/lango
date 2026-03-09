## 1. Provider Token Capture

- [ ] 1.1 Add `Usage` struct and `Usage *Usage` field to `StreamEvent` in `internal/provider/provider.go`
- [ ] 1.2 Capture OpenAI token usage with `StreamOptions.IncludeUsage` in `internal/provider/openai/openai.go`
- [ ] 1.3 Capture Anthropic token usage from `stream.Message.Usage` in `internal/provider/anthropic/anthropic.go`
- [ ] 1.4 Capture Gemini token usage from `resp.UsageMetadata` in `internal/provider/gemini/gemini.go`

## 2. Observability Core

- [ ] 2.1 Create `internal/observability/types.go` with TokenUsage, ToolMetric, AgentMetric, SessionMetric, SystemSnapshot types
- [ ] 2.2 Create `internal/observability/collector.go` with thread-safe MetricsCollector
- [ ] 2.3 Write table-driven tests for MetricsCollector in `collector_test.go`

## 3. Health Check System

- [ ] 3.1 Create `internal/observability/health/types.go` with Checker interface and Status types
- [ ] 3.2 Create `internal/observability/health/registry.go` with HealthRegistry
- [ ] 3.3 Create `internal/observability/health/checks.go` with DatabaseCheck, MemoryCheck, ProviderCheck
- [ ] 3.4 Write tests for health registry in `registry_test.go`

## 4. Token Cost Calculator

- [ ] 4.1 Create `internal/observability/token/cost.go` with model pricing table and Calculate function
- [ ] 4.2 Write tests for cost calculator with prefix matching in `cost_test.go`

## 5. Event Bus Wiring

- [ ] 5.1 Create `TokenUsageEvent` in `internal/eventbus/observability_events.go`
- [ ] 5.2 Create `TokenTracker` in `internal/observability/token/tracker.go`
- [ ] 5.3 Write tests for TokenTracker in `tracker_test.go`

## 6. ModelAdapter Token Forwarding

- [ ] 6.1 Add `OnTokenUsage` callback to `ModelAdapter` in `internal/adk/model.go`
- [ ] 6.2 Forward `evt.Usage` to callback on `StreamEventDone` in both streaming and non-streaming paths

## 7. Persistent Token Storage

- [ ] 7.1 Create Ent schema `internal/ent/schema/token_usage.go` and run `go generate`
- [ ] 7.2 Create `EntTokenStore` in `internal/observability/token/store.go` with Save, Query, Aggregate, Cleanup

## 8. ToolExecutedEvent Duration Fix

- [ ] 8.1 Add PreToolHook to `EventBusHook` with `sync.Map` timing in `internal/toolchain/hook_eventbus.go`
- [ ] 8.2 Update wiring to register EventBusHook as both pre and post hook

## 9. Config and Wiring

- [ ] 9.1 Create `ObservabilityConfig` in `internal/config/types_observability.go`
- [ ] 9.2 Add `Observability` field to `Config` in `internal/config/types.go`
- [ ] 9.3 Create `initObservability()` and `wireModelAdapterTokenUsage()` in `internal/app/wiring_observability.go`
- [ ] 9.4 Wire observability into `app.New()` and pass event bus to `initAgent()`
- [ ] 9.5 Add observability fields to `App` struct in `types.go`
- [ ] 9.6 Register lifecycle component for token store cleanup

## 10. CLI and API Exposure

- [ ] 10.1 Create `lango metrics` CLI commands in `internal/cli/metrics/`
- [ ] 10.2 Create gateway API routes (`/metrics`, `/metrics/*`, `/health/detailed`) in `internal/app/routes_observability.go`
- [ ] 10.3 Register metrics CLI command in `cmd/lango/main.go`

## 11. Audit Recorder

- [ ] 11.1 Create `AuditRecorder` in `internal/observability/audit/recorder.go`
- [ ] 11.2 Wire audit recorder to event bus in `app.New()` when `audit.enabled`

## 12. Verification

- [ ] 12.1 Run `go build ./...` — all packages compile
- [ ] 12.2 Run `go test ./...` — all tests pass
