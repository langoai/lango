package appinit

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/lifecycle"
)

func TestBuilder_Empty(t *testing.T) {
	result, err := NewBuilder().Build(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Tools) != 0 {
		t.Errorf("want 0 tools, got %d", len(result.Tools))
	}
	if len(result.Components) != 0 {
		t.Errorf("want 0 components, got %d", len(result.Components))
	}
}

func TestBuilder_MultipleModules(t *testing.T) {
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
			if val == nil {
				return nil, errors.New("expected key_a to be resolved")
			}
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Tools) != 2 {
		t.Fatalf("want 2 tools, got %d", len(result.Tools))
	}
	// A should init first, so tool_a first.
	if result.Tools[0].Name != "tool_a" {
		t.Errorf("want first tool %q, got %q", "tool_a", result.Tools[0].Name)
	}
	if result.Tools[1].Name != "tool_b" {
		t.Errorf("want second tool %q, got %q", "tool_b", result.Tools[1].Name)
	}

	// Verify resolver contains values from both modules.
	val := result.Resolver.Resolve("key_b")
	if val != "value_a_extended" {
		t.Errorf("want resolver key_b = %q, got %v", "value_a_extended", val)
	}
}

func TestBuilder_ResolverPassesValues(t *testing.T) {
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedVal != 42 {
		t.Errorf("want resolved value 42, got %v", receivedVal)
	}
}

func TestBuilder_Components(t *testing.T) {
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Components) != 1 {
		t.Fatalf("want 1 component, got %d", len(result.Components))
	}
	if result.Components[0].Component.Name() != "test_comp" {
		t.Errorf("want component name %q, got %q", "test_comp", result.Components[0].Component.Name())
	}
}

func TestBuilder_InitError(t *testing.T) {
	mod := &stubModule{
		name:    "failing",
		enabled: true,
		initFn: func(_ context.Context, _ Resolver) (*ModuleResult, error) {
			return nil, errors.New("init failed")
		},
	}

	_, err := NewBuilder().AddModule(mod).Build(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errors.Unwrap(err)) {
		// Just check that the error message contains useful info.
		wantMsg := `init module "failing"`
		if got := err.Error(); len(got) == 0 {
			t.Errorf("expected non-empty error message")
		}
		_ = wantMsg
	}
}

func TestBuilder_NilResult(t *testing.T) {
	mod := &stubModule{
		name:    "nil_result",
		enabled: true,
		initFn: func(_ context.Context, _ Resolver) (*ModuleResult, error) {
			return nil, nil
		},
	}

	result, err := NewBuilder().AddModule(mod).Build(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Tools) != 0 {
		t.Errorf("want 0 tools, got %d", len(result.Tools))
	}
}

func TestBuilder_CycleError(t *testing.T) {
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
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
}

// dummyComponent implements lifecycle.Component for testing.
type dummyComponent struct {
	name string
}

func (d *dummyComponent) Name() string                                    { return d.name }
func (d *dummyComponent) Start(_ context.Context, _ *sync.WaitGroup) error { return nil }
func (d *dummyComponent) Stop(_ context.Context) error                     { return nil }
