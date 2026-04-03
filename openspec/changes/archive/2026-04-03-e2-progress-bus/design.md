## Context

The system needs a lightweight, concurrent-safe mechanism for publishing and subscribing to progress events from heterogeneous sources: tools (`tool:web_search`), agents (`agent:operator`), and background tasks (`bg:task-123`). The existing `eventbus.Bus` is designed for domain-level application events and is not suited for high-frequency, fire-and-forget progress signals.

## Goals / Non-Goals

**Goals:**
- Provide a simple pub/sub bus for progress events with prefix-based filtering
- Support non-blocking emit so publishers are never stalled by slow consumers
- Allow multiple concurrent subscribers with independent filters
- Safe for concurrent use from multiple goroutines

**Non-Goals:**
- Persistence or replay of events
- Backpressure signaling to publishers
- Integration with eventbus.Bus

## Design

### ProgressType Constants

Four progress lifecycle types are defined as `ProgressType string`:

| Constant            | Value       | Meaning                          |
|---------------------|-------------|----------------------------------|
| `ProgressStarted`   | `"started"` | Operation has begun              |
| `ProgressUpdate`    | `"update"`  | Intermediate progress tick       |
| `ProgressCompleted` | `"completed"` | Operation finished successfully |
| `ProgressFailed`    | `"failed"`  | Operation terminated with error  |

### ProgressEvent Struct

```
ProgressEvent {
    Source   string          // namespaced identifier, e.g. "tool:web_search", "agent:operator", "bg:task-123"
    Type     ProgressType    // lifecycle phase
    Message  string          // human-readable progress text
    Progress float64         // 0.0 to 1.0, or -1 if indeterminate
    Metadata map[string]any  // optional additional data
}
```

The `Source` field uses a colon-separated namespace convention (`category:name`) enabling prefix-based subscription filtering.

### ProgressBus

The `ProgressBus` manages a slice of subscribers protected by `sync.RWMutex`:

- **`NewProgressBus()`** — Constructor, returns an empty bus.
- **`Emit(event ProgressEvent)`** — Acquires read lock, iterates subscribers. For each non-closed subscriber whose filter is a prefix of `event.Source` (or empty filter for all), attempts a non-blocking send on the subscriber's buffered channel. If the channel buffer is full, the event is silently dropped.
- **`Subscribe(filter string) (<-chan ProgressEvent, func())`** — Creates a subscriber with a buffered channel (capacity 64) and the given prefix filter. Returns a receive-only channel and a cancel function. The cancel function acquires the write lock, marks the subscriber closed, closes the channel, and removes it from the subscriber slice. Double-cancel is safe (no-op on already-closed subscriber).
- **`SubscribeAll() (<-chan ProgressEvent, func())`** — Convenience wrapper that calls `Subscribe("")` with an empty filter, matching all events.

### Subscriber Management

Each subscriber is a struct with:
- `filter string` — prefix to match against `event.Source`
- `ch chan ProgressEvent` — buffered channel (cap 64)
- `closed bool` — guard against double-close

On cancel, the subscriber is removed from the slice by identity comparison (pointer equality), preventing stale references.

### Concurrency Model

- `Emit` uses `RLock` — multiple goroutines can emit concurrently
- `Subscribe` and cancel functions use full `Lock` — subscriber list mutations are serialized
- Non-blocking channel sends prevent publisher goroutines from blocking on slow consumers
- Buffer capacity of 64 provides reasonable headroom for burst events

## Decisions

| Decision | Rationale |
|---|---|
| Standalone type, not wrapping eventbus.Bus | eventbus.Bus is for domain events with different lifecycle. Progress events are high-frequency, fire-and-forget. |
| Non-blocking emit with silent drop | Publishers must never block. Dropped events are acceptable for progress signals. |
| Prefix-based filtering | Matches the `category:name` source convention. Simple, zero-allocation matching via `strings.HasPrefix`. |
| Buffer capacity 64 | Balances memory usage with burst tolerance. Covers typical tool execution batches. |
| Pointer identity for subscriber removal | Avoids the need for subscriber IDs. Each `Subscribe` call returns a unique `*subscriber`. |

## Risks / Trade-offs

- **[Trade-off]** Silent event drop when buffer is full — consumers that fall behind lose events with no notification. Acceptable for progress display purposes.
- **[Trade-off]** Linear subscriber iteration in Emit — O(n) per emit where n is subscriber count. Acceptable given expected subscriber counts (< 10).
