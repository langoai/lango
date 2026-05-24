package extension

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadRegistry_ReturnsReadDirErrorForNonDirectory(t *testing.T) {
	t.Parallel()

	extensionsFile := filepath.Join(t.TempDir(), "extensions")
	require.NoError(t, os.WriteFile(extensionsFile, []byte("not a directory"), 0o644))

	reg, err := LoadRegistry(extensionsFile, false)

	require.Error(t, err)
	assert.Nil(t, reg)
	assert.Contains(t, err.Error(), "read extensions dir")
}

func TestLoadRegistry_SkipsDotPrefixedAndNonDirectoryEntries(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	installPack(t, inst, writeRegistryPack(t, "visible-pack"))
	require.NoError(t, os.MkdirAll(filepath.Join(inst.ExtensionsDir, ".hidden-pack"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(inst.ExtensionsDir, ".hidden-pack", manifestFileName),
		[]byte("not valid yaml {"),
		0o644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(inst.ExtensionsDir, "top-level-file"),
		[]byte("not a pack directory"),
		0o644,
	))

	reg, err := LoadRegistry(inst.ExtensionsDir, false)

	require.NoError(t, err)
	require.Len(t, reg.List(), 1)
	_, ok := reg.Lookup("visible-pack")
	assert.True(t, ok)
	_, ok = reg.Lookup("hidden-pack")
	assert.False(t, ok)
}

func TestLoadRegistry_InvalidInstalledMetaMarksPackBroken(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	installPack(t, inst, writeFakePack(t))
	packDir := filepath.Join(inst.ExtensionsDir, "fake-pack")
	require.NoError(t, os.WriteFile(filepath.Join(packDir, installedFileName), []byte("{"), 0o644))

	reg, err := LoadRegistry(inst.ExtensionsDir, false)

	require.NoError(t, err)
	require.Len(t, reg.List(), 1)
	pack := reg.List()[0]
	assert.Equal(t, StatusBroken, pack.Status)
	assert.NotNil(t, pack.Manifest)
	assert.Nil(t, pack.Meta)
	assert.Empty(t, reg.OKPacks())
	requireWarningContains(t, pack.Warnings, "parse .installed")
}

func TestLoadRegistry_MissingHashedFileMarksPackTampered(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	installPack(t, inst, writeFakePack(t))
	missingPrompt := filepath.Join(inst.ExtensionsDir, "fake-pack", "prompts", "hello.md")
	require.NoError(t, os.Remove(missingPrompt))

	reg, err := LoadRegistry(inst.ExtensionsDir, false)

	require.NoError(t, err)
	require.Len(t, reg.List(), 1)
	pack := reg.List()[0]
	assert.Equal(t, StatusTampered, pack.Status)
	assert.Empty(t, reg.OKPacks())
	requireWarningContains(t, pack.Warnings, `missing or unreadable "prompts/hello.md"`)

	enforced, err := LoadRegistry(inst.ExtensionsDir, true)
	require.NoError(t, err)
	require.Len(t, enforced.List(), 1)
	assert.Nil(t, enforced.List()[0].Manifest)
	assert.Empty(t, enforced.Modes())
	assert.Empty(t, enforced.PromptSources())
}

func TestLoadRegistry_UnsafeFileHashPathMarksPackTampered(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	installPack(t, inst, writeFakePack(t))
	packDir := filepath.Join(inst.ExtensionsDir, "fake-pack")
	meta := readInstalledMeta(t, packDir)
	meta.FileHashes["../outside.md"] = strings.Repeat("0", 64)
	require.NoError(t, WriteInstalledMeta(packDir, meta))

	reg, err := LoadRegistry(inst.ExtensionsDir, false)

	require.NoError(t, err)
	require.Len(t, reg.List(), 1)
	pack := reg.List()[0]
	assert.Equal(t, StatusTampered, pack.Status)
	assert.NotNil(t, pack.Manifest)
	assert.Empty(t, reg.OKPacks())
	requireWarningContains(t, pack.Warnings, `path-safety failed for "../outside.md"`)
}

func TestRegistry_LookupModesAndPromptSourcesUseOnlyHealthyPacks(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	installPack(t, inst, writeRegistryPack(t, "registry-pack"))
	broken := filepath.Join(inst.ExtensionsDir, "broken-pack")
	require.NoError(t, os.MkdirAll(broken, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(broken, manifestFileName), []byte("not valid yaml {"), 0o644))

	reg, err := LoadRegistry(inst.ExtensionsDir, false)

	require.NoError(t, err)
	require.Len(t, reg.List(), 2)
	pack, ok := reg.Lookup("registry-pack")
	require.True(t, ok)
	assert.Equal(t, StatusOK, pack.Status)
	_, ok = reg.Lookup("missing-pack")
	assert.False(t, ok)

	modes := reg.Modes()
	require.Len(t, modes, 1)
	assert.Equal(t, "registry-mode", modes[0].Name)
	assert.Equal(t, []string{"registry-skill"}, modes[0].Skills)

	prompts := reg.PromptSources()
	require.Len(t, prompts, 1)
	assert.Equal(t, filepath.Join(inst.ExtensionsDir, "registry-pack", "prompts", "guide.md"), prompts[0].AbsolutePath)
	assert.Equal(t, "registry-section", prompts[0].Section)
	assert.Equal(t, "registry-pack", prompts[0].PackName)
}

func TestLogOrphanSubdirs_LogsOnlyUnmatchedExtensionDirs(t *testing.T) {
	t.Parallel()

	inst := newTestInstaller(t)
	installPack(t, inst, writeFakePack(t))
	reg, err := LoadRegistry(inst.ExtensionsDir, false)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Join(inst.SkillsDir, "ext-orphan"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(inst.SkillsDir, "plain-dir"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(inst.SkillsDir, "ext-file"), []byte("not a dir"), 0o644))

	var log bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&log, nil))

	LogOrphanSubdirs(inst.SkillsDir, reg, logger)

	got := log.String()
	assert.Contains(t, got, "extension.orphan.detected")
	assert.Contains(t, got, "expected_pack=orphan")
	assert.NotContains(t, got, "expected_pack=fake-pack")
	assert.NotContains(t, got, "plain-dir")
	assert.NotContains(t, got, "ext-file")
}

func TestLogOrphanSubdirs_IgnoresReadDirError(t *testing.T) {
	t.Parallel()

	skillsFile := filepath.Join(t.TempDir(), "skills")
	require.NoError(t, os.WriteFile(skillsFile, []byte("not a directory"), 0o644))
	var log bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&log, nil))

	LogOrphanSubdirs(skillsFile, &Registry{}, logger)

	assert.Empty(t, log.String())
}

func TestWriteInstalledMeta_ReturnsWriteError(t *testing.T) {
	t.Parallel()

	notDir := filepath.Join(t.TempDir(), "pack-file")
	require.NoError(t, os.WriteFile(notDir, []byte("not a directory"), 0o644))

	err := WriteInstalledMeta(notDir, &InstalledMeta{Name: "pack"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write ")
	assert.Contains(t, err.Error(), installedFileName)
}

func installPack(t *testing.T, inst *Installer, packDir string) {
	t.Helper()

	src := NewLocalSource(packDir)
	_, wc, err := inst.Inspect(context.Background(), src)
	require.NoError(t, err)
	t.Cleanup(func() { _ = wc.Cleanup() })
	require.NoError(t, inst.Install(context.Background(), src, wc, InstallOptions{}))
}

func readInstalledMeta(t *testing.T, packDir string) *InstalledMeta {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(packDir, installedFileName))
	require.NoError(t, err)
	var meta InstalledMeta
	require.NoError(t, json.Unmarshal(data, &meta))
	return &meta
}

func requireWarningContains(t *testing.T, warnings []string, want string) {
	t.Helper()

	for _, warning := range warnings {
		if strings.Contains(warning, want) {
			return
		}
	}
	require.Failf(t, "missing warning", "warnings=%v want=%q", warnings, want)
}

func writeRegistryPack(t *testing.T, name string) string {
	t.Helper()

	dir := t.TempDir()
	manifest := `schema: lango.extension/v1
name: ` + name + `
version: 0.1.0
description: Registry test pack
contents:
  skills:
    - name: registry-skill
      path: skills/registry-skill/SKILL.md
  modes:
    - name: registry-mode
      systemHint: Use registry test mode.
      skills: [registry-skill]
  prompts:
    - path: prompts/guide.md
      section: registry-section
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, manifestFileName), []byte(manifest), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "skills", "registry-skill"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "skills", "registry-skill", "SKILL.md"),
		[]byte("---\nname: registry-skill\nstatus: active\n---\nregistry skill"),
		0o644,
	))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "prompts"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "prompts", "guide.md"),
		[]byte("registry prompt"),
		0o644,
	))
	return dir
}
