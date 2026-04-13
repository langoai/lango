package ontology

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/ontologypredicate"
	"github.com/langoai/lango/internal/ent/ontologytype"
)

// EntRegistry implements Registry backed by Ent ORM (SQLite).
type EntRegistry struct {
	client *ent.Client
}

// NewEntRegistry creates a new Ent-backed registry.
func NewEntRegistry(client *ent.Client) *EntRegistry {
	return &EntRegistry{client: client}
}

func (r *EntRegistry) RegisterType(ctx context.Context, t ObjectType) error {
	exists, err := r.client.OntologyType.Query().
		Where(ontologytype.Name(t.Name)).
		Exist(ctx)
	if err != nil {
		return fmt.Errorf("check type existence: %w", err)
	}
	if exists {
		return fmt.Errorf("object type %q already exists", t.Name)
	}

	propsJSON, err := marshalProperties(t.Properties)
	if err != nil {
		return fmt.Errorf("marshal properties: %w", err)
	}

	builder := r.client.OntologyType.Create().
		SetName(t.Name).
		SetDescription(t.Description).
		SetProperties(propsJSON).
		SetStatus(ontologytype.Status(t.Status)).
		SetVersion(t.Version)

	if t.Extends != "" {
		builder.SetExtends(t.Extends)
	}

	_, err = builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("create type: %w", err)
	}
	return nil
}

func (r *EntRegistry) GetType(ctx context.Context, name string) (*ObjectType, error) {
	e, err := r.client.OntologyType.Query().
		Where(ontologytype.Name(name)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get type %q: %w", name, err)
	}
	return entToObjectType(e), nil
}

func (r *EntRegistry) ListTypes(ctx context.Context) ([]ObjectType, error) {
	entries, err := r.client.OntologyType.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list types: %w", err)
	}
	result := make([]ObjectType, 0, len(entries))
	for _, e := range entries {
		result = append(result, *entToObjectType(e))
	}
	return result, nil
}

func (r *EntRegistry) DeprecateType(ctx context.Context, name string) error {
	n, err := r.client.OntologyType.Update().
		Where(ontologytype.Name(name)).
		SetStatus(ontologytype.StatusDeprecated).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("deprecate type %q: %w", name, err)
	}
	if n == 0 {
		return fmt.Errorf("type %q not found", name)
	}
	return nil
}

func (r *EntRegistry) RegisterPredicate(ctx context.Context, p PredicateDefinition) error {
	exists, err := r.client.OntologyPredicate.Query().
		Where(ontologypredicate.Name(p.Name)).
		Exist(ctx)
	if err != nil {
		return fmt.Errorf("check predicate existence: %w", err)
	}
	if exists {
		return fmt.Errorf("predicate %q already exists", p.Name)
	}

	builder := r.client.OntologyPredicate.Create().
		SetName(p.Name).
		SetDescription(p.Description).
		SetSourceTypes(p.SourceTypes).
		SetTargetTypes(p.TargetTypes).
		SetCardinality(ontologypredicate.Cardinality(p.Cardinality)).
		SetStatus(ontologypredicate.Status(p.Status)).
		SetVersion(p.Version)

	if p.Inverse != "" {
		builder.SetInverse(p.Inverse)
	}

	_, err = builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("create predicate: %w", err)
	}
	return nil
}

func (r *EntRegistry) GetPredicate(ctx context.Context, name string) (*PredicateDefinition, error) {
	e, err := r.client.OntologyPredicate.Query().
		Where(ontologypredicate.Name(name)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get predicate %q: %w", name, err)
	}
	return entToPredicateDefinition(e), nil
}

func (r *EntRegistry) ListPredicates(ctx context.Context) ([]PredicateDefinition, error) {
	entries, err := r.client.OntologyPredicate.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list predicates: %w", err)
	}
	result := make([]PredicateDefinition, 0, len(entries))
	for _, e := range entries {
		result = append(result, *entToPredicateDefinition(e))
	}
	return result, nil
}

func (r *EntRegistry) DeprecatePredicate(ctx context.Context, name string) error {
	n, err := r.client.OntologyPredicate.Update().
		Where(ontologypredicate.Name(name)).
		SetStatus(ontologypredicate.StatusDeprecated).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("deprecate predicate %q: %w", name, err)
	}
	if n == 0 {
		return fmt.Errorf("predicate %q not found", name)
	}
	return nil
}

func (r *EntRegistry) UpdateTypeStatus(ctx context.Context, name string, status SchemaStatus) error {
	n, err := r.client.OntologyType.Update().
		Where(ontologytype.Name(name)).
		SetStatus(ontologytype.Status(status)).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("update type status: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("type %q not found", name)
	}
	return nil
}

func (r *EntRegistry) UpdatePredicateStatus(ctx context.Context, name string, status SchemaStatus) error {
	n, err := r.client.OntologyPredicate.Update().
		Where(ontologypredicate.Name(name)).
		SetStatus(ontologypredicate.Status(status)).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("update predicate status: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("predicate %q not found", name)
	}
	return nil
}

// --- conversion helpers ---

func entToObjectType(e *ent.OntologyType) *ObjectType {
	var props []PropertyDef
	if e.Properties != nil {
		props = unmarshalProperties(e.Properties)
	}
	var ext string
	if e.Extends != nil {
		ext = *e.Extends
	}
	return &ObjectType{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		Properties:  props,
		Extends:     ext,
		Status:      SchemaStatus(e.Status),
		Version:     e.Version,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func entToPredicateDefinition(e *ent.OntologyPredicate) *PredicateDefinition {
	var inv string
	if e.Inverse != nil {
		inv = *e.Inverse
	}
	return &PredicateDefinition{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		SourceTypes: e.SourceTypes,
		TargetTypes: e.TargetTypes,
		Cardinality: Cardinality(e.Cardinality),
		Inverse:     inv,
		Status:      SchemaStatus(e.Status),
		Version:     e.Version,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func marshalProperties(props []PropertyDef) ([]map[string]interface{}, error) {
	if len(props) == 0 {
		return nil, nil
	}
	data, err := json.Marshal(props)
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func unmarshalProperties(raw []map[string]interface{}) []PropertyDef {
	if len(raw) == 0 {
		return nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var props []PropertyDef
	if err := json.Unmarshal(data, &props); err != nil {
		return nil
	}
	return props
}
