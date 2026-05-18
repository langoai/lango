package ontology_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/ontology"
)

func TestWave26ServiceSchemaHealthAndTypeUsageWithoutGovernance(t *testing.T) {
	createdAt := time.Date(2026, 5, 18, 9, 30, 0, 0, time.UTC)
	reg := newWave26Registry()
	reg.types["Agent"] = ontology.ObjectType{
		Name:      "Agent",
		Status:    ontology.SchemaActive,
		Version:   7,
		CreatedAt: createdAt,
	}
	reg.types["Skill"] = ontology.ObjectType{Name: "Skill", Status: ontology.SchemaProposed}
	reg.types["Legacy"] = ontology.ObjectType{Name: "Legacy", Status: ontology.SchemaDeprecated}
	reg.predicates["uses"] = ontology.PredicateDefinition{Name: "uses", Status: ontology.SchemaActive}
	reg.predicates["observes"] = ontology.PredicateDefinition{Name: "observes", Status: ontology.SchemaShadow}
	reg.predicates["old_rel"] = ontology.PredicateDefinition{Name: "old_rel", Status: ontology.SchemaDeprecated}

	svc := ontology.NewService(reg, nil)

	health, err := svc.SchemaHealth(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, health.Types[ontology.SchemaActive])
	assert.Equal(t, 1, health.Types[ontology.SchemaProposed])
	assert.Equal(t, 1, health.Types[ontology.SchemaDeprecated])
	assert.Equal(t, 1, health.Predicates[ontology.SchemaActive])
	assert.Equal(t, 1, health.Predicates[ontology.SchemaShadow])
	assert.Equal(t, 1, health.Predicates[ontology.SchemaDeprecated])

	usage, err := svc.TypeUsage(context.Background(), "Agent")
	require.NoError(t, err)
	assert.Equal(t, "Agent", usage.TypeName)
	assert.Equal(t, ontology.SchemaActive, usage.Status)
	assert.Equal(t, 7, usage.Version)
	assert.Equal(t, createdAt, usage.CreatedAt)
}

func TestWave26ServiceGovernanceDelegationErrors(t *testing.T) {
	errBoom := errors.New("boom")
	reg := newWave26Registry()
	reg.types["Agent"] = ontology.ObjectType{Name: "Agent", Status: ontology.SchemaActive}
	reg.listPredicatesErr = errBoom
	svc := ontology.NewService(reg, nil)
	svc.SetGovernanceEngine(ontology.NewGovernanceEngine(ontology.GovernancePolicy{}))

	health, err := svc.SchemaHealth(context.Background())
	require.Error(t, err)
	assert.Nil(t, health)
	assert.ErrorIs(t, err, errBoom)
	assert.ErrorContains(t, err, "list predicates")

	reg.listPredicatesErr = nil
	reg.getTypeErr = errBoom
	usage, err := svc.TypeUsage(context.Background(), "Agent")
	require.Error(t, err)
	assert.Nil(t, usage)
	assert.ErrorIs(t, err, errBoom)
	assert.ErrorContains(t, err, `get type "Agent"`)
}

func TestWave26ServicePermissionDeniedShortCircuitsDelegation(t *testing.T) {
	reg := newWave26Registry()
	svc := ontology.NewService(reg, nil)
	svc.SetACLPolicy(wave26DenyACL{})
	reg.resetCalls()

	ctx := ctxkeys.WithPrincipal(context.Background(), "blocked")

	health, err := svc.SchemaHealth(ctx)
	require.ErrorIs(t, err, ontology.ErrPermissionDenied)
	assert.Nil(t, health)
	assert.Zero(t, reg.calls["ListTypes"])
	assert.Zero(t, reg.calls["ListPredicates"])

	usage, err := svc.TypeUsage(ctx, "Agent")
	require.ErrorIs(t, err, ontology.ErrPermissionDenied)
	assert.Nil(t, usage)
	assert.Zero(t, reg.calls["GetType"])

	err = svc.StoreTriple(ctx, graph.Triple{Subject: "s", Predicate: "p", Object: "o"})
	require.ErrorIs(t, err, ontology.ErrPermissionDenied)
}

func TestWave26ServiceMissingDependencyDelegationErrors(t *testing.T) {
	svc := ontology.NewService(newWave26Registry(), nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		run     func() error
		wantErr string
	}{
		{
			name: "StoreTripleWithoutGraphStore",
			run: func() error {
				return svc.StoreTriple(ctx, graph.Triple{Subject: "s", Predicate: "p", Object: "o"})
			},
			wantErr: "graph store not available",
		},
		{
			name: "AssertFactWithoutTruthMaintainer",
			run: func() error {
				_, err := svc.AssertFact(ctx, ontology.AssertionInput{})
				return err
			},
			wantErr: "truth maintenance not initialized",
		},
		{
			name: "DeclareSameAsWithoutResolver",
			run: func() error {
				return svc.DeclareSameAs(ctx, "alias", "canonical", "test")
			},
			wantErr: "entity resolver not initialized",
		},
		{
			name: "QueryTriplesWithoutGraphStore",
			run: func() error {
				_, err := svc.QueryTriples(ctx, "subject")
				return err
			},
			wantErr: "graph store not available",
		},
		{
			name: "SetEntityPropertyWithoutPropertyStore",
			run: func() error {
				return svc.SetEntityProperty(ctx, "entity:1", "Agent", "status", "active")
			},
			wantErr: "property store not initialized",
		},
		{
			name: "PromoteTypeWithoutGovernance",
			run: func() error {
				return svc.PromoteType(ctx, "Agent", ontology.SchemaActive, "test")
			},
			wantErr: "governance not enabled",
		},
		{
			name: "ExecuteActionWithoutExecutor",
			run: func() error {
				_, err := svc.ExecuteAction(ctx, "missing", nil)
				return err
			},
			wantErr: "action executor not initialized",
		},
		{
			name: "GetActionLogWithoutExecutor",
			run: func() error {
				_, err := svc.GetActionLog(ctx, uuid.New())
				return err
			},
			wantErr: "action executor not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.run()
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}

	actions, err := svc.ListActions(ctx)
	require.NoError(t, err)
	assert.Nil(t, actions)

	logs, err := svc.ListActionLogs(ctx, "anything", 10)
	require.NoError(t, err)
	assert.Nil(t, logs)
}

func TestWave26ServiceActionExecutorDelegatesListAndPreLogErrors(t *testing.T) {
	svc := ontology.NewService(newWave26Registry(), nil)
	acl := ontology.NewRoleBasedPolicy(map[string]ontology.Permission{
		"admin":  ontology.PermAdmin,
		"reader": ontology.PermRead,
	})
	actionReg := ontology.NewActionRegistry()
	executor := ontology.NewActionExecutor(svc, actionReg, acl, nil)
	svc.SetActionExecutor(executor)

	require.NoError(t, actionReg.Register(&ontology.ActionType{
		Name:         "admin_only",
		Description:  "requires admin before log creation",
		RequiredPerm: ontology.PermAdmin,
		ParamSchema: map[string]string{
			"target": "Target entity",
		},
		Execute: func(context.Context, ontology.OntologyService, map[string]string) (*ontology.ActionEffects, error) {
			t.Fatal("Execute should not run when ACL denies access")
			return nil, nil
		},
	}))
	require.NoError(t, actionReg.Register(&ontology.ActionType{
		Name:         "guarded",
		Description:  "fails precondition before log creation",
		RequiredPerm: ontology.PermRead,
		ParamSchema: map[string]string{
			"required": "Required value",
		},
		Precondition: func(_ context.Context, _ ontology.OntologyService, params map[string]string) error {
			if params["required"] == "" {
				return errors.New("required param is empty")
			}
			return nil
		},
		Execute: func(context.Context, ontology.OntologyService, map[string]string) (*ontology.ActionEffects, error) {
			t.Fatal("Execute should not run when precondition fails")
			return nil, nil
		},
	}))

	summaries, err := svc.ListActions(context.Background())
	require.NoError(t, err)
	adminSummary := requireWave26ActionSummary(t, summaries, "admin_only")
	assert.Equal(t, "requires admin before log creation", adminSummary.Description)
	assert.Equal(t, ontology.PermAdmin, adminSummary.RequiredPerm)
	assert.Equal(t, "Target entity", adminSummary.ParamSchema["target"])

	readerCtx := ctxkeys.WithPrincipal(context.Background(), "reader")
	result, err := svc.ExecuteAction(readerCtx, "admin_only", map[string]string{"target": "entity:1"})
	require.ErrorIs(t, err, ontology.ErrPermissionDenied)
	assert.Nil(t, result)

	adminCtx := ctxkeys.WithPrincipal(context.Background(), "admin")
	result, err = svc.ExecuteAction(adminCtx, "guarded", map[string]string{})
	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "precondition failed")
	assert.ErrorContains(t, err, "required param is empty")

	result, err = svc.ExecuteAction(adminCtx, "missing", nil)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, `action "missing" not found`)
}

func requireWave26ActionSummary(
	t *testing.T,
	summaries []ontology.ActionSummary,
	name string,
) ontology.ActionSummary {
	t.Helper()
	for _, summary := range summaries {
		if summary.Name == name {
			return summary
		}
	}
	require.FailNowf(t, "missing action summary", "action %q not found in %v", name, summaries)
	return ontology.ActionSummary{}
}

type wave26DenyACL struct{}

func (wave26DenyACL) Check(principal string, required ontology.Permission) error {
	return fmt.Errorf("%w: principal %q requires %d", ontology.ErrPermissionDenied, principal, required)
}

type wave26Registry struct {
	types      map[string]ontology.ObjectType
	predicates map[string]ontology.PredicateDefinition
	calls      map[string]int

	getTypeErr        error
	listTypesErr      error
	listPredicatesErr error
}

func newWave26Registry() *wave26Registry {
	return &wave26Registry{
		types:      make(map[string]ontology.ObjectType),
		predicates: make(map[string]ontology.PredicateDefinition),
		calls:      make(map[string]int),
	}
}

func (r *wave26Registry) resetCalls() {
	r.calls = make(map[string]int)
}

func (r *wave26Registry) RegisterType(_ context.Context, objType ontology.ObjectType) error {
	r.calls["RegisterType"]++
	r.types[objType.Name] = objType
	return nil
}

func (r *wave26Registry) GetType(_ context.Context, name string) (*ontology.ObjectType, error) {
	r.calls["GetType"]++
	if r.getTypeErr != nil {
		return nil, r.getTypeErr
	}
	objType, ok := r.types[name]
	if !ok {
		return nil, fmt.Errorf("type %q not found", name)
	}
	return &objType, nil
}

func (r *wave26Registry) ListTypes(context.Context) ([]ontology.ObjectType, error) {
	r.calls["ListTypes"]++
	if r.listTypesErr != nil {
		return nil, r.listTypesErr
	}
	types := make([]ontology.ObjectType, 0, len(r.types))
	for _, objType := range r.types {
		types = append(types, objType)
	}
	return types, nil
}

func (r *wave26Registry) DeprecateType(_ context.Context, name string) error {
	r.calls["DeprecateType"]++
	objType, ok := r.types[name]
	if !ok {
		return fmt.Errorf("type %q not found", name)
	}
	objType.Status = ontology.SchemaDeprecated
	r.types[name] = objType
	return nil
}

func (r *wave26Registry) RegisterPredicate(
	_ context.Context,
	pred ontology.PredicateDefinition,
) error {
	r.calls["RegisterPredicate"]++
	r.predicates[pred.Name] = pred
	return nil
}

func (r *wave26Registry) GetPredicate(
	_ context.Context,
	name string,
) (*ontology.PredicateDefinition, error) {
	r.calls["GetPredicate"]++
	pred, ok := r.predicates[name]
	if !ok {
		return nil, fmt.Errorf("predicate %q not found", name)
	}
	return &pred, nil
}

func (r *wave26Registry) ListPredicates(context.Context) ([]ontology.PredicateDefinition, error) {
	r.calls["ListPredicates"]++
	if r.listPredicatesErr != nil {
		return nil, r.listPredicatesErr
	}
	preds := make([]ontology.PredicateDefinition, 0, len(r.predicates))
	for _, pred := range r.predicates {
		preds = append(preds, pred)
	}
	return preds, nil
}

func (r *wave26Registry) DeprecatePredicate(_ context.Context, name string) error {
	r.calls["DeprecatePredicate"]++
	pred, ok := r.predicates[name]
	if !ok {
		return fmt.Errorf("predicate %q not found", name)
	}
	pred.Status = ontology.SchemaDeprecated
	r.predicates[name] = pred
	return nil
}

func (r *wave26Registry) UpdateTypeStatus(
	_ context.Context,
	name string,
	status ontology.SchemaStatus,
) error {
	r.calls["UpdateTypeStatus"]++
	objType, ok := r.types[name]
	if !ok {
		return fmt.Errorf("type %q not found", name)
	}
	objType.Status = status
	r.types[name] = objType
	return nil
}

func (r *wave26Registry) UpdatePredicateStatus(
	_ context.Context,
	name string,
	status ontology.SchemaStatus,
) error {
	r.calls["UpdatePredicateStatus"]++
	pred, ok := r.predicates[name]
	if !ok {
		return fmt.Errorf("predicate %q not found", name)
	}
	pred.Status = status
	r.predicates[name] = pred
	return nil
}
