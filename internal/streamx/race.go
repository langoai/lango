package streamx

import (
	"context"
	"sync"
)

type raceResult[T any] struct {
	tag Tag[T]
	err error
}

// Race takes N named streams and yields events from the first stream to produce
// a value. Once one stream yields, the others are cancelled. The winning stream
// is drained completely (all its events are yielded).
func Race[T any](ctx context.Context, streams map[string]Stream[T]) Stream[Tag[T]] {
	return func(yield func(Tag[T], error) bool) {
		if len(streams) == 0 {
			return
		}

		// resultCh receives the first event from any stream.
		resultCh := make(chan raceResult[T], 1)

		raceCtx, raceCancel := context.WithCancel(ctx)
		defer raceCancel()

		var wg sync.WaitGroup

		for name, s := range streams {
			wg.Add(1)
			go func(name string, s Stream[T]) {
				defer wg.Done()
				for v, err := range s {
					select {
					case <-raceCtx.Done():
						return
					default:
					}
					if err != nil {
						select {
						case resultCh <- raceResult[T]{err: err}:
							// Won the race with an error.
						default:
							// Another goroutine already won.
						}
						return
					}
					select {
					case resultCh <- raceResult[T]{tag: Tag[T]{Source: name, Event: v}}:
						// Won the race.
					default:
						// Another goroutine already won.
					}
					return
				}
			}(name, s)
		}

		// Close channel once all goroutines finish.
		go func() {
			wg.Wait()
			close(resultCh)
		}()

		// Yield the single winning result (if any).
		for result := range resultCh {
			raceCancel()
			if result.err != nil {
				yield(Tag[T]{}, result.err)
				return
			}
			yield(result.tag, nil)
			return
		}
	}
}
