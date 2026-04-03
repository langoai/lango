## Overview

The `internal/streamx/` package provides generic stream combinator primitives built on Go's `iter.Seq2[T, error]` iterator protocol. All combinators are context-aware and ensure clean goroutine lifecycle (no leaks).

## Types

```go
type Stream[T any] iter.Seq2[T, error]

type Tag[T any] struct {
    Source string
    Event  T
}
```

`Stream[T]` is a push-based iterator that yields `(T, error)` pairs. `Tag[T]` wraps an event with its source identifier for multiplexed streams.

## Combinator Design

### Merge

- Launches one goroutine per named stream
- All goroutines send to a shared buffered channel (size = number of streams)
- Main loop reads from channel and yields tagged events via the iterator
- Context cancellation signals all goroutines to stop
- Uses `sync.WaitGroup` to wait for goroutine cleanup before closing channel

### Race

- Launches one goroutine per named stream
- Uses a derived context (`context.WithCancel`) so the first result cancels others
- First goroutine to send on the result channel wins
- Yields only the winning event (single iteration)
- Clean shutdown: cancel derived context, wait for all goroutines

### FanIn

- Launches one goroutine per named stream using `errgroup.Group`
- Each goroutine drains its stream into a local slice
- Results collected into `map[string][]T`
- First error from any stream causes the errgroup to cancel; function returns that error
- Uses `sync.Mutex` to protect shared result map

### Drain

- Simple: iterates the stream, appends each event to a slice
- Returns `([]T, error)` — stops on first error

## Goroutine Safety

- Every goroutine has a predictable stop condition (context cancellation or stream exhaustion)
- `sync.WaitGroup` or `errgroup` ensures no goroutine outlives the combinator call
- Channel operations respect context to avoid blocking on cancelled contexts

## Dependencies

- stdlib only: `context`, `iter`, `sync`, `sync/errgroup` (part of stdlib in Go 1.25)
