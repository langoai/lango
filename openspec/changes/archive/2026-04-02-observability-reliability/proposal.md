## Why

Two reliability issues in the observability stack: (1) `TestHeaderRoundTripper` sends a real HTTP request to RFC 5737 documentation address `192.0.2.1:1`, hanging on some networks and causing flaky CI failures. (2) `MetricsCollector` session map grows without bound — in long-running deployments with many sessions, this is a memory leak.

## What Changes

- Replace real HTTP request in MCP test with `capturingRoundTripper` mock that captures the request and returns a canned response — test verifies headers only, no network dependency
- Add `DefaultMaxSessions` constant (10,000) and `MaxSessions` field to `MetricsCollector`
- Add `evictOldestSession()` LRU eviction called before inserting new sessions in `RecordTokenUsage`
- Add `LastUpdated time.Time` field to `SessionMetric` for eviction ordering
- Zero or negative `MaxSessions` disables the cap

## Capabilities

### Modified Capabilities

- `observability-metrics`: Session tracking now bounded with configurable LRU eviction
- `mcp-testing`: Header round-tripper test no longer requires network access

## Impact

- `internal/mcp/connection_test.go` — mock round-tripper replacing real HTTP
- `internal/observability/collector.go` — `MaxSessions`, `evictOldestSession()`, eviction call in `RecordTokenUsage`
- `internal/observability/collector_test.go` — eviction tests (cap reached, oldest evicted, disabled cap)
- `internal/observability/types.go` — `LastUpdated` field on `SessionMetric`
