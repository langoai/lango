package extension

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/extension"
)

type exitPanic struct {
	code int
}

var extensionExitMu sync.Mutex

func executeExtensionCmd(
	t *testing.T,
	cmdArgs []string,
	cfg *config.Config,
	input io.Reader,
) (string, error, int) {
	t.Helper()

	extensionExitMu.Lock()
	defer extensionExitMu.Unlock()

	origExit := extensionExit
	extensionExit = func(code int) {
		panic(exitPanic{code: code})
	}
	defer func() {
		extensionExit = origExit
	}()

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
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				if got, ok := r.(exitPanic); ok {
					exitCode = got.code
					return
				}
				panic(r)
			}
		}()
		err = cmd.Execute()
	}()

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
	require.NoError(t, err)
	assert.Equal(t, exitUserDeclined, exitCode)
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
	require.NoError(t, cmdErr)
	assert.Equal(t, exitUserDeclined, exitCode)
	assert.Contains(t, out, "stdin is not a TTY; pass --yes for scripted runs")
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
