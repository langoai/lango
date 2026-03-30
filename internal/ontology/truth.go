package ontology

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/langoai/lango/internal/graph"
)

// ConflictStatus represents the lifecycle state of an ontology conflict.
type ConflictStatus string

const (
	ConflictOpen         ConflictStatus = "open"
	ConflictResolved     ConflictStatus = "resolved"
	ConflictAutoResolved ConflictStatus = "auto_resolved"
)

// AssertionInput carries a triple plus provenance metadata for fact assertion.
type AssertionInput struct {
	Triple     graph.Triple
	Source     string  // origin category (key in SourcePrecedence)
	Confidence float64 // 0.0–1.0
	ValidFrom  time.Time // zero = now
}

// AssertionResult reports the outcome of an AssertFact call.
type AssertionResult struct {
	Stored     bool
	ConflictID *uuid.UUID // non-nil when a conflict was created
	Message    string
}

// CandidateTriple is a serializable snapshot of a triple involved in a conflict.
type CandidateTriple struct {
	Subject    string `json:"subject"`
	Object     string `json:"object"`
	Source     string `json:"source"`
	Confidence string `json:"confidence"`
	RecordedAt string `json:"recorded_at"`
}

// Conflict represents a detected contradiction between triples.
type Conflict struct {
	ID         uuid.UUID
	Subject    string
	Predicate  string
	Candidates []CandidateTriple
	Status     ConflictStatus
	ResolvedAt *time.Time
	Resolution string
	CreatedAt  time.Time
}

// TruthMaintainer manages bi-temporal fact assertion, retraction, and conflict resolution.
type TruthMaintainer interface {
	// AssertFact stores a triple with temporal metadata and detects conflicts
	// based on predicate cardinality (OneToOne, ManyToOne).
	AssertFact(ctx context.Context, input AssertionInput) (*AssertionResult, error)

	// RetractFact sets ValidTo=now on a matching triple (soft delete).
	RetractFact(ctx context.Context, subject, predicate, object, reason string) error

	// ConflictSet returns conflicts for a given subject-predicate pair.
	ConflictSet(ctx context.Context, subject, predicate string) ([]Conflict, error)

	// ResolveConflict picks a winner and retracts losers.
	ResolveConflict(ctx context.Context, conflictID uuid.UUID, winnerObject, reason string) error

	// FactsAt returns triples valid at a specific point in time.
	FactsAt(ctx context.Context, subject string, validAt time.Time) ([]graph.Triple, error)

	// OpenConflicts returns all unresolved conflicts.
	OpenConflicts(ctx context.Context) ([]Conflict, error)
}

// truthMaintainer implements TruthMaintainer.
type truthMaintainer struct {
	svc           OntologyService
	store         graph.Store
	conflictStore *ConflictStore
}

// NewTruthMaintainer creates a TruthMaintainer backed by the given stores.
func NewTruthMaintainer(svc OntologyService, store graph.Store, cs *ConflictStore) TruthMaintainer {
	return &truthMaintainer{svc: svc, store: store, conflictStore: cs}
}

func (tm *truthMaintainer) AssertFact(ctx context.Context, input AssertionInput) (*AssertionResult, error) {
	// 1. Predicate validation — AssertFact must not bypass registry checks.
	if err := tm.svc.ValidateTriple(ctx, input.Triple); err != nil {
		return nil, fmt.Errorf("assert fact: %w", err)
	}

	// 2. Set temporal metadata BEFORE conflict detection
	//    (toCandidate reads metadata, so it must be populated first).
	if input.Triple.Metadata == nil {
		input.Triple.Metadata = make(map[string]string)
	}
	if input.ValidFrom.IsZero() {
		input.ValidFrom = time.Now()
	}
	now := time.Now()
	input.Triple.Metadata[MetaValidFrom] = input.ValidFrom.Format(time.RFC3339)
	input.Triple.Metadata[MetaRecordedAt] = now.Format(time.RFC3339)
	input.Triple.Metadata[MetaRecordedBy] = input.Source
	input.Triple.Metadata[MetaSource] = input.Source
	input.Triple.Metadata[MetaConfidence] = fmt.Sprintf("%.4f", input.Confidence)

	// 3. Cardinality-based conflict detection.
	var conflictID *uuid.UUID
	pred, err := tm.svc.GetPredicate(ctx, input.Triple.Predicate)
	if err == nil {
		switch pred.Cardinality {
		case OneToOne, ManyToOne:
			existing, qErr := tm.store.QueryBySubjectPredicate(ctx, input.Triple.Subject, input.Triple.Predicate)
			if qErr == nil {
				for _, e := range existing {
					if isCurrentlyValid(e) && e.Object != input.Triple.Object {
						if canAutoResolve(input.Source, e.Metadata[MetaSource]) {
							_ = tm.RetractFact(ctx, e.Subject, e.Predicate, e.Object, "auto: higher source precedence")
							// Record auto-resolved conflict for audit trail.
							_, _ = tm.conflictStore.Create(ctx, Conflict{
								Subject:    input.Triple.Subject,
								Predicate:  input.Triple.Predicate,
								Candidates: []CandidateTriple{toCandidate(e), toCandidate(input.Triple)},
								Status:     ConflictAutoResolved,
								Resolution: fmt.Sprintf("auto: source %q > %q", input.Source, e.Metadata[MetaSource]),
							})
						} else {
							c, cErr := tm.conflictStore.Create(ctx, Conflict{
								Subject:    input.Triple.Subject,
								Predicate:  input.Triple.Predicate,
								Candidates: []CandidateTriple{toCandidate(e), toCandidate(input.Triple)},
								Status:     ConflictOpen,
							})
							if cErr == nil {
								conflictID = &c.ID
							}
						}
					}
				}
			}
		case OneToMany, ManyToMany:
			// No subject-predicate conflict possible.
		}
	}
	// If GetPredicate failed (unregistered but validated): skip cardinality check, store anyway.

	// 4. Store the triple.
	if err := tm.store.AddTriple(ctx, input.Triple); err != nil {
		return nil, fmt.Errorf("assert fact store: %w", err)
	}

	msg := "stored"
	if conflictID != nil {
		msg = "stored with conflict"
	}
	return &AssertionResult{Stored: true, ConflictID: conflictID, Message: msg}, nil
}

func (tm *truthMaintainer) RetractFact(ctx context.Context, subject, predicate, object, reason string) error {
	triples, err := tm.store.QueryBySubjectPredicate(ctx, subject, predicate)
	if err != nil {
		return fmt.Errorf("retract fact query: %w", err)
	}

	// Retract the first currently-valid match. At most one valid triple should exist
	// per (subject, predicate, object) since AssertFact creates conflicts rather than
	// duplicates. If duplicates exist from direct store.AddTriple calls, only the first
	// is retracted — this is intentional to avoid cascading retractions.
	for _, t := range triples {
		if t.Object == object && isCurrentlyValid(t) {
			// Set ValidTo = now (soft delete). BoltDB has no update, so remove+add.
			if t.Metadata == nil {
				t.Metadata = make(map[string]string)
			}
			t.Metadata[MetaValidTo] = time.Now().Format(time.RFC3339)
			if err := tm.store.RemoveTriple(ctx, graph.Triple{
				Subject: t.Subject, Predicate: t.Predicate, Object: t.Object,
			}); err != nil {
				return fmt.Errorf("retract fact remove: %w", err)
			}
			if err := tm.store.AddTriple(ctx, t); err != nil {
				return fmt.Errorf("retract fact re-add: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("retract fact: no currently valid triple found for %s/%s/%s", subject, predicate, object)
}

func (tm *truthMaintainer) FactsAt(ctx context.Context, subject string, validAt time.Time) ([]graph.Triple, error) {
	all, err := tm.store.QueryBySubject(ctx, subject)
	if err != nil {
		return nil, err
	}

	var result []graph.Triple
	for _, t := range all {
		if isValidAt(t, validAt) {
			result = append(result, t)
		}
	}
	return result, nil
}

func (tm *truthMaintainer) ConflictSet(ctx context.Context, subject, predicate string) ([]Conflict, error) {
	return tm.conflictStore.ListBySubjectPredicate(ctx, subject, predicate)
}

func (tm *truthMaintainer) ResolveConflict(ctx context.Context, conflictID uuid.UUID, winnerObject, reason string) error {
	c, err := tm.conflictStore.Get(ctx, conflictID)
	if err != nil {
		return fmt.Errorf("resolve conflict get: %w", err)
	}

	// Retract all candidates except the winner.
	for _, cand := range c.Candidates {
		if cand.Object != winnerObject {
			_ = tm.RetractFact(ctx, c.Subject, c.Predicate, cand.Object, reason)
		}
	}

	return tm.conflictStore.Resolve(ctx, conflictID, reason)
}

func (tm *truthMaintainer) OpenConflicts(ctx context.Context) ([]Conflict, error) {
	return tm.conflictStore.ListOpen(ctx)
}

// isCurrentlyValid checks whether a triple is valid at the current moment.
// A triple is currently valid if:
//   - its ValidFrom is in the past or absent (legacy triples)
//   - its ValidTo is absent or in the future
func isCurrentlyValid(t graph.Triple) bool {
	return isValidAt(t, time.Now())
}

// isValidAt checks whether a triple is valid at a specific point in time.
func isValidAt(t graph.Triple, at time.Time) bool {
	if t.Metadata == nil {
		return true // legacy triple without temporal metadata
	}

	// Check ValidFrom — future facts are not yet valid.
	if vf, ok := t.Metadata[MetaValidFrom]; ok && vf != "" {
		validFrom, err := time.Parse(time.RFC3339, vf)
		if err == nil && at.Before(validFrom) {
			return false
		}
	}

	// Check ValidTo — retracted facts are no longer valid.
	if vt, ok := t.Metadata[MetaValidTo]; ok && vt != "" {
		validTo, err := time.Parse(time.RFC3339, vt)
		if err == nil && !at.Before(validTo) {
			return false
		}
	}

	return true
}

// canAutoResolve returns true if newSource has strictly higher precedence than existingSource.
func canAutoResolve(newSource, existingSource string) bool {
	newP := SourcePrecedence[newSource]
	existingP := SourcePrecedence[existingSource]
	return newP > existingP && existingP > 0
}

// toCandidate creates a CandidateTriple snapshot from a graph.Triple.
func toCandidate(t graph.Triple) CandidateTriple {
	c := CandidateTriple{
		Subject: t.Subject,
		Object:  t.Object,
	}
	if t.Metadata != nil {
		c.Source = t.Metadata[MetaSource]
		c.Confidence = t.Metadata[MetaConfidence]
		c.RecordedAt = t.Metadata[MetaRecordedAt]
	}
	return c
}
