package skill

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"go.uber.org/zap"
)

func newTestFileStore(t *testing.T) *FileSkillStore {
	dir := t.TempDir()
	logger := zap.NewNop().Sugar()
	return NewFileSkillStore(filepath.Join(dir, "skills"), logger)
}

func TestFileSkillStore_SaveAndGet(t *testing.T) {
	store := newTestFileStore(t)
	ctx := context.Background()

	entry := SkillEntry{
		Name:        "test-skill",
		Description: "A test skill",
		Type:        "script",
		Status:      "active",
		Definition:  map[string]interface{}{"script": "echo hello"},
	}

	if err := store.Save(ctx, entry); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := store.Get(ctx, "test-skill")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if got.Name != "test-skill" {
		t.Errorf("Name = %q, want %q", got.Name, "test-skill")
	}
	if got.Description != "A test skill" {
		t.Errorf("Description = %q, want %q", got.Description, "A test skill")
	}
	if got.Status != "active" {
		t.Errorf("Status = %q, want %q", got.Status, "active")
	}

	script, _ := got.Definition["script"].(string)
	if script != "echo hello" {
		t.Errorf("script = %q, want %q", script, "echo hello")
	}
}

func TestFileSkillStore_SaveEmptyName(t *testing.T) {
	store := newTestFileStore(t)
	ctx := context.Background()

	err := store.Save(ctx, SkillEntry{Name: ""})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestFileSkillStore_GetNotFound(t *testing.T) {
	store := newTestFileStore(t)
	ctx := context.Background()

	_, err := store.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent skill")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want to contain 'not found'", err.Error())
	}
}

func TestFileSkillStore_ListActive(t *testing.T) {
	store := newTestFileStore(t)
	ctx := context.Background()

	// Save active and draft skills.
	if err := store.Save(ctx, SkillEntry{
		Name:        "active-skill",
		Description: "active",
		Type:        "script",
		Status:      "active",
		Definition:  map[string]interface{}{"script": "echo active"},
	}); err != nil {
		t.Fatalf("Save active: %v", err)
	}

	if err := store.Save(ctx, SkillEntry{
		Name:        "draft-skill",
		Description: "draft",
		Type:        "script",
		Status:      "draft",
		Definition:  map[string]interface{}{"script": "echo draft"},
	}); err != nil {
		t.Fatalf("Save draft: %v", err)
	}

	entries, err := store.ListActive(ctx)
	if err != nil {
		t.Fatalf("ListActive: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	if entries[0].Name != "active-skill" {
		t.Errorf("entries[0].Name = %q, want %q", entries[0].Name, "active-skill")
	}
}

func TestFileSkillStore_ListActive_EmptyDir(t *testing.T) {
	store := newTestFileStore(t)
	ctx := context.Background()

	entries, err := store.ListActive(ctx)
	if err != nil {
		t.Fatalf("ListActive: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("len(entries) = %d, want 0", len(entries))
	}
}

func TestFileSkillStore_Activate(t *testing.T) {
	store := newTestFileStore(t)
	ctx := context.Background()

	if err := store.Save(ctx, SkillEntry{
		Name:        "my-skill",
		Description: "test",
		Type:        "script",
		Status:      "draft",
		Definition:  map[string]interface{}{"script": "echo hi"},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify it's not active.
	entries, _ := store.ListActive(ctx)
	if len(entries) != 0 {
		t.Fatalf("ListActive before activate: len = %d, want 0", len(entries))
	}

	if err := store.Activate(ctx, "my-skill"); err != nil {
		t.Fatalf("Activate: %v", err)
	}

	entries, _ = store.ListActive(ctx)
	if len(entries) != 1 {
		t.Fatalf("ListActive after activate: len = %d, want 1", len(entries))
	}
}

func TestFileSkillStore_Delete(t *testing.T) {
	store := newTestFileStore(t)
	ctx := context.Background()

	if err := store.Save(ctx, SkillEntry{
		Name:        "deleteme",
		Description: "test",
		Type:        "script",
		Status:      "active",
		Definition:  map[string]interface{}{"script": "echo hi"},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := store.Delete(ctx, "deleteme"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := store.Get(ctx, "deleteme")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestFileSkillStore_DeleteNotFound(t *testing.T) {
	store := newTestFileStore(t)
	ctx := context.Background()

	err := store.Delete(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent delete")
	}
}

func TestFileSkillStore_EnsureDefaults(t *testing.T) {
	store := newTestFileStore(t)

	// Create an in-memory FS with a default skill.
	defaultFS := fstest.MapFS{
		"serve/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: serve\ndescription: Start server\ntype: script\nstatus: active\n---\n\n```sh\nlango serve\n```\n"),
		},
		"version/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: version\ndescription: Show version\ntype: script\nstatus: active\n---\n\n```sh\nlango version\n```\n"),
		},
	}

	if err := store.EnsureDefaults(defaultFS); err != nil {
		t.Fatalf("EnsureDefaults: %v", err)
	}

	// Verify skills were deployed.
	ctx := context.Background()
	entries, err := store.ListActive(ctx)
	if err != nil {
		t.Fatalf("ListActive: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}

	// Run again — should not overwrite.
	// First, modify one to verify it's not replaced.
	customPath := filepath.Join(store.dir, "serve", "SKILL.md")
	if err := os.WriteFile(customPath, []byte("---\nname: serve\ndescription: Custom\ntype: script\nstatus: active\n---\n\n```sh\nlango serve --custom\n```\n"), 0o644); err != nil {
		t.Fatalf("write custom: %v", err)
	}

	if err := store.EnsureDefaults(defaultFS); err != nil {
		t.Fatalf("EnsureDefaults (second run): %v", err)
	}

	got, _ := store.Get(ctx, "serve")
	if got.Description != "Custom" {
		t.Errorf("Description = %q, want %q (should not be overwritten)", got.Description, "Custom")
	}
}
