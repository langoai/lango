package extension

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/extension"
)

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

func TestUnknownOutputFormatRejected(t *testing.T) {
	t.Parallel()

	err := validateOutput("yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown output format")
}
