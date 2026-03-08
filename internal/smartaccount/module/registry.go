package module

import (
	"fmt"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/common"

	sa "github.com/langoai/lango/internal/smartaccount"
)

// Registry manages available ERC-7579 module descriptors.
type Registry struct {
	mu      sync.RWMutex
	modules map[common.Address]*ModuleDescriptor
}

// NewRegistry creates a new module registry.
func NewRegistry() *Registry {
	return &Registry{
		modules: make(map[common.Address]*ModuleDescriptor),
	}
}

// Register adds a module descriptor to the registry.
func (r *Registry) Register(desc *ModuleDescriptor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.modules[desc.Address]; ok {
		return fmt.Errorf(
			"module %s: %w", desc.Address.Hex(),
			sa.ErrModuleAlreadyInstalled,
		)
	}
	cp := copyDescriptor(desc)
	r.modules[cp.Address] = cp
	return nil
}

// Get retrieves a module descriptor by address.
func (r *Registry) Get(addr common.Address) (*ModuleDescriptor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	desc, ok := r.modules[addr]
	if !ok {
		return nil, fmt.Errorf(
			"module %s: %w", addr.Hex(), sa.ErrModuleNotInstalled,
		)
	}
	return copyDescriptor(desc), nil
}

// List returns all registered module descriptors sorted by name.
func (r *Registry) List() []*ModuleDescriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*ModuleDescriptor, 0, len(r.modules))
	for _, desc := range r.modules {
		result = append(result, copyDescriptor(desc))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// ListByType returns all module descriptors matching the given type.
func (r *Registry) ListByType(t sa.ModuleType) []*ModuleDescriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ModuleDescriptor
	for _, desc := range r.modules {
		if desc.Type == t {
			result = append(result, copyDescriptor(desc))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// copyDescriptor returns a deep copy of a ModuleDescriptor.
func copyDescriptor(src *ModuleDescriptor) *ModuleDescriptor {
	cp := *src
	if src.InitData != nil {
		cp.InitData = make([]byte, len(src.InitData))
		copy(cp.InitData, src.InitData)
	}
	return &cp
}
