package ontology

import (
	"context"
	"strings"
)

// SeedDefaults registers the existing 9 predicates and 6 node types
// into the ontology registry. Idempotent — skips entries that already exist.
func SeedDefaults(ctx context.Context, svc OntologyService) error {
	if err := seedPredicates(ctx, svc); err != nil {
		return err
	}
	return seedTypes(ctx, svc)
}

func seedPredicates(ctx context.Context, svc OntologyService) error {
	predicates := []PredicateDefinition{
		{Name: "related_to", Description: "semantic relationship between entities", Cardinality: ManyToMany, Status: SchemaActive, Version: 1},
		{Name: "caused_by", Description: "causal relationship (effect → cause)", SourceTypes: []string{"ErrorPattern"}, TargetTypes: []string{"Tool", "ErrorPattern"}, Cardinality: ManyToMany, Status: SchemaActive, Version: 1},
		{Name: "resolved_by", Description: "resolution relationship (error → fix)", SourceTypes: []string{"ErrorPattern"}, TargetTypes: []string{"Fix"}, Cardinality: ManyToMany, Status: SchemaActive, Version: 1},
		{Name: "follows", Description: "temporal ordering (observation → observation)", SourceTypes: []string{"Observation"}, TargetTypes: []string{"Observation"}, Cardinality: OneToMany, Status: SchemaActive, Version: 1},
		{Name: "similar_to", Description: "similarity relationship (learning ↔ learning)", SourceTypes: []string{"ErrorPattern"}, TargetTypes: []string{"ErrorPattern"}, Cardinality: ManyToMany, Status: SchemaActive, Version: 1},
		{Name: "contains", Description: "containment (collection → member)", Cardinality: OneToMany, Status: SchemaActive, Version: 1},
		{Name: "in_session", Description: "session membership", TargetTypes: []string{"Session"}, Cardinality: ManyToOne, Status: SchemaActive, Version: 1},
		{Name: "reflects_on", Description: "reflection targets (reflection → observation)", SourceTypes: []string{"Reflection"}, TargetTypes: []string{"Observation"}, Cardinality: ManyToMany, Status: SchemaActive, Version: 1},
		{Name: "learned_from", Description: "provenance (learning → session)", SourceTypes: []string{"Fix"}, TargetTypes: []string{"Session"}, Cardinality: ManyToOne, Status: SchemaActive, Version: 1},
	}

	for _, p := range predicates {
		err := svc.RegisterPredicate(ctx, p)
		if err != nil && isAlreadyExists(err) {
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func seedTypes(ctx context.Context, svc OntologyService) error {
	types := []ObjectType{
		{
			Name:        "ErrorPattern",
			Description: "Normalized error pattern from tool execution",
			Properties: []PropertyDef{
				{Name: "tool_name", Type: TypeString, Required: true},
				{Name: "pattern", Type: TypeString, Required: true},
			},
			Status: SchemaActive, Version: 1,
		},
		{
			Name:        "Tool",
			Description: "Agent tool executor",
			Properties: []PropertyDef{
				{Name: "name", Type: TypeString, Required: true},
			},
			Status: SchemaActive, Version: 1,
		},
		{
			Name:        "Fix",
			Description: "Error resolution applied by user or agent",
			Properties: []PropertyDef{
				{Name: "description", Type: TypeString, Required: true},
			},
			Status: SchemaActive, Version: 1,
		},
		{
			Name:        "Session",
			Description: "Conversation session",
			Properties: []PropertyDef{
				{Name: "key", Type: TypeString, Required: true},
			},
			Status: SchemaActive, Version: 1,
		},
		{
			Name:        "Observation",
			Description: "Compressed conversation observation",
			Properties: []PropertyDef{
				{Name: "session_key", Type: TypeString, Required: true},
			},
			Status: SchemaActive, Version: 1,
		},
		{
			Name:        "Reflection",
			Description: "Distilled insight from observations",
			Properties: []PropertyDef{
				{Name: "session_key", Type: TypeString, Required: true},
			},
			Status: SchemaActive, Version: 1,
		},
	}

	for _, t := range types {
		err := svc.RegisterType(ctx, t)
		if err != nil && isAlreadyExists(err) {
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func isAlreadyExists(err error) bool {
	return err != nil && strings.Contains(err.Error(), "already exists")
}
