package extension

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/creack/pty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cli/cliexit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/extension"
)

func executeExtensionCmd(
	t *testing.T,
	cmdArgs []string,
	cfg *config.Config,
	input io.Reader,
) (string, error, int) {
	t.Helper()

	cmd := NewExtensionCmd(func() (*config.Config, error) { return cfg, nil })
	cmd.SetContext(context.Background())

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if input != nil {
		cmd.SetIn(input)
	}
	cmd.SetArgs(cmdArgs)

	exitCode := exitOK
	err := cmd.Execute()
	if code, ok := cliexit.Code(err); ok {
		exitCode = code
	}

	return out.String(), err, exitCode
}

func newTestConfig(t *testing.T) *config.Config {
	t.Helper()
	enabled := true
	return &config.Config{
		Extensions: config.ExtensionsConfig{
			Enabled: &enabled,
			Dir:     filepath.Join(t.TempDir(), "extensions"),
		},
		Skill: config.SkillConfig{SkillsDir: filepath.Join(t.TempDir(), "skills")},
	}
}

func writeSmokePack(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	manifest := `schema: lango.extension/v1
name: smoke-pack
version: 0.1.0
description: Smoke pack
contents:
  skills:
    - name: smoke
      path: skills/smoke/SKILL.md
  modes:
    - name: smoke-mode
      systemHint: Short hint.
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "extension.yaml"), []byte(manifest), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "skills", "smoke"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skills", "smoke", "SKILL.md"),
		[]byte("---\nname: smoke\ntype: script\nstatus: active\n---\n"), 0o644))
	return dir
}

func sampleInstalledPacks(installedAt time.Time) []extension.InstalledPack {
	return []extension.InstalledPack{
		{
			Manifest: &extension.Manifest{
				Name:    "smoke-pack",
				Version: "0.1.0",
				Author:  "Ops Team",
			},
			Meta: &extension.InstalledMeta{
				InstalledAt:    installedAt,
				Source:         "local:/tmp/smoke-pack",
				ManifestSHA256: "abc123",
			},
			Status: extension.StatusOK,
		},
		{
			Status: extension.StatusBroken,
		},
	}
}

func TestInspectJSONOutput(t *testing.T) {
	t.Parallel()

	cfg := newTestConfig(t)
	packDir := writeSmokePack(t)

	cmd := NewExtensionCmd(func() (*config.Config, error) { return cfg, nil })
	cmd.SetContext(context.Background())

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"inspect", packDir, "--output", "json"})

	require.NoError(t, cmd.Execute())

	var payload map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &payload))
	assert.Equal(t, "smoke-pack", payload["name"])
	assert.Equal(t, "0.1.0", payload["version"])
	assert.Contains(t, payload, "manifest_sha256")
	assert.Contains(t, payload, "planned_writes")
}

func TestListEmptyRegistry(t *testing.T) {
	t.Parallel()

	cfg := newTestConfig(t)
	cmd := NewExtensionCmd(func() (*config.Config, error) { return cfg, nil })
	cmd.SetContext(context.Background())

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"list", "--output", "json"})

	require.NoError(t, cmd.Execute())
	assert.Equal(t, "[]\n", out.String())
}

func TestListWithInstalledPack(t *testing.T) {
	t.Parallel()

	cfg := newTestConfig(t)
	inst := &extension.Installer{
		ExtensionsDir: cfg.Extensions.Dir,
		SkillsDir:     cfg.Skill.SkillsDir,
	}
	src := extension.NewLocalSource(writeSmokePack(t))
	_, wc, err := inst.Inspect(context.Background(), src)
	require.NoError(t, err)
	require.NoError(t, inst.Install(context.Background(), src, wc, extension.InstallOptions{}))
	_ = wc.Cleanup()

	cmd := NewExtensionCmd(func() (*config.Config, error) { return cfg, nil })
	cmd.SetContext(context.Background())

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"list", "--output", "json"})

	require.NoError(t, cmd.Execute())

	var rows []map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &rows))
	require.Len(t, rows, 1)
	assert.Equal(t, "smoke-pack", rows[0]["name"])
	assert.Equal(t, "0.1.0", rows[0]["version"])
	assert.Equal(t, "ok", rows[0]["status"])
}

func TestRenderList_EmptyOutputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		format outputFormat
		want   string
	}{
		{
			name:   "json",
			format: outputJSON,
			want:   "[]\n",
		},
		{
			name:   "plain",
			format: outputPlain,
			want:   "",
		},
		{
			name:   "table",
			format: outputTable,
			want:   "NAME                     VERSION    AUTHOR               INSTALLED              STATUS\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer
			require.NoError(t, renderList(&out, nil, tt.format))
			assert.Equal(t, tt.want, out.String())
		})
	}
}

func TestRenderList_NonEmptyOutputs(t *testing.T) {
	t.Parallel()

	installedAt := time.Date(2026, 5, 21, 8, 9, 10, 0, time.UTC)
	packs := sampleInstalledPacks(installedAt)

	t.Run("json includes manifest meta and broken row status", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		require.NoError(t, renderList(&out, packs, outputJSON))

		var rows []map[string]any
		require.NoError(t, json.Unmarshal(out.Bytes(), &rows))
		require.Len(t, rows, 2)
		assert.Equal(t, "smoke-pack", rows[0]["name"])
		assert.Equal(t, "0.1.0", rows[0]["version"])
		assert.Equal(t, "Ops Team", rows[0]["author"])
		assert.Equal(t, "local:/tmp/smoke-pack", rows[0]["source"])
		assert.Equal(t, "abc123", rows[0]["manifest_sha256"])
		assert.Equal(t, "ok", rows[0]["status"])
		assert.Equal(t, "broken", rows[1]["status"])
	})

	t.Run("plain writes tab-separated rows", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		require.NoError(t, renderList(&out, packs, outputPlain))
		assert.Equal(t, "smoke-pack\t0.1.0\tok\n\t\tbroken\n", out.String())
	})

	t.Run("table includes header install time and statuses", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		require.NoError(t, renderList(&out, packs, outputTable))
		got := out.String()
		assert.Contains(t, got, "NAME")
		assert.Contains(t, got, "smoke-pack")
		assert.Contains(t, got, "0.1.0")
		assert.Contains(t, got, "Ops Team")
		assert.Contains(t, got, installedAt.Format(time.RFC3339))
		assert.Contains(t, got, "broken")
	})
}

func TestRenderList_UnknownFormatReturnsError(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	err := renderList(&out, nil, outputFormat(99))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unhandled output format")
	assert.Empty(t, out.String())
}

func TestResolveOutput_ExplicitAndDefaultBehavior(t *testing.T) {
	t.Parallel()

	assert.Equal(t, outputJSON, resolveOutput("json", &bytes.Buffer{}))
	assert.Equal(t, outputPlain, resolveOutput("plain", &bytes.Buffer{}))
	assert.Equal(t, outputTable, resolveOutput("table", &bytes.Buffer{}))
	assert.Equal(t, outputPlain, resolveOutput("", &bytes.Buffer{}))
}

func TestResolveOutput_DefaultsToTableForTTY(t *testing.T) {
	t.Parallel()

	ptmx, tty, err := pty.Open()
	require.NoError(t, err)
	t.Cleanup(func() { _ = ptmx.Close() })
	t.Cleanup(func() { _ = tty.Close() })

	assert.Equal(t, outputTable, resolveOutput("", tty))
}

func TestUnknownOutputFormatRejected(t *testing.T) {
	t.Parallel()

	err := validateOutput("yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown output format")
}

func TestExpandTildeAbs_ExpandsHomeAndRelativePaths(t *testing.T) {
	t.Parallel()

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(home, "lango", "extensions"), expandTildeAbs("~/lango/extensions"))
	assert.True(t, filepath.IsAbs(expandTildeAbs("relative/extensions")))
	assert.True(t, strings.HasSuffix(expandTildeAbs("relative/extensions"), filepath.Join("relative", "extensions")))
}

func TestInstallerFor_ConfigErrorBranches(t *testing.T) {
	t.Parallel()

	t.Run("extensions disabled", func(t *testing.T) {
		t.Parallel()

		cfg := newTestConfig(t)
		disabled := false
		cfg.Extensions.Enabled = &disabled

		inst, err := installerFor(cfg)

		require.ErrorIs(t, err, extension.ErrNotEnabled)
		assert.Nil(t, inst)
	})

	t.Run("skills dir missing", func(t *testing.T) {
		t.Parallel()

		cfg := newTestConfig(t)
		cfg.Skill.SkillsDir = ""

		inst, err := installerFor(cfg)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "skills.skillsDir is not configured")
		assert.Nil(t, inst)
	})

	t.Run("returns absolute resolved dirs", func(t *testing.T) {
		t.Parallel()

		cfg := newTestConfig(t)

		inst, err := installerFor(cfg)

		require.NoError(t, err)
		require.NotNil(t, inst)
		assert.True(t, filepath.IsAbs(inst.ExtensionsDir))
		assert.True(t, filepath.IsAbs(inst.SkillsDir))
	})
}

func TestInstallCmd_ConfirmUsesCommandStreams(t *testing.T) {
	t.Parallel()

	cfg := newTestConfig(t)
	packDir := writeSmokePack(t)

	out, err, exitCode := executeExtensionCmd(
		t,
		[]string{"install", packDir},
		cfg,
		bytes.NewBufferString("y\n"),
	)
	require.NoError(t, err)
	assert.Equal(t, exitOK, exitCode)
	assert.Contains(t, out, "Install this pack? [y/N]: ")
	assert.Contains(t, out, "installed smoke-pack@0.1.0")
	_, statErr := os.Stat(filepath.Join(cfg.Extensions.Dir, "smoke-pack", ".installed"))
	require.NoError(t, statErr)
}

func TestInstallCmd_DenyCancelsWithoutWritingFiles(t *testing.T) {
	t.Parallel()

	cfg := newTestConfig(t)
	packDir := writeSmokePack(t)

	out, err, exitCode := executeExtensionCmd(
		t,
		[]string{"install", packDir},
		cfg,
		bytes.NewBufferString("n\n"),
	)
	require.Error(t, err)
	assert.Equal(t, exitUserDeclined, exitCode)
	assert.True(t, cliexit.Silent(err))
	assert.Contains(t, out, "Install this pack? [y/N]: ")
	assert.Contains(t, out, "install cancelled by user")
	_, statErr := os.Stat(filepath.Join(cfg.Extensions.Dir, "smoke-pack"))
	require.ErrorIs(t, statErr, os.ErrNotExist)
}

func TestInstallCmd_NonTTYWithoutYesReturnsGuidance(t *testing.T) {
	t.Parallel()

	cfg := newTestConfig(t)
	packDir := writeSmokePack(t)
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { _ = reader.Close() })
	t.Cleanup(func() { _ = writer.Close() })

	out, cmdErr, exitCode := executeExtensionCmd(t, []string{"install", packDir}, cfg, reader)
	require.Error(t, cmdErr)
	assert.Equal(t, exitUserDeclined, exitCode)
	assert.False(t, cliexit.Silent(cmdErr))
	assert.NotContains(t, out, "stdin is not a TTY; pass --yes for scripted runs")
	assert.Contains(t, cmdErr.Error(), "stdin is not a TTY; pass --yes for scripted runs")
}

func TestInspectCmd_InvalidManifestReturnsStructuredUserError(t *testing.T) {
	t.Parallel()

	cfg := newTestConfig(t)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "extension.yaml"), []byte("schema: lango.extension/v1\n"), 0o644))

	out, err, exitCode := executeExtensionCmd(t, []string{"inspect", dir}, cfg, nil)

	require.Error(t, err)
	assert.Equal(t, exitUserError, exitCode)
	assert.NotContains(t, out, "error:")
	assert.Contains(t, err.Error(), "invalid pack name")
}

func TestInspectCmd_UnknownOutputFormatReturnsStructuredInternalError(t *testing.T) {
	t.Parallel()

	cfg := newTestConfig(t)
	packDir := writeSmokePack(t)

	out, err, exitCode := executeExtensionCmd(t, []string{"inspect", packDir, "--output", "yaml"}, cfg, nil)

	require.Error(t, err)
	assert.Equal(t, exitInternal, exitCode)
	assert.NotContains(t, out, "error:")
	assert.Contains(t, err.Error(), "unknown output format")
}

func TestListCmd_LoadConfigErrorReturnsStructuredInternalError(t *testing.T) {
	t.Parallel()

	loadErr := errors.New("config unavailable")
	cmd := NewExtensionCmd(func() (*config.Config, error) { return nil, loadErr })
	cmd.SetContext(context.Background())

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	code, ok := cliexit.Code(err)

	require.Error(t, err)
	require.True(t, ok)
	assert.Equal(t, exitInternal, code)
	assert.NotContains(t, out.String(), "error:")
	assert.Contains(t, err.Error(), "load config: config unavailable")
}

func TestListCmd_RegistryReadErrorReturnsStructuredInternalError(t *testing.T) {
	t.Parallel()

	cfg := newTestConfig(t)
	registryFile := filepath.Join(t.TempDir(), "extensions-file")
	require.NoError(t, os.WriteFile(registryFile, []byte("not a directory"), 0o644))
	cfg.Extensions.Dir = registryFile

	out, err, exitCode := executeExtensionCmd(t, []string{"list"}, cfg, nil)

	require.Error(t, err)
	assert.Equal(t, exitInternal, exitCode)
	assert.NotContains(t, out, "error:")
	assert.Contains(t, err.Error(), "read extensions dir")
}

func TestListCmd_UnknownOutputFormatSkipsConfigLoad(t *testing.T) {
	t.Parallel()

	loaderCalled := false
	cmd := NewExtensionCmd(func() (*config.Config, error) {
		loaderCalled = true
		return newTestConfig(t), nil
	})
	cmd.SetContext(context.Background())

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"list", "--output", "yaml"})

	err := cmd.Execute()
	code, ok := cliexit.Code(err)

	require.Error(t, err)
	require.True(t, ok)
	assert.Equal(t, exitInternal, code)
	assert.False(t, loaderCalled)
	assert.NotContains(t, out.String(), "error:")
	assert.Contains(t, err.Error(), "unknown output format")
}

func TestRemoveCmd_ConfirmUsesCommandStreams(t *testing.T) {
	t.Parallel()

	cfg := newTestConfig(t)
	inst := &extension.Installer{
		ExtensionsDir: cfg.Extensions.Dir,
		SkillsDir:     cfg.Skill.SkillsDir,
	}
	src := extension.NewLocalSource(writeSmokePack(t))
	_, wc, err := inst.Inspect(context.Background(), src)
	require.NoError(t, err)
	require.NoError(t, inst.Install(context.Background(), src, wc, extension.InstallOptions{}))
	_ = wc.Cleanup()

	out, err, exitCode := executeExtensionCmd(
		t,
		[]string{"remove", "smoke-pack"},
		cfg,
		bytes.NewBufferString("y\n"),
	)
	require.NoError(t, err)
	assert.Equal(t, exitOK, exitCode)
	assert.Contains(t, out, "Will delete:")
	assert.Contains(t, out, "Remove pack? [y/N]: ")
	assert.Contains(t, out, "removed smoke-pack")
	_, statErr := os.Stat(filepath.Join(cfg.Extensions.Dir, "smoke-pack"))
	require.ErrorIs(t, statErr, os.ErrNotExist)
}

func TestRemoveCmd_MissingPackReturnsStructuredUserError(t *testing.T) {
	t.Parallel()

	cfg := newTestConfig(t)

	out, err, exitCode := executeExtensionCmd(
		t,
		[]string{"remove", "missing", "--yes"},
		cfg,
		nil,
	)

	require.Error(t, err)
	assert.Equal(t, exitUserError, exitCode)
	assert.Contains(t, out, "Will delete:")
	assert.Contains(t, err.Error(), "pack not found")
}

func TestRemoveCmd_LoadConfigErrorReturnsStructuredInternalError(t *testing.T) {
	t.Parallel()

	loadErr := errors.New("config unavailable")
	cmd := NewExtensionCmd(func() (*config.Config, error) { return nil, loadErr })
	cmd.SetContext(context.Background())

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"remove", "smoke-pack", "--yes"})

	err := cmd.Execute()
	code, ok := cliexit.Code(err)

	require.Error(t, err)
	require.True(t, ok)
	assert.Equal(t, exitInternal, code)
	assert.NotContains(t, out.String(), "error:")
	assert.Contains(t, err.Error(), "load config: config unavailable")
}

func TestRemoveCmd_DisabledExtensionsReturnUserErrorBeforePreview(t *testing.T) {
	t.Parallel()

	cfg := newTestConfig(t)
	disabled := false
	cfg.Extensions.Enabled = &disabled

	out, err, exitCode := executeExtensionCmd(
		t,
		[]string{"remove", "smoke-pack", "--yes"},
		cfg,
		nil,
	)

	require.Error(t, err)
	assert.Equal(t, exitUserError, exitCode)
	assert.Empty(t, out)
	assert.ErrorIs(t, err, extension.ErrNotEnabled)
}
