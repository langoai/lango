package adk

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSkillToolset_EmptyDirReturnsNilToolset(t *testing.T) {
	ts, err := BuildSkillToolset(context.Background(), "")
	require.NoError(t, err)
	assert.Nil(t, ts)
}

func TestBuildSkillToolset_NonexistentDirReturnsErr(t *testing.T) {
	ts, err := BuildSkillToolset(context.Background(), "/nonexistent/path/should/not/exist")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSkillsDirNotExist)
	assert.Nil(t, ts)
}

func TestBuildSkillToolset_FileInsteadOfDirReturnsErr(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "lango-skill-*.txt")
	require.NoError(t, err)
	tmpFile.Close()
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	ts, err := BuildSkillToolset(context.Background(), tmpFile.Name())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
	assert.Nil(t, ts)
}

func TestBuildSkillToolset_ValidSkillReturnsToolset(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "example-skill")
	require.NoError(t, os.Mkdir(skillDir, 0o755))
	skillMD := []byte(`---
name: example-skill
description: Example skill for unit test
---
# Example Skill

This skill is a unit-test fixture.
`)
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), skillMD, 0o644))

	ts, err := BuildSkillToolset(context.Background(), dir)
	require.NoError(t, err)
	require.NotNil(t, ts)
	assert.Equal(t, "SkillToolset", ts.Name())

	tools, err := ts.Tools(nil)
	require.NoError(t, err)
	// Upstream skilltoolset exposes load_skill, list_skill_resources,
	// load_skill_resource — three tools total.
	assert.Len(t, tools, 3)
}

func TestBuildSkillToolset_EmptySkillDirReturnsToolsetWithNoSkills(t *testing.T) {
	// An empty directory is valid (no skills present yet). Tools count is still 3.
	dir := t.TempDir()
	ts, err := BuildSkillToolset(context.Background(), dir)
	require.NoError(t, err)
	require.NotNil(t, ts)
	tools, err := ts.Tools(nil)
	require.NoError(t, err)
	assert.Len(t, tools, 3)
}
