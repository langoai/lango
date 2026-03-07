package team

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeam_AddAndGetMember(t *testing.T) {
	t.Parallel()

	tm := NewTeam("t1", "test-team", "solve problem", "did:leader", 5)

	m := &Member{DID: "did:1", Name: "worker-1", Role: RoleWorker}
	require.NoError(t, tm.AddMember(m))

	got := tm.GetMember("did:1")
	require.NotNil(t, got)
	assert.Equal(t, "worker-1", got.Name)
	assert.False(t, got.JoinedAt.IsZero(), "JoinedAt should be set")
}

func TestTeam_AddDuplicate(t *testing.T) {
	t.Parallel()

	tm := NewTeam("t1", "test-team", "goal", "did:leader", 5)
	m := &Member{DID: "did:1", Name: "worker-1"}

	_ = tm.AddMember(m)
	err := tm.AddMember(m)
	assert.ErrorIs(t, err, ErrAlreadyMember)
}

func TestTeam_MaxCapacity(t *testing.T) {
	t.Parallel()

	tm := NewTeam("t1", "test-team", "goal", "did:leader", 1)
	_ = tm.AddMember(&Member{DID: "did:1"})

	err := tm.AddMember(&Member{DID: "did:2"})
	assert.ErrorIs(t, err, ErrTeamFull)
}

func TestTeam_AddToDisbanded(t *testing.T) {
	t.Parallel()

	tm := NewTeam("t1", "test-team", "goal", "did:leader", 5)
	tm.Disband()

	err := tm.AddMember(&Member{DID: "did:1"})
	assert.ErrorIs(t, err, ErrTeamDisbanded)
}

func TestTeam_RemoveMember(t *testing.T) {
	t.Parallel()

	tm := NewTeam("t1", "test-team", "goal", "did:leader", 5)
	_ = tm.AddMember(&Member{DID: "did:1"})

	require.NoError(t, tm.RemoveMember("did:1"))
	assert.Equal(t, 0, tm.MemberCount())
}

func TestTeam_RemoveNotMember(t *testing.T) {
	t.Parallel()

	tm := NewTeam("t1", "test-team", "goal", "did:leader", 5)

	err := tm.RemoveMember("did:nonexistent")
	assert.ErrorIs(t, err, ErrNotMember)
}

func TestTeam_Lifecycle(t *testing.T) {
	t.Parallel()

	tm := NewTeam("t1", "test-team", "goal", "did:leader", 5)
	assert.Equal(t, StatusForming, tm.Status)

	tm.Activate()
	assert.Equal(t, StatusActive, tm.Status)

	tm.Disband()
	assert.Equal(t, StatusDisbanded, tm.Status)
	assert.False(t, tm.DisbandedAt.IsZero(), "DisbandedAt should be set after Disband")
}

func TestScopedContext_Roundtrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	sc := ScopedContext{TeamID: "t1", MemberDID: "did:1", Role: RoleWorker}

	ctx = WithScopedContext(ctx, sc)
	got, ok := ScopedContextFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, "t1", got.TeamID)
	assert.Equal(t, "did:1", got.MemberDID)
	assert.Equal(t, RoleWorker, got.Role)
}

func TestScopedContext_Missing(t *testing.T) {
	t.Parallel()

	_, ok := ScopedContextFromContext(context.Background())
	assert.False(t, ok)
}

func TestTeam_ActiveMembers(t *testing.T) {
	t.Parallel()

	tm := NewTeam("t1", "test-team", "goal", "did:leader", 10)
	_ = tm.AddMember(&Member{DID: "did:1", Status: MemberIdle})
	_ = tm.AddMember(&Member{DID: "did:2", Status: MemberBusy})
	_ = tm.AddMember(&Member{DID: "did:3", Status: MemberLeft})
	_ = tm.AddMember(&Member{DID: "did:4", Status: MemberFailed})

	active := tm.ActiveMembers()
	assert.Len(t, active, 2, "idle + busy")
}

func TestTeam_Budget(t *testing.T) {
	t.Parallel()

	tm := NewTeam("t1", "test-team", "goal", "did:leader", 5)
	tm.Budget = 10.0

	require.NoError(t, tm.AddSpend(5.0))
	assert.Equal(t, 5.0, tm.Spent)

	err := tm.AddSpend(6.0)
	assert.Error(t, err, "should fail when exceeding budget")
}

func TestContextFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give        string
		filter      ContextFilter
		metadata    map[string]string
		wantKeys    []string
		wantMissing []string
	}{
		{
			give:        "exclude keys",
			filter:      ContextFilter{ExcludeKeys: []string{"secret"}},
			metadata:    map[string]string{"name": "test", "secret": "hidden"},
			wantKeys:    []string{"name"},
			wantMissing: []string{"secret"},
		},
		{
			give:        "allow keys",
			filter:      ContextFilter{AllowedKeys: []string{"name"}},
			metadata:    map[string]string{"name": "test", "other": "data"},
			wantKeys:    []string{"name"},
			wantMissing: []string{"other"},
		},
		{
			give:        "allow and exclude",
			filter:      ContextFilter{AllowedKeys: []string{"name", "secret"}, ExcludeKeys: []string{"secret"}},
			metadata:    map[string]string{"name": "test", "secret": "hidden", "other": "data"},
			wantKeys:    []string{"name"},
			wantMissing: []string{"secret", "other"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			result := tt.filter.Filter(tt.metadata)
			for _, k := range tt.wantKeys {
				assert.Contains(t, result, k)
			}
			for _, k := range tt.wantMissing {
				assert.NotContains(t, result, k)
			}
		})
	}
}

func TestContextFilter_NilMetadata(t *testing.T) {
	t.Parallel()

	f := ContextFilter{AllowedKeys: []string{"name"}}
	result := f.Filter(nil)
	assert.Nil(t, result)
}
