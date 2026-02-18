package approval

import (
	"sync"
	"testing"
)

func TestGrantStore_GrantAndIsGranted(t *testing.T) {
	gs := NewGrantStore()

	if gs.IsGranted("session-1", "exec") {
		t.Error("expected no grant before granting")
	}

	gs.Grant("session-1", "exec")

	if !gs.IsGranted("session-1", "exec") {
		t.Error("expected grant after granting")
	}
}

func TestGrantStore_GrantIsolation(t *testing.T) {
	gs := NewGrantStore()
	gs.Grant("session-1", "exec")

	tests := []struct {
		give       string
		giveTool   string
		wantResult bool
	}{
		{give: "session-1", giveTool: "exec", wantResult: true},
		{give: "session-1", giveTool: "fs_delete", wantResult: false},
		{give: "session-2", giveTool: "exec", wantResult: false},
	}

	for _, tt := range tests {
		t.Run(tt.give+":"+tt.giveTool, func(t *testing.T) {
			if got := gs.IsGranted(tt.give, tt.giveTool); got != tt.wantResult {
				t.Errorf("IsGranted(%q, %q) = %v, want %v", tt.give, tt.giveTool, got, tt.wantResult)
			}
		})
	}
}

func TestGrantStore_Revoke(t *testing.T) {
	gs := NewGrantStore()
	gs.Grant("session-1", "exec")
	gs.Grant("session-1", "fs_write")

	gs.Revoke("session-1", "exec")

	if gs.IsGranted("session-1", "exec") {
		t.Error("expected exec grant to be revoked")
	}
	if !gs.IsGranted("session-1", "fs_write") {
		t.Error("expected fs_write grant to remain")
	}
}

func TestGrantStore_RevokeSession(t *testing.T) {
	gs := NewGrantStore()
	gs.Grant("session-1", "exec")
	gs.Grant("session-1", "fs_write")
	gs.Grant("session-2", "exec")

	gs.RevokeSession("session-1")

	if gs.IsGranted("session-1", "exec") {
		t.Error("expected session-1 exec grant to be revoked")
	}
	if gs.IsGranted("session-1", "fs_write") {
		t.Error("expected session-1 fs_write grant to be revoked")
	}
	if !gs.IsGranted("session-2", "exec") {
		t.Error("expected session-2 exec grant to remain")
	}
}

func TestGrantStore_ConcurrentAccess(t *testing.T) {
	gs := NewGrantStore()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gs.Grant("session-1", "exec")
			gs.IsGranted("session-1", "exec")
			gs.Grant("session-2", "fs_write")
			gs.Revoke("session-2", "fs_write")
		}()
	}

	wg.Wait()

	if !gs.IsGranted("session-1", "exec") {
		t.Error("expected session-1 exec grant after concurrent access")
	}
}

func TestGrantStore_RevokeNonExistent(t *testing.T) {
	gs := NewGrantStore()

	// Should not panic
	gs.Revoke("nonexistent", "tool")
	gs.RevokeSession("nonexistent")
}
