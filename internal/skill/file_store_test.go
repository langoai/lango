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

func TestFileSkillStore_DiscoverProjectSkills(t *testing.T) {
	t.Parallel()

	t.Run("discovers valid project skills", func(t *testing.T) {
		t.Parallel()

		store := newTestFileStore(t)
		ctx := context.Background()

		projectRoot := t.TempDir()
		skillsDir := filepath.Join(projectRoot, ".lango", "skills", "sample")
		require.NoError(t, os.MkdirAll(skillsDir, 0o700))

		skillMD := []byte("---\nname: sample\ndescription: A sample skill\ntype: instruction\nstatus: active\n---\n\nSample content here.\n")
		require.NoError(t, os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), skillMD, 0o644))

		entries, err := store.DiscoverProjectSkills(ctx, projectRoot)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, "sample", entries[0].Name)
		assert.Equal(t, "A sample skill", entries[0].Description)
		assert.Equal(t, SkillTypeInstruction, entries[0].Type)
	})

	t.Run("missing directory returns nil", func(t *testing.T) {
		t.Parallel()

		store := newTestFileStore(t)
		ctx := context.Background()

		projectRoot := filepath.Join(t.TempDir(), "nonexistent")
		entries, err := store.DiscoverProjectSkills(ctx, projectRoot)
		require.NoError(t, err)
		assert.Nil(t, entries)
	})

	t.Run("empty directory returns nil", func(t *testing.T) {
		t.Parallel()

		store := newTestFileStore(t)
		ctx := context.Background()

		projectRoot := t.TempDir()
		skillsDir := filepath.Join(projectRoot, ".lango", "skills")
		require.NoError(t, os.MkdirAll(skillsDir, 0o700))

		entries, err := store.DiscoverProjectSkills(ctx, projectRoot)
		require.NoError(t, err)
		assert.Nil(t, entries)
	})

	t.Run("skips dot-prefixed directories", func(t *testing.T) {
		t.Parallel()

		store := newTestFileStore(t)
		ctx := context.Background()

		projectRoot := t.TempDir()
		skillsDir := filepath.Join(projectRoot, ".lango", "skills")

		// Create a dot-prefixed directory with a valid SKILL.md.
		hiddenDir := filepath.Join(skillsDir, ".hidden")
		require.NoError(t, os.MkdirAll(hiddenDir, 0o700))
		skillMD := []byte("---\nname: hidden\ndescription: hidden\ntype: instruction\nstatus: active\n---\n\nHidden.\n")
		require.NoError(t, os.WriteFile(filepath.Join(hiddenDir, "SKILL.md"), skillMD, 0o644))

		// Create a valid skill too.
		visibleDir := filepath.Join(skillsDir, "visible")
		require.NoError(t, os.MkdirAll(visibleDir, 0o700))
		visibleMD := []byte("---\nname: visible\ndescription: visible\ntype: instruction\nstatus: active\n---\n\nVisible.\n")
		require.NoError(t, os.WriteFile(filepath.Join(visibleDir, "SKILL.md"), visibleMD, 0o644))

		entries, err := store.DiscoverProjectSkills(ctx, projectRoot)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, "visible", entries[0].Name)
	})

	t.Run("skips directories without SKILL.md", func(t *testing.T) {
		t.Parallel()

		store := newTestFileStore(t)
		ctx := context.Background()

		projectRoot := t.TempDir()
		skillsDir := filepath.Join(projectRoot, ".lango", "skills")

		// Create a directory without SKILL.md.
		emptySkillDir := filepath.Join(skillsDir, "incomplete")
		require.NoError(t, os.MkdirAll(emptySkillDir, 0o700))

		entries, err := store.DiscoverProjectSkills(ctx, projectRoot)
		require.NoError(t, err)
		assert.Nil(t, entries)
	})

	t.Run("skips skills with invalid frontmatter", func(t *testing.T) {
		t.Parallel()

		store := newTestFileStore(t)
		ctx := context.Background()

		projectRoot := t.TempDir()
		skillsDir := filepath.Join(projectRoot, ".lango", "skills", "bad")
		require.NoError(t, os.MkdirAll(skillsDir, 0o700))

		// Write a SKILL.md with missing required name field.
		badMD := []byte("---\ndescription: no name\ntype: instruction\n---\n\nContent.\n")
		require.NoError(t, os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), badMD, 0o644))

		entries, err := store.DiscoverProjectSkills(ctx, projectRoot)
		require.NoError(t, err)
		assert.Nil(t, entries)
	})

	t.Run("discovers multiple skills", func(t *testing.T) {
		t.Parallel()

		store := newTestFileStore(t)
		ctx := context.Background()

		projectRoot := t.TempDir()
		skillsDir := filepath.Join(projectRoot, ".lango", "skills")

		for _, name := range []string{"alpha", "beta"} {
			dir := filepath.Join(skillsDir, name)
			require.NoError(t, os.MkdirAll(dir, 0o700))
			md := []byte("---\nname: " + name + "\ndescription: " + name + " skill\ntype: instruction\nstatus: active\n---\n\nContent.\n")
			require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), md, 0o644))
		}

		entries, err := store.DiscoverProjectSkills(ctx, projectRoot)
		require.NoError(t, err)
		require.Len(t, entries, 2)
	})
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
