package ontology_test

import (
	"context"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/ontology"
	"github.com/langoai/lango/internal/testutil"
)

// newTruthTestEnv creates a full truth maintenance test environment
// with seeded predicates and graph store.
func newTruthTestEnv(t *testing.T) (ontology.OntologyService, *testutil.MockGraphStore) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	reg := ontology.NewEntRegistry(client)
	gs := testutil.NewMockGraphStore()
	svc := ontology.NewService(reg, gs)

	// Seed defaults so predicates are available for cardinality checks.
	require.NoError(t, ontology.SeedDefaults(context.Background(), svc))

	// Wire truth maintenance.
	cs := ontology.NewConflictStore(client)
	tm := ontology.NewTruthMaintainer(svc, gs, cs)
	svc.SetTruthMaintainer(tm)

	return svc, gs
}

func TestTruthMaintainer_AssertFact_Basic(t *testing.T) {
	svc, gs := newTruthTestEnv(t)
	ctx := context.Background()

	result, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{
			Subject: "error:timeout", Predicate: graph.CausedBy, Object: "tool:http",
			SubjectType: "ErrorPattern", ObjectType: "Tool",
		},
		Source:     "graph_engine",
		Confidence: 0.8,
	})
	require.NoError(t, err)
	assert.True(t, result.Stored)
	assert.Nil(t, result.ConflictID)

	// Verify temporal metadata was set.
	triples, err := gs.QueryBySubject(ctx, "error:timeout")
	require.NoError(t, err)
	require.Len(t, triples, 1)
	assert.NotEmpty(t, triples[0].Metadata[ontology.MetaValidFrom])
	assert.NotEmpty(t, triples[0].Metadata[ontology.MetaRecordedAt])
	assert.Equal(t, "graph_engine", triples[0].Metadata[ontology.MetaSource])
	assert.Equal(t, "graph_engine", triples[0].Metadata[ontology.MetaRecordedBy])
	assert.Equal(t, "0.8000", triples[0].Metadata[ontology.MetaConfidence])
}

func TestTruthMaintainer_AssertFact_OneToOneConflict(t *testing.T) {
	svc, _ := newTruthTestEnv(t)
	ctx := context.Background()

	// Register a OneToOne predicate for testing.
	require.NoError(t, svc.RegisterPredicate(ctx, ontology.PredicateDefinition{
		Name: "primary_cause", Cardinality: ontology.OneToOne, Status: ontology.SchemaActive,
	}))

	// First assertion — no conflict.
	r1, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "error:x", Predicate: "primary_cause", Object: "cause_a"},
		Source: "graph_engine", Confidence: 0.7,
	})
	require.NoError(t, err)
	assert.Nil(t, r1.ConflictID)

	// Second assertion with different object — should create open conflict.
	r2, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "error:x", Predicate: "primary_cause", Object: "cause_b"},
		Source: "graph_engine", Confidence: 0.6,
	})
	require.NoError(t, err)
	assert.True(t, r2.Stored)
	assert.NotNil(t, r2.ConflictID, "should create a conflict")

	// Verify conflict exists.
	conflicts, err := svc.ConflictSet(ctx, "error:x", "primary_cause")
	require.NoError(t, err)
	assert.Len(t, conflicts, 1)
	assert.Equal(t, ontology.ConflictOpen, conflicts[0].Status)
	assert.Len(t, conflicts[0].Candidates, 2)
}

func TestTruthMaintainer_AssertFact_ManyToOneConflict(t *testing.T) {
	svc, _ := newTruthTestEnv(t)
	ctx := context.Background()

	// in_session is ManyToOne — one subject can only be in one session.
	// First assertion.
	r1, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "obs:1", Predicate: graph.InSession, Object: "session:a"},
		Source: "memory_hook", Confidence: 0.9,
	})
	require.NoError(t, err)
	assert.Nil(t, r1.ConflictID)

	// Second assertion with different object for same subject+predicate → conflict.
	r2, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "obs:1", Predicate: graph.InSession, Object: "session:b"},
		Source: "memory_hook", Confidence: 0.9,
	})
	require.NoError(t, err)
	assert.NotNil(t, r2.ConflictID, "ManyToOne should detect conflict on subject+predicate")
}

func TestTruthMaintainer_AssertFact_AutoResolve(t *testing.T) {
	svc, _ := newTruthTestEnv(t)
	ctx := context.Background()

	require.NoError(t, svc.RegisterPredicate(ctx, ontology.PredicateDefinition{
		Name: "root_cause", Cardinality: ontology.OneToOne, Status: ontology.SchemaActive,
	}))

	// Low-priority source asserts first.
	_, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "error:x", Predicate: "root_cause", Object: "cause_low"},
		Source: "llm_extraction", Confidence: 0.5,
	})
	require.NoError(t, err)

	// Higher-priority source asserts different object → auto-resolve.
	r2, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "error:x", Predicate: "root_cause", Object: "cause_high"},
		Source: "manual", Confidence: 0.9,
	})
	require.NoError(t, err)
	assert.True(t, r2.Stored)
	assert.Nil(t, r2.ConflictID, "auto-resolved should not leave open conflict")

	// Verify auto_resolved audit trail exists.
	conflicts, err := svc.ConflictSet(ctx, "error:x", "root_cause")
	require.NoError(t, err)
	require.Len(t, conflicts, 1)
	assert.Equal(t, ontology.ConflictAutoResolved, conflicts[0].Status)
	assert.Contains(t, conflicts[0].Resolution, "auto:")
}

func TestTruthMaintainer_AssertFact_OneToMany_NoConflict(t *testing.T) {
	svc, _ := newTruthTestEnv(t)
	ctx := context.Background()

	// "contains" is OneToMany — no subject+predicate conflict.
	r1, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "session:a", Predicate: graph.Contains, Object: "ref:1"},
		Source: "knowledge", Confidence: 0.9,
	})
	require.NoError(t, err)
	assert.Nil(t, r1.ConflictID)

	r2, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "session:a", Predicate: graph.Contains, Object: "ref:2"},
		Source: "knowledge", Confidence: 0.9,
	})
	require.NoError(t, err)
	assert.Nil(t, r2.ConflictID, "OneToMany should not create conflict")
}

func TestTruthMaintainer_AssertFact_ManyToMany_NoConflict(t *testing.T) {
	svc, _ := newTruthTestEnv(t)
	ctx := context.Background()

	// "caused_by" is ManyToMany — no conflict.
	r1, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "err:1", Predicate: graph.CausedBy, Object: "tool:a"},
		Source: "graph_engine", Confidence: 0.7,
	})
	require.NoError(t, err)
	assert.Nil(t, r1.ConflictID)

	r2, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "err:1", Predicate: graph.CausedBy, Object: "tool:b"},
		Source: "graph_engine", Confidence: 0.7,
	})
	require.NoError(t, err)
	assert.Nil(t, r2.ConflictID, "ManyToMany should not create conflict")
}

func TestTruthMaintainer_AssertFact_SourceMetadata(t *testing.T) {
	svc, gs := newTruthTestEnv(t)
	ctx := context.Background()

	_, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "a", Predicate: graph.RelatedTo, Object: "b"},
		Source: "manual", Confidence: 0.95,
	})
	require.NoError(t, err)

	triples, err := gs.QueryBySubject(ctx, "a")
	require.NoError(t, err)
	require.Len(t, triples, 1)
	assert.Equal(t, "manual", triples[0].Metadata[ontology.MetaSource], "_source must be recorded")
}

func TestTruthMaintainer_AssertFact_UnknownPredicate(t *testing.T) {
	svc, _ := newTruthTestEnv(t)
	ctx := context.Background()

	// Unknown predicate should be rejected by ValidateTriple.
	_, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "a", Predicate: "invented_pred", Object: "b"},
		Source: "manual", Confidence: 1.0,
	})
	assert.Error(t, err, "should reject unknown predicate")
	assert.Contains(t, err.Error(), "unknown or deprecated predicate")
}

func TestTruthMaintainer_RetractFact(t *testing.T) {
	svc, gs := newTruthTestEnv(t)
	ctx := context.Background()

	// Assert a fact.
	_, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "a", Predicate: graph.RelatedTo, Object: "b"},
		Source: "manual", Confidence: 0.9,
	})
	require.NoError(t, err)

	// Retract it.
	err = svc.RetractFact(ctx, "a", graph.RelatedTo, "b", "test retraction")
	require.NoError(t, err)

	// Verify ValidTo is set.
	triples, err := gs.QueryBySubject(ctx, "a")
	require.NoError(t, err)
	require.Len(t, triples, 1)
	assert.NotEmpty(t, triples[0].Metadata[ontology.MetaValidTo], "ValidTo should be set after retraction")
}

func TestTruthMaintainer_RetractFact_NotFound(t *testing.T) {
	svc, _ := newTruthTestEnv(t)
	ctx := context.Background()

	err := svc.RetractFact(ctx, "nonexistent", graph.RelatedTo, "x", "test")
	assert.Error(t, err, "should error when no valid triple found")
}

func TestTruthMaintainer_FactsAt_Current(t *testing.T) {
	svc, _ := newTruthTestEnv(t)
	ctx := context.Background()

	// Assert and then retract.
	_, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "node:1", Predicate: graph.RelatedTo, Object: "node:2"},
		Source: "manual", Confidence: 0.9,
	})
	require.NoError(t, err)

	_, err = svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "node:1", Predicate: graph.CausedBy, Object: "node:3"},
		Source: "manual", Confidence: 0.8,
	})
	require.NoError(t, err)

	// Retract one.
	err = svc.RetractFact(ctx, "node:1", graph.RelatedTo, "node:2", "test")
	require.NoError(t, err)

	// FactsAt(now) should only return the non-retracted one.
	facts, err := svc.FactsAt(ctx, "node:1", time.Now())
	require.NoError(t, err)
	assert.Len(t, facts, 1)
	assert.Equal(t, graph.CausedBy, facts[0].Predicate)
}

func TestTruthMaintainer_FactsAt_Past(t *testing.T) {
	svc, _ := newTruthTestEnv(t)
	ctx := context.Background()

	pastTime := time.Now().Add(-1 * time.Hour)

	// Assert with ValidFrom in the past.
	_, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple:   graph.Triple{Subject: "node:1", Predicate: graph.RelatedTo, Object: "node:2"},
		Source:   "manual",
		Confidence: 0.9,
		ValidFrom: pastTime,
	})
	require.NoError(t, err)

	// Retract it (sets ValidTo=now).
	err = svc.RetractFact(ctx, "node:1", graph.RelatedTo, "node:2", "test")
	require.NoError(t, err)

	// FactsAt(past + 30min) should include the retracted fact.
	facts, err := svc.FactsAt(ctx, "node:1", pastTime.Add(30*time.Minute))
	require.NoError(t, err)
	assert.Len(t, facts, 1, "retracted fact should be visible at past time point")
}

func TestTruthMaintainer_ResolveConflict(t *testing.T) {
	svc, _ := newTruthTestEnv(t)
	ctx := context.Background()

	require.NoError(t, svc.RegisterPredicate(ctx, ontology.PredicateDefinition{
		Name: "assigned_to", Cardinality: ontology.OneToOne, Status: ontology.SchemaActive,
	}))

	// Create two conflicting facts.
	_, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "task:1", Predicate: "assigned_to", Object: "user:alice"},
		Source: "manual", Confidence: 0.9,
	})
	require.NoError(t, err)

	r2, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "task:1", Predicate: "assigned_to", Object: "user:bob"},
		Source: "manual", Confidence: 0.8,
	})
	require.NoError(t, err)
	require.NotNil(t, r2.ConflictID)

	// Resolve: alice wins.
	err = svc.ResolveConflict(ctx, *r2.ConflictID, "user:alice", "manager decided")
	require.NoError(t, err)

	// Verify conflict is resolved.
	conflicts, err := svc.OpenConflicts(ctx)
	require.NoError(t, err)
	assert.Len(t, conflicts, 0, "no open conflicts after resolution")

	// FactsAt(now) should show only the winner.
	facts, err := svc.FactsAt(ctx, "task:1", time.Now())
	require.NoError(t, err)
	assert.Len(t, facts, 1)
	assert.Equal(t, "user:alice", facts[0].Object)
}

func TestTruthMaintainer_OpenConflicts(t *testing.T) {
	svc, _ := newTruthTestEnv(t)
	ctx := context.Background()

	require.NoError(t, svc.RegisterPredicate(ctx, ontology.PredicateDefinition{
		Name: "owner", Cardinality: ontology.OneToOne, Status: ontology.SchemaActive,
	}))

	// Create a conflict.
	_, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "item:1", Predicate: "owner", Object: "a"},
		Source: "manual", Confidence: 0.9,
	})
	require.NoError(t, err)

	_, err = svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "item:1", Predicate: "owner", Object: "b"},
		Source: "manual", Confidence: 0.8,
	})
	require.NoError(t, err)

	conflicts, err := svc.OpenConflicts(ctx)
	require.NoError(t, err)
	assert.Len(t, conflicts, 1)
}

func TestTruthMaintainer_BackwardCompat(t *testing.T) {
	svc, gs := newTruthTestEnv(t)
	ctx := context.Background()

	// Add a legacy triple directly to the store (no temporal metadata).
	require.NoError(t, gs.AddTriple(ctx, graph.Triple{
		Subject: "legacy:a", Predicate: graph.RelatedTo, Object: "legacy:b",
	}))

	// FactsAt should treat it as valid (no metadata = always valid).
	facts, err := svc.FactsAt(ctx, "legacy:a", time.Now())
	require.NoError(t, err)
	assert.Len(t, facts, 1, "legacy triple without temporal metadata should be considered valid")
}

func TestConflictStore_CRUD(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	cs := ontology.NewConflictStore(client)

	// Create.
	c, err := cs.Create(ctx, ontology.Conflict{
		Subject:   "s1",
		Predicate: "p1",
		Candidates: []ontology.CandidateTriple{
			{Subject: "s1", Object: "o1", Source: "manual"},
			{Subject: "s1", Object: "o2", Source: "llm_extraction"},
		},
		Status: ontology.ConflictOpen,
	})
	require.NoError(t, err)
	assert.Equal(t, "s1", c.Subject)
	assert.Equal(t, ontology.ConflictOpen, c.Status)
	assert.Len(t, c.Candidates, 2)

	// Get.
	got, err := cs.Get(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, c.ID, got.ID)

	// ListOpen.
	open, err := cs.ListOpen(ctx)
	require.NoError(t, err)
	assert.Len(t, open, 1)

	// Resolve.
	err = cs.Resolve(ctx, c.ID, "picked o1")
	require.NoError(t, err)

	// After resolve, no open conflicts.
	open, err = cs.ListOpen(ctx)
	require.NoError(t, err)
	assert.Len(t, open, 0)

	// ListBySubjectPredicate shows resolved.
	all, err := cs.ListBySubjectPredicate(ctx, "s1", "p1")
	require.NoError(t, err)
	assert.Len(t, all, 1)
	assert.Equal(t, ontology.ConflictResolved, all[0].Status)

	// Delete.
	err = cs.Delete(ctx, c.ID)
	require.NoError(t, err)
	all, err = cs.ListBySubjectPredicate(ctx, "s1", "p1")
	require.NoError(t, err)
	assert.Len(t, all, 0)
}

func TestTruthMaintainer_NotInitialized(t *testing.T) {
	svc := newTestService(t) // no truth maintainer set
	ctx := context.Background()

	_, err := svc.AssertFact(ctx, ontology.AssertionInput{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "truth maintenance not initialized")
}

func TestTruthMaintainer_AssertFact_LowerSourceCreatesOpenConflict(t *testing.T) {
	svc, _ := newTruthTestEnv(t)
	ctx := context.Background()

	require.NoError(t, svc.RegisterPredicate(ctx, ontology.PredicateDefinition{
		Name: "owner_pred", Cardinality: ontology.OneToOne, Status: ontology.SchemaActive,
	}))

	// High-priority source asserts first.
	_, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "item:1", Predicate: "owner_pred", Object: "user:alice"},
		Source: "manual", Confidence: 0.9,
	})
	require.NoError(t, err)

	// Lower-priority source asserts different object → open conflict (NOT auto-resolved).
	r2, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "item:1", Predicate: "owner_pred", Object: "user:bob"},
		Source: "graph_engine", Confidence: 0.5,
	})
	require.NoError(t, err)
	assert.True(t, r2.Stored)
	assert.NotNil(t, r2.ConflictID, "lower→higher should create open conflict, not auto-resolve")

	// Verify it's open (not auto_resolved).
	conflicts, err := svc.ConflictSet(ctx, "item:1", "owner_pred")
	require.NoError(t, err)
	require.Len(t, conflicts, 1)
	assert.Equal(t, ontology.ConflictOpen, conflicts[0].Status)
}

func TestTruthMaintainer_AssertFact_EqualPrecedenceCreatesOpenConflict(t *testing.T) {
	svc, _ := newTruthTestEnv(t)
	ctx := context.Background()

	require.NoError(t, svc.RegisterPredicate(ctx, ontology.PredicateDefinition{
		Name: "lead_pred", Cardinality: ontology.OneToOne, Status: ontology.SchemaActive,
	}))

	// Same source asserts first.
	_, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "team:1", Predicate: "lead_pred", Object: "user:alice"},
		Source: "manual", Confidence: 0.9,
	})
	require.NoError(t, err)

	// Same source asserts different object → open conflict (equal precedence = no auto-resolve).
	r2, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "team:1", Predicate: "lead_pred", Object: "user:bob"},
		Source: "manual", Confidence: 0.8,
	})
	require.NoError(t, err)
	assert.NotNil(t, r2.ConflictID, "equal precedence should create open conflict")

	conflicts, err := svc.ConflictSet(ctx, "team:1", "lead_pred")
	require.NoError(t, err)
	require.Len(t, conflicts, 1)
	assert.Equal(t, ontology.ConflictOpen, conflicts[0].Status)
}

func TestTruthMaintainer_AssertFact_FutureValidFromExcludedFromConflict(t *testing.T) {
	svc, gs := newTruthTestEnv(t)
	ctx := context.Background()

	require.NoError(t, svc.RegisterPredicate(ctx, ontology.PredicateDefinition{
		Name: "next_owner", Cardinality: ontology.OneToOne, Status: ontology.SchemaActive,
	}))

	// Add a future-dated fact directly to the store (simulating a scheduled fact).
	futureTime := time.Now().Add(24 * time.Hour)
	require.NoError(t, gs.AddTriple(ctx, graph.Triple{
		Subject: "item:1", Predicate: "next_owner", Object: "user:future",
		Metadata: map[string]string{
			ontology.MetaValidFrom:  futureTime.Format(time.RFC3339),
			ontology.MetaSource:     "manual",
			ontology.MetaConfidence: "0.9000",
			ontology.MetaRecordedAt: time.Now().Format(time.RFC3339),
		},
	}))

	// Assert a current fact — should NOT conflict with future-dated fact.
	r, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "item:1", Predicate: "next_owner", Object: "user:current"},
		Source: "manual", Confidence: 0.9,
	})
	require.NoError(t, err)
	assert.Nil(t, r.ConflictID, "future-dated fact should not trigger present-time conflict")
}
