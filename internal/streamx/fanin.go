package streamx

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

// FanIn collects ALL events from N named streams, returning them grouped by
// source name. It waits for all streams to complete. The first error from any
// stream causes all others to be cancelled and the error is returned.
func FanIn[T any](ctx context.Context, streams map[string]Stream[T]) (map[string][]T, error) {
	if len(streams) == 0 {
		return make(map[string][]T), nil
	}

	result := make(map[string][]T, len(streams))
	var mu sync.Mutex

	g, gctx := errgroup.WithContext(ctx)

	for name, s := range streams {
		g.Go(func() error {
			var local []T
			for v, err := range s {
				select {
				case <-gctx.Done():
					return gctx.Err()
				default:
				}
				if err != nil {
					return err
				}
				local = append(local, v)
			}
			mu.Lock()
			result[name] = local
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}
