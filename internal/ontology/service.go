package ontology

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/langoai/lango/internal/ctxkeys"
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

	// Property Store — Change 1.5-1
	SetEntityProperty(ctx context.Context, entityID, entityType, property, value string) error
	GetEntityProperties(ctx context.Context, entityID string) (map[string]string, error)
	QueryEntities(ctx context.Context, q PropertyQuery) ([]EntityResult, error)
	GetEntity(ctx context.Context, entityID string) (*EntityResult, error)
	DeleteEntityProperties(ctx context.Context, entityID string) error

	// Action Types — Change 2-2
	ExecuteAction(ctx context.Context, actionName string, params map[string]string) (*ActionResult, error)
	ListActions(ctx context.Context) ([]ActionSummary, error)
	GetActionLog(ctx context.Context, logID uuid.UUID) (*ActionLogEntry, error)
	ListActionLogs(ctx context.Context, actionName string, limit int) ([]ActionLogEntry, error)
}

// ServiceImpl implements OntologyService.
type ServiceImpl struct {
	registry         Registry
	graphStore       graph.Store
	truth            TruthMaintainer
	resolver         EntityResolver
	propStore        *PropertyStore
	acl              ACLPolicy
	executor         *ActionExecutor
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

// SetPropertyStore injects the PropertyStore after construction.
func (s *ServiceImpl) SetPropertyStore(ps *PropertyStore) {
	s.propStore = ps
}

// SetACLPolicy injects the ACL policy after construction.
// When nil, all operations are permitted (backward compatible).
func (s *ServiceImpl) SetACLPolicy(p ACLPolicy) {
	s.acl = p
}

// SetActionExecutor injects the ActionExecutor after construction.
func (s *ServiceImpl) SetActionExecutor(e *ActionExecutor) {
	s.executor = e
}

// checkPermission verifies the calling principal has the required permission.
// Returns nil when acl is nil (allow-all default).
func (s *ServiceImpl) checkPermission(ctx context.Context, perm Permission) error {
	if s.acl == nil {
		return nil
	}
	principal := ctxkeys.PrincipalFromContext(ctx)
	if principal == "" {
		principal = "system"
	}
	return s.acl.Check(principal, perm)
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
	if err := s.checkPermission(ctx, PermRead); err != nil {
		return nil, err
	}
	return s.registry.GetType(ctx, name)
}

func (s *ServiceImpl) ListTypes(ctx context.Context) ([]ObjectType, error) {
	if err := s.checkPermission(ctx, PermRead); err != nil {
		return nil, err
	}
	return s.registry.ListTypes(ctx)
}

func (s *ServiceImpl) GetPredicate(ctx context.Context, name string) (*PredicateDefinition, error) {
	if err := s.checkPermission(ctx, PermRead); err != nil {
		return nil, err
	}
	return s.registry.GetPredicate(ctx, name)
}

func (s *ServiceImpl) ListPredicates(ctx context.Context) ([]PredicateDefinition, error) {
	if err := s.checkPermission(ctx, PermRead); err != nil {
		return nil, err
	}
	return s.registry.ListPredicates(ctx)
}

func (s *ServiceImpl) RegisterType(ctx context.Context, t ObjectType) error {
	if err := s.checkPermission(ctx, PermWrite); err != nil {
		return err
	}
	if err := s.registry.RegisterType(ctx, t); err != nil {
		return err
	}
	s.version.Add(1)
	return nil
}

func (s *ServiceImpl) RegisterPredicate(ctx context.Context, p PredicateDefinition) error {
	if err := s.checkPermission(ctx, PermWrite); err != nil {
		return err
	}
	if err := s.registry.RegisterPredicate(ctx, p); err != nil {
		return err
	}
	s.refreshPredicateCache()
	s.version.Add(1)
	return nil
}

func (s *ServiceImpl) DeprecateType(ctx context.Context, name string) error {
	if err := s.checkPermission(ctx, PermAdmin); err != nil {
		return err
	}
	if err := s.registry.DeprecateType(ctx, name); err != nil {
		return err
	}
	s.version.Add(1)
	return nil
}

func (s *ServiceImpl) DeprecatePredicate(ctx context.Context, name string) error {
	if err := s.checkPermission(ctx, PermAdmin); err != nil {
		return err
	}
	if err := s.registry.DeprecatePredicate(ctx, name); err != nil {
		return err
	}
	s.refreshPredicateCache()
	s.version.Add(1)
	return nil
}

func (s *ServiceImpl) ValidateTriple(ctx context.Context, t graph.Triple) error {
	if err := s.checkPermission(ctx, PermRead); err != nil {
		return err
	}
	if !s.predicateIsActive(t.Predicate) {
		return fmt.Errorf("unknown or deprecated predicate %q", t.Predicate)
	}
	return nil
}

func (s *ServiceImpl) SchemaVersion(ctx context.Context) (int, error) {
	if err := s.checkPermission(ctx, PermRead); err != nil {
		return 0, err
	}
	return int(s.version.Load()), nil
}

func (s *ServiceImpl) StoreTriple(ctx context.Context, t graph.Triple) error {
	if err := s.checkPermission(ctx, PermWrite); err != nil {
		return err
	}
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
	if err := s.checkPermission(ctx, PermWrite); err != nil {
		return nil, err
	}
	if s.truth == nil {
		return nil, fmt.Errorf("truth maintenance not initialized")
	}
	return s.truth.AssertFact(ctx, input)
}

func (s *ServiceImpl) RetractFact(ctx context.Context, subject, predicate, object, reason string) error {
	if err := s.checkPermission(ctx, PermWrite); err != nil {
		return err
	}
	if s.truth == nil {
		return fmt.Errorf("truth maintenance not initialized")
	}
	return s.truth.RetractFact(ctx, subject, predicate, object, reason)
}

func (s *ServiceImpl) ConflictSet(ctx context.Context, subject, predicate string) ([]Conflict, error) {
	if err := s.checkPermission(ctx, PermRead); err != nil {
		return nil, err
	}
	if s.truth == nil {
		return nil, fmt.Errorf("truth maintenance not initialized")
	}
	return s.truth.ConflictSet(ctx, subject, predicate)
}

func (s *ServiceImpl) ResolveConflict(ctx context.Context, conflictID uuid.UUID, winnerObject, reason string) error {
	if err := s.checkPermission(ctx, PermAdmin); err != nil {
		return err
	}
	if s.truth == nil {
		return fmt.Errorf("truth maintenance not initialized")
	}
	return s.truth.ResolveConflict(ctx, conflictID, winnerObject, reason)
}

func (s *ServiceImpl) FactsAt(ctx context.Context, subject string, validAt time.Time) ([]graph.Triple, error) {
	if err := s.checkPermission(ctx, PermRead); err != nil {
		return nil, err
	}
	if s.truth == nil {
		return nil, fmt.Errorf("truth maintenance not initialized")
	}
	return s.truth.FactsAt(ctx, subject, validAt)
}

func (s *ServiceImpl) OpenConflicts(ctx context.Context) ([]Conflict, error) {
	if err := s.checkPermission(ctx, PermRead); err != nil {
		return nil, err
	}
	if s.truth == nil {
		return nil, fmt.Errorf("truth maintenance not initialized")
	}
	return s.truth.OpenConflicts(ctx)
}

// --- Entity Resolution delegation ---

func (s *ServiceImpl) Resolve(ctx context.Context, rawID string) (string, error) {
	if err := s.checkPermission(ctx, PermRead); err != nil {
		return "", err
	}
	if s.resolver == nil {
		return rawID, nil // no resolver = identity
	}
	return s.resolver.Resolve(ctx, rawID)
}

func (s *ServiceImpl) DeclareSameAs(ctx context.Context, nodeA, nodeB, source string) error {
	if err := s.checkPermission(ctx, PermWrite); err != nil {
		return err
	}
	if s.resolver == nil {
		return fmt.Errorf("entity resolver not initialized")
	}
	return s.resolver.DeclareSameAs(ctx, nodeA, nodeB, source)
}

func (s *ServiceImpl) MergeEntities(ctx context.Context, canonical, duplicate string) (*MergeResult, error) {
	if err := s.checkPermission(ctx, PermAdmin); err != nil {
		return nil, err
	}
	if s.resolver == nil {
		return nil, fmt.Errorf("entity resolver not initialized")
	}
	return s.resolver.Merge(ctx, canonical, duplicate)
}

func (s *ServiceImpl) SplitEntity(ctx context.Context, canonical, splitOut string) error {
	if err := s.checkPermission(ctx, PermAdmin); err != nil {
		return err
	}
	if s.resolver == nil {
		return fmt.Errorf("entity resolver not initialized")
	}
	return s.resolver.Split(ctx, canonical, splitOut)
}

func (s *ServiceImpl) Aliases(ctx context.Context, canonicalID string) ([]string, error) {
	if err := s.checkPermission(ctx, PermRead); err != nil {
		return nil, err
	}
	if s.resolver == nil {
		return nil, nil
	}
	return s.resolver.Aliases(ctx, canonicalID)
}

// QueryTriples resolves subject via alias then queries graph store.
func (s *ServiceImpl) QueryTriples(ctx context.Context, subject string) ([]graph.Triple, error) {
	if err := s.checkPermission(ctx, PermRead); err != nil {
		return nil, err
	}
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

// --- Property Store delegation ---

func (s *ServiceImpl) SetEntityProperty(ctx context.Context, entityID, entityType, property, value string) error {
	if err := s.checkPermission(ctx, PermWrite); err != nil {
		return err
	}
	if s.propStore == nil {
		return fmt.Errorf("property store not initialized")
	}

	// Validate entityType exists in registry.
	objType, err := s.registry.GetType(ctx, entityType)
	if err != nil {
		return fmt.Errorf("set entity property: unknown type %q: %w", entityType, err)
	}

	// Validate property name exists in ObjectType.Properties.
	var propDef *PropertyDef
	for i := range objType.Properties {
		if objType.Properties[i].Name == property {
			propDef = &objType.Properties[i]
			break
		}
	}
	if propDef == nil {
		return fmt.Errorf("set entity property: property %q not defined in type %q", property, entityType)
	}

	// Canonicalize entity_id.
	if s.resolver != nil {
		if canonical, err := s.resolver.Resolve(ctx, entityID); err == nil {
			entityID = canonical
		}
	}

	return s.propStore.SetProperty(ctx, entityID, entityType, property, value, string(propDef.Type))
}

func (s *ServiceImpl) GetEntityProperties(ctx context.Context, entityID string) (map[string]string, error) {
	if err := s.checkPermission(ctx, PermRead); err != nil {
		return nil, err
	}
	if s.propStore == nil {
		return nil, fmt.Errorf("property store not initialized")
	}
	// Canonicalize entity_id.
	if s.resolver != nil {
		if canonical, err := s.resolver.Resolve(ctx, entityID); err == nil {
			entityID = canonical
		}
	}
	return s.propStore.GetProperties(ctx, entityID)
}

func (s *ServiceImpl) QueryEntities(ctx context.Context, q PropertyQuery) ([]EntityResult, error) {
	if err := s.checkPermission(ctx, PermRead); err != nil {
		return nil, err
	}
	if s.propStore == nil {
		return nil, fmt.Errorf("property store not initialized")
	}

	entityIDs, err := s.propStore.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	results := make([]EntityResult, 0, len(entityIDs))
	for _, id := range entityIDs {
		props, err := s.propStore.GetProperties(ctx, id)
		if err != nil {
			return nil, err
		}
		er := EntityResult{
			EntityID:   id,
			EntityType: q.EntityType,
			Properties: props,
		}
		// QueryEntities: outgoing triples only (incoming is too expensive for list queries).
		if s.graphStore != nil {
			if outgoing, qErr := s.graphStore.QueryBySubject(ctx, id); qErr == nil {
				er.Outgoing = toResultTriples(outgoing)
			}
		}
		results = append(results, er)
	}
	return results, nil
}

func (s *ServiceImpl) GetEntity(ctx context.Context, entityID string) (*EntityResult, error) {
	if err := s.checkPermission(ctx, PermRead); err != nil {
		return nil, err
	}
	if s.propStore == nil {
		return nil, fmt.Errorf("property store not initialized")
	}

	// Canonicalize entity_id.
	if s.resolver != nil {
		if canonical, err := s.resolver.Resolve(ctx, entityID); err == nil {
			entityID = canonical
		}
	}

	props, err := s.propStore.GetProperties(ctx, entityID)
	if err != nil {
		return nil, err
	}

	// Determine entity type from property store.
	entityType, _ := s.propStore.GetEntityType(ctx, entityID)

	result := &EntityResult{
		EntityID:   entityID,
		EntityType: entityType,
		Properties: props,
	}

	// Fetch outgoing + incoming triples from graph store.
	if s.graphStore != nil {
		if outgoing, err := s.graphStore.QueryBySubject(ctx, entityID); err == nil {
			result.Outgoing = toResultTriples(outgoing)
		}
		if incoming, err := s.graphStore.QueryByObject(ctx, entityID); err == nil {
			result.Incoming = toResultTriples(incoming)
		}
	}

	return result, nil
}

func (s *ServiceImpl) DeleteEntityProperties(ctx context.Context, entityID string) error {
	if err := s.checkPermission(ctx, PermAdmin); err != nil {
		return err
	}
	if s.propStore == nil {
		return fmt.Errorf("property store not initialized")
	}
	if s.resolver != nil {
		if canonical, err := s.resolver.Resolve(ctx, entityID); err == nil {
			entityID = canonical
		}
	}
	return s.propStore.DeleteProperties(ctx, entityID)
}

// --- Action Types delegation ---

func (s *ServiceImpl) ExecuteAction(ctx context.Context, actionName string, params map[string]string) (*ActionResult, error) {
	if s.executor == nil {
		return nil, fmt.Errorf("action executor not initialized")
	}
	return s.executor.Execute(ctx, actionName, params)
}

func (s *ServiceImpl) ListActions(_ context.Context) ([]ActionSummary, error) {
	if s.executor == nil {
		return nil, nil
	}
	actions := s.executor.registry.List()
	summaries := make([]ActionSummary, len(actions))
	for i, a := range actions {
		summaries[i] = ActionSummary{
			Name:         a.Name,
			Description:  a.Description,
			RequiredPerm: a.RequiredPerm,
			ParamSchema:  a.ParamSchema,
		}
	}
	return summaries, nil
}

func (s *ServiceImpl) GetActionLog(ctx context.Context, logID uuid.UUID) (*ActionLogEntry, error) {
	if s.executor == nil {
		return nil, fmt.Errorf("action executor not initialized")
	}
	return s.executor.logStore.Get(ctx, logID)
}

func (s *ServiceImpl) ListActionLogs(ctx context.Context, actionName string, limit int) ([]ActionLogEntry, error) {
	if s.executor == nil {
		return nil, nil
	}
	return s.executor.logStore.List(ctx, actionName, limit)
}

// toResultTriples converts graph.Triple slice to ResultTriple slice.
func toResultTriples(triples []graph.Triple) []ResultTriple {
	if len(triples) == 0 {
		return nil
	}
	result := make([]ResultTriple, len(triples))
	for i, t := range triples {
		result[i] = ResultTriple{
			Subject:     t.Subject,
			Predicate:   t.Predicate,
			Object:      t.Object,
			SubjectType: t.SubjectType,
			ObjectType:  t.ObjectType,
		}
	}
	return result
}
