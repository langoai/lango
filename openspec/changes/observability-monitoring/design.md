## Context

Lango uses three LLM providers that all return token usage data, but this data is silently discarded in the streaming code. The event bus infrastructure already exists with 20+ event types, but no observability events. An Ent AuditLog schema exists but is unwired. The `/health` endpoint returns only `{"status":"ok"}`.

## Goals / Non-Goals

**Goals:**
- Capture real token usage from all providers without breaking existing streaming consumers
- Provide real-time in-memory metrics with optional DB persistence
- Expose metrics via both CLI commands and Gateway HTTP API
- Maintain zero-dependency approach (no OpenTelemetry/Prometheus required)
- Fix existing ToolExecutedEvent.Duration bug

**Non-Goals:**
- OpenTelemetry/Prometheus integration (future work)
- Distributed tracing across P2P calls
- Real-time dashboards or web UI
- Token usage estimation replacement (keep existing `memory/token.go`)

## Decisions

1. **In-memory collector + optional DB persistence** over pure DB storage.
   - Real-time aggregation with zero I/O latency for hot-path metrics.
   - Ent `token_usage` table for historical queries when `persistHistory: true`.
   - Alternative: Pure DB → too slow for per-request recording in streaming path.

2. **Callback on ModelAdapter** over direct event bus import in provider package.
   - Avoids import cycle: `provider` → `eventbus` would couple core to infra.
   - `TokenUsageCallback` type on `ModelAdapter` wired via closure in app wiring.
   - Alternative: Context-based propagation → too implicit, hard to test.

3. **`Usage *Usage` on StreamEvent** over separate channel.
   - Nil pointer is backward compatible — existing consumers simply ignore it.
   - No new goroutine or channel complexity.
   - Alternative: Separate usage stream → requires consumer changes everywhere.

4. **PreToolHook + sync.Map for Duration timing** over context-based approach.
   - EventBusHook implements both PreToolHook and PostToolHook.
   - Start times stored in `sync.Map` keyed by session+tool+agent.
   - Alternative: HookContext mutation → interface doesn't support writable state.

5. **Routes in `app` package** over `observability` package.
   - Avoids import cycle: `observability` → `observability/token` → `observability`.
   - Routes need both parent and child packages; `app` already imports both.

## Risks / Trade-offs

- [Memory growth] In-memory collector grows unbounded per-session → Reset() available, retention cleanup on shutdown for DB.
- [Cost accuracy] Pricing table is static, models get price updates → RegisterPricing() allows runtime updates, prefix matching for model variants.
- [Streaming overhead] Token capture adds nil check per chunk → negligible, only populated on Done event.
- [Audit volume] AuditRecorder writes per tool call + per token event → gated behind `observability.audit.enabled` config flag.
