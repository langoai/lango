package agentregistry

import (
	"fmt"
	"sort"
	"sync"

	"github.com/langoai/lango/internal/orchestration"
)

// Registry manages agent definitions from multiple sources.
type Registry struct {
	mu     sync.RWMutex
	agents map[string]*AgentDefinition
	order  []string // insertion order for deterministic iteration
}

// New creates a new empty Registry.
func New() *Registry {
	return &Registry{
		agents: make(map[string]*AgentDefinition),
	}
}

// Register adds or overwrites an agent definition by name.
func (r *Registry) Register(def *AgentDefinition) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[def.Name]; !exists {
		r.order = append(r.order, def.Name)
	}
	r.agents[def.Name] = def
}

// RegisterAll registers multiple agent definitions.
func (r *Registry) RegisterAll(defs []*AgentDefinition) {
	for _, d := range defs {
		r.Register(d)
	}
}

// Get returns the agent definition with the given name.
func (r *Registry) Get(name string) (*AgentDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	def, ok := r.agents[name]
	return def, ok
}

// Active returns all agent definitions with status "active", sorted by name.
func (r *Registry) Active() []*AgentDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*AgentDefinition, 0, len(r.agents))
	for _, def := range r.agents {
		if def.Status == StatusActive {
			result = append(result, def)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// All returns all agent definitions in insertion order.
func (r *Registry) All() []*AgentDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*AgentDefinition, 0, len(r.order))
	for _, name := range r.order {
		result = append(result, r.agents[name])
	}
	return result
}

// Specs converts all active agent definitions to orchestration.AgentSpec format.
func (r *Registry) Specs() []orchestration.AgentSpec {
	active := r.Active()
	specs := make([]orchestration.AgentSpec, 0, len(active))
	for _, def := range active {
		specs = append(specs, orchestration.AgentSpec{
			Name:             def.Name,
			Description:      def.Description,
			Instruction:      def.Instruction,
			Prefixes:         def.Prefixes,
			Keywords:         def.Keywords,
			Capabilities:     def.Capabilities,
			Accepts:          def.Accepts,
			Returns:          def.Returns,
			CannotDo:         def.CannotDo,
			AlwaysInclude:    def.AlwaysInclude,
			SessionIsolation: def.SessionIsolation,
		})
	}
	return specs
}

// LoadFromStore loads agent definitions from a Store and registers them all.
func (r *Registry) LoadFromStore(store Store) error {
	defs, err := store.Load()
	if err != nil {
		return fmt.Errorf("load from store: %w", err)
	}
	r.RegisterAll(defs)
	return nil
}
