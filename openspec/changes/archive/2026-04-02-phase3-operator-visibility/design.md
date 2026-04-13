## Context

The runtime lacked standard observability integrations. Metrics were JSON-only, tracing was absent, alerts were internal-only (EventBus → audit), and several files used unstructured `log.Printf`. Phase 3 addresses all four gaps.

## Goals / Non-Goals

**Goals:**
- Prometheus metrics via `/metrics/prometheus` (event-driven, not snapshot-based)
- OpenTelemetry tool-level tracing (stdout exporter, outermost middleware)
- Webhook alert delivery (async, severity-filtered)
- Zero `log.Printf` in non-test, non-generated code

**Non-Goals:**
- OTLP exporter (deferred to next phase)
- Slack/Discord native delivery (webhook covers generic integration)
- TurnRunner/ADK-level span instrumentation (tool-level only in phase 3)

## Decisions

1. **`/metrics` JSON preserved, Prometheus at `/metrics/prometheus`** — CLI `metrics.go` parses `/metrics` JSON directly. Breaking that would require CLI changes. Prometheus gets its own route, controlled by `observability.metrics.format == "prometheus"`.

2. **Event-driven Prometheus, not snapshot-based** — Exporter subscribes to EventBus events and updates counters/gauges in real-time rather than periodically scraping `Snapshot()`. This avoids stale data and race conditions.

3. **`lango_tracked_sessions` not `active_sessions`** — The collector tracks sessions that have recorded token usage, not "currently active" sessions. The gauge is updated from token events by querying the collector snapshot.

4. **Tracing middleware outermost** — Placed after ExecPolicy (B4f) so blocked calls are also traced. Operators can see "which tools were blocked and how often" in traces.

5. **TracingConfig: stdout/none only** — Phase 3 scope. OTLP requires additional config (endpoint, auth) deferred to avoid scope creep.

6. **Async webhook delivery** — EventBus.Publish is synchronous. Webhook Send runs in a goroutine to avoid blocking the event pipeline on slow/unreachable endpoints.

7. **TUI delivery preservation** — When saving webhook URL from TUI, only update/remove the webhook entry, preserving other channel types in the delivery slice.

## Risks / Trade-offs

- **[Risk] Async delivery drops errors silently** → Mitigated: errors are logged via structured logging.
- **[Risk] Tracer batcher may drop spans on crash** → Mitigated: `TracerShutdown` registered in `App.Stop()` for graceful flush.
- **[Trade-off] tracked_sessions gauge accuracy** → Updated on token events, may lag slightly. Acceptable for monitoring purposes.
