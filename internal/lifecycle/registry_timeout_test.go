package lifecycle

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type blockingComponent struct {
	name        string
	stopStarted chan struct{}
	release     chan struct{}
}

func (b *blockingComponent) Name() string { return b.name }

func (b *blockingComponent) Start(_ context.Context, _ *sync.WaitGroup) error { return nil }

func (b *blockingComponent) Stop(_ context.Context) error {
	close(b.stopStarted)
	<-b.release
	return nil
}

type signalComponent struct {
	name       string
	stopCalled chan struct{}
}

func (s *signalComponent) Name() string { return s.name }

func (s *signalComponent) Start(_ context.Context, _ *sync.WaitGroup) error { return nil }

func (s *signalComponent) Stop(_ context.Context) error {
	close(s.stopCalled)
	return nil
}

func TestRegistryStopAll_TimeoutDoesNotBlockRemainingStops(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	blocker := &blockingComponent{
		name:        "network",
		stopStarted: make(chan struct{}),
		release:     make(chan struct{}),
	}
	follower := &signalComponent{
		name:       "buffer",
		stopCalled: make(chan struct{}),
	}

	r.Register(follower, PriorityBuffer)
	r.Register(blocker, PriorityNetwork)

	var wg sync.WaitGroup
	require.NoError(t, r.StartAll(context.Background(), &wg))

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := r.StopAll(ctx)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))

	select {
	case <-blocker.stopStarted:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected blocking component stop to be attempted")
	}

	select {
	case <-follower.stopCalled:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected remaining component stop to be attempted after timeout")
	}

	close(blocker.release)
}
