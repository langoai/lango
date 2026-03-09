package skill

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/zap"
)

func newTestFileStore(t *testing.T) *FileSkillStore {
	dir := t.TempDir()
	logger := zap.NewNop().Sugar()
	return NewFileSkillStore(filepath.Join(dir, "skills"), logger)
}

func TestFileSkillStore_SaveAndGet(t *testing.T) {
	t.Parallel()

	store := newTestFileStore(t)
	ctx := context.Background()

	entry := SkillEntry{
		Name:        "test-skill",
		Description: "A test skill",
		Type:        "script",
		Status:      "active",
		Definition:  map[string]interface{}{"script": "echo hello"},
	}

	require.NoError(t, store.Save(ctx, entry))

	got, err := store.Get(ctx, "test-skill")
	require.NoError(t, err)

	assert.Equal(t, "test-skill", got.Name)
	assert.Equal(t, "A test skill", got.Description)
	assert.Equal(t, SkillStatusActive, got.Status)

	script, _ := got.Definition["script"].(string)
	assert.Equal(t, "echo hello", script)
}

func TestFileSkillStore_SaveEmptyName(t *testing.T) {
	t.Parallel()

	store := newTestFileStore(t)
	ctx := context.Background()

	err := store.Save(ctx, SkillEntry{Name: ""})
	require.Error(t, err)
}

func TestFileSkillStore_GetNotFound(t *testing.T) {
	t.Parallel()

	store := newTestFileStore(t)
	ctx := context.Background()

	_, err := store.Get(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFileSkillStore_ListActive(t *testing.T) {
	t.Parallel()

	store := newTestFileStore(t)
	ctx := context.Background()

	// Save active and draft skills.
	require.NoError(t, store.Save(ctx, SkillEntry{
		Name:        "active-skill",
		Description: "active",
		Type:        "script",
		Status:      "active",
		Definition:  map[string]interface{}{"script": "echo active"},
	}))

	require.NoError(t, store.Save(ctx, SkillEntry{
		Name:        "draft-skill",
		Description: "draft",
		Type:        "script",
		Status:      "draft",
		Definition:  map[string]interface{}{"script": "echo draft"},
	}))

	entries, err := store.ListActive(ctx)
	require.NoError(t, err)

	require.Len(t, entries, 1)
	assert.Equal(t, "active-skill", entries[0].Name)
}

func TestFileSkillStore_ListActive_EmptyDir(t *testing.T) {
	t.Parallel()

	store := newTestFileStore(t)
	ctx := context.Background()

	entries, err := store.ListActive(ctx)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestFileSkillStore_Activate(t *testing.T) {
	t.Parallel()

	store := newTestFileStore(t)
	ctx := context.Background()

	require.NoError(t, store.Save(ctx, SkillEntry{
		Name:        "my-skill",
		Description: "test",
		Type:        "script",
		Status:      "draft",
		Definition:  map[string]interface{}{"script": "echo hi"},
	}))

	// Verify it's not active.
	entries, _ := store.ListActive(ctx)
	require.Empty(t, entries)

	require.NoError(t, store.Activate(ctx, "my-skill"))

	entries, _ = store.ListActive(ctx)
	require.Len(t, entries, 1)
}

func TestFileSkillStore_Delete(t *testing.T) {
	t.Parallel()

	store := newTestFileStore(t)
	ctx := context.Background()

	require.NoError(t, store.Save(ctx, SkillEntry{
		Name:        "deleteme",
		Description: "test",
		Type:        "script",
		Status:      "active",
		Definition:  map[string]interface{}{"script": "echo hi"},
	}))

	require.NoError(t, store.Delete(ctx, "deleteme"))

	_, err := store.Get(ctx, "deleteme")
	require.Error(t, err)
}

func TestFileSkillStore_DeleteNotFound(t *testing.T) {
	t.Parallel()

	store := newTestFileStore(t)
	ctx := context.Background()

	err := store.Delete(ctx, "nonexistent")
	require.Error(t, err)
}

func TestFileSkillStore_SaveResource(t *testing.T) {
	t.Parallel()

	store := newTestFileStore(t)
	ctx := context.Background()

	// Ensure skill directory exists first.
	require.NoError(t, store.Save(ctx, SkillEntry{
		Name:       "my-skill",
		Type:       "instruction",
		Status:     "active",
		Definition: map[string]interface{}{"content": "test"},
	}))

	data := []byte("#!/bin/bash\necho hello")
	require.NoError(t, store.SaveResource(ctx, "my-skill", "scripts/setup.sh", data))

	// Verify the file was written.
	got, err := os.ReadFile(filepath.Join(store.dir, "my-skill", "scripts", "setup.sh"))
	require.NoError(t, err)
	assert.Equal(t, string(data), string(got))
}

func TestFileSkillStore_SaveResource_NestedDir(t *testing.T) {
	t.Parallel()

	store := newTestFileStore(t)
	ctx := context.Background()

	require.NoError(t, store.Save(ctx, SkillEntry{
		Name:       "nested-skill",
		Type:       "instruction",
		Status:     "active",
		Definition: map[string]interface{}{"content": "test"},
	}))

	data := []byte("reference content")
	require.NoError(t, store.SaveResource(ctx, "nested-skill", "references/deep/nested/doc.md", data))

	got, err := os.ReadFile(filepath.Join(store.dir, "nested-skill", "references", "deep", "nested", "doc.md"))
	require.NoError(t, err)
	assert.Equal(t, string(data), string(got))
}

func TestFileSkillStore_EnsureDefaults(t *testing.T) {
	t.Parallel()

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

	require.NoError(t, store.EnsureDefaults(defaultFS))

	// Verify skills were deployed.
	ctx := context.Background()
	entries, err := store.ListActive(ctx)
	require.NoError(t, err)
	require.Len(t, entries, 2)

	// Run again — should not overwrite.
	// First, modify one to verify it's not replaced.
	customPath := filepath.Join(store.dir, "serve", "SKILL.md")
	require.NoError(t, os.WriteFile(customPath, []byte("---\nname: serve\ndescription: Custom\ntype: script\nstatus: active\n---\n\n```sh\nlango serve --custom\n```\n"), 0o644))

	require.NoError(t, store.EnsureDefaults(defaultFS))

	got, _ := store.Get(ctx, "serve")
	assert.Equal(t, "Custom", got.Description)
}
