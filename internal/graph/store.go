package graph

import "context"

// Predicate represents a relationship type in the knowledge graph.
type Predicate string

// Predicate constants define relationship types in the knowledge graph.
// These are untyped string constants to allow direct use in Triple struct literals
// where the Predicate field is string. Use Predicate(x).Valid() for validation.
const (
	RelatedTo   = "related_to"   // semantic relationship between entities
	CausedBy    = "caused_by"    // causal relationship (effect → cause)
	ResolvedBy  = "resolved_by"  // resolution relationship (error → fix)
	Follows     = "follows"      // temporal ordering (observation → observation)
	SimilarTo   = "similar_to"   // similarity relationship (learning ↔ learning)
	Contains    = "contains"     // containment (session → observation)
	InSession   = "in_session"   // session membership
	ReflectsOn  = "reflects_on"  // reflection targets (reflection → observation)
	LearnedFrom = "learned_from" // provenance (learning → session)
)

// Valid reports whether p is a known predicate type.
func (p Predicate) Valid() bool {
	switch string(p) {
	case RelatedTo, CausedBy, ResolvedBy, Follows, SimilarTo, Contains, InSession, ReflectsOn, LearnedFrom:
		return true
	}
	return false
}

// Values returns all known predicate types.
func (p Predicate) Values() []Predicate {
	return []Predicate{RelatedTo, CausedBy, ResolvedBy, Follows, SimilarTo, Contains, InSession, ReflectsOn, LearnedFrom}
}

// Triple represents a Subject-Predicate-Object relationship in the graph.
type Triple struct {
	Subject   string
	Predicate string
	Object    string
	Metadata  map[string]string
}

// Store provides graph CRUD and traversal operations.
type Store interface {
	// AddTriple adds a single triple to the graph.
	AddTriple(ctx context.Context, t Triple) error

	// AddTriples adds multiple triples atomically.
	AddTriples(ctx context.Context, triples []Triple) error

	// RemoveTriple removes a triple from the graph.
	RemoveTriple(ctx context.Context, t Triple) error

	// QueryBySubject returns all triples with the given subject.
	QueryBySubject(ctx context.Context, subject string) ([]Triple, error)

	// QueryByObject returns all triples with the given object.
	QueryByObject(ctx context.Context, object string) ([]Triple, error)

	// QueryBySubjectPredicate returns triples matching subject and predicate.
	QueryBySubjectPredicate(ctx context.Context, subject, predicate string) ([]Triple, error)

	// Traverse performs a breadth-first traversal from a start node.
	// predicates filters which edge types to follow (empty = all).
	Traverse(ctx context.Context, startNode string, maxDepth int, predicates []string) ([]Triple, error)

	// Count returns the total number of triples in the store.
	Count(ctx context.Context) (int, error)

	// PredicateStats returns the number of triples for each predicate type.
	PredicateStats(ctx context.Context) (map[string]int, error)

	// ClearAll removes all triples from the store.
	ClearAll(ctx context.Context) error

	// Close closes the underlying store.
	Close() error
}
