package ontology

import (
	"context"
	"fmt"
	"sort"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/entityproperty"
)

// PropertyStore provides Ent-backed EAV storage for per-entity property values.
type PropertyStore struct {
	client *ent.Client
}

// NewPropertyStore creates a PropertyStore backed by the given Ent client.
func NewPropertyStore(client *ent.Client) *PropertyStore {
	return &PropertyStore{client: client}
}

// SetProperty stores or updates a property value for an entity (upsert).
func (s *PropertyStore) SetProperty(ctx context.Context, entityID, entityType, property, value, valueType string) error {
	existing, err := s.client.EntityProperty.Query().
		Where(
			entityproperty.EntityIDEQ(entityID),
			entityproperty.PropertyEQ(property),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return fmt.Errorf("property set check: %w", err)
	}

	if existing != nil {
		_, err = existing.Update().
			SetValue(value).
			SetValueType(valueType).
			SetEntityType(entityType).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("property set update: %w", err)
		}
		return nil
	}

	_, err = s.client.EntityProperty.Create().
		SetEntityID(entityID).
		SetEntityType(entityType).
		SetProperty(property).
		SetValue(value).
		SetValueType(valueType).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("property set create: %w", err)
	}
	return nil
}

// GetProperties returns all properties for an entity as a key-value map.
func (s *PropertyStore) GetProperties(ctx context.Context, entityID string) (map[string]string, error) {
	entries, err := s.client.EntityProperty.Query().
		Where(entityproperty.EntityIDEQ(entityID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("property get: %w", err)
	}

	result := make(map[string]string, len(entries))
	for _, e := range entries {
		result[e.Property] = e.Value
	}
	return result, nil
}

// GetEntityType returns the entity type for a given entity ID.
// Returns empty string if no properties are stored for this entity.
func (s *PropertyStore) GetEntityType(ctx context.Context, entityID string) (string, error) {
	entry, err := s.client.EntityProperty.Query().
		Where(entityproperty.EntityIDEQ(entityID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", nil
		}
		return "", fmt.Errorf("property get type: %w", err)
	}
	return entry.EntityType, nil
}

// DeleteProperties removes all properties for an entity.
func (s *PropertyStore) DeleteProperties(ctx context.Context, entityID string) error {
	_, err := s.client.EntityProperty.Delete().
		Where(entityproperty.EntityIDEQ(entityID)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("property delete: %w", err)
	}
	return nil
}

// Query returns entity IDs matching type + property filters (AND semantics).
func (s *PropertyStore) Query(ctx context.Context, q PropertyQuery) ([]string, error) {
	if q.EntityType == "" {
		return nil, fmt.Errorf("property query: entityType is required")
	}
	if q.Limit <= 0 {
		q.Limit = 100
	}

	if len(q.Filters) == 0 {
		// No filters — return all entities of this type.
		entries, err := s.client.EntityProperty.Query().
			Where(entityproperty.EntityTypeEQ(q.EntityType)).
			Select(entityproperty.FieldEntityID).
			All(ctx)
		if err != nil {
			return nil, fmt.Errorf("property query all: %w", err)
		}
		return uniqueEntityIDs(entries, q.Limit, q.Offset), nil
	}

	// For each filter, find matching entity IDs, then intersect.
	var resultSet map[string]bool
	for i, f := range q.Filters {
		query := s.client.EntityProperty.Query().
			Where(
				entityproperty.EntityTypeEQ(q.EntityType),
				entityproperty.PropertyEQ(f.Property),
			)

		switch f.Op {
		case FilterEq:
			query = query.Where(entityproperty.ValueEQ(f.Value))
		case FilterNeq:
			query = query.Where(entityproperty.ValueNEQ(f.Value))
		case FilterContains:
			query = query.Where(entityproperty.ValueContains(f.Value))
		default:
			return nil, fmt.Errorf("property query: unsupported filter op %q", f.Op)
		}

		entries, err := query.Select(entityproperty.FieldEntityID).All(ctx)
		if err != nil {
			return nil, fmt.Errorf("property query filter %d: %w", i, err)
		}

		ids := make(map[string]bool, len(entries))
		for _, e := range entries {
			ids[e.EntityID] = true
		}

		if i == 0 {
			resultSet = ids
		} else {
			// AND intersection.
			for id := range resultSet {
				if !ids[id] {
					delete(resultSet, id)
				}
			}
		}

		if len(resultSet) == 0 {
			return nil, nil
		}
	}

	// Apply offset + limit. Sort for deterministic pagination.
	result := make([]string, 0, len(resultSet))
	for id := range resultSet {
		result = append(result, id)
	}
	sort.Strings(result)
	if q.Offset >= len(result) {
		return nil, nil
	}
	result = result[q.Offset:]
	if len(result) > q.Limit {
		result = result[:q.Limit]
	}
	return result, nil
}

func uniqueEntityIDs(entries []*ent.EntityProperty, limit, offset int) []string {
	seen := make(map[string]bool, len(entries))
	var result []string
	for _, e := range entries {
		if !seen[e.EntityID] {
			seen[e.EntityID] = true
			result = append(result, e.EntityID)
		}
	}
	if offset >= len(result) {
		return nil
	}
	result = result[offset:]
	if len(result) > limit {
		result = result[:limit]
	}
	return result
}
