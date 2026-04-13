package ontology_test

import (
	"context"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/ontology"
	"github.com/langoai/lango/internal/testutil"
)

func newActionTestEnv(t *testing.T) (*ontology.ServiceImpl, *ontology.ActionRegistry, *ontology.ActionExecutor) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	reg := ontology.NewEntRegistry(client)
	gs := testutil.NewMockGraphStore()
	svc := ontology.NewService(reg, gs)
	require.NoError(t, ontology.SeedDefaults(context.Background(), svc))

	cs := ontology.NewConflictStore(client)
	tm := ontology.NewTruthMaintainer(svc, gs, cs)
	svc.SetTruthMaintainer(tm)

	as := ontology.NewAliasStore(client)
	er := ontology.NewEntityResolver(as, gs, tm)
	svc.SetEntityResolver(er)

	ps := ontology.NewPropertyStore(client)
	svc.SetPropertyStore(ps)

	// ACL: ontologist=admin, chronicler=read
	acl := ontology.NewRoleBasedPolicy(map[string]ontology.Permission{
		"ontologist": ontology.PermAdmin,
		"chronicler": ontology.PermRead,
	})
	svc.SetACLPolicy(acl)

	actionReg := ontology.NewActionRegistry()
	logStore := ontology.NewActionLogStore(client)
	executor := ontology.NewActionExecutor(svc, actionReg, acl, logStore)

	return svc, actionReg, executor
}

func TestActionRegistry_RegisterAndGet(t *testing.T) {
	reg := ontology.NewActionRegistry()
	action := &ontology.ActionType{Name: "test_action", Description: "test"}

	require.NoError(t, reg.Register(action))

	got, ok := reg.Get("test_action")
	assert.True(t, ok)
	assert.Equal(t, "test_action", got.Name)
}

func TestActionRegistry_Duplicate(t *testing.T) {
	reg := ontology.NewActionRegistry()
	action := &ontology.ActionType{Name: "test_action"}

	require.NoError(t, reg.Register(action))
	assert.Error(t, reg.Register(action))
}

func TestActionRegistry_List(t *testing.T) {
	reg := ontology.NewActionRegistry()
	require.NoError(t, reg.Register(&ontology.ActionType{Name: "a"}))
	require.NoError(t, reg.Register(&ontology.ActionType{Name: "b"}))

	list := reg.List()
	assert.Len(t, list, 2)
}

func TestActionExecutor_HappyPath(t *testing.T) {
	_, actionReg, executor := newActionTestEnv(t)

	require.NoError(t, actionReg.Register(&ontology.ActionType{
		Name:         "noop",
		RequiredPerm: ontology.PermRead,
		Execute: func(ctx context.Context, svc ontology.OntologyService, params map[string]string) (*ontology.ActionEffects, error) {
			return &ontology.ActionEffects{}, nil
		},
	}))

	ctx := ctxkeys.WithPrincipal(context.Background(), "ontologist")
	result, err := executor.Execute(ctx, "noop", map[string]string{})
	require.NoError(t, err)
	assert.Equal(t, ontology.ActionCompleted, result.Status)
	assert.NotEqual(t, [16]byte{}, result.LogID)
}

func TestActionExecutor_PreconditionFail(t *testing.T) {
	_, actionReg, executor := newActionTestEnv(t)

	require.NoError(t, actionReg.Register(&ontology.ActionType{
		Name:         "guarded",
		RequiredPerm: ontology.PermRead,
		Precondition: func(ctx context.Context, svc ontology.OntologyService, params map[string]string) error {
			return errors.New("not ready")
		},
		Execute: func(ctx context.Context, svc ontology.OntologyService, params map[string]string) (*ontology.ActionEffects, error) {
			t.Fatal("Execute should not be called")
			return nil, nil
		},
	}))

	ctx := ctxkeys.WithPrincipal(context.Background(), "ontologist")
	_, err := executor.Execute(ctx, "guarded", nil)
	assert.ErrorContains(t, err, "precondition failed")
}

func TestActionExecutor_EffectFail_Compensate(t *testing.T) {
	_, actionReg, executor := newActionTestEnv(t)
	compensated := false

	require.NoError(t, actionReg.Register(&ontology.ActionType{
		Name:         "fail_and_comp",
		RequiredPerm: ontology.PermRead,
		Execute: func(ctx context.Context, svc ontology.OntologyService, params map[string]string) (*ontology.ActionEffects, error) {
			return &ontology.ActionEffects{
				FactsAsserted: []ontology.FactEffect{{Subject: "s", Predicate: "p", Object: "o"}},
			}, errors.New("boom")
		},
		Compensate: func(ctx context.Context, svc ontology.OntologyService, effects *ontology.ActionEffects) error {
			compensated = true
			return nil
		},
	}))

	ctx := ctxkeys.WithPrincipal(context.Background(), "ontologist")
	result, err := executor.Execute(ctx, "fail_and_comp", nil)
	require.NoError(t, err) // executor returns result, not error
	assert.Equal(t, ontology.ActionCompensated, result.Status)
	assert.True(t, compensated)
}

func TestActionExecutor_ACL_Denied(t *testing.T) {
	_, actionReg, executor := newActionTestEnv(t)

	require.NoError(t, actionReg.Register(&ontology.ActionType{
		Name:         "admin_action",
		RequiredPerm: ontology.PermAdmin,
		Execute: func(ctx context.Context, svc ontology.OntologyService, params map[string]string) (*ontology.ActionEffects, error) {
			t.Fatal("should not be called")
			return nil, nil
		},
	}))

	ctx := ctxkeys.WithPrincipal(context.Background(), "chronicler") // read-only
	_, err := executor.Execute(ctx, "admin_action", nil)
	assert.ErrorIs(t, err, ontology.ErrPermissionDenied)
}

func TestActionExecutor_NotFound(t *testing.T) {
	_, _, executor := newActionTestEnv(t)
	ctx := ctxkeys.WithPrincipal(context.Background(), "ontologist")
	_, err := executor.Execute(ctx, "nonexistent", nil)
	assert.ErrorContains(t, err, "not found")
}

func TestBuiltinLinkEntities(t *testing.T) {
	svc, actionReg, executor := newActionTestEnv(t)
	_ = svc

	require.NoError(t, actionReg.Register(ontology.BuiltinLinkEntities()))

	ctx := ctxkeys.WithPrincipal(context.Background(), "ontologist")
	result, err := executor.Execute(ctx, "link_entities", map[string]string{
		"subject":   "entity:1",
		"predicate": "related_to",
		"object":    "entity:2",
		"source":    "manual",
	})
	require.NoError(t, err)
	assert.Equal(t, ontology.ActionCompleted, result.Status)
	assert.Len(t, result.Effects.FactsAsserted, 1)
}

func TestBuiltinLinkEntities_MissingParam(t *testing.T) {
	_, actionReg, executor := newActionTestEnv(t)

	require.NoError(t, actionReg.Register(ontology.BuiltinLinkEntities()))

	ctx := ctxkeys.WithPrincipal(context.Background(), "ontologist")
	_, err := executor.Execute(ctx, "link_entities", map[string]string{
		"subject": "entity:1",
		// missing predicate, object, source
	})
	assert.ErrorContains(t, err, "missing required parameter")
}
