package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWave43DefaultServeAwaitShutdownWaitsForContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		serveAwaitShutdownFn(ctx)
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("serve await shutdown did not return after context cancellation")
	}
}

func TestWave43WatchServeSignalsLogsStopErrorAndStillCancels(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopErr := errors.New("wave43 stop refused")
	app := &fakeServeApp{
		stopFn: func(context.Context) error {
			return stopErr
		},
	}
	sigChan := make(chan os.Signal, 1)
	done := make(chan struct{})
	forceExit := make(chan int, 1)
	go func() {
		watchServeSignals(ctx, app, zap.NewNop().Sugar(), sigChan, time.Second, cancel, func(int) {
			forceExit <- 1
		})
		close(done)
	}()

	sigChan <- os.Interrupt

	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("shutdown error path did not cancel serve context")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("signal watcher did not return after cancellation")
	}
	select {
	case <-forceExit:
		t.Fatal("single interrupt must not force exit")
	default:
	}
}

func TestWave43RootChatCommandRoutesModeThroughRunnerSeam(t *testing.T) {
	origInteractive := isInteractiveFn
	origRunChat := runChatFn
	t.Cleanup(func() {
		isInteractiveFn = origInteractive
		runChatFn = origRunChat
	})

	isInteractiveFn = func() bool { return true }
	var gotMode string
	runChatFn = func(mode string) error {
		gotMode = mode
		return errors.New("wave43 stop before TUI")
	}

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--mode", "research", "chat"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()

	require.Error(t, err)
	assert.Equal(t, "research", gotMode)
	assert.Contains(t, err.Error(), "wave43 stop before TUI")
}

func TestWave43ConfigKeysRunsThroughComposedConfigCommand(t *testing.T) {
	cmd := configCmd()
	var out bytes.Buffer
	cmd.SetArgs([]string{"keys", "agent"})
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "agent.provider")
	assert.NotContains(t, out.String(), "server.host")
}
