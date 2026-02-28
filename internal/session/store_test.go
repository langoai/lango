package session

import (
	"errors"
	"os"
	"testing"
	"time"
)

func newTestEntStore(t *testing.T, opts ...StoreOption) *EntStore {
	tmpFile, err := os.CreateTemp("", "sessions_test_*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	store, err := NewEntStore(tmpFile.Name(), opts...)
	if err != nil {
		t.Fatalf("NewEntStore: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	return store
}

func TestEntStore_CreateAndGet(t *testing.T) {
	store := newTestEntStore(t)

	session := &Session{
		Key:         "sess-1",
		AgentID:     "agent-1",
		ChannelType: "telegram",
		ChannelID:   "12345",
		Model:       "claude-sonnet-4-20250514",
		History:     []Message{},
		Metadata:    map[string]string{"user": "test"},
	}

	if err := store.Create(session); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := store.Get("sess-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Key != "sess-1" {
		t.Errorf("Key: want %q, got %q", "sess-1", got.Key)
	}
	if got.Model != "claude-sonnet-4-20250514" {
		t.Errorf("Model: want %q, got %q", "claude-sonnet-4-20250514", got.Model)
	}
	if got.ChannelType != "telegram" {
		t.Errorf("ChannelType: want %q, got %q", "telegram", got.ChannelType)
	}
}

func TestEntStore_Get_NotFound(t *testing.T) {
	store := newTestEntStore(t)

	_, err := store.Get("non-existent")
	if err == nil {
		t.Fatal("expected error for non-existent session")
	}
}

func TestEntStore_Update(t *testing.T) {
	store := newTestEntStore(t)

	session := &Session{
		Key:   "sess-update",
		Model: "gpt-4",
	}
	if err := store.Create(session); err != nil {
		t.Fatalf("Create: %v", err)
	}

	session.Model = "claude-3-opus"
	if err := store.Update(session); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := store.Get("sess-update")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Model != "claude-3-opus" {
		t.Errorf("Model after update: want %q, got %q", "claude-3-opus", got.Model)
	}
}

func TestEntStore_Update_NotFound(t *testing.T) {
	store := newTestEntStore(t)

	err := store.Update(&Session{Key: "ghost"})
	if err == nil {
		t.Fatal("expected error for updating non-existent session")
	}
}

func TestEntStore_Delete_Idempotent(t *testing.T) {
	store := newTestEntStore(t)

	session := &Session{Key: "sess-del"}
	if err := store.Create(session); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// First delete
	if err := store.Delete("sess-del"); err != nil {
		t.Fatalf("first Delete: %v", err)
	}

	// Second delete should return nil (already deleted)
	if err := store.Delete("sess-del"); err != nil {
		t.Fatalf("second Delete: expected nil, got %v", err)
	}
}

func TestEntStore_AppendMessage_WithToolCalls(t *testing.T) {
	store := newTestEntStore(t)

	session := &Session{Key: "sess-tc"}
	if err := store.Create(session); err != nil {
		t.Fatalf("Create: %v", err)
	}

	msg := Message{
		Role:    "assistant",
		Content: "Let me check that",
		ToolCalls: []ToolCall{
			{ID: "tc-1", Name: "exec", Input: `{"cmd":"ls"}`, Output: "file.txt"},
		},
	}
	if err := store.AppendMessage("sess-tc", msg); err != nil {
		t.Fatalf("AppendMessage: %v", err)
	}

	got, err := store.Get("sess-tc")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(got.History) != 1 {
		t.Fatalf("History length: want 1, got %d", len(got.History))
	}
	if len(got.History[0].ToolCalls) != 1 {
		t.Fatalf("ToolCalls length: want 1, got %d", len(got.History[0].ToolCalls))
	}
	tc := got.History[0].ToolCalls[0]
	if tc.ID != "tc-1" {
		t.Errorf("ToolCall ID: want %q, got %q", "tc-1", tc.ID)
	}
	if tc.Name != "exec" {
		t.Errorf("ToolCall Name: want %q, got %q", "exec", tc.Name)
	}
}

func TestEntStore_AppendMessage_NotFound(t *testing.T) {
	store := newTestEntStore(t)

	err := store.AppendMessage("ghost", Message{Role: "user", Content: "hi"})
	if err == nil {
		t.Fatal("expected error for appending to non-existent session")
	}
}

func TestEntStore_MaxHistoryTurns(t *testing.T) {
	store := newTestEntStore(t, WithMaxHistoryTurns(3))

	session := &Session{Key: "sess-max"}
	if err := store.Create(session); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Add 5 messages
	for i := 0; i < 5; i++ {
		msg := Message{
			Role:      "user",
			Content:   "message",
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
		}
		if err := store.AppendMessage("sess-max", msg); err != nil {
			t.Fatalf("AppendMessage %d: %v", i, err)
		}
	}

	got, err := store.Get("sess-max")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(got.History) != 3 {
		t.Errorf("History length: want 3, got %d", len(got.History))
	}
}

func TestEntStore_TTL(t *testing.T) {
	store := newTestEntStore(t, WithTTL(50*time.Millisecond))

	sess := &Session{Key: "sess-ttl"}
	if err := store.Create(sess); err != nil {
		t.Fatalf("Create: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	_, err := store.Get("sess-ttl")
	if err == nil {
		t.Fatal("expected session expired error")
	}
	if !errors.Is(err, ErrSessionExpired) {
		t.Errorf("expected ErrSessionExpired, got: %v", err)
	}
}

func TestEntStore_TTL_DeleteAndRecreate(t *testing.T) {
	store := newTestEntStore(t, WithTTL(50*time.Millisecond))

	sess := &Session{Key: "sess-ttl-renew", Model: "old-model"}
	if err := store.Create(sess); err != nil {
		t.Fatalf("Create: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify expired
	_, err := store.Get("sess-ttl-renew")
	if !errors.Is(err, ErrSessionExpired) {
		t.Fatalf("expected ErrSessionExpired, got: %v", err)
	}

	// Delete expired session
	if err := store.Delete("sess-ttl-renew"); err != nil {
		t.Fatalf("Delete expired: %v", err)
	}

	// Recreate with new data
	newSess := &Session{Key: "sess-ttl-renew", Model: "new-model"}
	if err := store.Create(newSess); err != nil {
		t.Fatalf("Recreate: %v", err)
	}

	got, err := store.Get("sess-ttl-renew")
	if err != nil {
		t.Fatalf("Get after recreate: %v", err)
	}
	if got.Model != "new-model" {
		t.Errorf("Model: want %q, got %q", "new-model", got.Model)
	}
}

func TestEntStore_GetSetSalt(t *testing.T) {
	store := newTestEntStore(t)

	salt := []byte("random-salt-data")
	if err := store.SetSalt("test-salt", salt); err != nil {
		t.Fatalf("SetSalt: %v", err)
	}

	got, err := store.GetSalt("test-salt")
	if err != nil {
		t.Fatalf("GetSalt: %v", err)
	}
	if string(got) != string(salt) {
		t.Errorf("salt: want %q, got %q", salt, got)
	}
}

func TestEntStore_GetSalt_NotFound(t *testing.T) {
	store := newTestEntStore(t)

	_, err := store.GetSalt("missing")
	if err == nil {
		t.Fatal("expected error for missing salt")
	}
}

func TestEntStore_SetGetChecksum(t *testing.T) {
	store := newTestEntStore(t)

	// Must set salt first
	if err := store.SetSalt("chk-test", []byte("salt")); err != nil {
		t.Fatalf("SetSalt: %v", err)
	}

	checksum := []byte("checksum-data")
	if err := store.SetChecksum("chk-test", checksum); err != nil {
		t.Fatalf("SetChecksum: %v", err)
	}

	got, err := store.GetChecksum("chk-test")
	if err != nil {
		t.Fatalf("GetChecksum: %v", err)
	}
	if string(got) != string(checksum) {
		t.Errorf("checksum: want %q, got %q", checksum, got)
	}
}

func TestEntStore_SetChecksum_NoSalt(t *testing.T) {
	store := newTestEntStore(t)

	err := store.SetChecksum("no-salt", []byte("checksum"))
	if err == nil {
		t.Fatal("expected error when setting checksum without salt")
	}
}
