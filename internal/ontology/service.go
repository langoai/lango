package ontology

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
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

	// Triple storage facade — Resolve → store.AddTriple.
	// When EntityResolver is available, canonicalizes subject and object
	// before storing. Predicate validation is NOT performed here — use
	// AssertFact for validated storage with temporal metadata and conflict detection.
	StoreTriple(ctx context.Context, t graph.Triple) error

	// PredicateValidator returns a context-free closure for hot-path
	// predicate validation. Uses a cached map, refreshed on schema changes.
	PredicateValidator() func(name string) bool

	// Truth Maintenance — Change 1-3
	AssertFact(ctx context.Context, input AssertionInput) (*AssertionResult, error)
	RetractFact(ctx context.Context, subject, predicate, object, reason string) error
	ConflictSet(ctx context.Context, subject, predicate string) ([]Conflict, error)
	ResolveConflict(ctx context.Context, conflictID uuid.UUID, winnerObject, reason string) error
	FactsAt(ctx context.Context, subject string, validAt time.Time) ([]graph.Triple, error)
	OpenConflicts(ctx context.Context) ([]Conflict, error)

	// Entity Resolution — Change 1-4
	Resolve(ctx context.Context, rawID string) (string, error)
	DeclareSameAs(ctx context.Context, nodeA, nodeB, source string) error
	MergeEntities(ctx context.Context, canonical, duplicate string) (*MergeResult, error)
	SplitEntity(ctx context.Context, canonical, splitOut string) error
	Aliases(ctx context.Context, canonicalID string) ([]string, error)
	// QueryTriples resolves subject via alias before querying graph store.
	QueryTriples(ctx context.Context, subject string) ([]graph.Triple, error)
}

// ServiceImpl implements OntologyService.
type ServiceImpl struct {
	registry         Registry
	graphStore       graph.Store
	truth            TruthMaintainer
	resolver         EntityResolver
	cacheMu          sync.RWMutex
	activePredicates map[string]bool
	version          atomic.Int64
}

// SetTruthMaintainer injects the TruthMaintainer after construction.
func (s *ServiceImpl) SetTruthMaintainer(tm TruthMaintainer) {
	s.truth = tm
}

// SetEntityResolver injects the EntityResolver after construction.
func (s *ServiceImpl) SetEntityResolver(er EntityResolver) {
	s.resolver = er
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
	// Entity Resolution: canonicalize subject and object if resolver is available.
	if s.resolver != nil {
		if canonical, err := s.resolver.Resolve(ctx, t.Subject); err == nil {
			t.Subject = canonical
		}
		if canonical, err := s.resolver.Resolve(ctx, t.Object); err == nil {
			t.Object = canonical
		}
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

// --- Truth Maintenance delegation ---

func (s *ServiceImpl) AssertFact(ctx context.Context, input AssertionInput) (*AssertionResult, error) {
	if s.truth == nil {
		return nil, fmt.Errorf("truth maintenance not initialized")
	}
	return s.truth.AssertFact(ctx, input)
}

func (s *ServiceImpl) RetractFact(ctx context.Context, subject, predicate, object, reason string) error {
	if s.truth == nil {
		return fmt.Errorf("truth maintenance not initialized")
	}
	return s.truth.RetractFact(ctx, subject, predicate, object, reason)
}

func (s *ServiceImpl) ConflictSet(ctx context.Context, subject, predicate string) ([]Conflict, error) {
	if s.truth == nil {
		return nil, fmt.Errorf("truth maintenance not initialized")
	}
	return s.truth.ConflictSet(ctx, subject, predicate)
}

func (s *ServiceImpl) ResolveConflict(ctx context.Context, conflictID uuid.UUID, winnerObject, reason string) error {
	if s.truth == nil {
		return fmt.Errorf("truth maintenance not initialized")
	}
	return s.truth.ResolveConflict(ctx, conflictID, winnerObject, reason)
}

func (s *ServiceImpl) FactsAt(ctx context.Context, subject string, validAt time.Time) ([]graph.Triple, error) {
	if s.truth == nil {
		return nil, fmt.Errorf("truth maintenance not initialized")
	}
	return s.truth.FactsAt(ctx, subject, validAt)
}

func (s *ServiceImpl) OpenConflicts(ctx context.Context) ([]Conflict, error) {
	if s.truth == nil {
		return nil, fmt.Errorf("truth maintenance not initialized")
	}
	return s.truth.OpenConflicts(ctx)
}

// --- Entity Resolution delegation ---

func (s *ServiceImpl) Resolve(ctx context.Context, rawID string) (string, error) {
	if s.resolver == nil {
		return rawID, nil // no resolver = identity
	}
	return s.resolver.Resolve(ctx, rawID)
}

func (s *ServiceImpl) DeclareSameAs(ctx context.Context, nodeA, nodeB, source string) error {
	if s.resolver == nil {
		return fmt.Errorf("entity resolver not initialized")
	}
	return s.resolver.DeclareSameAs(ctx, nodeA, nodeB, source)
}

func (s *ServiceImpl) MergeEntities(ctx context.Context, canonical, duplicate string) (*MergeResult, error) {
	if s.resolver == nil {
		return nil, fmt.Errorf("entity resolver not initialized")
	}
	return s.resolver.Merge(ctx, canonical, duplicate)
}

func (s *ServiceImpl) SplitEntity(ctx context.Context, canonical, splitOut string) error {
	if s.resolver == nil {
		return fmt.Errorf("entity resolver not initialized")
	}
	return s.resolver.Split(ctx, canonical, splitOut)
}

func (s *ServiceImpl) Aliases(ctx context.Context, canonicalID string) ([]string, error) {
	if s.resolver == nil {
		return nil, nil
	}
	return s.resolver.Aliases(ctx, canonicalID)
}

// QueryTriples resolves subject via alias then queries graph store.
func (s *ServiceImpl) QueryTriples(ctx context.Context, subject string) ([]graph.Triple, error) {
	if s.graphStore == nil {
		return nil, fmt.Errorf("graph store not available")
	}
	if s.resolver != nil {
		if canonical, err := s.resolver.Resolve(ctx, subject); err == nil {
			subject = canonical
		}
	}
	return s.graphStore.QueryBySubject(ctx, subject)
}
