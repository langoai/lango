package extension

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type staticSource struct{}

func (staticSource) Fetch(context.Context) (*WorkingCopy, error) {
	return nil, nil
}

func TestInstaller_Install_RollbackOnCopyPackFilesError(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	src := NewLocalSource(writeFakePack(t))
	_, wc, err := inst.Inspect(context.Background(), src)
	require.NoError(t, err)
	t.Cleanup(func() { _ = wc.Cleanup() })

	wc.Manifest.Contents.Prompts = []PromptRef{{Path: "prompts/missing.md"}}

	err = inst.Install(context.Background(), src, wc, InstallOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing.md")
	assert.NoFileExists(t, filepath.Join(inst.ExtensionsDir, "fake-pack", manifestFileName))
	assert.NoDirExists(t, filepath.Join(inst.SkillsDir, "ext-fake-pack"))
	assertNoStagingDirs(t, inst.ExtensionsDir)
}

func TestInstaller_Install_RollbackOnCopySkillsToStoreError(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	skillsFile := filepath.Join(t.TempDir(), "skills-file")
	require.NoError(t, os.WriteFile(skillsFile, []byte("not a directory"), 0o644))
	inst.SkillsDir = skillsFile

	src := NewLocalSource(writeFakePack(t))
	_, wc, err := inst.Inspect(context.Background(), src)
	require.NoError(t, err)
	t.Cleanup(func() { _ = wc.Cleanup() })

	err = inst.Install(context.Background(), src, wc, InstallOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create skills dir")
	assert.NoDirExists(t, filepath.Join(inst.ExtensionsDir, "fake-pack"))
	assertNoStagingDirs(t, inst.ExtensionsDir)
}

func TestInstaller_Install_RollbackOnRenameError(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	require.NoError(t, os.MkdirAll(inst.ExtensionsDir, 0o755))
	finalPath := filepath.Join(inst.ExtensionsDir, "fake-pack")
	require.NoError(t, os.WriteFile(finalPath, []byte("blocks directory rename"), 0o644))

	src := NewLocalSource(writeFakePack(t))
	_, wc, err := inst.Inspect(context.Background(), src)
	require.NoError(t, err)
	t.Cleanup(func() { _ = wc.Cleanup() })

	err = inst.Install(context.Background(), src, wc, InstallOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "move staging")
	assert.FileExists(t, finalPath)
	assert.NoDirExists(t, filepath.Join(inst.SkillsDir, "ext-fake-pack"))
	assertNoStagingDirs(t, inst.ExtensionsDir)
}

func TestInstaller_Install_AllowOverwriteReinstallsPack(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	packDir := writeFakePack(t)
	staleSourcePath := filepath.Join(packDir, "skills", "foo", "references", "old.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(staleSourcePath), 0o755))
	require.NoError(t, os.WriteFile(staleSourcePath, []byte("old resource"), 0o644))

	src := NewLocalSource(packDir)
	_, wc, err := inst.Inspect(context.Background(), src)
	require.NoError(t, err)
	require.NoError(t, inst.Install(context.Background(), src, wc, InstallOptions{}))

	updatedSkill := []byte("---\nname: foo\nstatus: active\n---\nupdated")
	require.NoError(t, os.WriteFile(
		filepath.Join(packDir, "skills", "foo", "SKILL.md"),
		updatedSkill,
		0o644,
	))
	require.NoError(t, os.Remove(staleSourcePath))
	_, updatedWC, err := inst.Inspect(context.Background(), src)
	require.NoError(t, err)

	require.NoError(t, inst.Install(
		context.Background(),
		src,
		updatedWC,
		InstallOptions{AllowOverwrite: true},
	))

	packCopy, err := os.ReadFile(filepath.Join(
		inst.ExtensionsDir,
		"fake-pack",
		"skills",
		"foo",
		"SKILL.md",
	))
	require.NoError(t, err)
	assert.Equal(t, updatedSkill, packCopy)

	storeCopy, err := os.ReadFile(filepath.Join(
		inst.SkillsDir,
		"ext-fake-pack",
		"foo",
		"SKILL.md",
	))
	require.NoError(t, err)
	assert.Equal(t, updatedSkill, storeCopy)

	assert.NoFileExists(t, filepath.Join(
		inst.ExtensionsDir,
		"fake-pack",
		"skills",
		"foo",
		"references",
		"old.md",
	))
	assert.NoFileExists(t, filepath.Join(
		inst.SkillsDir,
		"ext-fake-pack",
		"foo",
		"references",
		"old.md",
	))
}

func TestInstaller_Install_AllowOverwritePreservesExistingSkillsOnFailure(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	packDir := writeFakePack(t)
	src := NewLocalSource(packDir)
	_, wc, err := inst.Inspect(context.Background(), src)
	require.NoError(t, err)
	require.NoError(t, inst.Install(context.Background(), src, wc, InstallOptions{}))

	originalSkillPath := filepath.Join(inst.SkillsDir, "ext-fake-pack", "foo", "SKILL.md")
	originalSkill, err := os.ReadFile(originalSkillPath)
	require.NoError(t, err)

	_, badWC, err := inst.Inspect(context.Background(), src)
	require.NoError(t, err)
	badWC.Manifest.Contents.Skills[0].Path = "../escape.md"

	err = inst.Install(context.Background(), src, badWC, InstallOptions{AllowOverwrite: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `skill "foo"`)

	preservedSkill, err := os.ReadFile(originalSkillPath)
	require.NoError(t, err)
	assert.Equal(t, originalSkill, preservedSkill)
	assert.FileExists(t, filepath.Join(inst.ExtensionsDir, "fake-pack", manifestFileName))
}

func TestReplaceInstallDirsRestoresExistingInstallWhenSkillCommitFails(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	finalDir := filepath.Join(root, "extensions", "fake-pack")
	extSkillDir := filepath.Join(root, "skills", "ext-fake-pack")
	require.NoError(t, os.MkdirAll(finalDir, 0o755))
	require.NoError(t, os.MkdirAll(extSkillDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(finalDir, manifestFileName), []byte("old manifest"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(extSkillDir, "old.md"), []byte("old skill"), 0o644))

	packStaging := filepath.Join(root, "extensions", ".fake-pack.staging")
	require.NoError(t, os.MkdirAll(packStaging, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(packStaging, manifestFileName), []byte("new manifest"), 0o644))
	missingSkillStaging := filepath.Join(root, "skills", ".missing-staging")

	err := replaceInstallDirs(packStaging, finalDir, missingSkillStaging, extSkillDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "move skills staging")

	preservedManifest, err := os.ReadFile(filepath.Join(finalDir, manifestFileName))
	require.NoError(t, err)
	assert.Equal(t, "old manifest", string(preservedManifest))
	preservedSkill, err := os.ReadFile(filepath.Join(extSkillDir, "old.md"))
	require.NoError(t, err)
	assert.Equal(t, "old skill", string(preservedSkill))
}

func TestInstaller_Remove_MissingInstalledButExistingPack(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	packDir := filepath.Join(inst.ExtensionsDir, "half-removed")
	require.NoError(t, os.MkdirAll(packDir, 0o755))
	extSkillDir := filepath.Join(inst.SkillsDir, "ext-half-removed")
	require.NoError(t, os.MkdirAll(extSkillDir, 0o755))

	require.NoError(t, inst.Remove(context.Background(), "half-removed"))

	assert.NoDirExists(t, packDir)
	assert.NoDirExists(t, extSkillDir)
}

func TestCopyPackFiles_PathError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, manifestFileName), []byte("manifest"), 0o644))
	wc := &WorkingCopy{
		RootDir: root,
		Manifest: &Manifest{
			Contents: Contents{
				Prompts: []PromptRef{{Path: "../escape.md"}},
			},
		},
	}

	err := copyPackFiles(wc, t.TempDir())
	require.Error(t, err)
	assert.Contains(t, err.Error(), `path "../escape.md"`)
}

func TestCopySkillsToStore_PathError(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	src := NewLocalSource(writeFakePack(t))
	_, wc, err := inst.Inspect(context.Background(), src)
	require.NoError(t, err)
	t.Cleanup(func() { _ = wc.Cleanup() })
	wc.Manifest.Contents.Skills[0].Path = "../escape.md"

	err = inst.copySkillsToStore(wc, filepath.Join(inst.SkillsDir, "ext-fake-pack"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), `skill "foo"`)
	assert.Contains(t, err.Error(), "parent-directory")
}

func TestSourceDescription(t *testing.T) {
	t.Parallel()

	localDir := t.TempDir()
	localWant, err := filepath.Abs(localDir)
	require.NoError(t, err)

	assert.Equal(t, "local:"+localWant, sourceDescription(NewLocalSource(localDir)))
	assert.Equal(t, "git:https://example.com/pack.git", sourceDescription(
		NewGitSource("https://example.com/pack.git"),
	))
	assert.Empty(t, sourceDescription(staticSource{}))
}

func TestInstaller_Install_CrossExtModeCollision(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	dirA := writeModeOnlyPack(t, "mode-pack-a", "shared-mode")
	srcA := NewLocalSource(dirA)
	_, wcA, err := inst.Inspect(context.Background(), srcA)
	require.NoError(t, err)
	require.NoError(t, inst.Install(context.Background(), srcA, wcA, InstallOptions{}))

	dirB := writeModeOnlyPack(t, "mode-pack-b", "shared-mode")
	srcB := NewLocalSource(dirB)
	_, wcB, err := inst.Inspect(context.Background(), srcB)
	require.NoError(t, err)

	err = inst.Install(context.Background(), srcB, wcB, InstallOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `mode name "shared-mode"`)
	assert.Contains(t, err.Error(), "mode-pack-a")
}

func writeModeOnlyPack(t *testing.T, name, mode string) string {
	t.Helper()

	dir := t.TempDir()
	manifest := `schema: lango.extension/v1
name: ` + name + `
version: 0.1.0
description: Mode-only pack for tests
contents:
  modes:
    - name: ` + mode + `
      tools:
        - shell
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, manifestFileName), []byte(manifest), 0o644))
	return dir
}

func assertNoStagingDirs(t *testing.T, extensionsDir string) {
	t.Helper()

	entries, err := os.ReadDir(extensionsDir)
	if os.IsNotExist(err) {
		return
	}
	require.NoError(t, err)
	for _, entry := range entries {
		assert.False(t, entry.IsDir() && entry.Name()[0] == '.', "staging dir left behind: %s", entry.Name())
	}
}
