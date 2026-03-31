package ontology

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

// --- Full ↔ Slim Conversions ---

// TypeToSlim converts a full ObjectType to a slim wire type.
func TypeToSlim(t ObjectType) SchemaTypeSlim {
	props := make([]SchemaPropertySlim, len(t.Properties))
	for i, p := range t.Properties {
		props[i] = SchemaPropertySlim{
			Name:     p.Name,
			Type:     string(p.Type),
			Required: p.Required,
		}
	}
	return SchemaTypeSlim{
		Name:        t.Name,
		Description: t.Description,
		Properties:  props,
		Extends:     t.Extends,
	}
}

// PredicateToSlim converts a full PredicateDefinition to a slim wire type.
func PredicateToSlim(p PredicateDefinition) SchemaPredicateSlim {
	return SchemaPredicateSlim{
		Name:        p.Name,
		Description: p.Description,
		SourceTypes: p.SourceTypes,
		TargetTypes: p.TargetTypes,
		Cardinality: string(p.Cardinality),
		Inverse:     p.Inverse,
	}
}

// SlimToType converts a slim wire type to a full ObjectType with generated local fields.
func SlimToType(s SchemaTypeSlim, status SchemaStatus) ObjectType {
	props := make([]PropertyDef, len(s.Properties))
	for i, p := range s.Properties {
		props[i] = PropertyDef{
			Name:     p.Name,
			Type:     PropertyType(p.Type),
			Required: p.Required,
		}
	}
	now := time.Now()
	return ObjectType{
		ID:          uuid.New(),
		Name:        s.Name,
		Description: s.Description,
		Properties:  props,
		Extends:     s.Extends,
		Status:      status,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// SlimToPredicate converts a slim wire type to a full PredicateDefinition with generated local fields.
func SlimToPredicate(s SchemaPredicateSlim, status SchemaStatus) PredicateDefinition {
	now := time.Now()
	return PredicateDefinition{
		ID:          uuid.New(),
		Name:        s.Name,
		Description: s.Description,
		SourceTypes: s.SourceTypes,
		TargetTypes: s.TargetTypes,
		Cardinality: Cardinality(s.Cardinality),
		Inverse:     s.Inverse,
		Status:      status,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// --- Digest ---

// digestPayload is the canonical structure for digest computation.
type digestPayload struct {
	Types      []SchemaTypeSlim      `json:"types"`
	Predicates []SchemaPredicateSlim `json:"predicates"`
}

// ComputeDigest produces a SHA256 hex string from the canonical JSON of types and predicates.
// Types and predicates are sorted by name for order independence.
func ComputeDigest(types []SchemaTypeSlim, predicates []SchemaPredicateSlim) string {
	sorted := digestPayload{
		Types:      make([]SchemaTypeSlim, len(types)),
		Predicates: make([]SchemaPredicateSlim, len(predicates)),
	}
	copy(sorted.Types, types)
	copy(sorted.Predicates, predicates)

	slices.SortFunc(sorted.Types, func(a, b SchemaTypeSlim) int {
		if a.Name < b.Name {
			return -1
		}
		if a.Name > b.Name {
			return 1
		}
		return 0
	})
	slices.SortFunc(sorted.Predicates, func(a, b SchemaPredicateSlim) int {
		if a.Name < b.Name {
			return -1
		}
		if a.Name > b.Name {
			return 1
		}
		return 0
	})

	data, _ := json.Marshal(sorted)
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}

// --- Export/Import Implementation ---

// exportSchema creates a SchemaBundle from the current ontology state.
func exportSchema(ctx context.Context, registry Registry, schemaVersion int, exportedBy string) (*SchemaBundle, error) {
	types, err := registry.ListTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("list types: %w", err)
	}
	preds, err := registry.ListPredicates(ctx)
	if err != nil {
		return nil, fmt.Errorf("list predicates: %w", err)
	}

	var slimTypes []SchemaTypeSlim
	for _, t := range types {
		if t.Status == SchemaActive || t.Status == SchemaShadow {
			slimTypes = append(slimTypes, TypeToSlim(t))
		}
	}
	var slimPreds []SchemaPredicateSlim
	for _, p := range preds {
		if p.Status == SchemaActive || p.Status == SchemaShadow {
			slimPreds = append(slimPreds, PredicateToSlim(p))
		}
	}

	digest := ComputeDigest(slimTypes, slimPreds)

	return &SchemaBundle{
		Version:       1,
		SchemaVersion: schemaVersion,
		ExportedAt:    time.Now(),
		ExportedBy:    exportedBy,
		Types:         slimTypes,
		Predicates:    slimPreds,
		Digest:        digest,
	}, nil
}

// importSchema imports a SchemaBundle into the local ontology.
func importSchema(ctx context.Context, registry Registry, bundle *SchemaBundle, opts ImportOptions, governanceEnabled bool) (*ImportResult, error) {
	status := SchemaShadow
	switch opts.Mode {
	case ImportGoverned:
		status = SchemaProposed
	case ImportDryRun:
		// no mutations
	case ImportShadow:
		// default
	}

	result := &ImportResult{}

	// Import types
	for _, slim := range bundle.Types {
		existing, err := registry.GetType(ctx, slim.Name)
		if err == nil && existing != nil {
			if slimTypesEqual(TypeToSlim(*existing), slim) {
				result.TypesSkipped++
			} else {
				result.TypesConflicting = append(result.TypesConflicting, slim.Name)
			}
			continue
		}
		if opts.Mode == ImportDryRun {
			result.TypesAdded++
			continue
		}
		full := SlimToType(slim, status)
		if err := registry.RegisterType(ctx, full); err != nil {
			return nil, fmt.Errorf("register type %q: %w", slim.Name, err)
		}
		result.TypesAdded++
	}

	// Import predicates
	for _, slim := range bundle.Predicates {
		existing, err := registry.GetPredicate(ctx, slim.Name)
		if err == nil && existing != nil {
			if slimPredsEqual(PredicateToSlim(*existing), slim) {
				result.PredsSkipped++
			} else {
				result.PredsConflicting = append(result.PredsConflicting, slim.Name)
			}
			continue
		}
		if opts.Mode == ImportDryRun {
			result.PredsAdded++
			continue
		}
		full := SlimToPredicate(slim, status)
		if err := registry.RegisterPredicate(ctx, full); err != nil {
			return nil, fmt.Errorf("register predicate %q: %w", slim.Name, err)
		}
		result.PredsAdded++
	}

	return result, nil
}

// slimTypesEqual compares two slim types for semantic equality.
func slimTypesEqual(a, b SchemaTypeSlim) bool {
	if a.Name != b.Name || a.Description != b.Description || a.Extends != b.Extends {
		return false
	}
	if len(a.Properties) != len(b.Properties) {
		return false
	}
	for i := range a.Properties {
		if a.Properties[i] != b.Properties[i] {
			return false
		}
	}
	return true
}

// slimPredsEqual compares two slim predicates for semantic equality.
func slimPredsEqual(a, b SchemaPredicateSlim) bool {
	if a.Name != b.Name || a.Description != b.Description || a.Cardinality != b.Cardinality || a.Inverse != b.Inverse {
		return false
	}
	if !slices.Equal(a.SourceTypes, b.SourceTypes) || !slices.Equal(a.TargetTypes, b.TargetTypes) {
		return false
	}
	return true
}
