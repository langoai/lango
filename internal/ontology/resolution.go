package ontology

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/langoai/lango/internal/graph"
)

// EntityResolver manages entity identity — aliases, canonicalization, merge/split.
type EntityResolver interface {
	// Resolve returns the canonical ID for a raw ID. If no alias exists, returns rawID unchanged.
	Resolve(ctx context.Context, rawID string) (string, error)

	// RegisterAlias creates a mapping from rawID to canonicalID.
	RegisterAlias(ctx context.Context, rawID, canonicalID, source string) error

	// DeclareSameAs declares that nodeA and nodeB refer to the same entity.
	// The canonical is chosen by source precedence; ties go to the first-created alias.
	DeclareSameAs(ctx context.Context, nodeA, nodeB, source string) error

	// Merge moves all relationships from duplicate to canonical and retracts the originals.
	// Order: snapshot → replicate → retract → alias (last, to avoid mid-merge canonicalization).
	Merge(ctx context.Context, canonical, duplicate string) (*MergeResult, error)

	// Split removes the alias for splitOut from canonical. Relationship restoration is manual.
	Split(ctx context.Context, canonical, splitOut string) error

	// Aliases returns all raw IDs that map to canonicalID.
	Aliases(ctx context.Context, canonicalID string) ([]string, error)
}

// MergeResult reports the outcome of a Merge operation.
type MergeResult struct {
	TriplesUpdated int
	AliasesCreated int
}

// entityResolver implements EntityResolver.
type entityResolver struct {
	aliasStore *AliasStore
	store      graph.Store
	truth      TruthMaintainer
}

// NewEntityResolver creates an EntityResolver backed by the given stores.
func NewEntityResolver(as *AliasStore, store graph.Store, tm TruthMaintainer) EntityResolver {
	return &entityResolver{aliasStore: as, store: store, truth: tm}
}

func (r *entityResolver) Resolve(ctx context.Context, rawID string) (string, error) {
	return r.aliasStore.Resolve(ctx, rawID)
}

func (r *entityResolver) RegisterAlias(ctx context.Context, rawID, canonicalID, source string) error {
	return r.aliasStore.Register(ctx, rawID, canonicalID, source)
}

func (r *entityResolver) DeclareSameAs(ctx context.Context, nodeA, nodeB, source string) error {
	canonA, err := r.aliasStore.Resolve(ctx, nodeA)
	if err != nil {
		return fmt.Errorf("declare same as resolve %q: %w", nodeA, err)
	}
	canonB, err := r.aliasStore.Resolve(ctx, nodeB)
	if err != nil {
		return fmt.Errorf("declare same as resolve %q: %w", nodeB, err)
	}

	if canonA == canonB {
		return nil // already the same entity
	}

	// Convention: second argument (nodeB) is treated as canonical on tie.
	// This matches the intuitive call pattern: DeclareSameAs(alias, canonical).
	canonical, duplicate := canonB, canonA

	return r.aliasStore.Register(ctx, duplicate, canonical, source)
}

func (r *entityResolver) Merge(ctx context.Context, canonical, duplicate string) (*MergeResult, error) {
	result := &MergeResult{}

	// 1. Snapshot: capture all triples BEFORE alias registration.
	outgoing, err := r.store.QueryBySubject(ctx, duplicate)
	if err != nil {
		return nil, fmt.Errorf("merge snapshot outgoing: %w", err)
	}
	incoming, err := r.store.QueryByObject(ctx, duplicate)
	if err != nil {
		return nil, fmt.Errorf("merge snapshot incoming: %w", err)
	}

	// 2. Replicate: copy triples with canonical IDs.
	for _, t := range outgoing {
		newT := t
		newT.Subject = canonical
		if err := r.store.AddTriple(ctx, newT); err != nil {
			return nil, fmt.Errorf("merge replicate outgoing: %w", err)
		}
		result.TriplesUpdated++
	}
	for _, t := range incoming {
		newT := t
		newT.Object = canonical
		if err := r.store.AddTriple(ctx, newT); err != nil {
			return nil, fmt.Errorf("merge replicate incoming: %w", err)
		}
		result.TriplesUpdated++
	}

	// 3. Retract: invalidate original triples.
	var retractErrs int
	var firstErrTriple graph.Triple
	reason := fmt.Sprintf("merge: %s→%s", duplicate, canonical)
	for _, t := range outgoing {
		if err := r.truth.RetractFact(ctx, t.Subject, t.Predicate, t.Object, reason); err != nil {
			retractErrs++
			if retractErrs == 1 {
				firstErrTriple = t
			}
		}
	}
	for _, t := range incoming {
		if err := r.truth.RetractFact(ctx, t.Subject, t.Predicate, t.Object, reason); err != nil {
			retractErrs++
			if retractErrs == 1 {
				firstErrTriple = t
			}
		}
	}
	if retractErrs > 0 {
		slog.Warn("merge retraction partial failure",
			"canonical", canonical, "duplicate", duplicate,
			"failedCount", retractErrs,
			"firstSubject", firstErrTriple.Subject,
			"firstPredicate", firstErrTriple.Predicate,
		)
	}

	// 4. Alias: register LAST to avoid mid-merge canonicalization interference.
	if err := r.aliasStore.Register(ctx, duplicate, canonical, "merge"); err != nil {
		return nil, fmt.Errorf("merge alias: %w", err)
	}
	result.AliasesCreated = 1

	// 5. Transitive alias update: any existing aliases pointing to duplicate
	// must now point to canonical, otherwise Resolve(oldAlias) → duplicate (stale).
	existingAliases, err := r.aliasStore.ListByCanonical(ctx, duplicate)
	if err == nil {
		for _, rawID := range existingAliases {
			if rawID == duplicate {
				continue // skip the alias we just registered
			}
			if err := r.aliasStore.Register(ctx, rawID, canonical, "merge"); err == nil {
				result.AliasesCreated++
			}
		}
	}

	return result, nil
}

func (r *entityResolver) Split(ctx context.Context, _, splitOut string) error {
	return r.aliasStore.Remove(ctx, splitOut)
}

func (r *entityResolver) Aliases(ctx context.Context, canonicalID string) ([]string, error) {
	return r.aliasStore.ListByCanonical(ctx, canonicalID)
}
