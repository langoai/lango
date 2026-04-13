package ontology

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/ontologyconflict"
)

// ConflictStore provides Ent-backed CRUD for ontology conflicts.
type ConflictStore struct {
	client *ent.Client
}

// NewConflictStore creates a ConflictStore backed by the given Ent client.
func NewConflictStore(client *ent.Client) *ConflictStore {
	return &ConflictStore{client: client}
}

// Create persists a new conflict record.
func (s *ConflictStore) Create(ctx context.Context, c Conflict) (*Conflict, error) {
	candidatesJSON, err := marshalCandidates(c.Candidates)
	if err != nil {
		return nil, fmt.Errorf("conflict store create marshal: %w", err)
	}

	builder := s.client.OntologyConflict.Create().
		SetSubject(c.Subject).
		SetPredicate(c.Predicate).
		SetCandidates(candidatesJSON)

	switch c.Status {
	case ConflictResolved:
		builder.SetStatus(ontologyconflict.StatusResolved)
	case ConflictAutoResolved:
		builder.SetStatus(ontologyconflict.StatusAutoResolved)
		now := time.Now()
		builder.SetResolvedAt(now)
		if c.Resolution != "" {
			builder.SetResolution(c.Resolution)
		}
	default:
		builder.SetStatus(ontologyconflict.StatusOpen)
	}

	e, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("conflict store create: %w", err)
	}

	return entToConflict(e)
}

// Get retrieves a conflict by ID.
func (s *ConflictStore) Get(ctx context.Context, id uuid.UUID) (*Conflict, error) {
	e, err := s.client.OntologyConflict.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("conflict store get %s: %w", id, err)
	}
	return entToConflict(e)
}

// ListBySubjectPredicate returns conflicts matching subject and predicate.
func (s *ConflictStore) ListBySubjectPredicate(ctx context.Context, subject, predicate string) ([]Conflict, error) {
	entries, err := s.client.OntologyConflict.Query().
		Where(
			ontologyconflict.SubjectEQ(subject),
			ontologyconflict.PredicateEQ(predicate),
		).
		Order(ent.Desc(ontologyconflict.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("conflict store list by s/p: %w", err)
	}
	return entsToConflicts(entries)
}

// ListOpen returns all conflicts with status "open".
func (s *ConflictStore) ListOpen(ctx context.Context) ([]Conflict, error) {
	entries, err := s.client.OntologyConflict.Query().
		Where(ontologyconflict.StatusEQ(ontologyconflict.StatusOpen)).
		Order(ent.Desc(ontologyconflict.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("conflict store list open: %w", err)
	}
	return entsToConflicts(entries)
}

// Resolve marks a conflict as resolved with a reason.
func (s *ConflictStore) Resolve(ctx context.Context, id uuid.UUID, resolution string) error {
	now := time.Now()
	_, err := s.client.OntologyConflict.UpdateOneID(id).
		SetStatus(ontologyconflict.StatusResolved).
		SetResolution(resolution).
		SetResolvedAt(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("conflict store resolve %s: %w", id, err)
	}
	return nil
}

// Delete removes a conflict record (for reconciliation cleanup).
func (s *ConflictStore) Delete(ctx context.Context, id uuid.UUID) error {
	return s.client.OntologyConflict.DeleteOneID(id).Exec(ctx)
}

// entToConflict converts an Ent entity to a domain Conflict.
func entToConflict(e *ent.OntologyConflict) (*Conflict, error) {
	candidates, err := unmarshalCandidates(e.Candidates)
	if err != nil {
		return nil, err
	}

	c := &Conflict{
		ID:         e.ID,
		Subject:    e.Subject,
		Predicate:  e.Predicate,
		Candidates: candidates,
		Status:     ConflictStatus(e.Status),
		CreatedAt:  e.CreatedAt,
	}
	if e.Resolution != nil {
		c.Resolution = *e.Resolution
	}
	if e.ResolvedAt != nil {
		c.ResolvedAt = e.ResolvedAt
	}
	return c, nil
}

func entsToConflicts(entries []*ent.OntologyConflict) ([]Conflict, error) {
	result := make([]Conflict, 0, len(entries))
	for _, e := range entries {
		c, err := entToConflict(e)
		if err != nil {
			return nil, err
		}
		result = append(result, *c)
	}
	return result, nil
}

// marshalCandidates converts typed CandidateTriple to Ent's JSON field type.
// Uses []map[string]interface{} because Ent's field.JSON requires the Go type
// to match the schema declaration. CandidateTriple lives in the ontology package
// (not schema), so we can't use it directly in the Ent schema.
func marshalCandidates(candidates []CandidateTriple) ([]map[string]interface{}, error) {
	data, err := json.Marshal(candidates)
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func unmarshalCandidates(raw []map[string]interface{}) ([]CandidateTriple, error) {
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var result []CandidateTriple
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}
