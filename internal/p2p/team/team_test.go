package team

import (
	"context"
	"testing"
)

func TestTeam_AddAndGetMember(t *testing.T) {
	tm := NewTeam("t1", "test-team", "solve problem", "did:leader", 5)

	m := &Member{DID: "did:1", Name: "worker-1", Role: RoleWorker}
	if err := tm.AddMember(m); err != nil {
		t.Fatalf("AddMember() error = %v", err)
	}

	got := tm.GetMember("did:1")
	if got == nil {
		t.Fatal("GetMember() returned nil")
	}
	if got.Name != "worker-1" {
		t.Errorf("Name = %q, want %q", got.Name, "worker-1")
	}
	if got.JoinedAt.IsZero() {
		t.Error("JoinedAt should be set")
	}
}

func TestTeam_AddDuplicate(t *testing.T) {
	tm := NewTeam("t1", "test-team", "goal", "did:leader", 5)
	m := &Member{DID: "did:1", Name: "worker-1"}

	_ = tm.AddMember(m)
	err := tm.AddMember(m)
	if err != ErrAlreadyMember {
		t.Errorf("AddMember duplicate: got %v, want ErrAlreadyMember", err)
	}
}

func TestTeam_MaxCapacity(t *testing.T) {
	tm := NewTeam("t1", "test-team", "goal", "did:leader", 1)
	_ = tm.AddMember(&Member{DID: "did:1"})

	err := tm.AddMember(&Member{DID: "did:2"})
	if err != ErrTeamFull {
		t.Errorf("AddMember over capacity: got %v, want ErrTeamFull", err)
	}
}

func TestTeam_AddToDisbanded(t *testing.T) {
	tm := NewTeam("t1", "test-team", "goal", "did:leader", 5)
	tm.Disband()

	err := tm.AddMember(&Member{DID: "did:1"})
	if err != ErrTeamDisbanded {
		t.Errorf("AddMember to disbanded: got %v, want ErrTeamDisbanded", err)
	}
}

func TestTeam_RemoveMember(t *testing.T) {
	tm := NewTeam("t1", "test-team", "goal", "did:leader", 5)
	_ = tm.AddMember(&Member{DID: "did:1"})

	if err := tm.RemoveMember("did:1"); err != nil {
		t.Fatalf("RemoveMember() error = %v", err)
	}
	if tm.MemberCount() != 0 {
		t.Errorf("MemberCount() = %d, want 0", tm.MemberCount())
	}
}

func TestTeam_RemoveNotMember(t *testing.T) {
	tm := NewTeam("t1", "test-team", "goal", "did:leader", 5)

	err := tm.RemoveMember("did:nonexistent")
	if err != ErrNotMember {
		t.Errorf("RemoveMember nonexistent: got %v, want ErrNotMember", err)
	}
}

func TestTeam_Lifecycle(t *testing.T) {
	tm := NewTeam("t1", "test-team", "goal", "did:leader", 5)

	if tm.Status != StatusForming {
		t.Errorf("initial Status = %q, want %q", tm.Status, StatusForming)
	}

	tm.Activate()
	if tm.Status != StatusActive {
		t.Errorf("after Activate: Status = %q, want %q", tm.Status, StatusActive)
	}

	tm.Disband()
	if tm.Status != StatusDisbanded {
		t.Errorf("after Disband: Status = %q, want %q", tm.Status, StatusDisbanded)
	}
	if tm.DisbandedAt.IsZero() {
		t.Error("DisbandedAt should be set after Disband")
	}
}

func TestScopedContext_Roundtrip(t *testing.T) {
	ctx := context.Background()
	sc := ScopedContext{TeamID: "t1", MemberDID: "did:1", Role: RoleWorker}

	ctx = WithScopedContext(ctx, sc)
	got, ok := ScopedContextFromContext(ctx)
	if !ok {
		t.Fatal("ScopedContextFromContext returned false")
	}
	if got.TeamID != "t1" {
		t.Errorf("TeamID = %q, want %q", got.TeamID, "t1")
	}
	if got.MemberDID != "did:1" {
		t.Errorf("MemberDID = %q, want %q", got.MemberDID, "did:1")
	}
	if got.Role != RoleWorker {
		t.Errorf("Role = %q, want %q", got.Role, RoleWorker)
	}
}

func TestScopedContext_Missing(t *testing.T) {
	_, ok := ScopedContextFromContext(context.Background())
	if ok {
		t.Error("ScopedContextFromContext(empty) should return false")
	}
}

func TestTeam_ActiveMembers(t *testing.T) {
	tm := NewTeam("t1", "test-team", "goal", "did:leader", 10)
	_ = tm.AddMember(&Member{DID: "did:1", Status: MemberIdle})
	_ = tm.AddMember(&Member{DID: "did:2", Status: MemberBusy})
	_ = tm.AddMember(&Member{DID: "did:3", Status: MemberLeft})
	_ = tm.AddMember(&Member{DID: "did:4", Status: MemberFailed})

	active := tm.ActiveMembers()
	if len(active) != 2 {
		t.Errorf("ActiveMembers() = %d, want 2 (idle + busy)", len(active))
	}
}

func TestTeam_Budget(t *testing.T) {
	tm := NewTeam("t1", "test-team", "goal", "did:leader", 5)
	tm.Budget = 10.0

	if err := tm.AddSpend(5.0); err != nil {
		t.Fatalf("AddSpend(5.0) error = %v", err)
	}
	if tm.Spent != 5.0 {
		t.Errorf("Spent = %f, want 5.0", tm.Spent)
	}

	err := tm.AddSpend(6.0)
	if err == nil {
		t.Error("AddSpend(6.0) should fail when exceeding budget")
	}
}

func TestContextFilter(t *testing.T) {
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
			result := tt.filter.Filter(tt.metadata)
			for _, k := range tt.wantKeys {
				if _, ok := result[k]; !ok {
					t.Errorf("expected key %q in result", k)
				}
			}
			for _, k := range tt.wantMissing {
				if _, ok := result[k]; ok {
					t.Errorf("expected key %q to be filtered out", k)
				}
			}
		})
	}
}

func TestContextFilter_NilMetadata(t *testing.T) {
	f := ContextFilter{AllowedKeys: []string{"name"}}
	result := f.Filter(nil)
	if result != nil {
		t.Errorf("Filter(nil) = %v, want nil", result)
	}
}
