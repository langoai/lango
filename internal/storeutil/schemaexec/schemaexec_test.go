package schemaexec

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunExclusiveSerializesCallsAndPropagatesErrors(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("schema failed")
	require.ErrorIs(t, RunExclusive(func() error {
		return wantErr
	}), wantErr)

	const workers = 16
	var (
		mu       sync.Mutex
		active   int
		maxSeen  int
		started  sync.WaitGroup
		released sync.WaitGroup
	)
	started.Add(workers)
	released.Add(workers)
	errs := make(chan error, workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer released.Done()
			errs <- RunExclusive(func() error {
				started.Done()
				mu.Lock()
				active++
				if active > maxSeen {
					maxSeen = active
				}
				active--
				mu.Unlock()
				return nil
			})
		}()
	}

	started.Wait()
	released.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	require.Equal(t, 1, maxSeen)
}
