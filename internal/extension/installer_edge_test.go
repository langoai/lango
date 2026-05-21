package extension

import (
	"context"
	"errors"
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

type errSource struct {
	err error
}

func (s errSource) Fetch(context.Context) (*WorkingCopy, error) {
	return nil, s.err
}

func TestInstaller_Inspect_ReturnsFetchError(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	wantErr := errors.New("fetch failed")

	report, wc, err := inst.Inspect(context.Background(), errSource{err: wantErr})

	require.ErrorIs(t, err, wantErr)
	assert.Nil(t, report)
	assert.Nil(t, wc)
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

func TestReplaceInstallDirsInstallsWhenNoExistingDirs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	finalDir := filepath.Join(root, "extensions", "fake-pack")
	extSkillDir := filepath.Join(root, "skills", "ext-fake-pack")
	packStaging := filepath.Join(root, "extensions", ".fake-pack.staging")
	skillStaging := filepath.Join(root, "skills", ".fake-pack.skills-staging")
	require.NoError(t, os.MkdirAll(packStaging, 0o755))
	require.NoError(t, os.MkdirAll(skillStaging, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(packStaging, manifestFileName), []byte("new manifest"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(skillStaging, "SKILL.md"), []byte("new skill"), 0o644))

	require.NoError(t, replaceInstallDirs(packStaging, finalDir, skillStaging, extSkillDir))

	assert.FileExists(t, filepath.Join(finalDir, manifestFileName))
	assert.FileExists(t, filepath.Join(extSkillDir, "SKILL.md"))
	assert.NoDirExists(t, packStaging)
	assert.NoDirExists(t, skillStaging)
}

func TestReplaceInstallDirsRestoresPackWhenSkillBackupFails(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	finalDir := filepath.Join(root, "extensions", "fake-pack")
	require.NoError(t, os.MkdirAll(finalDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(finalDir, manifestFileName), []byte("old manifest"), 0o644))

	packStaging := filepath.Join(root, "extensions", ".fake-pack.staging")
	skillStaging := filepath.Join(root, "skills", ".fake-pack.skills-staging")
	require.NoError(t, os.MkdirAll(packStaging, 0o755))
	require.NoError(t, os.MkdirAll(skillStaging, 0o755))

	err := replaceInstallDirs(packStaging, finalDir, skillStaging, filepath.Join(root, "skills", "bad\x00name"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "backup existing skills")

	preservedManifest, readErr := os.ReadFile(filepath.Join(finalDir, manifestFileName))
	require.NoError(t, readErr)
	assert.Equal(t, "old manifest", string(preservedManifest))
	assert.DirExists(t, packStaging)
}

func TestMoveToBackupBranches(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	missingBackup, had, err := moveToBackup(filepath.Join(root, "missing"))
	require.NoError(t, err)
	assert.False(t, had)
	assert.Empty(t, missingBackup)

	dir := filepath.Join(root, "installed")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "marker.txt"), []byte("old"), 0o644))
	backup, had, err := moveToBackup(dir)
	require.NoError(t, err)
	assert.True(t, had)
	assert.NotEmpty(t, backup)
	assert.NoDirExists(t, dir)
	assert.FileExists(t, filepath.Join(backup, "marker.txt"))

	badBackup, had, err := moveToBackup(filepath.Join(root, "bad\x00path"))
	require.Error(t, err)
	assert.False(t, had)
	assert.Empty(t, badBackup)
}

func TestTempBackupPathReturnsMkdirTempError(t *testing.T) {
	t.Parallel()

	parentFile := filepath.Join(t.TempDir(), "not-a-dir")
	require.NoError(t, os.WriteFile(parentFile, []byte("file"), 0o644))

	backup, err := tempBackupPath(parentFile, ".backup-")

	require.Error(t, err)
	assert.Empty(t, backup)
}

func TestRestoreBackupBranches(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "target")
	require.NoError(t, os.MkdirAll(target, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(target, "current.txt"), []byte("current"), 0o644))

	restoreBackup("", target)
	assert.FileExists(t, filepath.Join(target, "current.txt"))

	backup := filepath.Join(root, "backup")
	require.NoError(t, os.MkdirAll(backup, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(backup, "restored.txt"), []byte("restored"), 0o644))

	restoreBackup(backup, target)

	assert.NoDirExists(t, backup)
	assert.NoFileExists(t, filepath.Join(target, "current.txt"))
	assert.FileExists(t, filepath.Join(target, "restored.txt"))
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

func TestInstaller_Remove_ReturnsInstalledMetadataRemoveError(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	packDir := filepath.Join(inst.ExtensionsDir, "bad-meta")
	require.NoError(t, os.MkdirAll(filepath.Join(packDir, installedFileName), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(packDir, installedFileName, "child.txt"), []byte("blocks remove"), 0o644))

	err := inst.Remove(context.Background(), "bad-meta")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "remove")
	assert.Contains(t, err.Error(), installedFileName)
	assert.DirExists(t, packDir)
}

func TestInstaller_Remove_ReturnsStatError(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)

	err := inst.Remove(context.Background(), "bad\x00name")

	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrPackNotFound)
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

func TestCopyFileReturnsLstatError(t *testing.T) {
	t.Parallel()

	dst := filepath.Join(t.TempDir(), "out.txt")

	err := copyFile(filepath.Join(t.TempDir(), "missing.txt"), dst)

	require.Error(t, err)
	assert.NoFileExists(t, dst)
}

func TestCopyFileReturnsDestinationDirectoryError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	require.NoError(t, os.WriteFile(src, []byte("content"), 0o644))
	parentFile := filepath.Join(dir, "not-a-dir")
	require.NoError(t, os.WriteFile(parentFile, []byte("file"), 0o644))

	err := copyFile(src, filepath.Join(parentFile, "out.txt"))

	require.Error(t, err)
}

func TestCopyTreeReturnsRootAndSourceResolveErrors(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	src := filepath.Join(root, "skills", "foo")
	require.NoError(t, os.MkdirAll(src, 0o755))

	err := copyTree(src, t.TempDir(), filepath.Join(root, "missing-root"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolve root dir")

	err = copyTree(filepath.Join(root, "missing-src"), t.TempDir(), root)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolve src dir")
}

func TestCopyTreeReturnsDestinationDirectoryError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	src := filepath.Join(root, "skills", "foo")
	require.NoError(t, os.MkdirAll(src, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("skill"), 0o644))
	dstFile := filepath.Join(t.TempDir(), "not-a-dir")
	require.NoError(t, os.WriteFile(dstFile, []byte("file"), 0o644))

	err := copyTree(src, dstFile, root)

	require.Error(t, err)
}

func TestPlannedWritesFallbackWhenRootMissing(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	manifest := &Manifest{
		Name: "fallback-pack",
		Contents: Contents{
			Skills:  []SkillRef{{Name: "fallback", Path: "skills/fallback/SKILL.md"}},
			Prompts: []PromptRef{{Path: "prompts/hello.md"}},
		},
	}

	got := inst.plannedWrites(manifest, "")

	packDir := filepath.Join(inst.ExtensionsDir, "fallback-pack")
	extSkillDir := filepath.Join(inst.SkillsDir, "ext-fallback-pack", "fallback")
	assert.Contains(t, got, filepath.Join(packDir, manifestFileName))
	assert.Contains(t, got, filepath.Join(packDir, installedFileName))
	assert.Contains(t, got, filepath.Join(packDir, "skills", "fallback", "SKILL.md"))
	assert.Contains(t, got, filepath.Join(extSkillDir, "SKILL.md"))
	assert.Contains(t, got, filepath.Join(packDir, "prompts", "hello.md"))
}

func TestCopyTreeReturnsDanglingSymlinkResolveError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	src := filepath.Join(root, "skills", "foo")
	require.NoError(t, os.MkdirAll(src, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("skill"), 0o644))
	require.NoError(t, os.Symlink(filepath.Join(root, "missing.md"), filepath.Join(src, "dangling.md")))

	err := copyTree(src, t.TempDir(), root)

	require.Error(t, err)
	assert.Contains(t, err.Error(), filepath.Join("skills", "foo", "dangling.md"))
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
