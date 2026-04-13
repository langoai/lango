# Observability Reliability — Tasks

- [x] 1.1 Replace real HTTP in `TestHeaderRoundTripper` with `httptest.NewServer()`
- [x] 1.2 Add `LastUpdated` field to `SessionMetric` in `types.go`
- [x] 1.3 Add `DefaultMaxSessions` constant and `MaxSessions` field to `MetricsCollector`
- [x] 1.4 Implement `evictOldestSession()` with LRU linear scan
- [x] 1.5 Call eviction in `RecordTokenUsage` before new session insertion
- [x] 1.6 Set `MaxSessions: DefaultMaxSessions` in `NewCollector()`
- [x] 2.1 Test MCP header round-tripper without network
- [x] 2.2 Test eviction triggers at capacity
- [x] 2.3 Test eviction removes oldest session
- [x] 2.4 Test disabled cap (MaxSessions=0)
