## ADDED Requirements

### Requirement: Stream Type

The package SHALL define `Stream[T any]` as a type alias for `iter.Seq2[T, error]` representing a typed event stream.

#### Scenario: Stream yields events and errors
WHEN a Stream is iterated
THEN it SHALL yield `(T, error)` pairs where either T contains an event or error is non-nil

### Requirement: Tag Type

The package SHALL define `Tag[T any]` struct with `Source string` and `Event T` fields to wrap events with source identifiers.

#### Scenario: Tag wraps event with source
WHEN an event is tagged
THEN the Tag SHALL contain the source name and the original event value

### Requirement: Merge Combinator

The package SHALL provide `Merge[T any](ctx context.Context, streams map[string]Stream[T]) Stream[Tag[T]]` that yields tagged events from all input streams as they arrive.

#### Scenario: Merge yields from multiple streams
WHEN Merge is called with N named streams
THEN it SHALL yield all events from all streams tagged with their source name

#### Scenario: Merge with empty streams map
WHEN Merge is called with an empty streams map
THEN it SHALL yield zero events

#### Scenario: Merge context cancellation
WHEN the context is cancelled during Merge iteration
THEN all internal goroutines SHALL stop and no goroutine leaks SHALL occur

#### Scenario: Merge error propagation
WHEN a stream yields an error
THEN Merge SHALL propagate the error through the iterator

### Requirement: Race Combinator

The package SHALL provide `Race[T any](ctx context.Context, streams map[string]Stream[T]) Stream[Tag[T]]` that yields events from the first stream to produce a value, then cancels others.

#### Scenario: Race yields first result
WHEN Race is called with N named streams
THEN it SHALL yield only the event(s) from the first stream to produce a value

#### Scenario: Race cancels losers
WHEN one stream produces a value
THEN Race SHALL cancel the context for all other streams

#### Scenario: Race with empty streams map
WHEN Race is called with an empty streams map
THEN it SHALL yield zero events

#### Scenario: Race context cancellation
WHEN the context is cancelled before any stream yields
THEN Race SHALL stop cleanly with no goroutine leaks

#### Scenario: Race error propagation
WHEN the winning stream yields an error
THEN Race SHALL propagate the error

### Requirement: FanIn Combinator

The package SHALL provide `FanIn[T any](ctx context.Context, streams map[string]Stream[T]) (map[string][]T, error)` that collects all events from all streams grouped by source name.

#### Scenario: FanIn collects all events
WHEN FanIn is called with N named streams
THEN it SHALL return a map where each key is a source name and the value is all events from that stream

#### Scenario: FanIn with empty streams
WHEN FanIn is called with an empty streams map
THEN it SHALL return an empty map and nil error

#### Scenario: FanIn error stops collection
WHEN any stream yields an error
THEN FanIn SHALL return that error and stop collecting

### Requirement: Drain Combinator

The package SHALL provide `Drain[T any](stream Stream[T]) ([]T, error)` that consumes a stream and returns all events as a slice.

#### Scenario: Drain collects all events
WHEN Drain is called on a stream with N events
THEN it SHALL return a slice of all N events

#### Scenario: Drain empty stream
WHEN Drain is called on a stream that yields zero events
THEN it SHALL return an empty slice and nil error

#### Scenario: Drain error propagation
WHEN the stream yields an error after K events
THEN Drain SHALL return the K collected events and the error

### Requirement: Goroutine Safety

All combinators that use goroutines (Merge, Race, FanIn) SHALL ensure no goroutine leaks on normal completion, error, or context cancellation.

#### Scenario: No goroutine leak on cancellation
WHEN a context is cancelled mid-stream
THEN all internal goroutines SHALL terminate within a bounded time
