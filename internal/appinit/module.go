package appinit

import (
	"context"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/lifecycle"
)

// Provides identifies what a module provides to other modules.
type Provides string

// Well-known module provides keys.
const (
	ProvidesSessionStore Provides = "session_store"
	ProvidesSecurity     Provides = "security"
	ProvidesKnowledge    Provides = "knowledge"
	ProvidesMemory       Provides = "memory"
	ProvidesEmbedding    Provides = "embedding"
	ProvidesGraph        Provides = "graph"
	ProvidesPayment      Provides = "payment"
	ProvidesP2P          Provides = "p2p"
	ProvidesLibrarian    Provides = "librarian"
	ProvidesAutomation   Provides = "automation"
	ProvidesGateway      Provides = "gateway"
	ProvidesAgent        Provides = "agent"
)

// Resolver provides access to initialized module results.
type Resolver interface {
	// Resolve returns the value registered by a module for the given key.
	// Returns nil if the key hasn't been provided yet.
	Resolve(key Provides) interface{}
}

// mapResolver is the default Resolver backed by a map.
type mapResolver struct {
	values map[Provides]interface{}
}

func newMapResolver() *mapResolver {
	return &mapResolver{values: make(map[Provides]interface{})}
}

func (r *mapResolver) Resolve(key Provides) interface{} {
	return r.values[key]
}

func (r *mapResolver) set(key Provides, val interface{}) {
	r.values[key] = val
}

// ModuleResult is what Init returns.
type ModuleResult struct {
	// Tools are agent tools contributed by this module.
	Tools []*agent.Tool
	// Components are lifecycle components that need Start/Stop management.
	Components []lifecycle.ComponentEntry
	// Values are named values this module provides to other modules via Resolver.
	Values map[Provides]interface{}
}

// Module represents an initialization unit.
type Module interface {
	Name() string
	Provides() []Provides
	DependsOn() []Provides
	Enabled() bool
	Init(ctx context.Context, resolver Resolver) (*ModuleResult, error)
}
