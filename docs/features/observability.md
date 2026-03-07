---
title: Observability
---

# Observability

!!! warning "Experimental"

    The observability system is experimental. Metrics format and gateway endpoints may change in future releases.

Lango includes an observability subsystem for metrics collection, token usage tracking, health monitoring, and audit logging. All data is accessible through gateway HTTP endpoints when running `lango serve`.

## Metrics Collector

The metrics collector provides a system-level snapshot including:

- Goroutine count, memory usage, and process uptime
- Per-session, per-agent, and per-tool breakdowns
- Request counts and latency distributions

**Gateway endpoint:** `GET /metrics`

## Token Tracking

Token tracking records LLM provider token usage via the event bus (`TokenUsageEvent`). Usage data is stored in an Ent-backed persistent store with configurable retention.

- Subscribes to `token.usage` events from the event bus
- Tracks input, output, cache, and total tokens per session/agent/model
- Configurable retention period (default: 30 days)
- Supports historical queries by time range

**Gateway endpoints:**

| Endpoint | Description |
|----------|-------------|
| `GET /metrics/sessions` | Per-session token usage |
| `GET /metrics/tools` | Per-tool metrics |
| `GET /metrics/agents` | Per-agent metrics |
| `GET /metrics/history` | Historical metrics (`?days=N` parameter) |

## Health Checks

The health check system uses a registry-based architecture where components register their own health check functions.

- Built-in memory check (512 MB threshold)
- Configurable check interval
- Returns per-component status with details

**Gateway endpoint:** `GET /health/detailed`

## Audit Logging

The audit recorder subscribes to event bus events and writes audit log entries to the database:

- **Tool execution events** -- Records tool name, duration, success/failure, and error details via `ToolExecutedEvent`
- **Token usage events** -- Records provider, model, and token counts via `TokenUsageEvent`
- Default retention: 90 days

## Gateway Endpoints

All observability endpoints are available when the gateway is running (`lango serve`):

| Endpoint | Description |
|----------|-------------|
| `GET /metrics` | System metrics snapshot (goroutines, memory, uptime) |
| `GET /metrics/sessions` | Per-session token usage |
| `GET /metrics/tools` | Per-tool metrics |
| `GET /metrics/agents` | Per-agent metrics |
| `GET /metrics/history` | Historical metrics (`?days=N` parameter) |
| `GET /health/detailed` | Detailed health check results per component |

## Configuration

> **Settings:** `lango settings` -> Observability

```json
{
  "observability": {
    "enabled": true,
    "tokens": {
      "enabled": true,
      "persistHistory": true,
      "retentionDays": 30
    },
    "health": {
      "enabled": true,
      "interval": "30s"
    },
    "audit": {
      "enabled": true,
      "retentionDays": 90
    },
    "metrics": {
      "enabled": true,
      "format": "json"
    }
  }
}
```

| Key | Default | Description |
|-----|---------|-------------|
| `observability.enabled` | `false` | Activates the observability subsystem |
| `observability.tokens.enabled` | `true` | Activates token tracking (when observability is enabled) |
| `observability.tokens.persistHistory` | `false` | Enables DB-backed persistent storage |
| `observability.tokens.retentionDays` | `30` | Days to keep token usage records |
| `observability.health.enabled` | `true` | Activates health checks (when observability is enabled) |
| `observability.health.interval` | `30s` | Health check interval |
| `observability.audit.enabled` | `false` | Activates audit logging |
| `observability.audit.retentionDays` | `90` | Days to keep audit records |
| `observability.metrics.enabled` | `false` | Activates metrics export endpoint |
| `observability.metrics.format` | `"json"` | Metrics export format |

See the [Metrics CLI Reference](../cli/metrics.md) for command documentation.
