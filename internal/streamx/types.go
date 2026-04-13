// Package streamx provides generic stream combinator primitives built on Go iterators.
//
// Stream[T] is a typed event stream defined as iter.Seq2[T, error]. Combinators
// such as Merge, Race, FanIn, and Drain compose streams with context-aware
// cancellation and clean goroutine lifecycle.
package streamx

import "iter"

// Stream is a typed event stream built on Go iterators.
// Each iteration yields either a valid event T or a non-nil error.
type Stream[T any] iter.Seq2[T, error]

// Tag wraps an event with its source identifier.
type Tag[T any] struct {
	Source string
	Event  T
}
