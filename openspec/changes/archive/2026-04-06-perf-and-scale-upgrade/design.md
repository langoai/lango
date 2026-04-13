## Context

After completing Phase 1-5, analysis of cockpit TUI performance characteristics revealed that render() O(n) loops and per-entry style allocation cause 5-15ms/render latency at 1000+ entries, and unbounded growth of transcript/task/grant causes memory issues in long-running sessions.

## Goals / Non-Goals

**Goals:**
- Eliminate per-render style allocation cost via entry block memoization
- Eliminate NewStyle() calls via module-level style pre-allocation
- Bound memory and render cost with transcript entry cap (2000)
- Bound BackgroundManager memory with terminal task cap (500)
- Bound GrantStore memory with grant lazy cleanup
- Eliminate re-styling during scroll with approval diff styled line cache

**Non-Goals:**
- Viewport virtualization (requires changes to Bubble Tea viewport model itself — out of scope)
- Adding EventBus Unsubscribe (not a practical problem)
- MetricsCollector/ChannelTracker cleanup (already bounded)

## Decisions

### D1: Entry-level memoization via cachedBlock field
- Add cachedBlock string to transcriptItem
- In render(), use cache if cachedBlock != ""
- Invalidate on width change, tool finalize, thinking finalize
- **Why**: strings.Join + viewport.SetContent costs remain, but the most expensive parts (style creation + markdown render) are eliminated

### D2: Module-level style pre-allocation (sidebar.go pattern)
- lipgloss.Style is an immutable value type — sharing a base style only produces copies (no allocation)
- Variable properties like colors are maintained via per-call chaining
- **Why**: Proven pattern (sidebar.go already uses it)

### D3: Terminal task FIFO cap (eviction, not deletion)
- maxTerminalTasks = 500, oldest eviction based on CompletedAt
- Status/Result/retry work normally within the cap
- **Why**: TTL deletion breaks Status/Result/retry — cap provides the same memory bound while minimizing UX impact

### D4: Lazy cleanup in GrantStore List()
- Extract cleanExpiredLocked() internal helper (no lock required)
- List() upgrades from RLock → Lock, then performs cleanup + listing
- **Why**: Cross-calling RLock and Lock causes deadlock — Lock upgrade is necessary

### D5: Diff line cache — exclude scrollOffset from key
- Cache key: content + width + splitMode
- Scroll only windows into the cached lines slice
- **Why**: 100% cache hit on the most expensive interaction (scrolling)

### D6: Transcript trimming — preserve in-flight entries + new backing array
- Collect active tool/thinking entries from the trim range and preserve them after the tombstone
- Create a new slice with make() so the old backing array becomes eligible for GC
- Accumulate tombstone count (adjust boundary based on tombstone presence)
- **Why**: Guarantees that finalizeToolResult/finalizeThinking can always find active entries

## Risks / Trade-offs

- [Risk] Terminal task cap exceeded (500) → oldest task not found → Mitigation: Only occurs after 500+ completed tasks; documented
- [Risk] GrantStore List() Lock upgrade reduces read concurrency → Mitigation: List() is only called on UI ticks (every 2 seconds); not a bottleneck
- [Risk] cachedBlock increases memory usage → Mitigation: Entry trimming limits to 2000 cap; old backing array is released via new backing array
