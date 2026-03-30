package ontology

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/entityalias"
)

// AliasStore provides Ent-backed CRUD for entity aliases.
type AliasStore struct {
	client *ent.Client
}

// NewAliasStore creates an AliasStore backed by the given Ent client.
func NewAliasStore(client *ent.Client) *AliasStore {
	return &AliasStore{client: client}
}

// Resolve returns the canonical ID for a raw ID.
// If no alias exists, returns rawID unchanged.
func (s *AliasStore) Resolve(ctx context.Context, rawID string) (string, error) {
	e, err := s.client.EntityAlias.Query().
		Where(entityalias.RawIDEQ(rawID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return rawID, nil // no alias = identity
		}
		return "", fmt.Errorf("alias resolve %q: %w", rawID, err)
	}
	return e.CanonicalID, nil
}

// Register creates or updates an alias mapping from rawID to canonicalID.
func (s *AliasStore) Register(ctx context.Context, rawID, canonicalID, source string) error {
	// Check if alias already exists.
	existing, err := s.client.EntityAlias.Query().
		Where(entityalias.RawIDEQ(rawID)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return fmt.Errorf("alias register check: %w", err)
	}

	if existing != nil {
		// Update existing alias.
		_, err = existing.Update().
			SetCanonicalID(canonicalID).
			SetSource(source).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("alias register update: %w", err)
		}
		return nil
	}

	// Create new alias.
	_, err = s.client.EntityAlias.Create().
		SetRawID(rawID).
		SetCanonicalID(canonicalID).
		SetSource(source).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("alias register create: %w", err)
	}
	return nil
}

// ListByCanonical returns all raw IDs that map to the given canonical ID.
func (s *AliasStore) ListByCanonical(ctx context.Context, canonicalID string) ([]string, error) {
	entries, err := s.client.EntityAlias.Query().
		Where(entityalias.CanonicalIDEQ(canonicalID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("alias list by canonical: %w", err)
	}
	result := make([]string, len(entries))
	for i, e := range entries {
		result[i] = e.RawID
	}
	return result, nil
}

// Remove deletes an alias for a raw ID (used by Split).
func (s *AliasStore) Remove(ctx context.Context, rawID string) error {
	_, err := s.client.EntityAlias.Delete().
		Where(entityalias.RawIDEQ(rawID)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("alias remove %q: %w", rawID, err)
	}
	return nil
}
