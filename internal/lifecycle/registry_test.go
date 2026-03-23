package lifecycle

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type orderTracker struct {
	mu    sync.Mutex
	order []string
}

func (o *orderTracker) record(action string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.order = append(o.order, action)
}

type mockComponent struct {
	name     string
	tracker  *orderTracker
	startErr error
}

func (m *mockComponent) Name() string { return m.name }

func (m *mockComponent) Start(_ context.Context, _ *sync.WaitGroup) error {
	if m.startErr != nil {
		return m.startErr
	}
	m.tracker.record("start:" + m.name)
	return nil
}

func (m *mockComponent) Stop(_ context.Context) error {
	m.tracker.record("stop:" + m.name)
	return nil
}

func TestRegistry_StartOrder(t *testing.T) {
	t.Parallel()

	tracker := &orderTracker{}
	r := NewRegistry()

	r.Register(&mockComponent{name: "network", tracker: tracker}, PriorityNetwork)
	r.Register(&mockComponent{name: "buffer", tracker: tracker}, PriorityBuffer)
	r.Register(&mockComponent{name: "infra", tracker: tracker}, PriorityInfra)

	var wg sync.WaitGroup
	err := r.StartAll(context.Background(), &wg)
	require.NoError(t, err)

	assert.Equal(t, []string{"start:infra", "start:buffer", "start:network"}, tracker.order)
}

func TestRegistry_StopReverseOrder(t *testing.T) {
	t.Parallel()

	tracker := &orderTracker{}
	r := NewRegistry()

	r.Register(&mockComponent{name: "infra", tracker: tracker}, PriorityInfra)
	r.Register(&mockComponent{name: "buffer", tracker: tracker}, PriorityBuffer)
	r.Register(&mockComponent{name: "network", tracker: tracker}, PriorityNetwork)

	var wg sync.WaitGroup
	err := r.StartAll(context.Background(), &wg)
	require.NoError(t, err)

	tracker.order = nil // reset
	err = r.StopAll(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []string{"stop:network", "stop:buffer", "stop:infra"}, tracker.order)
}

func TestRegistry_RollbackOnFailure(t *testing.T) {
	t.Parallel()

	tracker := &orderTracker{}
	errBoom := errors.New("boom")
	r := NewRegistry()

	r.Register(&mockComponent{name: "a", tracker: tracker}, PriorityInfra)
	r.Register(&mockComponent{name: "b", tracker: tracker}, PriorityBuffer)
	r.Register(&mockComponent{name: "c", tracker: tracker, startErr: errBoom}, PriorityNetwork)

	var wg sync.WaitGroup
	err := r.StartAll(context.Background(), &wg)
	require.Error(t, err)
	assert.ErrorIs(t, err, errBoom)

	// a and b started, then c failed, so b and a should be stopped in reverse
	assert.Equal(t, []string{"start:a", "start:b", "stop:b", "stop:a"}, tracker.order)
}

func TestRegistry_EmptyRegistry(t *testing.T) {
	t.Parallel()

	r := NewRegistry()

	var wg sync.WaitGroup
	err := r.StartAll(context.Background(), &wg)
	require.NoError(t, err)

	err = r.StopAll(context.Background())
	require.NoError(t, err)
}

func TestRegistry_Names(t *testing.T) {
	t.Parallel()

	tracker := &orderTracker{}
	r := NewRegistry()

	r.Register(&mockComponent{name: "alpha", tracker: tracker}, PriorityInfra)
	r.Register(&mockComponent{name: "beta", tracker: tracker}, PriorityBuffer)
	r.Register(&mockComponent{name: "gamma", tracker: tracker}, PriorityNetwork)

	names := r.Names()
	assert.Equal(t, []string{"alpha", "beta", "gamma"}, names)
}

func TestRegistry_Names_Empty(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	names := r.Names()
	assert.Empty(t, names)
}

func TestRegistry_SetMaxPriority_SkipsHighPriority(t *testing.T) {
	t.Parallel()

	tracker := &orderTracker{}
	r := NewRegistry()

	r.Register(&mockComponent{name: "infra", tracker: tracker}, PriorityInfra)
	r.Register(&mockComponent{name: "core", tracker: tracker}, PriorityCore)
	r.Register(&mockComponent{name: "buffer", tracker: tracker}, PriorityBuffer)
	r.Register(&mockComponent{name: "network", tracker: tracker}, PriorityNetwork)
	r.Register(&mockComponent{name: "automation", tracker: tracker}, PriorityAutomation)

	r.SetMaxPriority(PriorityBuffer) // only start infra, core, buffer

	var wg sync.WaitGroup
	err := r.StartAll(context.Background(), &wg)
	require.NoError(t, err)

	assert.Equal(t, []string{"start:infra", "start:core", "start:buffer"}, tracker.order)

	// StopAll should only stop the components that were started.
	tracker.order = nil
	err = r.StopAll(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []string{"stop:buffer", "stop:core", "stop:infra"}, tracker.order)
}

func TestRegistry_SetMaxPriority_RollbackOnFailure(t *testing.T) {
	t.Parallel()

	tracker := &orderTracker{}
	errBoom := errors.New("boom")
	r := NewRegistry()

	r.Register(&mockComponent{name: "infra", tracker: tracker}, PriorityInfra)
	r.Register(&mockComponent{name: "buffer", tracker: tracker, startErr: errBoom}, PriorityBuffer)
	r.Register(&mockComponent{name: "network", tracker: tracker}, PriorityNetwork)

	r.SetMaxPriority(PriorityBuffer)

	var wg sync.WaitGroup
	err := r.StartAll(context.Background(), &wg)
	require.Error(t, err)
	assert.ErrorIs(t, err, errBoom)

	// infra started, buffer failed → rollback infra. network was never attempted.
	assert.Equal(t, []string{"start:infra", "stop:infra"}, tracker.order)
}

func TestRegistry_SamePriorityPreservesOrder(t *testing.T) {
	t.Parallel()

	tracker := &orderTracker{}
	r := NewRegistry()

	r.Register(&mockComponent{name: "first", tracker: tracker}, PriorityBuffer)
	r.Register(&mockComponent{name: "second", tracker: tracker}, PriorityBuffer)
	r.Register(&mockComponent{name: "third", tracker: tracker}, PriorityBuffer)

	var wg sync.WaitGroup
	err := r.StartAll(context.Background(), &wg)
	require.NoError(t, err)

	assert.Equal(t, []string{"start:first", "start:second", "start:third"}, tracker.order)
}
