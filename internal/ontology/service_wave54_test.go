package ontology_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/ontology"

	_ "github.com/mattn/go-sqlite3"
)

func TestWave54ServiceSchemaAccessorsPropagateRegistryErrors(t *testing.T) {
	ctx := context.Background()
	reg := newWave26Registry()
	reg.types["Agent"] = ontology.ObjectType{Name: "Agent", Status: ontology.SchemaActive}
	reg.predicates["uses"] = ontology.PredicateDefinition{Name: "uses", Status: ontology.SchemaActive}
	svc := ontology.NewService(reg, nil)

	gotType, err := svc.GetType(ctx, "Agent")
	require.NoError(t, err)
	assert.Equal(t, "Agent", gotType.Name)

	gotPred, err := svc.GetPredicate(ctx, "uses")
	require.NoError(t, err)
	assert.Equal(t, "uses", gotPred.Name)

	reg.getTypeErr = errors.New("type lookup failed")
	gotType, err = svc.GetType(ctx, "Agent")
	require.Error(t, err)
	assert.Nil(t, gotType)
	assert.ErrorContains(t, err, "type lookup failed")

	reg.getTypeErr = nil
	reg.listTypesErr = errors.New("type list failed")
	gotTypes, err := svc.ListTypes(ctx)
	require.Error(t, err)
	assert.Nil(t, gotTypes)
	assert.ErrorContains(t, err, "type list failed")

	gotPred, err = svc.GetPredicate(ctx, "missing")
	require.Error(t, err)
	assert.Nil(t, gotPred)
	assert.ErrorContains(t, err, `predicate "missing" not found`)

	reg.listPredicatesErr = errors.New("predicate list failed")
	gotPreds, err := svc.ListPredicates(ctx)
	require.Error(t, err)
	assert.Nil(t, gotPreds)
	assert.ErrorContains(t, err, "predicate list failed")
}

func TestWave54ServiceDeprecateAndPromoteBranchesUpdateVersionAndCache(t *testing.T) {
	ctx := context.Background()
	reg := newWave26Registry()
	reg.types["Agent"] = ontology.ObjectType{Name: "Agent", Status: ontology.SchemaActive}
	reg.types["Candidate"] = ontology.ObjectType{Name: "Candidate", Status: ontology.SchemaProposed}
	reg.predicates["uses"] = ontology.PredicateDefinition{Name: "uses", Status: ontology.SchemaActive}
	reg.predicates["draft_rel"] = ontology.PredicateDefinition{Name: "draft_rel", Status: ontology.SchemaProposed}
	svc := ontology.NewService(reg, nil)

	assert.True(t, svc.PredicateValidator()("uses"))
	require.NoError(t, svc.DeprecateType(ctx, "Agent"))
	require.NoError(t, svc.DeprecatePredicate(ctx, "uses"))
	assert.False(t, svc.PredicateValidator()("uses"))
	version, err := svc.SchemaVersion(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, version)

	err = svc.DeprecateType(ctx, "missing")
	require.Error(t, err)
	assert.ErrorContains(t, err, `type "missing" not found`)
	version, err = svc.SchemaVersion(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, version)

	err = svc.PromotePredicate(ctx, "draft_rel", ontology.SchemaShadow, "without governance")
	require.Error(t, err)
	assert.ErrorContains(t, err, "governance not enabled")

	svc.SetGovernanceEngine(ontology.NewGovernanceEngine(ontology.GovernancePolicy{}))
	require.NoError(t, svc.PromoteType(ctx, "Candidate", ontology.SchemaShadow, "reviewed"))
	require.NoError(t, svc.PromotePredicate(ctx, "draft_rel", ontology.SchemaShadow, "reviewed"))
	assert.True(t, svc.PredicateValidator()("draft_rel"))
	version, err = svc.SchemaVersion(ctx)
	require.NoError(t, err)
	assert.Equal(t, 4, version)

	err = svc.PromoteType(ctx, "Candidate", ontology.SchemaDeprecated, "invalid direct transition")
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid schema transition")
}

func TestWave54ServiceDelegatesTruthResolverAndP2PFactBranches(t *testing.T) {
	ctx := context.Background()
	conflictID := uuid.New()
	validAt := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	truth := &wave54TruthMaintainer{
		assertResult: &ontology.AssertionResult{Stored: true},
		conflicts: []ontology.Conflict{{
			ID:        conflictID,
			Subject:   "agent:1",
			Predicate: "status",
			Status:    ontology.ConflictOpen,
		}},
		factsAt: []graph.Triple{{Subject: "agent:1", Predicate: "status", Object: "ready"}},
		openConflicts: []ontology.Conflict{{
			ID:        conflictID,
			Subject:   "agent:1",
			Predicate: "status",
			Status:    ontology.ConflictOpen,
		}},
	}
	resolver := &wave54Resolver{
		resolve: map[string]string{"alias": "canonical"},
		aliases: []string{"alias", "legacy"},
		mergeResult: &ontology.MergeResult{
			TriplesUpdated: 3,
			AliasesCreated: 1,
		},
	}
	svc := ontology.NewService(newWave26Registry(), nil)
	svc.SetTruthMaintainer(truth)
	svc.SetEntityResolver(resolver)

	asserted, err := svc.AssertFact(ctx, ontology.AssertionInput{
		Triple: graph.Triple{Subject: "agent:1", Predicate: "status", Object: "ready"},
		Source: "test",
	})
	require.NoError(t, err)
	assert.True(t, asserted.Stored)

	require.NoError(t, svc.RetractFact(ctx, "agent:1", "status", "ready", "stale"))
	conflicts, err := svc.ConflictSet(ctx, "agent:1", "status")
	require.NoError(t, err)
	assert.Equal(t, conflictID, conflicts[0].ID)
	require.NoError(t, svc.ResolveConflict(ctx, conflictID, "ready", "manual"))
	facts, err := svc.FactsAt(ctx, "agent:1", validAt)
	require.NoError(t, err)
	assert.Equal(t, "ready", facts[0].Object)
	open, err := svc.OpenConflicts(ctx)
	require.NoError(t, err)
	assert.Equal(t, conflictID, open[0].ID)

	resolved, err := svc.Resolve(ctx, "alias")
	require.NoError(t, err)
	assert.Equal(t, "canonical", resolved)
	require.NoError(t, svc.DeclareSameAs(ctx, "alias", "canonical", "manual"))
	merged, err := svc.MergeEntities(ctx, "canonical", "duplicate")
	require.NoError(t, err)
	assert.Equal(t, 3, merged.TriplesUpdated)
	require.NoError(t, svc.SplitEntity(ctx, "canonical", "alias"))
	aliases, err := svc.Aliases(ctx, "canonical")
	require.NoError(t, err)
	assert.Equal(t, []string{"alias", "legacy"}, aliases)

	p2pResult, err := svc.AssertP2PFact(ctx, ontology.P2PFactInput{
		Triple:     graph.Triple{Subject: "agent:1", Predicate: "status", Object: "ready"},
		PeerDID:    "did:peer:alice",
		PeerTrust:  0.75,
		Confidence: 0.9,
	})
	require.NoError(t, err)
	assert.True(t, p2pResult.Stored)
	require.Len(t, truth.assertInputs, 2)
	assert.Equal(t, "did:peer:alice", truth.assertInputs[1].Triple.Metadata[ontology.MetaRecordedBy])
	assert.Equal(t, 1, truth.retractCalls)
	assert.Equal(t, 1, truth.resolveConflictCalls)
	assert.Equal(t, 1, resolver.declareCalls)
	assert.Equal(t, 1, resolver.mergeCalls)
	assert.Equal(t, 1, resolver.splitCalls)
}

func TestWave54ServiceImportExportActionLogAndPermissionBranches(t *testing.T) {
	ctx := context.Background()
	reg := newWave26Registry()
	reg.types["Agent"] = ontology.ObjectType{Name: "Agent", Status: ontology.SchemaActive}
	reg.predicates["uses"] = ontology.PredicateDefinition{Name: "uses", Status: ontology.SchemaActive}
	svc := ontology.NewService(reg, nil)

	bundle, err := svc.ExportSchema(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, bundle.SchemaVersion)
	assert.Equal(t, "local", bundle.ExportedBy)
	require.Len(t, bundle.Types, 1)
	require.Len(t, bundle.Predicates, 1)

	imported, err := svc.ImportSchema(ctx, &ontology.SchemaBundle{
		Types: []ontology.SchemaTypeSlim{{
			Name:        "Skill",
			Description: "A callable capability",
			Properties:  []ontology.SchemaPropertySlim{{Name: "name", Type: "string", Required: true}},
		}},
		Predicates: []ontology.SchemaPredicateSlim{{
			Name:        "supports",
			Description: "Capability support",
			Cardinality: string(ontology.ManyToMany),
		}},
	}, ontology.ImportOptions{Mode: ontology.ImportShadow})
	require.NoError(t, err)
	assert.Equal(t, 1, imported.TypesAdded)
	assert.Equal(t, 1, imported.PredsAdded)
	assert.True(t, svc.PredicateValidator()("supports"))

	client := enttest.Open(t, "sqlite3", "file:wave54actions?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	actionReg := ontology.NewActionRegistry()
	require.NoError(t, actionReg.Register(&ontology.ActionType{
		Name:         "record_noop",
		Description:  "records a completed action",
		RequiredPerm: ontology.PermRead,
		Execute: func(context.Context, ontology.OntologyService, map[string]string) (*ontology.ActionEffects, error) {
			return &ontology.ActionEffects{
				FactsAsserted: []ontology.FactEffect{{Subject: "s", Predicate: "p", Object: "o"}},
			}, nil
		},
	}))
	svc.SetActionExecutor(ontology.NewActionExecutor(
		svc,
		actionReg,
		ontology.NewRoleBasedPolicy(map[string]ontology.Permission{"reader": ontology.PermRead}),
		ontology.NewActionLogStore(client),
	))

	readerCtx := ctxkeys.WithPrincipal(ctx, "reader")
	actionResult, err := svc.ExecuteAction(readerCtx, "record_noop", map[string]string{"k": "v"})
	require.NoError(t, err)
	assert.Equal(t, ontology.ActionCompleted, actionResult.Status)

	entry, err := svc.GetActionLog(ctx, actionResult.LogID)
	require.NoError(t, err)
	assert.Equal(t, "record_noop", entry.ActionName)
	assert.Equal(t, ontology.ActionCompleted, entry.Status)
	logs, err := svc.ListActionLogs(ctx, "record_noop", 1)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.Equal(t, actionResult.LogID, logs[0].ID)

	svc.SetACLPolicy(wave26DenyACL{})
	deniedCtx := ctxkeys.WithPrincipal(ctx, "blocked")
	_, err = svc.ExportSchema(deniedCtx)
	require.ErrorIs(t, err, ontology.ErrPermissionDenied)
	_, err = svc.ImportSchema(deniedCtx, bundle, ontology.ImportOptions{})
	require.ErrorIs(t, err, ontology.ErrPermissionDenied)
	_, err = svc.SchemaVersion(deniedCtx)
	require.ErrorIs(t, err, ontology.ErrPermissionDenied)
}

type wave54TruthMaintainer struct {
	assertResult         *ontology.AssertionResult
	conflicts            []ontology.Conflict
	factsAt              []graph.Triple
	openConflicts        []ontology.Conflict
	assertInputs         []ontology.AssertionInput
	retractCalls         int
	resolveConflictCalls int
}

func (tm *wave54TruthMaintainer) AssertFact(_ context.Context, input ontology.AssertionInput) (*ontology.AssertionResult, error) {
	tm.assertInputs = append(tm.assertInputs, input)
	return tm.assertResult, nil
}

func (tm *wave54TruthMaintainer) RetractFact(context.Context, string, string, string, string) error {
	tm.retractCalls++
	return nil
}

func (tm *wave54TruthMaintainer) ConflictSet(context.Context, string, string) ([]ontology.Conflict, error) {
	return append([]ontology.Conflict(nil), tm.conflicts...), nil
}

func (tm *wave54TruthMaintainer) ResolveConflict(context.Context, uuid.UUID, string, string) error {
	tm.resolveConflictCalls++
	return nil
}

func (tm *wave54TruthMaintainer) FactsAt(context.Context, string, time.Time) ([]graph.Triple, error) {
	return append([]graph.Triple(nil), tm.factsAt...), nil
}

func (tm *wave54TruthMaintainer) OpenConflicts(context.Context) ([]ontology.Conflict, error) {
	return append([]ontology.Conflict(nil), tm.openConflicts...), nil
}

type wave54Resolver struct {
	resolve      map[string]string
	aliases      []string
	mergeResult  *ontology.MergeResult
	declareCalls int
	mergeCalls   int
	splitCalls   int
}

func (r *wave54Resolver) Resolve(_ context.Context, rawID string) (string, error) {
	if canonical, ok := r.resolve[rawID]; ok {
		return canonical, nil
	}
	return rawID, nil
}

func (*wave54Resolver) RegisterAlias(context.Context, string, string, string) error { return nil }

func (r *wave54Resolver) DeclareSameAs(context.Context, string, string, string) error {
	r.declareCalls++
	return nil
}

func (r *wave54Resolver) Merge(context.Context, string, string) (*ontology.MergeResult, error) {
	r.mergeCalls++
	return r.mergeResult, nil
}

func (r *wave54Resolver) Split(context.Context, string, string) error {
	r.splitCalls++
	return nil
}

func (r *wave54Resolver) Aliases(context.Context, string) ([]string, error) {
	return append([]string(nil), r.aliases...), nil
}
