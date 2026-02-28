package appinit

import (
	"context"
	"testing"
)

// stubModule is a minimal Module implementation for testing.
type stubModule struct {
	name      string
	provides  []Provides
	dependsOn []Provides
	enabled   bool
	initFn    func(ctx context.Context, r Resolver) (*ModuleResult, error)
}

func (s *stubModule) Name() string            { return s.name }
func (s *stubModule) Provides() []Provides     { return s.provides }
func (s *stubModule) DependsOn() []Provides    { return s.dependsOn }
func (s *stubModule) Enabled() bool            { return s.enabled }
func (s *stubModule) Init(ctx context.Context, r Resolver) (*ModuleResult, error) {
	if s.initFn != nil {
		return s.initFn(ctx, r)
	}
	return &ModuleResult{}, nil
}

func TestTopoSort(t *testing.T) {
	tests := []struct {
		give      string
		modules   []Module
		wantOrder []string
		wantErr   bool
	}{
		{
			give:      "empty input",
			modules:   nil,
			wantOrder: nil,
		},
		{
			give: "single module no deps",
			modules: []Module{
				&stubModule{name: "a", provides: []Provides{"key_a"}, enabled: true},
			},
			wantOrder: []string{"a"},
		},
		{
			give: "two modules with dependency",
			modules: []Module{
				&stubModule{name: "b", provides: []Provides{"key_b"}, dependsOn: []Provides{"key_a"}, enabled: true},
				&stubModule{name: "a", provides: []Provides{"key_a"}, enabled: true},
			},
			wantOrder: []string{"a", "b"},
		},
		{
			give: "chain A -> B -> C",
			modules: []Module{
				&stubModule{name: "c", provides: []Provides{"key_c"}, dependsOn: []Provides{"key_b"}, enabled: true},
				&stubModule{name: "a", provides: []Provides{"key_a"}, enabled: true},
				&stubModule{name: "b", provides: []Provides{"key_b"}, dependsOn: []Provides{"key_a"}, enabled: true},
			},
			wantOrder: []string{"a", "b", "c"},
		},
		{
			give: "diamond dependency",
			modules: []Module{
				&stubModule{name: "d", provides: []Provides{"key_d"}, dependsOn: []Provides{"key_b", "key_c"}, enabled: true},
				&stubModule{name: "b", provides: []Provides{"key_b"}, dependsOn: []Provides{"key_a"}, enabled: true},
				&stubModule{name: "c", provides: []Provides{"key_c"}, dependsOn: []Provides{"key_a"}, enabled: true},
				&stubModule{name: "a", provides: []Provides{"key_a"}, enabled: true},
			},
			wantOrder: []string{"a", "b", "c", "d"},
		},
		{
			give: "disabled module skipped",
			modules: []Module{
				&stubModule{name: "a", provides: []Provides{"key_a"}, enabled: true},
				&stubModule{name: "b", provides: []Provides{"key_b"}, dependsOn: []Provides{"key_a"}, enabled: false},
				&stubModule{name: "c", provides: []Provides{"key_c"}, dependsOn: []Provides{"key_b"}, enabled: true},
			},
			// c depends on key_b but b is disabled, so the dep is ignored; a and c both have no real deps.
			wantOrder: []string{"a", "c"},
		},
		{
			give: "dependency on missing key ignored",
			modules: []Module{
				&stubModule{name: "a", provides: []Provides{"key_a"}, dependsOn: []Provides{"nonexistent"}, enabled: true},
			},
			wantOrder: []string{"a"},
		},
		{
			give: "cycle detection",
			modules: []Module{
				&stubModule{name: "a", provides: []Provides{"key_a"}, dependsOn: []Provides{"key_b"}, enabled: true},
				&stubModule{name: "b", provides: []Provides{"key_b"}, dependsOn: []Provides{"key_a"}, enabled: true},
			},
			wantErr: true,
		},
		{
			give: "three-way cycle detection",
			modules: []Module{
				&stubModule{name: "a", provides: []Provides{"key_a"}, dependsOn: []Provides{"key_c"}, enabled: true},
				&stubModule{name: "b", provides: []Provides{"key_b"}, dependsOn: []Provides{"key_a"}, enabled: true},
				&stubModule{name: "c", provides: []Provides{"key_c"}, dependsOn: []Provides{"key_b"}, enabled: true},
			},
			wantErr: true,
		},
		{
			give: "self-dependency ignored",
			modules: []Module{
				&stubModule{name: "a", provides: []Provides{"key_a"}, dependsOn: []Provides{"key_a"}, enabled: true},
			},
			wantOrder: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got, err := TopoSort(tt.modules)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.wantOrder) {
				names := moduleNames(got)
				t.Fatalf("want %d modules %v, got %d modules %v",
					len(tt.wantOrder), tt.wantOrder, len(got), names)
			}

			for i, m := range got {
				if m.Name() != tt.wantOrder[i] {
					t.Errorf("position %d: want %q, got %q", i, tt.wantOrder[i], m.Name())
				}
			}
		})
	}
}

func moduleNames(modules []Module) []string {
	names := make([]string, len(modules))
	for i, m := range modules {
		names[i] = m.Name()
	}
	return names
}
