package main

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type fakeServeApp struct {
	stopFn func(ctx context.Context) error
}

func (f *fakeServeApp) Start(ctx context.Context) error { return nil }

func (f *fakeServeApp) Stop(ctx context.Context) error {
	if f.stopFn != nil {
		return f.stopFn(ctx)
	}
	return nil
}

func TestWatchServeSignals_FirstSignalStartsGracefulShutdown(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopped := make(chan struct{})
	app := &fakeServeApp{
		stopFn: func(ctx context.Context) error {
			close(stopped)
			return nil
		},
	}

	sigChan := make(chan os.Signal, 2)
	forced := make(chan int, 1)

	go watchServeSignals(ctx, app, zap.NewNop().Sugar(), sigChan, time.Second, cancel, func(code int) {
		forced <- code
	})

	sigChan <- os.Interrupt

	select {
	case <-stopped:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected graceful shutdown to start")
	}

	select {
	case code := <-forced:
		t.Fatalf("unexpected forced exit with code %d", code)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestWatchServeSignals_SecondSignalForcesExit(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	release := make(chan struct{})
	app := &fakeServeApp{
		stopFn: func(ctx context.Context) error {
			<-release
			return nil
		},
	}

	sigChan := make(chan os.Signal, 2)
	forced := make(chan int, 1)
	var once sync.Once

	go watchServeSignals(ctx, app, zap.NewNop().Sugar(), sigChan, time.Second, cancel, func(code int) {
		once.Do(func() { forced <- code })
	})

	sigChan <- os.Interrupt
	sigChan <- os.Interrupt

	select {
	case code := <-forced:
		assert.Equal(t, 130, code)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected forced exit on second signal")
	}

	close(release)
}
