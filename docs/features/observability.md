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

## Policy Metrics

The metrics collector tracks exec policy decisions (block and observe verdicts) published via the event bus as `PolicyDecisionEvent`. Allow verdicts are not tracked.

Collected counters:

- **Blocks** -- Total commands blocked by exec policy
- **Observes** -- Total commands flagged for observation
- **By Reason** -- Per-reason breakdown (e.g., `catastrophic_pattern`, `destructive_command`)

The collector's `RecordPolicyDecision(verdict, reason)` method aggregates these counters in memory. They are included in the `SystemSnapshot` used by the gateway endpoint and CLI command.

**Gateway endpoint:** `GET /metrics/policy`

**Response format:**

```json
{
  "blocks": 3,
  "observes": 12,
  "byReason": {
    "catastrophic_pattern": 2,
    "destructive_command": 1,
    "network_exfiltration": 5,
    "suspicious_pipe": 7
  }
}
```

**CLI command:** `lango metrics policy` (see [CLI Reference](../cli/metrics.md#lango-metrics-policy))

## Audit Logging

The audit recorder subscribes to event bus events and writes audit log entries to the database:

- **Tool execution events** -- Records tool name, duration, success/failure, and error details via `ToolExecutedEvent`
- **Token usage events** -- Records provider, model, and token counts via `TokenUsageEvent`
- **Policy decision events** -- Records exec policy block/observe verdicts via `PolicyDecisionEvent`
- Default retention: 90 days

### Policy Decision Audit Logging

When the exec policy evaluator blocks or flags a command, it publishes a `PolicyDecisionEvent` on the event bus. The audit recorder subscribes to these events and writes a database entry with:

| Field | Source | Description |
|-------|--------|-------------|
| `action` | `policy_decision` | Audit log action type |
| `actor` | `PolicyDecisionEvent.AgentName` | Agent that attempted the command (or `"system"`) |
| `target` | `PolicyDecisionEvent.Command` | The original command string |
| `details.verdict` | `PolicyDecisionEvent.Verdict` | `"block"` or `"observe"` |
| `details.reason` | `PolicyDecisionEvent.Reason` | Machine-readable reason code |
| `details.unwrapped` | `PolicyDecisionEvent.Unwrapped` | Command after shell wrapper unwrap |
| `details.message` | `PolicyDecisionEvent.Message` | Human-readable explanation (if present) |

This enables operators to query the audit log for all policy decisions in a session and correlate them with tool execution events.

## Recovery Decision Events

When the coordinating executor handles an agent execution failure, it publishes a `RecoveryDecisionEvent` on the event bus with structured metadata for observability. This event is published for every recovery decision (retry, retry with hint, direct answer, escalate).

**Event fields:**

| Field | Type | Description |
|-------|------|-------------|
| `CauseClass` | string | Error classification: `rate_limit`, `transient`, `malformed_tool_call`, `timeout`, or empty for non-agent errors |
| `Action` | string | Recovery decision: `retry`, `retry_with_hint`, `direct_answer`, `escalate`, or `none` |
| `Attempt` | int | Current retry attempt number (0-based) |
| `Backoff` | duration | Computed backoff duration before the next retry (zero for non-retry actions) |
| `SessionKey` | string | Session identifier |

**Event name:** `agent.recovery.decision`

### Exponential Backoff

Retry actions use exponential backoff before the next attempt. The formula is:

```
backoff = min(baseDelay * 2^attempt, maxBackoff)
```

| Parameter | Value |
|-----------|-------|
| Base delay | 1 second |
| Max backoff | 30 seconds |

Example progression: 1s, 2s, 4s, 8s, 16s, 30s, 30s, ...

Backoff sleeps are context-aware and will abort immediately if the context is cancelled.

### Per-Error-Class Retry Limits

In addition to the global `maxRetries` setting (default: 2), each error class has its own maximum retry count. When the per-class limit is reached for a specific error type, the recovery policy escalates even if the global limit has not been reached.

| Cause Class | Default Max Retries | Description |
|-------------|---------------------|-------------|
| `rate_limit` | 5 | Provider rate-limiting (429 responses) |
| `transient` | 3 | Transient provider errors |
| `malformed_tool_call` | 1 | Invalid function call schema |
| `timeout` | 3 | Execution or idle timeout |
| _(other)_ | Global `maxRetries` | Falls through to the global setting |

Per-class retry counts are tracked independently within a single run. The global `maxRetries` is configured via `recovery.maxRetries` in the config.

## Gateway Endpoints

All observability endpoints are available when the gateway is running (`lango serve`):

| Endpoint | Description |
|----------|-------------|
| `GET /metrics` | System metrics snapshot (goroutines, memory, uptime) |
| `GET /metrics/sessions` | Per-session token usage |
| `GET /metrics/tools` | Per-tool metrics |
| `GET /metrics/agents` | Per-agent metrics |
| `GET /metrics/policy` | Policy decision statistics (blocks, observes, by-reason) |
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
