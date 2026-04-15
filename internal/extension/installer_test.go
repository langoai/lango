package extension

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestInstaller(t *testing.T) *Installer {
	t.Helper()
	return &Installer{
		ExtensionsDir: filepath.Join(t.TempDir(), "extensions"),
		SkillsDir:     filepath.Join(t.TempDir(), "skills"),
	}
}

func TestInstaller_Inspect_HappyPath(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	src := NewLocalSource(writeFakePack(t))
	report, wc, err := inst.Inspect(context.Background(), src)
	require.NoError(t, err)
	t.Cleanup(func() { _ = wc.Cleanup() })

	require.NotNil(t, report.Manifest)
	assert.Equal(t, "fake-pack", report.Manifest.Name)
	assert.NotEmpty(t, report.ManifestSHA256)
	assert.NotEmpty(t, report.PlannedWrites)
	assert.Contains(t, report.SkippedWrites[0], "tools")

	// Inspect must not create anything under ExtensionsDir.
	_, err = os.Stat(inst.ExtensionsDir)
	assert.True(t, os.IsNotExist(err), "Inspect must be side-effect free")
}

func TestInstaller_Install_HappyPath(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	src := NewLocalSource(writeFakePack(t))
	_, wc, err := inst.Inspect(context.Background(), src)
	require.NoError(t, err)
	t.Cleanup(func() { _ = wc.Cleanup() })

	require.NoError(t, inst.Install(context.Background(), src, wc, InstallOptions{}))

	// Pack dir exists with manifest and .installed.
	packDir := filepath.Join(inst.ExtensionsDir, "fake-pack")
	assert.FileExists(t, filepath.Join(packDir, manifestFileName))
	assert.FileExists(t, filepath.Join(packDir, installedFileName))
	assert.FileExists(t, filepath.Join(packDir, "skills", "foo", "SKILL.md"))
	assert.FileExists(t, filepath.Join(packDir, "prompts", "hello.md"))

	// Skill copy landed under ext-<name>/ in skills dir.
	skillCopy := filepath.Join(inst.SkillsDir, "ext-fake-pack", "foo", "SKILL.md")
	assert.FileExists(t, skillCopy)
}

func TestInstaller_Install_DuplicateRejected(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	src := NewLocalSource(writeFakePack(t))

	_, wc1, err := inst.Inspect(context.Background(), src)
	require.NoError(t, err)
	require.NoError(t, inst.Install(context.Background(), src, wc1, InstallOptions{}))

	_, wc2, err := inst.Inspect(context.Background(), src)
	require.NoError(t, err)
	err = inst.Install(context.Background(), src, wc2, InstallOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already installed")
}

func TestInstaller_Install_CrossExtCollision(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	// Pack A declares skill "foo".
	dirA := writeFakePack(t)
	srcA := NewLocalSource(dirA)
	_, wcA, _ := inst.Inspect(context.Background(), srcA)
	require.NoError(t, inst.Install(context.Background(), srcA, wcA, InstallOptions{}))

	// Pack B reuses skill name "foo" under a different pack name.
	dirB := t.TempDir()
	manifest := `schema: lango.extension/v1
name: other-pack
version: 0.1.0
description: collides
contents:
  skills:
    - name: foo
      path: skills/foo/SKILL.md
`
	require.NoError(t, os.WriteFile(filepath.Join(dirB, manifestFileName), []byte(manifest), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(dirB, "skills", "foo"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dirB, "skills", "foo", "SKILL.md"),
		[]byte("---\nname: foo\nstatus: active\n---\nhi"), 0o644))

	srcB := NewLocalSource(dirB)
	_, wcB, err := inst.Inspect(context.Background(), srcB)
	require.NoError(t, err)
	err = inst.Install(context.Background(), srcB, wcB, InstallOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already owned by installed pack")
	assert.Contains(t, err.Error(), "fake-pack")
}

func TestInstaller_Remove_HappyPath(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	src := NewLocalSource(writeFakePack(t))
	_, wc, _ := inst.Inspect(context.Background(), src)
	require.NoError(t, inst.Install(context.Background(), src, wc, InstallOptions{}))

	require.NoError(t, inst.Remove(context.Background(), "fake-pack"))

	_, err := os.Stat(filepath.Join(inst.ExtensionsDir, "fake-pack"))
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(inst.SkillsDir, "ext-fake-pack"))
	assert.True(t, os.IsNotExist(err))
}

func TestInstaller_Remove_Unknown(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	err := inst.Remove(context.Background(), "missing")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPackNotFound)
}

func TestRegistry_LoadsInstalledPack(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	src := NewLocalSource(writeFakePack(t))
	_, wc, _ := inst.Inspect(context.Background(), src)
	require.NoError(t, inst.Install(context.Background(), src, wc, InstallOptions{}))

	reg, err := LoadRegistry(inst.ExtensionsDir, false)
	require.NoError(t, err)
	require.Len(t, reg.List(), 1)
	require.Len(t, reg.OKPacks(), 1)
	assert.Equal(t, StatusOK, reg.List()[0].Status)
}

func TestRegistry_TamperDetection(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	src := NewLocalSource(writeFakePack(t))
	_, wc, _ := inst.Inspect(context.Background(), src)
	require.NoError(t, inst.Install(context.Background(), src, wc, InstallOptions{}))

	// Tamper with a bundled skill file.
	tamperPath := filepath.Join(inst.ExtensionsDir, "fake-pack", "skills", "foo", "SKILL.md")
	require.NoError(t, os.WriteFile(tamperPath, []byte("tampered"), 0o644))

	// Default mode: logs warning, still loads.
	reg, err := LoadRegistry(inst.ExtensionsDir, false)
	require.NoError(t, err)
	require.Len(t, reg.List(), 1)
	assert.Equal(t, StatusTampered, reg.List()[0].Status)

	// Enforce mode: manifest stripped so OKPacks is empty.
	regEnforce, err := LoadRegistry(inst.ExtensionsDir, true)
	require.NoError(t, err)
	assert.Empty(t, regEnforce.OKPacks())
	assert.Len(t, regEnforce.List(), 1, "still visible in List() for CLI status reporting")
}

func TestRegistry_EmptyDirIsNoop(t *testing.T) {
	t.Parallel()

	reg, err := LoadRegistry(filepath.Join(t.TempDir(), "nope"), false)
	require.NoError(t, err)
	assert.Empty(t, reg.List())
}

func TestRegistry_BrokenPackSkipped(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pack := filepath.Join(dir, "broken")
	require.NoError(t, os.MkdirAll(pack, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pack, manifestFileName), []byte("not valid yaml {"), 0o644))

	reg, err := LoadRegistry(dir, false)
	require.NoError(t, err)
	require.Len(t, reg.List(), 1)
	assert.Equal(t, StatusBroken, reg.List()[0].Status)
	assert.Empty(t, reg.OKPacks())
}
