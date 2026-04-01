# Operational Alerting

## Overview

The alerting system monitors operational signals (policy decisions, recovery events, budget usage) and generates alerts when configurable thresholds are exceeded. Alerts flow through a 3-tier delivery pipeline:

1. **EventBus** -- `AlertEvent` published for real-time subscribers
2. **Audit log** -- persisted to the database for historical queries
3. **CLI** -- `lango alerts list` and `lango alerts summary` for operators

## Architecture

```
PolicyDecisionEvent ──▶ Alerting Dispatcher ──▶ AlertEvent ──▶ EventBus
                         (sliding window)                        │
                         (deduplication)                          ├──▶ Audit Recorder ──▶ DB
                                                                 └──▶ Other subscribers

lango alerts list ──▶ GET /alerts ──▶ Query audit DB (action="alert")
```

## Alert Conditions

| Condition | Type | Severity | Trigger |
|-----------|------|----------|---------|
| Policy block rate | `policy_block_rate` | warning | Block count exceeds threshold in 5min window |
| Recovery retries | `recovery_retries` | warning | Retry count exceeds threshold per session |
| Circuit breaker | `circuit_breaker` | critical | Circuit breaker tripped |
| Config drift | `config_drift` | warning | Configuration or provenance drift detected |

## Deduplication

The dispatcher deduplicates alerts by type within each 5-minute window. Only one alert per type per window is published. This prevents alert storms when a persistent condition repeatedly triggers the threshold.

## Configuration

```yaml
alerting:
  enabled: true                      # Master switch (default: false)
  policyBlockRateThreshold: 10       # Max blocks per 5min window
  recoveryRetryThreshold: 5          # Max retries per session
  adminChannel: ""                   # Optional: route to configured channel
```

All thresholds are configurable. The system is disabled by default and must be explicitly enabled.

## CLI Usage

```bash
# List recent alerts
lango alerts list --days=7

# Alert summary by type
lango alerts summary

# JSON output
lango alerts list --output json
```

## HTTP API

```
GET /alerts?days=7
```

Returns:
```json
{
  "alerts": [
    {
      "id": "uuid",
      "type": "policy_block_rate",
      "actor": "system",
      "details": {
        "severity": "warning",
        "message": "policy block rate exceeded threshold",
        "count": 12,
        "threshold": 10,
        "window": "5m0s"
      },
      "timestamp": "2026-04-01T12:00:00Z"
    }
  ],
  "total": 1,
  "days": 7
}
```
