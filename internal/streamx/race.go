package streamx

import (
	"context"
	"sync/atomic"
)

// Race takes N named streams and yields all events from the first stream to
// produce a value. Once one stream wins, the others are cancelled via context
// and the winner is drained completely.
//
// Loser goroutines exit when their stream returns or when the parent context
// is cancelled. Callers should use a cancellable context if streams may block.
func Race[T any](ctx context.Context, streams map[string]Stream[T]) Stream[Tag[T]] {
	return func(yield func(Tag[T], error) bool) {
		if len(streams) == 0 {
			return
		}

		type result struct {
			source string
			event  T
			err    error
		}

		firstCh := make(chan result, 1)
		restCh := make(chan result, 64)

		var won atomic.Bool
		var remaining atomic.Int32
		remaining.Store(int32(len(streams)))

		_, loserCancel := context.WithCancel(ctx)
		defer loserCancel()

		for name, s := range streams {
			go func(name string, s Stream[T]) {
				isWinner := false
				defer func() {
					if isWinner {
						close(restCh)
					}
					if remaining.Add(-1) == 0 && !won.Load() {
						// All streams exhausted, no winner — unblock main.
						close(firstCh)
					}
				}()

				for v, err := range s {
					if !isWinner {
						if err != nil {
							if won.CompareAndSwap(false, true) {
								firstCh <- result{source: name, err: err}
								isWinner = true
								loserCancel()
							}
							return
						}
						if won.CompareAndSwap(false, true) {
							firstCh <- result{source: name, event: v}
							isWinner = true
							loserCancel()
							// Continue to drain remaining events.
							continue
						}
						// Lost the race.
						return
					}

					// Winner: subsequent events. Check parent ctx only.
					select {
					case <-ctx.Done():
						return
					default:
					}
					if err != nil {
						select {
						case restCh <- result{source: name, err: err}:
						default:
						}
						return
					}
					select {
					case restCh <- result{source: name, event: v}:
					case <-ctx.Done():
						return
					}
				}
			}(name, s)
		}

		r, ok := <-firstCh
		if !ok {
			return // all streams empty
		}

		if r.err != nil {
			yield(Tag[T]{}, r.err)
			return
		}
		if !yield(Tag[T]{Source: r.source, Event: r.event}, nil) {
			return
		}

		for r := range restCh {
			if r.err != nil {
				yield(Tag[T]{}, r.err)
				return
			}
			if !yield(Tag[T]{Source: r.source, Event: r.event}, nil) {
				return
			}
		}
	}
}
