## Context

The MCP test `TestHeaderRoundTripper` sends a real HTTP request to `192.0.2.1:1` (RFC 5737 TEST-NET). On some networks, this hangs waiting for ICMP unreachable instead of failing fast, causing intermittent CI timeouts. Separately, `MetricsCollector` tracks sessions in an unbounded `map[string]*SessionMetric` — a memory leak in long-running deployments.

## Goals / Non-Goals

**Goals:**
- MCP header test runs deterministically without network I/O
- MetricsCollector session map has a configurable upper bound
- Eviction is simple, correct, and low-overhead

**Non-Goals:**
- Changing MCP transport architecture (test-only fix)
- Time-based session expiry (LRU by update time is sufficient)
- Heap-based eviction structure (linear scan at 10K cap is fast enough)

## Decisions

1. **`capturingRoundTripper` mock** — Custom `http.RoundTripper` implementation that stores the request and returns `http.StatusOK`. Rationale: simpler than `httptest.NewServer`, test only needs to inspect request headers.

2. **LRU by `LastUpdated`** — Linear scan for oldest session on eviction. Rationale: eviction only triggers when inserting a new session at capacity (rare), and 10K map scan is ~microseconds.

3. **Evict before insert** — Call `evictOldestSession()` before creating the new `SessionMetric`. Rationale: ensures the map never exceeds `MaxSessions`.

4. **Cap disabled by zero/negative** — `MaxSessions <= 0` skips eviction entirely. Rationale: backward compatible with existing behavior where no cap existed.

## Risks / Trade-offs

- [Linear eviction scan is O(n)] → Acceptable at 10K cap; only runs on new session insertion when at capacity. Upgrade to heap if profiling shows hot path.
- [`LastUpdated` not set on existing sessions] → Only new/updated sessions get timestamps; oldest sessions (time.Time zero value) are evicted first, which is the correct behavior.
