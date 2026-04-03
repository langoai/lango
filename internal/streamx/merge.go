package streamx

import (
	"context"
	"sync"
)

type mergeItem[T any] struct {
	tag Tag[T]
	err error
}

// Merge takes N named streams and yields tagged events as they arrive from any
// stream. One goroutine is launched per stream; all send to a shared channel.
// Context cancellation stops all goroutines cleanly.
func Merge[T any](ctx context.Context, streams map[string]Stream[T]) Stream[Tag[T]] {
	return func(yield func(Tag[T], error) bool) {
		if len(streams) == 0 {
			return
		}

		ch := make(chan mergeItem[T], len(streams))
		var wg sync.WaitGroup

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		for name, s := range streams {
			wg.Add(1)
			go func(name string, s Stream[T]) {
				defer wg.Done()
				for v, err := range s {
					select {
					case <-ctx.Done():
						return
					default:
					}
					if err != nil {
						select {
						case ch <- mergeItem[T]{err: err}:
						case <-ctx.Done():
						}
						return
					}
					select {
					case ch <- mergeItem[T]{tag: Tag[T]{Source: name, Event: v}}:
					case <-ctx.Done():
						return
					}
				}
			}(name, s)
		}

		// Close channel once all goroutines finish.
		go func() {
			wg.Wait()
			close(ch)
		}()

		for item := range ch {
			if item.err != nil {
				if !yield(Tag[T]{}, item.err) {
					cancel()
					return
				}
				continue
			}
			if !yield(item.tag, nil) {
				cancel()
				return
			}
		}
	}
}
