## 1. Delete Cost Calculator Files

- [x] 1.1 Delete `internal/observability/token/cost.go` (pricing table, Calculate, GetPricing, RegisterPricing)
- [x] 1.2 Delete `internal/observability/token/cost_test.go`
- [x] 1.3 Delete `internal/cli/metrics/cost.go` (lango metrics cost CLI command)

## 2. Remove EstimatedCost from Types

- [x] 2.1 Remove `EstimatedCost float64` from `TokenUsage`, `AgentMetric`, `SessionMetric`, `TokenUsageSummary` in `internal/observability/types.go`

## 3. Remove Cost from Collector

- [x] 3.1 Remove `EstimatedCost` accumulation from `RecordTokenUsage()` in `internal/observability/collector.go`
- [x] 3.2 Remove `EstimatedCost` fields and assertions from `internal/observability/collector_test.go`

## 4. Remove Cost from Tracker

- [x] 4.1 Remove `costCalc` field and parameter from `Tracker` and `NewTracker` in `internal/observability/token/tracker.go`
- [x] 4.2 Remove cost calculation logic from `handle()` method
- [x] 4.3 Update `internal/observability/token/tracker_test.go` — remove costFn mock, wantCost, update NewTracker calls

## 5. Update Wiring

- [x] 5.1 Remove `token.Calculate` argument from `NewTracker` call in `internal/app/wiring_observability.go`

## 6. Remove Cost from API Responses

- [x] 6.1 Remove `estimatedCost` from `/metrics`, `/metrics/sessions`, `/metrics/agents` responses in `internal/app/routes_observability.go`
- [x] 6.2 Delete entire `/metrics/cost` endpoint
- [x] 6.3 Remove `estimatedCost` and `totalCost` from `/metrics/history` response

## 7. Remove Cost from CLI

- [x] 7.1 Remove `newCostCmd()` registration from `internal/cli/metrics/metrics.go`
- [x] 7.2 Remove `Estimated Cost` line from summary output
- [x] 7.3 Remove `EstimatedCost` field and COST column from `internal/cli/metrics/sessions.go`
- [x] 7.4 Remove `EstimatedCost` field and COST column from `internal/cli/metrics/agents.go`
- [x] 7.5 Remove `EstimatedCost` field, `Cost:` output, and COST column from `internal/cli/metrics/history.go`

## 8. Update Ent Schema and Store

- [x] 8.1 Remove `field.Float("estimated_cost")` from `internal/ent/schema/token_usage.go`
- [x] 8.2 Run `go generate ./internal/ent` to regenerate Ent code
- [x] 8.3 Remove `SetEstimatedCost()` call from `Save()` in `internal/observability/token/store.go`
- [x] 8.4 Remove `TotalCost` from `AggregateResult` and aggregation logic
- [x] 8.5 Remove `EstimatedCost` mapping from `toTokenUsages()`

## 9. Verification

- [x] 9.1 Run `go build ./...` — no compilation errors
- [x] 9.2 Run `go test ./...` — all tests pass
- [x] 9.3 Grep for remaining `EstimatedCost`/`estimatedCost`/`estimated_cost`/`costCalc`/`TotalCost` references — zero matches in `internal/`
