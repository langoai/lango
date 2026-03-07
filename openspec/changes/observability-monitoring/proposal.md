## Why

All three LLM providers (OpenAI, Anthropic, Gemini) return token usage data that the streaming code completely discards. There is no agent-level call tracking, no session-level cost calculation, no tool execution performance metrics, and no system health monitoring. This makes it impossible to understand resource consumption, optimize costs, or diagnose performance issues.

## What Changes

- Capture actual token usage from all three providers (OpenAI, Anthropic, Gemini) via streaming events
- Add `Usage` struct to `StreamEvent` for backward-compatible token data propagation
- Create in-memory `MetricsCollector` for real-time aggregation (sessions, agents, tools)
- Add `TokenTracker` that subscribes to event bus and forwards to collector + persistent store
- Create model pricing table (`CostCalculator`) for estimated cost per request
- Add health check registry with built-in checks (database, memory, provider)
- Fix `ToolExecutedEvent.Duration` field (was always zero)
- Create Ent schema `token_usage` for persistent token usage history
- Wire `ModelAdapter.OnTokenUsage` callback to publish `TokenUsageEvent` to event bus
- Add `ObservabilityConfig` with nested tokens/health/audit/metrics settings
- Add Gateway API endpoints: `/metrics`, `/metrics/sessions`, `/metrics/tools`, `/metrics/agents`, `/metrics/cost`, `/metrics/history`, `/health/detailed`
- Add CLI commands: `lango metrics [sessions|tools|agents|cost|history]`
- Add `AuditRecorder` that writes tool calls and token usage to existing `AuditLog` Ent schema

## Capabilities

### New Capabilities
- `observability`: Token usage capture, metrics collection, cost calculation, health checks, audit recording, CLI/API exposure

### Modified Capabilities

## Impact

- `internal/provider/provider.go`: New `Usage` struct + field on `StreamEvent`
- `internal/provider/openai/`, `anthropic/`, `gemini/`: Token capture in streaming loops
- `internal/adk/model.go`: `OnTokenUsage` callback on `ModelAdapter`
- `internal/observability/`: New package (types, collector, routes)
- `internal/observability/token/`: Tracker, cost calculator, persistent store
- `internal/observability/health/`: Health check registry and built-in checks
- `internal/observability/audit/`: Audit recorder wiring existing AuditLog
- `internal/eventbus/`: New `TokenUsageEvent`
- `internal/toolchain/hook_eventbus.go`: Fix Duration, add PreToolHook
- `internal/config/types_observability.go`: New config types
- `internal/app/`: Wiring, routes, lifecycle registration
- `internal/ent/schema/token_usage.go`: New Ent schema
- `internal/cli/metrics/`: New CLI command group
- `cmd/lango/main.go`: Register metrics CLI command
