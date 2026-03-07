package appinit

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/lifecycle"
)

func TestBuilder_Empty(t *testing.T) {
	t.Parallel()

	result, err := NewBuilder().Build(context.Background())
	require.NoError(t, err)
	assert.Empty(t, result.Tools)
	assert.Empty(t, result.Components)
}

func TestBuilder_MultipleModules(t *testing.T) {
	t.Parallel()

	toolA := &agent.Tool{Name: "tool_a", Description: "Tool A"}
	toolB := &agent.Tool{Name: "tool_b", Description: "Tool B"}

	modA := &stubModule{
		name:     "a",
		provides: []Provides{"key_a"},
		enabled:  true,
		initFn: func(_ context.Context, _ Resolver) (*ModuleResult, error) {
			return &ModuleResult{
				Tools:  []*agent.Tool{toolA},
				Values: map[Provides]interface{}{"key_a": "value_a"},
			}, nil
		},
	}

	modB := &stubModule{
		name:      "b",
		provides:  []Provides{"key_b"},
		dependsOn: []Provides{"key_a"},
		enabled:   true,
		initFn: func(_ context.Context, r Resolver) (*ModuleResult, error) {
			// Verify we can resolve the dependency from module A.
			val := r.Resolve("key_a")
			require.NotNil(t, val)
			return &ModuleResult{
				Tools:  []*agent.Tool{toolB},
				Values: map[Provides]interface{}{"key_b": val.(string) + "_extended"},
			}, nil
		},
	}

	result, err := NewBuilder().
		AddModule(modB). // added out of order intentionally
		AddModule(modA).
		Build(context.Background())
	require.NoError(t, err)

	require.Len(t, result.Tools, 2)
	// A should init first, so tool_a first.
	assert.Equal(t, "tool_a", result.Tools[0].Name)
	assert.Equal(t, "tool_b", result.Tools[1].Name)

	// Verify resolver contains values from both modules.
	val := result.Resolver.Resolve("key_b")
	assert.Equal(t, "value_a_extended", val)
}

func TestBuilder_ResolverPassesValues(t *testing.T) {
	t.Parallel()

	var receivedVal interface{}

	modA := &stubModule{
		name:     "provider",
		provides: []Provides{ProvidesMemory},
		enabled:  true,
		initFn: func(_ context.Context, _ Resolver) (*ModuleResult, error) {
			return &ModuleResult{
				Values: map[Provides]interface{}{ProvidesMemory: 42},
			}, nil
		},
	}

	modB := &stubModule{
		name:      "consumer",
		dependsOn: []Provides{ProvidesMemory},
		enabled:   true,
		initFn: func(_ context.Context, r Resolver) (*ModuleResult, error) {
			receivedVal = r.Resolve(ProvidesMemory)
			return &ModuleResult{}, nil
		},
	}

	_, err := NewBuilder().
		AddModule(modB).
		AddModule(modA).
		Build(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 42, receivedVal)
}

func TestBuilder_Components(t *testing.T) {
	t.Parallel()

	comp := &dummyComponent{name: "test_comp"}
	mod := &stubModule{
		name:    "comp_module",
		enabled: true,
		initFn: func(_ context.Context, _ Resolver) (*ModuleResult, error) {
			return &ModuleResult{
				Components: []lifecycle.ComponentEntry{
					{Component: comp, Priority: lifecycle.PriorityCore},
				},
			}, nil
		},
	}

	result, err := NewBuilder().AddModule(mod).Build(context.Background())
	require.NoError(t, err)
	require.Len(t, result.Components, 1)
	assert.Equal(t, "test_comp", result.Components[0].Component.Name())
}

func TestBuilder_InitError(t *testing.T) {
	t.Parallel()

	mod := &stubModule{
		name:    "failing",
		enabled: true,
		initFn: func(_ context.Context, _ Resolver) (*ModuleResult, error) {
			return nil, errors.New("init failed")
		},
	}

	_, err := NewBuilder().AddModule(mod).Build(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failing")
}

func TestBuilder_NilResult(t *testing.T) {
	t.Parallel()

	mod := &stubModule{
		name:    "nil_result",
		enabled: true,
		initFn: func(_ context.Context, _ Resolver) (*ModuleResult, error) {
			return nil, nil
		},
	}

	result, err := NewBuilder().AddModule(mod).Build(context.Background())
	require.NoError(t, err)
	assert.Empty(t, result.Tools)
}

func TestBuilder_CycleError(t *testing.T) {
	t.Parallel()

	modA := &stubModule{
		name:      "a",
		provides:  []Provides{"key_a"},
		dependsOn: []Provides{"key_b"},
		enabled:   true,
	}
	modB := &stubModule{
		name:      "b",
		provides:  []Provides{"key_b"},
		dependsOn: []Provides{"key_a"},
		enabled:   true,
	}

	_, err := NewBuilder().AddModule(modA).AddModule(modB).Build(context.Background())
	require.Error(t, err)
}

// dummyComponent implements lifecycle.Component for testing.
type dummyComponent struct {
	name string
}

func (d *dummyComponent) Name() string                                    { return d.name }
func (d *dummyComponent) Start(_ context.Context, _ *sync.WaitGroup) error { return nil }
func (d *dummyComponent) Stop(_ context.Context) error                     { return nil }
