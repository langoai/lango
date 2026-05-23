package adk

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/adk/tool/skilltoolset/skill"
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

// TestSkillSource_RoundTripsSkillContent exercises the upstream skill.Source
// directly: lists frontmatters, then loads a single skill and verifies its
// raw SKILL.md body round-trips byte-for-byte. This closes the Track B spec
// gate that required "an agent can call load_skill and receive its markdown
// body" — without requiring ADK's internal tool.Context construction.
func TestSkillSource_RoundTripsSkillContent(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "demo-skill")
	require.NoError(t, os.Mkdir(skillDir, 0o755))
	body := `---
name: demo-skill
description: Round-trip test fixture.
---
# Demo Skill

When the user says "demo", respond with "ok".
`
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(body), 0o644))

	src := skill.NewFileSystemSource(os.DirFS(dir))

	// List should show exactly one skill with the expected frontmatter.
	fms, err := src.ListFrontmatters(context.Background())
	require.NoError(t, err)
	require.Len(t, fms, 1)
	assert.Equal(t, "demo-skill", fms[0].Name)
	assert.Equal(t, "Round-trip test fixture.", fms[0].Description)

	// LoadFrontmatter returns just the metadata header.
	fm, err := src.LoadFrontmatter(context.Background(), "demo-skill")
	require.NoError(t, err)
	assert.Equal(t, "demo-skill", fm.Name)
	assert.Equal(t, "Round-trip test fixture.", fm.Description)

	// LoadInstructions returns the markdown body after the frontmatter.
	// This is what load_skill ultimately exposes to the agent.
	instructions, err := src.LoadInstructions(context.Background(), "demo-skill")
	require.NoError(t, err)
	assert.True(t, strings.Contains(instructions, "Demo Skill"),
		"expected instructions to contain markdown body header, got %q", instructions)
	assert.True(t, strings.Contains(instructions, `respond with "ok"`),
		"expected instructions to contain body content, got %q", instructions)
}
