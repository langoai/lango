package approval

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestGrantStore_TTLExpired(t *testing.T) {
	now := time.Now()
	gs := NewGrantStore()
	gs.nowFn = func() time.Time { return now }
	gs.SetTTL(10 * time.Minute)

	gs.Grant("session-1", "echo")

	// Still valid within TTL.
	gs.nowFn = func() time.Time { return now.Add(9 * time.Minute) }
	if !gs.IsGranted("session-1", "echo") {
		t.Error("expected grant to be valid within TTL")
	}

	// Expired after TTL.
	gs.nowFn = func() time.Time { return now.Add(11 * time.Minute) }
	if gs.IsGranted("session-1", "echo") {
		t.Error("expected grant to be expired after TTL")
	}
}

func TestGrantStore_TTLZeroMeansNoExpiry(t *testing.T) {
	now := time.Now()
	gs := NewGrantStore()
	gs.nowFn = func() time.Time { return now }
	// TTL = 0 (default).

	gs.Grant("session-1", "echo")

	// 100 hours later, still valid.
	gs.nowFn = func() time.Time { return now.Add(100 * time.Hour) }
	if !gs.IsGranted("session-1", "echo") {
		t.Error("expected grant to be valid indefinitely when TTL = 0")
	}
}

func TestGrantStore_CleanExpired(t *testing.T) {
	now := time.Now()
	gs := NewGrantStore()
	gs.nowFn = func() time.Time { return now }
	gs.SetTTL(5 * time.Minute)

	gs.Grant("session-1", "echo")
	gs.Grant("session-1", "exec")
	gs.Grant("session-2", "echo")

	// Advance time past TTL for the first two, but grant session-2:echo later.
	gs.nowFn = func() time.Time { return now.Add(3 * time.Minute) }
	gs.Grant("session-2", "echo") // refresh

	gs.nowFn = func() time.Time { return now.Add(6 * time.Minute) }
	removed := gs.CleanExpired()
	if removed != 2 {
		t.Errorf("expected 2 expired grants removed, got %d", removed)
	}

	if gs.IsGranted("session-1", "echo") {
		t.Error("session-1:echo should be cleaned")
	}
	if gs.IsGranted("session-1", "exec") {
		t.Error("session-1:exec should be cleaned")
	}
	if !gs.IsGranted("session-2", "echo") {
		t.Error("session-2:echo should still be valid (refreshed)")
	}
}

func TestGrantStore_CleanExpiredNoOpWhenTTLZero(t *testing.T) {
	gs := NewGrantStore()
	gs.Grant("session-1", "echo")

	removed := gs.CleanExpired()
	if removed != 0 {
		t.Errorf("expected 0 removed with TTL=0, got %d", removed)
	}
	if !gs.IsGranted("session-1", "echo") {
		t.Error("grant should remain when TTL=0")
	}
}

func TestGrantLazyCleanup(t *testing.T) {
	now := time.Now()
	gs := NewGrantStore()
	gs.nowFn = func() time.Time { return now }
	gs.SetTTL(10 * time.Minute)

	// Add grants that will be expired by the time we call List().
	gs.Grant("session-1", "old_tool_1")
	gs.Grant("session-1", "old_tool_2")
	gs.Grant("session-2", "old_tool_3")

	// Add a grant that will still be valid.
	gs.nowFn = func() time.Time { return now.Add(8 * time.Minute) }
	gs.Grant("session-3", "fresh_tool")

	// Advance clock past TTL for the first 3 grants.
	gs.nowFn = func() time.Time { return now.Add(11 * time.Minute) }

	// Before List(), map should have all 4 entries.
	gs.mu.RLock()
	beforeCount := len(gs.grants)
	gs.mu.RUnlock()
	assert.Equal(t, 4, beforeCount, "should have 4 grants before cleanup")

	// Call List() — this should lazily clean expired grants.
	result := gs.List()
	assert.Len(t, result, 1, "only 1 non-expired grant should be returned")
	assert.Equal(t, "session-3", result[0].SessionKey)
	assert.Equal(t, "fresh_tool", result[0].ToolName)

	// After List(), expired grants should be physically removed from the map.
	gs.mu.RLock()
	afterCount := len(gs.grants)
	gs.mu.RUnlock()
	assert.Equal(t, 1, afterCount, "expired grants should be removed from map after List()")
}

func TestGrantLazyCleanup_NoTTL(t *testing.T) {
	now := time.Now()
	gs := NewGrantStore()
	gs.nowFn = func() time.Time { return now }

	// TTL is zero — no expiry.
	gs.Grant("session-1", "tool_1")
	gs.Grant("session-2", "tool_2")

	// Advance clock significantly.
	gs.nowFn = func() time.Time { return now.Add(999 * time.Hour) }

	result := gs.List()
	assert.Len(t, result, 2, "all grants should remain when TTL is zero")
}

func TestGrantStore_List(t *testing.T) {
	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		give     string
		setup    func(gs *GrantStore)
		wantLen  int
		wantList []GrantInfo
	}{
		{
			give:     "empty store returns empty slice",
			setup:    func(gs *GrantStore) {},
			wantLen:  0,
			wantList: nil,
		},
		{
			give: "returns all active grants sorted",
			setup: func(gs *GrantStore) {
				gs.Grant("session-2", "exec")
				gs.Grant("session-1", "fs_write")
				gs.Grant("session-1", "echo")
			},
			wantLen: 3,
			wantList: []GrantInfo{
				{SessionKey: "session-1", ToolName: "echo", GrantedAt: now},
				{SessionKey: "session-1", ToolName: "fs_write", GrantedAt: now},
				{SessionKey: "session-2", ToolName: "exec", GrantedAt: now},
			},
		},
		{
			give: "excludes expired grants",
			setup: func(gs *GrantStore) {
				gs.SetTTL(10 * time.Minute)
				// Grant at now (will be 15 min old when listed).
				gs.Grant("session-1", "old_tool")
				// Advance clock and grant a fresh one.
				gs.nowFn = func() time.Time { return now.Add(15 * time.Minute) }
				gs.Grant("session-1", "new_tool")
			},
			wantLen: 1,
			wantList: []GrantInfo{
				{SessionKey: "session-1", ToolName: "new_tool", GrantedAt: now.Add(15 * time.Minute)},
			},
		},
		{
			give: "excludes revoked grants",
			setup: func(gs *GrantStore) {
				gs.Grant("session-1", "echo")
				gs.Grant("session-1", "exec")
				gs.Revoke("session-1", "echo")
			},
			wantLen: 1,
			wantList: []GrantInfo{
				{SessionKey: "session-1", ToolName: "exec", GrantedAt: now},
			},
		},
		{
			give: "sort order: session key then tool name",
			setup: func(gs *GrantStore) {
				gs.Grant("z-session", "alpha")
				gs.Grant("a-session", "zulu")
				gs.Grant("a-session", "alpha")
				gs.Grant("z-session", "zulu")
			},
			wantLen: 4,
			wantList: []GrantInfo{
				{SessionKey: "a-session", ToolName: "alpha", GrantedAt: now},
				{SessionKey: "a-session", ToolName: "zulu", GrantedAt: now},
				{SessionKey: "z-session", ToolName: "alpha", GrantedAt: now},
				{SessionKey: "z-session", ToolName: "zulu", GrantedAt: now},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			gs := NewGrantStore()
			gs.nowFn = func() time.Time { return now }
			tt.setup(gs)

			got := gs.List()
			require.Len(t, got, tt.wantLen)
			if tt.wantList != nil {
				assert.Equal(t, tt.wantList, got)
			}
		})
	}
}
