package ontology

import "context"

// Registry is the internal interface for ontology schema persistence.
// Only used by ServiceImpl — external consumers use OntologyService.
type Registry interface {
	// ObjectType operations
	RegisterType(ctx context.Context, t ObjectType) error
	GetType(ctx context.Context, name string) (*ObjectType, error)
	ListTypes(ctx context.Context) ([]ObjectType, error)
	DeprecateType(ctx context.Context, name string) error

	// PredicateDefinition operations
	RegisterPredicate(ctx context.Context, p PredicateDefinition) error
	GetPredicate(ctx context.Context, name string) (*PredicateDefinition, error)
	ListPredicates(ctx context.Context) ([]PredicateDefinition, error)
	DeprecatePredicate(ctx context.Context, name string) error
}
