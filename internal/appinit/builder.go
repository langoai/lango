package appinit

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/lifecycle"
)

// Builder collects modules and orchestrates their initialization.
type Builder struct {
	modules []Module
}

// NewBuilder creates an empty Builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// AddModule appends a module and returns the builder for chaining.
func (b *Builder) AddModule(m Module) *Builder {
	b.modules = append(b.modules, m)
	return b
}

// BuildResult holds the aggregated output from all initialized modules.
type BuildResult struct {
	Tools      []*agent.Tool
	Components []lifecycle.ComponentEntry
	Resolver   Resolver
}

// Build sorts modules by dependency order, initializes each in sequence,
// and returns the aggregated tools, components, and resolver.
func (b *Builder) Build(ctx context.Context) (*BuildResult, error) {
	sorted, err := TopoSort(b.modules)
	if err != nil {
		return nil, fmt.Errorf("appinit build: %w", err)
	}

	resolver := newMapResolver()
	var tools []*agent.Tool
	var components []lifecycle.ComponentEntry

	for _, m := range sorted {
		result, err := m.Init(ctx, resolver)
		if err != nil {
			return nil, fmt.Errorf("init module %q: %w", m.Name(), err)
		}
		if result == nil {
			continue
		}

		tools = append(tools, result.Tools...)
		components = append(components, result.Components...)

		for key, val := range result.Values {
			resolver.set(key, val)
		}
	}

	return &BuildResult{
		Tools:      tools,
		Components: components,
		Resolver:   resolver,
	}, nil
}
