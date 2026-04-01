package appinit

import (
	"context"
	"fmt"
	"time"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/lifecycle"
	"github.com/langoai/lango/internal/logging"
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

// ModuleTimingEntry records the duration of a module initialization.
type ModuleTimingEntry struct {
	Module   string        `json:"module"`
	Duration time.Duration `json:"duration"`
}

// BuildResult holds the aggregated output from all initialized modules.
type BuildResult struct {
	Tools          []*agent.Tool
	Components     []lifecycle.ComponentEntry
	CatalogEntries []CatalogEntry
	Resolver       Resolver
	ModuleTiming   []ModuleTimingEntry `json:"moduleTiming,omitempty"`
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
	var catalogEntries []CatalogEntry
	var moduleTiming []ModuleTimingEntry

	log := logging.SubsystemSugar("appinit")
	for _, m := range sorted {
		log.Infow("initializing module", "module", m.Name())
		start := time.Now()
		result, err := m.Init(ctx, resolver)
		elapsed := time.Since(start)
		moduleTiming = append(moduleTiming, ModuleTimingEntry{Module: m.Name(), Duration: elapsed})
		if err != nil {
			return nil, fmt.Errorf("init module %q: %w", m.Name(), err)
		}
		log.Infow("module initialized", "module", m.Name(), "duration_ms", elapsed.Milliseconds())
		if result == nil {
			continue
		}

		tools = append(tools, result.Tools...)
		components = append(components, result.Components...)
		catalogEntries = append(catalogEntries, result.CatalogEntries...)

		for key, val := range result.Values {
			resolver.set(key, val)
		}
	}

	return &BuildResult{
		Tools:          tools,
		Components:     components,
		CatalogEntries: catalogEntries,
		Resolver:       resolver,
		ModuleTiming:   moduleTiming,
	}, nil
}
