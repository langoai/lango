package lifecycle

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStartable struct {
	started bool
	stopped bool
}

func (m *mockStartable) Start(_ *sync.WaitGroup) { m.started = true }
func (m *mockStartable) Stop()                    { m.stopped = true }

func TestNewSimpleComponent(t *testing.T) {
	m := &mockStartable{}
	c := NewSimpleComponent("test-simple", m)

	assert.Equal(t, "test-simple", c.Name())

	var wg sync.WaitGroup
	err := c.Start(context.Background(), &wg)
	require.NoError(t, err)
	assert.True(t, m.started)

	err = c.Stop(context.Background())
	require.NoError(t, err)
	assert.True(t, m.stopped)
}

func TestSimpleComponent_Struct(t *testing.T) {
	started := false
	stopped := false
	c := &SimpleComponent{
		ComponentName: "test-struct",
		StartFunc:     func(_ *sync.WaitGroup) { started = true },
		StopFunc:      func() { stopped = true },
	}

	assert.Equal(t, "test-struct", c.Name())

	var wg sync.WaitGroup
	err := c.Start(context.Background(), &wg)
	require.NoError(t, err)
	assert.True(t, started)

	err = c.Stop(context.Background())
	require.NoError(t, err)
	assert.True(t, stopped)
}

func TestFuncComponent(t *testing.T) {
	started := false
	stopped := false
	c := &FuncComponent{
		ComponentName: "test-func",
		StartFunc: func(_ context.Context, _ *sync.WaitGroup) error {
			started = true
			return nil
		},
		StopFunc: func(_ context.Context) error {
			stopped = true
			return nil
		},
	}

	assert.Equal(t, "test-func", c.Name())

	var wg sync.WaitGroup
	err := c.Start(context.Background(), &wg)
	require.NoError(t, err)
	assert.True(t, started)

	err = c.Stop(context.Background())
	require.NoError(t, err)
	assert.True(t, stopped)
}

func TestFuncComponent_NilStop(t *testing.T) {
	c := &FuncComponent{
		ComponentName: "test-nil-stop",
		StartFunc:     func(_ context.Context, _ *sync.WaitGroup) error { return nil },
	}

	err := c.Stop(context.Background())
	require.NoError(t, err)
}

func TestErrorComponent(t *testing.T) {
	errBoom := errors.New("boom")
	c := &ErrorComponent{
		ComponentName: "test-error",
		StartFunc:     func(_ context.Context) error { return errBoom },
		StopFunc:      func() {},
	}

	assert.Equal(t, "test-error", c.Name())

	var wg sync.WaitGroup
	err := c.Start(context.Background(), &wg)
	assert.ErrorIs(t, err, errBoom)
}
