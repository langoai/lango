package ontology

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/langoai/lango/internal/graph"
)

// OntologyService is the single facade for all ontology operations.
// All consumers use this interface — never reference Registry or other
// internal components directly.
type OntologyService interface {
	// Schema queries
	GetType(ctx context.Context, name string) (*ObjectType, error)
	ListTypes(ctx context.Context) ([]ObjectType, error)
	GetPredicate(ctx context.Context, name string) (*PredicateDefinition, error)
	ListPredicates(ctx context.Context) ([]PredicateDefinition, error)

	// Schema mutations
	RegisterType(ctx context.Context, t ObjectType) error
	RegisterPredicate(ctx context.Context, p PredicateDefinition) error
	DeprecateType(ctx context.Context, name string) error
	DeprecatePredicate(ctx context.Context, name string) error

	// Validation
	ValidateTriple(ctx context.Context, t graph.Triple) error

	// Schema version — increments on register/deprecate
	SchemaVersion(ctx context.Context) (int, error)

	// Triple storage facade — Resolve → Validate → store.AddTriple
	// Change 1-1: delegates directly to graph.Store.AddTriple.
	// Change 1-4: adds Resolve and Validate pipeline.
	StoreTriple(ctx context.Context, t graph.Triple) error

	// PredicateValidator returns a context-free closure for hot-path
	// predicate validation. Uses a cached map, refreshed on schema changes.
	PredicateValidator() func(name string) bool
}

// ServiceImpl implements OntologyService.
type ServiceImpl struct {
	registry         Registry
	graphStore       graph.Store
	cacheMu          sync.RWMutex
	activePredicates map[string]bool
	version          atomic.Int64
}

// NewService creates a new OntologyService backed by the given registry.
// graphStore may be nil if not yet available (StoreTriple will return an error).
func NewService(reg Registry, graphStore graph.Store) *ServiceImpl {
	return &ServiceImpl{
		registry:         reg,
		graphStore:       graphStore,
		activePredicates: make(map[string]bool),
	}
}

func (s *ServiceImpl) GetType(ctx context.Context, name string) (*ObjectType, error) {
	return s.registry.GetType(ctx, name)
}

func (s *ServiceImpl) ListTypes(ctx context.Context) ([]ObjectType, error) {
	return s.registry.ListTypes(ctx)
}

func (s *ServiceImpl) GetPredicate(ctx context.Context, name string) (*PredicateDefinition, error) {
	return s.registry.GetPredicate(ctx, name)
}

func (s *ServiceImpl) ListPredicates(ctx context.Context) ([]PredicateDefinition, error) {
	return s.registry.ListPredicates(ctx)
}

func (s *ServiceImpl) RegisterType(ctx context.Context, t ObjectType) error {
	if err := s.registry.RegisterType(ctx, t); err != nil {
		return err
	}
	s.version.Add(1)
	return nil
}

func (s *ServiceImpl) RegisterPredicate(ctx context.Context, p PredicateDefinition) error {
	if err := s.registry.RegisterPredicate(ctx, p); err != nil {
		return err
	}
	s.refreshPredicateCache()
	s.version.Add(1)
	return nil
}

func (s *ServiceImpl) DeprecateType(ctx context.Context, name string) error {
	if err := s.registry.DeprecateType(ctx, name); err != nil {
		return err
	}
	s.version.Add(1)
	return nil
}

func (s *ServiceImpl) DeprecatePredicate(ctx context.Context, name string) error {
	if err := s.registry.DeprecatePredicate(ctx, name); err != nil {
		return err
	}
	s.refreshPredicateCache()
	s.version.Add(1)
	return nil
}

func (s *ServiceImpl) ValidateTriple(_ context.Context, t graph.Triple) error {
	if !s.predicateIsActive(t.Predicate) {
		return fmt.Errorf("unknown or deprecated predicate %q", t.Predicate)
	}
	return nil
}

func (s *ServiceImpl) SchemaVersion(_ context.Context) (int, error) {
	return int(s.version.Load()), nil
}

func (s *ServiceImpl) StoreTriple(ctx context.Context, t graph.Triple) error {
	if s.graphStore == nil {
		return fmt.Errorf("graph store not available")
	}
	return s.graphStore.AddTriple(ctx, t)
}

// PredicateValidator returns a context-free closure for hot-path use.
// The closure checks the cached active predicate set without DB queries.
func (s *ServiceImpl) PredicateValidator() func(name string) bool {
	s.refreshPredicateCache()
	return s.predicateIsActive
}

func (s *ServiceImpl) predicateIsActive(name string) bool {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	return s.activePredicates[name]
}

// refreshPredicateCache reloads the active predicate set from registry.
func (s *ServiceImpl) refreshPredicateCache() {
	preds, err := s.registry.ListPredicates(context.Background())
	if err != nil {
		return
	}
	cache := make(map[string]bool, len(preds))
	for _, p := range preds {
		if p.Status == SchemaActive {
			cache[p.Name] = true
		}
	}
	s.cacheMu.Lock()
	s.activePredicates = cache
	s.cacheMu.Unlock()
}
