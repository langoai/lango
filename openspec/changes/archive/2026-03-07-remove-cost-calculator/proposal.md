## Why

No LLM provider (OpenAI, Anthropic, Gemini) offers a model pricing API. Hardcoded price tables become inaccurate on every model release and inaccurate cost estimates are worse than none. The system should track token counts only and remove all cost-related code.

## What Changes

- **BREAKING**: Remove `EstimatedCost` field from all observability types (`TokenUsage`, `AgentMetric`, `SessionMetric`, `TokenUsageSummary`)
- **BREAKING**: Remove `/metrics/cost` HTTP endpoint
- **BREAKING**: Remove `lango metrics cost` CLI command
- **BREAKING**: Remove `estimated_cost` column from Ent `TokenUsage` schema
- Delete `internal/observability/token/cost.go` (model pricing table, `Calculate`, `GetPricing`, `RegisterPricing`)
- Remove `costCalc` parameter from `token.NewTracker`
- Remove cost-related columns from CLI table output (`sessions`, `agents`, `history`)
- Remove cost fields from all API JSON responses

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `observability`: Remove "Token cost estimation" requirement and all `EstimatedCost` references from metrics collection, API responses, and CLI output

## Impact

- **Code**: `internal/observability/`, `internal/cli/metrics/`, `internal/app/routes_observability.go`, `internal/app/wiring_observability.go`, `internal/ent/schema/token_usage.go`
- **APIs**: `/metrics`, `/metrics/sessions`, `/metrics/agents`, `/metrics/history` lose cost fields; `/metrics/cost` removed entirely
- **CLI**: `lango metrics cost` subcommand removed; cost columns removed from `sessions`, `agents`, `history` output
- **Database**: `estimated_cost` column removed from `token_usage` table (requires Ent schema regeneration)
- **Dependencies**: No new dependencies; no removed dependencies
