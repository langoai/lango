## Why

Lango's agent runtime needs generic stream combinator primitives to compose concurrent event streams. The existing `asyncbuf` package handles buffered async operations, but there is no reusable library for merging, racing, or fan-in of typed iterator-based streams. As multi-agent orchestration grows, combinators over `iter.Seq2[T, error]` are needed to avoid ad-hoc goroutine/channel patterns scattered across the codebase.

## What Changes

- Introduce `internal/streamx/` package with generic stream combinators
- Define `Stream[T]` type alias over `iter.Seq2[T, error]` and `Tag[T]` wrapper
- Implement `Merge` — yields tagged events from N named streams as they arrive
- Implement `Race` — yields events from the first stream to produce a value, cancels others
- Implement `FanIn` — collects all events from N streams grouped by name
- Implement `Drain` — consumes a stream into a slice
- Full test coverage for all combinators

## Capabilities

### New Capabilities
- `stream-combinators`: Generic stream combinator primitives (Merge, Race, FanIn, Drain) built on Go iterators with context cancellation and clean goroutine lifecycle

### Modified Capabilities

## Impact

- New package: `internal/streamx/`
- No external dependencies beyond stdlib
- No changes to existing packages — purely additive
- Future consumers: `orchestration`, `turnrunner`, `agentrt`, multi-agent patterns
